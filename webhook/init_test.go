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

//go:build unit

package webhook

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

func TestSlice_Metrics(t *testing.T) {
	// GIVEN a WebHooks.
	tests := map[string]struct {
		slice *WebHooks
	}{
		"nil": {
			slice: nil},
		"empty": {
			slice: &WebHooks{}},
		"with one": {
			slice: &WebHooks{
				"foo": &WebHook{
					Main: &Defaults{}}}},
		"no Main, no metrics": {
			slice: &WebHooks{
				"foo": &WebHook{}}},
		"multiple": {
			slice: &WebHooks{
				"bish": &WebHook{
					Main: &Defaults{}},
				"bash": &WebHook{
					Main: &Defaults{}},
				"bosh": &WebHook{
					Main: &Defaults{}}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			if tc.slice != nil {
				for name, s := range *tc.slice {
					s.ID = name
					s.ServiceStatus = &status.Status{}
					s.ServiceStatus.ServiceInfo.ID = name + "-service"
				}
			}

			// WHEN the Prometheus metrics are initialised with initMetrics.
			had := testutil.CollectAndCount(metric.WebHookResultTotal)
			tc.slice.InitMetrics()

			// THEN it can be counted.
			got := testutil.CollectAndCount(metric.WebHookResultTotal)
			want := had
			if tc.slice != nil {
				want += 2 * len(*tc.slice)
			}
			if got != want {
				t.Errorf("%s\nmetric count mismatch after InitMetrics()\nwant: %d\ngot:  %d",
					name, want, got)
			}

			// AND the metrics can be deleted.
			tc.slice.DeleteMetrics()
			got = testutil.CollectAndCount(metric.WebHookResultTotal)
			if got != had {
				t.Errorf("%s\nmetric count mismatch after DeleteMetrics()\nwant: %d\ngot:  %d",
					packageName, had, got)
			}
		})
	}
}

func TestWebHook_Metrics(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		isNil bool
	}{
		"nil":     {isNil: true},
		"non-nil": {isNil: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			webhook := testWebHook(true, false, false)
			webhook.ID = name + "TestInitMetrics"
			webhook.ServiceStatus.ServiceInfo.ID = name + "TestInitMetrics"
			if tc.isNil {
				webhook = nil
			}

			// WHEN the Prometheus metrics are initialised with initMetrics.
			hadC := testutil.CollectAndCount(metric.WebHookResultTotal)
			webhook.initMetrics()

			// THEN it can be collected.
			// counters:
			gotC := testutil.CollectAndCount(metric.WebHookResultTotal)
			wantC := 2
			if tc.isNil {
				wantC = 0
			}
			if (gotC - hadC) != wantC {
				t.Errorf("%s\nmetric count mismatch after initMetrics()\nwant: %d\ngot:  %d",
					packageName, wantC, gotC-hadC)
			}

			// AND it can be deleted.
			webhook.deleteMetrics()
			gotC = testutil.CollectAndCount(metric.WebHookResultTotal)
			if gotC != hadC {
				t.Errorf("%s\nmetric count mismatch after deleteMetrics()\nwant: %d\ngot:  %d",
					packageName, hadC, gotC)
			}
		})
	}
}

func TestWebHook_Init(t *testing.T) {
	// GIVEN a WebHook and vars for the Init.
	webhook := testWebHook(true, false, false)
	var notifiers shoutrrr.Shoutrrrs
	var main Defaults
	var defaults, hardDefaults Defaults
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 1,
		"TestInit", "", "",
		&dashboard.Options{
			WebURL: "https://example.com"})

	// WHEN Init is called on it.
	webhook.Init(
		&svcStatus,
		&main, &defaults, &hardDefaults,
		&notifiers,
		webhook.ParentInterval)
	webhook.ID = "TestInit"

	// THEN pointers to those vars are handed out to the WebHook:
	// 	Main:
	if webhook.Main != &main {
		t.Errorf("%s\nMain was not handed to the WebHook correctly\nwant: %v\ngot:  %v",
			packageName, &main, webhook.Main)
	}
	// 	Defaults:
	if webhook.Defaults != &defaults {
		t.Errorf("%s\nDefaults were not handed to the WebHook correctly\nwant: %v\ngot:  %v",
			packageName, &defaults, webhook.Defaults)
	}
	// 	HardDefaults:
	if webhook.HardDefaults != &hardDefaults {
		t.Errorf("%s\nHardDefaults were not handed to the WebHook correctly\nwant: %v\ngot:  %v",
			packageName, &hardDefaults, webhook.HardDefaults)
	}
	// 	Status:
	if webhook.ServiceStatus != &svcStatus {
		t.Errorf("%s\nStatus was not handed to the WebHook correctly\nwant: %v\ngot:  %v",
			packageName, &svcStatus, webhook.ServiceStatus)
	}
	// 	Options:
	if webhook.Notifiers.Shoutrrr != &notifiers {
		t.Errorf("%s\nNotifiers were not handed to the WebHook correctly\nwant: %v\ngot:  %v",
			packageName, &notifiers, webhook.Notifiers.Shoutrrr)
	}
}

