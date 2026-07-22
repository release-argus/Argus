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

package store

import (
	"database/sql"
	"fmt"
	"slices"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
)

// TestStore_FaultInjection drives every repository error branch that only a
// failing SQL statement can reach, by arming the fault driver against a
// statement-identifying SQL substring.
func TestStore__faultInjection(t *testing.T) {
	// GIVEN: an initialised store whose next matching statement will fail.
	enabled := true
	disabled := false

	tests := []struct {
		name    string
		setup   func(t *testing.T, store *Store)
		pattern string
		invoke  func(t *testing.T, store *Store) error
	}{
		// users.go
		{
			name:    "CreateUser/username check fails",
			pattern: `COUNT(*) FROM users WHERE username`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateUser(t.Context(), "u", "", "", "", nil)
				return err
			},
		},
		{
			name:    "CreateUser/insert fails",
			pattern: `INSERT INTO users`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateUser(t.Context(), "u", "", "", "", nil)
				return err
			},
		},
		{
			name:    "CreateUser/group lookup fails",
			pattern: `SELECT id FROM groups WHERE name`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateUser(t.Context(), "u", "", "", "", []string{GroupViewer})
				return err
			},
		},
		{
			name:    "CreateUser/membership insert fails",
			pattern: `INSERT OR IGNORE INTO user_groups`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateUser(t.Context(), "u", "", "", "", []string{GroupViewer})
				return err
			},
		},
		{
			name:    "UpdateUser/load fails",
			pattern: `SELECT enabled FROM users WHERE id`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.UpdateUser(t.Context(), "any", UserPatch{})
				return err
			},
		},
		{
			name: "UpdateUser/admin-membership check fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `AND u.id = ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				_, err := store.UpdateUser(t.Context(), user.UserID,
					UserPatch{Enabled: &disabled})
				return err
			},
		},
		{
			name: "UpdateUser/other-admin count fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "", GroupAdmin)
			},
			pattern: `AND u.id != ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				_, err := store.UpdateUser(t.Context(), user.UserID,
					UserPatch{Enabled: &disabled})
				return err
			},
		},
		{
			name: "UpdateUser/display_name update fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `, display_name = ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				_, err := store.UpdateUser(t.Context(), user.UserID,
					UserPatch{DisplayName: new("x")})
				return err
			},
		},
		{
			name: "UpdateUser/email update fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `, email = ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				_, err := store.UpdateUser(t.Context(), user.UserID,
					UserPatch{Email: new("x@example.com")})
				return err
			},
		},
		{
			name: "UpdateUser/enabled update fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `, enabled = ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				// Enabling an enabled user skips the rails.
				_, err := store.UpdateUser(t.Context(), user.UserID,
					UserPatch{Enabled: &enabled})
				return err
			},
		},
		{
			name: "UpdateUser/password update fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `, password_hash = ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				_, err := store.UpdateUser(t.Context(), user.UserID,
					UserPatch{PasswordHash: new("x")})
				return err
			},
		},
		{
			name: "UpdateUser/clear memberships fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `DELETE FROM user_groups WHERE user_id`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				_, err := store.UpdateUser(t.Context(), user.UserID,
					UserPatch{Groups: new([]string{GroupViewer})})
				return err
			},
		},
		{
			name: "UpdateUser/updated_at fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `SET updated_at`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				_, err := store.UpdateUser(t.Context(), user.UserID, UserPatch{})
				return err
			},
		},
		{
			name:    "DeleteUser/load fails",
			pattern: `SELECT enabled FROM users`,
			invoke: func(t *testing.T, store *Store) error {
				return store.DeleteUser(t.Context(), "any")
			},
		},
		{
			name: "DeleteUser/admin-membership check fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `AND u.id = ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				return store.DeleteUser(t.Context(), user.UserID)
			},
		},
		{
			name: "DeleteUser/other-admin count fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "", GroupAdmin)
			},
			pattern: `AND u.id != ?`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				return store.DeleteUser(t.Context(), user.UserID)
			},
		},
		{
			name: "DeleteUser/cascade delete fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `DELETE FROM user_groups WHERE user_id`,
			invoke: func(t *testing.T, store *Store) error {
				user, _ := store.LocalCredentials(t.Context(), "target")
				return store.DeleteUser(t.Context(), user.UserID)
			},
		},
		{
			name:    "CreateFirstAdmin/count fails",
			pattern: `SELECT COUNT(*) FROM users;`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateFirstAdmin(t.Context(), "root", "", "password")
				return err
			},
		},
		{
			name:    "CreateFirstAdmin/create fails",
			pattern: `INSERT INTO users`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateFirstAdmin(t.Context(), "root", "", "password")
				return err
			},
		},
		{
			name:    "CreateFirstAdmin/membership insert fails",
			pattern: `INSERT OR IGNORE INTO user_groups`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateFirstAdmin(t.Context(), "root", "", "password")
				return err
			},
		},
		{
			name: "ResetUserPassword/load fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `SELECT id FROM users WHERE username`,
			invoke: func(t *testing.T, store *Store) error {
				return store.ResetUserPassword(t.Context(), "target", "new")
			},
		},
		{
			name: "ResetUserPassword/update fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `SET password_hash = ?, updated_at`,
			invoke: func(t *testing.T, store *Store) error {
				return store.ResetUserPassword(t.Context(), "target", "new")
			},
		},
		{
			name: "ResetUserPassword/session revocation fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `DELETE FROM sessions WHERE user_id`,
			invoke: func(t *testing.T, store *Store) error {
				return store.ResetUserPassword(t.Context(), "target", "new")
			},
		},
		// groups.go
		{
			name:    "CreateGroup/name check fails",
			pattern: `SELECT id FROM groups WHERE name`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateGroup(t.Context(), "g", "", nil)
				return err
			},
		},
		{
			name:    "CreateGroup/insert fails",
			pattern: `INSERT INTO groups`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateGroup(t.Context(), "g", "", nil)
				return err
			},
		},
		{
			name:    "UpdateGroup/load fails",
			pattern: `SELECT name, system FROM groups`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.UpdateGroup(t.Context(), "any", GroupPatch{})
				return err
			},
		},
		{
			name: "UpdateGroup/rename check fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `SELECT id FROM groups WHERE name`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.UpdateGroup(t.Context(), group.ID,
					GroupPatch{Name: new("renamed")})
				return err
			},
		},
		{
			name: "UpdateGroup/rename update fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `, name = ?`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.UpdateGroup(t.Context(), group.ID,
					GroupPatch{Name: new("renamed")})
				return err
			},
		},
		{
			name: "UpdateGroup/description update fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `, description = ?`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.UpdateGroup(t.Context(), group.ID,
					GroupPatch{Description: new("x")})
				return err
			},
		},
		{
			name: "UpdateGroup/clear grants fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `DELETE FROM group_permissions WHERE group_id`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.UpdateGroup(t.Context(), group.ID,
					GroupPatch{Grants: new([]rbac.Grant{})})
				return err
			},
		},
		{
			name: "UpdateGroup/permission load fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `SELECT id, resource, action FROM permissions`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.UpdateGroup(t.Context(), group.ID,
					GroupPatch{Grants: new([]rbac.Grant{
						globalGrant(rbac.ResourceService, rbac.ActionRead),
					})})
				return err
			},
		},
		{
			name: "UpdateGroup/insert grant fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `INSERT OR IGNORE INTO group_permissions`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.UpdateGroup(t.Context(), group.ID,
					GroupPatch{Grants: new([]rbac.Grant{
						globalGrant(rbac.ResourceService, rbac.ActionRead),
					})})
				return err
			},
		},
		{
			name: "UpdateGroup/updated_at fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `SET updated_at`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.UpdateGroup(t.Context(), group.ID, GroupPatch{})
				return err
			},
		},
		{
			name:    "DeleteGroup/load fails",
			pattern: `SELECT system FROM groups`,
			invoke: func(t *testing.T, store *Store) error {
				return store.DeleteGroup(t.Context(), "any")
			},
		},
		{
			name: "DeleteGroup/cascade delete fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `DELETE FROM user_groups WHERE group_id`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				return store.DeleteGroup(t.Context(), group.ID)
			},
		},
		{
			name:    "migrate/version read fails",
			pattern: `SELECT version FROM schema_migrations`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.migrate(t.Context())
				return err
			},
		},
		{
			name: "UserByID/group-name query fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `ORDER BY g.name`,
			invoke: func(t *testing.T, store *Store) error {
				creds, err := store.LocalCredentials(t.Context(), "target")
				if err != nil {
					return err
				}
				_, err = store.UserByID(t.Context(), creds.UserID)
				return err
			},
		},
		{
			name: "GroupByID/grants query fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			pattern: `WHERE gp.group_id = ?`,
			invoke: func(t *testing.T, store *Store) error {
				group := findGroup(t, store, "custom")
				_, err := store.GroupByID(t.Context(), group.ID)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store, state := testFaultStore(t)
			if tc.setup != nil {
				tc.setup(t, store)
			}

			// WHEN: the fault is armed and the method invoked.
			state.Set(tc.pattern)
			err := tc.invoke(t, store)
			state.Set("")

			prefix := fmt.Sprintf("%s\nfault injection", packageName)

			// THEN: the failure is surfaced.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

// TestNew_FaultInjection drives the initialisation error branches.
func TestNew__faultInjection(t *testing.T) {
	// GIVEN: databases whose next matching statement will fail.
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *sql.DB) // Optional pre-initialisation.
		pattern string
	}{
		{
			name:    "seed group insert fails",
			pattern: `INSERT INTO groups`,
		},
		{
			name:    "permission catalogue read fails",
			pattern: `SELECT id, resource, action FROM permissions`,
		},
		{
			name:    "permission insert fails",
			pattern: `INSERT INTO permissions`,
		},
		{
			name:    "seed grant clear fails",
			pattern: `DELETE FROM group_permissions WHERE group_id`,
		},
		{
			name:    "admin grant re-sync fails",
			pattern: `INSERT OR IGNORE INTO group_permissions`,
		},
		{
			name:    "stale grant delete fails",
			setup:   presetStalePermission,
			pattern: `DELETE FROM group_permissions WHERE permission_id`,
		},
		{
			name:    "stale permission delete fails",
			setup:   presetStalePermission,
			pattern: `DELETE FROM permissions WHERE id`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, state := testFaultDB(t)
			if tc.setup != nil {
				tc.setup(t, db)
			}

			// WHEN: the fault is armed and New invoked.
			state.Set(tc.pattern)
			_, err := New(t.Context(), db)
			state.Set("")

			prefix := fmt.Sprintf("%s\nNew() fault injection", packageName)

			// THEN: initialisation fails.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestSeedGroups__newPermissionGrantError(t *testing.T) {
	// GIVEN: an initialised database missing a catalogue pair (so the next start
	// treats it as new), with grant inserts armed to fail. The seed order is
	// reversed so a starter group's insert fires before admin's re-sync.
	seededGroupsHad := seededGroups
	seededGroups = slices.Clone(seededGroups)
	slices.Reverse(seededGroups)
	t.Cleanup(func() { seededGroups = seededGroupsHad })

	db, state := testFaultDB(t)
	if _, err := New(t.Context(), db); err != nil {
		t.Fatalf(
			"%s\nsetup New() failed: %v",
			packageName, err,
		)
	}
	removePermission(t, db, rbac.ResourceConfig, rbac.ActionRead)

	prefix := fmt.Sprintf("%s\nNew() with failing starter-group grant insert", packageName)

	// WHEN: New runs again with the fault armed.
	state.Set(`INSERT OR IGNORE INTO group_permissions`)
	_, err := New(t.Context(), db)
	state.Set("")

	// THEN: the failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

// presetStalePermission initialises the schema (fault disarmed) then plants
// a permission (and a grant on it) that is no longer in the catalogue.
func presetStalePermission(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := New(t.Context(), db); err != nil {
		t.Fatalf(
			"%s\nsetup New() failed: %v",
			packageName, err,
		)
	}
	for _, statement := range []string{
		`INSERT INTO permissions (resource, action) VALUES ('legacy', 'something');`,
		`INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
			VALUES (
				(SELECT id FROM groups WHERE name = 'viewer'),
				(SELECT id FROM permissions WHERE resource = 'legacy'),
				'global', '');`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf(
				"%s\nsetup failed: %v",
				packageName, err,
			)
		}
	}
}

func TestStore_corruptRows(t *testing.T) {
	// GIVEN: stores holding rows the scanners cannot decode.
	tests := []struct {
		name   string
		setup  func(t *testing.T, store *Store)
		invoke func(t *testing.T, store *Store) error
	}{
		{
			name: "UserByID/corrupt timestamps",
			setup: func(t *testing.T, store *Store) {
				mustExec(t, store, `
					INSERT INTO users (id, username, created_at, updated_at)
					VALUES ('corrupt', 'corrupt', 'not-a-time', 'not-a-time');`)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.UserByID(t.Context(), "corrupt")
				return err
			},
		},
		{
			name: "Users/corrupt updated_at only",
			setup: func(t *testing.T, store *Store) {
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO users (id, username, created_at, updated_at)
					VALUES ('corrupt', 'corrupt', '%s', 'not-a-time');`,
					timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.Users(t.Context())
				return err
			},
		},
		{
			name: "UserByID/NULL group name",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "target", "")
				dropTable(t, store, "groups")
				mustExec(t, store, `CREATE TABLE groups (id TEXT, name TEXT, system INTEGER DEFAULT 0);`)
				mustExec(t, store, `INSERT INTO groups (id, name) VALUES ('gid', NULL);`)
				mustExec(t, store,
					fmt.Sprintf(`INSERT INTO user_groups (user_id, group_id) VALUES ('%s', 'gid');`, user.ID))
			},
			invoke: func(t *testing.T, store *Store) error {
				creds, err := store.LocalCredentials(t.Context(), "target")
				if err != nil {
					return err
				}
				_, err = store.UserByID(t.Context(), creds.UserID)
				return err
			},
		},
		{
			name: "Groups/corrupt updated_at only",
			setup: func(t *testing.T, store *Store) {
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO groups (id, name, created_at, updated_at)
					VALUES ('corrupt', 'corrupt', '%s', 'not-a-time');`,
					timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.Groups(t.Context())
				return err
			},
		},
		{
			name: "GroupByID/corrupt timestamps",
			setup: func(t *testing.T, store *Store) {
				mustExec(t, store, `
					INSERT INTO groups (id, name, created_at, updated_at)
					VALUES ('corrupt', 'corrupt', 'not-a-time', 'not-a-time');`)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.GroupByID(t.Context(), "corrupt")
				return err
			},
		},
		{
			name: "GrantsForUser/NULL scope_ref",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "target", "", GroupViewer)
				dropTable(t, store, "group_permissions")
				mustExec(t, store, `CREATE TABLE group_permissions (group_id TEXT, permission_id INTEGER, scope_type TEXT, scope_ref TEXT);`)
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
					SELECT g.id, p.id, 'global', NULL FROM groups g, permissions p
					WHERE g.name = '%s' LIMIT 1;`, GroupViewer))
				_ = user
			},
			invoke: func(t *testing.T, store *Store) error {
				creds, err := store.LocalCredentials(t.Context(), "target")
				if err != nil {
					return err
				}
				_, err = store.GrantsForUser(t.Context(), creds.UserID)
				return err
			},
		},
		{
			name: "SessionByTokenHash/corrupt expires_at only",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "target", "")
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO sessions (token_hash, user_id, created_at, expires_at, last_seen_at)
					VALUES ('corrupt', '%s', '%s', 'not-a-time', '%s');`,
					user.ID, timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.SessionByTokenHash(t.Context(), "corrupt")
				return err
			},
		},
		{
			name: "SessionByTokenHash/corrupt last_seen_at only",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "target", "")
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO sessions (token_hash, user_id, created_at, expires_at, last_seen_at)
					VALUES ('corrupt', '%s', '%s', '%s', 'not-a-time');`,
					user.ID, timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.SessionByTokenHash(t.Context(), "corrupt")
				return err
			},
		},
		{
			name: "GroupByID/NULL grant scope_ref",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
				dropTable(t, store, "group_permissions")
				mustExec(t, store, `CREATE TABLE group_permissions (group_id TEXT, permission_id INTEGER, scope_type TEXT, scope_ref TEXT);`)
				mustExec(t, store, `
					INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
					SELECT g.id, p.id, 'global', NULL FROM groups g, permissions p
					WHERE g.name = 'custom' LIMIT 1;`)
			},
			invoke: func(t *testing.T, store *Store) error {
				var id string
				if err := store.db.QueryRow(
					`SELECT id FROM groups WHERE name = 'custom';`,
				).Scan(&id); err != nil {
					t.Fatalf(
						"%s\nlookup failed: %v",
						packageName, err,
					)
				}
				_, err := store.GroupByID(t.Context(), id)
				return err
			},
		},
		{
			name: "GroupByID/NULL group name",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "groups")
				mustExec(t, store, `
					CREATE TABLE groups (
						id TEXT,
						name TEXT,
						description TEXT DEFAULT '',
						system INTEGER DEFAULT 0,
						created_at DATETIME,
						updated_at DATETIME
					);`,
				)
				mustExec(t, store, fmt.Sprintf(
					`INSERT INTO groups (id, name, created_at, updated_at) VALUES ('gid', NULL, '%s', '%s');`,
					timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.GroupByID(t.Context(), "gid")
				return err
			},
		},
		{
			name: "UserByID/NULL username",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "users")
				mustExec(t, store, `
					CREATE TABLE users (
						id TEXT,
						username TEXT,
						display_name TEXT DEFAULT '',
						email TEXT DEFAULT '',
						password_hash TEXT,
						enabled INTEGER DEFAULT 1,
						created_at DATETIME,
						updated_at DATETIME
					);`,
				)
				mustExec(t, store, fmt.Sprintf(
					`INSERT INTO users (id, username, created_at, updated_at) VALUES ('uid', NULL, '%s', '%s');`,
					timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.UserByID(t.Context(), "uid")
				return err
			},
		},
		{
			name: "New/NULL permission row",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "permissions")
				mustExec(t, store, `CREATE TABLE permissions (id INTEGER PRIMARY KEY AUTOINCREMENT, resource TEXT, action TEXT);`)
				mustExec(t, store, `INSERT INTO permissions (resource, action) VALUES (NULL, NULL);`)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := New(t.Context(), store.db)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			tc.setup(t, store)

			// WHEN: the method scans the corrupt rows.
			err := tc.invoke(t, store)

			prefix := fmt.Sprintf("%s\ncorrupt rows", packageName)

			// THEN: the decode failure is surfaced.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestStore_rowsErr_on_SecondPass(t *testing.T) {
	// GIVEN: methods whose second row iteration fails.
	tests := []struct {
		name       string
		failOnCall int
		setup      func(t *testing.T, store *Store)
		invoke     func(t *testing.T, store *Store) error
	}{
		{
			name:       "Users/membership iteration fails",
			failOnCall: 2,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.Users(t.Context())
				return err
			},
		},
		{
			name:       "UserByID/group-name iteration fails",
			failOnCall: 1,
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			invoke: func(t *testing.T, store *Store) error {
				creds, err := store.LocalCredentials(t.Context(), "target")
				if err != nil {
					return err
				}
				_, err = store.UserByID(t.Context(), creds.UserID)
				return err
			},
		},
		{
			name:       "Groups/grant iteration fails",
			failOnCall: 2,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.Groups(t.Context())
				return err
			},
		},
		{
			name:       "GroupByID/grant iteration fails",
			failOnCall: 1,
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
			},
			invoke: func(t *testing.T, store *Store) error {
				// Resolve the ID without iterating rows (the fault is armed).
				var id string
				if err := store.db.QueryRow(
					`SELECT id FROM groups WHERE name = 'custom';`,
				).Scan(&id); err != nil {
					t.Fatalf(
						"%s\nlookup failed: %v",
						packageName, err,
					)
				}
				_, err := store.GroupByID(t.Context(), id)
				return err
			},
		},
		{
			name:       "GrantsForUser/iteration fails",
			failOnCall: 1,
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "", GroupViewer)
			},
			invoke: func(t *testing.T, store *Store) error {
				creds, err := store.LocalCredentials(t.Context(), "target")
				if err != nil {
					return err
				}
				_, err = store.GrantsForUser(t.Context(), creds.UserID)
				return err
			},
		},
		{
			name:       "New/permission-sync iteration fails",
			failOnCall: 2, // 1st: migrate versions; 2nd: permission sync.
			invoke: func(t *testing.T, store *Store) error {
				_, err := New(t.Context(), store.db)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			if tc.setup != nil {
				tc.setup(t, store)
			}
			overrideRowsErrOnCall(t, tc.failOnCall)

			// WHEN: the method iterates rows.
			err := tc.invoke(t, store)

			prefix := fmt.Sprintf("%s\nsecond-pass iteration", packageName)

			// THEN: the iteration failure is surfaced.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestInTx_beginError(t *testing.T) {
	// GIVEN: a Store whose database is closed.
	db := testDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s\nNew() failed: %v",
			packageName, err,
		)
	}
	_ = db.Close()

	prefix := fmt.Sprintf("%s\ninTx() begin error", packageName)

	// WHEN: a transactional method runs.
	_, err = store.CreateUser(t.Context(), "u", "", "", "", nil)

	// THEN: the begin failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

// findGroup returns the named group, failing the test if absent.
func findGroup(t *testing.T, store *Store, name string) *Group {
	t.Helper()

	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf(
			"%s\nGroups() failed: %v",
			packageName, err,
		)
	}
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i]
		}
	}
	t.Fatalf(
		"%s\ngroup %q not found",
		packageName, name,
	)
	return nil
}
