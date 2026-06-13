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

//go:build unit

package command

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

func TestNewController(t *testing.T) {
	// GIVEN: a Command.
	tests := []struct {
		name              string
		command           Commands
		shoutrrrNotifiers shoutrrr.Shoutrrrs
		parentInterval    *string
		expectNil         bool
	}{
		{
			name:      "nil Commands",
			expectNil: true,
		},
		{
			name: "non-nil Commands",
			command: Commands{
				{"date", "+%m-%d-%Y"},
			},
		},
		{
			name: "multiple Commands",
			command: Commands{
				{"date", "+%m-%d-%Y"},
				{"true"},
				{"false"},
			},
		},
		{
			name: "nil Notifiers",
			command: Commands{
				{"date", "+%m-%d-%Y"},
			},
		},
		{
			name: "non-nil Notifiers",
			command: Commands{
				{"date", "+%m-%d-%Y"},
			},
			shoutrrrNotifiers: shoutrrr.Shoutrrrs{
				"test": shoutrrrtest.Shoutrrr(false, false),
			},
		},
		{
			name: "nil parentInterval",
			command: Commands{
				{"date", "+%m-%d-%Y"},
			},
		},
		{
			name:           "non-nil parentInterval",
			parentInterval: test.Ptr("11m"),
			command: Commands{
				{"date", "+%m-%d-%Y"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Status.
			svcStatus := status.Status{
				ServiceInfo: serviceinfo.ServiceInfo{
					ID: "TestNewController",
				},
			}

			// WHEN: a controller is created with NewController.
			controller := NewController(
				&svcStatus,
				tc.command,
				tc.shoutrrrNotifiers,
				tc.parentInterval,
			)

			prefix := fmt.Sprintf("%s\nNewController()", packageName)

			// THEN: the result is expected.
			if controller == nil {
				if !tc.expectNil {
					t.Fatalf(
						"%s result mismatch\ngot:  nil\nwant: non-nil",
						prefix,
					)
				}
				return
			}
			// 	command:
			if err := test.AssertSlicesEqualFunc(
				t,
				controller.Command,
				tc.command,
				func(a, b Command) bool { return util.AreSlicesEqual(a, b) },
				prefix,
				"Command",
			); err != nil {
				t.Fatal(err)
			}
			// Pointers:
			fieldTests := []test.FieldAssertion{
				{Name: "ServiceStatus", Got: controller.ServiceStatus, Want: &svcStatus, Mode: test.CompareSamePointer},
				{Name: "ParentInterval", Got: controller.ParentInterval, Want: tc.parentInterval, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Controller"); err != nil {
				t.Fatal(err)
			}
			// Notifiers.
			if err := test.AssertMapEqual(
				t,
				controller.Notifiers.Shoutrrr,
				tc.shoutrrrNotifiers,
				prefix,
				"Notifier",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestController_Metrics(t *testing.T) {
	// GIVEN: a Controller with multiple Commands.
	controller := NewController(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "TestController_Metrics",
			},
		},
		Commands{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"},
		},
		nil,
		test.Ptr("11m"),
	)
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute),
	}
	for i := range nextRunnables {
		controller.SetNextRunnable(i, nextRunnables[i])
	}

	// WHEN: the Prometheus metrics are initialised with initMetrics.
	hadC := testutil.CollectAndCount(metric.CommandResultTotal)
	hadG := testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	controller.InitMetrics()

	prefix := fmt.Sprintf("%s\nInitMetrics()", packageName)

	// THEN: it can be collected.
	// 	Counters:
	gotC := testutil.CollectAndCount(metric.CommandResultTotal)
	wantC := 2 * len(controller.Command)
	if delta := gotC - hadC; delta != wantC {
		t.Errorf(
			"%s Counter metrics mismatch\ngot:  %d\nwant:  %d",
			prefix, delta, wantC,
		)
	}
	// 	Gauges:
	gotG := testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	wantG := 0
	if delta := gotG - hadG; delta != wantG {
		t.Errorf(
			"%s Gauge metrics mismatch\ngot:	%d\nwant:  %d",
			prefix, delta, wantG,
		)
	}

	prefix = fmt.Sprintf("%s\nDeleteMetrics()", packageName)

	// AND: it can be deleted.
	// 	Counters:
	controller.DeleteMetrics()
	gotC = testutil.CollectAndCount(metric.CommandResultTotal)
	if gotC != hadC {
		t.Errorf(
			"%s\nCounter metrics mismatch\ngot:	%d\nwant: %d",
			prefix, gotC, hadC,
		)
	}
	// 	Gauges:
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf(
			"%s Gauge metrics mismatch\ngot:  %d\nwant: %d",
			prefix, gotG, hadG,
		)
	}

	prefix = fmt.Sprintf("%s\nInitMetrics()", packageName)

	// AND: a nil Controller doesn't panic.
	controller = nil
	// InitMetrics:
	controller.InitMetrics()
	// 	Counter:
	gotC = testutil.CollectAndCount(metric.CommandResultTotal)
	if gotC != hadC {
		t.Errorf(
			"%s on nil Controller shouldn't have changed the Counter metrics\ngot:  %d\nwant: %d",
			prefix, gotC, hadC,
		)
	}
	// 	Gauges:
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf(
			"%s on nil Controller shouldn't have changed the Gauge metrics\ngot:  %d\nwant: %d",
			prefix, gotG, hadG,
		)
	}

	prefix = fmt.Sprintf("%s\nDeleteMetrics()", packageName)

	// DeleteMetrics:
	controller.DeleteMetrics()
	// 	Counters:
	gotC = testutil.CollectAndCount(metric.CommandResultTotal)
	if gotC != hadC {
		t.Errorf(
			"%s on nil Controller shouldn't have changed the Counter metrics\ngot:  %d\nwant: %d",
			prefix, gotC, hadC,
		)
	}
	// 	Gauges:
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf(
			"%s on nil Controller shouldn't have changed the Gauge metrics\ngot:  %d\nwant: %d",
			prefix, gotG, hadG,
		)
	}
}

