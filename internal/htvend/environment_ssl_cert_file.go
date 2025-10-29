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

package htvend

import (
	"fmt"
	"os"
	"path/filepath"
)

func sslCertFileAppender(e *envCtx) error {
	if len(e.Options.CertFileEnvVars) != 0 {
		resultPath := e.Options.TlsCertPem
		if resultPath == "" {
			resultPath = filepath.Join(e.TempDir, "cacerts.pem")
			if err := os.WriteFile(resultPath, e.CAPem, 0o444); err != nil {
				return fmt.Errorf("error writing CA PEM file: %w", err)
			}
		}
		for _, ev := range e.Options.CertFileEnvVars {
			e.EnvOverrides = append(e.EnvOverrides, ev+"="+resultPath)
		}
	}
	return nil
}
