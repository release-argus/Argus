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

// Package store persists the authentication and authorisation entities
// (users, groups, permission grants, sessions, API tokens) in the
// Argus SQLite database, and owns their schema migrations.
//
// Deletions cascade explicitly inside transactions rather than relying on
// SQLite foreign-key enforcement (which is off by default, per connection).
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/internal/logx"
)

// Sentinel errors of the store.
var (
	ErrNotFound       = errors.New("not found")
	ErrUsernameTaken  = errors.New("username already taken")
	ErrGroupNameTaken = errors.New("group name already taken")
	ErrUnknownGroup   = errors.New("unknown group")
	ErrInvalidGrant   = errors.New("invalid grant")
	ErrLastAdmin      = errors.New("cannot delete, disable or demote the last enabled admin")
	ErrSystemGroup    = errors.New("system group cannot be modified this way")
	ErrSetupComplete  = errors.New("setup already completed")
)

// isUniqueViolation reports whether err is a SQLite UNIQUE-constraint failure.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// Well-known seeded group names.
const (
	GroupAdmin    = "admin"
	GroupOperator = "operator"
	GroupViewer   = "viewer"
)

// timeFormat is how timestamps are stored (TEXT, UTC).
const timeFormat = time.RFC3339Nano

// newID mints entity IDs (overridable for tests).
// see [uuid.NewString].
var newID = uuid.NewString

// rowsErr reports iteration errors from a query (overridable for tests).
// see [sql.Rows.Err].
var rowsErr = func(rows *sql.Rows) error {
	return rows.Err()
}

// scanAll collects every row of rows via scan.
// what names the rows in the iteration error, e.g. "users".
// The result is non-nil, so an empty set marshals as [] rather than null.
func scanAll[T any](rows *sql.Rows, what string, scan func(scanner) (*T, error)) ([]T, error) {
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

// randRead fills b with cryptographically random bytes (overridable for tests).
// see [rand.Read].
var randRead = rand.Read

// timeNow returns the current UTC time (overridable for tests).
// see [time.Now] followed by [time.Time.UTC].
var timeNow = func() time.Time { return time.Now().UTC() }

// Store persists auth entities in SQLite.
type Store struct {
	db *sql.DB
}

// New creates a [Store] on db, applying pending schema migrations, syncing the
// permission catalogue, and seeding the built-in groups (admin re-synced every
// start; operator/viewer seeded once on first install).
func New(ctx context.Context, db *sql.DB) (*Store, error) {
	s := &Store{db: db}

	freshInstall, err := s.migrate(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth store migrate: %w", err)
	}
	if err := s.syncPermissions(ctx); err != nil {
		return nil, fmt.Errorf("auth store sync permissions: %w", err)
	}
	if err := s.seedGroups(ctx, freshInstall); err != nil {
		return nil, fmt.Errorf("auth store seed groups: %w", err)
	}

	return s, nil
}

// migration is a versioned set of schema statements.
type migration struct {
	version    int
	statements []string
}

// migrations lists every schema migration, in order. Append-only.
var migrations = []migration{
	{version: 1, statements: []string{
		`CREATE TABLE IF NOT EXISTS users (
			id            TEXT NOT NULL PRIMARY KEY,
			username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
			display_name  TEXT NOT NULL DEFAULT '',
			email         TEXT NOT NULL DEFAULT '',
			password_hash TEXT,
			enabled       INTEGER NOT NULL DEFAULT 1,
			created_at    DATETIME NOT NULL,
			updated_at    DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS groups (
			id          TEXT NOT NULL PRIMARY KEY,
			name        TEXT NOT NULL UNIQUE COLLATE NOCASE,
			description TEXT NOT NULL DEFAULT '',
			system      INTEGER NOT NULL DEFAULT 0,
			seed_key    TEXT UNIQUE,
			created_at  DATETIME NOT NULL,
			updated_at  DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS user_groups (
			user_id  TEXT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
			group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
			PRIMARY KEY (user_id, group_id)
		);`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			resource TEXT NOT NULL,
			action   TEXT NOT NULL,
			UNIQUE (resource, action)
		);`,
		`CREATE TABLE IF NOT EXISTS group_permissions (
			group_id      TEXT    NOT NULL REFERENCES groups(id)      ON DELETE CASCADE,
			permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
			scope_type    TEXT    NOT NULL DEFAULT 'global',
			scope_ref     TEXT    NOT NULL DEFAULT '',
			PRIMARY KEY (group_id, permission_id, scope_type, scope_ref)
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token_hash   TEXT NOT NULL PRIMARY KEY,
			user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at   DATETIME NOT NULL,
			expires_at   DATETIME NOT NULL,
			last_seen_at DATETIME NOT NULL,
			ip           TEXT NOT NULL DEFAULT '',
			user_agent   TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS api_tokens (
			id           TEXT NOT NULL PRIMARY KEY,
			user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name         TEXT NOT NULL,
			token_hash   TEXT NOT NULL UNIQUE,
			prefix       TEXT NOT NULL,
			created_at   DATETIME NOT NULL,
			expires_at   DATETIME,
			last_used_at DATETIME
		);`,
	}},
}

// migrate applies the migrations not yet recorded in schema_migrations,
// reporting whether the database was empty (no migrations applied) beforehand.
func (s *Store) migrate(ctx context.Context) (freshInstall bool, err error) {
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER NOT NULL PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);`); err != nil {
		return false, fmt.Errorf("create schema_migrations: %w", err)
	}

	// Find applied migrations.
	applied := map[int]bool{}
	rows, err := s.db.QueryContext(ctx, `SELECT version FROM schema_migrations;`)
	if err != nil {
		return false, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return false, fmt.Errorf("scan schema_migrations: %w", err)
		}
		applied[version] = true
	}
	if err := rowsErr(rows); err != nil {
		return false, fmt.Errorf("iterate schema_migrations: %w", err)
	}

	// A database with no recorded migrations is a first-time install;
	// starter groups are seeded only then.
	freshInstall = len(applied) == 0

	// Apply missing migrations.
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if err := s.applyMigration(ctx, m); err != nil {
			return false, fmt.Errorf("apply migration %d: %w", m.version, err)
		}
	}

	return freshInstall, nil
}

