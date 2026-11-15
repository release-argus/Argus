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
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/dashboard"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_HandleNewVersion(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()

	type versions struct {
		initialLatestVersion, initialDeployedVersion string
		newVersion, releaseDate                      string
	}

	tests := []struct {
		name          string
		versions      versions
		wantAnnounces int
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
		},
		{
			name: "same version found - shouldn't occur in practice",
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantAnnounces: 0,
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
			lookup.Options = opttest.PlainOptions(optCfg)
			lookup.Defaults = dvCfg.Soft
			lookup.HardDefaults = dvCfg.Hard

			lookup.HandleNewVersion(
				tc.versions.newVersion,
				tc.versions.releaseDate,
				true,
				true,
				logFrom,
			)

			if gotAnnounces := len(lookup.Status.AnnounceChannel); gotAnnounces != tc.wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotAnnounces, tc.wantAnnounces,
				)
			}

			if got := lookup.Status.LatestVersion(); got != tc.versions.initialLatestVersion {
				t.Errorf(
					"Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.versions.initialLatestVersion,
				)
			}

			if got := lookup.Status.DeployedVersion(); got != tc.versions.newVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.versions.newVersion,
				)
			}
		})
	}
}
