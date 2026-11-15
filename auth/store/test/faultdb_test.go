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
	"testing"

	"modernc.org/sqlite"
)

func TestFaultDB(t *testing.T) {
	db, state := FaultDB(t)
	if _, err := db.Exec(`CREATE TABLE things (id TEXT);`); err != nil {
		t.Fatalf("DB.Exec() failed: %v", err)
	}

	state.Set(`INSERT INTO things`)

	if _, err := db.Exec(`INSERT INTO things (id) VALUES ('x');`); !errors.Is(err, ErrInjected) {
		t.Errorf("armed: DB.Exec() error type mismatch\ngot:  %v\nwant: %v", err, ErrInjected)
	}

	if _, err := db.Exec(`DELETE FROM things;`); err != nil {
		t.Errorf("unmatched: DB.Exec() failed: %v", err)
	}

	state.Set(`SELECT id FROM things`)
	if _, err := db.Query(`SELECT id FROM things;`); !errors.Is(err, ErrInjected) {
		t.Errorf("armed: DB.Query() error type mismatch\ngot:  %v\nwant: %v", err, ErrInjected)
	}

	state.Set("")

	if _, err := db.Exec(`INSERT INTO things (id) VALUES ('x');`); err != nil {
		t.Errorf("disarmed: DB.Exec() failed: %v", err)
	}

	state.SetRowsAffected(`DELETE FROM things`)

	result, err := db.Exec(`DELETE FROM things;`)
	if err != nil {
		t.Fatalf("RowsAffected-armed: DB.Exec() failed: %v", err)
	}
	if _, err := result.RowsAffected(); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"RowsAffected-armed: Result.RowsAffected() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInjected,
		)
	}
	state.Set("")
}

