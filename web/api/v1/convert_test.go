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

package v1

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	dvbase "github.com/release-argus/Argus/service/deployed_version/types/base"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	dockertest "github.com/release-argus/Argus/service/latest_version/filter/docker/test"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	statustest "github.com/release-argus/Argus/service/status/test"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

//
// Defaults.
//

func TestConvertAndCensorDefaults(t *testing.T) {
	// GIVEN: a config.Defaults.
	tests := []struct {
		name  string
		input *config.Defaults
		want  apitype.Defaults
	}{
		{
			name:  "nil",
			input: nil,
			want:  apitype.Defaults{},
		},
		{
			name: "bare",
			input: &config.Defaults{
				Service: service.Defaults{
					Options:               opt.Defaults{},
					LatestVersion:         lvbase.Defaults{},
					DeployedVersionLookup: dvbase.Defaults{},
					Dashboard:             dashboard.Defaults{},
				},
			},
			want: apitype.Defaults{
				Service: apitype.ServiceDefaults{
					Options: apitype.ServiceOptions{},
					LatestVersion: apitype.LatestVersionDefaults{
						Require: &apitype.LatestVersionRequireDefaults{},
					},
					Command:               apitype.Commands{},
					DeployedVersionLookup: apitype.DeployedVersionLookupDefaults{},
					Dashboard:             apitype.DashboardOptions{},
				},
			},
		},
		{
			name: "censor service.latest_version",
			input: &config.Defaults{
				Service: service.Defaults{
					Options: opt.Defaults{},
					LatestVersion: lvbase.Defaults{
						AccessToken: "censor",
						Require: *test.Must(t, func() (*filter.RequireDefaults, error) {
							return filter.DecodeDefaults(
								"yaml", []byte(test.TrimYAML(`
									docker:
										type: hub
										image: i
										tag: t
										registry:
											ghcr:
												image: iGHCR
												tag: tGHCR
												auth:
													username: something
													token: ghp_X
											hub:
												image: iHub
												tag: tHub
												auth:
													username: something
													token: hub_X
											quay:
												image: iQuay
												tag: tQuay
												auth:
													username: something
													token: quay_X
								`)),
							)
						}),
					},
					DeployedVersionLookup: dvbase.Defaults{},
					Dashboard:             dashboard.Defaults{},
				},
			},
			want: apitype.Defaults{
				Service: apitype.ServiceDefaults{
					Options: apitype.ServiceOptions{},
					LatestVersion: apitype.LatestVersionDefaults{
						AccessToken: util.SecretValue,
						Require: &apitype.LatestVersionRequireDefaults{
							Docker: apitype.RequireDockerDefaults{
								Type:  "hub",
								Image: "i",
								Tag:   "t",
								Registry: apitype.RequireDockerRegistriesDefaults{
									GHCR: &apitype.RequireDockerRegistryDefaultsToken{
										RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
											Image: "iGHCR",
											Tag:   "tGHCR",
										},
										RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
											Token: util.SecretValue,
										},
									},
									Hub: &apitype.RequireDockerCheckRegistryDefaultsTokenWithUsername{
										RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
											Image: "iHub",
											Tag:   "tHub",
										},
										RequireDockerRegistryDefaultsAuthWithUsername: apitype.RequireDockerRegistryDefaultsAuthWithUsername{
											Username: "something",
											RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
												Token: util.SecretValue,
											},
										},
									},
									Quay: &apitype.RequireDockerRegistryDefaultsToken{
										RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
											Image: "iQuay",
											Tag:   "tQuay",
										},
										RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
											Token: util.SecretValue,
										},
									},
								},
							},
						},
					},
					Command:               apitype.Commands{},
					DeployedVersionLookup: apitype.DeployedVersionLookupDefaults{},
					Dashboard:             apitype.DashboardOptions{},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorDefaults is called.
			result := convertAndCensorDefaults(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorDefaults()", packageName)

			// THEN: the result should be as expected.
			got := decode.ToYAMLString(result, "")
			want := decode.ToYAMLString(tc.want, "")
			if got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

//
// Service.
//

func TestConvertAndCensorService(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: a service.Service.
	tests := []struct {
		name  string
		input *service.Service
		want  *apitype.Service
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "bare",
			input: &service.Service{},
			want: &apitype.Service{
				Options:       apitype.ServiceOptions{},
				LatestVersion: nil,
				Command:       apitype.Commands{},
				Notify:        apitype.Notifiers{},
				WebHook:       apitype.WebHooks{},
				Dashboard:     apitype.DashboardOptions{},
			},
		},
		{
			name: "filled",
			input: test.Must(t, func() (*service.Service, error) {
				svc, _ := service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						name: Something
						comment: Comment on the Service
						options:
							active: false
						latest_version:
							type: github
							access_token: lv_accessToken
							url: `+test.ArgusGitHubRepo+`
						deployed_version:
							type: url
							url: https://example.com
							allow_invalid_certs: true
						notify:
							gotify:
								url_fields:
									host: https://gotify.example.com
									token: foo
						command:
							- ["echo", "foo"]
						webhook:
							test_wh:
								url: https://example.com
								allow_invalid_certs: true
								secret: foo
						dashboard:
							icon: https://example.com/icon.png
					`)),
					"test",
					svcCfg, notifyCfg, whCfg,
				)
				svcStatus, _ := statustest.New(
					"yaml", []byte(test.TrimYAML(`
						approved_version: 2.0.0
						deployed_version: 1.0.0
						deployed_version_timestamp: 2020-01-01T00:00:00Z
						latest_version: 3.0.0
						latest_version_timestamp: 2020-01-02T00:00:00Z
					`)),
				)
				svc.Status = *svcStatus.Copy(true)
				return svc, nil
			}),
			want: &apitype.Service{
				Name:    "Something",
				Comment: "Comment on the Service",
				Options: apitype.ServiceOptions{
					Active: test.Ptr(false),
				},
				LatestVersion: &apitype.LatestVersion{
					Type:        "github",
					URL:         test.ArgusGitHubRepo,
					AccessToken: util.SecretValue,
				},
				Command: apitype.Commands{
					{"echo", "foo"},
				},
				Notify: apitype.Notifiers{
					"gotify": &apitype.Notify{
						URLFields: map[string]string{
							"host":  "gotify.example.com",
							"token": util.SecretValue,
						},
					},
				},
				WebHook: apitype.WebHooks{
					"test_wh": apitype.WebHook{
						URL:               "https://example.com",
						AllowInvalidCerts: test.Ptr(true),
						Secret:            util.SecretValue,
					},
				},
				DeployedVersionLookup: &apitype.DeployedVersionLookup{
					Type:              "url",
					URL:               "https://example.com",
					AllowInvalidCerts: test.Ptr(true),
				},
				Dashboard: apitype.DashboardOptions{
					Icon: "https://example.com/icon.png",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var had string
			if tc.input != nil {
				if err, _ := tc.input.CheckValues(); err != nil {
					fmt.Printf(
						"%s\ninvalid test input: %v\n",
						packageName, err,
					)
				}
				had = tc.input.String("")
			}

			// WHEN: convertAndCensorService is called.
			result := convertAndCensorService(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorService()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original notifier\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

//
// Latest Version.
//

func TestConvertAndCensorLatestVersion(t *testing.T) {
	lvCfg := lvtest.PlainDefaultsConfig(t)
	// GIVEN: a latestver.Lookup.
	tests := []struct {
		name  string
		input latestver.Lookup
		want  *apitype.LatestVersion
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "github/bare",
			input: &lvgithub.Lookup{},
			want:  &apitype.LatestVersion{},
		},
		{
			name: "github/filled",
			input: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: owner/repo
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
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			want: &apitype.LatestVersion{
				Type:          "github",
				URL:           "owner/repo",
				AccessToken:   util.SecretValue,
				UsePreRelease: test.Ptr(false),
				URLCommands: apitype.URLCommands{
					{Type: "replace", Old: "this", New: "withThis"},
					{Type: "split", Text: "splitThis", Index: test.Ptr(8)},
					{Type: "regex", Regex: `([0-9.]+)`},
				},
				Require: &apitype.LatestVersionRequire{
					RegexContent: ".*",
				},
			},
		},
		{
			name:  "url/bare",
			input: &lvweb.Lookup{},
			want:  &apitype.LatestVersion{},
		},
		{
			name: "url/filled",
			input: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
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
								image: test/app
								tag: '{{ version }}'
								auth:
									token: not_telling_you
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			want: &apitype.LatestVersion{
				Type:              "url",
				URL:               "https://example.com",
				AllowInvalidCerts: test.Ptr(true),
				URLCommands: apitype.URLCommands{
					{Type: "replace", Old: "this", New: "withThis"},
					{Type: "split", Text: "splitThis", Index: test.Ptr(8)},
					{Type: "regex", Regex: `([0-9.]+)`},
				},
				Require: &apitype.LatestVersionRequire{
					Docker: &apitype.RequireDocker{
						Type:  "ghcr",
						Image: "test/app",
						Tag:   "{{ version }}",
						Token: util.SecretValue,
					},
				},
			},
		},
		{
			name:  "unknown",
			input: &lvtest.MockLookup{},
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var had string
			if tc.input != nil {
				had = tc.input.String("")
			}

			// WHEN: convertAndCensorLatestVersion is called.
			result := convertAndCensorLatestVersion(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorLatestVersion()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the original input is unchanged.
			if result == nil || tc.input == nil {
				if tc.input == nil && result != nil {
					t.Fatalf(
						"%s mismatch\ngot:  %q\nwant: nil",
						prefix, result,
					)
				}
				if result == nil && tc.want != nil {
					t.Fatalf(
						"%s mismatch\ngot:  nil\nwant: non-nil",
						prefix,
					)
				}
				return
			}
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorLatestVersionRequireDefaults(t *testing.T) {
	// GIVEN: a filter.RequireDefaults.
	tests := []struct {
		name  string
		input *filter.RequireDefaults
		want  *apitype.LatestVersionRequireDefaults
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "bare",
			input: &filter.RequireDefaults{},
			want:  &apitype.LatestVersionRequireDefaults{},
		},
		{
			name: "docker/bare",
			input: &filter.RequireDefaults{
				Docker: docker.Defaults{},
			},
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerDefaults{},
			},
		},
		{
			name: "docker/globals",
			input: test.Must(t, func() (*filter.RequireDefaults, error) {
				return filter.DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: hub
							image: i
							tag: t
					`)),
				)
			}),
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerDefaults{
					Type:  "hub",
					Image: "i",
					Tag:   "t",
				},
			},
		},
		{
			name: "docker/ghcr",
			input: test.Must(t, func() (*filter.RequireDefaults, error) {
				return filter.DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						docker:
							registry:
								ghcr:
									image: iGHCR
									tag: tGHCR
									auth:
										username: something
										token: ghp_X
					`)),
				)
			}),
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerDefaults{
					Registry: apitype.RequireDockerRegistriesDefaults{
						GHCR: &apitype.RequireDockerRegistryDefaultsToken{
							RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
								Image: "iGHCR",
								Tag:   "tGHCR",
							},
							RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
								Token: util.SecretValue,
							},
						},
					},
				},
			},
		},
		{
			name: "docker/hub",
			input: test.Must(t, func() (*filter.RequireDefaults, error) {
				return filter.DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						docker:
							registry:
								hub:
									image: iHub
									tag: tHub
									auth:
										username: something
										token: hub_X
					`)),
				)
			}),
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerDefaults{
					Registry: apitype.RequireDockerRegistriesDefaults{
						Hub: &apitype.RequireDockerCheckRegistryDefaultsTokenWithUsername{
							RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
								Image: "iHub",
								Tag:   "tHub",
							},
							RequireDockerRegistryDefaultsAuthWithUsername: apitype.RequireDockerRegistryDefaultsAuthWithUsername{
								Username: "something",
								RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
									Token: util.SecretValue,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "docker/quay",
			input: test.Must(t, func() (*filter.RequireDefaults, error) {
				return filter.DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						docker:
							registry:
								quay:
									image: iQuay
									tag: tQuay
									auth:
										username: something
										token: ghp_X
					`)),
				)
			}),
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerDefaults{
					Registry: apitype.RequireDockerRegistriesDefaults{
						Quay: &apitype.RequireDockerRegistryDefaultsToken{
							RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
								Image: "iQuay",
								Tag:   "tQuay",
							},
							RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
								Token: util.SecretValue,
							},
						},
					},
				},
			},
		},
		{
			name: "filled",
			input: test.Must(t, func() (*filter.RequireDefaults, error) {
				return filter.DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: hub
							image: i
							tag: t
							registry:
								ghcr:
									image: iGHCR
									tag: tGHCR
									auth:
										username: something
										token: ghp_X
								hub:
									image: iHub
									tag: tHub
									auth:
										username: something
										token: hub_X
								quay:
									image: iQuay
									tag: tQuay
									auth:
										username: something
										token: quay_X
					`)),
				)
			}),
			want: &apitype.LatestVersionRequireDefaults{
				Docker: apitype.RequireDockerDefaults{
					Type:  "hub",
					Image: "i",
					Tag:   "t",
					Registry: apitype.RequireDockerRegistriesDefaults{
						GHCR: &apitype.RequireDockerRegistryDefaultsToken{
							RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
								Image: "iGHCR",
								Tag:   "tGHCR",
							},
							RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
								Token: util.SecretValue,
							},
						},
						Hub: &apitype.RequireDockerCheckRegistryDefaultsTokenWithUsername{
							RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
								Image: "iHub",
								Tag:   "tHub",
							},
							RequireDockerRegistryDefaultsAuthWithUsername: apitype.RequireDockerRegistryDefaultsAuthWithUsername{
								Username: "something",
								RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
									Token: util.SecretValue,
								},
							},
						},
						Quay: &apitype.RequireDockerRegistryDefaultsToken{
							RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
								Image: "iQuay",
								Tag:   "tQuay",
							},
							RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
								Token: util.SecretValue,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := decode.ToYAMLString(tc.input, "")

			// WHEN: convertAndCensorLatestVersionRequireDefaults is called.
			result := convertAndCensorLatestVersionRequireDefaults(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorLatestVersionRequireDefaults()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := decode.ToYAMLString(tc.input, ""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorLatestVersionRequire(t *testing.T) {
	configDefaults, _ := plainDefaults(t)
	defaults := configDefaults.Service.LatestVersion.Require
	// GIVEN: a filter.Require.
	tests := []struct {
		name  string
		input *filter.Require
		want  *apitype.LatestVersionRequire
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "bare",
			input: &filter.Require{},
			want:  &apitype.LatestVersionRequire{},
		},
		{
			name: "docker/bare",
			input: &filter.Require{
				Docker: docker.RegistryMap["hub"](),
			},
			want: &apitype.LatestVersionRequire{
				Docker: &apitype.RequireDocker{},
			},
		},
		{
			name: "docker/ghcr",
			input: test.Must(t, func() (*filter.Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return filter.Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: test/app
							tag: '{{ version }}'
							auth:
								username: something
								token: ghp_X
					`)),
					svcStatus,
					&defaults,
				)
			}),
			want: &apitype.LatestVersionRequire{
				Docker: &apitype.RequireDocker{
					Type:  "ghcr",
					Image: "test/app",
					Tag:   "{{ version }}",
					Token: util.SecretValue,
				},
			},
		},
		{
			name: "docker/hub",
			input: test.Must(t, func() (*filter.Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return filter.Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: hub
							image: test/app
							tag: '{{ version }}'
							auth:
								username: user
								token: hub_X
					`)),
					svcStatus,
					&defaults,
				)
			}),
			want: &apitype.LatestVersionRequire{
				Docker: &apitype.RequireDocker{
					Type:     "hub",
					Image:    "test/app",
					Tag:      "{{ version }}",
					Username: "user",
					Token:    util.SecretValue,
				},
			},
		},
		{
			name: "docker/quay",
			input: test.Must(t, func() (*filter.Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return filter.Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: quay
							image: test/app
							tag: '{{ version }}'
							auth:
								username: something
								token: quay_X
					`)),
					svcStatus,
					&defaults,
				)
			}),
			want: &apitype.LatestVersionRequire{
				Docker: &apitype.RequireDocker{
					Type:  "quay",
					Image: "test/app",
					Tag:   "{{ version }}",
					Token: util.SecretValue,
				},
			},
		},
		{
			name: "filled",
			input: &filter.Require{
				Status: status.New(
					nil, nil, nil,
					"2.0.0",
					"1.0.0", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					"3.0.0", time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					&dashboard.Options{},
				),
				RegexContent: ".*",
				RegexVersion: `([0-9.]+)`,
				Command:      command.Command{"echo", "hello"},
				Docker: test.Must(t, func() (docker.Registry, error) {
					req, _ := filter.Decode(
						"yaml", []byte(test.TrimYAML(`
							docker:
								type: hub
								image: test/app
								tag: '{{ version }}'
								auth:
									username: something
									token: hub_X
						`)),
						test.Must(t, func() (*status.Status, error) {
							return statustest.New("yaml", nil)
						}),
						&defaults,
					)
					return req.Docker, nil
				}),
			},
			want: &apitype.LatestVersionRequire{
				Command: []string{"echo", "hello"},
				Docker: &apitype.RequireDocker{
					Type:     "hub",
					Image:    "test/app",
					Tag:      "{{ version }}",
					Username: "something",
					Token:    util.SecretValue,
				},
				RegexContent: ".*",
				RegexVersion: `([0-9.]+)`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorLatestVersionRequire is called.
			result := convertAndCensorLatestVersionRequire(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorLatestVersionRequire()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					packageName,
					got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorRequireDockerRegistryDefaults(t *testing.T) {
	// GIVEN: a docker.RegistryDefaults.
	tests := []struct {
		name  string
		input docker.RegistryDefaults
		want  apitype.RequireDockerRegistryDefaults
	}{
		{
			name:  "docker/ghcr/IsZero",
			input: &docker.GHCRRegistryDefaults{},
			want:  nil,
		},
		{
			name: "docker/ghcr/converted",
			input: &docker.GHCRRegistryDefaults{
				CommonRegistryDefaults: docker.CommonRegistryDefaults{
					ContainerDetail: docker.ContainerDetail{
						Image: "test/app",
						Tag:   "{{ version }}",
					},
					Auth: &docker.GHCRAuthDefaults{
						Token: "ghcr_X",
					},
				},
			},
			want: &apitype.RequireDockerRegistryDefaultsToken{
				RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
					Image: "test/app",
					Tag:   "{{ version }}",
				},
				RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
					Token: util.SecretValue,
				},
			},
		},
		{
			name:  "docker/hub/IsZero",
			input: &docker.HubRegistryDefaults{},
			want:  nil,
		},
		{
			name: "docker/hub/converted",
			input: &docker.HubRegistryDefaults{
				CommonRegistryDefaults: docker.CommonRegistryDefaults{
					ContainerDetail: docker.ContainerDetail{
						Image: "test/app",
						Tag:   "{{ version }}",
					},
					Auth: &docker.HubAuthDefaults{
						Username: "something",
						Token:    "ghcr_X",
					},
				},
			},
			want: &apitype.RequireDockerCheckRegistryDefaultsTokenWithUsername{
				RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
					Image: "test/app",
					Tag:   "{{ version }}",
				},
				RequireDockerRegistryDefaultsAuthWithUsername: apitype.RequireDockerRegistryDefaultsAuthWithUsername{
					Username: "something",
					RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
						Token: util.SecretValue,
					},
				},
			},
		},
		{
			name:  "docker/quay/IsZero",
			input: &docker.QuayRegistryDefaults{},
			want:  nil,
		},
		{
			name: "docker/quay/converted",
			input: &docker.QuayRegistryDefaults{
				CommonRegistryDefaults: docker.CommonRegistryDefaults{
					ContainerDetail: docker.ContainerDetail{
						Image: "test/app",
						Tag:   "{{ version }}",
					},
					Auth: &docker.QuayAuthDefaults{
						Token: "quay_X",
					},
				},
			},
			want: &apitype.RequireDockerRegistryDefaultsToken{
				RequireDockerRegistryDefaultsBase: apitype.RequireDockerRegistryDefaultsBase{
					Image: "test/app",
					Tag:   "{{ version }}",
				},
				RequireDockerRegistryDefaultsAuth: apitype.RequireDockerRegistryDefaultsAuth{
					Token: util.SecretValue,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorRequireDockerRegistryDefaults is called.
			result := convertAndCensorRequireDockerRegistryDefaults(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorRequireDockerRegistryDefaults()", packageName)

			// THEN: the result should be as expected.
			got := decode.ToYAMLString(result, "")
			want := decode.ToYAMLString(tc.want, "")
			if got != want {
				t.Errorf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorRequireDockerRegistryDefaults__unknownType(t *testing.T) {
	// GIVEN: a docker.RegistryDefaults of an unknown type.
	input := &dockertest.MockRegistryDefaults{}

	// WHEN: convertAndCensorRequireDockerRegistryDefaults is called.
	result := convertAndCensorRequireDockerRegistryDefaults(input)

	// THEN: the result should be as expected.
	if result != nil {
		t.Errorf(
			"%s\nconvertAndCensorRequireDockerRegistryDefaults() result mismatch with unknown struct type\ngot:  %q\nwant: nil",
			packageName, result,
		)
	}
}

func TestConvertURLCommands(t *testing.T) {
	// GIVEN: a list of URL Commands.
	tests := []struct {
		name  string
		input filter.URLCommands
		want  apitype.URLCommands
	}{
		{

			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty",
			input: filter.URLCommands{},
			want:  nil,
		},
		{
			name: "regex",
			input: filter.URLCommands{
				{Type: "regex", Regex: "[0-9.]+"},
			},
			want: apitype.URLCommands{
				{Type: "regex", Regex: "[0-9.]+"},
			},
		},
		{
			name: "replace",
			input: filter.URLCommands{
				{Type: "replace", Old: "foo", New: "bar"},
			},
			want: apitype.URLCommands{
				{Type: "replace", Old: "foo", New: "bar"},
			},
		},
		{
			name: "split",
			input: filter.URLCommands{
				{Type: "split", Index: test.Ptr(7)},
			},
			want: apitype.URLCommands{
				{Type: "split", Index: test.Ptr(7)},
			},
		},
		{
			name: "one of each",
			input: filter.URLCommands{
				{Type: "regex", Regex: "[0-9.]+"},
				{Type: "replace", Old: "foo", New: "bar"},
				{Type: "split", Index: test.Ptr(7)},
			},
			want: apitype.URLCommands{
				{Type: "regex", Regex: "[0-9.]+"},
				{Type: "replace", Old: "foo", New: "bar"},
				{Type: "split", Index: test.Ptr(7)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String()

			// WHEN: convertURLCommands is called on it.
			result := convertURLCommands(tc.input)

			prefix := fmt.Sprintf("%s\nconvertURLCommands()", packageName)

			// THEN: the WebHooks is converted correctly.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s\nconvertURLCommands() mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

//
// Deployed Version.
//

func TestConvertAndCensorDeployedVersionLookup(t *testing.T) {
	optCfg := opttest.PlainDefaultsConfig(t)
	dvCfg := dvtest.PlainDefaultsConfig(t)

	// GIVEN: a DeployedVersionLookup.
	tests := []struct {
		name                                      string
		input                                     deployedver.Lookup
		inputStatus                               *status.Status
		approvedVersion                           string
		deployedVersion, deployedVersionTimestamp string
		latestVersion, latestVersionTimestamp     string
		lastQueried                               string
		regexMissesContent, regexMissesVersion    int

		want *apitype.DeployedVersionLookup
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "url/bare",
			input: &dvweb.Lookup{},
			want: &apitype.DeployedVersionLookup{
				Type: "url",
			},
		},
		{
			name: "url/minimal",
			input: test.Must(t, func() (deployedver.Lookup, error) {
				return deployedver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
						json: version
					`)),
					nil,
					nil,
					dvCfg,
				)
			}),
			want: &apitype.DeployedVersionLookup{
				Type: "url",
				URL:  "https://example.com",
				JSON: "version",
			},
		},
		{
			name: "url/censor basic_auth.password",
			input: test.Must(t, func() (deployedver.Lookup, error) {
				return deployedver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
						basic_auth:
							username: alan
							password: pass123
					`)),
					nil,
					nil,
					dvCfg,
				)
			}),
			want: &apitype.DeployedVersionLookup{
				Type: "url",
				URL:  "https://example.com",
				BasicAuth: &apitype.BasicAuth{
					Username: "alan",
					Password: util.SecretValue,
				},
			},
		},
		{
			name: "url/censor headers",
			input: test.Must(t, func() (deployedver.Lookup, error) {
				return deployedver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
						headers:
							- key: X-Test-0
								value: `+util.SecretValue+`
							- key: X-Test-1
								value: `+util.SecretValue+`
					`)),
					nil,
					nil,
					dvCfg,
				)
			}),
			want: &apitype.DeployedVersionLookup{
				Type: "url",
				URL:  "https://example.com",
				Headers: []apitype.Header{
					{Key: "X-Test-0", Value: util.SecretValue},
					{Key: "X-Test-1", Value: util.SecretValue},
				},
			},
		},
		{
			name:               "url/filled",
			regexMissesContent: 1,
			regexMissesVersion: 3,
			input: test.Must(t, func() (deployedver.Lookup, error) {
				options, _ := opt.Decode(
					"yaml", []byte(test.TrimYAML(`
						interval: 10m
						semantic_versioning: true
					`)),
					optCfg,
				)

				return deployedver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						method: POST
						url: https://release-argus.io
						allow_invalid_certs: true
						target_header: X-Version
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
					`)),
					options,
					&status.Status{},
					dvCfg,
				)
			}),
			inputStatus: &status.Status{
				Fails: status.Fails{},
				ServiceInfo: serviceinfo.ServiceInfo{
					ID:     "service-id",
					WebURL: "https://release-argus.io",
				},
			},
			want: &apitype.DeployedVersionLookup{
				Type:              "url",
				Method:            http.MethodPost,
				URL:               "https://release-argus.io",
				AllowInvalidCerts: test.Ptr(true),
				TargetHeader:      "X-Version",
				BasicAuth: &apitype.BasicAuth{
					Username: "jim",
					Password: util.SecretValue,
				},
				Headers: []apitype.Header{
					{Key: "X-Test-0", Value: util.SecretValue},
					{Key: "X-Test-1", Value: util.SecretValue},
				},
				Body:          "body_here",
				JSON:          "version",
				Regex:         `([0-9]+\.[0-9]+\.[0-9]+)`,
				RegexTemplate: "$1.$2.$3",
			},
		},
		{
			name: "manual/filled",
			input: test.Must(t, func() (deployedver.Lookup, error) {
				dvStatus := status.New(
					nil, nil, nil,
					"1.0.0",
					"1.1.0", time.Now().UTC().Format(time.RFC3339),
					"1.2.3", time.Now().Add(time.Minute).UTC().Format(time.RFC3339),
					time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
					&dashboard.Options{},
				)
				// Need the ServiceID for the Query that's done because we have both Status and Options.
				dvStatus.ServiceInfo.ID = "manual"
				options, _ := opt.Decode(
					"yaml", []byte("semantic_versioning: true"),
					optCfg,
				)
				return deployedver.Decode(
					"yaml", []byte(`type: manual`),
					options,
					dvStatus,
					dvCfg,
				)
			}),
			want: &apitype.DeployedVersionLookup{
				Type:    "manual",
				Version: "1.1.0",
			},
		},
		{
			name:  "unknown type",
			input: &dvtest.MockLookup{},
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var had string
			if tc.input != nil {
				had = tc.input.String("")

				var dvStatus *status.Status
				if dv, ok := tc.input.(*dvweb.Lookup); ok {
					dvStatus = dv.Status
				}

				if tc.approvedVersion != "" {
					dvStatus.SetDeployedVersion("1.2.3", "", false)
					dvStatus.SetLatestVersion("1.2.3", time.Now().Format(time.RFC3339), false)
					dvStatus.SetApprovedVersion("1.2.3", false)
					dvStatus.SetLastQueried(time.Now().Format(time.RFC3339))
				}
				if tc.inputStatus != nil {
					dvStatus.Fails.Copy(&tc.inputStatus.Fails)
					dvStatus.ServiceInfo.ID = tc.inputStatus.ServiceInfo.ID
					dvStatus.ServiceInfo.WebURL = tc.inputStatus.ServiceInfo.WebURL
				}
				for range tc.regexMissesContent {
					dvStatus.RegexMissContent()
				}
				for range tc.regexMissesVersion {
					dvStatus.RegexMissVersion()
				}
			}

			// WHEN: convertAndCensorDeployedVersionLookup is called on it.
			result := convertAndCensorDeployedVersionLookup(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorDeployedVersionLookup", packageName)

			// THEN: the WebHooks is converted correctly.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}

			// AND: the original input is unchanged.
			if result == nil || tc.input == nil {
				if tc.input == nil && result != nil {
					t.Fatalf(
						"%s mismatch\ngot:  %q\nwant: nil",
						prefix, result,
					)
				}
				if result == nil && tc.want != nil {
					t.Fatalf(
						"%s mismatch\ngot:  nil\nwant: non-nil",
						prefix,
					)
				}
				return
			}
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

//
// Notify.
//

func TestConvertAndCensorNotifierDefaults(t *testing.T) {
	// GIVEN: shoutrrr.ShoutrrrsDefaults.
	tests := []struct {
		name  string
		input shoutrrr.ShoutrrrsDefaults
		want  apitype.Notifiers
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty",
			input: shoutrrr.ShoutrrrsDefaults{},
			want:  apitype.Notifiers{},
		},
		{
			name: "one",
			input: shoutrrr.ShoutrrrsDefaults{
				"test": shoutrrr.NewDefaults(
					"discord",
					map[string]string{
						"test": "1",
					},
					map[string]string{
						"test": "2",
					},
					map[string]string{
						"test": "3",
					},
				),
			},
			want: apitype.Notifiers{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"test": "1",
					},
					URLFields: map[string]string{
						"test": "2",
					},
					Params: map[string]string{
						"test": "3",
					},
				},
			},
		},
		{
			name: "multiple",
			input: shoutrrr.ShoutrrrsDefaults{
				"test": shoutrrr.NewDefaults(
					"discord",
					map[string]string{
						"test": "1",
					},
					map[string]string{
						"test": "2",
					},
					map[string]string{
						"test": "3",
					},
				),
				"other": shoutrrr.NewDefaults(
					"discord",
					map[string]string{
						"message": "release {{ version }} is available",
					},
					map[string]string{
						"apikey": "censor?",
					},
					map[string]string{
						"devices": "censor this too",
						"avatar":  "https://example.com",
					},
				),
			},
			want: apitype.Notifiers{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"test": "1",
					},
					URLFields: map[string]string{
						"test": "2",
					},
					Params: map[string]string{
						"test": "3",
					},
				},
				"other": {
					Type: "discord",
					Options: map[string]string{
						"message": "release {{ version }} is available",
					},
					URLFields: map[string]string{
						"apikey": util.SecretValue,
					},
					Params: map[string]string{
						"devices": util.SecretValue,
						"avatar":  "https://example.com",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorNotifiersDefaults is called.
			result := convertAndCensorNotifiersDefaults(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorNotifiersDefaults()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}

			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorNotifiers(t *testing.T) {
	// GIVEN: shoutrrr.Shoutrrrs.
	tests := []struct {
		name  string
		input shoutrrr.Shoutrrrs
		want  apitype.Notifiers
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty",
			input: shoutrrr.Shoutrrrs{},
			want:  nil,
		},
		{
			name: "one",
			input: shoutrrr.Shoutrrrs{
				"test": shoutrrr.New(
					nil,
					"", "discord",
					map[string]string{
						"test": "1",
					},
					map[string]string{
						"altid":    "VALUE",
						"apikey":   "VALUE",
						"botkey":   "VALUE",
						"password": "VALUE",
						"token":    "VALUE",
						"tokena":   "VALUE",
						"tokenb":   "VALUE",
					},
					map[string]string{
						"devices": "VALUE",
					},
					nil, nil, nil,
				),
			},
			want: apitype.Notifiers{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"test": "1",
					},
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
		},
		{
			name: "multiple",
			input: shoutrrr.Shoutrrrs{
				"test": shoutrrr.New(
					nil,
					"", "discord",
					map[string]string{
						"test": "1",
					},
					map[string]string{
						"altid":    "VALUE",
						"apikey":   "VALUE",
						"botkey":   "VALUE",
						"password": "VALUE",
						"token":    "VALUE",
						"tokena":   "VALUE",
						"tokenb":   "VALUE",
					},
					map[string]string{
						"devices": "VALUE",
					},
					nil, nil, nil,
				),
				"other": shoutrrr.New(
					nil,
					"", "discord",
					map[string]string{
						"message": "release {{ version }} is available",
					},
					map[string]string{
						"apikey": "censor?",
					},
					map[string]string{
						"devices": "censor this too",
						"avatar":  "https://example.com",
					},
					nil, nil, nil,
				),
			},
			want: apitype.Notifiers{
				"test": {
					Type: "discord",
					Options: map[string]string{
						"test": "1",
					},
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
				"other": {
					Type: "discord",
					Options: map[string]string{
						"message": "release {{ version }} is available",
					},
					URLFields: map[string]string{
						"apikey": util.SecretValue,
					},
					Params: map[string]string{
						"devices": util.SecretValue,
						"avatar":  "https://example.com",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorNotifiers is called.
			result := convertAndCensorNotifiers(tc.input)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s\nconvertAndCensorNotifiers() mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s\nconvertAndCensorNotifiers() changed original input\ngot:  %q\nwant: %q",
					packageName, got, had,
				)
			}
		})
	}
}

//
// Command.
//

func TestConvertCommands(t *testing.T) {
	// GIVEN: Commands.
	tests := []struct {
		name  string
		input command.Commands
		want  apitype.Commands
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty",
			input: command.Commands{},
			want:  apitype.Commands{},
		},
		{
			name: "one",
			input: command.Commands{
				{"ls", "-lah"},
			},
			want: apitype.Commands{
				{"ls", "-lah"},
			},
		},
		{
			name: "two",
			input: command.Commands{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"},
			},
			want: apitype.Commands{
				{"ls", "-lah"},
				{"/bin/bash", "something.sh"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := decode.ToYAMLString(tc.input, "")

			// WHEN: convertCommands is called on it.
			got := convertCommands(tc.input)

			prefix := fmt.Sprintf(
				"%s\nconvertCommands(%+v)",
				packageName, tc.input,
			)

			// THEN: the Commands is converted correctly.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b apitype.Command) bool { return util.AreSlicesEqual(a, b) },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: the original input is unchanged.
			if got := decode.ToYAMLString(tc.input, ""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

//
// WebHook.
//

func TestConvertAndCensorWebHooksDefaults(t *testing.T) {
	// GIVEN: webhook.WebHooksDefaults.
	tests := []struct {
		name  string
		input webhook.WebHooksDefaults
		want  apitype.WebHooks
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty",
			input: webhook.WebHooksDefaults{},
			want:  apitype.WebHooks{},
		},
		// {
		// 	name: "nil and empty elements",
		// 	input: &webhook.WebHooksDefaults{
		// 		"test":  webhook.Defaults{},
		// 		"other": nil,
		// 	},
		// 	want: &apitype.WebHooks{
		// 		"test":  {},
		// 		"other": nil,
		// 	},
		// },
		{
			name: "one",
			input: webhook.WebHooksDefaults{
				"test": test.Must(t, func() (*webhook.Defaults, error) {
					return webhook.DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							headers:
								- key: X-Test
								  value: foo
							secret: censor
							type: github
							url: https://example.com
						`)),
					)
				}),
			},
			want: apitype.WebHooks{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					Headers: []apitype.Header{
						{Key: "X-Test", Value: util.SecretValue},
					},
				},
			},
		},
		{
			name: "multiple",
			input: webhook.WebHooksDefaults{
				"test": test.Must(t, func() (*webhook.Defaults, error) {
					return webhook.DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							headers:
								- key: X-Test
								  value: foo
							secret: censor
							type: github
							url: https://example.com
						`)),
					)
				}),
				"other": test.Must(t, func() (*webhook.Defaults, error) {
					return webhook.DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							type: gitlab
							url: https://release-argus.io
						`)),
					)
				}),
			},
			want: apitype.WebHooks{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					Headers: []apitype.Header{
						{Key: "X-Test", Value: util.SecretValue},
					},
				},
				"other": {
					Type: "gitlab",
					URL:  "https://release-argus.io",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorWebHooksDefaults is called.
			result := convertAndCensorWebHooksDefaults(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorWebHooksDefaults()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorWebHookDefaults(t *testing.T) {
	// GIVEN: a webhook.Defaults.
	tests := []struct {
		name  string
		input webhook.Defaults
		want  apitype.WebHook
	}{
		{
			name:  "empty",
			input: webhook.Defaults{},
			want:  apitype.WebHook{},
		},
		{
			name: "filled",
			input: *test.Must(t, func() (*webhook.Defaults, error) {
				return webhook.DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						allow_invalid_certs: true
						delay: 1s
						desired_status_code: 200
						headers:
							- key: X-Test
								value: foo
						max_tries: 3
						secret: censor
						silent_fails: true
						type: github
						url: https://example.com
					`)),
				)
			}),
			want: apitype.WebHook{
				Type:              "github",
				URL:               "https://example.com",
				AllowInvalidCerts: test.Ptr(true),
				Secret:            util.SecretValue,
				Headers: []apitype.Header{
					{Key: "X-Test", Value: util.SecretValue},
				},
				DesiredStatusCode: test.Ptr[uint16](200),
				Delay:             "1s",
				MaxTries:          test.Ptr[uint8](3),
				SilentFails:       test.Ptr(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorWebHookDefaults is called.
			result := convertAndCensorWebHookDefaults(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorWebHookDefaults()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(""), tc.want.String(""); got != want {
				t.Errorf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorWebHooks(t *testing.T) {
	// GIVEN: webhook.WebHooks.
	tests := []struct {
		name  string
		input webhook.WebHooks
		want  apitype.WebHooks
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty",
			input: webhook.WebHooks{},
			want:  apitype.WebHooks{},
		},
		{
			name: "one",
			input: webhook.WebHooks{
				"test": webhook.New(
					nil,
					webhook.Headers{
						{Key: "X-Test", Value: "foo"},
					},
					"",
					nil, nil,
					"test",
					nil, webhook.Notifiers{}, nil,
					"censor",
					nil,
					"github", "https://example.com",
					nil, nil, nil,
				),
			},
			want: apitype.WebHooks{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					Headers: []apitype.Header{
						{Key: "X-Test", Value: util.SecretValue},
					},
				},
			},
		},
		{
			name: "multiple",
			input: webhook.WebHooks{
				"test": webhook.New(
					nil,
					webhook.Headers{
						{Key: "X-Test", Value: "foo"},
					},
					"", nil, nil,
					"test",
					nil, webhook.Notifiers{}, nil,
					"censor",
					nil,
					"github", "https://example.com",
					nil, nil, nil,
				),
				"other": webhook.New(
					nil, nil, "", nil, nil,
					"other",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"gitlab", "https://release-argus.io",
					nil, nil, nil,
				),
			},
			want: apitype.WebHooks{
				"test": {
					Type:   "github",
					URL:    "https://example.com",
					Secret: util.SecretValue,
					Headers: []apitype.Header{
						{Key: "X-Test", Value: util.SecretValue},
					},
				},
				"other": {
					Type: "gitlab",
					URL:  "https://release-argus.io",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String()

			// WHEN: convertAndCensorWebHooks is called.
			result := convertAndCensorWebHooks(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorWebHooks()", packageName)

			// THEN: the result should be as expected.
			if got, want := result.String(), tc.want.String(); got != want {
				t.Errorf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertAndCensorWebHook(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name  string
		input *webhook.WebHook
		want  apitype.WebHook
	}{
		{
			name:  "nil",
			input: nil,
			want:  apitype.WebHook{},
		},
		{
			name:  "empty",
			input: &webhook.WebHook{},
			want:  apitype.WebHook{},
		},
		{
			name: "censor secret",
			input: webhook.New(
				nil, nil,
				"",
				nil, nil,
				"wh",
				nil, webhook.Notifiers{}, nil,
				"shazam",
				nil,
				"", "",
				nil, nil, nil,
			),
			want: apitype.WebHook{
				Secret: util.SecretValue,
			},
		},
		{
			name: "copy and censor headers",
			input: webhook.New(
				nil,
				webhook.Headers{
					{Key: "X-Something", Value: "foo"},
					{Key: "X-Another", Value: "bar"},
				},
				"",
				nil, nil,
				"wh",
				nil, webhook.Notifiers{}, nil,
				"",
				nil,
				"", "",
				nil, nil, nil,
			),
			want: apitype.WebHook{
				Headers: []apitype.Header{
					{Key: "X-Something", Value: util.SecretValue},
					{Key: "X-Another", Value: util.SecretValue},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := tc.input.String("")

			// WHEN: convertAndCensorWebHook is called on it.
			result := convertAndCensorWebHook(tc.input)

			prefix := fmt.Sprintf("%s\nconvertAndCensorWebHook()", packageName)

			// THEN: the WebHook is converted correctly.
			if got, want := result.String(""), tc.want.String(""); got != want {
				t.Errorf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					*config.WebRoutePrefix, got, want,
				)
			}

			// AND: the original input is unchanged.
			if got := tc.input.String(""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}

func TestConvertWebHookHeaders(t *testing.T) {
	// GIVEN: a webhook.Headers.
	tests := []struct {
		name  string
		input webhook.Headers
		want  []apitype.Header
	}{
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty",
			input: webhook.Headers{},
			want:  []apitype.Header{},
		},
		{
			name: "one header",
			input: webhook.Headers{
				{Key: "X-Test", Value: "foo"},
			},
			want: []apitype.Header{
				{Key: "X-Test", Value: "foo"},
			},
		},
		{
			name: "multiple headers",
			input: webhook.Headers{
				{Key: "X-Test-1", Value: "foo"},
				{Key: "X-Test-2", Value: "bar"},
			},
			want: []apitype.Header{
				{Key: "X-Test-1", Value: "foo"},
				{Key: "X-Test-2", Value: "bar"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := decode.ToYAMLString(tc.input, "")

			// WHEN: convertWebHookHeaders is called.
			got := convertWebHookHeaders(tc.input)

			prefix := fmt.Sprintf(
				"%s\nconvertWebHookHeaders(%+v)",
				packageName, tc.input,
			)

			// THEN: the result should be as expected.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b apitype.Header) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: the original input is unchanged.
			if got := decode.ToYAMLString(tc.input, ""); got != had {
				t.Errorf(
					"%s changed original input\ngot:  %q\nwant: %q",
					prefix, got, had,
				)
			}
		})
	}
}
