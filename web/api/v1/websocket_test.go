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

package v1

import (
	"testing"
	"time"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestConvertDeployedVersionLookupToAPITypeDeployedVersionLookup(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	tests := map[string]struct {
		dvl                      *deployedver.Lookup
		approvedVersion          string
		deployedVersion          string
		deployedVersionTimestamp string
		latestVersion            string
		latestVersionTimestamp   string
		lastQueried              string

		want *api_type.DeployedVersionLookup
	}{
		"nil": {
			dvl:  nil,
			want: nil,
		},
		"empty": {
			dvl:  &deployedver.Lookup{},
			want: &api_type.DeployedVersionLookup{},
		},
		"minimal": {
			dvl: &deployedver.Lookup{
				URL:  "https://example.com",
				JSON: "version"},
			want: &api_type.DeployedVersionLookup{
				URL:  "https://example.com",
				JSON: "version"},
		},
		"censor basic_auth.password": {
			dvl: &deployedver.Lookup{
				URL: "https://example.com",
				BasicAuth: &deployedver.BasicAuth{
					Username: "alan",
					Password: "pass123"}},
			want: &api_type.DeployedVersionLookup{
				URL: "https://example.com",
				BasicAuth: &api_type.BasicAuth{
					Username: "alan",
					Password: "<secret>"}},
		},
		"censor headers": {
			dvl: &deployedver.Lookup{
				URL: "https://example.com",
				Headers: []deployedver.Header{
					{Key: "X-Test-0", Value: "foo"},
					{Key: "X-Test-1", Value: "bar"}}},
			want: &api_type.DeployedVersionLookup{
				URL: "https://example.com",
				Headers: []api_type.Header{
					{Key: "X-Test-0", Value: "<secret>"},
					{Key: "X-Test-1", Value: "<secret>"},
				}},
		},
		"full": {
			dvl: &deployedver.Lookup{
				URL:               "https://release-argus.io",
				AllowInvalidCerts: boolPtr(true),
				BasicAuth: &deployedver.BasicAuth{
					Username: "jim",
					Password: "whoops"},
				Headers: []deployedver.Header{
					{Key: "X-Test-0", Value: "foo"},
					{Key: "X-Test-1", Value: "bar"}},
				JSON:  "version",
				Regex: `([0-9]+\.[0-9]+\.[0-9]+)`,
				Options: &opt.Options{
					Active:             boolPtr(true),
					Interval:           "10m",
					SemanticVersioning: boolPtr(true),
					Defaults:           &opt.Options{},
					HardDefaults:       &opt.Options{}},
				Status: &svcstatus.Status{
					RegexMissesContent: 1,
					RegexMissesVersion: 3,
					Fails:              svcstatus.Fails{},
					ServiceID:          stringPtr("service-id"),
					WebURL:             stringPtr("https://release-argus.io")},
				Defaults:     &deployedver.Lookup{},
				HardDefaults: &deployedver.Lookup{}},
			want: &api_type.DeployedVersionLookup{
				URL:               "https://release-argus.io",
				AllowInvalidCerts: boolPtr(true),
				BasicAuth: &api_type.BasicAuth{
					Username: "jim",
					Password: "<secret>"},
				Headers: []api_type.Header{
					{Key: "X-Test-0", Value: "<secret>"},
					{Key: "X-Test-1", Value: "<secret>"}},
				JSON:  "version",
				Regex: `([0-9]+\.[0-9]+\.[0-9]+)`},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.approvedVersion != "" {
				tc.dvl.Status.SetApprovedVersion("1.2.3")
				tc.dvl.Status.SetDeployedVersion("1.2.3", false)
				tc.dvl.Status.SetDeployedVersionTimestamp(time.Now().Format(time.RFC3339))
				tc.dvl.Status.SetLatestVersion("1.2.3", false)
				tc.dvl.Status.SetLatestVersionTimestamp(time.Now().Format(time.RFC3339))
				tc.dvl.Status.SetLastQueried(time.Now().Format(time.RFC3339))
			}

			// WHEN convertDeployedVersionLookupToAPITypeDeployedVersionLookup is called on it
			got := convertDeployedVersionLookupToAPITypeDeployedVersionLookup(tc.dvl)

			// THEN the WebHookSlice is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertURLCommandSliceToAPITypeURLCommandSlice(t *testing.T) {
	// GIVEN a URL Command slice
	tests := map[string]struct {
		slice *filter.URLCommandSlice
		want  *api_type.URLCommandSlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &filter.URLCommandSlice{},
			want:  &api_type.URLCommandSlice{},
		},
		"regex": {
			slice: &filter.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")}},
			want: &api_type.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")}},
		},
		"replace": {
			slice: &filter.URLCommandSlice{
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")}},
			want: &api_type.URLCommandSlice{
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")}},
		},
		"split": {
			slice: &filter.URLCommandSlice{
				{Type: "split", Index: 7}},
			want: &api_type.URLCommandSlice{
				{Type: "split", Index: 7}},
		},
		"one of each": {
			slice: &filter.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")},
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")},
				{Type: "split", Index: 7}},
			want: &api_type.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("[0-9.]+")},
				{Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")},
				{Type: "split", Index: 7}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertURLCommandSliceToAPITypeURLCommandSlice is called on it
			got := convertURLCommandSliceToAPITypeURLCommandSlice(tc.slice)

			// THEN the WebHookSlice is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want: %q, got: %q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertNotifySliceToAPITypeNotifySlice(t *testing.T) {
	// GIVEN a Notify slice
	tests := map[string]struct {
		slice *shoutrrr.Slice
		want  *api_type.NotifySlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &shoutrrr.Slice{},
			want:  &api_type.NotifySlice{},
		},
		"one": {
			slice: &shoutrrr.Slice{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"message": "something {{ version }}"},
					URLFields: map[string]string{
						"port":  "25",
						"other": "something"},
					Params: map[string]string{
						"avatar": "fizz"}}},
			want: &api_type.NotifySlice{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"message": "something {{ version }}"},
					URLFields: map[string]string{
						"port":  "25",
						"other": "something"},
					Params: map[string]string{
						"avatar": "fizz"}}},
		},
		"one, does censor": {
			slice: &shoutrrr.Slice{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"message": "something {{ version }}"},
					URLFields: map[string]string{
						"port":   "25",
						"other":  "something",
						"apikey": "shazam"},
					Params: map[string]string{
						"avatar":  "fizz",
						"devices": "shazam"}}},
			want: &api_type.NotifySlice{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"message": "something {{ version }}"},
					URLFields: map[string]string{
						"port":   "25",
						"other":  "something",
						"apikey": "<secret>"},
					Params: map[string]string{
						"avatar":  "fizz",
						"devices": "<secret>"}}},
		},
		"multiple": {
			slice: &shoutrrr.Slice{
				"test": {
					URLFields: map[string]string{
						"port":     "25",
						"password": "something"}},
				"test2": {
					Type: "discord",
					Params: map[string]string{
						"avatar":  "fizz",
						"devices": "shazam"}}},
			want: &api_type.NotifySlice{
				"test": {
					URLFields: map[string]string{
						"port":     "25",
						"password": "<secret>"}},
				"test2": {
					Type: "discord",
					Params: map[string]string{
						"avatar":  "fizz",
						"devices": "<secret>"}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertNotifySliceToAPITypeNotifySlice is called on it
			got := convertNotifySliceToAPITypeNotifySlice(tc.slice)

			// THEN the WebHookSlice is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertNotifySliceToAPITypeNotifySliceDoesCensor(t *testing.T) {
	// GIVEN a Notify slice
	slice := shoutrrr.Slice{
		"test": {
			Type: "discord",
			Options: map[string]string{
				"message": "something {{ version }}",
			},
			URLFields: map[string]string{
				"port":   "25",
				"apikey": "fizz",
			},
			Params: map[string]string{
				"avatar":  "argus",
				"devices": "buzz",
			},
		},
	}

	// WHEN convertNotifySliceToAPITypeNotifySlice is called on it
	got := convertNotifySliceToAPITypeNotifySlice(&slice)

	// THEN the slice was converted correctly
	if (*got)["test"].URLFields["port"] != slice["test"].URLFields["port"] ||
		(*got)["test"].URLFields["apikey"] != "<secret>" ||
		(*got)["test"].Params["avatar"] != slice["test"].Params["avatar"] ||
		(*got)["test"].Params["devices"] != "<secret>" {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			slice, got)
	}
}

func TestConvertCommandSliceToAPITypeCommandSlice(t *testing.T) {
	// GIVEN a Command slice
	tests := map[string]struct {
		slice *command.Slice
		want  *api_type.CommandSlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &command.Slice{},
			want:  &api_type.CommandSlice{},
		},
		"one": {
			slice: &command.Slice{
				{"ls", "-lah"}},
			want: &api_type.CommandSlice{
				{"ls", "-lah"}},
		},
		"two": {
			slice: &command.Slice{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"}},
			want: &api_type.CommandSlice{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertCommandSliceToAPITypeCommandSlice is called on it
			got := convertCommandSliceToAPITypeCommandSlice(tc.slice)

			// THEN the CommandSlice is converted correctly
			if got == tc.want { // both nil
				return
			}
			// check number of commands
			if len(*got) != len(*tc.want) {
				t.Errorf("want:\n%v\ngot:\n%v",
					tc.want, got)
				return
			}
			for cI := range *got {
				// check number of args
				if len((*got)[cI]) != len((*tc.want)[cI]) {
					t.Errorf("want:\n%v\ngot:\n%v",
						tc.want, got)
				}
				// check args
				for aI := range (*got)[cI] {
					if (*got)[cI][aI] != (*tc.want)[cI][aI] {
						t.Errorf("want:\n%v\ngot:\n%v",
							tc.want, got)
					}
				}
			}
		})
	}
}

func TestConvertWebHookToAPITypeWebHook(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		wh   *webhook.WebHook
		want *api_type.WebHook
	}{
		"nil": {
			wh:   nil,
			want: nil,
		},
		"empty": {
			wh:   &webhook.WebHook{},
			want: &api_type.WebHook{},
		},
		"censor secret": {
			wh: &webhook.WebHook{
				Secret: "shazam"},
			want: &api_type.WebHook{
				Secret: stringPtr("<secret>")},
		},
		"copy and censor headers": {
			wh: &webhook.WebHook{
				CustomHeaders: &webhook.Headers{
					{Key: "X-Something", Value: "foo"},
					{Key: "X-Another", Value: "bar"}}},
			want: &api_type.WebHook{
				CustomHeaders: []api_type.Header{
					{Key: "X-Something", Value: "<secret>"},
					{Key: "X-Another", Value: "<secret>"}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertWebHookToAPITypeWebHook is called on it
			got := convertWebHookToAPITypeWebHook(tc.wh)

			// THEN the WebHook is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertWebHookSliceToAPITypeWebHookSlice(t *testing.T) {
	// GIVEN a WebHook slice
	tests := map[string]struct {
		slice *webhook.Slice
		want  *api_type.WebHookSlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &webhook.Slice{},
			want:  &api_type.WebHookSlice{},
		},
		"single element": {
			slice: &webhook.Slice{
				"test": {URL: "https://example.com"}},
			want: &api_type.WebHookSlice{
				"test": {URL: stringPtr("https://example.com")}},
		},
		"multiple elements": {
			slice: &webhook.Slice{
				"test":  {URL: "https://example.com"},
				"other": {URL: "https://release-argus.io"}},
			want: &api_type.WebHookSlice{
				"test":  {URL: stringPtr("https://example.com")},
				"other": {URL: stringPtr("https://release-argus.io")}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertWebHookSliceToAPITypeWebHookSlice is called on it
			got := convertWebHookSliceToAPITypeWebHookSlice(tc.slice)

			// THEN the WebHookSlice is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:%q",
					tc.want.String(), got.String())
			}
		})
	}
}
