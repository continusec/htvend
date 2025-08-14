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

package caf

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

const (
	cafStateInitial = iota
	cafStateWriting
	cafStateClosed
	cafStateCommitted
	cafStateCanceled
)

type ContentAddressableFile struct {
	fn FilenameResolver // set during init

	mw io.Writer // created on first Write()
	tf *os.File  // created on first Write()
	dg hash.Hash // created on first Write()

	size int // updated in Write
}

type FilenameResolver func(digest []byte) string

// NewContentAddressableFile creates a writeable file in a directory
// with a temp name that is renamed to the canonical hash only when it
// has been fully written out successfully. Temp file not created until
// first write.
func NewContentAddressableFile(fn FilenameResolver) *ContentAddressableFile {
	return &ContentAddressableFile{
		fn: fn,
	}
}

func (caf *ContentAddressableFile) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return caf.writeAlways(p)
}

// make sure file exists, then write to it
func (caf *ContentAddressableFile) writeAlways(p []byte) (int, error) {
	if caf.mw == nil {
		caf.dg = sha256.New()
		// resolve with a bogus digest just to get the right parent dir - in this manner its likely near the final location
		filename := caf.fn(caf.dg.Sum(nil))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			return 0, fmt.Errorf("error creating parent dir: %w", err)
		}
		var err error
		if caf.tf, err = os.CreateTemp(filepath.Dir(filename), "tmp"); err != nil {
			return 0, fmt.Errorf("error creating temp file: %w", err)
		}
		caf.mw = io.MultiWriter(caf.tf, caf.dg)
	}

	bw, err := caf.mw.Write(p)
	caf.size += bw
	return bw, err
}

func (caf *ContentAddressableFile) Commit() ([]byte, error) {
	// special-case empty file... other code won't write until first byte received,
	// here we force the issue which populates caf.tf etc
	if caf.size == 0 {
		if _, err := caf.writeAlways(nil); err != nil {
			return nil, fmt.Errorf("error writing empty file: %w", err)
		}
	}

	if err := caf.tf.Close(); err != nil {
		return nil, fmt.Errorf("err closing temp file: %w", err)
	}
	rv := caf.dg.Sum(nil)
	path := caf.fn(rv)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("error creating parent dir: %w", err)
	}
	if err := os.Rename(caf.tf.Name(), path); err != nil {
		return nil, fmt.Errorf("err renaming temp file: %w", err)
	}
	caf.mw, caf.tf, caf.dg = nil, nil, nil
	return rv, nil
}

// Tidy up any temporary files. Safe to call after Commit() in which case written data is left there,
// but if not, temp files are deleted
func (caf *ContentAddressableFile) Cleanup() (retErr error) {
	if caf.tf != nil {
		if err := caf.tf.Close(); err != nil && retErr == nil {
			retErr = err
		}
		if err := os.Remove(caf.tf.Name()); err != nil && retErr == nil {
			retErr = err
		}
		caf.tf = nil
	}
	caf.mw, caf.dg = nil, nil
	return nil
}
