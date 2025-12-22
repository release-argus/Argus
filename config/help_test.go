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

//go:build unit || integration

package config

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
	logutil "github.com/release-argus/Argus/util/log"
)

var packageName = "config"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logutil.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty",
			packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

func testConfig() Config {
	logLevel := "DEBUG"
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)

	return Config{
		File: "/root/inaccessible",
		Settings: Settings{
			Indentation: 4,
			SettingsBase: SettingsBase{
				Log: LogSettings{
					Level: logLevel}}},
		HardDefaults: Defaults{
			Service: service.Defaults{
				Status: status.NewDefaults(
					nil, databaseChannel, saveChannel)}},
		DatabaseChannel: databaseChannel,
		SaveChannel:     saveChannel,
	}
}

func testSettings() Settings {
	logTimestamps := true
	return Settings{
		SettingsBase: SettingsBase{
			Log: LogSettings{
				Timestamps: &logTimestamps,
				Level:      "DEBUG",
			},
			Data: DataSettings{
				DatabaseFile: "somewhere.db",
			},
			Web: WebSettings{
				ListenHost:  "test",
				ListenPort:  "123",
				RoutePrefix: "/something",
				CertFile:    "../README.md",
				KeyFile:     "../LICENSE",
			}},
	}
}

var loadMutex sync.RWMutex

func testLoadBasic(t *testing.T, file string) *Config {
	config := &Config{}

	config.File = file

	//#nosec G304 -- Loading the test config file
	data, err := os.ReadFile(file)
	if err != nil {
		logutil.Log.Fatal(
			fmt.Sprintf("%s\nError reading %q\n%s",
				packageName, file, err),
			logutil.LogFrom{})
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		logutil.Log.Fatal(
			fmt.Sprintf("%q\nUnmarshal of %q failed\n%s",
				packageName, file, err),
			logutil.LogFrom{})
	}

	saveChannel := make(chan bool, 32)
	config.SaveChannel = saveChannel
	config.HardDefaults.Service.Status.SaveChannel = config.SaveChannel

	databaseChannel := make(chan dbtype.Message, 32)
	config.DatabaseChannel = databaseChannel
	config.HardDefaults.Service.Status.DatabaseChannel = config.DatabaseChannel

	config.GetOrder(data)
	config.Init()
	config.CheckValues()
	t.Logf("%s - Loaded %q",
		packageName, file)

	return config
}

func testServiceURL(id string) *service.Service {
	var (
		announceChannel = make(chan []byte, 5)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
		defaults        = &service.Defaults{}
		hardDefaults    = &service.Defaults{}
	)
	hardDefaults.Default()

	options := opt.New(
		test.BoolPtr(true), "5s", test.BoolPtr(true),
		&defaults.Options, &hardDefaults.Options)

	lv, err := latestver.New(
		"url",
		"yaml", test.TrimYAML(`
			url: `+test.LookupPlain["url_valid"]+`
			url_commands:
				- type: regex
					regex: 'v([0-9.]+)'
			require:
				regex_content: "{{ version }}-beta"
				regex_version: "[0-9]+"
			access_token: `+os.Getenv("GITHUB_TOKEN")+`
		`),
		nil, nil,
		&defaults.LatestVersion, &hardDefaults.LatestVersion)
	if err != nil {
		panic(err)
	}

	dv := test.IgnoreError(nil, func() (deployedver.Lookup, error) {
		return deployedver.New(
			"url",
			"yaml", test.TrimYAML(`
				method: GET
				url: `+test.LookupJSON["url_valid"]+`
				json: version
		`),
			nil,
			nil,
			&defaults.DeployedVersionLookup, &hardDefaults.DeployedVersionLookup)
	})

	svc := &service.Service{
		ID:                    id,
		LatestVersion:         lv,
		DeployedVersionLookup: dv,
		Dashboard: *dashboard.NewOptions(
			test.BoolPtr(false), "test", "https://release-argus.io", "https://release-argus.io/docs", nil,
			&dashboard.OptionsDefaults{}, &dashboard.OptionsDefaults{}),
		Options: *options,
		Status: *status.New(
			announceChannel, databaseChannel, saveChannel,
			"",
			"", "",
			"", "",
			"",
			nil),
		Defaults:     &service.Defaults{},
		HardDefaults: &service.Defaults{}}

	svc.Init(
		defaults, hardDefaults,
		nil, nil, nil,
		nil, nil, nil)

	svc.Status.SetLastQueried("")
	svc.Status.SetApprovedVersion("1.1.1", false)
	svc.Status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	svc.Status.SetDeployedVersion("0.0.0", "2001-01-01T01:01:01Z", false)
	return svc
}

func testOkMatch(t *testing.T, want bool, channel chan bool, releaseStdout func() string) {
	select {
	case got := <-channel:
		drainAndDebounce(t.Context(), logutil.ExitCodeChannel(), 200*time.Millisecond)
		// Ok value as expected.
		if got != want {
			t.Errorf("%s\nok mismatch:\nwant: %t\ngot:  %t",
				packageName, want, got)
		}
	case <-time.After(2500 * time.Millisecond):
		drainAndDebounce(t.Context(), logutil.ExitCodeChannel(), 200*time.Millisecond)
		if releaseStdout != nil {
			_ = releaseStdout()
		}
	}
}
