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
	"net/textproto"

	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = &BuildCommand{}

type FetchOptions struct {
	NoCache     []string `long:"no-cache-response" default:"^http.*/v2/$" default:"/token\\?" description:"Regex list of URLs to never store in cache. Useful for token endpoints."`
	CacheHeader []string `long:"cache-header" default:"Content-Length" default:"Docker-Content-Digest" default:"Content-Type" default:"Content-Encoding" default:"X-Checksum-Sha1" description:"List of headers for which we will cache the first value."`
}

func (fo FetchOptions) CacheHeaderMap() map[string]bool {
	rv := make(map[string]bool)
	for _, h := range fo.CacheHeader {
		rv[textproto.CanonicalMIMEHeaderKey(h)] = true
	}
	return rv
}

type BuildCommand struct {
	ManifestOptions
	ListenerOptions
	FetchOptions

	ForceRefresh bool `long:"force-refresh" description:"If set, ignore any existing SHA256 values"`
}

func (rc *BuildCommand) Execute(args []string) (retErr error) {
	bs, err := rc.ManifestOptions.MakeBlobStore(true)
	if err != nil {
		return fmt.Errorf("error making directory blob store: %w", err)
	}

	mf, err := rc.ManifestOptions.MakeManifestFile(&manifestContextOptions{
		Writable:    true,
		NoCacheList: rc.NoCache,
	})
	if err != nil {
		return fmt.Errorf("error getting manifest file: %w", err)
	}
	defer func() {
		if err := mf.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if err := mf.Reset(rc.ForceRefresh); err != nil {
		return fmt.Errorf("error resetting manifest file: %w", err)
	}

	return rc.ListenerOptions.RunListenerWithSubprocess(&listenerCtx{
		Assets:         mf,
		Blobs:          bs,
		FetchIfMissing: true,
		HeadersToCache: rc.FetchOptions.CacheHeaderMap(),
	}, "htvend build", args)
}