// applyMigration runs one migration's statements in a transaction.
func (s *Store) applyMigration(ctx context.Context, m migration) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, statement := range m.statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return err //nolint:wrapcheck
			}
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?);`,
			m.version, timeNow().Format(timeFormat),
		)
		return err //nolint:wrapcheck
	})
}

// syncPermissions reconciles the permissions table with the catalogue:
// missing (resource, action) pairs are inserted; pairs no longer in the
// catalogue are removed together with any grants referencing them.
func (s *Store) syncPermissions(ctx context.Context) error {
	valid := map[rbac.Permission]bool{}
	for _, rp := range rbac.Catalogue() {
		for _, ap := range rp.Actions {
			valid[rbac.Permission{Resource: rp.Resource, Action: ap.Action}] = true
		}
	}

	return s.inTx(ctx, func(tx *sql.Tx) error {
		existing, err := permissionIDsByPair(ctx, tx)
		if err != nil {
			return err //nolint:wrapcheck
		}

		// Insert catalogue pairs not yet in the table.
		for permission := range valid {
			if _, exists := existing[permission]; exists {
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO permissions (resource, action) VALUES (?, ?);`,
				string(permission.Resource), string(permission.Action),
			); err != nil {
				return err //nolint:wrapcheck
			}
		}

		// Remove stale pairs (and their grants).
		logFrom := logx.LogFrom{Primary: "auth", Secondary: "permissions"}
		for permission, id := range existing {
			if valid[permission] {
				continue
			}
			result, err := tx.ExecContext(ctx,
				`DELETE FROM group_permissions WHERE permission_id = ?;`, id,
			)
			if err != nil {
				return err //nolint:wrapcheck
			}
			// Record grants destroyed.
			grants, err := result.RowsAffected()
			if err != nil {
				return err //nolint:wrapcheck
			}
			logx.Warn(
				fmt.Sprintf(
					"dropped permission %s:%s (no longer in the catalogue), removing it from %d group(s)",
					permission.Resource, permission.Action, grants,
				),
				logFrom, true,
			)
			if _, err := tx.ExecContext(ctx,
				`DELETE FROM permissions WHERE id = ?;`, id,
			); err != nil {
				return err //nolint:wrapcheck
			}
		}

		return nil
	})
}

