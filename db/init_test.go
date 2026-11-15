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
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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
	// Uses the DB that TestMain runs.
	target := "keep0"
	cell := dbtype.Cell{Column: "latest_version", Value: "9.9.9"}
	cfg.DatabaseChannel <- dbtype.Message{
		ServiceID: target,
		Cells:     []dbtype.Cell{cell},
	}
	time.Sleep(time.Second)

	otherCfg := testConfig(t)
	otherCfg.Settings.Data.DatabaseFile = t.Name() + "-copy.db"
	bytesRead, err := os.ReadFile(cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(otherCfg.Settings.Data.DatabaseFile) })
	err = os.WriteFile(otherCfg.Settings.Data.DatabaseFile, bytesRead, os.FileMode(0644))
	if err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}
	tAPI := api{config: otherCfg}
	t.Cleanup(func() { dbCleanup(&tAPI) })
	tAPI.initialise()
	got := queryRow(t, tAPI.db, target)
	if got.LatestVersion() != cell.Value {
		t.Errorf(
			"Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
			got.LatestVersion(), cell.Value,
		)
	}
}

func TestAPI_Get__fail(t *testing.T) {
	tests := []struct {
		name     string
		setupDB  func(t *testing.T, cfg *config.Config)
		ok       bool
		errRegex string
	}{
		{
			name: "initialise fails",
			setupDB: func(t *testing.T, cfg *config.Config) {
				// Directory rather than file.
				cfg.Settings.Data.DatabaseFile = t.TempDir()
			},
			ok:       false,
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db .+ path "[^"]+" .*is a directory, not a file` + "\n$",
		},
		{
			name: "removeUnknownServices fails",
			setupDB: func(t *testing.T, cfg *config.Config) {
				tAPI := testAPI(t)
				tAPI.initialise()
				defer tAPI.db.Close()
				// Recreate status without an id column so api.removeUnknownServices fails.
				test.DropTable(t, tAPI.db, "status")
				test.ExecSQL(
					t, tAPI.db,
					`
					CREATE TABLE IF NOT EXISTS status (
						name TEXT PRIMARY KEY,
						latest_version TEXT,
						latest_version_timestamp TEXT,
						deployed_version TEXT,
						deployed_version_timestamp TEXT,
						approved_version TEXT
					);`,
					`
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
				)
				cfg.Settings.Data.DatabaseFile = tAPI.config.Settings.Data.DatabaseFile
			},
			ok:       false,
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db .*no such column.*` + "\n$",
		},
		{
			name: "extractServiceStatus fails",
			setupDB: func(t *testing.T, cfg *config.Config) {
				tAPI := testAPI(t)
				tAPI.initialise()
				defer tAPI.db.Close()
				// Drop a required column so api.extractServiceStatus fails.
				if _, err := tAPI.db.Exec("ALTER TABLE status DROP COLUMN deployed_version;"); err != nil {
					t.Fatalf("drop column: DB.Exec() failed: %v", err)
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
					t.Fatalf("insert: DB.Exec() failed: %v", err)
				}
				cfg.Settings.Data.DatabaseFile = tAPI.config.Settings.Data.DatabaseFile
			},
			ok:       false,
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db .* no such column.*` + "\n$",
		},
		{
			name:     "success",
			setupDB:  func(t *testing.T, cfg *config.Config) {},
			ok:       true,
			errRegex: `^(?:INFO: [^\n]*\n)*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			tAPI := testAPI(t)
			tAPI.initialise()
			tc.setupDB(t, tAPI.config)

			go tAPI.Handler(t.Context())

			test.AssertCallLog(
				t,
				tc.ok,
				func() bool {
					gotAPI := Get(tAPI.config)
					if gotAPI == nil {
						return false
					}
					t.Cleanup(func() { _ = gotAPI.db.Close() })
					return true
				},
				logx.ExitCodeChannel(),
				releaseStdout,
				tc.errRegex,
			)
		})
	}
}

func TestAPI_dbHandles(t *testing.T) {
	tAPI := testAPI(t)
	tAPI.initialise()

	got := tAPI.DB()
	if got != tAPI.db {
		t.Errorf("open: DB mismatch\ngot:  %p\nwant: %p", got, tAPI.db)
	}

	tAPI.CloseDB()
	if err := tAPI.db.Ping(); err == nil {
		t.Error("after CloseDB: DB.Ping() error mismatch\ngot:  nil\nwant: non-nil")
	}

	unopened := &api{}
	unopened.CloseDB()
	gotUnopened := unopened.DB()
	if gotUnopened != nil {
		t.Errorf("unopened: DB mismatch\ngot:  %p\nwant: nil", gotUnopened)
	}
}

