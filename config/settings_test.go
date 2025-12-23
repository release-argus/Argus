// Copyright [2025] [Argus]
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

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestSettings_String(t *testing.T) {
	// GIVEN a Settings struct.
	tests := map[string]struct {
		settings *Settings
		prefix   string
		want     string
	}{
		"nil settings": {
			settings: nil,
			prefix:   "",
			want:     "",
		},
		"empty settings": {
			settings: &Settings{},
			prefix:   "",
			want:     "{}\n",
		},
		"settings": {
			settings: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Level: "INFO",
					},
				},
			},
			prefix: "",
			want:   "log:\n  level: INFO\n",
		},
		"settings with prefix": {
			settings: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Level: "INFO",
					},
				},
			},
			prefix: "test_",
			want:   "test_log:\ntest_  level: INFO\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN String is called on it.
			got := tc.settings.String(tc.prefix)

			// THEN it's stringified as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestSettingsBase_CheckValues(t *testing.T) {
	// GIVEN a Settings struct with some values set.
	tests := map[string]struct {
		env                                map[string]string
		had, want                          Settings
		wantUsernameHash, wantPasswordHash string
		ok                                 bool
	}{
		"BasicAuth - empty": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{}}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: nil}}},
			ok: true,
		},
		"BasicAuth - hashed Username and str env Password": {
			env: map[string]string{
				"TEST_SETTINGS_BASE__CHECK_VALUES__ONE": "ass"},
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: util.FmtHash(util.GetHash("user")),
							Password: "p${TEST_SETTINGS_BASE__CHECK_VALUES__ONE}"}}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: util.FmtHash(util.GetHash("user")),
							Password: "p${TEST_SETTINGS_BASE__CHECK_VALUES__ONE}"}}}},
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
			ok:               true,
		},
		"Route prefix - empty": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: ""}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: ""}}},
			ok: true,
		},
		"Route prefix - no leading /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "test"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test"}}},
			ok: true,
		},
		"Route prefix - leading /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test"}}},
			ok: true,
		},
		"Route prefix - multiple leading /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "///test"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test"}}},
			ok: true,
		},
		"Route prefix - trailing /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test/"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test"}}},
			ok: true,
		},
		"Route prefix - multiple trailing /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test///"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/test"}}},
			ok: true,
		},
		"Route prefix - only a /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/"}}},
			ok: true,
		},
		"Route prefix - only multiple /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "///"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/"}}},
			ok: true,
		},
		"Favicon - empty": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						Favicon: &FaviconSettings{}}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						Favicon: nil}}},
			ok: true,
		},
		"Favicon - full": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						Favicon: &FaviconSettings{
							SVG: "https://example.com/favicon.svg",
							PNG: "https://example.com/favicon.png"}}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						Favicon: &FaviconSettings{
							SVG: "https://example.com/favicon.svg",
							PNG: "https://example.com/favicon.png"}}}},
			ok: true,
		},
		"Web.CertFile - not found": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.pem"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.pem"}}},
			ok: false,
		},
		"Web.KeyFile - not found": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						KeyFile: "privkey.pem"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						KeyFile: "privkey.pem"}}},
			ok: false,
		},
		"Web.CertFile + Web.KeyFile - both not found": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.pem",
						KeyFile:  "privkey.pem"}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.pem",
						KeyFile:  "privkey.pem"}}},
			ok: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}

			resultChannel := make(chan bool, 1)
			// WHEN CheckValues is called on it.
			go func() { resultChannel <- tc.had.CheckValues() }()

			// THEN the ok value is as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			// AND the Settings are converted/removed where necessary.
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("%s\nmismatch\nwant: %v\ngot:  %v",
					packageName, wantStr, hadStr)
			}
			// AND the expected error is returned.
			tc.errRegex = util.ValueOrValue(tc.errRegex, `^$`)
			select {
			case err := <-errChannel:
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, e)
				}
			default:
				t.Fatalf("%s\nerror not returned", packageName)
			}
			// AND the BasicAuth username and password are hashed (if they exist).
			if tc.want.Web.BasicAuth != nil {
				wantUsernameHash := util.FmtHash(util.GetHash(tc.want.Web.BasicAuth.Username))
				if tc.wantUsernameHash != "" {
					wantUsernameHash = tc.wantUsernameHash
				}
				if util.FmtHash(tc.had.Web.BasicAuth.UsernameHash) != wantUsernameHash {
					t.Errorf("%s\nUsername hash mismatch\nwant: %q\ngot:  %q",
						packageName, wantUsernameHash, tc.had.Web.BasicAuth.UsernameHash)
				}
				wantPasswordHash := util.FmtHash(util.GetHash(tc.want.Web.BasicAuth.Password))
				if tc.wantPasswordHash != "" {
					wantPasswordHash = tc.wantPasswordHash
				}
				if util.FmtHash(tc.had.Web.BasicAuth.PasswordHash) != wantPasswordHash {
					t.Errorf("%s\nPassword hash mismatch\nwant: %q\ngot:  %q",
						packageName, wantPasswordHash, tc.had.Web.BasicAuth.PasswordHash)
				}
			}
		})
	}
}

