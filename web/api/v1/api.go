// Copyright [2022] [Hymenaios]
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
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hymenaios-io/Hymenaios/config"
	"github.com/hymenaios-io/Hymenaios/utils"
	"github.com/hymenaios-io/Hymenaios/web/ui"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

// API is the API to use for the webserver.
type API struct {
	Config      *config.Config
	Log         *utils.JLog
	BaseRouter  *mux.Router
	Router      *mux.Router
	RoutePrefix string
}

// NewAPI will create a new API with the provided config.
func NewAPI(cfg *config.Config, log *utils.JLog) *API {
	baseRouter := mux.NewRouter().StrictSlash(true)
	routePrefix := "/" + strings.TrimPrefix(cfg.Settings.GetWebRoutePrefix(), "/")
	api := &API{
		Config:      cfg,
		Log:         log,
		BaseRouter:  baseRouter,
		Router:      baseRouter.PathPrefix(routePrefix).Subrouter().StrictSlash(true),
		RoutePrefix: routePrefix,
	}
	baseRouter.Handle(routePrefix, http.RedirectHandler(routePrefix+"/", 302))
	return api
}

// SetupRoutesNodeJS will setup the HTTP routes to the NodeJS files.
func (api *API) SetupRoutesNodeJS() {
	nodeRoutes := []string{
		"/approvals",
		"/config",
		"/flags",
		"/status",
	}
	for _, route := range nodeRoutes {
		prefix := strings.TrimRight(api.RoutePrefix, "/") + route
		api.Router.Handle(route, http.StripPrefix(prefix, statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
	}

	// Catch-all for JS, CSS, etc...
	api.Router.PathPrefix("/").Handler(http.StripPrefix(api.RoutePrefix, statigz.FileServer(ui.GetFS().(fs.ReadDirFS), brotli.AddEncoding)))
}
