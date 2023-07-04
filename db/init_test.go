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
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	_ "modernc.org/sqlite"
)

func TestCheckFile(t *testing.T) {
	// GIVEN various paths
	tests := map[string]struct {
		removeBefore     string
		createDirBefore  string
		createFileBefore string
		path             string
		panicRegex       string
	}{
		"file doesn't exist": {
			path:         "something_doesnt_exist.db",
			removeBefore: "something_doesnt_exist.db"},
		"dir doesn't exist, so is created": {
			path:         "dir_doesnt_exist/argus.db",
			removeBefore: "dir_doesnt_exist"},
		"dir exists but not file": {
			path:            "dir_does_exist/argus.db",
			createDirBefore: "dir_does_exist"},
		"file is dir": {
			path:            "folder.db",
			createDirBefore: "folder.db",
			panicRegex:      "path .* is a directory, not a file"},
		"dir is file": {
			path:             "item_not_a_dir/argus.db",
			createFileBefore: "item_not_a_dir",
			panicRegex:       "path .* is not a directory"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			os.RemoveAll(tc.removeBefore)
			os.RemoveAll(tc.createDirBefore)
			if tc.createDirBefore != "" {
				err := os.Mkdir(tc.createDirBefore, os.ModeDir|0755)
				if err != nil {
					t.Fatalf("%s",
						err)
				}
				defer os.RemoveAll(tc.createDirBefore)
			}
			if tc.createFileBefore != "" {
				file, err := os.Create(tc.createFileBefore)
				if err != nil {
					t.Fatalf("%s",
						err)
				}
				file.Close()
				defer os.Remove(tc.createFileBefore)
			}
			if tc.panicRegex != "" {
				defer func() {
					r := recover()
					rStr := fmt.Sprint(r)
					re := regexp.MustCompile(tc.panicRegex)
					match := re.MatchString(rStr)
					if !match {
						t.Errorf("want match for %q\nnot: %q",
							tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN checkFile is called on that same dir
			checkFile(tc.path)

			// THEN we get here only when we should
			if tc.panicRegex != "" {
				t.Fatalf("Expected panic with %q",
					tc.panicRegex)
			}
		})
	}
}

func TestAPI_Initialise(t *testing.T) {
	// GIVEN a config with a database location
	cfg := testConfig()
	api := api{config: cfg}
	*api.config.Settings.Data.DatabaseFile = "TestInitialise.db"

	// WHEN the db is initialised with it
	api.initialise()

	// THEN the status table was created in the db
	rows, err := api.db.Query(`
		SELECT	id,
				latest_version,
				latest_version_timestamp,
				deployed_version,
				deployed_version_timestamp,
				approved_version
		 FROM status;`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id  string
			lv  string
			lvt string
			dv  string
			dvt string
			av  string
		)
		err = rows.Scan(&id, &lv, &lvt, &dv, &dvt, &av)
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}

func TestDBQueryService(t *testing.T) {
	// GIVEN a blank DB
	cfg := testConfig()
	api := api{config: cfg}
	*api.config.Settings.Data.DatabaseFile = "TestQueryService.db"
	api.initialise()
	// Get a Service from the Config
	var serviceName string
	for k := range api.config.Service {
		serviceName = k
		break
	}
	svc := api.config.Service[serviceName]

	// WHEN the database contains data for a Service
	api.updateRow(
		serviceName,
		[]dbtype.Cell{
			{Column: "id", Value: serviceName},
			{Column: "latest_version", Value: (*svc).Status.LatestVersion()},
			{Column: "latest_version_timestamp", Value: (*svc).Status.LatestVersionTimestamp()},
			{Column: "deployed_version", Value: (*svc).Status.DeployedVersion()},
			{Column: "deployed_version_timestamp", Value: (*svc).Status.DeployedVersionTimestamp()},
			{Column: "approved_version", Value: (*svc).Status.ApprovedVersion()}})

	// THEN that data can be queried
	got := queryRow(t, api.db, serviceName)
	if (*svc).Status.LatestVersion() != got.LatestVersion() {
		t.Errorf("LatestVersion %q was not pushed to the db. Got %q",
			(*svc).Status.LatestVersion(), got.LatestVersion())
	}
	if (*svc).Status.LatestVersionTimestamp() != got.LatestVersionTimestamp() {
		t.Errorf("LatestVersionTimestamp %q was not pushed to the db. Got %q",
			(*svc).Status.LatestVersionTimestamp(), got.LatestVersionTimestamp())
	}
	if (*svc).Status.DeployedVersion() != got.DeployedVersion() {
		t.Errorf("DeployedVersion %q was not pushed to the db. Got %q\n%v\n%s",
			(*svc).Status.DeployedVersion(), got.DeployedVersion(), got, (*svc).Status.String())
	}
	if (*svc).Status.DeployedVersionTimestamp() != got.DeployedVersionTimestamp() {
		t.Errorf("DeployedVersionTimestamp %q was not pushed to the db. Got %q",
			(*svc).Status.DeployedVersionTimestamp(), got.DeployedVersionTimestamp())
	}
	if (*svc).Status.ApprovedVersion() != got.ApprovedVersion() {
		t.Errorf("ApprovedVersion %q was not pushed to the db. Got %q",
			(*svc).Status.ApprovedVersion(), got.ApprovedVersion())
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}

func TestAPI_RemoveUnknownServices(t *testing.T) {
	// GIVEN a DB with loads of service status'
	cfg := testConfig()
	api := api{config: cfg}
	*api.config.Settings.Data.DatabaseFile = "TestRemoveUnknownServices.db"
	api.initialise()
	sqlStmt := `
	INSERT OR REPLACE INTO status
		(
			id,
			latest_version,
			latest_version_timestamp,
			deployed_version,
			deployed_version_timestamp,
			approved_version
		)
	VALUES`
	for id, svc := range api.config.Service {
		sqlStmt += fmt.Sprintf(" (%q, %q, %q, %q, %q, %q),",
			id,
			svc.Status.LatestVersion(),
			svc.Status.LatestVersionTimestamp(),
			svc.Status.DeployedVersion(),
			svc.Status.DeployedVersionTimestamp(),
			svc.Status.ApprovedVersion(),
		)
	}
	_, err := api.db.Exec(sqlStmt[:len(sqlStmt)-1] + ";")
	if err != nil {
		t.Fatal(err)
	}

	// WHEN the unknown Services are removed with removeUnknownServices
	api.removeUnknownServices()

	// THEN the rows of Services not in .All are returned
	rows, err := api.db.Query(`
	SELECT	id,
			latest_version,
			latest_version_timestamp,
			deployed_version,
			deployed_version_timestamp,
			approved_version
	FROM status;`)
	if err != nil {
		t.Error(err)
	}
	count := 0
	defer rows.Close()
	for rows.Next() {
		count++
		var (
			id  string
			lv  string
			lvt string
			dv  string
			dvt string
			av  string
		)
		err = rows.Scan(&id, &lv, &lvt, &dv, &dvt, &av)
		svc := api.config.Service[id]
		if svc == nil || !util.Contains(api.config.Order, id) {
			t.Errorf("%q should have been removed from the table",
				id)
		}
	}
	if count != len(api.config.Order) {
		t.Errorf("Only %d were left in the table. Expected %d",
			count, len(api.config.Order))
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}

func TestAPI_Run(t *testing.T) {
	// GIVEN a DB is running (see TestMain)

	// WHEN a message is send to the DatabaseChannel targeting latest_version
	target := "keep0"
	cell := dbtype.Cell{Column: "latest_version", Value: "9.9.9"}
	*cfg.DatabaseChannel <- dbtype.Message{
		ServiceID: target,
		Cells:     []dbtype.Cell{cell},
	}
	time.Sleep(time.Second)

	// THEN the cell was changed in the DB
	otherCfg := testConfig()
	*otherCfg.Settings.Data.DatabaseFile = "TestAPI_Run-copy.db"
	bytesRead, err := os.ReadFile(*cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(*otherCfg.Settings.Data.DatabaseFile, bytesRead, os.FileMode(0644))
	if err != nil {
		t.Fatal(err)
	}
	api := api{config: otherCfg}
	api.initialise()
	got := queryRow(t, api.db, target)
	want := svcstatus.Status{}
	want.Init(
		0, 0, 0,
		&target,
		stringPtr("https://example.com"))
	want.SetLatestVersion("9.9.9", false)
	want.SetLatestVersionTimestamp("2022-01-01T01:01:01Z")
	want.SetApprovedVersion("0.0.1", false)
	want.SetDeployedVersion("0.0.0", false)
	want.SetDeployedVersionTimestamp("2020-01-01T01:01:01Z")
	if got.LatestVersion() != want.LatestVersion() {
		t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
			cell.Column, cell.Value, got, want.String())
	}
	api.db.Close()
	os.Remove(*cfg.Settings.Data.DatabaseFile)
	os.Remove(*otherCfg.Settings.Data.DatabaseFile)
}

func TestAPI_extractServiceStatus(t *testing.T) {
	// GIVEN an API on a DB containing atleast 1 row
	cfg := testConfig()
	*cfg.Settings.Data.DatabaseFile = "TestAPI_extractServiceStatus.db"
	defer os.Remove(*cfg.Settings.Data.DatabaseFile)
	go func() {
		api := api{config: cfg}
		api.initialise()
		api.handler()
	}()
	wantStatus := make([]svcstatus.Status, len(cfg.Service))
	// push a random Status for each Service to the DB
	index := 0
	for id, svc := range cfg.Service {
		id := id
		wantStatus[index].ServiceID = &id
		wantStatus[index].SetLatestVersion(fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)), false)
		wantStatus[index].SetLatestVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
		wantStatus[index].SetDeployedVersion(fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)), false)
		wantStatus[index].SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
		wantStatus[index].SetApprovedVersion(fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)), false)

		*cfg.DatabaseChannel <- dbtype.Message{
			ServiceID: id,
			Cells: []dbtype.Cell{
				{Column: "id", Value: id},
				{Column: "latest_version", Value: wantStatus[index].LatestVersion()},
				{Column: "latest_version_timestamp", Value: wantStatus[index].LatestVersionTimestamp()},
				{Column: "deployed_version", Value: wantStatus[index].DeployedVersion()},
				{Column: "deployed_version_timestamp", Value: wantStatus[index].DeployedVersionTimestamp()},
				{Column: "approved_version", Value: wantStatus[index].ApprovedVersion()}}}
		// Clear the Status in the Config
		svc.Status = *svcstatus.New(
			svc.Status.AnnounceChannel, svc.Status.DatabaseChannel, svc.Status.SaveChannel,
			"", "", "", "", "", "")
		index++
	}
	time.Sleep(250 * time.Millisecond)
	api := api{config: cfg}
	api.initialise()

	// WHEN extractServiceStatus is called
	api.extractServiceStatus()

	// THEN the Status in the Config is updated
	for i := range wantStatus {
		row := queryRow(t, api.db, *wantStatus[i].ServiceID)
		if row.LatestVersion() != wantStatus[i].LatestVersion() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"latest_version", row.LatestVersion(), row, wantStatus[i].String())
		}
		if row.LatestVersionTimestamp() != wantStatus[i].LatestVersionTimestamp() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"latest_version_timestamp", row.LatestVersionTimestamp(), row, wantStatus[i].String())
		}
		if row.DeployedVersion() != wantStatus[i].DeployedVersion() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"deployed_version", row.DeployedVersion(), row, wantStatus[i].String())
		}
		if row.DeployedVersionTimestamp() != wantStatus[i].DeployedVersionTimestamp() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"deployed_version_timestamp", row.DeployedVersionTimestamp(), row, wantStatus[i].String())
		}
		if row.ApprovedVersion() != wantStatus[i].ApprovedVersion() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"approved_version", row.ApprovedVersion(), row, wantStatus[i].String())
		}
	}
}

