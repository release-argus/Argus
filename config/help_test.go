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

//go:build unit || integration

package config

import (
	"fmt"
	"os"
	"sync"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

func boolPtr(val bool) *bool {
	return &val
}
func intPtr(val int) *int {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func uintPtr(val int) *uint {
	converted := uint(val)
	return &converted
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
					Level: &logLevel}}},
		HardDefaults: Defaults{
			Service: service.Defaults{
				Status: svcstatus.NewStatusDefaults(
					nil, &databaseChannel, &saveChannel)}},
		DatabaseChannel: &databaseChannel,
		SaveChannel:     &saveChannel,
	}
}

func testSettings() Settings {
	logTimestamps := true
	logLevel := "DEBUG"
	dataDatabaseFile := "somewhere.db"
	webListenHost := "test"
	webListenPort := "123"
	webRoutePrefix := "/something"
	webCertFile := "../README.md"
	webKeyFile := "../LICENSE"
	return Settings{
		SettingsBase: SettingsBase{
			Log: LogSettings{
				Timestamps: &logTimestamps,
				Level:      &logLevel,
			},
			Data: DataSettings{
				DatabaseFile: &dataDatabaseFile,
			},
			Web: WebSettings{
				ListenHost:  &webListenHost,
				ListenPort:  &webListenPort,
				RoutePrefix: &webRoutePrefix,
				CertFile:    &webCertFile,
				KeyFile:     &webKeyFile,
			}},
	}
}

var loadMutex sync.RWMutex

func testLoad(file string, t *testing.T) (config *Config) {
	config = &Config{}

	flags := make(map[string]bool)
	log := util.NewJLog("WARN", true)
	loadMutex.Lock()
	defer loadMutex.Unlock()
	config.Load(file, &flags, log)
	t.Cleanup(func() { os.Remove(*config.Settings.DataDatabaseFile()) })

	return
}

var mutex sync.Mutex

func testLoadBasic(file string, t *testing.T) (config *Config) {
	config = &Config{}

	config.File = file

	//#nosec G304 -- Loading the test config file
	data, err := os.ReadFile(file)
	jLog.Fatal(fmt.Sprintf("Error reading %q\n%s", file, err),
		util.LogFrom{}, err != nil)

	err = yaml.Unmarshal(data, config)
	jLog.Fatal(fmt.Sprintf("Unmarshal of %q failed\n%s", file, err),
		util.LogFrom{}, err != nil)

	saveChannel := make(chan bool, 32)
	config.SaveChannel = &saveChannel
	config.HardDefaults.Service.Status.SaveChannel = config.SaveChannel

	databaseChannel := make(chan dbtype.Message, 32)
	config.DatabaseChannel = &databaseChannel
	config.HardDefaults.Service.Status.DatabaseChannel = config.DatabaseChannel

	config.GetOrder(data)
	mutex.Lock()
	defer mutex.Unlock()
	config.Init()
	for name, service := range config.Service {
		service.ID = name
	}
	config.CheckValues()

	return
}

func testServiceURL(id string) *service.Service {
	var (
		announceChannel = make(chan []byte, 5)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
	)
	svc := &service.Service{
		ID: id,
		LatestVersion: *latestver.New(
			nil,
			boolPtr(false),
			nil,
			&opt.Options{},
			&filter.Require{
				RegexContent: "{{ version }}-beta",
				RegexVersion: "[0-9]+"},
			nil,
			"url",
			"https://valid.release-argus.io/plain",
			&filter.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("v([0-9.]+)")}},
			boolPtr(false),
			&latestver.LookupDefaults{}, &latestver.LookupDefaults{}),
		DeployedVersionLookup: deployedver.New(
			boolPtr(false),
			nil, nil,
			"version",
			nil, "", nil,
			"https://valid.release-argus.io/json",
			&deployedver.LookupDefaults{}, &deployedver.LookupDefaults{}),
		Dashboard: *service.NewDashboardOptions(
			boolPtr(false), "test", "https://release-argus.io", "https://release-argus.io/docs",
			&service.DashboardOptionsDefaults{}, &service.DashboardOptionsDefaults{}),
		Status: *svcstatus.New(
			&announceChannel, &databaseChannel, &saveChannel,
			"", "", "", "", "", ""),
		Options: *opt.New(
			boolPtr(true), "5s", boolPtr(true),
			&opt.OptionsDefaults{}, &opt.OptionsDefaults{}),
		Defaults:     &service.Defaults{},
		HardDefaults: &service.Defaults{}}
	svc.Status.ServiceID = &svc.ID
	svc.Status.Init(
		len(svc.Notify), len(svc.Command), len(svc.WebHook),
		&svc.ID,
		&svc.Dashboard.WebURL)
	svc.LatestVersion.Init(
		&latestver.LookupDefaults{}, &latestver.LookupDefaults{},
		&svc.Status,
		&svc.Options)
	svc.DeployedVersionLookup.Init(
		&deployedver.LookupDefaults{}, &deployedver.LookupDefaults{},
		&svc.Status,
		&svc.Options)
	svc.Status.WebURL = &svc.Dashboard.WebURL

	svc.Status.SetLastQueried("")
	svc.Status.SetApprovedVersion("1.1.1", false)
	svc.Status.SetLatestVersion("2.2.2", false)
	svc.Status.SetLatestVersionTimestamp("2002-02-02T02:02:02Z")
	svc.Status.SetDeployedVersion("0.0.0", false)
	svc.Status.SetDeployedVersionTimestamp("2001-01-01T01:01:01Z")
	return svc
}

func TestMain(m *testing.M) {
	log := util.NewJLog("DEBUG", true)
	log.Testing = true
	LogInit(log)
	os.Exit(m.Run())
}
