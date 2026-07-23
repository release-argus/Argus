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

package deployedver

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestRefresh__manual(t *testing.T) {
	// An hour ago, to stay clear of the rate-limit on manual updates.
	anHourAgo := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)

	type args struct {
		overrides                []byte
		previousType             string
		deployedVersion          string
		deployedVersionTimestamp string
	}
	type wants struct {
		version          string // Returned by Refresh.
		statusVersion    string
		dbMessages       int
		announces        int
		timestampChanged bool
	}
	// A refresh consumes the Version on its own copy of the Lookup, so it never needs
	// to save the config.
	const wantSaves = 0

	// GIVEN: a `manual` Lookup, and overrides for a deployed_version/refresh.
	tests := []struct {
		name     string
		args     args
		errRegex string
		wants    wants
	}{
		{
			name: "version override is persisted and announced",
			args: args{
				overrides: []byte(`{"type": "manual", "version": "7.8.9"}`),
			},
			errRegex: `^$`,
			wants: wants{
				version:          "7.8.9",
				statusVersion:    "7.8.9",
				dbMessages:       1,
				announces:        1,
				timestampChanged: true,
			},
		},
		{
			name: "version override without a type is persisted and announced",
			args: args{
				overrides: []byte(`{"version": "7.8.9"}`),
			},
			errRegex: `^$`,
			wants: wants{
				version:          "7.8.9",
				statusVersion:    "7.8.9",
				dbMessages:       1,
				announces:        1,
				timestampChanged: true,
			},
		},
		{
			name: "version override replacing an existing version is persisted",
			args: args{
				overrides:                []byte(`{"version": "7.8.9"}`),
				deployedVersion:          "6.7.8",
				deployedVersionTimestamp: anHourAgo,
			},
			errRegex: `^$`,
			wants: wants{
				version:          "7.8.9",
				statusVersion:    "7.8.9",
				dbMessages:       1,
				announces:        1,
				timestampChanged: true,
			},
		},
		{
			name: "unchanged version writes nothing",
			args: args{
				overrides:                []byte(`{"version": "7.8.9"}`),
				deployedVersion:          "7.8.9",
				deployedVersionTimestamp: anHourAgo,
			},
			errRegex: `^$`,
			wants: wants{
				version:       "7.8.9",
				statusVersion: "7.8.9",
			},
		},
		{
			name: "no overrides writes nothing",
			args: args{
				deployedVersion:          "7.8.9",
				deployedVersionTimestamp: anHourAgo,
			},
			errRegex: `^$`,
			wants: wants{
				version:       "7.8.9",
				statusVersion: "7.8.9",
			},
		},
		{
			name: "version override switching type is not persisted",
			args: args{
				overrides:    []byte(`{"type": "manual", "version": "7.8.9"}`),
				previousType: web.Type,
			},
			errRegex: `^$`,
			wants: wants{
				version: "7.8.9",
			},
		},
		{
			name: "switching type away from manual still validates",
			args: args{
				overrides:    []byte(`{"type": "url", "url": ""}`),
				previousType: manual.Type,
			},
			errRegex: `url: <required>`,
		},
		{
			name: "non-semantic version is rejected",
			args: args{
				overrides: []byte(`{"version": "not-a-version"}`),
			},
			errRegex: `failed to convert "not-a-version" to a semantic version`,
		},
		{
			name: "update within the rate-limit is rejected",
			args: args{
				overrides:                []byte(`{"version": "7.8.9"}`),
				deployedVersion:          "6.7.8",
				deployedVersionTimestamp: time.Now().UTC().Add(time.Second).Format(time.RFC3339),
			},
			errRegex: `rate-limited`,
			wants: wants{
				statusVersion: "6.7.8",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			previousType := util.ValueOr(tc.args.previousType, manual.Type)
			lookup := testLookup(t, previousType, false, "")
			svcStatus := lookup.GetStatus()
			if tc.args.deployedVersion != "" {
				svcStatus.SetDeployedVersion(
					tc.args.deployedVersion,
					tc.args.deployedVersionTimestamp,
					false,
				)
			}
			hadTimestamp := svcStatus.DeployedVersionTimestamp()

			// WHEN: Refresh is called with those overrides.
			got, err := Refresh(
				lookup,
				previousType,
				tc.args.overrides,
				nil,
				nil,
			)

			prefix := fmt.Sprintf("%s\nRefresh()", packageName)

			// THEN: we get an error only when expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the version found is returned to the caller.
			if err == nil && got != tc.wants.version {
				t.Errorf("%s mismatch on version returned\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.version,
				)
			}

			// AND: only a commit reaches the Service's Status.
			if gotVersion := svcStatus.DeployedVersion(); gotVersion != tc.wants.statusVersion {
				t.Errorf("%s mismatch on Status.DeployedVersion()\ngot:  %q\nwant: %q",
					prefix, gotVersion, tc.wants.statusVersion,
				)
			}

			// AND: it reaches the database, so that it survives a restart.
			if gotDBMessages := len(svcStatus.DatabaseChannel); gotDBMessages != tc.wants.dbMessages {
				t.Errorf("%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotDBMessages, tc.wants.dbMessages,
				)
			} else if gotDBMessages > 0 {
				message := <-svcStatus.DatabaseChannel
				var gotCell string
				for _, cell := range message.Cells {
					if cell.Column == "deployed_version" {
						gotCell = cell.Value
					}
				}
				if gotCell != tc.wants.version {
					t.Errorf("%s mismatch on the `deployed_version` cell sent to the DB\ngot:  %q\nwant: %q",
						prefix, gotCell, tc.wants.version,
					)
				}
			}

			// AND: it reaches the WebSocket clients, carrying the new version.
			if gotAnnounces := len(svcStatus.AnnounceChannel); gotAnnounces != tc.wants.announces {
				t.Errorf("%s AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotAnnounces, tc.wants.announces,
				)
			} else if gotAnnounces > 0 {
				got := string(<-svcStatus.AnnounceChannel)
				if !strings.Contains(got, tc.wants.version) {
					t.Errorf("%s announced payload missing version %q\ngot: %s",
						prefix, tc.wants.version, got,
					)
				}
			}

			// AND: the timestamp only moves when the version changed.
			gotTimestampChanged := svcStatus.DeployedVersionTimestamp() != hadTimestamp
			if gotTimestampChanged != tc.wants.timestampChanged {
				t.Errorf("%s mismatch on Status.DeployedVersionTimestamp() changing\ngot:  %t\nwant: %t",
					prefix, gotTimestampChanged, tc.wants.timestampChanged,
				)
			}

			// AND: no save is queued - a refresh consumes the Version on its own copy.
			if gotSaves := len(svcStatus.SaveChannel); gotSaves != wantSaves {
				t.Errorf("%s SaveChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotSaves, wantSaves,
				)
			}
		})
	}
}