func TestSettings_NilUndefinedFlags(t *testing.T) {
	// GIVEN tests with flags set/unset.
	var settings Settings
	tests := map[string]struct {
		flagSet   bool
		setStrTo  *string
		setBoolTo *bool
	}{
		"flag set": {
			flagSet:   true,
			setStrTo:  test.StringPtr("test"),
			setBoolTo: test.BoolPtr(true)},
		"flag not set": {
			flagSet:   false,
			setStrTo:  test.StringPtr("foo"),
			setBoolTo: test.BoolPtr(false)},
	}
	flagStr := "log.level"
	flagBool := "log.timestamps"
	flagset := map[string]bool{
		flagStr:  false,
		flagBool: false,
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing some env vars.

			// WHEN flags are set/unset and NilUndefinedFlags is called.
			flagset[flagStr] = tc.flagSet
			flagset[flagBool] = tc.flagSet
			LogLevel = tc.setStrTo
			LogTimestamps = tc.setBoolTo
			settings.NilUndefinedFlags(&flagset)

			// THEN the flags are defined/undefined correctly.
			gotStr := LogLevel
			if (tc.flagSet && gotStr == nil) ||
				(!tc.flagSet && gotStr != nil) {
				t.Errorf("%s\nmismatch on %s - %s:\nwant: %s\ngot:  %v",
					packageName,
					flagStr, name,
					*tc.setStrTo, util.DereferenceOrValue(gotStr, "<nil>"))
			}
			gotBool := LogTimestamps
			if (tc.flagSet && gotBool == nil) ||
				(!tc.flagSet && gotBool != nil) {
				t.Errorf("%s\n%s:\nwant: %v\ngot:  %v",
					packageName, flagBool, *tc.setBoolTo, gotBool)
			}
		})
	}
}

