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
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

func RunWithFlags(opts any, globalInit, globalShutdown func() error) {
	// create parser - this should include execution of any flags.Executor(s)
	parser := flags.NewParser(opts, flags.Default)
	parser.CommandHandler = func(command flags.Commander, args []string) (retErr error) {
		if err := globalInit(); err != nil {
			return err
		}
		defer func() {
			if err := globalShutdown(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		if command == nil {
			return nil
		}
		return command.Execute(args)
	}
	if _, err := parser.ParseArgs(os.Args[1:]); err != nil {
		if flags.WroteHelp(err) {
			return
		}
		logrus.Fatalf("%v", err)
	}
}
