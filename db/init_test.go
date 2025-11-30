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
	"database/sql"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	_ "modernc.org/sqlite"
)

func TestCheckFile(t *testing.T) {
	type createOptions struct {
		path  string
		perms fs.FileMode
	}
	// GIVEN various paths.
	tests := map[string]struct {
		removeBefore     string
		createDirBefore  []createOptions
		createFileBefore string
		path             string
		panicRegex       string
	}{
		"file doesn't exist": {
			path:         "something_does_not_exist.db",
			removeBefore: "something_does_not_exist.db"},
		"dir doesn't exist, so is created": {
			path:         "dir_does_not_exist_1/argus.db",
			removeBefore: "dir_does_not_exist_1"},
		"dir exists but not file": {
			path:            "dir_does_exist_2/argus.db",
			createDirBefore: []createOptions{{path: "dir_does_exist_2", perms: 0_750}}},
		"file is dir": {
			path:            "folder.db",
			createDirBefore: []createOptions{{path: "folder.db", perms: 0_750}},
			panicRegex:      `path .* is a directory, not a file`},
		"dir is file": {
			path:             "item_not_a_dir/argus.db",
			createFileBefore: "item_not_a_dir",
			panicRegex:       `path .* is not a directory`},
		"no perms to create dir": {
			path: "dir_no_perms_1/dir_no_perms_2/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_perms_1", perms: 0_555}},
			panicRegex: `mkdir .* permission denied`},
		"no perms to check for file in dir": {
			path: "dir_no_perms_3/dir_no_perms_4/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_perms_3", perms: 0_444},
				{path: "dir_no_perms_3/dir_no_perms_4", perms: 0_444}},
			panicRegex: `stat .* permission denied`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			os.RemoveAll(tc.removeBefore)
			t.Cleanup(func() { os.RemoveAll(tc.removeBefore) })
			if len(tc.createDirBefore) > 0 {
				os.RemoveAll(tc.createDirBefore[0].path)
				t.Cleanup(func() {
					for _, dir := range tc.createDirBefore {
						os.Chmod(dir.path, 0_750)
					}
					os.RemoveAll(tc.createDirBefore[0].path)
				})
				for _, dir := range tc.createDirBefore {
					err := os.Mkdir(dir.path, 0_750)
					if err != nil {
						t.Fatalf("%s",
							err)
					}
				}
				// Set perms (in reverse order).
				for i := len(tc.createDirBefore) - 1; i >= 0; i-- {
					if err := os.Chmod(tc.createDirBefore[i].path, tc.createDirBefore[i].perms); err != nil {
						t.Fatalf("%s",
							err)
					}
				}
			}
			if tc.createFileBefore != "" {
				file, err := os.Create(tc.createFileBefore)
				if err != nil {
					t.Fatalf("%s",
						err)
				}
				file.Close()
				t.Cleanup(func() { os.Remove(tc.createFileBefore) })
			}
			if tc.panicRegex != "" {
				defer func() {
					r := recover()

					rStr := fmt.Sprint(r)
					if !util.RegexCheck(tc.panicRegex, rStr) {
						t.Errorf("%s\nwant: %q\ngot:  %q",
							packageName, tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN checkFile is called on that same dir.
			checkFile(tc.path)

			// THEN we get here only when we should.
			if tc.panicRegex != "" {
				t.Fatalf("%s\nExpected panic with %q",
					packageName, tc.panicRegex)
			}
		})
	}
}

