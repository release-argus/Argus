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

package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/release-argus/Argus/config"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

var (
	jLog    *utils.JLog
	logFrom *utils.LogFrom
)

func checkFile(path string) {
	dir := filepath.Dir(path)
	err := os.Mkdir(dir, 0755)
	if err == nil {
		return
	}
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dir)
		jLog.Fatal(
			fmt.Sprintf("db path %q exists but is not a directory", dir),
			*logFrom,
			!info.IsDir())
		jLog.Fatal(utils.ErrorToString(err), *logFrom, err != nil)
	}
	info, _ := os.Stat(path)
	jLog.Fatal(
		fmt.Sprintf("db path %q exists but is a directory, not a file", path),
		*logFrom,
		info != nil && info.IsDir())
}

func Run(cfg *config.Config, log *utils.JLog) {
	jLog = log
	databaseFile := cfg.Settings.GetDataDatabaseFile()
	logFrom = &utils.LogFrom{Primary: "db", Secondary: *databaseFile}

	api := api{config: cfg}
	api.initialise()
	defer api.db.Close()
	api.removeUnknownServices()
	api.convertServiceStatus()
	api.extractServiceStatus()

	api.handler()
}

func (api *api) initialise() {
	databaseFile := api.config.Settings.GetDataDatabaseFile()
	checkFile(*databaseFile)
	db, err := sql.Open("sqlite3", *databaseFile)
	jLog.Fatal(err, *logFrom, err != nil)

	// Create the table
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS status
		(
			id STRING NOT NULL PRIMARY KEY,
			latest_version STRING DEFAULT '',
			latest_version_timestamp DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			deployed_version STRING DEFAULT '',
			deployed_version_timestamp DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			approved_version STRING DEFAULT ''
		);`
	_, err = db.Exec(sqlStmt)
	jLog.Fatal(utils.ErrorToString(err), *logFrom, err != nil)

	api.db = db
}

// removeUnknownServices will remove rows with an id not in config.All
func (api *api) removeUnknownServices() {
	var allServices string
	for _, id := range api.config.All {
		allServices += fmt.Sprintf(`'%s',`, id)
	}
	sqlStmt := fmt.Sprintf(`
		DELETE FROM status
		WHERE id NOT IN (%s);`,
		allServices[:len(allServices)-1])
	_, err := api.db.Exec(sqlStmt)
	jLog.Fatal(
		fmt.Sprintf("removeUnknownServices: %s", utils.ErrorToString(err)),
		*logFrom,
		err != nil)
}

// extractServiceStatus will query the database and add the data found
// into the Service.Status inside the config
func (api *api) extractServiceStatus() {
	rows, err := api.db.Query(`
	SELECT
		id,
		latest_version,
		latest_version_timestamp,
		deployed_version,
		deployed_version_timestamp,
		approved_version
	FROM status;`)
	jLog.Fatal(err, *logFrom, err != nil)
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
		jLog.Fatal(
			fmt.Sprintf("extractServiceStatus row: %s", utils.ErrorToString(err)),
			*logFrom,
			err != nil)
		api.config.Service[id].Status.LatestVersion = lv
		api.config.Service[id].Status.LatestVersionTimestamp = lvt
		api.config.Service[id].Status.DeployedVersion = dv
		api.config.Service[id].Status.DeployedVersionTimestamp = dvt
		api.config.Service[id].Status.ApprovedVersion = av
	}
	err = rows.Err()
	jLog.Fatal(
		fmt.Sprintf("extractServiceStatus: %s", utils.ErrorToString(err)),
		*logFrom,
		err != nil)
}

// TODO: Deprecate
// convertServiceStatus will push Service.*.OldStatus to the DB
func (api *api) convertServiceStatus() {
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
	servicesToConvert := 0
	for _, id := range api.config.All {
		if (*api.config.Service[id]).OldStatus != nil {
			servicesToConvert++
			sqlStmt += fmt.Sprintf(" ('%s', '%s', '%s', '%s', '%s', '%s'),",
				id,
				(*api.config.Service[id]).OldStatus.LatestVersion,
				(*api.config.Service[id]).OldStatus.LatestVersionTimestamp,
				(*api.config.Service[id]).OldStatus.DeployedVersion,
				(*api.config.Service[id]).OldStatus.DeployedVersionTimestamp,
				(*api.config.Service[id]).OldStatus.ApprovedVersion,
			)
			(*api.config.Service[id]).Status = &service_status.Status{
				LatestVersion:            (*api.config.Service[id]).OldStatus.LatestVersion,
				LatestVersionTimestamp:   (*api.config.Service[id]).OldStatus.LatestVersionTimestamp,
				DeployedVersion:          (*api.config.Service[id]).OldStatus.DeployedVersion,
				DeployedVersionTimestamp: (*api.config.Service[id]).OldStatus.DeployedVersionTimestamp,
				ApprovedVersion:          (*api.config.Service[id]).OldStatus.ApprovedVersion,
			}
		}
	}
	if servicesToConvert != 0 {
		*api.config.SaveChannel <- true
		_, err := api.db.Exec(sqlStmt[:len(sqlStmt)-1] + ";")
		jLog.Fatal(
			fmt.Sprintf("convertServiceStatus: %s\n%s",
				utils.ErrorToString(err), sqlStmt),
			*logFrom,
			err != nil)
		for _, id := range api.config.All {
			(*api.config.Service[id]).OldStatus = nil
		}
	}
}
