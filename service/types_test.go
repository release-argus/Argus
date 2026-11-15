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

package service

import (
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service/dashboard"
	dashtest "github.com/release-argus/Argus/service/dashboard/test"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestService_Marshal(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	tests := []struct {
		name               string
		svc                *Service
		wantJSON, wantYAML string
		errRegex           string
	}{
		{
			name:     "empty service",
			svc:      &Service{},
			wantJSON: "{}",
			wantYAML: "{}\n",
			errRegex: `^$`,
		},
		{
			name: "service with comment",
			svc: &Service{
				Comment: "test comment",
			},
			wantJSON: `{"comment":"test comment"}`,
			wantYAML: "comment: test comment\n",
			errRegex: `^$`,
		},
		{
			name: "service with options",
			svc: &Service{
				Options: opt.Options{
					Active: new(true),
				},
			},
			wantJSON: test.TrimJSON(`{
				"options": {
					"active": true
				}
			}`),
			wantYAML: test.TrimYAML(`
				options:
					active: true
			`),
			errRegex: `^$`,
		},
		{
			name: "service with latest version (GitHub)",
			svc: &Service{
				LatestVersion: &github.Lookup{
					Type: "github",
					URL:  test.ArgusGitHubRepo,
				},
			},
			wantJSON: test.TrimJSON(`{
				"latest_version":{
					"type":"github",
					"url":"` + test.ArgusGitHubRepo + `"
				}
			}`),
			wantYAML: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name: "service with latest version (URL)",
			svc: &Service{
				LatestVersion: &lvweb.Lookup{
					Type: "url",
					URL:  "https://example.com",
				},
			},
			wantJSON: test.TrimJSON(`{
				"latest_version":{
					"type":"url",
					"url":"https://example.com"
				}
			}`),
			wantYAML: test.TrimYAML(`
				latest_version:
					type: url
					url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name: "service with deployed version (URL)",
			svc: &Service{
				DeployedVersionLookup: &dvweb.Lookup{
					Type: "url",
					URL:  "https://example.com",
				},
			},
			wantJSON: test.TrimJSON(`{
				"deployed_version":{
					"type":"url",
					"url":"https://example.com"
				}
			}`),
			wantYAML: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name: "name",
			svc: &Service{
				Name: "foo",
			},
			wantJSON: `{"name":"foo"}`,
			wantYAML: "name: foo\n",
			errRegex: `^$`,
		},
		{
			name: "NotifyFromDefaults",
			svc: test.Must(t, func() (*Service, error) {
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							foo:
							bar:
					`)),
					"NotifyFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.NotifyFromDefaults = true
				return svc, nil
			}),
			wantJSON: test.TrimJSON(`{
				"latest_version":{
					"type":"github",
					"url":"` + test.ArgusGitHubRepo + `"
				}
			}`),
			wantYAML: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name: "CommandFromDefaults",
			svc: test.Must(t, func() (*Service, error) {
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						command:
							- ["ls", "-la"]
					`)),
					"CommandFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.CommandFromDefaults = true
				return svc, nil
			}),
			wantJSON: test.TrimJSON(`{
				"latest_version":{
					"type":"github",
					"url":"` + test.ArgusGitHubRepo + `"
				}
			}`),
			wantYAML: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name: "WebhookFromDefaults",
			svc: test.Must(t, func() (*Service, error) {
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						webhook:
							foo:
							bar:
					`)),
					"WebhookFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				svc.WebhookFromDefaults = true
				return svc, nil
			}),
			wantJSON: test.TrimJSON(`{
				"latest_version":{
					"type":"github",
					"url":"` + test.ArgusGitHubRepo + `"
				}
			}`),
			wantYAML: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name: "tags/single",
			svc: &Service{
				Dashboard: dashboard.Options{
					Tags: []string{"foo"},
				},
			},
			wantJSON: test.TrimJSON(`{
				"dashboard":{
					"tags":["foo"]
				}
			}`),
			wantYAML: test.TrimYAML(`
				dashboard:
					tags:
						- foo
			`),
			errRegex: `^$`,
		},
		{
			name: "tags/multiple",
			svc: &Service{
				Dashboard: dashboard.Options{
					Tags: []string{"foo", "bar"},
				},
			},
			wantJSON: test.TrimJSON(`{
				"dashboard":{
					"tags":["foo","bar"]
				}
			}`),
			wantYAML: test.TrimYAML(`
				dashboard:
					tags:
						- foo
						- bar
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			results := map[string]string{
				"json": tc.wantJSON,
				"yaml": tc.wantYAML,
			}

			for _, typ := range util.SortedKeys(results) {
				t.Run(typ, func(t *testing.T) {
					t.Parallel()

					want := results[typ]

					bytes, err := decode.Marshal(typ, tc.svc)
					e := errfmt.FormatError(err)
					if !util.RegexCheck(tc.errRegex, e) {
						t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
					}
					if err != nil {
						return
					}

					if got := string(bytes); got != want {
						t.Errorf("decode.Marshal() mismatch\ngot:  %q\nwant: %q", got, want)
					}
				})
			}
		})
	}
}

func TestService_Unmarshal(t *testing.T) {
	svcCfg := plainDefaultsConfig()

	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			errRegex: `^[^\s]+ invalid character .*`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			errRegex: `^[^\s]+ string was used .*`,
		},
		{
			name:   "JSON/latest_version",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/latest_version",
			format: "yaml",
			data: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name:   "latest_version err",
			format: "yaml",
			data: test.TrimYAML(`
				latest_version:
					type: github
					url: [` + test.ArgusGitHubRepo + `]
			`),
			errRegex: test.TrimYAML(`
				latest_version:
					[^\s]+ .*unmarshal.*
						 \d .*
						 \d .*
					[^\s] .*
					\s+\^$`,
			),
		},
		{
			name:   "JSON/deployed_version",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`),
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/deployed_version",
			format: "yaml",
			data: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name:   "deployed_version err",
			format: "yaml",
			data: test.TrimYAML(`
				deployed_version:
					type: url
					url: [https://example.com]
			`),
			errRegex: test.TrimYAML(`
				deployed_version:
					[^\s]+ .*unmarshal.*
						 \d .*
						 \d .*
					[^\s] .*
					\s+\^$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Service, error) {
					v := Service{
						Defaults:     svcCfg.Soft,
						HardDefaults: svcCfg.Hard,
					}
					err := decode.Unmarshal(format, data, &v)
					return &v, err
				},
				tc.format, tc.data,
				func(v *Service) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"Service.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestService_UnmarshalLatestVersion(t *testing.T) {
	svcCfg := plainDefaultsConfig()

	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid",
			format: "json",
			data:   "foo",
			errRegex: test.TrimYAML(`
				^latest_version:
					extract "latest_version":
						jsontext: invalid character .*
							[^\s]+ .*$`,
			),
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/invalid",
			format: "yaml",
			data:   "foo",
			errRegex: test.TrimYAML(`
				^latest_version:
					extract "latest_version":
						[^\s]+ string was used .*`,
			),
		},
		{
			name:   "JSON/latest_version",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/latest_version/full",
			format: "yaml",
			data: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/latest_version/err",
			format: "yaml",
			data: test.TrimYAML(`
				latest_version:
					type: github
					url: [` + test.ArgusGitHubRepo + `]
			`),
			errRegex: test.TrimYAML(`
				latest_version:
					[^\s]+ .*unmarshal.*
						 \d .*
						 \d .*
					[^\s] .*
					\s+\^$`,
			),
		},
		{
			name:   "JSON/deployed_version/ignored",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`),
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/deployed_version/ignored",
			format: "yaml",
			data: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/deployed_version/err, not reached",
			format: "yaml",
			data: test.TrimYAML(`
				deployed_version:
					type: url
					url: [https://example.com]
			`),
			want:     "{}\n",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := Service{
				Defaults:     svcCfg.Soft,
				HardDefaults: svcCfg.Hard,
			}

			err := svc.unmarshalLatestVersion(tc.format, []byte(tc.data))
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got := svc.String(""); got != tc.want {
				t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestService_UnmarshalDeployedVersion(t *testing.T) {
	svcCfg := plainDefaultsConfig()

	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid",
			format: "json",
			data:   "foo",
			errRegex: test.TrimYAML(`
				^deployed_version:
					extract "deployed_version":
						jsontext: invalid character .*
							[^\s]+ .*$`,
			),
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/invalid",
			format: "yaml",
			data:   "foo",
			errRegex: test.TrimYAML(`
				^deployed_version:
					extract "deployed_version":
						[^\s]+ string was used .*`,
			),
		},
		{
			name:   "JSON/latest_version/ignored",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/latest_version/ignored",
			format: "yaml",
			data: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/latest_version/err, not reached",
			format: "yaml",
			data: test.TrimYAML(`
				latest_version:
					type: github
					url: [` + test.ArgusGitHubRepo + `]
			`),
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/deployed_version",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`),
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/deployed_version/full",
			format: "yaml",
			data: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/deployed_version/err",
			format: "yaml",
			data: test.TrimYAML(`
				deployed_version:
					type: url
					url: [https://example.com]
			`),
			errRegex: test.TrimYAML(`
				deployed_version:
					[^\s]+ .*unmarshal.*
						 \d .*
						 \d .*
					[^\s] .*
					\s+\^$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := Service{
				Defaults:     svcCfg.Soft,
				HardDefaults: svcCfg.Hard,
			}

			err := svc.unmarshalDeployedVersion(tc.format, []byte(tc.data))
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got := svc.String(""); got != tc.want {
				t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestService_String(t *testing.T) {
	dashCfg := dashtest.PlainDefaultsConfig()
	lvCfg := lvtest.PlainDefaultsConfig()
	dvCfg := dvtest.PlainDefaultsConfig()

	tests := []struct {
		name string
		svc  *Service
		want string
	}{
		{
			name: "nil",
			svc:  nil,
			want: "",
		},
		{
			name: "empty",
			svc:  &Service{},
			want: "{}\n",
		},
		{
			name: "filled",
			svc: &Service{
				Comment: "svc for blah",
				Options: opt.Options{
					Active: new(false),
				},
				LatestVersion: test.Must(t, func() (latestver.Lookup, error) {
					return latestver.Decode(
						"yaml", []byte(test.TrimYAML(`
							type: github
							url: `+test.ArgusGitHubRepo+`
							url_commands:
								- type: regex
									regex: foo
									index: 1
							require:
								regex_version: v.+
								docker:
									image: test/app
									tag: '{{ version }}'
						`)),
						nil,
						nil,
						lvCfg,
					)
				}),
				DeployedVersionLookup: test.Must(t, func() (deployedver.Lookup, error) {
					return deployedver.Decode(
						"yaml", []byte(test.TrimYAML(`
							type: url
							method: GET
							url: `+test.LookupPlain.URLValid+`
							basic_auth:
								username: user
								password: pass
							headers:
								- key: foo
									value: bar
							json: version
						`)),
						nil,
						nil,
						dvCfg,
					)
				}),
				Notify: shoutrrr.Shoutrrrs{
					"foo": shoutrrr.New(
						nil,
						"", "discord",
						nil,
						map[string]string{
							"token": "bar",
						},
						nil,
						nil,
						nil, nil,
					),
				},
				Command: command.Commands{
					{"ls", "-la"},
				},
				Webhook: webhook.Webhooks{
					"foo": webhook.New(
						nil, nil,
						"",
						nil, nil, "foo",
						nil, webhook.Notifiers{}, nil,
						"",
						nil,
						"github", "https://example.com",
						nil, nil, nil,
					),
				},
				Dashboard: *test.Must(t, func() (*dashboard.Options, error) {
					return dashboard.Decode(
						"yaml", []byte("auto_approve: true"),
						dashCfg,
					)
				}),
				Defaults: &Defaults{
					Options: *test.Must(t, func() (*opt.Defaults, error) {
						return opt.DecodeDefaults("yaml", []byte("semantic_versioning: false"))
					}),
				},
				HardDefaults: &Defaults{
					Options: *test.Must(t, func() (*opt.Defaults, error) {
						return opt.DecodeDefaults("yaml", []byte("semantic_versioning: false"))
					}),
				},
			},
			want: test.TrimYAML(`
				comment: svc for blah
				options:
					active: false
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
					url_commands:
						- type: regex
							regex: foo
							index: 1
					require:
						regex_version: v.+
						docker:
							image: test/app
							tag: '{{ version }}'
				deployed_version:
					type: url
					method: GET
					url: ` + test.LookupPlain.URLValid + `
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
					auto_approve: true
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"Service.String",
				tc.svc.String,
				tc.want,
			)
		})
	}
}

