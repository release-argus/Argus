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

package web

import (
	"fmt"
	"net/http"

	"github.com/hymenaios-io/Hymenaios/config"
	"github.com/hymenaios-io/Hymenaios/utils"
	api_v1 "github.com/hymenaios-io/Hymenaios/web/api/v1"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var jLog *utils.JLog

// NewRouter that serves the Prometheus metrics,
// WebSocket and NodeJS frontend at the RoutePrefix.
func NewRouter(cfg *config.Config, jLog *utils.JLog, hub *api_v1.Hub) *mux.Router {
	// Go
	api := api_v1.NewAPI(cfg, jLog)

	// Prometheus metrics
	api.Router.Handle("/metrics", promhttp.Handler())

	// WebSocket
	api.Router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		api_v1.ServeWs(api, hub, w, r)
	})

	// NodeJS
	api.SetupRoutesNodeJS()

	return api.BaseRouter
}

// Main will set up everything web-related for Hymenaios.
func Main(cfg *config.Config, log *utils.JLog) {
	jLog = log
	hub := api_v1.NewHub()
	go hub.Run(jLog)
	router := NewRouter(cfg, jLog, hub)

	// Hand out the broadcast channel
	for sKey := range cfg.Service {
		cfg.Service[sKey].Announce = &hub.Broadcast
		if cfg.Service[sKey].WebHook != nil {
			for whKey := range *cfg.Service[sKey].WebHook {
				(*cfg.Service[sKey].WebHook)[whKey].ServiceID = cfg.Service[sKey].ID
				(*cfg.Service[sKey].WebHook)[whKey].Announce = &hub.Broadcast
			}
		}
	}

	listenAddress := fmt.Sprintf("%s:%s", cfg.Settings.GetWebListenHost(), cfg.Settings.GetWebListenPort())
	jLog.Info("Listening on "+listenAddress+cfg.Settings.GetWebRoutePrefix(), utils.LogFrom{}, true)

	if cfg.Settings.GetWebCertFile() != nil && cfg.Settings.GetWebKeyFile() != nil {
		jLog.Fatal(http.ListenAndServeTLS(listenAddress, *cfg.Settings.GetWebCertFile(), *cfg.Settings.GetWebKeyFile(), router), utils.LogFrom{}, true)
	} else {
		jLog.Fatal(http.ListenAndServe(listenAddress, router), utils.LogFrom{}, true)
	}
}
