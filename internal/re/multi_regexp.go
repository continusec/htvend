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

package re

import (
	"fmt"
	"regexp"
)

type MultiRegexMatcher struct {
	regexes []*regexp.Regexp
}

func NewMultiRegexMatcher(regexes []string) (*MultiRegexMatcher, error) {
	rv := &MultiRegexMatcher{
		regexes: make([]*regexp.Regexp, 0, len(regexes)),
	}
	for _, r := range regexes {
		re, err := regexp.Compile(r)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", r, err)
		}
		rv.regexes = append(rv.regexes, re)
	}
	return rv, nil
}

func (m *MultiRegexMatcher) Match(s string) bool {
	if m == nil {
		return false
	}
	for _, re := range m.regexes {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
