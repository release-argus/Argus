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
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	db_types "github.com/release-argus/Argus/db/types"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	_ "modernc.org/sqlite"
)

func TestCheckFile(t *testing.T) {
	// GIVEN various paths
	initLogging()
	tests := map[string]struct {
		removeBefore     string
		createDirBefore  string
		createFileBefore string
		path             string
		panicContains    string
	}{
		"file doesn't exist":      {path: "something_doesnt_exist.db", removeBefore: "something_doesnt_exist.db"},
		"dir doesn't exist":       {path: "dir_doesnt_exist/argus.db", removeBefore: "dir_doesnt_exist"},
		"dir exists but not file": {path: "dir_does_exist/argus.db", createDirBefore: "dir_does_exist"},
		"file is dir": {path: "folder.db", createDirBefore: "folder.db",
			panicContains: " exists but is a directory"},
		"dir is file": {path: "folder_not_a_dir/argus.db", createFileBefore: "folder_not_a_dir",
			panicContains: " exists but is not a directory"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.RemoveAll(tc.removeBefore)
			os.RemoveAll(tc.createDirBefore)
			if tc.createDirBefore != "" {
				err := os.Mkdir(tc.createDirBefore, os.ModeDir|0755)
				if err != nil {
					t.Fatalf("%s:\n%s",
						name, err)
				}
			}
			if tc.createFileBefore != "" {
				file, err := os.Create(tc.createFileBefore)
				if err != nil {
					t.Fatalf("%s:\n%s",
						name, err)
				}
				file.Close()
			}
			if tc.panicContains != "" {
				defer func() {
					r := recover()
					rStr := fmt.Sprint(r)
					os.RemoveAll(tc.createDirBefore)
					os.RemoveAll(tc.createFileBefore)
					if !strings.Contains(r.(string), tc.panicContains) {
						t.Errorf("%s:\nshould have panic'd with:\n%q, not:\n%q",
							name, tc.panicContains, rStr)
					}
				}()
			}

			// WHEN checkFile is called on that same dir
			checkFile(tc.path)

			// THEN we get here only when we should
			if tc.panicContains != "" {
				t.Fatalf("%s:\nExpected panic with %q",
					name, tc.panicContains)
			}
		})
	}
}

func TestInitialise(t *testing.T) {
	// GIVEN a config with a database location
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
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

func TestConvertServiceStatus(t *testing.T) {
	// GIVEN a blank DB
	initLogging()
	tests := map[string]struct {
		runs int
	}{
		"one run":       {runs: 1},
		"multiple runs": {runs: 2},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := testConfig()
			api := api{config: &cfg}
			*api.config.Settings.Data.DatabaseFile = "TestConvertServiceStatus.db"
			api.initialise()

			// WHEN we call convertServiceStatus
			for i := 0; i < tc.runs; i++ {
				api.convertServiceStatus()
			}

			// THEN each Service.*.OldStatus is pushed to the DB and can be queried
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
				if svc == nil {
					t.Errorf("%q was not pushed to the table",
						id)
				} else if svc.Status == nil || svc.OldStatus != nil {
					if svc.OldStatus != nil {
						t.Errorf("%q OldStatus should be nil, not %v",
							id, svc.OldStatus)
					}
					if svc.Status == nil {
						t.Errorf("%q Status was not converted. Status=%v",
							id, svc.Status)
					}
				} else {
					if (*svc).Status.LatestVersion != lv {
						t.Errorf("LatestVersion %q was not pushed to the db. Got %q",
							(*svc).Status.LatestVersion, lv)
					}
					if (*svc).Status.LatestVersionTimestamp != lvt {
						t.Errorf("LatestVersionTimestamp %q was not pushed to the db. Got %q",
							(*svc).Status.LatestVersionTimestamp, lvt)
					}
					if (*svc).Status.DeployedVersion != dv {
						t.Errorf("DeployedVersion %q was not pushed to the db. Got %q",
							(*svc).Status.DeployedVersion, dv)
					}
					if (*svc).Status.DeployedVersionTimestamp != dvt {
						t.Errorf("DeployedVersionTimestamp %q was not pushed to the db. Got %q",
							(*svc).Status.DeployedVersionTimestamp, dvt)
					}
					if (*svc).Status.ApprovedVersion != av {
						t.Errorf("ApprovedVersion %q was not pushed to the db. Got %q",
							(*svc).Status.ApprovedVersion, av)
					}
				}
			}
			if count != len(api.config.All) {
				t.Errorf("%d were pushed to the table. Expected %d",
					count, len(api.config.All))
			}
			api.db.Close()
			os.Remove(*api.config.Settings.Data.DatabaseFile)
		})
	}
}

func TestRemoveUnknownServices(t *testing.T) {
	// GIVEN a DB with loads of service status'
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
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
		if svc.OldStatus != nil {
			sqlStmt += fmt.Sprintf(" (%q, %q, %q, %q, %q, %q),",
				id,
				svc.OldStatus.LatestVersion,
				svc.OldStatus.LatestVersionTimestamp,
				svc.OldStatus.DeployedVersion,
				svc.OldStatus.DeployedVersionTimestamp,
				svc.OldStatus.ApprovedVersion,
			)
		}
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
		if svc == nil || !utils.Contains(api.config.All, id) {
			t.Errorf("%q should have been removed from the table",
				id)
		}
	}
	if count != len(api.config.All) {
		t.Errorf("Only %d were left in the table. Expected %d",
			count, len(api.config.All))
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}

func TestRun(t *testing.T) {
	// GIVEN a DB is running
	initLogging()
	cfg := testConfig()
	*cfg.Settings.Data.DatabaseFile = "TestRun.db"
	go Run(&cfg, jLog)

	// WHEN a message is send to the DatabaseChannel targeting latest_version
	target := "keep0"
	cell := db_types.Cell{Column: "latest_version", Value: "9.9.9"}
	*cfg.DatabaseChannel <- db_types.Message{
		ServiceID: target,
		Cells:     []db_types.Cell{cell},
	}
	time.Sleep(time.Second)

	// THEN the cell was changed in the DB
	otherCfg := testConfig()
	*otherCfg.Settings.Data.DatabaseFile = "TestRun-copy.db"
	bytesRead, err := ioutil.ReadFile(*cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(*otherCfg.Settings.Data.DatabaseFile, bytesRead, os.FileMode(0644))
	if err != nil {
		t.Fatal(err)
	}
	api := api{config: &otherCfg}
	api.initialise()
	got := queryRow(t, api.db, target)
	want := service_status.Status{
		LatestVersion:            "9.9.9",
		LatestVersionTimestamp:   "2022-01-01T01:01:01Z",
		ApprovedVersion:          "0.0.1",
		DeployedVersion:          "0.0.0",
		DeployedVersionTimestamp: "2020-01-01T01:01:01Z",
	}
	if got != want {
		t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
			cell.Column, cell.Value, got, want)
	}
	api.db.Close()
	os.Remove(*cfg.Settings.Data.DatabaseFile)
	os.Remove(*otherCfg.Settings.Data.DatabaseFile)
}
