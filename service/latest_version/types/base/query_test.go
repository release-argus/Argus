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
		name         string
		versions     versions
		wantAnnounce bool
		wantNotify   bool
	}{
		{
			name: "first version found/no deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounce: true,
			wantNotify:   false,
		},
		{
			name: "first version found/have newer deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "1.0.1",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounce: true,
			wantNotify:   false,
		},
		{
			name: "first version found/have older deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "0.9.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounce: true,
			wantNotify:   false,
		},
		{
			name: "new version found",
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.1.0",
				releaseDate:            "2024-02-01",
			},
			wantNotify: true,
		},
		{
			name: "same version found, which shouldn't occur in practice",
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantNotify: true, // Not compared here, so it still notifies.
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
			lookup.Status.Init(
				0, 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			lookup.Status.SetLatestVersion(tc.versions.initialLatestVersion, "", false)
			lookup.Status.SetDeployedVersion(tc.versions.initialDeployedVersion, "", false)

			notify, err := lookup.HandleNewVersion(tc.versions.newVersion, tc.versions.releaseDate, logFrom)
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
		name        string
		checkNumber int
		stdoutRegex string
	}{
		{
			name:        "the first check",
			checkNumber: 0,
			stdoutRegex: `^$`,
		},
		{
			name:        "the second check, which also logs that it is staying put",
			checkNumber: 1,
			stdoutRegex: `^VERBOSE: .*Staying on "1\.2\.3" as that's the latest version in the second check\n$`,
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
				logx.LogFrom{Primary: "TestLookup_HandleNoVersionChange", Secondary: tc.name},
			)

			stdout := releaseLog()
			if got, want := len(lookup.Status.AnnounceChannel), 1; got != want {
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
