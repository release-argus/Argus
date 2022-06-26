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

package webhook

import (
	"encoding/json"
	"testing"

	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestAnnounceSendWithNilChannel(t *testing.T) {
	// GIVEN a WebHook with a nil Announce channel
	whID := "test"
	webhook := WebHook{
		ID:     &whID,
		Failed: nil,
	}

	// WHEN AnnounceSend is called
	webhook.AnnounceSend()

	// THEN the function returns without hanging
}

func TestAnnounceSendWithChannel(t *testing.T) {
	// GIVEN a WebHook with an Announce channel
	whID := "test"
	whFailed := true
	channel := make(chan []byte, 5)
	webhook := WebHook{
		ID:       &whID,
		Failed:   &whFailed,
		Announce: &channel,
	}

	// WHEN AnnounceSend is called
	go webhook.AnnounceSend()

	// THEN the WebHook status is announce to the channel
	msg := <-channel
	var unmarshalled api_types.WebSocketMessage
	json.Unmarshal(msg, &unmarshalled)
	if *unmarshalled.WebHookData["test"].Failed != whFailed {
		t.Errorf("AnnounceSend should have given %v Failed, but got \n%v", whFailed, unmarshalled)
	}
}
