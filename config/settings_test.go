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
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestSettings_MapEnvToStruct(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		want     *Settings
		errRegex string
	}{
		{
			name: "empty vars ignored",
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "",
			},
			want:     &Settings{},
			errRegex: `^$`,
		},
		{
			name: "data.readonly",
			env: map[string]string{
				"ARGUS_DATA_READONLY": "true",
			},
			want: &Settings{
				Data: DataSettings{
					Readonly: new(true),
				},
			},
			errRegex: `^$`,
		},
		{
			name: "web.demo.username",
			env: map[string]string{
				"ARGUS_WEB_DEMO_USERNAME": "demo",
			},
			want: &Settings{
				Web: WebSettings{
					Demo: &WebSettingsDemo{
						Username: "demo",
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "web.demo.password",
			env: map[string]string{
				"ARGUS_WEB_DEMO_PASSWORD": "demo",
			},
			want: &Settings{
				Web: WebSettings{
					Demo: &WebSettingsDemo{
						Password: "demo",
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "web.demo.group",
			env: map[string]string{
				"ARGUS_WEB_DEMO_GROUP": "demo",
			},
			want: &Settings{
				Web: WebSettings{
					Demo: &WebSettingsDemo{
						Group: "demo",
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "auth.session.idle_timeout",
			env: map[string]string{
				"ARGUS_AUTH_SESSION_IDLE_TIMEOUT": "15m",
			},
			want: &Settings{
				Auth: AuthSettings{
					Session: AuthSessionSettings{
						IdleTimeout: "15m",
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "auth.session.lifetime",
			env: map[string]string{
				"ARGUS_AUTH_SESSION_LIFETIME": "48h",
			},
			want: &Settings{
				Auth: AuthSettings{
					Session: AuthSessionSettings{
						Lifetime: "48h",
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "log.level",
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "ERROR",
			},
			want: &Settings{
				Log: LogSettings{
					Level: "ERROR",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "log.timestamps/valid",
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "true",
			},
			want: &Settings{
				Log: LogSettings{
					Timestamps: new(true),
				},
			},
			errRegex: `^$`,
		},
		{
			name: "log.timestamps/invalid/not a bool",
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "abc",
			},
			want: &Settings{},
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
				Web: WebSettings{
					ListenHost: "test",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "web.listen-port",
			env: map[string]string{
				"ARGUS_WEB_LISTEN_PORT": "123",
			},
			want: &Settings{
				Web: WebSettings{
					ListenPort: "123",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "web.cert-file",
			env: map[string]string{
				"ARGUS_WEB_CERT_FILE": "cert.test",
			},
			want: &Settings{
				Web: WebSettings{
					CertFile: "cert.test",
				},
			},
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
				Web: WebSettings{
					KeyFile: "pkey.test",
				},
			},
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
				Web: WebSettings{
					RoutePrefix: "/prefix", // CheckValues adds the leading '/'.
				},
			},
			errRegex: `^$`,
		},
		{
			name: "web.disabled_routes",
			env: map[string]string{
				"ARGUS_WEB_DISABLED_ROUTES": "service_delete, notify_test",
			},
			want: &Settings{
				Web: WebSettings{
					DisabledRoutes: []string{"service_delete", "notify_test"},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "web.basic_auth",
			env: map[string]string{
				"ARGUS_WEB_BASIC_AUTH_USERNAME": "user",
				"ARGUS_WEB_BASIC_AUTH_PASSWORD": "pass",
			},
			want: &Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Username: "user",
						Password: util.FmtHash(util.GetHash("pass")), // CheckValues hashes it.
					},
				},
			},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			test.IsolateEnv(t, tc.env)
			settings := Settings{}

			var err error
			test.AssertCallLog(
				t,
				tc.errRegex == `^$`,
				func() bool {
					err = settings.MapEnvToStruct()
					return err == nil
				},
				logx.ExitCodeChannel(),
				releaseStdout,
				`^$`,
			)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			gotStr := settings.String("")
			wantStr := tc.want.String("")
			if gotStr != wantStr {
				t.Errorf("Settings.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}
		})
	}
}

func TestDataSettings_IsZero(t *testing.T) {
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
			name: "non-empty/DatabaseFile",
			data: DataSettings{
				DatabaseFile: "db.sqlite",
			},
			want: false,
		},
		{
			name: "non-empty/Readonly",
			data: DataSettings{
				Readonly: new(true),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestLogSettings_IsZero(t *testing.T) {
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
			name: "non-empty/Timestamps",
			data: LogSettings{
				Timestamps: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/Level",
			data: LogSettings{
				Level: "INFO",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: LogSettings{
				Timestamps: new(true),
				Level:      "INFO",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestWebSettingsBasicAuth_String(t *testing.T) {
	tests := []struct {
		name string
		auth *WebSettingsBasicAuth
		want string
	}{
		{
			name: "nil auth",
			auth: nil,
			want: "",
		},
		{
			name: "empty auth",
			auth: &WebSettingsBasicAuth{},
			want: "{}\n",
		},
		{
			name: "auth with values",
			auth: &WebSettingsBasicAuth{
				Username: "user",
				Password: "pass",
			},
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
				"WebSettingsBasicAuth.String",
				tc.auth.String,
				tc.want,
			)
		})
	}
}

func TestWebSettingsBasicAuth_CheckValues(t *testing.T) {
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
				Password: "",
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

			test.IsolateEnv(t, tc.env)

			target := tc.input()

			target.CheckValues()

			// A Password that env expansion leaves unchanged is stored hashed.
			gotStr := target.String("")
			wantStr := tc.want.String("")
			if gotStr != wantStr {
				t.Errorf(
					"WebSettingsBasicAuth.String() mismatch\ngot:  %q\nwant: %q",
					gotStr, wantStr,
				)
			}

			want := util.FmtHash(
				util.GetHash(
					util.EvalEnvVars(tc.want.Username),
				),
			)
			if tc.wantUsernameHash != "" {
				want = tc.wantUsernameHash
			}
			got := util.FmtHash(target.UsernameHash)
			if got != want {
				t.Errorf(
					"WebSettingsBasicAuth.UsernameHash mismatch\ngot:  %q\nwant: %q",
					got, want,
				)
			}

			want = util.FmtHash(
				util.GetHash(
					util.EvalEnvVars(tc.want.Password),
				),
			)
			if tc.wantPasswordHash != "" {
				want = tc.wantPasswordHash
			}
			got = util.FmtHash(target.PasswordHash)
			if got != want {
				t.Errorf(
					"WebSettingsBasicAuth.PasswordHash mismatch\ngot:  %q\nwant: %q",
					got, want,
				)
			}
		})
	}
}

func TestWebSettings_IsZero(t *testing.T) {
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
			name: "non-empty/ListenHost",
			data: WebSettings{
				ListenHost: "0.0.0.0",
			},
			want: false,
		},
		{
			name: "non-empty/ListenPort",
			data: WebSettings{
				ListenPort: "8080",
			},
			want: false,
		},
		{
			name: "non-empty/RoutePrefix",
			data: WebSettings{
				RoutePrefix: "/test",
			},
			want: false,
		},
		{
			name: "non-empty/CertFile",
			data: WebSettings{
				CertFile: "cert.pem",
			},
			want: false,
		},
		{
			name: "non-empty/KeyFile",
			data: WebSettings{
				KeyFile: "privkey.pem",
			},
			want: false,
		},
		{
			name: "BasicAuth/empty",
			data: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{},
			},
			want: false,
		},
		{
			name: "BasicAuth/non-empty",
			data: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: "pass",
				},
			},
			want: false,
		},
		{
			name: "non-empty/DisabledRoutes",
			data: WebSettings{
				DisabledRoutes: []string{"route1", "route2"},
			},
			want: false,
		},
		{
			name: "Favicon/empty",
			data: WebSettings{
				Favicon: &FaviconSettings{},
			},
			want: false,
		},
		{
			name: "Favicon/non-empty",
			data: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "favicon.svg",
				},
			},
			want: false,
		},
		{
			name: "Demo/empty",
			data: WebSettings{
				Demo: &WebSettingsDemo{},
			},
			want: false,
		},
		{
			name: "Demo/non-empty",
			data: WebSettings{
				Demo: &WebSettingsDemo{
					Username: "demo",
					Password: "demo",
					Group:    "demo",
				},
			},
			want: false,
		},
		{
			name: "non-empty/ListenHost + ListenPort",
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

			got := tc.data.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestWebSettings_String(t *testing.T) {
	tests := []struct {
		name        string
		webSettings *WebSettings
		want        string
	}{
		{
			name:        "nil webSettings",
			webSettings: nil,
			want:        "",
		},
		{
			name:        "empty webSettings",
			webSettings: &WebSettings{},
			want:        "{}\n",
		},
		{
			name: "webSettings with values",
			webSettings: &WebSettings{
				ListenHost: "0.0.0.0",
				ListenPort: "8080",
			},
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
				"WebSettings.String",
				tc.webSettings.String,
				tc.want,
			)
		})
	}
}

func TestWebSettings_CheckValues(t *testing.T) {
	tests := []struct {
		name                               string
		env                                map[string]string
		input                              *WebSettings
		want                               string
		wantUsernameHash, wantPasswordHash string
		errRegex                           string
	}{
		{
			name: "BasicAuth/empty",
			input: &WebSettings{
				BasicAuth: &WebSettingsBasicAuth{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "BasicAuth/str Username and Password already hashed",
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
			errRegex:         `^$`,
		},
		{
			name: "BasicAuth/hashed Username and str Password",
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
			errRegex:         `^$`,
		},
		{
			name: "BasicAuth/Username and password from env vars",
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
			errRegex:         `^$`,
		},
		{
			name: "BasicAuth/Username and password from env vars partial",
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
			errRegex:         `^$`,
		},
		{
			name: "Favicon/empty",
			input: &WebSettings{
				Favicon: &FaviconSettings{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "Favicon/SVG",
			input: &WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg",
				},
			},
			want: test.TrimYAML(`
				favicon:
					svg: https://example.com/favicon.svg
			`),
			errRegex: `^$`,
		},
		{
			name: "Favicon/PNG",
			input: &WebSettings{
				Favicon: &FaviconSettings{
					PNG: "https://example.com/favicon.png",
				},
			},
			want: test.TrimYAML(`
				favicon:
					png: https://example.com/favicon.png
			`),
			errRegex: `^$`,
		},
		{
			name: "Favicon/full",
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
			errRegex: `^$`,
		},
		{
			name: "Demo/empty",
			input: &WebSettings{
				Demo: &WebSettingsDemo{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "Demo/credentials",
			input: &WebSettings{
				Demo: &WebSettingsDemo{
					Username: "demo",
					Password: "demo",
				},
			},
			want: test.TrimYAML(`
				demo:
					username: demo
					password: demo
			`),
			errRegex: `^$`,
		},
		{
			name: "Demo/group only, kept",
			input: &WebSettings{
				Demo: &WebSettingsDemo{
					Group: "demo",
				},
			},
			want: test.TrimYAML(`
				demo:
					group: demo
			`),
			errRegex: `^$`,
		},
		{
			name: "Web.CertFile, not found",
			input: &WebSettings{
				CertFile: "cert.pem",
			},
			want:     "cert_file: cert.pem\n",
			errRegex: `^cert_file: .*no such file.*$`,
		},
		{
			name: "Web.KeyFile, not found",
			input: &WebSettings{
				KeyFile: "privkey.pem",
			},
			want:     "pkey_file: privkey.pem\n",
			errRegex: `^pkey_file: .*no such file.*$`,
		},
		{
			name: "Web.CertFile + Web.KeyFile, both not found",
			input: &WebSettings{
				CertFile: "cert.pem",
				KeyFile:  "privkey.pem",
			},
			want: test.TrimYAML(`
				cert_file: cert.pem
				pkey_file: privkey.pem
			`),
			errRegex: test.TrimYAML(`
				^cert_file: .*no such file.*
				pkey_file: .*no such file.*$`,
			),
		},
		{
			name: "TrustedProxies/valid IP and CIDR",
			input: &WebSettings{
				TrustedProxies: []string{"10.0.0.1", "192.168.0.0/16", "::1"},
			},
			want: test.TrimYAML(`
				trusted_proxies:
					- 10.0.0.1
					- 192.168.0.0/16
					- ::1
			`),
			errRegex: `^$`,
		},
		{
			name: "TrustedProxies/invalid entry",
			input: &WebSettings{
				TrustedProxies: []string{"10.0.0.1", "not-an-ip"},
			},
			want: test.TrimYAML(`
				trusted_proxies:
					- 10.0.0.1
					- not-an-ip
			`),
			errRegex: `^trusted_proxies: "not-an-ip" <invalid>.*IP address or CIDR`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.IsolateEnv(t, tc.env)

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				tc.input.CheckValues,
			)

			gotStr := tc.input.String("")
			if gotStr != tc.want {
				t.Errorf("WebSettings.String() mismatch\ngot:  %q\nwant: %q", gotStr, tc.want)
			}

			if tc.wantUsernameHash != "" {
				got := util.FmtHash(tc.input.BasicAuth.UsernameHash)
				if got != tc.wantUsernameHash {
					t.Errorf(
						"WebSettings.BasicAuth.UsernameHash mismatch\ngot:  %q\nwant: %q",
						got, tc.wantUsernameHash,
					)
				}
			}
			if tc.wantPasswordHash != "" {
				got := util.FmtHash(tc.input.BasicAuth.PasswordHash)
				if got != tc.wantPasswordHash {
					t.Errorf(
						"WebSettings.BasicAuth.PasswordHash mismatch\ngot:  %q\nwant: %q",
						got, tc.wantPasswordHash,
					)
				}
			}
		})
	}
}

func TestSettings_IsZero(t *testing.T) {
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
			name: "non-empty/Auth",
			data: Settings{
				Auth: AuthSettings{
					Enabled: new(true),
				},
			},
			want: false,
		},
		{
			name: "non-empty/Log",
			data: Settings{
				Log: LogSettings{
					Timestamps: new(true),
				},
			},
			want: false,
		},
		{
			name: "non-empty/Data",
			data: Settings{
				Data: DataSettings{
					DatabaseFile: "db.sqlite",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Web",
			data: Settings{
				Web: WebSettings{
					ListenHost: "0.0.0.0",
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: Settings{
				Auth: AuthSettings{
					Enabled: new(true),
				},
				Log: LogSettings{
					Timestamps: new(true),
				},
				Data: DataSettings{
					DatabaseFile: "db.sqlite",
				},
				Web: WebSettings{
					ListenHost: "0.0.0.0",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestSettings_String(t *testing.T) {
	tests := []struct {
		name     string
		settings *Settings
		want     string
	}{
		{
			name:     "nil settings",
			settings: nil,
			want:     "",
		},
		{
			name:     "empty settings",
			settings: &Settings{},
			want:     "{}\n",
		},
		{
			name: "settings",
			settings: &Settings{
				Log: LogSettings{
					Level: "INFO",
				},
			},
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
				"Settings.String",
				tc.settings.String,
				tc.want,
			)
		})
	}
}

func TestSettings_CheckValues(t *testing.T) {
	tests := []struct {
		name                               string
		env                                map[string]string
		input                              *Settings
		want                               string
		wantUsernameHash, wantPasswordHash string
		wantIndent                         uint8
		errRegex                           string
	}{
		{
			name: "indentation/0",
			input: &Settings{
				Indentation: 0,
			},
			want:       "{}\n",
			wantIndent: yaml.DefaultIndentSpaces,
			errRegex:   `^$`,
		},
		{
			name: "indentation/1",
			input: &Settings{
				Indentation: 1,
			},
			want:       "{}\n",
			wantIndent: 1,
			errRegex:   `^$`,
		},
		{
			name: "indentation/2",
			input: &Settings{
				Indentation: 2,
			},
			want:       "{}\n",
			wantIndent: 2,
			errRegex:   `^$`,
		},
		{
			name: "indentation/3",
			input: &Settings{
				Indentation: 3,
			},
			want:       "{}\n",
			wantIndent: 3,
			errRegex:   `^$`,
		},
		{
			name: "indentation/4",
			input: &Settings{
				Indentation: 4,
			},
			want:       "{}\n",
			wantIndent: 4,
			errRegex:   `^$`,
		},
		{
			name: "BasicAuth/empty",
			input: &Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{},
				},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "BasicAuth/hashed Username and str env Password",
			env: map[string]string{
				"TEST_SETTINGS_BASE__CHECK_VALUES__ONE": "ass",
			},
			input: &Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Username: util.FmtHash(util.GetHash("user")),
						Password: "p${TEST_SETTINGS_BASE__CHECK_VALUES__ONE}",
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
			errRegex:         `^$`,
		},
		{
			name: "BasicAuth/username without a password",
			input: &Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Username: "user",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					basic_auth:
						username: user
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			wantPasswordHash: util.FmtHash(util.GetHash("")),
			errRegex:         `^$`,
		},
		{
			name: "BasicAuth/password without a username",
			input: &Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Password: "pass",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					basic_auth:
						password: ` + util.FmtHash(util.GetHash("pass")) + `
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
			errRegex:         `^$`,
		},
		{
			name: "BasicAuth/empty credentials from the flags",
			input: &Settings{
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{},
					},
				},
			},
			want: "{}\n",
			errRegex: test.TrimYAML(`
				^web:
					basic_auth:
						a username and\/or a password is required$`,
			),
		},
		{
			name: "BasicAuth/username from the flags, password from the config",
			input: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass",
						},
					},
				},
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user",
						},
					},
				},
			},
			want: test.TrimYAML(`
				web:
					basic_auth:
						password: ` + util.FmtHash(util.GetHash("pass")) + `
			`),
			wantUsernameHash: util.FmtHash(util.GetHash("")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
			errRegex:         `^$`,
		},
		{
			name: "Route prefix/empty",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "",
				},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "Route prefix/no leading /",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "test",
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			errRegex: `^$`,
		},
		{
			name: "Route prefix/leading /",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "/test",
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			errRegex: `^$`,
		},
		{
			name: "Route prefix/multiple leading /",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "///test",
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			errRegex: `^$`,
		},
		{
			name: "Route prefix/trailing /",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "/test/",
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			errRegex: `^$`,
		},
		{
			name: "Route prefix/multiple trailing /",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "/test///",
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /test
			`),
			errRegex: `^$`,
		},
		{
			name: "Route prefix/only a /",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "/",
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /
			`),
			errRegex: `^$`,
		},
		{
			name: "Route prefix/only multiple /",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "///",
				},
			},
			want: test.TrimYAML(`
				web:
					route_prefix: /
			`),
			errRegex: `^$`,
		},
		{
			name: "Favicon/empty",
			input: &Settings{
				Web: WebSettings{
					Favicon: &FaviconSettings{},
				},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "Favicon/full",
			input: &Settings{
				Web: WebSettings{
					Favicon: &FaviconSettings{
						SVG: "https://example.com/favicon.svg",
						PNG: "https://example.com/favicon.png",
					},
				},
			},
			want: test.TrimYAML(`
				web:
					favicon:
						svg: https://example.com/favicon.svg
						png: https://example.com/favicon.png
			`),
			errRegex: `^$`,
		},
		{
			name: "Web.CertFile, not found",
			input: &Settings{
				Web: WebSettings{
					CertFile: "cert.pem",
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
		},
		{
			name: "Web.KeyFile, not found",
			input: &Settings{
				Web: WebSettings{
					KeyFile: "privkey.pem",
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
		},
		{
			name: "Web.CertFile + Web.KeyFile, both not found",
			input: &Settings{
				Web: WebSettings{
					CertFile: "cert.pem",
					KeyFile:  "privkey.pem",
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

			if tc.wantIndent == 0 {
				tc.wantIndent = yaml.DefaultIndentSpaces
			}
			test.IsolateEnv(t, tc.env)

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				tc.input.CheckValues,
			)

			if got := tc.input.String(""); got != tc.want {
				t.Errorf("Settings.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			if gotIndent := tc.input.Indentation; gotIndent != tc.wantIndent {
				t.Errorf(
					"Settings.Indentation mismatch\ngot:  %d\nwant: %d",
					gotIndent, tc.wantIndent,
				)
			}

			if tc.input.Web.BasicAuth != nil {
				if got := util.FmtHash(tc.input.Web.BasicAuth.UsernameHash); got != tc.wantUsernameHash {
					t.Errorf(
						"Settings.Web.BasicAuth.UsernameHash mismatch\ngot:  %q\nwant: %q",
						got, tc.wantUsernameHash,
					)
				}
				if got := util.FmtHash(tc.input.Web.BasicAuth.PasswordHash); got != tc.wantPasswordHash {
					t.Errorf(
						"Settings.Web.BasicAuth.PasswordHash mismatch\ngot:  %q\nwant: %q",
						got, tc.wantPasswordHash,
					)
				}
			}
		})
	}
}

func TestSettings_NilUndefinedFlags(t *testing.T) {
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
			setStrTo:  new("test"),
			setBoolTo: new(true),
		},
		{
			name:      "flag not set",
			flagSet:   false,
			setStrTo:  new("foo"),
			setBoolTo: new(false),
		},
	}
	flagStr := "log.level"
	flagBool := "log.timestamps"
	flagset := map[string]bool{
		flagStr:  false,
		flagBool: false,
	}

	loadMu.Lock()
	t.Cleanup(loadMu.Unlock)
	t.Cleanup(snapshotFlags())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares the flag vars.

			flagset[flagStr] = tc.flagSet
			flagset[flagBool] = tc.flagSet
			LogLevel = tc.setStrTo
			LogTimestamps = tc.setBoolTo
			settings.NilUndefinedFlags(&flagset)

			want := "nil"
			if tc.flagSet {
				want = "non-nil"
			}
			if (LogLevel != nil) != tc.flagSet {
				t.Errorf("LogLevel mismatch\ngot:  %s\nwant: %s", test.StringifyPtr(LogLevel), want)
			}
			if (LogTimestamps != nil) != tc.flagSet {
				t.Errorf(
					"LogTimestamps mismatch\ngot:  %s\nwant: %s",
					test.StringifyPtr(LogTimestamps), want,
				)
			}
		})
	}
}

func TestSettings_Default(t *testing.T) {
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
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			test.IsolateEnv(t, tc.env)
			settings := Settings{}

			test.AssertCallLog(
				t,
				tc.ok,
				settings.Default,
				logx.ExitCodeChannel(),
				releaseStdout,
				tc.stdoutRegex,
			)
		})
	}
}

