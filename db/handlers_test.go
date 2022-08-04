// Copyright [2022] [Argus]
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
	"os"
	"strings"
	"testing"
	"time"

	db_types "github.com/release-argus/Argus/db/types"
	_ "modernc.org/sqlite"
)

func TestUpdateRow(t *testing.T) {
	// GIVEN a DB with a few service status'
	initLogging()
	tests := map[string]struct {
		cells  []db_types.Cell
		target string
	}{
		"update single column of a row": {target: "keep0", cells: []db_types.Cell{{Column: "latest_version", Value: "9.9.9"}}},
		"update multiple columns of a row": {target: "keep0", cells: []db_types.Cell{{Column: "deployed_version", Value: "8.8.8"},
			{Column: "deployed_version_timestamp", Value: time.Now().UTC().Format(time.RFC3339)}}},
		"update single column of a non-existing row (new service)": {target: "new0", cells: []db_types.Cell{{Column: "latest_version", Value: "9.9.9"}}},
		"update multiple columns of a non-existing  row (new service)": {target: "new1", cells: []db_types.Cell{{Column: "deployed_version", Value: "8.8.8"},
			{Column: "deployed_version_timestamp", Value: time.Now().UTC().Format(time.RFC3339)}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := testConfig()
			api := api{config: &cfg}
			*api.config.Settings.Data.DatabaseFile = fmt.Sprintf("%s.db", strings.ReplaceAll(name, " ", "_"))
			api.initialise()
			api.convertServiceStatus()

			// WHEN updateRow is called targeting single/multiple cells
			api.updateRow(tc.target, tc.cells)
			time.Sleep(100 * time.Millisecond)

			// THEN those cell(s) are changed in the DB
			row := queryRow(t, api.db, tc.target)
			for _, cell := range tc.cells {
				var got *string
				switch cell.Column {
				case "latest_version":
					got = &row.LatestVersion
				case "latest_version_timestamp":
					got = &row.LatestVersionTimestamp
				case "deployed_version":
					got = &row.DeployedVersion
				case "deployed_version_timestamp":
					got = &row.DeployedVersionTimestamp
				case "approved_version":
					got = &row.ApprovedVersion
				}
				if *got != cell.Value {
					t.Errorf("expecting %s to have been updated to %q. got %q",
						cell.Column, cell.Value, *got)
				}
			}
			api.db.Close()
			os.Remove(*api.config.Settings.Data.DatabaseFile)
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestHandler(t *testing.T) {
	// GIVEN a DB with a few service status'
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestHandler.db"
	api.initialise()
	api.convertServiceStatus()
	go api.handler()

	// WHEN a message is send to the DatabaseChannel targeting latest_version
	target := "keep0"
	cell := db_types.Cell{Column: "latest_version", Value: "9.9.9"}
	want := queryRow(t, api.db, target)
	want.LatestVersion = cell.Value
	*api.config.DatabaseChannel <- db_types.Message{
		ServiceID: target,
		Cells:     []db_types.Cell{cell},
	}
	time.Sleep(time.Second)

	// THEN the cell was changed in the DB
	got := queryRow(t, api.db, target)
	if got.LatestVersion != want.LatestVersion {
		t.Errorf("Expected %q to be updated to %q\ngot  %#v\nwant %#v",
			cell.Column, cell.Value, got, want)
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}
