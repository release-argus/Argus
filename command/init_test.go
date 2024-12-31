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

//go:build unit

package command

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_test "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	metric "github.com/release-argus/Argus/web/metric"
)

func TestController_SetExecuting(t *testing.T) {
	// GIVEN a Controller with various Commands
	controller := Controller{}
	controller.Init(
		&status.Status{ServiceID: test.StringPtr("service_id")},
		&Slice{
			{"date", "+%m-%d-%Y"}, {"true"}, {"false"},
			{"date", "+%m-%d-%Y"}, {"true"}, {"false"}},
		nil,
		test.StringPtr("11m"))
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	controller.Failed.Set(4, false)
	controller.Failed.Set(5, true)
	tests := map[string]struct {
		index                                int
		executing                            bool
		timeDifferenceMin, timeDifferenceMax time.Duration
	}{
		"index out of range": {
			index:             6,
			timeDifferenceMin: -time.Second,
			timeDifferenceMax: time.Second,
		},
		"command that hasn't been run and isn't currently running": {
			index:             0,
			timeDifferenceMin: 14 * time.Second,
			timeDifferenceMax: 16 * time.Second,
		},
		"command that hasn't been run and is currently running": {
			index:             3,
			executing:         true,
			timeDifferenceMin: time.Hour + 14*time.Second,
			timeDifferenceMax: time.Hour + 16*time.Second,
		},
		"command that didn't fail and isn't currently running": {
			index:             1,
			timeDifferenceMin: 22*time.Minute - time.Second,
			timeDifferenceMax: 22*time.Minute + time.Second,
		},
		"command that didn't fail and is currently running": {
			index:             4,
			executing:         true,
			timeDifferenceMin: time.Hour + (22*time.Minute - time.Second),
			timeDifferenceMax: time.Hour + (22*time.Minute + time.Second),
		},
		"command that did fail and isn't currently running": {
			index:             2,
			timeDifferenceMin: 14 * time.Second,
			timeDifferenceMax: 16 * time.Second,
		},
		"command that did fail and is currently running": {
			index:             5,
			executing:         true,
			timeDifferenceMin: time.Hour + 14*time.Second,
			timeDifferenceMax: time.Hour + 16*time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN SetNextRunnable is called
			ranAt := time.Now().UTC()
			controller.SetExecuting(tc.index, tc.executing)

			// THEN the result is expected
			got := ranAt
			if tc.index < len(*controller.Command) {
				got = (controller.NextRunnable(tc.index))
			}
			minTime := ranAt.Add(tc.timeDifferenceMin)
			maxTime := ranAt.Add(tc.timeDifferenceMax)
			if !(minTime.Before(got)) || !(maxTime.After(got)) {
				t.Fatalf("ran at\n%s\nwant between:\n%s and\n%s\ngot:\n%s",
					ranAt, minTime, maxTime, got)
			}
		})
	}
}

func TestController_IsRunnable(t *testing.T) {
	// GIVEN a Controller with various Commands
	controller := Controller{}
	controller.Init(
		&status.Status{ServiceID: test.StringPtr("service_id")},
		&Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		nil,
		test.StringPtr("11m"))
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute)}
	for i := range nextRunnables {
		controller.SetNextRunnable(i, nextRunnables[i])
	}
	tests := map[string]struct {
		index int
		want  bool
	}{
		"NextRunnable just passed": {
			index: 0, want: true},
		"NextRunnable a minute ago": {
			index: 1, want: true},
		"NextRunnable in a minute": {
			index: 2, want: false},
		"NextRunnable out of range": {
			index: 3, want: false},
	}

	time.Sleep(5 * time.Millisecond)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN IsRunnable is called
			got := controller.IsRunnable(tc.index)

			// THEN the result is expected
			if got != tc.want {
				t.Errorf("want: %t\ngot:\n%t",
					tc.want, got)
			}
		})
	}
}

func TestController_NextRunnable(t *testing.T) {
	// GIVEN a Controller with various Commands
	controller := Controller{}
	controller.Init(
		&status.Status{ServiceID: test.StringPtr("service_id")},
		&Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		nil,
		test.StringPtr("11m"))
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute)}
	for i := range nextRunnables {
		controller.SetNextRunnable(i, nextRunnables[i])
	}
	tests := map[string]struct {
		index      int
		setTo      time.Time
		want       bool
		outOfRange bool
	}{
		"NextRunnable just passed": {
			index: 0, want: true},
		"NextRunnable a minute ago": {
			index: 1, want: true},
		"NextRunnable in a minute": {
			index: 2, want: false},
		"NextRunnable out of range": {
			index: 3, outOfRange: true, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN NextRunnable is called
			got := controller.NextRunnable(tc.index)

			// THEN the result is expected
			if tc.outOfRange {
				var defaultTime time.Time
				if got != defaultTime {
					t.Fatalf("want: %s\ngot:\n%s",
						defaultTime, got)
				}
			} else if got != controller.NextRunnable(tc.index) {
				t.Fatalf("want: %s\ngot:\n%s",
					controller.NextRunnable(tc.index), got)
			}
		})
	}
}

