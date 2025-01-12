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
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/util"
)

// API holds the configuration and routing information.
type API struct {
	Config      *config.Config
	BaseRouter  *mux.Router
	Router      *mux.Router
	RoutePrefix string
}

// NewAPI will create a new API with the provided config.
func NewAPI(cfg *config.Config, log *util.JLog) *API {
	LogInit(log)

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
	baseRouter.Path(fmt.Sprintf("%s/api/v1/healthcheck", routePrefix)).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logFrom := util.LogFrom{Primary: "apiHealthcheck", Secondary: getIP(r)}
		jLog.Verbose("-", logFrom, true)
		w.Header().Set("Connection", "close")
		fmt.Fprintf(w, "Alive")
	})
	api.Router = baseRouter.PathPrefix(routePrefix).Subrouter().StrictSlash(true)

	baseRouter.Handle(routePrefix, http.RedirectHandler(routePrefix+"/", http.StatusPermanentRedirect))
	if api.Config.Settings.Web.BasicAuth != nil ||
		api.Config.Settings.FromFlags.Web.BasicAuth != nil ||
		api.Config.Settings.HardDefaults.Web.BasicAuth != nil {
		api.Router.Use(api.basicAuth())
	}
	return api
}

func (api *API) writeJSON(w http.ResponseWriter, v interface{}, logFrom util.LogFrom) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Encoding error, 500.
		w.WriteHeader(http.StatusInternalServerError)
		jLog.Error(err, logFrom, true)
		return
	}
}
