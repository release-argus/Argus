// Copyright [2022] [Hymenaios]
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

package slack

import (
	"encoding/json"
	"testing"

	"github.com/hymenaios-io/Hymenaios/utils"
)

func TestSlackMessage(t *testing.T) {
	var (
		defaults         Slack  = Slack{}
		defaultsUsername string = "defaultUsername"

		main                     Slice  = Slice{}
		mainWantDefaultsURL      string = "https://gotify.main.com/hooks/ZZZZ"
		mainWantDefaultsUsername string = "mainUsername"
		mainWantDefaultsMessage  string = "mainMessage"

		service                   Slice  = Slice{}
		serviceNoDefaultsID       string = "noDefaults"
		serviceNoDefaultsURL      string = "https://gotify.example.com/hooks/ZZZZ"
		serviceNoDefaultsUsername string = "serviceUsername"
		serviceNoDefaultsMessage  string = "serviceMessage"

		serviceWantDefaultsID string = "wantDefaults"

		wantUsername string
		wantText     string
		gotPayload   Payload
	)

	// Check defaults are overridden
	service["noDefaults"] = &Slack{
		ID:           &serviceNoDefaultsID,
		URL:          &serviceNoDefaultsURL,
		Username:     &serviceNoDefaultsUsername,
		Message:      &serviceNoDefaultsMessage,
		Main:         &Slack{},
		Defaults:     &Slack{},
		HardDefaults: &Slack{},
	}
	wantText = "serviceMessage"
	wantUsername = "serviceUsername"
	got := (*service["noDefaults"]).GetPayload("", &utils.ServiceInfo{})
	err := json.Unmarshal(got, &gotPayload)

	if err != nil {
		t.Fatalf("Slack payload failed to unmarshal")
	}
	if wantText != gotPayload.Text {
		t.Fatalf("%s.Payload.Message = %q, want match for %q", *service["noDefaults"].ID, gotPayload.Text, wantText)
	}
	if wantUsername != gotPayload.Username {
		t.Fatalf("%s.Payload.Username = %q, want match for %q", *service["noDefaults"].ID, gotPayload.Username, wantUsername)
	}

	// Check defaults
	defaults = Slack{
		Username: &defaultsUsername,
	}
	main["wantDefaults"] = &Slack{
		URL:      &mainWantDefaultsURL,
		Username: &mainWantDefaultsUsername,
		Message:  &mainWantDefaultsMessage,
	}
	service["wantDefaults"] = &Slack{
		ID:           &serviceWantDefaultsID,
		Main:         main["wantDefaults"],
		Defaults:     &defaults,
		HardDefaults: &Slack{},
	}
	got = (*service["wantDefaults"]).GetPayload("", &utils.ServiceInfo{})
	err = json.Unmarshal(got, &gotPayload)
	wantText = "mainMessage"
	wantUsername = "mainUsername"

	if err != nil {
		t.Fatalf("Slack payload failed to unmarshal")
	}
	if wantText != gotPayload.Text {
		t.Fatalf(`%s.Payload.Message = %q, want match for %q`, *service["wantDefaults"].ID, gotPayload.Text, wantText)
	}
	if wantUsername != gotPayload.Username {
		t.Fatalf(`%s.Payload.Username = %q, want match for %q`, *service["wantDefaults"].ID, gotPayload.Username, wantUsername)
	}

}
