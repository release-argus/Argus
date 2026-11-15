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
	"strings"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestNew(t *testing.T) {
	db := test.MemoryDB(t)

	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := test.CountRows(t, store.db, "permissions"), cataloguePairCount(); got != want {
		t.Errorf("permissions row count mismatch\ngot:  %d\nwant: %d", got, want)
	}

	if got, want := test.CountRows(t, store.db, "user_preferences"), 0; got != want {
		t.Errorf("user_preferences row count mismatch\ngot:  %d\nwant: %d", got, want)
	}

	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf("Store.Groups() failed: %v", err)
	}
	wantSystem := map[string]bool{
		GroupAdmin:    true,
		GroupOperator: false,
		GroupViewer:   false,
	}
	if got, want := len(groups), len(wantSystem); got != want {
		t.Fatalf("len(Store.Groups()) mismatch\ngot:  %d\nwant: %d", got, want)
	}
	for _, group := range groups {
		want, known := wantSystem[group.Name]
		if !known {
			t.Errorf(
				"Group.Name mismatch\ngot:  %q\nwant: one of %q",
				group.Name, []string{GroupAdmin, GroupOperator, GroupViewer},
			)
			continue
		}
		if group.System != want {
			t.Errorf(
				"Group[%q].System mismatch\ngot:  %t\nwant: %t",
				group.Name, group.System, want,
			)
		}

		if group.SeedKey != group.Name {
			t.Errorf(
				"Group[%q].SeedKey mismatch\ngot:  %q\nwant: %q",
				group.Name, group.SeedKey, group.Name,
			)
		}

		wantGrants := map[string]int{
			GroupAdmin:    cataloguePairCount(),
			GroupOperator: seedGrantCount(t, GroupOperator),
			GroupViewer:   seedGrantCount(t, GroupViewer),
		}[group.Name]
		if len(group.Grants) != wantGrants {
			t.Errorf(
				"len(Group[%q].Grants) mismatch\ngot:  %d\nwant: %d",
				group.Name, len(group.Grants), wantGrants,
			)
		}
	}

	for _, group := range groups {
		for _, grant := range group.Grants {
			if !grant.Valid() {
				t.Errorf(
					"Group[%q].Grants[%q].Valid() mismatch\ngot:  false\nwant: true",
					group.Name, grantString(grant),
				)
			}
		}
	}

	byName := groupsByName(t, store)
	remitChecks := []struct {
		name     string
		group    string
		resource rbac.Resource
		action   rbac.Action
		want     bool
	}{
		{name: "operator/metric:read", group: GroupOperator, resource: rbac.ResourceMetric, action: rbac.ActionRead, want: false},
		{name: "operator/notify:execute", group: GroupOperator, resource: rbac.ResourceNotify, action: rbac.ActionExecute, want: true},
		{name: "operator/service:create", group: GroupOperator, resource: rbac.ResourceService, action: rbac.ActionCreate, want: true},
		{name: "viewer/config:read", group: GroupViewer, resource: rbac.ResourceConfig, action: rbac.ActionRead, want: true},
		{name: "viewer/metric:read", group: GroupViewer, resource: rbac.ResourceMetric, action: rbac.ActionRead, want: false},
		{name: "viewer/service:create", group: GroupViewer, resource: rbac.ResourceService, action: rbac.ActionCreate, want: false},
		{name: "viewer/service:read", group: GroupViewer, resource: rbac.ResourceService, action: rbac.ActionRead, want: true},
	}
	for _, check := range remitChecks {
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()

			group := groupNamed(t, byName, check.group)
			if got := groupHasGlobalGrant(group, check.resource, check.action); got != check.want {
				t.Errorf("groupHasGlobalGrant() mismatch\ngot:  %t\nwant: %t", got, check.want)
			}
		})
	}
}

