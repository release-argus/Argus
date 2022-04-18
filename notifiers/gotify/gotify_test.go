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

package gotify

import (
	"encoding/json"
	"testing"

	"github.com/hymenaios-io/Hymenaios/utils"
)

func TestGotifyMessage(t *testing.T) {
	var (
		hardDefaults         Gotify = Gotify{}
		hardDefaultsPriority        = 1

		defaults      Gotify
		defaultsTitle string = "defaultTitle"

		main                     Slice  = Slice{}
		mainWantDefaultsURL      string = "https://gotify.main.com/hooks/ZZZZ"
		mainWantDefaultsToken    string = "YYYY"
		mainWantDefaultsMessage  string = "mainMessage"
		mainWantDefaultsPriority int    = 1

		service                   Slice  = Slice{}
		serviceNoDefaultsID       string = "NoDefaults"
		serviceNoDefaultsURL      string = "https://gotify.example.com/hooks/ZZZZ"
		serviceNoDefaultsToken    string = "Go"
		serviceNoDefaultsPriority int    = 5
		serviceNoDefaultsTitle    string = "serviceTitle"
		serviceNoDefaultsMessage  string = "serviceMessage"

		serviceWantDefaultsID      string = "wantDefaults"
		serviceWantDefaultsMessage string = "serviceMessage"

		wantMessage  string
		wantTitle    string
		wantPriority int
	)
	hardDefaults.Priority = &hardDefaultsPriority
	defaults.Title = &defaultsTitle
	main["wantDefaults"] = &Gotify{
		URL:      &mainWantDefaultsURL,
		Token:    &mainWantDefaultsToken,
		Message:  &mainWantDefaultsMessage,
		Priority: &mainWantDefaultsPriority,
	}
	service["noDefaults"] = &Gotify{
		ID:           &serviceNoDefaultsID,
		URL:          &serviceNoDefaultsURL,
		Token:        &serviceNoDefaultsToken,
		Priority:     &serviceNoDefaultsPriority,
		Title:        &serviceNoDefaultsTitle,
		Message:      &serviceNoDefaultsMessage,
		Main:         &Gotify{},
		Defaults:     &Gotify{},
		HardDefaults: &hardDefaults,
	}

	// Check defaults are overridden
	wantMessage = "serviceMessage"
	wantTitle = "serviceTitle"
	wantPriority = 5
	got := (*service["noDefaults"]).GetPayload("", "", &utils.ServiceInfo{})
	var payload Payload
	err := json.Unmarshal(got, &payload)

	if err != nil {
		t.Fatalf("Slack payload failed to unmarshal")
	}
	if wantMessage != payload.Message {
		t.Fatalf(`%s.Payload.Message = %v, want match for %q`, *service["noDefaults"].ID, payload.Message, wantMessage)
	}
	if wantPriority != payload.Priority {
		t.Fatalf(`%s.Payload.Priority = %v, want match for %q`, *service["noDefaults"].ID, payload.Priority, wantPriority)
	}
	if wantTitle != payload.Title {
		t.Fatalf(`%s.Payload.Title = %v, want match for %q`, *service["noDefaults"].ID, payload.Title, wantTitle)
	}

	service["wantDefaults"] = &Gotify{
		ID:           &serviceWantDefaultsID,
		Message:      &serviceWantDefaultsMessage,
		Main:         &Gotify{},
		Defaults:     &defaults,
		HardDefaults: &hardDefaults,
	}
	// Check defaults
	wantMessage = "serviceMessage"
	wantTitle = "defaultTitle"
	wantPriority = 1
	got = (*service["wantDefaults"]).GetPayload("", "", &utils.ServiceInfo{})
	err = json.Unmarshal(got, &payload)

	if err != nil {
		t.Fatalf("Slack payload failed to unmarshal")
	}
	if wantMessage != payload.Message {
		t.Fatalf(`%s.Payload.Message = %v, want match for %q`, *service["wantDefaults"].ID, payload.Message, wantMessage)
	}
	if wantPriority != payload.Priority {
		t.Fatalf(`%s.Payload.Priority = %v, want match for %q`, *service["wantDefaults"].ID, payload.Priority, wantPriority)
	}
	if wantTitle != payload.Title {
		t.Fatalf(`%s.Payload.Title = %v, want match for %q`, *service["1"].ID, payload.Title, wantTitle)
	}
}
