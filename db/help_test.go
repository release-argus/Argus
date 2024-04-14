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

//go:build unit || integration

package db

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var cfg *config.Config

func TestMain(m *testing.M) {
	log := util.NewJLog("DEBUG", false)
	log.Testing = true
	databaseFile := "TestRun.db"
	LogInit(log, databaseFile)

	cfg = testConfig()
	*cfg.Settings.Data.DatabaseFile = databaseFile
	defer os.Remove(*cfg.Settings.Data.DatabaseFile)
	go Run(cfg)
	time.Sleep(250 * time.Millisecond) // Time for db to start

	os.Exit(m.Run())
}

func stringPtr(val string) *string {
	return &val
}

func testConfig() (cfg *config.Config) {
	databaseFile := "test.db"
	databaseChannel := make(chan dbtype.Message, 16)
	saveChannel := make(chan bool, 16)
	cfg = &config.Config{
		Settings: config.Settings{
			SettingsBase: config.SettingsBase{
				Data: config.DataSettings{
					DatabaseFile: &databaseFile,
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
	// Services
	for svcName := range cfg.Service {
		svc := service.Service{
			ID:     "foo",
			Status: svcstatus.Status{},
		}
		svc.Status.Init(
			len(svc.Notify), len(svc.Command), len(svc.WebHook),
			&svc.ID,
			stringPtr("https://example.com"))
		svc.Status.SetApprovedVersion("1.0.0", false)
		svc.Status.SetDeployedVersion("2.0.0", false)
		svc.Status.SetDeployedVersionTimestamp(time.Now().Format(time.RFC3339))
		svc.Status.SetLatestVersion("3.0.0", false)
		svc.Status.SetLatestVersionTimestamp(time.Now().Add(time.Hour).Format(time.RFC3339))

		// Add service to Config
		cfg.Service[svcName] = &svc
	}

	return
}

func queryRow(t *testing.T, db *sql.DB, serviceID string) *svcstatus.Status {
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
	// Retry up-to 10 times incase 'database is locked'
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
	defer row.Close()

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
	status := svcstatus.Status{}
	status.Init(
		0, 0, 0,
		&id,
		stringPtr("https://example.com"))
	status.SetLatestVersion(lv, false)
	status.SetLatestVersionTimestamp(lvt)
	status.SetDeployedVersion(dv, false)
	status.SetDeployedVersionTimestamp(dvt)
	status.SetApprovedVersion(av, false)

	return &status
}
