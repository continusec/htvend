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

	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = &CleanCommand{}

type CleanCommand struct {
	ManifestOptions

	RmGlobalCache bool `long:"all" description:"If set, instead remove shared global cache."`
}

func (rc *CleanCommand) Execute(args []string) (retErr error) {
	if rc.RmGlobalCache {
		// first open global manifest file
		mf, err := rc.ManifestOptions.MakeGlobalCacheManifestFile(nil)
		if err != nil {
			return fmt.Errorf("error opening global manifest file: %w", err)
		}

		// then destroy it
		if err := mf.CloseAndDestroy(); err != nil && retErr == nil {
			return err
		}

		// now open the actual blob store
		bs, err := rc.ManifestOptions.MakeBlobStore(true)
		if err != nil {
			return fmt.Errorf("error opening blob store: %w", err)
		}

		// and destroy that too
		if err := bs.Destroy(); err != nil {
			return fmt.Errorf("error removing global manifest: %w", err)
		}
	}

	return nil
}
