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
	state  int
	fn     FilenameResolver
	mw     io.Writer // writing until closed
	dg     hash.Hash // writing until closed
	tf     *os.File  // writing until committed
	path   string    // closed until cafStateCanceled
	digest []byte    // committed until cafStateCanceled
}

type FilenameResolver func(digest []byte) string

// NewContentAddressableFile creates a writeable file in a directory
// with a temp name that is renamed to the canonical hash only when it
// has been fully written out successfully. Temp file not created until
// first write.
func NewContentAddressableFile(fn FilenameResolver) *ContentAddressableFile {
	return &ContentAddressableFile{
		state: cafStateInitial,
		fn:    fn,
	}
}

func (caf *ContentAddressableFile) doInitToWriting() error {
	caf.dg = sha256.New()
	// create temp file with resolve and empty digest - in this manner its likely near the final location
	filename := caf.fn(caf.dg.Sum(nil))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("error creating parent dir: %w", err)
	}
	var err error
	if caf.tf, err = os.CreateTemp(filepath.Dir(filename), "tmp"); err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	caf.mw = io.MultiWriter(caf.tf, caf.dg)
	caf.state = cafStateWriting
	return nil
}

func (caf *ContentAddressableFile) doWritingToClosed() error {
	err := caf.tf.Close()
	if err != nil {
		return fmt.Errorf("err closing temp file: %w", err)
	}
	caf.digest = caf.dg.Sum(nil)
	caf.mw, caf.dg = nil, nil
	caf.state = cafStateClosed
	return nil
}

func (caf *ContentAddressableFile) doClosedToCommitted() error {
	caf.path = caf.fn(caf.digest)
	if err := os.MkdirAll(filepath.Dir(caf.path), 0o755); err != nil {
		return fmt.Errorf("error creating parent dir: %w", err)
	}
	if err := os.Rename(caf.tf.Name(), caf.path); err != nil {
		return fmt.Errorf("err renaming temp file: %w", err)
	}
	caf.tf = nil
	caf.state = cafStateCommitted
	return nil
}

func (caf *ContentAddressableFile) Write(p []byte) (int, error) {
	if caf.state == cafStateInitial {
		err := caf.doInitToWriting()
		if err != nil {
			return 0, fmt.Errorf("err initing content addressable file: %w", err)
		}
	}
	if caf.state != cafStateWriting {
		return 0, fmt.Errorf("invalid state for content addressable file: %d", caf.state)
	}
	return caf.mw.Write(p)
}

func (caf *ContentAddressableFile) doClosedToCancelled() error {
	err := os.Remove(caf.tf.Name())
	caf.tf = nil
	caf.state = cafStateCanceled
	return err
}

// Close will close and rename the file with the digest that it has calculated
// If file has never been written to, it will Close() without error
func (caf *ContentAddressableFile) Close() error {
	switch caf.state {
	case cafStateInitial:
		caf.state = cafStateClosed
		return nil
	case cafStateWriting:
		err := caf.doWritingToClosed()
		if err != nil {
			return fmt.Errorf("err transitioning from writing to closed: %w", err)
		}
		return caf.doClosedToCommitted()
	default:
		return fmt.Errorf("invalid state for close: %d", caf.state)
	}
}

// If not yet committed, cleanup as best we can
func (caf *ContentAddressableFile) Cleanup() error {
	switch caf.state {
	case cafStateInitial, cafStateCommitted:
		return nil
	case cafStateWriting:
		closeErr := caf.doWritingToClosed()
		cancelErr := caf.doClosedToCancelled()
		if closeErr != nil {
			return closeErr
		}
		return cancelErr
	case cafStateClosed:
		return caf.doClosedToCancelled()

	default:
		return fmt.Errorf("invalid state for cleanup: %d", caf.state)
	}
}

// Committed if we successfully wrote data to a hash
func (caf *ContentAddressableFile) Committed() bool {
	return caf.state == cafStateCommitted
}

// Only valid if Committed()
func (caf *ContentAddressableFile) Digest() []byte {
	return caf.digest
}

// Only valid if Committed()
func (caf *ContentAddressableFile) Path() string {
	return caf.path
}
