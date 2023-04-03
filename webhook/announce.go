// Copyright [2022] [Argus]
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

package webhook

import (
	"encoding/json"

	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

// AnnounceSend of the WebHook to the `w.Announce` channel
// (Broadcast to all WebSocket clients).
func (w *WebHook) AnnounceSend() {
	w.SetExecuting(false, false)
	webhookSummary := make(map[string]*api_type.WebHookSummary)
	webhookSummary[w.ID] = &api_type.WebHookSummary{
		Failed:       w.Failed.Get(w.ID),
		NextRunnable: w.GetNextRunnable(),
	}

	// WebHook pass/fail
	wsPage := "APPROVALS"
	wsType := "WEBHOOK"
	wsSubType := "EVENT"
	payloadData, _ := json.Marshal(api_type.WebSocketMessage{
		Page:    wsPage,
		Type:    wsType,
		SubType: wsSubType,
		ServiceData: &api_type.ServiceSummary{
			ID: util.DefaultIfNil(w.ServiceStatus.ServiceID),
		},
		WebHookData: webhookSummary,
	})

	if w.ServiceStatus.AnnounceChannel != nil {
		*w.ServiceStatus.AnnounceChannel <- payloadData
	}
}