func TestAPI_Initialise(t *testing.T) {
	// GIVEN a config with a database location.
	tests := map[string]struct {
		unreadableDB bool
	}{
		"DB is readable": {
			unreadableDB: false},
		"DB is unreadable": {
			unreadableDB: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(name, "TestAPI_Initialise")
			if tc.unreadableDB {
				os.Create(tAPI.config.Settings.DataDatabaseFile())
				os.Chmod(tAPI.config.Settings.DataDatabaseFile(), 0_000)
			}
			defer dbCleanup(tAPI)
			// Catch fatal panics.
			defer func() {
				r := recover()
				// Ignore nil panics.
				if r == nil {
					return
				}

				if !tc.unreadableDB {
					t.Fatalf("%s\nunexpected panic: %v",
						packageName, r)
				}
			}()

			// WHEN the db is initialised with it.
			tAPI.initialise()

			// THEN the app panic'd if the db was unreadable.
			if tc.unreadableDB {
				t.Errorf("%s\nExpected a panic",
					packageName)
				return
			}
			// THEN the status table was created in the db.
			rows, err := tAPI.db.Query(`
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
			t.Cleanup(func() { rows.Close() })
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
		})
	}
}

func TestDBQueryService(t *testing.T) {
	// GIVEN a blank DB.
	tAPI := testAPI("TestDBQueryService", "db")
	t.Cleanup(func() { dbCleanup(tAPI) })
	tAPI.initialise()
	// Get a Service from the Config.
	var serviceName string
	for k := range tAPI.config.Service {
		serviceName = k
		break
	}
	svc := tAPI.config.Service[serviceName]

	// WHEN the database contains data for a Service.
	tAPI.updateRow(
		serviceName,
		[]dbtype.Cell{
			{Column: "id", Value: serviceName},
			{Column: "latest_version", Value: svc.Status.LatestVersion()},
			{Column: "latest_version_timestamp", Value: svc.Status.LatestVersionTimestamp()},
			{Column: "deployed_version", Value: svc.Status.DeployedVersion()},
			{Column: "deployed_version_timestamp", Value: svc.Status.DeployedVersionTimestamp()},
			{Column: "approved_version", Value: svc.Status.ApprovedVersion()}},
	)

	// THEN that data can be queried.
	got := queryRow(t, tAPI.db, serviceName)
	if got.LatestVersion() != svc.Status.LatestVersion() {
		t.Errorf("%s\nLatestVersion db mismatch\nwant: %q\ngot:  %q",
			packageName, svc.Status.LatestVersion(), got.LatestVersion())
	}
	if got.LatestVersionTimestamp() != svc.Status.LatestVersionTimestamp() {
		t.Errorf("%s\nLatestVersionTimestamp db mismatch\nwant: %q\ngot:  %q",
			packageName, svc.Status.LatestVersionTimestamp(), got.LatestVersionTimestamp())
	}
	if got.DeployedVersion() != svc.Status.DeployedVersion() {
		t.Errorf("%s\nDeployedVersion db mismatch\nwant: %q\ngot:  %q",
			packageName, svc.Status.DeployedVersion(), got.DeployedVersion())
	}
	if got.DeployedVersionTimestamp() != svc.Status.DeployedVersionTimestamp() {
		t.Errorf("%s\nDeployedVersionTimestamp db mismatch\nwant: %q\ngot:  %q",
			packageName, svc.Status.DeployedVersionTimestamp(), got.DeployedVersionTimestamp())
	}
	if got.ApprovedVersion() != svc.Status.ApprovedVersion() {
		t.Errorf("%s\nApprovedVersion db mismatch\nwant: %q\ngot:  %q",
			packageName, svc.Status.ApprovedVersion(), got.ApprovedVersion())
	}
}

func TestAPI_RemoveUnknownServices(t *testing.T) {
	// GIVEN a DB with many service status'.
	tests := map[string]struct {
		databaseDeleted bool
	}{
		"DB is not deleted": {
			databaseDeleted: false},
		"DB is deleted": {
			databaseDeleted: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(name, "TestAPI_RemoveUnknownServices")
			t.Cleanup(func() { dbCleanup(tAPI) })
			tAPI.initialise()
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
			for id, svc := range tAPI.config.Service {
				sqlStmt += fmt.Sprintf(" (%q, %q, %q, %q, %q, %q),",
					id,
					svc.Status.LatestVersion(),
					svc.Status.LatestVersionTimestamp(),
					svc.Status.DeployedVersion(),
					svc.Status.DeployedVersionTimestamp(),
					svc.Status.ApprovedVersion(),
				)
			}
			if len(tAPI.config.Order) <= 1 {
				t.Fatalf("%s\nWant to test with more services present, have=%d",
					packageName, len(tAPI.config.Order))
			}
			util.RemoveIndex(&tAPI.config.Order, 0)
			_, err := tAPI.db.Exec(sqlStmt[:len(sqlStmt)-1] + ";")
			if err != nil {
				t.Fatal(err)
			}
			// Catch fatal panics.
			defer func() {
				r := recover()
				// Ignore nil panics.
				if r == nil {
					return
				}

				if !tc.databaseDeleted {
					t.Fatalf("%s\nunexpected panic: %v",
						packageName, r)
				}
			}()
			// Delete the DB file.
			if tc.databaseDeleted {
				os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			}

			// WHEN the unknown Services are removed with removeUnknownServices.
			tAPI.removeUnknownServices()

			// THEN the app panic'd if the db was deleted.
			if tc.databaseDeleted {
				t.Errorf("%s\nExpected a panic",
					packageName)
				return
			}
			// AND the rows of Services not in .All are returned.
			rows, err := tAPI.db.Query(`
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
			t.Cleanup(func() { rows.Close() })
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
				svc := tAPI.config.Service[id]
				if svc == nil || !util.Contains(tAPI.config.Order, id) {
					t.Errorf("%s\n%q should have been removed from the table",
						packageName, id)
				}
			}
			if count != len(tAPI.config.Order) {
				t.Errorf("%s\nOrder length mismatch\nwant: %d\ngot:  %d",
					packageName, len(tAPI.config.Order), count)
			}
		})
	}
}

