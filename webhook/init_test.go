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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

func TestWebhooks_Init(t *testing.T) {
	var notifiers shoutrrr.Shoutrrrs
	tests := []struct {
		name     string
		webhooks *Webhooks
		mains    WebhooksDefaults
	}{
		{
			name:     "nil map",
			webhooks: nil,
		},
		{
			name:     "empty map",
			webhooks: &Webhooks{},
		},
		{
			name: "no mains",
			webhooks: &Webhooks{
				"fail": testWebhook(true, false),
				"pass": testWebhook(false, false),
			},
		},
		{
			name: "map with nil element and matching main",
			webhooks: &Webhooks{
				"fail": nil,
			},
			mains: WebhooksDefaults{
				"fail": testDefaults(t, false),
			},
		},
		{
			name: "have matching mains",
			webhooks: &Webhooks{
				"fail": testWebhook(true, false),
				"pass": testWebhook(false, false),
			},
			mains: WebhooksDefaults{
				"fail": testDefaults(t, false),
				"pass": testDefaults(t, true),
			},
		},
		{
			name: "some matching mains",
			webhooks: &Webhooks{
				"fail": testWebhook(true, false),
				"pass": testWebhook(false, false),
			},
			mains: WebhooksDefaults{
				"other": testDefaults(t, false),
				"pass":  testDefaults(t, true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serviceStatus := status.Status{}
			serviceStatus.Init(
				0, 0, len(tc.mains),
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			parentInterval := "10s"
			defaults, hardDefaults := &Defaults{}, &Defaults{}

			tc.webhooks.Init(
				&serviceStatus,
				Config{
					Root:         tc.mains,
					Defaults:     defaults,
					HardDefaults: hardDefaults,
				},
				&notifiers,
				&parentInterval,
			)

			// A nil map is a no-op.
			if tc.webhooks == nil {
				return
			}
			for _, webhook := range *tc.webhooks {
				t.Run(webhook.ID, func(t *testing.T) {
					t.Parallel()

					if webhook.Main == nil {
						t.Error("Webhook.Main mismatch\ngot:  nil\nwant: non-nil")
					} else if wantMain, ok := tc.mains[webhook.ID]; ok && webhook.Main != wantMain {
						t.Errorf(
							"Webhook.Main mismatch\ngot:  %p\nwant: %p",
							webhook.Main, wantMain,
						)
					}

					fieldTests := []test.FieldAssertion{
						{Name: "Defaults", Got: webhook.Defaults, Want: defaults, Mode: test.CompareSamePointer},
						{Name: "HardDefaults", Got: webhook.HardDefaults, Want: hardDefaults, Mode: test.CompareSamePointer},
						{Name: "ServiceStatus", Got: webhook.ServiceStatus, Want: &serviceStatus, Mode: test.CompareSamePointer},
						{Name: "Notifiers.Shoutrrr", Got: webhook.Notifiers.Shoutrrr, Want: &notifiers, Mode: test.CompareSamePointer},
						{Name: "ParentInterval", Got: webhook.ParentInterval, Want: &parentInterval, Mode: test.CompareSamePointer},
					}
					if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
						t.Fatal(testErr)
					}
				})
			}
		})
	}
}

func TestWebhook_Init(t *testing.T) {
	webhook := testWebhook(true, false)
	var notifiers shoutrrr.Shoutrrrs
	var main Defaults
	var defaults, hardDefaults Defaults
	parentInterval := "10s"
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 1,
		status.ServiceInfo{
			ID: t.Name(),
		},
		&dashboard.Options{
			WebURL: "https://example.com",
		},
	)

	webhook.init(
		&svcStatus,
		&main,
		Config{
			Root:         nil,
			Defaults:     &defaults,
			HardDefaults: &hardDefaults,
		},
		&notifiers,
		&parentInterval,
	)

	fieldTests := []test.FieldAssertion{
		{Name: "Main", Got: webhook.Main, Want: &main, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: webhook.Defaults, Want: &defaults, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: webhook.HardDefaults, Want: &hardDefaults, Mode: test.CompareSamePointer},
		{Name: "ServiceStatus", Got: webhook.ServiceStatus, Want: &svcStatus, Mode: test.CompareSamePointer},
		{Name: "Notifiers.Shoutrrr", Got: webhook.Notifiers.Shoutrrr, Want: &notifiers, Mode: test.CompareSamePointer},
		{Name: "ParentInterval", Got: webhook.ParentInterval, Want: &parentInterval, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Webhook"); testErr != nil {
		t.Fatal(testErr)
	}
}
