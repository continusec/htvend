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
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/continusec/htvend/internal/app"
	"github.com/continusec/htvend/internal/blobs"
	"github.com/continusec/htvend/internal/lockfile"
	"github.com/continusec/htvend/internal/proxyserver"
	"github.com/continusec/htvend/internal/re"
	"github.com/sirupsen/logrus"
)

type ListenerOptions struct {
	app.SubprocessOptions `positional-args:"yes"`

	ListenAddr string `short:"l" long:"listen-addr" default:"127.0.0.1:0" description:"Listen address for proxy server (:0) will allocate a dynamic open port"`

	CertFileEnvVars  []string `long:"set-env-var-ssl-cert-file" default:"SSL_CERT_FILE" description:"List of environment variables that will be set pointing to the temporary certificate file."`
	HttpProxyEnvVars []string `long:"set-env-var-http-proxy" default:"HTTP_PROXY" default:"HTTPS_PROXY" default:"http_proxy" default:"https_proxy" description:"List of environment variables that will be set pointing to the proxy host:port."`
	NoProxyEnvVars   []string `long:"set-env-var-no-proxy" default:"NO_PROXY" default:"no_proxy" description:"List of environment variables that will be set blank."`
}

type mutateEnvFunc func(ectx *envCtx) error

type envCtx struct {
	// input
	TempDir   string
	ProxyAddr string
	CAPem     []byte
	Options   *ListenerOptions

	// output
	BuildahArgs  []string
	EnvOverrides []string
}

type listenerCtx struct {
	Assets *lockfile.File
	Blobs  *blobs.DirectoryStore

	FetchIfMissing bool
	FailIfMissing  bool

	// "offline" options
	DummyOK *re.MultiRegexMatcher

	// "build" options
	HeadersToCache map[string]bool
}

func (o *ListenerOptions) RunListenerWithSubprocess(lctx *listenerCtx, prompt string, args []string) error {
	return app.RunUntilSignals(func(parCtx context.Context) error {
		return proxyserver.ServeUntilDone(parCtx, o.ListenAddr, func(w http.ResponseWriter, r *http.Request) {
			if err := handleMainServerRequest(lctx, w, r); err != nil {
				logrus.Warnf("error handling request: %v", err)
				http.Error(w, "see proxy server log for details", http.StatusInternalServerError)
			}
		}, func(ctx context.Context, proxyAddr string, caPem []byte) error {
			return app.WithTempDir(func(tempDir string) error {
				ectx := envCtx{
					TempDir:   tempDir,
					ProxyAddr: proxyAddr,
					CAPem:     caPem,
					Options:   o,
				}
				for _, f := range []mutateEnvFunc{
					stdProxyVarsAppender,
					sslCertFileAppender,
				} {
					if err := f(&ectx); err != nil {
						return fmt.Errorf("error modifying env: %w", err)
					}
				}
				return app.RunSubprocess(ctx, prompt, o.SubprocessOptions, ectx.EnvOverrides)
			})
		})
	})
}

func handleMainServerRequest(lctx *listenerCtx, w http.ResponseWriter, r *http.Request) error {
	u := extractReqFromURL(r)

	if lctx.DummyOK.Match(u.Redacted()) {
		// we return a dummy 200 OK response
		w.WriteHeader(http.StatusOK)
		return nil
	}

	bi, found, err := lctx.Assets.GetBlob(u)
	if err != nil {
		return fmt.Errorf("error looking up asset: %w", err)
	}

	if found {
		return serveFoundBlob(lctx, bi, w)
	}

	if lctx.FetchIfMissing {
		return fetchAndSaveBlob(lctx.Assets, lctx.Blobs, r.Method, r.Body, u, http.DefaultClient, lctx.HeadersToCache, func(newReq *http.Request) error {
			for k, v := range r.Header {
				for _, v1 := range v {
					newReq.Header.Add(k, v1)
				}
			}
			return nil
		}, w)
	}

	if lctx.FailIfMissing {
		logrus.Warnf("missing asset for URL: %s", u.Redacted())
		http.Error(w, "missing asset", http.StatusNotFound)
		return nil // as this is unremarkable for the proxy server
	}

	return errors.New("missing logic path - should not have gotten here")
}

