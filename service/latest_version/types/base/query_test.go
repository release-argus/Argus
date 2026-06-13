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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_HandleNewVersion(t *testing.T) {
	type versions struct {
		initialLatestVersion, initialDeployedVersion string
		newVersion, releaseDate                      string
	}

	// GIVEN: a Lookup.
	tests := []struct {
		name         string
		versions     versions
		wantAnnounce bool
		wantNotify   bool
	}{
		{
			name: "first version found, no deployed version",
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
			name: "first version found, have newer deployed version",
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
			name: "first version found, have older deployed version",
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
			name: "same version found", // shouldn't occur in practice.
			versions: versions{
				initialLatestVersion:   "1.0.0",
				initialDeployedVersion: "1.0.0",
				newVersion:             "1.0.0",
				releaseDate:            "2024-01-01",
			},
			wantNotify: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logFrom := logx.LogFrom{Primary: "TestLookup_HandleNewVersion", Secondary: tc.name}
			lookup := &Lookup{
				Status: &status.Status{},
			}
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

			// WHEN: HandleNewVersion is called on it.
			notify, err := lookup.HandleNewVersion(tc.versions.newVersion, tc.versions.releaseDate, logFrom)

			prefix := fmt.Sprintf(
				"%s\nLookup.HandleNewVersion(version=%q, releaseDate=%s)",
				packageName, tc.versions.newVersion, tc.versions.releaseDate,
			)

			// THEN: the error should always be nil.
			if err != nil {
				t.Errorf(
					"%s unexpected error\n%v",
					prefix, err,
				)
			}

			// AND: the notify value should match the expected value.
			if notify != tc.wantNotify {
				t.Errorf(
					"%s notify bool mismatch\ngot:  %t\nwant: %t",
					prefix, notify, tc.wantNotify,
				)
			}

			// AND: an announcement should be made if expected.
			wantLen := 0
			if tc.wantAnnounce {
				wantLen = 1
			}
			if gotLen := len(lookup.Status.AnnounceChannel); gotLen != wantLen {
				t.Errorf(
					"%s Announce channel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}

			// AND: the LatestVersion should be set to the new version.
			if got := lookup.Status.LatestVersion(); got != tc.versions.newVersion {
				t.Errorf(
					"%s LatestVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.versions.newVersion,
				)
			}

			// AND: the DeployedVersion should be set to the new version if it was previously unset.
			if got := lookup.Status.DeployedVersion(); got != tc.versions.newVersion && tc.versions.initialDeployedVersion == "" {
				t.Errorf(
					"%s DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.versions.newVersion,
				)
			}
		})
	}
}
