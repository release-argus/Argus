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

package v1

import (
	"net/http"
	"strings"

	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/web/ui"
)

// SetupRoutesNodeJS sets up the HTTP routes to the Node.js files.
func (api *API) SetupRoutesNodeJS() {
	nodeRoutes := []string{
		"/account",
		"/account/tokens",
		"/admin",
		"/admin/groups",
		"/admin/users",
		"/approvals",
		"/config",
		"/flags",
		"/login",
		"/status",
	}

	// One server over the embedded UI, with the app shell
	// rewritten for this route prefix.
	fSys, err := ui.Rewritten(api.RoutePrefix)
	if err != nil {
		logx.Error(err, logx.LogFrom{Primary: "SetupRoutesNodeJS"}, true)
	}
	fileServer := statigz.FileServer(fSys, brotli.AddEncoding)

	// Serve the Node.js files.
	for _, route := range nodeRoutes {
		prefix := strings.TrimRight(api.RoutePrefix, "/") + route
		api.Router.Handle(route, http.StripPrefix(prefix, fileServer))
	}

	// Favicon override.
	api.SetupRoutesFavicon()

	// Catch-all for JS, CSS, etc... and for the route prefix's own root, which
	// resolves to the shell.
	api.Router.PathPrefix("/").Handler(
		http.StripPrefix(api.RoutePrefix, fileServer),
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
