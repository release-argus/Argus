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
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestDataSettings_IsZero(t *testing.T) {
	// GIVEN: a DataSettings struct.
	tests := []struct {
		name string
		data DataSettings
		want bool
	}{
		{
			name: "empty",
			data: DataSettings{},
			want: true,
		},
		{
			name: "non-empty DatabaseFile",
			data: DataSettings{
				DatabaseFile: "db.sqlite",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.data.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nDataSettings.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLogSettings_IsZero(t *testing.T) {
	// GIVEN: a LogSettings struct.
	tests := []struct {
		name string
		data LogSettings
		want bool
	}{
		{
			name: "empty",
			data: LogSettings{},
			want: true,
		},
		{
			name: "non-empty Timestamps",
			data: LogSettings{
				Timestamps: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "non-empty Level",
			data: LogSettings{
				Level: "INFO",
			},
			want: false,
		},
		{
			name: "filled",
			data: LogSettings{
				Timestamps: test.Ptr(true),
				Level:      "INFO",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.data.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLogSettings.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebSettings_IsZero(t *testing.T) {
	// GIVEN: a WebSettings struct.
	tests := []struct {
		name string
		data WebSettings
		want bool
	}{
		{
			name: "empty",
			data: WebSettings{},
			want: true,
		},
		{
			name: "non-empty ListenHost",
			data: WebSettings{
				ListenHost: "0.0.0.0",
			},
			want: false,
		},
		{
			name: "non-empty ListenPort",
			data: WebSettings{
				ListenPort: "8080",
			},
			want: false,
		},
		{
			name: "non-empty RoutePrefix",
			data: WebSettings{
				RoutePrefix: "/test",
			},
			want: false,
		},
		{
			name: "non-empty CertFile",
			data: WebSettings{
				CertFile: "cert.pem",
			},
			want: false,
		},
		{
			name: "non-empty KeyFile",
			data: WebSettings{
				KeyFile: "privkey.pem",
			},
			want: false,
		},
		{
			name: "empty BasicAuth",
			data: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{},
			},
			want: false,
		},
		{
			name: "non-empty BasicAuth",
			data: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: "pass",
				},
			},
			want: false,
		},
		{
			name: "non-empty DisabledRouted",
			data: WebSettings{
				DisabledRoutes: []string{"route1", "route2"},
			},
			want: false,
		},
		{
			name: "empty Favicon",
			data: WebSettings{
				Favicon: &FaviconSettings{},
			},
			want: false,
		},
		{
			name: "non-empty favicon",
			data: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "favicon.svg",
				},
			},
			want: false,
		},
		{
			name: "filled",
			data: WebSettings{
				ListenHost: "0.0.0.0",
				ListenPort: "8080",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.data.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nWebSettings.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebSettings_String(t *testing.T) {
	// GIVEN: a WebSettings struct.
	tests := []struct {
		name        string
		webSettings *WebSettings
		prefix      string
		want        string
	}{
		{
			name:        "nil webSettings",
			webSettings: nil,
			prefix:      "",
			want:        "",
		},
		{
			name:        "empty webSettings",
			webSettings: &WebSettings{},
			prefix:      "",
			want:        "{}\n",
		},
		{
			name: "webSettings with values",
			webSettings: &WebSettings{
				ListenHost: "0.0.0.0",
				ListenPort: "8080",
			},
			prefix: "",
			want: test.TrimYAML(`
				listen_host: 0.0.0.0
				listen_port: '8080'
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.webSettings.String,
				tc.want,
			)
		})
	}
}

func TestWebSettings_CheckValues(t *testing.T) {
	// GIVEN: a WebSettings struct with some values set.
	tests := []struct {
		name                               string
		env                                map[string]string
		input                              *WebSettings
		want                               string
		wantUsernameHash, wantPasswordHash string
		ok                                 bool
		errRegex                           string
	}{
		{
			name: "BasicAuth - empty",
			input: &WebSettings{
				BasicAuth: &WebSettingsBasicAuth{},
			},
			want: "{}\n",
			ok:   true,
		},
		{
			name: "BasicAuth - str Username and Password already hashed",
			input: &WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: util.FmtHash(util.GetHash("pass")),
				},
			},
			want: test.TrimYAML(`
				basic_auth:
					username: user
					password: ` + util.FmtHash(util.GetHash("pass")) + `
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
			ok:               true,
		},
		{
			name: "BasicAuth - hashed Username and str Password",
			input: &WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: "pass",
				},
			},
			want: test.TrimYAML(`
				basic_auth:
					username: user
					password: ` + util.FmtHash(util.GetHash("pass")) + `
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
			ok:               true,
		},
		{
			name: "BasicAuth - Username and password from env vars",
			env: map[string]string{
				"TEST_WEB_SETTINGS__CHECK_VALUES__ONE": "user",
				"TEST_WEB_SETTINGS__CHECK_VALUES__TWO": "pass",
			},
			input: &WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "${TEST_WEB_SETTINGS__CHECK_VALUES__ONE}",
					Password: "${TEST_WEB_SETTINGS__CHECK_VALUES__TWO}",
				},
			},
			want: test.TrimYAML(`
				basic_auth:
					username: ${TEST_WEB_SETTINGS__CHECK_VALUES__ONE}
					password: ${TEST_WEB_SETTINGS__CHECK_VALUES__TWO}
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			ok:               true,
		},
		{
			name: "BasicAuth - Username and password from env vars partial",
			env: map[string]string{
				"TEST_WEB_SETTINGS__CHECK_VALUES__THREE": "er",
				"TEST_WEB_SETTINGS__CHECK_VALUES__FOUR":  "ss",
			},
			input: &WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "us${TEST_WEB_SETTINGS__CHECK_VALUES__THREE}",
					Password: "pa${TEST_WEB_SETTINGS__CHECK_VALUES__FOUR}",
				},
			},
			want: test.TrimYAML(`
				basic_auth:
					username: us${TEST_WEB_SETTINGS__CHECK_VALUES__THREE}
					password: pa${TEST_WEB_SETTINGS__CHECK_VALUES__FOUR}
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
			ok:               true,
		},
		{
			name: "Favicon - empty",
			input: &WebSettings{
				Favicon: &FaviconSettings{},
			},
			want: "{}\n",
			ok:   true,
		},
		{
			name: "Favicon - SVG",
			input: &WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg",
				},
			},
			want: test.TrimYAML(`
				favicon:
					svg: https://example.com/favicon.svg
			`),
			ok: true,
		},
		{
			name: "Favicon - PNG",
			input: &WebSettings{
				Favicon: &FaviconSettings{
					PNG: "https://example.com/favicon.png",
				},
			},
			want: test.TrimYAML(`
				favicon:
					png: https://example.com/favicon.png
			`),
			ok: true,
		},
		{
			name: "Favicon - Full",
			input: &WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg",
					PNG: "https://example.com/favicon.png",
				},
			},
			want: test.TrimYAML(`
				favicon:
					svg: https://example.com/favicon.svg
					png: https://example.com/favicon.png
			`),
			ok: true,
		},
		{
			name: "Web.CertFile - not found",
			input: &WebSettings{
				CertFile: "cert.pem",
			},
			want:     "cert_file: cert.pem\n",
			ok:       false,
			errRegex: `^cert_file: .*no such file.*$`,
		},
		{
			name: "Web.KeyFile - not found",
			input: &WebSettings{
				KeyFile: "privkey.pem",
			},
			want:     "pkey_file: privkey.pem\n",
			ok:       false,
			errRegex: `^pkey_file: .*no such file.*$`,
		},
		{
			name: "Web.CertFile + Web.KeyFile - both not found",
			input: &WebSettings{
				CertFile: "cert.pem",
				KeyFile:  "privkey.pem",
			},
			want: test.TrimYAML(`
				cert_file: cert.pem
				pkey_file: privkey.pem
			`),
			ok: false,
			errRegex: test.TrimYAML(`
				^cert_file: .*no such file.*
				pkey_file: .*no such file.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)

			prefix := fmt.Sprintf("%s\nWebSettings.CheckValues()", packageName)

			// THEN: the Settings are converted/removed where necessary.
			gotStr := tc.input.String("")
			if gotStr != tc.want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, tc.want,
				)
			}
			if tc.wantUsernameHash != "" {
				got := util.FmtHash(tc.input.BasicAuth.UsernameHash)
				if got != tc.wantUsernameHash {
					t.Errorf(
						"%s Username hash mismatch\ngot:  %q\nwant: %q",
						prefix, got, tc.wantUsernameHash,
					)
				}
			}
		})
	}
}

// WebSettingsBasicAuth.

func TestWebSettingsBasicAuth_String(t *testing.T) {
	// GIVEN: a WebSettingsBasicAuth struct.
	tests := []struct {
		name   string
		auth   *WebSettingsBasicAuth
		prefix string
		want   string
	}{
		{
			name:   "nil auth",
			auth:   nil,
			prefix: "",
			want:   "",
		},
		{
			name:   "empty auth",
			auth:   &WebSettingsBasicAuth{},
			prefix: "",
			want:   "{}\n",
		},
		{
			name: "auth with values",
			auth: &WebSettingsBasicAuth{
				Username: "user",
				Password: "pass",
			},
			prefix: "",
			want: test.TrimYAML(`
				username: user
				password: pass
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.auth.String,
				tc.want,
			)
		})
	}
}

func TestWebSettingsBasicAuth_CheckValues(t *testing.T) {
	// GIVEN: a WebSettingsBasicAuth struct with some values set.
	tests := []struct {
		name                               string
		env                                map[string]string
		input                              func() WebSettingsBasicAuth
		want                               WebSettingsBasicAuth
		wantUsernameHash, wantPasswordHash string
	}{
		{
			name: "str Username",
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "test",
				}
			},
			want: WebSettingsBasicAuth{
				Username: "test",
				Password: util.FmtHash(util.GetHash("")),
			},
		},
		{
			name: "str Web.BasicAuth.Password",
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Password: "just a password here",
				}
			},
			want: WebSettingsBasicAuth{
				Username: "",
				Password: util.FmtHash(util.GetHash("just a password here")),
			},
		},
		{
			name: "str Web.BasicAuth.Username and str Web.BasicAuth.Password",
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "user",
					Password: "pass",
				}
			},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass")),
			},
		},
		{
			name: "str env Web.BasicAuth.Username and str env Web.BasicAuth.Password",
			env: map[string]string{
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__ONE": "user",
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__TWO": "pass",
			},
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__ONE}",
					Password: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__TWO}",
				}
			},
			want: WebSettingsBasicAuth{
				Username: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__ONE}",
				Password: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__TWO}",
			},
		},
		{
			name: "str env partial Web.BasicAuth.Username and str env partial Web.BasicAuth.Password",
			env: map[string]string{
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__THREE": "user",
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__FOUR":  "pass",
			},
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__THREE}",
					Password: "b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__FOUR}"}
			},
			want: WebSettingsBasicAuth{
				Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__THREE}",
				Password: "b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__FOUR}",
			},
		},
		{
			name: "str env undefined Web.BasicAuth.Username and str env undefined Web.BasicAuth.Password",
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}",
					Password: "b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}"}
			},
			want: WebSettingsBasicAuth{
				Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}",
				Password: util.FmtHash(util.GetHash("b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}")),
			},
			wantUsernameHash: util.FmtHash(util.GetHash("a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}")),
			wantPasswordHash: util.FmtHash(util.GetHash("b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}")),
		},
		{
			name: "str Web.BasicAuth.Username and Web.BasicAuth.Password already hashed",
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "user",
					Password: util.FmtHash(util.GetHash("pass")),
				}
			},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass")),
			},
		},
		{
			name: "hashed Web.BasicAuth.Username and str Web.BasicAuth.Password",
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "user",
					Password: "pass",
				}
			},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass")),
			},
		},
		{
			name: "hashed Web.BasicAuth.Username and hashed Web.BasicAuth.Password",
			input: func() WebSettingsBasicAuth {
				return WebSettingsBasicAuth{
					Username: "user",
					Password: util.FmtHash(util.GetHash("pass")),
				}
			},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass")),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)

			had := tc.input()

			// WHEN: CheckValues is called on it.
			had.CheckValues()

			prefix := fmt.Sprintf("%s\nWebSettings BasicAuth", packageName)

			// THEN: the Settings are converted/removed where necessary.
			gotStr := had.String("")
			wantStr := tc.want.String("")
			if gotStr != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %v\nwant: %v",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the UsernameHash is calculated correctly.
			want := util.FmtHash(
				util.GetHash(
					util.EvalEnvVars(tc.want.Username),
				),
			)
			if tc.wantUsernameHash != "" {
				want = tc.wantUsernameHash
			}
			got := util.FmtHash(had.UsernameHash)
			if got != want {
				t.Errorf(
					"%s Username Hash mismatch\ngot:  %s\nwant: %s",
					prefix, got, want,
				)
			}

			// AND: the PasswordHash is calculated correctly.
			want = util.FmtHash(
				util.GetHash(
					util.EvalEnvVars(tc.want.Password),
				),
			)
			if tc.wantPasswordHash != "" {
				want = tc.wantPasswordHash
			}
			got = util.FmtHash(had.PasswordHash)
			if got != want {
				t.Errorf(
					"%s Password Hash mismatch\ngot:  %s\nwant: %s",
					prefix, got, want,
				)
			}
		})
	}
}

// Settings.

func TestSettings_IsZero(t *testing.T) {
	// GIVEN: a Settings struct.
	tests := []struct {
		name string
		data Settings
		want bool
	}{
		{
			name: "empty",
			data: Settings{},
			want: true,
		},
		{
			name: "non-empty Log",
			data: Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Timestamps: test.Ptr(true),
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty Data",
			data: Settings{
				SettingsBase: SettingsBase{
					Data: DataSettings{
						DatabaseFile: "db.sqlite",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty Web",
			data: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						ListenHost: "0.0.0.0",
					},
				},
			},
			want: false,
		},
		{
			name: "filled",
			data: Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Timestamps: test.Ptr(true),
					},
					Data: DataSettings{
						DatabaseFile: "db.sqlite",
					},
					Web: WebSettings{
						ListenHost: "0.0.0.0",
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.data.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nSettings.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// Settings.

func TestSettings_CheckValues(t *testing.T) {
	// GIVEN: a Settings struct with some values set.
	tests := []struct {
		name                               string
		env                                map[string]string
		input                              *Settings
		want                               string
		wantUsernameHash, wantPasswordHash string
		ok                                 bool
		errRegex                           string
	}{
		{
			name: "BasicAuth - empty",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{},
					},
				},
			},
			want: "{}\n",
			ok:   true,
		},
		{
			name: "BasicAuth - hashed Username and str env Password",
			env: map[string]string{
				"TEST_SETTINGS_BASE__CHECK_VALUES__ONE": "ass",
			},
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: util.FmtHash(util.GetHash("user")),
							Password: "p${TEST_SETTINGS_BASE__CHECK_VALUES__ONE}",
						},
					},
				},
			},
			want: test.TrimYAML(`
				web:
					basic_auth:
						username: ` + util.FmtHash(util.GetHash("user")) + `
						password: p${TEST_SETTINGS_BASE__CHECK_VALUES__ONE}
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
			ok:               true,
		},
		{
			name: "Route prefix - empty",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "",
					},
				},
			},
			want: "{}\n",
			ok:   true,
		},
		{
			name: "Route prefix - no leading /",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "test",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			ok: true,
		},
		{
			name: "Route prefix - leading /",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			ok: true,
		},
		{
			name: "Route prefix - multiple leading /",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "///test",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			ok: true,
		},
		{
			name: "Route prefix - trailing /",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test/",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			ok: true,
		},
		{
			name: "Route prefix - multiple trailing /",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test///",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			ok: true,
		},
		{
			name: "Route prefix - only a /",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /
			`),
			ok: true,
		},
		{
			name: "Route prefix - only multiple /",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "///",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /
			`),
			ok: true,
		},
		{
			name: "Favicon - empty",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						Favicon: &FaviconSettings{},
					},
				},
			},
			want: "{}\n",
			ok:   true,
		},
		{
			name: "Favicon - full",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						Favicon: &FaviconSettings{
							SVG: "https://example.com/favicon.svg",
							PNG: "https://example.com/favicon.png",
						},
					},
				},
			},
			want: test.TrimYAML(`
				web:
					favicon:
						svg: https://example.com/favicon.svg
						png: https://example.com/favicon.png
			`),
			ok: true,
		},
		{
			name: "Web.CertFile - not found",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.pem",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					cert_file: cert.pem
			`),
			errRegex: test.TrimYAML(`
				^web:
					cert_file: .*no such file.*$`,
			),
			ok: false,
		},
		{
			name: "Web.KeyFile - not found",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						KeyFile: "privkey.pem",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					pkey_file: privkey.pem
			`),
			errRegex: test.TrimYAML(`
				^web:
					pkey_file: .*no such file.*$`,
			),
			ok: false,
		},
		{
			name: "Web.CertFile + Web.KeyFile - both not found",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.pem",
						KeyFile:  "privkey.pem",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					cert_file: cert.pem
					pkey_file: privkey.pem
			`),
			errRegex: test.TrimYAML(`
				^web:
					cert_file: .*no such file.*
					pkey_file: .*no such file.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)

			prefix := fmt.Sprintf("%s\nSettings.CheckValues()", packageName)

			// THEN: The Settings returned stringifies as expected.
			if got := tc.input.String(""); got != tc.want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the BasicAuth username and password are hashed (if they exist).
			if tc.input.Web.BasicAuth != nil {
				wantUsernameHash := util.FmtHash(util.GetHash(tc.input.Web.BasicAuth.Username))
				if tc.wantUsernameHash != "" {
					wantUsernameHash = tc.wantUsernameHash
				}
				if got := util.FmtHash(tc.input.Web.BasicAuth.UsernameHash); got != wantUsernameHash {
					t.Errorf(
						"%s Username hash mismatch\ngot:  %q\nwant: %q",
						prefix, got, wantUsernameHash,
					)
				}
				wantPasswordHash := util.FmtHash(util.GetHash(tc.input.Web.BasicAuth.Password))
				if tc.wantPasswordHash != "" {
					wantPasswordHash = tc.wantPasswordHash
				}
				if got := util.FmtHash(tc.input.Web.BasicAuth.PasswordHash); got != wantPasswordHash {
					t.Errorf(
						"%s Password hash mismatch\ngot:  %q\nwant: %q",
						prefix, got, wantPasswordHash,
					)
				}
			}
		})
	}
}