func TestNew__idempotent(t *testing.T) {
	db := test.MemoryDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	// User renames viewer's description and strips all grants from every group.
	newDescription := "customised"
	if _, err := store.db.Exec(`
			UPDATE groups
			SET description = ?
			WHERE name = ?;`,
		newDescription, GroupViewer,
	); err != nil {
		t.Fatalf("update: DB.Exec() failed: %v", err)
	}
	if _, err := store.db.Exec(
		`DELETE FROM group_permissions;`,
	); err != nil {
		t.Fatalf("delete: DB.Exec() failed: %v", err)
	}

	store, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := test.CountRows(t, store.db, "groups"); got != 3 {
		t.Errorf("groups row count mismatch\ngot:  %d\nwant: 3", got)
	}

	var description string
	if err := store.db.QueryRow(
		`SELECT description FROM groups WHERE name = ?;`, GroupViewer,
	).Scan(&description); err != nil || description != newDescription {
		t.Errorf(
			"viewer: Group.Description mismatch\ngot:  %q, err=%v\nwant: %q, err=nil",
			description, err, newDescription,
		)
	}

	// Starters stay stripped: only admin is restored to the full catalogue.
	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf("Store.Groups() failed: %v", err)
	}
	for _, group := range groups {
		switch group.Name {
		case GroupAdmin:
			if len(group.Grants) != cataloguePairCount() {
				t.Errorf(
					"len(Group[%q].Grants) mismatch\ngot:  %d\nwant: %d",
					group.Name, len(group.Grants), cataloguePairCount(),
				)
			}
		default:
			if len(group.Grants) != 0 {
				t.Errorf(
					"len(Group[%q].Grants) mismatch\ngot:  %d\nwant: 0",
					group.Name, len(group.Grants),
				)
			}
		}
	}
}

func TestNew__starterGroupDeletionPersists(t *testing.T) {
	db := test.MemoryDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	operator := mustGroupByName(t, store, GroupOperator)
	if err := store.DeleteGroup(t.Context(), operator.ID); err != nil {
		t.Fatalf("Store.DeleteGroup() failed: %v", err)
	}

	store, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only a fresh install seeds starters.
	byName := groupsByName(t, store)
	if _, present := byName[GroupOperator]; present {
		t.Error("byName[GroupOperator] mismatch\ngot:  present\nwant: absent (deleted)")
	}
	for _, want := range []string{GroupAdmin, GroupViewer} {
		if _, present := byName[want]; !present {
			t.Errorf("byName[%q] mismatch\ngot:  absent\nwant: present", want)
		}
	}
}

func TestNew__starterGroupRenamePersists(t *testing.T) {
	db := test.MemoryDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	viewer := mustGroupByName(t, store, GroupViewer)
	newName := "watchers"
	if _, err := store.UpdateGroup(
		t.Context(),
		viewer.ID,
		GroupPatch{Name: &newName},
	); err != nil {
		t.Fatalf("Store.UpdateGroup() failed: %v", err)
	}

	store, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Seeding finds the group by its seed_key.
	byName := groupsByName(t, store)
	if got, want := len(byName), 3; got != want {
		t.Errorf("len(byName) mismatch\ngot:  %d\nwant: %d", got, want)
	}
	if _, present := byName[GroupViewer]; present {
		t.Error("byName[GroupViewer] mismatch\ngot:  present\nwant: absent (renamed)")
	}
	renamed, present := byName[newName]
	if !present {
		t.Fatal("byName[newName] mismatch\ngot:  absent\nwant: present")
	}
	if renamed.SeedKey != GroupViewer {
		t.Errorf("Group.SeedKey mismatch\ngot:  %q\nwant: %q", renamed.SeedKey, GroupViewer)
	}
}