func TestSlice_Init(t *testing.T) {
	// GIVEN a WebHooks and vars for the Init.
	var notifiers shoutrrr.Shoutrrrs
	tests := map[string]struct {
		slice                  *WebHooks
		nilSlice               bool
		mains                  *WebHooksDefaults
		defaults, hardDefaults *Defaults
	}{
		"nil slice": {
			slice: nil, nilSlice: true,
		},
		"empty slice": {
			slice: &WebHooks{},
		},
		"no mains": {
			slice: &WebHooks{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false)},
		},
		"slice with nil element and matching main": {
			slice: &WebHooks{
				"fail": nil},
			mains: &WebHooksDefaults{
				"fail": testDefaults(false, false)},
		},
		"have matching mains": {
			slice: &WebHooks{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false)},
			mains: &WebHooksDefaults{
				"fail": testDefaults(false, false),
				"pass": testDefaults(true, false),
			},
		},
		"some matching mains": {
			slice: &WebHooks{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false)},
			mains: &WebHooksDefaults{
				"other": testDefaults(false, false),
				"pass":  testDefaults(true, false)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !tc.nilSlice {
				for i := range *tc.slice {
					if (*tc.slice)[i] != nil {
						(*tc.slice)[i].ID = name + i
					}
				}
			}
			serviceStatus := status.Status{}
			serviceStatus.ServiceInfo.ID = name
			mainCount := 0
			if tc.mains != nil {
				mainCount = len(*tc.mains)
			}
			serviceStatus.Init(
				0, 0, mainCount,
				name, "", "",
				&dashboard.Options{})
			parentInterval := "10s"

			// WHEN Init is called on it.
			tc.slice.Init(
				&serviceStatus,
				tc.mains, tc.defaults, tc.hardDefaults,
				&notifiers,
				&parentInterval)

			// THEN pointers to those vars are handed out to the WebHook:
			if tc.nilSlice {
				if tc.slice != nil {
					t.Fatalf("%s\nslice mismatch\nwant: nil\ngot:  %v",
						packageName, *tc.slice)
				}
				return
			}
			for _, webhook := range *tc.slice {
				// 	Main:
				if webhook.Main == nil {
					t.Errorf("%s\nMain of the WebHook was not initialised\ngot: %v",
						packageName, webhook.Main)
				} else if tc.mains != nil && (*tc.mains)[webhook.ID] != nil && webhook.Main != (*tc.mains)[webhook.ID] {
					t.Errorf("%s\nMain were not handed to the WebHook correctly\nwant: %v\ngot:  %v",
						packageName, (*tc.mains)[webhook.ID], webhook.Main)
				}
				// 	Defaults:
				if webhook.Defaults != tc.defaults {
					t.Errorf("%s\nDefaults were not handed to the WebHook correctly\nwant: %v\ngot:  %v",
						packageName, &tc.defaults, webhook.Defaults)
				}
				// 	HardDefaults:
				if webhook.HardDefaults != tc.hardDefaults {
					t.Errorf("%s\nHardDefaults were not handed to the WebHook correctly\nwant: %v\ngot:  %v",
						packageName, &tc.hardDefaults, webhook.HardDefaults)
				}
				// 	Status:
				if webhook.ServiceStatus != &serviceStatus {
					t.Errorf("%s\nStatus was not handed to the WebHook correctly\nwant: %v\ngot:  %v",
						packageName, &serviceStatus, webhook.ServiceStatus)
				}
				// 	Notifiers:
				if webhook.Notifiers.Shoutrrr != &notifiers {
					t.Errorf("%s\nNotifiers were not handed to the WebHook correctly\nwant: %v\ngot:  %v",
						packageName, &notifiers, webhook.Notifiers.Shoutrrr)
				}
			}
		})
	}
}
