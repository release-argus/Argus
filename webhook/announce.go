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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// marshalWebhookPayload serialises WebHook payloads (overridable for tests).
var marshalWebhookPayload = func(v any) ([]byte, error) {
	return decode.Marshal("json", v)
}

// AnnounceSend broadcasts the WebHook's send result to all WebSocket clients.
func (w *WebHook) AnnounceSend() {
	w.SetExecuting(false, false)
	webhookSummary := make(map[string]*apitype.WebHookSummary)
	webhookSummary[w.ID] = &apitype.WebHookSummary{
		Failed:       w.DidFail(),
		NextRunnable: w.NextRunnable(),
	}

	// WebHook pass/fail.
	payloadData, err := marshalWebhookPayload(
		apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "WEBHOOK",
			SubType: "EVENT",
			ServiceData: &apitype.ServiceSummary{
				ID: w.ServiceStatus.ServiceInfo.ID,
			},
			WebHookData: webhookSummary,
		},
	)
	if err != nil {
		logx.Error(err, logx.LogFrom{Primary: w.ID, Secondary: w.ServiceStatus.ServiceInfo.ID}, true)
		return
	}

	w.ServiceStatus.SendAnnounce(payloadData)
}
