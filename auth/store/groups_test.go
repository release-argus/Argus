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
	"slices"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestStore_CreateGroup_and_GroupByID(t *testing.T) {
	tests := []struct {
		name          string
		groupName     string
		grants        []rbac.Grant
		existingGroup string
		setup         func(t *testing.T, store *Store)
		wantErr       error
		errRegex      string
	}{
		{
			name:      "group without grants",
			groupName: "empty-handed",
			errRegex:  `^$`,
		},
		{
			name:      "group with grants across scopes",
			groupName: "scoped",
			grants: []rbac.Grant{
				globalGrant(rbac.ResourceService, rbac.ActionRead),
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "argus"),
			},
			errRegex: `^$`,
		},
		{
			name:          "duplicate name rejected",
			groupName:     "taken",
			existingGroup: "taken",
			wantErr:       ErrGroupNameTaken,
			errRegex:      `^group name already taken$`,
		},
		{
			name:          "duplicate name rejected case-insensitively",
			groupName:     "TAKEN", // groupIDByName matches COLLATE NOCASE.
			existingGroup: "taken",
			wantErr:       ErrGroupNameTaken,
			errRegex:      `^group name already taken$`,
		},
		{
			name:      "clash with seeded group rejected",
			groupName: GroupAdmin,
			wantErr:   ErrGroupNameTaken,
			errRegex:  `^group name already taken$`,
		},
		{
			name:      "name claimed between the check and the insert",
			groupName: "raced",
			setup:     raceOnGroupInsert,
			wantErr:   ErrGroupNameTaken, // Mapped from the UNIQUE violation.
			errRegex:  `^group name already taken$`,
		},
		{
			name:      "invalid grant rejected",
			groupName: "bad-grants",
			grants: []rbac.Grant{
				{Resource: "nonsense", Action: "read", Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
			},
			wantErr:  ErrInvalidGrant,
			errRegex: `^invalid grant: nonsense:read @ global$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			if tc.existingGroup != "" {
				mustCreateGroup(t, store, tc.existingGroup)
			}
			if tc.setup != nil {
				tc.setup(t, store)
			}
			groupDescription := "A test group."

			group, err := store.CreateGroup(
				t.Context(),
				tc.groupName,
				groupDescription,
				tc.grants,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error type mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got := mustGroupByID(t, store, group.ID)
			if testErr := test.AssertFields(t, []test.FieldAssertion{
				{Name: "Name", Got: got.Name, Want: tc.groupName, Mode: test.CompareEqual},
				{Name: "Description", Got: got.Description, Want: groupDescription, Mode: test.CompareEqual},
				{Name: "System", Got: got.System, Want: false, Mode: test.CompareEqual},
				{Name: "Members", Got: got.Members, Want: 0, Mode: test.CompareEqual},
			}, "Group"); testErr != nil {
				t.Error(testErr)
			}
			if gotLen, wantLen := len(got.Grants), len(tc.grants); gotLen != wantLen {
				t.Errorf("len(Group.Grants) mismatch\ngot:  %d\nwant: %d", gotLen, wantLen)
			}
			for _, want := range tc.grants {
				if !slices.Contains(got.Grants, want) {
					t.Errorf(
						"Group.Grants mismatch\ngot:  %q\nwant: %q",
						grantStrings(got.Grants), grantString(want),
					)
				}
			}
		})
	}
}

func TestStore_Groups(t *testing.T) {
	store := testStore(t)
	mustCreateGroup(
		t,
		store,
		"custom",
		globalGrant(rbac.ResourceService, rbac.ActionRead),
	)
	mustCreateUser(
		t,
		store,
		"member",
		"",
		"custom",
		GroupViewer,
	)

	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantOrder := []string{GroupAdmin, "custom", GroupOperator, GroupViewer}
	gotNames := make([]string, len(groups))
	for i, group := range groups {
		gotNames[i] = group.Name
	}
	if testErr := test.AssertSlicesEqual(t, gotNames, wantOrder, "Group.Name"); testErr != nil {
		t.Fatal(testErr)
	}

	for _, group := range groups {
		switch group.Name {
		case "custom":
			if got, want := group.Members, 1; got != want {
				t.Errorf("Group[%q].Members mismatch\ngot:  %d\nwant: %d", group.Name, got, want)
			}
			if got, want := len(group.Grants), 1; got != want {
				t.Errorf(
					"len(Group[%q].Grants) mismatch\ngot:  %d\nwant: %d",
					group.Name, got, want,
				)
			}
		case GroupViewer:
			if got, want := group.Members, 1; got != want {
				t.Errorf("Group[%q].Members mismatch\ngot:  %d\nwant: %d", group.Name, got, want)
			}
		}
	}
}

func TestStore_Groups__errors(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, store *Store)
		errRegex string
	}{
		{
			name: "groups table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "groups")
			},
			errRegex: test.TrimYAML(`
				^list groups:
					SQL logic error: no such table: groups`),
		},
		{
			name: "corrupt group timestamp",
			setup: func(t *testing.T, store *Store) {
				if _, err := store.db.Exec(`
					INSERT INTO groups (id, name, created_at, updated_at)
					VALUES ('x', 'x', 'not-a-time', 'not-a-time');`,
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			},
			errRegex: test.TrimYAML(`
				^scan group:
					sql: Scan error on column index \d+, name "created_at":
						parse stored timestamp "not-a-time"`,
			),
		},
		{
			name: "group_permissions table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "group_permissions")
			},
			errRegex: test.TrimYAML(`
				^attach grants:
					SQL logic error: no such table: group_permissions`),
		},
		{
			name: "NULL grant row",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "group_permissions")
				test.ExecSQL(
					t, store.db,
					`CREATE TABLE group_permissions (group_id TEXT, permission_id INTEGER, scope_type TEXT, scope_ref TEXT);`,
					`INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
						VALUES (NULL, (SELECT id FROM permissions LIMIT 1), NULL, NULL);`,
				)
			},
			errRegex: test.TrimYAML(`
				^attach grants:
					scan grant:
						sql: Scan error on column index 0, name "group_id":
							converting NULL to string is unsupported$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			tc.setup(t, store)

			_, err := store.Groups(t.Context())
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestStore_Groups__rowsErr(t *testing.T) {
	store := testStore(t)
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	_, err := store.Groups(t.Context())

	if !errors.Is(err, wantErr) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, wantErr)
	}
}

func TestStore_GroupByID__notFound(t *testing.T) {
	store := testStore(t)

	_, err := store.GroupByID(t.Context(), "no-such-id")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}
}

