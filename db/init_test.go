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
	"database/sql"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestAPI_Get(t *testing.T) {
	// GIVEN: a DB is running (see TestMain).

	// WHEN: a message is send to the DatabaseChannel targeting latest_version.
	target := "keep0"
	cell := dbtype.Cell{Column: "latest_version", Value: "9.9.9"}
	cfg.DatabaseChannel <- dbtype.Message{
		ServiceID: target,
		Cells:     []dbtype.Cell{cell},
	}
	time.Sleep(time.Second)

	// THEN: the cell was changed in the DB.
	otherCfg := testConfig(t)
	otherCfg.Settings.Data.DatabaseFile = "TestAPI_Get-copy.db"
	bytesRead, err := os.ReadFile(cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatalf("%s\n%v", packageName, err)
	}
	t.Cleanup(func() { _ = os.Remove(otherCfg.Settings.Data.DatabaseFile) })
	err = os.WriteFile(otherCfg.Settings.Data.DatabaseFile, bytesRead, os.FileMode(0644))
	if err != nil {
		t.Fatalf("%s\n%v", packageName, err)
	}
	tAPI := api{config: otherCfg}
	t.Cleanup(func() { dbCleanup(&tAPI) })
	tAPI.initialise()
	got := queryRow(t, tAPI.db, target)
	want := status.Status{}
	want.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: target,
		},
		&dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				WebURL: "https://example.com",
			},
		},
	)
	want.SetLatestVersion("9.9.9", "2022-01-01T01:01:01Z", false)
	want.SetDeployedVersion("0.0.0", "2020-01-01T01:01:01Z", false)
	want.SetApprovedVersion("0.0.1", false)
	if got.LatestVersion() != want.LatestVersion() {
		t.Errorf(
			"%s\nwant: %q to be updated to %q\ngot:  %q",
			packageName, cell.Column, cell.Value, got,
		)
	}
}

