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
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// API holds the configuration and routing information.
type API struct {
	Config      *config.Config
	BaseRouter  *mux.Router
	Router      *mux.Router
	RoutePrefix string
}

// NewAPI creates a new API with the provided config.
func NewAPI(cfg *config.Config) *API {
	baseRouter := mux.NewRouter().StrictSlash(true)
	routePrefix := cfg.Settings.WebRoutePrefix()

	api := &API{
		Config:      cfg,
		BaseRouter:  baseRouter,
		RoutePrefix: routePrefix,
	}

	// In cases where routePrefix equals "/", trim to prevent "//".
	routePrefix = strings.TrimSuffix(routePrefix, "/")
	// On baseRouter as Router may have basicAuth.
	baseRouter.Path(routePrefix + "/api/v1/healthcheck").
		Handler(
			loggerMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, "Alive")
				}),
			),
		)

	api.Router = baseRouter.PathPrefix(routePrefix).Subrouter().StrictSlash(true)

	baseRouter.Handle(
		routePrefix,
		http.RedirectHandler(routePrefix+"/", http.StatusPermanentRedirect),
	)
	// Add basic auth middleware.
	if api.Config.Settings.Web.BasicAuth != nil ||
		api.Config.Settings.FromFlags.Web.BasicAuth != nil ||
		api.Config.Settings.HardDefaults.Web.BasicAuth != nil {
		api.Router.Use(api.basicAuthMiddleware())
	}
	return api
}

// writeJSON marshals v as JSON and writes it to w with standard API response headers.
func (api *API) writeJSON(w http.ResponseWriter, v any, logFrom logx.LogFrom) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	b, err := decode.Marshal("json", v)
	if err != nil {
		// Encoding error, 500.
		w.WriteHeader(http.StatusInternalServerError)

		logx.Error(err, logFrom, true)

		api.writeJSON(
			w,
			apitype.Response{
				Error: err.Error(),
			},
			logFrom,
		)
		return
	}

	if _, err := w.Write(append(b, '\n')); err != nil {
		logx.Error(err, logFrom, true)
	}
}