func TestStore_GroupByID__errors(t *testing.T) {
	store := testStore(t)
	group := mustCreateGroup(t, store, "custom")
	test.DropTable(t, store.db, "group_permissions")

	_, err := store.GroupByID(t.Context(), group.ID)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^list group's grants:
			SQL logic error: no such table: group_permissions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_UpdateGroup(t *testing.T) {
	customTarget := "custom"

	tests := []struct {
		name     string
		target   string // customTarget (created) or a seeded group name.
		preset   string // Additional group created beforehand.
		patch    GroupPatch
		wantErr  error
		errRegex string
		check    func(t *testing.T, store *Store, groupID string)
	}{
		{
			name:     "rename custom group",
			target:   customTarget,
			patch:    GroupPatch{Name: new("renamed")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				if group.Name != "renamed" {
					t.Errorf("Group.Name mismatch\ngot:  %q\nwant: %q", group.Name, "renamed")
				}
			},
		},
		{
			name:     "rename to existing name rejected",
			target:   customTarget,
			preset:   "occupied",
			patch:    GroupPatch{Name: new("occupied")},
			wantErr:  ErrGroupNameTaken,
			errRegex: `^group name already taken$`,
		},
		{
			name:     "rename to own name is a no-op",
			target:   customTarget,
			patch:    GroupPatch{Name: new(customTarget)},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				if group.Name != customTarget {
					t.Errorf("Group.Name mismatch\ngot:  %q\nwant: %q", group.Name, customTarget)
				}
			},
		},
		{
			name:     "rename to own name in a different case",
			target:   customTarget,
			patch:    GroupPatch{Name: new("CUSTOM")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				if group.Name != "CUSTOM" {
					t.Errorf("Group.Name mismatch\ngot:  %q\nwant: %q", group.Name, "CUSTOM")
				}
			},
		},
		{
			name:     "update description",
			target:   customTarget,
			patch:    GroupPatch{Description: new("updated")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				if group.Description != "updated" {
					t.Errorf(
						"Group.Description mismatch\ngot:  %q\nwant: %q",
						group.Description, "updated",
					)
				}
			},
		},
		{
			name:   "replace grants",
			target: customTarget,
			patch: GroupPatch{Grants: new([]rbac.Grant{
				serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "argus"),
			})},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				want := serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "argus")
				if len(group.Grants) != 1 || group.Grants[0] != want {
					t.Errorf(
						"Group.Grants mismatch\ngot:  %q\nwant: %q",
						grantStrings(group.Grants), []string{grantString(want)},
					)
				}
			},
		},
		{
			name:   "invalid grant rejected",
			target: customTarget,
			patch: GroupPatch{Grants: new([]rbac.Grant{
				// Not a catalogued resource.
				{Resource: rbac.Resource("user"), Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "argus"}},
			})},
			wantErr:  ErrInvalidGrant,
			errRegex: `^invalid grant: user:read @ service/argus$`,
		},
		{
			name:     "rename admin rejected",
			target:   GroupAdmin,
			patch:    GroupPatch{Name: new("superusers")},
			wantErr:  ErrSystemGroup,
			errRegex: `^system group cannot be modified this way: cannot rename a system group$`,
		},
		{
			name:     "starter group renameable",
			target:   GroupViewer,
			patch:    GroupPatch{Name: new("watchers")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				if group.Name != "watchers" {
					t.Errorf("Group.Name mismatch\ngot:  %q\nwant: %q", group.Name, "watchers")
				}

				if group.SeedKey != GroupViewer {
					t.Errorf(
						"Group.SeedKey mismatch\ngot:  %q\nwant: %q",
						group.SeedKey, GroupViewer,
					)
				}
			},
		},
		{
			name:     "starter group description editable",
			target:   GroupViewer,
			patch:    GroupPatch{Description: new("customised")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				if group.Description != "customised" {
					t.Errorf(
						"Group.Description mismatch\ngot:  %q\nwant: %q",
						group.Description, "customised",
					)
				}
			},
		},
		{
			name:   "starter group grants editable",
			target: GroupOperator,
			patch: GroupPatch{Grants: new([]rbac.Grant{
				globalGrant(rbac.ResourceService, rbac.ActionRead),
			})},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, groupID string) {
				group := mustGroupByID(t, store, groupID)
				want := globalGrant(rbac.ResourceService, rbac.ActionRead)
				if len(group.Grants) != 1 || group.Grants[0] != want {
					t.Errorf(
						"Group.Grants mismatch\ngot:  %q\nwant: %q",
						grantStrings(group.Grants), []string{grantString(want)},
					)
				}
			},
		},
		{
			name:   "admin grants immutable",
			target: GroupAdmin,
			patch: GroupPatch{Grants: new([]rbac.Grant{
				globalGrant(rbac.ResourceService, rbac.ActionRead),
			})},
			wantErr:  ErrSystemGroup,
			errRegex: `^system group cannot be modified this way: admin grants are fixed$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			if tc.preset != "" {
				mustCreateGroup(t, store, tc.preset)
			}

			var groupID string
			if tc.target == customTarget {
				groupID = mustCreateGroup(t, store, customTarget).ID
			} else {
				groupID = mustGroupByName(t, store, tc.target).ID
			}

			_, err := store.UpdateGroup(t.Context(), groupID, tc.patch)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error type mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if tc.check != nil {
				tc.check(t, store, groupID)
			}
		})
	}
}

func TestStore_UpdateGroup__permissionLoadRowsErr(t *testing.T) {
	store := testStore(t)
	group := mustCreateGroup(t, store, "custom")
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	// permissionIDsByPair iterates the permissions.
	_, err := store.UpdateGroup(t.Context(), group.ID,
		GroupPatch{Grants: new([]rbac.Grant{
			globalGrant(rbac.ResourceService, rbac.ActionRead),
		})},
	)

	if !errors.Is(err, wantErr) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, wantErr)
	}
}

func TestStore_UpdateGroup__notFound(t *testing.T) {
	store := testStore(t)

	_, err := store.UpdateGroup(
		t.Context(),
		"no-such-id",
		GroupPatch{},
	)

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}
}

func TestStore_DeleteGroup(t *testing.T) {
	customGroup := "custom"

	tests := []struct {
		name     string
		target   string
		wantErr  error
		errRegex string
	}{
		{
			name:     "custom group",
			target:   customGroup,
			errRegex: `^$`,
		},
		{
			name:     "starter group deletable",
			target:   GroupOperator,
			errRegex: `^$`,
		},
		{
			name:     "admin rejected",
			target:   GroupAdmin,
			wantErr:  ErrSystemGroup,
			errRegex: `^system group cannot be modified this way: cannot delete a system group$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			var groupID string
			if tc.target == customGroup {
				groupID = mustCreateGroup(
					t,
					store,
					customGroup,
					globalGrant(rbac.ResourceService, rbac.ActionRead),
				).ID
				mustCreateUser(
					t,
					store,
					"member",
					"",
					customGroup,
				)
			} else {
				groupID = mustGroupByName(t, store, tc.target).ID
			}

			err := store.DeleteGroup(t.Context(), groupID)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error type mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if _, err := store.GroupByID(
				t.Context(),
				groupID,
			); !errors.Is(err, ErrNotFound) {
				t.Errorf(
					"Store.GroupByID() error type mismatch\ngot:  %v\nwant: %v",
					err, ErrNotFound,
				)
			}
			var count int
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM user_groups WHERE group_id = ?;`,
				groupID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"user_groups row count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil (deleted)",
					count, err,
				)
			}
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM group_permissions WHERE group_id = ?;`,
				groupID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"group_permissions row count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil (deleted)",
					count, err,
				)
			}
		})
	}
}

