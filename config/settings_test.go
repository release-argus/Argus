// Copyright [2022] [Argus]
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
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestSettingsBase_CheckValues(t *testing.T) {
	// GIVEN a Settings struct with some values set
	tests := map[string]struct {
		env              map[string]string
		had              Settings
		want             Settings
		wantUsernameHash string
		wantPasswordHash string
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
		},
		"BasicAuth - hashed Username and str env Password": {
			env: map[string]string{
				"TESTSETTINGSBASE_CHECKVALUES_ONE": "ass"},
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: util.FmtHash(util.GetHash("user")),
							Password: "p${TESTSETTINGSBASE_CHECKVALUES_ONE}"}}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: util.FmtHash(util.GetHash("user")),
							Password: "p${TESTSETTINGSBASE_CHECKVALUES_ONE}"}}}},
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
			wantPasswordHash: util.FmtHash(util.GetHash("pass")),
		},
		"Route prefix - empty": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/")}}},
		},
		"Route prefix - no leading /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("test")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test")}}},
		},
		"Route prefix - leading /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test")}}},
		},
		"Route prefix - multiple leading /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("///test")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test")}}},
		},
		"Route prefix - trailing /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test/")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test")}}},
		},
		"Route prefix - multiple trailing /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test///")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/test")}}},
		},
		"Route prefix - only a /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/")}}},
		},
		"Route prefix - only multiple /": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("///")}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/")}}},
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
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN CheckValues is called on it
			tc.had.CheckValues()

			// THEN the Settings are converted/removed where necessary
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("want:\n%v\ngot:\n%v",
					wantStr, hadStr)
			}
			// AND the BasicAuth username and password are hashed (if they exist)
			if tc.want.Web.BasicAuth != nil {
				wantUsernameHash := util.FmtHash(util.GetHash(tc.want.Web.BasicAuth.Username))
				if tc.wantUsernameHash != "" {
					wantUsernameHash = tc.wantUsernameHash
				}
				if util.FmtHash(tc.had.Web.BasicAuth.UsernameHash) != wantUsernameHash {
					t.Errorf("want: %q\ngot:  %q",
						wantUsernameHash, tc.had.Web.BasicAuth.UsernameHash)
				}
				wantPasswordHash := util.FmtHash(util.GetHash(tc.want.Web.BasicAuth.Password))
				if tc.wantPasswordHash != "" {
					wantPasswordHash = tc.wantPasswordHash
				}
				if util.FmtHash(tc.had.Web.BasicAuth.PasswordHash) != wantPasswordHash {
					t.Errorf("want: %q\ngot:  %q",
						wantPasswordHash, tc.had.Web.BasicAuth.PasswordHash)
				}
			}
		})
	}
}

func TestSettings_NilUndefinedFlags(t *testing.T) {
	// GIVEN tests with flags set/unset
	var settings Settings
	tests := map[string]struct {
		flagSet   bool
		setStrTo  *string
		setBoolTo *bool
	}{
		"flag set": {
			flagSet:   true,
			setStrTo:  stringPtr("test"),
			setBoolTo: boolPtr(true)},
		"flag not set": {
			flagSet:   false,
			setStrTo:  stringPtr("foo"),
			setBoolTo: boolPtr(false)},
	}
	flagStr := "log.level"
	flagBool := "log.timestamps"
	flagset := map[string]bool{
		flagStr:  false,
		flagBool: false,
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN flags are set/unset and NilUndefinedFlags is called
			flagset[flagStr] = tc.flagSet
			flagset[flagBool] = tc.flagSet
			LogLevel = tc.setStrTo
			LogTimestamps = tc.setBoolTo
			settings.NilUndefinedFlags(&flagset)

			// THEN the flags are defined/undefined correctly
			gotStr := LogLevel
			if (tc.flagSet && gotStr == nil) ||
				(!tc.flagSet && gotStr != nil) {
				t.Errorf("%s %s:\nwant: %s\ngot:  %v",
					flagStr, name, *tc.setStrTo, util.EvalNilPtr(gotStr, "<nil>"))
			}
			gotBool := LogTimestamps
			if (tc.flagSet && gotBool == nil) ||
				(!tc.flagSet && gotBool != nil) {
				t.Errorf("%s %s:\nwant: %v\ngot:  %v",
					flagBool, name, *tc.setBoolTo, gotBool)
			}
		})
	}
}

