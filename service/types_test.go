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

package service

import (
	"net/http"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	deployedver_base "github.com/release-argus/Argus/service/deployed_version/types/base"
	dv_web "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	lv_web "github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestSlice_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected Slice
		errRegex string
	}{
		"empty JSON": {
			input:    "{}",
			expected: Slice{},
			errRegex: `^$`,
		},
		"single service": {
			input: test.TrimJSON(`{
				"service1": {
					"latest_version": {
						"type": "github",
						"url": "owner/repo"
					}
				}
			}`),
			expected: Slice{
				"service1": &Service{
					ID: "service1",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo"},
					},
				},
			},
			errRegex: `^$`,
		},
		"multiple services": {
			input: test.TrimJSON(`{
				"service1": {
					"latest_version": {
						"type": "github",
						"url": "owner/repo1"
					}
				},
				"service2": {
					"latest_version": {
						"type": "github",
						"url": "owner/repo2"
					}
				}
			}`),
			expected: Slice{
				"service1": &Service{
					ID: "service1",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo1"},
					},
				},
				"service2": &Service{
					ID: "service2",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo2"},
					},
				},
			},
			errRegex: `^$`,
		},
		"nil service is removed": {
			input: test.TrimJSON(`{
				"service1": null,
				"service2": {
					"latest_version": {
						"type": "github",
						"url": "owner/repo"
					}
				}
			}`),
			expected: Slice{
				"service2": &Service{
					ID: "service2",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo"},
					},
				},
			},
			errRegex: `^$`,
		},
		"invalid JSON var": {
			input: `{"invalid": "json"}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service\.Slice:
					failed to unmarshal service\.Service:
						cannot unmarshal string .*$`),
		},
		"invalid JSON format": {
			input: `{"invalid": json`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service\.Slice:
					invalid character.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN the YAML is unmarshalled into a Slice.
			var got Slice
			err := got.UnmarshalJSON([]byte(tc.input))

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("UnmarshalYAML() error mismatch\n%v\ngot:\n%v",
					tc.errRegex, err)
				return
			}
			// AND the length is as expected.
			if len(got) != len(tc.expected) {
				t.Errorf("got length %v, expected %v", len(got), len(tc.expected))
			}
			// AND the services are as expected.
			for id, expectedService := range tc.expected {
				gotService, exists := got[id]
				if !exists {
					t.Errorf("service %q not found in result", id)
					continue
				}
				if gotService.ID != expectedService.ID {
					t.Errorf("service %q: got ID %q, expected %q",
						id, gotService.ID, expectedService.ID)
				}
			}
		})
	}
}

