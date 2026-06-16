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
	"fmt"
	"log"
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
	// GIVEN: a string to write.
	tests := []struct {
		name string
		s    string
		want string
	}{
		{name: "simple string", s: "foo", want: "`foo`"},
		{name: "string with backticks", s: "foo`bar", want: "`foo`bar`"},
		{name: "empty string", s: "", want: "``"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a strings.Builder to write to.
			var b strings.Builder

			// WHEN: writeQuoted is called with the string and builder.
			writeQuoted(&b, tc.s)

			prefix := fmt.Sprintf(
				"%s\nwriteQuoted(%q)",
				packageName, tc.s,
			)

			// THEN: the builder contains the string wrapped in backticks.
			if got := b.String(); got != tc.want {
				t.Errorf(
					"%s quoted string mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestAPI_Handler(t *testing.T) {
	// GIVEN: a DB with a few service status'.
	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())

	// WHEN: a message is sent to the DatabaseChannel targeting latest_version.
	target := "keep0"
	cell1 := dbtype.Cell{
		Column: "latest_version", Value: "9.9.9",
	}
	cell2 := dbtype.Cell{
		Column: cell1.Column, Value: cell1.Value + "-dev",
	}
	want := queryRow(t, tAPI.db, target)
	want.SetLatestVersion(cell1.Value, "", false)
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

	prefix := fmt.Sprintf("%s\napi.handler()", packageName)

	// THEN: the cell was changed in the DB.
	got := queryRow(t, tAPI.db, target)
	if got.LatestVersion() != want.LatestVersion() {
		t.Errorf(
			"%s LatestVersion mismatch (sent %+v)\ngot:  %#v\nwant: %#v",
			prefix, msg1,
			got, want,
		)
	}

	// ------------------------------

	// WHEN: a message is sent to the DatabaseChannel deleting a row.
	msg3 := dbtype.Message{
		ServiceID: target,
		Delete:    true,
	}
	tAPI.config.DatabaseChannel <- msg3
	time.Sleep(250 * time.Millisecond)

	// THEN: the row is deleted from the DB.
	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != "" ||
		got.DeployedVersion() != "" {
		t.Errorf(
			"%s row not deleted (sent %+v)\ngot:  %#v\nwant: deleted (empty status)",
			prefix, msg3,
			got,
		)
	}

	// ------------------------------

	// WHEN: multiple messages are targeting the same row in quick succession.
	tAPI.config.DatabaseChannel <- msg1
	wantLatestVersion := msg2.Cells[0].Value
	tAPI.config.DatabaseChannel <- msg2
	time.Sleep(250 * time.Millisecond)

	// THEN: the last message is the one applied.
	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != wantLatestVersion {
		t.Errorf(
			"%s LatestVersion mismatch (sent %+v, then %+v)\ngot:  %#v\nwant: %#v",
			prefix, msg1, msg2,
			got, want,
		)
	}
}

func TestAPI_Handler__fail(t *testing.T) {
	// GIVEN: a DB with a few service status'.
	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())
	releaseStdout := test.CaptureLog(t, logx.Default())

	// WHEN: the DatabaseChannel is closed.
	close(tAPI.config.DatabaseChannel)
	time.Sleep(10 * time.Millisecond)

	// THEN: the handler exits cleanly.
	stdout := releaseStdout()
	re := `FATAL: .*Database closed`
	if !util.RegexCheck(re, stdout) {
		t.Errorf(
			"%s stdout mismatch after closing DatabaseChannel:\ngot:  %q\nwant: %q",
			log.Prefix(), stdout, re,
		)
	}
	<-logx.ExitCodeChannel()
}

func TestBuildUpdateRowStatement(t *testing.T) {
	// GIVEN: a serviceID and a set of cells to upsert.
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
			name:      "empty cells",
			serviceID: "service4",
			cells:     []dbtype.Cell{},
			wantSQL: `` +
				"INSERT INTO status (`id`) " +
				`VALUES (?) ` +
				`ON CONFLICT("id") DO UPDATE SET `,
			wantParams: []any{"service4"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: buildUpdateRowStatement is called with the serviceID and cells.
			gotSQL, gotParams := buildUpdateRowStatement(tc.serviceID, tc.cells)

			prefix := fmt.Sprintf(
				"%s\nbuildUpdateRowStatement(%q, %+v)",
				packageName, tc.serviceID, tc.cells,
			)

			// THEN: the SQL statement matches the expected upsert statement.
			if gotSQL != tc.wantSQL {
				t.Errorf(
					"%s SQL mismatch\ngot:  %q\nwant: %q",
					prefix, gotSQL, tc.wantSQL,
				)
			}

			// AND: params has serviceID first, then each cell's Value in order.
			if !util.AreSlicesEqual(gotParams, tc.wantParams) {
				t.Fatalf(
					"%s params mismatch\ngot:  %v\nwant: %v",
					prefix, gotParams, tc.wantParams,
				)
			}
		})
	}
}

