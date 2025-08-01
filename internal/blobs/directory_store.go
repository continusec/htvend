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

package blobs

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/continusec/htvend/internal/caf"
	"github.com/sirupsen/logrus"
)

type DirectoryStore struct {
	dir      string
	writable bool
}

func NewDirectoryStore(dir string, writable bool) *DirectoryStore {
	return &DirectoryStore{
		dir:      dir,
		writable: writable,
	}
}

// key is raw hash
// caller must call Close()
func (s *DirectoryStore) Get(k []byte) (io.ReadCloser, error) {
	return os.Open(s.resolve(k))
}

func (s *DirectoryStore) resolve(k []byte) string {
	return filepath.Join(s.dir, hex.EncodeToString(k))
}

func (s *DirectoryStore) Put() (*caf.ContentAddressableFile, error) {
	if !s.writable {
		return nil, errors.New("blob store is not writable")
	}
	return caf.NewContentAddressableFile(s.resolve), nil
}

func (s *DirectoryStore) Destroy() error {
	if !s.writable {
		return errors.New("blob store is not writable and therefore cannot be destroyed")
	}
	logrus.Infof("rm -rf %s", s.dir)
	return os.RemoveAll(s.dir)
}

func (s *DirectoryStore) RemoveExcept(keep map[string]bool) error {
	if !s.writable {
		return errors.New("blob store is not writable and therefore cannot be modified")
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// no work required
			return nil
		}
		return fmt.Errorf("error listing blobs dir: %w", err)
	}
	for _, e := range entries {
		if !keep[e.Name()] {
			pathToRm := filepath.Join(s.dir, e.Name())
			logrus.Infof("rm -f %s", pathToRm)
			if err := os.Remove(pathToRm); err != nil {
				return err
			}
		}
	}
	return nil
}
