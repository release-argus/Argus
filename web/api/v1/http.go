// Copyright [2023] [Argus]
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

package v1

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/ui"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

func (api *API) basicAuth() mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			username, password, ok := r.BasicAuth()
			if ok {
				// Hash purely to prevent ConstantTimeCompare leaking lengths
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))
				expectedUsernameHash := sha256.Sum256([]byte(api.Config.Settings.Web.BasicAuth.Username))
				expectedPasswordHash := sha256.Sum256([]byte(api.Config.Settings.Web.BasicAuth.Password))

				// Protect from possible timing attacks
				usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
				passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

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

// SetupRoutesAPI will setup the HTTP API routes.
func (api *API) SetupRoutesAPI() {
	// /config
	//   GET, config
	api.Router.HandleFunc("/api/v1/config", api.httpConfig).Methods("GET")
	// /status
	//   GET, runtime info
	api.Router.HandleFunc("/api/v1/status/runtime", api.httpRuntimeInfo).Methods("GET")
	//   GET, build info
	api.Router.HandleFunc("/api/v1/version", api.httpVersion).Methods("GET")
	// /flags
	//   GET, flags
	api.Router.HandleFunc("/api/v1/flags", api.httpFlags).Methods("GET")
	// /approvals
	//   GET, service order
	api.Router.HandleFunc("/api/v1/service/order", api.httpServiceOrder).Methods("GET")
	//   GET, service summary
	api.Router.HandleFunc("/api/v1/service/summary/{service_name:.+}", api.httpServiceSummary).Methods("GET")
	//   GET, service actions (webhooks/commands)
	api.Router.HandleFunc("/api/v1/service/actions/{service_name:.+}", api.httpServiceGetActions).Methods("GET")
	//   POST, service actions
	api.Router.HandleFunc("/api/v1/service/actions/{service_name:.+}", api.httpServiceRunActions).Methods("POST")
	//   GET, service-edit - get details
	api.Router.HandleFunc("/api/v1/service/edit", api.httpOtherServiceDetails).Methods("GET")
	api.Router.HandleFunc("/api/v1/service/edit/{service_name:.+}", api.httpServiceDetail).Methods("GET")
	//   GET, service-edit - refresh unsaved service
	api.Router.HandleFunc("/api/v1/latest_version/refresh", api.httpVersionRefreshUncreated).Methods("GET")
	api.Router.HandleFunc("/api/v1/deployed_version/refresh", api.httpVersionRefreshUncreated).Methods("GET")
	//   GET, service-edit - refresh
	api.Router.HandleFunc("/api/v1/latest_version/refresh/{service_name:.+}", api.httpVersionRefresh).Methods("GET")
	api.Router.HandleFunc("/api/v1/deployed_version/refresh/{service_name:.+}", api.httpVersionRefresh).Methods("GET")
	//   PUT, service-edit - update details
	api.Router.HandleFunc("/api/v1/service/edit/{service_name:.+}", api.httpServiceEdit).Methods("PUT")
	//   POST, service-edit - new service
	api.Router.HandleFunc("/api/v1/service/new", api.httpServiceEdit).Methods("POST")
	//   DELETE, service-edit
	api.Router.HandleFunc("/api/v1/service/delete/{service_name:.+}", api.httpServiceDelete).Methods("DELETE")
}

// SetupRoutesNodeJS will setup the HTTP routes to the NodeJS files.
func (api *API) SetupRoutesNodeJS() {
	nodeRoutes := []string{
		"/approvals",
		"/config",
		"/flags",
		"/status",
	}
	// Serve the NodeJS files
	for _, route := range nodeRoutes {
		prefix := strings.TrimRight(api.RoutePrefix, "/") + route
		api.Router.Handle(route, http.StripPrefix(prefix, statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
	}

	// Catch-all for JS, CSS, etc...
	api.Router.PathPrefix("/").Handler(http.StripPrefix(api.RoutePrefix, statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
}

// httpVersion serves Argus version JSON over HTTP.
func (api *API) httpVersion(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpVersion", Secondary: getIP(r)}
	api.Log.Verbose("-", logFrom, true)

	// Set headers
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(api_type.VersionAPI{
		Version:   util.Version,
		BuildDate: util.BuildDate,
		GoVersion: util.GoVersion,
	})
	api.Log.Error(err, logFrom, err != nil)
}

// failRequest with a JSON response containing a message and status code.
func failRequest(w *http.ResponseWriter, message string, statusCodeOverride ...int) {
	// Default to 400 Bad Request
	statusCode := http.StatusBadRequest
	// Override if provided
	if len(statusCodeOverride) > 0 {
		statusCode = statusCodeOverride[0]
	}

	// Write response
	(*w).WriteHeader(statusCode)
	resp := map[string]string{
		"message": message,
	}
	jsonResp, _ := json.Marshal(resp)
	//nolint:errcheck
	(*w).Write(jsonResp)
}