func TestAPI_Get__fail(t *testing.T) {
	// GIVEN: a DB to set up.
	tests := []struct {
		name         string
		setupDB      func(cfg *config.Config)
		deleteDBFile bool
		ok           bool
		errRegex     string
	}{
		{
			name: "initialise fails",
			setupDB: func(cfg *config.Config) {
				// Directory rather than file.
				cfg.Settings.Data.DatabaseFile = t.TempDir()
				fmt.Println()
			},
			ok:       false,
			errRegex: `^FATAL: db .+ path "[^"]+" .*is a directory, not a file` + "\n$",
		},
		{
			name: "removeUnknownServices fails",
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
						name TEXT PRIMARY KEY,
						latest_version TEXT,
						latest_version_timestamp TEXT,
						deployed_version TEXT,
						deployed_version_timestamp TEXT,
						approved_version TEXT
					);`,
				); err != nil {
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
					  '0.0.1')`,
				); err != nil {
					t.Fatalf(
						"%s\nerror inserting row in setupDB:\n%v",
						packageName, err,
					)
				}
				cfg.Settings.Data.DatabaseFile = tAPI.config.Settings.Data.DatabaseFile
			},
			ok:       false,
			errRegex: `^FATAL: db .*no such column.*` + "\n$",
		},
		{
			name: "extractServiceStatus fails",
			setupDB: func(cfg *config.Config) {
				// Drop required column so extractServiceStatus fails.
				tAPI := testAPI(t)
				t.Cleanup(func() { dbCleanup(tAPI) })
				tAPI.initialise()
				defer tAPI.db.Close()
				if _, err := tAPI.db.Exec("ALTER TABLE status DROP COLUMN deployed_version;"); err != nil {
					t.Fatalf("%s\nerror dropping column in setupDB:\n%v", packageName, err)
				}
				if _, err := tAPI.db.Exec(`
					INSERT INTO status (
						id,
						latest_version,
						latest_version_timestamp,
						deployed_version_timestamp,
						approved_version
					) VALUES (
					  'service1',
					  '0.0.0',
					  '2020-01-01T01:01:01Z',
					  '2020-01-01T01:01:01Z',
					  '0.0.1')`,
				); err != nil {
					t.Fatalf(
						"%s\nerror inserting row in setupDB:\n%v",
						packageName, err,
					)
				}
				cfg.Settings.Data.DatabaseFile = tAPI.config.Settings.Data.DatabaseFile
			},
			ok:       false,
			errRegex: `^FATAL: db .* no such column.*` + "\n$",
		},
		{
			name:     "success",
			setupDB:  func(cfg *config.Config) {},
			ok:       true,
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			tAPI := testAPI(t)
			tAPI.initialise()
			tc.setupDB(tAPI.config)

			go tAPI.Handler(t.Context())

			resultChannel := make(chan bool, 1)
			// WHEN: that DB is fetched.
			api := Get(tAPI.config)
			if api != nil {
				defer api.db.Close()
			}
			resultChannel <- api != nil

			prefix := fmt.Sprintf("%s\napi.removeUnknownServices()", packageName)

			// THEN: the Status in the Config is updated.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.errRegex,
				)
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
	// GIVEN: a config with a database location.
	tests := []struct {
		name  string
		ok    bool
		paths []pathInfo
	}{
		{
			name: "DB rw-r-r",
			ok:   true,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_644, itemType: "file"},
			},
		},
		{
			name: "DB -w--w--w-",
			ok:   false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_444, itemType: "file"},
			},
		},
		{
			name: "DB r--r--r--",
			ok:   false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_222, itemType: "file"},
			},
		},
		{
			name: "DB --x--x--x",
			ok:   false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_111, itemType: "file"},
			},
		},
		{
			name: "Dir not writable",
			ok:   false,
			paths: []pathInfo{
				{path: "test", perms: 0_555, itemType: "dir"},
				{path: "test/argus.db", perms: 0_644, itemType: "file"},
			},
		},
		{
			name: "Dir not readable",
			ok:   false,
			paths: []pathInfo{
				{path: "test", perms: 0_444, itemType: "dir"},
				{path: "test/argus.db", perms: 0_444, itemType: "file"},
			},
		},
		{
			name: "Non-DB file data",
			ok:   false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_644, itemType: "file", data: "something"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
						t.Fatalf(
							"%s\npaths chmod on %q failed (%q)\nerror: %s",
							packageName, tc.paths[i].path, tc.paths[i].perms, err,
						)
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
			// WHEN: the db is initialised with it.
			resultChannel <- tAPI.initialise()

			prefix := fmt.Sprintf("%s\napi.initialise()", packageName)

			// THEN: the function failed if the db was unreadable.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				nil,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			if !tc.ok {
				return
			}
			// THEN: the status table was created in the db.
			rows, err := tAPI.db.Query(`
				SELECT	id,
						latest_version,
						latest_version_timestamp,
						deployed_version,
						deployed_version_timestamp,
						approved_version
				FROM status;`,
			)
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

func TestAPI_Initialise__openError(t *testing.T) {
	releaseStdout := test.CaptureLog(t, logx.Default())

	// GIVEN: a failing database open.
	original := openDatabase
	customErr := fmt.Errorf("open failed")
	openDatabase = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, customErr
	}
	t.Cleanup(func() { openDatabase = original })
	errRegex := `^FATAL: db .*` + customErr.Error()

	// AND: an API.
	tAPI := testAPI(t)

	// WHEN: the DB is initialised.
	resultChannel := make(chan bool, 1)
	resultChannel <- tAPI.initialise()

	prefix := fmt.Sprintf("%s\napi.initialise()", packageName)

	// THEN: initialisation fails with a fatal open error.
	if err := test.AssertChannelBool(
		t,
		false,
		resultChannel,
		logx.ExitCodeChannel(),
		releaseStdout,
	); err != nil {
		t.Fatal(prefix + err.Error())
	}

	// AND: stdout recorded the error.
	stdout := releaseStdout()
	if !util.RegexCheck(errRegex, stdout) {
		t.Errorf(
			"%s stdout mismatch\ngot:  %q\nwant: %q",
			prefix, stdout, errRegex,
		)
	}
}

