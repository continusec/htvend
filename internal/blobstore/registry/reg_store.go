// Copyright 2025 Continusec Pty Ltd.
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

package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/continusec/htvend/internal/blobstore"
	"github.com/sirupsen/logrus"
)

var _ blobstore.Store = &RegistryStore{}

type RegistryStore struct {
	base     string
	writable bool
	client   *http.Client
}

func NewRegistryStore(url string, writable bool) *RegistryStore {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return &RegistryStore{
		base:     url,
		writable: writable,
		client:   http.DefaultClient,
	}
}

func (r *RegistryStore) Exists(k []byte) (bool, error) {
	resp, err := r.client.Head(r.base + "blobs/sha256:" + hex.EncodeToString(k))
	if err != nil {
		return false, fmt.Errorf("error checking blob existence from registry store: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		bb, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logrus.Debugf("error response from HEAD blob: %s", bb)
		return false, fmt.Errorf("bad status code in registry store for HEAD blob: %d", resp.StatusCode)
	}
}

func (r *RegistryStore) Get(k []byte) (io.ReadCloser, error) {
	resp, err := r.client.Get(r.base + "blobs/sha256:" + hex.EncodeToString(k))
	if err != nil {
		return nil, fmt.Errorf("error fetching blob from registry store: %w", err)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return resp.Body, nil
	case http.StatusNotFound:
		resp.Body.Close()
		return nil, fmt.Errorf("can't find blob in registry: %w", blobstore.ErrBlobNotExist)
	default:
		bb, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logrus.Debugf("error response from GET blob: %s", bb)
		return nil, fmt.Errorf("bad status code in registry store for blob: %d", resp.StatusCode)
	}
}

func (r *RegistryStore) Put() (blobstore.ContentAddressableBlob, error) {
	if !r.writable {
		return nil, fmt.Errorf("attempt to write to unwriteable blobstore")
	}

	// do a POST to uploads/
	resp, err := r.client.Post(r.base+"blobs/uploads/", "", nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching blob from registry store: %w", err)
	}
	switch resp.StatusCode {
	case http.StatusAccepted:
		err = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("unexpected error on close: %w", err)
		}

		urlToUpload := resp.Header.Get("Location")
		if len(urlToUpload) == 0 {
			return nil, fmt.Errorf("no location returned therefore cannot put blob!")
		}

		return r.newPutAddressableBlob(urlToUpload)
	default:
		bb, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logrus.Debugf("errors response from POST blob: %s", bb)
		return nil, fmt.Errorf("bad status code in registry store for putting blob: %d", resp.StatusCode)
	}

}

func (r *RegistryStore) Destroy() error {
	return fmt.Errorf("destroy not implemented for registry store")
}

func (r *RegistryStore) RemoveExcept(map[string]bool) error {
	return fmt.Errorf("RemoveExcept not implemented for registry store")
}

func (r *RegistryStore) newPutAddressableBlob(u string) (*blobBeingUploaded, error) {
	return &blobBeingUploaded{
		digest: sha256.New(),
		url:    u,
		client: r.client,
	}, nil
}

type blobBeingUploaded struct {
	url    string
	client *http.Client
	buf    []byte
	digest hash.Hash // set to nil when done
	bw     int
}

func (b *blobBeingUploaded) Write(bb []byte) (int, error) {
	if len(bb) == 0 {
		// special-case else we have to get clever later...
		return 0, nil
	}
	req, err := http.NewRequest(http.MethodPatch, b.url, io.TeeReader(bytes.NewReader(bb), b.digest))
	if err != nil {
		return 0, fmt.Errorf("error making PATCH req: %w", err)
	}
	req.Header.Set("Content-Length", strconv.Itoa(len(bb)))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/*", b.bw, b.bw+len(bb)-1))
	resp, err := b.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error doing PATCH: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return 0, fmt.Errorf("bad status code: %s", resp.Status)
	}
	newLoc := resp.Header.Get("Location")
	if len(newLoc) == 0 {
		return 0, fmt.Errorf("no location set on response!")
	}
	b.url = newLoc
	b.bw += len(bb)
	return len(bb), nil
}

func (b *blobBeingUploaded) Commit() ([]byte, error) {
	digest := b.digest.Sum(nil)

	req, err := http.NewRequest(http.MethodPut, b.url+"&digest=sha256:"+hex.EncodeToString(digest), nil)
	if err != nil {
		return nil, fmt.Errorf("error making PUT req: %w", err)
	}
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing PUT: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		br, _ := io.ReadAll(resp.Body)
		logrus.Debugf("COMMIT server error resp: %s", br)
		return nil, fmt.Errorf("bad status code from COMMIT: %s", resp.Status)
	}
	b.digest = nil // signifies done
	return digest, nil
}

func (b *blobBeingUploaded) Cleanup() error {
	if b.digest == nil {
		return nil // no-op we are already successfully committed or cleaned up
	}

	req, err := http.NewRequest(http.MethodDelete, b.url, nil)
	if err != nil {
		return fmt.Errorf("error making DELETE req: %w", err)
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("error doing PATCH: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("bad status code: %s", resp.Status)
	}

	b.digest = nil
	return nil
}
