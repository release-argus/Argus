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
	"encoding/json"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNotify_Censor(t *testing.T) {
	// GIVEN a Notify
	tests := map[string]struct {
		notify, want *Notify
	}{
		"nil": {
			notify: nil,
			want:   nil,
		},
		"url_fields": {
			notify: &Notify{
				URLFields: map[string]string{
					"altid":    "alpha",
					"apikey":   "bravo",
					"botkey":   "charlie",
					"password": "delta",
					"token":    "echo",
					"tokena":   "foxtrot",
					"tokenb":   "golf"}},
			want: &Notify{
				URLFields: map[string]string{
					"altid":    util.SecretValue,
					"apikey":   util.SecretValue,
					"botkey":   util.SecretValue,
					"password": util.SecretValue,
					"token":    util.SecretValue,
					"tokena":   util.SecretValue,
					"tokenb":   util.SecretValue}},
		},
		"params": {
			notify: &Notify{
				Params: map[string]string{
					"devices": "foo"}},
			want: &Notify{
				Params: map[string]string{
					"devices": util.SecretValue}},
		},
		"all censorable": {
			notify: &Notify{
				URLFields: map[string]string{
					"altid":    "alpha",
					"apikey":   "bravo",
					"botkey":   "charlie",
					"password": "delta",
					"token":    "echo",
					"tokena":   "foxtrot",
					"tokenb":   "golf"},
				Params: map[string]string{
					"devices": "hotel"}},
			want: &Notify{
				URLFields: map[string]string{
					"altid":    util.SecretValue,
					"apikey":   util.SecretValue,
					"botkey":   util.SecretValue,
					"password": util.SecretValue,
					"token":    util.SecretValue,
					"tokena":   util.SecretValue,
					"tokenb":   util.SecretValue},
				Params: map[string]string{
					"devices": util.SecretValue}},
		},
		"all censorable, plus non-censored": {
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
					"events":  "lima"}},
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
					"username": "india"},
				Params: map[string]string{
					"devices": util.SecretValue,
					"rooms":   "kilo",
					"events":  "lima"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Censor is called on it
			tc.notify.Censor()

			// THEN nil Notifiers are kept
			if tc.notify == tc.want {
				return
			}
			// AND defined fields are censored as expected
			for k := range tc.want.URLFields {
				if tc.want.URLFields[k] != tc.notify.URLFields[k] {
					t.Errorf("URLField %q:\ngot %q, want %q",
						k, tc.notify.URLFields[k], tc.want.URLFields[k])
				}
			}
			for k := range tc.want.Params {
				if tc.want.Params[k] != tc.notify.Params[k] {
					t.Errorf("Param %q:\ngot %q, want %q",
						k, tc.notify.Params[k], tc.want.Params[k])
				}
			}
		})
	}
}

func TestNotifySlice_Censor(t *testing.T) {
	// GIVEN a Notify
	tests := map[string]struct {
		notify, want *NotifySlice
	}{
		"nil": {
			notify: nil,
			want:   nil,
		},
		"non-nil": {
			notify: &NotifySlice{
				"0": &Notify{
					URLFields: map[string]string{
						"password": "alpha",
						"port":     "bravo"},
					Params: map[string]string{
						"devices": "charlie",
						"rooms":   "delta"}},
				"1": &Notify{
					URLFields: map[string]string{
						"altid": "echo",
						"port":  "foxtrot"},
					Params: map[string]string{
						"devices": "hotel",
						"rooms":   "golf"}},
			},
			want: &NotifySlice{
				"0": &Notify{
					URLFields: map[string]string{
						"password": util.SecretValue,
						"port":     "bravo"},
					Params: map[string]string{
						"devices": util.SecretValue,
						"rooms":   "delta"}},
				"1": &Notify{
					URLFields: map[string]string{
						"altid": util.SecretValue,
						"port":  "foxtrot"},
					Params: map[string]string{
						"devices": util.SecretValue,
						"rooms":   "golf"}},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Censor is called on it
			tc.notify.Censor()

			// THEN nil NotifySlices are kept
			if tc.notify == tc.want {
				return
			}
			// AND defined fields are censored as expected
			for i := range *tc.notify {
				for k := range (*tc.want)[i].URLFields {
					if (*tc.want)[i].URLFields[k] != (*tc.notify)[i].URLFields[k] {
						t.Errorf("URLField %q:\ngot %q, want %q",
							k, (*tc.notify)[i].URLFields[k], (*tc.want)[i].URLFields[k])
					}
				}
				for k := range (*tc.want)[i].Params {
					if (*tc.want)[i].Params[k] != (*tc.notify)[i].Params[k] {
						t.Errorf("Param %q:\ngot %q, want %q",
							k, (*tc.notify)[i].Params[k], (*tc.want)[i].Params[k])
					}
				}
			}
		})
	}
}

