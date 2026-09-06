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
	"slices"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

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
			name: "data.readonly",
			env: map[string]string{
				"ARGUS_DATA_READONLY": "true",
			},
			want: &Settings{
				Data: DataSettings{
					Readonly: new(true),
				},
			},
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
		},
		{
			name: "log.timestamps/invalid - not a bool",
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
				Web: WebSettings{
					ListenHost: "test",
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
				Web: WebSettings{
					ListenPort: "123",
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
				Web: WebSettings{
					CertFile: "cert.test",
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
				Web: WebSettings{
					KeyFile: "pkey.test",
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
				Web: WebSettings{
					RoutePrefix: "/prefix",
				},
			},
			ok: true,
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
			ok: true,
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
						Password: util.FmtHash(util.GetHash("pass")),
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
				t.Fatalf("%s error expected but not returned", prefix)
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
			name: "non-empty/DisabledRouted",
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
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
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
			name: "BasicAuth/empty",
			input: &WebSettings{
				BasicAuth: &WebSettingsBasicAuth{},
			},
			want: "{}\n",
			ok:   true,
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
			ok:               true,
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
			ok:               true,
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
			ok:               true,
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
			ok:               true,
		},
		{
			name: "Favicon/empty",
			input: &WebSettings{
				Favicon: &FaviconSettings{},
			},
			want: "{}\n",
			ok:   true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
		},
		{
			name: "Demo/empty",
			input: &WebSettings{
				Demo: &WebSettingsDemo{},
			},
			want: "{}\n",
			ok:   true,
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
			ok: true,
		},
		{
			name: "Web.CertFile, not found",
			input: &WebSettings{
				CertFile: "cert.pem",
			},
			want:     "cert_file: cert.pem\n",
			ok:       false,
			errRegex: `^cert_file: .*no such file.*$`,
		},
		{
			name: "Web.KeyFile, not found",
			input: &WebSettings{
				KeyFile: "privkey.pem",
			},
			want:     "pkey_file: privkey.pem\n",
			ok:       false,
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
			ok: false,
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
			ok: true,
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
			ok:       false,
			errRegex: `^trusted_proxies: "not-an-ip" <invalid>.*IP address or CIDR`,
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
				Log: LogSettings{
					Level: "INFO",
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

func TestSettings_CheckValues(t *testing.T) {
	// GIVEN: a Settings struct with some values set.
	tests := []struct {
		name                               string
		env                                map[string]string
		input                              *Settings
		want                               string
		wantUsernameHash, wantPasswordHash string
		wantIndent                         uint8
		ok                                 bool
		errRegex                           string
	}{
		{
			name: "indentation/0",
			input: &Settings{
				Indentation: 0,
			},
			want:       "{}\n",
			wantIndent: yaml.DefaultIndentSpaces,
		},
		{
			name: "indentation/1",
			input: &Settings{
				Indentation: 1,
			},
			want:       "{}\n",
			wantIndent: 1,
		},
		{
			name: "indentation/2",
			input: &Settings{
				Indentation: 2,
			},
			want:       "{}\n",
			wantIndent: 2,
		},
		{
			name: "indentation/3",
			input: &Settings{
				Indentation: 3,
			},
			want:       "{}\n",
			wantIndent: 3,
		},
		{
			name: "indentation/4",
			input: &Settings{
				Indentation: 4,
			},
			want:       "{}\n",
			wantIndent: 4,
		},
		{
			name: "BasicAuth/empty",
			input: &Settings{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{},
				},
			},
			want: "{}\n",
			ok:   true,
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
			ok:               true,
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
			ok: true,
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
			ok: true,
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
			ok: false,
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
			ok: true,
		},
		{
			name: "Route prefix/empty",
			input: &Settings{
				Web: WebSettings{
					RoutePrefix: "",
				},
			},
			want: "{}\n",
			ok:   true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
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
			ok: true,
		},
		{
			name: "Favicon/empty",
			input: &Settings{
				Web: WebSettings{
					Favicon: &FaviconSettings{},
				},
			},
			want: "{}\n",
			ok:   true,
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
			ok: true,
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
			ok: false,
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
			ok: false,
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
			test.SetEnv(t, tc.env)

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)

			prefix := fmt.Sprintf("%s\nSettings.CheckValues()", packageName)

			// THEN: the Settings returned stringifies as expected.
			if got := tc.input.String(""); got != tc.want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the Indentation is min yaml.Default
			if gotIndent := uint8(tc.input.Indentation); gotIndent != tc.wantIndent {
				t.Errorf(
					"%s Indentation mismatch\ngot:  %d\nwant: %d",
					prefix, gotIndent, tc.wantIndent,
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

func TestSettings_Default__BasicAuthFromFlags(t *testing.T) {
	// GIVEN: the web.basic-auth.username/password flags may or may not be provided.
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

	loadMu.Lock() // Protect flag vars.
	t.Cleanup(loadMu.Unlock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing flag vars.

			hadUsername, hadPassword := WebBasicAuthUsername, WebBasicAuthPassword
			WebBasicAuthUsername, WebBasicAuthPassword = tc.usernameFlag, tc.passwordFlag
			t.Cleanup(func() {
				WebBasicAuthUsername, WebBasicAuthPassword = hadUsername, hadPassword
			})

			settings := Settings{}

			// WHEN: Default is called.
			settings.Default()

			// THEN: FromFlags.Web.BasicAuth is only populated when a flag was provided.
			got := settings.FromFlags.Web.BasicAuth
			switch {
			case tc.want == nil:
				if got != nil {
					t.Errorf("%s\nSettings.Default() BasicAuth\ngot:  %v\nwant: nil",
						packageName, got)
				}
			case got == nil:
				t.Errorf("%s\nSettings.Default() BasicAuth\ngot:  nil\nwant: %v",
					packageName, tc.want)
			case got.Username != tc.want.Username || got.Password != tc.want.Password:
				t.Errorf("%s\nSettings.Default() BasicAuth\ngot:  %v\nwant: %v",
					packageName, got, tc.want)
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
	// GIVEN: different flags/env vars are set that may impact the result of .Default().
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
		Log: LogSettings{},
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

func TestSettings_WebDisabledRoutes(t *testing.T) {
	// GIVEN: a Settings struct with disabled routes from YAML and/or the
	// hard defaults (where ARGUS_WEB_DISABLED_ROUTES lands).
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
			name:         "hard default fallback",
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

			prefix := fmt.Sprintf("%s\nSettings.WebDisabledRoutes()", packageName)

			// WHEN: the accessor resolves the layered value.
			got := settings.WebDisabledRoutes()

			// THEN: it matches, so an env-set value is actually reachable.
			if !slices.Equal(got, tc.want) {
				t.Errorf(
					"%s mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestSettings_WebDemo(t *testing.T) {
	// GIVEN: a Settings struct with demo credentials from YAML and/or hard defaults.
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
			// t.Parallel() - Cannot run in parallel since we're setting env vars.
			test.SetEnv(t, tc.env)

			settings := Settings{
				Web: WebSettings{Demo: tc.value},
				HardDefaults: SettingsBase{
					Web: WebSettings{Demo: tc.hardDefault},
				},
			}

			// WHEN: the accessor resolves the layered value.
			got := settings.WebDemo()

			// THEN: the credentials resolve as expected.
			if got == nil || tc.want == nil {
				if got != tc.want {
					t.Fatalf(
						"%s\nSettings.WebDemo() mismatch\ngot:  %v\nwant: %v",
						packageName, got, tc.want,
					)
				}
				return
			}
			if *got != *tc.want {
				t.Errorf(
					"%s\nSettings.WebDemo() mismatch\ngot:  %v\nwant: %v",
					packageName, *got, *tc.want,
				)
			}
		})
	}
}

func TestSettings_WebTrustedProxies(t *testing.T) {
	// GIVEN: a Settings struct with some values set.
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

			// WHEN: WebTrustedProxies is called.
			got := settings.WebTrustedProxies()

			prefix := fmt.Sprintf("%s\nSettings.WebTrustedProxies()", packageName)

			// THEN: the resolved prefixes match expectations.
			gotStrs := make([]string, len(got))
			for i, prefix := range got {
				gotStrs[i] = prefix.String()
			}
			if !slices.Equal(gotStrs, tc.want) {
				t.Errorf(
					"%s mismatch\ngot:  %v\nwant: %v",
					prefix, gotStrs, tc.want,
				)
			}
		})
	}
}

func TestSettings_WebBasicAuthUsernameHash(t *testing.T) {
	// GIVEN: a Settings struct with a username in none, one, or several layers.
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

			_ = tc.had.CheckValues()
			_ = tc.had.FromFlags.CheckValues()
			_ = tc.had.HardDefaults.CheckValues()

			// WHEN: WebBasicAuthUsernameHash is called on it.
			got := tc.had.WebBasicAuthUsernameHash()

			// THEN: the layers resolve.
			if got != tc.want {
				t.Errorf(
					"%s\nWebBasicAuthUsernameHash() mismatch\ngot:  %s\nwant: %s",
					packageName, util.FmtHash(got), util.FmtHash(tc.want),
				)
			}
		})
	}
}

func TestSettings_WebBasicAuthPasswordHash(t *testing.T) {
	// GIVEN: a Settings struct with a password in none, one, or several layers.
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

			_ = tc.had.CheckValues()
			_ = tc.had.FromFlags.CheckValues()
			_ = tc.had.HardDefaults.CheckValues()

			// WHEN: WebBasicAuthPasswordHash is called on it.
			got := tc.had.WebBasicAuthPasswordHash()

			// THEN: the layers resolve.
			if got != tc.want {
				t.Errorf(
					"%s\nWebBasicAuthPasswordHash() mismatch\ngot:  %s\nwant: %s",
					packageName, util.FmtHash(got), util.FmtHash(tc.want),
				)
			}
		})
	}
}
