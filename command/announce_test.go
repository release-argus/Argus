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
	// GIVEN: Controllers with various failed Command announces.
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
			failed:         test.Ptr(true),
		},
		{
			name:           "success does delay by 2*Interval",
			index:          1,
			timeDifference: 22 * time.Minute,
			failed:         test.Ptr(false),
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
				test.Ptr("11m"),
			)
			if !tc.nilChannel {
				announceChannel := make(chan []byte, 4)
				controller.ServiceStatus.AnnounceChannel = announceChannel
			}
			if tc.failed != nil {
				controller.Failed.Set(tc.index, *tc.failed)
			}
			time.Sleep(time.Millisecond)

			// WHEN: AnnounceCommand is run.
			serviceInfo := controller.ServiceStatus.GetServiceInfo()
			go controller.AnnounceCommand(tc.index, serviceInfo)

			prefix := fmt.Sprintf(
				"%s\nController.AnnounceCommand(index=%d, info=%v)",
				packageName, tc.index, serviceInfo,
			)

			// THEN: the correct response is received.
			if controller.ServiceStatus.AnnounceChannel == nil {
				return
			}
			m := <-controller.ServiceStatus.AnnounceChannel
			var parsed apitype.WebSocketMessage
			_ = decode.Unmarshal("json", m, &parsed)

			cmdStr := controller.Command[tc.index].String()
			if parsed.CommandData[cmdStr] == nil {
				t.Fatalf(
					"%s Command data not seen in AnnounceChannel message\ngot:  %+v\nwant: %q",
					prefix, parsed.CommandData, cmdStr,
				)
			}

			// AND: the failed state matches.
			want := test.StringifyPtr(controller.Failed.Get(tc.index))
			got := test.StringifyPtr(parsed.CommandData[controller.Command[tc.index].String()].Failed)
			if got != want {
				t.Errorf(
					"%s Controller.Failed state mismatch at index %d\ngot:  %s\nwant: %s",
					packageName, tc.index, got, want,
				)
			}

			// AND: 'next runnable' is within expected range.
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.CommandData[controller.Command[tc.index].String()].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf(
					"%s Controller.NextRunnable[%d] mismatch\nran at\n%s\ngot:\n%s\nwant between:\n%s and\n%s",
					packageName, tc.index,
					now,
					gotTime,
					minTime, maxTime,
				)
			}
		})
	}
}

func TestController_Find(t *testing.T) {
	// GIVEN: we have a Controller with Commands.
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
			command:  "bash upgrade.sh 1.2.3",
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
			controllerCommand := controller.Command

			controller.ServiceStatus.Init(
				len(controller.Command), 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			controller.Failed = &controller.ServiceStatus.Fails.Command
			controller.ServiceStatus.SetLatestVersion("1.2.3", "", false)
			if tc.nilController {
				controller = nil
				controllerCommand = nil
			}

			// WHEN: Find is run for a command.
			got, err := controller.Find(tc.command)

			prefix := fmt.Sprintf(
				"%s\nController Find(%q) (commands=%v)",
				packageName, tc.command, controllerCommand,
			)

			// THEN: the index is returned if it exists.
			if got != tc.want {
				t.Errorf(
					"%s unexpected index\ngot:  %d\nwant: %d",
					prefix, got, tc.want,
				)
			}

			// AND: an error is returned if it doesn't.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}