func TestNotifySlice_Flatten(t *testing.T) {
	// GIVEN a NotifySlice
	tests := map[string]struct {
		notify *NotifySlice
		want   *[]Notify
	}{
		"nil": {
			notify: nil,
			want:   nil,
		},
		"empty": {
			notify: &NotifySlice{},
			want:   &[]Notify{},
		},
		"ordered": {
			notify: &NotifySlice{
				"zulu": &Notify{
					URLFields: map[string]string{
						"port": "alpha"},
					Params: map[string]string{
						"hosts": "bravo"}},
				"yankee": &Notify{
					URLFields: map[string]string{
						"path": "charlie"},
					Params: map[string]string{
						"rooms": "delta"}}},
			want: &[]Notify{
				{ID: "yankee",
					URLFields: map[string]string{
						"path": "charlie"},
					Params: map[string]string{
						"rooms": "delta"}},
				{ID: "zulu",
					URLFields: map[string]string{
						"port": "alpha"},
					Params: map[string]string{
						"hosts": "bravo"}}},
		},
		"ordered and censored": {
			notify: &NotifySlice{
				"hotel": &Notify{
					URLFields: map[string]string{
						"port":  "alpha",
						"altid": "echo"},
					Params: map[string]string{
						"hosts":   "bravo",
						"devices": "foxtrot"}},
				"golf": &Notify{
					URLFields: map[string]string{
						"path":   "charlie",
						"botkey": "india"},
					Params: map[string]string{
						"rooms": "delta"}}},
			want: &[]Notify{
				{ID: "golf",
					URLFields: map[string]string{
						"path":   "charlie",
						"botkey": util.SecretValue},
					Params: map[string]string{
						"rooms": "delta"}},
				{ID: "hotel",
					URLFields: map[string]string{
						"port":  "alpha",
						"altid": util.SecretValue},
					Params: map[string]string{
						"hosts":   "bravo",
						"devices": util.SecretValue}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Flatten is called on it
			got := tc.notify.Flatten()

			// THEN nil NotifySlices are kept
			if tc.notify == nil && tc.want == nil {
				return
			}
			// AND defined fields are censored as expected
			for i := range *tc.want {
				if (*tc.want)[i].ID != got[i].ID {
					t.Errorf("ID:\ngot %q, want %q",
						got[i].ID, (*tc.want)[i].ID)
				}
				for k := range (*tc.want)[i].URLFields {
					if (*tc.want)[i].URLFields[k] != got[i].URLFields[k] {
						t.Errorf("URLField %q:\ngot %q, want %q",
							k, got[i].URLFields[k], (*tc.want)[i].URLFields[k])
					}
				}
				for k := range (*tc.want)[i].Params {
					if (*tc.want)[i].Params[k] != got[i].Params[k] {
						t.Errorf("Param %q:\ngot %q, want %q",
							k, got[i].Params[k], (*tc.want)[i].Params[k])
					}
				}
			}
		})
	}
}

func TestWebHook_Censor(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		webhook, want *WebHook
	}{
		"nil": {
			webhook: nil,
			want:    nil,
		},
		"secret": {
			webhook: &WebHook{
				Secret: "shazam"},
			want: &WebHook{
				Secret: util.SecretValue},
		},
		"custom_headers": {
			webhook: &WebHook{
				CustomHeaders: &[]Header{
					{Key: "X-Header", Value: "something"},
					{Key: "X-Bing", Value: "Bam"}}},
			want: &WebHook{
				CustomHeaders: &[]Header{
					{Key: "X-Header", Value: util.SecretValue},
					{Key: "X-Bing", Value: util.SecretValue}}},
		},
		"all": {
			webhook: &WebHook{
				Secret: "shazam",
				CustomHeaders: &[]Header{
					{Key: "X-Header", Value: "something"},
					{Key: "X-Bing", Value: "Bam"}}},
			want: &WebHook{
				Secret: util.SecretValue,
				CustomHeaders: &[]Header{
					{Key: "X-Header", Value: util.SecretValue},
					{Key: "X-Bing", Value: util.SecretValue}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Censor is called on it
			tc.webhook.Censor()

			// THEN nil WebHooks are kept
			if tc.webhook == tc.want {
				return
			}
			// AND the Secret is censored
			if tc.want.Secret != tc.webhook.Secret {
				t.Errorf("Secret uncensored\ngot %q, want %q",
					tc.webhook.Secret, tc.want.Secret)
			}
			if tc.webhook.CustomHeaders != nil {
				for i := range *tc.want.CustomHeaders {
					if (*tc.want.CustomHeaders)[i] != (*tc.webhook.CustomHeaders)[i] {
						t.Errorf("Header %d:\ngot %v, want %v",
							i, (*tc.webhook.CustomHeaders)[i], (*tc.want.CustomHeaders)[i])
					}
				}
			}
		})
	}
}

func TestWebHookSlice_Flatten(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		webhook *WebHookSlice
		want    []*WebHook
	}{
		"nil": {
			webhook: nil,
			want:    nil,
		},
		"empty": {
			webhook: &WebHookSlice{},
			want:    []*WebHook{},
		},
		"webhooks ordered": {
			webhook: &WebHookSlice{
				"alpha": &WebHook{URL: "https://example.com"},
				"bravo": &WebHook{URL: "https://example.com/other"}},
			want: []*WebHook{
				{ID: "alpha", URL: "https://example.com"},
				{ID: "bravo", URL: "https://example.com/other"}},
		},
		"webhooks ordered and censored": {
			webhook: &WebHookSlice{
				"alpha": &WebHook{
					URL:    "https://example.com",
					Secret: "foo"},
				"bravo": &WebHook{
					URL:    "https://example.com/other",
					Secret: "bar"}},
			want: []*WebHook{
				{ID: "alpha",
					URL:    "https://example.com",
					Secret: util.SecretValue},
				{ID: "bravo",
					URL:    "https://example.com/other",
					Secret: util.SecretValue}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Flatten is called on it
			got := tc.webhook.Flatten()

			// THEN the map is flattened, ordered and censored
			gotBytes, _ := json.Marshal(got)
			wantBytes, _ := json.Marshal(tc.want)
			if string(gotBytes) != string(wantBytes) {
				t.Errorf("Flatten\ngot %q, want %q",
					string(gotBytes), string(wantBytes))
			}
		})
	}
}

