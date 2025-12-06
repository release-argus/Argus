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

package db

import (
	"os"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestAPI_UpdateRow(t *testing.T) {
	// GIVEN a DB with a few service status'.
	tests := map[string]struct {
		cells           []dbtype.Cell
		target          string
		exists          bool
		databaseDeleted bool
	}{
		"no cells": {
			target: "new",
		},
		"update single column of a row": {
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "latest_version",
					Value: "9.9.9"}},
			exists: true,
		},
		"trailing 0 is kept": {
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "latest_version",
					Value: "1.20"}},
			exists: true,
		},
		"update multiple columns of a row": {
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "deployed_version",
					Value: "8.8.8"},
				{Column: "deployed_version_timestamp",
					Value: time.Now().UTC().Format(time.RFC3339)}},
			exists: true,
		},
		"insert single column (new service)": {
			target: "new",
			cells: []dbtype.Cell{
				{Column: "latest_version",
					Value: "9.9.9"}},
		},
		"insert multiple columns (new service)": {
			target: "new",
			cells: []dbtype.Cell{
				{Column: "deployed_version",
					Value: "8.8.8"},
				{Column: "deployed_version_timestamp",
					Value: time.Now().UTC().Format(time.RFC3339)}},
		},
		"fail insert with invalid timestamp": {
			target: "new",
			cells: []dbtype.Cell{
				{Column: "deployed_version",
					Value: "8.8.8"},
				{Column: "deployed_version_timestamp",
					Value: "invalid"}},
		},
		"fail update with unknown column": {
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "foo",
					Value: "bar"}},
			exists: true,
		},
		"fail as db is deleted": {
			target: "existing",
			cells: []dbtype.Cell{
				{Column: "latest_version",
					Value: "9.9.9"}},
			exists:          true,
			databaseDeleted: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(name, "TestAPI_UpdateRow")
			t.Cleanup(func() { dbCleanup(tAPI) })
			tAPI.initialise()

			// Ensure the row exists when tc.exists.
			if tc.exists {
				tAPI.db.Exec("INSERT INTO status (id) VALUES (?)",
					tc.target)
			}
			// Delete the DB file.
			if tc.databaseDeleted {
				os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			}

			// WHEN updateRow is called targeting single/multiple cells.
			tAPI.updateRow(tc.target, tc.cells)
			time.Sleep(100 * time.Millisecond)

			// THEN those cells are changed in the DB.
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
				if got != cell.Value {
					if !tc.databaseDeleted {
						t.Errorf("%s\nwant: %s to have been updated to %q\ngot:  %q",
							packageName, cell.Column, cell.Value, got)
					}
				} else if tc.databaseDeleted {
					t.Errorf("%s\nwant: %s to not have been updated\ngot:  %q",
						packageName, cell.Column, got)
				}
			}
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestAPI_DeleteRow(t *testing.T) {
	// GIVEN a DB with a few service status'.
	tests := map[string]struct {
		serviceID       string
		exists          bool
		databaseDeleted bool
	}{
		"delete a row": {
			serviceID: "TestDeleteRow0",
			exists:    true},
		"delete a non-existing row": {
			serviceID: "TestDeleteRow1",
			exists:    false},
		"delete a row on a deleted DB": {
			serviceID:       "TestDeleteRow2",
			exists:          true,
			databaseDeleted: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			tAPI := testAPI(name, "TestAPI_DeleteRow")
			os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			t.Cleanup(func() { dbCleanup(tAPI) })
			tAPI.initialise()

			// Ensure the row exists if tc.exists.
			if tc.exists {
				tAPI.updateRow(
					tc.serviceID,
					[]dbtype.Cell{
						{Column: "latest_version", Value: "9.9.9"}, {Column: "deployed_version", Value: "8.8.8"}},
				)
				time.Sleep(100 * time.Millisecond)
			}
			// Check the row existence before the test.
			row := queryRow(t, tAPI.db, tc.serviceID)
			if tc.exists && (row.LatestVersion() == "" || row.DeployedVersion() == "") {
				t.Errorf("%s\nwant: row to exist\ngot:  %#v",
					packageName, row)
			}
			// Delete the DB file.
			if tc.databaseDeleted {
				os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			}

			// WHEN deleteRow is called targeting a row.
			tAPI.deleteRow(tc.serviceID)
			time.Sleep(100 * time.Millisecond)

			// THEN if we deleted the DB before the statement, we should have logged an error.
			stdout := releaseStdout()
			deleteFailRegex := `ERROR: [^)]+\), deleteRow`
			if tc.databaseDeleted != util.RegexCheck(deleteFailRegex, stdout) {
				t.Errorf("%s\nstdout mismatch:\nwant=%t (%q)\ngot: %q",
					packageName, tc.databaseDeleted, deleteFailRegex, stdout)
			}
			// AND the row is deleted from the DB (if it existed and the DB wasn't deleted).
			row = queryRow(t, tAPI.db, tc.serviceID)
			if row.LatestVersion() != "" || row.DeployedVersion() != "" {
				// no delete if we deleted the db.
				if !tc.databaseDeleted {
					t.Errorf("%s\nwant: row deleted\ngot:  %#v",
						packageName, row)
				}
			} else if tc.databaseDeleted {
				t.Errorf("%s\nwant: row exist as we deleted the db\ngot:  %#v",
					packageName, row)
			}
		})
	}
}

func TestAPI_Handler(t *testing.T) {
	// GIVEN a DB with a few service status'.
	tAPI := testAPI("TestAPI_Handler", "db")
	t.Cleanup(func() { dbCleanup(tAPI) })
	tAPI.initialise()
	go tAPI.handler()

	// WHEN a message is sent to the DatabaseChannel targeting latest_version.
	target := "keep0"
	cell1 := dbtype.Cell{
		Column: "latest_version", Value: "9.9.9"}
	cell2 := dbtype.Cell{
		Column: cell1.Column, Value: cell1.Value + "-dev"}
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
	*tAPI.config.DatabaseChannel <- msg1
	time.Sleep(250 * time.Millisecond)

	// THEN the cell was changed in the DB.
	got := queryRow(t, tAPI.db, target)
	if got.LatestVersion() != want.LatestVersion() {
		t.Errorf("%s\nExpected %q to be updated to %q\nwant: %#v\ngot:  %#v",
			packageName, cell1.Column, cell1.Value, want, got)
	}

	// ------------------------------

	// WHEN a message is sent to the DatabaseChannel deleting a row.
	*tAPI.config.DatabaseChannel <- dbtype.Message{
		ServiceID: target,
		Delete:    true,
	}
	time.Sleep(250 * time.Millisecond)

	// THEN the row is deleted from the DB.
	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != "" ||
		got.DeployedVersion() != "" {
		t.Errorf("%s\nExpected row to be deleted\nwant: %#v\ngot:  %#v",
			packageName, want, got)
	}

	// ------------------------------

	// WHEN multiple messages are targeting the same row in quick succession.
	*tAPI.config.DatabaseChannel <- msg1
	wantLatestVersion := msg2.Cells[0].Value
	*tAPI.config.DatabaseChannel <- msg2
	time.Sleep(250 * time.Millisecond)

	// THEN the last message is the one that is applied.
	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != wantLatestVersion {
		t.Errorf("%s\nExpected %q to be updated to %q\nwant: %#v\ngot:  %#v",
			packageName, cell2.Column, cell2.Value, want, got)
	}
}