func TestSettings_GetString(t *testing.T) {
	// GIVEN vars set in different at different priority levels in Settings.
	settings := testSettings()
	tests := map[string]struct {
		flag         **string
		flagVal      *string
		env          map[string]string
		want         string
		wantSettings *Settings
		nilConfig    bool
		configPtr    *string
		getFunc      func() string
		getFuncPtr   func() *string
	}{
		"log.level hard default": {
			getFunc: settings.LogLevel,
			flag:    &LogLevel, want: "DEBUG",
			nilConfig: true,
			configPtr: &settings.Log.Level,
		},
		"log.level config": {
			getFunc: settings.LogLevel,
			flag:    &LogLevel, want: "DEBUG",
		},
		"log.level flag": {
			getFunc: settings.LogLevel,
			flag:    &LogLevel, flagVal: test.StringPtr("ERROR"),
			want: "ERROR",
		},
		"data.database-file hard default": {
			getFunc: settings.DataDatabaseFile,
			flag:    &DataDatabaseFile, want: "data/argus.db",
			nilConfig: true, configPtr: &settings.Data.DatabaseFile,
		},
		"data.database-file config": {
			getFunc: settings.DataDatabaseFile,
			flag:    &DataDatabaseFile, want: "somewhere.db",
		},
		"data.database-file flag": {
			getFunc: settings.DataDatabaseFile,
			flag:    &DataDatabaseFile, flagVal: test.StringPtr("ERROR"),
			want: "ERROR",
		},
		"web.listen-host hard default": {
			getFunc: settings.WebListenHost,
			flag:    &WebListenHost, want: "0.0.0.0",
			nilConfig: true, configPtr: &settings.Web.ListenHost,
		},
		"web.listen-host config": {
			getFunc: settings.WebListenHost,
			flag:    &WebListenHost, want: "test",
		},
		"web.listen-host flag": {
			getFunc: settings.WebListenHost,
			flag:    &WebListenHost, flagVal: test.StringPtr("127.0.0.1"),
			want: "127.0.0.1",
		},
		"web.listen-port hard default": {
			getFunc: settings.WebListenPort,
			flag:    &WebListenPort, want: "8080",
			nilConfig: true, configPtr: &settings.Web.ListenPort,
		},
		"web.listen-port config": {
			getFunc: settings.WebListenPort,
			flag:    &WebListenPort, want: "123",
		},
		"web.listen-port flag": {
			getFunc: settings.WebListenPort,
			flag:    &WebListenPort, flagVal: test.StringPtr("54321"),
			want: "54321",
		},
		"web.cert-file hard default": {
			getFunc: settings.WebCertFile,
			flag:    &WebCertFile, want: "",
			nilConfig: true, configPtr: &settings.Web.CertFile,
		},
		"web.cert-file config": {
			getFunc: settings.WebCertFile,
			flag:    &WebCertFile, want: "../README.md",
		},
		"web.cert-file flag": {
			getFunc: settings.WebCertFile,
			flag:    &WebCertFile, flagVal: test.StringPtr("settings_test.go"),
			want: "settings_test.go",
		},
		"web.pkey-file hard default": {
			getFunc: settings.WebKeyFile,
			flag:    &WebPKeyFile, want: "",
			nilConfig: true, configPtr: &settings.Web.KeyFile,
		},
		"web.pkey-file config": {
			getFunc: settings.WebKeyFile,
			flag:    &WebPKeyFile, want: "../LICENSE",
		},
		"web.pkey-file flag": {
			getFunc: settings.WebKeyFile,
			flag:    &WebPKeyFile, flagVal: test.StringPtr("settings_test.go"),
			want: "settings_test.go",
		},
		"web.route-prefix hard default": {
			getFunc: settings.WebRoutePrefix,
			flag:    &WebRoutePrefix, want: "/",
			nilConfig: true, configPtr: &settings.Web.RoutePrefix,
		},
		"web.route-prefix config": {
			getFunc: settings.WebRoutePrefix,
			flag:    &WebRoutePrefix, want: "/something",
		},
		"web.route-prefix flag": {
			getFunc: settings.WebRoutePrefix,
			flag:    &WebRoutePrefix, flagVal: test.StringPtr("/flag"),
			want: "/flag",
		},
		"set from env": {
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "ERROR"},
			wantSettings: &Settings{
				HardDefaults: SettingsBase{
					Log: LogSettings{
						Level: "ERROR"}}},
		},
	}

	loadMutex.Lock() // Protect flag env vars.
	t.Cleanup(loadMutex.Unlock)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing some env vars.
			releaseStdout := test.CaptureLog(logutil.Log)
			defer releaseStdout()

			settings = testSettings()
			if tc.flag != nil {
				*tc.flag = tc.flagVal
			}

			// WHEN Default is called on it.
			settings.Default()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = ""
				t.Cleanup(func() { *tc.configPtr = had })
			}

			// THEN the Service part is initialised to the defined defaults.
			var got string
			if tc.getFunc != nil {
				got = tc.getFunc()
			}
			if tc.getFuncPtr != nil {
				got = util.DereferenceOrValue(tc.getFuncPtr(), "<nil>")
			}
			if got != tc.want {
				t.Errorf("%s\nmismatch\nwant: %s\ngot:  %v",
					packageName, tc.want, got)
			}
		})
	}
}

