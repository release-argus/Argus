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

package webhook

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestWebHook_AnnounceSend(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name           string
		nilChannel     bool
		failed         *bool
		timeDifference time.Duration
	}{
		{
			name:       "no channel",
			nilChannel: true,
		},
		{
			name:           "not tried (failed=nil) does delay by 15s",
			timeDifference: 15 * time.Second,
			failed:         nil,
		},
		{
			name:           "failed (failed=true) does delay by 15s",
			timeDifference: 15 * time.Second,
			failed:         test.Ptr(true),
		},
		{
			name:           "success (failed=false) does delay by 2*Interval",
			timeDifference: 24 * time.Minute,
			failed:         test.Ptr(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)

			webhook.Failed.Set(webhook.ID, tc.failed)
			webhook.ServiceStatus.AnnounceChannel = nil
			if !tc.nilChannel {
				announceChannel := make(chan []byte, 4)
				webhook.ServiceStatus.AnnounceChannel = announceChannel
			}

			// WH AnnounceCommand is run.
			go webhook.AnnounceSend()

			prefix := fmt.Sprintf("%s\nWebHook.AnnounceSend()", packageName)

			// THEN: the correct response is received.
			if webhook.ServiceStatus.AnnounceChannel == nil {
				return
			}
			m := <-webhook.ServiceStatus.AnnounceChannel
			var parsed apitype.WebSocketMessage
			_ = decode.Unmarshal("json", m, &parsed)

			if parsed.WebHookData[webhook.ID] == nil {
				t.Fatalf(
					"%s message mismatch\ngot:  %+v\nwant: message for service %q",
					prefix, parsed.WebHookData, webhook.ID,
				)
			}

			// if they failed status matches.
			got := test.StringifyPtr(parsed.WebHookData[webhook.ID].Failed)
			want := test.StringifyPtr(webhook.Failed.Get(webhook.ID))
			if got != want {
				t.Errorf(
					"%s 'Failed' part of message didn't match\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// next runnable is within expected range.
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.WebHookData[webhook.ID].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf(
					"%s NextRunnable mismatch for WebHook that ran at:\n     %s\ngot: %s\nwant between:\n     %s\n     %s",
					prefix,
					now,
					gotTime,
					minTime, maxTime,
				)
			}
		})
	}
}

func TestWebHook_AnnounceSend__MarshalError(t *testing.T) {
	// GIVEN: a failing marshal function.
	original := marshalWebhookPayload
	customErr := fmt.Errorf("marshal failed")
	marshalWebhookPayload = func(v any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalWebhookPayload = original })

	// AND: a WebHook with an AnnounceChannel.
	announceChannel := make(chan []byte, 1)
	webhook := testWebHook(true, false, false)
	webhook.ServiceStatus.AnnounceChannel = announceChannel

	// WHEN: AnnounceSend is called.
	webhook.AnnounceSend()

	prefix := fmt.Sprintf("%s\nWebHook.AnnounceSend(marshal error)", packageName)

	// THEN: no message is sent to the announce channel.
	select {
	case msg := <-announceChannel:
		t.Fatalf(
			"%s unexpected message on AnnounceChannel\ngot:  %q\nwant: none",
			prefix, msg,
		)
	default:
	}
}
