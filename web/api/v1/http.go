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

// Package v1 provides the API for the webserver.
package v1

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"slices"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/ui"
)

// SetupRoutesAPI sets up the HTTP API routes.
func (api *API) SetupRoutesAPI() {
	// Create a subrouter for "/api/v1".
	v1Router := api.Router.PathPrefix("/api/v1").Subrouter()

	// Only if VERBOSE or DEBUG.
	if logx.Level() >= 3 {
		// Apply loggerMiddleware to only the "/api/v1" routes.
		v1Router.Use(loggerMiddleware)
	}

	// Session/RBAC authentication.
	if api.auth != nil {
		v1Router.Use(api.originCheckMiddleware, api.authMiddleware())

		//   POST, log in (exempt from the auth middleware).
		v1Router.HandleFunc("/auth/login", api.httpAuthLogin).Methods(http.MethodPost)
		//   First-run setup - GET, whether it is pending; POST, create the
		//   first administrator (both exempt from the auth middleware).
		v1Router.HandleFunc("/auth/setup", api.httpAuthSetupState).Methods(http.MethodGet)
		v1Router.HandleFunc("/auth/setup", api.httpAuthSetup).Methods(http.MethodPost)
		//   POST, log out.
		v1Router.HandleFunc("/auth/logout", api.httpAuthLogout).Methods(http.MethodPost)
		//   GET, the authenticated user and their permissions.
		v1Router.HandleFunc("/auth/me", api.httpAuthMe).Methods(http.MethodGet)
		// Users - CRUD (admin-only).
		v1Router.HandleFunc("/users",
			api.requireAdmin(api.httpUserList)).Methods(http.MethodGet)
		v1Router.HandleFunc("/users",
			api.requireAdmin(api.httpUserCreate)).Methods(http.MethodPost)
		v1Router.HandleFunc("/users/{id}",
			api.requireAdmin(api.httpUserGet)).Methods(http.MethodGet)
		v1Router.HandleFunc("/users/{id}",
			api.requireAdmin(api.httpUserUpdate)).Methods(http.MethodPatch)
		v1Router.HandleFunc("/users/{id}",
			api.requireAdmin(api.httpUserDelete)).Methods(http.MethodDelete)
		// Groups - CRUD (admin-only).
		v1Router.HandleFunc("/groups",
			api.requireAdmin(api.httpGroupList)).Methods(http.MethodGet)
		v1Router.HandleFunc("/groups",
			api.requireAdmin(api.httpGroupCreate)).Methods(http.MethodPost)
		v1Router.HandleFunc("/groups/{id}",
			api.requireAdmin(api.httpGroupGet)).Methods(http.MethodGet)
		v1Router.HandleFunc("/groups/{id}",
			api.requireAdmin(api.httpGroupUpdate)).Methods(http.MethodPatch)
		v1Router.HandleFunc("/groups/{id}",
			api.requireAdmin(api.httpGroupDelete)).Methods(http.MethodDelete)
		//   GET, the valid permission matrix (read-only; grants are edited on groups).
		v1Router.HandleFunc("/permissions",
			api.requireAdmin(api.httpPermissionCatalogue)).Methods(http.MethodGet)
		// API tokens.
		v1Router.HandleFunc("/tokens", api.httpAPITokenList).Methods(http.MethodGet)
		v1Router.HandleFunc("/tokens", api.httpAPITokenCreate).Methods(http.MethodPost)
		v1Router.HandleFunc("/tokens/{id}", api.httpAPITokenDelete).Methods(http.MethodDelete)
	}

	// /config
	// Apply the logging middleware globally.
	//   GET, config.
	v1Router.HandleFunc("/config",
		api.guard(rbac.ResourceConfig, rbac.ActionRead, nil, api.httpConfig)).Methods(http.MethodGet)
	// /status
	//   GET, runtime info.
	v1Router.HandleFunc("/status/runtime",
		api.guard(rbac.ResourceConfig, rbac.ActionRead, nil, api.httpRuntimeInfo)).Methods(http.MethodGet)
	//   GET, build info.
	v1Router.HandleFunc("/version",
		api.guard(rbac.ResourceConfig, rbac.ActionRead, nil, api.httpVersion)).Methods(http.MethodGet)
	//   GET, short-lived token for authenticating the "/ws" WebSocket handshake (only used when basic_auth is enabled).
	v1Router.HandleFunc("/ws-token", api.httpWebSocketToken).Methods(http.MethodGet)
	// /flags
	//   GET, flags.
	v1Router.HandleFunc("/flags",
		api.guard(rbac.ResourceConfig, rbac.ActionRead, nil, api.httpFlags)).Methods(http.MethodGet)
	// /approvals
	//   GET, service order (filtered per-user inside the handler:
	//   scoped users receive only the services they may read).
	v1Router.HandleFunc("/service/order", api.httpServiceOrderGet).Methods(http.MethodGet)
	//   PUT, service order (disable=order_edit).
	v1Router.HandleFunc("/service/order",
		api.guard(rbac.ResourceServiceOrder, rbac.ActionUpdate, nil, api.httpServiceOrderSet)).Methods(http.MethodPut)
	//   GET, service summary.
	v1Router.HandleFunc("/service/summary",
		api.guard(rbac.ResourceService, rbac.ActionRead, api.serviceTarget, api.httpServiceSummary)).Methods(http.MethodGet)
	//   GET, service actions (webhooks/commands).
	v1Router.HandleFunc("/service/actions",
		api.guard(rbac.ResourceService, rbac.ActionRead, api.serviceTarget, api.httpServiceGetActions)).Methods(http.MethodGet)
	//   POST, service actions (disable=service_actions).
	v1Router.HandleFunc("/service/actions",
		api.guardReadable(rbac.ResourceServiceAction, rbac.ActionExecute, api.serviceTarget, api.httpServiceRunActions)).Methods(http.MethodPost)
	//   GET, service - get details on specific service.
	v1Router.HandleFunc("/service/config",
		api.guard(rbac.ResourceService, rbac.ActionRead, api.serviceTarget, api.httpServiceDetail)).Methods(http.MethodGet)
	//   GET, service - get details on service defaults.
	v1Router.HandleFunc("/service/defaults",
		api.guardAnyScope(rbac.ResourceService, rbac.ActionRead, api.httpOtherServiceDetails)).Methods(http.MethodGet)
	//   GET, service - refresh unsaved service (disable=[ld]v_refresh_new).
	v1Router.HandleFunc("/latest_version/refresh_uncreated",
		api.guard(rbac.ResourceService, rbac.ActionCreate, nil, api.httpLatestVersionRefreshUncreated)).Methods(http.MethodGet)
	v1Router.HandleFunc("/deployed_version/refresh_uncreated",
		api.guard(rbac.ResourceService, rbac.ActionCreate, nil, api.httpDeployedVersionRefreshUncreated)).Methods(http.MethodGet)
	//   GET, service - refresh service (disable=[ld]v_refresh).
	v1Router.HandleFunc("/latest_version/refresh",
		api.guardReadable(rbac.ResourceVersionRefresh, rbac.ActionExecute, api.serviceTarget, api.httpLatestVersionRefresh)).Methods(http.MethodGet)
	v1Router.HandleFunc("/deployed_version/refresh",
		api.guardReadable(rbac.ResourceVersionRefresh, rbac.ActionExecute, api.serviceTarget, api.httpDeployedVersionRefresh)).Methods(http.MethodGet)
	//   POST, service - test notify (disable=notify_test).
	v1Router.HandleFunc("/notify/test",
		api.guard(rbac.ResourceNotify, rbac.ActionExecute, nil, api.httpNotifyTest)).Methods(http.MethodPost)
	//   PUT, service - update details (disable=service_edit).
	v1Router.HandleFunc("/service/config",
		api.guard(rbac.ResourceService, rbac.ActionUpdate, api.serviceTarget, api.httpServiceEdit)).Methods(http.MethodPut)
	//   PUT, service - new service (disable=service_create).
	v1Router.HandleFunc("/service/new",
		api.guard(rbac.ResourceService, rbac.ActionCreate, nil, api.httpServiceEdit)).Methods(http.MethodPut)
	//   DELETE, service - delete service (disable=service_delete).
	v1Router.HandleFunc("/service/delete",
		api.guard(rbac.ResourceService, rbac.ActionDelete, api.serviceTarget, api.httpServiceDelete)).Methods(http.MethodDelete)
	//   GET, service - template strings.
	v1Router.HandleFunc("/template",
		api.guardAnyScope(rbac.ResourceService, rbac.ActionRead, api.httpTemplateParse)).Methods(http.MethodGet)
	// GET, counts for Heimdall.
	v1Router.HandleFunc("/counts",
		api.guard(rbac.ResourceConfig, rbac.ActionRead, nil, api.httpCounts)).Methods(http.MethodGet)

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
		webRoutePrefix + "/api/v1/service/order":                      {name: "order_edit", method: http.MethodPut},
		webRoutePrefix + "/api/v1/service/new":                        {name: "service_create", method: http.MethodPut},
		webRoutePrefix + "/api/v1/service/config":                     {name: "service_update", method: http.MethodPut},
		webRoutePrefix + "/api/v1/service/delete":                     {name: "service_delete", method: http.MethodDelete},
		webRoutePrefix + "/api/v1/notify/test":                        {name: "notify_test", method: http.MethodPost},
		webRoutePrefix + "/api/v1/latest_version/refresh":             {name: "lv_refresh", method: http.MethodGet},
		webRoutePrefix + "/api/v1/deployed_version/refresh":           {name: "dv_refresh", method: http.MethodGet},
		webRoutePrefix + "/api/v1/latest_version/refresh_uncreated":   {name: "lv_refresh_new", method: http.MethodGet},
		webRoutePrefix + "/api/v1/deployed_version/refresh_uncreated": {name: "dv_refresh_new", method: http.MethodGet},
		webRoutePrefix + "/api/v1/service/actions":                    {name: "service_actions", method: http.MethodPost},
	}
	for _, r := range routes {
		r.disabled = slices.Contains(api.Config.Settings.Web.DisabledRoutes, r.name)
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
				failRequest(
					&w,
					errors.New("route disabled"),
					http.StatusNotFound,
				)
				return
			}

			// Call the original handler for other methods.
			handler.(http.HandlerFunc)(w, r) // Cast the handler to http.HandlerFunc before calling it.
		})

		return nil
	})
}

