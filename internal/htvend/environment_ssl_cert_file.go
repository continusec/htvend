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
	resultPath := filepath.Join(e.TempDir, "cacerts.pem")
	if err := os.WriteFile(resultPath, e.CAPem, 0o444); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	for idx, ev := range e.Options.CertFileEnvVars {
		e.EnvOverrides = append(e.EnvOverrides, ev+"="+resultPath)
		if idx == 0 {
			e.BuildahArgs = append(e.BuildahArgs,
				"--secret=id=SSL_CERT_FILE_DATA,type=file,src="+resultPath,
				"--secret=id=SSL_CERT_FILE_PATH,type=env,env="+ev,
				"--run-mount=type=secret,id=SSL_CERT_FILE_DATA,required,target="+resultPath, // put in same spot so we can re-use env-var above
			)
		}
		e.BuildahArgs = append(e.BuildahArgs,
			"--run-mount=type=secret,id=SSL_CERT_FILE_PATH,required,env="+ev,
		)
	}
	return nil
}
