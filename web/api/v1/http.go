// Copyright [2024] [Argus]
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
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/ui"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

func (api *API) basicAuth() mux.MiddlewareFunc {
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

// setCommonHeaders sets common headers for JSON API responses.
func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")
}

// SetupRoutesAPI will set up the HTTP API routes.
func (api *API) SetupRoutesAPI() {
	// /config
	//   GET, config.
	api.Router.HandleFunc("/api/v1/config", api.httpConfig).Methods("GET")
	// /status
	//   GET, runtime info.
	api.Router.HandleFunc("/api/v1/status/runtime", api.httpRuntimeInfo).Methods("GET")
	//   GET, build info.
	api.Router.HandleFunc("/api/v1/version", api.httpVersion).Methods("GET")
	// /flags
	//   GET, flags.
	api.Router.HandleFunc("/api/v1/flags", api.httpFlags).Methods("GET")
	// /approvals
	//   GET, service order.
	api.Router.HandleFunc("/api/v1/service/order", api.httpServiceOrder).Methods("GET")
	//   GET, service summary.
	api.Router.HandleFunc("/api/v1/service/summary/{service_name:.+}", api.httpServiceSummary).Methods("GET")
	//   GET, service actions (webhooks/commands).
	api.Router.HandleFunc("/api/v1/service/actions/{service_name:.+}", api.httpServiceGetActions).Methods("GET")
	//   POST, service actions (disable=service_actions).
	api.Router.HandleFunc("/api/v1/service/actions/{service_name:.+}", api.httpServiceRunActions).Methods("POST")
	//   GET, service-edit - get details.
	api.Router.HandleFunc("/api/v1/service/update", api.httpOtherServiceDetails).Methods("GET")
	api.Router.HandleFunc("/api/v1/service/update/{service_name:.+}", api.httpServiceDetail).Methods("GET")
	//   GET, service-edit - refresh unsaved service (disable=[ld]v_refresh_new).
	api.Router.HandleFunc("/api/v1/latest_version/refresh", api.httpLatestVersionRefreshUncreated).Methods("GET")
	api.Router.HandleFunc("/api/v1/deployed_version/refresh", api.httpDeployedVersionRefreshUncreated).Methods("GET")
	//   GET, service-edit - refresh service (disable=[ld]v_refresh).
	api.Router.HandleFunc("/api/v1/latest_version/refresh/{service_name:.+}", api.httpLatestVersionRefresh).Methods("GET")
	api.Router.HandleFunc("/api/v1/deployed_version/refresh/{service_name:.+}", api.httpDeployedVersionRefresh).Methods("GET")
	//   POST, service-edit - test notify (disable=notify_test).
	api.Router.HandleFunc("/api/v1/notify/test", api.httpNotifyTest).Methods("POST")
	//   PUT, service-edit - update details (disable=service_edit).
	api.Router.HandleFunc("/api/v1/service/update/{service_name:.+}", api.httpServiceEdit).Methods("PUT")
	//   PUT, service-edit - new service (disable=service_create).
	api.Router.HandleFunc("/api/v1/service/new", api.httpServiceEdit).Methods("PUT")
	//   DELETE, service-edit - delete service (disable=service_delete).
	api.Router.HandleFunc("/api/v1/service/delete/{service_name:.+}", api.httpServiceDelete).Methods("DELETE")

	// Disable specified routes.
	api.DisableRoutesAPI()
}

