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
	"errors"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/webhook"
)

func TestConfig_Unmarshal(t *testing.T) {
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
			data: test.TrimYAML(`
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
							- "-"
							- ls
						webhook:
							hi:
								url: https://example.com/webhook
								secret: foo
						dashboard:
							icon: https://example.com/icon.png
			`),
			want:     "{}\n",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Config, error) {
					var cfg Config
					err := decode.Unmarshal(format, data, &cfg)
					return &cfg, err
				},
				tc.format, tc.data,
				func(v *Config) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				"Config.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestConfig_Decode(t *testing.T) {
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

			cfg := &Config{}

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
				"Config.Decode",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestConfig_Decode__serviceExtractError(t *testing.T) {
	original := extractServiceSubtree
	extractServiceSubtree = func(format string, data []byte, key string) ([]byte, error) {
		return nil, &decode.ErrKeyField{
			Key: "service",
			Err: errors.New("ERROR MESSAGE"),
		}
	}
	t.Cleanup(func() { extractServiceSubtree = original })

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

	cfg := &Config{}
	err := cfg.Decode([]byte(data))
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestConfig_GetDefaults(t *testing.T) {
	name := t.Name()

	cfg := testConfig(t)

	svcCfg, notifyCfg, whCfg := cfg.GetDefaults()

	fieldTests := []test.FieldAssertion{
		{Name: "shoutrrr.Config.Root[name]", Got: notifyCfg.Root[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Config.Notify[name]", Got: cfg.Notify[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "shoutrrr.Config.Defaults[name]", Got: notifyCfg.Defaults[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Config.Defaults.Notify[name]", Got: cfg.Defaults.Notify[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "shoutrrr.Config.HardDefaults[name]", Got: notifyCfg.HardDefaults[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Config.HardDefaults.Notify[name]", Got: cfg.HardDefaults.Notify[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "webhook.Config.Root[name]", Got: whCfg.Root[name], Want: nil, Mode: test.CompareSamePointer},
		{Name: "Config.Webhook[name]", Got: cfg.Webhook[name], Want: nil, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, ""); testErr != nil {
		t.Fatal(testErr)
	}

	fieldTests = []test.FieldAssertion{
		{Name: "service.DefaultsConfig.Soft", Got: svcCfg.Soft, Want: &cfg.Defaults.Service, Mode: test.CompareSamePointer},
		{Name: "service.DefaultsConfig.Hard", Got: svcCfg.Hard, Want: &cfg.HardDefaults.Service, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, ""); testErr != nil {
		t.Fatal(testErr)
	}

	rootNotify := &shoutrrr.Defaults{}
	defaultNotify := &shoutrrr.Defaults{}
	hardDefaultNotify := &shoutrrr.Defaults{}
	cfg.Notify[name] = rootNotify
	cfg.Defaults.Notify[name] = defaultNotify
	cfg.HardDefaults.Notify[name] = hardDefaultNotify
	fieldTests = []test.FieldAssertion{
		{Name: "shoutrrr.Config.Root[name]", Got: notifyCfg.Root[name], Want: rootNotify, Mode: test.CompareSamePointer},
		{Name: "shoutrrr.Config.Defaults[name]", Got: notifyCfg.Defaults[name], Want: defaultNotify, Mode: test.CompareSamePointer},
		{Name: "shoutrrr.Config.HardDefaults[name]", Got: notifyCfg.HardDefaults[name], Want: hardDefaultNotify, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, ""); testErr != nil {
		t.Fatal(testErr)
	}

	rootWebhook := &webhook.Defaults{}
	cfg.Webhook[name] = rootWebhook
	fieldTests = []test.FieldAssertion{
		{Name: "webhook.Config.Root[name]", Got: whCfg.Root[name], Want: rootWebhook, Mode: test.CompareSamePointer},
		{Name: "webhook.Config.Defaults", Got: whCfg.Defaults, Want: &cfg.Defaults.Webhook, Mode: test.CompareSamePointer},
		{Name: "webhook.Config.HardDefaults", Got: whCfg.HardDefaults, Want: &cfg.HardDefaults.Webhook, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, ""); testErr != nil {
		t.Fatal(testErr)
	}
}