// closeFailConn is a connection that cannot be closed.
type closeFailConn struct {
	driver.Conn
}

func (closeFailConn) Close() error { return errors.New("close failed") }

// closeFailConnector hands out connections that cannot be closed.
type closeFailConnector struct {
	inner driver.Driver
}

func (c closeFailConnector) Connect(context.Context) (driver.Conn, error) {
	conn, err := c.inner.Open(":memory:")
	if err != nil {
		return nil, err
	}
	return closeFailConn{Conn: conn}, nil
}

func (c closeFailConnector) Driver() driver.Driver { return c.inner }

func TestAPI_CloseDB__error(t *testing.T) {
	releaseStdout := test.CaptureLog(t, logx.Default())

	probe, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() failed: %v", err)
	}
	tAPI := &api{db: sql.OpenDB(closeFailConnector{inner: probe.Driver()})}
	_ = probe.Close()

	if err := tAPI.db.Ping(); err != nil {
		t.Fatalf("DB.Ping() failed: %v", err)
	}

	tAPI.CloseDB()

	stdout := releaseStdout()
	wantRegex := `close failed`
	if !util.RegexCheck(wantRegex, stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, wantRegex)
	}
}

func TestAPI_Initialise(t *testing.T) {
	type pathInfo struct {
		path     string
		perms    os.FileMode
		itemType string
		data     string
	}

	tests := []struct {
		name  string
		ok    bool
		paths []pathInfo
	}{
		{
			name: "DB/rw-r--r--",
			ok:   true,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_644, itemType: "file"},
			},
		},
		{
			name: "DB/r--r--r--",
			ok:   false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_444, itemType: "file"},
			},
		},
		{
			name: "DB/-w--w--w-",
			ok:   false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_222, itemType: "file"},
			},
		},
		{
			name: "DB/--x--x--x",
			ok:   false,
			paths: []pathInfo{
				{path: "argus.db", perms: 0_111, itemType: "file"},
			},
		},
		{
			name: "Dir/not writable",
			ok:   false,
			paths: []pathInfo{
				{path: "test", perms: 0_555, itemType: "dir"},
				{path: "test/argus.db", perms: 0_644, itemType: "file"},
			},
		},
		{
			name: "Dir/not readable",
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
					if err := os.WriteFile(dbFile, []byte(p.data), 0_644); err != nil {
						t.Fatalf("os.WriteFile(%q) failed: %v", dbFile, err)
					}
				} else if err := os.Mkdir(tc.paths[i].path, 0_755); err != nil {
					t.Fatalf("os.Mkdir(%q) failed: %v", tc.paths[i].path, err)
				}
			}
			for _, v := range slices.Backward(tc.paths) {
				if err := os.Chmod(v.path, v.perms); err != nil {
					t.Fatalf("os.Chmod(%q, %#o) failed: %v", v.path, v.perms, err)
				}
			}
			tAPI.config.Settings.Data.DatabaseFile = dbFile
			t.Cleanup(func() {
				for _, p := range tc.paths {
					_ = os.Chmod(p.path, 0_755)
				}
				dbCleanup(tAPI)
			})

			test.AssertCallLog(t, tc.ok, tAPI.initialise, logx.ExitCodeChannel(), nil, "")
			if !tc.ok {
				return
			}

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
				t.Fatalf("DB.Query() failed: %v", err)
			}
			_ = rows.Close()
		})
	}
}

func TestAPI_Initialise__chmodError(t *testing.T) {
	releaseStdout := test.CaptureLog(t, logx.Default())

	original := chmodFile
	customErr := errors.New("chmod failed")
	chmodFile = func(_ string, _ os.FileMode) error {
		return customErr
	}
	t.Cleanup(func() { chmodFile = original })
	errRegex := `^ERROR: db .*` + customErr.Error()

	tAPI := testAPI(t)

	// A failed chmod is logged, not fatal.
	test.AssertCallLog(t, true, tAPI.initialise, logx.ExitCodeChannel(), releaseStdout, errRegex)
}

