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
func stringPtr(val string) *string {
	return &val
}

func testConfig() Config {
	logLevel := "DEBUG"
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)

	return Config{
		File: "/root/inaccessible",
		Settings: Settings{
			Indentation: 4,
			Log: LogSettings{
				Level: &logLevel}},
		HardDefaults: Defaults{
			Service: service.Service{
				Status: svcstatus.Status{
					DatabaseChannel: &databaseChannel,
					SaveChannel:     &saveChannel}}},
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
		},
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
	t.Cleanup(func() { os.Remove(*config.Settings.GetDataDatabaseFile()) })

	return
}

func testLoadBasic(file string, t *testing.T) (config *Config) {
	config = &Config{}

	config.File = file

	//#nosec G304 -- Loading the test config file
	data, err := os.ReadFile(file)
	msg := fmt.Sprintf("Error reading %q\n%s", file, err)
	jLog.Fatal(msg, util.LogFrom{}, err != nil)

	err = yaml.Unmarshal(data, config)
	msg = fmt.Sprintf("Unmarshal of %q failed\n%s", file, err)
	jLog.Fatal(msg, util.LogFrom{}, err != nil)

	saveChannel := make(chan bool, 32)
	config.SaveChannel = &saveChannel
	config.HardDefaults.Service.Status.SaveChannel = config.SaveChannel

	databaseChannel := make(chan dbtype.Message, 32)
	config.DatabaseChannel = &databaseChannel
	config.HardDefaults.Service.Status.DatabaseChannel = config.DatabaseChannel

	config.GetOrder(data)
	config.Init()
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
		LatestVersion: latestver.Lookup{
			Type: "url",
			URL:  "https://valid.release-argus.io/plain",
			Require: &filter.Require{
				RegexContent: "{{ version }}-beta",
				RegexVersion: "[0-9]+",
			},
			URLCommands: filter.URLCommandSlice{
				{Type: "regex", Regex: stringPtr("v([0-9.]+)")},
			},
			AllowInvalidCerts: boolPtr(true),
			UsePreRelease:     boolPtr(false),
		},
		DeployedVersionLookup: &deployedver.Lookup{
			URL:               "https://valid.release-argus.io/json",
			JSON:              "version",
			AllowInvalidCerts: boolPtr(false),
		},
		Dashboard: service.DashboardOptions{
			AutoApprove:  boolPtr(false),
			Icon:         "test",
			IconLinkTo:   "https://release-argus.io",
			WebURL:       "https://release-argus.io",
			Defaults:     &service.DashboardOptions{},
			HardDefaults: &service.DashboardOptions{},
		},
		Status: svcstatus.Status{
			ServiceID:       stringPtr("test"),
			AnnounceChannel: &announceChannel,
			DatabaseChannel: &databaseChannel,
			SaveChannel:     &saveChannel,
		},
		Options: opt.Options{
			Interval:           "5s",
			SemanticVersioning: boolPtr(true),
			Defaults:           &opt.Options{},
			HardDefaults:       &opt.Options{},
		},
		Defaults: &service.Service{},
		HardDefaults: &service.Service{
			Options: opt.Options{
				Active: boolPtr(true)},
			DeployedVersionLookup: &deployedver.Lookup{},
		},
	}
	svc.Status.ServiceID = &svc.ID
	svc.Status.Init(
		len(svc.Notify), len(svc.Command), len(svc.WebHook),
		&svc.ID,
		&svc.Dashboard.WebURL)
	svc.LatestVersion.Init(
		&latestver.Lookup{}, &latestver.Lookup{},
		&svc.Status,
		&svc.Options)
	svc.DeployedVersionLookup.Init(
		&deployedver.Lookup{}, &deployedver.Lookup{},
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
	LogInit(util.NewJLog("DEBUG", true))
	os.Exit(m.Run())
}