func TestAPI_RemoveUnknownServices(t *testing.T) {
	// GIVEN: a DB with many service status'.
	tests := []struct {
		databaseDeleted bool
	}{
		{databaseDeleted: false},
		{databaseDeleted: true},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("deleted=%t", tc.databaseDeleted)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tAPI := testAPI(t)
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
				sqlStmt += fmt.Sprintf(
					" (%q, %q, %q, %q, %q, %q),",
					id,
					svc.Status.LatestVersion(),
					svc.Status.LatestVersionTimestamp(),
					svc.Status.DeployedVersion(),
					svc.Status.DeployedVersionTimestamp(),
					svc.Status.ApprovedVersion(),
				)
			}
			if got := len(tAPI.config.Order); got <= 1 {
				t.Fatalf(
					"%s\nexpected more services present in the API.Config.Order, have=%d",
					packageName, got,
				)
			}
			tAPI.config.Order = util.RemoveAt(tAPI.config.Order, 0)
			_, err := tAPI.db.Exec(sqlStmt[:len(sqlStmt)-1] + ";")
			if err != nil {
				t.Fatal(err)
			}
			// Delete the DB file.
			if tc.databaseDeleted {
				_ = os.Remove(tAPI.config.Settings.Data.DatabaseFile)
			}
			wantOk := !tc.databaseDeleted

			resultChannel := make(chan bool, 1)
			// WHEN: the unknown Services are removed with removeUnknownServices.
			resultChannel <- tAPI.removeUnknownServices()

			prefix := fmt.Sprintf("%s\napi.RemoveUnknownServices()", packageName)

			// THEN: the app panicked if the db was deleted.
			if err := test.AssertChannelBool(
				t,
				wantOk,
				resultChannel,
				logx.ExitCodeChannel(),
				nil,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			if !wantOk {
				return
			}

			// AND: the rows of Services not in .All are returned.
			rows, err := tAPI.db.Query(`
				SELECT	id,
						latest_version,
						latest_version_timestamp,
						deployed_version,
						deployed_version_timestamp,
						approved_version
				FROM status;`,
			)
			if err != nil {
				t.Fatal(err)
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
					t.Errorf(
						"%s %q should have been removed from the table",
						prefix, id,
					)
				}
			}
			if got := len(tAPI.config.Order); got != count {
				t.Errorf(
					"%s Config.Order length mismatch\ngot:  %d\nwant: %d",
					prefix, got, count,
				)
			}
		})
	}
}

func TestAPI_RemoveUnknownServices__fail(t *testing.T) {
	// GIVEN: an API with no 'id' column in the 'status' table.
	tests := []struct {
		name            string
		errRegex        string
		tableCreateStmt string
		insertDataStmt  string
	}{
		{
			name:     "id missing",
			errRegex: `^FATAL: db.*no such column: id`,
			tableCreateStmt: `
				CREATE TABLE IF NOT EXISTS status (
					name TEXT PRIMARY KEY,
					latest_version TEXT,
					latest_version_timestamp TEXT,
					deployed_version TEXT,
					deployed_version_timestamp TEXT,
					approved_version TEXT
				);`,
			insertDataStmt: `
				INSERT INTO status (
					name,
					latest_version,
					latest_version_timestamp,
					deployed_version,
					deployed_version_timestamp,
					approved_version
				)
				VALUES (
					'service1',
					'0.0.0',
					'2020-01-01T01:01:01Z',
					'0.0.0',
					'2020-01-01T01:01:01Z',
					'0.0.1');`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			tAPI := testAPI(t)
			tAPI.initialise()
			_, _ = tAPI.db.Exec("DROP TABLE IF EXISTS status;")
			_, _ = tAPI.db.Exec(tc.tableCreateStmt)
			if tc.insertDataStmt != "'" {
				_, _ = tAPI.db.Exec(tc.insertDataStmt)
			}

			go tAPI.Handler(t.Context())

			resultChannel := make(chan bool, 1)
			// WHEN: removeUnknownServices is called.
			resultChannel <- tAPI.removeUnknownServices()

			prefix := fmt.Sprintf("%s\napi.removeUnknownServices()", packageName)

			// THEN: the Status in the Config is updated.
			if err := test.AssertChannelBool(
				t,
				false,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.errRegex,
				)
			}
		})
	}
}

