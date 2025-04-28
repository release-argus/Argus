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

package command

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_test "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/test"
	metric "github.com/release-argus/Argus/web/metric"
)

func TestController_SetExecuting(t *testing.T) {
	// GIVEN a Controller with various Commands.
	controller := Controller{}
	controller.Init(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id"}},
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

			// WHEN SetNextRunnable is called.
			ranAt := time.Now().UTC()
			controller.SetExecuting(tc.index, tc.executing)

			// THEN the result is expected.
			got := ranAt
			if tc.index < len(*controller.Command) {
				got = (controller.NextRunnable(tc.index))
			}
			minTime := ranAt.Add(tc.timeDifferenceMin)
			maxTime := ranAt.Add(tc.timeDifferenceMax)
			if !(minTime.Before(got)) || !(maxTime.After(got)) {
				t.Fatalf("%s\nran at\n%s\nwant between:\n%s and\n%s\ngot:\n%s",
					packageName, ranAt, minTime, maxTime, got)
			}
		})
	}
}

func TestController_IsRunnable(t *testing.T) {
	// GIVEN a Controller with various Commands.
	controller := Controller{}
	controller.Init(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id"}},
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

			// WHEN IsRunnable is called.
			got := controller.IsRunnable(tc.index)

			// THEN the result is expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestController_NextRunnable(t *testing.T) {
	// GIVEN a Controller with various Commands.
	controller := Controller{}
	controller.Init(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id"}},
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

			// WHEN NextRunnable is called.
			got := controller.NextRunnable(tc.index)

			// THEN the result is expected.
			if tc.outOfRange {
				var defaultTime time.Time
				// out of range index should return the default time.
				if got != defaultTime {
					t.Fatalf("%s\nout of range\nwant: %s\ngot:  %s",
						packageName, defaultTime, got)
				}
			} else if got != controller.NextRunnable(tc.index) {
				t.Fatalf("%s\nwant: %s\ngot:  %s",
					packageName, controller.NextRunnable(tc.index), got)
			}
		})
	}
}

func TestController_SetNextRunnable(t *testing.T) {
	// GIVEN a Controller with various Commands.
	controller := Controller{}
	controller.Init(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "service_id"}},
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

			// Reset nextRunnables.
			for i := range nextRunnables {
				controller.SetNextRunnable(i, nextRunnables[i])
			}

			// WHEN SetNextRunnable is called.
			controller.SetNextRunnable(tc.index, tc.setTo)

			// THEN the NextRunnable is changed if the index is in range.
			if tc.index < len(*controller.Command) {
				got := controller.NextRunnable(tc.index)
				if !got.Equal(tc.setTo) {
					t.Errorf("%s\nwant: %s\ngot:  %s",
						packageName, tc.setTo, got)
				}
			} else {
				// Ensure out of range index does not panic and does not change anything.
				for i := range nextRunnables {
					got := controller.NextRunnable(i)
					if !got.Equal(nextRunnables[i]) {
						t.Errorf("%s\nindex out of range should not change nextRunnable. want: %s\ngot:  %s",
							packageName, nextRunnables[i], got)
					}
				}
			}
		})
	}
}