func TestStore_DeleteGroup__notFound(t *testing.T) {
	store := testStore(t)

	err := store.DeleteGroup(t.Context(), "no-such-id")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}
}

func TestStore_UserIDsInGroup(t *testing.T) {
	store := testStore(t)
	crew := mustCreateGroup(t, store, "crew")
	empty := mustCreateGroup(t, store, "empty")
	mustCreateGroup(t, store, "other")
	userA := mustCreateUser(t, store, "user-a", "", "crew")
	userB := mustCreateUser(t, store, "user-b", "", "crew", "other")
	mustCreateUser(t, store, "user-c", "", "other")

	got, err := store.UserIDsInGroup(t.Context(), crew.ID)
	if err != nil {
		t.Fatalf("before DropTable: unexpected error: %v", err)
	}

	want := []string{userA.ID, userB.ID}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("before DropTable: user IDs mismatch\ngot:  %q\nwant: %q", got, want)
	}

	noMembersTests := []struct {
		name    string
		groupID string
	}{
		{name: "memberless", groupID: empty.ID},
		{name: "unknown", groupID: "no-such-id"},
	}
	for _, tc := range noMembersTests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares the user_groups table, which the parent drops after the loop.

			if ids, err := store.UserIDsInGroup(t.Context(), tc.groupID); err != nil || ids != nil {
				t.Errorf("user IDs mismatch\ngot:  %q, err=%v\nwant: nil, err=nil", ids, err)
			}
		})
	}

	test.DropTable(t, store.db, "user_groups")
	_, err = store.UserIDsInGroup(t.Context(), crew.ID)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^list group's users:
			SQL logic error: no such table: user_groups`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("missing table: error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_GrantsForUser(t *testing.T) {
	store := testStore(t)
	mustCreateGroup(
		t,
		store,
		"scoped",
		serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "argus"),
	)
	multi := mustCreateUser(t, store, "multi", "", GroupViewer, "scoped")
	single := mustCreateUser(t, store, "single", "", GroupViewer)
	loner := mustCreateUser(t, store, "loner", "")
	disabled := mustCreateUser(t, store, "disabled", "", GroupViewer)
	if _, err := store.db.Exec(
		`UPDATE users SET enabled = 0 WHERE id = ?;`, disabled.ID,
	); err != nil {
		t.Fatalf("DB.Exec() failed: %v", err)
	}

	viewerGrantCount := seedGrantCount(t, GroupViewer)

	tests := []struct {
		name      string
		userID    string
		wantCount int
		wantHas   []rbac.Grant
	}{
		{
			name:      "multiple groups merge (union)",
			userID:    multi.ID,
			wantCount: viewerGrantCount + 1,
			wantHas: []rbac.Grant{
				globalGrant(rbac.ResourceService, rbac.ActionRead),
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "argus"),
			},
		},
		{
			name:      "single group",
			userID:    single.ID,
			wantCount: viewerGrantCount,
		},
		{
			name:   "no groups",
			userID: loner.ID,
		},
		{
			name:   "disabled user gets nothing",
			userID: disabled.ID, // The query filters out disabled users.
		},
		{
			name:   "unknown user gets nothing",
			userID: "no-such-id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			grants, err := store.GrantsForUser(t.Context(), tc.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(grants) != tc.wantCount {
				t.Errorf(
					"len(grants) mismatch\ngot:  %d - %q\nwant: %d",
					len(grants), grantStrings(grants), tc.wantCount,
				)
			}
			for _, want := range tc.wantHas {
				if !slices.Contains(grants, want) {
					t.Errorf(
						"grants mismatch\ngot:  %q\nwant: %q",
						grantStrings(grants), grantString(want),
					)
				}
			}
		})
	}
}

func TestStore_GrantsForUser__errors(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "member", "", GroupViewer)
	test.DropTable(t, store.db, "group_permissions")

	_, err := store.GrantsForUser(t.Context(), user.ID)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^list user's grants:
			SQL logic error: no such table: group_permissions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_ServiceScopeRefs(t *testing.T) {
	store := testStore(t)
	group := mustCreateGroup(t, store, "scoped",
		globalGrant(rbac.ResourceService, rbac.ActionRead),
		serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"),
		serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "other-service"),
	)

	oldID, newID := "old-name", "new-name"
	_, err := store.RenameServiceScopeRefs(t.Context(), oldID, newID)
	if err != nil {
		t.Fatalf("rename: Store.RenameServiceScopeRefs() failed: %v", err)
	}

	got, err := store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf("rename: Store.GroupByID() failed: %v", err)
	}
	renamed := serviceGrant(rbac.ResourceService, rbac.ActionUpdate, newID)
	if !slices.Contains(got.Grants, renamed) {
		t.Errorf(
			"rename: Group.Grants mismatch\ngot:  %q\nwant: %q",
			grantStrings(got.Grants), grantString(renamed),
		)
	}
	unrelated := serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "other-service")
	if !slices.Contains(got.Grants, unrelated) {
		t.Errorf(
			"rename: Group.Grants mismatch\ngot:  %q\nwant: %q",
			grantStrings(got.Grants), grantString(unrelated),
		)
	}

	err = store.DeleteServiceScopeRefs(t.Context(), newID)
	if err != nil {
		t.Fatalf("delete: Store.DeleteServiceScopeRefs() failed: %v", err)
	}

	got, err = store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf("delete: Store.GroupByID() failed: %v", err)
	}
	wantRemaining := []rbac.Grant{globalGrant(rbac.ResourceService, rbac.ActionRead), unrelated}
	if len(got.Grants) != len(wantRemaining) {
		t.Errorf(
			"delete: len(Group.Grants) mismatch\ngot:  %d - %q\nwant: %d - %q",
			len(got.Grants), grantStrings(got.Grants),
			len(wantRemaining), grantStrings(wantRemaining),
		)
	}
	var missing []rbac.Grant
	for _, want := range wantRemaining {
		if !slices.Contains(got.Grants, want) {
			missing = append(missing, want)
		}
	}
	if len(missing) != 0 {
		t.Errorf(
			"delete: Group.Grants mismatch\ngot:  %q\nwant: %q",
			grantStrings(got.Grants), grantStrings(missing),
		)
	}

	test.DropTable(t, store.db, "group_permissions")
	_, err = store.RenameServiceScopeRefs(t.Context(), "a", "b")
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^list service-scoped grants:
			SQL logic error: no such table: group_permissions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf(
			"missing table: Store.RenameServiceScopeRefs() error mismatch\ngot:  %q\nwant: %q",
			e, errRegex,
		)
	}
	e = errfmt.FormatError(store.DeleteServiceScopeRefs(t.Context(), "a"))
	errRegex = test.TrimYAML(`
		^delete service scope refs:
			SQL logic error: no such table: group_permissions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf(
			"missing table: Store.DeleteServiceScopeRefs() error mismatch\ngot:  %q\nwant: %q",
			e, errRegex,
		)
	}
}