func TestAPI_ExtractServiceStatus(t *testing.T) {
	// GIVEN: an API on a DB containing at least 1 row.
	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())
	wantStatus := make([]status.Status, len(cfg.Service))
	// Push a random Status for each Service to the DB.
	index := 0
	for id, svc := range tAPI.config.Service {
		wantStatus[index].Init(
			0, 0, 0,
			status.ServiceInfo{
				ID: id,
			},
			&dashboard.Options{},
		)

		wantStatus[index].ServiceInfo.ID = id
		wantStatus[index].SetLatestVersion(
			fmt.Sprintf(
				"%d.%d.%d",
				rand.Intn(10), rand.Intn(10), rand.Intn(10),
			),
			time.Now().UTC().Format(time.RFC3339),
			false,
		)
		wantStatus[index].SetDeployedVersion(
			fmt.Sprintf(
				"%d.%d.%d",
				rand.Intn(10), rand.Intn(10), rand.Intn(10),
			),
			"",
			false,
		)
		wantStatus[index].SetApprovedVersion(
			fmt.Sprintf(
				"%d.%d.%d",
				rand.Intn(10), rand.Intn(10), rand.Intn(10),
			),
			false,
		)

		tAPI.config.DatabaseChannel <- dbtype.Message{
			ServiceID: id,
			Cells: []dbtype.Cell{
				{Column: "id", Value: id},
				{Column: "latest_version", Value: wantStatus[index].LatestVersion()},
				{Column: "latest_version_timestamp", Value: wantStatus[index].LatestVersionTimestamp()},
				{Column: "deployed_version", Value: wantStatus[index].DeployedVersion()},
				{Column: "deployed_version_timestamp", Value: wantStatus[index].DeployedVersionTimestamp()},
				{Column: "approved_version", Value: wantStatus[index].ApprovedVersion()},
			},
		}

		// Clear the Status in the Config.
		svc.Status = *status.New(
			svc.Status.AnnounceChannel, svc.Status.DatabaseChannel, svc.Status.SaveChannel,
			"",
			"", "",
			"", "",
			"",
			nil,
		)
		svc.Status.Init(
			0, 0, 0,
			status.ServiceInfo{
				ID: svc.ID,
			},
			&svc.Dashboard,
		)

		index++
	}
	time.Sleep(250 * time.Millisecond)

	// WHEN: extractServiceStatus is called.
	tAPI.extractServiceStatus()

	// THEN: the Status in the Config is updated.
	for i := range wantStatus {
		row := queryRow(t, tAPI.db, wantStatus[i].ServiceInfo.ID)
		w := &wantStatus[i]
		prefix := fmt.Sprintf(
			"%s\napi.extractServiceStatus() for service %q",
			packageName, w.ServiceInfo.ID,
		)

		fieldTests := []test.FieldAssertion{
			{Name: "LatestVersion", Got: row.LatestVersion(), Want: w.LatestVersion(), Mode: test.CompareEqual},
			{Name: "LatestVersionTimestamp", Got: row.LatestVersionTimestamp(), Want: w.LatestVersionTimestamp(), Mode: test.CompareEqual},
			{Name: "DeployedVersion", Got: row.DeployedVersion(), Want: w.DeployedVersion(), Mode: test.CompareEqual},
			{Name: "DeployedVersionTimestamp", Got: row.DeployedVersionTimestamp(), Want: w.DeployedVersionTimestamp(), Mode: test.CompareEqual},
			{Name: "ApprovedVersion", Got: row.ApprovedVersion(), Want: w.ApprovedVersion(), Mode: test.CompareEqual},
		}
		if err := test.AssertFields(t, fieldTests, prefix, "DB"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAPI_ExtractServiceStatus__fail(t *testing.T) {
	// GIVEN: an API with different 'status' columns.
	tests := []struct {
		name            string
		errRegex        string
		tableCreateStmt string
		insertDataStmt  string
	}{
		{
			name:     "latest_version missing",
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
		{
			name:     "latest_version unexpected type",
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			tAPI := testAPI(t)
			tAPI.initialise()
			_, _ = tAPI.db.Exec("DROP TABLE IF EXISTS status;")
			_, _ = tAPI.db.Exec(tc.tableCreateStmt)
			if tc.insertDataStmt != "'" {
				_, _ = tAPI.db.Exec(tc.insertDataStmt)
			}

			go tAPI.Handler(t.Context())

			resultChannel := make(chan bool, 1)
			// WHEN: extractServiceStatus is called.
			resultChannel <- tAPI.extractServiceStatus()

			prefix := fmt.Sprintf("%s\napi.extractServiceStatus()", packageName)

			// THEN: the Status in the Config is updated.
			if err := test.AssertChannelBool(
				t,
				false,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			stdout := releaseStdout()
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.errRegex,
				)
			}
		})
	}
}

func TestAPI_ExtractServiceStatus__rowsErr(t *testing.T) {
	// GIVEN: a failing rows iteration check.
	original := serviceStatusRowsErr
	customErr := fmt.Errorf("rows iteration failed")
	serviceStatusRowsErr = func(rows *sql.Rows) error {
		return customErr
	}
	t.Cleanup(func() { serviceStatusRowsErr = original })
	errRegex := `^FATAL: db .*` + customErr.Error()

	// AND: an api.
	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())

	releaseStdout := test.CaptureLog(t, logx.Default())

	// WHEN: extractServiceStatus is called.
	resultChannel := make(chan bool, 1)
	resultChannel <- tAPI.extractServiceStatus()

	prefix := fmt.Sprintf("%s\napi.extractServiceStatus()", packageName)

	// THEN: initialisation fails with a fatal rows error.
	if err := test.AssertChannelBool(
		t,
		false,
		resultChannel,
		logx.ExitCodeChannel(),
		releaseStdout,
	); err != nil {
		t.Fatal(prefix + err.Error())
	}
	stdout := releaseStdout()
	if !util.RegexCheck(errRegex, stdout) {
		t.Errorf(
			"%s stdout mismatch\ngot:  %q\nwant: %q",
			prefix, stdout, errRegex,
		)
	}
}