func TestSettings_Default__BasicAuthFromFlags(t *testing.T) {
	tests := []struct {
		name         string
		usernameFlag *string
		passwordFlag *string
		want         *WebSettingsBasicAuth
	}{
		{
			name:         "neither flag provided",
			usernameFlag: nil,
			passwordFlag: nil,
			want:         nil,
		},
		{
			name:         "only username flag provided",
			usernameFlag: new("test-user"),
			passwordFlag: nil,
			want: &WebSettingsBasicAuth{
				Username: "test-user",
				Password: "",
			},
		},
		{
			name:         "only password flag provided",
			usernameFlag: nil,
			passwordFlag: new("test-pass"),
			want: &WebSettingsBasicAuth{
				Password: util.FmtHash(util.GetHash("test-pass")),
			},
		},
		{
			name:         "both flags provided",
			usernameFlag: new("test-user"),
			passwordFlag: new("test-pass"),
			want: &WebSettingsBasicAuth{
				Username: "test-user",
				Password: util.FmtHash(util.GetHash("test-pass")),
			},
		},
	}

	// loadMu guards the flag vars.
	loadMu.Lock()
	t.Cleanup(loadMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares the flag vars.

			hadUsername, hadPassword := WebBasicAuthUsername, WebBasicAuthPassword
			WebBasicAuthUsername, WebBasicAuthPassword = tc.usernameFlag, tc.passwordFlag
			t.Cleanup(func() {
				WebBasicAuthUsername, WebBasicAuthPassword = hadUsername, hadPassword
			})

			settings := Settings{}

			settings.Default()

			got := settings.FromFlags.Web.BasicAuth
			switch {
			case tc.want == nil:
				if got != nil {
					t.Error("Settings.FromFlags.Web.BasicAuth mismatch\ngot:  non-nil\nwant: nil")
				}
			case got == nil:
				t.Error("Settings.FromFlags.Web.BasicAuth mismatch\ngot:  nil\nwant: non-nil")
			default:
				if got.Username != tc.want.Username {
					t.Errorf(
						"Settings.FromFlags.Web.BasicAuth.Username mismatch\ngot:  %q\nwant: %q",
						got.Username, tc.want.Username,
					)
				}
				if got.Password != tc.want.Password {
					t.Errorf(
						"Settings.FromFlags.Web.BasicAuth.Password mismatch\ngot:  %q\nwant: %q",
						got.Password, tc.want.Password,
					)
				}
			}
		})
	}
}

