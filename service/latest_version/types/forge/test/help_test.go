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

//go:build unit

package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var packageName = "forgetest"

// get makes a GET against `server` for `path`, carrying `header`.
func get(
	t *testing.T,
	server *Server,
	path string,
	header http.Header,
) (*http.Response, string) {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
	if err != nil {
		t.Fatalf(
			"%s\ncould not build a request for %q: %v",
			packageName, path, err,
		)
	}
	for key, values := range header {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf(
			"%s\nGET %q failed: %v",
			packageName, path, err,
		)
	}
	t.Cleanup(func() { _ = response.Body.Close() })

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf(
			"%s\ncould not read the body of %q: %v",
			packageName, path, err,
		)
	}

	return response, string(body)
}

// startServing starts `server` over plain HTTP.
func startServing(t *testing.T, server *Server) *Server {
	t.Helper()

	return Start(t, server, httptest.NewServer)
}
