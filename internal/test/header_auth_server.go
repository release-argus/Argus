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

//go:build unit || integration

package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// HeaderAuthServer starts a stand-in for the [LookupWithHeaderAuth] hook, and returns its URL.
// It responds with version when the [LookupWithHeaderAuth].HeaderKey header holds its
// HeaderValuePass, and with the hook's rejection otherwise.
func HeaderAuthServer(t *testing.T, version string) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(LookupWithHeaderAuth.HeaderKey) != LookupWithHeaderAuth.HeaderValuePass {
			_, _ = w.Write([]byte("Hook rules were not satisfied."))
			return
		}
		_, _ = w.Write([]byte(version))
	}))
	t.Cleanup(server.Close)
	return server.URL + "/hooks/single-header"
}
