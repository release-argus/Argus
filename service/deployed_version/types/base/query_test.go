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

//go:build unit

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"testing"

	"github.com/release-argus/Argus/service/dashboard"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestVerifySemanticVersioning(t *testing.T) {
	type versions struct {
		new      string
		current  string
		deployed string
	}

	// GIVEN a Lookup and a set of versions.
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
			errRegex: `failed to convert "invalid-version" to a semantic version`,
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
			errRegex: `^$`, // Can downgrade.
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logFrom := logutil.LogFrom{Primary: "TestVerifySemanticVersioning", Secondary: name}
			lookup := &Lookup{
				Status: &status.Status{}}
			lookup.Status.Init(
				0, 0, 0,
				"", "", "",
				&dashboard.Options{})
			lookup.Status.SetLatestVersion(tc.versions.current, "", false)
			lookup.Status.SetDeployedVersion(tc.versions.deployed, "", false)

			// WHEN VerifySemanticVersioning is called.
			err := lookup.VerifySemanticVersioning(tc.versions.new, tc.versions.current, logFrom)

			// THEN the error message should match the expected regex.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestLookup_HandleNewVersion(t *testing.T) {
	type versions struct {
		initialLatestVersion, initialDeployedVersion string
		newVersion, releaseDate                      string
	}

	// GIVEN a Lookup.
	tests := map[string]struct {
		versions      versions
		wantAnnounces int
		wantNotify    bool
	}{
		"first version found, no deployed version": {
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantAnnounces: 1,
			wantNotify:    false,
		},
		"first version found, have newer deployed version": {
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "1.0.1",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantAnnounces: 1,
			wantNotify:    false,
		},
		"first version found, have older deployed version": {
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "0.9.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantAnnounces: 1,
			wantNotify:    false,
		},
		"new version found": {
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.1.0",
				releaseDate:            "2024-02-01"},
			wantAnnounces: 1,
			wantNotify:    true,
		},
		"same version found": { // shouldn't occur in practice.
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01"},
			wantAnnounces: 0,
			wantNotify:    true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logFrom := logutil.LogFrom{Primary: "TestLookup_HandleNewVersion", Secondary: name}
			lookup := &Lookup{
				Status: &status.Status{},
			}
			// Status.
			announceChannel := make(chan []byte, 2)
			lookup.Status.AnnounceChannel = announceChannel
			lookup.Status.Init(
				0, 0, 0,
				name, "", "",
				&dashboard.Options{})
			lookup.Status.SetLatestVersion(tc.versions.initialLatestVersion, "", false)
			lookup.Status.SetDeployedVersion(tc.versions.initialDeployedVersion, "", false)
			// Options.
			lookup.Options = opt.New(
				nil, "", nil,
				&opt.Defaults{}, &opt.Defaults{})
			lookup.Options.HardDefaults.Default()
			// Defaults.
			defaults := &Defaults{}
			lookup.Defaults = defaults
			// HardDefaults.
			hardDefaults := &Defaults{}
			hardDefaults.Default()
			lookup.HardDefaults = hardDefaults

			// WHEN HandleNewVersion is called on it.
			lookup.HandleNewVersion(tc.versions.newVersion, tc.versions.releaseDate, true, logFrom)

			// THEN an announcement should be made when expected.
			gotLen := len(lookup.Status.AnnounceChannel)
			if gotLen != tc.wantAnnounces {
				t.Errorf("%s\nAnnounce channel length mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wantAnnounces, gotLen)
			}
			// AND the LatestVersion should not be changed.
			if tc.versions.initialLatestVersion != lookup.Status.LatestVersion() {
				t.Errorf("%s\nLatestVersion mismatch\nwant: %q\ngot:  %q",
					packageName, tc.versions.initialLatestVersion, lookup.Status.LatestVersion())
			}
			// AND the DeployedVersion should be set to the new version if it was previously unset.
			if tc.versions.initialDeployedVersion == "" &&
				tc.versions.newVersion != lookup.Status.DeployedVersion() {
				t.Errorf("%s\nDeployedVersion mismatch\nwant: %q\ngot:  %q",
					packageName, tc.versions.newVersion, lookup.Status.DeployedVersion())
			}
		})
	}
}
