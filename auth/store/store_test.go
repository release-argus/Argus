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
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
)

func TestNew(t *testing.T) {
	// GIVEN: a fresh database.
	db := testDB(t)

	// WHEN: New is called.
	store, err := New(t.Context(), db)

	prefix := fmt.Sprintf("%s\nNew()", packageName)

	// THEN: it succeeds.
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// AND: the permission catalogue is synced.
	if got, want := countRows(t, store, "permissions"), cataloguePairCount(); got != want {
		t.Errorf(
			"%s permissions row count mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: the three groups are seeded (admin protected; the rest not).
	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf(
			"%s Groups() failed: %v",
			prefix, err,
		)
	}
	wantSystem := map[string]bool{
		GroupAdmin:    true,
		GroupOperator: false,
		GroupViewer:   false,
	}
	if got, want := len(groups), len(wantSystem); got != want {
		t.Fatalf(
			"%s group count mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
	for _, group := range groups {
		want, known := wantSystem[group.Name]
		if !known {
			t.Errorf(
				"%s unexpected group %q",
				prefix, group.Name,
			)
			continue
		}
		if group.System != want {
			t.Errorf(
				"%s group %q system mismatch\ngot:  %t\nwant: %t",
				prefix, group.Name, group.System, want,
			)
		}

		// AND: each seeded group carries its immutable seed_key.
		if group.SeedKey != group.Name {
			t.Errorf(
				"%s group %q seed_key mismatch\ngot:  %q\nwant: %q",
				prefix, group.Name, group.SeedKey, group.Name,
			)
		}

		// AND: admin holds the full catalogue; the others hold their remits.
		wantGrants := map[string]int{
			GroupAdmin:    cataloguePairCount(),
			GroupOperator: seedGrantCount(t, GroupOperator),
			GroupViewer:   seedGrantCount(t, GroupViewer),
		}[group.Name]
		if len(group.Grants) != wantGrants {
			t.Errorf(
				"%s group %q grant count mismatch\ngot:  %d\nwant: %d",
				prefix, group.Name, len(group.Grants), wantGrants,
			)
		}
	}

	// AND: every seeded grant is valid against the catalogue.
	for _, group := range groups {
		for _, grant := range group.Grants {
			if !grant.Valid() {
				t.Errorf(
					"%s group %q holds an invalid grant: %+v",
					prefix, group.Name, grant,
				)
			}
		}
	}

	// AND: the remits hold.
	byName := groupsByName(t, store)
	remitChecks := []struct {
		group    string
		resource rbac.Resource
		action   rbac.Action
		want     bool
	}{
		{GroupOperator, rbac.ResourceService, rbac.ActionCreate, true},
		{GroupOperator, rbac.ResourceNotify, rbac.ActionExecute, true},
		{GroupViewer, rbac.ResourceService, rbac.ActionRead, true},
		{GroupViewer, rbac.ResourceConfig, rbac.ActionRead, true},
		{GroupViewer, rbac.ResourceService, rbac.ActionCreate, false},
	}
	for _, check := range remitChecks {
		if got := groupHasGlobalGrant(byName[check.group], check.resource, check.action); got != check.want {
			t.Errorf(
				"%s group %q grant on %s:%s\ngot:  %v\nwant: %v",
				prefix, check.group, check.resource, check.action, got, check.want,
			)
		}
	}
}

func TestNew__idempotent(t *testing.T) {
	// GIVEN: a database already initialised once, with user changes on top.
	db := testDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s\nfirst New() failed: %v",
			packageName, err,
		)
	}
	// User renames viewer's description and strips admin grants directly.
	newDescription := "customised"
	if _, err := store.db.Exec(`
			UPDATE groups
			SET description = '`+newDescription+`'
			WHERE name = ?;`,
		GroupViewer,
	); err != nil {
		t.Fatalf(
			"%s\nsetup update failed: %v",
			packageName, err,
		)
	}
	if _, err := store.db.Exec(
		`DELETE FROM group_permissions;`,
	); err != nil {
		t.Fatalf(
			"%s\nsetup delete failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nNew() again", packageName)

	// WHEN: New runs again on the same database.
	store, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: no duplicate groups appear.
	if got := countRows(t, store, "groups"); got != 3 {
		t.Errorf(
			"%s group count mismatch\ngot:  %d\nwant: 3",
			prefix, got,
		)
	}

	// AND: user customisation of non-admin groups survives.
	var description string
	if err := store.db.QueryRow(
		`SELECT description FROM groups WHERE name = ?;`, GroupViewer,
	).Scan(&description); err != nil || description != newDescription {
		t.Errorf(
			"%s viewer description mismatch\ngot:  %q, err=%v\nwant: %q",
			prefix, description, err, newDescription,
		)
	}

	// AND: admin grants are restored to the full catalogue.
	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf(
			"%s Groups() failed: %v",
			prefix, err,
		)
	}
	for _, group := range groups {
		switch group.Name {
		case GroupAdmin:
			if len(group.Grants) != cataloguePairCount() {
				t.Errorf(
					"%s %s grants not re-synced\ngot:  %d\nwant: %d",
					prefix, group.Name, len(group.Grants), cataloguePairCount(),
				)
			}
		default:
			if len(group.Grants) != 0 {
				t.Errorf(
					"%s %s grants shouldn't have re-synced\ngot:  %d\nwant: 0",
					prefix, group.Name, len(group.Grants),
				)
			}
		}
	}
}

func TestNew__starterGroupDeletionPersists(t *testing.T) {
	// GIVEN: an initialised database with a deleted starter group.
	db := testDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s\nfirst New() failed: %v",
			packageName, err,
		)
	}
	operator, ok := groupsByName(t, store)[GroupOperator]
	if !ok {
		t.Fatalf("%s\noperator group not seeded", packageName)
	}
	if err := store.DeleteGroup(t.Context(), operator.ID); err != nil {
		t.Fatalf(
			"%s\nDeleteGroup failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nNew() after deleting a starter group", packageName)

	// WHEN: New runs again.
	store, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: the deleted starter group is not re-created, while the others remain.
	byName := groupsByName(t, store)
	if _, present := byName[GroupOperator]; present {
		t.Errorf("%s operator group was re-added after deletion", prefix)
	}
	for _, want := range []string{GroupAdmin, GroupViewer} {
		if _, present := byName[want]; !present {
			t.Errorf(
				"%s group %q should still be present",
				prefix, want,
			)
		}
	}
}

