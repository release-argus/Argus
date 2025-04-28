// Copyright [2025] [Argus]
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

// Package v1 provides the API for the webserver.
package v1

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	logutil "github.com/release-argus/Argus/util/log"
)

// basicAuthMiddleware handles basic authentication with hashed credentials.
// It rejects unauthorised requests with a 401 and closes the connection.
func (api *API) basicAuthMiddleware() mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if ok {
				// Hash purely to prevent ConstantTimeCompare leaking lengths.
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))

				// Protect from possible timing attacks.
				usernameMatch := ConstantTimeCompare(usernameHash, api.Config.Settings.WebBasicAuthUsernameHash())
				passwordMatch := ConstantTimeCompare(passwordHash, api.Config.Settings.WebBasicAuthPasswordHash())

				if usernameMatch && passwordMatch {
					h.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Connection", "close")
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// loggerMiddleware logs the HTTP request method, client IP, and URL path before
// passing the request to the next handler in the chain.
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request.
		logutil.Log.Verbose(
			fmt.Sprintf("%s (%s), %s",
				r.Method, getIP(r), r.URL.Path,
			),
			logutil.LogFrom{},
			true)

		// Process request.
		next.ServeHTTP(w, r)
	})
}