func TestSettings_GetString(t *testing.T) {
	// GIVEN vars set in different at different priority levels in Settings
	settings := testSettings()
	tests := map[string]struct {
		flag         **string
		flagVal      *string
		env          map[string]string
		want         string
		wantSettings *Settings
		nilConfig    bool
		configPtr    **string
		getFunc      func() string
		getFuncPtr   func() *string
	}{
		"log.level hard default": {
			getFunc: settings.LogLevel,
			flag:    &LogLevel, want: "INFO",
			nilConfig: true,
			configPtr: &settings.Log.Level,
		},
		"log.level config": {
			getFunc: settings.LogLevel,
			flag:    &LogLevel, want: "DEBUG",
		},
		"log.level flag": {
			getFunc: settings.LogLevel,
			flag:    &LogLevel, flagVal: stringPtr("ERROR"),
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
			flag:    &DataDatabaseFile, flagVal: stringPtr("ERROR"),
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
			flag:    &WebListenHost, flagVal: stringPtr("127.0.0.1"),
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
			flag:    &WebListenPort, flagVal: stringPtr("54321"),
			want: "54321",
		},
		"web.cert-file hard default": {
			getFuncPtr: settings.WebCertFile,
			flag:       &WebCertFile, want: "<nil>",
			nilConfig: true, configPtr: &settings.Web.CertFile,
		},
		"web.cert-file config": {
			getFuncPtr: settings.WebCertFile,
			flag:       &WebCertFile, want: "../README.md",
		},
		"web.cert-file flag": {
			getFuncPtr: settings.WebCertFile,
			flag:       &WebCertFile, flagVal: stringPtr("settings_test.go"),
			want: "settings_test.go",
		},
		"web.pkey-file hard default": {
			getFuncPtr: settings.WebKeyFile,
			flag:       &WebPKeyFile, want: "<nil>",
			nilConfig: true, configPtr: &settings.Web.KeyFile,
		},
		"web.pkey-file config": {
			getFuncPtr: settings.WebKeyFile,
			flag:       &WebPKeyFile, want: "../LICENSE",
		},
		"web.pkey-file flag": {
			getFuncPtr: settings.WebKeyFile,
			flag:       &WebPKeyFile, flagVal: stringPtr("settings_test.go"),
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
			flag:    &WebRoutePrefix, flagVal: stringPtr("/flag"),
			want: "/flag",
		},
		"set from env": {
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "ERROR"},
			wantSettings: &Settings{
				HardDefaults: SettingsBase{
					Log: LogSettings{
						Level: stringPtr("ERROR")}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			settings = testSettings()
			if tc.flag != nil {
				*tc.flag = tc.flagVal
			}

			// WHEN SetDefaults is called on it
			settings.SetDefaults()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = nil
				defer func() { *tc.configPtr = had }()
			}

			// THEN the Service part is initialised to the defined defaults
			var got string
			if tc.getFunc != nil {
				got = tc.getFunc()
			}
			if tc.getFuncPtr != nil {
				got = util.EvalNilPtr(tc.getFuncPtr(), "<nil>")
			}
			if got != tc.want {
				t.Errorf("want: %s\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

func TestSettings_MapEnvToStruct(t *testing.T) {
	// GIVEN vars set for Settings vars
	tests := map[string]struct {
		env      map[string]string
		want     *Settings
		errRegex string
	}{
		"empty vars ignored": {
			env: map[string]string{
				"ARGUS_LOG_LEVEL": ""},
			want: &Settings{},
		},
		"log.level": {
			env: map[string]string{
				"ARGUS_LOG_LEVEL": "ERROR"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Level: stringPtr("ERROR")}}},
		},
		"log.timestamps": {
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "true"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Log: LogSettings{
						Timestamps: boolPtr(true)}}},
		},
		"log.timestamps - invalid, not a bool": {
			env: map[string]string{
				"ARGUS_LOG_TIMESTAMPS": "abc"},
			errRegex: `invalid bool for ARGUS_LOG_TIMESTAMPS: "abc"`,
		},
		"web.listen-host": {
			env: map[string]string{
				"ARGUS_WEB_LISTEN_HOST": "test"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						ListenHost: stringPtr("test")}}},
		},
		"web.listen-port": {
			env: map[string]string{
				"ARGUS_WEB_LISTEN_PORT": "123"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						ListenPort: stringPtr("123")}}},
		},
		"web.cert-file": {
			env: map[string]string{
				"ARGUS_WEB_CERT_FILE": "cert.test"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: stringPtr("cert.test")}}},
		},
		"web.pkey-file": {
			env: map[string]string{
				"ARGUS_WEB_PKEY_FILE": "pkey.test"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						KeyFile: stringPtr("pkey.test")}}},
		},
		"web.route-prefix": {
			env: map[string]string{
				"ARGUS_WEB_ROUTE_PREFIX": "prefix"},
			want: &Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						RoutePrefix: stringPtr("/prefix")}}},
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
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			settings := Settings{}
			// Catch fatal panics.
			defer func() {
				r := recover()
				if r != nil {
					if tc.errRegex == "" {
						t.Fatalf("unexpected panic: %v", r)
					}
					switch r.(type) {
					case string:
						if !regexp.MustCompile(tc.errRegex).MatchString(r.(string)) {
							t.Errorf("want error matching:\n%v\ngot:\n%v",
								tc.errRegex, t)
						}
					default:
						t.Fatalf("unexpected panic: %v", r)
					}
				}
			}()

			// WHEN MapEnvToStruct is called on it
			settings.MapEnvToStruct()

			// THEN any error is as expected
			if tc.errRegex != "" { // Expected a FATAL panic to be caught above
				t.Fatalf("expected an error matching %q, but got none", tc.errRegex)
			}
			// AND the settings are set to the appropriate env vars
			if settings.String("") != tc.want.String("") {
				t.Errorf("want:\n%v\ngot:\n%v",
					tc.want.String(""), settings.String(""))
			}
		})
	}
}