// SetupWebSocket fills the handler for the "/ws" route.
// The route must be outside the basic-auth subrouter because Safari/WebKit
// doesn't forward cached credentials on WebSocket handshakes.
// When Basic Auth is enabled the client instead passes a short-lived token
// as a query parameter; when session auth is enabled, the same-origin
// handshake carries the session cookie, which is validated directly.
func (api *API) SetupWebSocket(hub *Hub, wsRoute *mux.Route) {
	api.hub = hub
	wsRoute.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// auth identifies the session/user behind the connection
		// (nil = unrestricted).
		var auth *clientAuth

		// Session auth: the handshake carries the session cookie.
		if api.auth != nil {
			cookie, err := r.Cookie(authCookieName)
			if err != nil {
				http.Error(w, errUnauthorised.Error(), http.StatusUnauthorized)
				return
			}
			authCtx, err := api.authenticateSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, errUnauthorised.Error(), http.StatusUnauthorized)
				return
			}
			sessionHash := session.HashToken(cookie.Value)
			auth = &clientAuth{
				userID:          authCtx.User.ID,
				sessionHash:     sessionHash,
				allowedServices: api.allowedServices(authCtx),
				sessionAlive: func() bool {
					// Bounded so a stalled store can't hang the write pump goroutine.
					ctx, cancel := context.WithTimeout(context.Background(), writeWait)
					defer cancel()
					return api.auth.Sessions.Alive(ctx, sessionHash)
				},
			}
		} else if api.wsTokens != nil && !api.wsTokens.Validate(r.URL.Query().Get("token")) {
			http.Error(w, errUnauthorised.Error(), http.StatusUnauthorized)
			return
		}
		w.Header().Set("Connection", "keep-alive")
		defer r.Body.Close()
		ServeWs(hub, w, r, auth)
	})
}

