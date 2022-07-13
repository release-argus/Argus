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
	"database/sql"
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config"
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func initLogging() {
	jLog = utils.NewJLog("WARN", false)
	jLog.Testing = true
	logFrom = &utils.LogFrom{}
}

func testConfig() config.Config {
	databaseFile := "test.db"
	svc := service.Service{
		Status: &service_status.Status{},
		OldStatus: &service_status.OldStatus{
			LatestVersion:            "0.0.2",
			LatestVersionTimestamp:   "2022-01-01T01:01:01Z",
			DeployedVersion:          "0.0.0",
			DeployedVersionTimestamp: "2020-01-01T01:01:01Z",
			ApprovedVersion:          "0.0.1",
		},
	}
	databaseChannel := make(chan db_types.Message, 5)
	saveChannel := make(chan bool, 16)
	return config.Config{
		Settings: config.Settings{
			Data: config.DataSettings{
				DatabaseFile: &databaseFile,
			},
		},
		Service: service.Slice{
			"delete0": &svc,
			"keep0":   &svc,
			"delete1": &svc,
			"keep1":   &svc,
			"delete2": &svc,
			"keep2":   &svc,
			"delete3": &svc,
		},
		All: []string{
			"keep0",
			"keep1",
			"keep2",
		},
		DatabaseChannel: &databaseChannel,
		SaveChannel:     &saveChannel,
	}
}

func queryRow(db *sql.DB, serviceID string, t *testing.T) service_status.Status {
	sqlStmt := fmt.Sprintf(`
	SELECT
		id,
		latest_version,
		latest_version_timestamp,
		deployed_version,
		deployed_version_timestamp,
		approved_version
	FROM status
	WHERE id = '%s';`,
		serviceID)
	row, err := db.Query(sqlStmt)
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
	return service_status.Status{
		LatestVersion:            lv,
		LatestVersionTimestamp:   lvt,
		DeployedVersion:          dv,
		DeployedVersionTimestamp: dvt,
		ApprovedVersion:          av,
	}
}
