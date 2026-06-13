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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

func TestWebHooks_Init(t *testing.T) {
	// GIVEN: a WebHooks and vars for the Init.
	var notifiers shoutrrr.Shoutrrrs
	tests := []struct {
		name                   string
		webhooks               *WebHooks
		nilMap                 bool
		mains                  WebHooksDefaults
		defaults, hardDefaults *Defaults
	}{
		{
			name:     "nil map",
			webhooks: nil, nilMap: true,
		},
		{
			name:     "empty map",
			webhooks: &WebHooks{},
		},
		{
			name: "no mains",
			webhooks: &WebHooks{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false),
			},
		},
		{
			name: "map with nil element and matching main",
			webhooks: &WebHooks{
				"fail": nil,
			},
			mains: WebHooksDefaults{
				"fail": testDefaults(false, false),
			},
		},
		{
			name: "have matching mains",
			webhooks: &WebHooks{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false),
			},
			mains: WebHooksDefaults{
				"fail": testDefaults(false, false),
				"pass": testDefaults(true, false),
			},
		},
		{
			name: "some matching mains",
			webhooks: &WebHooks{
				"fail": testWebHook(true, false, false),
				"pass": testWebHook(false, false, false),
			},
			mains: WebHooksDefaults{
				"other": testDefaults(false, false),
				"pass":  testDefaults(true, false),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !tc.nilMap {
				for i := range *tc.webhooks {
					if (*tc.webhooks)[i] != nil {
						(*tc.webhooks)[i].ID = tc.name + i
					}
				}
			}
			serviceStatus := status.Status{}
			serviceStatus.ServiceInfo.ID = tc.name
			mainCount := 0
			if tc.mains != nil {
				mainCount = len(tc.mains)
			}
			serviceStatus.Init(
				0, 0, mainCount,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			parentInterval := "10s"

			// WHEN: Init is called on it.
			tc.webhooks.Init(
				&serviceStatus,
				Config{
					Root:         tc.mains,
					Defaults:     tc.defaults,
					HardDefaults: tc.hardDefaults,
				},
				&notifiers,
				&parentInterval,
			)

			prefix := fmt.Sprintf("%s\nWebHooks.Init()", packageName)

			// THEN: pointers to those vars are handed out to the WebHook:
			if tc.nilMap {
				if tc.webhooks != nil {
					t.Fatalf(
						"%s map mismatch\ngot:  %v\nwant: nil",
						prefix, *tc.webhooks,
					)
				}
				return
			}

			for _, webhook := range *tc.webhooks {
				// 	Main:
				if webhook.Main == nil {
					t.Errorf(
						"%s .Main of the WebHook[%q] was not initialised\ngot: nil",
						prefix, webhook.ID,
					)
				} else if len(tc.mains) != 0 {
					if wantMain, ok := tc.mains[webhook.ID]; ok {
						gotMain := webhook.Main
						gotMain.Delay = tc.name
						if wantMain.Delay != gotMain.Delay {
							t.Errorf(
								"%s Main[%q] was not handed to the WebHook correctly\ngot:  %v\nwant: %v",
								prefix, webhook.ID,
								webhook.Main, wantMain,
							)
						}
					}
				}

				fieldTests := []test.FieldAssertion{
					{Name: "Defaults", Got: webhook.Defaults, Want: tc.defaults, Mode: test.CompareSamePointer},
					{Name: "HardDefaults", Got: webhook.HardDefaults, Want: tc.hardDefaults, Mode: test.CompareSamePointer},
					{Name: "Status", Got: webhook.ServiceStatus, Want: &serviceStatus, Mode: test.CompareSamePointer},
					{Name: "Notifiers.Shoutrrr", Got: webhook.Notifiers.Shoutrrr, Want: &notifiers, Mode: test.CompareSamePointer},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "WebHook"); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestWebHook_Init(t *testing.T) {
	// GIVEN: a WebHook and vars for the Init.
	webhook := testWebHook(true, false, false)
	var notifiers shoutrrr.Shoutrrrs
	var main Defaults
	var defaults, hardDefaults Defaults
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 1,
		status.ServiceInfo{
			ID: "TestWebHook_Init",
		},
		&dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				WebURL: "https://example.com",
			},
		},
	)

	// WHEN: Init is called on it.
	webhook.init(
		&svcStatus,
		&main,
		Config{
			Root:         nil,
			Defaults:     &defaults,
			HardDefaults: &hardDefaults,
		},
		&notifiers,
		webhook.ParentInterval,
	)
	webhook.ID = "TestInit"

	prefix := fmt.Sprintf("%s\nWebHook.Init()", packageName)

	// THEN: pointers to those vars are handed out to the WebHook:
	fieldTests := []test.FieldAssertion{
		{Name: "Main", Got: webhook.Main, Want: &main, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: webhook.Defaults, Want: &defaults, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: webhook.HardDefaults, Want: &hardDefaults, Mode: test.CompareSamePointer},
		{Name: "Status", Got: webhook.ServiceStatus, Want: &svcStatus, Mode: test.CompareSamePointer},
		{Name: "Notifiers.Shoutrrr", Got: webhook.Notifiers.Shoutrrr, Want: &notifiers, Mode: test.CompareSamePointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "WebHook"); err != nil {
		t.Fatal(err)
	}
}