func TestSettings_GetStrings(t *testing.T) {
	settings := testSettings()
	tests := []struct {
		name      string
		flag      **string
		flagVal   *string
		env       map[string]string
		want      string
		nilConfig bool
		configPtr *string
		getFunc   func() string
	}{
		{
			name:    "log.level hard default",
			getFunc: settings.LogLevel,
			flag:    &LogLevel, want: "INFO",
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
			flagVal: new("ERROR"),
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
			flagVal: new("ERROR"),
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
			flagVal: new("127.0.0.1"),
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
			flagVal: new("54321"),
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
			flagVal: new("settings_test.go"),
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
			flagVal: new("settings_test.go"),
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
			flagVal: new("/flag"),
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

	// loadMu guards the flag vars.
	loadMu.Lock()
	t.Cleanup(loadMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars and captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())
			defer releaseStdout()
			test.IsolateEnv(t, tc.env)

			settings = testSettings()
			if tc.flag != nil {
				had := *tc.flag
				*tc.flag = tc.flagVal
				t.Cleanup(func() { *tc.flag = had })
			}

			settings.Default()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = ""
				t.Cleanup(func() { *tc.configPtr = had })
			}

			if got := tc.getFunc(); got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestSettings_GetBool(t *testing.T) {
	settings := testSettings()
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
			name:      "data.readonly hard default",
			getFunc:   settings.DataReadonly,
			flag:      &DataReadonly,
			want:      "false",
			nilConfig: true,
			configPtr: &settings.Data.Readonly,
		},
		{
			name:    "data.readonly config",
			getFunc: settings.DataReadonly,
			flag:    &DataReadonly,
			want:    "true",
		},
		{
			name:    "data.readonly flag",
			getFunc: settings.DataReadonly,
			flag:    &DataReadonly,
			flagVal: new(false),
			want:    "false",
		},
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
			flagVal:    new(false),
			want:       "false",
		},
	}

	// loadMu guards the flag vars.
	loadMu.Lock()
	t.Cleanup(loadMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares the flag vars.

			had := *tc.flag
			*tc.flag = tc.flagVal
			t.Cleanup(func() { *tc.flag = had })

			settings.Default()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = nil
				t.Cleanup(func() { *tc.configPtr = had })
			}

			var got string
			switch {
			case tc.getFunc != nil:
				got = strconv.FormatBool(tc.getFunc())
			case tc.getFuncPtr != nil:
				ptr := tc.getFuncPtr()
				got = "nil"
				if ptr != nil {
					got = strconv.FormatBool(*ptr)
				}
			default:
				t.Fatal("invalid test case: no getFunc or getFuncPtr specified")
			}
			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestSettings_GetWebFile__notExist(t *testing.T) {
	settings := Settings{
		Log: LogSettings{},
		FromFlags: SettingsBase{
			Log: LogSettings{},
		},
		HardDefaults: SettingsBase{
			Log: LogSettings{},
		},
	}

	tests := []struct {
		name      string
		getFunc   func() string
		changeVar *string
	}{
		{
			name:      "hard default cert file",
			changeVar: &settings.HardDefaults.Web.CertFile,
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
			changeVar: &settings.HardDefaults.Web.KeyFile,
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
			// Serial: shares settings.

			t.Cleanup(func() { *tc.changeVar = "" })

			*tc.changeVar = ""
			got := tc.getFunc()

			if got != "" {
				t.Errorf("unset: mismatch\ngot:  %q\nwant: \"\"", got)
			}

			file := fmt.Sprintf(
				"test_%s.pem",
				strings.ReplaceAll(strings.ToLower(tc.name), " ", "_"),
			)
			*tc.changeVar = file
			got = tc.getFunc()

			// Returned even though no such file exists.
			if got != file {
				t.Errorf("set: mismatch\ngot:  %q\nwant: %q", got, file)
			}
		})
	}
}