func TestServiceSummary_String(t *testing.T) {
	// GIVEN a ServiceSummary
	tests := map[string]struct {
		summary *ServiceSummary
		want    string
	}{
		"nil": {
			summary: nil,
			want:    "",
		},
		"empty": {
			summary: &ServiceSummary{},
			want:    "{}",
		},
		"some": {
			summary: &ServiceSummary{
				ID:      "foo",
				Name:    test.StringPtr("bar"),
				Type:    "github",
				Command: test.IntPtr(1),
				WebHook: test.IntPtr(2)},
			want: `
				{
					"id": "foo",
					"name": "bar",
					"type": "github",
					"command": 1,
					"webhook": 2
				}`,
		},
		"full": {
			summary: &ServiceSummary{
				ID:                       "bar",
				Name:                     test.StringPtr("foo"),
				Active:                   test.BoolPtr(true),
				Comment:                  "test",
				Type:                     "gitlab",
				WebURL:                   test.StringPtr("http://example.com"),
				Icon:                     test.StringPtr("https://example.com/icon.png"),
				IconLinkTo:               test.StringPtr("https://release-argus.io"),
				HasDeployedVersionLookup: test.BoolPtr(true),
				Command:                  test.IntPtr(2),
				WebHook:                  test.IntPtr(1),
				Status: &Status{
					ApprovedVersion: "1.2.3"}},
			want: `
				{
					"id": "bar",
					"name": "foo",
					"active": true,
					"comment": "test",
					"type": "gitlab",
					"url": "http://example.com",
					"icon": "https://example.com/icon.png",
					"icon_link_to": "https://release-argus.io",
					"has_deployed_version": true,
					"command": 2,
					"webhook": 1,
					"status": {
						"approved_version": "1.2.3"
				}}`,
		},
	}

	// WHEN String is called on it
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN the Summary is stringified with String
			got := tc.summary.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestNilUnchanged(t *testing.T) {
	tests := map[string]struct {
		oldValue *int
		newValue *int
		want     *int
	}{
		"unchanged - nil->nil": {
			oldValue: nil,
			newValue: nil,
			want:     nil,
		},
		"unchanged - value->value": {
			oldValue: test.IntPtr(1),
			newValue: test.IntPtr(1),
			want:     nil,
		},
		"removed - non-nil->nil": {
			oldValue: test.IntPtr(1),
			newValue: nil,
			want:     test.IntPtr(0),
		},
		"added - nil->non-nil": {
			oldValue: nil,
			newValue: test.IntPtr(1),
			want:     test.IntPtr(1),
		},
		"changed - non-nil->other-non-nil": {
			oldValue: test.IntPtr(1),
			newValue: test.IntPtr(2),
			want:     test.IntPtr(2),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN nilUnchanged is called
			nilUnchanged(tc.oldValue, &tc.newValue)

			// THEN the result is as expected
			if (tc.want == nil && tc.newValue != nil) || (tc.want != nil && tc.newValue == nil) || (tc.want != nil && tc.newValue != nil && *tc.want != *tc.newValue) {
				t.Errorf("got: %v, want: %v", tc.newValue, tc.want)
			}
		})
	}
}