func TestNew__starterGroupRenamePersists(t *testing.T) {
	// GIVEN: an initialised database whose admin has renamed a starter group.
	db := testDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s\nfirst New() failed: %v",
			packageName, err,
		)
	}
	viewer, ok := groupsByName(t, store)[GroupViewer]
	if !ok {
		t.Fatalf("%s\nviewer group not seeded", packageName)
	}
	newName := "watchers"
	if _, err := store.UpdateGroup(
		t.Context(),
		viewer.ID,
		GroupPatch{Name: &newName},
	); err != nil {
		t.Fatalf(
			"%s\nUpdateGroup failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nNew() after renaming a starter group", packageName)

	// WHEN: New runs again.
	store, err = New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: the original name is not re-seeded; the renamed group persists,
	// keeping its immutable seed_key.
	byName := groupsByName(t, store)
	if got, want := len(byName), 3; got != want {
		t.Errorf(
			"%s group count mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
	if _, present := byName[GroupViewer]; present {
		t.Errorf(
			"%s original %q group name was re-seeded",
			prefix, GroupViewer,
		)
	}
	renamed, present := byName[newName]
	if !present {
		t.Fatalf(
			"%s renamed group %q missing",
			prefix, newName,
		)
	}
	if renamed.SeedKey != GroupViewer {
		t.Errorf(
			"%s seed_key mismatch\ngot:  %q\nwant: %q",
			prefix, renamed.SeedKey, GroupViewer,
		)
	}
}

func TestNew__permissionsIntroducedLater(t *testing.T) {
	// GIVEN: initialised databases missing a permission pair (simulating a database
	// seeded before the pair entered the catalogue), with various user edits on top.
	tests := []struct {
		name  string
		setup func(t *testing.T, store *Store)
		check func(t *testing.T, prefix string, byName map[string]Group)
	}{
		{
			name: "admin only/new non-read permission stays out of starter groups",
			setup: func(t *testing.T, store *Store) {
				removePermission(t, store.db, rbac.ResourceNotify, rbac.ActionExecute)
			},
			check: func(t *testing.T, prefix string, byName map[string]Group) {
				if !groupHasGlobalGrant(byName[GroupAdmin], rbac.ResourceNotify, rbac.ActionExecute) {
					t.Errorf("%s admin should have gained notify:execute", prefix)
				}
				for _, groupName := range []string{GroupOperator, GroupViewer} {
					if groupHasGlobalGrant(byName[groupName], rbac.ResourceNotify, rbac.ActionExecute) {
						t.Errorf(
							"%s %s should not have gained notify:execute",
							prefix, groupName,
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
			check: func(t *testing.T, prefix string, byName map[string]Group) {
				if !groupHasGlobalGrant(byName[GroupAdmin], rbac.ResourceConfig, rbac.ActionRead) {
					t.Errorf("%s admin should have gained config:read", prefix)
				}
				for _, groupName := range []string{GroupOperator, GroupViewer} {
					if groupHasGlobalGrant(byName[groupName], rbac.ResourceConfig, rbac.ActionRead) {
						t.Errorf(
							"%s %s should not have gained config:read",
							prefix, groupName,
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
					t.Fatalf(
						"%s\nsetup strip failed: %v",
						packageName, err,
					)
				}
			},
			check: func(t *testing.T, prefix string, byName map[string]Group) {
				// The stripped grant is not restored.
				if groupHasGlobalGrant(byName[GroupOperator], rbac.ResourceService, rbac.ActionCreate) {
					t.Errorf("%s operator's stripped service:create was restored", prefix)
				}
			},
		},
		{
			name: "deleted group/skipped",
			setup: func(t *testing.T, store *Store) {
				removePermission(t, store.db, rbac.ResourceNotify, rbac.ActionExecute)
				operator := groupsByName(t, store)[GroupOperator]
				if err := store.DeleteGroup(t.Context(), operator.ID); err != nil {
					t.Fatalf(
						"%s\nsetup DeleteGroup failed: %v",
						packageName, err,
					)
				}
			},
			check: func(t *testing.T, prefix string, byName map[string]Group) {
				if _, present := byName[GroupOperator]; present {
					t.Errorf("%s deleted operator group was re-created", prefix)
				}
			},
		},
		{
			name: "renamed group/grants left untouched",
			setup: func(t *testing.T, store *Store) {
				removePermission(t, store.db, rbac.ResourceNotify, rbac.ActionExecute)
				operator := groupsByName(t, store)[GroupOperator]
				newName := "ops"
				if _, err := store.UpdateGroup(
					t.Context(),
					operator.ID,
					GroupPatch{Name: &newName},
				); err != nil {
					t.Fatalf(
						"%s\nsetup UpdateGroup failed: %v",
						packageName, err,
					)
				}
			},
			check: func(t *testing.T, prefix string, byName map[string]Group) {
				if groupHasGlobalGrant(byName["ops"], rbac.ResourceNotify, rbac.ActionExecute) {
					t.Errorf(
						"%s renamed operator group should not have gained notify:execute",
						prefix,
					)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := testDB(t)
			store, err := New(t.Context(), db)
			if err != nil {
				t.Fatalf(
					"%s\nfirst New() failed: %v",
					packageName, err,
				)
			}
			tc.setup(t, store)

			prefix := fmt.Sprintf("%s\nNew() with a late permission (%s)", packageName, tc.name)

			// WHEN: New runs again (the removed pair re-enters the catalogue).
			store, err = New(t.Context(), db)
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// THEN: the new pair flows to the surviving seed groups' remits.
			tc.check(t, prefix, groupsByName(t, store))
		})
	}
}

func TestNew_RemovesStalePermissions(t *testing.T) {
	// GIVEN: an initialised database holding a permission (and a grant on it)
	// that is no longer in the code catalogue.
	db := testDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s\nfirst New() failed: %v",
			packageName, err,
		)
	}
	if _, err := store.db.Exec(
		`INSERT INTO permissions (resource, action)
		VALUES ('legacy', 'shazam');`,
	); err != nil {
		t.Fatalf(
			"%s\nsetup insert failed: %v",
			packageName, err,
		)
	}
	if _, err := store.db.Exec(`
		INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
		VALUES (
			(SELECT id FROM groups WHERE name = 'viewer'),
			(SELECT id FROM permissions WHERE resource = 'legacy'),
			'global',
			''
		);`,
	); err != nil {
		t.Fatalf(
			"%s\nsetup grant insert failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nNew() with stale permission", packageName)

	// WHEN: New runs again.
	if _, err := New(t.Context(), db); err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: the stale permission and its grants are gone.
	var count int
	if err := store.db.QueryRow(
		`SELECT COUNT(*) FROM permissions WHERE resource = 'legacy';`,
	).Scan(&count); err != nil || count != 0 {
		t.Errorf(
			"%s stale permission should be removed\ngot:  %d rows, err=%v",
			prefix, count, err,
		)
	}
	if err := store.db.QueryRow(`
		SELECT COUNT(*) FROM group_permissions gp
		LEFT JOIN permissions p ON p.id = gp.permission_id
		WHERE p.id IS NULL;`,
	).Scan(&count); err != nil || count != 0 {
		t.Errorf(
			"%s orphaned grants should be removed\ngot:  %d rows, err=%v",
			prefix, count, err,
		)
	}
}

func TestNew__errors(t *testing.T) {
	// GIVEN: databases in states that break initialisation.
	tests := []struct {
		name  string
		setup func(t *testing.T, db *sql.DB)
	}{
		{
			name: "closed database",
			setup: func(_ *testing.T, db *sql.DB) {
				_ = db.Close()
			},
		},
		{
			name: "malformed schema_migrations rows",
			setup: func(t *testing.T, db *sql.DB) {
				for _, statement := range []string{
					`CREATE TABLE schema_migrations (version TEXT, applied_at DATETIME);`,
					`INSERT INTO schema_migrations (version, applied_at) VALUES (NULL, 'x');`,
				} {
					if _, err := db.Exec(statement); err != nil {
						t.Fatalf(
							"%s\nsetup failed: %v",
							packageName, err,
						)
					}
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := testDB(t)
			tc.setup(t, db)

			// WHEN: New is called.
			_, err := New(t.Context(), db)

			prefix := fmt.Sprintf("%s\nNew()", packageName)

			// THEN: it errors rather than continuing on a broken schema.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestMigrate__badMigration(t *testing.T) {
	// GIVEN: a pending migration with invalid SQL.
	migrationsHad := migrations
	migrations = append(
		append([]migration{}, migrations...),
		migration{version: 99, statements: []string{`THIS IS NOT SQL;`}},
	)
	t.Cleanup(func() { migrations = migrationsHad })

	db := testDB(t)

	// WHEN: New is called.
	_, err := New(t.Context(), db)

	prefix := fmt.Sprintf("%s\nNew() with a bad migration", packageName)

	// THEN: the bad migration surfaces as an error.
	if err == nil {
		t.Fatalf("%s expected an error, got nil", prefix)
	}
}

func TestSyncPermissions__rowsErr(t *testing.T) {
	// GIVEN: row iteration fails.
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	db := testDB(t)

	// WHEN: New is called (migrate + syncPermissions both iterate rows).
	_, err := New(t.Context(), db)

	prefix := fmt.Sprintf("%s\nNew() with failing row iteration", packageName)

	// THEN: the iteration failure is surfaced.
	if !errors.Is(err, wantErr) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, wantErr,
		)
	}
}

func TestSeedGroups__groupIDBySeedKeyError(t *testing.T) {
	// GIVEN: an initialised database whose groups table breaks lookups
	// (seed_key column removed).
	db := testDB(t)
	if _, err := New(t.Context(), db); err != nil {
		t.Fatalf(
			"%s\nfirst New() failed: %v",
			packageName, err,
		)
	}
	for _, statement := range []string{
		`DROP TABLE groups;`,
		`CREATE TABLE groups (id TEXT NOT NULL PRIMARY KEY);`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf(
				"%s\nsetup failed: %v", packageName, err,
			)
		}
	}

	// WHEN: New runs again (seedGroups looks groups up by seed_key).
	_, err := New(t.Context(), db)

	prefix := fmt.Sprintf("%s\nNew() with broken groups table", packageName)

	// THEN: the lookup failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error\ngot: nil", prefix)
	}
}

func TestFullCatalogueGrants(t *testing.T) {
	// GIVEN: the code catalogue.
	prefix := fmt.Sprintf("%s\nfullCatalogueGrants()", packageName)

	// WHEN: fullCatalogueGrants is called.
	grants := fullCatalogueGrants()

	// THEN: there is one valid global grant per catalogue pair.
	if len(grants) != cataloguePairCount() {
		t.Fatalf(
			"%s grant count mismatch\ngot:  %d\nwant: %d",
			prefix, len(grants), cataloguePairCount(),
		)
	}
	for _, grant := range grants {
		if !grant.Valid() || grant.Scope.Type != rbac.ScopeGlobal {
			t.Errorf(
				"%s invalid or non-global grant: %+v",
				prefix, grant,
			)
		}
	}
}

func TestInTx__commitError(t *testing.T) {
	// GIVEN: a Store, and a transaction function that itself commits.
	store := testStore(t)

	prefix := fmt.Sprintf("%s\ninTx() commit error", packageName)

	// WHEN: inTx tries to commit the already-committed transaction.
	err := store.inTx(t.Context(), func(tx *sql.Tx) error {
		_ = tx.Commit()
		return nil
	})

	// THEN: the commit failure is surfaced.
	if err == nil || !errors.Is(err, sql.ErrTxDone) {
		t.Errorf(
			"%s expected a commit error\ngot: %v",
			prefix, err,
		)
	}
}

func TestParseTime__invalid(t *testing.T) {
	// GIVEN: a corrupt stored timestamp.
	ts := "not-a-timestamp"

	// WHEN: parseTime parses it.
	_, err := parseTime(ts)

	prefix := fmt.Sprintf("%s\nparseTime()", packageName)

	// THEN: it errors.
	if err == nil {
		t.Errorf("%s expected an error\ngot: nil", prefix)
	}
}