func TestSettings_WebDisabledRoutes(t *testing.T) {
	tests := []struct {
		name         string
		values       []string
		hardDefaults []string
		want         []string
	}{
		{
			name: "unset",
			want: nil,
		},
		{
			name:   "explicit values",
			values: []string{"version", "service_delete"},
			want:   []string{"version", "service_delete"},
		},
		{
			name: "hard default fallback",
			// ARGUS_WEB_DISABLED_ROUTES lands in the hard defaults.
			hardDefaults: []string{"service_delete"},
			want:         []string{"service_delete"},
		},
		{
			name:         "empty, not nil, still falls back",
			values:       []string{},
			hardDefaults: []string{"service_delete"},
			want:         []string{"service_delete"},
		},
		{
			name:         "explicit values override the hard defaults",
			values:       []string{"version"},
			hardDefaults: []string{"service_delete"},
			want:         []string{"version"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{
				Web: WebSettings{DisabledRoutes: tc.values},
				HardDefaults: SettingsBase{
					Web: WebSettings{DisabledRoutes: tc.hardDefaults},
				},
			}

			got := settings.WebDisabledRoutes()

			if !slices.Equal(got, tc.want) {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestSettings_WebDemo(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		value       *WebSettingsDemo
		hardDefault *WebSettingsDemo
		want        *WebSettingsDemo
	}{
		{
			name: "unset",
			want: nil,
		},
		{
			name:  "resolved/explicit values",
			value: &WebSettingsDemo{Username: "demo", Password: "pass"},
			want:  &WebSettingsDemo{Username: "demo", Password: "pass"},
		},
		{
			name:        "resolved/hard default fallback",
			hardDefault: &WebSettingsDemo{Username: "demo", Password: "pass"},
			want:        &WebSettingsDemo{Username: "demo", Password: "pass"},
		},
		{
			name:        "resolved/explicit values override the hard defaults",
			value:       &WebSettingsDemo{Username: "demo", Password: "pass"},
			hardDefault: &WebSettingsDemo{Username: "other", Password: "other"},
			want:        &WebSettingsDemo{Username: "demo", Password: "pass"},
		},
		{
			name:        "resolved/layers merge per-field",
			value:       &WebSettingsDemo{Username: "demo"},
			hardDefault: &WebSettingsDemo{Username: "other", Password: "pass"},
			want:        &WebSettingsDemo{Username: "demo", Password: "pass"},
		},
		{
			name: "env/expanded",
			env: map[string]string{
				"TEST_SETTINGS__WEB_DEMO__PASSWORD": "pass",
			},
			value: &WebSettingsDemo{
				Username: "demo",
				Password: "${TEST_SETTINGS__WEB_DEMO__PASSWORD}",
			},
			want: &WebSettingsDemo{Username: "demo", Password: "pass"},
		},
		{
			name: "env/unset expands to its own literal",
			value: &WebSettingsDemo{
				Username: "demo",
				Password: "${TEST_SETTINGS__WEB_DEMO__UNSET}",
			},
			want: &WebSettingsDemo{
				Username: "demo",
				Password: "${TEST_SETTINGS__WEB_DEMO__UNSET}",
			},
		},
		{
			name:  "incomplete/empty",
			value: &WebSettingsDemo{},
			want:  nil,
		},
		{
			name:  "incomplete/username only",
			value: &WebSettingsDemo{Username: "demo"},
			want:  nil,
		},
		{
			name:  "incomplete/password only",
			value: &WebSettingsDemo{Password: "pass"},
			want:  nil,
		},
		{
			name:        "incomplete/hard defaults, username only",
			hardDefault: &WebSettingsDemo{Username: "demo"},
			want:        nil,
		},
		{
			name:        "incomplete/hard defaults, password only",
			hardDefault: &WebSettingsDemo{Password: "pass"},
			want:        nil,
		},
		{
			name: "incomplete/env var expands to empty",
			env: map[string]string{
				"TEST_SETTINGS__WEB_DEMO__EMPTY": "",
			},
			value: &WebSettingsDemo{
				Username: "demo",
				Password: "${TEST_SETTINGS__WEB_DEMO__EMPTY}",
			},
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			test.IsolateEnv(t, tc.env)

			settings := Settings{
				Web: WebSettings{Demo: tc.value},
				HardDefaults: SettingsBase{
					Web: WebSettings{
						Demo: tc.hardDefault,
					},
				},
			}

			got := settings.WebDemoCredentials()

			gotStr := test.StringifyPtr(got)
			wantStr := test.StringifyPtr(tc.want)
			if gotStr != wantStr {
				t.Fatalf("mismatch\ngot:  %s\nwant: %s", gotStr, wantStr)
			}
		})
	}
}

func TestSettings_WebDemoGroup(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		value       *WebSettingsDemo
		hardDefault *WebSettingsDemo
		want        string
	}{
		{
			name: "unset",
			want: "",
		},
		{
			name:  "resolved/explicit value",
			value: &WebSettingsDemo{Group: "demo"},
			want:  "demo",
		},
		{
			name:        "resolved/hard default fallback",
			hardDefault: &WebSettingsDemo{Group: "demo"},
			want:        "demo",
		},
		{
			name:        "resolved/explicit value overrides the hard default",
			value:       &WebSettingsDemo{Group: "demo"},
			hardDefault: &WebSettingsDemo{Group: "other"},
			want:        "demo",
		},
		{
			name:  "resolved/group independent of the prefill credentials",
			value: &WebSettingsDemo{Group: "demo"},
			want:  "demo",
		},
		{
			name: "env/expanded",
			env: map[string]string{
				"TEST_SETTINGS__WEB_DEMO__GROUP": "demo",
			},
			value: &WebSettingsDemo{Group: "${TEST_SETTINGS__WEB_DEMO__GROUP}"},
			want:  "demo",
		},
		{
			name:  "env/unset expands to its own literal",
			value: &WebSettingsDemo{Group: "${TEST_SETTINGS__WEB_DEMO__UNSET}"},
			want:  "${TEST_SETTINGS__WEB_DEMO__UNSET}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			test.IsolateEnv(t, tc.env)

			settings := Settings{
				Web: WebSettings{Demo: tc.value},
				HardDefaults: SettingsBase{
					Web: WebSettings{
						Demo: tc.hardDefault,
					},
				},
			}

			got := settings.WebDemoGroup()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestSettings_WebTrustedProxies(t *testing.T) {
	tests := []struct {
		name         string
		values       []string
		hardDefaults []string
		want         []string
	}{
		{
			name: "unset",
			want: []string{},
		},
		{
			name:   "explicit values, bare IPs become single-address ranges",
			values: []string{"10.0.0.1", "192.168.0.0/16", "::1"},
			want:   []string{"10.0.0.1/32", "192.168.0.0/16", "::1/128"},
		},
		{
			name:         "hard default fallback",
			hardDefaults: []string{"172.16.0.0/12"},
			want:         []string{"172.16.0.0/12"},
		},
		{
			name:   "unparseable entries skipped",
			values: []string{"10.0.0.1", "not-an-ip"},
			want:   []string{"10.0.0.1/32"},
		},
		{
			name:   "IPv4-mapped IPv6 CIDR rewritten as its IPv4 equivalent",
			values: []string{"::ffff:10.0.0.0/104"},
			want:   []string{"10.0.0.0/8"},
		},
		{
			name:   "IPv4-mapped IPv6 CIDR with a prefix shorter than 96 bits is kept as-is",
			values: []string{"::ffff:10.0.0.0/64"},
			want:   []string{"::ffff:10.0.0.0/64"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{}
			settings.Web.TrustedProxies = tc.values
			settings.HardDefaults.Web.TrustedProxies = tc.hardDefaults

			got := settings.WebTrustedProxies()

			gotStrs := make([]string, len(got))
			for i, prefix := range got {
				gotStrs[i] = prefix.String()
			}
			if !slices.Equal(gotStrs, tc.want) {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", gotStrs, tc.want)
			}
		})
	}
}

func TestSettings_WebBasicAuthUsernameHash(t *testing.T) {
	tests := []struct {
		name string
		had  Settings
		want [32]byte
	}{
		{
			name: "unset in every layer",
			want: util.GetHash(""),
		},
		{
			name: "set in config",
			had: Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Username: "user",
					},
				},
			},
			want: util.GetHash("user"),
		},
		{
			name: "set in flag",
			had: Settings{
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user",
						},
					},
				},
			},
			want: util.GetHash("user"),
		},
		{
			name: "set in hard defaults",
			had: Settings{
				HardDefaults: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "env-user",
						},
					},
				},
			},
			want: util.GetHash("env-user"),
		},
		{
			name: "config has only a password, username from the hard defaults",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass",
						},
					},
				},
				HardDefaults: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "env-user",
						},
					},
				},
			},
			want: util.GetHash("env-user"),
		},
		{
			name: "config has only a password, an empty username authenticates",
			had: Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Password: "pass",
					},
				},
			},
			want: util.GetHash(""),
		},
		{
			name: "set everywhere, use flag",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "config",
						},
					},
				},
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "flag",
						},
					},
				},
				HardDefaults: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "env-user",
						},
					},
				},
			},
			want: util.GetHash("flag"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.had.CheckValues(); err != nil {
				t.Fatalf("Settings.CheckValues() failed: %v", err)
			}
			if err := tc.had.FromFlags.CheckValues(); err != nil {
				t.Fatalf("Settings.FromFlags.CheckValues() failed: %v", err)
			}
			if err := tc.had.HardDefaults.CheckValues(); err != nil {
				t.Fatalf("Settings.HardDefaults.CheckValues() failed: %v", err)
			}

			got := tc.had.WebBasicAuthUsernameHash()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", util.FmtHash(got), util.FmtHash(tc.want))
			}
		})
	}
}

