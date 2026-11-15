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
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/release-argus/Argus/auth/rbac"
	storetest "github.com/release-argus/Argus/auth/store/test"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// TestStore__faultInjection drives every repository error branch that only a
// failing SQL statement can reach, by arming the fault driver against a
// statement-identifying SQL substring.
func TestStore__faultInjection(t *testing.T) {
	enabled := true
	disabled := false

	tests := []struct {
		name           string
		setup          func(t *testing.T, store *Store)
		pattern        string                                 // SQL pattern that will fail.
		rowsAffected   bool                                   // Fail RowsAffected of pattern rather than the statement.
		invoke         func(t *testing.T, store *Store) error // Called with the fault armed.
		wantRolledBack func(t *testing.T, store *Store)       // Checks the failed call left no partial write.
	}{
		{
			name:    "CreateUser/insert fails",
			pattern: `INSERT INTO users`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateUser(t.Context(), "u", "", "", testPasswordHash, nil)
				return err
			},
		},
		{
			name:    "CreateUser/group lookup fails",
			pattern: `SELECT id FROM groups WHERE name`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateUser(t.Context(), "u", "", "", testPasswordHash, []string{GroupViewer})
				return err
			},
		},
		{
			name:    "CreateUser/membership insert fails",
			pattern: `INSERT OR IGNORE INTO user_groups`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateUser(t.Context(), "u", "", "", testPasswordHash, []string{GroupViewer})
				return err
			},
			wantRolledBack: func(t *testing.T, store *Store) {
				// The users row is written before the membership.
				if got, want := test.CountRows(t, store.db, "users"), 0; got != want {
					t.Errorf("users row count mismatch\ngot:  %d\nwant: %d", got, want)
				}
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
			name: "UpdateUser/admin-status check fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "", GroupAdmin)
			},
			pattern: `SUM(u.id = ?)`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.UpdateUser(t.Context(), mustUserID(t, store, "target"),
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
				_, err := store.UpdateUser(t.Context(), mustUserID(t, store, "target"),
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
				_, err := store.UpdateUser(t.Context(), mustUserID(t, store, "target"),
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
				// Enabling an enabled user skips the rails.
				_, err := store.UpdateUser(t.Context(), mustUserID(t, store, "target"),
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
				_, err := store.UpdateUser(t.Context(), mustUserID(t, store, "target"),
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
				_, err := store.UpdateUser(t.Context(), mustUserID(t, store, "target"),
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
				_, err := store.UpdateUser(t.Context(), mustUserID(t, store, "target"), UserPatch{})
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
			name: "DeleteUser/admin-status check fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "", GroupAdmin)
			},
			pattern: `SUM(u.id = ?)`,
			invoke: func(t *testing.T, store *Store) error {
				return store.DeleteUser(t.Context(), mustUserID(t, store, "target"))
			},
		},
		{
			name: "DeleteUser/cascade delete fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `DELETE FROM user_groups WHERE user_id`,
			invoke: func(t *testing.T, store *Store) error {
				return store.DeleteUser(t.Context(), mustUserID(t, store, "target"))
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
			wantRolledBack: func(t *testing.T, store *Store) {
				// The users row is written before the membership.
				if got := test.CountRows(t, store.db, "users"); got != 0 {
					t.Errorf("users row count mismatch\ngot:  %d\nwant: 0", got)
				}
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
		{
			name:    "UserWithGrants/query fails",
			pattern: `LEFT JOIN group_permissions gp ON gp.group_id = ug.group_id`,
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.UserWithGrants(t.Context(), "any")
				return err
			},
		},
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
				group := mustGroupByName(t, store, "custom")
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
				group := mustGroupByName(t, store, "custom")
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
				group := mustGroupByName(t, store, "custom")
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
				group := mustGroupByName(t, store, "custom")
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
				group := mustGroupByName(t, store, "custom")
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
				group := mustGroupByName(t, store, "custom")
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
				group := mustGroupByName(t, store, "custom")
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
				group := mustGroupByName(t, store, "custom")
				return store.DeleteGroup(t.Context(), group.ID)
			},
		},
		{
			name: "RenameServiceScopeRefs/rename fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "scoped",
					serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"))
			},
			pattern: `UPDATE OR REPLACE group_permissions`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.RenameServiceScopeRefs(t.Context(), "old-name", "new-name")
				return err
			},
		},
		{
			name: "UndoServiceScopeMove/repoint fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "scoped",
					serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"))
			},
			pattern: `AND group_id = ? AND permission_id = ?`,
			invoke: func(t *testing.T, store *Store) error {
				move, err := store.RenameServiceScopeRefs(t.Context(), "old-name", "new-name")
				if err != nil {
					return err
				}
				return store.UndoServiceScopeMove(t.Context(), move)
			},
		},
		{
			name: "UndoServiceScopeMove/absorbed grant restore fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "collide",
					serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"),
					serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "new-name"))
			},
			pattern: `INSERT OR IGNORE INTO group_permissions`,
			invoke: func(t *testing.T, store *Store) error {
				move, err := store.RenameServiceScopeRefs(t.Context(), "old-name", "new-name")
				if err != nil {
					return err
				}
				return store.UndoServiceScopeMove(t.Context(), move)
			},
		},
		{
			name:         "CreateFirstAdmin/insert row count fails",
			pattern:      `INSERT INTO users`,
			rowsAffected: true,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.CreateFirstAdmin(t.Context(), "root", "", "setup-password")
				return err
			},
		},
		{
			name: "TrimSessionsForUser/delete fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateSessions(t, store, "target", 2)
			},
			pattern: `DELETE FROM sessions WHERE token_hash`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.TrimSessionsForUser(t.Context(), mustUserID(t, store, "target"), 1)
				return err
			},
			wantRolledBack: func(t *testing.T, store *Store) {
				if got, want := test.CountRows(t, store.db, "sessions"), 2; got != want {
					t.Errorf("sessions row count mismatch\ngot:  %d\nwant: %d", got, want)
				}
			},
		},
		{
			name: "CreateAPIToken/insert fails (rowsAffected)",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern:      `INSERT INTO api_tokens`,
			rowsAffected: true,
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.CreateAPIToken(t.Context(), mustUserID(t, store, "target"), "ci", nil)
				return err
			},
		},
		{
			name: "CreateAPIToken/insert fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `INSERT INTO api_tokens`,
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.CreateAPIToken(t.Context(), mustUserID(t, store, "target"), "ci", nil)
				return err
			},
		},
		{
			name:    "PreferencesForUser/load fails",
			pattern: `SELECT hide, view, timestamps`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.PreferencesForUser(t.Context(), "any")
				return err
			},
		},
		{
			name:    "SetPreferences/upsert fails",
			pattern: `INSERT INTO user_preferences`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.SetPreferences(t.Context(), "any", DashboardPreferences{
					View: ViewTable,
				})
				return err
			},
		},
		{
			name:    "SetPreferences/zero set deletes, and that fails",
			pattern: `DELETE FROM user_preferences WHERE user_id`,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.SetPreferences(t.Context(), "any", DashboardPreferences{})
				return err
			},
		},
		{
			name:    "DeletePreferences/delete fails",
			pattern: `DELETE FROM user_preferences WHERE user_id`,
			invoke: func(t *testing.T, store *Store) error {
				return store.DeletePreferences(t.Context(), "any")
			},
		},
		{
			name: "DeleteUser/preferences cascade fails",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "")
			},
			pattern: `DELETE FROM user_preferences WHERE user_id`,
			invoke: func(t *testing.T, store *Store) error {
				return store.DeleteUser(t.Context(), mustUserID(t, store, "target"))
			},
			wantRolledBack: func(t *testing.T, store *Store) {
				// The cascade runs before the user row is removed.
				if got, want := test.CountRows(t, store.db, "users"), 1; got != want {
					t.Errorf("users row count mismatch\ngot:  %d\nwant: %d", got, want)
				}
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
				group := mustGroupByName(t, store, "custom")
				_, err := store.GroupByID(t.Context(), group.ID)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store, state := testFaultStore(t)
			if tc.setup != nil {
				tc.setup(t, store)
			}

			if tc.rowsAffected {
				state.SetRowsAffected(tc.pattern)
			} else {
				state.Set(tc.pattern)
			}
			err := tc.invoke(t, store)
			state.Set("")

			if !errors.Is(err, errInjected) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, errInjected)
			}

			if tc.wantRolledBack != nil {
				tc.wantRolledBack(t, store)
			}
		})
	}
}

