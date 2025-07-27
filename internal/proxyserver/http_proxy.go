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

package proxyserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type httpServer struct {
	listenAddr string
	caPrivKey  *rsa.PrivateKey
	ca         *x509.Certificate
	caPEM      []byte
	tlsAddr    string
}

func newSelfSignedServer(listenAddr string) (*httpServer, error) {
	var rv httpServer
	rv.listenAddr = listenAddr

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "httpvendor"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	var err error
	rv.caPrivKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("err generating priv key: %w", err)
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &rv.caPrivKey.PublicKey, rv.caPrivKey)
	if err != nil {
		return nil, fmt.Errorf("err signing cert: %w", err)
	}

	rv.ca, err = x509.ParseCertificate(caBytes)
	if err != nil {
		return nil, fmt.Errorf("err parsing cert: %w", err)
	}

	caPEM := &bytes.Buffer{}
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("err writing PEM: %w", err)
	}

	rv.caPEM = caPEM.Bytes()

	return &rv, nil
}

func ServeUntilDone(parCtx context.Context, listAddr string, handlerF http.HandlerFunc, childProcess func(ctx context.Context, proxyAddr string, caPemBytes []byte) error) (retErr error) {
	s, err := newSelfSignedServer(listAddr)
	if err != nil {
		return fmt.Errorf("error creating keys for servier: %w", err)
	}

	ctx, cancel := context.WithCancel(parCtx)
	cancelled := false
	defer func() {
		// if we get to the end successfully we deliberately call cancel to close all the listeners etc
		// The reason we check here is in case we didn't make it, but we don't want to call twice (simply because not sure what happens if we do)
		if !cancelled {
			cancel()
		}
	}()

	// this is the main listener which accepts proxy request, e.g. can
	// handle raw HTTP, as well as CONNECT requests
	list, err := net.Listen("tcp4", s.listenAddr)
	if err != nil {
		return fmt.Errorf("error making listener: %w", err)
	}
	defer list.Close()

	mainServerErr := make(chan error, 1)
	defer func() {
		if err := <-mainServerErr; err != nil && retErr == nil && !errors.Is(err, http.ErrServerClosed) {
			retErr = fmt.Errorf("error from main server: %w", err)
		}
	}()
	go func() {
		defer close(mainServerErr)
		server := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodConnect {
					s.handleConnect(w, r)
				} else {
					handlerF(w, r)
				}
			}),
		}
		go func() {
			<-ctx.Done()
			server.Shutdown(context.Background())
		}()
		mainServerErr <- server.Serve(list)
	}()

	// creates a second listener. This recieves HTTPS request sent by ourselves, to ourself
	// when handling CONNECT requests. We could probably do this better, but for now, this works
	list2, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("error making internal listener: %w", err)
	}
	defer list2.Close()

	s.tlsAddr = list2.Addr().String()

	tlsServerErr := make(chan error, 1)
	defer func() {
		if err := <-tlsServerErr; retErr == nil && err != nil && !errors.Is(err, http.ErrServerClosed) {
			retErr = fmt.Errorf("error from tls server: %w", err)
		}
	}()
	go func() {
		defer close(tlsServerErr)
		server := &http.Server{
			Handler: http.HandlerFunc(handlerF),
		}
		go func() {
			<-ctx.Done()
			server.Shutdown(context.Background())
		}()
		tlsServerErr <- server.Serve(tls.NewListener(list2, &tls.Config{
			GetCertificate: s.makeCertFor,
		}))
	}()

	defer func() {
		defer logrus.Debugf("(terminated)")
		cancel()
		cancelled = true
	}()
	return childProcess(ctx, list.Addr().String(), s.caPEM)
}

func (s *httpServer) handleConnect(w http.ResponseWriter, _ *http.Request) {
	if err := func() (retErr error) {
		// connect to our other server which handles MITM
		destConn, err := net.DialTimeout("tcp", s.tlsAddr, 10*time.Second)
		if err != nil {
			return fmt.Errorf("error dialing upstream: %w", err)
		}
		defer func() {
			if err := destConn.Close(); err != nil && retErr == nil {
				retErr = fmt.Errorf("error closing dest conn: %w", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		h, ok := w.(http.Hijacker)
		if !ok {
			return fmt.Errorf("oops, webserver does not support hijacking")
		}

		srcConn, extraBufferedData, err := h.Hijack() // TODO buffered reader?
		if err != nil {
			return fmt.Errorf("error hijacking conn: %w", err)
		}
		defer func() {
			if err := srcConn.Close(); err != nil && retErr == nil {
				retErr = fmt.Errorf("error closing src conn: %w", err)
			}
		}()

		// send data to server
		sendToServerErr := make(chan error, 1)
		defer func() {
			// make sure we read this err as that forces us to wait for go func to finish
			if err := <-sendToServerErr; retErr == nil && err != nil {
				retErr = err
			}
		}()
		go func() {
			defer close(sendToServerErr)
			_, err := io.Copy(destConn, io.MultiReader(extraBufferedData, srcConn))
			if err != nil {
				err = fmt.Errorf("error sending to dest: %w %T", err, err)
			}
			sendToServerErr <- err
		}()

		// and read from resp
		_, err = io.Copy(srcConn, destConn)
		if err != nil {
			return fmt.Errorf("error getting response from server: %w", err)
		}
		return nil
	}(); err != nil {
		fmt.Printf("error handling request somewhere: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *httpServer) makeCertFor(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	leaf := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		DNSNames:              []string{chi.ServerName},
		Subject:               pkix.Name{CommonName: chi.ServerName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, leaf, s.ca, &s.caPrivKey.PublicKey, s.caPrivKey)
	if err != nil {
		return nil, fmt.Errorf("err signing cert: %w", err)
	}
	return &tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  s.caPrivKey,
		Leaf:        leaf,
	}, nil
}
