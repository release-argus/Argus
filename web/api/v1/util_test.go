// Copyright [2023] [Argus]
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
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestConvertAndCensorNotifySliceDefaults(t *testing.T) {
	// GIVEN a shoutrrr.SliceDefaults
	tests := map[string]struct {
		input *shoutrrr.SliceDefaults
		want  *api_type.NotifySlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &shoutrrr.SliceDefaults{},
			want:  &api_type.NotifySlice{},
		},
		"one": {
			input: &shoutrrr.SliceDefaults{
				"test": shoutrrr.NewDefaults(
					"discord",
					&map[string]string{
						"test": "1"},
					&map[string]string{
						"test": "3"},
					&map[string]string{
						"test": "2"})},
			want: &api_type.NotifySlice{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"test": "1"},
					URLFields: map[string]string{
						"test": "2"},
					Params: map[string]string{
						"test": "3"}}},
		},
		"multiple": {
			input: &shoutrrr.SliceDefaults{
				"test": shoutrrr.NewDefaults(
					"discord",
					&map[string]string{
						"test": "1"},
					&map[string]string{
						"test": "3"},
					&map[string]string{
						"test": "2"}),
				"other": shoutrrr.NewDefaults(
					"discord",
					&map[string]string{
						"message": "release {{ version }} is available"},
					&map[string]string{
						"devices": "censor this too",
						"avatar":  "https://example.com"},
					&map[string]string{
						"apikey": "censor?"})},
			want: &api_type.NotifySlice{
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

			// WHEN convertAndCensorNotifySliceDefaults is called
			got := convertAndCensorNotifySliceDefaults(tc.input)

			// THEN the result should be as expected
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

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
				"test": shoutrrr.New(
					nil, "",
					&map[string]string{
						"test": "1"},
					&map[string]string{
						"test": "3"},
					"discord",
					&map[string]string{
						"test": "2"},
					nil, nil, nil)},
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
				"test": shoutrrr.New(
					nil, "",
					&map[string]string{
						"test": "1"},
					&map[string]string{
						"test": "3"},
					"discord",
					&map[string]string{
						"test": "2"},
					nil, nil, nil),
				"other": shoutrrr.New(
					nil, "",
					&map[string]string{
						"message": "release {{ version }} is available"},
					&map[string]string{
						"devices": "censor this too",
						"avatar":  "https://example.com"},
					"discord",
					&map[string]string{
						"apikey": "censor?"},
					nil, nil, nil)},
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

