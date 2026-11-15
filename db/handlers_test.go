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

package db

import (
	"os"
	"strings"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestWriteQuoted(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{name: "simple string", s: "foo", want: "`foo`"},
		{name: "string with backticks (not escaped)", s: "foo`bar", want: "`foo`bar`"},
		{name: "empty string", s: "", want: "``"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var b strings.Builder

			writeQuoted(&b, tc.s)

			if got := b.String(); got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestAPI_Handler(t *testing.T) {
	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())

	target := "keep0"
	cell1 := dbtype.Cell{
		Column: "latest_version", Value: "9.9.9",
	}
	cell2 := dbtype.Cell{
		Column: cell1.Column, Value: cell1.Value + "-dev",
	}
	msg1 := dbtype.Message{
		ServiceID: target,
		Cells:     []dbtype.Cell{cell1},
	}
	msg2 := dbtype.Message{
		ServiceID: target,
		Cells:     []dbtype.Cell{cell2},
	}
	tAPI.config.DatabaseChannel <- msg1
	time.Sleep(250 * time.Millisecond)

	got := queryRow(t, tAPI.db, target)
	if got.LatestVersion() != cell1.Value {
		t.Errorf(
			"update: Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
			got.LatestVersion(), cell1.Value,
		)
	}

	// ------------------------------

	msg3 := dbtype.Message{
		ServiceID: target,
		Delete:    true,
	}
	tAPI.config.DatabaseChannel <- msg3
	time.Sleep(250 * time.Millisecond)

	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != "" ||
		got.DeployedVersion() != "" {
		t.Errorf(
			"delete: Status mismatch\n"+
				"got:  LatestVersion()=%q, DeployedVersion()=%q\n"+
				"want: LatestVersion()=\"\", DeployedVersion()=\"\"",
			got.LatestVersion(), got.DeployedVersion(),
		)
	}

	// ------------------------------

	tAPI.config.DatabaseChannel <- msg1
	wantLatestVersion := msg2.Cells[0].Value
	tAPI.config.DatabaseChannel <- msg2
	time.Sleep(250 * time.Millisecond)

	// The handler applies messages in channel order.
	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != wantLatestVersion {
		t.Errorf(
			"successive updates: Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
			got.LatestVersion(), wantLatestVersion,
		)
	}
}

func TestAPI_Handler__fail(t *testing.T) {
	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())
	releaseStdout := test.CaptureLog(t, logx.Default())

	close(tAPI.config.DatabaseChannel)

	// Fatal logs before it signals, so the line is captured once the signal arrives.
	select {
	case <-logx.ExitCodeChannel():
	case <-time.After(time.Second):
		t.Fatal("logx.ExitCodeChannel() mismatch\ngot:  absent\nwant: present")
	}

	stdout := releaseStdout()
	re := `FATAL: .*Database closed`
	if !util.RegexCheck(re, stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, re)
	}
}

func TestBuildUpdateRowStatement(t *testing.T) {
	tests := []struct {
		name       string
		serviceID  string
		cells      []dbtype.Cell
		wantSQL    string
		wantParams []any
	}{
		{
			name:      "single cell",
			serviceID: "service1",
			cells: []dbtype.Cell{
				{Column: "latest_version", Value: "1.2.3"},
			},
			wantSQL: `` +
				"INSERT INTO status (`id`,`latest_version`) " +
				`VALUES (?,?) ` +
				`ON CONFLICT("id") DO UPDATE SET ` +
				"`latest_version` = excluded.`latest_version`",
			wantParams: []any{"service1", "1.2.3"},
		},
		{
			name:      "multiple cells",
			serviceID: "service2",
			cells: []dbtype.Cell{
				{Column: "latest_version", Value: "1.2.3"},
				{Column: "deployed_version", Value: "1.2.2"},
				{Column: "deployed_version_timestamp", Value: "2026-06-15T00:00:00Z"},
			},
			wantSQL: `` +
				"INSERT INTO status (`id`,`latest_version`,`deployed_version`,`deployed_version_timestamp`) " +
				`VALUES (?,?,?,?) ` +
				`ON CONFLICT("id") DO UPDATE SET ` +
				"`latest_version` = excluded.`latest_version`," +
				"`deployed_version` = excluded.`deployed_version`," +
				"`deployed_version_timestamp` = excluded.`deployed_version_timestamp`",
			wantParams: []any{"service2", "1.2.3", "1.2.2", "2026-06-15T00:00:00Z"},
		},
		{
			name:       "empty cells",
			serviceID:  "service4",
			cells:      []dbtype.Cell{},
			wantSQL:    "",
			wantParams: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotSQL, gotParams := buildUpdateRowStatement(tc.serviceID, tc.cells)

			if gotSQL != tc.wantSQL {
				t.Errorf("SQL mismatch\ngot:  %q\nwant: %q", gotSQL, tc.wantSQL)
			}

			if !util.AreSlicesEqual(gotParams, tc.wantParams) {
				t.Fatalf("params mismatch\ngot:  %v\nwant: %v", gotParams, tc.wantParams)
			}
		})
	}
}

