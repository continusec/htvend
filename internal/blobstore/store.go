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

package blobstore

import (
	"errors"
	"io"
)

var ErrBlobNotExist = errors.New("blob does not exist")

type Store interface {
	// Get thing with this hash
	Get(k []byte) (io.ReadCloser, error)

	// Does this exist?
	Exists(k []byte) (bool, error)

	// Put a thing
	Put() (ContentAddressableBlob, error)

	// clean up everything - delete it all
	Destroy() error

	// delete everything except these (by string?)
	RemoveExcept(keep map[string]bool) error
}

type ContentAddressableBlob interface {
	io.Writer

	// Called when complete successfully. Returns hash and nil if successful.
	Commit() ([]byte, error)

	// Call if failed and should cleanup after ourselves. No-op if called after successful Commit()
	Cleanup() error
}
