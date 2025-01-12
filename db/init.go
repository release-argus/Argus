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

// Package db provides database functionality for Argus to keep track of versions found/deployed/approved.
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
	// Only set the log if it has not been set (avoid RACE condition).
	if jLog == nil {
		jLog = log
		logFrom = util.LogFrom{Primary: "db", Secondary: databaseFile}
	}
}

func checkFile(path string) {
	file := filepath.Base(path)
	// Check the directory exists.
	dir := filepath.Dir(path)
	fileInfo, err := os.Stat(dir)
	if err != nil {
		// Directory doesn't exist.
		if os.IsNotExist(err) {
			// Create the dir.
			if err := os.MkdirAll(dir, 0_750); err != nil {
				jLog.Fatal(err, logFrom, true)
			}
		} else {
			// other error.
			jLog.Fatal(err, logFrom, true)
		}

		// Directory exists, but is not a directory.
	} else if fileInfo == nil || !fileInfo.IsDir() {
		jLog.Fatal(
			fmt.Sprintf("path %q (for %q) is not a directory",
				dir, file),
			logFrom, true)
	}

	// Check the file exists.
	fileInfo, err = os.Stat(path)
	if err != nil {
		// File doesn't exist.
		jLog.Fatal(err, logFrom, os.IsExist(err))

		// Item exists, but is a directory.
	} else if fileInfo != nil && fileInfo.IsDir() {
		jLog.Fatal(
			fmt.Sprintf("path %q (for %q) is a directory, not a file",
				path, file),
			logFrom, true)
	}
}

// Run will start the database, initialise it and run the handler for messages in the background.
func Run(cfg *config.Config, log *util.JLog) {
	if log != nil {
		LogInit(log, cfg.Settings.DataDatabaseFile())
	}
	api := api{config: cfg}

	api.initialise()
	runningHandler := false
	defer func() {
		if !runningHandler {
			api.db.Close()
		}
	}()

	if len(api.config.Order) > 0 {
		api.removeUnknownServices()
		api.extractServiceStatus()
	}

	go api.handler()
	runningHandler = true
}

func (api *api) initialise() {
	databaseFile := api.config.Settings.DataDatabaseFile()
	checkFile(databaseFile)
	db, err := sql.Open("sqlite", databaseFile)
	jLog.Fatal(err, logFrom, err != nil)

	// Create the table.
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS status (
			id                         TEXT     NOT NULL PRIMARY KEY,
			latest_version             TEXT     DEFAULT  '',
			latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			deployed_version           TEXT     DEFAULT  '',
			deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			approved_version           TEXT     DEFAULT  ''
		);`
	if _, err := db.Exec(sqlStmt); err != nil {
		jLog.Fatal(err, logFrom, true)
	}

	updateTable(db)

	api.db = db
}

// removeUnknownServices will remove rows with an ID not in config.Order.
func (api *api) removeUnknownServices() {
	// ? for each service.
	servicePlaceholders := strings.Repeat("?,", len(api.config.Order))
	servicePlaceholders = servicePlaceholders[:len(servicePlaceholders)-1]

	// SQL statement to remove unknown services.
	//#nosec G201 -- servicePlaceholders is safe.
	sqlStmt := fmt.Sprintf(`
		DELETE FROM status
		WHERE id NOT IN (%s);`,
		servicePlaceholders)

	// Get the vars for the SQL statement.
	params := make([]interface{}, len(api.config.Order))
	for i, name := range api.config.Order {
		params[i] = name
	}

	if _, err := api.db.Exec(sqlStmt, params...); err != nil {
		jLog.Fatal(
			fmt.Sprintf("removeUnknownServices: %s",
				err),
			logFrom, true)
	}
}

// extractServiceStatus will query the database and add the data found
// into the Service.Status inside the config.
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
		if err := rows.Scan(&id, &lv, &lvt, &dv, &dvt, &av); err != nil {
			jLog.Fatal(
				fmt.Sprintf("extractServiceStatus row: %s",
					err),
				logFrom, true)
		}
		api.config.Service[id].Status.SetLatestVersion(lv, lvt, false)
		api.config.Service[id].Status.SetDeployedVersion(dv, dvt, false)
		api.config.Service[id].Status.SetApprovedVersion(av, false)
	}
	if err := rows.Err(); err != nil {
		jLog.Fatal(
			fmt.Sprintf("extractServiceStatus: %s",
				err),
			logFrom, true)
	}
}

// updateTable will update the table for the latest version.
func updateTable(db *sql.DB) {
	// Get the type of the *_version columns.
	var columnType string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('status') WHERE name = 'latest_version'").Scan(&columnType); err != nil {
		jLog.Fatal(
			fmt.Sprintf("updateTable: %s",
				err),
			logFrom, true)
	}
	// Update if the column type is not TEXT.
	if columnType != "TEXT" {
		jLog.Verbose("Updating column types", logFrom, true)
		updateColumnTypes(db)
		jLog.Verbose("Finished updating column types", logFrom, true)
	}
}

// updateColumnTypes will recreate the table with the correct column types.
func updateColumnTypes(db *sql.DB) {
	// Create the new table.
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS status_backup (
			id                         TEXT     NOT NULL PRIMARY KEY,
			latest_version             TEXT     DEFAULT  '',
			latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			deployed_version           TEXT     DEFAULT  '',
			deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			approved_version           TEXT     DEFAULT  ''
		);`
	if _, err := db.Exec(sqlStmt); err != nil {
		jLog.Fatal(
			fmt.Sprintf("updateColumnTypes - create: %s", err),
			logFrom, true)
	}

	// Copy the data from the old table to the new table.
	if _, err := db.Exec(`INSERT INTO status_backup SELECT * FROM status;`); err != nil {
		jLog.Fatal(
			fmt.Sprintf("updateColumnTypes - copy: %s",
				err),
			logFrom, true)
	}

	// Drop the table.
	if _, err := db.Exec("DROP TABLE status;"); err != nil {
		jLog.Fatal(
			fmt.Sprintf("updateColumnTypes - drop: %s",
				err),
			logFrom, true)
	}

	// Rename the new table to the old table.
	if _, err := db.Exec("ALTER TABLE status_backup RENAME TO status;"); err != nil {
		jLog.Fatal(
			fmt.Sprintf("updateColumnTypes - rename: %s",
				err),
			logFrom, true)
	}
}
