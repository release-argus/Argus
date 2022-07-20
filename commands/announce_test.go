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
	"fmt"
	"testing"
	"time"

	service_status "github.com/release-argus/Argus/service/status"
	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestAnnounceCommand(t *testing.T) {
	// GIVEN Controllers with various failed Command announces
	controller := Controller{
		ServiceID: stringPtr("some_service_id"),
		Command: &Slice{
			Command{"ls", "-lah", "/root"},
			Command{"ls", "-lah"},
			Command{"ls", "-lah", "a"},
		},
		Failed:         Fails{boolPtr(true), boolPtr(false), nil},
		NextRunnable:   make([]time.Time, 3),
		ParentInterval: stringPtr("11m")}
	tests := map[string]struct {
		nilChannel     bool
		channel        chan []byte
		index          int
		failed         *bool
		timeDifference time.Duration
	}{
		"no channel": {nilChannel: true},
		"failed nil does delay by 15s": {
			channel:        make(chan []byte, 4),
			index:          2,
			timeDifference: 15 * time.Second,
		},
		"failed true does delay by 15s": {
			channel:        make(chan []byte, 4),
			index:          0,
			timeDifference: 15 * time.Second,
		},
		"failed false does delay by 2*Interval": {
			channel:        make(chan []byte, 4),
			index:          1,
			timeDifference: 22 * time.Minute,
		},
	}

	for name, tc := range tests {
		if tc.nilChannel {
			controller.Announce = nil
		} else {
			controller.Announce = &tc.channel
		}

		// WHEN AnnounceCommand is run
		go controller.AnnounceCommand(tc.index)

		// THEN the correct response is received
		if controller.Announce == nil {
			return
		}
		m := <-*controller.Announce
		var parsed api_types.WebSocketMessage
		json.Unmarshal(m, &parsed)

		if parsed.CommandData[(*controller.Command)[tc.index].String()] == nil {
			t.Fatalf("%s:\nmessage wasn't for %q\ngot %v",
				name, (*controller.Command)[tc.index].String(), parsed.CommandData)
		}

		// if they failed status matches
		got := "nil"
		if parsed.CommandData[(*controller.Command)[tc.index].String()].Failed != nil {
			got = fmt.Sprint(*parsed.CommandData[(*controller.Command)[tc.index].String()].Failed)
		}
		want := "nil"
		if controller.Failed[tc.index] != nil {
			want = fmt.Sprint(*controller.Failed[tc.index])
		}
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
	}
}

func TestFind(t *testing.T) {
	// GIVEN we have a Controller with Command's
	controller := Controller{
		ServiceID: stringPtr("some_service_id"),
		Command: &Slice{
			Command{"ls", "-lah"},
			Command{"ls", "-lah", "a"},
			Command{"ls", "-lah", "b"},
			Command{"bash", "upgrade.sh", "{{ version }}"},
		},
		ServiceStatus: &service_status.Status{LatestVersion: "1.2.3"},
		Failed:        make(Fails, 3),
	}
	tests := map[string]struct {
		command string
		want    *int
	}{
		"command at first index":      {command: "ls -lah", want: intPtr(0)},
		"command at second index":     {command: "ls -lah a", want: intPtr(1)},
		"command with service_status": {command: "bash upgrade.sh 1.2.3", want: intPtr(3)},
		"unknwon command":             {command: "ls -lah /root", want: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN Find is run for a command
			index := controller.Find(tc.command)

			// THEN the index is returned if it exists
			got := "nil"
			if index != nil {
				got = fmt.Sprint(*index)
			}
			want := "nil"
			if tc.want != nil {
				want = fmt.Sprint(*tc.want)
			}
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
			Failed: Fails{boolPtr(true), boolPtr(true)}}},
		"controller with no fails": {controller: &Controller{
			Failed: Fails{boolPtr(false), boolPtr(false)}}},
		"controller with some fails": {controller: &Controller{
			Failed: Fails{boolPtr(true), boolPtr(false)}}},
		"controller with nil fails": {controller: &Controller{
			Failed: Fails{nil, nil}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN ResetFails is run on this controller
			tc.controller.ResetFails()

			// THEN all the Failed's are reset to nil
			if tc.controller == nil {
				return
			}
			for i := range tc.controller.Failed {
				if tc.controller.Failed[i] != nil {
					t.Errorf("%s:\nfails weren't reset to nil. got %v",
						name, tc.controller.Failed)
				}
			}
		})
	}
}
