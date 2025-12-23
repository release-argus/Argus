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
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	logtest "github.com/release-argus/Argus/test/log"
	logutil "github.com/release-argus/Argus/util/log"
)

var packageName = "db"
var cfg *config.Config

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	databaseFile := "TestMain.db.test"

	cfg = testConfig(nil)
	cfg.Settings.Data.DatabaseFile = databaseFile
	// AND a cancellable context for shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	Run(ctx, cfg)

	// Run other tests.
	exitCode := m.Run()
	cancel()
	_ = os.Remove(cfg.Settings.Data.DatabaseFile)

	if len(logutil.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty",
			packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

func testConfig(t *testing.T) (cfg *config.Config) {
	databaseFile := "test.db"
	var baseDir = ""
	if t != nil {
		baseDir = t.TempDir()
		databaseFile = filepath.Join(baseDir, "test.db")
		_, _ = os.Create(databaseFile)
	}

	databaseChannel := make(chan dbtype.Message, 16)
	saveChannel := make(chan bool, 16)
	cfg = &config.Config{
		Settings: config.Settings{
			SettingsBase: config.SettingsBase{
				Data: config.DataSettings{
					DatabaseFile: databaseFile,
				}},
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

	// Services.
	for svcName := range cfg.Service {
		svc := service.Service{
			ID:     "foo",
			Status: status.Status{},
			Dashboard: dashboard.Options{
				WebURL: "https://example.com"}}
		svc.Status.Init(
			len(svc.Notify), len(svc.Command), len(svc.WebHook),
			svc.ID, "", "",
			&svc.Dashboard)
		svc.Status.SetApprovedVersion("1.0.0", false)
		svc.Status.SetDeployedVersion("2.0.0", "", false)
		svc.Status.SetLatestVersion("3.0.0", time.Now().Add(time.Hour).Format(time.RFC3339), false)

		// Add service to Config.
		cfg.Service[svcName] = &svc
	}

	return
}

func testAPI(t *testing.T) *api {
	return &api{config: testConfig(t)}

}

func dbCleanup(api *api) {
	if api.db != nil {
		_ = api.db.Close()
	}
	_ = os.Remove(api.config.Settings.Data.DatabaseFile)
	_ = os.Remove(api.config.Settings.Data.DatabaseFile + "-journal")
}

func queryRow(t *testing.T, db *sql.DB, serviceID string) *status.Status {
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
	// Retry up-to 10 times in case 'database is locked'.
	var row *sql.Rows
	var err error
	for i := 0; i < 10; i++ {
		row, err = db.Query(sqlStmt, serviceID)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = row.Close() })

	var (
		id  string
		lv  string
		lvt string
		dv  string
		dvt string
		av  string
	)
	for row.Next() {
		err = row.Scan(&id, &lv, &lvt, &dv, &dvt, &av)
		if err != nil {
			t.Fatal(err)
		}
	}
	svcStatus := status.Status{}
	svcStatus.Init(
		0, 0, 0,
		id, "", "",
		&dashboard.Options{
			WebURL: "https://example.com"})
	svcStatus.SetLatestVersion(lv, lvt, false)
	svcStatus.SetDeployedVersion(dv, dvt, false)
	svcStatus.SetApprovedVersion(av, false)

	return &svcStatus
}
