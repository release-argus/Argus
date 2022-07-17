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
	"strings"
	"testing"

	"github.com/release-argus/Argus/utils"
)

func TestNilUndefinedFlags(t *testing.T) {
	// GIVEN tests with flags set/unset
	var settings Settings
	tests := map[string]struct {
		flagSet bool
		setTo   *string
	}{
		"flag set":     {flagSet: true, setTo: stringPtr("test")},
		"flag not set": {flagSet: false, setTo: stringPtr("test")},
	}
	flagset := map[string]bool{
		"log.level": false,
	}
	flag := "log.level"

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN a flag is set/unset and NilUndefinedFlags is called
			flagset[flag] = tc.flagSet
			LogLevel = tc.setTo
			settings.NilUndefinedFlags(&flagset)

			// THEN the flag is defined/undefined correctly
			if (tc.flagSet && LogLevel == nil) ||
				(!tc.flagSet && LogLevel != nil) {
				t.Errorf("%s %s:\nwant: %s\ngot:  %v",
					flag, name, *tc.setTo, utils.EvalNilPtr(LogLevel, "<nil>"))
			}
		})
	}
}

func TestSettingsGetString(t *testing.T) {
	// GIVEN vars set in different at different priority levels in Settings
	settings := testSettings()
	tests := map[string]struct {
		flag       **string
		flagVal    *string
		want       string
		nilConfig  bool
		configPtr  **string
		getFunc    func() string
		getFuncPtr func() *string
	}{
		"log.level hard default":          {getFunc: settings.GetLogLevel, flag: &LogLevel, want: "INFO", nilConfig: true, configPtr: &settings.Log.Level},
		"log.level config":                {getFunc: settings.GetLogLevel, flag: &LogLevel, want: "DEBUG"},
		"log.level flag":                  {getFunc: settings.GetLogLevel, flag: &LogLevel, flagVal: stringPtr("ERROR"), want: "ERROR"},
		"data.database-file hard default": {getFuncPtr: settings.GetDataDatabaseFile, flag: &DataDatabaseFile, want: "data/argus.db", nilConfig: true, configPtr: &settings.Data.DatabaseFile},
		"data.database-file config":       {getFuncPtr: settings.GetDataDatabaseFile, flag: &DataDatabaseFile, want: "somewhere.db"},
		"data.database-file flag":         {getFuncPtr: settings.GetDataDatabaseFile, flag: &DataDatabaseFile, flagVal: stringPtr("ERROR"), want: "ERROR"},
		"web.listen-host hard default":    {getFunc: settings.GetWebListenHost, flag: &WebListenHost, want: "0.0.0.0", nilConfig: true, configPtr: &settings.Web.ListenHost},
		"web.listen-host config":          {getFunc: settings.GetWebListenHost, flag: &WebListenHost, want: "test"},
		"web.listen-host flag":            {getFunc: settings.GetWebListenHost, flag: &WebListenHost, flagVal: stringPtr("127.0.0.1"), want: "127.0.0.1"},
		"web.listen-port hard default":    {getFunc: settings.GetWebListenPort, flag: &WebListenPort, want: "8080", nilConfig: true, configPtr: &settings.Web.ListenPort},
		"web.listen-port config":          {getFunc: settings.GetWebListenPort, flag: &WebListenPort, want: "123"},
		"web.listen-port flag":            {getFunc: settings.GetWebListenPort, flag: &WebListenPort, flagVal: stringPtr("54321"), want: "54321"},
		"web.cert-file hard default":      {getFuncPtr: settings.GetWebCertFile, flag: &WebCertFile, want: "<nil>", nilConfig: true, configPtr: &settings.Web.CertFile},
		"web.cert-file config":            {getFuncPtr: settings.GetWebCertFile, flag: &WebCertFile, want: "../test/ordering_0.yml"},
		"web.cert-file flag":              {getFuncPtr: settings.GetWebCertFile, flag: &WebCertFile, flagVal: stringPtr("settings_test.go"), want: "settings_test.go"},
		"web.pkey-file hard default":      {getFuncPtr: settings.GetWebKeyFile, flag: &WebPKeyFile, want: "<nil>", nilConfig: true, configPtr: &settings.Web.KeyFile},
		"web.pkey-file config":            {getFuncPtr: settings.GetWebKeyFile, flag: &WebPKeyFile, want: "../test/ordering_1.yml"},
		"web.pkey-file flag":              {getFuncPtr: settings.GetWebKeyFile, flag: &WebPKeyFile, flagVal: stringPtr("settings_test.go"), want: "settings_test.go"},
		"web.route-prefix hard default":   {getFunc: settings.GetWebRoutePrefix, flag: &WebRoutePrefix, want: "/", nilConfig: true, configPtr: &settings.Web.RoutePrefix},
		"web.route-prefix config":         {getFunc: settings.GetWebRoutePrefix, flag: &WebRoutePrefix, want: "/something"},
		"web.route-prefix flag":           {getFunc: settings.GetWebRoutePrefix, flag: &WebRoutePrefix, flagVal: stringPtr("/flag"), want: "/flag"},
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
				got = tc.getFunc()
			}
			if tc.getFuncPtr != nil {
				got = utils.EvalNilPtr(tc.getFuncPtr(), "<nil>")
			}
			if got != tc.want {
				t.Errorf("%s:\nwant: %s\ngot:  %v",
					name, tc.want, got)
			}
		})
	}
}

func TestSettingsGetBool(t *testing.T) {
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
		"log.timestamps hard default": {getFuncPtr: settings.GetLogTimestamps, flag: &LogTimestamps, want: "false", nilConfig: true, configPtr: &settings.Log.Timestamps},
		"log.timestamps config":       {getFuncPtr: settings.GetLogTimestamps, flag: &LogTimestamps, want: "true"},
		"log.timestamps flag":         {getFuncPtr: settings.GetLogTimestamps, flag: &LogTimestamps, flagVal: boolPtr(true), want: "true"},
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
				t.Errorf("%s:\nwant: %s\ngot:  %v",
					name, tc.want, got)
			}
		})
	}
}

func TestGetWebWebFileNotExist(t *testing.T) {
	// GIVEN strings that point to files that don't exist
	jLog = utils.NewJLog("WARN", false)
	settings := Settings{}
	tests := map[string]struct {
		file      string
		getFunc   func() *string
		changeVar **string
		want      string
	}{
		"hard default cert file": {getFunc: settings.GetWebCertFile},
		"config cert file":       {file: "cert_file_that_shouldnt_exist.hope", changeVar: &settings.Web.CertFile, getFunc: settings.GetWebCertFile},
		"flag cert file":         {file: "cert_file_that_shouldnt_exist.hope", changeVar: &WebCertFile, getFunc: settings.GetWebCertFile},
		"hard default pkey file": {getFunc: settings.GetWebKeyFile},
		"config pkey file":       {file: "pkey_file_that_shouldnt_exist.hope", changeVar: &settings.Web.KeyFile, getFunc: settings.GetWebKeyFile},
		"flag pkey file":         {file: "pkey_file_that_shouldnt_exist.hope", changeVar: &WebPKeyFile, getFunc: settings.GetWebKeyFile},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Switch Fatal to panic and disable this panic.
			jLog.Testing = true
			defer func() {
				r := recover()
				if r != nil && !strings.Contains(r.(string), "no such file or directory") {
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
				t.Errorf("%s:\n%q shouldn't exist, so this call should have been Fatal",
					name, tc.file)
			}
		})
	}
}
