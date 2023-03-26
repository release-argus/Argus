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

package service

import (
	"fmt"
	"os"

	dbtype "github.com/release-argus/Argus/db/types"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
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
func stringifyPointer[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}
func testLogging() {
	jLog = util.NewJLog("WARN", false)
	jLog.Testing = true
	var webhookLogs *webhook.Slice
	webhookLogs.Init(jLog, nil, nil, nil, nil, nil, nil)
	var latestVersion latestver.Lookup
	latestVersion.Init(jLog, nil, nil, &svcstatus.Status{ServiceID: stringPtr("foo")}, nil)
	var deployedVersion *deployedver.Lookup
	deployedVersion.Init(jLog, nil, nil, &svcstatus.Status{ServiceID: stringPtr("foo")}, nil)
}

func testServiceGitHub(id string) *Service {
	var (
		announceChannel chan []byte         = make(chan []byte, 2)
		saveChannel     chan bool           = make(chan bool, 5)
		databaseChannel chan dbtype.Message = make(chan dbtype.Message, 5)
	)
	svc := &Service{
		ID: id,
		LatestVersion: latestver.Lookup{
			Type:        "github",
			AccessToken: stringPtr(os.Getenv("GITHUB_TOKEN")),
			URL:         "release-argus/Argus",
			Require: &filter.Require{
				RegexContent: "content",
				RegexVersion: "version",
			},
			AllowInvalidCerts: boolPtr(true),
			UsePreRelease:     boolPtr(false),
		},
		Dashboard: DashboardOptions{
			AutoApprove: boolPtr(false),
			Icon:        "test",
			IconLinkTo:  "https://example.com",
			WebURL:      "https://release-argus.io",
		},
		Status: svcstatus.Status{
			ApprovedVersion:          "1.1.1",
			LatestVersion:            "2.2.2",
			LatestVersionTimestamp:   "2002-02-02T02:02:02Z",
			DeployedVersion:          "0.0.0",
			DeployedVersionTimestamp: "2001-01-01T01:01:01Z",
			AnnounceChannel:          &announceChannel,
			DatabaseChannel:          &databaseChannel,
			SaveChannel:              &saveChannel,
		},
		Options: opt.Options{
			Interval:           "5s",
			SemanticVersioning: boolPtr(true),
		},
		Defaults: &Service{},
		HardDefaults: &Service{
			Options: opt.Options{
				Active: boolPtr(true)},
		},
	}
	svc.Init(jLog, &Service{}, &Service{}, nil, nil, nil, nil, nil, nil)
	svc.Status.ServiceID = &svc.ID
	svc.Status.WebURL = &svc.Dashboard.WebURL
	return svc
}

func testLatestVersionLookupURL(fail bool) latestver.Lookup {
	return latestver.Lookup{
		Type: "url",
		URL:  "https://invalid.release-argus.io/plain",
		Require: &filter.Require{
			RegexContent: "{{ version }}-beta",
			RegexVersion: "[0-9]+",
		},
		URLCommands: filter.URLCommandSlice{
			{Type: "regex", Regex: stringPtr("v([0-9.]+)")},
		},
		AllowInvalidCerts: boolPtr(!fail),
		UsePreRelease:     boolPtr(false),
	}
}

func testDeployedVersionLookup(fail bool) *deployedver.Lookup {
	return &deployedver.Lookup{
		URL:               "https://invalid.release-argus.io/json",
		JSON:              "version",
		AllowInvalidCerts: boolPtr(!fail),
	}
}

func testServiceURL(id string) *Service {
	var (
		announceChannel = make(chan []byte, 5)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
	)
	svc := &Service{
		ID:                    id,
		LatestVersion:         testLatestVersionLookupURL(false),
		DeployedVersionLookup: testDeployedVersionLookup(false),
		Dashboard: DashboardOptions{
			AutoApprove:  boolPtr(false),
			Icon:         "test",
			IconLinkTo:   "https://release-argus.io",
			WebURL:       "https://release-argus.io",
			Defaults:     &DashboardOptions{},
			HardDefaults: &DashboardOptions{},
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
		},
		Options: opt.Options{
			Interval:           "5s",
			SemanticVersioning: boolPtr(true),
			Defaults:           &opt.Options{},
			HardDefaults:       &opt.Options{},
		},
		Defaults: &Service{},
		HardDefaults: &Service{
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

func testWebHook(failing bool) *webhook.WebHook {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	wh := &webhook.WebHook{
		ID:                "test",
		Type:              "github",
		URL:               "https://valid.release-argus.io/hooks/github-style",
		Secret:            "argus",
		AllowInvalidCerts: boolPtr(false),
		DesiredStatusCode: &desiredStatusCode,
		Delay:             "0s",
		SilentFails:       boolPtr(false),
		MaxTries:          &whMaxTries,
		ParentInterval:    stringPtr("12m"),
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
	if failing {
		wh.Secret = "notArgus"
	}
	return wh
}