// GuardMetrics wraps the /metrics handler: when auth is enabled it requires
// authentication plus config:read.
func (api *API) GuardMetrics(handler http.Handler) http.Handler {
	if api.auth == nil {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCtx, err := api.authenticate(r)
		if err != nil {
			failAuthenticationError(w, r, err)
			return
		}
		if !authCtx.Permissions.Allowed(rbac.ResourceConfig, rbac.ActionRead, nil) {
			failRequest(&w, errForbidden, http.StatusForbidden)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// SetupRoutesNodeJS sets up the HTTP routes to the Node.js files.
func (api *API) SetupRoutesNodeJS() {
	nodeRoutes := []string{
		"/approvals",
		"/config",
		"/flags",
		"/groups",
		"/login",
		"/status",
		"/tokens",
		"/users",
	}
	// Serve the Node.js files.
	for _, route := range nodeRoutes {
		prefix := strings.TrimRight(api.RoutePrefix, "/") + route
		api.Router.Handle(
			route,
			http.StripPrefix(
				prefix,
				statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding),
			),
		)
	}

	// Favicon override.
	api.SetupRoutesFavicon()

	// Catch-all for JS, CSS, etc...
	api.Router.PathPrefix("/").Handler(
		http.StripPrefix(
			api.RoutePrefix,
			statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding),
		),
	)
}

// SetupRoutesFavicon adds any favicon route overrides.
func (api *API) SetupRoutesFavicon() {
	if api.Config.Settings.Web.Favicon == nil {
		return
	}

	if api.Config.Settings.Web.Favicon.SVG != "" {
		api.Router.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(
				w, r,
				api.Config.Settings.Web.Favicon.SVG,
				http.StatusPermanentRedirect,
			)
		})
	}
	if api.Config.Settings.Web.Favicon.PNG != "" {
		api.Router.HandleFunc("/apple-touch-icon.png", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(
				w, r,
				api.Config.Settings.Web.Favicon.PNG,
				http.StatusPermanentRedirect,
			)
		})
	}
}

