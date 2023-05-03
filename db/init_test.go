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

//go:build unit

package db

import (
	"fmt"
	"math/rand"
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
	cfg := testConfig()
	api := api{config: cfg}
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

func TestDBQueryService(t *testing.T) {
	// GIVEN a blank DB
	cfg := testConfig()
	api := api{config: cfg}
	*api.config.Settings.Data.DatabaseFile = "TestQueryService.db"
	api.initialise()
	// Get a Service from the Config
	var serviceName string
	for k := range api.config.Service {
		serviceName = k
		break
	}
	svc := api.config.Service[serviceName]

	// WHEN the database contains data for a Service
	api.updateRow(
		serviceName,
		[]dbtype.Cell{
			{Column: "id", Value: serviceName},
			{Column: "latest_version", Value: (*svc).Status.GetLatestVersion()},
			{Column: "latest_version_timestamp", Value: (*svc).Status.GetLatestVersionTimestamp()},
			{Column: "deployed_version", Value: (*svc).Status.GetDeployedVersion()},
			{Column: "deployed_version_timestamp", Value: (*svc).Status.GetDeployedVersionTimestamp()},
			{Column: "approved_version", Value: (*svc).Status.GetApprovedVersion()}})

	// THEN that data can be queried
	got := queryRow(t, api.db, serviceName)
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
	cfg := testConfig()
	api := api{config: cfg}
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
		sqlStmt += fmt.Sprintf(" (%q, %q, %q, %q, %q, %q),",
			id,
			svc.Status.GetLatestVersion(),
			svc.Status.GetLatestVersionTimestamp(),
			svc.Status.GetDeployedVersion(),
			svc.Status.GetDeployedVersionTimestamp(),
			svc.Status.GetApprovedVersion(),
		)
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

func TestAPI_Run(t *testing.T) {
	// GIVEN a DB is running (see TestMain)

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
	*otherCfg.Settings.Data.DatabaseFile = "TestAPI_Run-copy.db"
	bytesRead, err := os.ReadFile(*cfg.Settings.Data.DatabaseFile)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(*otherCfg.Settings.Data.DatabaseFile, bytesRead, os.FileMode(0644))
	if err != nil {
		t.Fatal(err)
	}
	api := api{config: otherCfg}
	api.initialise()
	got := queryRow(t, api.db, target)
	want := svcstatus.Status{}
	want.Init(
		0, 0, 0,
		&target,
		stringPtr("https://example.com"))
	want.SetLatestVersion("9.9.9", false)
	want.SetLatestVersionTimestamp("2022-01-01T01:01:01Z")
	want.SetApprovedVersion("0.0.1", false)
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

func TestAPI_extractServiceStatus(t *testing.T) {
	// GIVEN an API on a DB containing atleast 1 row
	cfg := testConfig()
	*cfg.Settings.Data.DatabaseFile = "TestAPI_extractServiceStatus.db"
	defer os.Remove(*cfg.Settings.Data.DatabaseFile)
	go func() {
		api := api{config: cfg}
		api.initialise()
		api.handler()
	}()
	wantStatus := make([]svcstatus.Status, len(cfg.Service))
	// push a random Status for each Service to the DB
	index := 0
	for id, svc := range cfg.Service {
		id := id
		wantStatus[index].ServiceID = &id
		wantStatus[index].SetLatestVersion(fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)), false)
		wantStatus[index].SetLatestVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
		wantStatus[index].SetDeployedVersion(fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)), false)
		wantStatus[index].SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
		wantStatus[index].SetApprovedVersion(fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)), false)

		*cfg.DatabaseChannel <- dbtype.Message{
			ServiceID: id,
			Cells: []dbtype.Cell{
				{Column: "id", Value: id},
				{Column: "latest_version", Value: wantStatus[index].GetLatestVersion()},
				{Column: "latest_version_timestamp", Value: wantStatus[index].GetLatestVersionTimestamp()},
				{Column: "deployed_version", Value: wantStatus[index].GetDeployedVersion()},
				{Column: "deployed_version_timestamp", Value: wantStatus[index].GetDeployedVersionTimestamp()},
				{Column: "approved_version", Value: wantStatus[index].GetApprovedVersion()}}}
		// Clear the Status in the Config
		svc.Status = *svcstatus.New(
			svc.Status.AnnounceChannel, svc.Status.DatabaseChannel, svc.Status.SaveChannel,
			"", "", "", "", "", "")
		index++
	}
	time.Sleep(250 * time.Millisecond)
	api := api{config: cfg}
	api.initialise()

	// WHEN extractServiceStatus is called
	api.extractServiceStatus()

	// THEN the Status in the Config is updated
	for i := range wantStatus {
		row := queryRow(t, api.db, *wantStatus[i].ServiceID)
		if row.GetLatestVersion() != wantStatus[i].GetLatestVersion() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"latest_version", row.GetLatestVersion(), row, wantStatus[i].String())
		}
		if row.GetLatestVersionTimestamp() != wantStatus[i].GetLatestVersionTimestamp() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"latest_version_timestamp", row.GetLatestVersionTimestamp(), row, wantStatus[i].String())
		}
		if row.GetDeployedVersion() != wantStatus[i].GetDeployedVersion() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"deployed_version", row.GetDeployedVersion(), row, wantStatus[i].String())
		}
		if row.GetDeployedVersionTimestamp() != wantStatus[i].GetDeployedVersionTimestamp() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"deployed_version_timestamp", row.GetDeployedVersionTimestamp(), row, wantStatus[i].String())
		}
		if row.GetApprovedVersion() != wantStatus[i].GetApprovedVersion() {
			t.Errorf("Expected %q to be updated to %q\ngot %q, want %q",
				"approved_version", row.GetApprovedVersion(), row, wantStatus[i].String())
		}
	}
}