func TestSettings_MapEnvToStruct(t *testing.T) {
	// Unset ARGUS_LOG_LEVEL.
	logLevel := os.Getenv("ARGUS_LOG_LEVEL")
	_ = os.Unsetenv("ARGUS_LOG_LEVEL")
	t.Cleanup(func() { _ = os.Setenv("ARGUS_LOG_LEVEL", logLevel) })
	// GIVEN vars set for Settings vars.
	tests := map[string]struct {
		env         map[string]string
		want        *Settings
		stdoutRegex string
		ok          bool
	}{
		"empty vars ignored": {
			env: map[string]string{
				"ARGUS_LOG_LEVEL": ""},
			want: &Settings{},
			ok:   true,
		},
		"log.level": {
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "ERROR"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Level: "ERROR"}}},
			ok: true,
		},
		"log.timestamps": {
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "true"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Timestamps: test.BoolPtr(true)}}},
			ok: true,
		},
		"log.timestamps - invalid, not a bool": {
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "abc"},
			want:        &Settings{},
			stdoutRegex: `ARGUS_LOG_TIMESTAMPS: "abc" <invalid>`,
			ok:          false,
		},
		"web.listen-host": {
			env: map[string]string{
				"ARGUS_WEB_LISTEN_HOST": "test"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						ListenHost: "test"}}},
			ok: true,
		},
		"web.listen-port": {
			env: map[string]string{
				"ARGUS_WEB_LISTEN_PORT": "123"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						ListenPort: "123"}}},
			ok: true,
		},
		"web.cert-file": {
			env: map[string]string{
				"ARGUS_WEB_CERT_FILE": "cert.test"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "cert.test"}}},
			ok: false,
		},
		"web.pkey-file": {
			env: map[string]string{
				"ARGUS_WEB_PKEY_FILE": "pkey.test"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						KeyFile: "pkey.test"}}},
			ok: false,
		},
		"web.route-prefix": {
			env: map[string]string{
				"ARGUS_WEB_ROUTE_PREFIX": "prefix"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: "/prefix"}}},
			ok: true,
		},
		"web.basic_auth": {
			env: map[string]string{
				"ARGUS_WEB_BASIC_AUTH_USERNAME": "user",
				"ARGUS_WEB_BASIC_AUTH_PASSWORD": "pass"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user",
							Password: util.FmtHash(util.GetHash("pass"))}}}},
			ok: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}
			settings := Settings{}

			resultChannel := make(chan bool, 1)
			// WHEN MapEnvToStruct is called on it.
			resultChannel <- settings.MapEnvToStruct()

			// THEN the ok value is as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), releaseStdout); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			// AND any error is as expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
					packageName, tc.stdoutRegex, stdout)
			}
			// AND the settings are set to the appropriate env vars.
			if settings.String("") != tc.want.String("") {
				t.Errorf("%s\nstruct mismatch\nwant: %v\ngot:  %v",
					packageName, tc.want.String(""), settings.String(""))
			}
		})
	}
}

func TestSettings_Default(t *testing.T) {
	// GIVEN a set of env vars.
	tests := map[string]struct {
		env         map[string]string
		stdoutRegex string
		ok          bool
	}{
		"no env vars": {
			env:         map[string]string{},
			stdoutRegex: `^$`,
			ok:          true,
		},
		"valid env var": {
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "false"},
			stdoutRegex: `^$`,
			ok:          true,
		},
		"invalid env var": {
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "abc"},
			stdoutRegex: `^FATAL.*environment variable.*incorrect.*\s.*ARGUS_LOG_TIMESTAMPS.*\s$`,
			ok:          false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}
			settings := Settings{}

			resultChannel := make(chan bool, 1)
			// WHEN Default is called.
			resultChannel <- settings.Default()

			// THEN the ok value is as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), releaseStdout); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			// AND any error is as expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
					packageName, tc.stdoutRegex, stdout)
			}
		})
	}
}

