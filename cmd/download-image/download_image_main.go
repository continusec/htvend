package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/containers/image/v5/manifest"
	"github.com/continusec/htvend/internal/registryauthclient"
	"github.com/hashicorp/go-multierror"
	"github.com/opencontainers/go-digest"
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
	mBySha, shaCT, err := pullManifest(client, img.ToManifestUrlWithKey(hm[:]))
	if err != nil {
		return fmt.Errorf("error pulling manifest by tag: %w", err)
	}

	if !bytes.Equal(mByTag, mBySha) {
		fmt.Printf("by tag: %s\nby sha: %s\n", mByTag, mBySha)
		return fmt.Errorf("that's weird, manifest changed!")
	}

	ml, err := manifest.ListFromBlob(mBySha, shaCT)
	if err != nil {
		return fmt.Errorf("error parsing manifest: %w", err)
	}

	ourDigest, err := ml.ChooseInstance(nil)
	if err != nil {
		return fmt.Errorf("error choosing operating system digest: %w", err)
	}

	ourPlatformManifestBytes, ourPlatformManifestCT, err := pullManifest(client, img.ToManifestUrlWithDigest(ourDigest))
	if err != nil {
		return fmt.Errorf("can't pull actual platform manifest: %w", err)
	}

	ourPlatformManifest, err := manifest.FromBlob(ourPlatformManifestBytes, ourPlatformManifestCT)
	if err != nil {
		return fmt.Errorf("can't parse platform manifest: %w", err)
	}

	if err := pullBlob(client, img.ToBlobUrlWithDigest(ourPlatformManifest.ConfigInfo().Digest)); err != nil {
		return fmt.Errorf("error pulling config: %w", err)
	}

	for _, li := range ourPlatformManifest.LayerInfos() {
		if err := pullBlob(client, img.ToBlobUrlWithDigest(li.Digest)); err != nil {
			return fmt.Errorf("error pulling layer: %w", err)
		}
	}

	return nil
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
