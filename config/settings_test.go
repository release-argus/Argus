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
	"testing"

	"github.com/release-argus/Argus/utils"
)

func TestNilUndefinedFlagsDoesntResetDefined(t *testing.T) {
	// GIVEN the log.level flag has been set
	newLevel := "DEBUG"
	tmp := ""
	LogLevel = &tmp
	*LogLevel = newLevel
	flagset := map[string]bool{
		"log.level":        true,
		"log.timestamp":    false,
		"web.listen-host":  false,
		"web.listen-port":  false,
		"web.cert-file":    false,
		"web.pkey-file":    false,
		"web.route-prefix": false,
	}
	var settings Settings

	// GIVEN the flags that weren't set are nil'd with NilUndefinedFlags
	settings.NilUndefinedFlags(&flagset)

	// THEN LogLevel is not nil
	if LogLevel == nil || *LogLevel != newLevel {
		got := utils.DefaultIfNil(LogLevel)
		t.Errorf("LogLevel shouldn't have been reset to %s. Want %s", got, *LogLevel)
	}
}

func TestNilUndefinedFlagsDoesResetUndefined(t *testing.T) {
	// GIVEN the log.level flag has not been set
	newLevel := "DEBUG"
	*LogLevel = newLevel
	flagset := map[string]bool{
		"log.level":        false,
		"log.timestamp":    false,
		"web.listen-host":  false,
		"web.listen-port":  false,
		"web.cert-file":    false,
		"web.pkey-file":    false,
		"web.route-prefix": false,
	}
	var settings Settings

	// GIVEN the flags that weren't set are nil'd with NilUndefinedFlags
	settings.NilUndefinedFlags(&flagset)

	// THEN LogLevel is not nil
	if LogLevel != nil {
		t.Errorf("LogLevel should have been reset to nil, not %s", *LogLevel)
	}
}

func TestSettingsSetDefaultsWebFromFlags(t *testing.T) {
	// GIVEN the web.listen-port flag has been set
	want := "123"
	WebListenPort = &want
	var settings Settings

	// WHEN SetDefaults is called on it
	settings.SetDefaults()

	// THEN the Service part is initialised to the defined defaults
	got := settings.GetWebListenPort()
	if got != want {
		t.Errorf("settings.Web.CertFile should have been %s, but got %s", want, got)
	}
}

func TestSettingsSetDefaultsLogsFromFlags(t *testing.T) {
	// GIVEN the log.level flag has been set
	want := "DEBUG"
	tmp := ""
	LogLevel = &tmp
	*LogLevel = want
	var settings Settings

	// WHEN SetDefaults is called on it
	settings.SetDefaults()

	// THEN the Service part is initialised to the defined defaults
	got := settings.GetLogLevel()
	if got != want {
		t.Errorf("settings.Log.Level should have been %s, but got %s", want, got)
	}
}

func testSettings() Settings {
	logTimestamps := true
	logLevel := "DEBUG"
	webListenHost := "test"
	webListenPort := "123"
	webRoutePrefix := "something"
	webCertFile := "../test/ordering_0.yml"
	webKeyFile := "../test/ordering_1.yml"
	return Settings{
		Log: LogSettings{
			Timestamps: &logTimestamps,
			Level:      &logLevel,
		},
		Web: WebSettings{
			ListenHost:  &webListenHost,
			ListenPort:  &webListenPort,
			RoutePrefix: &webRoutePrefix,
			CertFile:    &webCertFile,
			KeyFile:     &webKeyFile,
		},
	}
}

func TestGetLogLevel(t *testing.T) {
	// GIVEN Log.Level is defined in the Settings
	settings := testSettings()

	// WHEN GetLogLevel is called
	got := settings.Log.Level

	// THEN it'll return the value from Settings.Log.Level
	want := settings.Log.Level
	if want != got {
		t.Errorf("Settings.Log.Level should have been returned (%s), not %s", *want, *got)
	}
}

func TestGetLogTimestamps(t *testing.T) {
	// GIVEN Log.Timestamps is defined in the Settings
	settings := testSettings()

	// WHEN GetLogTimestamps is called
	got := settings.Log.Timestamps

	// THEN it'll return the value from Settings.Log.Timestamps
	want := settings.Log.Timestamps
	if want != got {
		t.Errorf("Settings.Log.Timestamps should have been returned (%t), not %t", *want, *got)
	}
}

