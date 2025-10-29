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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/continusec/htvend/internal/app"
	"github.com/continusec/htvend/internal/blobstore"
	"github.com/continusec/htvend/internal/lockfile"
	"github.com/continusec/htvend/internal/proxyserver"
	"github.com/continusec/htvend/internal/re"
	"github.com/sirupsen/logrus"
)

type ListenerOptions struct {
	app.SubprocessOptions `positional-args:"yes"`

	ListenAddr      string `short:"l" long:"listen-addr" default:"127.0.0.1:0" description:"Listen address for proxy server (:0) will allocate a dynamic open port"`
	TlsListenAddr   string `long:"tls-listen-addr" default:"127.0.0.1:0" description:"Listen address for a TLS proxy server (:0) will allocate a dynamic open port"`
	Daemon          bool   `short:"d" long:"daemon" description:"Run as a daemon until terminated"`
	DaemonRpcSocket string `long:"daemon-rpc-socket" description:"If set, create a unix socket here that can be used to accept updates such as new blobs."`

	TlsCertPem           string `long:"tls-cert-pem" description:"If set use this as the TLS cert. Must be a CA pem"`
	TlsKeyPem            string `long:"tls-key-pem" description:"If set use this as the TLS key. Must match the cert"`
	TlsGenerateIfMissing bool   `long:"tls-generate-if-missing" description:"If set, generate and save if files missing"`

	TmpDirs          []string `long:"with-temp-dir" short:"t" description:"List of temporary directories to be creating when running this command. Env vars will be be pointing to these for the sub-process."`
	CertFileEnvVars  []string `long:"set-env-var-ssl-cert-file" default:"SSL_CERT_FILE" description:"List of environment variables that will be set pointing to the temporary CA certificates file in PEM format."`
	JksKeyStoreVars  []string `long:"set-env-var-jks-keystore" default:"JKS_KEYSTORE_FILE" description:"List of environment variables that will be set pointing to the temporary CA certificates file in JKS format."`
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
	Blobs  blobstore.Store

	FetchIfMissing bool
	FailIfMissing  bool

	// "offline" options
	DummyOK *re.MultiRegexMatcher

	// "build" options
	HeadersToCache map[string]bool
}

type KeyValue struct {
	Key   *url.URL
	Value lockfile.BlobInfo
}

