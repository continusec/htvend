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

package lockfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"sync"

	"github.com/continusec/htvend/internal/re"
	"github.com/danjacques/gofslock/fslock"
	"github.com/sirupsen/logrus"
)

type BlobInfo struct {
	Sha256  string
	Size    int
	Headers map[string]string
}

func blobEquals(a, b BlobInfo) bool {
	return a.Size == b.Size && a.Sha256 == b.Sha256 && maps.Equal(a.Headers, b.Headers)
}

type blobMap map[string]BlobInfo

type File struct {
	options MapFileOptions

	mu       sync.Mutex
	blobs    blobMap
	dirty    bool
	lock     fslock.Handle
	lockPath string
}

type MapFileOptions struct {
	// Path to where manifest is to be saved
	Path string

	// Is this read-only or one we will write out?
	Writable bool

	// If we get a new value for an existing entry, should we allow overwriting it?
	AllowOverwrite bool

	// Upstream cache to consult
	Fallback *File

	// If set, always never return blob (ie force a new fetch will be cached)
	AlwaysFetch bool

	// Should we save the file out as each entry is written (useful during debugging) else it is saved once at end
	IncrementalSave bool

	// List of regexes that we never return a value for
	NoCache *re.MultiRegexMatcher
}

// if writable, then we get an exclusive lock on this file,
// and caller must release it by calling Close()
func NewMapFile(options MapFileOptions) (retFile *File, retErr error) {
	rv := &File{
		options: options,
	}
	if options.Writable {
		var err error
		rv.lockPath = options.Path + ".lock"
		rv.lock, err = fslock.Lock(rv.lockPath)
		if err != nil {
			return nil, fmt.Errorf("error getting lock on map path (%s): %w", options.Path, err)
		}
		defer func() {
			// if we return an error, then we must release any lock we have
			if retErr != nil {
				retFile.Close()
			}
		}()
	}

	return rv, rv.load()
}

func (f *File) SkipSave(u *url.URL) bool {
	return f.options.NoCache.Match(u.Redacted())
}

func (f *File) GetBlob(u *url.URL) (BlobInfo, bool, error) {
	if f.options.AlwaysFetch {
		return BlobInfo{}, false, nil
	}

	k := u.Redacted()

	// if we don't want to cache it, stop early
	if f.options.NoCache.Match(k) {
		return BlobInfo{}, false, nil
	}

	// first, see if fallback has it
	var fallbackRV BlobInfo
	var fallbackOK bool
	if f.options.Fallback != nil {
		var err error
		if fallbackRV, fallbackOK, err = f.options.Fallback.GetBlob(u); err != nil {
			return BlobInfo{}, false, fmt.Errorf("error getting fallback blob: %w", err)
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// now, if we have it, then use it
	rv, ok := f.blobs[k]
	if ok {
		logrus.Debugf("cache hit for %s", k)
		return rv, ok, nil
	}

	// else if fallback has it populate it here so that we save this out as used
	if fallbackOK {
		logrus.Debugf("fallback cache hit for %s", k)
		f.blobs[k] = fallbackRV
		f.dirty = true
	} else {
		logrus.Debugf("cache miss for %s", k)
	}

	return fallbackRV, fallbackOK, f.save(false)
}

func (f *File) AddBlob(u *url.URL, info BlobInfo) error {
	// first put in backing cache
	if f.options.Fallback != nil {
		if err := f.options.Fallback.AddBlob(u, info); err != nil {
			return fmt.Errorf("error adding blob to fallback (cache): %w", err)
		}
	}

	// then do our thing
	k := u.Redacted()

	if f.options.NoCache.Match(k) {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	v0, ok := f.blobs[k]
	if ok {
		if blobEquals(v0, info) {
			return nil
		}
		if !f.options.AllowOverwrite {
			return fmt.Errorf("wrong SHA256 for %s: expected: %s received: %s (or different headers)", k, v0.Sha256, info.Sha256)
		}
	}
	f.blobs[k] = info
	f.dirty = true
	return f.save(false)
}

func (f *File) Reset() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.blobs) == 0 {
		return nil
	}
	f.blobs = make(blobMap)
	f.dirty = true
	return f.save(false)
}

func (f *File) CloseAndDestroy() error {
	if !f.options.Writable {
		return fmt.Errorf("hmm, file should not be destroyable")
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing manifest file prior to destruction: %w", err)
	}
	return os.Remove(f.options.Path)
}

func (f *File) ForEach(cb func(k *url.URL, v BlobInfo) error) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for k, v := range f.blobs {
		u, err := url.Parse(k)
		if err != nil {
			return err
		}
		if err := cb(u, v); err != nil {
			return err
		}
	}

	return nil
}

// writes file out, releases any locks we have
// ok to call if read-only
func (f *File) Close() (retErr error) {
	// make sure fallback is Closed()
	if f.options.Fallback != nil {
		defer func() {
			if err := f.options.Fallback.Close(); err != nil && retErr == nil {
				retErr = fmt.Errorf("error closing fallback cache map file: %w", err)
			}
		}()
	}

	if !f.options.Writable {
		return nil
	}

	defer func() {
		if err := f.lock.Unlock(); err != nil && retErr == nil {
			retErr = err
		}
		if err := os.Remove(f.lockPath); err != nil && retErr == nil {
			retErr = err
		}
	}()

	f.mu.Lock()
	defer f.mu.Unlock()

	return f.save(true)
}

// caller must get mutex
func (f *File) save(final bool) (retErr error) {
	// if no need to write, then don't
	if !f.dirty {
		return nil
	}

	// first handle reading read-only
	if !f.options.Writable {
		return fmt.Errorf("%s is not writable! should not get here", f.lockPath)
	}

	// next, handle non-final save
	if !final && !f.options.IncrementalSave {
		return nil
	}

	bb, err := json.MarshalIndent(f.blobs, "", "  ") // uses the JSON marshaller which docs say will sort keys
	if err != nil {
		return fmt.Errorf("error marshalling: %w", err)
	}

	if err = os.WriteFile(f.options.Path, bb, 0o666); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	f.dirty = false

	return nil
}

// caller must get mutex
func (f *File) load() (retErr error) {
	f.blobs = make(blobMap)
	bb, err := os.ReadFile(f.options.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && f.options.Writable {
			f.dirty = true
			return nil
		}
		return fmt.Errorf("error opening map: %w", err)
	}
	return json.Unmarshal(bb, &f.blobs)
}