func TestStore_RenameServiceScopeRefs__collision(t *testing.T) {
	store := testStore(t)
	oldID, newID := "old-name", "new-name"
	group := mustCreateGroup(
		t,
		store,
		"collide",
		serviceGrant(rbac.ResourceService, rbac.ActionUpdate, oldID),
		serviceGrant(rbac.ResourceService, rbac.ActionUpdate, newID),
	)

	// A primary-key collision the UPDATE OR REPLACE must absorb.
	_, err := store.RenameServiceScopeRefs(t.Context(), oldID, newID)
	if err != nil {
		t.Fatalf("Store.RenameServiceScopeRefs() failed: %v", err)
	}

	got := mustGroupByID(t, store, group.ID)
	want := serviceGrant(rbac.ResourceService, rbac.ActionUpdate, newID)
	if !slices.Contains(got.Grants, want) {
		t.Errorf(
			"Group.Grants mismatch\ngot:  %q\nwant: %q",
			grantStrings(got.Grants), grantString(want),
		)
	}
	if len(got.Grants) != 1 {
		t.Errorf(
			"len(Group.Grants) mismatch\ngot:  %d - %q\nwant: 1",
			len(got.Grants), grantStrings(got.Grants),
		)
	}
}

func TestStore_UndoServiceScopeMove(t *testing.T) {
	tests := []struct {
		name   string
		grants []rbac.Grant
	}{
		{
			name: "moved only",
			grants: []rbac.Grant{
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"),
			},
		},
		{
			name: "held at the target only",
			grants: []rbac.Grant{
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "new-name"),
			},
		},
		{
			name: "both, so the rename absorbs one",
			grants: []rbac.Grant{
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"),
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "new-name"),
			},
		},
		{
			name: "an unrelated scope is never touched",
			grants: []rbac.Grant{
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"),
				serviceGrant(rbac.ResourceService, rbac.ActionRead, "elsewhere"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			group := mustCreateGroup(t, store, "scoped", tc.grants...)

			before, err := store.GroupByID(t.Context(), group.ID)
			if err != nil {
				t.Fatalf("before: Store.GroupByID() failed: %v", err)
			}

			oldID, newID := "old-name", "new-name"
			move, err := store.RenameServiceScopeRefs(t.Context(), oldID, newID)
			if err != nil {
				t.Fatalf("rename: Store.RenameServiceScopeRefs() failed: %v", err)
			}
			err = store.UndoServiceScopeMove(t.Context(), move)
			if err != nil {
				t.Fatalf("undo: Store.UndoServiceScopeMove() failed: %v", err)
			}

			if got, want := move.From(), oldID; got != want {
				t.Errorf("ServiceScopeMove.From() mismatch\ngot:  %q\nwant: %q", got, want)
			}
			if got, want := move.To(), newID; got != want {
				t.Errorf("ServiceScopeMove.To() mismatch\ngot:  %q\nwant: %q", got, want)
			}

			// The undo re-creates any absorbed row.
			got, err := store.GroupByID(t.Context(), group.ID)
			if err != nil {
				t.Fatalf("after UndoServiceScopeMove: Store.GroupByID() failed: %v", err)
			}
			if gotLen, wantLen := len(got.Grants), len(before.Grants); gotLen != wantLen {
				t.Fatalf(
					"len(Group.Grants) mismatch\ngot:  %d - %q\nwant: %d - %q",
					gotLen, grantStrings(got.Grants),
					wantLen, grantStrings(before.Grants),
				)
			}
			for _, want := range before.Grants {
				if !slices.Contains(got.Grants, want) {
					t.Errorf(
						"Group.Grants mismatch\ngot:  %q\nwant: %q",
						grantStrings(got.Grants), grantString(want),
					)
				}
			}
		})
	}
}

