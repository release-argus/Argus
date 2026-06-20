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

package config

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/webhook"
)

func TestConfig_Unmarshal(t *testing.T) {
	// GIVEN: a JSON string to unmarshal into a Config.
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
			want:   "",
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
			name:     "JSON/invalid format",
			format:   "json",
			data:     `{"service": abc}`,
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid format",
			format:   "yaml",
			data:     `service: [abc`,
			errRegex: `[^\s]+ sequence end token.*`,
		},
		{
			name:   "JSON/static fields",
			format: "json",
			data: test.TrimJSON(`{
				"settings": {
						"log": {
							"level": "INFO"
						}
				},
				"defaults": {
					"service": {
						"options": {
							"interval": "10s"
						}
					}
				},
				"notify": {
					"hello": {
						"options": {
							"webhook_url": "https://example.com/webhook"
						}
					}
				},
				"webhook": {
					"hi": {
						"url": "https://example.com/webhook"
					}
				}
			}`),
			want: test.TrimYAML(`
				settings:
					log:
						level: INFO
				defaults:
					service:
						options:
							interval: 10s
				notify:
					hello:
						options:
							webhook_url: https://example.com/webhook
				webhook:
					hi:
						url: https://example.com/webhook
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/static fields",
			format: "yaml",
			data: test.TrimYAML(`
				settings:
						log:
							level: "INFO"
				defaults:
					service:
						options:
							interval: "10s"
				notify:
					hello:
						options:
							webhook_url: "https://example.com/webhook"
				webhook:
					hi:
						url: "https://example.com/webhook"
			`),
			want: test.TrimYAML(`
				settings:
					log:
						level: INFO
				defaults:
					service:
						options:
							interval: 10s
				notify:
					hello:
						options:
							webhook_url: https://example.com/webhook
				webhook:
					hi:
						url: https://example.com/webhook
			`),
			errRegex: `^$`,
		},
		{
			name:   "JSON/service subtree, ignored in Unmarshal",
			format: "json",
			data: test.TrimJSON(`{
				"service": {
					"a": {
						"name": "hi",
						"comment": "hello",
						"options": {
							"interval": "10s"
						},
						"latest_version": {
							"type": "github",
							"url": "` + test.ArgusGitHubRepo + `"
						},
						"deployed_version": {
							"type": "url",
							"url": "https://example.com"
						},
						"notify": {
							"smtp": {
								"url_fields": {
									"host": "smtp.example.com"
								},
								"params": {
									"fromaddress": "test@example.com",
									"toaddresses": "me@example.com"
								}
							}
						},
						"command": [
							"-",
							"ls"
						],
						"webhook": {
							"hi": {
								"url": "https://example.com/webhook",
								"secret": "foo"
							}
						},
						"dashboard": {
							"icon": "https://example.com/icon.png"
						}
					}
				}
			}`),
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/service subtree, ignored in Unmarshal",
			format: "yaml",
			data: test.TrimJSON(`{
				"service": {
					"a": {
						"name": "hi",
						"comment": "hello",
						"options": {
							"interval": "10s"
						},
						"latest_version": {
							"type": "github",
							"url": "` + test.ArgusGitHubRepo + `"
						},
						"deployed_version": {
							"type": "url",
							"url": "https://example.com"
						},
						"notify": {
							"smtp": {
								"url_fields": {
									"host": "smtp.example.com"
								},
								"params": {
									"fromaddress": "test@example.com",
									"toaddresses": "me@example.com"
								}
							}
						},
						"command": [
							"-",
							"ls"
						],
						"webhook": {
							"hi": {
								"url": "https://example.com/webhook",
								"secret": "foo"
							}
						},
						"dashboard": {
							"icon": "https://example.com/icon.png"
						}
					}
				}
			}`),
			want:     "{}\n",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Config.
			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Config, error) {
					var zero Config
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(v *Config) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Config",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestConfig_UnmarshalYAML(t *testing.T) {
	// GIVEN: a YAML string to unmarshal into a Config.
	tests := []struct {
		name     string
		data     string
		want     string
		errRegex string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Config, error) {
					var zero Config
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				"yaml", tc.data,
				func(v *Config) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Config",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestConfig_Decode(t *testing.T) {
	// GIVEN: a string to decode into a Config.
	tests := []struct {
		name     string
		data     string
		want     string
		errRegex string
	}{
		{
			name:     "YAML/empty",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/invalid 'service' subtree",
			data: `{"service": []}`,
			errRegex: test.TrimYAML(`
				^service:
					[^\s]+ sequence was used where mapping is expected.*
					[^\s]+.*
					\s+\^$`,
			),
		},
		{
			name: "YAML/invalid static fields",
			data: test.TrimYAML(`
				settings:
					log: INFO
			`),
			errRegex: test.TrimYAML(`
				^[^\s]+ string was used where mapping is expected
					 1 \| settings:
				\>  2 \|   log: INFO
				\s+\^$`,
			),
		},
		{
			name: "YAML/invalid defaults subtree/duplicate keys",
			data: test.TrimYAML(`
				defaults:
					service:
						options:
							interval: "10s"
					service:
						options:
							interval: "10s"
			`),
			errRegex: test.TrimYAML(`
				^[^\s]+ mapping key "service" already defined.*`,
			),
		},
		{
			name: "YAML/invalid defaults subtree/invalid data types",
			data: test.TrimYAML(`
				defaults:
					service:
						options: foo
			`),
			errRegex: test.TrimYAML(`
				^[^\s]+ string was used where mapping is expected
				[^\s]+.*
				\s+\^$`,
			),
		},
		{
			name: "YAML/static fields",
			data: test.TrimYAML(`
				settings:
					log:
						level: INFO
				defaults:
					service:
						options:
							interval: "10s"
				notify:
					hello:
						options:
							webhook_url: "https://example.com/webhook"
				webhook:
					hi:
						url: "https://example.com/webhook"
			`),
			want: test.TrimYAML(`
				settings:
					log:
						level: INFO
				defaults:
					service:
						options:
							interval: 10s
				notify:
					hello:
						options:
							webhook_url: https://example.com/webhook
				webhook:
					hi:
						url: https://example.com/webhook
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/service subtree",
			data: test.TrimYAML(`
				service:
					a:
						name: hi
						comment: hello
						options:
							interval: "10s"
						latest_version:
							type: github
							url: "` + test.ArgusGitHubRepo + `"
						deployed_version:
							type: url
							url: "https://example.com"
						notify:
							smtp:
								url_fields:
									host: smtp.example.com
								params:
									fromaddress: "test@example.com"
									toaddresses: "me@example.com"
						command:
							- - ls
						webhook:
							hi:
								url: https://example.com/webhook
								secret: foo
						dashboard:
							icon: https://example.com/icon.png
			`),
			want: test.TrimYAML(`
				service:
					a:
						name: hi
						comment: hello
						options:
							interval: 10s
						latest_version:
							type: github
							url: ` + test.ArgusGitHubRepo + `
						deployed_version:
							type: url
							url: https://example.com
						notify:
							smtp:
								url_fields:
									host: smtp.example.com
								params:
									fromaddress: test@example.com
									toaddresses: me@example.com
						command:
							- - ls
						webhook:
							hi:
								url: https://example.com/webhook
								secret: foo
						dashboard:
							icon: https://example.com/icon.png
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Config var.
			cfg := &Config{}

			// WHEN: Decode is called on it.
			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Config, error) {
					err := cfg.Decode(data)
					return cfg, err
				},
				"-", tc.data,
				func(v *Config) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Config.Decode",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestConfig_Decode__serviceExtractError(t *testing.T) {
	// GIVEN: a failing service extract.
	original := extractServiceSubtree
	extractServiceSubtree = func(format string, data []byte, key string) ([]byte, error) {
		return nil, &decode.KeyFieldError{
			Key: "service",
			Err: fmt.Errorf("ERROR MESSAGE"),
		}
	}
	t.Cleanup(func() { extractServiceSubtree = original })

	// AND: valid config.
	data := test.TrimYAML(`
		settings:
			log:
				level: INFO
		service:
			foo:
				latest_version:
					type: url
					url: https://example.com
	`)
	errRegex := test.TrimYAML(`
		^service:
			ERROR MESSAGE$`,
	)

	// WHEN: Decode is called on it.
	cfg := &Config{}
	err := cfg.Decode([]byte(data))

	prefix := fmt.Sprintf("%s\nConfig.Decode()", packageName)

	// THEN: the extract error is returned.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, errRegex,
		)
	}
}

func TestConfig_GetDefaults(t *testing.T) {
	name := "TestConfig_GetDefaults"

	// GIVEN: a Config.
	cfg := testConfig(t)

	// WHEN: GetDefaults is called.
	svcCfg, notifyCfg, whCfg := cfg.GetDefaults()

	prefix := fmt.Sprintf("%s\nConfig.GetDefaults()", packageName)

	// THEN: Notify/WebHook maps are initially empty at the 'name' key.
	fieldTests := []test.FieldAssertion{
		{Name: "Notify.Root (got)", Got: notifyCfg.Root[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Notify.Root (had)", Got: cfg.Notify[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Notify.Defaults (got)", Got: notifyCfg.Defaults[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Notify.Defaults (had)", Got: cfg.Defaults.Notify[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Notify.HardDefaults (got)", Got: notifyCfg.HardDefaults[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Notify.HardDefaults (had)", Got: cfg.HardDefaults.Notify[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "WebHook.Root (got)", Got: whCfg.Root[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "WebHook.Root (had)", Got: cfg.WebHook[name], Want: nil, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
		t.Fatal(testErr)
	}

	// THEN: the service defaults are as expected.
	fieldTests = []test.FieldAssertion{
		{Name: "Defaults.Service", Got: svcCfg.Soft, Want: &cfg.Defaults.Service, Mode: test.CompareSamePointer},
		{Name: "HardDefaults.Service", Got: svcCfg.Hard, Want: &cfg.HardDefaults.Service, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
		t.Fatal(testErr)
	}

	// AND: the notify defaults are as expected.
	rootNotify := &shoutrrr.Defaults{}
	defaultNotify := &shoutrrr.Defaults{}
	hardDefaultNotify := &shoutrrr.Defaults{}
	cfg.Notify[name] = rootNotify
	cfg.Defaults.Notify[name] = defaultNotify
	cfg.HardDefaults.Notify[name] = hardDefaultNotify
	fieldTests = []test.FieldAssertion{
		{Name: "Notify.Root", Got: notifyCfg.Root[name], Want: rootNotify, Mode: test.CompareSamePointer},
		{Name: "Notify.Defaults", Got: notifyCfg.Defaults[name], Want: defaultNotify, Mode: test.CompareSamePointer},
		{Name: "Notify.HardDefaults", Got: notifyCfg.HardDefaults[name], Want: hardDefaultNotify, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
		t.Fatal(testErr)
	}

	// AND: the webhook defaults are as expected.
	rootWebHook := &webhook.Defaults{}
	cfg.WebHook[name] = rootWebHook
	fieldTests = []test.FieldAssertion{
		{Name: "WebHook.Root", Got: whCfg.Root[name], Want: rootWebHook, Mode: test.CompareSamePointer},
		{Name: "WebHook.Defaults", Got: whCfg.Defaults, Want: &cfg.Defaults.WebHook, Mode: test.CompareSamePointer},
		{Name: "WebHook.HardDefaults", Got: whCfg.HardDefaults, Want: &cfg.HardDefaults.WebHook, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
		t.Fatal(testErr)
	}
}
