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

//go:build unit

package test

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
)

var packageName = "auth_store_test"

func TestFaultDB(t *testing.T) {
	// GIVEN: a fault-injecting database.
	db, state := FaultDB(t)
	if _, err := db.Exec(`CREATE TABLE things (id TEXT);`); err != nil {
		t.Fatalf(
			"%s\nsetup failed: %v",
			packageName, err,
		)
	}

	// WHEN: a statement fault is armed.
	state.Set(`INSERT INTO things`)

	prefix := fmt.Sprintf("%s\nFaultDB()", packageName)

	// THEN: matching statements fail with ErrInjected.
	if _, err := db.Exec(`INSERT INTO things (id) VALUES ('x');`); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s armed statement should fail\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}
	// AND: non-matching statements pass.
	if _, err := db.Exec(`DELETE FROM things;`); err != nil {
		t.Errorf(
			"%s unmatched statement should pass\ngot: %v",
			prefix, err,
		)
	}
	// AND: matching queries fail too.
	state.Set(`SELECT id FROM things`)
	if _, err := db.Query(`SELECT id FROM things;`); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s armed query should fail\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}

	// WHEN: the fault is disarmed.
	state.Set("")

	// THEN: everything passes again.
	if _, err := db.Exec(`INSERT INTO things (id) VALUES ('x');`); err != nil {
		t.Errorf(
			"%s disarmed statement should pass\ngot: %v",
			prefix, err,
		)
	}

	// WHEN: a RowsAffected fault is armed.
	state.SetRowsAffected(`DELETE FROM things`)

	// THEN: the statement runs, but RowsAffected fails.
	result, err := db.Exec(`DELETE FROM things;`)
	if err != nil {
		t.Fatalf(
			"%s RowsAffected-armed statement should run\ngot: %v",
			prefix, err,
		)
	}
	if _, err := result.RowsAffected(); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s RowsAffected should fail\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}
	state.Set("")
}

func TestFaultDB_DelegatePaths(t *testing.T) {
	// GIVEN: a fault-injecting database (fault disarmed).
	db, _ := FaultDB(t)
	if _, err := db.Exec(`CREATE TABLE things (id TEXT);`); err != nil {
		t.Fatalf(
			"%s\nsetup failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nFaultDB() delegate paths", packageName)

	// WHEN: statements are prepared, run in a transaction, and bind args.
	statement, err := db.Prepare(`INSERT INTO things (id) VALUES (?);`)
	if err != nil {
		t.Fatalf(
			"%s Prepare failed: %v",
			prefix, err,
		)
	}
	if _, err := statement.Exec("bound"); err != nil {
		t.Errorf(
			"%s prepared exec failed: %v",
			prefix, err,
		)
	}
	_ = statement.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf(
			"%s Begin failed: %v",
			prefix, err,
		)
	}
	if _, err := tx.Exec(`INSERT INTO things (id) VALUES (?);`, "in-tx"); err != nil {
		t.Errorf(
			"%s transactional exec failed: %v",
			prefix, err,
		)
	}
	if err := tx.Commit(); err != nil {
		t.Errorf(
			"%s Commit failed: %v",
			prefix, err,
		)
	}

	// THEN: both rows landed.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM things;`).Scan(&count); err != nil || count != 2 {
		t.Errorf(
			"%s row count mismatch\ngot:  %d, err=%v\nwant: 2",
			prefix, count, err,
		)
	}
}

func TestFaultConnector(t *testing.T) {
	// GIVEN: a connector to an unopenable DSN (a directory).
	state := &FaultState{}
	connector := &faultConnector{
		dsn:   "file:" + t.TempDir() + "?mode=ro",
		d:     sqliteDriver(t),
		state: state,
	}

	prefix := fmt.Sprintf("%s\nfaultConnector", packageName)

	// WHEN: a connection is attempted.
	if _, err := connector.Connect(t.Context()); err == nil {
		t.Errorf("%s Connect to a directory should fail", prefix)
	}

	// AND: Driver returns the wrapped driver.
	if connector.Driver() == nil {
		t.Errorf("%s Driver should return the inner driver", prefix)
	}
}

// stubConn is a driver.Conn implementing none of the optional interfaces,
// to exercise the fault wrapper's fallback paths.
type stubConn struct {
	prepared string // Last prepared query.
}

func (s *stubConn) Prepare(query string) (driver.Stmt, error) {
	s.prepared = query
	return nil, errors.New("stub: no statements")
}

func (s *stubConn) Close() error { return nil }

func (s *stubConn) Begin() (driver.Tx, error) { return nil, errors.New("stub: no transactions") }

