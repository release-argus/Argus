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
	"path/filepath"
	"strings"
	"testing"
	"time"

	db_types "github.com/release-argus/Argus/db/types"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	_ "modernc.org/sqlite"
)

func TestCheckFilePassNoDirBefore(t *testing.T) {
	// GIVEN a dir that doesn't exist
	initLogging()
	path := "TestCheckFilePassNoDirBefore/argus.db"
	folder := filepath.Dir(path)
	os.RemoveAll(folder)

	// WHEN checkFile is called on that same dir
	checkFile(path)

	// THEN we get here with no panic
	os.RemoveAll(folder)
}

func TestCheckFilePassHadDirBefore(t *testing.T) {
	// GIVEN the a dir that already exists
	initLogging()
	path := "TestCheckFilePassHadDirBefore/argus.db"
	folder := filepath.Dir(path)
	os.RemoveAll(folder)
	err := os.Mkdir(folder, os.ModeDir)
	if err != nil {
		t.Fatal(err)
	}

	// WHEN checkFile is called on that same dir
	checkFile(path)

	// THEN we get here with no panic
	os.RemoveAll(folder)
}

func TestCheckFileFailNotDirectoryOnTheWay(t *testing.T) {
	// GIVEN a dir that is actually a file on the way to the db
	initLogging()
	path := "TestCheckFileFailNotDirectoryOnTheWay/argus.db"
	folder := filepath.Dir(path)
	os.RemoveAll(folder)
	_, err := os.Create(folder)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		r := recover()
		os.RemoveAll(folder)
		if !strings.Contains(r.(string), " exists but is not a directory") {
			t.Error(r)
		}
	}()

	// WHEN checkFile is called on that same dir
	checkFile(path)

	// THEN we panic and don't reach this
	t.Errorf("Shouldn't reach this as %q is used as a dir in %q, but should be a file",
		folder, path)
}

func TestCheckFileFailNotDirectoryAtTheEnd(t *testing.T) {
	// GIVEN a dir that is actually a file at the end
	initLogging()
	path := "TestCheckFileFailNotDirectoryAtTheEnd/argus.db"
	folder := filepath.Dir(path)
	os.RemoveAll(folder)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		r := recover()
		os.RemoveAll(folder)
		if !strings.Contains(r.(string), " exists but is a directory, not a file") {
			t.Error(r)
		}
	}()

	// WHEN checkFile is called on that same dir
	checkFile(path)

	// THEN we panic and don't reach this
	t.Errorf("Shouldn't reach this as %q is used as a dir in %q, but should be a file",
		folder, path)
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

func TestConvertServiceStatusWhenNotConverted(t *testing.T) {
	// GIVEN a blank DB
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestConvertServiceStatusWhenNotConverted.db"
	api.initialise()

	// WHEN we call convertServiceStatus
	api.convertServiceStatus()

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
}

func TestConvertServiceStatusWhenAlreadyConverted(t *testing.T) {
	// GIVEN a blank DB
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestConvertServiceStatusWhenAlreadyConverted.db"
	api.initialise()

	// WHEN every Service.*.Status is pushed to the DB with convertServiceStatus
	// and then convertServiceStatus is ran again with nothing to convert
	api.convertServiceStatus()
	api.convertServiceStatus()

	// THEN each Service.*.Status can be queried from the DB
	// (they weren't removed when the conversion was ran a second time)
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
}

func TestQueryService(t *testing.T) {
	// GIVEN a blank DB
	initLogging()
	cfg := testConfig()
	api := api{config: &cfg}
	*api.config.Settings.Data.DatabaseFile = "TestQueryService.db"
	api.initialise()

	// WHEN every Service.*.Status is pushed to the DB with convertServiceStatus
	api.convertServiceStatus()

	// THEN a Services that was copied over can be queried
	target := "keep0"
	got := queryRow(api.db, target, t)
	svc := api.config.Service[target]
	if (*svc).Status.LatestVersion != got.LatestVersion {
		t.Errorf("LatestVersion %q was not pushed to the db. Got %q",
			(*svc).Status.LatestVersion, got.LatestVersion)
	}
	if (*svc).Status.LatestVersionTimestamp != got.LatestVersionTimestamp {
		t.Errorf("LatestVersionTimestamp %q was not pushed to the db. Got %q",
			(*svc).Status.LatestVersionTimestamp, got.LatestVersionTimestamp)
	}
	if (*svc).Status.DeployedVersion != got.DeployedVersion {
		t.Errorf("DeployedVersion %q was not pushed to the db. Got %q\n%v\n%v",
			(*svc).Status.DeployedVersion, got.DeployedVersion, got, (*svc).Status)
	}
	if (*svc).Status.DeployedVersionTimestamp != got.DeployedVersionTimestamp {
		t.Errorf("DeployedVersionTimestamp %q was not pushed to the db. Got %q",
			(*svc).Status.DeployedVersionTimestamp, got.DeployedVersionTimestamp)
	}
	if (*svc).Status.ApprovedVersion != got.ApprovedVersion {
		t.Errorf("ApprovedVersion %q was not pushed to the db. Got %q",
			(*svc).Status.ApprovedVersion, got.ApprovedVersion)
	}
	api.db.Close()
	os.Remove(*api.config.Settings.Data.DatabaseFile)
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
	err = ioutil.WriteFile(*otherCfg.Settings.Data.DatabaseFile, bytesRead, 0644)
	if err != nil {
		t.Fatal(err)
	}
	api := api{config: &otherCfg}
	api.initialise()
	got := queryRow(api.db, target, t)
	want := service_status.Status{
		LatestVersion:            "9.9.9",
		LatestVersionTimestamp:   "2022-01-01T01:01:01Z",
		ApprovedVersion:          "0.0.1",
		DeployedVersion:          "0.0.0",
		DeployedVersionTimestamp: "2020-01-01T01:01:01Z",
	}
	if got.LatestVersion != want.LatestVersion {
		t.Errorf("Expected %q to be updated to %q\ngot  %v\nwant %v",
			cell.Column, cell.Value, got, want)
	}
	api.db.Close()
	os.Remove(*cfg.Settings.Data.DatabaseFile)
	os.Remove(*otherCfg.Settings.Data.DatabaseFile)
}
