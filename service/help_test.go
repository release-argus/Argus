// Copyright [2023] [Argus]
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
	"strings"
	"testing"

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

func TestMain(m *testing.M) {
	// initialize jLog
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true
	LogInit(jLog)

	// run other tests
	exitCode := m.Run()

	// exit
	os.Exit(exitCode)
}

func testServiceGitHub(id string) *Service {
	var (
		announceChannel chan []byte         = make(chan []byte, 2)
		saveChannel     chan bool           = make(chan bool, 5)
		databaseChannel chan dbtype.Message = make(chan dbtype.Message, 5)
	)
	svc := &Service{
		ID: id,
		LatestVersion: *latestver.New(
			stringPtr(os.Getenv("GITHUB_TOKEN")),
			nil, nil, nil,
			&filter.Require{
				RegexContent: "content",
				RegexVersion: "version"},
			nil,
			"github",
			"release-argus/Argus",
			nil,
			boolPtr(false),
			&latestver.LookupDefaults{},
			&latestver.LookupDefaults{}),
		Dashboard: *NewDashboardOptions(
			boolPtr(false), "test", "https://example.com", "https://release-argus.io",
			nil, nil),
		Options: *opt.New(
			boolPtr(true), "5s", boolPtr(true),
			&opt.OptionsDefaults{}, &opt.OptionsDefaults{}),
		Defaults:     &Defaults{},
		HardDefaults: &Defaults{},
	}
	// Status
	svc.Status = *svcstatus.New(
		&announceChannel, &databaseChannel, &saveChannel,
		"", "", "", "", "", "")
	svc.Status.ServiceID = &svc.ID
	svc.Status.SetApprovedVersion("1.1.1", false)
	svc.Status.SetLatestVersion("2.2.2", false)
	svc.Status.SetLatestVersionTimestamp("2002-02-02T02:02:02Z")
	svc.Status.SetDeployedVersion("0.0.0", false)
	svc.Status.SetDeployedVersionTimestamp("2001-01-01T01:01:01Z")

	svc.Init(
		&Defaults{}, &Defaults{},
		nil, nil, nil,
		nil, nil, nil)
	svc.Status.ServiceID = &svc.ID
	svc.Status.WebURL = &svc.Dashboard.WebURL
	return svc
}

func testServiceURL(id string) *Service {
	var (
		announceChannel = make(chan []byte, 5)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
	)
	svc := &Service{
		ID:                    id,
		LatestVersion:         *testLatestVersionLookupURL(false),
		DeployedVersionLookup: testDeployedVersionLookup(false),
		Dashboard: *NewDashboardOptions(
			boolPtr(false), "test", "https://release-argus.io", "https://release-argus.io",
			&DashboardOptionsDefaults{}, &DashboardOptionsDefaults{}),
		Status: *svcstatus.New(
			&announceChannel, &databaseChannel, &saveChannel,
			"", "", "", "", "", ""),
		Options: *opt.New(
			boolPtr(true), "5s", boolPtr(true),
			&opt.OptionsDefaults{}, &opt.OptionsDefaults{}),
		Defaults: &Defaults{},
		HardDefaults: &Defaults{
			DeployedVersionLookup: deployedver.LookupDefaults{},
			Status:                svcstatus.StatusDefaults{}}}
	svc.HardDefaults.Status.AnnounceChannel = svc.Status.AnnounceChannel
	svc.HardDefaults.Status.DatabaseChannel = svc.Status.DatabaseChannel
	svc.HardDefaults.Status.SaveChannel = svc.Status.SaveChannel

	// Status
	svc.Status.Init(
		0, 0, 0,
		&svc.ID,
		&svc.Dashboard.WebURL)
	svc.Status.SetApprovedVersion("1.1.1", false)
	svc.Status.SetLatestVersion("2.2.2", false)
	svc.Status.SetLatestVersionTimestamp("2002-02-02T02:02:02Z")
	svc.Status.SetDeployedVersion("0.0.0", false)
	svc.Status.SetDeployedVersionTimestamp("2001-01-01T01:01:01Z")

	svc.LatestVersion.Init(
		&latestver.LookupDefaults{}, &latestver.LookupDefaults{},
		&svc.Status,
		&svc.Options)
	svc.DeployedVersionLookup.Init(
		&deployedver.LookupDefaults{}, &deployedver.LookupDefaults{},
		&svc.Status,
		&svc.Options)
	return svc
}

func testLatestVersionLookupURL(fail bool) *latestver.Lookup {
	lv := latestver.New(
		stringPtr(os.Getenv("GITHUB_TOKEN")),
		boolPtr(!fail),
		nil,
		opt.New(
			nil, "", boolPtr(true),
			&opt.OptionsDefaults{}, &opt.OptionsDefaults{}),
		&filter.Require{
			RegexContent: "{{ version }}-beta",
			RegexVersion: "[0-9]+",
		},
		nil,
		"url",
		"https://invalid.release-argus.io/plain",
		&filter.URLCommandSlice{
			{Type: "regex", Regex: stringPtr("v([0-9.]+)")}},
		boolPtr(false),
		&latestver.LookupDefaults{},
		&latestver.LookupDefaults{})
	lv.Status.ServiceID = stringPtr("foo")
	return lv
}

func testDeployedVersionLookup(fail bool) (dvl *deployedver.Lookup) {
	dvl = deployedver.New(
		boolPtr(!fail),
		nil, nil,
		"version",
		opt.New(
			nil, "", boolPtr(true),
			&opt.OptionsDefaults{}, &opt.OptionsDefaults{}),
		"", nil,
		&svcstatus.Status{},
		"https://invalid.release-argus.io/json",
		&deployedver.LookupDefaults{},
		&deployedver.LookupDefaults{})

	announceChannel := make(chan []byte, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	saveChannel := make(chan bool, 5)
	dvl.Status = svcstatus.New(
		&announceChannel, &databaseChannel, &saveChannel,
		"", "", "", "", "", "")

	return
}

func testWebHook(failing bool) *webhook.WebHook {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	wh := webhook.New(
		boolPtr(false),
		nil,
		"0s",
		&desiredStatusCode,
		nil,
		&whMaxTries,
		nil,
		stringPtr("12m"),
		"argus",
		boolPtr(false),
		"github",
		"https://valid.release-argus.io/hooks/github-style",
		nil, nil, nil)
	if failing {
		wh.Secret = "notArgus"
	}
	return wh
}

func trimJSON(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, ": ", ":")
	return str
}