func TestAPI_UpdateRow(t *testing.T) {
	tests := []struct {
		name            string
		cells           []dbtype.Cell
		target          string
		exists          bool
		databaseDeleted bool
		errorLogRegex   string
	}{
		{
			name:   "no cells",
			target: "new",
		},
		{
			name:   "update single column of a row",
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "latest_version", Value: "9.9.9"},
			},
			exists: true,
		},
		{
			name:   "trailing 0 is kept",
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "latest_version", Value: "1.20"}, // TEXT column, so not read as a number.
			},
			exists: true,
		},
		{
			name:   "update multiple columns of a row",
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "deployed_version", Value: "8.8.8"},
				{Column: "deployed_version_timestamp", Value: time.Now().UTC().Format(time.RFC3339)},
			},
			exists: true,
		},
		{
			name:   "insert single column (new service)",
			target: "new",
			cells: []dbtype.Cell{
				{Column: "latest_version", Value: "9.9.9"},
			},
		},
		{
			name:   "insert multiple columns (new service)",
			target: "new",
			cells: []dbtype.Cell{
				{Column: "deployed_version", Value: "8.8.8"},
				{Column: "deployed_version_timestamp", Value: time.Now().UTC().Format(time.RFC3339)},
			},
		},
		{
			name:   "insert with non-RFC3339 timestamp is stored as-is",
			target: "new",
			cells: []dbtype.Cell{
				{Column: "deployed_version", Value: "8.8.8"},
				{Column: "deployed_version_timestamp", Value: "invalid"},
			},
		},
		{
			name:   "fail update with unknown column",
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "foo", Value: "bar"},
			},
			exists:        true,
			errorLogRegex: `ERROR: .*failed to upsert status row: .*no column named foo`,
		},
		{
			name:   "fail as db is deleted",
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "latest_version", Value: "9.9.9"},
			},
			exists:          true,
			databaseDeleted: true,
			errorLogRegex:   `ERROR: .*failed to upsert status row: `,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			tAPI := testAPI(t)
			tAPI.initialise()

			if tc.exists {
				if _, err := tAPI.db.Exec("INSERT INTO status (id) VALUES (?)", tc.target); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			}
			if tc.databaseDeleted {
				if err := os.Remove(tAPI.config.Settings.Data.DatabaseFile); err != nil {
					t.Fatalf("os.Remove() failed: %v", err)
				}
			}

			tAPI.updateRow(tc.target, tc.cells)
			time.Sleep(100 * time.Millisecond)

			stdout := releaseStdout()
			if tc.errorLogRegex == "" {
				if errPrefix := "ERROR:"; strings.Contains(stdout, errPrefix) {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: NOT %q", stdout, errPrefix)
				}
			} else if !util.RegexCheck(tc.errorLogRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.errorLogRegex)
			}

			row := queryRow(t, tAPI.db, tc.target)
			if len(tc.cells) == 0 {
				if id := row.GetServiceInfo().ID; id != "" {
					t.Errorf("Status.GetServiceInfo().ID mismatch\ngot:  %q\nwant: \"\"", id)
				}
			}

			for _, cell := range tc.cells {
				var got, getter string
				switch cell.Column {
				case "latest_version":
					got, getter = row.LatestVersion(), "LatestVersion()"
				case "latest_version_timestamp":
					got, getter = row.LatestVersionTimestamp(), "LatestVersionTimestamp()"
				case "deployed_version":
					got, getter = row.DeployedVersion(), "DeployedVersion()"
				case "deployed_version_timestamp":
					got, getter = row.DeployedVersionTimestamp(), "DeployedVersionTimestamp()"
				case "approved_version":
					got, getter = row.ApprovedVersion(), "ApprovedVersion()"
				default:
					continue // Not readable; the error log covers it.
				}
				t.Run(cell.Column, func(t *testing.T) {
					if got != cell.Value {
						if !tc.databaseDeleted {
							t.Errorf(
								"Status.%s mismatch\ngot:  %q\nwant: %q",
								getter, got, cell.Value,
							)
						}
					} else if tc.databaseDeleted {
						t.Errorf(
							"Status.%s mismatch\ngot:  %q\nwant: NOT %q",
							getter, got, cell.Value,
						)
					}
				})
			}
		})
	}
}