// TestNew__faultInjection drives the initialisation error branches.
func TestNew__faultInjection(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T, db *sql.DB) // Optional pre-initialisation.
		pattern      string
		rowsAffected bool // Fail RowsAffected of pattern rather than the statement.
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
			name:    "stale permission retire fails",
			setup:   presetStalePermission,
			pattern: `SET retired_at = ?`,
		},
		{
			name:    "retired permission grant count fails",
			setup:   presetStalePermission,
			pattern: `SELECT COUNT(*) FROM group_permissions WHERE permission_id`,
		},
		{
			name:    "catalogued permission un-retire fails",
			setup:   presetStalePermission,
			pattern: `SET retired_at = NULL`,
		},
		{
			name:         "stale permission retire row count fails",
			setup:        presetStalePermission,
			pattern:      `SET retired_at = ?`,
			rowsAffected: true,
		},
		{
			name:    "starter group grant insert fails",
			setup:   reverseSeedOrder,
			pattern: `INSERT OR IGNORE INTO group_permissions`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps seededGroups.

			db, state := storetest.FaultDB(t)
			if tc.setup != nil {
				tc.setup(t, db)
			}

			if tc.rowsAffected {
				state.SetRowsAffected(tc.pattern)
			} else {
				state.Set(tc.pattern)
			}
			_, err := New(t.Context(), db)
			state.Set("")

			if !errors.Is(err, errInjected) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, errInjected)
			}
		})
	}
}

