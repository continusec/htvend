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
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/continusec/htvend/internal/app"
	"software.sslmate.com/src/go-pkcs12"
)

func createJksInHere(resultPath string) error {
	pemLoc, ok := os.LookupEnv("SSL_CERT_FILE")
	if !ok {
		return fmt.Errorf("SSL_CERT_FILE must be specified")
	}

	caPem, err := os.ReadFile(pemLoc)
	if err != nil {
		return fmt.Errorf("error loading %s: %w", pemLoc, err)
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

	if err := os.WriteFile(resultPath, bb, 0o444); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func main() {
	opts := &struct {
		app.FlagsCommon
		app.SubprocessOptions `positional-args:"yes"`

		JksEnvVarPath []string `short:"j" description:"This tool takes the certificate file at SSL_CERT_FILE and creates a pkcs12 file with the same data, and put it at a temp location referenced by this path."`
	}{}
	app.RunWithFlags(opts, func() error {
		if err := opts.FlagsCommon.Apply(); err != nil {
			return err
		}

		return app.WithTempDir(func(tempDirRoot string) error {
			jksPath := filepath.Join(tempDirRoot, "cacerts.jks")
			if err := createJksInHere(jksPath); err != nil {
				return err
			}

			var extraEnv []string
			for _, ev := range opts.JksEnvVarPath {
				extraEnv = append(extraEnv, ev+"="+jksPath)
			}
			return app.RunSubprocess(context.Background(), "with-java-keystore-main", opts.SubprocessOptions, extraEnv)
		})
	})
}
