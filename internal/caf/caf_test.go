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
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaf(t *testing.T) {
	td := t.TempDir()

	// test empty file
	caf1 := NewContentAddressableFile(func(digest []byte) string {
		return filepath.Join(td, hex.EncodeToString(digest))
	})
	assert.Nil(t, caf1.Commit())
	assert.Nil(t, caf1.Cleanup())

	// test simple file
	caf2 := NewContentAddressableFile(func(digest []byte) string {
		return filepath.Join(td, hex.EncodeToString(digest))
	})
	_, err := caf2.Write([]byte{1, 2})
	assert.Nil(t, err)
	assert.Nil(t, caf2.Commit())
	assert.Nil(t, caf2.Cleanup())

	// test simple file with Cleanup early
	caf3 := NewContentAddressableFile(func(digest []byte) string {
		return filepath.Join(td, hex.EncodeToString(digest))
	})
	_, err = caf3.Write([]byte{3, 5})
	assert.Nil(t, err)
	assert.Nil(t, caf3.Cleanup())

	// test simple file
	caf4 := NewContentAddressableFile(func(digest []byte) string {
		return filepath.Join(td, hex.EncodeToString(digest))
	})
	_, err = caf4.Write([]byte{1, 2})
	assert.Nil(t, err)
	assert.Nil(t, caf4.Commit())

	// count final files
	entries, err := os.ReadDir(td)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(entries)) // shoudl be for caf1 and caf2,caf4 only
}
