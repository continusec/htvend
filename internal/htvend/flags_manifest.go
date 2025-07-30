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
	"strings"

	"github.com/adrg/xdg"
	"github.com/continusec/htvend/internal/blobs"
	"github.com/continusec/htvend/internal/lockfile"
	"github.com/continusec/htvend/internal/re"
)

type ManifestOptions struct {
	ManifestFile string `short:"m" long:"manifest" default:"./blobs.yml" description:"File to put manifest data in"`
	BlobsDir     string `long:"blobs-dir" default:"${XDG_DATA_HOME}/htvend/blobs" description:"Common directory to store downloaded blobs in"`
	CacheMap     string `long:"cache-manifest" default:"${XDG_DATA_HOME}/htvend/cache.yml" description:"Cache of all downloaded assets"`
}

type manifestContextOptions struct {
	Writable       bool
	FetchAlways    bool
	AllowOverwrite bool

	NoCacheList []string
}

func (o *ManifestOptions) MakeGlobalCacheManifestFile(noCache *re.MultiRegexMatcher) (*lockfile.File, error) {
	cmPath, err := xdgIt(o.CacheMap)
	if err != nil {
		return nil, fmt.Errorf("error getting cache map path with xdg: %w", err)
	}
	return lockfile.NewMapFile(lockfile.MapFileOptions{
		Path:           cmPath,
		AllowOverwrite: true, // global cache should overwrite new vals
		Writable:       true,
		NoCache:        noCache,
	})
}

func (o *ManifestOptions) MakeManifestFile(opts *manifestContextOptions) (*lockfile.File, error) {
	noCache, err := re.NewMultiRegexMatcher(opts.NoCacheList)
	if err != nil {
		return nil, fmt.Errorf("error creating no-cache regex matcher: %w", err)
	}

	var cache *lockfile.File
	if opts.Writable {
		cache, err = o.MakeGlobalCacheManifestFile(noCache)
		if err != nil {
			return nil, fmt.Errorf("error creating global manifest file: %w", err)
		}
	}

	return lockfile.NewMapFile(lockfile.MapFileOptions{
		Path:           o.ManifestFile,
		Writable:       opts.Writable,
		AllowOverwrite: opts.AllowOverwrite,
		AlwaysFetch:    opts.FetchAlways,

		Fallback: cache,
		NoCache:  noCache,
	})
}

func xdgIt(origPath string) (string, error) {
	d := origPath
	if remPath, ok := strings.CutPrefix(d, "${XDG_DATA_HOME}/"); ok {
		var err error
		if d, err = xdg.DataFile(remPath); err != nil {
			return "", fmt.Errorf("error getting path with xdg (%s): %w", origPath, err)
		}
	}
	return d, nil
}

func (o *ManifestOptions) MakeBlobStore(writable bool) (*blobs.DirectoryStore, error) {
	d, err := xdgIt(o.BlobsDir)
	if err != nil {
		return nil, fmt.Errorf("error getting blob store with xdg: %w", err)
	}
	return blobs.NewDirectoryStore(d, writable), nil
}