func TestController_IsRunnable(t *testing.T) {
	// GIVEN: a Controller with various Commands.
	controller := NewController(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id",
			},
		},
		Commands{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"},
		},
		nil,
		test.Ptr("11m"),
	)
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute),
	}
	for i := range nextRunnables {
		controller.SetNextRunnable(i, nextRunnables[i])
	}
	tests := []struct {
		name  string
		index int
		want  bool
	}{
		{
			name:  "NextRunnable just passed",
			index: 0,
			want:  true,
		},
		{
			name:  "NextRunnable a minute ago",
			index: 1,
			want:  true,
		},
		{
			name:  "NextRunnable in a minute",
			index: 2,
			want:  false,
		},
		{
			name:  "NextRunnable out of range",
			index: 3,
			want:  false,
		},
	}

	time.Sleep(5 * time.Millisecond)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsRunnable is called.
			got := controller.IsRunnable(tc.index)

			// THEN: the result is expected.
			if got != tc.want {
				t.Errorf(
					"%s\nController IsRunnable(%d) mismatch (%+v)\ngot:  %t\nwant: %t",
					packageName, tc.index, controller.nextRunnable,
					got, tc.want,
				)
			}
		})
	}
}

func TestController_NextRunnable(t *testing.T) {
	// GIVEN: a Controller with various Commands.
	controller := NewController(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id",
			},
		},
		Commands{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"},
		},
		nil,
		test.Ptr("11m"),
	)
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute),
	}
	for i := range nextRunnables {
		controller.SetNextRunnable(i, nextRunnables[i])
	}
	tests := []struct {
		name       string
		index      int
		setTo      time.Time
		want       bool
		outOfRange bool
	}{
		{
			name:  "NextRunnable just passed",
			index: 0,
			want:  true,
		},
		{
			name:  "NextRunnable a minute ago",
			index: 1,
			want:  true,
		},
		{
			name:  "NextRunnable in a minute",
			index: 2,
			want:  false,
		},
		{
			name:       "NextRunnable out of range",
			index:      3,
			outOfRange: true,
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: NextRunnable is called.
			got := controller.NextRunnable(tc.index)

			// THEN: the result is expected.
			if tc.outOfRange {
				var defaultTime time.Time
				// out-of-range index should return the default time.
				if got != defaultTime {
					t.Fatalf(
						"%s\nController.NextRunnable(%d) out of range index (length=%d)\ngot:  %s\nwant:  %s",
						packageName, tc.index, len(controller.Command),
						got, defaultTime,
					)
				}
			} else if want := controller.nextRunnable[tc.index]; got != want {
				t.Fatalf(
					"%s\nController.NextRunnable(%d) mismatch\ngot:  %s\nwant: %s",
					packageName, tc.index,
					got, want,
				)
			}
		})
	}
}

