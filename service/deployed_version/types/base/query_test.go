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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/dashboard"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_HandleNewVersion(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type versions struct {
		initialLatestVersion, initialDeployedVersion string
		newVersion, releaseDate                      string
	}

	// GIVEN: a Lookup.
	tests := []struct {
		name          string
		versions      versions
		wantAnnounces int
		wantNotify    bool
	}{
		{
			name: "first version found/no deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounces: 1,
			wantNotify:    false,
		},
		{
			name: "first version found/have newer deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "1.0.1",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounces: 1,
			wantNotify:    false,
		},
		{
			name: "first version found/have older deployed version",
			versions: versions{
				initialLatestVersion:   "",
				initialDeployedVersion: "0.9.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounces: 1,
			wantNotify:    false,
		},
		{
			name: "new version found",
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.1.0",
				releaseDate:            "2024-02-01",
			},
			wantAnnounces: 1,
			wantNotify:    true,
		},
		{
			name: "same version found", // shouldn't occur in practice.
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounces: 0,
			wantNotify:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logFrom := logx.LogFrom{Primary: "TestLookup_HandleNewVersion", Secondary: tc.name}
			lookup := &Lookup{
				Status: &status.Status{},
			}
			// Status.
			announceChannel := make(chan []byte, 2)
			lookup.Status.AnnounceChannel = announceChannel
			lookup.Status.Init(
				0, 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			lookup.Status.SetLatestVersion(tc.versions.initialLatestVersion, "", false)
			lookup.Status.SetDeployedVersion(tc.versions.initialDeployedVersion, "", false)
			// Options.
			lookup.Options, _ = opt.Decode(
				"yaml", nil,
				optCfg,
			)
			// Defaults.
			lookup.Defaults = dvCfg.Soft
			// HardDefaults.
			lookup.HardDefaults = dvCfg.Hard

			// WHEN: HandleNewVersion is called on it.
			lookup.HandleNewVersion(tc.versions.newVersion, tc.versions.releaseDate, true, logFrom)

			prefix := fmt.Sprintf(
				"%s\nLookup.HandleNewVersion(%q,%q)",
				packageName, tc.versions.newVersion, tc.versions.releaseDate,
			)

			// THEN: an announcement should be made when expected.
			if gotAnnounces := len(lookup.Status.AnnounceChannel); gotAnnounces != tc.wantAnnounces {
				t.Errorf(
					"%s Announce channel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotAnnounces, tc.wantAnnounces,
				)
			}

			// AND: the LatestVersion should not be changed.
			if got := lookup.Status.LatestVersion(); got != tc.versions.initialLatestVersion {
				t.Errorf(
					"%s .LatestVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.versions.initialLatestVersion,
				)
			}

			// AND: the DeployedVersion should be set to the new version if it was previously unset.
			if tc.versions.initialDeployedVersion == "" {
				if got := lookup.Status.DeployedVersion(); got != tc.versions.newVersion {
					t.Errorf(
						"%s .DeployedVersion() mismatch\ngot:  %q\nwant: %q",
						prefix, got, tc.versions.newVersion,
					)
				}
			}
		})
	}
}
