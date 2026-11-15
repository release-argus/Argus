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

			got, err := Refresh(
				lookup,
				previousType,
				tc.args.overrides,
				nil,
				nil,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if err == nil && got != tc.wants.version {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", got, tc.wants.version)
			}

			if gotVersion := svcStatus.DeployedVersion(); gotVersion != tc.wants.statusVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					gotVersion, tc.wants.statusVersion,
				)
			}

			if gotDBMessages := len(svcStatus.DatabaseChannel); gotDBMessages != tc.wants.dbMessages {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					gotDBMessages, tc.wants.dbMessages,
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
					t.Errorf(
						"Status.DatabaseChannel deployed_version cell mismatch\ngot:  %q\nwant: %q",
						gotCell, tc.wants.version,
					)
				}
			}

			if gotAnnounces := len(svcStatus.AnnounceChannel); gotAnnounces != tc.wants.announces {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotAnnounces, tc.wants.announces,
				)
			} else if gotAnnounces > 0 {
				payload := string(<-svcStatus.AnnounceChannel)
				if !strings.Contains(payload, tc.wants.version) {
					t.Errorf(
						"Status.AnnounceChannel payload mismatch\ngot:  %s\nwant: %q",
						payload, tc.wants.version,
					)
				}
			}

			gotTimestamp := svcStatus.DeployedVersionTimestamp()
			if tc.wants.timestampChanged {
				if gotTimestamp == hadTimestamp {
					t.Errorf(
						"Status.DeployedVersionTimestamp() mismatch\ngot:  %q\nwant: NOT %q",
						gotTimestamp, hadTimestamp,
					)
				}
			} else if gotTimestamp != hadTimestamp {
				t.Errorf(
					"Status.DeployedVersionTimestamp() mismatch\ngot:  %q\nwant: %q",
					gotTimestamp, hadTimestamp,
				)
			}

			if gotSaves := len(svcStatus.SaveChannel); gotSaves != wantSaves {
				t.Errorf(
					"Status.SaveChannel message count mismatch\ngot:  %d\nwant: %d",
					gotSaves, wantSaves,
				)
			}
		})
	}
}

func TestRefresh__previewKeepsLiveStatus(t *testing.T) {
	tests := []struct {
		name               string
		semanticVersioning *string
	}{
		{
			name:               "semantic_versioning toggled off",
			semanticVersioning: new("false"),
		},
		{
			name:               "semantic_versioning reset to the default",
			semanticVersioning: new("null"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, manual.Type, false, "")
			hadStatus := lookup.GetStatus()

			_, err := Refresh(
				lookup,
				manual.Type,
				nil,
				tc.semanticVersioning,
				nil,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// applyOverridesJSON returns this same Lookup, so a detached Lookup.GetStatus would cut the
			// Service off from the database and the WebSocket clients.
			if gotStatus := lookup.GetStatus(); gotStatus != hadStatus {
				t.Errorf("Lookup.GetStatus() mismatch\ngot:  %p\nwant: %p", gotStatus, hadStatus)
			}

			if lookup.GetStatus().AnnounceChannel == nil {
				t.Error("Status.AnnounceChannel mismatch\ngot:  nil\nwant: non-nil")
			}
			if lookup.GetStatus().DatabaseChannel == nil {
				t.Error("Status.DatabaseChannel mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestRefresh__overridesRemovingTheLookup(t *testing.T) {
	lookup := testLookup(t, manual.Type, false, "")

	got, err := Refresh(
		lookup,
		lookup.GetType(),
		[]byte("null"),
		nil,
		nil,
	)
	e := errfmt.FormatError(err)
	if wantErrRegex := `^deployed_version: removed by overrides$`; !util.RegexCheck(wantErrRegex, e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, wantErrRegex)
	}

	if got != "" {
		t.Errorf("version mismatch\ngot:  %q\nwant: \"\"", got)
	}
}

func TestApplyOverridesJSON(t *testing.T) {
	type args struct {
		lookup             Lookup
		overrides          []byte
		semanticVerDiff    bool
		semanticVersioning *string
	}
	defaultURL := test.LookupBare.URLInvalid + "/"
	tests := []struct {
		name                   string
		args                   args
		wantURL                string
		wantSemanticVersioning string
		errRegex               string
	}{
		{
			name: "no overrides, no semantic versioning change",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantURL:                defaultURL,
			wantSemanticVersioning: "nil",
			errRegex:               `^$`,
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
			wantURL:                defaultURL,
			wantSemanticVersioning: "true",
			errRegex:               `^$`,
		},
		{
			name: "overrides/valid",
			args: args{
				lookup:             testLookup(t, web.Type, false, ""),
				overrides:          []byte(`{"url": "` + test.LookupJSON.URLValid + `"}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantURL:                test.LookupJSON.URLValid,
			wantSemanticVersioning: "nil",
			errRegex:               `^$`,
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

			got, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if gotURL := got.(*web.Lookup).URL; gotURL != tc.wantURL {
				t.Errorf("Lookup.URL mismatch\ngot:  %q\nwant: %q", gotURL, tc.wantURL)
			}
			gotSemanticVersioning := test.StringifyPtr(got.GetOptions().SemanticVersioning)
			if gotSemanticVersioning != tc.wantSemanticVersioning {
				t.Errorf(
					"Lookup.GetOptions().SemanticVersioning mismatch\ngot:  %s\nwant: %s",
					gotSemanticVersioning, tc.wantSemanticVersioning,
				)
			}
		})
	}
}