func TestStore_UndoServiceScopeMove__leavesUnmovedGrantsAlone(t *testing.T) {
	store := testStore(t)
	// Scoped to an ID no rename touches, which happens to be the rename's target.
	group := mustCreateGroup(t, store,
		"bystander",
		serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "target"),
	)

	// The rename matches no rows, so the undo must move nothing.
	oldID, newID := "absent", "target"
	move, err := store.RenameServiceScopeRefs(t.Context(), oldID, newID)
	if err != nil {
		t.Fatalf("Store.RenameServiceScopeRefs() failed: %v", err)
	}
	err = store.UndoServiceScopeMove(t.Context(), move)
	if err != nil {
		t.Fatalf("Store.UndoServiceScopeMove() failed: %v", err)
	}

	got := mustGroupByID(t, store, group.ID)
	want := serviceGrant(rbac.ResourceService, rbac.ActionUpdate, newID)
	if !slices.Contains(got.Grants, want) {
		t.Errorf(
			"Group.Grants mismatch\ngot:  %q\nwant: %q",
			grantStrings(got.Grants), grantString(want),
		)
	}
	neverHeld := serviceGrant(rbac.ResourceService, rbac.ActionUpdate, oldID)
	if slices.Contains(got.Grants, neverHeld) {
		t.Errorf(
			"Group.Grants mismatch\ngot:  %q\nwant: NOT %q",
			grantStrings(got.Grants), grantString(neverHeld),
		)
	}
}

