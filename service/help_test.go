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

package service

import (
	"fmt"
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestMain(m *testing.M) {
	// initialise jLog
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true
	LogInit(jLog)

	// run other tests
	exitCode := m.Run()

	// exit
	os.Exit(exitCode)
}

func testOptions() *opt.Options {
	hardDefaults := opt.Defaults{}
	hardDefaults.Default()

	return opt.New(
		nil, "5s", test.BoolPtr(true),
		&opt.Defaults{}, &hardDefaults)
}

func testStatus() *status.Status {
	var (
		announceChannel = make(chan []byte, 5)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
	)

	return status.New(
		&announceChannel, &databaseChannel, &saveChannel,
		"", "", "", "", "", "")
}

func testService(t *testing.T, id string, sType string) *Service {
	hardDefaults := Defaults{}
	hardDefaults.Default()

	svc := &Service{
		ID:                    id,
		LatestVersion:         testLatestVersion(t, sType, false),
		DeployedVersionLookup: testDeployedVersionLookup(t, false),
		Dashboard: *NewDashboardOptions(
			test.BoolPtr(false), "test", "https://release-argus.io", "https://release-argus.io", nil,
			&DashboardOptionsDefaults{}, &DashboardOptionsDefaults{}),
		Status:       *testStatus(),
		Options:      *testOptions(),
		Defaults:     &Defaults{},
		HardDefaults: &hardDefaults}
	svc.HardDefaults.Status.AnnounceChannel = svc.Status.AnnounceChannel
	svc.HardDefaults.Status.DatabaseChannel = svc.Status.DatabaseChannel
	svc.HardDefaults.Status.SaveChannel = svc.Status.SaveChannel

	// Status
	svc.Status.Init(
		0, 0, 0,
		&svc.ID, &svc.ID,
		&svc.Dashboard.WebURL)
	svc.Status.SetApprovedVersion("1.1.1", false)
	svc.Status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	svc.Status.SetDeployedVersion("0.0.0", "2001-01-01T01:01:01Z", false)

	svc.LatestVersion.Init(
		&svc.Options,
		&svc.Status,
		svc.LatestVersion.GetDefaults(), &hardDefaults.LatestVersion)
	svc.DeployedVersionLookup.Init(
		&svc.Options,
		&svc.Status,
		&deployedver.Defaults{}, &hardDefaults.DeployedVersionLookup)

	err := svc.LatestVersion.CheckValues("")
	if err != nil {
		t.Fatalf("testService(), latest_version.CheckValues() error: %v", err)
	}
	err = svc.DeployedVersionLookup.CheckValues("")
	if err != nil {
		t.Fatalf("testService(), deployed_version.CheckValues() error: %v", err)
	}

	return svc
}

func testLatestVersionGitHub(t *testing.T, fail bool) latestver_base.Interface {
	hardDefaults := latestver_base.Defaults{}
	hardDefaults.Default()
	accessToken := os.Getenv("GITHUB_TOKEN")
	if fail {
		accessToken = "invalid"
	}

	lv, _ := latestver.New(
		"github",
		"yaml", test.TrimYAML(`
				url: release-argus/Argus
				access_token: `+accessToken+`
				require:
					regex_content: content
					regex_version: version
			`),
		testOptions(),
		testStatus(),
		&latestver_base.Defaults{}, &hardDefaults)

	err := lv.CheckValues("")
	if err != nil {
		t.Fatalf("testLatestVersionGitHub(), CheckValues() error: %v", err)
	}

	return lv
}

func testLatestVersionWeb(t *testing.T, fail bool) latestver_base.Interface {
	hardDefaults := latestver_base.Defaults{}
	hardDefaults.Default()

	lv, _ := latestver.New(
		"url",
		"yaml", test.TrimYAML(`
				url: https://invalid.release-argus.io/plain
				allow_invalid_certs: `+fmt.Sprint(!fail)+`
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
				require:
					regex_content: "{{ version }}-beta"
					regex_version: "[0-9]+"
			`),
		testOptions(),
		testStatus(),
		&latestver_base.Defaults{}, &hardDefaults)

	err := lv.CheckValues("")
	if err != nil {
		t.Fatalf("testLatestVersionWeb(), CheckValues() error: %v", err)
	}

	return lv
}

func testLatestVersion(t *testing.T, lvType string, fail bool) (lv latestver.Lookup) {
	if lvType == "url" {
		lv = testLatestVersionWeb(t, fail)
	} else {
		lv = testLatestVersionGitHub(t, fail)
	}

	lv.Init(
		lv.GetOptions(),
		lv.GetStatus(),
		lv.GetDefaults(), lv.GetHardDefaults())
	lv.GetStatus().ServiceID = test.StringPtr("TEST_LV")

	err := lv.CheckValues("")
	if err != nil {
		t.Fatalf("testLatestVersion(), CheckValues() error: %v", err)
	}

	return lv
}

func testDeployedVersionLookup(t *testing.T, fail bool) *deployedver.Lookup {
	hardDefaults := deployedver.Defaults{}
	hardDefaults.Default()

	dv := test.IgnoreError(t, func() (*deployedver.Lookup, error) {
		return deployedver.New(
			"yaml", test.TrimYAML(`
				method: GET
				url: https://invalid.release-argus.io/json
				allow_invalid_certs: `+fmt.Sprint(!fail)+`
				json: version
			`),
			testOptions(),
			testStatus(),
			&deployedver.Defaults{}, &hardDefaults)
	})

	err := dv.CheckValues("")
	if err != nil {
		t.Fatalf("testDeployedVersionLookup(), CheckValues() error: %v", err)
	}

	return dv
}