func TestSettings_GetBool(t *testing.T) {
	// GIVEN vars set in different at different priority levels in Settings
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
			flag:       &LogTimestamps, flagVal: boolPtr(true),
			want: "true"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			*tc.flag = tc.flagVal

			// WHEN SetDefaults is called on it
			settings.SetDefaults()
			if tc.nilConfig {
				had := *tc.configPtr
				*tc.configPtr = nil
				defer func() { *tc.configPtr = had }()
			}

			// THEN the Service part is initialised to the defined defaults
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
				t.Errorf("want: %s\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

func TestSettings_GetWebFile_NotExist(t *testing.T) {
	// GIVEN strings that point to files that don't exist
	settings := Settings{}
	tests := map[string]struct {
		file      string
		getFunc   func() *string
		changeVar **string
		want      string
	}{
		"hard default cert file": {
			getFunc: settings.WebCertFile},
		"config cert file": {
			file:      "cert_file_that_shouldnt_exist.hope",
			changeVar: &settings.Web.CertFile,
			getFunc:   settings.WebCertFile},
		"flag cert file": {
			file:      "cert_file_that_shouldnt_exist.hope",
			changeVar: &WebCertFile,
			getFunc:   settings.WebCertFile},
		"hard default pkey file": {
			getFunc: settings.WebKeyFile},
		"config pkey file": {
			file:      "pkey_file_that_shouldnt_exist.hope",
			changeVar: &settings.Web.KeyFile,
			getFunc:   settings.WebKeyFile},
		"flag pkey file": {
			file:      "pkey_file_that_shouldnt_exist.hope",
			changeVar: &WebPKeyFile,
			getFunc:   settings.WebKeyFile},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// Catch fatal panics.
			defer func() {
				r := recover()
				if r != nil &&
					!(strings.Contains(r.(string), "no such file or directory") ||
						strings.Contains(r.(string), "cannot find the file specified")) {
					t.Errorf("expected an error about the file not existing, not %s",
						r.(string))
				}
				tc.changeVar = nil
			}()

			// WHEN a get is called on files that don't exist
			if tc.file != "" {
				os.Remove(tc.file)
			}
			if tc.changeVar != nil {
				*tc.changeVar = &tc.file
			}
			got := tc.getFunc()

			// THEN this call will crash the program if a file is wanted
			if got != nil && *got != tc.file {
				t.Errorf("%q shouldn't exist, so this call should have been Fatal",
					tc.file)
			}
		})
	}
}