func TestSettings_GetBool(t *testing.T) {
	// GIVEN vars set in different at different priority levels in Settings.
	settings := testSettings()
	tests := map[string]struct {
		flag       **bool
		flagVal    *bool
		want       string
		nilConfig  bool
		configPtr  **bool
		getFunc    func() bool
		getFuncPtr func() *bool
	}{
		"log.timestamps hard default": {
			getFuncPtr: settings.LogTimestamps,
			flag:       &LogTimestamps,
			want:       "false",
			nilConfig:  true, configPtr: &settings.Log.Timestamps},
		"log.timestamps config": {
			getFuncPtr: settings.LogTimestamps,
			flag:       &LogTimestamps,
			want:       "true"},
		"log.timestamps flag": {
			getFuncPtr: settings.LogTimestamps,
			flag:       &LogTimestamps, flagVal: test.BoolPtr(true),
			want: "true"},
	}

	loadMutex.Lock() // Protect flag env vars.
	t.Cleanup(loadMutex.Unlock)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing flag vars.

			*tc.flag = tc.flagVal

			// WHEN Default is called on it.
			settings.Default()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = nil
				t.Cleanup(func() { *tc.configPtr = had })
			}

			// THEN the Service part is initialised to the defined defaults.
			var got string
			if tc.getFunc != nil {
				got = fmt.Sprint(tc.getFunc())
			}
			if tc.getFuncPtr != nil {
				ptr := tc.getFuncPtr()
				got = "<nil>"
				if ptr != nil {
					got = fmt.Sprint(*tc.getFuncPtr())
				}
			}
			if got != tc.want {
				t.Errorf("%s\nwant: %s\ngot:  %v",
					packageName, tc.want, got)
			}
		})
	}
}

func TestSettings_GetWebFile_NotExist(t *testing.T) {
	settings := Settings{
		SettingsBase: SettingsBase{
			Log: LogSettings{}},
		FromFlags: SettingsBase{
			Log: LogSettings{}},
		HardDefaults: SettingsBase{
			Log: LogSettings{}}}

	// GIVEN different target vars, and their respective 'get' functions.
	tests := map[string]struct {
		getFunc   func() string
		changeVar any
	}{
		"hard default cert file": {
			changeVar: &settings.Web.CertFile,
			getFunc:   settings.WebCertFile},
		"config cert file": {
			changeVar: &settings.Web.CertFile,
			getFunc:   settings.WebCertFile},
		"flag cert file": {
			changeVar: &settings.FromFlags.Web.CertFile,
			getFunc:   settings.WebCertFile},
		"hard default pkey file": {
			changeVar: &settings.Web.KeyFile,
			getFunc:   settings.WebKeyFile},
		"config pkey file": {
			changeVar: &settings.Web.KeyFile,
			getFunc:   settings.WebKeyFile},
		"flag pkey file": {
			changeVar: &settings.FromFlags.Web.KeyFile,
			getFunc:   settings.WebKeyFile},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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
			// WHEN a get is called with no file path set.
			got := tc.getFunc()

			// THEN the empty string is returned.
			if got != file {
				t.Errorf("%s\nempty string\nwant: %q\ngot:  %q",
					packageName, file, got)
			}

			//
			// Test 2: file path.
			//
			file = fmt.Sprintf("test_%s.pem",
				strings.ReplaceAll(strings.ToLower(name), " ", "_"))
			if ptr, ok := tc.changeVar.(*string); ok {
				*ptr = file
			} else if ptrPtr, ok := tc.changeVar.(**string); ok {
				*ptrPtr = &file
			}
			// WHEN a get is called with a file path set.
			got = tc.getFunc()

			// THEN the file path is returned.
			if got != file {
				t.Errorf("%s\nfile path\nwant: %q\ngot:  %q",
					packageName, file, got)
			}
		})
	}
}

