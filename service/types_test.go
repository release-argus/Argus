// Copyright [2024] [Argus]
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

package service

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
	"gopkg.in/yaml.v3"
)

func TestService_String(t *testing.T) {
	tests := map[string]struct {
		svc  *Service
		want string
	}{
		"nil": {
			svc:  nil,
			want: "",
		},
		"empty": {
			svc:  &Service{},
			want: "{}",
		},
		"all fields defined": {
			svc: &Service{
				Comment: "svc for blah",
				Options: opt.Options{
					Active: test.BoolPtr(false)},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
						url: release-argus/Argus
						url_commands:
							- type: regex
								regex: foo
								index: 1
						require:
							regex_version: v.+
							docker:
								image: releaseargus/argus
								tag: '{{ version }}'
					`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (*deployedver.Lookup, error) {
					return deployedver.New(
						"yaml", test.TrimYAML(`
						url: https://valid.release-argus.io/plain
						basic_auth:
							username: user
							password: pass
						headers:
							- key: foo
								value: bar
						json: version
					`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "",
						"discord",
						nil,
						map[string]string{
							"token": "bar"},
						nil,
						nil, nil, nil)},
				Command: command.Slice{
					{"ls", "-la"}},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"https://example.com",
						nil, nil, nil)},
				Dashboard: *NewDashboardOptions(
					test.BoolPtr(true), "", "", "",
					nil, nil),
				Defaults: &Defaults{
					Options: *opt.NewDefaults(
						"", test.BoolPtr(false))},
				HardDefaults: &Defaults{
					Options: *opt.NewDefaults(
						"", test.BoolPtr(false))}},
			want: test.TrimYAML(`
				comment: svc for blah
				options:
					active: false
				latest_version:
					type: github
					url: release-argus/Argus
					url_commands:
						- type: regex
							regex: foo
							index: 1
					require:
						regex_version: v.+
						docker:
							image: releaseargus/argus
							tag: '{{ version }}'
				deployed_version:
					url: https://valid.release-argus.io/plain
					basic_auth:
						username: user
						password: pass
					headers:
						- key: foo
							value: bar
					json: version
				notify:
					foo:
						type: discord
						url_fields:
							token: bar
				command:
					- - ls
						- -la
				webhook:
					foo:
						type: github
						url: https://example.com
				dashboard:
					auto_approve: true`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Service is stringified with String
				got := tc.svc.String(prefix)

				// THEN the result is as expected
				if got != want {
					t.Errorf("Service.String() mismatch (prefix=%q)\nwant: %q\ngot:  %q",
						prefix, want, got)
					return // no need to check other prefixes
				}
			}
		})
	}
}

func TestService_Summary(t *testing.T) {
	// GIVEN a Service
	tests := map[string]struct {
		svc                                       *Service
		approvedVersion                           string
		deployedVersion, deployedVersionTimestamp string
		latestVersion, latestVersionTimestamp     string
		lastQueried                               string
		want                                      *apitype.ServiceSummary
	}{
		"nil": {
			svc:  nil,
			want: nil,
		},
		"empty": {
			svc: &Service{},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only id": {
			svc: &Service{
				ID: "foo"},
			want: &apitype.ServiceSummary{
				ID:                       "foo",
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only options.active": {
			svc: &Service{
				Options: opt.Options{
					Active: test.BoolPtr(false)}},
			want: &apitype.ServiceSummary{
				Active:                   test.BoolPtr(false),
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only latest_version.type": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				})},
			want: &apitype.ServiceSummary{
				Type:                     ("github"),
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, and it's a url": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "https://example.com/icon.png"}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     ("https://example.com/icon.png"),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, and it's not a url": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "smile"}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, from notify": {
			svc: &Service{
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "",
						nil,
						nil,
						map[string]string{
							"icon": "https://example.com/notify.png"},
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil))}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     ("https://example.com/notify.png"),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, dashboard overrides notify": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "https://example.com/icon.png"},
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "",
						map[string]string{
							"icon": "https://example.com/notify.png"},
						nil, nil,
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil),
						shoutrrr.NewDefaults(
							"", nil, nil, nil))}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     ("https://example.com/icon.png"),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon_link_to": {
			svc: &Service{
				Dashboard: DashboardOptions{
					IconLinkTo: "https://example.com"}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               ("https://example.com"),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"only deployed_version": {
			svc: &Service{
				DeployedVersionLookup: &deployedver.Lookup{}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(true),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"no commands": {
			svc: &Service{
				Command: command.Slice{}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"3 commands": {
			svc: &Service{
				Command: command.Slice{
					{"ls", "-la"},
					{"true"},
					{"false"}}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (3),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"0 webhooks": {
			svc: &Service{
				WebHook: webhook.Slice{}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status:                   &apitype.Status{}},
		},
		"3 webhooks": {
			svc: &Service{
				WebHook: webhook.Slice{
					"bish": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil),
					"bash": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil),
					"bosh": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"gitlab",
						"", nil, nil, nil)}},
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (3),
				Status:                   &apitype.Status{}},
		},
		"only status": {
			svc: &Service{
				Status: status.Status{}},
			approvedVersion:          "1",
			deployedVersion:          "2",
			deployedVersionTimestamp: "2-",
			latestVersion:            "3",
			latestVersionTimestamp:   "3-",
			lastQueried:              "4",
			want: &apitype.ServiceSummary{
				Type:                     "",
				Icon:                     (""),
				IconLinkTo:               (""),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  (0),
				WebHook:                  (0),
				Status: &apitype.Status{
					ApprovedVersion:          "1",
					DeployedVersion:          "2",
					DeployedVersionTimestamp: "2-",
					LatestVersion:            "3",
					LatestVersionTimestamp:   "3-",
					LastQueried:              "4"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// status
			if tc.svc != nil {
				tc.svc.Status.Init(
					len(tc.svc.Notify), len(tc.svc.Command), len(tc.svc.WebHook),
					&tc.svc.ID,
					&tc.svc.Dashboard.WebURL)
				if tc.approvedVersion != "" {
					tc.svc.Status.SetApprovedVersion(tc.approvedVersion, false)
					tc.svc.Status.SetDeployedVersion(tc.deployedVersion, tc.deployedVersionTimestamp, false)
					tc.svc.Status.SetLatestVersion(tc.latestVersion, tc.latestVersionTimestamp, false)
					tc.svc.Status.SetLastQueried(tc.lastQueried)
				}
			}

			// WHEN the Service is converted to a ServiceSummary
			got := tc.svc.Summary()

			// THEN the result is as expected
			if got.String() != tc.want.String() {
				t.Errorf("got:\n%q\nwant:\n%q",
					got.String(), tc.want.String())
			}
		})
	}
}

func TestService_UsingDefaults(t *testing.T) {
	// GIVEN a Service that may/may not be using defaults
	tests := map[string]struct {
		nilService                                               bool
		usingNotifyDefaults, usingCommandDefaults, usingDefaults bool
	}{
		"nil Service": {
			nilService:           true,
			usingNotifyDefaults:  false,
			usingCommandDefaults: false,
			usingDefaults:        false,
		},
		"using all defaults": {
			usingNotifyDefaults:  true,
			usingCommandDefaults: true,
			usingDefaults:        true,
		},
		"using no defaults": {
			usingNotifyDefaults:  false,
			usingCommandDefaults: false,
			usingDefaults:        false,
		},
		"using Notify defaults": {
			usingNotifyDefaults:  true,
			usingCommandDefaults: false,
			usingDefaults:        false,
		},
		"using Command defaults": {
			usingNotifyDefaults:  false,
			usingCommandDefaults: true,
			usingDefaults:        false,
		},
		"using WebHook defaults": {
			usingNotifyDefaults:  false,
			usingCommandDefaults: false,
			usingDefaults:        true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var svc *Service
			if !tc.nilService {
				svc = &Service{}
				svc.notifyFromDefaults = tc.usingNotifyDefaults
				svc.commandFromDefaults = tc.usingCommandDefaults
				svc.webhookFromDefaults = tc.usingDefaults
			}

			// WHEN UsingDefaults is called
			usingNotifyDefaults, usingCommandDefaults, usingDefaults := svc.UsingDefaults()

			// THEN the Service is using defaults as expected
			if tc.usingNotifyDefaults != usingNotifyDefaults {
				t.Errorf("got: %v, want: %v",
					usingNotifyDefaults, tc.usingNotifyDefaults)
			}
			if tc.usingCommandDefaults != usingCommandDefaults {
				t.Errorf("got: %v, want: %v",
					usingCommandDefaults, tc.usingCommandDefaults)
			}
			if tc.usingDefaults != usingDefaults {
				t.Errorf("got: %v, want: %v",
					usingDefaults, tc.usingDefaults)
			}
		})
	}
}

func TestService_UnmarshalJSON(t *testing.T) {
	// GIVEN a JSON string that represents a Service
	tests := map[string]struct {
		svc      *Service
		jsonData string
		errRegex string
		want     *Service
	}{
		"valid type - github": {
			jsonData: `{
				"latest_version": {
					"type": "github",
					"url": "release-argus/Argus"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus"}},
			},
		},
		"valid type - github (full)": {
			jsonData: `{
				"latest_version": {
					"type": "github",
					"url": "release-argus/Argus",
					"require": {
						"docker": {
							"image": "releaseargus/argus"}},
					"access_token": "foo",
					"url_commands": [
						{"type": "regex", "regex": ".*"}],
					"use_prerelease": true
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus",
						URLCommands: filter.URLCommandSlice{
							filter.URLCommand{Type: "regex", Regex: `.*`}},
						Require: &filter.Require{
							Docker: &filter.DockerCheck{
								Image: "releaseargus/argus"}}},
					AccessToken:   "foo",
					UsePreRelease: test.BoolPtr(true)},
			},
		},
		"github - invalid json": {
			jsonData: `{
				"latest_version": {
					"type": "github",
					"url": ["release-argus/Argus"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				failed to unmarshal github.Lookup:
				json: cannot unmarshal array into Go struct field .url of type string`),
		},
		"valid type - url": {
			jsonData: `{
				"latest_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com",
					},
				},
			},
		},
		"valid type - url (full)": {
			jsonData: `{
				"latest_version": {
					"type": "url",
					"url": "https://example.com",
					"require": {
						"docker": {
							"image": "releaseargus/argus"}},
					"allow_invalid_certs": true,
					"url_commands": [
						{"type": "regex", "regex": ".*"}
					]
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com",
						URLCommands: filter.URLCommandSlice{
							filter.URLCommand{Type: "regex", Regex: `.*`}},
						Require: &filter.Require{
							Docker: &filter.DockerCheck{
								Image: "releaseargus/argus"}}},
					AllowInvalidCerts: test.BoolPtr(true)},
			},
		},
		"url - invalid json": {
			jsonData: `{
				"latest_version": {
					"type": "url",
					"url": ["https://example.com"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				failed to unmarshal web.Lookup:
				json: cannot unmarshal array into Go struct field .url of type string`),
		},
		"valid type - web (url alias)": {
			jsonData: `{
				"latest_version": {
					"type": "web",
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}},
			},
		},
		"unknown type": {
			jsonData: `{
				"latest_version": {
					"type": "unsupported"
				}
			}`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				failed to unmarshal latestver.Lookup:
				type: "unsupported" <invalid> \(expected one of \[github, url\]\)$`),
			want: &Service{},
		},
		"missing type": {
			jsonData: `{
			"latest_version": {
				"url": "https://example.com"
				}
			}`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				type: <required> \[github, url\]$`),
			want: &Service{},
		},
		"invalid type format": {
			jsonData: `{
				"latest_version": {
					"type": ["unsupported"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				type: <invalid> \(cannot unmarshal array.*$`),
			want: &Service{},
		},
		"invalid json": {
			jsonData: `{invalid: json}`,
			errRegex: test.TrimYAML(`
				failed to unmarshal Service:
				invalid character.*$`),
			want: &Service{},
		},
		"nil latest_version": {
			jsonData: `{
				"latest_version": null
			}`,
			errRegex: "",
			want:     &Service{},
		},
		"type from LatestVersion - GitHub": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return github.New(
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			jsonData: `{
				"latest_version": {
					"url": "release-argus/Argus"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus"}},
			},
		},
		"type from LatestVersion - URL": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return web.New(
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			jsonData: `{
				"latest_version": {
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}}},
		},
		"no latest_version": {
			jsonData: `{
				"deployed_version": {
					"method": "GET",
					"url": "https://valid.release-argus.io/plain"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					Method: "GET",
					URL:    "https://valid.release-argus.io/plain",
				}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty Service
			if tc.svc == nil {
				tc.svc = &Service{}
			}

			// WHEN the JSON is unmarshalled into a Service
			err := tc.svc.UnmarshalJSON([]byte(test.TrimJSON(tc.jsonData)))

			// THEN the error is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("unmarshalJSON() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}
			// AND the result is as expected
			if tc.want != nil && tc.svc.String("") != tc.want.String("") {
				t.Errorf("unmarshalJSON() result mismatch\n%q\ngot:\n%q",
					tc.want.String(""), tc.svc.String(""))
			}
		})
	}
}

func TestService_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		svc      *Service
		want     string
		errRegex string
	}{
		"empty service": {
			svc: &Service{},
			want: test.TrimJSON(`{
				"options":{},
				"dashboard":{}
			}`),
			errRegex: `^$`,
		},
		"service with comment": {
			svc: &Service{
				Comment: "test comment",
			},
			want: test.TrimJSON(`{
				"comment":"test comment",
				"options":{},
				"dashboard":{}
			}`),
			errRegex: `^$`,
		},
		"service with options": {
			svc: &Service{
				Options: opt.Options{
					Active: test.BoolPtr(true),
				},
			},
			want: test.TrimJSON(`{
				"options":{"active":true},
				"dashboard":{}
			}`),
			errRegex: `^$`,
		},
		"service with latest version (GitHub)": {
			svc: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus",
					},
				},
			},
			want: test.TrimJSON(`{
				"options":{},
				"latest_version":{
					"type":"github",
					"url":"release-argus/Argus"
				},
				"dashboard":{}
			}`),
			errRegex: `^$`,
		},
		"service with latest version (URL)": {
			svc: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com",
					},
				},
			},
			want: test.TrimJSON(`{
				"options":{},
				"latest_version":{
					"type":"url",
					"url":"https://example.com"
				},
				"dashboard":{}
			}`),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Service is marshalled to JSON
			gotBytes, err := tc.svc.MarshalJSON()

			// THEN the error is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("MarshalJSON() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}

			// AND the result is as expected
			gotString := string(gotBytes)
			if gotString != tc.want {
				t.Errorf("MarshalJSON() result mismatch\nwant: %q\ngot:  %q",
					tc.want, gotString)
			}
		})
	}
}

func TestService_UnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		svc      *Service
		yamlData string
		errRegex string
		want     *Service
	}{
		"valid type - github": {
			yamlData: `
				latest_version:
					type: github
					url: release-argus/Argus
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus"}},
			},
		},
		"valid type - github (full)": {
			yamlData: `
				latest_version:
					type: github
					url: release-argus/Argus
					require:
						docker:
							image: releaseargus/argus
					access_token: foo
					url_commands:
					- type: regex
						regex: .*
					use_prerelease: true
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus",
						URLCommands: filter.URLCommandSlice{
							filter.URLCommand{Type: "regex", Regex: `.*`}},
						Require: &filter.Require{
							Docker: &filter.DockerCheck{
								Image: "releaseargus/argus"}}},
					AccessToken:   "foo",
					UsePreRelease: test.BoolPtr(true)},
			},
		},
		"github - invalid json": {
			yamlData: `
				latest_version:
					type: github
					url: ["https://example.com"]
			`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				failed to unmarshal github.Lookup:
				yaml: unmarshal errors:
				.*cannot unmarshal.*$`),
		},
		"valid type - url": {
			yamlData: `
				latest_version:
					type: url
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}},
			},
		},
		"valid type - url (full)": {
			yamlData: `
				latest_version:
					type: url
					url: https://example.com
					require:
						docker:
							image: releaseargus/argus
					allow_invalid_certs: true
					url_commands:
					- type: regex
						regex: .*
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com",
						URLCommands: filter.URLCommandSlice{
							filter.URLCommand{Type: "regex", Regex: `.*`}},
						Require: &filter.Require{
							Docker: &filter.DockerCheck{
								Image: "releaseargus/argus"}}},
					AllowInvalidCerts: test.BoolPtr(true)},
			},
		},
		"valid type - web (url alias)": {
			yamlData: `
				latest_version:
					type: web
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}},
			},
		},
		"url - invalid json": {
			yamlData: `
				latest_version:
					type: url
					url: ["https://example.com"]
			`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				failed to unmarshal web.Lookup:
				yaml: unmarshal errors:
				.*cannot unmarshal.*$`),
		},
		"unknown type": {
			yamlData: `
				latest_version:
					type: unsupported
			`,
			errRegex: test.TrimYAML(`
			error in latest_version field:
			type: "unsupported" <invalid> \(expected one of \[github, url\]\)$`),
			want: &Service{},
		},
		"missing type": {
			yamlData: `
				latest_version:
					url: https://example.com
			`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				type: <required> \[github, url\]$`),
			want: &Service{},
		},
		"invalid type format": {
			yamlData: `
				latest_version:
					type: ["unsupported"]
			`,
			errRegex: test.TrimYAML(`
				^error in latest_version field:
				type: <invalid> \(".*cannot unmarshal.*"\)$`),
			want: &Service{},
		},
		"invalid yaml": {
			yamlData: `invalid yaml`,
			errRegex: test.TrimYAML(`
			failed to unmarshal Service:
			yaml: unmarshal errors:
			  .*cannot unmarshal.*$`),
			want: &Service{},
		},
		"nil latest_version": {
			yamlData: `
				latest_version: null
			`,
			errRegex: "",
			want:     &Service{},
		},
		"type from LatestVersion - GitHub": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return github.New(
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			yamlData: `
				latest_version:
					url: release-argus/Argus
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus"}},
			},
		},
		"type from LatestVersion - URL": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return web.New(
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			yamlData: `
				latest_version:
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}}},
		},
		"no latest_version": {
			yamlData: `
				deployed_version:
					method: GET
					url: https://valid.release-argus.io/plain
			`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					Method: "GET",
					URL:    "https://valid.release-argus.io/plain",
				}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty Service
			if tc.svc == nil {
				tc.svc = &Service{}
			}

			// WHEN the YAML is unmarshalled into a Service
			err := yaml.Unmarshal([]byte(test.TrimYAML(tc.yamlData)), &tc.svc)

			// THEN the error is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}
			// AND the result is as expected
			if tc.want != nil && tc.svc.String("") != tc.want.String("") {
				t.Errorf("result mismatch\nwant: %s\ngot:  %s",
					tc.want.String(""), tc.svc.String(""))
			}
		})
	}
}

func TestService_MarshalYAML(t *testing.T) {
	tests := map[string]struct {
		svc      *Service
		want     string
		errRegex string
	}{
		"empty service": {
			svc:      &Service{},
			want:     "{}\n",
			errRegex: `^$`,
		},
		"service with comment": {
			svc: &Service{
				Comment: "test comment",
			},
			want: test.TrimYAML(`
				comment: test comment
			`),
			errRegex: `^$`,
		},
		"service with options": {
			svc: &Service{
				Options: opt.Options{
					Active: test.BoolPtr(true),
				},
			},
			want: test.TrimYAML(`
				options:
					  active: true
			`),
			errRegex: `^$`,
		},
		"service with latest version (GitHub)": {
			svc: &Service{
				LatestVersion: &github.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "github",
						URL:  "release-argus/Argus",
					},
				},
			},
			want: test.TrimYAML(`
				latest_version:
					  type: github
					  url: release-argus/Argus
			`),
			errRegex: `^$`,
		},
		"service with latest version (URL)": {
			svc: &Service{
				LatestVersion: &web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com",
					},
				},
			},
			want: test.TrimYAML(`
				latest_version:
					  type: url
					  url: https://example.com
			`),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Service is marshalled to YAML
			got, err := tc.svc.MarshalYAML()

			// THEN the error is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("MarshalYAML() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}

			// AND the result is as expected
			gotBytes, err := yaml.Marshal(got)
			gotString := string(gotBytes)
			if gotString != tc.want {
				t.Errorf("MarshalYAML() result mismatch\nwant: %q\ngot:  %q",
					tc.want, gotString)
			}
		})
	}
}