func TestController_SetNextRunnable(t *testing.T) {
	// GIVEN: a Controller with various Commands.
	controller := NewController(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id",
			},
		},
		Commands{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"},
		},
		nil,
		test.Ptr("11m"),
	)
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute),
	}

	tests := []struct {
		name  string
		index int
		setTo time.Time
	}{
		{
			name:  "valid index 0",
			index: 0,
			setTo: time.Now().UTC().Add(10 * time.Minute),
		},
		{
			name:  "valid index 1",
			index: 1,
			setTo: time.Now().UTC().Add(20 * time.Minute),
		},
		{
			name:  "valid index 2",
			index: 2,
			setTo: time.Now().UTC().Add(30 * time.Minute),
		},
		{
			name:  "index out of range",
			index: 3,
			setTo: time.Now().UTC().Add(40 * time.Minute),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// Reset nextRunnables.
			for i := range nextRunnables {
				controller.SetNextRunnable(i, nextRunnables[i])
			}

			// WHEN: SetNextRunnable is called.
			controller.SetNextRunnable(tc.index, tc.setTo)

			prefix := fmt.Sprintf(
				"%s\nController.SetNextRunnable(index=%d, value=%q)",
				packageName, tc.index, tc.setTo,
			)

			// THEN: the NextRunnable is changed if the index is in range.
			if tc.index < len(controller.Command) {
				got := controller.NextRunnable(tc.index)
				if !got.Equal(tc.setTo) {
					t.Errorf(
						"%s NextRunnable[%d] not set correctly\ngot:  %s\nwant: %s",
						prefix, tc.index,
						got, tc.setTo,
					)
				}
			} else {
				// Ensure an out of range index does not panic and does not change anything.
				for i := range nextRunnables {
					got := controller.NextRunnable(i)
					if !got.Equal(nextRunnables[i]) {
						t.Errorf(
							"%s index out of range should not change nextRunnable (len=%d)\ngot:  %s\nwant: %s",
							packageName, len(controller.Command),
							got, nextRunnables[i],
						)
					}
				}
			}
		})
	}
}

func TestController_SetExecuting(t *testing.T) {
	// GIVEN: a Controller with various Commands.
	controller := NewController(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id",
			},
		},
		Commands{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"},
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"},
		},
		nil,
		test.Ptr("11m"),
	)
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	controller.Failed.Set(4, false)
	controller.Failed.Set(5, true)
	tests := []struct {
		name                                 string
		index                                int
		executing                            bool
		timeDifferenceMin, timeDifferenceMax time.Duration
	}{
		{
			name:              "index out of range",
			index:             6,
			timeDifferenceMin: -time.Second,
			timeDifferenceMax: time.Second,
		},
		{
			name:              "command that hasn't been run and isn't currently running",
			index:             0,
			timeDifferenceMin: 14 * time.Second,
			timeDifferenceMax: 16 * time.Second,
		},
		{
			name:              "command that hasn't been run and is currently running",
			index:             3,
			executing:         true,
			timeDifferenceMin: time.Hour + 14*time.Second,
			timeDifferenceMax: time.Hour + 16*time.Second,
		},
		{
			name:              "command that didn't fail and isn't currently running",
			index:             1,
			timeDifferenceMin: 22*time.Minute - time.Second,
			timeDifferenceMax: 22*time.Minute + time.Second,
		},
		{
			name:              "command that didn't fail and is currently running",
			index:             4,
			executing:         true,
			timeDifferenceMin: time.Hour + (22*time.Minute - time.Second),
			timeDifferenceMax: time.Hour + (22*time.Minute + time.Second),
		},
		{
			name:              "command that did fail and isn't currently running",
			index:             2,
			timeDifferenceMin: 14 * time.Second,
			timeDifferenceMax: 16 * time.Second,
		},
		{
			name:              "command that did fail and is currently running",
			index:             5,
			executing:         true,
			timeDifferenceMin: time.Hour + 14*time.Second,
			timeDifferenceMax: time.Hour + 16*time.Second,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: SetNextRunnable is called.
			ranAt := time.Now().UTC()
			controller.SetExecuting(tc.index, tc.executing)

			// THEN: the result is expected.
			got := ranAt
			if tc.index < len(controller.Command) {
				got = controller.NextRunnable(tc.index)
			}
			minTime := ranAt.Add(tc.timeDifferenceMin)
			maxTime := ranAt.Add(tc.timeDifferenceMax)
			if !(minTime.Before(got)) || !(maxTime.After(got)) {
				t.Fatalf(
					"%s\n[%d]NextRunnable not set correctly\nran at\n%s\ngot:\n%s\nwant between:\n%s and\n%s",
					packageName, tc.index,
					ranAt,
					got,
					minTime, maxTime,
				)
			}
		})
	}
}
