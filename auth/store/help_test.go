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
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

// testStore creates a Store on a fresh (in-memory) database.
func testStore(t *testing.T) *Store {
	t.Helper()

	store, err := New(t.Context(), test.MemoryDB(t))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	return store
}

// testPasswordHash is a dummy argon2id hash.
const testPasswordHash = "$argon2id$v=19$m=65536,t=2,p=1$c2FsdHNhbHRzYWx0$aGFzaGhhc2hoYXNoaGFzaA"

// mustCreateUser creates a user, failing the test on error.
// An empty passwordHash gets [testPasswordHash].
func mustCreateUser(
	t *testing.T,
	store *Store,
	username, passwordHash string,
	groups ...string,
) *auth.User {
	t.Helper()

	if passwordHash == "" {
		passwordHash = testPasswordHash
	}
	user, err := store.CreateUser(
		t.Context(),
		username, "Test "+username, username+"@example.com", passwordHash,
		groups,
	)
	if err != nil {
		t.Fatalf("Store.CreateUser(%q) failed: %v", username, err)
	}
	return user
}

// mustCreateGroup creates a group, failing the test on error.
func mustCreateGroup(t *testing.T, store *Store, name string, grants ...rbac.Grant) *Group {
	t.Helper()

	group, err := store.CreateGroup(t.Context(), name, "Test "+name, grants)
	if err != nil {
		t.Fatalf("Store.CreateGroup(%q) failed: %v", name, err)
	}
	return group
}

// mustUserByID fetches a user with the given id, failing the test on error.
func mustUserByID(t *testing.T, store *Store, id string) *auth.User {
	t.Helper()

	user, err := store.UserByID(t.Context(), id)
	if err != nil {
		t.Fatalf("Store.UserByID(%q) failed: %v", id, err)
	}
	return user
}

// mustLocalCredentials fetches the credentials of username, failing the test on error or none.
func mustLocalCredentials(t *testing.T, store *Store, username string) *auth.Credentials {
	t.Helper()

	creds, err := store.LocalCredentials(t.Context(), username)
	if err != nil {
		t.Fatalf("Store.LocalCredentials(%q) failed: %v", username, err)
	}
	if creds == nil {
		t.Fatalf("Store.LocalCredentials(%q) mismatch\ngot:  nil\nwant: non-nil", username)
	}
	return creds
}

// mustGroupByID fetches a group with the given id, failing the test on error.
func mustGroupByID(t *testing.T, store *Store, id string) *Group {
	t.Helper()

	group, err := store.GroupByID(t.Context(), id)
	if err != nil {
		t.Fatalf("Store.GroupByID(%q) failed: %v", id, err)
	}
	return group
}

// groupsByName indexes a store's groups by name.
func groupsByName(t *testing.T, store *Store) map[string]Group {
	t.Helper()

	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf("Store.Groups() failed: %v", err)
	}
	byName := make(map[string]Group, len(groups))
	for _, group := range groups {
		byName[group.Name] = group
	}
	return byName
}

// mustGroupByName returns store's group called name, failing when it is absent.
func mustGroupByName(t *testing.T, store *Store, name string) Group {
	t.Helper()

	return groupNamed(t, groupsByName(t, store), name)
}

// groupNamed returns byName[name], failing when it is absent.
func groupNamed(t *testing.T, byName map[string]Group, name string) Group {
	t.Helper()

	group, ok := byName[name]
	if !ok {
		t.Fatalf("byName[%q] mismatch\ngot:  absent\nwant: present", name)
	}
	return group
}

// globalGrant builds a valid global-scope Grant.
func globalGrant(resource rbac.Resource, action rbac.Action) rbac.Grant {
	return rbac.Grant{Resource: resource, Action: action, Scope: rbac.Scope{Type: rbac.ScopeGlobal}}
}

// serviceGrant builds a valid service-scope Grant.
func serviceGrant(resource rbac.Resource, action rbac.Action, serviceID string) rbac.Grant {
	return rbac.Grant{Resource: resource, Action: action, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: serviceID}}
}

// grantString renders a Grant as `resource:action@scope[/ref]`.
func grantString(grant rbac.Grant) string {
	str := fmt.Sprintf("%s:%s@%s", grant.Resource, grant.Action, grant.Scope.Type)
	if grant.Scope.Ref != "" {
		str += "/" + grant.Scope.Ref
	}
	return str
}

// grantStrings renders each of grants with [grantString].
func grantStrings(grants []rbac.Grant) []string {
	strs := make([]string, len(grants))
	for i, grant := range grants {
		strs[i] = grantString(grant)
	}
	return strs
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

	t.Fatalf("seededGroups key %q mismatch\ngot:  absent\nwant: present", key)
	return -1
}

// groupHasGlobalGrant reports whether group holds a global-scope grant on
// (resource, action).
func groupHasGlobalGrant(group Group, resource rbac.Resource, action rbac.Action) bool {
	return slices.Contains(group.Grants, globalGrant(resource, action))
}

// raceOnGroupInsert claims the group name being inserted from inside the
// INSERT, standing in for a concurrent create landing between the duplicate
// check and the insert.
func raceOnGroupInsert(t *testing.T, store *Store) {
	t.Helper()

	test.ExecSQL(t, store.db, `
		CREATE TRIGGER race_groups BEFORE INSERT ON groups BEGIN
			INSERT INTO groups (id, name, description, system, created_at, updated_at)
			VALUES ('racer', NEW.name, '', 0, '', '');
		END;`)
}

// removePermission deletes the (resource, action) pair (and grants on it),
// simulating a database initialised before the pair entered the catalogue.
func removePermission(t *testing.T, db *sql.DB, resource rbac.Resource, action rbac.Action) {
	t.Helper()

	for i, statement := range []string{
		`DELETE FROM group_permissions WHERE permission_id =
			(SELECT id FROM permissions WHERE resource = ? AND action = ?);`,
		`DELETE FROM permissions WHERE resource = ? AND action = ?;`,
	} {
		if _, err := db.Exec(statement, string(resource), string(action)); err != nil {
			t.Fatalf(
				"statement %d: DB.Exec(resource=%q, action=%q) failed: %v",
				i, resource, action, err,
			)
		}
	}
}