func TestAPI_UpdateRow(t *testing.T) {
	// GIVEN: a DB with a few service status'.
	tests := []struct {
		name            string
		cells           []dbtype.Cell
		target          string
		exists          bool
		databaseDeleted bool
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
				{Column: "latest_version", Value: "1.20"},
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
			name:   "fail insert with invalid timestamp",
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
			exists: true,
		},
		{
			name:   "fail as db is deleted",
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "latest_version", Value: "9.9.9"},
			},
			exists:          true,
			databaseDeleted: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(t)
			tAPI.initialise()

			// Ensure the row exists when tc.exists.
			if tc.exists {
				_, _ = tAPI.db.Exec("INSERT INTO status (id) VALUES (?)",
					tc.target,
				)
			}
			// Delete the DB file.
			if tc.databaseDeleted {
				_ = os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			}

			// WHEN: updateRow is called targeting single/multiple cells.
			tAPI.updateRow(tc.target, tc.cells)
			time.Sleep(100 * time.Millisecond)

			// THEN: those cells are changed in the DB.
			row := queryRow(t, tAPI.db, tc.target)
			for _, cell := range tc.cells {
				var got string
				switch cell.Column {
				case "latest_version":
					got = row.LatestVersion()
				case "latest_version_timestamp":
					got = row.LatestVersionTimestamp()
				case "deployed_version":
					got = row.DeployedVersion()
				case "deployed_version_timestamp":
					got = row.DeployedVersionTimestamp()
				case "approved_version":
					got = row.ApprovedVersion()
				default:
					continue // Skip unknown columns.
				}
				prefix := fmt.Sprintf(
					"%s\nAPI UpdateRow(target: %q, cell: %q)",
					packageName, tc.target, cell.Column,
				)

				if got != cell.Value {
					if !tc.databaseDeleted {
						t.Errorf(
							"%s mismatch\ngot:  %q\nwant: %q",
							prefix, got, cell.Value,
						)
					}
				} else if tc.databaseDeleted {
					t.Errorf(
						"%s shouldn't have done anything as the DB was deleted\ngot:  %q\nwant: %q",
						prefix, got, cell.Value,
					)
				}
			}
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestAPI_DeleteRow(t *testing.T) {
	// GIVEN: a DB with multiple service status'.
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
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			tAPI := testAPI(t)
			tAPI.initialise()

			// Ensure the row exists if tc.exists.
			if tc.exists {
				tAPI.updateRow(
					tc.serviceID,
					[]dbtype.Cell{
						{Column: "latest_version", Value: "9.9.9"},
						{Column: "deployed_version", Value: "8.8.8"},
					},
				)
				time.Sleep(100 * time.Millisecond)
			}
			// Check the row existence before the test.
			row := queryRow(t, tAPI.db, tc.serviceID)
			if tc.exists && (row.LatestVersion() == "" || row.DeployedVersion() == "") {
				t.Errorf(
					"%s\napi.queryRow(%q) mismatch\ngot:  %#v\nwant: row to exist",
					packageName, tc.serviceID,
					row,
				)
			}
			// Delete the DB file.
			if tc.databaseDeleted {
				_ = os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			}

			// WHEN: deleteRow is called targeting a row.
			tAPI.deleteRow(tc.serviceID)
			time.Sleep(100 * time.Millisecond)

			prefix := fmt.Sprintf(
				"%s\napi.DeleteRow(%q)",
				packageName, tc.serviceID,
			)

			// THEN: if we deleted the DB before the statement, we should have logged an error.
			stdout := releaseStdout()
			deleteFailRegex := `ERROR: [^)]+\), deleteRow`
			if tc.databaseDeleted != util.RegexCheck(deleteFailRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch (DB deleted=%t):\ngot:  %q\nwant: %q (want match: %t)",
					prefix, tc.databaseDeleted,
					stdout, deleteFailRegex, tc.databaseDeleted,
				)
			}

			// AND: the row is deleted from the DB (if it existed and the DB was not deleted).
			row = queryRow(t, tAPI.db, tc.serviceID)
			if row.LatestVersion() != "" || row.DeployedVersion() != "" {
				// Get values as we didn't delete the db.
				if !tc.databaseDeleted {
					t.Errorf(
						"%s row mismatch when DB not deleted:\ngot:  %#v\nwant: row deleted",
						prefix, row,
					)
				}
			} else if tc.databaseDeleted {
				t.Errorf(
					"%s row mismatch when DB deleted:\ngot:  %#v\nwant: row deleted",
					packageName, row,
				)
			}
		})
	}
}