// seededGroups are the groups created when the database is first initialised.
// [GroupAdmin] is a protected system group whose grants are re-synced to the
// full catalogue on every start; [GroupOperator]/[GroupViewer] are starter
// groups seeded once and thereafter user-owned (deletable, renameable, grants
// editable). key is an immutable identity (stored as seed_key), independent of
// the user-editable name.
var seededGroups = []struct {
	key         string
	name        string
	system      bool
	description string
	matches     func(rbac.Permission) bool
}{
	{
		GroupAdmin, GroupAdmin, true,
		"Full administrative access.",
		func(rbac.Permission) bool { return true },
	},
	{
		GroupOperator, GroupOperator, false,
		"Operate and configure services; no user/group administration.",
		// User/group administration is admin-only, gated by requireAdmin
		// rather than by a grantable permission.
		func(rbac.Permission) bool { return true },
	},
	{
		GroupViewer, GroupViewer, false,
		"Read-only dashboard access.",
		func(p rbac.Permission) bool {
			return p.Action == rbac.ActionRead
		},
	},
}

// seedGroups maintains the seeded groups:
//   - fresh install: creates all of them, granting each its remit.
//   - every start: (re-)ensures admin exists, re-synced to the full catalogue,
//     so newly-added permissions are always held by at least one group.
//   - starter groups (operator/viewer) are user-owned after creation: found by
//     their immutable seed_key (so renames survive), never re-created after
//     deletion, and their grants are never touched again - permissions
//     catalogued by later releases stay admin-only until granted deliberately.
func (s *Store) seedGroups(ctx context.Context, freshInstall bool) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, seed := range seededGroups {
			id, err := groupIDBySeedKey(ctx, tx, seed.key)
			if err != nil && !errors.Is(err, ErrNotFound) {
				return err
			}

			// Create the group if absent (and freshInstall).
			if errors.Is(err, ErrNotFound) {
				if !seed.system && !freshInstall {
					continue
				}
				id = newID()
				now := timeNow().Format(timeFormat)
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO groups (id, name, description, system, seed_key, created_at, updated_at)
						VALUES (?, ?, ?, ?, ?, ?, ?);`,
					id, seed.name, seed.description, seed.system, seed.key, now, now,
				); err != nil {
					return err //nolint:wrapcheck
				}
			}

			// admin always grants the full catalogue.
			if seed.key == GroupAdmin {
				if err := setGroupGrants(ctx, tx, id, fullCatalogueGrants()); err != nil {
					return err
				}
				continue
			}

			// Starter groups: seed their remit once; user-owned thereafter.
			if !freshInstall {
				continue
			}
			var grants []rbac.Grant
			for _, grant := range fullCatalogueGrants() {
				if seed.matches(grant.Permission) {
					grants = append(grants, grant)
				}
			}
			if err := insertGroupGrants(ctx, tx, id, grants); err != nil {
				return err
			}
		}

		return nil
	})
}

// fullCatalogueGrants returns a global-scope grant for every valid permission.
func fullCatalogueGrants() []rbac.Grant {
	var grants []rbac.Grant
	for _, rp := range rbac.Catalogue() {
		for _, ap := range rp.Actions {
			grants = append(grants, rbac.Grant{
				Permission: rbac.Permission{Resource: rp.Resource, Action: ap.Action},
				Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
			})
		}
	}
	return grants
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
	t, err := time.Parse(timeFormat, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse stored timestamp %q: %w", value, err)
	}
	return t, nil
}