func TestService_Summary(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	tests := []struct {
		name string
		svc  *Service
		want *apitype.ServiceSummary
	}{
		{
			name: "nil",
			svc:  nil,
			want: nil,
		},
		{
			name: "empty",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", nil,
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				ID:                  "",
				Name:                nil,
				Type:                "",
				Icon:                nil,
				IconLinkTo:          nil,
				DeployedVersionType: new(""),
				Command:             nil,
				Webhook:             nil,
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only id",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", nil,
					"foo",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				ID:                  "foo",
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only name",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: bar
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Name:                new("bar"),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only options.active",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							active: false
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Active:              new(false),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only latest_version.type",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Type:                "github",
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only dashboard.icon/is a url",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						dashboard:
							icon: https://example.com/icon.png
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Icon:                new("https://example.com/icon.png"),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only dashboard.icon/is not a url",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						dashboard:
							icon: smile
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Icon:                new("smile"),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only dashboard.icon/from notify",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						notify:
							foo:
								type: discord
								params:
									icon: https://example.com/notify.png
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Icon:                new("https://example.com/notify.png"),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only dashboard.icon/dashboard overrides notify",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						dashboard:
							icon: https://example.com/icon.png
						notify:
							foo:
								type: discord
								params:
									icon: https://example.com/notify.png
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Icon:                new("https://example.com/icon.png"),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only dashboard.icon_link_to",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						dashboard:
							icon_link_to: https://example.com
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				IconLinkTo:          new("https://example.com"),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only dashboard.web_url",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						dashboard:
							web_url: https://example.com
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				WebURL:              new("https://example.com"),
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only dashboard.tags",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						dashboard:
							tags: ["hello", "there"]
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				Tags:                &[]string{"hello", "there"},
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only deployed_version",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version: {}
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				DeployedVersionType: new(dvweb.Type),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "no commands",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						command: []
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "3 commands",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						command:
							- ['ls', '-la']
							- ['true']
							- ['false']
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				DeployedVersionType: new(""),
				Command:             new(3),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "0 webhooks",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						webhook: {}
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				DeployedVersionType: new(""),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "3 webhooks",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						webhook:
							bish:
								type: github
							bash:
								type: github
							bosh:
								type: gitlab
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: &apitype.ServiceSummary{
				DeployedVersionType: new(""),
				Webhook:             new(3),
				Status:              &apitype.Status{},
			},
		},
		{
			name: "only status",
			svc: &Service{
				Status: *status.New(
					nil, nil, nil,
					"1",
					"2", "2-",
					"3", "3-",
					"4",
					&dashboard.Options{},
				),
			},
			want: &apitype.ServiceSummary{
				DeployedVersionType: new(""),
				Status: &apitype.Status{
					ApprovedVersion:          "1",
					DeployedVersion:          "2",
					DeployedVersionTimestamp: "2-",
					LatestVersion:            "3",
					LatestVersionTimestamp:   "3-",
					LastQueried:              "4",
				},
			},
		},
		{
			name: "all",
			svc: test.Must(t, func() (*Service, error) {
				svc, err := DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: bar
						comment: svc for blah

						options:
							active: false

						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						deployed_version:
							type: url
							method: GET
							url: `+test.LookupPlain.URLValid+`
							json: version

						notify:
							foo:
								type: discord

						command:
							- ['true']
							- ['false']

						webhook:
							bish:
								type: github
								url: https://example.com
							bash:
								type: github
								url: https://example.com
							bosh:
								type: gitlab
								url: https://example.com

						dashboard:
							icon: https://example.com/icon.png
							icon_link_to: https://example.com
							web_url: https://example.com
							tags: ["hello", "there"]
					`)),
					"foo",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}

				svc.Status = *status.New(
					nil, nil, nil,
					"1",
					"2", "2-",
					"3", "3-",
					"4",
					&svc.Dashboard,
				)
				svc.Status.Init(
					len(svc.Command), len(svc.Notify), len(svc.Webhook),
					status.ServiceInfo{
						ID:         svc.ID,
						Name:       svc.Name,
						Comment:    svc.Comment,
						ServiceURL: svc.LatestVersion.ServiceURL(),
					},
					&svc.Dashboard,
				)
				svc.Status.RefreshServiceInfo()

				return svc, nil
			}),
			want: &apitype.ServiceSummary{
				ID:                  "foo",
				Name:                new("bar"),
				Active:              new(false),
				Comment:             new("svc for blah"),
				Type:                "github",
				WebURL:              new("https://example.com"),
				Icon:                new("https://example.com/icon.png"),
				IconLinkTo:          new("https://example.com"),
				DeployedVersionType: new(dvweb.Type),
				Command:             new(2),
				Webhook:             new(3),
				Status: &apitype.Status{
					ApprovedVersion:          "1",
					DeployedVersion:          "2",
					DeployedVersionTimestamp: "2-",
					LatestVersion:            "3",
					LatestVersionTimestamp:   "3-",
					LastQueried:              "4",
				},
				Tags: &[]string{"hello", "there"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.svc.Summary()

			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf("ServiceSummary.String() mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

func TestService_UsingDefaults(t *testing.T) {
	tests := []struct {
		name                                                            string
		nilService                                                      bool
		usingNotifyDefaults, usingCommandDefaults, usingWebhookDefaults bool
	}{
		{
			name:                 "nil Service",
			nilService:           true,
			usingNotifyDefaults:  false,
			usingCommandDefaults: false,
			usingWebhookDefaults: false,
		},
		{
			name:                 "using all defaults",
			usingNotifyDefaults:  true,
			usingCommandDefaults: true,
			usingWebhookDefaults: true,
		},
		{
			name:                 "using no defaults",
			usingNotifyDefaults:  false,
			usingCommandDefaults: false,
			usingWebhookDefaults: false,
		},
		{
			name:                 "using Notify defaults",
			usingNotifyDefaults:  true,
			usingCommandDefaults: false,
			usingWebhookDefaults: false,
		},
		{
			name:                 "using Command defaults",
			usingNotifyDefaults:  false,
			usingCommandDefaults: true,
			usingWebhookDefaults: false,
		},
		{
			name:                 "using Webhook defaults",
			usingNotifyDefaults:  false,
			usingCommandDefaults: false,
			usingWebhookDefaults: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var svc *Service
			if !tc.nilService {
				svc = &Service{}
				svc.NotifyFromDefaults = tc.usingNotifyDefaults
				svc.CommandFromDefaults = tc.usingCommandDefaults
				svc.WebhookFromDefaults = tc.usingWebhookDefaults
			}

			usingNotifyDefaults, usingCommandDefaults, usingWebhookDefaults := svc.UsingDefaults()

			if usingNotifyDefaults != tc.usingNotifyDefaults ||
				usingCommandDefaults != tc.usingCommandDefaults ||
				usingWebhookDefaults != tc.usingWebhookDefaults {
				t.Errorf(
					"mismatch\ngot:  notify=%t, command=%t, webhook=%t\n"+
						"want: notify=%t, command=%t, webhook=%t",
					usingNotifyDefaults, usingCommandDefaults, usingWebhookDefaults,
					tc.usingNotifyDefaults, tc.usingCommandDefaults, tc.usingWebhookDefaults,
				)
			}
		})
	}
}

func TestService_GetName(t *testing.T) {
	tests := []struct {
		name string
		svc  *Service
		want string
	}{
		{
			name: "empty",
			svc:  &Service{},
			want: "",
		},
		{
			name: "ID used when no Name",
			svc:  &Service{ID: "foo"},
			want: "foo",
		},
		{
			name: "Name used when no ID",
			svc: &Service{
				Name: "bar",
			},
			want: "bar",
		},
		{
			name: "Name overrides ID",
			svc: &Service{
				ID:   "foo",
				Name: "bar",
			},
			want: "bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.svc.GetName()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
