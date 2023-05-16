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

package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/util"
	api_v1 "github.com/release-argus/Argus/web/api/v1"
)

var jLog *util.JLog

// NewRouter that serves the Prometheus metrics,
// WebSocket and NodeJS frontend at the RoutePrefix.
func NewRouter(cfg *config.Config, jLog *util.JLog, hub *api_v1.Hub) *mux.Router {
	// Go
	api := api_v1.NewAPI(cfg, jLog)

	// Prometheus metrics
	api.Router.Handle("/metrics", promhttp.Handler())

	// WebSocket
	api.Router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		defer r.Body.Close()
		api_v1.ServeWs(api, hub, w, r)
	})

	// HTTP API
	api.SetupRoutesAPI()
	// NodeJS
	api.SetupRoutesNodeJS()

	return api.BaseRouter
}

// newWebUI will set up everything web-related for Argus.
func newWebUI(cfg *config.Config) *mux.Router {
	hub := api_v1.NewHub()
	go hub.Run(jLog)
	router := NewRouter(cfg, jLog, hub)

	// Hand out the broadcast channel
	cfg.HardDefaults.Service.Status.AnnounceChannel = &hub.Broadcast
	for sKey := range cfg.Service {
		cfg.Service[sKey].Status.SetAnnounceChannel(&hub.Broadcast)
	}

	return router
}

func Run(cfg *config.Config, log *util.JLog) {
	jLog = log
	router := newWebUI(cfg)

	listenAddress := fmt.Sprintf("%s:%s", cfg.Settings.WebListenHost(), cfg.Settings.WebListenPort())
	jLog.Info("Listening on "+listenAddress+cfg.Settings.WebRoutePrefix(), util.LogFrom{}, true)

	if cfg.Settings.WebCertFile() != nil && cfg.Settings.WebKeyFile() != nil {
		jLog.Fatal(
			http.ListenAndServeTLS(
				listenAddress, *cfg.Settings.WebCertFile(), *cfg.Settings.WebKeyFile(), router),
			util.LogFrom{}, true)
	} else {
		jLog.Fatal(
			http.ListenAndServe(
				listenAddress, router),
			util.LogFrom{}, true)
	}
}
