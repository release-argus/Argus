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
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/util"
)

// API is the API to use for the webserver.
type API struct {
	Config      *config.Config
	Log         *util.JLog
	BaseRouter  *mux.Router
	Router      *mux.Router
	RoutePrefix string
}

// NewAPI will create a new API with the provided config.
func NewAPI(cfg *config.Config, log *util.JLog) *API {
	baseRouter := mux.NewRouter().StrictSlash(true)
	routePrefix := "/" + strings.TrimPrefix(cfg.Settings.WebRoutePrefix(), "/")

	api := &API{
		Config:      cfg,
		Log:         log,
		BaseRouter:  baseRouter,
		RoutePrefix: routePrefix,
	}
	// On baseRouter as Router may have basicAuth
	baseRouter.Path(fmt.Sprintf("%s/api/v1/healthcheck", strings.TrimSuffix(routePrefix, "/"))).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logFrom := util.LogFrom{Primary: "apiHealthcheck", Secondary: getIP(r)}
		api.Log.Verbose("-", logFrom, true)
		w.Header().Set("Connection", "close")
		fmt.Fprintf(w, "Alive")
	})
	api.Router = baseRouter.PathPrefix(routePrefix).Subrouter().StrictSlash(true)

	baseRouter.Handle(routePrefix, http.RedirectHandler(routePrefix+"/", http.StatusPermanentRedirect))
	if api.Config.Settings.Web.BasicAuth != nil {
		api.Router.Use(api.basicAuth())
	}
	return api
}
