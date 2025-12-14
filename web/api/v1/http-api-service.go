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
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	logutil "github.com/release-argus/Argus/util/log"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// ServiceOrderAPI is the API response for the service order.
type ServiceOrderAPI struct {
	Order []string `json:"order"`
}

func (api *API) httpServiceOrderGet(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceOrderGet", Secondary: getIP(r)}

	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	api.writeJSON(w, ServiceOrderAPI{Order: api.Config.Order}, logFrom)
}

func (api *API) httpServiceOrderSet(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceOrderSet", Secondary: getIP(r)}

	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()

	currentOrder := api.Config.Order

	// Read the payload.
	payload := http.MaxBytesReader(w, r.Body, int64(512+(128*len(currentOrder))))
	defer payload.Close()
	body, err := io.ReadAll(payload)
	if err != nil {
		failRequest(&w,
			err.Error(),
			http.StatusBadRequest)
		return
	}
	// Unmarshal the new order from the payload.
	var newOrder ServiceOrderAPI
	if err := json.Unmarshal(body, &newOrder); err != nil {
		failRequest(&w,
			"Invalid JSON - "+err.Error(),
			http.StatusBadRequest)
		return
	}

	// Trim unknown services.
	trimmedOrder := make([]string, 0, len(newOrder.Order))
	for _, svc := range newOrder.Order {
		if api.Config.Service[svc] != nil {
			trimmedOrder = append(trimmedOrder, svc)
		}
	}

	// Set the new order.
	api.Config.Order = trimmedOrder
	api.writeJSON(w, apitype.Response{
		Message: "order updated"},
		logFrom)

	// Announce to the WebSocket.
	api.announceOrder()
	// Trigger save.
	api.Config.HardDefaults.Service.Status.SaveChannel <- true
}

func (api *API) httpServiceSummary(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceSummary", Secondary: getIP(r)}
	targetService, _ := url.QueryUnescape(mux.Vars(r)["service_id"])

	// Check Service still exists in this ordering.
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	svc := api.Config.Service[targetService]
	if svc == nil {
		err := fmt.Sprintf("service %q not found", targetService)
		logutil.Log.Error(err, logFrom, true)
		failRequest(&w,
			err,
			http.StatusNotFound)
		return
	}

	// Get ServiceSummary.
	summary := svc.Summary()

	api.writeJSON(w, summary, logFrom)
}
