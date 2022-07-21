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
	"os"
	"testing"
	"time"

	db_types "github.com/release-argus/Argus/db/types"
	service_status "github.com/release-argus/Argus/service/status"
	_ "modernc.org/sqlite"
)

func TestUpdateRow(t *testing.T) {
	// GIVEN a DB with a few service status'
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestUpdateRow.db"
	api.initialise()
	api.convertServiceStatus()
	tests := map[string]struct {
		row   string
		cells []db_types.Cell
		want  service_status.Status
	}{
		"single_column": {row: "keep0", cells: []db_types.Cell{{Column: "latest_version", Value: "9.9.9"}},
			want: service_status.Status{ApprovedVersion: "0.0.1",
				DeployedVersion: "0.0.0", DeployedVersionTimestamp: "2020-01-01T01:01:01Z",
				LatestVersion: "9.9.9", LatestVersionTimestamp: "2022-01-01T01:01:01Z"}},
		"multiple_columns": {row: "keep1", cells: []db_types.Cell{{Column: "deployed_version", Value: "9.9.9"}, {Column: "deployed_version_timestamp", Value: "2022-02-02T02:02:02Z"}},
			want: service_status.Status{ApprovedVersion: "0.0.1",
				DeployedVersion: "9.9.9", DeployedVersionTimestamp: "2022-02-02T02:02:02Z",
				LatestVersion: "0.0.2", LatestVersionTimestamp: "2022-01-01T01:01:01Z"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN updateRow is called targeting latest_version
			api.updateRow(tc.row, tc.cells)

			// THEN the row was changed correctly
			got := queryRow(t, api.db, tc.row)
			if got != tc.want {
				t.Errorf("Expected row to be updated with %v\ngot  %v\nwant %v",
					tc.cells, got, tc.want)
			}
		})
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
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
	// t.Fatal(want)
	if got != want {
		t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
			cell.Column, cell.Value, got, want)
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}
