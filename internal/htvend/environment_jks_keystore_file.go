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
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"software.sslmate.com/src/go-pkcs12"
)

func jksKeystoreAppender(e *envCtx) error {
	if len(e.Options.JksKeyStoreVars) != 0 {
		jks := bytes.NewBuffer(nil)
		if err := pemToJks(bytes.NewReader(e.CAPem), jks); err != nil {
			return fmt.Errorf("error converting PEM to JKS file: %w", err)
		}

		resultPath := filepath.Join(e.TempDir, "cacerts.jks")
		if err := os.WriteFile(resultPath, jks.Bytes(), 0o444); err != nil {
			return fmt.Errorf("error writing CA PEM file: %w", err)
		}

		for _, ev := range e.Options.JksKeyStoreVars {
			e.EnvOverrides = append(e.EnvOverrides, ev+"="+resultPath)
		}
	}
	return nil
}

func pemToJks(in io.Reader, out io.Writer) error {
	caPem, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("error loading PEM from stdin: %w", err)
	}

	var entries []pkcs12.TrustStoreEntry
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

	_, err = out.Write(bb)
	return err
}
