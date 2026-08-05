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

// Package test provides fault-injection database fixtures for testing code
// built on the auth store: an in-memory SQLite database whose statements
// can be failed selectively by SQL substring.
package test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

// ErrInjected is the error produced by the fault-injecting driver.
var ErrInjected = errors.New("injected fault")

// FaultState selects which statements the fault driver fails.
type FaultState struct {
	mu               sync.Mutex
	pattern          string // Substring of the SQL to fail; empty fails nothing.
	failRowsAffected bool   // Fail RowsAffected of matching statements instead.
}

// Set arms the fault for statements containing pattern
// (empty disarms the fault).
func (f *FaultState) Set(pattern string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pattern = pattern
	f.failRowsAffected = false
}

// SetRowsAffected arms a RowsAffected fault for statements containing pattern.
func (f *FaultState) SetRowsAffected(pattern string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pattern = pattern
	f.failRowsAffected = true
}

// failing reports whether query should fail, and how.
func (f *FaultState) failing(query string) (statement, rowsAffected bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.pattern == "" || !strings.Contains(query, f.pattern) {
		return false, false
	}
	return !f.failRowsAffected, f.failRowsAffected
}

// FaultDB opens an in-memory SQLite database whose statements can be failed
// selectively via the returned FaultState.
func FaultDB(t *testing.T) (*sql.DB, *FaultState) {
	t.Helper()

	state := &FaultState{}
	db := sql.OpenDB(&faultConnector{
		dsn:   ":memory:",
		d:     sqliteDriver(t),
		state: state,
	})

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)
	t.Cleanup(func() { _ = db.Close() })

	return db, state
}

// sqliteDriver returns the registered modernc SQLite driver.
func sqliteDriver(t *testing.T) driver.Driver {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("auth_store_test\nopen for driver lookup: %v", err)
	}
	d := db.Driver()
	_ = db.Close()
	return d
}

// faultConnector produces fault-injecting connections to a fixed DSN.
type faultConnector struct {
	dsn   string // Data Source Name.
	d     driver.Driver
	state *FaultState
}

func (c *faultConnector) Connect(_ context.Context) (driver.Conn, error) {
	conn, err := c.d.Open(c.dsn)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return &faultConn{inner: conn, state: c.state}, nil
}

func (c *faultConnector) Driver() driver.Driver { return c.d }

// faultConn wraps a [driver.Conn], failing statements per its FaultState.
type faultConn struct {
	inner driver.Conn
	state *FaultState
}

// Prepare intercepts statements matching configured failures.
func (c *faultConn) Prepare(query string) (driver.Stmt, error) {
	if statement, _ := c.state.failing(query); statement {
		return nil, ErrInjected
	}
	return c.inner.Prepare(query) //nolint:wrapcheck
}

// PrepareContext intercepts statements matching configured failures.
func (c *faultConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if statement, _ := c.state.failing(query); statement {
		return nil, ErrInjected
	}
	if p, ok := c.inner.(driver.ConnPrepareContext); ok {
		return p.PrepareContext(ctx, query) //nolint:wrapcheck
	}
	return c.inner.Prepare(query) //nolint:wrapcheck
}

// ExecContext intercepts configured failures and optionally injects result errors.
func (c *faultConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	statement, rowsAffected := c.state.failing(query)
	if statement {
		return nil, ErrInjected
	}
	if e, ok := c.inner.(driver.ExecerContext); ok {
		result, err := e.ExecContext(ctx, query, args)
		if err == nil && rowsAffected {
			result = faultResult{inner: result}
		}
		return result, err //nolint:wrapcheck
	}
	return nil, driver.ErrSkip
}

// QueryContext intercepts queries matching configured failures.
func (c *faultConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if statement, _ := c.state.failing(query); statement {
		return nil, ErrInjected
	}
	if q, ok := c.inner.(driver.QueryerContext); ok {
		return q.QueryContext(ctx, query, args) //nolint:wrapcheck
	}
	return nil, driver.ErrSkip
}

// BeginTx delegates transaction creation to the underlying connection.
func (c *faultConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if b, ok := c.inner.(driver.ConnBeginTx); ok {
		return b.BeginTx(ctx, opts) //nolint:wrapcheck
	}
	return c.inner.Begin() //nolint:staticcheck,wrapcheck
}

// Begin delegates transaction creation to the underlying connection.
func (c *faultConn) Begin() (driver.Tx, error) {
	return c.inner.Begin() //nolint:staticcheck,wrapcheck
}

// Close closes the underlying connection.
func (c *faultConn) Close() error { return c.inner.Close() } //nolint:wrapcheck

// CheckNamedValue delegates named value validation to the underlying connection.
func (c *faultConn) CheckNamedValue(nv *driver.NamedValue) error {
	if n, ok := c.inner.(driver.NamedValueChecker); ok {
		return n.CheckNamedValue(nv) //nolint:wrapcheck
	}
	return driver.ErrSkip
}

// faultResult delegates LastInsertId but fails RowsAffected.
type faultResult struct {
	inner driver.Result
}

// LastInsertId delegates to the underlying result.
func (r faultResult) LastInsertId() (int64, error) {
	return r.inner.LastInsertId() //nolint:wrapcheck
}

// RowsAffected injects a configured failure.
func (faultResult) RowsAffected() (int64, error) {
	return 0, ErrInjected
}
