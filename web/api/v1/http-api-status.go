// Copyright [2024] [Argus]
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
	"net/http"
	"os"
	"runtime"

	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpRuntimeInfo returns runtime info about the server.
func (api *API) httpRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpBuildInfo", Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	// Create and send status page data.
	msg := apitype.RuntimeInfo{
		StartTime:      util.StartTime,
		CWD:            util.CWD,
		GoRoutineCount: runtime.NumGoroutine(),
		GOMAXPROCS:     runtime.GOMAXPROCS(0),
		GoGC:           os.Getenv("GOGC"),
		GoDebug:        os.Getenv("GODEBUG")}

	err := json.NewEncoder(w).Encode(msg)
	jLog.Error(err, logFrom, err != nil)
}