func TestAPI_Initialise__openError(t *testing.T) {
	releaseStdout := test.CaptureLog(t, logx.Default())

	original := openDatabase
	customErr := errors.New("open failed")
	openDatabase = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, customErr
	}
	t.Cleanup(func() { openDatabase = original })
	errRegex := `^FATAL: db .*` + customErr.Error()

	tAPI := testAPI(t)

	test.AssertCallLog(t, false, tAPI.initialise, logx.ExitCodeChannel(), releaseStdout, errRegex)
}

func TestAPI_Initialise__pingError(t *testing.T) {
	releaseStdout := test.CaptureLog(t, logx.Default())

	original := openDatabase
	openDatabase = func(_, _ string) (*sql.DB, error) {
		// sql.Open is lazy, so this only fails once Ping forces a connection.
		return sql.Open("sqlite", filepath.Join(t.TempDir(), "absent-dir", "x.db"))
	}
	t.Cleanup(func() { openDatabase = original })
	errRegex := `^FATAL: db .* unable to open database file`

	tAPI := testAPI(t)

	test.AssertCallLog(t, false, tAPI.initialise, logx.ExitCodeChannel(), releaseStdout, errRegex)
}

func TestAPI_RemoveUnknownServices(t *testing.T) {
	tests := []struct {
		name            string
		databaseDeleted bool
	}{
		{name: "DB file present", databaseDeleted: false},
		{name: "DB file deleted", databaseDeleted: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			values := make([]string, 0, len(tAPI.config.Service))
			for id, svc := range tAPI.config.Service {
				values = append(values, fmt.Sprintf(
					" (%q, %q, %q, %q, %q, %q)",
					id,
					svc.Status.LatestVersion(),
					svc.Status.LatestVersionTimestamp(),
					svc.Status.DeployedVersion(),
					svc.Status.DeployedVersionTimestamp(),
					svc.Status.ApprovedVersion(),
				))
			}
			if got := len(tAPI.config.Order); got <= 1 {
				t.Fatalf("len(Config.Order) mismatch\ngot:  %d\nwant: > 1", got)
			}
			tAPI.config.Order = util.RemoveAt(tAPI.config.Order, 0)
			_, err := tAPI.db.Exec(sqlStmt + strings.Join(values, ",") + ";")
			if err != nil {
				t.Fatalf("DB.Exec() failed: %v", err)
			}
			if tc.databaseDeleted {
				if err := os.Remove(tAPI.config.Settings.Data.DatabaseFile); err != nil {
					t.Fatalf("os.Remove() failed: %v", err)
				}
			}
			wantOk := !tc.databaseDeleted

			test.AssertCallLog(t, wantOk, tAPI.removeUnknownServices, logx.ExitCodeChannel(), nil, "")
			if !wantOk {
				return
			}

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
				t.Fatalf("DB.Query() failed: %v", err)
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
				if err := rows.Scan(&id, &lv, &lvt, &dv, &dvt, &av); err != nil {
					t.Fatalf("Rows.Scan() failed: %v", err)
				}
				inService := tAPI.config.Service[id] != nil
				inOrder := slices.Contains(tAPI.config.Order, id)
				if !inService || !inOrder {
					t.Errorf(
						"status row %q: Config.Service/Config.Order mismatch\n"+
							"got:  Config.Service=%s, Config.Order=%s\n"+
							"want: Config.Service=present, Config.Order=present",
						id, test.Presence(inService), test.Presence(inOrder),
					)
				}
			}
			if got, want := count, len(tAPI.config.Order); got != want {
				t.Errorf("status row count mismatch\ngot:  %d\nwant: %d", got, want)
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("Rows.Err() failed: %v", err)
			}
		})
	}
}

