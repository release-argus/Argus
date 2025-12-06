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
	"encoding/json"
	"testing"
	"time"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/test"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestController_AnnounceCommand(t *testing.T) {
	// GIVEN Controllers with various failed Command announces.
	tests := map[string]struct {
		nilChannel     bool
		index          int
		failed         *bool
		timeDifference time.Duration
	}{
		"no channel": {
			nilChannel: true,
		},
		"not tried does delay by 15s": {
			index:          2,
			timeDifference: 15 * time.Second,
		},
		"failed does delay by 15s": {
			index:          0,
			timeDifference: 15 * time.Second,
			failed:         test.BoolPtr(true),
		},
		"success does delay by 2*Interval": {
			index:          1,
			timeDifference: 22 * time.Minute,
			failed:         test.BoolPtr(false),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			controller := Controller{
				ParentInterval: test.StringPtr("11m"),
				ServiceStatus: &status.Status{
					ServiceInfo: serviceinfo.ServiceInfo{
						ID: "some_service_id"}}}
			controller.Init(
				&status.Status{
					ServiceInfo: serviceinfo.ServiceInfo{
						ID: "some_service_id"}},
				&Commands{
					{"ls", "-lah", "/root"},
					{"ls", "-lah"},
					{"ls", "-lah", "a"}},
				nil,
				test.StringPtr("11m"))
			if !tc.nilChannel {
				announceChannel := make(chan []byte, 4)
				controller.ServiceStatus.AnnounceChannel = &announceChannel
			}
			if tc.failed != nil {
				controller.Failed.Set(tc.index, *tc.failed)
			}
			time.Sleep(time.Millisecond)

			// WHEN AnnounceCommand is run.
			go controller.AnnounceCommand(tc.index, controller.ServiceStatus.GetServiceInfo())

			// THEN the correct response is received.
			if controller.ServiceStatus.AnnounceChannel == nil {
				return
			}
			m := <-*controller.ServiceStatus.AnnounceChannel
			var parsed apitype.WebSocketMessage
			json.Unmarshal(m, &parsed)

			if parsed.CommandData[(*controller.Command)[tc.index].String()] == nil {
				t.Fatalf("%s\nmessage wasn't for %q\ngot %v",
					packageName, (*controller.Command)[tc.index].String(), parsed.CommandData)
			}

			// if they failed status matches.
			want := test.StringifyPtr(controller.Failed.Get(tc.index))
			got := test.StringifyPtr(parsed.CommandData[(*controller.Command)[tc.index].String()].Failed)
			if got != want {
				t.Errorf("%s\nfailed mismatch\nwant: %s\ngot:  %s",
					packageName, want, got)
			}

			// next runnable is within expected range.
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.CommandData[(*controller.Command)[tc.index].String()].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf("%s\nran at\n%s\nwant between:\n%s and\n%s\ngot\n%s",
					packageName, now, minTime, maxTime, gotTime)
			}
		})
	}
}

func TestController_Find(t *testing.T) {
	// GIVEN we have a Controller with Commands.
	tests := map[string]struct {
		command       string
		want          int
		err           bool
		nilController bool
	}{
		"command at first index": {
			command: "ls -lah",
			want:    0},
		"command at second index": {
			command: "ls -lah a",
			want:    1},
		"command with status": {
			command: "bash upgrade.sh 1.2.3",
			want:    3},
		"unknown command": {
			command: "ls -lah /root",
			err:     true},
		"nil controller": {
			command:       "ls -lah /root",
			err:           true,
			nilController: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			controller := &Controller{
				Command: &Commands{
					Command{"ls", "-lah"},
					Command{"ls", "-lah", "a"},
					Command{"ls", "-lah", "b"},
					Command{"bash", "upgrade.sh", "{{ version }}"},
				},
				ServiceStatus: &status.Status{
					ServiceInfo: serviceinfo.ServiceInfo{
						ID: "some_service_id"}},
			}
			controller.ServiceStatus.Init(
				0, len(*controller.Command), 0,
				name, "", "",
				&dashboard.Options{})
			controller.Failed = &controller.ServiceStatus.Fails.Command
			controller.ServiceStatus.SetLatestVersion("1.2.3", "", false)
			if tc.nilController {
				controller = nil
			}

			// WHEN Find is run for a command.
			got, err := controller.Find(tc.command)

			// THEN the index is returned if it exists.
			if got != tc.want {
				t.Errorf("%s\nunexpected index\nwant: %d\ngot:  %d",
					packageName, tc.want, got)
			}
			// AND an error is returned if it doesn't.
			if err != nil != tc.err {
				t.Errorf("%s\nerror mismatch\nwant: err=%t\ngot:  %v",
					packageName, tc.err, err)
			}
		})
	}
}