func TestSetGroupGrants__lookupError(t *testing.T) {
	store := testStore(t)
	// Emptied behind the catalogue's back, so the grant still validates.
	if _, err := store.db.Exec(`DELETE FROM permissions;`); err != nil {
		t.Fatalf("DB.Exec() failed: %v", err)
	}

	_, err := store.CreateGroup(
		t.Context(),
		"custom",
		"",
		[]rbac.Grant{
			globalGrant(rbac.ResourceService, rbac.ActionRead),
		},
	)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^look up permission service:read:
			not found$`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestPermissionIDsByPair(t *testing.T) {
	store := testStore(t)
	tx, err := store.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("DB.BeginTx() failed: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	ids, err := permissionIDsByPair(t.Context(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, rp := range rbac.Catalogue() {
		for _, ap := range rp.Actions {
			perm := rbac.Permission{Resource: rp.Resource, Action: ap.Action}
			if _, ok := ids[perm]; !ok {
				t.Errorf(
					"ids[%s:%s] mismatch\ngot:  absent\nwant: present",
					perm.Resource, perm.Action,
				)
			}
		}
	}

	if got, want := len(ids), cataloguePairCount(); got != want {
		t.Errorf("len(ids) mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestPermissionIDsByPair__scanError(t *testing.T) {
	store := testStore(t)
	test.DropTable(t, store.db, "permissions")
	test.ExecSQL(
		t, store.db,
		`CREATE TABLE permissions (id TEXT, resource TEXT, action TEXT);`,
		`INSERT INTO permissions (id, resource, action) VALUES ('not-an-int', 'service', 'read');`,
	)

	tx, err := store.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("DB.BeginTx() failed: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	_, err = permissionIDsByPair(t.Context(), tx)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^scan permission:
			[^\s]+.*name "id":
				converting driver.Value type string \("not-an-int"\)`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}
