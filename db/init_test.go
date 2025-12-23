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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	_ "modernc.org/sqlite"
)

func TestCheckFile(t *testing.T) {
	type createOptions struct {
		path  string
		perms fs.FileMode
	}
	// GIVEN various paths.
	tests := map[string]struct {
		createDirBefore  []createOptions
		dirPerms         []createOptions
		createFileBefore string
		path             string
		stdoutRegex      string
	}{
		"file doesn't exist": {
			path:        "something_does_not_exist.db",
			stdoutRegex: `^$`},
		"dir doesn't exist, so is created": {
			path:        "dir_does_not_exist_1/argus.db",
			stdoutRegex: `^$`},
		"dir exists but not file": {
			path:            "dir_does_exist_2/argus.db",
			createDirBefore: []createOptions{{path: "dir_does_exist_2", perms: 0_750}},
			stdoutRegex:     `^$`},
		"file is dir": {
			path:            "folder.db",
			createDirBefore: []createOptions{{path: "folder.db", perms: 0_750}},
			stdoutRegex:     `path .* is a directory, not a file`},
		"dir is file": {
			path:             "item_not_a_dir/argus.db",
			createFileBefore: "item_not_a_dir",
			stdoutRegex:      `path .* is not a directory`},
		"no perms to create dir": {
			path: "dir_no_perms_1/dir_no_perms_2/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_perms_1", perms: 0_555}},
			stdoutRegex: `mkdir .* permission denied`},
		"no perms to check for file in dir": {
			path: "dir_no_perms_3/dir_no_perms_4/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_perms_3", perms: 0_444},
				{path: "dir_no_perms_3/dir_no_perms_4", perms: 0_444}},
			stdoutRegex: `stat .* permission denied`},
		"cannot stat existing file due to dir perms": {
			path: "dir_no_exec/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_exec", perms: 0_666},
			},
			dirPerms: []createOptions{
				{path: "dir_no_exec", perms: 0_666},
			},
			createFileBefore: "dir_no_exec/argus.db",
			stdoutRegex:      `stat .* permission denied`},
		"file does exist": {
			path:             "file_does_exist.db",
			createFileBefore: "file_does_exist.db",
			stdoutRegex:      `^$`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			baseDir := t.TempDir()
			// Append baseDir to all paths.
			for i, dir := range tc.createDirBefore {
				tc.createDirBefore[i].path = filepath.Join(baseDir, dir.path)
			}
			for i, dir := range tc.dirPerms {
				tc.dirPerms[i].path = filepath.Join(baseDir, dir.path)
			}
			if tc.createFileBefore != "" {
				tc.createFileBefore = filepath.Join(baseDir, tc.createFileBefore)
			}
			tc.path = filepath.Join(baseDir, tc.path)

			if len(tc.createDirBefore) > 0 {
				_ = os.RemoveAll(tc.createDirBefore[0].path)
				t.Cleanup(func() {
					for _, dir := range tc.createDirBefore {
						_ = os.Chmod(dir.path, 0_750)
					}
					_ = os.RemoveAll(tc.createDirBefore[0].path)
				})
				for _, dir := range tc.createDirBefore {
					err := os.Mkdir(dir.path, 0_750)
					if err != nil {
						t.Fatalf("%s\ncreateDirBefore mkdir on %q-%q\nerror: %s",
							packageName, dir.path, 0_750, err)
					}
				}
			}
			if tc.createFileBefore != "" {
				file, err := os.Create(tc.createFileBefore)
				if err != nil {
					t.Fatalf("%s\ncreateFileBefore on %q\nerror: %s",
						packageName, tc.createFileBefore, err)
				}
				_ = file.Close()
				t.Cleanup(func() { _ = os.Remove(tc.createFileBefore) })
			}
			// Set perms (in reverse order).
			for i := len(tc.createDirBefore) - 1; i >= 0; i-- {
				if err := os.Chmod(tc.createDirBefore[i].path, tc.createDirBefore[i].perms); err != nil {
					t.Fatalf("%s\ncreateDirBefore chmod on %q-%q\nerror: %s",
						packageName, tc.createDirBefore[i].path, tc.createDirBefore[i].perms, err)
				}
			}

			resultChannel := make(chan bool, 1)
			// WHEN checkFile is called on that same dir.
			resultChannel <- checkFile(tc.path)

			// THEN if false is returned, the error is logged.
			wantOk := tc.stdoutRegex == `^$`
			if err := test.OkMatch(t, wantOk, resultChannel, logutil.ExitCodeChannel(), releaseStdout); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			// AND the stdout matches the expected result.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
					packageName, tc.stdoutRegex, stdout)
			}
		})
	}
}