func TestSettings_MapEnvToStruct(t *testing.T) {
	// Unset ARGUS_LOG_LEVEL.
	argusLogLevelEnvKey := "ARGUS_LOG_LEVEL"
	argusLogLevel := os.Getenv(argusLogLevelEnvKey)
	os.Setenv(argusLogLevelEnvKey, "")
	t.Cleanup(func() {
		os.Setenv(argusLogLevelEnvKey, argusLogLevel)
	})

	// GIVEN: vars set for Settings vars.
	tests := []struct {
		name                  string
		env                   map[string]string
		want                  *Settings
		stdoutRegex, errRegex string
		ok                    bool
	}{
		{
			name: "empty vars ignored",
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "",
			},
			want: &Settings{},
			ok:   true,
		},
		{
			name: "log.level",
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "ERROR",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Level: "ERROR",
					},
				},
			},
			ok: true,
		},
		{
			name: "log.timestamps",
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "true",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Timestamps: test.Ptr(true),
					},
				},
			},
			ok: true,
		},
		{
			name: "log.timestamps - invalid, not a bool",
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "abc",
			},
			want: &Settings{},
			ok:   false,
			errRegex: test.TrimYAML(`
				one or more.* environment variables.*
					ARGUS_LOG_TIMESTAMPS: .*$`,
			),
		},
		{
			name: "web.listen-host",
			env: map[string]string{
				"ARGUS_WEB_LISTEN_HOST": "test",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						ListenHost: "test",
					},
				},
			},
			ok: true,
		},
		{
			name: "web.listen-port",
			env: map[string]string{
				"ARGUS_WEB_LISTEN_PORT": "123",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						ListenPort: "123",
					},
				},
			},
			ok: true,
		},
		{
			name: "web.cert-file",
			env: map[string]string{
				"ARGUS_WEB_CERT_FILE": "cert.test",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.test",
					},
				},
			},
			ok: false,
			errRegex: test.TrimYAML(`
				^hard_defaults:
					settings:
						web:
							cert_file: .*no such file.*$`,
			),
		},
		{
			name: "web.pkey-file",
			env: map[string]string{
				"ARGUS_WEB_PKEY_FILE": "pkey.test",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						KeyFile: "pkey.test",
					},
				},
			},
			ok: false,
			errRegex: test.TrimYAML(`
				^hard_defaults:
					settings:
						web:
							pkey_file: .*no such file.*$`,
			),
		},
		{
			name: "web.route-prefix",
			env: map[string]string{
				"ARGUS_WEB_ROUTE_PREFIX": "prefix",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/prefix",
					},
				},
			},
			ok: true,
		},
		{
			name: "web.basic_auth",
			env: map[string]string{
				"ARGUS_WEB_BASIC_AUTH_USERNAME": "user",
				"ARGUS_WEB_BASIC_AUTH_PASSWORD": "pass",
			},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user",
							Password: util.FmtHash(util.GetHash("pass")),
						},
					},
				},
			},
			ok: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			test.SetEnv(t, tc.env)
			settings := Settings{}

			errChannel := make(chan error, 1)
			resultChannel := make(chan bool, 1)
			// WHEN: MapEnvToStruct is called on it.
			go func() {
				err := settings.MapEnvToStruct()
				errChannel <- err
				resultChannel <- err == nil
			}()

			prefix := fmt.Sprintf(
				"%s\nSettings.MapEnvToStruct(%+v)",
				packageName, tc.env,
			)

			// THEN: the ok value is as expected.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}

			// AND: any stdout error is as expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}

			// AND: any returned error is as expected.
			tc.errRegex = util.ValueOr(tc.errRegex, `^$`)
			select {
			case err := <-errChannel:
				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Errorf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, e, tc.errRegex,
					)
				}
			default:
				t.Fatalf("%s\nerror expected but not returned", prefix)
			}

			// AND: the settings are set to the appropriate env vars.
			gotStr := settings.String("")
			wantStr := tc.want.String("")
			if gotStr != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %v\nwant: %v",
					prefix, gotStr, wantStr,
				)
			}
		})
	}
}