// DisableRoutesAPI disables HTTP API routes marked as disabled in the config.
func (api *API) DisableRoutesAPI() {
	// Trim suffix to ensure no trailing slash and prevent '//api/v1/...' routes.
	webRoutePrefix := strings.TrimSuffix(api.Config.Settings.WebRoutePrefix(), "/")
	routes := map[string]*struct {
		name         string
		method       string
		otherMethods map[string]func(w http.ResponseWriter, r *http.Request)
		disabled     bool
	}{
		webRoutePrefix + "/api/v1/service/new":                                {name: "service_create", method: "PUT"},
		webRoutePrefix + "/api/v1/service/update/{service_name:.+}":           {name: "service_update", method: "PUT"},
		webRoutePrefix + "/api/v1/service/delete/{service_name:.+}":           {name: "service_delete", method: "DELETE"},
		webRoutePrefix + "/api/v1/notify/test":                                {name: "notify_test", method: "POST"},
		webRoutePrefix + "/api/v1/latest_version/refresh/{service_name:.+}":   {name: "lv_refresh", method: "GET"},
		webRoutePrefix + "/api/v1/deployed_version/refresh/{service_name:.+}": {name: "dv_refresh", method: "GET"},
		webRoutePrefix + "/api/v1/latest_version/refresh":                     {name: "lv_refresh_new", method: "GET"},
		webRoutePrefix + "/api/v1/deployed_version/refresh":                   {name: "dv_refresh_new", method: "GET"},
		webRoutePrefix + "/api/v1/service/actions/{service_name:.+}":          {name: "service_actions", method: "POST"},
	}
	for _, r := range routes {
		r.disabled = util.Contains(api.Config.Settings.Web.DisabledRoutes, r.name)
	}

	_ = api.Router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		// Return an error if the route does not define a path.
		routePath, _ := route.GetPathTemplate()

		// Ignore routes not defined above or disabled.
		r := routes[routePath]
		if r == nil || !r.disabled {
			return nil
		}

		handler := route.GetHandler()

		// Set the new handler for the route.
		disabledMethod := r.method
		route.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			disabledMethod := disabledMethod
			if r.Method == disabledMethod {
				failRequest(&w, "Route disabled", http.StatusNotFound)
				return
			}

			// Call the original handler for other methods.
			handler.(http.HandlerFunc)(w, r) // Cast the handler to http.HandlerFunc before calling it.
		})

		return nil
	})
}

// SetupRoutesNodeJS will set up the HTTP routes to the Node.js files.
func (api *API) SetupRoutesNodeJS() {
	nodeRoutes := []string{
		"/approvals",
		"/config",
		"/flags",
		"/status",
	}
	// Serve the Node.js files.
	for _, route := range nodeRoutes {
		prefix := strings.TrimRight(api.RoutePrefix, "/") + route
		api.Router.Handle(route, http.StripPrefix(prefix,
			statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
	}

	// Favicon override.
	api.SetupRoutesFavicon()

	// Catch-all for JS, CSS, etc...
	api.Router.PathPrefix("/").Handler(http.StripPrefix(api.RoutePrefix,
		statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
}

// SetupRoutesFavicon adds any favicon route overrides.
func (api *API) SetupRoutesFavicon() {
	if api.Config.Settings.Web.Favicon == nil {
		return
	}

	if api.Config.Settings.Web.Favicon.SVG != "" {
		api.Router.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, api.Config.Settings.Web.Favicon.SVG, http.StatusPermanentRedirect)
		})
	}
	if api.Config.Settings.Web.Favicon.PNG != "" {
		api.Router.HandleFunc("/apple-touch-icon.png", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, api.Config.Settings.Web.Favicon.PNG, http.StatusPermanentRedirect)
		})
	}
}

// httpVersion serves Argus version JSON over HTTP.
func (api *API) httpVersion(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	logFrom := util.LogFrom{Primary: "httpVersion", Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	err := json.NewEncoder(w).Encode(apitype.VersionAPI{
		Version:   util.Version,
		BuildDate: util.BuildDate,
		GoVersion: util.GoVersion,
	})
	jLog.Error(err, logFrom, err != nil)
}

// failRequest returns a JSON response containing a message and status code.
func failRequest(w *http.ResponseWriter, message string, statusCode int) {
	// Write response.
	(*w).WriteHeader(statusCode)
	resp := map[string]string{
		"message": message,
	}
	jsonResp, _ := json.Marshal(resp)
	//#nosec G104 -- Disregard.
	//nolint:errcheck -- ^
	(*w).Write(jsonResp)
}