func TestSettings_WebBasicAuthUsernameHash(t *testing.T) {
	// GIVEN a Settings struct with some values set
	tests := map[string]struct {
		want string // The string that was hashed
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
			tc.had.CheckValues()
			tc.had.FromFlags.CheckValues()
			// HardDefaults.Web.BasicAuth will never be nil if Basic Auth is in use
			tc.had.HardDefaults = SettingsBase{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						UsernameHash: util.GetHash(""),
						PasswordHash: util.GetHash("")}}}

			// WHEN WebBasicAuthUsernameHash is called on it
			got := tc.had.WebBasicAuthUsernameHash()

			// THEN the hash is returned
			if got != want {
				t.Errorf("want: %s\ngot:  %s",
					want, got)
			}
		})
	}
}

func TestSettings_WebBasicAuthPasswordHash(t *testing.T) {
	// GIVEN a Settings struct with some values set
	tests := map[string]struct {
		want string // The string that was hashed
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
			tc.had.CheckValues()
			tc.had.FromFlags.CheckValues()
			// HardDefaults.Web.BasicAuth will never be nil if Basic Auth is in use
			tc.had.HardDefaults = SettingsBase{
				Web: WebSettings{
					BasicAuth: &WebSettingsBasicAuth{
						UsernameHash: util.GetHash(""),
						PasswordHash: util.GetHash("")}}}

			// WHEN WebBasicAuthPasswordHash is called on it
			got := tc.had.WebBasicAuthPasswordHash()

			// THEN the hash is returned
			if got != want {
				t.Errorf("want: %s\ngot:  %s",
					want, got)
			}
		})
	}
}

func TestWebSettings_CheckValues(t *testing.T) {
	// GIVEN a WebSettings struct with some values set
	tests := map[string]struct {
		env              map[string]string
		had              WebSettings
		want             WebSettings
		wantUsernameHash string
	}{
		"BasicAuth - empty": {
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{}},
			want: WebSettings{
				BasicAuth: nil},
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
		},
		"BasicAuth - Username and password from env vars": {
			env: map[string]string{
				"TESTWEBSETTINGS_CHECKVALUES_ONE": "user",
				"TESTWEBSETTINGS_CHECKVALUES_TWO": "pass"},
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "${TESTWEBSETTINGS_CHECKVALUES_ONE}",
					Password: "${TESTWEBSETTINGS_CHECKVALUES_TWO}"}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "${TESTWEBSETTINGS_CHECKVALUES_ONE}",
					Password: "${TESTWEBSETTINGS_CHECKVALUES_TWO}"}},
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
		},
		"BasicAuth - Username and password from env vars partial": {
			env: map[string]string{
				"TESTWEBSETTINGS_CHECKVALUES_THREE": "er",
				"TESTWEBSETTINGS_CHECKVALUES_FOUR":  "ss"},
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "us${TESTWEBSETTINGS_CHECKVALUES_THREE}",
					Password: "pa${TESTWEBSETTINGS_CHECKVALUES_FOUR}"}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: "us${TESTWEBSETTINGS_CHECKVALUES_THREE}",
					Password: "pa${TESTWEBSETTINGS_CHECKVALUES_FOUR}"}},
			wantUsernameHash: util.FmtHash(util.GetHash("user")),
		},
		"Favicon - empty": {
			had: WebSettings{
				Favicon: &FaviconSettings{}},
			want: WebSettings{
				Favicon: nil},
		},
		"Favicon - SVG": {
			had: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg"}},
			want: WebSettings{
				Favicon: &FaviconSettings{
					SVG: "https://example.com/favicon.svg"}},
		},
		"Favicon - PNG": {
			had: WebSettings{
				Favicon: &FaviconSettings{
					PNG: "https://example.com/favicon.png"}},
			want: WebSettings{
				Favicon: &FaviconSettings{
					PNG: "https://example.com/favicon.png"}},
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
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN CheckValues is called on it
			tc.had.CheckValues()

			// THEN the Settings are converted/removed where necessary
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("want:\n%v\ngot:\n%v",
					wantStr, hadStr)
			}
			if tc.wantUsernameHash != "" {
				got := util.FmtHash(tc.had.BasicAuth.UsernameHash)
				if got != tc.wantUsernameHash {
					t.Errorf("Username hash\nwant: %q\ngot:  %q",
						tc.wantUsernameHash, got)
				}
			}
		})
	}
}