func TestConvertAndCensorWebHook(t *testing.T) {
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
			wh: webhook.New(
				nil, nil, "", nil, nil, nil, nil, nil,
				"shazam",
				nil, "", "", nil, nil, nil),
			want: &api_type.WebHook{
				Secret: stringPtr("<secret>")},
		},
		"copy and censor headers": {
			wh: webhook.New(
				nil,
				&webhook.Headers{
					{Key: "X-Something", Value: "foo"},
					{Key: "X-Another", Value: "bar"}},
				"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
			want: &api_type.WebHook{
				CustomHeaders: &[]api_type.Header{
					{Key: "X-Something", Value: "<secret>"},
					{Key: "X-Another", Value: "<secret>"}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorWebHook is called on it
			got := convertAndCensorWebHook(tc.wh)

			// THEN the WebHook is converted correctly
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertAndCensorWebHookSliceDefaults(t *testing.T) {
	// GIVEN a webhook.SliceDefaults
	tests := map[string]struct {
		input *webhook.SliceDefaults
		want  *api_type.WebHookSlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &webhook.SliceDefaults{},
			want:  &api_type.WebHookSlice{},
		},
		"nil and empty elements": {
			input: &webhook.SliceDefaults{
				"test":  &webhook.WebHookDefaults{},
				"other": nil},
			want: &api_type.WebHookSlice{
				"test":  {},
				"other": nil},
		},
		"one": {
			input: &webhook.SliceDefaults{
				"test": webhook.NewDefaults(
					nil,
					&webhook.Headers{
						{Key: "X-Test", Value: "foo"}},
					"", nil, nil,
					"censor",
					nil,
					"github",
					"https://example.com")},
			want: &api_type.WebHookSlice{
				"test": {
					Type:   stringPtr("github"),
					URL:    stringPtr("https://example.com"),
					Secret: stringPtr("<secret>"),
					CustomHeaders: &[]api_type.Header{
						{Key: "X-Test", Value: "<secret>"}}}},
		},
		"multiple": {
			input: &webhook.SliceDefaults{
				"test": webhook.NewDefaults(
					nil,
					&webhook.Headers{
						{Key: "X-Test", Value: "foo"}},
					"", nil, nil,
					"censor",
					nil,
					"github",
					"https://example.com"),
				"other": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil,
					"gitlab",
					"https://release-argus.io")},
			want: &api_type.WebHookSlice{
				"test": {
					Type:   stringPtr("github"),
					URL:    stringPtr("https://example.com"),
					Secret: stringPtr("<secret>"),
					CustomHeaders: &[]api_type.Header{
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

			// WHEN convertAndCensorWebHookSliceDefaults is called
			got := convertAndCensorWebHookSliceDefaults(tc.input)

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
				"test": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "X-Test", Value: "foo"}},
					"",
					nil, nil, nil, nil, nil,
					"censor",
					nil,
					"github",
					"https://example.com",
					nil, nil, nil)},
			want: &api_type.WebHookSlice{
				"test": {
					Type:   stringPtr("github"),
					URL:    stringPtr("https://example.com"),
					Secret: stringPtr("<secret>"),
					CustomHeaders: &[]api_type.Header{
						{Key: "X-Test", Value: "<secret>"}}}},
		},
		"multiple": {
			input: &webhook.Slice{
				"test": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "X-Test", Value: "foo"}},
					"", nil, nil, nil, nil, nil,
					"censor",
					nil,
					"github",
					"https://example.com",
					nil, nil, nil),
				"other": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"",
					nil,
					"gitlab",
					"https://release-argus.io",
					nil, nil, nil)},
			want: &api_type.WebHookSlice{
				"test": {
					Type:   stringPtr("github"),
					URL:    stringPtr("https://example.com"),
					Secret: stringPtr("<secret>"),
					CustomHeaders: &[]api_type.Header{
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

func TestConvertAndCensorLatestVersionRequireDefaults(t *testing.T) {
	// GIVEN a filter.RequireDefaults
	tests := map[string]struct {
		input *filter.RequireDefaults
		want  *api_type.LatestVersionRequireDefaults
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"bare": {
			input: &filter.RequireDefaults{},
			want:  &api_type.LatestVersionRequireDefaults{},
		},
		"bare with bare Docker": {
			input: &filter.RequireDefaults{
				Docker: filter.DockerCheckDefaults{}},
			want: &api_type.LatestVersionRequireDefaults{
				Docker: api_type.RequireDockerCheckDefaults{}},
		},
		"docker.ghcr": {
			input: &filter.RequireDefaults{
				Docker: *filter.NewDockerCheckDefaults(
					"quay",
					"tokenForGHCR",
					"",
					"usernameForHub",
					"",
					nil)},
			want: &api_type.LatestVersionRequireDefaults{
				Docker: api_type.RequireDockerCheckDefaults{
					Type: "quay",
					GHCR: &api_type.RequireDockerCheckRegistryDefaults{
						Token: "<secret>"},
					Hub: &api_type.RequireDockerCheckRegistryDefaultsWithUsername{
						Username: "usernameForHub"}}},
		},
		"docker.hub": {
			input: &filter.RequireDefaults{
				Docker: *filter.NewDockerCheckDefaults(
					"ghcr",
					"",
					"tokenForHub",
					"",
					"",
					nil)},
			want: &api_type.LatestVersionRequireDefaults{
				Docker: api_type.RequireDockerCheckDefaults{
					Type: "ghcr",
					Hub: &api_type.RequireDockerCheckRegistryDefaultsWithUsername{
						RequireDockerCheckRegistryDefaults: api_type.RequireDockerCheckRegistryDefaults{
							Token: "<secret>"}}}},
		},
		"docker.quay": {
			input: &filter.RequireDefaults{
				Docker: *filter.NewDockerCheckDefaults(
					"quay",
					"",
					"",
					"",
					"tokenForQuay",
					nil)},
			want: &api_type.LatestVersionRequireDefaults{
				Docker: api_type.RequireDockerCheckDefaults{
					Type: "quay",
					Quay: &api_type.RequireDockerCheckRegistryDefaults{
						Token: "<secret>"}}},
		},
		"filled": {
			input: &filter.RequireDefaults{
				Docker: *filter.NewDockerCheckDefaults(
					"quay",
					"tokenForGHCR",
					"tokenForHub",
					"usernameForHub",
					"tokenForQuay",
					nil)},
			want: &api_type.LatestVersionRequireDefaults{
				Docker: api_type.RequireDockerCheckDefaults{
					Type: "quay",
					GHCR: &api_type.RequireDockerCheckRegistryDefaults{
						Token: "<secret>"},
					Hub: &api_type.RequireDockerCheckRegistryDefaultsWithUsername{
						RequireDockerCheckRegistryDefaults: api_type.RequireDockerCheckRegistryDefaults{
							Token: "<secret>"},
						Username: "usernameForHub"},
					Quay: &api_type.RequireDockerCheckRegistryDefaults{
						Token: "<secret>"}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorLatestVersionRequireDefaults is called
			got := convertAndCensorLatestVersionRequireDefaults(tc.input)

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
				Service: service.Defaults{
					Options:               opt.OptionsDefaults{},
					LatestVersion:         latestver.LookupDefaults{},
					DeployedVersionLookup: deployedver.LookupDefaults{},
					Dashboard:             service.DashboardOptionsDefaults{}},
			},
			want: &api_type.Defaults{
				Service: api_type.ServiceDefaults{
					Service: api_type.Service{
						Options:               &api_type.ServiceOptions{},
						DeployedVersionLookup: &api_type.DeployedVersionLookup{},
						Dashboard:             &api_type.DashboardOptions{}},
					LatestVersion: &api_type.LatestVersionDefaults{
						Require: &api_type.LatestVersionRequireDefaults{}}}},
		},
		"censor service.latest_version": {
			input: &config.Defaults{
				Service: service.Defaults{
					Options: opt.OptionsDefaults{},
					LatestVersion: *latestver.NewDefaults(
						stringPtr("censor"),
						nil, nil,
						filter.NewRequireDefaults(
							filter.NewDockerCheckDefaults(
								"ghcr",
								"tokenGHCR",
								"tokenHub",
								"usernameHub",
								"tokenQuay",
								nil))),
					DeployedVersionLookup: deployedver.LookupDefaults{},
					Dashboard:             service.DashboardOptionsDefaults{}},
			},
			want: &api_type.Defaults{
				Service: api_type.ServiceDefaults{
					Service: api_type.Service{
						Options:               &api_type.ServiceOptions{},
						DeployedVersionLookup: &api_type.DeployedVersionLookup{},
						Dashboard:             &api_type.DashboardOptions{}},
					LatestVersion: &api_type.LatestVersionDefaults{
						LatestVersion: api_type.LatestVersion{
							AccessToken: "<secret>"},
						Require: &api_type.LatestVersionRequireDefaults{
							Docker: api_type.RequireDockerCheckDefaults{
								Type: "ghcr",
								GHCR: &api_type.RequireDockerCheckRegistryDefaults{
									Token: "<secret>"},
								Hub: &api_type.RequireDockerCheckRegistryDefaultsWithUsername{
									RequireDockerCheckRegistryDefaults: api_type.RequireDockerCheckRegistryDefaults{
										Token: "<secret>"},
									Username: "usernameHub"},
								Quay: &api_type.RequireDockerCheckRegistryDefaults{
									Token: "<secret>"},
							}}}}},
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
