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

package store

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

var packageName = "auth_store"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// testDB opens a fresh (in-memory) SQLite database.
func testDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf(
			"%s\nopen test database: %v",
			packageName, err,
		)
	}
	// In-memory SQLite gives each connection its own database;
	// pin the pool to one connection so all queries share it.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	return db
}

// testStore creates a Store on a fresh (in-memory) database.
func testStore(t *testing.T) *Store {
	t.Helper()

	store, err := New(t.Context(), testDB(t))
	if err != nil {
		t.Fatalf(
			"%s\nNew() failed: %v",
			packageName, err,
		)
	}
	return store
}

// mustCreateUser creates a user, failing the test on error.
func mustCreateUser(t *testing.T, store *Store, username, passwordHash string, groups ...string) *auth.User {
	t.Helper()

	user, err := store.CreateUser(
		t.Context(),
		username, "Test "+username, username+"@example.com", passwordHash,
		groups,
	)
	if err != nil {
		t.Fatalf(
			"%s\nCreateUser(%q) failed: %v",
			packageName, username, err,
		)
	}
	return user
}

// mustCreateGroup creates a group, failing the test on error.
func mustCreateGroup(t *testing.T, store *Store, name string, grants ...rbac.Grant) *Group {
	t.Helper()

	group, err := store.CreateGroup(t.Context(), name, "Test "+name, grants)
	if err != nil {
		t.Fatalf(
			"%s\nCreateGroup(%q) failed: %v",
			packageName, name, err,
		)
	}
	return group
}

// groupsByName indexes a store's groups by name.
func groupsByName(t *testing.T, store *Store) map[string]Group {
	t.Helper()

	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf(
			"%s\nGroups() failed: %v",
			packageName, err,
		)
	}
	byName := make(map[string]Group, len(groups))
	for _, group := range groups {
		byName[group.Name] = group
	}
	return byName
}

// globalGrant builds a valid global-scope Grant.
func globalGrant(resource rbac.Resource, action rbac.Action) rbac.Grant {
	return rbac.Grant{
		Permission: rbac.Permission{Resource: resource, Action: action},
		Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
	}
}

// serviceGrant builds a valid service-scope Grant.
func serviceGrant(resource rbac.Resource, action rbac.Action, serviceID string) rbac.Grant {
	return rbac.Grant{
		Permission: rbac.Permission{Resource: resource, Action: action},
		Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: serviceID},
	}
}

// cataloguePairCount counts the valid (resource, action) pairs.
func cataloguePairCount() int {
	count := 0
	for _, rp := range rbac.Catalogue() {
		count += len(rp.Actions)
	}
	return count
}

// seedGrantCount counts the catalogue pairs within seed group key's remit.
func seedGrantCount(t *testing.T, key string) int {
	t.Helper()

	for _, seed := range seededGroups {
		if seed.key != key {
			continue
		}
		count := 0
		for _, rp := range rbac.Catalogue() {
			for _, ap := range rp.Actions {
				if seed.matches(rbac.Permission{Resource: rp.Resource, Action: ap.Action}) {
					count++
				}
			}
		}
		return count
	}

	t.Fatalf(
		"%s\nunknown seed group %q",
		packageName, key,
	)
	return -1
}

// groupHasGlobalGrant reports whether group holds a global-scope grant on
// (resource, action).
func groupHasGlobalGrant(group Group, resource rbac.Resource, action rbac.Action) bool {
	return slices.Contains(group.Grants, globalGrant(resource, action))
}

// mustExec runs a statement, failing the test on error.
func mustExec(t *testing.T, store *Store, statement string) {
	t.Helper()

	if _, err := store.db.Exec(statement); err != nil {
		t.Fatalf(
			"%s\nexec failed: %v\n%s",
			packageName, err, statement,
		)
	}
}

// removePermission deletes the (resource, action) pair (and grants on it),
// simulating a database initialised before the pair entered the catalogue.
func removePermission(t *testing.T, db *sql.DB, resource rbac.Resource, action rbac.Action) {
	t.Helper()

	for _, statement := range []string{
		`DELETE FROM group_permissions WHERE permission_id =
			(SELECT id FROM permissions WHERE resource = ? AND action = ?);`,
		`DELETE FROM permissions WHERE resource = ? AND action = ?;`,
	} {
		if _, err := db.Exec(statement, string(resource), string(action)); err != nil {
			t.Fatalf(
				"%s\nremove permission %s:%s: %v",
				packageName, resource, action, err,
			)
		}
	}
}

// countRows counts the rows of a table.
func countRows(t *testing.T, store *Store, table string) int {
	t.Helper()

	var count int
	//#nosec G201 -- table names come from the tests themselves.
	if err := store.db.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, table),
	).Scan(&count); err != nil {
		t.Fatalf(
			"%s\ncount %s: %v",
			packageName, table, err,
		)
	}
	return count
}

// dropTable drops a table to force errors on statements touching it.
func dropTable(t *testing.T, store *Store, table string) {
	t.Helper()

	//#nosec G201 -- table names come from the tests themselves.
	if _, err := store.db.Exec(
		fmt.Sprintf(`DROP TABLE %s;`, table),
	); err != nil {
		t.Fatalf(
			"%s\ndrop %s: %v",
			packageName, table, err,
		)
	}
}