func TestAPI_Run(t *testing.T) {
	// GIVEN a DB is running (see TestMain).

	// WHEN a message is send to the DatabaseChannel targeting latest_version.
	target := "keep0"
	cell := dbtype.Cell{Column: "latest_version", Value: "9.9.9"}
	*cfg.DatabaseChannel <- dbtype.Message{
		ServiceID: target,
		Cells:     []dbtype.Cell{cell},
	}
	time.Sleep(time.Second)

	// THEN the cell was changed in the DB.
	otherCfg := testConfig()
	otherCfg.Settings.Data.DatabaseFile = "TestAPI_Run-copy.db"
	bytesRead, err := os.ReadFile(cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatalf("%s\n%v",
			packageName, err)
	}
	t.Cleanup(func() { os.Remove(otherCfg.Settings.Data.DatabaseFile) })
	err = os.WriteFile(otherCfg.Settings.Data.DatabaseFile, bytesRead, os.FileMode(0644))
	if err != nil {
		t.Fatalf("%s\n%v",
			packageName, err)
	}
	tAPI := api{config: otherCfg}
	t.Cleanup(func() { dbCleanup(&tAPI) })
	tAPI.initialise()
	got := queryRow(t, tAPI.db, target)
	want := status.Status{}
	want.Init(
		0, 0, 0,
		target, "", "",
		&dashboard.Options{
			WebURL: "https://example.com"})
	want.SetLatestVersion("9.9.9", "2022-01-01T01:01:01Z", false)
	want.SetApprovedVersion("0.0.1", false)
	want.SetDeployedVersion("0.0.0", "2020-01-01T01:01:01Z", false)
	if got.LatestVersion() != want.LatestVersion() {
		t.Errorf("%s\nwant: %q to be updated to %q\ngot:  %q",
			packageName, cell.Column, cell.Value, got)
	}
}

