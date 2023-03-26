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
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
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
		File:            "/root/inaccessible",
		DatabaseChannel: &databaseChannel,
		SaveChannel:     &saveChannel,
		Settings: Settings{
			Indentation: 4,
			Log: LogSettings{
				Level: &logLevel,
			},
		},
	}
}

func testConfigEdit() Config {
	jLog = util.NewJLog("WARN", true)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)

	return Config{
		Order: []string{"alpha", "bravo", "charlie"},
		Service: map[string]*service.Service{
			"alpha":   testServiceURL("alpha"),
			"bravo":   testServiceURL("bravo"),
			"charlie": testServiceURL("charlie"),
		},
		// Set defaults so the test doesn't fail
		Defaults: Defaults{Service: service.Service{Options: opt.Options{}}},
		HardDefaults: Defaults{
			Service: service.Service{
				Options: opt.Options{Interval: "1y"},
				Status: svcstatus.Status{
					SaveChannel:     &saveChannel,
					DatabaseChannel: &databaseChannel,
				},
			},
		},
		DatabaseChannel: &databaseChannel,
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

func testLoad(file string) *Config {
	var config Config

	flags := make(map[string]bool)
	log := util.NewJLog("WARN", true)
	config.Load(file, &flags, log)

	return &config
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
			ServiceID:                stringPtr("test"),
			ApprovedVersion:          "1.1.1",
			LatestVersion:            "2.2.2",
			LatestVersionTimestamp:   "2002-02-02T02:02:02Z",
			DeployedVersion:          "0.0.0",
			DeployedVersionTimestamp: "2001-01-01T01:01:01Z",
			AnnounceChannel:          &announceChannel,
			DatabaseChannel:          &databaseChannel,
			SaveChannel:              &saveChannel,
			LastQueried:              time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
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
	svc.LatestVersion.Init(jLog, &latestver.Lookup{}, &latestver.Lookup{}, &svc.Status, &svc.Options)
	svc.DeployedVersionLookup.Init(jLog, &deployedver.Lookup{}, &deployedver.Lookup{}, &svc.Status, &svc.Options)
	svc.Status.WebURL = &svc.Dashboard.WebURL
	return svc
}