func TestCommand_String(t *testing.T) {
	// GIVEN a Command.
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

			// WHEN the command is stringified with String().
			got := tc.cmd.String()

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestCommand_FormattedString(t *testing.T) {
	// GIVEN a Command.
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

			// WHEN the command is stringified with FormattedString.
			got := tc.cmd.FormattedString()

			// THEN the result is expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestController_Metrics(t *testing.T) {
	// GIVEN a Controller with multiple Commands.
	controller := &Controller{}
	controller.Init(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: "TestController_Metrics"}},
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

	// WHEN the Prometheus metrics are initialised with initMetrics.
	hadC := testutil.CollectAndCount(metric.CommandResultTotal)
	hadG := testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	controller.InitMetrics()

	// THEN it can be collected.
	// 	Counters:
	gotC := testutil.CollectAndCount(metric.CommandResultTotal)
	wantC := 2 * len(*controller.Command)
	if (gotC - hadC) != wantC {
		t.Errorf("%s\nCounter metrics mismatch after InitMetrics() \nwant: %d\ngot:  %d",
			packageName, wantC, (gotC - hadC))
	}
	// 	Gauges:
	gotG := testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	wantG := 0
	if (gotG - hadG) != wantG {
		t.Errorf("%s\nGauge metrics mismatch after InitMetrics()\nwant: %d\ngot:  %d",
			packageName, wantG, (gotG - hadG))
	}

	// AND it can be deleted.
	// 	Counters:
	controller.DeleteMetrics()
	gotC = testutil.CollectAndCount(metric.CommandResultTotal)
	if gotC != hadC {
		t.Errorf("%s\nCounter metrics mismatch after DeleteMetrics()\nwant: %d\ngot:  %d",
			packageName, hadC, gotC)
	}
	// 	Gauges:
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf("%s\nGauge metrics mismatch after DeleteMetrics()\nwant: %d\ngot:  %d",
			packageName, hadG, gotG)
	}

	// AND a nil Controller doesn't panic.
	controller = nil
	// InitMetrics:
	controller.InitMetrics()
	// 	Counter:
	gotC = testutil.CollectAndCount(metric.CommandResultTotal)
	if gotC != hadC {
		t.Errorf("%s\nInitMetrics() on nil Controller shouldn't have changed the Counter metrics\nwant: %d\ngot:  %d",
			packageName, hadC, gotC)
	}
	// 	Gauges:
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf("%s\nInitMetrics() on nil Controller shouldn't have changed the Gauge metrics\nwant: %d\ngot:  %d",
			packageName, hadG, gotG)
	}
	// DeleteMetrics:
	controller.DeleteMetrics()
	// 	Counters:
	gotC = testutil.CollectAndCount(metric.CommandResultTotal)
	if gotC != hadC {
		t.Errorf("%s\nDeleteMetrics() on nil Controller shouldn't have changed the Counter metrics\nwant: %d\ngot:  %d",
			packageName, hadC, gotC)
	}
	// 	Gauges:
	gotG = testutil.CollectAndCount(metric.LatestVersionIsDeployed)
	if gotG != hadG {
		t.Errorf("%s\nDeleteMetrics() on nil Controller shouldn't have changed the Gauge metrics\nwant: %d\ngot:  %d",
			packageName, hadG, gotG)
	}
}

func TestCommand_Init(t *testing.T) {
	// GIVEN a Command.
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

			// WHEN a controller is initialised with Init.
			var controller *Controller
			if !tc.nilController {
				controller = &Controller{}
			}
			serviceStatus := status.Status{
				ServiceInfo: serviceinfo.ServiceInfo{
					ID: "TestInit"}}
			controller.Init(
				&serviceStatus,
				tc.command,
				tc.shoutrrrNotifiers,
				tc.parentInterval)

			// THEN the result is expected.
			// 	nilController:
			if tc.nilController {
				if controller != nil {
					t.Fatalf("%s\nInit of nil Controller gave %v",
						packageName, controller)
				}
				return
			}
			// 	serviceStatus:
			if controller.ServiceStatus != &serviceStatus {
				t.Errorf("%s\nwant: ServiceStatus=%v\ngot:  ServiceStatus=%v",
					packageName, controller.ServiceStatus, &serviceStatus)
			}
			// 	command:
			if controller.Command != tc.command {
				t.Errorf("%s\nwant: Command=%v\ngot:  Command=%v",
					packageName, controller.Command, tc.command)
			}
			// 	shoutrrrNotifiers:
			if controller.Notifiers.Shoutrrr != tc.shoutrrrNotifiers {
				t.Errorf("%s\nwant: Notifiers.Shoutrrr=%v\ngot:  Notifiers.Shoutrrr=%v",
					packageName, controller.Notifiers.Shoutrrr, tc.shoutrrrNotifiers)
			}
			// 	parentInterval:
			if controller.ParentInterval != tc.parentInterval {
				t.Errorf("%s\nwant: ParentInterval=%v\ngot:  ParentInterval=%v",
					packageName, controller.ParentInterval, tc.parentInterval)
			}
		})
	}
}
