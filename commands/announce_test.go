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
	"encoding/json"
	"testing"
	"time"

	service_status "github.com/release-argus/Argus/service/status"
	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestAnnounceCommand(t *testing.T) {
	// GIVEN Controllers with various failed Command announces
	fails := make([]*bool, 3)
	tests := map[string]struct {
		nilChannel     bool
		index          int
		failed         *bool
		timeDifference time.Duration
	}{
		"no channel": {nilChannel: true},
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
		t.Run(name, func(t *testing.T) {
			controller := Controller{
				Command: &Slice{
					{"ls", "-lah", "/root"},
					{"ls", "-lah"},
					{"ls", "-lah", "a"},
				},
				Failed:         &fails,
				NextRunnable:   make([]time.Time, 3),
				ParentInterval: stringPtr("11m"),
				ServiceStatus:  &service_status.Status{ServiceID: stringPtr("some_service_id"), AnnounceChannel: nil}}
			if !tc.nilChannel {
				announceChannel := make(chan []byte, 4)
				controller.ServiceStatus.AnnounceChannel = &announceChannel
			}
			(*controller.Failed)[tc.index] = tc.failed
			time.Sleep(time.Millisecond)

			// WHEN AnnounceCommand is run
			go controller.AnnounceCommand(tc.index)

			// THEN the correct response is received
			if controller.ServiceStatus.AnnounceChannel == nil {
				return
			}
			m := <-*controller.ServiceStatus.AnnounceChannel
			var parsed api_types.WebSocketMessage
			json.Unmarshal(m, &parsed)

			if parsed.CommandData[(*controller.Command)[tc.index].String()] == nil {
				t.Fatalf("%s:\nmessage wasn't for %q\ngot %v",
					name, (*controller.Command)[tc.index].String(), parsed.CommandData)
			}

			// if they failed status matches
			got := stringifyPointer(parsed.CommandData[(*controller.Command)[tc.index].String()].Failed)
			want := stringifyPointer((*controller.Failed)[tc.index])
			if got != want {
				t.Errorf("%s:\nwant failed=%s\ngot  failed=%s",
					name, want, got)
			}

			// next runnable is within expectred range
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.CommandData[(*controller.Command)[tc.index].String()].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf("%s:\nran at\n%s\nwant between:\n%s and\n%s\ngot\n%s",
					name, now, minTime, maxTime, gotTime)
			}
		})
	}
}

func TestFind(t *testing.T) {
	// GIVEN we have a Controller with Command's
	fails := make([]*bool, 3)
	controller := Controller{
		Command: &Slice{
			Command{"ls", "-lah"},
			Command{"ls", "-lah", "a"},
			Command{"ls", "-lah", "b"},
			Command{"bash", "upgrade.sh", "{{ version }}"},
		},
		ServiceStatus: &service_status.Status{ServiceID: stringPtr("some_service_id"), LatestVersion: "1.2.3"},
		Failed:        &fails,
	}
	tests := map[string]struct {
		command       string
		want          *int
		nilController bool
	}{
		"command at first index":      {command: "ls -lah", want: intPtr(0)},
		"command at second index":     {command: "ls -lah a", want: intPtr(1)},
		"command with service_status": {command: "bash upgrade.sh 1.2.3", want: intPtr(3)},
		"unknown command":             {command: "ls -lah /root", want: nil},
		"nil controller":              {command: "ls -lah /root", want: nil, nilController: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var target *Controller
			if !tc.nilController {
				target = &controller
			}
			// WHEN Find is run for a command
			index := target.Find(tc.command)

			// THEN the index is returned if it exists
			got := stringifyPointer(index)
			want := stringifyPointer(tc.want)
			if got != want {
				t.Errorf("%s:\nwant: %s\ngot:  %s",
					name, want, got)
			}
		})
	}
}

func TestResetFails(t *testing.T) {
	// GIVEN we have a Controller
	tests := map[string]struct {
		controller *Controller
	}{
		"nil controller": {controller: nil},
		"controller with all fails": {controller: &Controller{
			Failed: &[]*bool{boolPtr(true), boolPtr(true)}}},
		"controller with no fails": {controller: &Controller{
			Failed: &[]*bool{boolPtr(false), boolPtr(false)}}},
		"controller with some fails": {controller: &Controller{
			Failed: &[]*bool{boolPtr(true), boolPtr(false)}}},
		"controller with nil fails": {controller: &Controller{
			Failed: &[]*bool{nil, nil}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN ResetFails is run on this controller
			tc.controller.ResetFails()

			// THEN all the Failed's are reset to nil
			if tc.controller == nil {
				return
			}
			for i := range *tc.controller.Failed {
				if (*tc.controller.Failed)[i] != nil {
					t.Errorf("%s:\nfails weren't reset to nil. got %v",
						name, tc.controller.Failed)
				}
			}
		})
	}
}
