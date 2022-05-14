// Copyright [2022] [Argus]
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
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/release-argus/Argus/utils"
	api_types "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/web/ui"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

// SetupRoutesAPI will setup the HTTP API routes.
func (api *API) SetupRoutesAPI() {
	api.Router.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		api.httpVersion(w, r)
	})
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

// httpVersion serves Argus version JSON over HTTP.
func (api *API) httpVersion(w http.ResponseWriter, r *http.Request) {
	logFrom := utils.LogFrom{Primary: "apiVersion"}
	api.Log.Verbose("-", logFrom, true)
	err := json.NewEncoder(w).Encode(api_types.VersionAPI{
		Version:   utils.Version,
		BuildDate: utils.BuildDate,
		GoVersion: utils.GoVersion,
	})
	api.Log.Error(err, logFrom, err != nil)
}
