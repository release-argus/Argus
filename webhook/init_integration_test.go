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
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

func TestWebHooks_Metrics(t *testing.T) {
	// GIVEN: a WebHooks.
	tests := []struct {
		name     string
		webhooks *WebHooks
	}{
		{
			name:     "nil",
			webhooks: nil,
		},
		{
			name:     "empty",
			webhooks: &WebHooks{},
		},
		{
			name: "with one",
			webhooks: &WebHooks{
				"foo": &WebHook{
					Main: &Defaults{},
				},
			},
		},
		{
			name: "no Main, no metrics",
			webhooks: &WebHooks{
				"foo": &WebHook{},
			},
		},
		{
			name: "multiple",
			webhooks: &WebHooks{
				"bish": &WebHook{
					Main: &Defaults{},
				},
				"bash": &WebHook{
					Main: &Defaults{},
				},
				"bosh": &WebHook{
					Main: &Defaults{},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			if tc.webhooks != nil {
				for name, s := range *tc.webhooks {
					s.ID = name
					s.ServiceStatus = &status.Status{}
					s.ServiceStatus.ServiceInfo.ID = name + "-service"
				}
			}

			// WHEN: the Prometheus metrics are initialised with initMetrics.
			had := testutil.CollectAndCount(metric.WebHookResultTotal)
			tc.webhooks.InitMetrics()

			// THEN: it can be counted.
			got := testutil.CollectAndCount(metric.WebHookResultTotal)
			want := had
			if tc.webhooks != nil {
				want += 2 * len(*tc.webhooks)
			}
			if got != want {
				t.Errorf(
					"%s\nWebHooks.InitMetrics() metric count mismatch\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}

			// AND: the metrics can be deleted.
			tc.webhooks.DeleteMetrics()
			got = testutil.CollectAndCount(metric.WebHookResultTotal)
			if got != had {
				t.Errorf(
					"%s\nWebHooks.DeleteMetrics() metric count mismatch\ngot:  %d\nwant: %d",
					packageName, got, had,
				)
			}
		})
	}
}

func TestWebHook_Metrics(t *testing.T) {
	// GIVEN: a WebHook.
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
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			webhook := testWebHook(true, false, false)
			webhook.ID = tc.name + "TestInitMetrics"
			webhook.ServiceStatus.ServiceInfo.ID = webhook.ID + "_SVC"
			if tc.isNil {
				webhook = nil
			}

			// WHEN: the Prometheus metrics are initialised with initMetrics.
			hadC := testutil.CollectAndCount(metric.WebHookResultTotal)
			webhook.initMetrics()

			// THEN: it can be collected.
			// counters:
			gotC := testutil.CollectAndCount(metric.WebHookResultTotal)
			wantC := 2
			if tc.isNil {
				wantC = 0
			}
			if resC := gotC - hadC; resC != wantC {
				t.Errorf(
					"%s\nWebHook.initMetrics() metric count mismatch\ngot:  %d\nwant: %d",
					packageName, resC, wantC,
				)
			}

			// AND: it can be deleted.
			webhook.deleteMetrics()
			gotC = testutil.CollectAndCount(metric.WebHookResultTotal)
			if gotC != hadC {
				t.Errorf(
					"%s\nWebHook.deleteMetrics() metric count mismatch\ngot:  %d\nwant: %d",
					packageName, gotC, hadC,
				)
			}
		})
	}
}