// r and w are optional - if they are specified, then we are in a reverse proxy request
// ELSE we happily ignore them being nil and assume GET with no body or headers
// as this is called by validate
func fetchAndSaveBlob(
	assets *lockfile.File,
	blobs *blobs.DirectoryStore,
	method string,
	body io.Reader,
	u *url.URL,
	client *http.Client,
	hdrsToCache map[string]bool,
	preprocessRequest func(*http.Request) error,
	w http.ResponseWriter,
) (retErr error) {
	newReq, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return fmt.Errorf("error making request object: %w", err)
	}
	logrus.Debugf("req for URL: %s", newReq.URL)
	if preprocessRequest != nil {
		err = preprocessRequest(newReq)
		if err != nil {
			return fmt.Errorf("error preprocessing request (%s): %w", u, err)
		}
	}
	resp, err := client.Do(newReq)
	if err != nil {
		return fmt.Errorf("error performing request (%s): %w", u, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("error closing response: %w", err)
		}
	}()

	if w != nil {
		for k, v := range resp.Header {
			for _, v1 := range v {
				w.Header().Add(k, v1)
			}
		}
		w.WriteHeader(resp.StatusCode)
	}

	// if we don't need to save, then exit early
	if assets.SkipSave(u) {
		if w != nil {
			_, err = io.Copy(w, resp.Body)
			return err
		}
		return nil
	}

	// we save to this file
	caf, err := blobs.Put()
	if err != nil {
		return fmt.Errorf("error creating caf to put: %w", err)
	}
	defer func() {
		// if we aren't clean on return, then delete the CAF
		if retErr != nil {
			_ = caf.Cleanup() // we are already returning an error, so we swallow this one
		}
	}()

	if w != nil {
		_, err = io.Copy(w, io.TeeReader(resp.Body, caf))
		if err != nil {
			return fmt.Errorf("error copying response via tee: %w", err)
		}
	} else {
		_, err = io.Copy(caf, resp.Body)
		if err != nil {
			return fmt.Errorf("error copying response direct to CAF: %w", err)
		}
	}

	err = caf.Close()
	if err != nil {
		return fmt.Errorf("error committing blob: %w", err)
	}

	// record asset belonging to this build
	err = assets.AddBlob(u, lockfile.BlobInfo{
		Size:    caf.Size(),
		Sha256:  hex.EncodeToString(caf.Digest()),
		Headers: filterHeaders(hdrsToCache, resp.Header),
	})
	if err != nil {
		return fmt.Errorf("error updating asset file: %w", err)
	}
	return nil
}

func filterHeaders(desired map[string]bool, actual http.Header) map[string]string {
	rv := make(map[string]string)
	for k, vs := range actual {
		if desired[k] {
			rv[k] = vs[0]
		}
	}
	return rv
}

func serveFoundBlob(lctx *listenerCtx, bi lockfile.BlobInfo, w http.ResponseWriter) (retErr error) {
	k, err := hex.DecodeString(bi.Sha256)
	if err != nil {
		return fmt.Errorf("bad hex key: %w", err)
	}
	br, err := lctx.Blobs.Get(k)
	if err != nil {
		return fmt.Errorf("error opening blob: %w", err)
	}
	defer func() {
		if err := br.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	hdrs := w.Header()
	for k, v := range bi.Headers {
		hdrs.Set(k, v)
	}

	_, err = io.Copy(w, br)
	return err
}

func extractReqFromURL(r *http.Request) *url.URL {
	var protocol string
	if r.TLS == nil {
		protocol = "http"
	} else {
		protocol = "https"
	}
	return &url.URL{
		Scheme:   protocol, // r.URL has empty scheme and host, so we make our own
		Host:     r.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
		Fragment: r.URL.Fragment,
	}
}
