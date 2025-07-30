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

var _ flags.Commander = &ExportCommand{}

type ExportCommand struct {
	ManifestOptions

	OutputDir string `short:"o" long:"output-directory" default:"./blobs" description:"Directory to export blobs to."`
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

	bs, err := rc.ManifestOptions.MakeBlobStore(false)
	if err != nil {
		return fmt.Errorf("error creating blob store: %w", err)
	}

	return doValidate(&validateCtx{
		Assets:         mf,
		Blobs:          bs,
		ExportDir:      rc.OutputDir,
		FailIfMissing:  true,
		ValidateSHA256: true,
		SaveToExport:   true,
	})
}
