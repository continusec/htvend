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

	"github.com/continusec/htvend/internal/re"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = &OfflineCommand{}

type OfflineCommand struct {
	ManifestOptions
	ListenerOptions

	DummyOK []string `long:"dummy-ok-response" default:"^http.*/v2/$" description:"Regex list of URLs that we return a dummy 200 OK reply to. Useful for some Docker clients."`
}

func (rc *OfflineCommand) Execute(args []string) (retErr error) {
	bs, err := rc.ManifestOptions.MakeBlobStore(false)
	if err != nil {
		return fmt.Errorf("error making directory blob store: %w", err)
	}

	mf, err := rc.ManifestOptions.MakeManifestFile(&manifestContextOptions{}) // read-only!
	if err != nil {
		return fmt.Errorf("error getting manifest file: %w", err)
	}
	defer func() {
		if err := mf.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	dummyOK, err := re.NewMultiRegexMatcher(rc.DummyOK)
	if err != nil {
		return fmt.Errorf("error creating dummy OK regex matcher: %w", err)
	}
	return rc.ListenerOptions.RunListenerWithSubprocess(&listenerCtx{
		Assets:        mf,
		Blobs:         bs,
		FailIfMissing: true,
		DummyOK:       dummyOK,
	}, "htvend offline", args)
}