func TestServiceSummary_RemoveUnchanged(t *testing.T) {
	// GIVEN two ServiceSummaries
	tests := map[string]struct {
		old, new, want *ServiceSummary
	}{
		"compare to nil": {
			old: nil,
			new: &ServiceSummary{ID: "foo"},
			want: &ServiceSummary{
				ID:      "foo",
				Status:  &Status{},
				Command: test.IntPtr(0),
				WebHook: test.IntPtr(0)},
		},
		"same id": {
			old: &ServiceSummary{
				ID: "foo"},
			new: &ServiceSummary{
				ID: "foo"},
			want: &ServiceSummary{},
		},
		"different id": {
			old: &ServiceSummary{
				ID: "foo"},
			new: &ServiceSummary{
				ID: "bar"},
			want: &ServiceSummary{
				ID: "bar"},
		},
		"name added": {
			old: &ServiceSummary{},
			new: &ServiceSummary{
				Name: test.StringPtr("foo")},
			want: &ServiceSummary{
				Name: test.StringPtr("foo")},
		},
		"name removed": {
			old: &ServiceSummary{
				Name: test.StringPtr("foo")},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				Name: test.StringPtr("")},
		},
		"same name": {
			old: &ServiceSummary{
				Name: test.StringPtr("foo")},
			new: &ServiceSummary{
				Name: test.StringPtr("foo")},
			want: &ServiceSummary{},
		},
		"different name": {
			old: &ServiceSummary{
				Name: test.StringPtr("foo")},
			new: &ServiceSummary{
				Name: test.StringPtr("bar")},
			want: &ServiceSummary{
				Name: test.StringPtr("bar")},
		},
		"same active": {
			old: &ServiceSummary{
				Active: test.BoolPtr(false)},
			new: &ServiceSummary{
				Active: test.BoolPtr(false)},
			want: &ServiceSummary{},
		},
		"different active": {
			old: &ServiceSummary{
				Active: test.BoolPtr(true)},
			new: &ServiceSummary{
				Active: test.BoolPtr(false)},
			want: &ServiceSummary{
				Active: test.BoolPtr(false)},
		},
		"same type": {
			old: &ServiceSummary{
				Type: "github"},
			new: &ServiceSummary{
				Type: "github"},
			want: &ServiceSummary{},
		},
		"different type": {
			old: &ServiceSummary{
				Type: "github"},
			new: &ServiceSummary{
				Type: "gitlab"},
			want: &ServiceSummary{
				Type: "gitlab"},
		},
		"same icon": {
			old: &ServiceSummary{
				Icon: test.StringPtr("https://example.com/icon.png")},
			new: &ServiceSummary{
				Icon: test.StringPtr("https://example.com/icon.png")},
			want: &ServiceSummary{},
		},
		"different icon": {
			old: &ServiceSummary{
				Icon: test.StringPtr("https://example.com/icon.png")},
			new: &ServiceSummary{
				Icon: test.StringPtr("https://example.com/icon2.png")},
			want: &ServiceSummary{
				Icon: test.StringPtr("https://example.com/icon2.png")},
		},
		"same icon_link_to": {
			old: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io")},
			new: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io")},
			want: &ServiceSummary{},
		},
		"different icon_link_to": {
			old: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io")},
			new: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io/other")},
			want: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io/other")},
		},
		"same has_deployed_version_lookup": {
			old: &ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(true)},
			new: &ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(true)},
			want: &ServiceSummary{},
		},
		"different has_deployed_version_lookup": {
			old: &ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(true)},
			new: &ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(false)},
			want: &ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(false)},
		},
		"same approved_version": {
			old: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "1.2.3"}},
			new: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "1.2.3"}},
			want: &ServiceSummary{},
		},
		"different approved_version": {
			old: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "1.2.3"}},
			new: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "4.5.6"}},
			want: &ServiceSummary{
				Status: &Status{
					ApprovedVersion: "4.5.6"}},
		},
		"same deployed_version": {
			old: &ServiceSummary{
				Status: &Status{
					DeployedVersion: "1.2.3"}},
			new: &ServiceSummary{
				Status: &Status{
					DeployedVersion: "1.2.3"}},
			want: &ServiceSummary{},
		},
		"same deployed_version, different timestamps ignored": {
			old: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-01-01T00:00:00Z"}},
			new: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z"}},
			want: &ServiceSummary{},
		},
		"different deployed_version": {
			old: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-01-01T00:00:00Z"}},
			new: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z"}},
			want: &ServiceSummary{
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z"}},
		},
		"same latest_version": {
			old: &ServiceSummary{
				Status: &Status{
					LatestVersion: "1.2.3"}},
			new: &ServiceSummary{
				Status: &Status{
					LatestVersion: "1.2.3"}},
			want: &ServiceSummary{},
		},
		"same latest_version, different timestamps ignored": {
			old: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "1.2.3",
					LatestVersionTimestamp: "2020-01-01T00:00:00Z"}},
			new: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "1.2.3",
					LatestVersionTimestamp: "2020-02-02T00:00:00Z"}},
			want: &ServiceSummary{},
		},
		"different latest_version": {
			old: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "1.2.3",
					LatestVersionTimestamp: "2020-01-01T00:00:00Z"}},
			new: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "4.5.6",
					LatestVersionTimestamp: "2020-02-02T00:00:00Z"}},
			want: &ServiceSummary{
				Status: &Status{
					LatestVersion:          "4.5.6",
					LatestVersionTimestamp: "2020-02-02T00:00:00Z"}},
		},
		"multiple differences": {
			old: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io"),
				Status: &Status{
					DeployedVersion:          "1.2.3",
					DeployedVersionTimestamp: "2020-01-01T00:00:00Z"}},
			new: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io/other"),
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z"}},
			want: &ServiceSummary{
				IconLinkTo: test.StringPtr("https://release-argus.io/other"),
				Status: &Status{
					DeployedVersion:          "4.5.6",
					DeployedVersionTimestamp: "2020-02-02T00:00:00Z"}},
		},
		"tags added": {
			old: &ServiceSummary{},
			new: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"foo"})},
			want: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"foo"})},
		},
		"tags removed": {
			old: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"foo"})},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{})},
		},
		"same tags": {
			old: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"foo"})},
			new: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"foo"})},
			want: &ServiceSummary{},
		},
		"different tags": {
			old: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"foo"})},
			new: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"bar"})},
			want: &ServiceSummary{
				Tags: test.StringSlicePtr([]string{"bar"})},
		},
		"command added": {
			old: &ServiceSummary{},
			new: &ServiceSummary{
				Command: test.IntPtr(1)},
			want: &ServiceSummary{
				Command: test.IntPtr(1)},
		},
		"command removed": {
			old: &ServiceSummary{
				Command: test.IntPtr(1)},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				Command: test.IntPtr(0)},
		},
		"same command": {
			old: &ServiceSummary{
				Command: test.IntPtr(1)},
			new: &ServiceSummary{
				Command: test.IntPtr(1)},
			want: &ServiceSummary{
				Command: nil},
		},
		"webhook added": {
			old: &ServiceSummary{},
			new: &ServiceSummary{
				WebHook: test.IntPtr(1)},
			want: &ServiceSummary{
				WebHook: test.IntPtr(1)},
		},
		"webhook removed": {
			old: &ServiceSummary{
				WebHook: test.IntPtr(1)},
			new: &ServiceSummary{},
			want: &ServiceSummary{
				WebHook: test.IntPtr(0)},
		},
		"same webhook": {
			old: &ServiceSummary{
				WebHook: test.IntPtr(1)},
			new: &ServiceSummary{
				WebHook: test.IntPtr(1)},
			want: &ServiceSummary{
				WebHook: nil},
		},
	}

	// Helper function to initialize fields
	initialiseFields := func(instance *ServiceSummary) {
		if instance.Status == nil {
			instance.Status = &Status{}
		}
		if instance.Command == nil {
			instance.Command = test.IntPtr(0)
		}
		if instance.WebHook == nil {
			instance.WebHook = test.IntPtr(0)
		}
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Give them non-nil Status, Command and WebHook.
			if tc.old != nil {
				initialiseFields(tc.old)
			}
			if tc.new != nil {
				initialiseFields(tc.new)
			}

			// WHEN RemoveUnchanged is called, comparing new to old
			tc.new.RemoveUnchanged(tc.old)

			// THEN the values that are unchanged are removed
			if tc.want.String() != tc.new.String() {
				t.Errorf("got:\n%q\nwant:\n%q",
					tc.new.String(), tc.want.String())
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	// GIVEN a Status
	tests := map[string]struct {
		status *Status
		want   string
	}{
		"nil": {
			status: nil,
			want:   "",
		},
		"empty": {
			status: &Status{},
			want:   "{}",
		},
		"all fields": {
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
				}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN the Status is stringified with String
			got := tc.status.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestWebHook_String(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		webhook *WebHook
		want    string
	}{
		"nil": {
			webhook: nil,
			want:    "",
		},
		"empty": {
			webhook: &WebHook{},
			want:    "{}",
		},
		"all fields": {
			webhook: &WebHook{
				ServiceID:         "something",
				ID:                "foobar",
				Type:              "url",
				URL:               "https://release-argus.io",
				AllowInvalidCerts: test.BoolPtr(true),
				Secret:            "secret",
				CustomHeaders: &[]Header{
					{Key: "X-Header", Value: "bosh"}},
				DesiredStatusCode: test.UInt16Ptr(200),
				Delay:             "1h",
				MaxTries:          test.UInt8Ptr(7),
				SilentFails:       test.BoolPtr(false),
			},
			want: `
				{
					"name": "foobar",
					"type": "url",
					"url": "https://release-argus.io",
					"allow_invalid_certs": true,
					"secret": "secret",
					"custom_headers": [{"key": "X-Header","value": "bosh"}],
					"desired_status_code": 200,
					"delay": "1h",
					"max_tries": 7,
					"silent_fails": false
				}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the WebHook is stringified with String
			got := tc.webhook.String()

			// THEN the result is as expected
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q",
					got, tc.want)
			}
		})
	}
}

func TestWebHookSlice_String(t *testing.T) {
	// GIVEN a WebHookSlice
	tests := map[string]struct {
		slice *WebHookSlice
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "",
		},
		"empty": {
			slice: &WebHookSlice{},
			want:  "{}",
		},
		"single webhook, all fields": {
			slice: &WebHookSlice{
				"0": {ServiceID: "something",
					ID:                "foobar",
					Type:              "url",
					URL:               "https://release-argus.io",
					AllowInvalidCerts: test.BoolPtr(true),
					Secret:            "secret",
					CustomHeaders: &[]Header{
						{Key: "X-Header", Value: "bosh"}},
					DesiredStatusCode: test.UInt16Ptr(200),
					Delay:             "1h",
					MaxTries:          test.UInt8Ptr(7),
					SilentFails:       test.BoolPtr(false)},
			},
			want: `
				{
					"0": {
						"name": "foobar",
						"type": "url",
						"url": "https://release-argus.io",
						"allow_invalid_certs": true,
						"secret": "secret",
						"custom_headers": [{"key": "X-Header","value": "bosh"}],
						"desired_status_code": 200,
						"delay": "1h",
						"max_tries": 7,
						"silent_fails": false
					}
				}`,
		},
		"multiple webhooks": {
			slice: &WebHookSlice{
				"0": {URL: "bish"},
				"1": {Secret: "bash"},
				"2": {Type: "github"}},
			want: `
				{
					"0": {"url": "bish"},
					"1": {"secret": "bash"},
					"2": {"type": "github"}
				}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the WebHook is stringified with String
			got := tc.slice.String()

			// THEN the result is as expected
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf("got: %q\nwant: %q",
					got, tc.want)
			}
		})
	}
}

