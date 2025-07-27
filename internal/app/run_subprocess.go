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

func RunSubprocess(ctx context.Context, prompt string, args []string, extraEnv []string) error {
	if len(args) == 0 { // then assume shell
		shell := os.Getenv("SHELL")
		if shell == "" {
			return fmt.Errorf("no args specified, and unable to find SHELL envariable to default to")
		}
		args = []string{shell}
		if strings.HasSuffix(shell, "bash") {
			// if we're bash, then add an extra prompt
			args = append(args, "--norc")
			extraEnv = append(extraEnv, fmt.Sprintf("PS1=(%s) \\$ ", prompt))
		}

		logrus.Infof("Entering shell with env set to use proxy. Type exit / ctrl-D to exit.")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	cmd.Env = append(cmd.Environ(), extraEnv...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for _, e := range extraEnv {
		logrus.Debugf("export %s", e)
	}
	logrus.Debugf("%s", strings.Join(args, " "))
	defer logrus.Debugf("(terminated)")

	return cmd.Run()
}