func TestAPI_RemoveUnknownServices__fail(t *testing.T) {
	tests := []struct {
		name       string
		errRegex   string
		statements []string // Run after the 'status' table is dropped.
	}{
		{
			name:     "id missing",
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db.*no such column: id`,
			statements: []string{
				`
				CREATE TABLE IF NOT EXISTS status (
					name TEXT PRIMARY KEY,
					latest_version TEXT,
					latest_version_timestamp TEXT,
					deployed_version TEXT,
					deployed_version_timestamp TEXT,
					approved_version TEXT
				);`,
				`
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
		},
		{
			name:     "status table missing",
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db.*no such table: status`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			assertFailsOnSchema(t, (*api).removeUnknownServices, tc.statements, tc.errRegex)
		})
	}
}

func TestAPI_ExtractServiceStatus(t *testing.T) {
	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())
	wantStatus := make([]status.Status, len(tAPI.config.Service))
	index := 0
	for id, svc := range tAPI.config.Service {
		wantStatus[index].Init(
			0, 0, 0,
			status.ServiceInfo{
				ID: id,
			},
			&dashboard.Options{},
		)
		wantStatus[index].SetLatestVersion(
			fmt.Sprintf("3.%d.0", index),
			time.Now().UTC().Format(time.RFC3339),
			false,
		)
		wantStatus[index].SetDeployedVersion(
			fmt.Sprintf("2.%d.0", index),
			time.Now().UTC().Format(time.RFC3339),
			false,
		)
		wantStatus[index].SetApprovedVersion(
			fmt.Sprintf("1.%d.0", index),
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

	ok := tAPI.extractServiceStatus()

	if !ok {
		t.Fatal("ok mismatch\ngot:  false\nwant: true")
	}

	for i := range wantStatus {
		w := &wantStatus[i]
		t.Run(w.ServiceInfo.ID, func(t *testing.T) {
			got := &tAPI.config.Service[w.ServiceInfo.ID].Status

			fieldTests := []test.FieldAssertion{
				{Name: "LatestVersion()", Got: got.LatestVersion(), Want: w.LatestVersion(), Mode: test.CompareEqual},
				{Name: "LatestVersionTimestamp()", Got: got.LatestVersionTimestamp(), Want: w.LatestVersionTimestamp(), Mode: test.CompareEqual},
				{Name: "DeployedVersion()", Got: got.DeployedVersion(), Want: w.DeployedVersion(), Mode: test.CompareEqual},
				{Name: "DeployedVersionTimestamp()", Got: got.DeployedVersionTimestamp(), Want: w.DeployedVersionTimestamp(), Mode: test.CompareEqual},
				{Name: "ApprovedVersion()", Got: got.ApprovedVersion(), Want: w.ApprovedVersion(), Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Service.Status"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestAPI_ExtractServiceStatus__fail(t *testing.T) {
	tests := []struct {
		name       string
		errRegex   string
		statements []string // Run after the 'status' table is dropped.
	}{
		{
			name:     "latest_version missing",
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db.*no such column: latest_version`,
			statements: []string{
				`
				CREATE TABLE IF NOT EXISTS status (
					id TEXT PRIMARY KEY,
					latest_version_timestamp TEXT,
					deployed_version TEXT,
					deployed_version_timestamp TEXT,
					approved_version TEXT
				);`,
			},
		},
		{
			name:     "latest_version unexpected type",
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db.*converting NULL to string is unsupported`,
			statements: []string{
				`
				CREATE TABLE IF NOT EXISTS status (
					id TEXT PRIMARY KEY,
					latest_version TEXT,
					latest_version_timestamp TEXT,
					deployed_version TEXT,
					deployed_version_timestamp TEXT,
					approved_version TEXT
				);`,
				`
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
		},
		{
			name:     "status table missing",
			errRegex: `^(?:INFO: [^\n]*\n)*FATAL: db.*no such table: status`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			assertFailsOnSchema(t, (*api).extractServiceStatus, tc.statements, tc.errRegex)
		})
	}
}

func TestAPI_ExtractServiceStatus__rowsErr(t *testing.T) {
	original := serviceStatusRowsErr
	customErr := errors.New("rows iteration failed")
	serviceStatusRowsErr = func(rows *sql.Rows) error {
		return customErr
	}
	t.Cleanup(func() { serviceStatusRowsErr = original })
	errRegex := `^FATAL: db .*` + customErr.Error()

	tAPI := testAPI(t)
	tAPI.initialise()
	go tAPI.Handler(t.Context())

	releaseStdout := test.CaptureLog(t, logx.Default())

	test.AssertCallLog(
		t,
		false,
		tAPI.extractServiceStatus,
		logx.ExitCodeChannel(),
		releaseStdout,
		errRegex,
	)
}