func TestNotifySlice_String(t *testing.T) {
	// GIVEN a NotifySlice
	tests := map[string]struct {
		slice *NotifySlice
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "",
		},
		"empty": {
			slice: &NotifySlice{},
			want:  "{}",
		},
		"one": {
			slice: &NotifySlice{
				"0": {
					ID:   "foo",
					Type: "discord",
					Options: map[string]string{
						"message": "hello world"},
					URLFields: map[string]string{
						"username": "bing"},
					Params: map[string]string{
						"devices": "bang"},
				}},
			want: `
				{
					"0": {
						"name": "foo",
						"type": "discord",
						"options": {
							"message": "hello world"},
						"url_fields": {
							"username": "bing"},
						"params": {
							"devices": "bang"}
					}
				}`,
		},
		"multiple": {
			slice: &NotifySlice{
				"0": {
					ID:   "foo",
					Type: "discord",
					Options: map[string]string{
						"message": "hello world"},
					URLFields: map[string]string{
						"username": "bing"},
					Params: map[string]string{
						"devices": "bang"},
				},
				"other": {
					Type: "gotify"}},
			want: `
				{
					"0": {
						"name": "foo",
						"type": "discord",
						"options": {
							"message": "hello world"},
						"url_fields": {
							"username": "bing"},
						"params": {
							"devices": "bang"}
					},
					"other": {
						"type": "gotify"
					}
				}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the NotifySlice is stringified with String
			got := tc.slice.String()

			// THEN the result is as expected
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestDeployedVersionLookup_String(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	tests := map[string]struct {
		dvl  *DeployedVersionLookup
		want string
	}{
		"nil": {
			dvl:  nil,
			want: "",
		},
		"empty": {
			dvl:  &DeployedVersionLookup{},
			want: "{}",
		},
		"all fields": {
			dvl: &DeployedVersionLookup{
				Method:            "POST",
				URL:               "https://release-argus.io",
				AllowInvalidCerts: test.BoolPtr(false),
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "pass"},
				Headers: []Header{
					{Key: "X-Header", Value: "bosh"},
					{Key: "X-Other", Value: "bash"}},
				Body:         "what",
				JSON:         "boo",
				Regex:        `bam`,
				HardDefaults: &DeployedVersionLookup{},
				Defaults:     &DeployedVersionLookup{}},
			want: `
				{
					"method": "POST",
					"url": "https://release-argus.io",
					"allow_invalid_certs": false,
					"basic_auth": {
						"username": "user",
						"password": "pass"},
					"headers": [
						{"key": "X-Header","value": "bosh"},
						{"key": "X-Other","value": "bash"}],
					"body": "what",
					"json": "boo",
					"regex": "bam"
				}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the DeployedVersionLookup is stringified with String
			got := tc.dvl.String()

			// THEN the result is as expected
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestURLCommandSlice_String(t *testing.T) {
	// GIVEN a URLCommandSlice
	tests := map[string]struct {
		slice *URLCommandSlice
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "",
		},
		"empty": {
			slice: &URLCommandSlice{},
			want:  "[]",
		},
		"one of each type": {
			slice: &URLCommandSlice{
				{Type: "regex", Regex: `bam`},
				{Type: "replace", Old: "want-rid", New: test.StringPtr("replacement")},
				{Type: "split", Text: "split on me", Index: test.IntPtr(5)},
			},
			want: `
				[
					{"type": "regex","regex": "bam"},
					{"type": "replace","new": "replacement","old": "want-rid"},
					{"type": "split","index": 5,"text": "split on me"}
				]`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the URLCommandSlice is stringified with String
			got := tc.slice.String()

			// THEN the result is as expected
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestDefaults_String(t *testing.T) {
	// GIVEN a Defaults
	tests := map[string]struct {
		defaults *Defaults
		want     string
	}{
		"nil": {
			defaults: nil,
			want:     "",
		},
		"empty": {
			defaults: &Defaults{},
			want:     `{}`,
		},
		"all types": {
			defaults: &Defaults{
				Service: ServiceDefaults{
					LatestVersion: &LatestVersionDefaults{
						AccessToken: "foo"}},
				Notify: NotifySlice{
					"gotify": &Notify{
						URLFields: map[string]string{
							"url": "https://gotify.example.com"}}},
				WebHook: WebHook{
					Secret: "bar"}},
			want: `
				{
					"service": {
						"latest_version": {
							"access_token": "foo"}},
					"notify": {
						"gotify": {
							"url_fields": {
								"url": "https://gotify.example.com"}}},
					"webhook": {
						"secret": "bar"}
				}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Defaults are stringified with String
			got := tc.defaults.String()

			// THEN the result is as expected
			tc.want = test.TrimJSON(tc.want)
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestService_String(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		input *Service
		want  string
	}{
		"nil": {
			input: nil,
			want:  "",
		},
		"empty": {
			input: &Service{},
			want:  `{}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Defaults are stringified with String
			got := tc.input.String()

			// THEN the result is as expected
			tc.want = strings.ReplaceAll(tc.want, "\n", "")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestLatestVersion_String(t *testing.T) {
	// GIVEN a LatestVersion
	tests := map[string]struct {
		input *LatestVersion
		want  string
	}{
		"nil": {
			input: nil,
			want:  ""},
		"empty": {
			input: &LatestVersion{},
			want:  `{}`},
		"all fields": {
			input: &LatestVersion{
				Type:              "github",
				URL:               "release-argus/argus",
				AccessToken:       util.SecretValue,
				AllowInvalidCerts: test.BoolPtr(true),
				UsePreRelease:     test.BoolPtr(false),
				URLCommands: &URLCommandSlice{
					{Type: "replace", Old: "this", New: test.StringPtr("withThis")},
					{Type: "split", Text: "splitThis", Index: test.IntPtr(8)},
					{Type: "regex", Regex: `([0-9.]+)`}},
				Require: &LatestVersionRequire{
					RegexContent: ".*"}},
			want: `
				{
					"type": "github",
					"url": "release-argus/argus",
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
				}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN the LatestVersion is stringified with String
			got := tc.input.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("api.latest_version.String() mismatch\ngot:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestLatestVersionRequireDefaults_String(t *testing.T) {
	// GIVEN a LatestVersionRequireDefaults
	tests := map[string]struct {
		lvRD *LatestVersionRequireDefaults
		want string
	}{
		"nil": {
			lvRD: nil,
			want: ""},
		"empty": {
			lvRD: &LatestVersionRequireDefaults{},
			want: `{}`},
		"all fields": {
			lvRD: &LatestVersionRequireDefaults{
				Docker: RequireDockerCheckDefaults{
					Type: "ghcr",
					GHCR: &RequireDockerCheckRegistryDefaults{
						Token: "tokenForGHCR"},
					Hub: &RequireDockerCheckRegistryDefaultsWithUsername{
						RequireDockerCheckRegistryDefaults: RequireDockerCheckRegistryDefaults{
							Token: "tokenForHub"},
						Username: "userForHub"},
					Quay: &RequireDockerCheckRegistryDefaults{
						Token: "tokenForQuay"}}},
			want: `
				{
					"docker": {
						"type": "ghcr",
						"ghcr": {
							"token": "tokenForGHCR"},
						"hub": {
							"token": "tokenForHub",
							"username": "userForHub"},
						"quay": {
							"token": "tokenForQuay"}
					}
				}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN the LatestVersionRequireDefaults are stringified with String
			got := tc.lvRD.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestLatestVersionRequire_String(t *testing.T) {
	// GIVEN a LatestVersionRequire
	tests := map[string]struct {
		input *LatestVersionRequire
		want  string
	}{
		"nil": {
			input: nil,
			want:  ""},
		"empty": {
			input: &LatestVersionRequire{},
			want:  `{}`},
		"all fields": {
			input: &LatestVersionRequire{
				Command: []string{"echo", "hello"},
				Docker: &RequireDockerCheck{
					Type:     "hub",
					Image:    "release-argus/argus",
					Tag:      "{{ version }}",
					Username: "user",
					Token:    util.SecretValue},
				RegexContent: ".*",
				RegexVersion: `([0-9.]+)`},
			want: `
				{
					"command": ["echo","hello"],
					"docker": {
						"type": "hub",
						"image": "release-argus/argus",
						"tag": "{{ version }}",
						"username": "user",
						"token": ` + secretValueMarshalled + `
					},
					"regex_content": ".*",
					"regex_version": "([0-9.]+)"
				}`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.want = test.TrimJSON(tc.want)

			// WHEN the LatestVersionRequire is stringified with String
			got := tc.input.String()

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
