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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"software.sslmate.com/src/go-pkcs12"
)

func jksTrustStoreAppender(e *envCtx) error {
	var entries []pkcs12.TrustStoreEntry
	caPem := e.CAPem
	for i := 0; true; i++ {
		var block *pem.Block
		block, caPem = pem.Decode(caPem)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing cert: %w", err)
		}
		entries = append(entries, pkcs12.TrustStoreEntry{
			Cert:         cert,
			FriendlyName: fmt.Sprintf("cert%d", i),
		})
	}
	bb, err := pkcs12.Passwordless.EncodeTrustStoreEntries(entries, "")
	if err != nil {
		return fmt.Errorf("error encoding JKS: %w", err)
	}

	resultPath := filepath.Join(e.TempDir, "truststore.jks")
	if err := os.WriteFile(resultPath, bb, 0o444); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	e.EnvOverrides = append(e.EnvOverrides, "JAVA_TRUST_STORE_FILE="+resultPath) // not needed for BUILDAH, but may be useful elsewhere
	e.BuildahArgs = append(e.BuildahArgs, "--secret=id=JAVA_TRUST_STORE_FILE,type=file,src="+resultPath)
	for _, ev := range e.Options.JksPaths {
		e.BuildahArgs = append(e.BuildahArgs, "--run-mount=type=secret,id=JAVA_TRUST_STORE_FILE,target="+ev+",required")
	}
	return nil
}