func TestFaultDB_DelegatePaths(t *testing.T) {
	db, _ := FaultDB(t)
	if _, err := db.Exec(`CREATE TABLE things (id TEXT);`); err != nil {
		t.Fatalf("DB.Exec() failed: %v", err)
	}

	statement, err := db.Prepare(`INSERT INTO things (id) VALUES (?);`)
	if err != nil {
		t.Fatalf("DB.Prepare() failed: %v", err)
	}
	if _, err := statement.Exec("bound"); err != nil {
		t.Errorf("Stmt.Exec() failed: %v", err)
	}
	_ = statement.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("DB.Begin() failed: %v", err)
	}
	if _, err := tx.Exec(`INSERT INTO things (id) VALUES (?);`, "in-tx"); err != nil {
		t.Errorf("Tx.Exec() failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Errorf("Tx.Commit() failed: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM things;`).Scan(&count); err != nil || count != 2 {
		t.Errorf("things row count mismatch\ngot:  %d, err=%v\nwant: 2, err=nil", count, err)
	}
}

func TestFaultConnector(t *testing.T) {
	state := &FaultState{}
	connector := &faultConnector{
		dsn:   "file:" + t.TempDir() + "?mode=ro", // A directory, so it cannot be opened.
		d:     &sqlite.Driver{},
		state: state,
	}

	_, err := connector.Connect(t.Context())
	if err == nil {
		t.Error("faultConnector.Connect() error mismatch\ngot:  nil\nwant: non-nil")
	}

	if connector.Driver() == nil {
		t.Error("faultConnector.Driver() mismatch\ngot:  nil\nwant: non-nil")
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
	inner := &stubConn{}
	state := &FaultState{}
	conn := &faultConn{inner: inner, state: state}
	ctx := t.Context()

	// Missing optional interfaces skip, so database/sql falls back.
	if _, err := conn.ExecContext(ctx, `EXEC;`, nil); !errors.Is(err, driver.ErrSkip) {
		t.Errorf(
			"unarmed: faultConn.ExecContext() error type mismatch\ngot:  %v\nwant: %v",
			err, driver.ErrSkip,
		)
	}
	if _, err := conn.QueryContext(ctx, `QUERY;`, nil); !errors.Is(err, driver.ErrSkip) {
		t.Errorf(
			"unarmed: faultConn.QueryContext() error type mismatch\ngot:  %v\nwant: %v",
			err, driver.ErrSkip,
		)
	}

	if _, err := conn.PrepareContext(ctx, `PREPARE;`); err == nil || inner.prepared != `PREPARE;` {
		t.Errorf(
			"unarmed: faultConn.PrepareContext() mismatch\n"+
				"got:  err=%v, prepared=%q\nwant: err=non-nil, prepared=%q",
			err, inner.prepared, `PREPARE;`,
		)
	}

	// BeginTx falls back to Begin.
	if _, err := conn.BeginTx(ctx, driver.TxOptions{}); err == nil {
		t.Error("unarmed: faultConn.BeginTx() error mismatch\ngot:  nil\nwant: non-nil")
	}
	if _, err := conn.Begin(); err == nil {
		t.Error("unarmed: faultConn.Begin() error mismatch\ngot:  nil\nwant: non-nil")
	}

	// ErrSkip lets the default converter run.
	if err := conn.CheckNamedValue(&driver.NamedValue{}); !errors.Is(err, driver.ErrSkip) {
		t.Errorf(
			"unarmed: faultConn.CheckNamedValue() error type mismatch\ngot:  %v\nwant: %v",
			err, driver.ErrSkip,
		)
	}

	if err := conn.Close(); err != nil {
		t.Errorf("unarmed: faultConn.Close() failed: %v", err)
	}

	if _, err := conn.Prepare(`PASSTHROUGH;`); err == nil || inner.prepared != `PASSTHROUGH;` {
		t.Errorf(
			"unarmed: faultConn.Prepare() mismatch\n"+
				"got:  err=%v, prepared=%q\nwant: err=non-nil, prepared=%q",
			err, inner.prepared, `PASSTHROUGH;`,
		)
	}

	state.Set(`PREPARE;`)
	_, err := conn.Prepare(`PREPARE;`)
	if !errors.Is(err, ErrInjected) {
		t.Errorf(
			"armed: faultConn.Prepare() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInjected,
		)
	}
	_, err = conn.PrepareContext(ctx, `PREPARE;`)
	if !errors.Is(err, ErrInjected) {
		t.Errorf(
			"armed: faultConn.PrepareContext() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInjected,
		)
	}
	state.Set(`EXEC;`)
	_, err = conn.ExecContext(ctx, `EXEC;`, nil)
	if !errors.Is(err, ErrInjected) {
		t.Errorf(
			"armed: faultConn.ExecContext() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInjected,
		)
	}
	state.Set(`QUERY;`)
	_, err = conn.QueryContext(ctx, `QUERY;`, nil)
	if !errors.Is(err, ErrInjected) {
		t.Errorf(
			"armed: faultConn.QueryContext() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInjected,
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

func TestFaultConn_CheckNamedValue__delegates(t *testing.T) {
	inner := &checkerConn{}
	conn := &faultConn{inner: inner, state: &FaultState{}}

	err := conn.CheckNamedValue(&driver.NamedValue{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !inner.checked {
		t.Error("checkerConn.checked mismatch\ngot:  false\nwant: true")
	}
}

// stubResult is a driver.Result with fixed answers.
type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 42, nil }
func (stubResult) RowsAffected() (int64, error) { return 1, nil }

func TestFaultResult(t *testing.T) {
	result := faultResult{inner: stubResult{}}

	if id, err := result.LastInsertId(); err != nil || id != 42 {
		t.Errorf(
			"faultResult.LastInsertId() mismatch\ngot:  %d, err=%v\nwant: 42, err=nil",
			id, err,
		)
	}
	if _, err := result.RowsAffected(); !errors.Is(err, ErrInjected) {
		t.Errorf(
			"faultResult.RowsAffected() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInjected,
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
