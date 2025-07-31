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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"software.sslmate.com/src/go-pkcs12"
)

func createJksInHere(in io.Reader, out io.Writer) error {
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

func main() {
	if err := createJksInHere(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error converting PEM to JKS. Expecting PEM on stdin\n")
		os.Exit(1)
	}
}