func TestController_SetNextRunnable(t *testing.T) {
	// GIVEN a Controller with various Commands
	controller := Controller{}
	controller.Init(
		&status.Status{ServiceID: test.StringPtr("service_id")},
		&Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		nil,
		test.StringPtr("11m"))
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute)}

	tests := map[string]struct {
		index int
		setTo time.Time
	}{
		"valid index 0": {
			index: 0, setTo: time.Now().UTC().Add(10 * time.Minute)},
		"valid index 1": {
			index: 1, setTo: time.Now().UTC().Add(20 * time.Minute)},
		"valid index 2": {
			index: 2, setTo: time.Now().UTC().Add(30 * time.Minute)},
		"index out of range": {
			index: 3, setTo: time.Now().UTC().Add(40 * time.Minute)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// Reset nextRunnables
			for i := range nextRunnables {
				controller.SetNextRunnable(i, nextRunnables[i])
			}

			// WHEN SetNextRunnable is called
			controller.SetNextRunnable(tc.index, tc.setTo)

			// THEN the NextRunnable is changed if the index is in range
			if tc.index < len(*controller.Command) {
				got := controller.NextRunnable(tc.index)
				if !got.Equal(tc.setTo) {
					t.Errorf("want: %s\ngot:\n%s",
						tc.setTo, got)
				}
			} else {
				// Ensure out of range index does not panic and does not change anything
				for i := range nextRunnables {
					got := controller.NextRunnable(i)
					if !got.Equal(nextRunnables[i]) {
						t.Errorf("index out of range should not change nextRunnable. want: %s\ngot:\n%s",
							nextRunnables[i], got)
					}
				}
			}
		})
	}
}

func TestCommand_String(t *testing.T) {
	// GIVEN a Command
	tests := map[string]struct {
		cmd  *Command
		want string
	}{
		"empty command": {
			cmd:  &Command{},
			want: ""},
		"nil command": {
			want: ""},
		"command with no args": {
			cmd:  &Command{"ls"},
			want: "ls"},
		"command with one arg": {
			cmd:  &Command{"ls", "-lah"},
			want: "ls -lah"},
		"command with multiple args": {
			cmd:  &Command{"ls", "-lah", "/root", "/tmp"},
			want: "ls -lah /root /tmp"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the command is stringified with String()
			got := tc.cmd.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("want: %q\ngot:\n%q",
					tc.want, got)
			}
		})
	}
}

func TestCommand_FormattedString(t *testing.T) {
	// GIVEN a Command
	tests := map[string]struct {
		cmd  Command
		want string
	}{
		"command with no args": {
			cmd:  Command{"ls"},
			want: "[ \"ls\" ]"},
		"command with one arg": {
			cmd:  Command{"ls", "-lah"},
			want: "[ \"ls\", \"-lah\" ]"},
		"command with multiple args": {
			cmd:  Command{"ls", "-lah", "/root", "/tmp"},
			want: "[ \"ls\", \"-lah\", \"/root\", \"/tmp\" ]"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the command is stringified with FormattedString
			got := tc.cmd.FormattedString()

			// THEN the result is expected
			if got != tc.want {
				t.Errorf("want: %q\ngot:\n%q",
					tc.want, got)
			}
		})
	}
}

