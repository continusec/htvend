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
	"encoding/hex"
	"fmt"
	"io"
	"net/url"

	blobs "github.com/continusec/htvend/internal/blobstore"
	"github.com/continusec/htvend/internal/jobs"
	"github.com/continusec/htvend/internal/lockfile"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

var _ flags.Commander = &ExportCommand{}

type ExportCommand struct {
	ManifestOptions

	Dest CacheOptions `group:"Destination blob store" namespace:"dest"`
}

func ensureBlobExported(src, dst blobs.Store, expectedH []byte) (retErr error) {
	exists, err := dst.Exists(expectedH)
	if err != nil {
		return fmt.Errorf("error checking if destination exists: %w", err)
	}
	if exists {
		logrus.Infof("%s already exists in destination blobstore, skipping fetch", hex.EncodeToString(expectedH))
		return nil
	}

	logrus.Infof("Fetching %s from upstream blobstore...", hex.EncodeToString(expectedH))

	// else we must fetch, write and check hash
	srcBlob, err := src.Get(expectedH)
	if err != nil {
		return fmt.Errorf("error fetching from upstream blobstore: %w", err)
	}

	// create file to write
	dstBlob, err := dst.Put()
	if err != nil {
		return fmt.Errorf("error creating destination blobstore to write: %w", err)
	}
	defer func() {
		if err := dstBlob.Cleanup(); err != nil && retErr == nil {
			retErr = fmt.Errorf("error committing to dest: %w", err)
		}
	}()

	// copy src to dest
	if _, err := io.Copy(dstBlob, srcBlob); err != nil {
		return fmt.Errorf("error reading from src to dst: %w", err)
	}

	actualH, err := dstBlob.Commit()
	if err != nil {
		return fmt.Errorf("error comming destination blob: %w", err)
	}

	if !bytes.Equal(expectedH, actualH) {
		return fmt.Errorf("actual hash (%s) received differs from desired hash (%s) for blob", hex.EncodeToString(actualH), hex.EncodeToString(expectedH))
	}

	return nil
}

func (rc *ExportCommand) Execute(args []string) (retErr error) {
	mf, err := rc.ManifestOptions.MakeManifestFile(&manifestContextOptions{})
	if err != nil {
		return fmt.Errorf("error getting manifest file: %w", err)
	}
	defer func() {
		if err := mf.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	srcBs, err := rc.ManifestOptions.MakeBlobStore(false)
	if err != nil {
		return fmt.Errorf("error creating source blob store: %w", err)
	}

	dstBs, err := rc.Dest.MakeBlobStore(true)
	if err != nil {
		return fmt.Errorf("error creating destination blob store: %w", err)
	}

	// first dedupe any hashes
	neededCanonShas := make(map[string]bool)
	if err := mf.ForEach(func(k *url.URL, v lockfile.BlobInfo) error {
		expectedH, err := hex.DecodeString(v.Sha256)
		if err != nil {
			return fmt.Errorf("error decoding hash: %w", err)
		}
		neededCanonShas[hex.EncodeToString(expectedH)] = true
		return nil
	}); err != nil {
		return fmt.Errorf("error iterating blobs: %w", err)
	}

	// now handle each
	mt := jobs.NewMultiTasker()
	for canonSha := range neededCanonShas {
		expectedH, err := hex.DecodeString(canonSha)
		if err != nil {
			return fmt.Errorf("error decoding hash: %w", err)
		}
		mt.Queue(func() error {
			return ensureBlobExported(srcBs, dstBs, expectedH)
		})
	}

	return mt.Wait(func(err error) {
		logrus.Errorf("error during parallel job: %v", err)
	})
}