func TestNew__permissionsIntroducedLater(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, store *Store)
		check func(t *testing.T, byName map[string]Group)
	}{
		{
			name: "admin only/new non-read permission stays out of starter groups",
			setup: func(t *testing.T, store *Store) {
				removePermission(t, store.db, rbac.ResourceNotify, rbac.ActionExecute)
			},
			check: func(t *testing.T, byName map[string]Group) {
				r, a := rbac.ResourceNotify, rbac.ActionExecute
				if !groupHasGlobalGrant(groupNamed(t, byName, GroupAdmin), r, a) {
					t.Error(
						"groupHasGlobalGrant(byName[GroupAdmin]) mismatch\ngot:  false\nwant: true",
					)
				}
				for _, groupName := range []string{GroupOperator, GroupViewer} {
					if groupHasGlobalGrant(groupNamed(t, byName, groupName), r, a) {
						t.Errorf(
							"groupHasGlobalGrant(byName[%q]) mismatch\ngot:  true\nwant: false",
							groupName,
						)
					}
				}
			},
		},
		{
			name: "admin only/new read permission stays out of starter groups",
			setup: func(t *testing.T, store *Store) {
				removePermission(t, store.db, rbac.ResourceConfig, rbac.ActionRead)
			},
			check: func(t *testing.T, byName map[string]Group) {
				r, a := rbac.ResourceConfig, rbac.ActionRead
				if !groupHasGlobalGrant(groupNamed(t, byName, GroupAdmin), r, a) {
					t.Error(
						"groupHasGlobalGrant(byName[GroupAdmin]) mismatch\ngot:  false\nwant: true",
					)
				}
				for _, groupName := range []string{GroupOperator, GroupViewer} {
					if groupHasGlobalGrant(groupNamed(t, byName, groupName), r, a) {
						t.Errorf(
							"groupHasGlobalGrant(byName[%q]) mismatch\ngot:  true\nwant: false",
							groupName,
						)
					}
				}
			},
		},
		{
			name: "user-stripped grants stay removed",
			setup: func(t *testing.T, store *Store) {
				// User strips operator's service:create beforehand.
				if _, err := store.db.Exec(`
					DELETE FROM group_permissions
					WHERE group_id = (SELECT id FROM groups WHERE seed_key = ?)
						AND permission_id =
							(SELECT id FROM permissions WHERE resource = ? AND action = ?);`,
					GroupOperator, string(rbac.ResourceService), string(rbac.ActionCreate),
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			},
			check: func(t *testing.T, byName map[string]Group) {
				r, a := rbac.ResourceService, rbac.ActionCreate
				if groupHasGlobalGrant(groupNamed(t, byName, GroupOperator), r, a) {
					t.Error(
						"groupHasGlobalGrant(byName[GroupOperator]) mismatch\n" +
							"got:  true\nwant: false",
					)
				}
			},
		},
		{
			name: "deleted group/skipped",
			setup: func(t *testing.T, store *Store) {
				removePermission(t, store.db, rbac.ResourceNotify, rbac.ActionExecute)
				operator := mustGroupByName(t, store, GroupOperator)
				if err := store.DeleteGroup(t.Context(), operator.ID); err != nil {
					t.Fatalf("Store.DeleteGroup() failed: %v", err)
				}
			},
			check: func(t *testing.T, byName map[string]Group) {
				if _, present := byName[GroupOperator]; present {
					t.Error("byName[GroupOperator] mismatch\ngot:  present\nwant: absent (deleted)")
				}
			},
		},
		{
			name: "renamed group/grants left untouched",
			setup: func(t *testing.T, store *Store) {
				removePermission(t, store.db, rbac.ResourceNotify, rbac.ActionExecute)
				operator := mustGroupByName(t, store, GroupOperator)
				newName := "ops"
				if _, err := store.UpdateGroup(
					t.Context(),
					operator.ID,
					GroupPatch{Name: &newName},
				); err != nil {
					t.Fatalf("Store.UpdateGroup() failed: %v", err)
				}
			},
			check: func(t *testing.T, byName map[string]Group) {
				r, a := rbac.ResourceNotify, rbac.ActionExecute
				if groupHasGlobalGrant(groupNamed(t, byName, "ops"), r, a) {
					t.Error(
						"groupHasGlobalGrant(byName[\"ops\"]) mismatch\ngot:  true\nwant: false",
					)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := test.MemoryDB(t)
			store, err := New(t.Context(), db)
			if err != nil {
				t.Fatalf("first: New() failed: %v", err)
			}
			tc.setup(t, store)

			store, err = New(t.Context(), db)
			if err != nil {
				t.Fatalf("second: New() failed: %v", err)
			}

			tc.check(t, groupsByName(t, store))
		})
	}
}

func TestNew__retiresStalePermissions(t *testing.T) {
	db := test.MemoryDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf("first: New() failed: %v", err)
	}
	test.ExecSQL(
		t, store.db,
		`INSERT INTO permissions (resource, action)
		VALUES ('legacy', 'shazam');`,
		`INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
		VALUES (
			(SELECT id FROM groups WHERE name = 'viewer'),
			(SELECT id FROM permissions WHERE resource = 'legacy'),
			'global',
			''
		);`,
	)

	// A catalogued permission wrongly marked retired.
	test.ExecSQL(
		t, store.db,
		`UPDATE permissions SET retired_at = '2020-01-01T00:00:00Z'
		WHERE resource = 'config' AND action = 'read';`,
	)

	_, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf("second: New() failed: %v", err)
	}

	// Retired rather than deleted, so a downgrade keeps the grants referencing it.
	var count int
	if err := store.db.QueryRow(
		`SELECT COUNT(*) FROM permissions
		WHERE resource = 'legacy' AND retired_at IS NOT NULL;`,
	).Scan(&count); err != nil || count != 1 {
		t.Errorf(
			"retired: permissions row count mismatch\ngot:  %d, err=%v\nwant: 1, err=nil",
			count, err,
		)
	}

	if err := store.db.QueryRow(`
		SELECT COUNT(*) FROM group_permissions gp
		JOIN permissions p ON p.id = gp.permission_id
		WHERE p.resource = 'legacy';`,
	).Scan(&count); err != nil || count != 1 {
		t.Errorf(
			"retired grant: group_permissions row count mismatch\n"+
				"got:  %d, err=%v\nwant: 1, err=nil",
			count, err,
		)
	}

	if err := store.db.QueryRow(`
		SELECT COUNT(*) FROM group_permissions gp
		LEFT JOIN permissions p ON p.id = gp.permission_id
		WHERE p.id IS NULL;`,
	).Scan(&count); err != nil || count != 0 {
		t.Errorf(
			"orphaned: group_permissions row count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil",
			count, err,
		)
	}

	if err := store.db.QueryRow(
		`SELECT COUNT(*) FROM permissions
		WHERE resource = 'config' AND action = 'read' AND retired_at IS NULL;`,
	).Scan(&count); err != nil || count != 1 {
		t.Errorf(
			"un-retired: permissions row count mismatch\ngot:  %d, err=%v\nwant: 1, err=nil",
			count, err,
		)
	}

	var retiredAt string
	if err := store.db.QueryRow(
		`SELECT retired_at FROM permissions WHERE resource = 'legacy';`,
	).Scan(&retiredAt); err != nil {
		t.Fatalf("before: DB.QueryRow().Scan() failed: %v", err)
	}
	_, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf("third: New() failed: %v", err)
	}
	var retiredAtAfter string
	if err := store.db.QueryRow(
		`SELECT retired_at FROM permissions WHERE resource = 'legacy';`,
	).Scan(&retiredAtAfter); err != nil || retiredAtAfter != retiredAt {
		t.Errorf(
			"after: permissions.retired_at mismatch\ngot:  %q, err=%v\nwant: %q, err=nil",
			retiredAtAfter, err, retiredAt,
		)
	}
}