func TestSettings_WebBasicAuthUsernameHash(t *testing.T) {
	// GIVEN a Settings struct with some values set.
	tests := map[string]struct {
		want string // The string that was hashed.
		had  Settings
	}{
		"empty": {
			want: "",
		},
		"set in config": {
			want: "user",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user"}}}},
		},
		"set in flag": {
			want: "user",
			had: Settings{
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: "user"}}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := util.GetHash(tc.want)
			_ = tc.had.CheckValues("")
			_ = tc.had.FromFlags.CheckValues("")
			// HardDefaults.Web.BasicAuth will never be nil if Basic Auth is in use.
			tc.had.HardDefaults = SettingsBase{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						UsernameHash: util.GetHash(""),
						PasswordHash: util.GetHash("")}}}

			// WHEN WebBasicAuthUsernameHash is called on it.
			got := tc.had.WebBasicAuthUsernameHash()

			// THEN the hash is returned.
			if got != want {
				t.Errorf("%s\nwant: %s\ngot:  %s",
					packageName, want, got)
			}
		})
	}
}

func TestSettings_WebBasicAuthPasswordHash(t *testing.T) {
	// GIVEN a Settings struct with some values set.
	tests := map[string]struct {
		want string // The string that was hashed.
		had  Settings
	}{
		"empty": {
			want: "",
		},
		"set in config": {
			want: "pass",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass"}}}},
		},
		"set in flag": {
			want: "pass",
			had: Settings{
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "pass"}}}},
		},
		"set everywhere, use flag": {
			want: "flag",
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "config"}}},
				FromFlags: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Password: "flag"}}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := util.GetHash(tc.want)
			_ = tc.had.CheckValues("")
			_ = tc.had.FromFlags.CheckValues("")
			// HardDefaults.Web.BasicAuth will never be nil if Basic Auth is in use.
			tc.had.HardDefaults = SettingsBase{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						UsernameHash: util.GetHash(""),
						PasswordHash: util.GetHash("")}}}

			// WHEN WebBasicAuthPasswordHash is called on it.
			got := tc.had.WebBasicAuthPasswordHash()

			// THEN the hash is returned.
			if got != want {
				t.Errorf("%s\nwant: %s\ngot:  %s",
					packageName, want, got)
			}
		})
	}
}