func TestSlice_UnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected Slice
		errRegex string
	}{
		"empty YAML": {
			input:    "{}",
			expected: Slice{},
			errRegex: `^$`,
		},
		"single service": {
			input: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo`),
			expected: Slice{
				"service1": &Service{
					ID: "service1",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo"},
					},
				},
			},
			errRegex: `^$`,
		},
		"multiple services": {
			input: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo1
				service2:
					latest_version:
						type: github
						url: owner/repo2`),
			expected: Slice{
				"service1": &Service{
					ID: "service1",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo1"},
					},
				},
				"service2": &Service{
					ID: "service2",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo2"},
					},
				},
			},
			errRegex: `^$`,
		},
		"nil service is removed": {
			input: test.TrimYAML(`
				service1: null
				service2:
					latest_version:
						type: github
						url: owner/repo`),
			expected: Slice{
				"service2": &Service{
					ID: "service2",
					LatestVersion: &github.Lookup{
						Lookup: latestver_base.Lookup{
							URL: "owner/repo"},
					},
				},
			},
			errRegex: `^$`,
		},
		"invalid YAML": {
			input:    "invalid: [yaml: syntax",
			errRegex: `yaml: line 1: did not find expected`,
		},
		"invalid Slice YAML": {
			input: test.TrimYAML(`
				service1:
					latest_version:
						type: something`),
			errRegex: test.TrimYAML(`
				^failed to unmarshal service\.Slice:
					failed to unmarshal latestver\.Lookup:
						type: "something" <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tc.input), &node); err != nil {
				if util.RegexCheck(tc.errRegex, err.Error()) {
					return
				}
				t.Fatalf("failed to parse YAML: %v", err)
			}

			// WHEN the YAML is unmarshalled into a Slice.
			var got Slice
			err := got.UnmarshalYAML(&node)

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("UnmarshalYAML() error mismatch\n%v\ngot:\n%v",
					tc.errRegex, err)
				return
			}
			// AND the length is as expected.
			if len(got) != len(tc.expected) {
				t.Errorf("got length %v, expected %v", len(got), len(tc.expected))
			}
			// AND the services are as expected.
			for id, expectedService := range tc.expected {
				gotService, exists := got[id]
				if !exists {
					t.Errorf("service %q not found in result", id)
					continue
				}
				if gotService.ID != expectedService.ID {
					t.Errorf("service %q: got ID %q, expected %q",
						id, gotService.ID, expectedService.ID)
				}
			}
		})
	}
}

func TestSlice_giveIDs(t *testing.T) {
	// GIVEN a Slice.
	tests := map[string]struct {
		slice    Slice
		expected Slice
	}{
		"nil slice": {
			slice:    nil,
			expected: nil,
		},
		"empty slice": {
			slice:    Slice{},
			expected: Slice{},
		},
		"slice with nil service": {
			slice: Slice{
				"s1": nil,
				"s2": &Service{},
			},
			expected: Slice{
				"s2": &Service{ID: "s2", Name: "s2"},
			},
		},
		"multiple services": {
			slice: Slice{
				"s1": &Service{},
				"s2": &Service{},
				"s3": &Service{},
			},
			expected: Slice{
				"s1": &Service{ID: "s1", Name: "s1"},
				"s2": &Service{ID: "s2", Name: "s2"},
				"s3": &Service{ID: "s3", Name: "s3"},
			},
		},
		"service with existing ID": {
			slice: Slice{
				"service1": &Service{ID: "oldID"},
			},
			expected: Slice{
				"service1": &Service{ID: "service1", Name: "service1"},
			},
		},
		"services with Name different to ID": {
			slice: Slice{
				"s1": &Service{ID: "s1"},
				"s2": &Service{ID: "s2", Name: "Name 2"},
			},
			expected: Slice{
				"s1": &Service{ID: "s1", Name: "s1"},
				"s2": &Service{ID: "s2", Name: "Name 2"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN giveIDs is called.
			tc.slice.giveIDs()

			// THEN the length is as expected.
			if len(tc.slice) != len(tc.expected) {
				t.Errorf("got length %v, expected %v",
					len(tc.slice), len(tc.expected))
			}

			// AND each Service is given its key as ID.
			for id, service := range tc.expected {
				got, exists := tc.slice[id]
				if !exists {
					t.Errorf("service %q not found in result", id)
					continue
				}
				if got.ID != service.ID {
					t.Errorf("service %q: got ID %q, expected %q",
						id, got.ID, service.ID)
				}
				if got.Name != service.Name {
					t.Errorf("service %q: got Name %q, expected %q",
						id, got.Name, service.Name)
				}
			}
		})
	}
}

func TestService_MarshalName(t *testing.T) {
	tests := []struct {
		name              string
		marshalName, want bool
	}{
		{
			name:        "true",
			marshalName: true,
			want:        true,
		},
		{
			name:        "false",
			marshalName: false,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				marshalName: tt.marshalName,
			}
			if got := s.MarshalName(); got != tt.want {
				t.Errorf("Service.MarshalName() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							method: GET
							url: `+test.LookupPlain["url_valid"]+`
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
					test.BoolPtr(true), "", "", "", nil,
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
					type: url
					method: GET
					url: ` + test.LookupPlain["url_valid"] + `
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

				// WHEN the Service is stringified with String.
				got := tc.svc.String(prefix)

				// THEN the result is as expected.
				if got != want {
					t.Errorf("Service.String() mismatch (prefix=%q)\nwant: %q\ngot:  %q",
						prefix, want, got)
					return // No need to check other prefixes.
				}
			}
		})
	}
}

func TestService_Summary(t *testing.T) {
	// GIVEN a Service.
	tests := map[string]struct {
		svc  *Service
		want *apitype.ServiceSummary
	}{
		"nil": {
			svc:  nil,
			want: nil,
		},
		"empty": {
			svc: &Service{},
			want: &apitype.ServiceSummary{
				ID:                       "",
				Name:                     nil,
				Type:                     "",
				Icon:                     nil,
				IconLinkTo:               nil,
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  nil,
				WebHook:                  nil,
				Status:                   &apitype.Status{}},
		},
		"only id": {
			svc: &Service{
				ID: "foo"},
			want: &apitype.ServiceSummary{
				ID:                       "foo",
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"only name": {
			svc: &Service{
				Name:        "bar",
				marshalName: true},
			want: &apitype.ServiceSummary{
				Name:                     test.StringPtr("bar"),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"only options.active": {
			svc: &Service{
				Options: opt.Options{
					Active: test.BoolPtr(false)}},
			want: &apitype.ServiceSummary{
				Active:                   test.BoolPtr(false),
				HasDeployedVersionLookup: test.BoolPtr(false),
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
				Type:                     "github",
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, and it's a url": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "https://example.com/icon.png"}},
			want: &apitype.ServiceSummary{
				Icon:                     test.StringPtr("https://example.com/icon.png"),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon, and it's not a url": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Icon: "smile"}},
			want: &apitype.ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(false),
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
				Icon:                     test.StringPtr("https://example.com/notify.png"),
				HasDeployedVersionLookup: test.BoolPtr(false),
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
				Icon:                     test.StringPtr("https://example.com/icon.png"),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.icon_link_to": {
			svc: &Service{
				Dashboard: DashboardOptions{
					IconLinkTo: "https://example.com"}},
			want: &apitype.ServiceSummary{
				IconLinkTo:               test.StringPtr("https://example.com"),
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"only dashboard.tags": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Tags: []string{"hello", "there"}}},
			want: &apitype.ServiceSummary{
				Tags:                     &[]string{"hello", "there"},
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"only deployed_version": {
			svc: &Service{
				DeployedVersionLookup: &dv_web.Lookup{}},
			want: &apitype.ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(true),
				Status:                   &apitype.Status{}},
		},
		"no commands": {
			svc: &Service{
				Command: command.Slice{}},
			want: &apitype.ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(false),
				Status:                   &apitype.Status{}},
		},
		"3 commands": {
			svc: &Service{
				Command: command.Slice{
					{"ls", "-la"},
					{"true"},
					{"false"}}},
			want: &apitype.ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(false),
				Command:                  test.IntPtr(3),
				Status:                   &apitype.Status{}},
		},
		"0 webhooks": {
			svc: &Service{
				WebHook: webhook.Slice{}},
			want: &apitype.ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(false),
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
				HasDeployedVersionLookup: test.BoolPtr(false),
				WebHook:                  test.IntPtr(3),
				Status:                   &apitype.Status{}},
		},
		"only status": {
			svc: &Service{
				Status: *status.New(
					nil, nil, nil,
					"1",
					"2", "2-",
					"3", "3-",
					"4")},
			want: &apitype.ServiceSummary{
				HasDeployedVersionLookup: test.BoolPtr(false),
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

			// Status.
			if tc.svc != nil {
				tc.svc.Status.Init(
					len(tc.svc.Notify), len(tc.svc.Command), len(tc.svc.WebHook),
					&tc.svc.ID, &name,
					&tc.svc.Dashboard.WebURL)
			}

			// WHEN the Service is converted to a ServiceSummary.
			got := tc.svc.Summary()

			// THEN the result is as expected.
			if got.String() != tc.want.String() {
				t.Errorf("got:\n%q\nwant:\n%q",
					got.String(), tc.want.String())
			}
		})
	}
}

func TestService_UsingDefaults(t *testing.T) {
	// GIVEN a Service that may/may not be using defaults.
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

			// WHEN UsingDefaults is called.
			usingNotifyDefaults, usingCommandDefaults, usingDefaults := svc.UsingDefaults()

			// THEN the Service is using defaults as expected.
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
	// GIVEN a JSON string that represents a Service.
	tests := map[string]struct {
		svc      *Service
		jsonData string
		errRegex string
		want     *Service
	}{
		"invalid JSON": {
			jsonData: `{invalid: json}`,
			errRegex: test.TrimYAML(`
				failed to unmarshal service\.Service:
					invalid character.*$`),
			want: &Service{},
		},
		"latest_version: valid type - github": {
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
		"latest_version: valid type - github (full)": {
			jsonData: `{
				"name": "foo",
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
				Name:        "foo",
				marshalName: true,
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
		"latest_version: github - invalid JSON": {
			jsonData: `{
				"latest_version": {
					"type": "github",
					"url": ["release-argus/Argus"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal github.Lookup:
					cannot unmarshal array into Go struct field \.Lookup\.url of type string`),
		},
		"latest_version: valid type - url": {
			jsonData: `{
				"latest_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com",
					},
				},
			},
		},
		"latest_version: valid type - url (full)": {
			jsonData: `{
				"name": "bar",
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
				Name:        "bar",
				marshalName: true,
				LatestVersion: &lv_web.Lookup{
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
		"latest_version: url - invalid JSON": {
			jsonData: `{
				"latest_version": {
					"type": "url",
					"url": ["https://example.com"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					cannot unmarshal array into Go struct field \.Lookup\.url of type string`),
		},
		"latest_version: valid type - web (url alias)": {
			jsonData: `{
				"latest_version": {
					"type": "web",
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}},
			},
		},
		"latest_version: unknown type": {
			jsonData: `{
				"latest_version": {
					"type": "unsupported"
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					type: "unsupported" <invalid> .*\[github, url\].*$`),
			want: &Service{},
		},
		"latest_version: missing type": {
			jsonData: `{
			"latest_version": {
				"url": "https://example.com"
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					type: <required> .*\[github, url\].*$`),
			want: &Service{},
		},
		"latest_version: invalid type format": {
			jsonData: `{
				"latest_version": {
					"type": ["unsupported"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service.Service.LatestVersion:
					cannot unmarshal array.* type string$`),
			want: &Service{},
		},
		"latest_version: nil": {
			jsonData: `{
				"latest_version": null
			}`,
			errRegex: "",
			want:     &Service{},
		},
		"latest_version: type from existing - GitHub": {
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
		"latest_version: type from existing - URL": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return lv_web.New(
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
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}}},
		},
		"no latest_version": {
			jsonData: `{
				"deployed_version": {
					"type": "url",
					"method": "GET",
					"url": "` + test.LookupPlain["url_valid"] + `"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					Method: http.MethodGet,
					URL:    test.LookupPlain["url_valid"],
				}},
		},
		"deployed_version: valid type - url": {
			jsonData: `{
				"deployed_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					URL: "https://example.com"},
			},
		},
		"deployed_version: valid type - url (full)": {
			jsonData: `{
				"name": "foo",
				"deployed_version": {
					"type": "url",
					"method": "GET",
					"url": "https://example.com",
					"allow_invalid_certs": true,
					"basic_auth": {
						"username": "foo",
						"password": "bar"
					},
					"headers": [
						{ "key": "foo", "value": "bar" },
						{ "key": "something", "value": "else" }
					],
					"body": "removed_on_verify",
					"regex": "(\\d+)\\.(\\d+)\\.(\\d+)",
					"regex_template": "$3.$2.$1"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				Name:        "foo",
				marshalName: true,
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					Method:            http.MethodGet,
					URL:               "https://example.com",
					AllowInvalidCerts: test.BoolPtr(true),
					BasicAuth: &dv_web.BasicAuth{
						Username: "foo",
						Password: "bar"},
					Headers: []dv_web.Header{
						{Key: "foo", Value: "bar"},
						{Key: "something", Value: "else"}},
					Body:          "removed_on_verify",
					Regex:         `(\d+)\.(\d+)\.(\d+)`,
					RegexTemplate: "$3.$2.$1"},
			},
		},
		"deployed_version: valid type - web (url alias)": {
			jsonData: `{
				"deployed_version": {
					"type": "web",
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					URL: "https://example.com"},
			},
		},
		"deployed_version: url - invalid JSON": {
			jsonData: `{
				"deployed_version": {
					"type": "url",
					"url": ["https://example.com"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					cannot unmarshal array.* type string$`),
		},
		"deployed_version: unknown type": {
			jsonData: `{
				"deployed_version": {
					"type": "unsupported"
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal deployedver.Lookup:
					type: "unsupported" <invalid> .*\[url, manual\].*$`),
			want: &Service{},
		},
		"deployed_version: missing type": {
			jsonData: `{
				"deployed_version": {
					"url": "https://example.com"
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal deployedver.Lookup:
					type: <required> .*\[url, manual\].*$`),
			want: &Service{},
		},
		"deployed_version: invalid type format": {
			jsonData: `{
				"deployed_version": {
					"type": ["unsupported"]
				}
			}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service.Service.DeployedVersion:
					cannot unmarshal.*$`),
			want: &Service{},
		},
		"deployed_version: null": {
			jsonData: `{
				"deployed_version": null
			}`,
			errRegex: "",
			want:     &Service{},
		},
		"deployed_version: type from existing - url": {
			svc: &Service{
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return dv_web.New(
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			jsonData: `{
				"deployed_version": {
					"url": "https://example.com"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					URL: "https://example.com"},
			},
		},
		"no deployed_version": {
			jsonData: `{
				"latest_version": {
					"type": "url",
					"url": "` + test.LookupPlain["url_valid"] + `"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  test.LookupPlain["url_valid"],
					}},
			},
		},
		"dashboard.tags - []string": {
			jsonData: `{
				"dashboard": {
					"tags": [
						"foo",
						"bar"
					]
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				Dashboard: *NewDashboardOptions(
					nil, "", "", "",
					[]string{"foo", "bar"},
					nil, nil),
			},
		},
		"dashboard.tags - string": {
			jsonData: `{
				"dashboard": {
					"tags": "foo"
				}
			}`,
			errRegex: `^$`,
			want: &Service{
				Dashboard: *NewDashboardOptions(
					nil, "", "", "",
					[]string{"foo"},
					nil, nil),
			},
		},
		"dashboard.tags - invalid": {
			jsonData: `{
				"dashboard": {
					"tags": {
						"foo": "bar"
					}
				}
			}`,
			errRegex: test.TrimYAML(`
				failed to unmarshal service\.Service:
					failed to unmarshal service\.DashboardOptions:
						tags: <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty Service.
			if tc.svc == nil {
				tc.svc = &Service{}
			}

			// WHEN the JSON is unmarshalled into a Service.
			err := tc.svc.UnmarshalJSON([]byte(test.TrimJSON(tc.jsonData)))

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Service.UnmarshalJSON() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}
			// AND the result is as expected.
			gotString := tc.svc.String("")
			wantString := tc.want.String("")
			if tc.want != nil && gotString != wantString {
				t.Errorf("Service.UnmarshalJSON() result mismatch\n%q\ngot:\n%q",
					wantString, gotString)
			}
			// AND marshalName is only set if Name is non-empty.
			if tc.svc.MarshalName() != (tc.svc.Name != "") {
				t.Errorf("Service.UnmarshalJSON() marshalName mismatch\nwant: %t\ngot:  %t",
					tc.svc.MarshalName(), (tc.svc.Name != ""))
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
				LatestVersion: &lv_web.Lookup{
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
		"name that marshals": {
			svc: &Service{
				Name:        "foo",
				marshalName: true,
			},
			want: test.TrimJSON(`{
				"name":"foo",
				"options":{},
				"dashboard":{}
			}`),
			errRegex: `^$`,
		},
		"name that doesn't marshal": {
			svc: &Service{
				Name: "bar",
			},
			want: test.TrimJSON(`{
				"options":{},
				"dashboard":{}
			}`),
			errRegex: `^$`,
		},
		"service with tag": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Tags: []string{"foo"}},
			},
			want: test.TrimJSON(`{
				"options":{},
				"dashboard":{
					"tags":["foo"]
				}
			}`),
			errRegex: `^$`,
		},
		"service with tags": {
			svc: &Service{
				Dashboard: DashboardOptions{
					Tags: []string{"foo", "bar"}},
			},
			want: test.TrimJSON(`{
				"options":{},
				"dashboard":{
					"tags":["foo","bar"]
				}
			}`),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Service is marshalled to JSON.
			gotBytes, err := tc.svc.MarshalJSON()

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("MarshalJSON() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}

			// AND the result is as expected.
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
		"invalid YAML": {
			yamlData: `invalid yaml`,
			errRegex: test.TrimYAML(`
				failed to unmarshal service\.Service:
					line \d: cannot unmarshal.*$`),
			want: &Service{},
		},
		"latest_version: valid type - github": {
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
		"latest_version: valid type - github (full)": {
			yamlData: `
				name: foo
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
				Name:        "foo",
				marshalName: true,
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
		"latest_version: github - invalid YAML": {
			yamlData: `
				latest_version:
					type: github
					url: ["https://example.com"]
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal github.Lookup:
					line \d: cannot unmarshal.*$`),
		},
		"latest_version: valid type - url": {
			yamlData: `
				latest_version:
					type: url
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}},
			},
		},
		"latest_version: valid type - url (full)": {
			yamlData: `
				name: foo
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
				Name:        "foo",
				marshalName: true,
				LatestVersion: &lv_web.Lookup{
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
		"latest_version: valid type - web (url alias)": {
			yamlData: `
				latest_version:
					type: web
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}},
			},
		},
		"latest_version: url - invalid YAML": {
			yamlData: `
				latest_version:
					type: url
					url: ["https://example.com"]
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					line \d: cannot unmarshal.*$`),
		},
		"latest_version: unknown type": {
			yamlData: `
				latest_version:
					type: unsupported
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					type: "unsupported" <invalid> .*\[github, url\].*$`),
			want: &Service{},
		},
		"latest_version: missing type": {
			yamlData: `
				latest_version:
					url: https://example.com
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					type: <required> .*\[github, url\].*$`),
			want: &Service{},
		},
		"latest_version: invalid type format": {
			yamlData: `
				latest_version:
					type: ["unsupported"]
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service.Service.LatestVersion:
					line \d: cannot unmarshal.*$`),
			want: &Service{},
		},
		"latest_version: nil": {
			yamlData: `
				latest_version: null
			`,
			errRegex: "",
			want:     &Service{},
		},
		"latest_version: type from existing - github": {
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
		"latest_version: type from existing - url": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return lv_web.New(
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
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  "https://example.com"}},
			},
		},
		"no latest_version": {
			yamlData: `
				deployed_version:
					type: url
					method: GET
					url: ` + test.LookupPlain["url_valid"] + `
			`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					Method: http.MethodGet,
					URL:    test.LookupPlain["url_valid"]},
			},
		},
		"deployed_version: valid type - url": {
			yamlData: `
				deployed_version:
					type: url
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					URL: "https://example.com"},
			},
		},
		"deployed_version: valid type - url (full)": {
			yamlData: `
				name: foo
				deployed_version:
					type: url
					method: GET
					url: https://example.com
					allow_invalid_certs: true
					basic_auth:
						username: foo
						password: bar
					headers:
						- key: foo
							value: bar
						- key: something
							value: else
					body: removed_on_verify
					regex: '(\d+)\.(\d+)\.(\d+)'
					regex_template: $3.$2.$1
			`,
			errRegex: `^$`,
			want: &Service{
				Name:        "foo",
				marshalName: true,
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					Method:            http.MethodGet,
					URL:               "https://example.com",
					AllowInvalidCerts: test.BoolPtr(true),
					BasicAuth: &dv_web.BasicAuth{
						Username: "foo",
						Password: "bar"},
					Headers: []dv_web.Header{
						{Key: "foo", Value: "bar"},
						{Key: "something", Value: "else"}},
					Body:          "removed_on_verify",
					Regex:         `(\d+)\.(\d+)\.(\d+)`,
					RegexTemplate: "$3.$2.$1"},
			},
		},
		"deployed_version: valid type - web (url alias)": {
			yamlData: `
				deployed_version:
					type: web
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					URL: "https://example.com"},
			},
		},
		"deployed_version: url - invalid YAML": {
			yamlData: `
				deployed_version:
					type: url
					url: ["https://example.com"]
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					line \d: cannot unmarshal.*$`),
		},
		"deployed_version: unknown type": {
			yamlData: `
				deployed_version:
					type: unsupported
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal deployedver.Lookup:
					type: "unsupported" <invalid> .*\[url, manual\].*$`),
			want: &Service{},
		},
		"deployed_version: missing type": {
			yamlData: `
				deployed_version:
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					URL: "https://example.com"},
			},
		},
		"deployed_version: invalid type format": {
			yamlData: `
				deployed_version:
					type: ["unsupported"]
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service.Service.DeployedVersion:
					line \d: cannot unmarshal.*$`),
			want: &Service{},
		},
		"deployed_version: nil": {
			yamlData: `
				deployed_version: null
			`,
			errRegex: "",
			want:     &Service{},
		},
		"deployed_version: type from existing - url": {
			svc: &Service{
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return dv_web.New(
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			yamlData: `
				deployed_version:
					url: https://example.com
			`,
			errRegex: `^$`,
			want: &Service{
				DeployedVersionLookup: &dv_web.Lookup{
					Lookup: deployedver_base.Lookup{
						Type: "url"},
					URL: "https://example.com"},
			},
		},
		"no deployed_version": {
			yamlData: `
				latest_version:
					type: url
					url: ` + test.LookupPlain["url_valid"] + `
			`,
			errRegex: `^$`,
			want: &Service{
				LatestVersion: &lv_web.Lookup{
					Lookup: latestver_base.Lookup{
						Type: "url",
						URL:  test.LookupPlain["url_valid"],
					}},
			},
		},
		"tags - []string": {
			yamlData: `
				dashboard:
					tags:
					- foo
					- bar
			`,
			errRegex: `^$`,
			want: &Service{
				Dashboard: DashboardOptions{
					Tags: []string{"foo", "bar"}},
			},
		},
		"tags - string": {
			yamlData: `
				dashboard:
					tags: foo
			`,
			errRegex: `^$`,
			want: &Service{
				Dashboard: DashboardOptions{
					Tags: []string{"foo"}},
			},
		},
		"tags - invalid": {
			yamlData: `
				dashboard:
					tags:
						foo: bar
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service\.Service:
					failed to unmarshal service\.DashboardOptions:
						tags: <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty Service.
			if tc.svc == nil {
				tc.svc = &Service{}
			}

			// WHEN the YAML is unmarshalled into a Service.
			err := yaml.Unmarshal([]byte(test.TrimYAML(tc.yamlData)), &tc.svc)

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Service.UnmarshalYAML() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}
			// AND the result is as expected.
			if tc.want != nil && tc.svc.String("") != tc.want.String("") {
				t.Errorf("Service.UnmarshalYAML() result mismatch\nwant: %s\ngot:  %s",
					tc.want.String(""), tc.svc.String(""))
			}
			// AND marshalName is only set if Name is non-empty.
			if tc.svc.MarshalName() != (tc.svc.Name != "") {
				t.Errorf("Service.UnmarshalYAML() marshalName mismatch\nwant: %t\ngot:  %t",
					tc.svc.MarshalName(), (tc.svc.Name != ""))
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
		"comment": {
			svc: &Service{
				Comment: "test comment",
			},
			want: test.TrimYAML(`
				comment: test comment
			`),
			errRegex: `^$`,
		},
		"options": {
			svc: &Service{
				Options: opt.Options{
					Active: test.BoolPtr(true)},
			},
			want: test.TrimYAML(`
				options:
					  active: true
			`),
			errRegex: `^$`,
		},
		"tags - single": {
			svc: &Service{
				Dashboard: *NewDashboardOptions(
					nil, "", "", "",
					[]string{"foo"},
					nil, nil),
			},
			want: test.TrimYAML(`
				dashboard:
						tags:
								- foo
			`),
			errRegex: `^$`,
		},
		"tags - multiple": {
			svc: &Service{
				Dashboard: *NewDashboardOptions(
					nil, "", "", "",
					[]string{"foo", "bar"},
					nil, nil),
			},
			want: test.TrimYAML(`
				dashboard:
						tags:
								- foo
								- bar
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
				LatestVersion: &lv_web.Lookup{
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
		"name that marshals": {
			svc: &Service{
				Name:        "foo",
				marshalName: true,
			},
			want: test.TrimYAML(`
				name: foo
			`),
			errRegex: `^$`,
		},
		"name that doesn't marshal": {
			svc: &Service{
				Name: "bar",
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Service is marshalled to YAML.
			got, err := tc.svc.MarshalYAML()

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("MarshalYAML() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}

			// AND the result is as expected.
			gotBytes, err := yaml.Marshal(got)
			gotString := string(gotBytes)
			if gotString != tc.want {
				t.Errorf("MarshalYAML() result mismatch\nwant: %q\ngot:  %q",
					tc.want, gotString)
			}
		})
	}
}
