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

package registryauthclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	www "github.com/gboddin/go-www-authenticate-parser"
)

var (
	dockerRegistryRegex = regexp.MustCompile("^(https?://.*/v2/)([^/]+/[^/]+)/(blobs|manifests)/.*$")
)

type Client struct {
	upstream http.RoundTripper

	mu     sync.Mutex
	tokens map[string]cachedToken
}

type cachedToken struct {
	Token string
	TTL   time.Time
}

type tokenResp struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

func NewClient(upstream http.RoundTripper) http.RoundTripper {
	return &Client{
		upstream: upstream,
		tokens:   make(map[string]cachedToken),
	}
}

func (c *Client) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method != http.MethodGet {
		return c.upstream.RoundTrip(r)
	}
	regexResult := dockerRegistryRegex.FindStringSubmatch(r.URL.String())
	if len(regexResult) != 4 {
		return c.upstream.RoundTrip(r)
	}

	// do we have a token?
	key := regexResult[1] + regexResult[2]
	c.mu.Lock()
	val, ok := c.tokens[key]
	if val.TTL.Before(time.Now()) {
		delete(c.tokens, key)
		ok = false
	}
	c.mu.Unlock()

	if ok {
		r.Header.Set("Authorization", "Bearer "+val.Token) // are we bad for modifying this?
		return c.upstream.RoundTrip(r)
	}

	// else we expect to get a failure and we will check the result
	resp, err := c.upstream.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	authenticateSettings := www.Parse(resp.Header.Get("Www-Authenticate"))
	if authenticateSettings.AuthType != "Bearer" {
		return resp, err
	}

	realm, ok := authenticateSettings.Params["realm"]
	if !ok {
		return resp, err
	}

	service, ok := authenticateSettings.Params["service"]
	if !ok {
		return resp, err
	}

	scope, ok := authenticateSettings.Params["scope"]
	if !ok {
		return resp, err
	}

	// OK, we will do our request, kill the old resp
	err = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing response that we're ignoring: %w", err)
	}

	ar, err := http.NewRequest(http.MethodGet, realm+"?"+url.Values{
		"scope":   []string{scope},
		"service": []string{service},
	}.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}

	resp, err = c.upstream.RoundTrip(ar)
	if err != nil {
		return nil, fmt.Errorf("error making upstream RT: %w", err)
	}
	defer resp.Body.Close()

	var tr tokenResp
	err = json.NewDecoder(resp.Body).Decode(&tr)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}
	if tr.Token == "" {
		return nil, fmt.Errorf("error got blank token")
	}

	c.mu.Lock()
	c.tokens[key] = cachedToken{
		Token: tr.Token,
		TTL:   time.Now().Add(time.Second * time.Duration(tr.ExpiresIn-10)),
	}
	c.mu.Unlock()

	fr, err := http.NewRequest(http.MethodGet, r.URL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error making final request: %w", err)
	}
	fr.Header.Set("Authorization", "Bearer "+tr.Token)

	return c.upstream.RoundTrip(fr)
}