func TestFaultConn_Fallbacks(t *testing.T) {
	// GIVEN: a fault wrapper around a minimal driver.Conn.
	inner := &stubConn{}
	state := &FaultState{}
	conn := &faultConn{inner: inner, state: state}
	ctx := t.Context()

	prefix := fmt.Sprintf("%s\nfaultConn fallbacks", packageName)

	// WHEN: the optional interfaces are missing.
	// THEN: ExecContext/QueryContext skip so database/sql falls back.
	if _, err := conn.ExecContext(ctx, `EXEC;`, nil); !errors.Is(err, driver.ErrSkip) {
		t.Errorf(
			"%s ExecContext\ngot:  %v\nwant: %v",
			prefix, err, driver.ErrSkip,
		)
	}
	if _, err := conn.QueryContext(ctx, `QUERY;`, nil); !errors.Is(err, driver.ErrSkip) {
		t.Errorf(
			"%s QueryContext\ngot:  %v\nwant: %v",
			prefix, err, driver.ErrSkip,
		)
	}
	// AND: PrepareContext falls back to Prepare.
	if _, err := conn.PrepareContext(ctx, `PREPARE;`); err == nil || inner.prepared != `PREPARE;` {
		t.Errorf(
			"%s PrepareContext should fall back to Prepare\ngot:  err=%v, prepared=%q",
			prefix, err, inner.prepared,
		)
	}
	// AND: BeginTx falls back to Begin.
	if _, err := conn.BeginTx(ctx, driver.TxOptions{}); err == nil {
		t.Errorf("%s BeginTx should fall back to the stub's Begin error", prefix)
	}
	if _, err := conn.Begin(); err == nil {
		t.Errorf("%s Begin should surface the stub's error", prefix)
	}
	// AND: CheckNamedValue skips so the default converter runs.
	if err := conn.CheckNamedValue(&driver.NamedValue{}); !errors.Is(err, driver.ErrSkip) {
		t.Errorf(
			"%s CheckNamedValue\ngot:  %v\nwant: %v",
			prefix, err, driver.ErrSkip,
		)
	}
	// AND: Close passes through.
	if err := conn.Close(); err != nil {
		t.Errorf(
			"%s Close\ngot:  %v",
			prefix, err,
		)
	}

	// AND: an unarmed Prepare passes through to the inner conn.
	if _, err := conn.Prepare(`PASSTHROUGH;`); err == nil || inner.prepared != `PASSTHROUGH;` {
		t.Errorf(
			"%s unarmed Prepare should pass through\ngot:  err=%v, prepared=%q",
			prefix, err, inner.prepared,
		)
	}

	// WHEN: faults are armed on the fallback paths.
	state.Set(`PREPARE;`)
	if _, err := conn.Prepare(`PREPARE;`); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s armed Prepare\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}
	if _, err := conn.PrepareContext(ctx, `PREPARE;`); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s armed PrepareContext\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}
	state.Set(`EXEC;`)
	if _, err := conn.ExecContext(ctx, `EXEC;`, nil); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s armed ExecContext\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}
	state.Set(`QUERY;`)
	if _, err := conn.QueryContext(ctx, `QUERY;`, nil); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s armed QueryContext\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}
}

// checkerConn is a stubConn that also implements driver.NamedValueChecker.
type checkerConn struct {
	stubConn
	checked bool
}

func (c *checkerConn) CheckNamedValue(_ *driver.NamedValue) error {
	c.checked = true
	return nil
}

func TestFaultConn_CheckNamedValueDelegates(t *testing.T) {
	// GIVEN: a fault wrapper around a conn that checks named values.
	inner := &checkerConn{}
	conn := &faultConn{inner: inner, state: &FaultState{}}

	prefix := fmt.Sprintf("%s\nCheckNamedValue delegation", packageName)

	// WHEN: a named value is checked.
	if err := conn.CheckNamedValue(&driver.NamedValue{}); err != nil {
		t.Errorf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: the inner conn's checker ran.
	if !inner.checked {
		t.Errorf("%s inner CheckNamedValue should have been called", prefix)
	}
}

// stubResult is a driver.Result with fixed answers.
type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 42, nil }
func (stubResult) RowsAffected() (int64, error) { return 1, nil }

func TestFaultResult(t *testing.T) {
	// GIVEN: a fault result wrapping a healthy one.
	result := faultResult{inner: stubResult{}}

	prefix := fmt.Sprintf("%s\nfaultResult", packageName)

	// WHEN/THEN: LastInsertId delegates; RowsAffected fails.
	if id, err := result.LastInsertId(); err != nil || id != 42 {
		t.Errorf(
			"%s LastInsertId should delegate\ngot:  %d, err=%v",
			prefix, id, err,
		)
	}
	if _, err := result.RowsAffected(); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"%s RowsAffected should fail\ngot:  %v\nwant: %v",
			prefix, err, ErrInjected,
		)
	}
}

// Compile-time interface checks for the wrapper.
var (
	_ driver.Conn               = (*faultConn)(nil)
	_ driver.ConnPrepareContext = (*faultConn)(nil)
	_ driver.ConnBeginTx        = (*faultConn)(nil)
	_ driver.ExecerContext      = (*faultConn)(nil)
	_ driver.QueryerContext     = (*faultConn)(nil)
	_ driver.NamedValueChecker  = (*faultConn)(nil)
)
