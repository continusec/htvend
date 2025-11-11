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
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type FlagsCommon struct {
	DirToChangeTo    string `short:"C" long:"chdir" default:"." description:"Directory to change to before running."`
	Verbose          bool   `short:"v" long:"verbose" description:"Set for verbose output. Equivalent to setting LOG_LEVEL=debug"`
	GithubOutputPath string `long:"github-output-path" description:"Set to append simple metric format to this path. Suitable for GITHUB_OUTPUT in Actions. Will write on success and most failures."`
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

func (fc FlagsCommon) OnShutdown() (retErr error) {
	if fc.GithubOutputPath != "" {
		fp, err := os.OpenFile(fc.GithubOutputPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			return fmt.Errorf("error opening file to write simple metrics to: %w", err)
		}
		defer func() {
			if err := fp.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		metrics, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			return fmt.Errorf("error gathering metrics: %w", err)
		}
		for _, m := range metrics {
			if m.Name != nil && strings.HasPrefix(*m.Name, "htvend_") && m.Type != nil && *m.Type == *prometheus.CounterMetricTypePtr {
				for _, mm := range m.Metric {
					if len(mm.Label) == 0 && mm.Counter != nil && mm.Counter.Value != nil {
						if _, err := fmt.Fprintf(fp, "%s=%d\n", *m.Name, uint64(*mm.Counter.Value)); err != nil {
							return fmt.Errorf("error writing simple metric out: %w", err)
						}
					}
				}
			}
		}
	}
	return nil
}
