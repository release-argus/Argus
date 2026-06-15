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
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	"github.com/release-argus/Argus/util/errfmt"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestDecodeServices(t *testing.T) {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: data in a given format to Decode into Services.
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
			name:     "YAML/empty",
			format:   "yaml",
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
			name:     "JSON/null",
			format:   "json",
			data:     "null",
			want:     "",
			errRegex: `^$`,
		},
		{
			name:     "YAML/null",
			format:   "yaml",
			data:     "null",
			want:     "",
			errRegex: `^$`,
		},
		{
			name:   "JSON/single service",
			format: "json",
			data: test.TrimJSON(`{
				"service1": {
					"latest_version": {
						"type": "github",
						"url": "owner/repo"
					}
				}
			}`),
			want: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/single service",
			format: "yaml",
			data: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo
			`),
			want: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo
			`),
			errRegex: `^$`,
		},
		{
			name:   "JSON/multiple services",
			format: "json",
			data: test.TrimJSON(`{
				"service1": {
					"latest_version": {
						"type": "github",
						"url": "owner/repo1"
					}
				},
				"service2": {
					"name": "service2",
					"latest_version": {
						"type": "github",
						"url": "owner/repo2"
					},
					"deployed_version": {
						"type": "url",
						"method": "GET",
						"url": "` + test.LookupPlain["url_valid"] + `"
					}
				}
			}`),
			want: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo1
				service2:
					name: service2
					latest_version:
						type: github
						url: owner/repo2
					deployed_version:
						type: url
						method: GET
						url: ` + test.LookupPlain["url_valid"] + `
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/multiple services",
			format: "yaml",
			data: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo1
				service2:
					name: service2
					latest_version:
						type: github
						url: owner/repo2
					deployed_version:
						type: url
						method: GET
						url: ` + test.LookupPlain["url_valid"] + `
			`),
			want: test.TrimYAML(`
				service1:
					latest_version:
						type: github
						url: owner/repo1
				service2:
					name: service2
					latest_version:
						type: github
						url: owner/repo2
					deployed_version:
						type: url
						method: GET
						url: ` + test.LookupPlain["url_valid"] + `
			`),
			errRegex: `^$`,
		},
		{
			name:   "nil service is removed",
			format: "json",
			data: test.TrimJSON(`{
				"service1": null,
				"service2": {
					"latest_version": {
						"type": "github",
						"url": "owner/repo"
					}
				}
			}`),
			want: test.TrimYAML(`
				service2:
					latest_version:
						type: github
						url: owner/repo
			`),
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid data type",
			format: "json",
			data:   `{"invalid": "json"}`,
			errRegex: test.TrimYAML(`
				^service:
					"invalid":
						json: .*unmarshal .*$`,
			),
		},
		{
			name:   "YAML/invalid",
			format: "yaml",
			data:   `invalid: [yaml: syntax`,
			errRegex: test.TrimYAML(`
				^service:
					[^\s]+ sequence end token.*
					[^\s]+ .*
					\s+\^$`,
			),
		},
		{
			name:   "JSON/invalid Service",
			format: "json",
			data:   `{"service1": []}`,
			errRegex: test.TrimYAML(`
				^service:
					"service1":
						json: .*unmarshal .*$`,
			),
		},
		{
			name:   "YAML/invalid Service",
			format: "yaml",
			data:   `service1: []`,
			errRegex: test.TrimYAML(`
				^service:
					"service1":
						[^\s]+ sequence was used.*
						[^\s]+.*
						\s+\^$`,
			),
		},
		{
			name:   "JSON/invalid latest_version",
			format: "json",
			data: test.TrimJSON(`{
				"service1": {
					"latest_version": {
						"type": "something"
					}
				}
			}`),
			errRegex: test.TrimYAML(`
				^service:
					"service1":
						latest_version:
							type: "something" <invalid>.*$`,
			),
		},
		{
			name:   "YAML/invalid latest_version",
			format: "yaml",
			data: test.TrimYAML(`
				service1:
					latest_version:
						type: something
			`),
			errRegex: test.TrimYAML(`
				^service:
					"service1":
						latest_version:
							type: "something" <invalid>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.errRegex = strings.ReplaceAll(tc.errRegex, "__name__", tc.name)

			_, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (Services, error) {
					return DecodeServices(
						format, data,
						svcCfg, notifyCfg, whCfg,
					)
				},
				tc.format, tc.data,
				func(v Services) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeServices",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDecodeServices__MarshalError(t *testing.T) {
	// GIVEN: a failing marshal function.
	original := marshalServiceRaw
	customErr := fmt.Errorf("marshal failed")
	marshalServiceRaw = func(format string, m any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalServiceRaw = original })

	// AND: Service, Notify and WebHook config defaults.
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// AND: data to decode.
	data := `{"service1": {"comment": "test"}}`

	// WHEN: DecodeServices is called.
	_, err := DecodeServices(
		"json", []byte(data),
		svcCfg, notifyCfg, whCfg,
	)

	prefix := fmt.Sprintf("%s\nDecodeServices(marshal error)", packageName)

	// THEN: the marshal error is returned.
	got := errfmt.FormatError(err)
	wantErr := &decode.KeyFieldError{
		Key: "service",
		Err: customErr,
	}
	want := errfmt.FormatError(wantErr)
	if got != want {
		t.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
}

func TestDecodeService(t *testing.T) {
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: data in a given format to Decode into a Service.
	tests := []struct {
		name              string
		format, data      string
		id                *string
		emptyHardDefaults bool
		want              string
		errRegex          string
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
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid",
			format: "json",
			data:   `{invalid: json}`,
			errRegex: test.TrimYAML(`
				^"__name__":
					jsontext: invalid character.*`,
			),
		},
		{
			name:     "JSON/invalid - empty ID",
			format:   "json",
			data:     `{invalid: json}`,
			id:       test.Ptr(""),
			errRegex: `invalid character`,
		},
		{
			name:   "JSON/latest_version: valid type - github",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
		},
		{
			name:   "JSON/latest_version, valid type - github (filled)",
			format: "json",
			data: test.TrimJSON(`{
				"name": "foo",
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `",
					"require": {
						"docker": {
							"image": "releaseargus/argus"
						}
					},
					"access_token": "foo",
					"url_commands": [
						{
							"type": "regex",
							"regex": ".*"
						}
					],
					"use_prerelease": true
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				name: foo
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
					url_commands:
						- type: regex
							regex: .*
					require:
						docker:
							image: releaseargus/argus
					access_token: foo
					use_prerelease: true
			`),
		},
		{
			name:   "JSON/latest_version, github - invalid",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": ["` + test.ArgusGitHubRepo + `"]
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					latest_version:
						json: .*unmarshal .* string.*$`,
			),
		},
		{
			name:   "JSON/latest_version, github - invalid, empty ID",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": ["` + test.ArgusGitHubRepo + `"]
				}
			}`),
			id: test.Ptr(""),
			errRegex: test.TrimYAML(`
				^latest_version:
					json: .*unmarshal .* string.*$`,
			),
		},
		{
			name:   "JSON/latest_version, valid type - url",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				latest_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name:   "JSON/latest_version, valid type - url (full)",
			format: "json",
			data: test.TrimJSON(`{
				"name": "bar",
				"latest_version": {
					"type": "url",
					"url": "https://example.com",
					"require": {
						"docker": {
							"image": "releaseargus/argus"
						}
					},
					"allow_invalid_certs": true,
					"url_commands": [
						{"type": "regex", "regex": ".*"}
					]
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				name: bar
				latest_version:
					type: url
					url: https://example.com
					url_commands:
						- type: regex
							regex: .*
					require:
						docker:
							image: releaseargus/argus
					allow_invalid_certs: true
			`),
		},
		{
			name:   "JSON/latest_version, url - invalid",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "url",
					"url": ["https://example.com"]
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					latest_version:
						json: .*unmarshal .* string.*$`,
			),
		},
		{
			name:   "JSONL latest_version, valid type - web (url alias)",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "web",
					"url": "https://example.com"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				latest_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name:   "JSON/latest_version: unknown type",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "unsupported"
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					latest_version:
						type: "unsupported" <invalid> .*\['github', 'url'\].*$`,
			),
		},
		{
			name:   "JSON/latest_version: missing type",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"url": "https://example.com"
				}
			}`),
			emptyHardDefaults: true,
			errRegex: test.TrimYAML(`
				^"__name__":
					latest_version:
						type: <required> .*\['github', 'url'\].*$`,
			),
		},
		{
			name:   "JSON/latest_version, invalid type format",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": ["unsupported"]
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					latest_version:
						json: .*unmarshal.* string.*`,
			),
		},
		{
			name:     "JSON/latest_version=null, no deployed_version",
			format:   "json",
			data:     `{"latest_version": null}`,
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/no latest_version, have deployed_version",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "url",
					"method": "GET",
					"url": "` + test.LookupPlain["url_valid"] + `"
				}
			}`),
			want: test.TrimYAML(`
				deployed_version:
					type: url
					method: GET
					url: https://valid.release-argus.io/plain
			`),
		},
		{
			name:   "JSON/deployed_version, valid type - url",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "url",
					"url": "https://example.com"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name:   "JSON/deployed_version, valid type - url (full)",
			format: "json",
			data: test.TrimJSON(`{
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
						{"key": "foo",       "value": "bar"},
						{"key": "something", "value": "else"}
					],
					"body": "removed_on_verify",
					"regex": "(\\d+)\\.(\\d+)\\.(\\d+)",
					"regex_template": "$3.$2.$1"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
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
			`),
		},
		{
			name:   "JSON/deployed_version, valid type - web (url alias)",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "web",
					"url": "https://example.com"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name:   "JSON/deployed_version, url - invalid",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "url",
					"url": ["https://example.com"]
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					deployed_version:
						json: .*unmarshal.*$`,
			),
		},
		{
			name:   "JSON/deployed_version, unknown type",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": "unsupported"
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					deployed_version:
						type: "unsupported" <invalid> .*\['manual', 'url'\].*$`,
			),
		},
		{
			name:   "JSON/deployed_version, missing type",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"url": "https://example.com"
				}
			}`),
			emptyHardDefaults: true,
			errRegex: test.TrimYAML(`
				^"__name__":
					deployed_version:
						type: <required> .*\['manual', 'url'\].*$`,
			),
		},
		{
			name:   "JSON/deployed_version, invalid type format",
			format: "json",
			data: test.TrimJSON(`{
				"deployed_version": {
					"type": ["unsupported"]
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					deployed_version:
						json: .*unmarshal.*$`,
			),
		},
		{
			name:     "JSON/no latest_version, deployed_version=null",
			format:   "json",
			data:     `{"deployed_version": null}`,
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/have latest_version, no deployed_version",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "url",
					"url": "` + test.LookupPlain["url_valid"] + `"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				latest_version:
					type: url
					url: ` + test.LookupPlain["url_valid"] + `
			`),
		},
		{
			name:   "JSON/dashboard.tags - []string",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				},
				"dashboard": {
					"tags": [
						"foo",
						"bar"
					]
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				dashboard:
					tags:
						- foo
						- bar
			`),
		},
		{
			name:   "JSON/dashboard.tags - string",
			format: "json",
			data: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				},
				"dashboard": {
					"tags": "foo"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				dashboard:
					tags:
						- foo
			`),
		},
		{
			name:   "JSON/dashboard.tags - invalid",
			format: "json",
			data: test.TrimJSON(`{
				"dashboard": {
					"tags": {
						"foo": "bar"
					}
				}
			}`),
			errRegex: test.TrimYAML(`
				^"__name__":
					json: .*unmarshal.*
						tags: expected a string or a list of strings, got map.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svcCfg := plainDefaultsConfig(t)
			if tc.emptyHardDefaults {
				svcCfg.Hard = &Defaults{}
				svcCfg.Soft.SetDefaults(svcCfg.Hard)
			}
			id := tc.name
			if tc.id != nil {
				id = *tc.id
			}
			tc.errRegex = strings.ReplaceAll(tc.errRegex, "__name__", id)

			got, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Service, error) {
					return DecodeService(
						format, data,
						id,
						svcCfg, notifyCfg, whCfg,
					)
				},
				tc.format, tc.data,
				func(v *Service) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeService",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if got == nil || err != nil {
				return
			}

			// THEN: the .ID is set want.
			if got.ID != tc.name {
				t.Errorf(
					"%s\nDecodeService(format=%q, data=%q) .ID mismatch\ngot:  %q\nwant: %q",
					packageName, tc.format, tc.data,
					got.ID, tc.name,
				)
			}
		})
	}
}

func TestApplyOverrides(t *testing.T) {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	type args struct {
		format string
		data   string
		target *Service
	}
	// GIVEN: args to apply overrides to a given Service.
	tests := []struct {
		name     string
		args     args
		want     string
		errRegex string
	}{
		{
			name: "no overrides - github",
			args: args{
				format: "yaml",
				data:   "",
				target: testService(t, "-id-", "github", "url"),
			},
			want:     testService(t, "-id-", "github", "url").String(""),
			errRegex: `^$`,
		},
		{
			name: "no overrides - url",
			args: args{
				format: "yaml",
				data:   "",
				target: testService(t, "-id-", "url", "url"),
			},
			want:     testService(t, "-id-", "url", "url").String(""),
			errRegex: `^$`,
		},
		{
			name: "null overrides = remove",
			args: args{
				format: "yaml",
				data:   "null",
				target: testService(t, "-id-", "url", "url"),
			},
			errRegex: `^$`,
		},
		{
			name: "nil target creates fresh",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					latest_version:
						type: github
						url: ` + test.ArgusGitHubRepo + `
				`),
				target: nil,
			},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name: "nil target creates fresh - error",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					latest_version:
						type: unknown-type
						url: ` + test.ArgusGitHubRepo + `
				`),
				target: nil,
			},
			errRegex: test.TrimYAML(`
				^"__name__":
					latest_version:
						type: "unknown-type" <invalid> .*$`,
			),
		},
		{
			name: "fail to extract 'latest_version'",
			args: args{
				format: "yaml",
				data: test.TrimJSON(`{
					"latest_version": {
						"type": "github",
						"url": "` + test.ArgusGitHubRepo + `"
				}`),
				target: testService(t, "-id-", "url", "url"),
			},
			errRegex: test.TrimYAML(`
				^"-id-":
					[^\s]+ could not find flow .*`,
			),
		},
		{
			name: "latest_version - change fields",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					latest_version:
						url: release-argus/Test
						access_token: abc
				`),
				target: test.Must(t, func() (*Service, error) {
					return DecodeService(
						"yaml", []byte(test.TrimYAML(`
							name: foo
							comment: bar
							options:
								interval: 1s
								active: false
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
							deployed_version:
							`+dvtest.Lookup(t, "url", false, "").String("  ")+`
							notify:
								"1":
							`+shoutrrrtest.Shoutrrr(t, true, true).String("    ")+`
								"2":
							`+shoutrrrtest.Shoutrrr(t, true, false).String("    ")+`
							command:
							  - ["ls", "-lah"]
							webhook:
								"a":
							`+whtest.WebHook(t, false, true, true).String("    ")+`
								"b":
							`+whtest.WebHook(t, false, false, true).String("    ")+`
						`)),
						"latest_version - change fields",
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
			want: test.TrimYAML(`
				name: foo
				comment: bar
				options:
					interval: 1s
					active: false
				latest_version:
					type: github
					url: release-argus/Test
					access_token: abc
				deployed_version:` + "\n" +
				dvtest.Lookup(t, "url", false, "").String("  ") +
				`notify:
					'1':` + "\n" + shoutrrrtest.Shoutrrr(t, true, true).String("    ") +
				`					'2':` + "\n" + shoutrrrtest.Shoutrrr(t, true, false).String("    ") +
				`command:
					- - ls
					  - -lah
				webhook:
					a:` + "\n" + whtest.WebHook(t, false, true, true).String("    ") +
				`					b:` + "\n" + whtest.WebHook(t, false, false, true).String("    "),
			),
			errRegex: `^$`,
		},
		{
			name: "latest_version - invalid field types",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					latest_version:
						url: ["https://example.com"]
				`),
				target: test.Must(t, func() (*Service, error) {
					return DecodeService(
						"yaml", []byte(test.TrimYAML(`
							latest_version:
								type: url
								url: "https://example.com"
						`)),
						"latest_version - invalid field types",
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
			errRegex: test.TrimYAML(`
				"latest_version - invalid field types":
					latest_version:
						[^\s]+ .*unmarshal.*`,
			),
		},
		{
			name: "deployed_version - change fields",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					deployed_version:
						url: https://example.com
						allow_invalid_certs: true
				`),
				target: test.Must(t, func() (*Service, error) {
					return DecodeService(
						"yaml", []byte(test.TrimYAML(`
							name: foo
							comment: bar
							options:
								interval: 1s
								active: false
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
							deployed_version:
								type: web
								url: https://release-argus.io
								allow_invalid_certs: false
								json: foo
							notify:
								"1":
							`+shoutrrrtest.Shoutrrr(t, true, true).String("    ")+`
								"2":
							`+shoutrrrtest.Shoutrrr(t, true, false).String("    ")+`
							command:
							  - ["ls", "-lah"]
								- ["docker", "compose", "up"]
							webhook:
								"a":
							`+whtest.WebHook(t, false, true, true).String("    ")+`
								"b":
							`+whtest.WebHook(t, false, false, true).String("    ")+`
						`)),
						"deployed_version - change fields",
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
			want: test.TrimYAML(`
				name: foo
				comment: bar
				options:
					interval: 1s
					active: false
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				deployed_version:
					type: url
					url: https://example.com
					allow_invalid_certs: true
					json: foo
				notify:
					'1':` + "\n" + shoutrrrtest.Shoutrrr(t, true, true).String("    ") +
				`					'2':` + "\n" + shoutrrrtest.Shoutrrr(t, true, false).String("    ") +
				`command:
					- - ls
					  - -lah
					- - docker
					  - compose
					  - up
				webhook:
					a:` + "\n" + whtest.WebHook(t, false, true, true).String("    ") +
				`					b:` + "\n" + whtest.WebHook(t, false, false, true).String("    "),
			),
			errRegex: `^$`,
		},
		{
			name: "deployed_version - invalid field types",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					deployed_version:
						type: web
						url: ["https://example.com"]
				`),
				target: test.Must(t, func() (*Service, error) {
					return DecodeService(
						"yaml", nil,
						"deployed_version - invalid field types",
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
			errRegex: test.TrimYAML(`
				^"deployed_version - invalid field types":
					deployed_version:
						[^\s]+ .*unmarshal.*`,
			),
		},
		{
			name: "comment - invalid data type",
			args: args{
				format: "yaml",
				data:   `comment: ["array","not","supported"]`,
				target: test.Must(t, func() (*Service, error) {
					return DecodeService(
						"yaml", nil,
						"comment - invalid data type",
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
			errRegex: test.TrimYAML(`
				^"__name__":
					[^\s]+ .*unmarshal.*
					[^\s]+ .*
					\s+\^$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id := tc.name
			if tc.args.target != nil && tc.args.target.ID != "" {
				id = tc.args.target.ID
			}
			tc.errRegex = strings.ReplaceAll(tc.errRegex, "__name__", tc.name)

			// WHEN: these overrides are applied to the Service.
			if _, _, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v *Service) (*Service, error) {
					return ApplyOverrides(
						format, data,
						v,
						id,
						svcCfg,
						notifyCfg,
						whCfg,
					)
				},
				tc.args.format, tc.args.data,
				func(v *Service) string { return v.String("") },
				tc.want,
				tc.errRegex,
				len(tc.args.data) == 0,
				packageName,
				"ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
