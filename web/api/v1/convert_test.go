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

package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	deployedver_base "github.com/release-argus/Argus/service/deployed_version/types/base"
	dv_web "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	lv_github "github.com/release-argus/Argus/service/latest_version/types/github"
	lv_web "github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

//
// Defaults.
//

func TestConvertAndCensorDefaults(t *testing.T) {
	// GIVEN a config.Defaults.
	tests := map[string]struct {
		input *config.Defaults
		want  *apitype.Defaults
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"bare": {
			input: &config.Defaults{
				Service: service.Defaults{
					Options:               opt.Defaults{},
					LatestVersion:         latestver_base.Defaults{},
					DeployedVersionLookup: deployedver_base.Defaults{},
					Dashboard:             service.DashboardOptionsDefaults{}},
			},
			want: &apitype.Defaults{
				Service: apitype.ServiceDefaults{
					Options: &apitype.ServiceOptions{},
					LatestVersion: &apitype.LatestVersionDefaults{
						Require: &apitype.LatestVersionRequireDefaults{}},
					DeployedVersionLookup: &apitype.DeployedVersionLookup{},
					Dashboard:             &apitype.DashboardOptions{}}},
		},
		"censor service.latest_version": {
			input: &config.Defaults{
				Service: service.Defaults{
					Options: opt.Defaults{},
					LatestVersion: latestver_base.Defaults{
						AccessToken: "censor",
						Require: filter.RequireDefaults{
							Docker: *filter.NewDockerCheckDefaults(
								"ghcr",
								"tokenGHCR",
								"tokenHub",
								"usernameHub",
								"tokenQuay",
								nil)}},
					DeployedVersionLookup: deployedver_base.Defaults{},
					Dashboard:             service.DashboardOptionsDefaults{}},
			},
			want: &apitype.Defaults{
				Service: apitype.ServiceDefaults{
					Options: &apitype.ServiceOptions{},
					LatestVersion: &apitype.LatestVersionDefaults{
						AccessToken: util.SecretValue,
						Require: &apitype.LatestVersionRequireDefaults{
							Docker: apitype.RequireDockerCheckDefaults{
								Type: "ghcr",
								GHCR: &apitype.RequireDockerCheckRegistryDefaults{
									Token: util.SecretValue},
								Hub: &apitype.RequireDockerCheckRegistryDefaultsWithUsername{
									RequireDockerCheckRegistryDefaults: apitype.RequireDockerCheckRegistryDefaults{
										Token: util.SecretValue},
									Username: "usernameHub"},
								Quay: &apitype.RequireDockerCheckRegistryDefaults{
									Token: util.SecretValue},
							}}},
					DeployedVersionLookup: &apitype.DeployedVersionLookup{},
					Dashboard:             &apitype.DashboardOptions{}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorDefaults is called.
			got := convertAndCensorDefaults(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

//
// Service.
//

func TestConvertAndCensorService(t *testing.T) {
	// GIVEN a service.Service.
	tests := map[string]struct {
		input *service.Service
		want  *apitype.Service
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"bare": {
			input: &service.Service{},
			want: &apitype.Service{
				Options:       &apitype.ServiceOptions{},
				LatestVersion: nil,
				Command:       &apitype.CommandSlice{},
				Notify:        &apitype.NotifySlice{},
				WebHook:       &apitype.WebHookSlice{},
				Dashboard:     &apitype.DashboardOptions{}},
		},
		"name ignored when no marshalName": {
			input: &service.Service{
				ID:      "Test",
				Name:    "Test",
				Comment: "Hi",
			},
			want: &apitype.Service{
				Comment:       "Hi",
				Options:       &apitype.ServiceOptions{},
				LatestVersion: nil,
				Command:       &apitype.CommandSlice{},
				Notify:        &apitype.NotifySlice{},
				WebHook:       &apitype.WebHookSlice{},
				Dashboard:     &apitype.DashboardOptions{}},
		},
		"all fields": {
			input: &service.Service{
				ID:      "Test",
				Name:    "Something",
				Comment: "Comment on the Service",
				Options: opt.Options{
					Active: test.BoolPtr(false)},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: lv_accessToken
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							allow_invalid_certs: true
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"gotify": shoutrrr.New(
						nil,
						"gotify",
						"",
						nil,
						map[string]string{
							"url": "http://gotify"},
						nil,
						nil, nil, nil)},
				Command: command.Slice{
					{"echo", "foo"}},
				WebHook: webhook.Slice{
					"test_wh": webhook.New(
						test.BoolPtr(true),
						nil, "", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
				Dashboard: *service.NewDashboardOptions(
					nil,
					"https://example.com/icon.png",
					"", "", nil,
					nil, nil),
				Status: *status.New(
					nil, nil, nil,
					"2.0.0",
					"1.0.0", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					"3.0.0", time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)),
			},
			want: &apitype.Service{
				Name:    "Something",
				Comment: "Comment on the Service",
				Options: &apitype.ServiceOptions{
					Active: test.BoolPtr(false)},
				LatestVersion: &apitype.LatestVersion{
					Type:        "github",
					AccessToken: util.SecretValue,
					URLCommands: &apitype.URLCommandSlice{}},
				Command: &apitype.CommandSlice{
					{"echo", "foo"}},
				Notify: &apitype.NotifySlice{
					"gotify": &apitype.Notify{
						URLFields: map[string]string{
							"url": "http://gotify"}}},
				WebHook: &apitype.WebHookSlice{
					"test_wh": &apitype.WebHook{
						AllowInvalidCerts: test.BoolPtr(true)}},
				DeployedVersionLookup: &apitype.DeployedVersionLookup{
					AllowInvalidCerts: test.BoolPtr(true)},
				Dashboard: &apitype.DashboardOptions{
					Icon: "https://example.com/icon.png"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// If it has a Name that is different from the ID.
			if tc.input != nil && tc.input.Name != "" && tc.input.ID != tc.input.Name {
				// Re-marshal so that Name will unmarshal.
				serviceJSON, _ := json.Marshal(tc.input)
				serviceJSON = []byte(strings.Replace(string(serviceJSON),
					"{", `{"name":"`+tc.input.Name+`",`, 1))
				json.Unmarshal(serviceJSON, tc.input)
			}

			// WHEN convertAndCensorService is called.
			got := convertAndCensorService(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

//
// Latest Version.
//

func TestConvertAndCensorLatestVersionRequireDefaults(t *testing.T) {
	// GIVEN a filter.RequireDefaults.
	tests := map[string]struct {
		input *filter.RequireDefaults
		want  *apitype.LatestVersionRequireDefaults
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"bare": {
			input: &filter.RequireDefaults{},
			want:  &apitype.LatestVersionRequireDefaults{},
		},
		"bare with bare Docker": {
			input: &filter.RequireDefaults{
				Docker: filter.DockerCheckDefaults{}},
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerCheckDefaults{}},
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
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerCheckDefaults{
					Type: "quay",
					GHCR: &apitype.RequireDockerCheckRegistryDefaults{
						Token: util.SecretValue},
					Hub: &apitype.RequireDockerCheckRegistryDefaultsWithUsername{
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
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerCheckDefaults{
					Type: "ghcr",
					Hub: &apitype.RequireDockerCheckRegistryDefaultsWithUsername{
						RequireDockerCheckRegistryDefaults: apitype.RequireDockerCheckRegistryDefaults{
							Token: util.SecretValue}}}},
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
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerCheckDefaults{
					Type: "quay",
					Quay: &apitype.RequireDockerCheckRegistryDefaults{
						Token: util.SecretValue}}},
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
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerCheckDefaults{
					Type: "quay",
					GHCR: &apitype.RequireDockerCheckRegistryDefaults{
						Token: util.SecretValue},
					Hub: &apitype.RequireDockerCheckRegistryDefaultsWithUsername{
						RequireDockerCheckRegistryDefaults: apitype.RequireDockerCheckRegistryDefaults{
							Token: util.SecretValue},
						Username: "usernameForHub"},
					Quay: &apitype.RequireDockerCheckRegistryDefaults{
						Token: util.SecretValue}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorLatestVersionRequireDefaults is called.
			got := convertAndCensorLatestVersionRequireDefaults(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertAndCensorLatestVersion(t *testing.T) {
	// GIVEN a latestver.Lookup.
	tests := map[string]struct {
		input latestver.Lookup
		want  string
	}{
		"nil": {
			input: nil,
			want:  "",
		},
		"github - bare": {
			input: &lv_github.Lookup{},
			want:  `{"url_commands":[]}`,
		},
		"github - filled": {
			input: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", test.TrimYAML(`
						url: release-argus/argus
						access_token: not_telling_you
						use_prerelease: false
						url_commands:
							- type: replace
								old: this
								new: withThis
							- type: split
								text: splitThis
								index: 8
							- type: regex
								regex: ([0-9.]+)
						require:
							regex_content: .*
					`),
					nil,
					nil,
					nil, nil)
			}),
			want: test.TrimJSON(`{
				"type": "github",
				"url": "release-argus/argus",
				"access_token": ` + secretValueMarshalled + `,
				"use_prerelease": false,
				"url_commands": [
					{"type": "replace", "new": "withThis", "old": "this"},
					{"type": "split", "index": 8, "text": "splitThis"},
					{"type": "regex", "regex": "([0-9.]+)"}
				],
				"require": {
					"regex_content": ".*"
				}
			}`),
		},
		"url - bare": {
			input: &lv_web.Lookup{},
			want:  `{"url_commands":[]}`,
		},
		"url - filled": {
			input: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"url",
					"yaml", test.TrimYAML(`
						allow_invalid_certs: true
						url: https://example.com
						url_commands:
							- type: replace
								old: this
								new: withThis
							- type: split
								text: splitThis
								index: 8
							- type: regex
								regex: ([0-9.]+)
						require:
							docker:
								type: ghcr
								image: release-argus/argus
								tag: '{{ version }}'
								token: not_telling_you
					`),
					nil,
					nil,
					nil, nil)
			}),
			want: test.TrimJSON(`{
				"type": "url",
				"url": "https://example.com",
				"allow_invalid_certs": true,
				"url_commands": [
					{"type": "replace", "new": "withThis", "old": "this"},
					{"type": "split", "index": 8, "text": "splitThis"},
					{"type": "regex", "regex": "([0-9.]+)"}
				],
				"require": {
					"docker": {
						"type": "ghcr",
						"image": "release-argus/argus",
						"tag": "{{ version }}",
						"token": ` + secretValueMarshalled + `
					}
				}
			}`),
		},
		"unknown": {
			input: &latestver_base.Lookup{},
			want:  "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorLatestVersion is called.
			got := convertAndCensorLatestVersion(tc.input)

			// THEN the result should be as expected.
			if tc.want != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want, got.String())
			}
		})
	}
}

func TestConvertAndCensorLatestVersionRequire(t *testing.T) {
	// GIVEN a filter.Require.
	tests := map[string]struct {
		input *filter.Require
		want  *apitype.LatestVersionRequire
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"bare": {
			input: &filter.Require{},
			want:  &apitype.LatestVersionRequire{},
		},
		"bare with bare Docker": {
			input: &filter.Require{
				Docker: &filter.DockerCheck{}},
			want: &apitype.LatestVersionRequire{
				Docker: &apitype.RequireDockerCheck{}},
		},
		"docker.ghcr": {
			input: &filter.Require{
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus", "{{ version }}",
					"", "tokenForGHCR",
					"", time.Now(),
					nil)},
			want: &apitype.LatestVersionRequire{
				Docker: &apitype.RequireDockerCheck{
					Type:  "ghcr",
					Image: "release-argus/argus",
					Tag:   "{{ version }}",
					Token: util.SecretValue}},
		},
		"docker.hub": {
			input: &filter.Require{
				Docker: filter.NewDockerCheck(
					"hub",
					"release-argus/argus", "{{ version }}",
					"user", "tokenForHub",
					"", time.Now(),
					nil)},
			want: &apitype.LatestVersionRequire{
				Docker: &apitype.RequireDockerCheck{
					Type:     "hub",
					Image:    "release-argus/argus",
					Tag:      "{{ version }}",
					Username: "user",
					Token:    util.SecretValue}},
		},
		"filled": {
			input: &filter.Require{
				Status: status.New(
					nil, nil, nil,
					"2.0.0",
					"1.0.0", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					"3.0.0", time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)),
				RegexContent: ".*",
				RegexVersion: `([0-9.]+)`,
				Command:      command.Command{"echo", "hello"},
				Docker: filter.NewDockerCheck(
					"hub",
					"release-argus/argus", "{{ version }}",
					"user", "tokenForHub",
					"", time.Now(),
					nil)},
			want: &apitype.LatestVersionRequire{
				Command: []string{"echo", "hello"},
				Docker: &apitype.RequireDockerCheck{
					Type:     "hub",
					Image:    "release-argus/argus",
					Tag:      "{{ version }}",
					Username: "user",
					Token:    util.SecretValue},
				RegexContent: ".*",
				RegexVersion: `([0-9.]+)`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorLatestVersionRequire is called.
			got := convertAndCensorLatestVersionRequire(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertURLCommandSlice(t *testing.T) {
	// GIVEN a URL Command slice.
	tests := map[string]struct {
		slice *filter.URLCommandSlice
		want  *apitype.URLCommandSlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &filter.URLCommandSlice{},
			want:  &apitype.URLCommandSlice{},
		},
		"regex": {
			slice: &filter.URLCommandSlice{
				{Type: "regex", Regex: "[0-9.]+"}},
			want: &apitype.URLCommandSlice{
				{Type: "regex", Regex: "[0-9.]+"}},
		},
		"replace": {
			slice: &filter.URLCommandSlice{
				{Type: "replace", Old: "foo", New: test.StringPtr("bar")}},
			want: &apitype.URLCommandSlice{
				{Type: "replace", Old: "foo", New: test.StringPtr("bar")}},
		},
		"split": {
			slice: &filter.URLCommandSlice{
				{Type: "split", Index: test.IntPtr(7)}},
			want: &apitype.URLCommandSlice{
				{Type: "split", Index: test.IntPtr(7)}},
		},
		"one of each": {
			slice: &filter.URLCommandSlice{
				{Type: "regex", Regex: "[0-9.]+"},
				{Type: "replace", Old: "foo", New: test.StringPtr("bar")},
				{Type: "split", Index: test.IntPtr(7)}},
			want: &apitype.URLCommandSlice{
				{Type: "regex", Regex: "[0-9.]+"},
				{Type: "replace", Old: "foo", New: test.StringPtr("bar")},
				{Type: "split", Index: test.IntPtr(7)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertURLCommandSlice is called on it.
			got := convertURLCommandSlice(tc.slice)

			// THEN the WebHookSlice is converted correctly.
			if got.String() != tc.want.String() {
				t.Errorf("want: %q, got: %q",
					tc.want.String(), got.String())
			}
		})
	}
}

//
// Deployed Version.
//

func TestConvertAndCensorDeployedVersionLookup(t *testing.T) {
	// GIVEN a DeployedVersionLookup.
	tests := map[string]struct {
		dvl                                       deployedver.Lookup
		dvlStatus                                 *status.Status
		approvedVersion                           string
		deployedVersion, deployedVersionTimestamp string
		latestVersion, latestVersionTimestamp     string
		lastQueried                               string
		regexMissesContent, regexMissesVersion    int

		want *apitype.DeployedVersionLookup
	}{
		"nil": {
			dvl:  nil,
			want: nil,
		},
		"empty": {
			dvl:  &dv_web.Lookup{},
			want: &apitype.DeployedVersionLookup{},
		},
		"minimal": {
			dvl: &dv_web.Lookup{
				URL:  "https://example.com",
				JSON: "version"},
			want: &apitype.DeployedVersionLookup{
				URL:  "https://example.com",
				JSON: "version"},
		},
		"censor basic_auth.password": {
			dvl: &dv_web.Lookup{
				URL: "https://example.com",
				BasicAuth: &dv_web.BasicAuth{
					Username: "alan",
					Password: "pass123"}},
			want: &apitype.DeployedVersionLookup{
				URL: "https://example.com",
				BasicAuth: &apitype.BasicAuth{
					Username: "alan",
					Password: util.SecretValue}},
		},
		"censor headers": {
			dvl: &dv_web.Lookup{
				URL: "https://example.com",
				Headers: []dv_web.Header{
					{Key: "X-Test-0", Value: "foo"},
					{Key: "X-Test-1", Value: "bar"}}},
			want: &apitype.DeployedVersionLookup{
				URL: "https://example.com",
				Headers: []apitype.Header{
					{Key: "X-Test-0", Value: util.SecretValue},
					{Key: "X-Test-1", Value: util.SecretValue},
				}},
		},
		"full": {
			regexMissesContent: 1,
			regexMissesVersion: 3,
			dvl: test.IgnoreError(t, func() (deployedver.Lookup, error) {
				return deployedver.New(
					"url",
					"yaml", test.TrimYAML(`
						method: POST
						url: https://release-argus.io
						allow_invalid_certs: true
						basic_auth:
							username: jim
							password: whoops
						body: body_here
						headers:
							- key: X-Test-0
								value: foo
							- key: X-Test-1
								value: bar
						json: version
						regex: ([0-9]+\.[0-9]+\.[0-9]+)
						regex_template: $1.$2.$3
					`),
					opt.New(
						test.BoolPtr(true), "10m", test.BoolPtr(true),
						&opt.Defaults{}, &opt.Defaults{}),
					&status.Status{},
					&deployedver_base.Defaults{}, &deployedver_base.Defaults{})
			}),
			dvlStatus: &status.Status{
				Fails:     status.Fails{},
				ServiceID: test.StringPtr("service-id"),
				WebURL:    test.StringPtr("https://release-argus.io")},
			want: &apitype.DeployedVersionLookup{
				Method:            http.MethodPost,
				URL:               "https://release-argus.io",
				AllowInvalidCerts: test.BoolPtr(true),
				BasicAuth: &apitype.BasicAuth{
					Username: "jim",
					Password: util.SecretValue},
				Headers: []apitype.Header{
					{Key: "X-Test-0", Value: util.SecretValue},
					{Key: "X-Test-1", Value: util.SecretValue}},
				Body:          "body_here",
				JSON:          "version",
				Regex:         `([0-9]+\.[0-9]+\.[0-9]+)`,
				RegexTemplate: "$1.$2.$3"},
		},
		"unknown type": {
			dvl:  &deployedver_base.Lookup{},
			want: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.dvl != nil {
				var dvStatus *status.Status
				if dv, ok := tc.dvl.(*dv_web.Lookup); ok {
					dvStatus = dv.Status
				}

				if tc.approvedVersion != "" {
					dvStatus.SetApprovedVersion("1.2.3", false)
					dvStatus.SetDeployedVersion("1.2.3", "", false)
					dvStatus.SetLatestVersion("1.2.3", time.Now().Format(time.RFC3339), false)
					dvStatus.SetLastQueried(time.Now().Format(time.RFC3339))
				}
				if tc.dvlStatus != nil {
					dvStatus.Fails.Copy(&tc.dvlStatus.Fails)
					dvStatus.ServiceID = tc.dvlStatus.ServiceID
					dvStatus.WebURL = tc.dvlStatus.WebURL
				}
				for i := 0; i < tc.regexMissesContent; i++ {
					dvStatus.RegexMissContent()
				}
				for i := 0; i < tc.regexMissesVersion; i++ {
					dvStatus.RegexMissVersion()
				}
			}

			// WHEN convertAndCensorDeployedVersionLookup is called on it.
			got := convertAndCensorDeployedVersionLookup(tc.dvl)

			// THEN the WebHookSlice is converted correctly.
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

//
// Notify.
//

func TestConvertAndCensorNotifySliceDefaults(t *testing.T) {
	// GIVEN a shoutrrr.SliceDefaults.
	tests := map[string]struct {
		input *shoutrrr.SliceDefaults
		want  *apitype.NotifySlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &shoutrrr.SliceDefaults{},
			want:  &apitype.NotifySlice{},
		},
		"one": {
			input: &shoutrrr.SliceDefaults{
				"test": shoutrrr.NewDefaults(
					"discord",
					map[string]string{
						"test": "1"},
					map[string]string{
						"test": "2"},
					map[string]string{
						"test": "3"})},
			want: &apitype.NotifySlice{
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
					map[string]string{
						"test": "1"},
					map[string]string{
						"test": "2"},
					map[string]string{
						"test": "3"}),
				"other": shoutrrr.NewDefaults(
					"discord",
					map[string]string{
						"message": "release {{ version }} is available"},
					map[string]string{
						"apikey": "censor?"},
					map[string]string{
						"devices": "censor this too",
						"avatar":  "https://example.com"})},
			want: &apitype.NotifySlice{
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
						"apikey": util.SecretValue},
					Params: map[string]string{
						"devices": util.SecretValue,
						"avatar":  "https://example.com"}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorNotifySliceDefaults is called.
			got := convertAndCensorNotifySliceDefaults(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertAndCensorNotifySlice(t *testing.T) {
	// GIVEN a shoutrrr.Slice.
	tests := map[string]struct {
		input *shoutrrr.Slice
		want  *apitype.NotifySlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &shoutrrr.Slice{},
			want:  &apitype.NotifySlice{},
		},
		"one": {
			input: &shoutrrr.Slice{
				"test": shoutrrr.New(
					nil, "",
					"discord",
					map[string]string{
						"test": "1"},
					map[string]string{
						"test": "2"},
					map[string]string{
						"test": "3"},
					nil, nil, nil)},
			want: &apitype.NotifySlice{
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
			input: &shoutrrr.Slice{
				"test": shoutrrr.New(
					nil, "",
					"discord",
					map[string]string{
						"test": "1"},
					map[string]string{
						"test": "2"},
					map[string]string{
						"test": "3"},
					nil, nil, nil),
				"other": shoutrrr.New(
					nil, "",
					"discord",
					map[string]string{
						"message": "release {{ version }} is available"},
					map[string]string{
						"apikey": "censor?"},
					map[string]string{
						"devices": "censor this too",
						"avatar":  "https://example.com"},
					nil, nil, nil)},
			want: &apitype.NotifySlice{
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
						"apikey": util.SecretValue},
					Params: map[string]string{
						"devices": util.SecretValue,
						"avatar":  "https://example.com"}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorNotifySlice is called.
			got := convertAndCensorNotifySlice(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

//
// Command.
//

func TestConvertCommandSlice(t *testing.T) {
	// GIVEN a Command slice.
	tests := map[string]struct {
		slice *command.Slice
		want  *apitype.CommandSlice
	}{
		"nil": {
			slice: nil,
			want:  nil,
		},
		"empty": {
			slice: &command.Slice{},
			want:  &apitype.CommandSlice{},
		},
		"one": {
			slice: &command.Slice{
				{"ls", "-lah"}},
			want: &apitype.CommandSlice{
				{"ls", "-lah"}},
		},
		"two": {
			slice: &command.Slice{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"}},
			want: &apitype.CommandSlice{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertCommandSlice is called on it.
			got := convertCommandSlice(tc.slice)

			// THEN the CommandSlice is converted correctly.
			if got == tc.want { // both nil.
				return
			}
			// check number of commands.
			if len(*got) != len(*tc.want) {
				t.Errorf("want:\n%v\ngot:\n%v",
					tc.want, got)
				return
			}
			for cI := range *got {
				// check number of args.
				if len((*got)[cI]) != len((*tc.want)[cI]) {
					t.Errorf("want:\n%v\ngot:\n%v",
						tc.want, got)
				}
				// check args.
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

//
// WebHook.
//

func TestConvertAndCensorWebHookSliceDefaults(t *testing.T) {
	// GIVEN a webhook.SliceDefaults.
	tests := map[string]struct {
		input *webhook.SliceDefaults
		want  *apitype.WebHookSlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &webhook.SliceDefaults{},
			want:  &apitype.WebHookSlice{},
		},
		"nil and empty elements": {
			input: &webhook.SliceDefaults{
				"test":  &webhook.Defaults{},
				"other": nil},
			want: &apitype.WebHookSlice{
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
			want: &apitype.WebHookSlice{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					CustomHeaders: &[]apitype.Header{
						{Key: "X-Test", Value: util.SecretValue}}}},
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
			want: &apitype.WebHookSlice{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					CustomHeaders: &[]apitype.Header{
						{Key: "X-Test", Value: util.SecretValue}}},
				"other": {
					Type: "gitlab",
					URL:  "https://release-argus.io"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorWebHookSliceDefaults is called.
			got := convertAndCensorWebHookSliceDefaults(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertAndCensorWebHookSlice(t *testing.T) {
	// GIVEN a webhook.Slice.
	tests := map[string]struct {
		input *webhook.Slice
		want  *apitype.WebHookSlice
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &webhook.Slice{},
			want:  &apitype.WebHookSlice{},
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
			want: &apitype.WebHookSlice{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					CustomHeaders: &[]apitype.Header{
						{Key: "X-Test", Value: util.SecretValue}}}},
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
			want: &apitype.WebHookSlice{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					CustomHeaders: &[]apitype.Header{
						{Key: "X-Test", Value: util.SecretValue}}},
				"other": {
					Type: "gitlab",
					URL:  "https://release-argus.io"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorWebHookSlice is called.
			got := convertAndCensorWebHookSlice(tc.input)

			// THEN the result should be as expected.
			if tc.want.String() != got.String() {
				t.Errorf("want\n%q\ngot\n%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertAndCensorWebHook(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		wh   *webhook.WebHook
		want *apitype.WebHook
	}{
		"nil": {
			wh:   nil,
			want: nil,
		},
		"empty": {
			wh:   &webhook.WebHook{},
			want: &apitype.WebHook{},
		},
		"censor secret": {
			wh: webhook.New(
				nil, nil, "", nil, nil, nil, nil, nil,
				"shazam",
				nil, "", "", nil, nil, nil),
			want: &apitype.WebHook{
				Secret: util.SecretValue},
		},
		"copy and censor headers": {
			wh: webhook.New(
				nil,
				&webhook.Headers{
					{Key: "X-Something", Value: "foo"},
					{Key: "X-Another", Value: "bar"}},
				"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
			want: &apitype.WebHook{
				CustomHeaders: &[]apitype.Header{
					{Key: "X-Something", Value: util.SecretValue},
					{Key: "X-Another", Value: util.SecretValue}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertAndCensorWebHook is called on it.
			got := convertAndCensorWebHook(tc.wh)

			// THEN the WebHook is converted correctly.
			if got.String() != tc.want.String() {
				t.Errorf("want:\n%q\ngot:%q",
					tc.want.String(), got.String())
			}
		})
	}
}

func TestConvertWebHookHeaders(t *testing.T) {
	// GIVEN a webhook.Headers.
	tests := map[string]struct {
		input *webhook.Headers
		want  *[]apitype.Header
	}{
		"nil": {
			input: nil,
			want:  nil,
		},
		"empty": {
			input: &webhook.Headers{},
			want:  &[]apitype.Header{},
		},
		"one header": {
			input: &webhook.Headers{
				{Key: "X-Test", Value: "foo"}},
			want: &[]apitype.Header{
				{Key: "X-Test", Value: "foo"}},
		},
		"multiple headers": {
			input: &webhook.Headers{
				{Key: "X-Test-1", Value: "foo"},
				{Key: "X-Test-2", Value: "bar"}},
			want: &[]apitype.Header{
				{Key: "X-Test-1", Value: "foo"},
				{Key: "X-Test-2", Value: "bar"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN convertWebHookHeaders is called.
			got := convertWebHookHeaders(tc.input)

			// THEN the result should be as expected.
			if got == nil && tc.want == nil {
				return
			}
			if got == nil || tc.want == nil {
				t.Errorf("want: %v, got: %v", tc.want, got)
				return
			}
			if len(*got) != len(*tc.want) {
				t.Errorf("want: %v, got: %v", tc.want, got)
				return
			}
			for i := range *got {
				if (*got)[i] != (*tc.want)[i] {
					t.Errorf("want: %v, got: %v", tc.want, got)
					return
				}
			}
		})
	}
}
