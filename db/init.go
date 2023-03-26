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
	"strings"

	"github.com/release-argus/Argus/config"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog    *util.JLog
	logFrom *util.LogFrom
)

func checkFile(path string) {
	file := filepath.Base(path)
	// Check that the directory exists
	dir := filepath.Dir(path)
	fileInfo, err := os.Stat(dir)
	if err != nil {
		// directory doesn't exist
		// create the dir
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			jLog.Fatal(util.ErrorToString(err), *logFrom, err != nil)
		} else {
			// other error
			jLog.Fatal(util.ErrorToString(err), *logFrom, true)
		}

		// directory exists but is not a directory
	} else if fileInfo == nil || !fileInfo.IsDir() {
		jLog.Fatal(fmt.Sprintf("path %q (for %q) is not a directory", dir, file), *logFrom, true)
	}

	// Check that the file exists
	fileInfo, err = os.Stat(path)
	if err != nil {
		// file doesn't exist
		jLog.Fatal(util.ErrorToString(err), *logFrom, os.IsExist(err))

		// item exists but is a directory
	} else if fileInfo != nil && fileInfo.IsDir() {
		jLog.Fatal(fmt.Sprintf("path %q (for %q) is a directory, not a file", path, file), *logFrom, true)
	}
}

func Run(cfg *config.Config, log *util.JLog) {
	jLog = log
	databaseFile := cfg.Settings.GetDataDatabaseFile()
	logFrom = &util.LogFrom{Primary: "db", Secondary: *databaseFile}

	api := api{config: cfg}
	api.initialise()
	defer api.db.Close()
	if len(api.config.Order) > 0 {
		api.removeUnknownServices()
		api.convertServiceStatus()
		api.extractServiceStatus()
	}

	api.handler()
}

func (api *api) initialise() {
	databaseFile := api.config.Settings.GetDataDatabaseFile()
	checkFile(*databaseFile)
	db, err := sql.Open("sqlite", *databaseFile)
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
	jLog.Fatal(util.ErrorToString(err), *logFrom, err != nil)

	api.db = db
}

// removeUnknownServices will remove rows with an id not in config.Order
func (api *api) removeUnknownServices() {
	// ? for each service
	services := strings.Repeat(`?,`, len(api.config.Order))

	// SQL statement to remove unknown services
	sqlStmt := fmt.Sprintf(`
		DELETE FROM status
		WHERE id NOT IN (%s);`,
		services[:len(services)-1])

	// Get the vars for the SQL statement
	params := make([]interface{}, len(api.config.Order))
	for i, name := range api.config.Order {
		params[i] = name
	}

	_, err := api.db.Exec(sqlStmt, params...)
	jLog.Fatal(
		fmt.Sprintf("removeUnknownServices: %s", util.ErrorToString(err)),
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
			fmt.Sprintf("extractServiceStatus row: %s", util.ErrorToString(err)),
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
		fmt.Sprintf("extractServiceStatus: %s", util.ErrorToString(err)),
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
	for _, id := range api.config.Order {
		if api.config.Service[id].OldStatus != nil {
			servicesToConvert++
			sqlStmt += fmt.Sprintf(" ('%s', '%s', '%s', '%s', '%s', '%s'),",
				id,
				api.config.Service[id].OldStatus.LatestVersion,
				api.config.Service[id].OldStatus.LatestVersionTimestamp,
				api.config.Service[id].OldStatus.DeployedVersion,
				api.config.Service[id].OldStatus.DeployedVersionTimestamp,
				api.config.Service[id].OldStatus.ApprovedVersion,
			)
			api.config.Service[id].Status = svcstatus.Status{
				LatestVersion:            api.config.Service[id].OldStatus.LatestVersion,
				LatestVersionTimestamp:   api.config.Service[id].OldStatus.LatestVersionTimestamp,
				DeployedVersion:          api.config.Service[id].OldStatus.DeployedVersion,
				DeployedVersionTimestamp: api.config.Service[id].OldStatus.DeployedVersionTimestamp,
				ApprovedVersion:          api.config.Service[id].OldStatus.ApprovedVersion,
			}
		}
	}
	if servicesToConvert != 0 {
		*api.config.SaveChannel <- true
		_, err := api.db.Exec(sqlStmt[:len(sqlStmt)-1] + ";")
		jLog.Fatal(
			fmt.Sprintf("convertServiceStatus: %s\n%s",
				util.ErrorToString(err), sqlStmt),
			*logFrom,
			err != nil)
		for _, id := range api.config.Order {
			api.config.Service[id].OldStatus = nil
		}
	}
}
