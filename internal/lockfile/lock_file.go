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
	Headers map[string]string
}

func blobEquals(a, b BlobInfo) bool {
	return a.Sha256 == b.Sha256 && maps.Equal(a.Headers, b.Headers)
}

type blobMap map[string]BlobInfo

type File struct {
	options MapFileOptions

	mu            sync.Mutex
	blobs         blobMap
	previousBlobs blobMap
	dirty         bool
	lock          fslock.Handle
	lockPath      string
}

type MapFileOptions struct {
	// Path to where manifest is to be saved
	Path string

	// Is this read-only or one we will write out?
	Writable bool

	// If we get a new value for an existing entry, should we allow overwriting it?
	AllowOverwrite bool

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

	if err := rv.load(); err != nil {
		return nil, fmt.Errorf("error loading initial file: %w", err)
	}

	return rv, nil
}

func (f *File) SkipSave(u *url.URL) bool {
	return f.options.NoCache.Match(u.Redacted())
}

func (f *File) GetBlob(u *url.URL) (BlobInfo, bool, error) {
	k := u.Redacted()

	// if we don't want to cache it, stop early
	if f.options.NoCache.Match(k) {
		return BlobInfo{}, false, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// now, if we have it, then use it
	rv, ok := f.blobs[k]
	if ok {
		logrus.Infof("Found (manifest): %s", k)
		return rv, true, nil
	}

	// ok, do we have it prior to us being reset()
	rv, ok = f.previousBlobs[k]
	if ok {
		if err := f.internalAddBlob(k, rv); err != nil {
			logrus.Infof("Found (previous run): %s", k)
			return BlobInfo{}, false, fmt.Errorf("error storing previously cached blob info: %w", err)
		}
		return rv, true, nil
	}

	logrus.Infof("Not cached: %s", k)
	return BlobInfo{}, false, nil
}

func (f *File) AddBlob(u *url.URL, info BlobInfo) error {
	k := u.Redacted()

	if f.options.NoCache.Match(k) {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	return f.internalAddBlob(k, info)
}

func (f *File) internalAddBlob(k string, info BlobInfo) error {
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

// remove from us only with no regard to fallback
func (f *File) RemoveEntry(u *url.URL) error {
	k := u.Redacted()

	if f.options.NoCache.Match(k) {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	_, ok := f.blobs[k]
	if !ok {
		return nil
	}

	delete(f.blobs, k)
	f.dirty = true
	return f.save(false)
}

func (f *File) Reset(forgetBlobs bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if forgetBlobs {
		f.previousBlobs = nil
	} else {
		f.previousBlobs = f.blobs
	}

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
	logrus.Infof("rm %s", f.options.Path)
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
	if !final {
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
	logrus.Infof("loading assets file from: %s", f.options.Path)
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
