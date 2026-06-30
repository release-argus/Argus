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
	"net/http"
	"os"
	"runtime"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpWebSocketToken issues a short-lived, single-use token for authenticating
// the "/ws" WebSocket handshake.
//
// Safari/WebKit doesn't forward cached HTTP Basic Auth credentials on
// WebSocket handshake requests, so when Basic Auth is configured, clients
// fetch a token from this (Basic Auth protected) endpoint and pass it as a
// "token" query parameter when connecting to "/ws".
func (api *API) httpWebSocketToken(w http.ResponseWriter, r *http.Request) {
	if api.wsTokens == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	logFrom := logx.LogFrom{Primary: "httpWebSocketToken", Secondary: getIP(r)}

	api.writeJSON(
		w,
		apitype.WebSocketTokenAPI{
			Token: api.wsTokens.New(),
		},
		logFrom,
	)
}

// httpRuntimeInfo returns runtime info about the server.
func (api *API) httpRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpBuildInfo", Secondary: getIP(r)}

	// Create and send status page data.
	msg := apitype.RuntimeInfo{
		StartTime:      util.StartTime,
		CWD:            util.CWD,
		GoRoutineCount: runtime.NumGoroutine(),
		GOMAXPROCS:     runtime.GOMAXPROCS(0),
		GoGC:           os.Getenv("GOGC"),
		GoDebug:        os.Getenv("GODEBUG"),
	}

	api.writeJSON(w, msg, logFrom)
}
