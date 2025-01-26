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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"encoding/json"

	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// AnnounceSend of the WebHook to the `w.Announce` channel.
// (Broadcast to all WebSocket clients).
func (w *WebHook) AnnounceSend() {
	w.SetExecuting(false, false)
	webhookSummary := make(map[string]*apitype.WebHookSummary)
	webhookSummary[w.ID] = &apitype.WebHookSummary{
		Failed:       w.Failed.Get(w.ID),
		NextRunnable: w.NextRunnable()}

	// WebHook pass/fail.
	payloadData, _ := json.Marshal(apitype.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "WEBHOOK",
		SubType: "EVENT",
		ServiceData: &apitype.ServiceSummary{
			ID: util.DereferenceOrDefault(w.ServiceStatus.ServiceID)},
		WebHookData: webhookSummary})

	w.ServiceStatus.SendAnnounce(&payloadData)
}