func TestSettings_String(t *testing.T) {
	// GIVEN: a Settings struct.
	tests := []struct {
		name     string
		settings *Settings
		prefix   string
		want     string
	}{
		{
			name:     "nil settings",
			settings: nil,
			prefix:   "",
			want:     "",
		},
		{
			name:     "empty settings",
			settings: &Settings{},
			prefix:   "",
			want:     "{}\n",
		},
		{
			name: "settings",
			settings: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Level: "INFO",
					},
				},
			},
			prefix: "",
			want: test.TrimYAML(`
				log:
					level: INFO
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.settings.String,
				tc.want,
			)
		})
	}
}

func TestSettings_NilUndefinedFlags(t *testing.T) {
	// GIVEN: tests with flags set/unset.
	var settings Settings
	tests := []struct {
		name      string
		flagSet   bool
		setStrTo  *string
		setBoolTo *bool
	}{
		{
			name:      "flag set",
			flagSet:   true,
			setStrTo:  test.Ptr("test"),
			setBoolTo: test.Ptr(true),
		},
		{
			name:      "flag not set",
			flagSet:   false,
			setStrTo:  test.Ptr("foo"),
			setBoolTo: test.Ptr(false),
		},
	}
	flagStr := "log.level"
	flagBool := "log.timestamps"
	flagset := map[string]bool{
		flagStr:  false,
		flagBool: false,
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing some env vars.

			// WHEN: flags are set/unset and NilUndefinedFlags is called.
			flagset[flagStr] = tc.flagSet
			flagset[flagBool] = tc.flagSet
			LogLevel = tc.setStrTo
			LogTimestamps = tc.setBoolTo
			settings.NilUndefinedFlags(&flagset)

			prefix := fmt.Sprintf("%s\nSettings.NilUndefinedFlags()", packageName)

			// THEN: the flags are defined/undefined correctly.
			gotStr := LogLevel
			if (tc.flagSet && gotStr == nil) ||
				(!tc.flagSet && gotStr != nil) {
				t.Errorf(
					"%s mismatch on %s - %s:\ngot:  %v\nwant: %s",
					prefix, flagStr, tc.name,
					util.DerefOr(gotStr, "<nil>"), *tc.setStrTo,
				)
			}
			gotBool := LogTimestamps
			if (tc.flagSet && gotBool == nil) ||
				(!tc.flagSet && gotBool != nil) {
				t.Errorf(
					"%s mismatch on %s - %s:\ngot:  %v\nwant: %v",
					prefix, flagBool, tc.name,
					gotBool, *tc.setBoolTo,
				)
			}
		})
	}
}

func TestSettings_Default(t *testing.T) {
	// GIVEN: a set of env vars.
	tests := []struct {
		name        string
		env         map[string]string
		stdoutRegex string
		ok          bool
	}{
		{
			name:        "no env vars",
			env:         map[string]string{},
			stdoutRegex: `^$`,
			ok:          true,
		},
		{
			name: "valid env var",
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "false",
			},
			stdoutRegex: `^$`,
			ok:          true,
		},
		{
			name: "invalid env var",
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "abc",
			},
			stdoutRegex: `^FATAL.*environment variable.*incorrect.*\s.*ARGUS_LOG_TIMESTAMPS.*\s$`,
			ok:          false,
		},
		{
			name: "web.cert-file that doesn't exist",
			env: map[string]string{
				"ARGUS_WEB_CERT_FILE": "cert.test",
			},
			stdoutRegex: test.TrimYAML(`
				^FATAL: hard_defaults:
					settings:
						web:
							cert_file: .*no such file.*`,
			),
			ok: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			test.SetEnv(t, tc.env)
			settings := Settings{}

			resultChannel := make(chan bool, 1)
			// WHEN: Default is called.
			resultChannel <- settings.Default()

			prefix := fmt.Sprintf("%s\nSettings.Default()", packageName)

			// THEN: the ok value is as expected.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}

			// AND: any error is as expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}

func TestSettings_GetStrings(t *testing.T) {
	// GIVEN: different flags/env vars are set that may impact the result of .Default().
	settings := testSettings(t)
	tests := []struct {
		name       string
		flag       **string
		flagVal    *string
		env        map[string]string
		want       string
		nilConfig  bool
		configPtr  *string
		getFunc    func() string
		getFuncPtr func() *string
	}{
		{
			name:    "log.level hard default",
			getFunc: settings.LogLevel,
			flag:    &LogLevel, want: "DEBUG",
			nilConfig: true,
			configPtr: &settings.Log.Level,
		},
		{
			name:    "log.level config",
			getFunc: settings.LogLevel,
			flag:    &LogLevel,
			want:    "DEBUG",
		},
		{
			name:    "log.level flag",
			getFunc: settings.LogLevel,
			flag:    &LogLevel,
			flagVal: test.Ptr("ERROR"),
			want:    "ERROR",
		},
		{
			name:      "data.database-file hard default",
			getFunc:   settings.DataDatabaseFile,
			flag:      &DataDatabaseFile,
			want:      "data/argus.db",
			nilConfig: true,
			configPtr: &settings.Data.DatabaseFile,
		},
		{
			name:    "data.database-file config",
			getFunc: settings.DataDatabaseFile,
			flag:    &DataDatabaseFile,
			want:    "somewhere.db",
		},
		{
			name:    "data.database-file flag",
			getFunc: settings.DataDatabaseFile,
			flag:    &DataDatabaseFile,
			flagVal: test.Ptr("ERROR"),
			want:    "ERROR",
		},
		{
			name:    "web.listen-host hard default",
			getFunc: settings.WebListenHost,
			flag:    &WebListenHost, want: "0.0.0.0",
			nilConfig: true,
			configPtr: &settings.Web.ListenHost,
		},
		{
			name:    "web.listen-host config",
			getFunc: settings.WebListenHost,
			flag:    &WebListenHost,
			want:    "test",
		},
		{
			name:    "web.listen-host flag",
			getFunc: settings.WebListenHost,
			flag:    &WebListenHost,
			flagVal: test.Ptr("127.0.0.1"),
			want:    "127.0.0.1",
		},
		{
			name:      "web.listen-port hard default",
			getFunc:   settings.WebListenPort,
			flag:      &WebListenPort,
			want:      "8080",
			nilConfig: true,
			configPtr: &settings.Web.ListenPort,
		},
		{
			name:    "web.listen-port config",
			getFunc: settings.WebListenPort,
			flag:    &WebListenPort,
			want:    "123",
		},
		{
			name:    "web.listen-port flag",
			getFunc: settings.WebListenPort,
			flag:    &WebListenPort,
			flagVal: test.Ptr("54321"),
			want:    "54321",
		},
		{
			name:      "web.cert-file hard default",
			getFunc:   settings.WebCertFile,
			flag:      &WebCertFile,
			want:      "",
			nilConfig: true,
			configPtr: &settings.Web.CertFile,
		},
		{
			name:    "web.cert-file config",
			getFunc: settings.WebCertFile,
			flag:    &WebCertFile,
			want:    "../README.md",
		},
		{
			name:    "web.cert-file flag",
			getFunc: settings.WebCertFile,
			flag:    &WebCertFile,
			flagVal: test.Ptr("settings_test.go"),
			want:    "settings_test.go",
		},
		{
			name:      "web.pkey-file hard default",
			getFunc:   settings.WebKeyFile,
			flag:      &WebPKeyFile,
			want:      "",
			nilConfig: true,
			configPtr: &settings.Web.KeyFile,
		},
		{
			name:    "web.pkey-file config",
			getFunc: settings.WebKeyFile,
			flag:    &WebPKeyFile,
			want:    "../LICENSE",
		},
		{
			name:    "web.pkey-file flag",
			getFunc: settings.WebKeyFile,
			flag:    &WebPKeyFile,
			flagVal: test.Ptr("settings_test.go"),
			want:    "settings_test.go",
		},
		{
			name:      "web.route-prefix hard default",
			getFunc:   settings.WebRoutePrefix,
			flag:      &WebRoutePrefix,
			want:      "/",
			nilConfig: true,
			configPtr: &settings.Web.RoutePrefix,
		},
		{
			name:    "web.route-prefix config",
			getFunc: settings.WebRoutePrefix,
			flag:    &WebRoutePrefix,
			want:    "/something",
		},
		{
			name:    "web.route-prefix flag",
			getFunc: settings.WebRoutePrefix,
			flag:    &WebRoutePrefix,
			flagVal: test.Ptr("/flag"),
			want:    "/flag",
		},
		{
			name: "set from env",
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "ERROR",
			},
			getFunc: func() string { return settings.HardDefaults.Log.Level },
			want:    "ERROR",
		},
	}

	loadMu.Lock() // Protect flag env vars.
	t.Cleanup(loadMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing some env vars.
			releaseStdout := test.CaptureLog(t, logx.Default())
			defer releaseStdout()
			test.SetEnv(t, tc.env)

			settings = testSettings(t)
			if tc.flag != nil {
				had := *tc.flag
				*tc.flag = tc.flagVal
				t.Cleanup(func() { *tc.flag = had })
			}

			// WHEN: Default is called on it.
			settings.Default()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = ""
				t.Cleanup(func() { *tc.configPtr = had })
			}

			// THEN: the specified part is initialised correctly.
			var got string
			switch {
			case tc.getFunc != nil:
				got = tc.getFunc()
			case tc.getFuncPtr != nil:
				got = util.DerefOr(tc.getFuncPtr(), "<nil>")
			default:
				t.Fatalf(
					"%s\ninvalid test case %q: no getFunc or getFuncPtr specified for Settings",
					packageName, tc.name,
				)
			}
			if got != tc.want {
				t.Errorf(
					"%s\nmismatch on Settings GetX\ngot:  %v\nwant: %s",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestSettings_GetBool(t *testing.T) {
	// GIVEN: vars set in dif	// GIVEN: different flags/env vars are set that may impact the result of .Default().
	settings := testSettings(t)
	tests := []struct {
		name       string
		flag       **bool
		flagVal    *bool
		want       string
		nilConfig  bool
		configPtr  **bool
		getFunc    func() bool
		getFuncPtr func() *bool
	}{
		{
			name:       "log.timestamps hard default",
			getFuncPtr: settings.LogTimestamps,
			flag:       &LogTimestamps,
			want:       "false",
			nilConfig:  true,
			configPtr:  &settings.Log.Timestamps,
		},
		{
			name:       "log.timestamps config",
			getFuncPtr: settings.LogTimestamps,
			flag:       &LogTimestamps,
			want:       "true",
		},
		{
			name:       "log.timestamps flag",
			getFuncPtr: settings.LogTimestamps,
			flag:       &LogTimestamps,
			flagVal:    test.Ptr(true),
			want:       "true",
		},
	}

	loadMu.Lock() // Protect flag env vars.
	t.Cleanup(loadMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing flag vars.

			had := *tc.flag
			*tc.flag = tc.flagVal
			t.Cleanup(func() { *tc.flag = had })

			// WHEN: Default is called on it.
			settings.Default()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = nil
				t.Cleanup(func() { *tc.configPtr = had })
			}

			// THEN: the specified part is initialised correctly.
			var got string
			switch {
			case tc.getFunc != nil:
				got = fmt.Sprint(tc.getFunc())
			case tc.getFuncPtr != nil:
				ptr := tc.getFuncPtr()
				got = "<nil>"
				if ptr != nil {
					got = fmt.Sprint(*tc.getFuncPtr())
				}
			default:
				t.Fatalf(
					"%s\ninvalid test case %q: no getFunc or getFuncPtr specified for Settings",
					packageName, tc.name,
				)
			}
			if got != tc.want {
				t.Errorf(
					"%s\nmismatch on Settings GetX\ngot:  %v\nwant: %s",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestSettings_GetWebFile__notExist(t *testing.T) {
	settings := Settings{
		SettingsBase: SettingsBase{
			Log: LogSettings{},
		},
		FromFlags: SettingsBase{
			Log: LogSettings{},
		},
		HardDefaults: SettingsBase{
			Log: LogSettings{},
		},
	}

	// GIVEN: different target vars, and their respective 'get' functions.
	tests := []struct {
		name      string
		getFunc   func() string
		changeVar any
	}{
		{
			name:      "hard default cert file",
			changeVar: &settings.Web.CertFile,
			getFunc:   settings.WebCertFile,
		},
		{
			name:      "config cert file",
			changeVar: &settings.Web.CertFile,
			getFunc:   settings.WebCertFile,
		},
		{
			name:      "flag cert file",
			changeVar: &settings.FromFlags.Web.CertFile,
			getFunc:   settings.WebCertFile,
		},
		{
			name:      "hard default pkey file",
			changeVar: &settings.Web.KeyFile,
			getFunc:   settings.WebKeyFile,
		},
		{
			name:      "config pkey file",
			changeVar: &settings.Web.KeyFile,
			getFunc:   settings.WebKeyFile,
		},
		{
			name:      "flag pkey file",
			changeVar: &settings.FromFlags.Web.KeyFile,
			getFunc:   settings.WebKeyFile,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing the Settings struct.

			t.Cleanup(func() {
				if ptr, ok := tc.changeVar.(*string); ok {
					*ptr = ""
				} else if ptrPtr, ok := tc.changeVar.(**string); ok {
					f := ""
					*ptrPtr = &f
				}
			})

			//
			// Test 1: empty string.
			//
			file := ""
			if ptr, ok := tc.changeVar.(*string); ok {
				*ptr = file
			} else if ptrPtr, ok := tc.changeVar.(**string); ok {
				*ptrPtr = &file
			}
			// WHEN: a get is called with no file path set.
			got := tc.getFunc()

			prefix := fmt.Sprintf(
				"%s\nGetWebFile(%s)",
				packageName, tc.name,
			)

			// THEN: the empty string is returned.
			if got != file {
				t.Errorf(
					"%s mismatch when unset\ngot:  %q\nwant: %q",
					prefix, got, file,
				)
			}

			//
			// Test 2: file path.
			//
			file = fmt.Sprintf(
				"test_%s.pem",
				strings.ReplaceAll(strings.ToLower(tc.name), " ", "_"),
			)
			if ptr, ok := tc.changeVar.(*string); ok {
				*ptr = file
			} else if ptrPtr, ok := tc.changeVar.(**string); ok {
				*ptrPtr = &file
			}
			// WHEN: a get is called with a file path set.
			got = tc.getFunc()

			// THEN: the file path is returned.
			if got != file {
				t.Errorf(
					"%s mismatch when set\ngot:  %q\nwant: %q",
					prefix, got, file,
				)
			}
		})
	}
}

func TestSettings_WebBasicAuthUsernameHash(t *testing.T) {
	// GIVEN: a Settings struct with some values set.
	tests := []struct {
		name string
		want string // The string that was hashed.
		had  Settings
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name: "set in config",
			want: "user",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user",
						},
					},
				},
			},
		},
		{
			name: "set in flag",
			want: "user",
			had: Settings{
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := util.GetHash(tc.want)
			_ = tc.had.CheckValues()
			_ = tc.had.FromFlags.CheckValues()
			// HardDefaults.Web.BasicAuth will never be nil if Basic Auth is in use.
			tc.had.HardDefaults = SettingsBase{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						UsernameHash: util.GetHash(""),
						PasswordHash: util.GetHash(""),
					},
				},
			}

			// WHEN: WebBasicAuthUsernameHash is called on it.
			got := tc.had.WebBasicAuthUsernameHash()

			// THEN: the hash is returned.
			if got != want {
				t.Errorf(
					"%s\nWebBasicAuthUsernameHash() mismatch\ngot:  %s\nwant: %s",
					packageName, got, want,
				)
			}
		})
	}
}

func TestSettings_WebBasicAuthPasswordHash(t *testing.T) {
	// GIVEN: a Settings struct with some values set.
	tests := []struct {
		name string
		want string // The string that was hashed.
		had  Settings
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name: "set in config",
			want: "pass",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass",
						},
					},
				},
			},
		},
		{
			name: "set in flag",
			want: "pass",
			had: Settings{
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass",
						},
					},
				},
			},
		},
		{
			name: "set everywhere, use flag",
			want: "flag",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "config",
						},
					},
				},
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "flag",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := util.GetHash(tc.want)
			_ = tc.had.CheckValues()
			_ = tc.had.FromFlags.CheckValues()
			// HardDefaults.Web.BasicAuth will never be nil if Basic Auth is in use.
			tc.had.HardDefaults = SettingsBase{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						UsernameHash: util.GetHash(""),
						PasswordHash: util.GetHash(""),
					},
				},
			}

			// WHEN: WebBasicAuthPasswordHash is called on it.
			got := tc.had.WebBasicAuthPasswordHash()

			// THEN: the hash is returned.
			if got != want {
				t.Errorf(
					"%s\nWebBasicAuthPasswordHash() mismatch\ngot:  %s\nwant: %s",
					packageName, got, want,
				)
			}
		})
	}
}
