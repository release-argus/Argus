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

package types

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestServiceSummary_IsZero(t *testing.T) {
	// GIVEN: a ServiceSummary struct.
	tests := []struct {
		name string
		data ServiceSummary
		want bool
	}{
		{
			name: "empty",
			data: ServiceSummary{},
			want: true,
		},
		{
			name: "non-empty/ID",
			data: ServiceSummary{
				ID: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/Name",
			data: ServiceSummary{
				Name: test.Ptr("foo"),
			},
			want: false,
		},
		{
			name: "non-empty/Active",
			data: ServiceSummary{
				Active: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "non-empty/Comment",
			data: ServiceSummary{
				Comment: test.Ptr("foo"),
			},
			want: false,
		},
		{
			name: "non-empty/Type",
			data: ServiceSummary{
				Type: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/WebURL",
			data: ServiceSummary{
				WebURL: test.Ptr("https://example.com"),
			},
			want: false,
		},
		{
			name: "non-empty/Icon",
			data: ServiceSummary{
				Icon: test.Ptr("https://example.com/icon.png"),
			},
			want: false,
		},
		{
			name: "non-empty/IconLinkTo",
			data: ServiceSummary{
				IconLinkTo: test.Ptr("https://example.com/somewhere"),
			},
			want: false,
		},
		{
			name: "non-empty/HasDeployedVersionLookup",
			data: ServiceSummary{
				HasDeployedVersionLookup: test.Ptr(false),
			},
			want: false,
		},
		{
			name: "non-empty/Command",
			data: ServiceSummary{
				Command: test.Ptr(1),
			},
			want: false,
		},
		{
			name: "non-empty/WebHook",
			data: ServiceSummary{
				WebHook: test.Ptr(2),
			},
			want: false,
		},
		{
			name: "non-empty/Status",
			data: ServiceSummary{
				Status: &Status{
					ApprovedVersion: "1.2.3",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Tags",
			data: ServiceSummary{
				Tags: test.Ptr([]string{"foo"}),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: ServiceSummary{
				ID:                       "foo",
				Name:                     test.Ptr("foo"),
				Active:                   test.Ptr(true),
				Comment:                  test.Ptr("foo"),
				Type:                     "foo",
				WebURL:                   test.Ptr("https://example.com"),
				Icon:                     test.Ptr("https://example.com/icon.png"),
				IconLinkTo:               test.Ptr("https://example.com/somewhere"),
				HasDeployedVersionLookup: test.Ptr(false),
				Command:                  test.Ptr(1),
				WebHook:                  test.Ptr(2),
				Status: &Status{
					ApprovedVersion: "1.2.3",
				},
				Tags: test.Ptr([]string{"foo"}),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.data.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nServiceSummary.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestServiceSummary_String(t *testing.T) {
	// GIVEN: a ServiceSummary.
	tests := []struct {
		name    string
		summary *ServiceSummary
		want    string
	}{
		{
			name:    "nil",
			summary: nil,
			want:    "",
		},
		{
			name:    "empty",
			summary: &ServiceSummary{},
			want:    "{}",
		},
		{
			name: "some",
			summary: &ServiceSummary{
				ID:      "foo",
				Name:    test.Ptr("bar"),
				Type:    "github",
				Command: test.Ptr(1),
				WebHook: test.Ptr(2),
			},
			want: `
				{
					"id": "foo",
					"name": "bar",
					"type": "github",
					"command": 1,
					"webhook": 2
				}`,
		},
		{
			name: "full",
			summary: &ServiceSummary{
				ID:                       "bar",
				Name:                     test.Ptr("foo"),
				Active:                   test.Ptr(true),
				Comment:                  test.Ptr("test"),
				Type:                     "url",
				WebURL:                   test.Ptr("https://example.com"),
				Icon:                     test.Ptr("https://example.com/icon.png"),
				IconLinkTo:               test.Ptr("https://release-argus.io"),
				HasDeployedVersionLookup: test.Ptr(true),
				Command:                  test.Ptr(2),
				WebHook:                  test.Ptr(1),
				Status: &Status{
					ApprovedVersion: "1.2.3",
				},
			},
			want: `
				{
					"id": "bar",
					"name": "foo",
					"active": true,
					"comment": "test",
					"type": "url",
					"url": "https://example.com",
					"icon": "https://example.com/icon.png",
					"icon_link_to": "https://release-argus.io",
					"has_deployed_version": true,
					"command": 2,
					"webhook": 1,
					"status": {
						"approved_version": "1.2.3"
					}
				}`,
		},
	}

	// WHEN: String is called on it.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN: the Summary is stringified with String.
			got := tc.summary.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nServiceSummary.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestServiceSummary_RemoveUnchanged(t *testing.T) {
	// GIVEN: two ServiceSummaries.
	tests := []struct {
		name           string
		old, new, want *ServiceSummary
	}{
		{
			name: "compare to nil",
			old:  nil,
			new:  &ServiceSummary{ID: "foo"},
			want: &ServiceSummary{
				ID:     "foo",
				Status: &Status{},
			},
		},
		{
			name: "same id",
			old: &ServiceSummary{
				ID: "foo",
			},
			new: &ServiceSummary{
				ID: "foo",
			},
			want: &ServiceSummary{},
		},
		{
			name: "different id",
			old: &ServiceSummary{
				ID: "foo",
			},
			new: &ServiceSummary{
				ID: "bar",
			},
			want: &ServiceSummary{
				ID: "bar",
			},
		},
		{
			name: "name added",
			old:  &ServiceSummary{},
			new: &ServiceSummary{
				Name: test.Ptr("foo"),
			},
			want: &ServiceSummary{
				Name: test.Ptr("foo"),
			},
		},
		{
			name: "name removed",
			old: &ServiceSummary{
				Name: test.Ptr("foo"),
			},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				Name: test.Ptr(""),
			},
		},
		{
			name: "same name",
			old: &ServiceSummary{
				Name: test.Ptr("foo"),
			},
			new: &ServiceSummary{
				Name: test.Ptr("foo"),
			},
			want: &ServiceSummary{},
		},
		{
			name: "different name",
			old: &ServiceSummary{
				Name: test.Ptr("foo"),
			},
			new: &ServiceSummary{
				Name: test.Ptr("bar"),
			},
			want: &ServiceSummary{
				Name: test.Ptr("bar"),
			},
		},
		{
			name: "same active",
			old: &ServiceSummary{
				Active: test.Ptr(false),
			},
			new: &ServiceSummary{
				Active: test.Ptr(false),
			},
			want: &ServiceSummary{},
		},
		{
			name: "different active",
			old: &ServiceSummary{
				Active: test.Ptr(true),
			},
			new: &ServiceSummary{
				Active: test.Ptr(false),
			},
			want: &ServiceSummary{
				Active: test.Ptr(false),
			},
		},
		{
			name: "comment added",
			old:  &ServiceSummary{},
			new: &ServiceSummary{
				Comment: test.Ptr("foo"),
			},
			want: &ServiceSummary{
				Comment: test.Ptr("foo"),
			},
		},
		{
			name: "comment changed",
			old: &ServiceSummary{
				Comment: test.Ptr("foo"),
			},
			new: &ServiceSummary{
				Comment: test.Ptr("bar"),
			},
			want: &ServiceSummary{
				Comment: test.Ptr("bar"),
			},
		},
		{
			name: "comment removed",
			old: &ServiceSummary{
				Comment: test.Ptr("foo"),
			},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				Comment: test.Ptr(""),
			},
		},
		{
			name: "same type",
			old: &ServiceSummary{
				Type: "github",
			},
			new: &ServiceSummary{
				Type: "github",
			},
			want: &ServiceSummary{},
		},
		{
			name: "different type",
			old: &ServiceSummary{
				Type: "github",
			},
			new: &ServiceSummary{
				Type: "url",
			},
			want: &ServiceSummary{
				Type: "url",
			},
		},
		{
			name: "same icon",
			old: &ServiceSummary{
				Icon: test.Ptr("https://example.com/icon.png"),
			},
			new: &ServiceSummary{
				Icon: test.Ptr("https://example.com/icon.png"),
			},
			want: &ServiceSummary{},
		},
		{
			name: "different icon",
			old: &ServiceSummary{
				Icon: test.Ptr("https://example.com/icon.png"),
			},
			new: &ServiceSummary{
				Icon: test.Ptr("https://example.com/icon2.png"),
			},
			want: &ServiceSummary{
				Icon: test.Ptr("https://example.com/icon2.png"),
			},
		},
		{
			name: "same icon_link_to",
			old: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io"),
			},
			new: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io"),
			},
			want: &ServiceSummary{},
		},
		{
			name: "different icon_link_to",
			old: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io"),
			},
			new: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io/other"),
			},
			want: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io/other"),
			},
		},
		{
			name: "same has_deployed_version_lookup",
			old: &ServiceSummary{
				HasDeployedVersionLookup: test.Ptr(true),
			},
			new: &ServiceSummary{
				HasDeployedVersionLookup: test.Ptr(true),
			},
			want: &ServiceSummary{},
		},
		{
			name: "different has_deployed_version_lookup",
			old: &ServiceSummary{
				HasDeployedVersionLookup: test.Ptr(true),
			},
			new: &ServiceSummary{
				HasDeployedVersionLookup: test.Ptr(false),
			},
			want: &ServiceSummary{
				HasDeployedVersionLookup: test.Ptr(false),
			},
		},
		{
			name: "same approved_version",
			old: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "1.2.3",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "1.2.3",
				},
			},
			want: &ServiceSummary{},
		},
		{
			name: "different approved_version",
			old: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "1.2.3",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "4.5.6",
				},
			},
			want: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "4.5.6",
				},
			},
		},
		{
			name: "same deployed_version/bare",
			old: &ServiceSummary{
				Status: &Status{
					DeployedVersion: "1.2.3",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					DeployedVersion: "1.2.3",
				},
			},
			want: &ServiceSummary{},
		},
		{
			name: "same deployed_version/different timestamps ignored",
			old: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-01-01T00:00:00Z",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
			want: &ServiceSummary{},
		},
		{
			name: "different deployed_version",
			old: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-01-01T00:00:00Z",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
			want: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
		},
		{
			name: "same latest_version/bare",
			old: &ServiceSummary{
				Status: &Status{
					LatestVersion: "1.2.3",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					LatestVersion: "1.2.3",
				},
			},
			want: &ServiceSummary{},
		},
		{
			name: "same latest_version/different timestamps ignored",
			old: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "1.2.3",
					LatestVersionTimestamp: "2020-01-01T00:00:00Z",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "1.2.3",
					LatestVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
			want: &ServiceSummary{},
		},
		{
			name: "different latest_version",
			old: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "1.2.3",
					LatestVersionTimestamp: "2020-01-01T00:00:00Z",
				},
			},
			new: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "4.5.6",
					LatestVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
			want: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "4.5.6",
					LatestVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
		},
		{
			name: "multiple differences",
			old: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io"),
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-01-01T00:00:00Z",
				},
			},
			new: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io/other"),
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
			want: &ServiceSummary{
				IconLinkTo: test.Ptr("https://release-argus.io/other"),
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z",
				},
			},
		},
		{
			name: "tags added",
			old:  &ServiceSummary{},
			new: &ServiceSummary{
				Tags: test.Ptr([]string{"foo"}),
			},
			want: &ServiceSummary{
				Tags: test.Ptr([]string{"foo"}),
			},
		},
		{
			name: "tags removed",
			old: &ServiceSummary{
				Tags: test.Ptr([]string{"foo"}),
			},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				Tags: test.Ptr([]string{}),
			},
		},
		{
			name: "same tags",
			old: &ServiceSummary{
				Tags: test.Ptr([]string{"foo"}),
			},
			new: &ServiceSummary{
				Tags: test.Ptr([]string{"foo"}),
			},
			want: &ServiceSummary{},
		},
		{
			name: "different tags",
			old: &ServiceSummary{
				Tags: test.Ptr([]string{"foo"}),
			},
			new: &ServiceSummary{
				Tags: test.Ptr([]string{"bar"}),
			},
			want: &ServiceSummary{
				Tags: test.Ptr([]string{"bar"}),
			},
		},
		{
			name: "command added",
			old:  &ServiceSummary{},
			new: &ServiceSummary{
				Command: test.Ptr(1),
			},
			want: &ServiceSummary{
				Command: test.Ptr(1),
			},
		},
		{
			name: "command removed",
			old: &ServiceSummary{
				Command: test.Ptr(1),
			},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				Command: test.Ptr(0),
			},
		},
		{
			name: "same command",
			old: &ServiceSummary{
				Command: test.Ptr(1),
			},
			new: &ServiceSummary{
				Command: test.Ptr(1),
			},
			want: &ServiceSummary{
				Command: nil,
			},
		},
		{
			name: "webhook added",
			old:  &ServiceSummary{},
			new: &ServiceSummary{
				WebHook: test.Ptr(1),
			},
			want: &ServiceSummary{
				WebHook: test.Ptr(1),
			},
		},
		{
			name: "webhook removed",
			old: &ServiceSummary{
				WebHook: test.Ptr(1),
			},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				WebHook: test.Ptr(0),
			},
		},
		{
			name: "same webhook",
			old: &ServiceSummary{
				WebHook: test.Ptr(1),
			},
			new: &ServiceSummary{
				WebHook: test.Ptr(1),
			},
			want: &ServiceSummary{
				WebHook: nil,
			},
		},
	}

	initialiseFields := func(instance *ServiceSummary) {
		if instance.Status == nil {
			instance.Status = &Status{}
		}
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Give them non-nil Status, Command and WebHook.
			if tc.old != nil {
				initialiseFields(tc.old)
			}
			if tc.new != nil {
				initialiseFields(tc.new)
			}

			// WHEN: RemoveUnchanged is called, comparing new to old.
			tc.new.RemoveUnchanged(tc.old)

			prefix := fmt.Sprintf(
				"%s\nServiceSummary.RemoveUnchanged(%+v)",
				packageName, tc.old,
			)

			// THEN: the values that are unchanged are removed.
			if gotStr, wantStr := tc.new.String(), tc.want.String(); gotStr != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults Defaults
		want     bool
	}{
		{
			name: "empty",
			defaults: Defaults{
				Service: ServiceDefaults{},
				Notify:  Notifiers{},
				WebHook: WebHook{},
			},
			want: true,
		},
		{
			name: "non-empty/Service",
			defaults: Defaults{
				Service: ServiceDefaults{
					Comment: "a",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Notify",
			defaults: Defaults{
				Notify: Notifiers{
					"a": &Notify{
						Type: "a",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/WebHook",
			defaults: Defaults{
				WebHook: WebHook{
					Type: "a",
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: Defaults{
				Service: ServiceDefaults{
					Comment: "a",
				},
				Notify: Notifiers{
					"a": &Notify{
						Type: "a",
					},
				},
				WebHook: WebHook{
					Type: "a",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_String(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
		want     string
	}{
		{
			name:     "nil",
			defaults: nil,
			want:     "",
		},
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     `{}`,
		},
		{
			name: "all types",
			defaults: &Defaults{
				Service: ServiceDefaults{
					LatestVersion: LatestVersionDefaults{
						AccessToken: "foo",
					},
				},
				Notify: Notifiers{
					"gotify": &Notify{
						URLFields: map[string]string{
							"url": "https://gotify.example.com",
						},
					},
				},
				WebHook: WebHook{
					Secret: "bar",
				},
			},
			want: `
				{
					"service": {
						"latest_version": {
							"access_token": "foo"
						}
					},
					"notify": {
						"gotify": {
							"url_fields": {
								"url": "https://gotify.example.com"
							}
						}
					},
					"webhook": {
						"secret": "bar"
					}
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the Defaults are stringified with String.
			got := tc.defaults.String()

			// THEN: the result is as expected.
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestSettings_IsZero(t *testing.T) {
	// GIVEN: Settings.
	tests := []struct {
		name     string
		settings Settings
		want     bool
	}{
		{
			name: "empty",
			settings: Settings{
				Log: LogSettings{},
				Web: WebSettings{},
			},
			want: true,
		},
		{
			name: "non-empty/Log",
			settings: Settings{
				Log: LogSettings{
					Level: "DEBUG",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Web",
			settings: Settings{
				Web: WebSettings{
					ListenPort: "9001",
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			settings: Settings{
				Log: LogSettings{
					Level: "DEBUG",
				},
				Web: WebSettings{
					ListenPort: "9001",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.settings.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nSettings.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLogSettings_IsZero(t *testing.T) {
	// GIVEN: LogSettings.
	tests := []struct {
		name        string
		logSettings LogSettings
		want        bool
	}{
		{
			name: "empty",
			logSettings: LogSettings{
				Timestamps: nil,
				Level:      "",
			},
			want: true,
		},
		{
			name: "non-empty/Timestamps",
			logSettings: LogSettings{
				Timestamps: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "non-empty/Level",
			logSettings: LogSettings{
				Level: "DEBUG",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			logSettings: LogSettings{
				Timestamps: test.Ptr(true),
				Level:      "DEBUG",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.logSettings.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLogSettings.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebSettings_IsZero(t *testing.T) {
	// GIVEN: WebSettings.
	tests := []struct {
		name        string
		webSettings WebSettings
		want        bool
	}{
		{
			name: "empty",
			webSettings: WebSettings{

				ListenHost:  "",
				ListenPort:  "",
				CertFile:    "",
				KeyFile:     "",
				RoutePrefix: "",
			},
			want: true,
		},
		{
			name: "non-empty/ListenHost",
			webSettings: WebSettings{
				ListenHost: "127.0.0.1",
			},
			want: false,
		},
		{
			name: "non-empty/ListenPort",
			webSettings: WebSettings{
				ListenPort: "9001",
			},
			want: false,
		},
		{
			name: "non-empty/CertFile",
			webSettings: WebSettings{
				CertFile: "file.pem",
			},
			want: false,
		},
		{
			name: "non-empty/KeyFile",
			webSettings: WebSettings{
				KeyFile: "file.pem",
			},
			want: false,
		},
		{
			name: "non-empty/RoutePrefix",
			webSettings: WebSettings{
				RoutePrefix: "/",
			},
			want: false,
		},
		{
			name: "non-empty/DisabledRoutes",
			webSettings: WebSettings{
				DisabledRoutes: []string{"route1", "route2"},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			webSettings: WebSettings{
				ListenHost:     "127.0.0.1",
				ListenPort:     "9001",
				CertFile:       "file.pem",
				KeyFile:        "file.pem",
				RoutePrefix:    "/",
				DisabledRoutes: []string{"route1", "route2"},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.webSettings.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nWebSettings.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestNotifiers_String(t *testing.T) {
	// GIVEN: Notifiers.
	tests := []struct {
		name      string
		notifiers *Notifiers
		want      string
	}{
		{
			name:      "nil",
			notifiers: nil,
			want:      "",
		},
		{
			name:      "empty",
			notifiers: &Notifiers{},
			want:      "{}",
		},
		{
			name: "one",
			notifiers: &Notifiers{
				"0": {
					ID:   "foo",
					Type: "discord",
					Options: map[string]string{
						"message": "hello world",
					},
					URLFields: map[string]string{
						"username": "bing",
					},
					Params: map[string]string{
						"devices": "bang",
					},
				},
			},
			want: `
				{
					"0": {
						"name": "foo",
						"type": "discord",
						"options": {
							"message": "hello world"
						},
						"url_fields": {
							"username": "bing"
						},
						"params": {
							"devices": "bang"
						}
					}
				}`,
		},
		{
			name: "multiple",
			notifiers: &Notifiers{
				"0": {
					ID:   "foo",
					Type: "discord",
					Options: map[string]string{
						"message": "hello world",
					},
					URLFields: map[string]string{
						"username": "bing",
					},
					Params: map[string]string{
						"devices": "bang",
					},
				},
				"other": {
					Type: "gotify",
				},
			},
			want: `
				{
					"0": {
						"name": "foo",
						"type": "discord",
						"options": {
							"message": "hello world"
						},
						"url_fields": {
							"username": "bing"
						},
						"params": {
							"devices": "bang"
						}
					},
					"other": {
						"type": "gotify"
					}
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the Notifiers is stringified with String.
			got := tc.notifiers.String()

			// THEN: the result is as expected.
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf(
					"%sNotifiers.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestNotifiers_Flatten(t *testing.T) {
	// GIVEN: a Notifiers.
	tests := []struct {
		name   string
		notify *Notifiers
		want   *[]Notify
	}{
		{
			name:   "nil",
			notify: nil,
			want:   nil,
		},
		{
			name:   "empty",
			notify: &Notifiers{},
			want:   &[]Notify{},
		},
		{
			name: "ordered",
			notify: &Notifiers{
				"zulu": &Notify{
					URLFields: map[string]string{
						"port": "alpha",
					},
					Params: map[string]string{
						"hosts": "bravo",
					},
				},
				"yankee": &Notify{
					URLFields: map[string]string{
						"path": "charlie",
					},
					Params: map[string]string{
						"rooms": "delta",
					},
				},
			},
			want: &[]Notify{
				{ID: "yankee",
					URLFields: map[string]string{
						"path": "charlie",
					},
					Params: map[string]string{
						"rooms": "delta",
					},
				},
				{ID: "zulu",
					URLFields: map[string]string{
						"port": "alpha",
					},
					Params: map[string]string{
						"hosts": "bravo",
					},
				},
			},
		},
		{
			name: "ordered and censored",
			notify: &Notifiers{
				"hotel": &Notify{
					URLFields: map[string]string{
						"port":  "alpha",
						"altid": "echo",
					},
					Params: map[string]string{
						"hosts":   "bravo",
						"devices": "foxtrot",
					},
				},
				"golf": &Notify{
					URLFields: map[string]string{
						"path":   "charlie",
						"botkey": "india",
					},
					Params: map[string]string{
						"rooms": "delta",
					},
				},
			},
			want: &[]Notify{
				{ID: "golf",
					URLFields: map[string]string{
						"path":   "charlie",
						"botkey": util.SecretValue,
					},
					Params: map[string]string{
						"rooms": "delta",
					},
				},
				{ID: "hotel",
					URLFields: map[string]string{
						"port":  "alpha",
						"altid": util.SecretValue,
					},
					Params: map[string]string{
						"hosts":   "bravo",
						"devices": util.SecretValue,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Flatten is called on it.
			got := tc.notify.Flatten()

			// THEN: nil Notifiers are kept.
			if tc.notify == nil && tc.want == nil {
				return
			}

			// AND: defined fields are censored as expected.
			for i := range *tc.want {
				prefix := fmt.Sprintf(
					"%s\nNotifiers.Flatten() Notify[%q]",
					packageName, i,
				)

				if gotID, wantID := got[i].ID, (*tc.want)[i].ID; gotID != wantID {
					t.Errorf(
						"%s .ID mismatch\ngot:  %q\nwant: %q",
						prefix, gotID, wantID,
					)
				}
				if testErr := test.AssertMapEqual(
					t,
					got[i].URLFields,
					(*tc.want)[i].URLFields,
					prefix,
					"URLFields",
				); testErr != nil {
					t.Error(testErr)
				}
				if testErr := test.AssertMapEqual(
					t,
					got[i].Params,
					(*tc.want)[i].Params,
					prefix,
					"Params",
				); testErr != nil {
					t.Error(testErr)
				}
			}
		})
	}
}

func TestNotify_Censor(t *testing.T) {
	// GIVEN: a Notify.
	tests := []struct {
		name         string
		notify, want *Notify
	}{
		{
			name:   "nil",
			notify: nil,
			want:   nil,
		},
		{
			name: "url_fields",
			notify: &Notify{
				URLFields: map[string]string{
					"altid":    "alpha",
					"apikey":   "bravo",
					"botkey":   "charlie",
					"password": "delta",
					"token":    "echo",
					"tokena":   "foxtrot",
					"tokenb":   "golf",
				},
			},
			want: &Notify{
				URLFields: map[string]string{
					"altid":    util.SecretValue,
					"apikey":   util.SecretValue,
					"botkey":   util.SecretValue,
					"password": util.SecretValue,
					"token":    util.SecretValue,
					"tokena":   util.SecretValue,
					"tokenb":   util.SecretValue,
				},
			},
		},
		{
			name: "params",
			notify: &Notify{
				Params: map[string]string{
					"devices": "foo",
				},
			},
			want: &Notify{
				Params: map[string]string{
					"devices": util.SecretValue,
				},
			},
		},
		{
			name: "all censorable/only",
			notify: &Notify{
				URLFields: map[string]string{
					"altid":    "alpha",
					"apikey":   "bravo",
					"botkey":   "charlie",
					"password": "delta",
					"token":    "echo",
					"tokena":   "foxtrot",
					"tokenb":   "golf",
				},
				Params: map[string]string{
					"devices": "hotel",
				},
			},
			want: &Notify{
				URLFields: map[string]string{
					"altid":    util.SecretValue,
					"apikey":   util.SecretValue,
					"botkey":   util.SecretValue,
					"password": util.SecretValue,
					"token":    util.SecretValue,
					"tokena":   util.SecretValue,
					"tokenb":   util.SecretValue,
				},
				Params: map[string]string{
					"devices": util.SecretValue,
				},
			},
		},
		{
			name: "all censorable/plus non-censored",
			notify: &Notify{
				URLFields: map[string]string{
					"altid":    "alpha",
					"apikey":   "bravo",
					"botkey":   "charlie",
					"password": "delta",
					"token":    "echo",
					"tokena":   "foxtrot",
					"tokenb":   "golf",
					"port":     "hotel",
					"username": "india",
				},
				Params: map[string]string{
					"devices": "juliette",
					"rooms":   "kilo",
					"events":  "lima",
				},
			},
			want: &Notify{
				URLFields: map[string]string{
					"altid":    util.SecretValue,
					"apikey":   util.SecretValue,
					"botkey":   util.SecretValue,
					"password": util.SecretValue,
					"token":    util.SecretValue,
					"tokena":   util.SecretValue,
					"tokenb":   util.SecretValue,
					"port":     "hotel",
					"username": "india",
				},
				Params: map[string]string{
					"devices": util.SecretValue,
					"rooms":   "kilo",
					"events":  "lima",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Censor is called on it.
			tc.notify.Censor()

			prefix := fmt.Sprintf("%s\nNotify.Censor()", packageName)

			// THEN: nil Notifiers are kept.
			if tc.notify == tc.want {
				return
			}

			// AND: defined fields are censored as expected.
			if testErr := test.AssertMapEqual(
				t,
				tc.want.URLFields,
				tc.notify.URLFields,
				prefix,
				"URLFields",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				tc.want.Params,
				tc.notify.Params,
				prefix,
				"Params",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestService_String(t *testing.T) {
	// GIVEN: a Service.
	tests := []struct {
		name  string
		input *Service
		want  string
	}{
		{
			name:  "nil",
			input: nil,
			want:  "",
		},
		{
			name:  "empty",
			input: &Service{},
			want:  `{}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the Defaults are stringified with String.
			got := tc.input.String()

			// THEN: the result is as expected.
			tc.want = strings.ReplaceAll(tc.want, "\n", "")
			if got != tc.want {
				t.Errorf(
					"%s\nService.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestServiceDefaults_IsZero(t *testing.T) {
	// GIVEN: ServiceDefaults.
	tests := []struct {
		name     string
		defaults ServiceDefaults
		want     bool
	}{
		{
			name: "empty",
			defaults: ServiceDefaults{
				Comment:               "",
				Options:               ServiceOptions{},
				LatestVersion:         LatestVersionDefaults{},
				Notify:                []string{},
				Command:               Commands{},
				WebHook:               []string{},
				DeployedVersionLookup: DeployedVersionLookupDefaults{},
				Dashboard:             DashboardOptions{},
			},
			want: true,
		},
		{
			name: "non-empty/Comment",
			defaults: ServiceDefaults{
				Comment: "abc",
			},
			want: false,
		},
		{
			name: "non-empty/Options",
			defaults: ServiceDefaults{
				Options: ServiceOptions{
					Interval: "1s",
				},
			},
			want: false,
		},
		{
			name: "non-empty/LatestVersion",
			defaults: ServiceDefaults{
				LatestVersion: LatestVersionDefaults{
					AccessToken: "hi",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Notify",
			defaults: ServiceDefaults{
				Notify: []string{"a"},
			},
			want: false,
		},
		{
			name: "non-empty/Command",
			defaults: ServiceDefaults{
				Command: Commands{
					{"ls"},
				},
			},
			want: false,
		},
		{
			name: "non-empty/WebHook",
			defaults: ServiceDefaults{
				WebHook: []string{"a"},
			},
			want: false,
		},
		{
			name: "non-empty/DeployedVersionLookup",
			defaults: ServiceDefaults{
				DeployedVersionLookup: DeployedVersionLookupDefaults{
					Method: "GET",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Dashboard",
			defaults: ServiceDefaults{
				Dashboard: DashboardOptions{
					Icon: "a",
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: ServiceDefaults{
				Comment: "abc",
				Options: ServiceOptions{
					Interval: "1s",
				},
				LatestVersion: LatestVersionDefaults{
					AccessToken: "hi",
				},
				Notify: []string{"a"},
				Command: Commands{
					{"ls"},
				},
				WebHook: []string{"a"},
				DeployedVersionLookup: DeployedVersionLookupDefaults{
					Method: "GET",
				},
				Dashboard: DashboardOptions{
					Icon: "a",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nServiceDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestServiceOptions_IsZero(t *testing.T) {
	// GIVEN: ServiceOptions.
	tests := []struct {
		name    string
		options ServiceOptions
		want    bool
	}{
		{
			name: "empty",
			options: ServiceOptions{
				Active:             nil,
				Interval:           "",
				SemanticVersioning: nil,
			},
			want: true,
		},
		{
			name: "non-empty/Active",
			options: ServiceOptions{
				Active: test.Ptr(false),
			},
			want: false,
		},
		{
			name: "non-empty/Interval",
			options: ServiceOptions{
				Interval: "1s",
			},
			want: false,
		},
		{
			name: "non-empty/SemanticVersioning",
			options: ServiceOptions{
				SemanticVersioning: test.Ptr(false),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			options: ServiceOptions{
				Active:             test.Ptr(false),
				Interval:           "1s",
				SemanticVersioning: test.Ptr(false),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.options.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nServiceOptions.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDashboardOptions_IsZero(t *testing.T) {
	// GIVEN: DashboardOptions.
	tests := []struct {
		name    string
		options DashboardOptions
		want    bool
	}{
		{
			name: "empty",
			options: DashboardOptions{
				AutoApprove: nil,
				Icon:        "",
				IconLinkTo:  "",
				WebURL:      "",
				Tags:        []string{},
			},
			want: true,
		},
		{
			name: "non-empty/AutoApprove",
			options: DashboardOptions{
				AutoApprove: test.Ptr(false),
			},
			want: false,
		},
		{
			name: "non-empty/Icon",
			options: DashboardOptions{
				Icon: "a",
			},
			want: false,
		},
		{
			name: "non-empty/IconLinkTo",
			options: DashboardOptions{
				IconLinkTo: "a",
			},
			want: false,
		},
		{
			name: "non-empty/WebURL",
			options: DashboardOptions{
				WebURL: "a",
			},
			want: false,
		},
		{
			name: "non-empty/Tag",
			options: DashboardOptions{
				Tags: []string{"a"},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			options: DashboardOptions{
				AutoApprove: test.Ptr(false),
				Icon:        "a",
				IconLinkTo:  "a",
				WebURL:      "a",
				Tags:        []string{"a"},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.options.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nDashboardOptions.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLatestVersion_String(t *testing.T) {
	// GIVEN: a LatestVersion.
	tests := []struct {
		name  string
		input *LatestVersion
		want  string
	}{
		{
			name:  "nil",
			input: nil,
			want:  "",
		},
		{
			name:  "empty",
			input: &LatestVersion{},
			want:  `{}`,
		},
		{
			name: "filled",
			input: &LatestVersion{
				Type:              "github",
				URL:               "owner/repo",
				AccessToken:       util.SecretValue,
				AllowInvalidCerts: test.Ptr(true),
				UsePreRelease:     test.Ptr(false),
				URLCommands: URLCommands{
					{Type: "replace", Old: "this", New: "withThis"},
					{Type: "split", Text: "splitThis", Index: test.Ptr(8)},
					{Type: "regex", Regex: `([0-9.]+)`},
				},
				Require: &LatestVersionRequire{
					RegexContent: ".*",
				},
			},
			want: `
				{
					"type": "github",
					"url": "owner/repo",
					"access_token": ` + secretValueMarshalled + `,
					"allow_invalid_certs": true,
					"use_prerelease": false,
					"url_commands": [
						{"type": "replace","new": "withThis","old": "this"},
						{"type": "split","index": 8,"text": "splitThis"},
						{"type": "regex","regex": "([0-9.]+)"}
					],
					"require": {
						"regex_content": ".*"
					}
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN: the LatestVersion is stringified with String.
			got := tc.input.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLatestVersion.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLatestVersionDefaults_IsZero(t *testing.T) {
	// GIVEN: LatestVersionDefaults.
	tests := []struct {
		name     string
		defaults LatestVersionDefaults
		want     bool
	}{
		{
			name: "empty",
			defaults: LatestVersionDefaults{
				URL:               "",
				AccessToken:       "",
				AllowInvalidCerts: nil,
				UsePreRelease:     nil,
				Require:           nil,
			},
			want: true,
		},
		{
			name: "non-empty/URL",
			defaults: LatestVersionDefaults{
				URL: "a",
			},
			want: false,
		},
		{
			name: "non-empty/AccessToken",
			defaults: LatestVersionDefaults{
				AccessToken: "a",
			},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			defaults: LatestVersionDefaults{
				AllowInvalidCerts: test.Ptr(false),
			},
			want: false,
		},
		{
			name: "non-empty/UsePreRelease",
			defaults: LatestVersionDefaults{
				UsePreRelease: test.Ptr(false),
			},
			want: false,
		},
		{
			name: "non-empty/Require",
			defaults: LatestVersionDefaults{
				Require: &LatestVersionRequireDefaults{
					Docker: RequireDockerDefaults{
						Tag: "a",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			defaults: LatestVersionDefaults{
				URL:               "a",
				AccessToken:       "a",
				AllowInvalidCerts: test.Ptr(false),
				UsePreRelease:     test.Ptr(false),
				Require: &LatestVersionRequireDefaults{
					Docker: RequireDockerDefaults{
						Tag: "a",
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLatestVersionDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLatestVersionRequire_String(t *testing.T) {
	// GIVEN: a LatestVersionRequire.
	tests := []struct {
		name  string
		input *LatestVersionRequire
		want  string
	}{
		{
			name:  "nil",
			input: nil,
			want:  "",
		},
		{
			name:  "empty",
			input: &LatestVersionRequire{},
			want:  `{}`,
		},
		{
			name: "filled",
			input: &LatestVersionRequire{
				Command: []string{"echo", "hello"},
				Docker: &RequireDocker{
					Type:     "hub",
					Image:    "test/app",
					Tag:      "{{ version }}",
					Username: "user",
					Token:    util.SecretValue,
				},
				RegexContent: ".*",
				RegexVersion: `([0-9.]+)`,
			},
			want: `
				{
					"command": ["echo","hello"],
					"docker": {
						"type": "hub",
						"image": "test/app",
						"tag": "{{ version }}",
						"username": "user",
						"token": ` + secretValueMarshalled + `
					},
					"regex_content": ".*",
					"regex_version": "([0-9.]+)"
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN: the LatestVersionRequire is stringified with String.
			got := tc.input.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLatestVersionRequire.String() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLatestVersionRequireDefaults_IsZero(t *testing.T) {
	// GIVEN: LatestVersionRequireDefaults.
	tests := []struct {
		name     string
		defaults LatestVersionRequireDefaults
		want     bool
	}{
		{
			name: "empty",
			defaults: LatestVersionRequireDefaults{
				Docker: RequireDockerDefaults{},
			},
			want: true,
		},
		{
			name: "non-empty",
			defaults: LatestVersionRequireDefaults{
				Docker: RequireDockerDefaults{
					Tag: "a",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLatestVersionRequireDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLatestVersionRequireDefaults_String(t *testing.T) {
	// GIVEN: a LatestVersionRequireDefaults.
	tests := []struct {
		name string
		lvRD *LatestVersionRequireDefaults
		want string
	}{
		{
			name: "nil",
			lvRD: nil,
			want: "",
		},
		{
			name: "empty",
			lvRD: &LatestVersionRequireDefaults{},
			want: `{}`,
		},
		{
			name: "filled",
			lvRD: &LatestVersionRequireDefaults{
				Docker: RequireDockerDefaults{
					Type: "ghcr",
					Registry: RequireDockerRegistriesDefaults{
						GHCR: &RequireDockerRegistryDefaultsToken{
							RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
								Token: "tokenForGHCR",
							},
						},
						Hub: &RequireDockerCheckRegistryDefaultsTokenWithUsername{
							RequireDockerRegistryDefaultsAuthWithUsername: RequireDockerRegistryDefaultsAuthWithUsername{
								Username: "userForHub",
								RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
									Token: "tokenForHub",
								},
							},
						},
						Quay: &RequireDockerRegistryDefaultsToken{
							RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
								Token: "tokenForQuay",
							},
						},
					},
				},
			},
			want: `
				{
					"docker": {
						"type": "ghcr",
						"registry": {
							"ghcr": {
								"auth": {
									"token": "tokenForGHCR"
								}
							},
							"hub": {
								"auth": {
									"username": "userForHub",
									"token": "tokenForHub"
								}
							},
							"quay": {
								"auth": {
									"token": "tokenForQuay"
								}
							}
						}
					}
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN: the LatestVersionRequireDefaults are stringified with String.
			got := tc.lvRD.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLatestVersionRequireDefaults.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerRegistriesDefaults_IsZero(t *testing.T) {
	// GIVEN: RequireDockerRegistriesDefaults.
	tests := []struct {
		name     string
		defaults RequireDockerRegistriesDefaults
		want     bool
	}{
		{
			name:     "empty",
			defaults: RequireDockerRegistriesDefaults{},
			want:     true,
		},
		{
			name: "non-empty",
			defaults: RequireDockerRegistriesDefaults{
				GHCR: &RequireDockerRegistryDefaultsToken{
					RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
						Token: "token",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerRegistriesDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerRegistryDefaultsAuth_IsZero(t *testing.T) {
	// GIVEN: RequireDockerRegistryDefaultsAuth.
	tests := []struct {
		name     string
		defaults RequireDockerRegistryDefaultsAuth
		want     bool
	}{
		{
			name:     "empty",
			defaults: RequireDockerRegistryDefaultsAuth{},
			want:     true,
		},
		{
			name: "non-empty",
			defaults: RequireDockerRegistryDefaultsAuth{
				Token: "t",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerRegistryDefaultsAuth.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerRegistryDefaultsAuthWithUsername_IsZero(t *testing.T) {
	// GIVEN: RequireDockerRegistryDefaultsAuthWithUsername.
	tests := []struct {
		name     string
		defaults RequireDockerRegistryDefaultsAuthWithUsername
		want     bool
	}{
		{
			name: "empty",
			defaults: RequireDockerRegistryDefaultsAuthWithUsername{
				RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{},
			},
			want: true,
		},
		{
			name: "non-empty/Token",
			defaults: RequireDockerRegistryDefaultsAuthWithUsername{
				RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
					Token: "t",
				},
			},
		},
		{
			name: "non-empty/Username",
			defaults: RequireDockerRegistryDefaultsAuthWithUsername{
				Username: "u",
			},
		},
		{
			name: "non-empty/all",
			defaults: RequireDockerRegistryDefaultsAuthWithUsername{
				Username: "u",
				RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
					Token: "t",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerRegistryDefaultsAuthWithUsername.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerRegistryDefaultsToken_IsZero(t *testing.T) {
	// GIVEN: RequireDockerRegistryDefaultsToken.
	tests := []struct {
		name     string
		defaults RequireDockerRegistryDefaultsToken
		want     bool
	}{
		{
			name:     "empty",
			defaults: RequireDockerRegistryDefaultsToken{},
			want:     true,
		},
		{
			name: "non-empty/RequireDockerRegistryDefaultsAuth",
			defaults: RequireDockerRegistryDefaultsToken{
				RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
					Token: "t",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerRegistryDefaultsToken.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerRegistryDefaultsToken_GetToken(t *testing.T) {
	// GIVEN: a RequireDockerRegistryDefaultsToken.
	tests := []struct {
		name     string
		defaults RequireDockerRegistryDefaultsToken
		want     string
	}{
		{
			name:     "empty",
			defaults: RequireDockerRegistryDefaultsToken{},
			want:     "",
		},
		{
			name: "non-empty",
			defaults: RequireDockerRegistryDefaultsToken{
				RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
					Token: "t",
				},
			},
			want: "t",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetToken() is called on it.
			got := tc.defaults.GetToken()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerRegistryDefaultsToken.GetToken() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerCheckRegistryDefaultsTokenWithUsername_IsZero(t *testing.T) {
	// GIVEN: RequireDockerCheckRegistryDefaultsTokenWithUsername.
	tests := []struct {
		name     string
		defaults RequireDockerCheckRegistryDefaultsTokenWithUsername
		want     bool
	}{
		{
			name:     "empty",
			defaults: RequireDockerCheckRegistryDefaultsTokenWithUsername{},
			want:     true,
		},
		{
			name: "non-empty/RequireDockerRegistryDefaultsAuthWithUsername",
			defaults: RequireDockerCheckRegistryDefaultsTokenWithUsername{
				RequireDockerRegistryDefaultsAuthWithUsername: RequireDockerRegistryDefaultsAuthWithUsername{
					Username: "u",
					RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
						Token: "t",
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.defaults.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerCheckRegistryDefaultsTokenWithUsername.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerCheckRegistryDefaultsTokenWithUsername_GetToken(t *testing.T) {
	// GIVEN: a RequireDockerCheckRegistryDefaultsTokenWithUsername.
	tests := []struct {
		name     string
		defaults RequireDockerCheckRegistryDefaultsTokenWithUsername
		want     string
	}{
		{
			name:     "empty",
			defaults: RequireDockerCheckRegistryDefaultsTokenWithUsername{},
			want:     "",
		},
		{
			name: "non-empty",
			defaults: RequireDockerCheckRegistryDefaultsTokenWithUsername{
				RequireDockerRegistryDefaultsAuthWithUsername: RequireDockerRegistryDefaultsAuthWithUsername{
					Username: "u",
					RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
						Token: "t",
					},
				},
			},
			want: "t",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetToken() is called on it.
			got := tc.defaults.GetToken()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerRegistryDefaultsToken.GetToken() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDockerDefaults_IsZero(t *testing.T) {
	// GIVEN: a RequireDockerDefaults.
	tests := []struct {
		name string
		d    RequireDockerDefaults
		want bool
	}{
		{
			name: "empty",
			d:    RequireDockerDefaults{},
			want: true,
		},
		{
			name: "non-empty/Type",
			d: RequireDockerDefaults{
				Type: "t",
			},
			want: false,
		},
		{
			name: "non-empty/Tag",
			d: RequireDockerDefaults{
				Tag: "t",
			},
			want: false,
		},
		{
			name: "non-empty/Registry",
			d: RequireDockerDefaults{
				Registry: RequireDockerRegistriesDefaults{
					GHCR: &RequireDockerRegistryDefaultsToken{
						RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
							Token: "token",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			d: RequireDockerDefaults{
				Type: "t",
				Tag:  "t",
				Registry: RequireDockerRegistriesDefaults{
					GHCR: &RequireDockerRegistryDefaultsToken{
						RequireDockerRegistryDefaultsAuth: RequireDockerRegistryDefaultsAuth{
							Token: "token",
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.d.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDockerDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDeployedVersionLookupDefaults_IsZero(t *testing.T) {
	// GIVEN: a DeployedVersionLookupDefaults.
	tests := []struct {
		name string
		d    DeployedVersionLookupDefaults
		want bool
	}{
		{
			name: "empty",
			d:    DeployedVersionLookupDefaults{},
			want: true,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			d: DeployedVersionLookupDefaults{
				AllowInvalidCerts: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "non-empty/Method",
			d: DeployedVersionLookupDefaults{
				Method: "m",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			d: DeployedVersionLookupDefaults{
				AllowInvalidCerts: test.Ptr(true),
				Method:            "m",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.d.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nDeployedVersionLookupDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDeployedVersionLookup_String(t *testing.T) {
	// GIVEN: a DeployedVersionLookup.
	tests := []struct {
		name string
		dvl  *DeployedVersionLookup
		want string
	}{
		{
			name: "nil",
			dvl:  nil,
			want: "",
		},
		{
			name: "empty",
			dvl:  &DeployedVersionLookup{},
			want: "{}",
		},
		{
			name: "filled",
			dvl: &DeployedVersionLookup{
				Method:            http.MethodPost,
				URL:               "https://release-argus.io",
				AllowInvalidCerts: test.Ptr(false),
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "pass",
				},
				Headers: []Header{
					{Key: "X-Header", Value: "bosh"},
					{Key: "X-Other", Value: "bash"},
				},
				Body:         "what",
				JSON:         "boo",
				Regex:        `bam`,
				HardDefaults: &DeployedVersionLookup{},
				Defaults:     &DeployedVersionLookup{},
			},
			want: `
				{
					"method": "POST",
					"url": "https://release-argus.io",
					"allow_invalid_certs": false,
					"basic_auth": {
						"username": "user",
						"password": "pass"
					},
					"headers": [
						{"key": "X-Header","value": "bosh"},
						{"key": "X-Other","value": "bash"}
					],
					"body": "what",
					"json": "boo",
					"regex": "bam"
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the DeployedVersionLookup is stringified with String.
			got := tc.dvl.String()

			// THEN: the result is as expected.
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf(
					"%s\nDeployedVersionLookup.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestURLCommands_String(t *testing.T) {
	// GIVEN: URLCommands.
	tests := []struct {
		name        string
		urlCommands *URLCommands
		want        string
	}{
		{
			name:        "nil",
			urlCommands: nil,
			want:        "",
		},
		{
			name:        "empty",
			urlCommands: &URLCommands{},
			want:        "[]",
		},
		{
			name: "one of each type",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `bam`},
				{Type: "replace", Old: "want-rid", New: "replacement"},
				{Type: "split", Text: "split on me", Index: test.Ptr(5)},
			},
			want: `
				[
					{"type": "regex","regex": "bam"},
					{"type": "replace","new": "replacement","old": "want-rid"},
					{"type": "split","index": 5,"text": "split on me"}
				]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the URLCommands is stringified with String.
			got := tc.urlCommands.String()

			// THEN: the result is as expected.
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf(
					"%s\nURLCommands.String() value mismatch\ngot:  %qs\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	// GIVEN: a Status.
	tests := []struct {
		name   string
		status *Status
		want   string
	}{
		{
			name:   "nil",
			status: nil,
			want:   "",
		},
		{
			name:   "empty",
			status: &Status{},
			want:   "{}",
		},
		{
			name: "filled",
			status: &Status{
				ApprovedVersion:          "1.2.4",
				DeployedVersion:          "1.2.3",
				DeployedVersionTimestamp: "2022-01-01T01:01:01Z",
				LatestVersion:            "1.2.4",
				LatestVersionTimestamp:   "2022-01-01T01:01:01Z",
				LastQueried:              "2022-01-01T01:01:01Z",
				RegexMissesContent:       1,
				RegexMissesVersion:       2,
			},
			want: `
				{
					"approved_version": "1.2.4",
					"deployed_version": "1.2.3",
					"deployed_version_timestamp": "2022-01-01T01:01:01Z",
					"latest_version": "1.2.4",
					"latest_version_timestamp": "2022-01-01T01:01:01Z",
					"last_queried": "2022-01-01T01:01:01Z",
					"regex_misses_content": 1,
					"regex_misses_version": 2
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN: the Status is stringified with String.
			got := tc.status.String()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nStatus.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHooks_String(t *testing.T) {
	// GIVEN: WebHooks.
	tests := []struct {
		name     string
		webhooks *WebHooks
		want     string
	}{
		{
			name:     "nil",
			webhooks: nil,
			want:     "",
		},
		{
			name:     "empty",
			webhooks: &WebHooks{},
			want:     "{}",
		},
		{
			name: "single webhook, filled",
			webhooks: &WebHooks{
				"0": {ServiceID: "something",
					ID:                "foobar",
					Type:              "url",
					URL:               "https://release-argus.io",
					AllowInvalidCerts: test.Ptr(true),
					Secret:            "secret",
					Headers: []Header{
						{Key: "X-Header", Value: "bosh"},
					},
					DesiredStatusCode: test.Ptr[uint16](200),
					Delay:             "1h",
					MaxTries:          test.Ptr[uint8](7),
					SilentFails:       test.Ptr(false),
				},
			},
			want: `
				{
					"0": {
						"name": "foobar",
						"type": "url",
						"url": "https://release-argus.io",
						"allow_invalid_certs": true,
						"secret": "secret",
						"headers": [{"key": "X-Header","value": "bosh"}],
						"desired_status_code": 200,
						"delay": "1h",
						"max_tries": 7,
						"silent_fails": false
					}
				}`,
		},
		{
			name: "multiple webhooks",
			webhooks: &WebHooks{
				"0": {URL: "bish"},
				"1": {Secret: "bash"},
				"2": {Type: "github"},
			},
			want: `
				{
					"0": {"url": "bish"},
					"1": {"secret": "bash"},
					"2": {"type": "github"}
				}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the WebHook is stringified with String.
			got := tc.webhooks.String()

			// THEN: the result is as expected.
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf(
					"%s\nWebHooks.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHooks_Flatten(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name    string
		webhook *WebHooks
		want    []*WebHook
	}{
		{
			name:    "nil",
			webhook: nil,
			want:    nil,
		},
		{
			name:    "empty",
			webhook: &WebHooks{},
			want:    []*WebHook{},
		},
		{
			name: "webhooks ordered",
			webhook: &WebHooks{
				"alpha": WebHook{URL: "https://example.com"},
				"bravo": WebHook{URL: "https://example.com/other"},
			},
			want: []*WebHook{
				{ID: "alpha", URL: "https://example.com"},
				{ID: "bravo", URL: "https://example.com/other"},
			},
		},
		{
			name: "webhooks ordered and censored",
			webhook: &WebHooks{
				"alpha": WebHook{
					URL:    "https://example.com",
					Secret: "foo",
				},
				"bravo": WebHook{
					URL:    "https://example.com/other",
					Secret: "bar",
				},
			},
			want: []*WebHook{
				{ID: "alpha",
					URL:    "https://example.com",
					Secret: util.SecretValue,
				},
				{ID: "bravo",
					URL:    "https://example.com/other",
					Secret: util.SecretValue,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Flatten is called on it.
			got := tc.webhook.Flatten()

			// THEN: the map is flattened, ordered and censored.
			gotBytes, _ := decode.Marshal("json", got)
			wantBytes, _ := decode.Marshal("json", tc.want)
			if gotStr, wantStr := string(gotBytes), string(wantBytes); gotStr != wantStr {
				t.Errorf(
					"%s\nWebHooks.Flatten()\ngot:  %q\nwant: %q",
					packageName, gotStr, wantStr,
				)
			}
		})
	}
}

func TestWebHook_IsZero(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name    string
		webhook WebHook
		want    bool
	}{
		{
			name:    "empty",
			webhook: WebHook{},
			want:    true,
		},
		{
			name: "non-empty/ServiceID",
			webhook: WebHook{
				ServiceID: "alpha",
			},
			want: false,
		},
		{
			name: "non-empty/ID",
			webhook: WebHook{
				ID: "alpha",
			},
			want: false,
		},
		{
			name: "non-empty/Type",
			webhook: WebHook{
				Type: "alpha",
			},
			want: false,
		},
		{
			name: "non-empty/URL",
			webhook: WebHook{
				URL: "alpha",
			},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			webhook: WebHook{
				AllowInvalidCerts: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "non-empty/Secret",
			webhook: WebHook{
				Secret: "alpha",
			},
			want: false,
		},
		{
			name: "non-empty/Headers",
			webhook: WebHook{
				Headers: []Header{
					{
						Key:   "alpha",
						Value: "alpha",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/DesiredStatusCode",
			webhook: WebHook{
				DesiredStatusCode: test.Ptr[uint16](200),
			},
			want: false,
		},
		{
			name: "non-empty/Delay",
			webhook: WebHook{
				Delay: "2s",
			},
			want: false,
		},
		{
			name: "non-empty/MaxTries",
			webhook: WebHook{
				MaxTries: test.Ptr[uint8](2),
			},
			want: false,
		},
		{
			name: "non-empty/SilentFails",
			webhook: WebHook{
				SilentFails: test.Ptr(true),
			},
		},
		{
			name: "non-empty/all",
			webhook: WebHook{
				ServiceID:         "alpha",
				ID:                "alpha",
				Type:              "alpha",
				URL:               "alpha",
				AllowInvalidCerts: test.Ptr(true),
				Secret:            "alpha",
				Headers: []Header{
					{Key: "alpha", Value: "alpha"},
				},
				DesiredStatusCode: test.Ptr[uint16](200),
				Delay:             "2s",
				MaxTries:          test.Ptr[uint8](2),
				SilentFails:       test.Ptr(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.webhook.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.IsZero() value mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_String(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name    string
		webhook *WebHook
		want    string
	}{
		{
			name:    "nil",
			webhook: nil,
			want:    "",
		},
		{
			name:    "empty",
			webhook: &WebHook{},
			want:    "{}\n",
		},
		{
			name: "filled",
			webhook: &WebHook{
				ServiceID:         "something",
				ID:                "foobar",
				Type:              "url",
				URL:               "https://release-argus.io",
				AllowInvalidCerts: test.Ptr(true),
				Secret:            "secret",
				Headers: []Header{
					{Key: "X-Header", Value: "bosh"},
				},
				DesiredStatusCode: test.Ptr[uint16](200),
				Delay:             "1h",
				MaxTries:          test.Ptr[uint8](7),
				SilentFails:       test.Ptr(false),
			},
			want: test.TrimYAML(`
				name: foobar
				type: url
				url: https://release-argus.io
				allow_invalid_certs: true
				secret: secret
				headers:
					- key: X-Header
						value: bosh
				desired_status_code: 200
				delay: 1h
				max_tries: 7
				silent_fails: false
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.webhook.String,
				tc.want,
			)
		})
	}
}

func TestWebHook_Censor(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name          string
		webhook, want *WebHook
	}{
		{
			name:    "nil",
			webhook: nil,
			want:    nil,
		},
		{
			name: "secret",
			webhook: &WebHook{
				Secret: "shazam",
			},
			want: &WebHook{
				Secret: util.SecretValue,
			},
		},
		{
			name: "headers",
			webhook: &WebHook{
				Headers: []Header{
					{Key: "X-Header", Value: "something"},
					{Key: "X-Bing", Value: "Bam"},
				},
			},
			want: &WebHook{
				Headers: []Header{
					{Key: "X-Header", Value: util.SecretValue},
					{Key: "X-Bing", Value: util.SecretValue},
				},
			},
		},
		{
			name: "all",
			webhook: &WebHook{
				Secret: "shazam",
				Headers: []Header{
					{Key: "X-Header", Value: "something"},
					{Key: "X-Bing", Value: "Bam"},
				},
			},
			want: &WebHook{
				Secret: util.SecretValue,
				Headers: []Header{
					{Key: "X-Header", Value: util.SecretValue},
					{Key: "X-Bing", Value: util.SecretValue},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Censor is called on it.
			tc.webhook.Censor()

			prefix := fmt.Sprintf("%s\nWebHook.Censor()", packageName)

			// THEN: nil WebHooks are kept.
			if tc.webhook == tc.want {
				return
			}

			// AND: the Secret is censored.
			if tc.webhook.Secret != tc.want.Secret {
				t.Errorf(
					"%s .Secret uncensored\ngot:  %q\nwant: %q",
					prefix, tc.webhook.Secret, tc.want.Secret,
				)
			}

			// AND: the headers are as expected.
			gotHeaders := tc.webhook.Headers
			wantHeaders := tc.want.Headers
			if testErr := test.AssertSlicesEqualFunc(
				t,
				gotHeaders,
				wantHeaders,
				func(a, b Header) bool { return a.Key == b.Key && a.Value == b.Value },
				prefix,
				"Headers",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestNilIfUnchanged(t *testing.T) {
	// GIVEN: two pointers to integers.
	tests := []struct {
		name     string
		oldValue *int
		newValue *int
		want     *int
	}{
		{
			name:     "unchanged/nil->nil",
			oldValue: nil,
			newValue: nil,
			want:     nil,
		},
		{
			name:     "unchanged/value->value",
			oldValue: test.Ptr(1),
			newValue: test.Ptr(1),
			want:     nil,
		},
		{
			name:     "removed, non-nil->nil",
			oldValue: test.Ptr(1),
			newValue: nil,
			want:     test.Ptr(0),
		},
		{
			name:     "added, nil->non-nil",
			oldValue: nil,
			newValue: test.Ptr(1),
			want:     test.Ptr(1),
		},
		{
			name:     "changed, non-nil->other-non-nil",
			oldValue: test.Ptr(1),
			newValue: test.Ptr(2),
			want:     test.Ptr(2),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: nilIfUnchanged is called.
			got := nilIfUnchanged(tc.oldValue, tc.newValue)

			prefix := fmt.Sprintf(
				"%s\nnilIfUnchanged(ptr=%v, val=%v)",
				packageName, tc.oldValue, tc.newValue,
			)

			// THEN: the newValue is nil'd if it's the same as oldValue.
			gotStr := test.StringifyPtr(got)
			wantStr := test.StringifyPtr(tc.want)
			if gotStr != wantStr {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}
		})
	}
}