func (o *ListenerOptions) RunListenerWithSubprocess(lctx *listenerCtx, prompt string, args []string) error {
	return app.RunUntilSignals(func(parCtx context.Context) error {
		return proxyserver.ServeUntilDone(parCtx, proxyserver.ProxyServerConfig{
			HttpListenAddr:       o.ListenAddr,
			HttpsListenAddr:      o.TlsListenAddr,
			TlsCertPath:          o.TlsCertPem,
			TlsKeyPath:           o.TlsKeyPem,
			TlsGenerateIfMissing: o.TlsGenerateIfMissing,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if err := handleMainServerRequest(lctx, w, r); err != nil {
					logrus.Warnf("error handling request: %v", err)
					http.Error(w, "see proxy server log for details", http.StatusInternalServerError)
				}
			}}, func(ctx context.Context, proxyAddr string, caPem []byte) error {
			return app.WithTempDir(func(tempDir string) (retErr error) {
				ectx := envCtx{
					TempDir:   tempDir,
					ProxyAddr: proxyAddr,
					CAPem:     caPem,
					Options:   o,
				}
				for _, f := range []mutateEnvFunc{
					stdProxyVarsAppender,
					sslCertFileAppender,
					tmpDirsAppender,
					jksKeystoreAppender,
				} {
					if err := f(&ectx); err != nil {
						return fmt.Errorf("error modifying env: %w", err)
					}
				}

				if !o.Daemon {
					return app.RunSubprocess(ctx, prompt, o.SubprocessOptions, ectx.EnvOverrides)
				}

				// we are a daemon
				if o.SubprocessOptions.Process != "" {
					return fmt.Errorf("if running as a daemon, no sub-process should be specified. Received: %s", o.SubprocessOptions.Process)
				}

				if o.DaemonRpcSocket != "" {
					rpcSocket, err := net.Listen("unix", o.DaemonRpcSocket)
					if err != nil {
						return fmt.Errorf("error opening unix socket for daemon rpc: %w", err)
					}
					defer func() {
						if err := os.Remove(o.DaemonRpcSocket); err != nil && retErr == nil {
							retErr = err
						}
					}()

					rpcHandler := http.NewServeMux()
					rpcHandler.HandleFunc("/exists", func(w http.ResponseWriter, r *http.Request) {
						if r.Method != http.MethodGet {
							http.Error(w, "bad method", http.StatusBadRequest)
							return
						}
						kb, err := hex.DecodeString(r.URL.Query().Get("key"))
						if err != nil {
							http.Error(w, "bad key", http.StatusBadRequest)
							return
						}
						exists, err := lctx.Blobs.Exists(kb)
						if err != nil {
							logrus.Errorf("unexpected error checking blob existence: %v", err)
							http.Error(w, "internal error, see server log", http.StatusInternalServerError)
							return
						}
						if exists {
							w.WriteHeader(http.StatusOK)
						} else {
							w.WriteHeader(http.StatusNotFound)
						}
					})
					rpcHandler.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
						if r.Method != http.MethodPut {
							http.Error(w, "bad method", http.StatusBadRequest)
							return
						}
						caf, err := lctx.Blobs.Put()
						if err != nil {
							logrus.Errorf("unexpected error creating caf: %v", err)
							http.Error(w, "internal error, see server log", http.StatusInternalServerError)
							return
						}
						_, err = io.Copy(caf, r.Body)
						if err != nil {
							logrus.Errorf("unexpected error uploading caf: %v", err)
							http.Error(w, "internal error, see server log", http.StatusInternalServerError)
							return
						}
						digest, err := caf.Commit()
						if err != nil {
							logrus.Errorf("unexpected error committing caf: %v", err)
							http.Error(w, "internal error, see server log", http.StatusInternalServerError)
							return
						}
						w.Header().Set("X-Sha256-Digest", hex.EncodeToString(digest))
						w.WriteHeader(http.StatusCreated)
					})
					rpcHandler.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
						if r.Method != http.MethodPost {
							http.Error(w, "bad method", http.StatusBadRequest)
							return
						}
						var kv KeyValue
						if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
							http.Error(w, "bad request", http.StatusBadRequest)
							return
						}
						if err := lctx.Assets.AddBlob(kv.Key, kv.Value); err != nil {
							logrus.Errorf("unexpected error adding blob: %v", err)
							http.Error(w, "internal error, see server log", http.StatusInternalServerError)
							return
						}
						w.WriteHeader(http.StatusAccepted)
					})

					rpcServer := &http.Server{
						Handler: rpcHandler,
					}
					defer func() {
						if err := rpcServer.Shutdown(context.Background()); err != nil && retErr == nil {
							retErr = err
						}
					}()

					go rpcServer.Serve(rpcSocket)
				}

				logrus.Infof("Daemon running...")
				for _, ev := range ectx.EnvOverrides {
					fmt.Printf("export %s\n", ev)
				}

				<-ctx.Done()
				return nil
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
	blobs blobstore.Store,
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
	logrus.Infof("Fetching URL: %s %s", method, newReq.URL)
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

	logrus.Debugf("Response (%d):", resp.StatusCode)
	for k, v := range resp.Header {
		for _, v1 := range v {
			logrus.Debugf("%s: %s", k, v1)
		}
	}

	if w != nil {
		for k, v := range resp.Header {
			for _, v1 := range v {
				w.Header().Add(k, v1)
			}
		}
		w.WriteHeader(resp.StatusCode)
	}

	// if we don't need to save, then exit early - don't save non-OK responses - for now don't filter HEAD - useful for Docker API call during k3s init
	if assets.SkipSave(u) || resp.StatusCode != http.StatusOK {
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
		// Cleanup() is safe to call (no-op) after a successful Commit()
		if err := caf.Cleanup(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if w != nil {
		if _, err = io.Copy(w, io.TeeReader(resp.Body, caf)); err != nil {
			return fmt.Errorf("error copying response via tee: %w", err)
		}
	} else {
		if _, err = io.Copy(caf, resp.Body); err != nil {
			return fmt.Errorf("error copying response direct to CAF: %w", err)
		}
	}

	digest, err := caf.Commit()
	if err != nil {
		return fmt.Errorf("error committing blob (url %s): %w", u.Redacted(), err)
	}

	// record asset belonging to this build
	err = assets.AddBlob(u, lockfile.BlobInfo{
		Sha256:  hex.EncodeToString(digest),
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
