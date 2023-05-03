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

//go:build unit

package webhook

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	metric "github.com/release-argus/Argus/web/metrics"
)

func TestSlice_Metrics(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice *Slice
	}{
		"nil": {
			slice: nil},
		"empty": {
			slice: &Slice{}},
		"with one": {
			slice: &Slice{
				"foo": &WebHook{
					Main: &WebHookDefaults{}}}},
		"no Main, no metrics": {
			slice: &Slice{
				"foo": &WebHook{}}},
		"multiple": {
			slice: &Slice{
				"bish": &WebHook{
					Main: &WebHookDefaults{}},
				"bash": &WebHook{
					Main: &WebHookDefaults{}},
				"bosh": &WebHook{
					Main: &WebHookDefaults{}}}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

			if tc.slice != nil {
				for name, s := range *tc.slice {
					s.ID = name
					s.ServiceStatus = &svcstatus.Status{ServiceID: stringPtr(name + "-service")}
				}
			}

			// WHEN the Prometheus metrics are initialised with initMetrics
			had := testutil.CollectAndCount(metric.WebHookMetric)
			tc.slice.InitMetrics()

			// THEN it can be counted
			got := testutil.CollectAndCount(metric.WebHookMetric)
			want := had
			if tc.slice != nil {
				want += 2 * len(*tc.slice)
			}
			if got != want {
				t.Errorf("got %d metrics, expecting %d",
					got, want)
			}

			// AND the metrics can be deleted
			tc.slice.DeleteMetrics()
			got = testutil.CollectAndCount(metric.WebHookMetric)
			if got != had {
				t.Errorf("deleted metrics but got %d, expecting %d",
					got, want)
			}
		})
	}
}

func TestWebHook_Metrics(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		isNil bool
	}{
		"nil":     {isNil: true},
		"non-nil": {isNil: false},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			webhook := testWebHook(true, false, false)
			webhook.ID = name + "TestInitMetrics"
			webhook.ServiceStatus.ServiceID = stringPtr(name + "TestInitMetrics")
			if tc.isNil {
				webhook = nil
			}

			// WHEN the Prometheus metrics are initialised with initMetrics
			hadC := testutil.CollectAndCount(metric.WebHookMetric)
			webhook.initMetrics()

			// THEN it can be collected
			// counters
			gotC := testutil.CollectAndCount(metric.WebHookMetric)
			wantC := 2
			if tc.isNil {
				wantC = 0
			}
			if (gotC - hadC) != wantC {
				t.Errorf("%d Counter metrics's were initialised, expecting %d",
					(gotC - hadC), wantC)
			}

			// AND it can be deleted
			webhook.deleteMetrics()
			gotC = testutil.CollectAndCount(metric.WebHookMetric)
			if gotC != hadC {
				t.Errorf("Counter metrics were not deleted, still have %d. expexting %d",
					gotC, hadC)
			}
		})
	}
}

func TestWebHook_Init(t *testing.T) {
	// GIVEN a WebHook and vars for the Init
	webhook := testWebHook(true, false, false)
	var notifiers shoutrrr.Slice
	var main WebHookDefaults
	var defaults WebHookDefaults
	var hardDefaults WebHookDefaults
	status := svcstatus.Status{ServiceID: stringPtr("TestInit")}
	status.Init(
		0, 0, 1,
		stringPtr("TestInit"),
		stringPtr("https://example.com"))

	// WHEN Init is called on it
	webhook.Init(
		&status,
		&main, &defaults, &hardDefaults,
		&notifiers,
		webhook.ParentInterval)
	webhook.ID = "TestInit"

	// THEN pointers to those vars are handed out to the WebHook
	// main
	if webhook.Main != &main {
		t.Errorf("Main was not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&main, webhook.Main)
	}
	// defaults
	if webhook.Defaults != &defaults {
		t.Errorf("Defaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&defaults, webhook.Defaults)
	}
	// hardDefaults
	if webhook.HardDefaults != &hardDefaults {
		t.Errorf("HardDefaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&hardDefaults, webhook.HardDefaults)
	}
	// status
	if webhook.ServiceStatus != &status {
		t.Errorf("Status was not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&status, webhook.ServiceStatus)
	}
	// options
	if webhook.Notifiers.Shoutrrr != &notifiers {
		t.Errorf("Notifiers were not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&notifiers, webhook.Notifiers.Shoutrrr)
	}
}

func TestSlice_Init(t *testing.T) {
	// GIVEN a Slice and vars for the Init
	var notifiers shoutrrr.Slice
	tests := map[string]struct {
		slice        *Slice
		nilSlice     bool
		mains        *SliceDefaults
		defaults     *WebHookDefaults
		hardDefaults *WebHookDefaults
	}{
		"nil slice": {
			slice: nil, nilSlice: true,
		},
		"empty slice": {
			slice: &Slice{},
		},
		"no mains": {
			slice: &Slice{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false)},
		},
		"slice with nil element and matching main": {
			slice: &Slice{
				"fail": nil},
			mains: &SliceDefaults{
				"fail": testWebHookDefaults(false, false, false)},
		},
		"have matching mains": {
			slice: &Slice{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false)},
			mains: &SliceDefaults{
				"fail": testWebHookDefaults(false, false, false),
				"pass": testWebHookDefaults(true, false, false),
			},
		},
		"some matching mains": {
			slice: &Slice{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false)},
			mains: &SliceDefaults{
				"other": testWebHookDefaults(false, false, false),
				"pass":  testWebHookDefaults(true, false, false)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !tc.nilSlice {
				for i := range *tc.slice {
					if (*tc.slice)[i] != nil {
						(*tc.slice)[i].ID = name + i
					}
				}
			}
			serviceStatus := svcstatus.Status{ServiceID: &name}
			mainCount := 0
			if tc.mains != nil {
				mainCount = len(*tc.mains)
			}
			serviceStatus.Init(
				0, 0, mainCount,
				&name,
				nil)
			parentInterval := "10s"

			// WHEN Init is called on it
			tc.slice.Init(
				&serviceStatus,
				tc.mains, tc.defaults, tc.hardDefaults,
				&notifiers,
				&parentInterval)

			// THEN pointers to those vars are handed out to the WebHook
			if tc.nilSlice {
				if tc.slice != nil {
					t.Fatalf("expecting the Slice to be nil, not %v",
						*tc.slice)
				}
				return
			}
			for _, webhook := range *tc.slice {
				// main
				if webhook.Main == nil {
					t.Errorf("Main of the WebHook was not initialised. got: %v",
						webhook.Main)
				} else if tc.mains != nil && (*tc.mains)[webhook.ID] != nil && webhook.Main != (*tc.mains)[webhook.ID] {
					t.Errorf("Main were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						(*tc.mains)[webhook.ID], webhook.Main)
				}
				// defaults
				if webhook.Defaults != tc.defaults {
					t.Errorf("Defaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&tc.defaults, webhook.Defaults)
				}
				// hardDefaults
				if webhook.HardDefaults != tc.hardDefaults {
					t.Errorf("HardDefaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&tc.hardDefaults, webhook.HardDefaults)
				}
				// status
				if webhook.ServiceStatus != &serviceStatus {
					t.Errorf("Status was not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&serviceStatus, webhook.ServiceStatus)
				}
				// notifiers
				if webhook.Notifiers.Shoutrrr != &notifiers {
					t.Errorf("Notifiers were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&notifiers, webhook.Notifiers.Shoutrrr)
				}
			}
		})
	}
}
