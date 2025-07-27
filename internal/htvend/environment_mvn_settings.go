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
	"encoding/xml"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func mvnSettingsAppender(e *envCtx) error {
	host, port, err := net.SplitHostPort(e.ProxyAddr)
	if err != nil {
		return fmt.Errorf("error splitting host port for proxy addr: %w", err)
	}

	type mvnProxy struct {
		ID            string `xml:"id"`
		Active        bool   `xml:"active"`
		Protocol      string `xml:"protocol"`
		Host          string `xml:"host"`
		Port          string `xml:"port"`
		NonProxyHosts string `xml:"nonProxyHosts"`
	}

	bb, err := xml.Marshal(struct {
		XMLName xml.Name   `xml:"settings"`
		Proxy   []mvnProxy `xml:"proxies>proxy"`
	}{
		Proxy: []mvnProxy{
			{
				ID:            "httpvendor",
				Active:        true,
				Protocol:      "http",
				Host:          host,
				Port:          port,
				NonProxyHosts: "",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error making XML: %w", err)
	}

	resultPath := filepath.Join(e.TempDir, "settings.xml")
	if err := os.WriteFile(resultPath, bb, 0o444); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	e.EnvOverrides = append(e.EnvOverrides, "MAVEN_SETTINGS_FILE="+resultPath) // not needed for BUILDAH, but may be useful elsewhere
	e.BuildahArgs = append(e.BuildahArgs, "--secret=id=MAVEN_SETTINGS_FILE,type=file,src="+resultPath)
	for _, ev := range e.Options.MvnSettingsPaths {
		e.BuildahArgs = append(e.BuildahArgs, "--run-mount=type=secret,id=MAVEN_SETTINGS_FILE,target="+ev+",required")
	}
	return nil
}