func TestWebSettingsBasicAuth_CheckValues(t *testing.T) {
	// GIVEN a WebSettingsBasicAuth struct with some values set
	tests := map[string]struct {
		env              map[string]string
		had              WebSettingsBasicAuth
		want             WebSettingsBasicAuth
		wantUsernameHash string
		wantPasswordHash string
	}{
		"str Username": {
			had: WebSettingsBasicAuth{
				Username: "test"},
			want: WebSettingsBasicAuth{
				Username: "test",
				Password: util.FmtHash(util.GetHash(""))},
		},
		"str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Password: "just a password here"},
			want: WebSettingsBasicAuth{
				Username: "",
				Password: util.FmtHash(util.GetHash("just a password here"))},
		},
		"str Web.BasicAuth.Username and str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: "pass"},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
		},
		"str env Web.BasicAuth.Username and str env Web.BasicAuth.Password": {
			env: map[string]string{
				"TESTWEBSETTINGSBASICAUTH_CHECKVALUES_ONE": "user",
				"TESTWEBSETTINGSBASICAUTH_CHECKVALUES_TWO": "pass"},
			had: WebSettingsBasicAuth{
				Username: "${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_ONE}",
				Password: "${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_TWO}"},
			want: WebSettingsBasicAuth{
				Username: "${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_ONE}",
				Password: "${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_TWO}"},
		},
		"str env partial Web.BasicAuth.Username and str env partial Web.BasicAuth.Password": {
			env: map[string]string{
				"TESTWEBSETTINGSBASICAUTH_CHECKVALUES_THREE": "user",
				"TESTWEBSETTINGSBASICAUTH_CHECKVALUES_FOUR":  "pass"},
			had: WebSettingsBasicAuth{
				Username: "a${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_THREE}",
				Password: "b${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_FOUR}"},
			want: WebSettingsBasicAuth{
				Username: "a${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_THREE}",
				Password: "b${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_FOUR}"},
		},
		"str env undefined Web.BasicAuth.Username and str env undefined Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "a${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_UNDEFINED}",
				Password: "b${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_UNDEFINED}"},
			want: WebSettingsBasicAuth{
				Username: "a${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_UNDEFINED}",
				Password: util.FmtHash(util.GetHash("b${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_UNDEFINED}"))},
			wantUsernameHash: util.FmtHash(util.GetHash("a${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_UNDEFINED}")),
			wantPasswordHash: util.FmtHash(util.GetHash("b${TESTWEBSETTINGSBASICAUTH_CHECKVALUES_UNDEFINED}")),
		},
		"str Web.BasicAuth.Username and Web.BasicAuth.Password already hashed": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
		},
		"hashed Web.BasicAuth.Username and str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: "pass"},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
		},
		"hashed Web.BasicAuth.Username and hashed Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
			want: WebSettingsBasicAuth{
				Username: "user",
				Password: util.FmtHash(util.GetHash("pass"))},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN CheckValues is called on it
			tc.had.CheckValues()

			// THEN the Settings are converted/removed where necessary
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("want:\n%v\ngot:\n%v",
					wantStr, hadStr)
			}
			// AND the UsernameHash is calculated correctly
			want := util.FmtHash(util.GetHash(
				util.EvalEnvVars(tc.want.Username)))
			if tc.wantUsernameHash != "" {
				want = tc.wantUsernameHash
			}
			got := util.FmtHash(tc.had.UsernameHash)
			if got != want {
				t.Errorf("Username Hash\nwant: %s\ngot:  %s",
					want, got)
			}
			// AND the PasswordHash is calculated correctly
			want = util.FmtHash(util.GetHash(
				util.EvalEnvVars(tc.want.Password)))
			if tc.wantPasswordHash != "" {
				want = tc.wantPasswordHash
			}
			got = util.FmtHash(tc.had.PasswordHash)
			if got != want {
				t.Errorf("Password Hash\nwant: %s\ngot:  %s",
					want, got)
			}
		})
	}
}
