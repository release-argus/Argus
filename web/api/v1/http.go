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
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/ui"
)

// SetupRoutesAPI will set up the HTTP API routes.
func (api *API) SetupRoutesAPI() {
	// Create a subrouter for "/api/v1".
	v1Router := api.Router.PathPrefix("/api/v1").Subrouter()

	// Only if VERBOSE or DEBUG.
	if logutil.Log.Level >= 3 {
		// Apply loggerMiddleware to only the /api/v1 routes.
		v1Router.Use(loggerMiddleware)
	}

	// /config
	// Apply the logging middleware globally.
	//   GET, config.
	v1Router.HandleFunc("/config", api.httpConfig).Methods(http.MethodGet)
	// /status
	//   GET, runtime info.
	v1Router.HandleFunc("/status/runtime", api.httpRuntimeInfo).Methods(http.MethodGet)
	//   GET, build info.
	v1Router.HandleFunc("/version", api.httpVersion).Methods(http.MethodGet)
	// /flags
	//   GET, flags.
	v1Router.HandleFunc("/flags", api.httpFlags).Methods(http.MethodGet)
	// /approvals
	//   GET, service order.
	v1Router.HandleFunc("/service/order", api.httpServiceOrderGet).Methods(http.MethodGet)
	//   PUT, service order (disable=order_edit).
	v1Router.HandleFunc("/service/order", api.httpServiceOrderSet).Methods(http.MethodPut)
	//   GET, service summary.
	v1Router.HandleFunc("/service/summary/{service_id:.+}", api.httpServiceSummary).Methods(http.MethodGet)
	//   GET, service actions (webhooks/commands).
	v1Router.HandleFunc("/service/actions/{service_id:.+}", api.httpServiceGetActions).Methods(http.MethodGet)
	//   POST, service actions (disable=service_actions).
	v1Router.HandleFunc("/service/actions/{service_id:.+}", api.httpServiceRunActions).Methods(http.MethodPost)
	//   GET, service - get details.
	v1Router.HandleFunc("/service/update", api.httpOtherServiceDetails).Methods(http.MethodGet)
	v1Router.HandleFunc("/service/update/{service_id:.+}", api.httpServiceDetail).Methods(http.MethodGet)
	//   GET, service - refresh unsaved service (disable=[ld]v_refresh_new).
	v1Router.HandleFunc("/latest_version/refresh", api.httpLatestVersionRefreshUncreated).Methods(http.MethodGet)
	v1Router.HandleFunc("/deployed_version/refresh", api.httpDeployedVersionRefreshUncreated).Methods(http.MethodGet)
	//   GET, service - refresh service (disable=[ld]v_refresh).
	v1Router.HandleFunc("/latest_version/refresh/{service_id:.+}", api.httpLatestVersionRefresh).Methods(http.MethodGet)
	v1Router.HandleFunc("/deployed_version/refresh/{service_id:.+}", api.httpDeployedVersionRefresh).Methods(http.MethodGet)
	//   POST, service - test notify (disable=notify_test).
	v1Router.HandleFunc("/notify/test", api.httpNotifyTest).Methods(http.MethodPost)
	//   PUT, service - update details (disable=service_edit).
	v1Router.HandleFunc("/service/update/{service_id:.+}", api.httpServiceEdit).Methods(http.MethodPut)
	//   PUT, service - new service (disable=service_create).
	v1Router.HandleFunc("/service/new", api.httpServiceEdit).Methods(http.MethodPut)
	//   DELETE, service - delete service (disable=service_delete).
	v1Router.HandleFunc("/service/delete/{service_id:.+}", api.httpServiceDelete).Methods(http.MethodDelete)
	//   GET, service - template strings.
	v1Router.HandleFunc("/template", api.httpTemplateParse).Methods(http.MethodGet)
	// GET, counts for Heimdall.
	v1Router.HandleFunc("/counts", api.httpCounts).Methods(http.MethodGet)

	// Disable specified routes.
	api.DisableRoutes()
}

// DisableRoutes disables HTTP API routes marked as disabled in the config.
func (api *API) DisableRoutes() {
	// Trim suffix to ensure no trailing slash and prevent '//api/v1/...' routes.
	webRoutePrefix := strings.TrimSuffix(api.Config.Settings.WebRoutePrefix(), "/")
	routes := map[string]*struct {
		name         string
		method       string
		otherMethods map[string]func(w http.ResponseWriter, r *http.Request)
		disabled     bool
	}{
		webRoutePrefix + "/api/v1/service/order":                            {name: "order_edit", method: http.MethodPut},
		webRoutePrefix + "/api/v1/service/new":                              {name: "service_create", method: http.MethodPut},
		webRoutePrefix + "/api/v1/service/update/{service_id:.+}":           {name: "service_update", method: http.MethodPut},
		webRoutePrefix + "/api/v1/service/delete/{service_id:.+}":           {name: "service_delete", method: http.MethodDelete},
		webRoutePrefix + "/api/v1/notify/test":                              {name: "notify_test", method: http.MethodPost},
		webRoutePrefix + "/api/v1/latest_version/refresh/{service_id:.+}":   {name: "lv_refresh", method: http.MethodGet},
		webRoutePrefix + "/api/v1/deployed_version/refresh/{service_id:.+}": {name: "dv_refresh", method: http.MethodGet},
		webRoutePrefix + "/api/v1/latest_version/refresh":                   {name: "lv_refresh_new", method: http.MethodGet},
		webRoutePrefix + "/api/v1/deployed_version/refresh":                 {name: "dv_refresh_new", method: http.MethodGet},
		webRoutePrefix + "/api/v1/service/actions/{service_id:.+}":          {name: "service_actions", method: http.MethodPost},
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
				failRequest(&w,
					"Route disabled",
					http.StatusNotFound)
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
			http.Redirect(w, r,
				api.Config.Settings.Web.Favicon.SVG,
				http.StatusPermanentRedirect)
		})
	}
	if api.Config.Settings.Web.Favicon.PNG != "" {
		api.Router.HandleFunc("/apple-touch-icon.png", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r,
				api.Config.Settings.Web.Favicon.PNG,
				http.StatusPermanentRedirect)
		})
	}
}

// httpVersion serves Argus version JSON over HTTP.
func (api *API) httpVersion(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpVersion", Secondary: getIP(r)}

	api.writeJSON(w,
		apitype.VersionAPI{
			Version:   util.Version,
			BuildDate: util.BuildDate,
			GoVersion: util.GoVersion,
		},
		logFrom)
}

// failRequest returns a JSON response containing a message and status code.
func failRequest(w *http.ResponseWriter, message string, statusCode int) {
	// Write the response.
	(*w).WriteHeader(statusCode)
	resp := map[string]string{
		"message": message,
	}
	jsonResp, _ := json.Marshal(resp)
	//#nosec G104 -- Disregard.
	//nolint:errcheck // ^
	(*w).Write(jsonResp)
}
