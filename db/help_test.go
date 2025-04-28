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
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	logtest "github.com/release-argus/Argus/test/log"
)

var packageName = "db"
var cfg *config.Config

func TestMain(m *testing.M) {
	databaseFile := "TestDB.db"

	// Log.
	logtest.InitLog()

	cfg = testConfig()
	cfg.Settings.Data.DatabaseFile = databaseFile
	Run(cfg)

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Remove(cfg.Settings.Data.DatabaseFile)
	os.Exit(exitCode)
}

func testConfig() (cfg *config.Config) {
	databaseFile := "test.db"
	databaseChannel := make(chan dbtype.Message, 16)
	saveChannel := make(chan bool, 16)
	cfg = &config.Config{
		Settings: config.Settings{
			SettingsBase: config.SettingsBase{
				Data: config.DataSettings{
					DatabaseFile: databaseFile,
				}},
		},
		Service: service.Slice{
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
		DatabaseChannel: &databaseChannel,
		SaveChannel:     &saveChannel,
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

func testAPI(primary string, secondary string) *api {
	testAPI := api{config: testConfig()}

	databaseFile := strings.ReplaceAll(
		fmt.Sprintf("%s-%s.db",
			primary, secondary),
		" ", "_")
	testAPI.config.Settings.Data.DatabaseFile = databaseFile

	return &testAPI
}

func dbCleanup(api *api) {
	if api.db != nil {
		api.db.Close()
	}
	os.Remove(api.config.Settings.Data.DatabaseFile)
	os.Remove(api.config.Settings.Data.DatabaseFile + "-journal")
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
	t.Cleanup(func() { row.Close() })

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
	status := status.Status{}
	status.Init(
		0, 0, 0,
		id, "", "",
		&dashboard.Options{
			WebURL: "https://example.com"})
	status.SetLatestVersion(lv, lvt, false)
	status.SetDeployedVersion(dv, dvt, false)
	status.SetApprovedVersion(av, false)

	return &status
}
