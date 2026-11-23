// Copyright [2026] [Argus]
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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestLookup_HandleNewVersion(t *testing.T) {
	type versions struct {
		initialLatestVersion, initialDeployedVersion string
		newVersion, releaseDate                      string
	}

	tests := []struct {
		name                 string
		versions             versions
		writeToDB            bool
		wantAnnounce         bool
		wantNotify           bool
		wantDatabaseMessages int
	}{
		{
			name: "first version found/no deployed version/DB write",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			writeToDB:            true,
			wantAnnounce:         true,
			wantNotify:           false,
			wantDatabaseMessages: 2,
		},
		{
			name: "first version found/no deployed version/no DB write",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			writeToDB:            false,
			wantAnnounce:         false,
			wantNotify:           false,
			wantDatabaseMessages: 0,
		},
		{
			name: "first version found/have newer deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "1.0.1",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			writeToDB:            true,
			wantAnnounce:         true,
			wantNotify:           false,
			wantDatabaseMessages: 1,
		},
		{
			name: "first version found/have older deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "0.9.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			writeToDB:            true,
			wantAnnounce:         true,
			wantNotify:           false,
			wantDatabaseMessages: 1,
		},
		{
			name: "new version found/DB write",
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.1.0",
				releaseDate:            "2024-02-01",
			},
			writeToDB:            true,
			wantNotify:           true,
			wantDatabaseMessages: 1,
		},
		{
			name: "new version found/no DB write",
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.1.0",
				releaseDate:            "2024-02-01",
			},
			writeToDB:            false,
			wantNotify:           true,
			wantDatabaseMessages: 0,
		},
		{
			name: "same version found, which shouldn't occur in practice",
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			writeToDB:            true,
			wantNotify:           true,
			wantDatabaseMessages: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logFrom := logx.LogFrom{Primary: "TestLookup_HandleNewVersion", Secondary: tc.name}
			lookup := &Lookup{
				Status: &status.Status{},
			}
			lookup.Status.AnnounceChannel = make(chan []byte, 2)
			lookup.Status.DatabaseChannel = make(chan dbtype.Message, 2)
			lookup.Status.Init(
				0, 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			lookup.Status.SetLatestVersion(tc.versions.initialLatestVersion, "", false)
			lookup.Status.SetDeployedVersion(tc.versions.initialDeployedVersion, "", false)

			notify, err := lookup.HandleNewVersion(
				tc.versions.newVersion, tc.versions.releaseDate,
				tc.writeToDB,
				logFrom,
			)
			if err != nil {
				t.Errorf("Lookup.HandleNewVersion() failed: %v", err)
			}

			if notify != tc.wantNotify {
				t.Errorf("notify mismatch\ngot:  %t\nwant: %t", notify, tc.wantNotify)
			}

			wantLen := 0
			if tc.wantAnnounce {
				wantLen = 1
			}
			if gotLen := len(lookup.Status.AnnounceChannel); gotLen != wantLen {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotLen, wantLen,
				)
			}

			if got := len(lookup.Status.DatabaseChannel); got != tc.wantDatabaseMessages {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantDatabaseMessages,
				)
			}

			if got := lookup.Status.LatestVersion(); got != tc.versions.newVersion {
				t.Errorf(
					"Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.versions.newVersion,
				)
			}

			wantDeployedVersion := util.ValueOr(tc.versions.initialDeployedVersion, tc.versions.newVersion)
			if got := lookup.Status.DeployedVersion(); got != wantDeployedVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					got, wantDeployedVersion,
				)
			}
		})
	}
}

func TestLookup_HandleNoVersionChange(t *testing.T) {
	tests := []struct {
		name          string
		checkNumber   int
		writeToDB     bool
		wantAnnounces int
		stdoutRegex   string
	}{
		{
			name:          "the first check",
			checkNumber:   0,
			writeToDB:     true,
			wantAnnounces: 1,
			stdoutRegex:   `^$`,
		},
		{
			name:          "the second check, which also logs that it is staying put",
			checkNumber:   1,
			writeToDB:     true,
			wantAnnounces: 1,
			stdoutRegex:   `^VERBOSE: .*Staying on "1\.2\.3" as that's the latest version in the second check\n$`,
		},
		{
			name:          "no DB write, so no announce",
			checkNumber:   0,
			writeToDB:     false,
			wantAnnounces: 0,
			stdoutRegex:   `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			lookup := &Lookup{Status: &status.Status{}}
			lookup.Status.AnnounceChannel = make(chan []byte, 2)
			lookup.Status.Init(
				0, 0, 0,
				status.ServiceInfo{ID: tc.name},
				&dashboard.Options{},
			)
			lookup.Status.SetLatestVersion("1.2.3", "", false)
			// Drain what setting the version announced.
			for len(lookup.Status.AnnounceChannel) > 0 {
				<-lookup.Status.AnnounceChannel
			}
			releaseLog := test.CaptureLog(t, logx.Default())

			lookup.HandleNoVersionChange(
				tc.checkNumber, "1.2.3",
				tc.writeToDB,
				logx.LogFrom{Primary: "TestLookup_HandleNoVersionChange", Secondary: tc.name},
			)

			stdout := releaseLog()
			if got, want := len(lookup.Status.AnnounceChannel), tc.wantAnnounces; got != want {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
			if got, want := lookup.Status.LatestVersion(), "1.2.3"; got != want {
				t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, want)
			}

			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
			}
		})
	}
}
