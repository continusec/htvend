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
	"github.com/continusec/htvend/internal/blobstore"
	"github.com/continusec/htvend/internal/blobstore/directory"
	"github.com/continusec/htvend/internal/blobstore/registry"
	"github.com/continusec/htvend/internal/blobstore/s3store"
	"github.com/continusec/htvend/internal/lockfile"
	"github.com/continusec/htvend/internal/re"
)

type CacheOptions struct {
	BlobsBackend  string `long:"blobs-backend" default:"filesystem" choice:"filesystem" choice:"registry" choice:"s3" description:"Type of blob store"`
	BlobsRegistry string `long:"blobs-registry" description:"URL for registry to store / fetch blobs from"`
	BlobsDir      string `long:"blobs-dir" default:"${XDG_DATA_HOME}/htvend/cache/blobs" description:"Common directory to store downloaded blobs in"`

	// S3 options - all other auth etc is with standard AWS env vars / metadata server
	BlobsBucket string `long:"blobs-bucket" description:"S3 bucket to use for blobs"`
	BlobsPrefix string `long:"blobs-prefix" default:"" description:"Prefix to prepend keys before uploading to S3 bucket"`
}

type ManifestOptions struct {
	CacheOptions
	ManifestFile string `short:"m" long:"manifest" default:"./assets.json" description:"File to put manifest data in"`
}

type manifestContextOptions struct {
	Writable        bool
	FetchAlways     bool
	AllowOverwrite  bool
	IncrementalSave bool

	NoCacheList []string
}

func (o *ManifestOptions) MakeManifestFile(opts *manifestContextOptions) (*lockfile.File, error) {
	noCache, err := re.NewMultiRegexMatcher(opts.NoCacheList)
	if err != nil {
		return nil, fmt.Errorf("error creating no-cache regex matcher: %w", err)
	}

	return lockfile.NewMapFile(lockfile.MapFileOptions{
		Path:            o.ManifestFile,
		Writable:        opts.Writable,
		AllowOverwrite:  opts.AllowOverwrite,
		AlwaysFetch:     opts.FetchAlways,
		IncrementalSave: opts.IncrementalSave,

		NoCache: noCache,
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

func (o *CacheOptions) MakeBlobStore(writable bool) (blobstore.Store, error) {
	switch o.BlobsBackend {
	case "filesystem":
		d, err := xdgIt(o.BlobsDir)
		if err != nil {
			return nil, fmt.Errorf("error getting blob store with xdg: %w", err)
		}
		return directory.NewDirectoryStore(d, writable), nil
	case "registry":
		return registry.NewRegistryStore(o.BlobsRegistry, writable), nil
	case "s3":
		return s3store.NewS3Store(s3store.S3StoreConfig{
			Bucket: o.BlobsBucket,
			Prefix: o.BlobsPrefix,
		})
	default:
		return nil, fmt.Errorf("bad blob store type: %s", o.BlobsBackend)
	}
}