func TestAPI_DeleteRow(t *testing.T) {
	tests := []struct {
		name            string
		serviceID       string
		exists          bool
		databaseDeleted bool
	}{
		{
			name:      "delete a row",
			serviceID: "TestDeleteRow0",
			exists:    true,
		},
		{
			name:      "delete a non-existing row",
			serviceID: "TestDeleteRow1",
			exists:    false,
		},
		{
			name:            "delete a row on a deleted DB",
			serviceID:       "TestDeleteRow2",
			exists:          true,
			databaseDeleted: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			tAPI := testAPI(t)
			tAPI.initialise()

			latestVersion, deployedVersion := "9.9.9", "8.8.8"
			if tc.exists {
				tAPI.updateRow(
					tc.serviceID,
					[]dbtype.Cell{
						{Column: "latest_version", Value: latestVersion},
						{Column: "deployed_version", Value: deployedVersion},
					},
				)
				time.Sleep(100 * time.Millisecond)
				row := queryRow(t, tAPI.db, tc.serviceID)
				if row.LatestVersion() != latestVersion || row.DeployedVersion() != deployedVersion {
					t.Fatalf(
						"before: Status mismatch\n"+
							"got:  LatestVersion()=%q, DeployedVersion()=%q\n"+
							"want: LatestVersion()=%q, DeployedVersion()=%q",
						row.LatestVersion(), row.DeployedVersion(),
						latestVersion, deployedVersion,
					)
				}
			}
			if tc.databaseDeleted {
				if err := os.Remove(tAPI.config.Settings.Data.DatabaseFile); err != nil {
					t.Fatalf("os.Remove() failed: %v", err)
				}
			}

			tAPI.deleteRow(tc.serviceID)
			time.Sleep(100 * time.Millisecond)

			stdout := releaseStdout()
			deleteFailRegex := `ERROR: [^)]+\), failed to delete status row`
			matched := util.RegexCheck(deleteFailRegex, stdout)
			if tc.databaseDeleted && !matched {
				t.Errorf(
					"after deleteRow: stdout mismatch\ngot:  %q\nwant: %q",
					stdout, deleteFailRegex,
				)
			} else if !tc.databaseDeleted && matched {
				t.Errorf(
					"after deleteRow: stdout mismatch\ngot:  %q\nwant: NOT %q",
					stdout, deleteFailRegex,
				)
			}

			row := queryRow(t, tAPI.db, tc.serviceID)
			wantLatestVersion, wantDeployedVersion := "", ""
			if tc.databaseDeleted {
				wantLatestVersion, wantDeployedVersion = latestVersion, deployedVersion
			}
			if row.LatestVersion() != wantLatestVersion || row.DeployedVersion() != wantDeployedVersion {
				t.Errorf(
					"after deleteRow: Status mismatch\n"+
						"got:  LatestVersion()=%q, DeployedVersion()=%q\n"+
						"want: LatestVersion()=%q, DeployedVersion()=%q",
					row.LatestVersion(), row.DeployedVersion(),
					wantLatestVersion, wantDeployedVersion,
				)
			}
		})
	}
}