func TestSeedGroups__newPermissionGrantError(t *testing.T) {
	db, state := storetest.FaultDB(t)
	if _, err := New(t.Context(), db); err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	removePermission(t, db, rbac.ResourceConfig, rbac.ActionRead)

	state.Set(`INSERT OR IGNORE INTO group_permissions`)
	_, err := New(t.Context(), db)
	state.Set("")

	// Admin's re-sync hits the fault: starters are granted only on a fresh install.
	if !errors.Is(err, errInjected) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, errInjected)
	}
}

// mustCreateSessions creates username with count sessions, oldest first.
func mustCreateSessions(t *testing.T, store *Store, username string, count int) {
	t.Helper()

	user := mustCreateUser(t, store, username, "")
	now := timeNow()
	for i := range count {
		session := testSession(
			fmt.Sprintf("%s-%d", username, i),
			user.ID,
			now.Add(time.Hour),
		)
		session.LastSeenAt = now.Add(time.Duration(i) * time.Minute)
		if err := store.InsertSession(t.Context(), session); err != nil {
			t.Fatalf("session %d: Store.InsertSession() failed: %v", i, err)
		}
	}
}

// reverseSeedOrder puts a starter group ahead of admin, so on a fresh install its
// grant insert fires before admin's re-sync.
func reverseSeedOrder(t *testing.T, _ *sql.DB) {
	t.Helper()

	seededGroupsHad := seededGroups
	seededGroups = slices.Clone(seededGroups)
	slices.Reverse(seededGroups)
	t.Cleanup(func() { seededGroups = seededGroupsHad })
}

