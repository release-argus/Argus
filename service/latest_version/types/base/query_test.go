// Copyright [2024] [Argus]
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

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestVerifySemanticVersioning(t *testing.T) {
	type versions struct {
		new      string
		current  string
		deployed string
	}

	// GIVEN a Lookup and a set of versions
	tests := map[string]struct {
		versions versions
		errRegex string
	}{
		"valid semantic version, no current version": {
			versions: versions{
				new:     "1.2.3",
				current: ""},
			errRegex: `^$`,
		},
		"invalid semantic version": {
			versions: versions{
				new:     "invalid-version",
				current: ""},
			errRegex: `failed converting "invalid-version" to a semantic version`,
		},
		"progressive check - valid semantic version, non-semantic deployed version": {
			versions: versions{
				new:      "1.2.4",
				current:  "1.2.3",
				deployed: "non-semantic"},
			errRegex: `^$`,
		},
		"progressive check - valid semantic version, newer than current version": {
			versions: versions{
				new:      "1.2.4",
				current:  "1.2.3",
				deployed: "1.2.3"},
			errRegex: `^$`,
		},
		"progressive check - valid semantic version, older than current version": {
			versions: versions{
				new:      "1.2.2",
				current:  "1.2.3",
				deployed: "1.2.3"},
			errRegex: `queried version "1.2.2" is less than the deployed version "1.2.3"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logFrom := util.LogFrom{Primary: "TestVerifySemanticVersioning", Secondary: name}
			lookup := &Lookup{
				Status: &status.Status{},
			}
			lookup.Status.SetLatestVersion(tc.versions.current, "", false)
			lookup.Status.SetDeployedVersion(tc.versions.deployed, "", false)

			// WHEN VerifySemanticVersioning is called
			err := lookup.VerifySemanticVersioning(tc.versions.new, tc.versions.current, logFrom)

			// THEN the error message should match the expected regex
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Expected error to match regex\n%q\ngot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestLookup_HandleNewVersion(t *testing.T) {
	type versions struct {
		initialLatestVersion, initialDeployedVersion string
		newVersion, releaseDate                      string
	}

	// GIVEN a Lookup
	tests := map[string]struct {
		versions     versions
		wantAnnounce bool
		wantNotify   bool
	}{
		"first version found, no deployed version": {
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantAnnounce: true,
			wantNotify:   false,
		},
		"first version found, have newer deployed version": {
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "1.0.1",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantAnnounce: true,
			wantNotify:   false,
		},
		"first version found, have older deployed version": {
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "0.9.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantAnnounce: true,
			wantNotify:   false,
		},
		"new version found": {
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.1.0",
				releaseDate:            "2024-02-01"},
			wantNotify: true,
		},
		"same version found": { // shouldn't occur in practice
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantNotify: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logFrom := util.LogFrom{Primary: "TestLookup_HandleNewVersion", Secondary: name}
			lookup := &Lookup{
				Status: &status.Status{},
			}
			announceChan := make(chan []byte, 2)
			lookup.Status.AnnounceChannel = &announceChan
			lookup.Status.ServiceID = &name
			lookup.Status.SetLatestVersion(tc.versions.initialLatestVersion, "", false)
			lookup.Status.SetDeployedVersion(tc.versions.initialDeployedVersion, "", false)

			// WHEN HandleNewVersion is called on it
			notify, err := lookup.HandleNewVersion(tc.versions.newVersion, tc.versions.releaseDate, logFrom)

			// THEN the error should always be nil
			if err != nil {
				t.Errorf("Expected error to be nil, got %v",
					err)
			}
			// AND the notify value should match the expected value
			if notify != tc.wantNotify {
				t.Errorf("notify bool mismatch\nwant: %t\ngot:  %t",
					tc.wantNotify, notify)
			}
			// AND an announcement should be made if expected
			wantLen := 0
			if tc.wantAnnounce {
				wantLen = 1
			}
			gotLen := len(*lookup.Status.AnnounceChannel)
			if wantLen != gotLen {
				t.Errorf("Announcement channel length mismatch\nwant: %d\ngot:  %d",
					wantLen, gotLen)
			}
			// AND the LatestVersion should be set to the new version
			if lookup.Status.LatestVersion() != tc.versions.newVersion {
				t.Errorf("LatestVersion mismatch\nwant: %q\ngot:  %q",
					tc.versions.newVersion, lookup.Status.LatestVersion())
			}
			// AND the DeployedVersion should be set to the new version if it was previously unset
			if tc.versions.initialDeployedVersion == "" &&
				lookup.Status.DeployedVersion() != tc.versions.newVersion {
				t.Errorf("DeployedVersion mismatch\nwant: %q\ngot:  %q",
					tc.versions.newVersion, lookup.Status.DeployedVersion())
			}
		})
	}
}
