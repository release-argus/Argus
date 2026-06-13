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
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

func TestWebHooksDefaults_CheckValues(t *testing.T) {
	// GIVEN: WebHooksDefaults.
	tests := []struct {
		name     string
		input    *WebHooksDefaults
		errRegex string
		changed  bool
	}{
		{
			name:  "nil map",
			input: (*WebHooksDefaults)(nil),
		},
		{
			name: "valid single element map",
			input: &WebHooksDefaults{
				"a": testDefaults(true, false),
			},
		},
		{
			name: "invalid single element map",
			errRegex: test.TrimYAML(`
				^a:
					delay: .* <invalid>.*$`,
			),
			input: &WebHooksDefaults{
				"a": test.Must(t, func() (*Defaults, error) {
					return DecodeDefaults("yaml", []byte("delay: 5x"))
				}),
			},
		},
		{
			name: "valid multi element map",
			input: &WebHooksDefaults{
				"a": testDefaults(true, false),
				"b": testDefaults(false, false),
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
			input: &WebHooksDefaults{
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
			name: "custom_headers -> headers",
			input: &WebHooksDefaults{
				"a": &Defaults{
					Base: Base{
						CustomHeaders: Headers{
							{Key: "foo", Value: "bar"},
						},
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
				packageName,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}

func TestWebHooks_CheckValues(t *testing.T) {
	whCfg := plainConfig()
	// GIVEN: WebHooks.
	tests := []struct {
		name     string
		input    func() *WebHooks
		errRegex string
		changed  bool
	}{
		{
			name:  "nil map",
			input: func() *WebHooks { return nil },
		},
		{
			name: "valid single element map",
			input: func() *WebHooks {
				return &WebHooks{
					"a": testWebHook(true, false, false),
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
			input: func() *WebHooks {
				return &WebHooks{
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
			name: "valid multi element map",
			input: func() *WebHooks {
				return &WebHooks{
					"a": testWebHook(true, false, false),
					"b": testWebHook(false, false, false),
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
			input: func() *WebHooks {
				return &WebHooks{
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
			name: "custom_headers -> headers",
			input: func() *WebHooks {
				return &WebHooks{
					"a": &WebHook{
						Base: Base{
							Type:   "github",
							URL:    "example.com",
							Secret: "Argus",
							CustomHeaders: Headers{
								{Key: "foo", Value: "bar"},
							},
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
				packageName,
				tc.errRegex,
				tc.changed,
				input.CheckValues,
			)
		})
	}
}

func TestWebHook_CheckValues(t *testing.T) {
	// GIVEN: a WebHook.
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
			name: "valid WebHook",
		},
		{
			name:     "invalid delay",
			errRegex: `^delay: .* <invalid>`,
			delay:    "5x",
		},
		{
			name:      "fix int delay",
			delay:     "5",
			wantDelay: "5s",
		},
		{
			name:     "invalid type",
			errRegex: `^type: .*foo.* <invalid>.*$`,
			whType:   test.Ptr("foo"),
		},
		{
			name:       "invalid main type",
			errRegex:   `^$`, // Invalid, but caught in the Defaults CheckValues.
			whType:     test.Ptr(""),
			whMainType: "bar",
		},
		{
			name:       "mismatching type and main type",
			errRegex:   `^type: "github" <invalid>.*"gitlab".*$`,
			whType:     test.Ptr("github"),
			whMainType: "gitlab",
		},
		{
			name:     "no type",
			errRegex: `^type: <required>.*$`,
			whType:   test.Ptr(""),
		},
		{
			name:     "invalid url template",
			errRegex: `^url: .* <invalid>.*$`,
			url:      test.Ptr("{{ version }"),
		},
		{
			name:     "no url",
			errRegex: `url: <required>.*$`,
			url:      test.Ptr(""),
		},
		{
			name:     "no secret",
			errRegex: `secret: <required>.*$`,
			secret:   test.Ptr(""),
		},
		{
			name: "valid headers",
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
			name: "custom_headers -> headers",
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
			whType: test.Ptr("foo"),
			url:    test.Ptr(""),
			secret: test.Ptr(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testWebHook(true, false, false)
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

			// THEN: any error is as expected, and changed state matches expected.
			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				input.CheckValues,
			)

			prefix := fmt.Sprintf("%s\nWebHook.CheckValues()", packageName)

			// AND: the delay is fixed when expected.
			if tc.wantDelay != "" && input.Delay != tc.wantDelay {
				t.Errorf(
					"%s Delay mismatch after WebHook CheckValues()\ngot:  %q\nwant: %q",
					prefix, tc.wantDelay, input.Delay,
				)
			}

			// AND: CustomHeaders are always moved to Headers.
			if input.CustomHeaders != nil && input.Headers == nil {
				t.Errorf(
					"%s CustomHeaders should have moved to .Headers after WebHook CheckValues()\nHeaders=%v\nCustomHeaders=%v",
					prefix, input.Headers, input.CustomHeaders,
				)
			}
		})
	}
}

// Base.CheckDefaults()
func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name      string
		input     *Defaults
		wantDelay string
		errRegex  string
		changed   bool
	}{
		{
			name:  "valid WebHook",
			input: testDefaults(false, false),
		},
		{
			name:     "invalid delay",
			errRegex: `^delay: "4y" <invalid>`,
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults("yaml", []byte("delay: 4y"))
			}),
		},
		{
			name: "fix int delay",
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
			name: "valid custom headers",
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults(
					"yaml", []byte(test.TrimYAML(`
						headers:
							- key: foo
							  value: bar
						url: 'https://example.com/{{ version }'
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
			name: "custom_headers -> headers",
			input: &Defaults{
				Base: Base{
					CustomHeaders: Headers{
						{Key: "foo", Value: "bar"},
					},
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
				packageName,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}

func TestWebHooksDefaults_Print(t *testing.T) {
	testValid := testDefaults(false, false)
	testInvalid := testDefaults(true, false)
	// GIVEN: a WebHooksDefaults.
	tests := []struct {
		name             string
		webhooksDefaults *WebHooksDefaults
		want             string
	}{
		{
			name:             "nil map",
			webhooksDefaults: nil,
			want:             "",
		},
		{
			name: "single element map",
			webhooksDefaults: &WebHooksDefaults{
				"single": testValid,
			},
			want: test.TrimYAML(`
				webhook:
					single:
						type: ` + testValid.Type + `
						url: ` + testValid.URL + `
						allow_invalid_certs: ` + fmt.Sprint(*testValid.AllowInvalidCerts) + `
						secret: ` + testValid.Secret + `
						desired_status_code: ` + fmt.Sprint(*testValid.DesiredStatusCode) + `
						delay: ` + testValid.Delay + `
						max_tries: ` + fmt.Sprint(*testValid.MaxTries) + `
						silent_fails: ` + fmt.Sprint(*testValid.SilentFails),
			),
		},
		{
			name: "multiple element map",
			webhooksDefaults: &WebHooksDefaults{
				"first":  testValid,
				"second": testInvalid,
			},
			want: test.TrimYAML(`
				webhook:
					first:
						type: ` + testValid.Type + `
						url: ` + testValid.URL + `
						allow_invalid_certs: ` + fmt.Sprint(*testValid.AllowInvalidCerts) + `
						secret: ` + testValid.Secret + `
						desired_status_code: ` + fmt.Sprint(*testValid.DesiredStatusCode) + `
						delay: ` + testValid.Delay + `
						max_tries: ` + fmt.Sprint(*testValid.MaxTries) + `
						silent_fails: ` + fmt.Sprint(*testValid.SilentFails) + `
					second:
						type: ` + testInvalid.Type + `
						url: ` + testInvalid.URL + `
						allow_invalid_certs: ` + fmt.Sprint(*testInvalid.AllowInvalidCerts) + `
						secret: ` + testInvalid.Secret + `
						desired_status_code: ` + fmt.Sprint(*testInvalid.DesiredStatusCode) + `
						delay: ` + testInvalid.Delay + `
						max_tries: ` + fmt.Sprint(*testInvalid.MaxTries) + `
						silent_fails: ` + fmt.Sprint(*testInvalid.SilentFails),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout(t)

			if tc.want != "" {
				tc.want += "\n"
			}

			// WHEN: Print is called.
			tc.webhooksDefaults.Print("")

			// THEN: it prints the expected output.
			stdout := releaseStdout()
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if stdout != tc.want {
				t.Errorf(
					"%s\nWebHooksDefaults.Print() stdout mismatch\ngot:  %q\nwant: %q",
					packageName, stdout, tc.want,
				)
			}
		})
	}
}
