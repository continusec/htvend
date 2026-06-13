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

package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func WithTempDir(f func(string) error) (retErr error) {
	td, err := os.MkdirTemp("", "htvend")
	if err != nil {
		return fmt.Errorf("error creating temp dir: %w", err)
	}
	logrus.Debugf("mkdir -p %s", td)
	defer func() {
		logrus.Debugf("rm -rf %s", td)
		if err := deleteDir(td); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return f(td)
}

func deleteDir(tempDir string) error {
	// first walk the tree and set perms to 0o700 for all dirs, we don't worry about error handling in here
	filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, _ error) error {
		if !d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Mode().Perm()&0o700 != 0o700 {
			_ = os.Chmod(path, 0o700)
		}
		return nil
	})

	// now try to delete, and return any errors
	return os.RemoveAll(tempDir)
}
