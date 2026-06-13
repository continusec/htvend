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

package main

import (
	"github.com/continusec/htvend/internal/app"
	"github.com/continusec/htvend/internal/htvend"
)

func main() {
	opts := &struct {
		app.FlagsCommon
		Build   htvend.BuildCommand   `command:"build" description:"Run command to create/update the manifest file"`
		Verify  htvend.VerifyCommand  `command:"verify" description:"Verify and fetch any missing assets in the manifest file"`
		Export  htvend.ExportCommand  `command:"export" description:"Export referenced assets to directory"`
		Offline htvend.OfflineCommand `command:"offline" description:"Serve assets to command, don't allow other outbound requests"`
	}{}
	// not 100% clear to me why we need to wrap opts.FlagsCommon.Apply, but I suspect it's because the value changes
	// and it's not a proper pointer? Anyway this works, and not doing so doesn't.
	app.RunWithFlags(opts, func() error { return opts.FlagsCommon.Apply() }, func() error { return opts.FlagsCommon.OnShutdown() })
}
