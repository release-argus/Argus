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

	_ "modernc.org/sqlite"
	db_types "github.com/release-argus/Argus/db/types"
)

func TestUpdateRowSingleValue(t *testing.T) {
	// GIVEN a DB with a few service status'
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestUpdateCellLatestVersion.db"
	api.initialise()
	api.convertServiceStatus()

	// WHEN updateRow is called targeting latest_version
	target := "keep0"
	cell := db_types.Cell{Column: "latest_version", Value: "9.9.9"}
	want := queryRow(api.db, target, t)
	want.LatestVersion = cell.Value
	api.updateRow(target, []db_types.Cell{cell})

	// THEN the cell was changed in the DB
	got := queryRow(api.db, target, t)
	// t.Fatal(want)
	if got != want {
		t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
			cell.Column, cell.Value, got, want)
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}

func TestUpdateRowMultiValue(t *testing.T) {
	// GIVEN a DB with a few service status'
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestUpdateCellLatestVersion.db"
	api.initialise()
	api.convertServiceStatus()

	// WHEN updateRow is called targeting deployed_version and deployed_version_timestamp
	target := "keep0"
	dvCell := db_types.Cell{Column: "deployed_version", Value: "9.9.9"}
	now := time.Now().UTC().Format(time.RFC3339)
	dvtCell := db_types.Cell{Column: "deployed_version_timestamp", Value: now}
	want := queryRow(api.db, target, t)
	want.DeployedVersion = dvCell.Value
	want.DeployedVersionTimestamp = dvtCell.Value
	api.updateRow(target, []db_types.Cell{dvCell, dvtCell})

	// THEN the cell was changed in the DB
	got := queryRow(api.db, target, t)
	// t.Fatal(want)
	if got != want {
		count := 0
		if got.DeployedVersion != dvCell.Value {
			count++
			t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
				dvCell.Column, dvCell.Value, got, want)
		}
		if got.DeployedVersionTimestamp != dvtCell.Value {
			count++
			t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
				dvtCell.Column, dvtCell.Value, got, want)
		}
		if count == 0 {
			t.Errorf("Didn't expect others to change\ngot  %v\nwant %v",
				got, want)
		}
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
	want := queryRow(api.db, target, t)
	want.LatestVersion = cell.Value
	*api.config.DatabaseChannel <- db_types.Message{
		ServiceID: target,
		Cells:     []db_types.Cell{cell},
	}
	time.Sleep(time.Second)

	// THEN the cell was changed in the DB
	got := queryRow(api.db, target, t)
	// t.Fatal(want)
	if got != want {
		t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
			cell.Column, cell.Value, got, want)
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}