func TestDBQueryService(t *testing.T) {
	// GIVEN: a blank DB.
	tAPI := testAPI(t)
	tAPI.initialise()

	// AND: a Service from a Config.
	var serviceName string
	for k := range tAPI.config.Service {
		serviceName = k
		break
	}
	svc := tAPI.config.Service[serviceName]
	cells := []dbtype.Cell{
		{Column: "id", Value: serviceName},
		{Column: "latest_version", Value: svc.Status.LatestVersion()},
		{Column: "latest_version_timestamp", Value: svc.Status.LatestVersionTimestamp()},
		{Column: "deployed_version", Value: svc.Status.DeployedVersion()},
		{Column: "deployed_version_timestamp", Value: svc.Status.DeployedVersionTimestamp()},
		{Column: "approved_version", Value: svc.Status.ApprovedVersion()},
	}

	// WHEN: the Service data is written to the DB.
	tAPI.updateRow(serviceName, cells)

	prefix := fmt.Sprintf(
		"%s\napi.updateRow(id=%q, cells=%+v)",
		packageName, serviceName, cells,
	)

	// THEN: that data can be queried and retrieved.
	got := queryRow(t, tAPI.db, serviceName)
	fieldTests := []test.FieldAssertion{
		{Name: "LatestVersion", Got: got.LatestVersion(), Want: svc.Status.LatestVersion(), Mode: test.CompareEqual},
		{Name: "LatestVersionTimestamp", Got: got.LatestVersionTimestamp(), Want: svc.Status.LatestVersionTimestamp(), Mode: test.CompareEqual},
		{Name: "DeployedVersion", Got: got.DeployedVersion(), Want: svc.Status.DeployedVersion(), Mode: test.CompareEqual},
		{Name: "DeployedVersionTimestamp", Got: got.DeployedVersionTimestamp(), Want: svc.Status.DeployedVersionTimestamp(), Mode: test.CompareEqual},
		{Name: "ApprovedVersion", Got: got.ApprovedVersion(), Want: svc.Status.ApprovedVersion(), Mode: test.CompareEqual},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "DB"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckFile(t *testing.T) {
	type createOptions struct {
		path  string
		perms fs.FileMode
	}
	// GIVEN: various paths.
	tests := []struct {
		name             string
		createDirBefore  []createOptions
		dirPerms         []createOptions
		createFileBefore string
		path             string
		stdoutRegex      string
	}{
		{
			name:        "file doesn't exist",
			path:        "something_does_not_exist.db",
			stdoutRegex: `^$`,
		},
		{
			name:        "dir doesn't exist, so is created",
			path:        "dir_does_not_exist_1/argus.db",
			stdoutRegex: `^$`,
		},
		{
			name: "dir exists but not file",
			path: "dir_does_exist_2/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_does_exist_2", perms: 0_750},
			},
			stdoutRegex: `^$`,
		},
		{
			name: "file is dir",
			path: "folder.db",
			createDirBefore: []createOptions{
				{path: "folder.db", perms: 0_750},
			},
			stdoutRegex: `path .* is a directory, not a file`,
		},
		{
			name:             "dir is file",
			path:             "item_not_a_dir/argus.db",
			createFileBefore: "item_not_a_dir",
			stdoutRegex:      `path .* is not a directory`,
		},
		{
			name: "no perms to create dir",
			path: "dir_no_perms_1/dir_no_perms_2/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_perms_1", perms: 0_555},
			},
			stdoutRegex: test.TrimYAML(`
				mkdir .*
					permission denied`,
			),
		},
		{
			name: "no perms to check for file in dir",
			path: "dir_no_perms_3/dir_no_perms_4/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_perms_3", perms: 0_444},
				{path: "dir_no_perms_3/dir_no_perms_4", perms: 0_444},
			},
			stdoutRegex: test.TrimYAML(`
				stat .*
					permission denied`,
			),
		},
		{
			name: "cannot stat existing file due to dir perms",
			path: "dir_no_exec/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_no_exec", perms: 0_666},
			},
			dirPerms: []createOptions{
				{path: "dir_no_exec", perms: 0_666},
			},
			createFileBefore: "dir_no_exec/argus.db",
			stdoutRegex: test.TrimYAML(`
				stat .*
					permission denied`,
			),
		},
		{
			name:             "file does exist",
			path:             "file_does_exist.db",
			createFileBefore: "file_does_exist.db",
			stdoutRegex:      `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

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
					perm := os.FileMode(0_750)
					if err := os.Mkdir(dir.path, perm); err != nil {
						t.Fatalf(
							"%s\ncreateDirBefore mkdir on %q-%q\nerror: %s",
							packageName, dir.path, perm, err,
						)
					}
				}
			}
			if tc.createFileBefore != "" {
				file, err := os.Create(tc.createFileBefore)
				if err != nil {
					t.Fatalf(
						"%s\ncreateFileBefore on %q\nerror: %s",
						packageName, tc.createFileBefore, err,
					)
				}
				_ = file.Close()
				t.Cleanup(func() { _ = os.Remove(tc.createFileBefore) })
			}
			// Set perms (in reverse order).
			for i := len(tc.createDirBefore) - 1; i >= 0; i-- {
				if err := os.Chmod(tc.createDirBefore[i].path, tc.createDirBefore[i].perms); err != nil {
					t.Fatalf(
						"%s\ncreateDirBefore chmod on %q-%q\nerror: %s",
						packageName, tc.createDirBefore[i].path, tc.createDirBefore[i].perms, err,
					)
				}
			}

			resultChannel := make(chan bool, 1)
			// WHEN: checkFile is called on that same dir.
			resultChannel <- checkFile(tc.path)

			prefix := fmt.Sprintf(
				"%s\ncheckFile(%q)",
				packageName, tc.path,
			)

			// THEN: if false is returned, the error is logged.
			wantOk := tc.stdoutRegex == `^$`
			if err := test.AssertChannelBool(
				t,
				wantOk,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}

			// AND: the stdout matches the expected result.
			if stdout := releaseStdout(); !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}

func TestUpdateTypes(t *testing.T) {
	// GIVEN: a DB with the *_version columns as STRING/TEXT.
	tests := []struct {
		name        string
		columnType  string
		setupDB     func(t *testing.T, db **sql.DB, dbFile string)
		ok          bool
		stdoutRegex string
	}{
		{
			name:       "No conversion necessary",
			columnType: "TEXT",
			ok:         true,
		},
		{
			name:       "Conversion wanted",
			columnType: "STRING",
			ok:         true,
		},
		{
			name:        "DB is deleted",
			columnType:  "STRING",
			stdoutRegex: `no rows in result set`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				_, _ = (*db).Exec("DROP TABLE IF EXISTS status;")
			},
			ok: false,
		},
		{
			name:        "Backup exists with different fields",
			columnType:  "STRING",
			stdoutRegex: `copy.*table .* has \d columns but \d values were supplied`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				_, _ = (*db).Exec(`
					CREATE TABLE IF NOT EXISTS status_backup (
						id TEXT NOT NULL PRIMARY KEY
					);`,
				)
			},
			ok: false,
		},
		{
			name:        "Cannot drop table because of foreign key",
			columnType:  "STRING",
			stdoutRegex: `drop.*FOREIGN KEY constraint failed`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				fkTable := "status_backup_fk"
				_, _ = (*db).Exec(`
					CREATE TABLE IF NOT EXISTS ` + fkTable + ` (
						id INTEGER NOT NULL PRIMARY KEY,
						fk_id TEXT NOT NULL,
						FOREIGN KEY(fk_id) REFERENCES status(id)
					);`,
				)
				_, _ = (*db).Exec(`INSERT OR REPLACE INTO ` + fkTable + ` (fk_id) VALUES ('keepMe');`)
			},
			ok: false,
		},
		{
			name:        "Cannot alter backup table because of foreign key",
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
					END;`,
				)
			},
			ok: false,
		},
		{
			name:        "latest_version column doesn't exist",
			stdoutRegex: `no rows in result set`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				if _, err := (*db).Exec(`
					ALTER TABLE status
					DROP COLUMN latest_version;
				`); err != nil {
					t.Fatalf(
						"%s\nfailed to drop column: %v",
						packageName, err,
					)
				}
			},
			ok: false,
		},
		{
			name:        "Read-only DB",
			stdoutRegex: `create.*attempt to write a readonly database`,
			setupDB: func(t *testing.T, db **sql.DB, dbFile string) {
				_ = (*db).Close()
				newDB, err := sql.Open("sqlite", "file:"+dbFile+"?mode=ro")
				if err != nil {
					t.Fatalf(
						"%s\nfailed to open read-only db: %v",
						packageName, err,
					)
				}
				*db = newDB
			},
			ok: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			databaseFile := filepath.Join(t.TempDir(), "test.db")
			db, _ := sql.Open("sqlite", databaseFile)
			t.Cleanup(func() {
				_ = db.Close()
				_ = os.Remove(databaseFile)
			})

			// Enable foreign key constraint enforcement.
			if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
				t.Fatalf(
					"%s\nFailed to enable foreign key constraints: %s",
					packageName, err,
				)
			}

			// Create the 'status' table.
			if err := createStatusTable(db, "status", tc.columnType); err != nil {
				t.Fatalf(
					"%s\nfailed to create status table: %v",
					packageName, err,
				)
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
			if err := insertStatusRow(
				db,
				id,
				latestVersion, latestVersionTimestamp,
				deployedVersion, deployedVersionTimestamp,
				approvedVersion,
			); err != nil {
				t.Fatalf("%s: failed to insert row: %v", packageName, err)
			}

			// Apply test-specific setup.
			if tc.setupDB != nil {
				tc.setupDB(t, &db, databaseFile)
			}

			resultChannel := make(chan bool, 1)
			// WHEN: updateTable is called.
			resultChannel <- updateTable(db)

			prefix := fmt.Sprintf("%s\nupdateTable()", packageName)

			// THEN: it returns ok only if the update succeeded.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			if !tc.ok {
				stdout := releaseStdout()
				if !util.RegexCheck(tc.stdoutRegex, stdout) {
					t.Fatalf(
						"%s stdout mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, tc.stdoutRegex,
					)
				}
				return
			}

			// AND: the ID column and all *_version columns are now TEXT.
			wantTextColumns := []string{"id", "latest_version", "deployed_version", "approved_version"}
			for _, row := range wantTextColumns {
				var columnType string
				db.QueryRow(`
					SELECT type
					FROM pragma_table_info('status')
					WHERE name = 'latest_version'`,
				).Scan(&columnType)
				if columnType != "TEXT" {
					t.Errorf(
						"%s column type mismatch on %q\ngot:  %q\nwant: TEXT",
						prefix, row,
						columnType,
					)
				}
			}

			// AND: all rows were carried over.
			got := queryRow(t, db, id)
			if got.LatestVersion() != latestVersion || got.LatestVersionTimestamp() != latestVersionTimestamp ||
				got.DeployedVersion() != deployedVersion || got.DeployedVersionTimestamp() != deployedVersionTimestamp ||
				got.ApprovedVersion() != approvedVersion {
				t.Errorf(
					"%s Row wasn't carried over correctly:\n"+
						"Got: lv=%q, lvt=%q, dv=%q, dvt=%q, av=%q\n"+
						"Had: lv=%q, lvt=%q, dv=%q, dvt=%q, av=%q",
					prefix,
					got.LatestVersion(), got.LatestVersionTimestamp(),
					got.DeployedVersion(), got.DeployedVersionTimestamp(),
					got.ApprovedVersion(),
					latestVersion, latestVersionTimestamp,
					deployedVersion, deployedVersionTimestamp,
					approvedVersion,
				)
			}

			// AND: the conversion was printed to stdout.
			stdout := releaseStdout()
			want := "Finished updating column types"
			contains := strings.Contains(stdout, want)
			if tc.columnType == "TEXT" && contains {
				t.Errorf(
					"%s table started as %q:\ngot:  %q\n want: no change",
					prefix, tc.columnType,
					stdout,
				)
			} else if tc.columnType == "STRING" && !contains {
				t.Errorf(
					"%s table started as %q\ngot:  %q\n want: contains %q",
					prefix, tc.columnType,
					stdout, want,
				)
			}
			time.Sleep(100 * time.Millisecond)
		})
	}
}

// ################
// # TEST HELPERS #
// ################

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
