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
			name: "non-nil Notifiers",
			command: Commands{
				{"date", "+%m-%d-%Y"},
			},
			shoutrrrNotifiers: shoutrrr.Shoutrrrs{
				"test": shoutrrrtest.Shoutrrr(false, false),
			},
		},
		{
			name:           "non-nil parentInterval",
			parentInterval: new("11m"),
			command: Commands{
				{"date", "+%m-%d-%Y"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svcStatus := status.Status{
				ServiceInfo: serviceinfo.ServiceInfo{
					ID: "TestNewController",
				},
			}

			controller := NewController(
				&svcStatus,
				tc.command,
				tc.shoutrrrNotifiers,
				tc.parentInterval,
			)

			if tc.expectNil {
				if controller != nil {
					t.Fatal("controller mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}
			if controller == nil {
				t.Fatal("controller mismatch\ngot:  nil\nwant: non-nil")
			}

			if testErr := test.AssertSlicesEqualFunc(
				t,
				controller.Command,
				tc.command,
				func(a, b Command) bool { return util.AreSlicesEqual(a, b) },
				"Controller.Command",
			); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "ServiceStatus", Got: controller.ServiceStatus, Want: &svcStatus, Mode: test.CompareSamePointer},
				{Name: "ParentInterval", Got: controller.ParentInterval, Want: tc.parentInterval, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Controller"); testErr != nil {
				t.Fatal(testErr)
			}

			if testErr := test.AssertMapEqual(
				t,
				controller.Notifiers.Shoutrrr,
				tc.shoutrrrNotifiers,
				"Controller.Notifiers.Shoutrrr",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestController_Metrics(t *testing.T) {
	controller, _ := testRunnableController(t.Name())

	hadC := testutil.CollectAndCount(metric.CommandResultTotal)
	hadG := testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	controller.InitMetrics()

	test.AssertMetricCount(
		t, "non-nil: Controller.InitMetrics() CommandResultTotal",
		metric.CommandResultTotal, hadC+2*len(controller.Command),
	)
	test.AssertMetricCount(
		t, "non-nil: Controller.InitMetrics() LatestVersionIsDeployed",
		metric.LatestVersionIsDeployed, hadG,
	)

	controller.DeleteMetrics()
	test.AssertMetricCount(
		t, "non-nil: Controller.DeleteMetrics() CommandResultTotal",
		metric.CommandResultTotal, hadC,
	)
	test.AssertMetricCount(
		t, "non-nil: Controller.DeleteMetrics() LatestVersionIsDeployed",
		metric.LatestVersionIsDeployed, hadG,
	)

	controller = nil
	controller.InitMetrics()
	test.AssertMetricCount(
		t, "nil: Controller.InitMetrics() CommandResultTotal",
		metric.CommandResultTotal, hadC,
	)
	test.AssertMetricCount(
		t, "nil: Controller.InitMetrics() LatestVersionIsDeployed",
		metric.LatestVersionIsDeployed, hadG,
	)
	controller.DeleteMetrics()
	test.AssertMetricCount(
		t, "nil: Controller.DeleteMetrics() CommandResultTotal",
		metric.CommandResultTotal, hadC,
	)
	test.AssertMetricCount(
		t, "nil: Controller.DeleteMetrics() LatestVersionIsDeployed",
		metric.LatestVersionIsDeployed, hadG,
	)
}

func TestController_IsRunnable(t *testing.T) {
	controller, _ := testRunnableController("service_id")
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

			got := controller.IsRunnable(tc.index)

			if got != tc.want {
				t.Errorf(
					"mismatch (NextRunnable=%s)\ngot:  %t\nwant: %t",
					controller.NextRunnable(tc.index), got, tc.want,
				)
			}
		})
	}
}

func TestController_NextRunnable(t *testing.T) {
	controller, nextRunnables := testRunnableController("service_id")
	tests := []struct {
		name       string
		index      int
		outOfRange bool
	}{
		{
			name:  "NextRunnable just passed",
			index: 0,
		},
		{
			name:  "NextRunnable a minute ago",
			index: 1,
		},
		{
			name:  "NextRunnable in a minute",
			index: 2,
		},
		{
			name:       "NextRunnable out of range",
			index:      3,
			outOfRange: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := controller.NextRunnable(tc.index)

			var want time.Time
			if !tc.outOfRange {
				want = nextRunnables[tc.index]
			}
			if !got.Equal(want) {
				t.Fatalf("mismatch\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}

func TestController_SetNextRunnable(t *testing.T) {
	controller, nextRunnables := testRunnableController("service_id")
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
			// Serial: shares controller and resets its NextRunnables.

			for i := range nextRunnables {
				controller.SetNextRunnable(i, nextRunnables[i])
			}

			controller.SetNextRunnable(tc.index, tc.setTo)

			if tc.index < len(controller.Command) {
				got := controller.NextRunnable(tc.index)
				if !got.Equal(tc.setTo) {
					t.Errorf(
						"Controller.NextRunnable() mismatch\ngot:  %s\nwant: %s",
						got, tc.setTo,
					)
				}
			} else {
				// An out-of-range index changes nothing.
				for i := range nextRunnables {
					got := controller.NextRunnable(i)
					if !got.Equal(nextRunnables[i]) {
						t.Errorf(
							"Controller.NextRunnable(%d) mismatch\ngot:  %s\nwant: %s",
							i, got, nextRunnables[i],
						)
					}
				}
			}
		})
	}
}

func TestController_SetExecuting(t *testing.T) {
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
		new("11m"),
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
			name:  "index out of range",
			index: 6,
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
			timeDifferenceMin: 22*time.Minute - time.Second, // 2× the 11m interval.
			timeDifferenceMax: 22*time.Minute + time.Second,
		},
		{
			name:              "command that didn't fail and is currently running",
			index:             4,
			executing:         true,
			timeDifferenceMin: time.Hour + (22*time.Minute - time.Second), // 2× the 11m interval.
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
			// Serial: reads every NextRunnable, which the parallel rows change.

			if tc.index >= len(controller.Command) {
				before := make([]time.Time, len(controller.Command))
				for i := range before {
					before[i] = controller.NextRunnable(i)
				}

				controller.SetExecuting(tc.index, tc.executing)

				for i := range before {
					if got := controller.NextRunnable(i); !got.Equal(before[i]) {
						t.Errorf(
							"Controller.NextRunnable(%d) mismatch\ngot:  %s\nwant: %s",
							i, got, before[i],
						)
					}
				}
				return
			}
			t.Parallel()

			ranAt := time.Now().UTC()
			controller.SetExecuting(tc.index, tc.executing)

			got := controller.NextRunnable(tc.index)
			minTime := ranAt.Add(tc.timeDifferenceMin)
			maxTime := ranAt.Add(tc.timeDifferenceMax)
			if !(minTime.Before(got)) || !(maxTime.After(got)) {
				t.Fatalf(
					"Controller.NextRunnable() mismatch (ran at %s)\n"+
						"got:  %s\nwant: between %s and %s",
					ranAt, got, minTime, maxTime,
				)
			}
		})
	}
}