// httpVersion handles GET /api/v1/version: serving the Argus build info
// (version, build date, Go version).
//
// Response:
//
//	200 OK: JSON object containing the build info.
func (api *API) httpVersion(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpVersion", Secondary: getIP(r)}

	api.writeJSON(
		w,
		apitype.VersionAPI{
			Version:   util.Version,
			BuildDate: util.BuildDate,
			GoVersion: util.GoVersion,
		},
		logFrom,
	)
}

// marshalFailRequestBody serialises failRequest responses (overridable for tests).
// see [decode.Marshal].
var marshalFailRequestBody = func(v map[string]string) ([]byte, error) {
	return decode.Marshal("json", v)
}

// failRequest returns a JSON response containing a message and status code.
func failRequest(w *http.ResponseWriter, err error, statusCode int) {
	// Write the response.
	(*w).WriteHeader(statusCode)
	resp := map[string]string{
		"message": errfmt.FormatError(err),
	}
	jsonResp, marshalErr := marshalFailRequestBody(resp)
	if marshalErr != nil {
		logx.Error(marshalErr, logx.LogFrom{Primary: "failRequest"}, true)
		jsonResp = []byte(`{"message":"failed to encode error response"}`)
	}
	if _, writeErr := (*w).Write(jsonResp); writeErr != nil {
		logx.Error(writeErr, logx.LogFrom{Primary: "failRequest"}, true)
	}
}