func TestWebSettings_String(t *testing.T) {
	// GIVEN a WebSettings struct.
	tests := map[string]struct {
		webSettings *WebSettings
		prefix      string
		want        string
	}{
		"nil webSettings": {
			webSettings: nil,
			prefix:      "",
			want:        "",
		},
		"empty webSettings": {
			webSettings: &WebSettings{},
			prefix:      "",
			want:        "{}\n",
		},
		"webSettings with values": {
			webSettings: &WebSettings{
				ListenHost: "0.0.0.0",
				ListenPort: "8080",
			},
			prefix: "",
			want: test.TrimYAML(`
				listen_host: 0.0.0.0
				listen_port: "8080"
			`),
		},
		"webSettings with prefix": {
			webSettings: &WebSettings{
				ListenHost: "0.0.0.0",
				ListenPort: "8080",
			},
			prefix: "test_",
			want: test.TrimYAML(`
				test_listen_host: 0.0.0.0
				test_listen_port: "8080"
			`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN String is called on it.
			got := tc.webSettings.String(tc.prefix)

			// THEN it's stringified as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebSettings_CheckValues(t *testing.T) {
	// GIVEN a WebSettings struct with some values set.
	tests := map[string]struct {
		env              map[string]string
		had, want        WebSettings
		wantUsernameHash string
		ok               bool
	}{
		"BasicAuth - empty": {
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{}},
			want: WebSettings{
				BasicAuth: nil},
			ok: true,
		},
		"BasicAuth - str Username and Password already hashed": {
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: util.FmtHash(util.GetHash("pass"))}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: util.FmtHash(util.GetHash("pass"))}},
			ok: true,
		},
		"BasicAuth - hashed Username and str Password": {
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: "pass"}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "user",
					Password: util.FmtHash(util.GetHash("pass"))}},
			ok: true,
		},
		"BasicAuth - Username and password from env vars": {
			env: map[string]string{
				"TEST_WEB_SETTINGS__CHECK_VALUES__ONE": "user",
				"TEST_WEB_SETTINGS__CHECK_VALUES__TWO": "pass"},
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "${TEST_WEB_SETTINGS__CHECK_VALUES__ONE}",
					Password: "${TEST_WEB_SETTINGS__CHECK_VALUES__TWO}"}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "${TEST_WEB_SETTINGS__CHECK_VALUES__ONE}",
					Password: "${TEST_WEB_SETTINGS__CHECK_VALUES__TWO}"}},
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			ok:               true,
		},
		"BasicAuth - Username and password from env vars partial": {
			env: map[string]string{
				"TEST_WEB_SETTINGS__CHECK_VALUES__THREE": "er",
				"TEST_WEB_SETTINGS__CHECK_VALUES__FOUR":  "ss"},
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "us${TEST_WEB_SETTINGS__CHECK_VALUES__THREE}",
					Password: "pa${TEST_WEB_SETTINGS__CHECK_VALUES__FOUR}"}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "us${TEST_WEB_SETTINGS__CHECK_VALUES__THREE}",
					Password: "pa${TEST_WEB_SETTINGS__CHECK_VALUES__FOUR}"}},
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			ok:               true,
		},
		"Favicon - empty": {
			had: WebSettings{
				Favicon: &FaviconSettings{}},
			want: WebSettings{
				Favicon: nil},
			ok: true,
		},
		"Favicon - SVG": {
			had: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg"}},
			want: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg"}},
			ok: true,
		},
		"Favicon - PNG": {
			had: WebSettings{
				Favicon: &FaviconSettings{
					PNG: "https://example.com/favicon.png"}},
			want: WebSettings{
				Favicon: &FaviconSettings{
					PNG: "https://example.com/favicon.png"}},
			ok: true,
		},
		"Favicon - Full": {
			had: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg",
					PNG: "https://example.com/favicon.png"}},
			want: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg",
					PNG: "https://example.com/favicon.png"}},
			ok: true,
		},
		"Web.CertFile - not found": {
			had: WebSettings{
				CertFile: "cert.pem"},
			want: WebSettings{
				CertFile: "cert.pem"},
			ok: false,
		},
		"Web.KeyFile - not found": {
			had: WebSettings{
				KeyFile: "privkey.pem"},
			want: WebSettings{
				KeyFile: "privkey.pem"},
			ok: false,
		},
		"Web.CertFile + Web.KeyFile - both not found": {
			had: WebSettings{
				CertFile: "cert.pem",
				KeyFile:  "privkey.pem"},
			want: WebSettings{
				CertFile: "cert.pem",
				KeyFile:  "privkey.pem"},
			ok: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}

			resultChannel := make(chan bool, 1)
			// WHEN CheckValues is called on it.
			go func() { resultChannel <- tc.had.CheckValues() }()

			// THEN the ok value is as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			// AND the Settings are converted/removed where necessary.
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("%s\nwant:\n%v\ngot:\n%v",
					packageName, wantStr, hadStr)
			}
			if tc.wantUsernameHash != "" {
				got := util.FmtHash(tc.had.BasicAuth.UsernameHash)
				if got != tc.wantUsernameHash {
					t.Errorf("%s\nUsername hash mismatch\nwant: %q\ngot:  %q",
						packageName, tc.wantUsernameHash, got)
				}
			}
		})
	}
}

