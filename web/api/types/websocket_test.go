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

package types

import (
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/test"
)

func TestWebSocketMessage_String(t *testing.T) {
	// GIVEN a WebSocketMessage.
	tests := map[string]struct {
		websocketMessage WebSocketMessage
		want             string
	}{
		"empty": {
			websocketMessage: WebSocketMessage{},
			want: `
				{
					"page": "",
					"type": ""
				}`,
		},
		"filled": {
			websocketMessage: WebSocketMessage{
				Version: test.IntPtr(1),
				Page:    "foo",
				Type:    "bar",
				SubType: "baz",
				Target:  test.StringPtr("bish"),
				Order: &[]string{
					"zing", "zap", "wallop"},
				ServiceData: &ServiceSummary{
					ID: "summary id"},
				CommandData: map[string]*CommandSummary{
					"alpha": {
						Failed:       test.BoolPtr(true),
						NextRunnable: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)}},
				WebHookData: map[string]*WebHookSummary{
					"omega": {
						Failed:       test.BoolPtr(true),
						NextRunnable: time.Date(2020, 2, 2, 0, 0, 0, 0, time.UTC)}}},
			want: `
				{
					"version": 1,
					"page": "foo",
					"type": "bar",
					"sub_type": "baz",
					"target": "bish",
					"order": [
					"zing",
					"zap",
					"wallop"
					],
					"service_data": {
						"id": "summary id"},
					"command_data": {
						"alpha": {"failed":
						true,
						"next_runnable": "2010-01-01T00:00:00Z"}},
						"webhook_data": {
							"omega": {
								"failed": true,
								"next_runnable": "2020-02-02T00:00:00Z"
							}
						}
				}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN the Data is stringified with String.
			got := tc.websocketMessage.String()

			// THEN the result is as expected.
			tc.want = strings.ReplaceAll(tc.want, "\n", "")
			tc.want = strings.ReplaceAll(tc.want, "\t", "")
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}