func TestGetWebListenHost(t *testing.T) {
	// GIVEN Web.ListenHost is defined in the Settings
	settings := testSettings()

	// WHEN TestGetWebListenHost is called
	got := settings.GetWebListenHost()

	// THEN it'll return the value from Settings.Web.ListenHost
	want := settings.Web.ListenHost
	if *want != got {
		t.Errorf("Settings.Web.ListenHost should have been returned (%s), not %s", *want, got)
	}
}

func TestGetWebListenPort(t *testing.T) {
	// GIVEN Web.ListenPort is defined in the Settings
	settings := testSettings()

	// WHEN TestGetWebListenPort is called
	got := settings.GetWebListenPort()

	// THEN it'll return the value from Settings.Web.ListenPort
	want := settings.Web.ListenPort
	if *want != got {
		t.Errorf("Settings.Web.ListenPort should have been returned (%s), not %s", *want, got)
	}
}

func TestGetWebRoutePrefix(t *testing.T) {
	// GIVEN Web.RoutePrefix is defined in the Settings
	settings := testSettings()

	// WHEN TestGetWebRoutePrefix is called
	got := settings.GetWebRoutePrefix()

	// THEN it'll return the value from Settings.Web.RoutePrefix
	want := settings.Web.RoutePrefix
	if *want != got {
		t.Errorf("Settings.Web.RoutePrefix should have been returned (%s), not %s", *want, got)
	}
}

func TestGetWebCertFile(t *testing.T) {
	// GIVEN Web.CertFile points to a file that exists
	settings := testSettings()

	// WHEN TestGetWebCertFile is called
	got := settings.GetWebCertFile()

	// THEN it'll return the value from Settings.Web.CertFile
	want := settings.Web.CertFile
	if want != got {
		t.Errorf("Settings.Web.CertFile should have been returned (%s), not %s", *want, *got)
	}
}

func TestGetWebCertFileUndefined(t *testing.T) {
	// GIVEN Web.CertFile is defined as an empty string
	settings := testSettings()
	*settings.Web.CertFile = ""

	// WHEN TestGetWebCertFile is called
	got := settings.GetWebCertFile()

	// THEN it'll return the value from Settings.Web.CertFile
	var want *string
	if want != got {
		t.Errorf("Settings.Web.CertFile should have been returned %v, not %s", want, *got)
	}
}

func TestGetWebCertFileNotExist(t *testing.T) {
	// GIVEN Web.CertFile points to a file that doesn't exist
	jLog = utils.NewJLog("WARN", false)
	settings := testSettings()
	*settings.Web.CertFile = "doesnt_exist"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN TestGetWebCertFile is called
	settings.GetWebCertFile()

	// THEN this call will crash the program
	t.Errorf("%s doesn't exist, so this call should have been Fatal", *settings.Web.CertFile)
}

func TestGetWebKeyFile(t *testing.T) {
	// GIVEN Web.KeyFile points to a file that exists
	settings := testSettings()

	// WHEN TestGetWebKeyFile is called
	got := settings.GetWebKeyFile()

	// THEN it'll return the value from Settings.Web.KeyFile
	want := settings.Web.KeyFile
	if want != got {
		t.Errorf("Settings.Web.KeyFile should have been returned (%s), not %s", *want, *got)
	}
}

func TestGetWebKeyFileUndefined(t *testing.T) {
	// GIVEN Web.KeyFile is defined as an empty string
	settings := testSettings()
	*settings.Web.KeyFile = ""

	// WHEN TestGetWebKeyFile is called
	got := settings.GetWebKeyFile()

	// THEN it'll return the value from Settings.Web.KeyFile
	var want *string
	if want != got {
		t.Errorf("Settings.Web.KeyFile should have been returned %v, not %s", want, *got)
	}
}

func TestGetWebKeyFileNotExist(t *testing.T) {
	// GIVEN Web.KeyFile points to a file that doesn't exist
	jLog = utils.NewJLog("WARN", false)
	settings := testSettings()
	*settings.Web.KeyFile = "doesnt_exist"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN TestGetWebKeyFile is called
	settings.GetWebKeyFile()

	// THEN this call will crash the program
	t.Errorf("%s doesn't exist, so this call should have been Fatal", *settings.Web.KeyFile)
}