func TestRefresh__previewKeepsLiveStatus(t *testing.T) {
	// GIVEN: a Lookup whose Status carries the live Service's channels.
	tests := map[string]struct {
		semanticVersioning *string
	}{
		"semantic_versioning toggled off": {
			semanticVersioning: new("false"),
		},
		"semantic_versioning reset to the default": {
			semanticVersioning: new("null"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, manual.Type, false, "")
			hadStatus := lookup.GetStatus()

			// WHEN: Refresh previews a `semantic_versioning` change with no overrides.
			// (`applyOverridesJSON` has nothing to apply, so it returns this same Lookup.)
			_, err := Refresh(
				lookup,
				manual.Type,
				nil,
				tc.semanticVersioning,
				nil,
			)

			prefix := fmt.Sprintf("%s\nRefresh()", packageName)

			// THEN: no error is given.
			if err != nil {
				t.Fatalf("%s unexpected error: %v", prefix, err)
			}

			// AND: the live Status is left in place - detaching it would sever the
			// Service from the database and the WebSocket clients.
			if gotStatus := lookup.GetStatus(); gotStatus != hadStatus {
				t.Errorf("%s replaced the live Status\ngot:  %p\nwant: %p",
					prefix, gotStatus, hadStatus,
				)
			}

			// AND: it keeps the channels that carry version changes outward.
			if lookup.GetStatus().AnnounceChannel == nil {
				t.Errorf("%s lost Status.AnnounceChannel", prefix)
			}
			if lookup.GetStatus().DatabaseChannel == nil {
				t.Errorf("%s lost Status.DatabaseChannel", prefix)
			}
		})
	}
}

func TestRefresh__overridesRemovingTheLookup(t *testing.T) {
	// GIVEN: a Lookup, and overrides that remove it.
	lookup := testLookup(t, manual.Type, false, "")
	overrides := []byte("null")

	// WHEN: Refresh is called with those overrides.
	got, err := Refresh(
		lookup,
		lookup.GetType(),
		overrides,
		nil,
		nil,
	)

	prefix := fmt.Sprintf("%s\nRefresh()", packageName)

	// THEN: an error is returned.
	e := errfmt.FormatError(err)
	if wantErrRegex := `removed by overrides`; !util.RegexCheck(wantErrRegex, e) {
		t.Fatalf("%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, wantErrRegex,
		)
	}

	// AND: no version is returned.
	if got != "" {
		t.Errorf("%s mismatch on version returned\ngot:  %q\nwant: %q",
			prefix, got, "",
		)
	}
}

func TestApplyOverridesJSON(t *testing.T) {
	type args struct {
		lookup             Lookup
		overrides          []byte
		semanticVerDiff    bool
		semanticVersioning *string
	}
	tests := []struct {
		name     string
		args     args
		errRegex string
	}{
		{
			name: "no overrides, no semantic versioning change",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		{
			name: "invalid semantic versioning",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: new("invalid"),
			},
			errRegex: test.TrimYAML(`
				^semantic_versioning:
					jsontext:
						invalid character .*$`,
			),
		},
		{
			name: "valid semantic versioning change",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: new("true"),
			},
			errRegex: `^$`,
		},
		{
			name: "overrides/valid",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          []byte(`{"url": "` + test.LookupJSON["url_valid"] + `"}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		{
			name: "overrides/invalid JSON",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          []byte(`{"url": "}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ could not find end character.*`,
			),
		},
		{
			name: "overrides/invalid var data type",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          []byte(`{"url": [""]}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning,
			)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\napplyOverridesJSON(%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.args.overrides,
					e, tc.errRegex,
				)
			}
		})
	}
}