func TestNew__errors(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, db *sql.DB)
		errRegex string
	}{
		{
			name: "closed database",
			errRegex: test.TrimYAML(`
				^auth store migrate:
					create schema_migrations:
						sql: database is closed$`),
			setup: func(_ *testing.T, db *sql.DB) {
				_ = db.Close()
			},
		},
		{
			name: "malformed schema_migrations rows",
			errRegex: test.TrimYAML(`
				^auth store migrate:
					scan schema_migrations:
						sql: Scan error on column index 0, name "version"`),
			setup: func(t *testing.T, db *sql.DB) {
				test.ExecSQL(
					t, db,
					`CREATE TABLE schema_migrations (version TEXT, applied_at DATETIME);`,
					`INSERT INTO schema_migrations (version, applied_at) VALUES (NULL, 'x');`,
				)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := test.MemoryDB(t)
			tc.setup(t, db)

			_, err := New(t.Context(), db)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestNew__Migrate__badMigration(t *testing.T) {
	migrationsHad := migrations
	migrations = append(
		append([]migration{}, migrations...),
		migration{version: 99, statements: []string{`THIS IS NOT SQL;`}},
	)
	t.Cleanup(func() { migrations = migrationsHad })

	db := test.MemoryDB(t)

	_, err := New(t.Context(), db)
	if err == nil {
		t.Fatal("error mismatch\ngot:  nil\nwant: non-nil")
	}
}

func TestNew__Migrate__rowsErr(t *testing.T) {
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	db := test.MemoryDB(t)

	// migrate iterates rows first.
	_, err := New(t.Context(), db)

	if !errors.Is(err, wantErr) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, wantErr)
	}
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^auth store migrate:
			iterate schema_migrations:
				iteration broke`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestMigrate__hotQueriesUseAnIndex(t *testing.T) {
	store := testStore(t)

	tests := []struct {
		name      string
		query     string
		args      []any
		wantIndex string
	}{
		{
			name: "sessions/trim over the per-user cap",
			query: `SELECT token_hash FROM sessions WHERE user_id = ?
				ORDER BY last_seen_at DESC, created_at DESC, rowid DESC
				LIMIT -1 OFFSET ?;`,
			args:      []any{"user", 10},
			wantIndex: "idx_sessions_user_id",
		},
		{
			name:      "sessions/delete every session of a user",
			query:     `DELETE FROM sessions WHERE user_id = ?;`,
			args:      []any{"user"},
			wantIndex: "idx_sessions_user_id",
		},
		{
			name:      "sessions/prune the expired",
			query:     `DELETE FROM sessions WHERE expires_at < ?;`,
			args:      []any{"2026-01-01T00:00:00.000000000Z"},
			wantIndex: "idx_sessions_expires_at",
		},
		{
			name:      "api_tokens/list a user's tokens",
			query:     `SELECT id FROM api_tokens WHERE user_id = ? ORDER BY created_at DESC;`,
			args:      []any{"user"},
			wantIndex: "idx_api_tokens_user_id",
		},
		{
			name:      "user_preferences/load a user's preferences",
			query:     `SELECT hide, view, timestamps FROM user_preferences WHERE user_id = ?;`,
			args:      []any{"user"},
			wantIndex: "sqlite_autoindex_user_preferences_1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := store.db.QueryContext(t.Context(), "EXPLAIN QUERY PLAN "+tc.query, tc.args...)
			if err != nil {
				t.Fatalf("DB.QueryContext() failed: %v", err)
			}
			defer rows.Close()

			var planLines []string
			for rows.Next() {
				var id, parent, notUsed int
				var detail string
				if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
					t.Fatalf("Rows.Scan() failed: %v", err)
				}
				planLines = append(planLines, detail)
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("Rows.Err() failed: %v", err)
			}
			plan := strings.Join(planLines, "; ")

			if !strings.Contains(plan, tc.wantIndex) {
				t.Errorf("query plan mismatch\ngot:  %q\nwant: %q", plan, tc.wantIndex)
			}
			for _, scan := range []string{"SCAN sessions", "SCAN api_tokens", "SCAN user_preferences"} {
				if strings.Contains(plan, scan) {
					t.Errorf("query plan mismatch\ngot:  %q\nwant: NOT %q", plan, scan)
				}
			}
		})
	}
}

func TestNew__SeedGroups_GroupIDBySeedKeyError(t *testing.T) {
	db := test.MemoryDB(t)
	if _, err := New(t.Context(), db); err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	test.ExecSQL(
		t, db,
		`DROP TABLE groups;`,
		`CREATE TABLE groups (id TEXT NOT NULL PRIMARY KEY);`,
	)

	// seedGroups looks groups up by seed_key.
	_, err := New(t.Context(), db)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^auth store seed groups:
			look up seed group "admin":
				SQL logic error: no such column: seed_key`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestFullCatalogueGrants(t *testing.T) {
	grants := fullCatalogueGrants()

	if got, want := len(grants), cataloguePairCount(); got != want {
		t.Fatalf("len(grants) mismatch\ngot:  %d\nwant: %d", got, want)
	}
	for _, grant := range grants {
		if !grant.Valid() || grant.Scope.Type != rbac.ScopeGlobal {
			t.Errorf(
				"Grant mismatch\ngot:  %s, Valid()=%t\nwant: Valid()=true, scope=%s",
				grantString(grant), grant.Valid(), rbac.ScopeGlobal,
			)
		}
	}
}

func TestStore_InTx__commitError(t *testing.T) {
	store := testStore(t)

	err := store.inTx(t.Context(), func(tx *sql.Tx) error {
		_ = tx.Commit()
		return nil
	})

	if !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, sql.ErrTxDone)
	}
}

func TestStore_InTx__rollsBackOnError(t *testing.T) {
	db := test.MemoryDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	before := test.CountRows(t, store.db, "groups")

	wantErr := errors.New("abandon the transaction")
	gotErr := store.inTx(t.Context(), func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), `
			INSERT INTO groups (id, name, description, system, created_at, updated_at)
			VALUES ('rollback-me', 'rollback-me', '', 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');`,
		); err != nil {
			t.Fatalf("Tx.ExecContext() failed: %v", err)
		}
		return wantErr
	})

	e := errfmt.FormatError(gotErr)
	errRegex := `^abandon the transaction$`
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
	if !errors.Is(gotErr, wantErr) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", gotErr, wantErr)
	}

	if got, want := test.CountRows(t, store.db, "groups"), before; got != want {
		t.Errorf("groups row count mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestParseTime__invalid(t *testing.T) {
	ts := "not-a-timestamp"

	_, err := parseTime(ts)
	if err == nil {
		t.Error("error mismatch\ngot:  nil\nwant: non-nil")
	}
}