func TestWebSettingsBasicAuth_String(t *testing.T) {
	// GIVEN a WebSettingsBasicAuth struct.
	tests := map[string]struct {
		auth   *WebSettingsBasicAuth
		prefix string
		want   string
	}{
		"nil auth": {
			auth:   nil,
			prefix: "",
			want:   "",
		},
		"empty auth": {
			auth:   &WebSettingsBasicAuth{},
			prefix: "",
			want:   "{}\n",
		},
		"auth with values": {
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
		"auth with prefix": {
			auth: &WebSettingsBasicAuth{
				Username: "user",
				Password: "pass",
			},
			prefix: "test_",
			want: test.TrimYAML(`
				test_username: user
				test_password: pass
			`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN String is called on it.
			got := tc.auth.String(tc.prefix)

			// THEN it's stringified as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebSettingsBasicAuth_CheckValues(t *testing.T) {
	// GIVEN a WebSettingsBasicAuth struct with some values set.
	tests := map[string]struct {
		env                                map[string]string
		had, want                          WebSettingsBasicAuth
		wantUsernameHash, wantPasswordHash string
		ok                                 bool
	}{
		"str Username": {
			had: WebSettingsBasicAuth{
				Username: "test"},
			want: WebSettingsBasicAuth{
				Username: "test",
				Password: util.FmtHash(util.GetHash(""))},
			ok: true,
		},
		"str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Password: "just a password here"},
			want: WebSettingsBasicAuth{
				Username: "",
				Password: util.FmtHash(util.GetHash("just a password here"))},
			ok: true,
		},
		"str Web.BasicAuth.Username and str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: "pass"},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			ok: true,
		},
		"str env Web.BasicAuth.Username and str env Web.BasicAuth.Password": {
			env: map[string]string{
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__ONE": "user",
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__TWO": "pass"},
			had: WebSettingsBasicAuth{
				Username: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__ONE}",
				Password: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__TWO}"},
			want: WebSettingsBasicAuth{
				Username: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__ONE}",
				Password: "${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__TWO}"},
			ok: true,
		},
		"str env partial Web.BasicAuth.Username and str env partial Web.BasicAuth.Password": {
			env: map[string]string{
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__THREE": "user",
				"TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__FOUR":  "pass"},
			had: WebSettingsBasicAuth{
				Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__THREE}",
				Password: "b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__FOUR}"},
			want: WebSettingsBasicAuth{
				Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__THREE}",
				Password: "b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__FOUR}"},
			ok: true,
		},
		"str env undefined Web.BasicAuth.Username and str env undefined Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}",
				Password: "b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}"},
			want: WebSettingsBasicAuth{
				Username: "a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}",
				Password: util.FmtHash(util.GetHash("b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}"))},
			wantUsernameHash: util.FmtHash(util.GetHash("a${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}")),
			wantPasswordHash: util.FmtHash(util.GetHash("b${TEST_WEB_SETTINGS_BASIC_AUTH__CHECK_VALUES__UNDEFINED}")),
			ok:               true,
		},
		"str Web.BasicAuth.Username and Web.BasicAuth.Password already hashed": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			ok: true,
		},
		"hashed Web.BasicAuth.Username and str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: "pass"},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			ok: true,
		},
		"hashed Web.BasicAuth.Username and hashed Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			ok: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}

			resultChannel := make(chan bool, 1)
			// WHEN CheckValues is called on it.
			resultChannel <- tc.had.CheckValues()

			// THEN the ok value is as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			// THEN the Settings are converted/removed where necessary.
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("%s\nString() mismatch\nwant: %v\ngot:  %v",
					packageName, wantStr, hadStr)
			}
			// AND the UsernameHash is calculated correctly.
			want := util.FmtHash(util.GetHash(
				util.EvalEnvVars(tc.want.Username)))
			if tc.wantUsernameHash != "" {
				want = tc.wantUsernameHash
			}
			got := util.FmtHash(tc.had.UsernameHash)
			if got != want {
				t.Errorf("%s\nUsername Hash mismatch\nwant: %s\ngot:  %s",
					packageName, want, got)
			}
			// AND the PasswordHash is calculated correctly.
			want = util.FmtHash(util.GetHash(
				util.EvalEnvVars(tc.want.Password)))
			if tc.wantPasswordHash != "" {
				want = tc.wantPasswordHash
			}
			got = util.FmtHash(tc.had.PasswordHash)
			if got != want {
				t.Errorf("%s\nPassword Hash mismatch\nwant: %s\ngot:  %s",
					packageName, want, got)
			}
		})
	}
}
