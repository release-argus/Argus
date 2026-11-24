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

package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Track(t *testing.T) {
	tests := []struct {
		name                                      string
		content                                   string
		missingFile                               bool
		startDeployedVersion, wantDeployedVersion string
		wantAnnounces, wantDatabaseMessages       int
		deleting                                  bool
	}{
		{
			name:                 "get version from file",
			content:              "1.2.3\n",
			wantDeployedVersion:  "1.2.3",
			wantAnnounces:        1,
			wantDatabaseMessages: 1,
		},
		{
			name:                 "unchanged version",
			content:              "1.2.3\n",
			startDeployedVersion: "1.2.3",
			wantDeployedVersion:  "1.2.3",
			wantAnnounces:        0,
			wantDatabaseMessages: 0,
		},
		{
			name:        "missing file",
			missingFile: true,
		},
		{
			name:     "deleting service stops track",
			content:  "1.2.3\n",
			deleting: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.
			releaseLog := test.CaptureLog(t, logx.Default())

			lookup := testLookup(t, tc.content)
			if tc.missingFile {
				lookup.Path = filepath.Join(filepath.Dir(lookup.Path), "missing")
			}
			if err := decode.Unmarshal("yaml", []byte("interval: 2s"), lookup.Options); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			lookup.Status.ServiceInfo.ID = t.Name()
			lookup.Status.SetDeployedVersion(tc.startDeployedVersion, "", false)
			if tc.deleting {
				lookup.Status.SetDeleting()
			}
			lookup.InitMetrics(lookup)
			t.Cleanup(func() {
				lookup.Status.SetDeleting()
				lookup.DeleteMetrics(lookup)
			})

			didFinish := make(chan bool, 1)
			go func() {
				lookup.Track()
				didFinish <- true
			}()

			if tc.deleting {
				select {
				case <-didFinish:
				case <-time.After(time.Second):
					t.Fatal("Lookup.Track() didn't return within 1s")
				}
				if got, want := len(lookup.Status.AnnounceChannel), 0; got != want {
					t.Errorf(
						"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
						got, want,
					)
				}
				if got, want := len(lookup.Status.DatabaseChannel), 0; got != want {
					t.Errorf(
						"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
						got, want,
					)
				}
				return
			}

			serviceID := lookup.GetServiceID()
			queries := func() float64 {
				return testutil.ToFloat64(
					metric.DeployedVersionQueryResultTotal.WithLabelValues(
						serviceID, lookup.GetType(), metric.ActionResultSuccess,
					),
				) + testutil.ToFloat64(
					metric.DeployedVersionQueryResultTotal.WithLabelValues(
						serviceID, lookup.GetType(), metric.ActionResultFail,
					),
				)
			}
			deadline := time.Now().Add(10 * time.Second)
			for queries() == 0 {
				if time.Now().After(deadline) {
					t.Fatal("Lookup.Track() didn't query within 10s")
				}
				time.Sleep(50 * time.Millisecond)
			}
			_ = releaseLog()

			if got := lookup.Status.DeployedVersion(); got != tc.wantDeployedVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.wantDeployedVersion,
				)
			}

			if got := len(lookup.Status.AnnounceChannel); got != tc.wantAnnounces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantAnnounces,
				)
				test.LogChannel(t, "Status.AnnounceChannel", lookup.Status.AnnounceChannel)
			}
			if got := len(lookup.Status.DatabaseChannel); got != tc.wantDatabaseMessages {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantDatabaseMessages,
				)
			}
		})
	}
}