func Test_UpdateColumnTypes(t *testing.T) {
	// GIVEN a DB with the *_version columns as STRING/TEXT
	tests := map[string]struct {
		columnType string
	}{
		"No conversion necessary": {
			columnType: "TEXT"},
		"Conversion wanted": {
			columnType: "STRING"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			db, err := sql.Open("sqlite", "Test_UpdateColumnTypes.db")
			defer os.Remove("Test_UpdateColumnTypes.db")
			if err != nil {
				t.Fatal(err)
			}
			sqlStmt := `
				CREATE TABLE IF NOT EXISTS status (
					id                         TYPE     NOT NULL PRIMARY KEY,
					latest_version             TYPE     DEFAULT  '',
					latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
					deployed_version           TYPE     DEFAULT  '',
					deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
					approved_version           TYPE     DEFAULT  ''
				);`
			sqlStmt = strings.ReplaceAll(sqlStmt, "TYPE", tc.columnType)
			_, err = db.Exec(sqlStmt)
			// Add a row to the table
			id := "keepMe"
			latest_version, latest_version_timestamp := "0.0.3", "2020-01-02T01:01:01Z"
			deployed_version, deployed_version_timestamp := "0.0.2", "2020-01-01T01:01:01Z"
			approved_version := "0.0.1"
			sqlStmt = `
				INSERT OR REPLACE INTO status (
					id,
					latest_version,
					latest_version_timestamp,
					deployed_version,
					deployed_version_timestamp,
					approved_version
				)
				VALUES (
					'` + id + `',
					'` + latest_version + `',
					'` + latest_version_timestamp + `',
					'` + deployed_version + `',
					'` + deployed_version_timestamp + `',
					'` + approved_version + `'
				);`
			_, err = db.Exec(sqlStmt)

			// WHEN updateTable is called
			updateTable(db)

			// THEN the id column and all *_version columns are now TEXT
			wantTextColumns := []string{"id", "latest_version", "deployed_version", "approved_version"}
			for _, row := range wantTextColumns {
				var columnType string
				db.QueryRow("SELECT type FROM pragma_table_info('status') WHERE name = 'latest_version'").Scan(&columnType)
				if columnType != "TEXT" {
					t.Errorf("Expected %q to be %q, not %q",
						row, "TEXT", columnType)
				}
			}
			// AND all rows were carried over
			got := queryRow(t, db, id)
			if got.LatestVersion() != latest_version || got.LatestVersionTimestamp() != latest_version_timestamp ||
				got.DeployedVersion() != deployed_version || got.DeployedVersionTimestamp() != deployed_version_timestamp ||
				got.ApprovedVersion() != approved_version {
				t.Errorf("Row wasn't carried over correctly.\nHad: lv=%q, lvt=%q, dv=%q, dvt=%q, av=%q\nGot: lv=%q, lvt=%q, dv=%q, dvt=%q, av=%q",
					latest_version, latest_version_timestamp, deployed_version, deployed_version_timestamp, approved_version,
					got.LatestVersion(), got.LatestVersionTimestamp(), got.DeployedVersion(), got.DeployedVersionTimestamp(), got.ApprovedVersion())
			}
			// AND the conversion was output to stdout
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			want := "Finished updating column types"
			contains := strings.Contains(output, want)
			if tc.columnType == "TEXT" && contains {
				t.Errorf("Table started as %q, so should not have been updated. Got %q",
					tc.columnType, output)
			} else if tc.columnType == "STRING" && !contains {
				t.Errorf("Table started as %q, so should have been updated. Got %q",
					tc.columnType, output)
			}
		})
	}
}