// presetStalePermission initialises the schema (fault disarmed) then plants
// a permission (and a grant on it) that is no longer in the catalogue.
func presetStalePermission(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := New(t.Context(), db); err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	test.ExecSQL(
		t, db,
		`INSERT INTO permissions (resource, action) VALUES ('legacy', 'something');`,
		`INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
			VALUES (
				(SELECT id FROM groups WHERE name = 'viewer'),
				(SELECT id FROM permissions WHERE resource = 'legacy'),
				'global', '');`,
	)
}

func TestStore__corruptRows(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, store *Store)
		invoke   func(t *testing.T, store *Store) error
		errRegex string
	}{
		{
			name: "UserByID/corrupt timestamps",
			setup: func(t *testing.T, store *Store) {
				test.ExecSQL(t, store.db, `
					INSERT INTO users
						(id, username, password_hash, created_at, updated_at)
					VALUES
						('corrupt', 'corrupt', 'x', 'not-a-time', 'not-a-time');`)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.UserByID(t.Context(), "corrupt")
				return err
			},
			errRegex: test.TrimYAML(`
				^scan user:
					[^\s]+.*name "created_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "Users/corrupt updated_at only",
			setup: func(t *testing.T, store *Store) {
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO users
						(id, username, password_hash, created_at, updated_at)
					VALUES
						('corrupt', 'corrupt', 'x', '%s', 'not-a-time');`,
					timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.Users(t.Context())
				return err
			},
			errRegex: test.TrimYAML(`
				^scan user:
					[^\s]+.*name "updated_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "UserByID/NULL group name",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "target", "")
				test.DropTable(t, store.db, "groups")
				test.ExecSQL(t, store.db, `
					CREATE TABLE groups
						(id TEXT, name TEXT, system INTEGER DEFAULT 0);`,
				)
				test.ExecSQL(t, store.db, `
					INSERT INTO groups
						(id, name)
					VALUES
						('gid', NULL);`,
				)
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO user_groups
						(user_id, group_id)
					VALUES
						('%s', 'gid');`,
					user.ID,
				))
			},
			invoke: func(t *testing.T, store *Store) error {
				creds, err := store.LocalCredentials(t.Context(), "target")
				if err != nil {
					return err
				}
				_, err = store.UserByID(t.Context(), creds.UserID)
				return err
			},
			errRegex: test.TrimYAML(`
				^scan user's groups:
					[^\s]+.*name "name":
						converting NULL to string is unsupported$`),
		},
		{
			name: "Groups/corrupt updated_at only",
			setup: func(t *testing.T, store *Store) {
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO groups
						(id, name, created_at, updated_at)
					VALUES
						('corrupt', 'corrupt', '%s', 'not-a-time');`,
					timeNow().Format(timeFormat),
				))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.Groups(t.Context())
				return err
			},
			errRegex: test.TrimYAML(`
				^scan group:
					[^\s]+.*name "updated_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "GroupByID/corrupt timestamps",
			setup: func(t *testing.T, store *Store) {
				test.ExecSQL(t, store.db, `
					INSERT INTO groups
						(id, name, created_at, updated_at)
					VALUES
						('corrupt', 'corrupt', 'not-a-time', 'not-a-time');`,
				)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.GroupByID(t.Context(), "corrupt")
				return err
			},
			errRegex: test.TrimYAML(`
				^scan group:
					[^\s]+.*name "created_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "GrantsForUser/NULL scope_ref",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "", GroupViewer)
				test.DropTable(t, store.db, "group_permissions")
				test.ExecSQL(t, store.db, `
					CREATE TABLE group_permissions
						(group_id TEXT, permission_id INTEGER, scope_type TEXT, scope_ref TEXT);`,
				)
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO group_permissions
						(group_id, permission_id, scope_type, scope_ref)
					SELECT
						g.id, p.id, 'global', NULL
					FROM groups g, permissions p
					WHERE g.name = '%s' LIMIT 1;`, GroupViewer))
			},
			invoke: func(t *testing.T, store *Store) error {
				creds, err := store.LocalCredentials(t.Context(), "target")
				if err != nil {
					return err
				}
				_, err = store.GrantsForUser(t.Context(), creds.UserID)
				return err
			},
			errRegex: test.TrimYAML(`
				^list user's grants:
					scan grant:
						[^\s]+.*name "scope_ref":
							converting NULL to string is unsupported$`),
		},
		{
			name: "SessionByTokenHash/corrupt expires_at only",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "target", "")
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO sessions
						(token_hash, user_id, created_at, expires_at, last_seen_at)
					VALUES
						('corrupt', '%s', '%s', 'not-a-time', '%s');`,
					user.ID, timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.SessionByTokenHash(t.Context(), "corrupt")
				return err
			},
			errRegex: test.TrimYAML(`
				^load session:
					[^\s]+.*name "expires_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "SessionByTokenHash/corrupt last_seen_at only",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "target", "")
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO sessions
						(token_hash, user_id, created_at, expires_at, last_seen_at)
					VALUES
						('corrupt', '%s', '%s', '%s', 'not-a-time');`,
					user.ID, timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.SessionByTokenHash(t.Context(), "corrupt")
				return err
			},
			errRegex: test.TrimYAML(`
				^load session:
					[^\s]+.*name "last_seen_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "GroupByID/NULL grant scope_ref",
			setup: func(t *testing.T, store *Store) {
				mustCreateGroup(t, store, "custom")
				test.DropTable(t, store.db, "group_permissions")
				test.ExecSQL(t, store.db, `
					CREATE TABLE group_permissions
						(group_id TEXT, permission_id INTEGER, scope_type TEXT, scope_ref TEXT);`,
				)
				test.ExecSQL(t, store.db, `
					INSERT INTO group_permissions
						(group_id, permission_id, scope_type, scope_ref)
					SELECT
						g.id, p.id, 'global', NULL FROM groups g, permissions p
					WHERE g.name = 'custom' LIMIT 1;`)
			},
			invoke: func(t *testing.T, store *Store) error {
				var id string
				if err := store.db.QueryRow(
					`SELECT id FROM groups WHERE name = 'custom';`,
				).Scan(&id); err != nil {
					t.Fatalf("DB.QueryRow().Scan() failed: %v", err)
				}
				_, err := store.GroupByID(t.Context(), id)
				return err
			},
			errRegex: test.TrimYAML(`
				^list group's grants:
					scan grant:
						[^\s]+.*name "scope_ref":
							converting NULL to string is unsupported$`),
		},
		{
			name: "GroupByID/NULL group name",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "groups")
				test.ExecSQL(t, store.db, `
					CREATE TABLE groups (
						id TEXT,
						name TEXT,
						description TEXT DEFAULT '',
						system INTEGER DEFAULT 0,
						seed_key TEXT,
						created_at DATETIME,
						updated_at DATETIME
					);`,
				)
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO groups
						(id, name, created_at, updated_at)
					VALUES
						('gid', NULL, '%s', '%s');`,
					timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.GroupByID(t.Context(), "gid")
				return err
			},
			errRegex: test.TrimYAML(`
				^scan group:
					[^\s]+.*name "name":
						converting NULL to string is unsupported$`),
		},
		{
			name: "UserByID/NULL username",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "users")
				test.ExecSQL(t, store.db, `
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
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO users
						(id, username, password_hash, created_at, updated_at)
					VALUES
						('uid', NULL, 'x', '%s', '%s');`,
					timeNow().Format(timeFormat), timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.UserByID(t.Context(), "uid")
				return err
			},
			errRegex: test.TrimYAML(`
				^scan user:
					[^\s]+.*name "username":
						converting NULL to string is unsupported$`),
		},
		{
			name: "UserWithGrants/corrupt timestamps",
			setup: func(t *testing.T, store *Store) {
				test.ExecSQL(t, store.db, `
					INSERT INTO users
						(id, username, password_hash, created_at, updated_at)
					VALUES
						('corrupt', 'corrupt', 'x', 'not-a-time', 'not-a-time');`)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.UserWithGrants(t.Context(), "corrupt")
				return err
			},
			errRegex: test.TrimYAML(`
				^scan user with grants:
					[^\s]+.*name "created_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "RenameServiceScopeRefs/NULL grant group_id",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "group_permissions")
				test.ExecSQL(t, store.db, `
					CREATE TABLE group_permissions (
						group_id TEXT,
						permission_id INTEGER,
						scope_type TEXT,
						scope_ref TEXT
					);`,
				)
				test.ExecSQL(t, store.db, `
					INSERT INTO group_permissions
						(group_id, permission_id, scope_type, scope_ref)
					VALUES
						(NULL, 1, 'service', 'old-name');`)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.RenameServiceScopeRefs(t.Context(), "old-name", "new-name")
				return err
			},
			errRegex: test.TrimYAML(`
				^scan service-scoped grant:
					[^\s]+.*name "group_id":
						converting NULL to string is unsupported$`),
		},
		{
			name: "New/NULL permission row",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "permissions")
				test.ExecSQL(t, store.db, `
				CREATE TABLE permissions
					(id INTEGER PRIMARY KEY AUTOINCREMENT, resource TEXT, action TEXT);`,
				)
				test.ExecSQL(t, store.db, `
					INSERT INTO permissions
						(resource, action)
					VALUES
						(NULL, NULL);`,
				)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := New(t.Context(), store.db)
				return err
			},
			errRegex: test.TrimYAML(`
				^auth store sync permissions:
					scan permission:
						[^\s]+.*name "resource":
							converting NULL to string is unsupported$`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			tc.setup(t, store)

			err := tc.invoke(t, store)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestStore__rowsErrOnSecondPass(t *testing.T) {
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
					t.Fatalf("DB.QueryRow().Scan() failed: %v", err)
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
			name:       "TrimSessionsForUser/over-cap iteration fails",
			failOnCall: 1,
			setup: func(t *testing.T, store *Store) {
				mustCreateSessions(t, store, "target", 2)
			},
			invoke: func(t *testing.T, store *Store) error {
				// Resolve the ID without iterating rows (the fault is armed).
				_, err := store.TrimSessionsForUser(t.Context(),
					mustUserID(t, store, "target"), 1)
				return err
			},
		},
		{
			name:       "UserWithGrants/iteration fails",
			failOnCall: 1,
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "target", "", GroupViewer)
			},
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.UserWithGrants(t.Context(),
					mustUserID(t, store, "target"))
				return err
			},
		},
		{
			name:       "RenameServiceScopeRefs/source iteration fails",
			failOnCall: 1,
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.RenameServiceScopeRefs(t.Context(), "old-name", "new-name")
				return err
			},
		},
		{
			name:       "RenameServiceScopeRefs/target iteration fails",
			failOnCall: 2, // 1st: the source scope; 2nd: the target scope.
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.RenameServiceScopeRefs(t.Context(), "old-name", "new-name")
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
			// Serial: swaps rowsErr.

			store := testStore(t)
			if tc.setup != nil {
				tc.setup(t, store)
			}
			overrideRowsErrOnCall(t, tc.failOnCall)

			err := tc.invoke(t, store)

			if !errors.Is(err, errInjected) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, errInjected)
			}
		})
	}
}

func TestStore_InTx__beginError(t *testing.T) {
	db := test.MemoryDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	_ = db.Close()

	_, err = store.CreateUser(t.Context(), "u", "", "", testPasswordHash, nil)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^begin transaction:
			sql: database is closed$`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}