func TestDBQueryService(t *testing.T) {
	tAPI := testAPI(t)
	tAPI.initialise()

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

	tAPI.updateRow(serviceName, cells)

	got := queryRow(t, tAPI.db, serviceName)
	fieldTests := []test.FieldAssertion{
		{Name: "LatestVersion()", Got: got.LatestVersion(), Want: svc.Status.LatestVersion(), Mode: test.CompareEqual},
		{Name: "LatestVersionTimestamp()", Got: got.LatestVersionTimestamp(), Want: svc.Status.LatestVersionTimestamp(), Mode: test.CompareEqual},
		{Name: "DeployedVersion()", Got: got.DeployedVersion(), Want: svc.Status.DeployedVersion(), Mode: test.CompareEqual},
		{Name: "DeployedVersionTimestamp()", Got: got.DeployedVersionTimestamp(), Want: svc.Status.DeployedVersionTimestamp(), Mode: test.CompareEqual},
		{Name: "ApprovedVersion()", Got: got.ApprovedVersion(), Want: svc.Status.ApprovedVersion(), Mode: test.CompareEqual},
	}
	if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
		t.Fatal(testErr)
	}
}

func TestCheckFile(t *testing.T) {
	type createOptions struct {
		path  string
		perms fs.FileMode
	}

	tests := []struct {
		name             string
		createDirBefore  []createOptions
		createFileBefore string
		path             string
		ok               bool
		stdoutRegex      string
	}{
		{
			name:        "file doesn't exist",
			path:        "something_does_not_exist.db",
			ok:          true,
			stdoutRegex: `^$`,
		},
		{
			name:        "dir doesn't exist, so is created",
			path:        "dir_does_not_exist_1/argus.db",
			ok:          true,
			stdoutRegex: `^$`,
		},
		{
			name: "dir exists but not file",
			path: "dir_does_exist_2/argus.db",
			createDirBefore: []createOptions{
				{path: "dir_does_exist_2", perms: 0_750},
			},
			ok:          true,
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
			ok:               true,
			stdoutRegex:      `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			baseDir := t.TempDir()
			for i, dir := range tc.createDirBefore {
				tc.createDirBefore[i].path = filepath.Join(baseDir, dir.path)
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
						t.Fatalf("os.Mkdir(%q, %#o) failed: %v", dir.path, perm, err)
					}
				}
			}
			if tc.createFileBefore != "" {
				file, err := os.Create(tc.createFileBefore)
				if err != nil {
					t.Fatalf("os.Create(%q) failed: %v", tc.createFileBefore, err)
				}
				_ = file.Close()
				t.Cleanup(func() { _ = os.Remove(tc.createFileBefore) })
			}
			// Innermost first, so a parent's perms cannot block its child's chmod.
			for _, v := range slices.Backward(tc.createDirBefore) {
				if err := os.Chmod(v.path, v.perms); err != nil {
					t.Fatalf("os.Chmod(%q, %#o) failed: %v", v.path, v.perms, err)
				}
			}

			test.AssertCallLog(
				t,
				tc.ok,
				func() bool { return checkFile(tc.path) },
				logx.ExitCodeChannel(),
				releaseStdout,
				tc.stdoutRegex,
			)
		})
	}
}

func TestUpdateTypes(t *testing.T) {
	tests := []struct {
		name              string
		columnType        string
		setupDB           func(t *testing.T, db **sql.DB, dbFile string)
		ok                bool
		wantConversionLog bool
		stdoutRegex       string
	}{
		{
			name:       "No conversion necessary",
			columnType: "TEXT",
			ok:         true,
		},
		{
			name:              "Conversion wanted",
			columnType:        "STRING",
			ok:                true,
			wantConversionLog: true,
		},
		{
			name:        "status table dropped",
			columnType:  "STRING",
			stdoutRegex: `no rows in result set`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				if _, err := (*db).Exec("DROP TABLE IF EXISTS status;"); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			},
			ok: false,
		},
		{
			name:        "Backup exists with different fields",
			columnType:  "STRING",
			stdoutRegex: `copy.*table .* has \d columns but \d values were supplied`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				if _, err := (*db).Exec(`
					CREATE TABLE IF NOT EXISTS status_backup (
						id TEXT NOT NULL PRIMARY KEY
					);`,
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			},
			ok: false,
		},
		{
			name:        "Cannot drop table because of foreign key",
			columnType:  "STRING",
			stdoutRegex: `drop.*FOREIGN KEY constraint failed`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				fkTable := "status_backup_fk"
				if _, err := (*db).Exec(`
					CREATE TABLE IF NOT EXISTS ` + fkTable + ` (
						id INTEGER NOT NULL PRIMARY KEY,
						fk_id TEXT NOT NULL,
						FOREIGN KEY(fk_id) REFERENCES status(id)
					);`,
				); err != nil {
					t.Fatalf("create table: DB.Exec() failed: %v", err)
				}
				if _, err := (*db).Exec(
					`INSERT OR REPLACE INTO ` + fkTable + ` (fk_id) VALUES ('keepMe');`,
				); err != nil {
					t.Fatalf("insert: DB.Exec() failed: %v", err)
				}
			},
			ok: false,
		},
		{
			name:        "Cannot alter backup table because of foreign key",
			columnType:  "STRING",
			stdoutRegex: `rename.*no such table`,
			setupDB: func(t *testing.T, db **sql.DB, _ string) {
				backupTableName := "status_backup"
				if err := createStatusTable(*db, backupTableName, "STRING"); err != nil {
					t.Fatalf("createStatusTable() failed: %v", err)
				}
				if _, err := (*db).Exec(`
					CREATE TRIGGER ` + backupTableName + `
					AFTER INSERT ON status_backup
					BEGIN
						INSERT INTO status (id) VALUES ('name');
					END;`,
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
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
					t.Fatalf("DB.Exec() failed: %v", err)
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
					t.Fatalf("sql.Open() failed: %v", err)
				}
				*db = newDB
			},
			ok: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			databaseFile := filepath.Join(t.TempDir(), "test.db")
			db, err := sql.Open("sqlite", databaseFile)
			if err != nil {
				t.Fatalf("sql.Open() failed: %v", err)
			}
			t.Cleanup(func() {
				_ = db.Close()
				_ = os.Remove(databaseFile)
			})

			// The foreign-key rows rely on enforcement, which is off by default.
			if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
				t.Fatalf("DB.Exec() failed: %v", err)
			}

			if err := createStatusTable(db, "status", tc.columnType); err != nil {
				t.Fatalf("createStatusTable() failed: %v", err)
			}

			var (
				id                       = "keepMe"
				latestVersion            = "0.0.3"
				latestVersionTimestamp   = "2020-01-02T01:01:01Z"
				deployedVersion          = "0.0.2"
				deployedVersionTimestamp = "2020-01-01T01:01:01Z"
				approvedVersion          = "0.0.1"
			)
			if err := insertStatusRow(
				db,
				id,
				latestVersion, latestVersionTimestamp,
				deployedVersion, deployedVersionTimestamp,
				approvedVersion,
			); err != nil {
				t.Fatalf("insertStatusRow() failed: %v", err)
			}

			if tc.setupDB != nil {
				tc.setupDB(t, &db, databaseFile)
			}

			test.AssertCallLog(
				t,
				tc.ok,
				func() bool { return updateTable(db) },
				logx.ExitCodeChannel(),
				nil,
				"",
			)
			if !tc.ok {
				stdout := releaseStdout()
				if !util.RegexCheck(tc.stdoutRegex, stdout) {
					t.Fatalf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
				}
				return
			}

			wantTextColumns := []string{"id", "latest_version", "deployed_version", "approved_version"}
			for _, column := range wantTextColumns {
				t.Run(column, func(t *testing.T) {
					t.Parallel()

					var columnType string
					if err := db.QueryRow(`
						SELECT type
						FROM pragma_table_info('status')
						WHERE name = ?`,
						column,
					).Scan(&columnType); err != nil {
						t.Fatalf("DB.QueryRow().Scan() failed: %v", err)
					}
					if columnType != "TEXT" {
						t.Errorf("type mismatch\ngot:  %q\nwant: \"TEXT\"", columnType)
					}
				})
			}

			got := queryRow(t, db, id)
			fieldTests := []test.FieldAssertion{
				{Name: "LatestVersion()", Got: got.LatestVersion(), Want: latestVersion, Mode: test.CompareEqual},
				{Name: "LatestVersionTimestamp()", Got: got.LatestVersionTimestamp(), Want: latestVersionTimestamp, Mode: test.CompareEqual},
				{Name: "DeployedVersion()", Got: got.DeployedVersion(), Want: deployedVersion, Mode: test.CompareEqual},
				{Name: "DeployedVersionTimestamp()", Got: got.DeployedVersionTimestamp(), Want: deployedVersionTimestamp, Mode: test.CompareEqual},
				{Name: "ApprovedVersion()", Got: got.ApprovedVersion(), Want: approvedVersion, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Status"); testErr != nil {
				t.Error(testErr)
			}

			stdout := releaseStdout()
			want := "Finished updating column types"
			contains := strings.Contains(stdout, want)
			if contains && !tc.wantConversionLog {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: NOT %q", stdout, want)
			} else if !contains && tc.wantConversionLog {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, want)
			}
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