func TestSettings_WebBasicAuthPasswordHash(t *testing.T) {
	tests := []struct {
		name string
		had  Settings
		want [32]byte
	}{
		{
			name: "unset in every layer",
			had:  Settings{},
			want: util.GetHash(""),
		},
		{
			name: "set in config",
			had: Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Password: "pass",
					},
				},
			},
			want: util.GetHash("pass"),
		},
		{
			name: "set in flag",
			had: Settings{
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass",
						},
					},
				},
			},
			want: util.GetHash("pass"),
		},
		{
			name: "set in the hard defaults",
			had: Settings{
				HardDefaults: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "env-pass",
						},
					},
				},
			},
			want: util.GetHash("env-pass"),
		},
		{
			name: "flag has only a username, password from the config",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass",
						},
					},
				},
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user",
						},
					},
				},
			},
			want: util.GetHash("pass"),
		},
		{
			name: "config has only a username, an empty password authenticates",
			had: Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						Username: "user",
					},
				},
			},
			want: util.GetHash(""),
		},
		{
			name: "set everywhere, use flag",
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
				HardDefaults: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "env-pass",
						},
					},
				},
			},
			want: util.GetHash("flag"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.had.CheckValues(); err != nil {
				t.Fatalf("Settings.CheckValues() failed: %v", err)
			}
			if err := tc.had.FromFlags.CheckValues(); err != nil {
				t.Fatalf("Settings.FromFlags.CheckValues() failed: %v", err)
			}
			if err := tc.had.HardDefaults.CheckValues(); err != nil {
				t.Fatalf("Settings.HardDefaults.CheckValues() failed: %v", err)
			}

			got := tc.had.WebBasicAuthPasswordHash()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", util.FmtHash(got), util.FmtHash(tc.want))
			}
		})
	}
}
