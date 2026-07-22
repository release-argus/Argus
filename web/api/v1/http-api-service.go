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
	"fmt"
	"io"
	"net/http"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// ServiceOrderAPI is the API response for the service order.
type ServiceOrderAPI struct {
	Order []string `json:"order"`
}

// httpServiceOrderGet handles GET /api/v1/service/order: returning the current
// service ordering (filtered to the services the caller may read).
//
// Response:
//
//	200 OK: JSON object containing the current ordering.
func (api *API) httpServiceOrderGet(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceOrderGet", Secondary: getIP(r)}

	// Users without global service:read see only the services their scoped
	// grants allow (possibly none).
	var allowed map[string]bool
	if api.auth != nil {
		if authCtx := authContextFrom(r); authCtx != nil {
			allowed = api.allowedServices(authCtx)
		}
	}

	api.Config.OrderMu.RLock()
	defer api.Config.OrderMu.RUnlock()
	order := api.Config.Order
	if allowed != nil {
		filtered := make([]string, 0, len(allowed))
		for _, serviceID := range order {
			if allowed[serviceID] {
				filtered = append(filtered, serviceID)
			}
		}
		order = filtered
	}
	api.writeJSON(w, ServiceOrderAPI{Order: order}, logFrom)
}

// httpServiceOrderSet handles PUT /api/v1/service/order: setting the ordering of services.
//
// Body:
//
//	JSON object containing the new order.
//
// Response:
//
//	200 OK: Success message.
//	400 Bad Request: Error message.
func (api *API) httpServiceOrderSet(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceOrderSet", Secondary: getIP(r)}

	api.Config.OrderMu.RLock()
	defer api.Config.OrderMu.RUnlock()

	currentOrder := api.Config.Order

	// Read the payload.
	payload := http.MaxBytesReader(w, r.Body, int64(512+(128*len(currentOrder))))
	defer payload.Close()
	body, err := io.ReadAll(payload)
	if err != nil {
		failRequest(&w, err, http.StatusBadRequest)
		return
	}
	// Unmarshal the new order from the payload.
	var newOrder ServiceOrderAPI
	if err := decode.Unmarshal("json", body, &newOrder); err != nil {
		failRequest(
			&w,
			fmt.Errorf("invalid JSON: %w", err),
			http.StatusBadRequest,
		)
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
	api.writeJSON(
		w,
		apitype.Response{
			Message: "order updated",
		},
		logFrom,
	)

	// Announce to the WebSocket.
	api.announceOrder()
	// Trigger save.
	api.Config.HardDefaults.Service.Status.SaveChannel <- true
}

// httpServiceSummary handles GET /api/v1/service/summary: returning the
// [apitype.ServiceSummary] for the given service.
//
// Query parameters:
//
//	service_id: The ID of the Service to get details for.
//
// Response:
//
//	200 OK: JSON object containing the service details.
func (api *API) httpServiceSummary(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpServiceSummary", Secondary: getIP(r)}
	serviceID, ok := requireQueryParam(w, r, "service_id")
	if !ok {
		return
	}

	// Check Service still exists in this ordering.
	api.Config.OrderMu.RLock()
	defer api.Config.OrderMu.RUnlock()
	svc := api.Config.Service[serviceID]
	if svc == nil {
		err := fmt.Errorf("service %q not found", serviceID)
		logx.Error(err, logFrom, true)
		failRequest(&w, err, http.StatusNotFound)
		return
	}

	// Get ServiceSummary.
	summary := svc.Summary()

	api.writeJSON(w, summary, logFrom)
}
