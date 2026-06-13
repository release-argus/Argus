// Copyright [2026] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package httpx provides a HTTP client.
package httpx

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

var (
	Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,

		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		ForceAttemptHTTP2:     true,
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	Client = &http.Client{
		Timeout:   15 * time.Second,
		Transport: Transport,
	}
	InsecureTransport = Transport.Clone()
	InsecureClient    = &http.Client{
		Timeout:   15 * time.Second,
		Transport: InsecureTransport,
	}
)

func init() {
	configureInsecureTransport(InsecureTransport)
}

func configureInsecureTransport(tr *http.Transport) {
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}

	//#nosec G402 -- explicitly wanted InsecureSkipVerify
	tr.TLSClientConfig.InsecureSkipVerify = true
}
