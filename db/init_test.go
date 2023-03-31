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
	"os"
	"regexp"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
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

func TestAPI_ConvertServiceStatus(t *testing.T) {
	// GIVEN a blank DB
	initLogging()
	tests := map[string]struct {
		runs int
	}{
		"one run": {
			runs: 1},
		"multiple runs": {
			runs: 2},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := testConfig()
			api := api{config: &cfg}
			*api.config.Settings.Data.DatabaseFile = "TestConvertServiceStatus" + name + ".db"
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
				} else if svc.OldStatus != nil {
					if svc.OldStatus != nil {
						t.Errorf("%q OldStatus should be nil, not %v",
							id, svc.OldStatus)
					}
				} else {
					if (*svc).Status.GetLatestVersion() != lv {
						t.Errorf("LatestVersion %q was not pushed to the db. Got %q",
							(*svc).Status.GetLatestVersion(), lv)
					}
					if (*svc).Status.GetLatestVersionTimestamp() != lvt {
						t.Errorf("LatestVersionTimestamp %q was not pushed to the db. Got %q",
							(*svc).Status.GetLatestVersionTimestamp(), lvt)
					}
					if (*svc).Status.GetDeployedVersion() != dv {
						t.Errorf("DeployedVersion %q was not pushed to the db. Got %q",
							(*svc).Status.GetDeployedVersion(), dv)
					}
					if (*svc).Status.GetDeployedVersionTimestamp() != dvt {
						t.Errorf("DeployedVersionTimestamp %q was not pushed to the db. Got %q",
							(*svc).Status.GetDeployedVersionTimestamp(), dvt)
					}
					if (*svc).Status.GetApprovedVersion() != av {
						t.Errorf("ApprovedVersion %q was not pushed to the db. Got %q",
							(*svc).Status.GetApprovedVersion(), av)
					}
				}
			}
			if count != len(api.config.Order) {
				t.Errorf("%d were pushed to the table. Expected %d",
					count, len(api.config.Order))
			}
			api.db.Close()
			os.Remove(*api.config.Settings.Data.DatabaseFile)
		})
	}
}

func TestDBQueryService(t *testing.T) {
	// GIVEN a blank DB
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestQueryService.db"
	api.initialise()

	// WHEN every Service.*.Status is pushed to the DB with convertServiceStatus
	api.convertServiceStatus()

	// THEN a Service that was copied over can be queried
	target := "keep0"
	got := queryRow(t, api.db, target)
	svc := api.config.Service[target]
	if (*svc).Status.GetLatestVersion() != got.GetLatestVersion() {
		t.Errorf("LatestVersion %q was not pushed to the db. Got %q",
			(*svc).Status.GetLatestVersion(), got.GetLatestVersion())
	}
	if (*svc).Status.GetLatestVersionTimestamp() != got.GetLatestVersionTimestamp() {
		t.Errorf("LatestVersionTimestamp %q was not pushed to the db. Got %q",
			(*svc).Status.GetLatestVersionTimestamp(), got.GetLatestVersionTimestamp())
	}
	if (*svc).Status.GetDeployedVersion() != got.GetDeployedVersion() {
		t.Errorf("DeployedVersion %q was not pushed to the db. Got %q\n%v\n%s",
			(*svc).Status.GetDeployedVersion(), got.GetDeployedVersion(), got, (*svc).Status.String())
	}
	if (*svc).Status.GetDeployedVersionTimestamp() != got.GetDeployedVersionTimestamp() {
		t.Errorf("DeployedVersionTimestamp %q was not pushed to the db. Got %q",
			(*svc).Status.GetDeployedVersionTimestamp(), got.GetDeployedVersionTimestamp())
	}
	if (*svc).Status.GetApprovedVersion() != got.GetApprovedVersion() {
		t.Errorf("ApprovedVersion %q was not pushed to the db. Got %q",
			(*svc).Status.GetApprovedVersion(), got.GetApprovedVersion())
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
}

func TestAPI_RemoveUnknownServices(t *testing.T) {
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

func TestRun(t *testing.T) {
	// GIVEN a DB is running
	initLogging()
	cfg := testConfig()
	*cfg.Settings.Data.DatabaseFile = "TestRun.db"
	go Run(&cfg, jLog)

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
	*otherCfg.Settings.Data.DatabaseFile = "TestRun-copy.db"
	bytesRead, err := os.ReadFile(*cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(*otherCfg.Settings.Data.DatabaseFile, bytesRead, os.FileMode(0644))
	if err != nil {
		t.Fatal(err)
	}
	api := api{config: &otherCfg}
	api.initialise()
	got := queryRow(t, api.db, target)
	want := svcstatus.Status{}
	want.Init(
		0, 0, 0,
		&target,
		stringPtr("https://example.com"))
	want.SetLatestVersion("9.9.9", false)
	want.SetLatestVersionTimestamp("2022-01-01T01:01:01Z")
	want.SetApprovedVersion("0.0.1")
	want.SetDeployedVersion("0.0.0", false)
	want.SetDeployedVersionTimestamp("2020-01-01T01:01:01Z")
	if got.GetLatestVersion() != want.GetLatestVersion() {
		t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
			cell.Column, cell.Value, got, want.String())
	}
	api.db.Close()
	os.Remove(*cfg.Settings.Data.DatabaseFile)
	os.Remove(*otherCfg.Settings.Data.DatabaseFile)
}
