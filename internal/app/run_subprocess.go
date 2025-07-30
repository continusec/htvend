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
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type SubprocessOptions struct {
	Process string   `positional-arg-name:"COMMAND" description:"Sub-process to run. If not specified an interactive-shell is opened"`
	Args    []string `positional-arg-name:"ARG" description:"Arguments to pass to the sub-process"`
}

func RunSubprocess(ctx context.Context, prompt string, opts SubprocessOptions, extraEnv []string) error {
	if opts.Process == "" { // then assume shell
		opts.Process = os.Getenv("SHELL")
		if opts.Process == "" {
			return fmt.Errorf("no args specified, and unable to find SHELL envariable to default to")
		}
		opts.Args = nil
		if strings.HasSuffix(opts.Process, "bash") {
			// if we're bash, then add an extra prompt
			opts.Args = append(opts.Args, "--norc")
			extraEnv = append(extraEnv, fmt.Sprintf("PS1=(%s) \\$ ", prompt))
		}

		logrus.Infof("Entering shell with env set. Type exit / ctrl-D to exit.")
	}

	cmd := exec.CommandContext(ctx, opts.Process, opts.Args...)
	cmd.Env = append(cmd.Environ(), extraEnv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for _, e := range extraEnv {
		logrus.Debugf("export %s", e)
	}
	logrus.Debugf("%s %s", opts.Process, strings.Join(opts.Args, " "))
	defer logrus.Debugf("(terminated)")

	return cmd.Run()
}
