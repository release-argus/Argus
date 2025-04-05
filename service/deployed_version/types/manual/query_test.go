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

package manual

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestLookup_Track(t *testing.T) {
	// GIVEN a Lookup.
	lookup := testLookup("1.2.3", false)
	didFinish := make(chan bool, 1)

	// WHEN Track is called on it.
	go func() {
		lookup.Track()
		didFinish <- true
	}()
	time.Sleep(10 * time.Millisecond)

	// THEN the function exits straight away.
	if len(didFinish) == 0 {
		t.Fatalf("%s\nshould have exited immediately",
			packageName)
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		previousLatestVersion       string
		previousDeployedVersion     string
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
		announces                   int
	}{
		"No version": {
			overrides: test.TrimYAML(`
				version: ""
			`),
			announces: 0,
		},
		"Same version": {
			previousLatestVersion:   "3.2.1",
			previousDeployedVersion: "3.2.1",
			overrides: test.TrimYAML(`
				version: "3.2.1"
			`),
			announces: 0,
		},
		"Newer version": {
			previousLatestVersion:   "1.2.3",
			previousDeployedVersion: "1.2.3",
			overrides: test.TrimYAML(`
				version: "1.2.4"
			`),
			announces: 1,
		},
		"Older version": {
			previousLatestVersion:   "1.2.3",
			previousDeployedVersion: "1.2.3",
			overrides: test.TrimYAML(`
				version: "1.2.2"
			`),
			announces: 1,
		},
		"handle non-semantic (only major) version": {
			overrides: test.TrimYAML(`
				version: 1
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			announces: 1,
		},
		"want semantic versioning but get non-semantic version": {
			overrides: test.TrimYAML(`
				version: 1_2_3
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: true
			`),
			errRegex:  `failed to convert "[^"]+" to a semantic version`,
			announces: 0,
		},
		"allow non-semantic versioning and get non-semantic version": {
			overrides: test.TrimYAML(`
				version: 1_2_3
			`),
			optionsOverrides: test.TrimYAML(`
				semantic_versioning: false
			`),
			announces: 1,
		},
		"valid semantic version": {
			overrides: test.TrimYAML(`
				version: 1.2.3
			`),
			wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`,
			announces:   1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dvl := testLookup("", false)
			err := yaml.Unmarshal([]byte(tc.overrides), dvl)
			if err != nil {
				t.Fatalf("%s\nfailed to unmarshal overrides: %s",
					packageName, err)
			}
			err = yaml.Unmarshal([]byte(tc.optionsOverrides), dvl.Options)
			if err != nil {
				t.Fatalf("%s\nfailed to unmarshal options overrides: %s",
					packageName, err)
			}
			oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
			dvl.Status.SetLatestVersion(tc.previousLatestVersion, oneMinuteAgo, false)
			dvl.Status.SetDeployedVersion(tc.previousDeployedVersion, oneMinuteAgo, false)

			// WHEN Query is called on it.
			err = dvl.Query(true, logutil.LogFrom{})

			// THEN any err is expected.
			if tc.wantVersion != "" {
				version := dvl.Status.DeployedVersion()
				if !util.RegexCheck(tc.wantVersion, version) {
					t.Errorf("%s\nversion mismatch\nwant: %q\ngot:  %q",
						packageName, tc.wantVersion, version)
				}
			}
			e := util.ErrorToString(err)
			if tc.errRegex == "" {
				tc.errRegex = `^$`
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			// AND the Version is cleared.
			wantVersion := ""
			if dvl.Version != wantVersion {
				t.Errorf("%s\nVersion not cleared\nwant: %q\ngot:  %q",
					packageName, wantVersion, dvl.Version)
			}
			// AND the correct number of announces are queued.
			gotAnnounces := len(*dvl.Status.AnnounceChannel)
			if gotAnnounces != tc.announces {
				t.Errorf("%s\nannounce count mismatch\nwant: %d\ngot:  %d",
					packageName, tc.announces, gotAnnounces)
			}
		})
	}
}

func TestLookup_Query__RateLimit(t *testing.T) {
	// GIVEN a Lookup that has just had its DeployedVersion set.
	dvl := testLookup("", false)
	oneMinuteAgo := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
	dvl.Status.SetLatestVersion("1.2.3", oneMinuteAgo, false)
	dvl.Status.SetDeployedVersion("1.2.3", "", false)
	dvl.Version = "1.2.4"

	// WHEN Query is called on it.
	err := dvl.Query(true, logutil.LogFrom{})

	// THEN it errors with a rate-limit message.
	e := util.ErrorToString(err)
	if util.RegexCheck("^$", e) {
		t.Fatalf("%s\nExpected a rate-limit error, got %q",
			packageName, e)
	}
	// AND the Version is cleared.
	if dvl.Version != "" {
		t.Errorf("%s\nVersion not cleared: %q",
			packageName, dvl.Version)
	}
	// AND no announces are queued.
	wantAnnounces := 0
	gotAnnounces := len(*dvl.Status.AnnounceChannel)
	if gotAnnounces != wantAnnounces {
		t.Errorf("%s\nannounce count mismatch\nwant: %d\ngot:  %d",
			packageName, wantAnnounces, gotAnnounces)
	}
}
