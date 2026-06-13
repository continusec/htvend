// Copyright 2026 Continusec Pty Ltd.
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

package jobs

import "sync"

type MultiTasker struct {
	wg     sync.WaitGroup
	errors chan error
}

func NewMultiTasker() *MultiTasker {
	return &MultiTasker{
		errors: make(chan error),
	}
}

func (mt *MultiTasker) Queue(f func() error) {
	mt.wg.Add(1)
	go func() {
		defer mt.wg.Done()
		mt.errors <- f()
	}()
}

func (mt *MultiTasker) Wait(errCallback func(error)) error {
	go func() {
		defer close(mt.errors)
		mt.wg.Wait()
	}()
	var rvErr error
	for err := range mt.errors {
		if err != nil && rvErr == nil {
			rvErr = err
			errCallback(err)
		}
	}
	return rvErr
}
