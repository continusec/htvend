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

package htvend

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/continusec/htvend/internal/blobs"
	"github.com/continusec/htvend/internal/lockfile"
	"github.com/continusec/htvend/internal/registryauthclient"
	"github.com/hashicorp/go-multierror"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

var _ flags.Commander = &VerifyCommand{}

type VerifyCommand struct {
	ManifestOptions
	FetchOptions

	Fetch  bool `long:"fetch" description:"If set, fetch missing assets"`
	Repair bool `long:"repair" description:"If set, replace any missing assets with new versions currently found (implies fetch). May still require a rebuild afterwards (e.g. if they trigger other new calls)."`
}

func (rc *VerifyCommand) Execute(args []string) (retErr error) {
	mf, err := rc.ManifestOptions.MakeManifestFile(&manifestContextOptions{
		Writable:       rc.Repair,
		AllowOverwrite: rc.Repair,
		NoCacheList:    rc.FetchOptions.NoCache,
	})
	if err != nil {
		return fmt.Errorf("error getting manifest file: %w", err)
	}
	defer func() {
		if err := mf.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	bs, err := rc.ManifestOptions.MakeBlobStore(rc.Fetch)
	if err != nil {
		return fmt.Errorf("error creating blob store: %w", err)
	}

	return doValidate(&validateCtx{
		Assets:         mf,
		Blobs:          bs,
		FailIfMissing:  !rc.Fetch && !rc.Repair,
		FetchIfMissing: rc.Fetch || rc.Repair,
		RepairIfWrong:  rc.Repair,
		ValidateSHA256: true,
		HeadersToCache: rc.FetchOptions.CacheHeaderMap(),
	})
}

type validateCtx struct {
	Assets    *lockfile.File
	Blobs     blobs.Store
	ExportDir string

	HeadersToCache map[string]bool // if repair

	FailIfMissing  bool
	FetchIfMissing bool
	RepairIfWrong  bool
	ValidateSHA256 bool
	SaveToExport   bool
}

// copy reader to all writers, returning any errors
// if no writers, then nothing is read an no error
func copyToWriters(r io.Reader, writers []io.Writer) error {
	var actualWriter io.Writer
	for _, w := range writers {
		if actualWriter == nil {
			actualWriter = w
		} else {
			r = io.TeeReader(r, w)
		}
	}
	if actualWriter == nil {
		return nil
	}
	_, err := io.Copy(actualWriter, r)
	return err
}

func doValidate(vctx *validateCtx) error {
	if vctx.SaveToExport {
		if err := os.MkdirAll(vctx.ExportDir, 0o755); err != nil {
			return fmt.Errorf("error creating export dir: %w", err)
		}
	}

	type toBeFetched struct {
		K       *url.URL
		V       lockfile.BlobInfo
		NewHash []byte
	}

	var missingList []toBeFetched
	var wrongHashList []toBeFetched
	if err := vctx.Assets.ForEach(func(k *url.URL, v lockfile.BlobInfo) (retErr error) {
		logrus.Infof("Verifying %s...", k)

		expectedH, err := hex.DecodeString(v.Sha256)
		if err != nil {
			return fmt.Errorf("error decoding sha key for %s: %w", k.String(), err)
		}

		r, err := vctx.Blobs.Get(expectedH)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				missingList = append(missingList, toBeFetched{
					K: k,
					V: v,
				})
				return nil // since we handle later, as we won't want to edit map while iterating it
			}
			return fmt.Errorf("unknown error finding blob for %s: %w", k.String(), err)
		}

		defer func() {
			if err := r.Close(); err != nil && retErr == nil {
				retErr = fmt.Errorf("error in clean-up close in validate for %s: %w", k.String(), err)
			}
		}()

		var writers []io.Writer
		if vctx.SaveToExport {
			var fileWriter io.WriteCloser
			if fileWriter, err = os.Create(filepath.Join(vctx.ExportDir, hex.EncodeToString(expectedH))); err != nil {
				return fmt.Errorf("error opening file to export to: %w", err)
			}
			writers = append(writers, fileWriter)
			defer func() {
				if err := fileWriter.Close(); err != nil && retErr == nil {
					retErr = fmt.Errorf("error closing export file: %w", err)
				}
			}()
		}

		if vctx.ValidateSHA256 {
			h2 := sha256.New()
			writers = append(writers, h2)
			defer func() {
				actualH := h2.Sum(nil)
				if !bytes.Equal(expectedH, actualH) {
					wrongHashList = append(wrongHashList, toBeFetched{
						K:       k,
						V:       v,
						NewHash: actualH,
					})
				}
			}()
		}

		return copyToWriters(r, writers)
	}); err != nil {
		return fmt.Errorf("error in verification: %w", err)
	}

	var rv error
	var client *http.Client
	for _, missing := range missingList {
		switch {
		case vctx.FailIfMissing:
			rv = multierror.Append(rv, fmt.Errorf("missing asset: %s", missing.K.Redacted()))
		case vctx.FetchIfMissing:
			if client == nil {
				client = &http.Client{
					Transport: registryauthclient.NewClient(http.DefaultTransport),
				}
			}
			if err := fetchAndSaveBlob(vctx.Assets, vctx.Blobs, http.MethodGet, nil, missing.K, client, vctx.HeadersToCache, nil, nil); err != nil {
				return fmt.Errorf("error fetching %s: %w", missing.K.Redacted(), err)
			}
		}
	}

	for _, wrongHash := range wrongHashList {
		rv = multierror.Append(rv, fmt.Errorf("wrong hash for: %s expected: %s have %s", wrongHash.K.Redacted(), wrongHash.V.Sha256, hex.EncodeToString(wrongHash.NewHash)))
	}

	return rv
}
