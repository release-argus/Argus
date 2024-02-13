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
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestSettings_NilUndefinedFlags(t *testing.T) {
	// GIVEN tests with flags set/unset
	var settings Settings
	tests := map[string]struct {
		flagSet bool
		setTo   *string
	}{
		"flag set": {
			flagSet: true, setTo: stringPtr("test")},
		"flag not set": {
			flagSet: false, setTo: stringPtr("foo")},
	}
	flagset := map[string]bool{
		"log.level": false,
	}
	flag := "log.level"
	var flagLock sync.Mutex

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN a flag is set/unset and NilUndefinedFlags is called
			flagLock.Lock()
			flagset[flag] = tc.flagSet
			LogLevel = tc.setTo
			settings.NilUndefinedFlags(&flagset)

			// THEN the flag is defined/undefined correctly
			got := LogLevel
			flagLock.Unlock()
			if (tc.flagSet && got == nil) ||
				(!tc.flagSet && got != nil) {
				t.Errorf("%s %s:\nwant: %s\ngot:  %v",
					flag, name, *tc.setTo, util.EvalNilPtr(got, "<nil>"))
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
			getFuncPtr: settings.DataDatabaseFile,
			flag:       &DataDatabaseFile, want: "data/argus.db",
			nilConfig: true, configPtr: &settings.Data.DatabaseFile,
		},
		"data.database-file config": {
			getFuncPtr: settings.DataDatabaseFile,
			flag:       &DataDatabaseFile, want: "somewhere.db",
		},
		"data.database-file flag": {
			getFuncPtr: settings.DataDatabaseFile,
			flag:       &DataDatabaseFile, flagVal: stringPtr("ERROR"),
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
						RoutePrefix: stringPtr("prefix")}}},
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

func TestWebSettingsBasicAuth_CheckValues(t *testing.T) {
	// GIVEN a WebSettingsBasicAuth struct with some values set
	tests := map[string]struct {
		had  WebSettingsBasicAuth
		want WebSettingsBasicAuth
	}{
		"str Username": {
			had: WebSettingsBasicAuth{
				Username: "test"},
			want: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("test"))),
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("")))},
		},
		"str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Password: "just a password here"},
			want: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte(""))),
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("just a password here")))},
		},
		"str Web.BasicAuth.Username and str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: "pass"},
			want: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))},
		},
		"str Web.BasicAuth.Username and Web.BasicAuth.Password already hashed": {
			had: WebSettingsBasicAuth{
				Username: "user",
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))},
			want: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))},
		},
		"hashed Web.BasicAuth.Username and str Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
				Password: "pass"},
			want: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))},
		},
		"hashed Web.BasicAuth.Username and hashed Web.BasicAuth.Password": {
			had: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))},
			want: WebSettingsBasicAuth{
				Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
				Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called on it
			tc.had.CheckValues()

			// THEN the Settings are converted/removed where necessary
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("want:\n%v\ngot:\n%v",
					wantStr, hadStr)
			}
		})
	}
}

func TestWebSettings_CheckValues(t *testing.T) {
	// GIVEN a WebSettings struct with some values set
	tests := map[string]struct {
		had  WebSettings
		want WebSettings
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
					Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
					Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))}},
		},
		"BasicAuth - hashed Username and str Password": {
			had: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
					Password: "pass"}},
			want: WebSettings{
				BasicAuth: &WebSettingsBasicAuth{
					Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
					Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))}},
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

			// WHEN CheckValues is called on it
			tc.had.CheckValues()

			// THEN the Settings are converted/removed where necessary
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("want:\n%v\ngot:\n%v",
					wantStr, hadStr)
			}
		})
	}
}

func TestSettings_CheckValues(t *testing.T) {
	// GIVEN a Settings struct with some values set
	tests := map[string]struct {
		had  Settings
		want Settings
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
		"BasicAuth - hashed Username and str Password": {
			had: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
							Password: "pass"}}}},
			want: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						BasicAuth: &WebSettingsBasicAuth{
							Username: fmt.Sprintf("h__%x", sha256.Sum256([]byte("user"))),
							Password: fmt.Sprintf("h__%x", sha256.Sum256([]byte("pass")))}}}},
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

			// WHEN CheckValues is called on it
			tc.had.CheckValues()

			// THEN the Settings are converted/removed where necessary
			hadStr := tc.had.String("")
			wantStr := tc.want.String("")
			if hadStr != wantStr {
				t.Errorf("want:\n%v\ngot:\n%v",
					wantStr, hadStr)
			}
		})
	}
}
