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

	// THEN: the function exits straight away.
	if len(didFinish) == 0 {
		t.Fatalf("%s\nLookup Track() should have exited immediately", packageName)
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
