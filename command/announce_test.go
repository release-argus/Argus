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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestController_AnnounceCommand(t *testing.T) {
	tests := []struct {
		name           string
		nilChannel     bool
		index          int
		failed         *bool
		timeDifference time.Duration
	}{
		{
			name:       "no channel",
			nilChannel: true,
		},
		{
			name:           "not tried does delay by 15s",
			index:          2,
			timeDifference: 15 * time.Second,
		},
		{
			name:           "failed does delay by 15s",
			index:          0,
			timeDifference: 15 * time.Second,
			failed:         new(true),
		},
		{
			name:           "success does delay by 2*Interval",
			index:          1,
			timeDifference: 22 * time.Minute,
			failed:         new(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			controller := NewController(
				&status.Status{
					ServiceInfo: serviceinfo.ServiceInfo{
						ID: "some_service_id",
					},
				},
				Commands{
					{"ls", "-lah", "/root"},
					{"ls", "-lah"},
					{"ls", "-lah", "a"},
				},
				nil,
				new("11m"),
			)
			if !tc.nilChannel {
				announceChannel := make(chan []byte, 4)
				controller.ServiceStatus.AnnounceChannel = announceChannel
			}
			if tc.failed != nil {
				controller.Failed.Set(tc.index, *tc.failed)
			}
			time.Sleep(time.Millisecond)

			serviceInfo := controller.ServiceStatus.GetServiceInfo()
			done := make(chan struct{})
			go func() {
				controller.AnnounceCommand(tc.index, serviceInfo)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("Controller.AnnounceCommand() did not return")
			}
			if tc.nilChannel {
				return
			}

			var m []byte
			select {
			case m = <-controller.ServiceStatus.AnnounceChannel:
			default:
				t.Fatal("Status.AnnounceChannel message count mismatch\ngot:  0\nwant: 1")
			}
			var parsed apitype.WebSocketMessage
			if err := decode.Unmarshal("json", m, &parsed); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}

			cmdStr := controller.Command[tc.index].String()
			if parsed.CommandData[cmdStr] == nil {
				t.Fatalf(
					"WebSocketMessage.CommandData[%q] mismatch\ngot:  absent (%q)\nwant: present",
					cmdStr, util.SortedKeys(parsed.CommandData),
				)
			}

			want := test.StringifyPtr(controller.Failed.Get(tc.index))
			got := test.StringifyPtr(parsed.CommandData[cmdStr].Failed)
			if got != want {
				t.Errorf(
					"WebSocketMessage.CommandData.Failed mismatch\ngot:  %s\nwant: %s",
					got, want,
				)
			}

			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.CommandData[cmdStr].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf(
					"WebSocketMessage.CommandData.NextRunnable mismatch (checked at %s)\n"+
						"got:  %s\nwant: between %s and %s",
					now, gotTime, minTime, maxTime,
				)
			}
		})
	}
}

func TestController_Find(t *testing.T) {
	latestVersion := "1.2.3"

	tests := []struct {
		name          string
		command       string
		nilController bool
		want          int
		errRegex      string
	}{
		{
			name:     "command at first index",
			command:  "ls -lah",
			want:     0,
			errRegex: `^$`,
		},
		{
			name:     "command at second index",
			command:  "ls -lah a",
			want:     1,
			errRegex: `^$`,
		},
		{
			name:     "command with status",
			command:  "bash upgrade.sh " + latestVersion,
			want:     3,
			errRegex: `^$`,
		},
		{
			name:     "unknown command",
			command:  "ls -lah /root",
			errRegex: `^command "[^"]+" not found$`,
		},
		{
			name:          "nil controller",
			command:       "ls -lah /root",
			nilController: true,
			errRegex:      `^controller is nil$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			controller := &Controller{
				Command: Commands{
					{"ls", "-lah"},
					{"ls", "-lah", "a"},
					{"ls", "-lah", "b"},
					{"bash", "upgrade.sh", "{{ version }}"},
				},
				ServiceStatus: &status.Status{
					ServiceInfo: serviceinfo.ServiceInfo{
						ID: "some_service_id",
					},
				},
			}
			controller.ServiceStatus.Init(
				len(controller.Command), 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			controller.Failed = &controller.ServiceStatus.Fails.Command
			controller.ServiceStatus.SetLatestVersion(latestVersion, "", false)
			if tc.nilController {
				controller = nil
			}

			got, err := controller.Find(tc.command)

			if got != tc.want {
				t.Errorf("index mismatch\ngot:  %d\nwant: %d", got, tc.want)
			}

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}
