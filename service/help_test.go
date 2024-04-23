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
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

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

func testService(id string, sType string) *Service {
	var (
		announceChannel = make(chan []byte, 5)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
	)
	svc := &Service{
		ID:                    id,
		LatestVersion:         *testLatestVersion(sType, false),
		DeployedVersionLookup: testDeployedVersionLookup(false),
		Dashboard: *NewDashboardOptions(
			test.BoolPtr(false), "test", "https://release-argus.io", "https://release-argus.io",
			&DashboardOptionsDefaults{}, &DashboardOptionsDefaults{}),
		Status: *svcstatus.New(
			&announceChannel, &databaseChannel, &saveChannel,
			"", "", "", "", "", ""),
		Options: *opt.New(
			test.BoolPtr(true), "5s", test.BoolPtr(true),
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

	svc.DeployedVersionLookup.Init(
		&deployedver.LookupDefaults{}, &deployedver.LookupDefaults{},
		&svc.Status,
		&svc.Options)
	return svc
}

func testLatestVersionGitHub(fail bool) (lv *latestver.Lookup) {
	lv = latestver.New(
		test.StringPtr(os.Getenv("GITHUB_TOKEN")),
		nil, nil, nil,
		&filter.Require{
			RegexContent: "content",
			RegexVersion: "version"},
		nil,
		"github",
		"release-argus/Argus",
		nil,
		test.BoolPtr(false),
		&latestver.LookupDefaults{},
		&latestver.LookupDefaults{})

	if fail {
		lv.AccessToken = test.StringPtr("invalid")
	}
	return
}
func testLatestVersionURL(fail bool) (lv *latestver.Lookup) {
	lv = latestver.New(
		nil,
		test.BoolPtr(!fail),
		nil,
		opt.New(
			nil, "", test.BoolPtr(true),
			&opt.OptionsDefaults{}, &opt.OptionsDefaults{}),
		&filter.Require{
			RegexContent: "{{ version }}-beta",
			RegexVersion: "[0-9]+",
		},
		nil,
		"url",
		"https://invalid.release-argus.io/plain",
		&filter.URLCommandSlice{
			{Type: "regex", Regex: test.StringPtr("v([0-9.]+)")}},
		test.BoolPtr(false),
		&latestver.LookupDefaults{},
		&latestver.LookupDefaults{})

	return
}

func testLatestVersion(lvType string, fail bool) (lv *latestver.Lookup) {
	if lvType == "url" {
		lv = testLatestVersionURL(fail)
	} else {
		lv = testLatestVersionGitHub(fail)
	}
	lv.Status.ServiceID = test.StringPtr("TEST_LV")
	return lv
}

func testDeployedVersionLookup(fail bool) (dvl *deployedver.Lookup) {
	dvl = deployedver.New(
		test.BoolPtr(!fail),
		nil, nil,
		"version",
		opt.New(
			nil, "", test.BoolPtr(true),
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
