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

//go:build unit || integration

package test

import (
	"database/sql"
	"fmt"
	"testing"
)

// DropTable drops a table to force errors on statements touching it.
func DropTable(t TB, db *sql.DB, table string) {
	t.Helper()

	//#nosec G201 -- table names come from the tests themselves.
	stmt := fmt.Sprintf(`DROP TABLE %s;`, table)
	if _, err := db.ExecContext(t.Context(), stmt); err != nil {
		t.Fatalf("DB.ExecContext(%q) failed: %v", stmt, err)
	}
}

// ExecSQL runs each statement in turn, failing on the first that errors.
func ExecSQL(t TB, db *sql.DB, statements ...string) {
	t.Helper()

	for i, statement := range statements {
		if _, err := db.ExecContext(t.Context(), statement); err != nil {
			t.Fatalf("statement %d: DB.ExecContext(%q) failed: %v", i, statement, err)
		}
	}
}

// MemoryDB opens a fresh in-memory SQLite database, closed on test cleanup.
// The caller's package must import the sqlite driver.
func MemoryDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() failed: %v", err)
	}
	// In-memory SQLite gives each connection its own database.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	return db
}

// CountRows counts the rows of a table.
func CountRows(t TB, db *sql.DB, table string) int {
	t.Helper()

	var count int
	//#nosec G201 -- table names come from the tests themselves.
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, table)
	if err := db.QueryRowContext(t.Context(), query).Scan(&count); err != nil {
		t.Fatalf("DB.QueryRowContext(%q) failed: %v", query, err)
	}
	return count
}