func TestAPI_extractServiceStatus(t *testing.T) {
	// GIVEN an API on a DB containing at least 1 row.
	tAPI := testAPI("TestAPI_extractServiceStatus", "db")
	t.Cleanup(func() { dbCleanup(tAPI) })
	tAPI.initialise()
	go tAPI.handler()
	wantStatus := make([]status.Status, len(cfg.Service))
	// Push a random Status for each Service to the DB.
	index := 0
	for id, svc := range tAPI.config.Service {
		id := id
		wantStatus[index].Init(
			0, 0, 0,
			"", "", "",
			&dashboard.Options{})

		wantStatus[index].ServiceInfo.ID = id
		wantStatus[index].SetLatestVersion(fmt.Sprintf("%d.%d.%d",
			rand.Intn(10), rand.Intn(10), rand.Intn(10)),
			time.Now().UTC().Format(time.RFC3339), false)
		wantStatus[index].SetDeployedVersion(fmt.Sprintf("%d.%d.%d",
			rand.Intn(10), rand.Intn(10), rand.Intn(10)),
			"", false)
		wantStatus[index].SetApprovedVersion(fmt.Sprintf("%d.%d.%d",
			rand.Intn(10), rand.Intn(10), rand.Intn(10)),
			false)

		*tAPI.config.DatabaseChannel <- dbtype.Message{
			ServiceID: id,
			Cells: []dbtype.Cell{
				{Column: "id", Value: id},
				{Column: "latest_version", Value: wantStatus[index].LatestVersion()},
				{Column: "latest_version_timestamp", Value: wantStatus[index].LatestVersionTimestamp()},
				{Column: "deployed_version", Value: wantStatus[index].DeployedVersion()},
				{Column: "deployed_version_timestamp", Value: wantStatus[index].DeployedVersionTimestamp()},
				{Column: "approved_version", Value: wantStatus[index].ApprovedVersion()}}}
		// Clear the Status in the Config.
		svc.Status = *status.New(
			svc.Status.AnnounceChannel, svc.Status.DatabaseChannel, svc.Status.SaveChannel,
			"",
			"", "",
			"", "",
			"",
			nil)
		svc.Status.Init(
			0, 0, 0,
			"", "", "",
			&svc.Dashboard)
		index++
	}
	time.Sleep(250 * time.Millisecond)

	// WHEN extractServiceStatus is called.
	tAPI.extractServiceStatus()

	// THEN the Status in the Config is updated.
	errMsg := "%s\n%q not updated correctly\nwant: %q\ngot:  %q (%+v)"
	for i := range wantStatus {
		row := queryRow(t, tAPI.db, wantStatus[i].ServiceInfo.ID)
		if row.LatestVersion() != wantStatus[i].LatestVersion() {
			t.Errorf(errMsg,
				packageName, "latest_version", wantStatus[i].LatestVersion(), row.LatestVersion(), row)
		}
		if row.LatestVersionTimestamp() != wantStatus[i].LatestVersionTimestamp() {
			t.Errorf(errMsg,
				packageName, "latest_version_timestamp", wantStatus[i].LatestVersionTimestamp(), row.LatestVersionTimestamp(), row)
		}
		if row.DeployedVersion() != wantStatus[i].DeployedVersion() {
			t.Errorf(errMsg,
				packageName, "deployed_version", wantStatus[i].DeployedVersion(), row.DeployedVersion(), row)
		}
		if row.DeployedVersionTimestamp() != wantStatus[i].DeployedVersionTimestamp() {
			t.Errorf(errMsg,
				packageName, "deployed_version_timestamp", wantStatus[i].DeployedVersionTimestamp(), row.DeployedVersionTimestamp(), row)
		}
		if row.ApprovedVersion() != wantStatus[i].ApprovedVersion() {
			t.Errorf(errMsg,
				packageName, "approved_version", wantStatus[i].ApprovedVersion(), row.ApprovedVersion(), row)
		}
	}
}

