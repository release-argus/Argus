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
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestWebhook_AnnounceSend(t *testing.T) {
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
			name:           "not tried (failed=nil) - does delay by 15s",
			timeDifference: 15 * time.Second,
			failed:         nil,
		},
		{
			name:           "failed (failed=true) - does delay by 15s",
			timeDifference: 15 * time.Second,
			failed:         new(true),
		},
		{
			name:           "success (failed=false) - does delay by 2*Interval",
			timeDifference: 24 * time.Minute,
			failed:         new(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebhook(true, false)

			webhook.Failed.Set(webhook.ID, tc.failed)
			webhook.ServiceStatus.AnnounceChannel = nil
			if !tc.nilChannel {
				webhook.ServiceStatus.AnnounceChannel = make(chan []byte, 1)
			}

			webhook.AnnounceSend()

			if tc.nilChannel {
				return
			}
			var m []byte
			select {
			case m = <-webhook.ServiceStatus.AnnounceChannel:
			default:
				t.Fatal("Status.AnnounceChannel message count mismatch\ngot:  0\nwant: 1")
			}
			var parsed apitype.WebSocketMessage
			if err := decode.Unmarshal("json", m, &parsed); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if parsed.WebhookData[webhook.ID] == nil {
				t.Fatal(
					"WebSocketMessage.WebhookData[webhook.ID] mismatch\n" +
						"got:  absent\nwant: present",
				)
			}

			got := test.StringifyPtr(parsed.WebhookData[webhook.ID].Failed)
			want := test.StringifyPtr(webhook.Failed.Get(webhook.ID))
			if got != want {
				t.Errorf(
					"WebSocketMessage.WebhookData[webhook.ID].Failed mismatch\ngot:  %q\nwant: %q",
					got, want,
				)
			}

			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := parsed.WebhookData[webhook.ID].NextRunnable
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf(
					"WebSocketMessage.WebhookData[webhook.ID].NextRunnable mismatch\n"+
						"got:  %s\nwant: between %s and %s",
					gotTime, minTime, maxTime,
				)
			}
		})
	}
}

func TestWebhook_AnnounceSend__marshalError(t *testing.T) {
	original := marshalWebhookPayload
	customErr := errors.New("marshal failed")
	marshalWebhookPayload = func(v any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalWebhookPayload = original })

	announceChannel := make(chan []byte, 1)
	webhook := testWebhook(true, false)
	webhook.ServiceStatus.AnnounceChannel = announceChannel
	releaseStdout := test.CaptureLog(t, logx.Default())

	webhook.AnnounceSend()

	stdoutRegex := `^ERROR: test \(testWebhook\), ` + regexp.QuoteMeta(customErr.Error()) + `\s$`
	if stdout := releaseStdout(); !util.RegexCheck(stdoutRegex, stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, stdoutRegex)
	}

	select {
	case msg := <-announceChannel:
		t.Fatalf("Status.AnnounceChannel message mismatch\ngot:  %q\nwant: absent", msg)
	default:
	}
}
