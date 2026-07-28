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

package manual

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Track(t *testing.T) {
	// GIVEN: a Lookup.
	lookup := testLookup(t, "1.2.3")
	didFinish := make(chan bool, 1)

	// WHEN: Track is called on it.
	go func() {
		lookup.Track()
		didFinish <- true
	}()
	time.Sleep(10 * time.Millisecond)

	prefix := fmt.Sprintf("%s\nLookup.Track()", packageName)

	// THEN: the function exits straight away.
	if len(didFinish) == 0 {
		t.Fatalf("%s should have exited immediately", prefix)
	}

	// AND: the configured version was applied.
	if got, want := lookup.Status.DeployedVersion(), "1.2.3"; got != want {
		t.Errorf(
			"%s .DeployedVersion() mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
}

func TestLookup_ApplyConfiguredVersion(t *testing.T) {
	// GIVEN: a Lookup with a version set in the config.
	tests := []struct {
		name                              string
		prevVersion, version, wantVersion string
		optionsOverrides                  string
		wantDBMessages                    int
		wantSaves                         int
	}{
		{
			name:           "no version, nothing to apply",
			prevVersion:    "",
			version:        "",
			wantVersion:    "",
			wantDBMessages: 0,
			wantSaves:      0,
		},
		{
			name:           "version is applied and persisted",
			version:        "1.2.3",
			wantVersion:    "1.2.3",
			wantDBMessages: 1,
			wantSaves:      1,
		},
		{
			name:           "version replacing an existing one is persisted",
			prevVersion:    "1.2.3",
			version:        "1.2.4",
			wantVersion:    "1.2.4",
			wantDBMessages: 1,
			wantSaves:      1,
		},
		{
			name:        "unchanged version writes nothing",
			version:     "1.2.3",
			prevVersion: "1.2.3",
			wantVersion: "1.2.3",
			// The Version was still consumed, so the config is saved without it.
			wantSaves: 1,
		},
		{
			name:        "non-semantic version is not applied",
			version:     "'1_2_3'",
			wantVersion: "",
			wantSaves:   0,
		},
		{
			name:             "non-semantic version applied when semantic versioning is off",
			version:          "'1_2_3'",
			optionsOverrides: "semantic_versioning: false",
			wantVersion:      "1_2_3",
			wantDBMessages:   1,
			wantSaves:        1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nLookup.ApplyConfiguredVersion()", packageName)

			dvl := testLookup(t, tc.version)
			if len(tc.optionsOverrides) != 0 {
				if err := decode.Unmarshal(
					"yaml", []byte(tc.optionsOverrides),
					dvl.Options,
				); err != nil {
					t.Fatalf(
						"%s failed to unmarshal Lookup.Options overrides: %s",
						prefix, err,
					)
				}
			}
			oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
			dvl.Status.SetDeployedVersion(tc.prevVersion, oneMinuteAgo, false)

			// WHEN: ApplyConfiguredVersion is called on it.
			dvl.ApplyConfiguredVersion()

			// THEN: the version reaches the Status.
			if got := dvl.Status.DeployedVersion(); got != tc.wantVersion {
				t.Errorf(
					"%s .DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantVersion,
				)
			}

			// AND: the Version is cleared.
			if dvl.Version != "" {
				t.Errorf(
					"%s .Version not cleared\ngot:  %q\nwant: %q",
					prefix, dvl.Version, "",
				)
			}

			// AND: it is persisted to the DB, so that it survives a restart.
			if got := len(dvl.Status.DatabaseChannel); got != tc.wantDBMessages {
				t.Errorf(
					"%s Database message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantDBMessages,
				)
			} else if tc.wantDBMessages > 0 {
				message := <-dvl.Status.DatabaseChannel
				var gotCell string
				for _, cell := range message.Cells {
					if cell.Column == "deployed_version" {
						gotCell = cell.Value
					}
				}
				if gotCell != tc.wantVersion {
					t.Errorf(
						"%s `deployed_version` cell mismatch\ngot:  %q\nwant: %q",
						prefix, gotCell, tc.wantVersion,
					)
				}
			}

			// AND: it is never announced - the caller broadcasts the change itself.
			if got := len(dvl.Status.AnnounceChannel); got != 0 {
				t.Errorf(
					"%s Announce message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, 0,
				)
			}

			// AND: a save is queued when the Version was consumed, so that the config
			// file stops declaring it and cannot re-apply it after a restart.
			if got := len(dvl.Status.SaveChannel); got != tc.wantSaves {
				t.Errorf(
					"%s Save message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantSaves,
				)
			}
		})
	}
}

func TestLookup_ApplyConfiguredVersion__notRateLimited(t *testing.T) {
	// GIVEN: a Lookup that has just had its DeployedVersion set.
	dvl := testLookup(t, "1.2.4")
	dvl.Status.SetDeployedVersion("1.2.3", "", false)

	// WHEN: ApplyConfiguredVersion is called on it.
	dvl.ApplyConfiguredVersion()

	prefix := fmt.Sprintf("%s\nLookup.ApplyConfiguredVersion()", packageName)

	// THEN: the rate-limit that guards Query does not apply, so the version is set.
	if got, want := dvl.Status.DeployedVersion(), "1.2.4"; got != want {
		t.Errorf(
			"%s .DeployedVersion() mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// AND: it is persisted to the DB.
	if got, want := len(dvl.Status.DatabaseChannel), 1; got != want {
		t.Errorf(
			"%s Database message count mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: no announces are queued.
	if got, want := len(dvl.Status.AnnounceChannel), 0; got != want {
		t.Errorf(
			"%s Announce message count mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                        string
		previousLatestVersion       string
		previousDeployedVersion     string
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
		announces                   int
	}{
		{
			name:      "No version",
			overrides: `version: ""`,
			errRegex:  `^$`,
			announces: 0,
		},
		{
			name:                    "Inherit version",
			previousLatestVersion:   "3.2.1",
			previousDeployedVersion: "3.2.1",
			overrides:               `version: "3.2.1"`,
			errRegex:                `^$`,
			announces:               0,
		},
		{
			name:                    "Newer version",
			previousLatestVersion:   "1.2.3",
			previousDeployedVersion: "1.2.3",
			overrides:               `version: "1.2.4"`,
			errRegex:                `^$`,
			announces:               1,
		},
		{
			name:                    "Older version",
			previousLatestVersion:   "1.2.3",
			previousDeployedVersion: "1.2.3",
			overrides:               `version: "1.2.2"`,
			errRegex:                `^$`,
			announces:               1,
		},
		{
			name:             "handle non-semantic (only major) version",
			overrides:        `version: 1`,
			optionsOverrides: `semantic_versioning: false`,
			errRegex:         `^$`,
			announces:        1,
		},
		{
			name:             "want semantic versioning but get non-semantic version",
			overrides:        `version: "1_2_3"`,
			optionsOverrides: `semantic_versioning: true`,
			errRegex:         `failed to convert "[^"]+" to a semantic version`,
			announces:        0,
		},
		{
			name:             "allow non-semantic versioning and get non-semantic version",
			overrides:        `version: "1_2_3"`,
			optionsOverrides: `semantic_versioning: false`,
			errRegex:         `^$`,
			announces:        1,
		},
		{
			name:        "valid semantic version",
			overrides:   `version: 1.2.3`,
			wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`,
			errRegex:    `^$`,
			announces:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

			dvl := testLookup(t, "")
			if err := dvl.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s failed to unmarshal Lookup overrides: %s",
					prefix, err,
				)
			}
			if len(tc.optionsOverrides) != 0 {
				if err := decode.Unmarshal(
					"yaml", []byte(tc.optionsOverrides),
					dvl.Options,
				); err != nil {
					t.Fatalf(
						"%s failed to unmarshal Lookup.Options overrides: %s",
						prefix, err,
					)
				}
			}
			oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
			dvl.Status.SetLatestVersion(tc.previousLatestVersion, oneMinuteAgo, false)
			dvl.Status.SetDeployedVersion(tc.previousDeployedVersion, oneMinuteAgo, false)

			// WHEN: Query is called on it.
			err := dvl.Query(true, logx.LogFrom{})

			// THEN: the error is as expected.
			if tc.wantVersion != "" {
				if version := dvl.Status.DeployedVersion(); !util.RegexCheck(tc.wantVersion, version) {
					t.Errorf(
						"%s .DeployedVersion() value mismatch\ngot:  %q\nwant: %q",
						prefix, version, tc.wantVersion,
					)
				}
			}
			e := errfmt.FormatError(err)
			if tc.errRegex == "" {
				tc.errRegex = `^$`
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the Version is cleared.
			wantVersion := ""
			if dvl.Version != wantVersion {
				t.Errorf(
					"%s Lookup.Version not cleared\ngot:  %q\nwant: %q",
					prefix, dvl.Version, wantVersion,
				)
			}

			// AND: the correct number of announces are queued.
			if got := len(dvl.Status.AnnounceChannel); got != tc.announces {
				t.Errorf(
					"%s Announce message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.announces,
				)
			}
		})
	}
}

func TestLookup_Query__rateLimit(t *testing.T) {
	// GIVEN: a Lookup that has just had its DeployedVersion set.
	dvl := testLookup(t, "")
	oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
	dvl.Status.SetLatestVersion("1.2.3", oneMinuteAgo, false)
	dvl.Status.SetDeployedVersion("1.2.3", "", false)
	dvl.Version = "1.2.4"

	// WHEN: Query is called on it.
	err := dvl.Query(true, logx.LogFrom{})

	prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

	// THEN: it errors with a rate-limit message.
	e := errfmt.FormatError(err)
	if util.RegexCheck("^$", e) {
		t.Fatalf(
			"%s expected a rate-limit error, got %q",
			prefix, e,
		)
	}

	// AND: the Version is cleared.
	if dvl.Version != "" {
		t.Errorf(
			"%s .Version not cleared: %q",
			prefix, dvl.Version,
		)
	}

	// AND: no announces are queued.
	wantAnnounces := 0
	gotAnnounces := len(dvl.Status.AnnounceChannel)
	if gotAnnounces != wantAnnounces {
		t.Errorf(
			"%s Announce message count mismatch\ngot:  %d\nwant: %d",
			prefix, gotAnnounces, wantAnnounces,
		)
	}
}
