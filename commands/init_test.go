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

package command

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	metrics "github.com/release-argus/Argus/web/metrics"
)

func TestSetNextRunnable(t *testing.T) {
	// GIVEN a Controller with various Command's
	controller := Controller{
		Command: &Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		ServiceStatus:  &service_status.Status{ServiceID: stringPtr("service_id")},
		Failed:         &[]*bool{nil, boolPtr(false), boolPtr(true)},
		NextRunnable:   make([]time.Time, 3),
		ParentInterval: stringPtr("11m"),
	}
	tests := map[string]struct {
		index             int
		executing         bool
		timeDifferenceMin time.Duration
		timeDifferenceMax time.Duration
	}{
		"index out of range": {index: 3, timeDifferenceMin: -time.Second, timeDifferenceMax: time.Second},
		"command that hasn't been run and isn't currently running": {index: 0, timeDifferenceMin: 14 * time.Second, timeDifferenceMax: 16 * time.Second},
		"command that hasn't been run and is currently running":    {index: 0, executing: true, timeDifferenceMin: time.Hour + 14*time.Second, timeDifferenceMax: time.Hour + 16*time.Second},
		"command that didn't fail and isn't currently running":     {index: 1, timeDifferenceMin: 22*time.Minute - time.Second, timeDifferenceMax: 22*time.Minute + time.Second},
		"command that didn't fail and is currently running":        {index: 1, executing: true, timeDifferenceMin: time.Hour + (22*time.Minute - time.Second), timeDifferenceMax: time.Hour + (22*time.Minute + time.Second)},
		"command that did fail and isn't currently running":        {index: 2, timeDifferenceMin: 14 * time.Second, timeDifferenceMax: 16 * time.Second},
		"command that did fail and is currently running":           {index: 2, executing: true, timeDifferenceMin: time.Hour + 14*time.Second, timeDifferenceMax: time.Hour + 16*time.Second},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN SetNextRunnable is called
			ranAt := time.Now().UTC()
			controller.SetNextRunnable(tc.index, tc.executing)

			// THEN the result is expected
			got := ranAt
			if tc.index < len(controller.NextRunnable) {
				got = (controller.NextRunnable[tc.index])
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

func TestIsRunnable(t *testing.T) {
	// GIVEN a Controller with various Command's
	controller := Controller{
		Command: &Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		ServiceStatus:  &service_status.Status{ServiceID: stringPtr("service_id")},
		Failed:         &[]*bool{nil, boolPtr(false), boolPtr(true)},
		NextRunnable:   []time.Time{time.Now().UTC(), time.Now().UTC().Add(-time.Minute), time.Now().UTC().Add(time.Minute)},
		ParentInterval: stringPtr("11m"),
	}
	tests := map[string]struct {
		index int
		want  bool
	}{
		"NextRunnable just passed":  {index: 0, want: true},
		"NextRunnable a minute ago": {index: 1, want: true},
		"NextRunnable in a minute":  {index: 2, want: false},
		"NextRunnable out of range": {index: 3, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

func TestGetNextRunnable(t *testing.T) {
	// GIVEN a Controller with various Command's
	controller := Controller{
		Command: &Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		ServiceStatus:  &service_status.Status{ServiceID: stringPtr("service_id")},
		Failed:         &[]*bool{nil, boolPtr(false), boolPtr(true)},
		NextRunnable:   []time.Time{time.Now().UTC(), time.Now().UTC().Add(-time.Minute), time.Now().UTC().Add(time.Minute)},
		ParentInterval: stringPtr("11m"),
	}
	tests := map[string]struct {
		index      int
		setTo      time.Time
		want       bool
		outOfRange bool
	}{
		"NextRunnable just passed":  {index: 0, want: true},
		"NextRunnable a minute ago": {index: 1, want: true},
		"NextRunnable in a minute":  {index: 2, want: false},
		"NextRunnable out of range": {index: 3, outOfRange: true, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN GetNextRunnable is called
			got := controller.GetNextRunnable(tc.index)

			// THEN the result is expected
			if tc.outOfRange {
				var defaultTime time.Time
				if got != defaultTime {
					t.Fatalf("want: %s\ngot:\n%s",
						defaultTime, got)
				}
			} else if got != controller.NextRunnable[tc.index] {
				t.Fatalf("want: %s\ngot:\n%s",
					controller.NextRunnable[tc.index], got)
			}
		})
	}
}

func TestString(t *testing.T) {
	// GIVEN a Command
	tests := map[string]struct {
		cmd  Command
		want string
	}{
		"command with no args":       {cmd: Command{"ls"}, want: "ls"},
		"command with one arg":       {cmd: Command{"ls", "-lah"}, want: "ls -lah"},
		"command with multiple args": {cmd: Command{"ls", "-lah", "/root", "/tmp"}, want: "ls -lah /root /tmp"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN the command is stringified with String
			got := tc.cmd.String()

			// THEN the result is expected
			if got != tc.want {
				t.Errorf("want: %q\ngot:\n%q",
					tc.want, got)
			}
		})
	}
}

func TestFormattedString(t *testing.T) {
	// GIVEN a Command
	tests := map[string]struct {
		cmd  Command
		want string
	}{
		"command with no args":       {cmd: Command{"ls"}, want: "[ \"ls\" ]"},
		"command with one arg":       {cmd: Command{"ls", "-lah"}, want: "[ \"ls\", \"-lah\" ]"},
		"command with multiple args": {cmd: Command{"ls", "-lah", "/root", "/tmp"}, want: "[ \"ls\", \"-lah\", \"/root\", \"/tmp\" ]"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

func TestInitMetrics(t *testing.T) {
	// GIVEN a Controller with multiple Command's
	controller := Controller{
		Command: &Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}},
		ServiceStatus:  &service_status.Status{ServiceID: stringPtr("InitMetrics")},
		Failed:         &[]*bool{nil, boolPtr(false), boolPtr(true)},
		NextRunnable:   []time.Time{time.Now().UTC(), time.Now().UTC().Add(-time.Minute), time.Now().UTC().Add(time.Minute)},
		ParentInterval: stringPtr("11m"),
	}

	// WHEN the Prometheus metrics are initialised with initMetrics
	hadC := testutil.CollectAndCount(metrics.CommandMetric)
	hadG := testutil.CollectAndCount(metrics.AckWaiting)
	controller.initMetrics()

	// THEN it can be collected
	// counters
	gotC := testutil.CollectAndCount(metrics.CommandMetric)
	wantC := 2 * len(*controller.Command)
	if (gotC - hadC) != wantC {
		t.Errorf("%d Counter metrics's were initialised, expecting %d",
			(gotC - hadC), wantC)
	}
	// gauges
	gotG := testutil.CollectAndCount(metrics.AckWaiting)
	wantG := 1
	if (gotG - hadG) != wantG {
		t.Errorf("%d Gauge metrics's were initialised, expecting %d",
			(gotG - hadG), wantG)
	}
}

func TestInit(t *testing.T) {
	// GIVEN a Command
	tests := map[string]struct {
		nilController     bool
		log               *utils.JLog
		serviceStatus     service_status.Status
		command           *Slice
		shoutrrrNotifiers *shoutrrr.Slice
		parentInterval    *string
	}{
		"nil log":            {log: nil, command: &Slice{{"date", "+%m-%d-%Y"}}},
		"non-nil log":        {log: utils.NewJLog("INFO", false), command: &Slice{{"date", "+%m-%d-%Y"}}},
		"nil Controller":     {nilController: true},
		"non-nil Controller": {command: &Slice{{"date", "+%m-%d-%Y"}}},
		"non-nil Command's": {command: &Slice{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"}}},
		"nil Notifiers":          {shoutrrrNotifiers: nil, command: &Slice{{"date", "+%m-%d-%Y"}}},
		"non-nil Notifiers":      {shoutrrrNotifiers: nil, command: &Slice{{"date", "+%m-%d-%Y"}}},
		"nil parentInterval":     {parentInterval: nil, command: &Slice{{"date", "+%m-%d-%Y"}}},
		"non-nil parentInterval": {parentInterval: stringPtr("11m"), command: &Slice{{"date", "+%m-%d-%Y"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN a controller is initialised with Init
			var controller *Controller
			if !tc.nilController {
				controller = &Controller{}
			}
			logBefore := jLog
			tc.serviceStatus.ServiceID = stringPtr("TestInit")
			controller.Init(tc.log, &tc.serviceStatus, tc.command, tc.shoutrrrNotifiers, tc.parentInterval)

			// THEN the result is expected
			// log
			if (tc.log == nil && jLog != logBefore) && jLog != tc.log {
				t.Errorf("want: jLog=%v\ngot:  jLog=%v",
					tc.log, jLog)
			}
			// nilController
			if tc.nilController {
				if controller != nil {
					t.Fatalf("Init of nil Controller gave %v",
						controller)
				}
				return
			}
			// serviceStatus
			if controller.ServiceStatus != &tc.serviceStatus {
				t.Errorf("want: ServiceStatus=%v\ngot:  ServiceStatus=%v",
					controller.ServiceStatus, &tc.serviceStatus)
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
