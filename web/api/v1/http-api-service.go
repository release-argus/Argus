// Copyright [2025] [Argus]
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
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/release-argus/Argus/util"
)

// ServiceOrderAPI is the API response for the service order.
type ServiceOrderAPI struct {
	Order []string `json:"order"`
}

func (api *API) httpServiceOrder(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpServiceOrder", Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	err := json.NewEncoder(w).Encode(ServiceOrderAPI{Order: api.Config.Order})
	jLog.Error(err, logFrom, err != nil)
}

func (api *API) httpServiceSummary(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpServiceSummary", Secondary: getIP(r)}
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])
	jLog.Verbose(targetService, logFrom, true)

	// Check Service still exists in this ordering.
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	service := api.Config.Service[targetService]
	if service == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		jLog.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Get ServiceSummary.
	summary := service.Summary()

	err := json.NewEncoder(w).Encode(summary)
	jLog.Error(err, logFrom, err != nil)
}