func TestController_Metrics(t *testing.T) {
	// GIVEN a Controller with multiple Commands
	controller := &Controller{}
	controller.Init(
		&status.Status{ServiceID: test.StringPtr("TestController_Metrics")},
		&Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		nil,
		test.StringPtr("11m"))
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute)}
	for i := range nextRunnables {
		controller.SetNextRunnable(i, nextRunnables[i])
	}

	// WHEN the Prometheus metrics are initialised with initMetrics
	hadC := testutil.CollectAndCount(metric.CommandMetric)
	hadG := testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	controller.InitMetrics()

	// THEN it can be collected
	// counters
	gotC := testutil.CollectAndCount(metric.CommandMetric)
	wantC := 2 * len(*controller.Command)
	if (gotC - hadC) != wantC {
		t.Errorf("Controller.InitMetrics() %d Counter metrics were initialised, expecting %d",
			(gotC - hadC), wantC)
	}
	// gauges
	gotG := testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	wantG := 0
	if (gotG - hadG) != wantG {
		t.Errorf("Controller.InitMetrics() %d Gauge metrics were initialised, expecting %d",
			(gotG - hadG), wantG)
	}

	// AND it can be deleted
	// counters
	controller.DeleteMetrics()
	gotC = testutil.CollectAndCount(metric.CommandMetric)
	if gotC != hadC {
		t.Errorf("Controller.DeleteMetrics() was called, but Counter metrics mismatch got %d, expecting %d",
			gotC, hadC)
	}
	// gauges
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf("Controller.DeleteMetrics() was called, but Gauge metrics mismatch. got %d, expecting %d",
			gotG, hadG)
	}

	// AND a nil Controller doesn't panic
	controller = nil
	// InitMetrics
	controller.InitMetrics()
	// counters
	gotC = testutil.CollectAndCount(metric.CommandMetric)
	if gotC != hadC {
		t.Errorf("Controller.InitMetrics() on nil Controller shouldn't have changed the Counter metrics, got %d. expecting %d",
			gotC, hadC)
	}
	// gauges
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf("Controller.InitMetrics() on nil Controller shouldn't have changed the Gauge metrics, got %d. expecting %d",
			gotG, hadG)
	}
	// DeleteMetrics
	controller.DeleteMetrics()
	// counters
	gotC = testutil.CollectAndCount(metric.CommandMetric)
	if gotC != hadC {
		t.Errorf("Controller.DeleteMetrics() on nil Controller shouldn't have changed the Counter metrics, got %d. expecting %d",
			gotC, hadC)
	}
	// gauges
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf("Controller.DeleteMetrics() on nil Controller shouldn't have changed the Gauge metrics, got %d. expecting %d",
			gotG, hadG)
	}
}

func TestCommand_Init(t *testing.T) {
	// GIVEN a Command
	tests := map[string]struct {
		nilController     bool
		command           *Slice
		shoutrrrNotifiers *shoutrrr.Slice
		parentInterval    *string
	}{
		"nil Controller": {
			nilController: true,
		},
		"non-nil Controller": {
			command: &Slice{
				{"date", "+%m-%d-%Y"}},
		},
		"non-nil Commands": {
			command: &Slice{
				{"date", "+%m-%d-%Y"},
				{"true"},
				{"false"}},
		},
		"nil Notifiers": {
			command: &Slice{
				{"date", "+%m-%d-%Y"}},
		},
		"non-nil Notifiers": {
			command: &Slice{
				{"date", "+%m-%d-%Y"}},
			shoutrrrNotifiers: &shoutrrr.Slice{
				"test": shoutrrr_test.Shoutrrr(false, false)},
		},
		"nil parentInterval": {
			command: &Slice{
				{"date", "+%m-%d-%Y"}},
		},
		"non-nil parentInterval": {
			parentInterval: test.StringPtr("11m"),
			command: &Slice{
				{"date", "+%m-%d-%Y"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN a controller is initialised with Init
			var controller *Controller
			if !tc.nilController {
				controller = &Controller{}
			}
			serviceStatus := status.Status{}
			serviceStatus.ServiceID = test.StringPtr("TestInit")
			controller.Init(
				&serviceStatus,
				tc.command,
				tc.shoutrrrNotifiers,
				tc.parentInterval)

			// THEN the result is expected
			// nilController
			if tc.nilController {
				if controller != nil {
					t.Fatalf("Init of nil Controller gave %v",
						controller)
				}
				return
			}
			// serviceStatus
			if controller.ServiceStatus != &serviceStatus {
				t.Errorf("want: ServiceStatus=%v\ngot:  ServiceStatus=%v",
					controller.ServiceStatus, &serviceStatus)
			}
			// command
			if controller.Command != tc.command {
				t.Errorf("want: Command=%v\ngot:  Command=%v",
					controller.Command, tc.command)
			}
			// shoutrrrNotifiers
			if controller.Notifiers.Shoutrrr != tc.shoutrrrNotifiers {
				t.Errorf("want: Notifiers.Shoutrrr=%v\ngot:  Notifiers.Shoutrrr=%v",
					controller.Notifiers.Shoutrrr, tc.shoutrrrNotifiers)
			}
			// parentInterval
			if controller.ParentInterval != tc.parentInterval {
				t.Errorf("want: ParentInterval=%v\ngot:  ParentInterval=%v",
					controller.ParentInterval, tc.parentInterval)
			}
		})
	}
}
