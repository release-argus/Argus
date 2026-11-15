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

//go:build unit || integration

package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var cfg *config.Config

func TestMain(m *testing.M) {
	logtest.InitLog()

	databaseFile := "TestMain.db.test"

	cfg = testConfig(nil)
	cfg.Settings.Data.DatabaseFile = databaseFile

	ctx, cancel := context.WithCancel(context.Background())
	dbAPI := Get(cfg)
	go dbAPI.Handler(ctx)

	exitCode := m.Run()
	cancel()
	_ = os.Remove(cfg.Settings.Data.DatabaseFile)

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

func testConfig(t *testing.T) *config.Config {
	databaseFile := "test.db"
	if t != nil {
		databaseFile = filepath.Join(t.TempDir(), "test.db")
		file, err := os.Create(databaseFile)
		if err != nil {
			t.Fatalf("os.Create() failed: %v", err)
		}
		_ = file.Close()
	}

	databaseChannel := make(chan dbtype.Message, 16)
	saveChannel := make(chan bool, 16)
	c := &config.Config{
		Settings: config.Settings{
			Data: config.DataSettings{
				DatabaseFile: databaseFile,
			},
		},
		Service: service.Services{
			"delete0": nil,
			"keep0":   nil,
			"delete1": nil,
			"keep1":   nil,
			"delete2": nil,
			"keep2":   nil,
			"delete3": nil,
		},
		Order: []string{
			"delete0",
			"keep0",
			"delete1",
			"keep1",
			"delete2",
			"keep2",
			"delete3",
		},
		DatabaseChannel: databaseChannel,
		SaveChannel:     saveChannel,
	}

	if t != nil {
		t.Cleanup(func() {
			_ = os.Remove(c.Settings.Data.DatabaseFile)
		})
	}

	for svcName := range c.Service {
		svc := service.Service{
			ID:     svcName,
			Status: status.Status{},
			Dashboard: dashboard.Options{
				WebURL: "https://example.com",
			},
		}

		svc.Status.Init(
			len(svc.Command), len(svc.Notify), len(svc.Webhook),
			status.ServiceInfo{
				ID: svc.ID,
			},
			&svc.Dashboard,
		)
		svc.Status.SetDeployedVersion("2.0.0", "", false)
		svc.Status.SetLatestVersion("3.0.0", time.Now().Add(time.Hour).Format(time.RFC3339), false)
		svc.Status.SetApprovedVersion("1.0.0", false)

		c.Service[svcName] = &svc
	}

	return c
}

func testAPI(t *testing.T) *api {
	tAPI := &api{config: testConfig(t)}

	t.Cleanup(func() {
		dbCleanup(tAPI)
	})

	return tAPI
}

// assertFailsOnSchema checks call fails with a log matching errRegex on a status table rebuilt by statements.
func assertFailsOnSchema(t *testing.T, call func(*api) bool, statements []string, errRegex string) {
	t.Helper()

	releaseStdout := test.CaptureLog(t, logx.Default())
	tAPI := testAPI(t)
	tAPI.initialise()
	test.DropTable(t, tAPI.db, "status")
	test.ExecSQL(t, tAPI.db, statements...)
	go tAPI.Handler(t.Context())

	test.AssertCallLog(
		t,
		false,
		func() bool { return call(tAPI) },
		logx.ExitCodeChannel(),
		releaseStdout,
		errRegex,
	)
}

func dbCleanup(tAPI *api) {
	if tAPI.db != nil {
		_ = tAPI.db.Close()
	}
	_ = os.Remove(tAPI.config.Settings.Data.DatabaseFile)
	_ = os.Remove(tAPI.config.Settings.Data.DatabaseFile + "-journal")
}

func queryRow(t *testing.T, db *sql.DB, serviceID string) *status.Status {
	t.Helper()

	sqlStmt := `
		SELECT
			id,
			latest_version,
			latest_version_timestamp,
			deployed_version,
			deployed_version_timestamp,
			approved_version
		FROM status
		WHERE id = ?;`
	var row *sql.Rows
	var err error
	// Retry while the database is locked.
	for range 10 {
		row, err = db.Query(sqlStmt, serviceID)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("DB.Query(%q) failed: %v", serviceID, err)
	}
	t.Cleanup(func() { _ = row.Close() })

	var (
		id  string
		av  string
		dv  string
		dvt string
		lv  string
		lvt string
	)
	for row.Next() {
		err = row.Scan(&id, &lv, &lvt, &dv, &dvt, &av)
		if err != nil {
			t.Fatalf("Rows.Scan() failed: %v", err)
		}
	}
	if err := row.Err(); err != nil {
		t.Fatalf("Rows.Err() failed: %v", err)
	}
	svcStatus, err := statustest.New(
		"yaml", []byte(test.TrimYAML(`
			approved_version: '`+av+`'
			deployed_version: '`+dv+`'
			deployed_version_timestamp: '`+dvt+`'
			latest_version: '`+lv+`'
			latest_version_timestamp: '`+lvt+`'
		`)),
	)
	if err != nil {
		t.Fatalf("statustest.New(%q) failed: %v", serviceID, err)
	}
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: id,
		},
		&dashboard.Options{
			WebURL: "https://example.com",
		},
	)

	return svcStatus
}