func Test_UpdateTypes(t *testing.T) {
	// GIVEN a DB with the *_version columns as STRING/TEXT.
	tests := map[string]struct {
		columnType                         string
		databaseDeleted, backupTableExists bool
		cannotDropTable, cannotAlterTable  bool
	}{
		"No conversion necessary": {
			columnType: "TEXT"},
		"Conversion wanted": {
			columnType: "STRING"},
		"DB is deleted": {
			columnType:      "STRING",
			databaseDeleted: true},
		"Backup exists with different fields": {
			columnType:        "STRING",
			backupTableExists: true},
		"Cannot drop table because of foreign key": {
			columnType:      "STRING",
			cannotDropTable: true},
		"Cannot alter backup table because of foreign key": {
			columnType:       "STRING",
			cannotAlterTable: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			databaseFile := strings.ReplaceAll(
				fmt.Sprintf("%s-Test_UpdateColumnTypes.db", name),
				" ", "_")
			db, err := sql.Open("sqlite", databaseFile)
			// Enable foreign key constraint enforcement.
			if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
				t.Fatalf("%s\nFailed to enable foreign key constraints: %s",
					packageName, err)
			}
			t.Cleanup(func() { os.Remove(databaseFile) })
			if err != nil {
				t.Fatalf("%s\n%v",
					packageName, err)
			}
			sqlStmtTable := `
				CREATE TABLE IF NOT EXISTS status (
					id                         TYPE     NOT NULL PRIMARY KEY,
					latest_version             TYPE     DEFAULT  '',
					latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
					deployed_version           TYPE     DEFAULT  '',
					deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
					approved_version           TYPE     DEFAULT  ''
				);`
			sqlStmtTable = strings.ReplaceAll(sqlStmtTable, "TYPE", tc.columnType)
			db.Exec(sqlStmtTable)
			// Add a row to the table.
			id := "keepMe"
			latestVersion, latestVersionTimestamp := "0.0.3", "2020-01-02T01:01:01Z"
			deployedVersion, deployedVersionTimestamp := "0.0.2", "2020-01-01T01:01:01Z"
			approvedVersion := "0.0.1"
			sqlStmt := `
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
					'` + latestVersion + `',
					'` + latestVersionTimestamp + `',
					'` + deployedVersion + `',
					'` + deployedVersionTimestamp + `',
					'` + approvedVersion + `'
				);`
			db.Exec(sqlStmt)
			// Catch fatal panics.
			expectPanic := tc.databaseDeleted ||
				tc.backupTableExists ||
				tc.cannotDropTable ||
				tc.cannotAlterTable
			defer func() {
				r := recover()
				// Ignore nil panics.
				if r == nil {
					return
				}

				releaseStdout()
				if !expectPanic {
					t.Fatalf("%s\nunexpected panic: %v",
						packageName, r)
				}
			}()
			// Delete the DB file.
			if tc.databaseDeleted {
				os.Remove(databaseFile)

				// Create a backup table with different fields.
			} else if tc.backupTableExists {
				db.Exec(`
					CREATE TABLE IF NOT EXISTS status_backup (
						id  TEXT  NOT NULL  PRIMARY KEY
					);`)

				// Create a foreign key to prevent dropping the status table.
			} else if tc.cannotDropTable {
				db.Exec(`
					CREATE TABLE IF NOT EXISTS fk_table (
						id     INTEGER  NOT NULL  PRIMARY KEY,
						fk_id  TEXT     NOT NULL,
						FOREIGN KEY (fk_id) REFERENCES status(id)
					);`)
				// Create a row in the fk_table.
				db.Exec(`
					INSERT OR REPLACE INTO fk_table ( fk_id )
					VALUES ( '` + id + `' );`)

				// Create a trigger to prevent altering the status_backup table.
			} else if tc.cannotAlterTable {
				db.Exec(strings.Replace(
					sqlStmtTable, "status", "status_backup", 1))
				db.Exec(`
					CREATE TRIGGER trigger_name
					AFTER INSERT ON status_backup
					BEGIN
						INSERT INTO status (id) VALUES ('name');
					END;`)
			}

			// WHEN updateTable is called.
			updateTable(db)

			// THEN the app panic'd if the db was deleted, or the table manipulation failed.
			if expectPanic {
				t.Errorf("%s\nExpected a panic",
					packageName)
				releaseStdout()
				return
			}
			// AND the ID column and all *_version columns are now TEXT.
			wantTextColumns := []string{"id", "latest_version", "deployed_version", "approved_version"}
			for _, row := range wantTextColumns {
				var columnType string
				db.QueryRow("SELECT type FROM pragma_table_info('status') WHERE name = 'latest_version'").Scan(&columnType)
				if columnType != "TEXT" {
					t.Errorf("%s\ncolumn type mismatch on %q\nwant: %q\ngot:  %q",
						packageName, row, "TEXT", columnType)
				}
			}
			// AND all rows were carried over.
			got := queryRow(t, db, id)
			if got.LatestVersion() != latestVersion || got.LatestVersionTimestamp() != latestVersionTimestamp ||
				got.DeployedVersion() != deployedVersion || got.DeployedVersionTimestamp() != deployedVersionTimestamp ||
				got.ApprovedVersion() != approvedVersion {
				t.Errorf("%s\nRow wasn't carried over correctly.\nHad: lv=%q, lvt=%q, dv=%q, dvt=%q, av=%q\nGot: lv=%q, lvt=%q, dv=%q, dvt=%q, av=%q",
					packageName,
					latestVersion, latestVersionTimestamp,
					deployedVersion, deployedVersionTimestamp,
					approvedVersion,
					got.LatestVersion(), got.LatestVersionTimestamp(),
					got.DeployedVersion(), got.DeployedVersionTimestamp(),
					got.ApprovedVersion())
			}
			// AND the conversion was printed to stdout.
			stdout := releaseStdout()
			want := "Finished updating column types"
			contains := strings.Contains(stdout, want)
			if tc.columnType == "TEXT" && contains {
				t.Errorf("%s\nTable started as %q, so should not have been updated\ngot: %q",
					packageName, tc.columnType, stdout)
			} else if tc.columnType == "STRING" && !contains {
				t.Errorf("%s\nTable started as %q, so should have been updated\ngot: %q",
					packageName, tc.columnType, stdout)
			}
			time.Sleep(100 * time.Millisecond)
		})
	}
}
