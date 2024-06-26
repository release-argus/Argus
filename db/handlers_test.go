// Copyright [2023] [Argus]
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
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
)

func TestAPI_UpdateRow(t *testing.T) {
	// GIVEN a DB with a few service status'
	tests := map[string]struct {
		cells  []dbtype.Cell
		target string
	}{
		"update single column of a row": {
			target: "keep0",
			cells: []dbtype.Cell{
				{Column: "latest_version",
					Value: "9.9.9"}},
		},
		"trailing 0 is kept": {
			target: "keep0",
			cells: []dbtype.Cell{
				{Column: "latest_version",
					Value: "1.20"}}},
		"update multiple columns of a row": {
			target: "keep0",
			cells: []dbtype.Cell{
				{Column: "deployed_version",
					Value: "8.8.8"},
				{Column: "deployed_version_timestamp",
					Value: time.Now().UTC().Format(time.RFC3339)}},
		},
		"update single column of a non-existing row (new service)": {
			target: "new0",
			cells: []dbtype.Cell{
				{Column: "latest_version",
					Value: "9.9.9"}},
		},
		"update multiple columns of a non-existing row (new service)": {
			target: "new1",
			cells: []dbtype.Cell{
				{Column: "deployed_version",
					Value: "8.8.8"},
				{Column: "deployed_version_timestamp",
					Value: time.Now().UTC().Format(time.RFC3339)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(name, "TestAPI_UpdateRow")
			defer dbCleanup(tAPI)
			tAPI.initialise()

			// WHEN updateRow is called targeting single/multiple cells
			tAPI.updateRow(tc.target, tc.cells)
			time.Sleep(100 * time.Millisecond)

			// THEN those cell(s) are changed in the DB
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
				}
				if got != cell.Value {
					t.Errorf("expecting %s to have been updated to %q. got %q",
						cell.Column, cell.Value, got)
				}
			}
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestAPI_DeleteRow(t *testing.T) {
	// GIVEN a DB with a few service status'
	tests := map[string]struct {
		serviceID string
		exists    bool
	}{
		"delete a row": {
			serviceID: "TestDeleteRow0",
			exists:    true},
		"delete a non-existing row": {
			serviceID: "TestDeleteRow1",
			exists:    false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(name, "TestAPI_DeleteRow")
			defer dbCleanup(tAPI)
			tAPI.initialise()

			// Ensure the row exists if tc.exists
			if tc.exists {
				tAPI.updateRow(
					tc.serviceID,
					[]dbtype.Cell{
						{Column: "latest_version", Value: "9.9.9"}, {Column: "deployed_version", Value: "8.8.8"}},
				)
				time.Sleep(100 * time.Millisecond)
			}
			// Check the row existance before the test
			row := queryRow(t, tAPI.db, tc.serviceID)
			if tc.exists && (row.LatestVersion() == "" || row.DeployedVersion() == "") {
				t.Errorf("expecting row to exist. got %#v", row)
			}

			// WHEN deleteRow is called targeting a row
			tAPI.deleteRow(tc.serviceID)
			time.Sleep(100 * time.Millisecond)

			// THEN the row is deleted from the DB
			row = queryRow(t, tAPI.db, tc.serviceID)
			if row.LatestVersion() != "" || row.DeployedVersion() != "" {
				t.Errorf("expecting row to be deleted. got %#v", row)
			}
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestAPI_Handler(t *testing.T) {
	// GIVEN a DB with a few service status'
	tAPI := testAPI("TestAPI_Handler", "db")
	defer dbCleanup(tAPI)
	tAPI.initialise()
	go tAPI.handler()

	// WHEN a message is sent to the DatabaseChannel targeting latest_version
	target := "keep0"
	cell1 := dbtype.Cell{
		Column: "latest_version", Value: "9.9.9"}
	cell2 := dbtype.Cell{
		Column: cell1.Column, Value: cell1.Value + "-dev"}
	want := queryRow(t, tAPI.db, target)
	want.SetLatestVersion(cell1.Value, false)
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

	// THEN the cell was changed in the DB
	got := queryRow(t, tAPI.db, target)
	if got.LatestVersion() != want.LatestVersion() {
		t.Errorf("Expected %q to be updated to %q\ngot  %#v\nwant %#v",
			cell1.Column, cell1.Value, got, want)
	}

	// ------------------------------

	// WHEN a message is sent to the DatabaseChannel deleting a row
	*tAPI.config.DatabaseChannel <- dbtype.Message{
		ServiceID: target,
		Delete:    true,
	}
	time.Sleep(250 * time.Millisecond)

	// THEN the row is deleted from the DB
	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != "" || got.DeployedVersion() != "" {
		t.Errorf("Expected row to be deleted\ngot  %#v\nwant %#v", got, want)
	}

	// ------------------------------

	// WHEN multiple messages are targeting the same row in quick succession
	*tAPI.config.DatabaseChannel <- msg1
	wantLatestVersion := msg2.Cells[0].Value
	*tAPI.config.DatabaseChannel <- msg2
	time.Sleep(250 * time.Millisecond)

	// THEN the last message is the one that is applied
	got = queryRow(t, tAPI.db, target)
	if got.LatestVersion() != wantLatestVersion {
		t.Errorf("Expected %q to be updated to %q\ngot  %#v\nwant %#v",
			cell2.Column, cell2.Value, got, want)
	}
}
