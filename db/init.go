// Copyright [2026] [Argus]
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
	"sync"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
)

var once sync.Once

// openDatabase opens the SQLite database (overridable for tests).
var openDatabase = sql.Open

// serviceStatusRowsErr reports iteration errors from a status query (overridable for tests).
var serviceStatusRowsErr = func(rows *sql.Rows) error {
	return rows.Err()
}

// Get opens the database, loads persisted status, and returns the API handle.
func Get(cfg *config.Config) *api {
	once.Do(func() {
		if logFrom.Secondary == "" {
			logFrom.Secondary = cfg.Settings.DataDatabaseFile()
		}
	})

	api := api{config: cfg}
	if ok := api.initialise(); !ok {
		return nil
	}

	if len(api.config.Order) > 0 {
		if ok := api.removeUnknownServices(); !ok {
			return nil
		}
		if ok := api.extractServiceStatus(); !ok {
			return nil
		}
	}

	return &api
}

// initialise opens the SQLite database and ensures the status table exists.
func (api *api) initialise() (ok bool) {
	databaseFile := api.config.Settings.DataDatabaseFile()
	if ok := checkFile(databaseFile); !ok {
		return ok
	}
	db, err := openDatabase("sqlite", databaseFile)
	if err != nil {
		logx.Fatal(err, logFrom)
		return
	}

	// Create the table.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS status (
			id                         TEXT     NOT NULL PRIMARY KEY,
			latest_version             TEXT     DEFAULT  '',
			latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			deployed_version           TEXT     DEFAULT  '',
			deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			approved_version           TEXT     DEFAULT  ''
	);`); err != nil {
		logx.Fatal(err, logFrom)
		return
	}

	ok = updateTable(db)

	api.db = db
	return
}

// removeUnknownServices deletes status rows whose IDs are not in config.Order.
func (api *api) removeUnknownServices() bool {
	// ? for each service.
	servicePlaceholders := strings.Repeat("?,", len(api.config.Order))
	servicePlaceholders = strings.TrimSuffix(servicePlaceholders, ",")

	// SQL statement to remove unknown services.
	//#nosec G201 -- servicePlaceholders is safe.
	sqlStmt := fmt.Sprintf(`
		DELETE FROM status
		WHERE id NOT IN (%s);`,
		servicePlaceholders,
	)

	// Get the vars for the SQL statement.
	params := make([]any, len(api.config.Order))
	for i, name := range api.config.Order {
		params[i] = name
	}

	if _, err := api.db.Exec(sqlStmt, params...); err != nil {
		logx.Fatal(fmt.Sprintf("removeUnknownServices: %s", err), logFrom)
		return false
	}

	return true
}

// extractServiceStatus loads persisted version data into each service's Status.
func (api *api) extractServiceStatus() (ok bool) {
	rows, err := api.db.Query(`
		SELECT
			id,
			latest_version,
			latest_version_timestamp,
			deployed_version,
			deployed_version_timestamp,
			approved_version
		FROM status;`,
	)
	if err != nil {
		logx.Fatal(err, logFrom)
		return
	}
	defer rows.Close()

	api.config.OrderMu.RLock()
	defer api.config.OrderMu.RUnlock()
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
			logx.Fatal(fmt.Sprintf("extractServiceStatus row: %s", err), logFrom)
			return
		}
		svc := api.config.Service[id]
		svc.Status.SetLatestVersion(lv, lvt, false)
		svc.Status.SetDeployedVersion(dv, dvt, false)
		svc.Status.SetApprovedVersion(av, false)
	}
	if err := serviceStatusRowsErr(rows); err != nil {
		logx.Fatal(
			fmt.Sprintf("extractServiceStatus: %s", err),
			logFrom,
		)
		return
	}

	return true
}

// checkFile ensures the database directory exists and path is not a directory.
func checkFile(path string) (ok bool) {
	file := filepath.Base(path)

	// Check the directory exists.
	dir := filepath.Dir(path)
	fileInfo, err := os.Stat(dir)
	if err != nil {
		// Directory doesn't exist.
		if os.IsNotExist(err) {
			// Create the dir.
			if err := os.MkdirAll(dir, 0_750); err != nil {
				logx.Fatal(err, logFrom)
				return
			}
		} else {
			// Other error.
			logx.Fatal(err, logFrom)
			return
		}

		// Something exists, but not a directory.
	} else if fileInfo == nil || !fileInfo.IsDir() {
		logx.Fatal(
			fmt.Sprintf(
				"path %q (for %q) is not a directory",
				dir, file,
			),
			logFrom,
		)
		return
	}

	// Check the file exists.
	fileInfo, err = os.Stat(path)
	if err != nil {
		// File doesn't exist.
		if os.IsNotExist(err) {
			return true
		}
		// Other errors accessing the file.
		logx.Fatal(err, logFrom)
		return

		// Item exists, but is a directory.
	} else if fileInfo != nil && fileInfo.IsDir() {
		logx.Fatal(
			fmt.Sprintf(
				"path %q (for %q) is a directory, not a file",
				path, file,
			),
			logFrom,
		)
		return
	}

	return true
}

// updateTable migrates the status table when legacy column types are detected.
func updateTable(db *sql.DB) bool {
	// Get the type of the *_version columns.
	var columnType string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('status') WHERE name = 'latest_version'").Scan(&columnType); err != nil {
		logx.Fatal(fmt.Sprintf("updateTable: %s", err), logFrom)
		return false
	}
	// Update if the column type is not TEXT.
	if columnType != "TEXT" {
		logx.Verbose("Updating column types", logFrom, true)
		if ok := updateColumnTypes(db); !ok {
			return false
		}
		logx.Verbose("Finished updating column types", logFrom, true)
	}

	return true
}

// updateColumnTypes recreates status with TEXT version columns and copies existing rows.
func updateColumnTypes(db *sql.DB) (ok bool) {
	// Create the new table.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS status_backup (
			id                         TEXT     NOT NULL PRIMARY KEY,
			latest_version             TEXT     DEFAULT  '',
			latest_version_timestamp   DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			deployed_version           TEXT     DEFAULT  '',
			deployed_version_timestamp DATETIME DEFAULT  (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			approved_version           TEXT     DEFAULT  ''
		);`); err != nil {
		logx.Fatal(fmt.Sprintf("updateColumnTypes - create: %s", err), logFrom)
		return
	}

	// Copy the data from the old table to the new table.
	if _, err := db.Exec(`INSERT INTO status_backup SELECT * FROM status;`); err != nil {
		logx.Fatal(fmt.Sprintf("updateColumnTypes - copy: %s", err), logFrom)
		return
	}

	// Drop the table.
	if _, err := db.Exec("DROP TABLE status;"); err != nil {
		logx.Fatal(fmt.Sprintf("updateColumnTypes - drop: %s", err), logFrom)
		return
	}

	// Rename the new table to the old table.
	if _, err := db.Exec("ALTER TABLE status_backup RENAME TO status;"); err != nil {
		logx.Fatal(fmt.Sprintf("updateColumnTypes - rename: %s", err), logFrom)
		return
	}

	return true
}
