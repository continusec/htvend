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

package app

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type FlagsCommon struct {
	DirToChangeTo string `short:"C" long:"chdir" default:"." description:"Directory to change to before running."`
	Verbose       bool   `short:"v" long:"verbose" description:"Set for verbose output. Equivalent to setting LOG_LEVEL=debug"`
}

func (fc FlagsCommon) Apply() error {
	ll := logrus.InfoLevel

	// is an env var set?
	lls, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		var err error
		ll, err = logrus.ParseLevel(lls)
		if err != nil {
			return fmt.Errorf("bad log level: %w", err)
		}
	}
	// is verbose set? (Will override any env vars)
	if fc.Verbose {
		ll = logrus.DebugLevel
	}
	logrus.SetLevel(ll)

	if fc.DirToChangeTo != "." {
		if err := os.Chdir(fc.DirToChangeTo); err != nil {
			return err
		}
	}
	return nil
}
