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
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Track(t *testing.T) {
	lookup := testLookup(t, "1.2.3")
	didFinish := make(chan bool, 1)

	go func() {
		lookup.Track()
		didFinish <- true
	}()

	select {
	case <-didFinish:
	case <-time.After(time.Second):
		t.Fatal("Lookup.Track() didn't return within 1s")
	}

	if got, want := lookup.Status.DeployedVersion(), "1.2.3"; got != want {
		t.Errorf("Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestLookup_ApplyConfiguredVersion(t *testing.T) {
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
			name:           "unchanged version is not written to the DB",
			version:        "1.2.3",
			prevVersion:    "1.2.3",
			wantVersion:    "1.2.3",
			wantDBMessages: 0,
			wantSaves:      1, // The Version was still consumed, so the config is saved without it.
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

			dvl := testLookup(t, tc.version)
			if len(tc.optionsOverrides) != 0 {
				if err := decode.Unmarshal(
					"yaml", []byte(tc.optionsOverrides),
					dvl.Options,
				); err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}
			}
			oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
			dvl.Status.SetDeployedVersion(tc.prevVersion, oneMinuteAgo, false)

			dvl.ApplyConfiguredVersion()

			if got := dvl.Status.DeployedVersion(); got != tc.wantVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.wantVersion,
				)
			}

			if got, want := dvl.Version, ""; got != want {
				t.Errorf("Lookup.Version mismatch\ngot:  %q\nwant: %q (consumed)", got, want)
			}

			if got := len(dvl.Status.DatabaseChannel); got != tc.wantDBMessages {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantDBMessages,
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
						"Status.DatabaseChannel deployed_version cell mismatch\ngot:  %q\nwant: %q",
						gotCell, tc.wantVersion,
					)
				}
			}

			if got, want := len(dvl.Status.AnnounceChannel), 0; got != want {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}

			// A consumed Version is saved so the config file cannot re-apply it after a restart.
			if got := len(dvl.Status.SaveChannel); got != tc.wantSaves {
				t.Errorf(
					"Status.SaveChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantSaves,
				)
			}
		})
	}
}

func TestLookup_ApplyConfiguredVersion__notRateLimited(t *testing.T) {
	dvl := testLookup(t, "1.2.4")
	// An empty timestamp stamps it now, inside the rate-limit window that guards Query.
	dvl.Status.SetDeployedVersion("1.2.3", "", false)

	dvl.ApplyConfiguredVersion()

	if got, want := dvl.Status.DeployedVersion(), "1.2.4"; got != want {
		t.Errorf("Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q", got, want)
	}

	if got, want := len(dvl.Status.DatabaseChannel), 1; got != want {
		t.Errorf("Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d", got, want)
	}

	if got, want := len(dvl.Status.AnnounceChannel), 0; got != want {
		t.Errorf("Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestLookup_Query(t *testing.T) {
	tests := []struct {
		name                        string
		previousLatestVersion       string
		previousDeployedVersion     string
		overrides, optionsOverrides string
		errRegex                    string
		writeToDB                   bool
		wantVersion                 string
		announces                   int
	}{
		{
			name:        "No version",
			overrides:   `version: ""`,
			wantVersion: `^$`,
			writeToDB:   true,
			errRegex:    `^$`,
			announces:   0,
		},
		{
			name:                    "Inherit version",
			previousLatestVersion:   "3.2.1",
			previousDeployedVersion: "3.2.1",
			overrides:               `version: "3.2.1"`,
			wantVersion:             `^3\.2\.1$`,
			writeToDB:               true,
			errRegex:                `^$`,
			announces:               0,
		},
		{
			name:                    "Newer version",
			previousLatestVersion:   "1.2.3",
			previousDeployedVersion: "1.2.3",
			overrides:               `version: "1.2.4"`,
			wantVersion:             `^1\.2\.4$`,
			writeToDB:               true,
			errRegex:                `^$`,
			announces:               1,
		},
		{
			name:                    "Older version",
			previousLatestVersion:   "1.2.3",
			previousDeployedVersion: "1.2.3",
			overrides:               `version: "1.2.2"`,
			wantVersion:             `^1\.2\.2$`,
			writeToDB:               true,
			errRegex:                `^$`,
			announces:               1,
		},
		{
			name:             "handle non-semantic (only major) version",
			overrides:        `version: 1`,
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^1$`,
			writeToDB:        true,
			errRegex:         `^$`,
			announces:        1,
		},
		{
			name:             "want semantic versioning but get non-semantic version",
			overrides:        `version: "1_2_3"`,
			optionsOverrides: `semantic_versioning: true`,
			wantVersion:      `^$`,
			writeToDB:        true,
			errRegex:         `failed to convert "[^"]+" to a semantic version`,
			announces:        0,
		},
		{
			name:             "allow non-semantic versioning and get non-semantic version",
			overrides:        `version: "1_2_3"`,
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^1_2_3$`,
			writeToDB:        true,
			errRegex:         `^$`,
			announces:        1,
		},
		{
			name:                    "Newer version/no DB write",
			previousLatestVersion:   "1.2.3",
			previousDeployedVersion: "1.2.3",
			overrides:               `version: "1.2.4"`,
			wantVersion:             `^1\.2\.4$`,
			errRegex:                `^$`,
			announces:               0,
		},
		{
			name:        "valid semantic version",
			overrides:   `version: 1.2.3`,
			wantVersion: `^1\.2\.3$`,
			writeToDB:   true,
			errRegex:    `^$`,
			announces:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dvl := testLookup(t, "")
			if err := dvl.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
			}
			if len(tc.optionsOverrides) != 0 {
				if err := decode.Unmarshal(
					"yaml", []byte(tc.optionsOverrides),
					dvl.Options,
				); err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}
			}
			oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
			dvl.Status.SetLatestVersion(tc.previousLatestVersion, oneMinuteAgo, false)
			dvl.Status.SetDeployedVersion(tc.previousDeployedVersion, oneMinuteAgo, false)

			err := dvl.Query(tc.writeToDB, logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if version := dvl.Status.DeployedVersion(); !util.RegexCheck(tc.wantVersion, version) {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					version, tc.wantVersion,
				)
			}

			if got, want := dvl.Version, ""; got != want {
				t.Errorf("Lookup.Version mismatch\ngot:  %q\nwant: %q (consumed)", got, want)
			}

			if got := len(dvl.Status.AnnounceChannel); got != tc.announces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.announces,
				)
			}
		})
	}
}

func TestLookup_Query__rateLimit(t *testing.T) {
	dvl := testLookup(t, "")
	oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
	dvl.Status.SetLatestVersion("1.2.3", oneMinuteAgo, false)
	// An empty timestamp stamps it now, inside the rate-limit window.
	dvl.Status.SetDeployedVersion("1.2.3", "", false)
	dvl.Version = "1.2.4"

	err := dvl.Query(true, logx.LogFrom{})
	e := errfmt.FormatError(err)
	if wantErrRegex := `rate-limited`; !util.RegexCheck(wantErrRegex, e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, wantErrRegex)
	}

	if got, want := dvl.Version, ""; got != want {
		t.Errorf("Lookup.Version mismatch\ngot:  %q\nwant: %q (consumed)", got, want)
	}

	if got, want := len(dvl.Status.AnnounceChannel), 0; got != want {
		t.Errorf("Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d", got, want)
	}
}
