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

//go:build integration

package webhook

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

func TestWebhooks_Metrics(t *testing.T) {
	tests := []struct {
		name     string
		webhooks *Webhooks
	}{
		{
			name:     "nil",
			webhooks: nil,
		},
		{
			name:     "empty",
			webhooks: &Webhooks{},
		},
		{
			name: "with one",
			webhooks: &Webhooks{
				"foo": &Webhook{
					Main: &Defaults{},
				},
			},
		},
		{
			name: "without Main",
			webhooks: &Webhooks{
				"foo": &Webhook{},
			},
		},
		{
			name: "multiple",
			webhooks: &Webhooks{
				"bish": &Webhook{
					Main: &Defaults{},
				},
				"bash": &Webhook{
					Main: &Defaults{},
				},
				"bosh": &Webhook{
					Main: &Defaults{},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares metric.WebhookResultTotal.

			if tc.webhooks != nil {
				for name, s := range *tc.webhooks {
					s.ID = name
					s.ServiceStatus = &status.Status{}
					s.ServiceStatus.ServiceInfo.ID = name + "-service"
				}
			}

			had := testutil.CollectAndCount(metric.WebhookResultTotal)
			tc.webhooks.InitMetrics()

			want := had
			if tc.webhooks != nil {
				want += 2 * len(*tc.webhooks)
			}
			test.AssertMetricCount(
				t, "Webhooks.InitMetrics() WebhookResultTotal",
				metric.WebhookResultTotal, want,
			)

			tc.webhooks.DeleteMetrics()
			test.AssertMetricCount(
				t, "Webhooks.DeleteMetrics() WebhookResultTotal",
				metric.WebhookResultTotal, had,
			)
		})
	}
}

func TestWebhook_Metrics(t *testing.T) {
	tests := []struct {
		name  string
		isNil bool
	}{
		{
			name:  "nil",
			isNil: true,
		},
		{
			name:  "non-nil",
			isNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares metric.WebhookResultTotal.

			webhook := testWebhook(true, false)
			webhook.ID = tc.name + "TestInitMetrics"
			webhook.ServiceStatus.ServiceInfo.ID = webhook.ID + "_SVC"
			if tc.isNil {
				webhook = nil
			}

			hadC := testutil.CollectAndCount(metric.WebhookResultTotal)
			webhook.initMetrics()

			wantC := hadC + 2
			if tc.isNil {
				wantC = hadC
			}
			test.AssertMetricCount(
				t, "Webhook.initMetrics() WebhookResultTotal",
				metric.WebhookResultTotal, wantC,
			)

			webhook.deleteMetrics()
			test.AssertMetricCount(
				t, "Webhook.deleteMetrics() WebhookResultTotal",
				metric.WebhookResultTotal, hadC,
			)
		})
	}
}
