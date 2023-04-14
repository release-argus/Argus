// Copyright [2023] [Argus]
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

	svcstatus "github.com/release-argus/Argus/service/status"
	api_type "github.com/release-argus/Argus/web/api/types"
)

func TestController_AnnounceCommand(t *testing.T) {
	// GIVEN Controllers with various failed Command announces
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
			failed:         nil,
		},
		"failed does delay by 15s": {
			index:          0,
			timeDifference: 15 * time.Second,
			failed:         boolPtr(true),
		},
		"success does delay by 2*Interval": {
			index:          1,
			timeDifference: 22 * time.Minute,
			failed:         boolPtr(false),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			controller := Controller{
				ParentInterval: stringPtr("11m"),
				ServiceStatus: &svcstatus.Status{
					ServiceID: stringPtr("some_service_id")}}
			controller.Init(
				&svcstatus.Status{
					ServiceID: stringPtr("some_service_id")},
				&Slice{
					{"ls", "-lah", "/root"},
					{"ls", "-lah"},
					{"ls", "-lah", "a"}},
				nil,
				stringPtr("11m"))
			if !tc.nilChannel {
				announceChannel := make(chan []byte, 4)
				controller.ServiceStatus.AnnounceChannel = &announceChannel
			}
			if tc.failed != nil {
				controller.Failed.Set(tc.index, *tc.failed)
			}
			time.Sleep(time.Millisecond)

			// WHEN AnnounceCommand is run
			go controller.AnnounceCommand(tc.index)

			// THEN the correct response is received
			if controller.ServiceStatus.AnnounceChannel == nil {
				return
			}
			m := <-*controller.ServiceStatus.AnnounceChannel
			var parsed api_type.WebSocketMessage
			json.Unmarshal(m, &parsed)

			if parsed.CommandData[(*controller.Command)[tc.index].String()] == nil {
				t.Fatalf("message wasn't for %q\ngot %v",
					(*controller.Command)[tc.index].String(), parsed.CommandData)
			}

			// if they failed status matches
			got := stringifyPointer(parsed.CommandData[(*controller.Command)[tc.index].String()].Failed)
			want := stringifyPointer(controller.Failed.Get(tc.index))
			if got != want {
				t.Errorf("want failed=%s\ngot  failed=%s",
					want, got)
			}

			// next runnable is within expectred range
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.CommandData[(*controller.Command)[tc.index].String()].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf("ran at\n%s\nwant between:\n%s and\n%s\ngot\n%s",
					now, minTime, maxTime, gotTime)
			}
		})
	}
}

func TestController_Find(t *testing.T) {
	// GIVEN we have a Controller with Command's
	tests := map[string]struct {
		command       string
		want          *int
		nilController bool
	}{
		"command at first index": {
			command: "ls -lah",
			want:    intPtr(0)},
		"command at second index": {
			command: "ls -lah a",
			want:    intPtr(1)},
		"command with svcstatus": {
			command: "bash upgrade.sh 1.2.3",
			want:    intPtr(3)},
		"unknown command": {
			command: "ls -lah /root",
			want:    nil},
		"nil controller": {
			command:       "ls -lah /root",
			want:          nil,
			nilController: true},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			controller := &Controller{
				Command: &Slice{
					Command{"ls", "-lah"},
					Command{"ls", "-lah", "a"},
					Command{"ls", "-lah", "b"},
					Command{"bash", "upgrade.sh", "{{ version }}"},
				},
				ServiceStatus: &svcstatus.Status{
					ServiceID: stringPtr("some_service_id")},
			}
			controller.ServiceStatus.Init(
				0, len(*controller.Command), 0,
				&name,
				nil)
			controller.Failed = &controller.ServiceStatus.Fails.Command
			controller.ServiceStatus.SetLatestVersion("1.2.3", false)
			if tc.nilController {
				controller = nil
			}

			// WHEN Find is run for a command
			index := controller.Find(tc.command)

			// THEN the index is returned if it exists
			got := stringifyPointer(index)
			want := stringifyPointer(tc.want)
			if got != want {
				t.Errorf("want: %s\ngot:  %s",
					want, got)
			}
		})
	}
}
