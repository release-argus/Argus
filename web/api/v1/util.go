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
	"net/url"

	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

var (
	jLog *util.JLog
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	// Only if unset (avoid RACE condition).
	if jLog == nil {
		jLog = log
	}
}

// getParam from a URL query string.
func getParam(queryParams *url.Values, param string) *string {
	if !queryParams.Has(param) {
		return nil
	}

	val := queryParams.Get(param)
	return &val
}

// announceDelete broadcasts a DELETE message to all WebSocket clients.
func (api *API) announceDelete(serviceID string) {
	payloadData, _ := json.Marshal(apitype.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "DELETE",
		SubType: serviceID})
	*api.Config.HardDefaults.Service.Status.AnnounceChannel <- payloadData
}

// AnnounceEdit broadcasts an EDIT message to all WebSocket clients
// if the displayed data changes.
func (api *API) announceEdit(oldData *apitype.ServiceSummary, newData *apitype.ServiceSummary) {
	serviceChanged := ""
	if oldData != nil {
		serviceChanged = oldData.ID
		newData.RemoveUnchanged(oldData)
	}

	payloadData, _ := json.Marshal(apitype.WebSocketMessage{
		Page:        "APPROVALS",
		Type:        "EDIT",
		SubType:     serviceChanged,
		ServiceData: newData})

	// Announce all edits to refresh caches.
	*api.Config.HardDefaults.Service.Status.AnnounceChannel <- payloadData
}

// ConstantTimeCompare returns whether the slices x and y have equal contents.
// The time taken depends on the length of the slices,
// and remains independent of the contents.
func ConstantTimeCompare(x, y [32]byte) bool {
	var result byte

	for i := 0; i < len(x); i++ {
		result |= x[i] ^ y[i]
	}

	return result == 0
}
