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

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestConvertAndCensorNotifySlice(t *testing.T) {
	// GIVEN a shoutrrr.Slice
	tests := map[string]struct {
		input *shoutrrr.Slice
		want  *api_type.NotifySlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &shoutrrr.Slice{},
			want:  &api_type.NotifySlice{},
		},
		"one": {
			input: &shoutrrr.Slice{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"test": "1"},
					URLFields: map[string]string{
						"test": "2"},
					Params: map[string]string{
						"test": "3"}}},
			want: &api_type.NotifySlice{
				"test": {
					ID:   "test",
					Type: "discord",
					Options: map[string]string{
						"test": "1"},
					URLFields: map[string]string{
						"test": "2"},
					Params: map[string]string{
						"test": "3"}}},
		},
		"multiple": {
			input: &shoutrrr.Slice{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"test": "1"},
					URLFields: map[string]string{
						"test": "2"},
					Params: map[string]string{
						"test": "3"}},
				"other": {
					Type: "discord",
					Options: map[string]string{
						"message": "release {{ version }} is available"},
					URLFields: map[string]string{
						"apikey": "censor?"},
					Params: map[string]string{
						"devices": "censor this too",
						"avatar":  "https://example.com"}}},
			want: &api_type.NotifySlice{
				"test": {
					ID:   "test",
					Type: "discord",
					Options: map[string]string{
						"test": "1"},
					URLFields: map[string]string{
						"test": "2"},
					Params: map[string]string{
						"test": "3"}},
				"other": {
					ID:   "other",
					Type: "discord",
					Options: map[string]string{
						"message": "release {{ version }} is available"},
					URLFields: map[string]string{
						"apikey": "<secret>"},
					Params: map[string]string{
						"devices": "<secret>",
						"avatar":  "https://example.com"}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorNotifySlice is called
			got := convertAndCensorNotifySlice(tc.input)

			// THEN the result should be as expected
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertAndCensorWebHookSlice(t *testing.T) {
	// GIVEN a webhook.Slice
	tests := map[string]struct {
		input *webhook.Slice
		want  *api_type.WebHookSlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &webhook.Slice{},
			want:  &api_type.WebHookSlice{},
		},
		"one": {
			input: &webhook.Slice{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: "censor",
					CustomHeaders: &webhook.Headers{
						{Key: "X-Test", Value: "foo"}}}},
			want: &api_type.WebHookSlice{
				"test": {
					Type:   stringPtr("github"),
					URL:    stringPtr("https://example.com"),
					Secret: stringPtr("<secret>"),
					CustomHeaders: []api_type.Header{
						{Key: "X-Test", Value: "<secret>"}}}},
		},
		"multiple": {
			input: &webhook.Slice{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: "censor",
					CustomHeaders: &webhook.Headers{
						{Key: "X-Test", Value: "foo"}}},
				"other": {
					Type: "gitlab",
					URL:  "https://release-argus.io"}},
			want: &api_type.WebHookSlice{
				"test": {
					Type:   stringPtr("github"),
					URL:    stringPtr("https://example.com"),
					Secret: stringPtr("<secret>"),
					CustomHeaders: []api_type.Header{
						{Key: "X-Test", Value: "<secret>"}}},
				"other": {
					Type: stringPtr("gitlab"),
					URL:  stringPtr("https://release-argus.io")}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorWebHookSlice is called
			got := convertAndCensorWebHookSlice(tc.input)

			// THEN the result should be as expected
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertAndCensorDefaults(t *testing.T) {
	// GIVEN a config.Defaults
	tests := map[string]struct {
		input *config.Defaults
		want  *api_type.Defaults
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"bare": {
			input: &config.Defaults{
				Service: service.Service{
					Options:               opt.Options{},
					LatestVersion:         latestver.Lookup{},
					DeployedVersionLookup: &deployedver.Lookup{},
					Dashboard:             service.DashboardOptions{}},
			},
			want: &api_type.Defaults{
				Service: api_type.Service{
					Options:               &api_type.ServiceOptions{},
					LatestVersion:         &api_type.LatestVersion{},
					DeployedVersionLookup: &api_type.DeployedVersionLookup{},
					Dashboard:             &api_type.DashboardOptions{}}},
		},
		"censor service.latest_version.access_token": {
			input: &config.Defaults{
				Service: service.Service{
					Options: opt.Options{},
					LatestVersion: latestver.Lookup{
						AccessToken: stringPtr("censor")},
					DeployedVersionLookup: &deployedver.Lookup{},
					Dashboard:             service.DashboardOptions{}},
			},
			want: &api_type.Defaults{
				Service: api_type.Service{
					Options: &api_type.ServiceOptions{},
					LatestVersion: &api_type.LatestVersion{
						AccessToken: "<secret>"},
					DeployedVersionLookup: &api_type.DeployedVersionLookup{},
					Dashboard:             &api_type.DashboardOptions{}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorDefaults is called
			got := convertAndCensorDefaults(tc.input)

			// THEN the result should be as expected
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}
