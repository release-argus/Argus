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

// Package web provides the web server for Argus.
package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	v1 "github.com/release-argus/Argus/web/api/v1"
)

// Run starts the web server.
// authDeps is non-nil only when auth is enabled.
func Run(ctx context.Context, cfg *config.Config, authDeps *v1.AuthDeps) error {
	router := newWebUI(cfg, authDeps)

	listenAddress := fmt.Sprintf(
		"%s:%s",
		cfg.Settings.WebListenHost(), cfg.Settings.WebListenPort(),
	)
	logx.Info(
		"Listening on "+listenAddress+cfg.Settings.WebRoutePrefix(),
		logx.LogFrom{},
		true,
	)

	srv := &http.Server{
		Addr:         listenAddress,
		Handler:      router,
		ReadTimeout:  10 * time.Second, // Max time to read request headers and body.
		WriteTimeout: 10 * time.Second, // Max time to write response.
	}

	errChan := make(chan error, 1)
	go func() {
		if cfg.Settings.WebCertFile() != "" && cfg.Settings.WebKeyFile() != "" {
			errChan <- srv.ListenAndServeTLS(
				cfg.Settings.WebCertFile(),
				cfg.Settings.WebKeyFile(),
			)
		} else {
			errChan <- srv.ListenAndServe()
		}
	}()

	select {
	// Graceful shutdown.
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx) //nolint:wrapcheck

	// Error.
	case err := <-errChan:
		return err
	}
}

// newRouter serves Prometheus metrics, WebSocket, and Node.js frontend at RoutePrefix.
func newRouter(cfg *config.Config, hub *v1.Hub, authDeps *v1.AuthDeps) *mux.Router {
	api, wsRoute := v1.NewAPI(cfg)
	if authDeps != nil {
		api.EnableAuth(authDeps)
	}

	api.Router.Handle("/metrics", api.GuardMetrics(promhttp.Handler()))

	api.SetupWebSocket(hub, wsRoute)
	api.SetupRoutesAPI()
	api.SetupRoutesNodeJS()

	return api.BaseRouter
}

// newWebUI sets up everything web-related for Argus.
func newWebUI(cfg *config.Config, authDeps *v1.AuthDeps) *mux.Router {
	hub := v1.NewHub()
	go hub.Run()
	router := newRouter(cfg, hub, authDeps)

	cfg.HardDefaults.Service.Status.AnnounceChannel = hub.Broadcast
	for _, svc := range cfg.Service {
		svc.Status.SetAnnounceChannel(hub.Broadcast)
	}

	return router
}