func TestLookup_Query(t *testing.T) {
	outsidePath := filepath.Join(t.TempDir(), "VERSION")
	if err := os.WriteFile(outsidePath, []byte("1.2.3"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	tests := []struct {
		name                        string
		content                     string
		previousDeployedVersion     string
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
		announces                   int
	}{
		{
			name:        "No version/empty",
			content:     "",
			wantVersion: `^$`,
			errRegex:    `^no version found in "[^"]+"$`,
			announces:   0,
		},
		{
			name:        "No version/whitespace only",
			content:     " \r\n\t",
			wantVersion: `^$`,
			errRegex:    `^no version found in "[^"]+"$`,
			announces:   0,
		},
		{
			name:                    "version/same",
			content:                 "3.2.1",
			previousDeployedVersion: "3.2.1",
			wantVersion:             `^3\.2\.1$`,
			errRegex:                `^$`,
			announces:               0,
		},
		{
			name:                    "version/newer",
			content:                 "1.2.4",
			previousDeployedVersion: "1.2.3",
			wantVersion:             `^1\.2\.4$`,
			errRegex:                `^$`,
			announces:               1,
		},
		{
			name:                    "version/older",
			content:                 "1.2.2",
			previousDeployedVersion: "1.2.3",
			wantVersion:             `^1\.2\.2$`,
			errRegex:                `^$`,
			announces:               1,
		},
		{
			name:        "file contents trimmed without regex",
			content:     "  1.2.3\n",
			wantVersion: `^1\.2\.3$`,
			errRegex:    `^$`,
			announces:   1,
		},
		{
			name:             "semantic versioning/fail on non-semantic version",
			content:          "1_2_3",
			optionsOverrides: `semantic_versioning: true`,
			wantVersion:      `^$`,
			errRegex:         `failed to convert "[^"]+" to a semantic version`,
			announces:        0,
		},
		{
			name:             "non-semantic versioning/accepted",
			content:          "1_2_3",
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^1_2_3$`,
			errRegex:         `^$`,
			announces:        1,
		},
		{
			name:        "regex/with capture group",
			content:     `productversion="1.2.3-54321"` + "\n",
			overrides:   `regex: 'productversion="([^"]+)"'`,
			wantVersion: `^1\.2\.3-54321$`,
			errRegex:    `^$`,
			announces:   1,
		},
		{
			name:        "regex/without capture group",
			content:     "release 1.2.3-54321",
			overrides:   `regex: '1\.2\.3-[0-9]+'`,
			wantVersion: `^1\.2\.3-54321$`,
			errRegex:    `^$`,
			announces:   1,
		},
		{
			name:    "regex/with template",
			content: `productversion="v1-2-3"` + "\n",
			overrides: test.TrimYAML(`
				regex: 'productversion="v([0-9]+)-([0-9]+)-([0-9]+)"'
				regex_template: '$1.$2.$3'
			`),
			wantVersion: `^1\.2\.3$`,
			errRegex:    `^$`,
			announces:   1,
		},
		{
			name:        "regex/no match",
			content:     "no version here\n",
			overrides:   `regex: 'productversion="([^"]+)"'`,
			wantVersion: `^$`,
			errRegex:    `^regex .* didn't return any matches on .*$`,
			announces:   0,
		},
		{
			name:        "regex/invalid",
			content:     "1.2.3",
			overrides:   `regex: '('`,
			wantVersion: `^$`,
			errRegex:    `^invalid regex "\(":\s+error parsing regexp: .*$`,
			announces:   0,
		},
		{
			name:        "regex/produced an empty version",
			content:     `productversion=""` + "\n",
			overrides:   `regex: 'productversion="([^"]*)"'`,
			wantVersion: `^$`,
			errRegex:    `^regex .* produced an empty version in .*$`,
			announces:   0,
		},
		{
			name:        "file/missing",
			overrides:   "path: " + filepath.Join(allowedDir, "missing", "VERSION"),
			wantVersion: `^$`,
			errRegex: test.TrimYAML(`
				^read deployed version from "[^"]+":
					stat file under allowed path:
						statat missing/VERSION:
							no such file or directory$`),
			announces: 0,
		},
		{
			name:        "file/outside file_read_paths",
			overrides:   "path: " + outsidePath,
			wantVersion: `^$`,
			errRegex: test.TrimYAML(`
				^read deployed version from "[^"]+":
					file path is outside configured settings\.data\.file_read_paths$`),
			announces: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dvl := testLookup(t, tc.content)
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
			dvl.Status.SetDeployedVersion(tc.previousDeployedVersion, oneMinuteAgo, false)

			err := dvl.Query(true, logx.LogFrom{})
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

			if got := len(dvl.Status.AnnounceChannel); got != tc.announces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.announces,
				)
			}
		})
	}
}
