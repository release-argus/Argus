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

package webhook

import (
	"strconv"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

func TestWebhooksDefaults_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *WebhooksDefaults
		errRegex string
		changed  bool
	}{
		{
			name:     "nil map",
			errRegex: `^$`,
			input:    (*WebhooksDefaults)(nil),
		},
		{
			name:     "valid single element map",
			errRegex: `^$`,
			input: &WebhooksDefaults{
				"a": testDefaults(t, true),
			},
		},
		{
			name: "invalid single element map",
			errRegex: test.TrimYAML(`
				^a:
					delay: .* <invalid>.*$`,
			),
			input: &WebhooksDefaults{
				"a": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults("yaml", []byte("delay: 5x"))
				}),
			},
		},
		{
			name:     "valid multi element map",
			errRegex: `^$`,
			input: &WebhooksDefaults{
				"a": testDefaults(t, true),
				"b": testDefaults(t, false),
			},
		},
		{
			name: "invalid multi element map",
			errRegex: test.TrimYAML(`
				^a:
					delay: "[^"]+" <invalid>.*
				b:
					type: "[^"]+" <invalid>.*
					url: "[^"]+" <invalid>.*
					delay: "[^"]+" <invalid>.*$`,
			),
			input: &WebhooksDefaults{
				"a": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults("yaml", []byte("delay: 5x"))
				}),
				"b": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							delay: 5x
							type: foo
							url: 'https://example.com/{{ version }'
						`)),
					)
				}),
			},
		},
		{
			name:     "custom_headers -> headers",
			errRegex: `^$`,
			input: &WebhooksDefaults{
				"a": &Defaults{
					CustomHeaders: Headers{
						{Key: "foo", Value: "bar"},
					},
				},
			},
			changed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}

func TestWebhooks_CheckValues(t *testing.T) {
	whCfg := plainConfig(t)

	tests := []struct {
		name     string
		input    func() *Webhooks
		errRegex string
		changed  bool
	}{
		{
			name:     "nil map",
			errRegex: `^$`,
			input:    func() *Webhooks { return nil },
		},
		{
			name:     "valid single element map",
			errRegex: `^$`,
			input: func() *Webhooks {
				return &Webhooks{
					"a": testWebhook(true, false),
				}
			},
		},
		{
			name: "invalid single element map",
			errRegex: test.TrimYAML(`
				^a:
					delay: "5x" <invalid>.*
					url: <required>.*
					secret: <required>.*$`,
			),
			input: func() *Webhooks {
				return &Webhooks{
					"a": New(
						nil, nil,
						"5x",
						nil, nil,
						"a",
						nil, Notifiers{}, nil,
						"",
						nil,
						"", "",
						nil, nil, nil,
					),
				}
			},
		},
		{
			name:     "valid multi element map",
			errRegex: `^$`,
			input: func() *Webhooks {
				return &Webhooks{
					"a": testWebhook(true, false),
					"b": testWebhook(false, false),
				}
			},
		},
		{
			name: "invalid multi element map",
			errRegex: test.TrimYAML(`
				^a:
					delay: "5x" <invalid>.*
					url: <required>.*
					secret: <required>.*
				b:
					type: "foo" <invalid>.*
					url: <required>.*
					secret: <required>.*$`,
			),
			input: func() *Webhooks {
				return &Webhooks{
					"a": New(
						nil, nil,
						"5x",
						nil, nil,
						"a",
						nil, Notifiers{}, nil,
						"",
						nil,
						"", "",
						nil, nil, nil,
					),
					"b": New(
						nil, nil,
						"",
						nil, nil,
						"b",
						nil, Notifiers{}, nil,
						"",
						nil,
						"foo", "",
						&Defaults{}, &Defaults{}, &Defaults{},
					),
				}
			},
		},
		{
			name:     "custom_headers -> headers",
			errRegex: `^$`,
			input: func() *Webhooks {
				return &Webhooks{
					"a": &Webhook{
						Type:   "github",
						URL:    "example.com",
						Secret: "Argus",
						CustomHeaders: Headers{
							{Key: "foo", Value: "bar"},
						},
					},
				}
			},
			changed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := tc.input()
			if input != nil {
				svcStatus := status.Status{}
				svcStatus.Init(
					0, 0, len(*input),
					status.ServiceInfo{
						ID: tc.name,
					},
					&dashboard.Options{},
				)
				input.Init(
					&svcStatus,
					whCfg,
					nil, nil,
				)
			}

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				input.CheckValues,
			)
		})
	}
}

func TestWebhook_CheckValues(t *testing.T) {
	tests := []struct {
		name             string
		delay, wantDelay string
		whType           *string
		whMainType       string
		url, secret      *string
		customHeaders    Headers
		headers          Headers
		errRegex         string
		changed          bool
	}{
		{
			name:     "valid Webhook",
			errRegex: `^$`,
		},
		{
			name:     "invalid delay",
			errRegex: `^delay: .* <invalid>`,
			delay:    "5x",
		},
		{
			name:      "fix int delay",
			errRegex:  `^$`,
			delay:     "5",
			wantDelay: "5s",
		},
		{
			name:     "invalid type",
			errRegex: `^type: .*foo.* <invalid>.*$`,
			whType:   new("foo"),
		},
		{
			name:       "invalid main type",
			errRegex:   `^$`, // Invalid, but caught in the Defaults CheckValues.
			whType:     new(""),
			whMainType: "bar",
		},
		{
			name:       "mismatching type and main type",
			errRegex:   `^type: "github" <invalid>.*"gitlab".*$`,
			whType:     new("github"),
			whMainType: "gitlab",
		},
		{
			name:     "no type",
			errRegex: `^type: <required>.*$`,
			whType:   new(""),
		},
		{
			name:     "invalid url template",
			errRegex: `^url: .* <invalid>.*$`,
			url:      new("{{ version }"),
		},
		{
			name:     "no url",
			errRegex: `url: <required>.*$`,
			url:      new(""),
		},
		{
			name:     "no secret",
			errRegex: `secret: <required>.*$`,
			secret:   new(""),
		},
		{
			name:     "valid headers",
			errRegex: `^$`,
			headers: Headers{
				{Key: "foo", Value: "bar"},
			},
			changed: false,
		},
		{
			name: "invalid headers",
			errRegex: test.TrimYAML(`
				^headers:
					bar: "[^"]+" <invalid>.*$`,
			),
			headers: Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bar", Value: "{{ version }"},
			},
		},
		{
			name:     "custom_headers -> headers",
			errRegex: `^$`,
			customHeaders: Headers{
				{Key: "foo", Value: "bar"},
			},
			changed: true,
		},
		{
			name: "all decode",
			errRegex: test.TrimYAML(`
				^type: "[^"]+" <invalid>.*
				delay: "[^"]+" <invalid>.*
				url: <required>.*
				secret: <required>.*$`,
			),
			delay:  "5x",
			whType: new("foo"),
			url:    new(""),
			secret: new(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testWebhook(true, false)
			if tc.whMainType != "" {
				input.Main.Type = tc.whMainType
			}
			if tc.delay != "" {
				input.Delay = tc.delay
			}
			if tc.whType != nil {
				input.Type = *tc.whType
			}
			if tc.url != nil {
				input.URL = *tc.url
			}
			if tc.secret != nil {
				input.Secret = *tc.secret
			}
			input.CustomHeaders = tc.customHeaders
			input.Headers = tc.headers

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				input.CheckValues,
			)

			if tc.wantDelay != "" && input.Delay != tc.wantDelay {
				t.Errorf("Webhook.Delay mismatch\ngot:  %q\nwant: %q", input.Delay, tc.wantDelay)
			}

			if tc.customHeaders != nil && tc.headers == nil {
				if testErr := test.AssertSlicesEqual(
					t,
					input.Headers,
					tc.customHeaders,
					"Webhook.Headers",
				); testErr != nil {
					t.Error(testErr)
				}
				if input.CustomHeaders != nil {
					keys := test.AttributesSorted(input.CustomHeaders, func(h Header) string { return h.Key })
					t.Errorf(
						"Webhook.CustomHeaders mismatch\ngot:  keys=%q\nwant: nil (moved)",
						keys,
					)
				}
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	tests := []struct {
		name      string
		input     *Defaults
		wantDelay string
		errRegex  string
		changed   bool
	}{
		{
			name:     "valid Webhook",
			errRegex: `^$`,
			input:    testDefaults(t, false),
		},
		{
			name:     "invalid delay",
			errRegex: `^delay: "4y" <invalid>`,
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults("yaml", []byte("delay: 4y"))
			}),
		},
		{
			name:     "fix int delay",
			errRegex: `^$`,
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults("yaml", []byte("delay: 3"))
			}),
			wantDelay: "3s",
		},
		{
			name:     "invalid type",
			errRegex: `^type: "foo" <invalid>`,
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults("yaml", []byte("type: foo"))
			}),
		},
		{
			name:     "invalid url template",
			errRegex: `url: ".+" <invalid>`,
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults("yaml", []byte("url: 'https://example.com/{{ version }'"))
			}),
		},
		{
			name:     "valid custom headers",
			errRegex: `^$`,
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						headers:
							- key: foo
							  value: bar
					`)),
				)
			}),
		},
		{
			name: "invalid custom headers",
			errRegex: test.TrimYAML(`
				^headers:
					bar: "[^"]+" <invalid>`,
			),
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						headers:
							- key: foo
							  value: bar
							- key: bar
							  value: '{{ version }'
					`)),
				)
			}),
		},
		{
			name:     "custom_headers -> headers",
			errRegex: `^$`,
			input: &Defaults{
				CustomHeaders: Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			changed: true,
		},
		{
			name: "all decode",
			errRegex: test.TrimYAML(`
				^type: "[^"]+" <invalid>.*
				url: "[^"]+" <invalid>.*
				headers:
					bar: "[^"]+" <invalid>.*
				delay: "[^"]+" <invalid>.*$`,
			),
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						headers:
							- key: foo
							  value: bar
							- key: bar
							  value: '{{ version }'
						delay: 5x
						type: shazam
						url: 'https://example.com/{{ version }'
					`)),
				)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)

			if tc.wantDelay != "" && tc.input.Delay != tc.wantDelay {
				t.Errorf(
					"Defaults.Delay mismatch\ngot:  %q\nwant: %q",
					tc.input.Delay, tc.wantDelay,
				)
			}
		})
	}
}

func TestWebhooksDefaults_Print(t *testing.T) {
	testValid := testDefaults(t, false)
	testFailing := testDefaults(t, true)

	tests := []struct {
		name             string
		webhooksDefaults *WebhooksDefaults
		want             string
	}{
		{
			name:             "nil map",
			webhooksDefaults: nil,
			want:             "",
		},
		{
			name: "single element map",
			webhooksDefaults: &WebhooksDefaults{
				"single": testValid,
			},
			want: test.TrimYAML(`
				webhook:
					single:
						type: ` + testValid.Type + `
						url: ` + testValid.URL + `
						allow_invalid_certs: ` + strconv.FormatBool(*testValid.AllowInvalidCerts) + `
						secret: ` + testValid.Secret + `
						desired_status_code: ` + strconv.FormatUint(uint64(*testValid.DesiredStatusCode), 10) + `
						delay: ` + testValid.Delay + `
						max_tries: ` + strconv.FormatUint(uint64(*testValid.MaxTries), 10) + `
						silent_fails: ` + strconv.FormatBool(*testValid.SilentFails),
			),
		},
		{
			name: "multiple element map",
			webhooksDefaults: &WebhooksDefaults{
				"first":  testValid,
				"second": testFailing,
			},
			want: test.TrimYAML(`
				webhook:
					first:
						type: ` + testValid.Type + `
						url: ` + testValid.URL + `
						allow_invalid_certs: ` + strconv.FormatBool(*testValid.AllowInvalidCerts) + `
						secret: ` + testValid.Secret + `
						desired_status_code: ` + strconv.FormatUint(uint64(*testValid.DesiredStatusCode), 10) + `
						delay: ` + testValid.Delay + `
						max_tries: ` + strconv.FormatUint(uint64(*testValid.MaxTries), 10) + `
						silent_fails: ` + strconv.FormatBool(*testValid.SilentFails) + `
					second:
						type: ` + testFailing.Type + `
						url: ` + testFailing.URL + `
						allow_invalid_certs: ` + strconv.FormatBool(*testFailing.AllowInvalidCerts) + `
						secret: ` + testFailing.Secret + `
						desired_status_code: ` + strconv.FormatUint(uint64(*testFailing.DesiredStatusCode), 10) + `
						delay: ` + testFailing.Delay + `
						max_tries: ` + strconv.FormatUint(uint64(*testFailing.MaxTries), 10) + `
						silent_fails: ` + strconv.FormatBool(*testFailing.SilentFails),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureStdout(t)

			if tc.want != "" {
				tc.want += "\n"
			}

			tc.webhooksDefaults.Print("")

			stdout := releaseStdout()
			if stdout != tc.want {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.want)
			}
		})
	}
}
