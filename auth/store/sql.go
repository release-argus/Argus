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

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"modernc.org/sqlite"
)

// timeFormat is how timestamps are written (TEXT, UTC). Fixed-width fractional
// seconds keep TEXT order chronological.
// [time.RFC3339Nano] trims trailing zeros and so loses this ordering.
const timeFormat = "2006-01-02T15:04:05.000000000Z07:00"

// timeParseFormat accepts any number of fractional digits, so timestamps
// written with either layout load.
const timeParseFormat = time.RFC3339Nano

// sqliteConstraintUnique is SQLITE_CONSTRAINT_UNIQUE, the result code for a
// UNIQUE-constraint failure.
const sqliteConstraintUnique = 2067

// isUniqueViolation reports whether err is an SQLite UNIQUE-constraint failure.
func isUniqueViolation(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqliteConstraintUnique
}

// scanner abstracts [sql.Row]/[sql.Rows] for the scan* helpers.
type scanner interface {
	Scan(dest ...any) error
}

// rowsErr reports iteration errors from a query (overridable for tests).
// see [sql.Rows.Err].
var rowsErr = func(rows *sql.Rows) error {
	return rows.Err()
}

// scanAll collects every row of rows via scan.
// what names the rows in the iteration error, e.g. "users".
// The result is non-nil, so an empty set marshals as [] rather than null.
func scanAll[T any](
	rows *sql.Rows,
	what string,
	scan func(scanner) (*T, error),
) ([]T, error) {
	items := []T{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rowsErr(rows); err != nil {
		return nil, fmt.Errorf("iterate %s: %w", what, err)
	}

	return items, nil
}

// scanColumn collects the first column of every row of rows.
// what names the rows in the scan/iteration errors, e.g. "group names".
func scanColumn[T any](rows *sql.Rows, what string) ([]T, error) {
	var values []T
	for rows.Next() {
		var value T
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("scan %s: %w", what, err)
		}
		values = append(values, value)
	}
	if err := rowsErr(rows); err != nil {
		return nil, fmt.Errorf("iterate %s: %w", what, err)
	}

	return values, nil
}

// inTx runs fn inside a transaction, committing on nil and
// rolling back on error.
func (s *Store) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// parseTime parses a stored timestamp.
func parseTime(value string) (time.Time, error) {
	t, err := time.Parse(timeParseFormat, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse stored timestamp %q: %w", value, err)
	}
	return t, nil
}

// timeText is an [sql.Scanner] that parses a stored timestamp string straight
// into a [time.Time].
type timeText struct{ dst *time.Time }

// Scan implements the sql.Scanner interface.
func (t timeText) Scan(src any) error {
	switch v := src.(type) {
	case string:
		parsed, err := parseTime(v)
		if err != nil {
			return err
		}
		*t.dst = parsed
	case time.Time:
		*t.dst = v
	default:
		return fmt.Errorf("timestamp: unsupported source %T", src)
	}
	return nil
}

// nullTimeText is the nullable counterpart of [timeText]:
// a NULL column leaves the target nil.
type nullTimeText struct{ dst **time.Time }

// Scan implements the sql.Scanner interface.
func (t nullTimeText) Scan(src any) error {
	if src == nil {
		*t.dst = nil
		return nil
	}
	var parsed time.Time
	if err := (timeText{&parsed}).Scan(src); err != nil {
		return err
	}
	*t.dst = &parsed
	return nil
}

// rowUpdate accumulates "col = ?" assignments (and their values) for an UPDATE.
type rowUpdate struct {
	sets []string
	args []any
}

// newRowUpdate starts an update that bumps updated_at to now.
func newRowUpdate() *rowUpdate {
	return &rowUpdate{
		sets: []string{"updated_at = ?"},
		args: []any{timeNow().Format(timeFormat)},
	}
}

// set adds "column = ?" with its value.
func (u *rowUpdate) set(column string, value any) {
	u.sets = append(u.sets, column+" = ?")
	u.args = append(u.args, value)
}

// exec runs "UPDATE <table> SET <assignments> WHERE id = ?" within tx.
func (u *rowUpdate) exec(ctx context.Context, tx *sql.Tx, table, id string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE `+table+` SET `+strings.Join(u.sets, ", ")+` WHERE id = ?;`,
		append(u.args, id)...,
	)
	return err //nolint:wrapcheck
}
