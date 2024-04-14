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
	"encoding/json"
	"net/http"

	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

// httpFlags returns the values of vars that can be set with flags.
func (api *API) httpFlags(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpFlags", Secondary: getIP(r)}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	msg := api_type.Flags{
		ConfigFile:       api.Config.File,
		LogLevel:         api.Config.Settings.LogLevel(),
		LogTimestamps:    api.Config.Settings.LogTimestamps(),
		DataDatabaseFile: api.Config.Settings.DataDatabaseFile(),
		WebListenHost:    api.Config.Settings.WebListenHost(),
		WebListenPort:    api.Config.Settings.WebListenPort(),
		WebCertFile:      api.Config.Settings.WebCertFile(),
		WebPKeyFile:      api.Config.Settings.WebKeyFile(),
		WebRoutePrefix:   api.Config.Settings.WebRoutePrefix()}

	err := json.NewEncoder(w).Encode(msg)
	api.Log.Error(err, logFrom, err != nil)
}
