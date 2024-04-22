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

package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/util"
)

// LogInit for this package.
func LogInit(log *util.JLog, databaseFile string) {
	// Only set the log if it hasn't been set (avoid RACE condition)
	if jLog == nil {
		jLog = log
		logFrom = &util.LogFrom{Primary: "db", Secondary: databaseFile}
	}
}

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
			jLog.Fatal(util.ErrorToString(err), logFrom, err != nil)
		} else {
			// other error
			jLog.Fatal(util.ErrorToString(err), logFrom, true)
		}

		// directory exists but is not a directory
	} else if fileInfo == nil || !fileInfo.IsDir() {
		jLog.Fatal(fmt.Sprintf("path %q (for %q) is not a directory", dir, file), logFrom, true)
	}

	// Check that the file exists
	fileInfo, err = os.Stat(path)
	if err != nil {
		// file doesn't exist
		jLog.Fatal(util.ErrorToString(err), logFrom, os.IsExist(err))

		// item exists but is a directory
	} else if fileInfo != nil && fileInfo.IsDir() {
		jLog.Fatal(fmt.Sprintf("path %q (for %q) is a directory, not a file", path, file), logFrom, true)
	}
}

func Run(cfg *config.Config, log *util.JLog) {
	api := api{config: cfg}
	if log != nil {
		LogInit(log, cfg.Settings.DataDatabaseFile())
	}
	api.initialise()
	defer api.db.Close()
	if len(api.config.Order) > 0 {
		api.removeUnknownServices()
		api.extractServiceStatus()
	}

	api.handler()
}

func (api *api) initialise() {
	databaseFile := api.config.Settings.DataDatabaseFile()
	checkFile(databaseFile)
	db, err := sql.Open("sqlite", databaseFile)
	jLog.Fatal(err, logFrom, err != nil)

	// Create the table
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS status (
			id                         TEXT     NOT NULL PRIMARY KEY,
			latest_version             TEXT     DEFAULT  '',
			latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			deployed_version           TEXT     DEFAULT  '',
			deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			approved_version           TEXT     DEFAULT  ''
		);`
	_, err = db.Exec(sqlStmt)
	jLog.Fatal(util.ErrorToString(err), logFrom, err != nil)

	updateTable(db)

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
		logFrom,
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
	jLog.Fatal(err, logFrom, err != nil)
	defer rows.Close()

	api.config.OrderMutex.RLock()
	defer api.config.OrderMutex.RUnlock()
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
			logFrom,
			err != nil)
		api.config.Service[id].Status.SetLatestVersion(lv, false)
		api.config.Service[id].Status.SetLatestVersionTimestamp(lvt)
		api.config.Service[id].Status.SetDeployedVersion(dv, false)
		api.config.Service[id].Status.SetDeployedVersionTimestamp(dvt)
		api.config.Service[id].Status.SetApprovedVersion(av, false)
	}
	err = rows.Err()
	jLog.Fatal(
		fmt.Sprintf("extractServiceStatus: %s", util.ErrorToString(err)),
		logFrom,
		err != nil)
}

// updateTable will update the table for the latest version
func updateTable(db *sql.DB) {
	// Get the type of the *_version columns
	var columnType string
	err := db.QueryRow("SELECT type FROM pragma_table_info('status') WHERE name = 'latest_version'").Scan(&columnType)
	jLog.Fatal(fmt.Sprintf("updateTable: %s", util.ErrorToString(err)), logFrom, err != nil)
	// Update if the column type is not TEXT
	if columnType != "TEXT" {
		jLog.Verbose("Updating column types", logFrom, true)
		updateColumnTypes(db)
		jLog.Verbose("Finished updating column types", logFrom, true)
	}
}

// updateColumnTypes will recreate the table with the correct column types
func updateColumnTypes(db *sql.DB) {
	// Create the new table
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS status_backup (
			id                         TEXT     NOT NULL PRIMARY KEY,
			latest_version             TEXT     DEFAULT  '',
			latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			deployed_version           TEXT     DEFAULT  '',
			deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			approved_version           TEXT     DEFAULT  ''
		);`
	_, err := db.Exec(sqlStmt)
	jLog.Fatal(fmt.Sprintf("updateColumnTypes - create: %s", util.ErrorToString(err)), logFrom, err != nil)

	// Copy the data from the old table to the new table
	_, err = db.Exec(`INSERT INTO status_backup SELECT * FROM status;`)
	jLog.Fatal(fmt.Sprintf("updateColumnTypes - copy: %s", util.ErrorToString(err)), logFrom, err != nil)

	// Drop the table
	_, err = db.Exec("DROP TABLE status;")
	jLog.Fatal(fmt.Sprintf("updateColumnTypes - drop: %s", util.ErrorToString(err)), logFrom, err != nil)

	// Rename the new table to the old table
	_, err = db.Exec("ALTER TABLE status_backup RENAME TO status;")
	jLog.Fatal(fmt.Sprintf("updateColumnTypes - rename: %s", util.ErrorToString(err)), logFrom, err != nil)
}
