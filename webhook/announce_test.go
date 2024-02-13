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

package webhook

import (
	"encoding/json"
	"testing"
	"time"

	api_type "github.com/release-argus/Argus/web/api/types"
)

func TestWebHook_AnnounceSend(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		nilChannel     bool
		failed         *bool
		timeDifference time.Duration
	}{
		"no channel": {
			nilChannel: true},
		"not tried (failed=nil) does delay by 15s": {
			timeDifference: 15 * time.Second,
			failed:         nil,
		},
		"failed (failed=true) does delay by 15s": {
			timeDifference: 15 * time.Second,
			failed:         boolPtr(true),
		},
		"success (failed=false) does delay by 2*Interval": {
			timeDifference: 24 * time.Minute,
			failed:         boolPtr(false),
		},
	}

	for name, tc := range tests {
		webhook := testWebHook(true, false, false)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook.Failed.Set(webhook.ID, tc.failed)
			webhook.ServiceStatus.AnnounceChannel = nil
			if !tc.nilChannel {
				announceChannel := make(chan []byte, 4)
				webhook.ServiceStatus.AnnounceChannel = &announceChannel
			}

			// WH AnnounceCommand is run
			go webhook.AnnounceSend()

			// THEN the correct response is received
			if webhook.ServiceStatus.AnnounceChannel == nil {
				return
			}
			m := <-*webhook.ServiceStatus.AnnounceChannel
			var parsed api_type.WebSocketMessage
			json.Unmarshal(m, &parsed)

			if parsed.WebHookData[webhook.ID] == nil {
				t.Fatalf("message wasn't for %q\ngot %v",
					webhook.ID, parsed.WebHookData)
			}

			// if they failed status matches
			got := stringifyPointer(parsed.WebHookData[webhook.ID].Failed)
			want := stringifyPointer(webhook.Failed.Get(webhook.ID))
			if got != want {
				t.Errorf("want failed=%s\ngot  failed=%s",
					want, got)
			}

			// next runnable is within expectred range
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.WebHookData[webhook.ID].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf("ran at\n%s\nwant between:\n%s and\n%s\ngot\n%s",
					now, minTime, maxTime, gotTime)
			}
		})
	}
}
