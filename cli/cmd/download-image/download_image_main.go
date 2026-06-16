// Copyright 2026 Continusec Pty Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/continusec/htvend/internal/registryauthclient"
	"github.com/hashicorp/go-multierror"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

type ImageName struct {
	Host string
	Org  string
	Name string
	Tag  string
}

func (in *ImageName) ToManifestURL() string {
	return fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s", in.Host, in.Org, in.Name, in.Tag) // TODO, how to escape?
}

func (in *ImageName) ToManifestUrlWithKey(key []byte) string {
	return fmt.Sprintf("https://%s/v2/%s/%s/manifests/sha256:%s", in.Host, in.Org, in.Name, hex.EncodeToString(key)) // TODO, how to escape?
}

func (in *ImageName) ToManifestUrlWithDigest(digest digest.Digest) string {
	return fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s", in.Host, in.Org, in.Name, digest) // TODO, how to escape?
}

func (in *ImageName) ToBlobUrlWithDigest(digest digest.Digest) string {
	return fmt.Sprintf("https://%s/v2/%s/%s/blobs/%s", in.Host, in.Org, in.Name, digest) // TODO, how to escape?
}

func pullBlob(client *http.Client, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error in GET: %w", err)
	}
	defer resp.Body.Close()
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		return err
	}
	return nil
}

func pullManifest(client *http.Client, url string) ([]byte, string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("error in GET: %w", err)
	}
	defer resp.Body.Close()
	rv, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return rv, resp.Header.Get("Content-Type"), nil
}

func downloadImage(client *http.Client, name string) error {
	img, err := NewImageName(name)
	if err != nil {
		return fmt.Errorf("error parsing name: %w", err)
	}

	mByTag, _, err := pullManifest(client, img.ToManifestURL())
	if err != nil {
		return fmt.Errorf("error pulling manifest by tag: %w", err)
	}

	// hash it
	hm := sha256.Sum256(mByTag)

	// now pull same but via SHA256 URL, as some clients need this
	mBySha, _, err := pullManifest(client, img.ToManifestUrlWithKey(hm[:]))
	if err != nil {
		return fmt.Errorf("error pulling manifest by tag: %w", err)
	}

	if !bytes.Equal(mByTag, mBySha) {
		fmt.Printf("by tag: %s\nby sha: %s\n", mByTag, mBySha)
		return fmt.Errorf("that's weird, manifest changed!")
	}

	var ml imgspecv1.Index
	if err := json.Unmarshal(mBySha, &ml); err != nil {
		return fmt.Errorf("error parsing manifest list: %w", err)
	}

	ourDigest, err := chooseInstance(ml.Manifests)
	if err != nil {
		return fmt.Errorf("error choosing operating system digest: %w", err)
	}

	ourPlatformManifestBytes, _, err := pullManifest(client, img.ToManifestUrlWithDigest(ourDigest))
	if err != nil {
		return fmt.Errorf("can't pull actual platform manifest: %w", err)
	}

	var ourPlatformManifest imgspecv1.Manifest
	if err := json.Unmarshal(ourPlatformManifestBytes, &ourPlatformManifest); err != nil {
		return fmt.Errorf("can't parse platform manifest: %w", err)
	}

	if err := pullBlob(client, img.ToBlobUrlWithDigest(ourPlatformManifest.Config.Digest)); err != nil {
		return fmt.Errorf("error pulling config: %w", err)
	}

	for _, li := range ourPlatformManifest.Layers {
		if err := pullBlob(client, img.ToBlobUrlWithDigest(li.Digest)); err != nil {
			return fmt.Errorf("error pulling layer: %w", err)
		}
	}

	return nil
}

// chooseInstance picks the manifest descriptor matching the host platform,
// replacing containers/image's manifest.List.ChooseInstance(nil). Docker
// schema2 manifest lists and OCI image indexes share the same JSON shape, so
// both unmarshal into imgspecv1.Index.
func chooseInstance(manifests []imgspecv1.Descriptor) (digest.Digest, error) {
	wantVariant := ""
	if runtime.GOARCH == "arm" {
		wantVariant = "v7" // matches containers/image's default for 32-bit arm
	}
	for _, m := range manifests {
		p := m.Platform
		if p == nil || p.OS != runtime.GOOS || p.Architecture != runtime.GOARCH {
			continue
		}
		// Treat an unset variant on either side as compatible (e.g. arm64/v8).
		if wantVariant != "" && p.Variant != "" && p.Variant != wantVariant {
			continue
		}
		return m.Digest, nil
	}
	return "", fmt.Errorf("no image found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func NewImageName(ref string) (ImageName, error) {
	var rv ImageName
	rv.Host = "docker.io"
	rv.Org = "library"
	rv.Tag = "latest"

	var img string
	bits := strings.Split(ref, "/")
	switch len(bits) {
	case 1:
		img = bits[0]
	case 2:
		rv.Org, img = bits[0], bits[1]
	case 3:
		rv.Host, rv.Org, img = bits[0], bits[1], bits[2]
	default:
		return rv, fmt.Errorf("unexpected name format: %s", ref)
	}

	imgBits := strings.Split(img, ":")
	switch len(imgBits) {
	case 1:
		rv.Name = imgBits[0]
	case 2:
		rv.Name, rv.Tag = imgBits[0], imgBits[1]
	default:
		return rv, fmt.Errorf("unexpected name format: %s", ref)
	}

	if rv.Host == "docker.io" {
		rv.Host = "registry-1.docker.io" // why?
	}

	return rv, nil
}

func main() {
	if err := func() error {
		client := http.Client{
			Transport: registryauthclient.NewClient(http.DefaultTransport),
		}
		var rv error
		for _, imgName := range os.Args[1:] {
			if err := downloadImage(&client, imgName); err != nil {
				rv = multierror.Append(rv, err)
			}
		}
		return rv
	}(); err != nil {
		logrus.Fatalf("error: %v", err)
	}
}
