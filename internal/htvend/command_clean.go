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
	"fmt"
	"net/url"

	"github.com/continusec/htvend/internal/lockfile"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = &CleanCommand{}

type CleanCommand struct {
	CacheOptions

	RmGlobalCache bool `long:"all" description:"If set, remove entire shared global cache."`
}

func (rc *CleanCommand) Execute(args []string) (retErr error) {
	// first open global manifest file
	mf, err := rc.CacheOptions.MakeGlobalCacheManifestFile(nil, 0)
	if err != nil {
		return fmt.Errorf("error opening global manifest file: %w", err)
	}
	defer func() {
		if mf != nil {
			if err := mf.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}
	}()

	// now open the actual blob store
	bs, err := rc.CacheOptions.MakeBlobStore(true)
	if err != nil {
		return fmt.Errorf("error opening blob store: %w", err)
	}

	// if we are blowing everything away, then simply do so
	if rc.RmGlobalCache {
		// then destroy it
		if err := mf.CloseAndDestroy(); err != nil {
			return err
		}
		mf = nil // so that we don't attempt Close it again

		// and destroy that too
		return bs.Destroy()
	}

	// else we find a list of dangling blobs and delete those only
	blobsToKeep := make(map[string]bool)
	if err := mf.ForEach(func(k *url.URL, v lockfile.BlobInfo) error {
		blobsToKeep[v.Sha256] = true
		return nil
	}); err != nil {
		return err
	}

	return bs.RemoveExcept(blobsToKeep)
}