func TestAPI_Initialise(t *testing.T) {
	type pathInfo struct {
		path     string
		perms    os.FileMode
		itemType string
		data     string
	}
	// GIVEN a config with a database location.
	tests := map[string]struct {
		ok    bool
		paths []pathInfo
	}{
		"DB rw-r-r": {
			ok: true,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_644, itemType: "file"},
			},
		},
		"DB -w--w--w-": {
			ok: false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_444, itemType: "file"},
			},
		},
		"DB r--r--r--": {
			ok: false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_222, itemType: "file"},
			},
		},
		"DB --x--x--x": {
			ok: false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_111, itemType: "file"},
			},
		},
		"Dir not writable": {
			ok: false,
			paths: []pathInfo{
				{path: "test", perms: 0_555, itemType: "dir"},
				{path: "test/argus.db", perms: 0_644, itemType: "file"},
			},
		},
		"Non-DB file data": {
			ok: false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_644, itemType: "file", data: "something"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(t)
			baseDir := filepath.Dir(tAPI.config.Settings.Data.DatabaseFile)
			var dbFile string
			for i, p := range tc.paths {
				tc.paths[i].path = filepath.Join(baseDir, p.path)
				if p.itemType == "file" {
					dbFile = tc.paths[i].path
					_, _ = os.Create(tc.paths[i].path)
					if p.data != "" {
						_ = os.WriteFile(tc.paths[i].path, []byte(p.data), 0_644)
					}
					_ = os.Chmod(tc.paths[i].path, 0_644)
				} else {
					_ = os.Mkdir(tc.paths[i].path, 0_755)
				}
			}
			for i := len(tc.paths) - 1; i >= 0; i-- {
				if err := os.Chmod(tc.paths[i].path, tc.paths[i].perms); err != nil {
					if tc.paths[i].itemType == "dir" {
						t.Fatalf("%s\npaths chmod on %q (%q)\nerror: %s",
							packageName, tc.paths[i].path, tc.paths[i].perms, err)
					}
				}
			}
			tAPI.config.Settings.Data.DatabaseFile = dbFile
			t.Cleanup(func() {
				for _, p := range tc.paths {
					_ = os.Chmod(p.path, 0_755)
				}
				dbCleanup(tAPI)
			})

			resultChannel := make(chan bool, 1)
			// WHEN the db is initialised with it.
			resultChannel <- tAPI.initialise()

			// THEN the function failed if the db was unreadable.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			if !tc.ok {
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
			t.Cleanup(func() { _ = rows.Close() })
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
	tAPI := testAPI(t)
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

			tAPI := testAPI(t)
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
			// Delete the DB file.
			if tc.databaseDeleted {
				os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			}
			wantOk := !tc.databaseDeleted

			resultChannel := make(chan bool, 1)
			// WHEN the unknown Services are removed with removeUnknownServices.
			resultChannel <- tAPI.removeUnknownServices()

			// THEN the app panicked if the db was deleted.
			if err := test.OkMatch(t, wantOk, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			if !wantOk {
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
			t.Cleanup(func() { _ = rows.Close() })
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
	cfg.DatabaseChannel <- dbtype.Message{
		ServiceID: target,
		Cells:     []dbtype.Cell{cell},
	}
	time.Sleep(time.Second)

	// THEN the cell was changed in the DB.
	otherCfg := testConfig(t)
	otherCfg.Settings.Data.DatabaseFile = "TestAPI_Run-copy.db"
	bytesRead, err := os.ReadFile(cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatalf("%s\n%v",
			packageName, err)
	}
	t.Cleanup(func() { _ = os.Remove(otherCfg.Settings.Data.DatabaseFile) })
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

func TestAPI_Run_Fail(t *testing.T) {
	// GIVEN a DB to setup.
	tests := map[string]struct {
		setupDB      func(cfg *config.Config)
		deleteDBFile bool
		ok           bool
	}{
		"initialise fails": {
			setupDB: func(cfg *config.Config) {
				// Invalid DB path.
				cfg.Settings.Data.DatabaseFile = "/invalid/path.db"
			},
			ok: false,
		},
		"removeUnknownServices returns false": {
			setupDB: func(cfg *config.Config) {
				// Drop required column so removeUnknownServices fails.
				tAPI := testAPI(t)
				t.Cleanup(func() { dbCleanup(tAPI) })
				tAPI.initialise()
				defer tAPI.db.Close()
				if _, err := tAPI.db.Exec("DROP TABLE status"); err != nil {
					t.Fatalf("%s\nerror dropping table in setupDB:\n%v", packageName, err)
				}
				if _, err := tAPI.db.Exec(`
				CREATE TABLE IF NOT EXISTS status (
					latest_version TEXT,
					latest_version_timestamp TEXT,
					deployed_version TEXT,
					deployed_version_timestamp TEXT,
					approved_version TEXT
				);`); err != nil {
					t.Fatalf("%s\nerror creating table in setupDB:\n%v", packageName, err)
				}
				if _, err := tAPI.db.Exec(`
					INSERT INTO status (
						latest_version,
						latest_version_timestamp,
						deployed_version,
						deployed_version_timestamp,
						approved_version
					) VALUES (
					  '0.0.0',
					  '2020-01-01T01:01:01Z',
					  '0.0.0',
					  '2020-01-01T01:01:01Z',
					  '0.0.1')`); err != nil {
					t.Fatalf("%s\nerror inserting row in setupDB:\n%v", packageName, err)
				}
				cfg.Settings.Data.DatabaseFile = tAPI.config.Settings.Data.DatabaseFile
			},
			ok: false,
		},
		"extractServiceStatus returns false": {
			setupDB: func(cfg *config.Config) {
				// Drop required column so extractServiceStatus fails.
				tAPI := testAPI(t)
				t.Cleanup(func() { dbCleanup(tAPI) })
				tAPI.initialise()
				defer tAPI.db.Close()
				if _, err := tAPI.db.Exec("ALTER TABLE status DROP COLUMN latest_version;"); err != nil {
					t.Fatalf("%s\nerror dropping column in setupDB:\n%v", packageName, err)
				}
				if _, err := tAPI.db.Exec(`
					INSERT INTO status (
						id,
						latest_version_timestamp,
						deployed_version,
						deployed_version_timestamp,
						approved_version
					) VALUES (
					  'service1',
					  '2020-01-01T01:01:01Z',
					  '0.0.0',
					  '2020-01-01T01:01:01Z',
					  '0.0.1')`); err != nil {
					t.Fatalf("%s\nerror inserting row in setupDB:\n%v", packageName, err)
				}
				cfg.Settings.Data.DatabaseFile = tAPI.config.Settings.Data.DatabaseFile
			},
			ok: false,
		},
		"success": {
			setupDB: func(cfg *config.Config) {},
			ok:      true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			cfg := testConfig(t)
			tc.setupDB(cfg)

			resultChannel := make(chan bool, 1)
			// WHEN that DB is run.
			go func() { resultChannel <- Run(t.Context(), cfg) }()

			// THEN the ok value is as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), releaseStdout); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			_ = releaseStdout()
		})
	}
}

func TestAPI_extractServiceStatus(t *testing.T) {
	// GIVEN an API on a DB containing at least 1 row.
	tAPI := testAPI(t)
	t.Cleanup(func() { dbCleanup(tAPI) })
	tAPI.initialise()
	go tAPI.handler(t.Context())
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

		tAPI.config.DatabaseChannel <- dbtype.Message{
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

func TestAPI_extractServiceStatus_Fail(t *testing.T) {
	// GIVEN an API with different 'status' columns.
	tests := map[string]struct {
		errRegex        string
		tableCreateStmt string
		insertDataStmt  string
	}{
		"latest_version missing": {
			errRegex: `^FATAL: db.*no such column: latest_version`,
			tableCreateStmt: `
				CREATE TABLE IF NOT EXISTS status (
					id TEXT PRIMARY KEY,
					latest_version_timestamp TEXT,
					deployed_version TEXT,
					deployed_version_timestamp TEXT,
					approved_version TEXT
				);`,
		},
		"latest_version unexpected type": {
			errRegex: `^FATAL: db.*converting NULL to string is unsupported`,
			tableCreateStmt: `
				CREATE TABLE IF NOT EXISTS status (
					id TEXT PRIMARY KEY,
					latest_version TEXT,
					latest_version_timestamp TEXT,
					deployed_version TEXT,
					deployed_version_timestamp TEXT,
					approved_version TEXT
				);`,
			insertDataStmt: `
				INSERT OR REPLACE INTO status (
					id,
					latest_version,
					latest_version_timestamp,
					deployed_version,
					deployed_version_timestamp,
					approved_version
				)
				VALUES (
					'service1',
					NULL,
					'2020-01-01T01:01:01Z',
					'0.0.1',
					'2020-01-01T01:01:01Z',
					'0.0.0'
				);`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			tAPI := testAPI(t)
			t.Cleanup(func() { dbCleanup(tAPI) })
			tAPI.initialise()
			_, _ = tAPI.db.Exec("DROP TABLE IF EXISTS status;")
			_, _ = tAPI.db.Exec(tc.tableCreateStmt)
			if tc.insertDataStmt != "'" {
				_, _ = tAPI.db.Exec(tc.insertDataStmt)
			}

			go tAPI.handler(t.Context())

			resultChannel := make(chan bool, 1)
			// WHEN extractServiceStatus is called.
			resultChannel <- tAPI.extractServiceStatus()

			// THEN the Status in the Config is updated.
			if err := test.OkMatch(t, false, resultChannel, logutil.ExitCodeChannel(), releaseStdout); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, stdout)
			}
		})
	}
}

func Test_UpdateTypes(t *testing.T) {
	// GIVEN a DB with the *_version columns as STRING/TEXT.
	tests := map[string]struct {
		columnType  string
		setupDB     func(t *testing.T, db **sql.DB, dbFile string)
		ok          bool
		stdoutRegex string
	}{
		"No conversion necessary": {
			columnType: "TEXT",
			ok:         true,
		},
		"Conversion wanted": {
			columnType: "STRING",
			ok:         true,
		},
		"DB is deleted": {
			columnType:  "STRING",
			stdoutRegex: `no rows in result set`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				_, _ = (*db).Exec("DROP TABLE IF EXISTS status;")
			},
			ok: false,
		},
		"Backup exists with different fields": {
			columnType:  "STRING",
			stdoutRegex: `copy.*table .* has \d columns but \d values were supplied`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				_, _ = (*db).Exec(`
					CREATE TABLE IF NOT EXISTS status_backup (
						id TEXT NOT NULL PRIMARY KEY
					);`)
			},
			ok: false,
		},
		"Cannot drop table because of foreign key": {
			columnType:  "STRING",
			stdoutRegex: `drop.*FOREIGN KEY constraint failed`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				fkTable := "status_backup_fk"
				_, _ = (*db).Exec(`
					CREATE TABLE IF NOT EXISTS ` + fkTable + ` (
						id INTEGER NOT NULL PRIMARY KEY,
						fk_id TEXT NOT NULL,
						FOREIGN KEY(fk_id) REFERENCES status(id)
					);`)
				_, _ = (*db).Exec(`INSERT OR REPLACE INTO ` + fkTable + ` (fk_id) VALUES ('keepMe');`)
			},
			ok: false,
		},
		"Cannot alter backup table because of foreign key": {
			columnType:  "STRING",
			stdoutRegex: `rename.*no such table`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				backupTableName := "status_backup"
				_ = createStatusTable(*db, backupTableName, "STRING")
				_, _ = (*db).Exec(`
					CREATE TRIGGER ` + backupTableName + `
					AFTER INSERT ON status_backup
					BEGIN
						INSERT INTO status (id) VALUES ('name');
					END;`)
			},
			ok: false,
		},
		"latest_version column doesn't exist": {
			stdoutRegex: `no rows in result set`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				if _, err := (*db).Exec(`
					ALTER TABLE status
					DROP COLUMN latest_version;
				`); err != nil {
					t.Fatalf("%s\nfailed to drop column: %v",
						packageName, err)
				}
			},
			ok: false,
		},
		"Read-only DB": {
			stdoutRegex: `create.*attempt to write a readonly database`,
			setupDB: func(t *testing.T, db **sql.DB, dbFile string) {
				_ = (*db).Close()
				newDB, err := sql.Open("sqlite", "file:"+dbFile+"?mode=ro")
				if err != nil {
					t.Fatalf("%s\nfailed to open read-only db: %v",
						packageName, err)
				}
				*db = newDB
			},
			ok: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			databaseFile := filepath.Join(t.TempDir(), "test.db")
			db, _ := sql.Open("sqlite", databaseFile)
			t.Cleanup(func() {
				_ = db.Close()
				_ = os.Remove(databaseFile)
			})

			// Enable foreign key constraint enforcement.
			if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
				t.Fatalf("%s\nFailed to enable foreign key constraints: %s",
					packageName, err)
			}

			// Create the 'status' table.
			if err := createStatusTable(db, "status", tc.columnType); err != nil {
				t.Fatalf("%s\nfailed to create status table: %v",
					packageName, err)
			}

			var (
				id                       = "keepMe"
				latestVersion            = "0.0.3"
				latestVersionTimestamp   = "2020-01-02T01:01:01Z"
				deployedVersion          = "0.0.2"
				deployedVersionTimestamp = "2020-01-01T01:01:01Z"
				approvedVersion          = "0.0.1"
			)
			// Add a row to the table.
			if err := insertStatusRow(db,
				id,
				latestVersion, latestVersionTimestamp,
				deployedVersion, deployedVersionTimestamp,
				approvedVersion); err != nil {
				t.Fatalf("%s: failed to insert row: %v",
					packageName, err)
			}

			// Apply test-specific setup.
			if tc.setupDB != nil {
				tc.setupDB(t, &db, databaseFile)
			}

			resultChannel := make(chan bool, 1)
			// WHEN updateTable is called.
			resultChannel <- updateTable(db)

			// THEN it returns ok only if the update succeeded.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), releaseStdout); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			if !tc.ok {
				stdout := releaseStdout()
				if !util.RegexCheck(tc.stdoutRegex, stdout) {
					t.Fatalf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
						packageName, tc.stdoutRegex, stdout)
				}
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

func createStatusTable(db *sql.DB, tableName string, columnType string) error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS ` + tableName + ` (
			id                         TYPE     NOT NULL PRIMARY KEY,
			latest_version             TYPE     DEFAULT '',
			latest_version_timestamp   DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			deployed_version           TYPE     DEFAULT '',
			deployed_version_timestamp DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			approved_version           TYPE     DEFAULT ''
		);`
	sqlStmt = strings.ReplaceAll(sqlStmt, "TYPE", columnType)
	_, err := db.Exec(sqlStmt)

	return err
}

func insertStatusRow(db *sql.DB, id, lv, lvt, dv, dvt, av string) error {
	sqlStmt := `
		INSERT OR REPLACE INTO status (
			id, latest_version, latest_version_timestamp,
			deployed_version, deployed_version_timestamp, approved_version
		) VALUES (?, ?, ?, ?, ?, ?);`
	_, err := db.Exec(sqlStmt, id, lv, lvt, dv, dvt, av)

	return err
}

func createBackupTableWithTrigger(db *sql.DB, tableName string, columnType string) {
}
