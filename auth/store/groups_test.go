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

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/internal/test"
)

func TestStore_CreateGroup_and_GroupByID(t *testing.T) {
	// GIVEN: a Store.
	tests := []struct {
		name          string
		groupName     string
		grants        []rbac.Grant
		existingGroup string
		wantErr       error
	}{
		{
			name:      "group without grants",
			groupName: "empty-handed",
		},
		{
			name:      "group with grants across scopes",
			groupName: "scoped",
			grants: []rbac.Grant{
				globalGrant(rbac.ResourceService, rbac.ActionRead),
				serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "argus"),
			},
		},
		{
			name:          "duplicate name rejected",
			groupName:     "taken",
			existingGroup: "taken",
			wantErr:       ErrGroupNameTaken,
		},
		{
			name:          "duplicate name rejected case-insensitively",
			groupName:     "TAKEN",
			existingGroup: "taken",
			wantErr:       ErrGroupNameTaken,
		},
		{
			name:      "clash with seeded group rejected",
			groupName: GroupAdmin,
			wantErr:   ErrGroupNameTaken,
		},
		{
			name:      "invalid grant rejected",
			groupName: "bad-grants",
			grants: []rbac.Grant{
				{
					Permission: rbac.Permission{Resource: "nonsense", Action: "read"},
					Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
				},
			},
			wantErr: ErrInvalidGrant,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			if tc.existingGroup != "" {
				mustCreateGroup(t, store, tc.existingGroup)
			}
			groupDescription := "A test group."

			// WHEN: CreateGroup is called.
			group, err := store.CreateGroup(
				t.Context(),
				tc.groupName,
				groupDescription,
				tc.grants,
			)

			prefix := fmt.Sprintf(
				"%s\nCreateGroup(%q)",
				packageName, tc.groupName,
			)

			// THEN: errors match expectations.
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf(
						"%s error mismatch\ngot:  %v\nwant: %v",
						prefix, err, tc.wantErr,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// AND: the group is found in GroupByID.
			got, err := store.GroupByID(t.Context(), group.ID)
			if err != nil {
				t.Fatalf(
					"%s GroupByID failed: %v",
					prefix, err,
				)
			}
			if got.Name != tc.groupName ||
				got.Description != groupDescription ||
				got.System ||
				got.Members != 0 {
				t.Errorf(
					"%s group mismatch\ngot:  %+v",
					prefix, *got,
				)
			}
			if len(got.Grants) != len(tc.grants) {
				t.Errorf(
					"%s grant count mismatch\ngot:  %d\nwant: %d",
					prefix, len(got.Grants), len(tc.grants),
				)
			}
			for _, want := range tc.grants {
				if !slices.Contains(got.Grants, want) {
					t.Errorf(
						"%s missing grant: %+v\ngot: %+v",
						prefix, want, got.Grants,
					)
				}
			}
		})
	}
}

func TestStore_Groups(t *testing.T) {
	// GIVEN: a Store with a custom group and a member.
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

	prefix := fmt.Sprintf("%s\nGroups()", packageName)

	// WHEN: Groups is called.
	groups, err := store.Groups(t.Context())
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: seeded groups plus the custom group are returned, ordered by name.
	wantOrder := []string{GroupAdmin, "custom", GroupOperator, GroupViewer}
	if testErr := test.AssertSlicesEqualFunc(
		t,
		groups,
		wantOrder,
		func(g Group, w string) bool {
			return g.Name == w
		},
		prefix,
		"",
	); testErr != nil {
		t.Fatal(testErr)
	}

	// AND: member counts and grants are attached.
	for _, group := range groups {
		switch group.Name {
		case GroupAdmin:
		case GroupOperator:
		case "custom":
			if group.Members != 1 || len(group.Grants) != 1 {
				t.Errorf(
					"%s %s group mismatch\ngot:  members=%d grants=%d\nwant: members=1 grants=1",
					prefix, group.Name, group.Members, len(group.Grants),
				)
			}
		case GroupViewer:
			if group.Members != 1 {
				t.Errorf(
					"%s viewer member count mismatch\ngot:  %d\nwant: 1",
					prefix, group.Members,
				)
			}
		}
	}
}

func TestStore_Groups__errors(t *testing.T) {
	// GIVEN: stores in states that break listing.
	tests := []struct {
		name  string
		setup func(t *testing.T, store *Store)
	}{
		{
			name: "groups table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "groups")
			},
		},
		{
			name: "corrupt group timestamp",
			setup: func(t *testing.T, store *Store) {
				if _, err := store.db.Exec(`
					INSERT INTO groups (id, name, created_at, updated_at)
					VALUES ('x', 'x', 'not-a-time', 'not-a-time');`,
				); err != nil {
					t.Fatalf(
						"%s\nsetup failed: %v",
						packageName, err,
					)
				}
			},
		},
		{
			name: "group_permissions table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "group_permissions")
			},
		},
		{
			name: "NULL grant row",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "group_permissions")
				for _, statement := range []string{
					`CREATE TABLE group_permissions (group_id TEXT, permission_id INTEGER, scope_type TEXT, scope_ref TEXT);`,
					`INSERT INTO group_permissions (group_id, permission_id, scope_type, scope_ref)
						VALUES (NULL, (SELECT id FROM permissions LIMIT 1), NULL, NULL);`,
				} {
					if _, err := store.db.Exec(statement); err != nil {
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
			store := testStore(t)
			tc.setup(t, store)

			// WHEN: Groups is called.
			_, err := store.Groups(t.Context())

			prefix := fmt.Sprintf("%s\nGroups()", packageName)

			// THEN: the failure is surfaced.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestStore_Groups__rowsErr(t *testing.T) {
	// GIVEN: a Store whose row iteration fails.
	store := testStore(t)
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	// WHEN: Groups is called.
	_, err := store.Groups(t.Context())

	prefix := fmt.Sprintf("%s\nGroups() with failing iteration", packageName)

	// THEN: the iteration failure is surfaced.
	if !errors.Is(err, wantErr) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, wantErr,
		)
	}
}

func TestStore_GroupByID__notFound(t *testing.T) {
	// GIVEN: a Store.
	store := testStore(t)

	// WHEN: GroupByID is called with an unknown ID.
	_, err := store.GroupByID(t.Context(), "no-such-id")

	prefix := fmt.Sprintf("%s\nGroupByID(unknown)", packageName)

	// THEN: ErrNotFound is returned.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}
}

func TestStore_GroupByID__errors(t *testing.T) {
	// GIVEN: a Store with a group whose grants are unreadable.
	store := testStore(t)
	group := mustCreateGroup(t, store, "custom")
	dropTable(t, store, "group_permissions")

	// WHEN: GroupByID is called.
	_, err := store.GroupByID(t.Context(), group.ID)

	prefix := fmt.Sprintf("%s\nGroupByID() with unreadable grants", packageName)

	// THEN: the failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

func TestStore_UpdateGroup(t *testing.T) {
	customTarget := "custom"
	// GIVEN: a Store with groups in various states.
	tests := []struct {
		name    string
		target  string // [customTarget] (created) or a seeded group name.
		preset  string // Additional group created beforehand.
		patch   GroupPatch
		wantErr error
		check   func(t *testing.T, store *Store, prefix string, groupID string)
	}{
		{
			name:   "rename custom group",
			target: customTarget,
			patch:  GroupPatch{Name: test.Ptr("renamed")},
			check: func(t *testing.T, store *Store, prefix, groupID string) {
				group, _ := store.GroupByID(t.Context(), groupID)
				if group.Name != "renamed" {
					t.Errorf(
						"%s name not updated: %+v",
						prefix, *group,
					)
				}
			},
		},
		{
			name:    "rename to existing name rejected",
			target:  customTarget,
			preset:  "occupied",
			patch:   GroupPatch{Name: test.Ptr("occupied")},
			wantErr: ErrGroupNameTaken,
		},
		{
			name:   "rename to own name is a no-op",
			target: customTarget,
			patch:  GroupPatch{Name: test.Ptr(customTarget)},
		},
		{
			name:   "update description",
			target: customTarget,
			patch:  GroupPatch{Description: test.Ptr("updated")},
			check: func(t *testing.T, store *Store, prefix, groupID string) {
				group, _ := store.GroupByID(t.Context(), groupID)
				if group.Description != "updated" {
					t.Errorf(
						"%s description not updated: %+v",
						prefix, *group,
					)
				}
			},
		},
		{
			name:   "replace grants",
			target: customTarget,
			patch: GroupPatch{Grants: test.Ptr([]rbac.Grant{
				serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "argus"),
			})},
			check: func(t *testing.T, store *Store, prefix, groupID string) {
				group, _ := store.GroupByID(t.Context(), groupID)
				want := serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "argus")
				if len(group.Grants) != 1 || group.Grants[0] != want {
					t.Errorf(
						"%s grants not replaced: %+v",
						prefix, group.Grants,
					)
				}
			},
		},
		{
			name:   "invalid grant rejected",
			target: customTarget,
			patch: GroupPatch{Grants: test.Ptr([]rbac.Grant{
				{Permission: rbac.Permission{Resource: rbac.Resource("user"), Action: rbac.ActionRead},
					Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "argus"}},
			})},
			wantErr: ErrInvalidGrant,
		},
		{
			name:    "rename admin rejected",
			target:  GroupAdmin,
			patch:   GroupPatch{Name: test.Ptr("superusers")},
			wantErr: ErrSystemGroup,
		},
		{
			name:   "starter group renameable",
			target: GroupViewer,
			patch:  GroupPatch{Name: test.Ptr("watchers")},
			check: func(t *testing.T, store *Store, prefix, groupID string) {
				group, _ := store.GroupByID(t.Context(), groupID)
				if group.Name != "watchers" {
					t.Errorf(
						"%s name not updated: %+v",
						prefix, *group,
					)
				}
				// AND: the seed identity is unchanged by the rename.
				if group.SeedKey != GroupViewer {
					t.Errorf(
						"%s seed_key changed: %+v",
						prefix, *group,
					)
				}
			},
		},
		{
			name:   "starter group description editable",
			target: GroupViewer,
			patch:  GroupPatch{Description: test.Ptr("customised")},
			check: func(t *testing.T, store *Store, prefix, groupID string) {
				group, _ := store.GroupByID(t.Context(), groupID)
				if group.Description != "customised" {
					t.Errorf(
						"%s description not updated: %+v",
						prefix, *group,
					)
				}
			},
		},
		{
			name:   "starter group grants editable",
			target: GroupOperator,
			patch: GroupPatch{Grants: test.Ptr([]rbac.Grant{
				globalGrant(rbac.ResourceService, rbac.ActionRead),
			})},
			check: func(t *testing.T, store *Store, prefix, groupID string) {
				group, _ := store.GroupByID(t.Context(), groupID)
				if len(group.Grants) != 1 {
					t.Errorf(
						"%s grants not replaced: %+v",
						prefix, group.Grants,
					)
				}
			},
		},
		{
			name:   "admin grants immutable",
			target: GroupAdmin,
			patch: GroupPatch{Grants: test.Ptr([]rbac.Grant{
				globalGrant(rbac.ResourceService, rbac.ActionRead),
			})},
			wantErr: ErrSystemGroup,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			if tc.preset != "" {
				mustCreateGroup(t, store, tc.preset)
			}

			// Resolve the target group's ID.
			var groupID string
			if tc.target == customTarget {
				groupID = mustCreateGroup(t, store, customTarget).ID
			} else {
				groups, err := store.Groups(t.Context())
				if err != nil {
					t.Fatalf(
						"%s setup Groups failed: %v",
						packageName, err,
					)
				}
				for _, group := range groups {
					if group.Name == tc.target {
						groupID = group.ID
						break
					}
				}
			}

			// WHEN: UpdateGroup is called.
			_, err := store.UpdateGroup(t.Context(), groupID, tc.patch)

			prefix := fmt.Sprintf(
				"%s UpdateGroup(id=%q, patch=%+v)",
				packageName, groupID, tc.patch,
			)

			// THEN: errors match expectations.
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf(
						"%s error mismatch\ngot:  %v\nwant: %v",
						prefix, err, tc.wantErr,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// AND: the patch took effect.
			if tc.check != nil {
				tc.check(t, store, prefix, groupID)
			}
		})
	}
}

func TestStore_UpdateGroup__permissionLoadRowsErr(t *testing.T) {
	// GIVEN: a Store with a group whose row iteration fails.
	store := testStore(t)
	group := mustCreateGroup(t, store, "custom")
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	// WHEN: grants are set (permissionIDsByPair iterates the permissions).
	_, err := store.UpdateGroup(t.Context(), group.ID,
		GroupPatch{Grants: test.Ptr([]rbac.Grant{
			globalGrant(rbac.ResourceService, rbac.ActionRead),
		})})

	prefix := fmt.Sprintf("%s\nUpdateGroup() with failing permission iteration", packageName)

	// THEN: the iteration failure is surfaced.
	if !errors.Is(err, wantErr) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, wantErr,
		)
	}
}

func TestStore_UpdateGroup__notFound(t *testing.T) {
	// GIVEN: a Store.
	store := testStore(t)

	// WHEN: UpdateGroup is called with an unknown ID.
	_, err := store.UpdateGroup(
		t.Context(),
		"no-such-id",
		GroupPatch{},
	)

	prefix := fmt.Sprintf("%s\nUpdateGroup(unknown)", packageName)

	// THEN: ErrNotFound is returned.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}
}

func TestStore_DeleteGroup(t *testing.T) {
	customGroup := "custom"
	// GIVEN: a Store with a member-carrying 'custom' group and the seeded groups.
	tests := []struct {
		name    string
		target  string
		wantErr error
	}{
		{
			name:   "custom group",
			target: customGroup,
		},
		{
			name:   "starter group deletable",
			target: GroupOperator,
		},
		{
			name:    "admin rejected",
			target:  GroupAdmin,
			wantErr: ErrSystemGroup,
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
				groups, err := store.Groups(t.Context())
				if err != nil {
					t.Fatalf(
						"%s\nsetup Groups failed: %v",
						packageName, err,
					)
				}
				for _, group := range groups {
					if group.Name == tc.target {
						groupID = group.ID
						break
					}
				}
			}

			// WHEN: DeleteGroup is called.
			err := store.DeleteGroup(t.Context(), groupID)

			prefix := fmt.Sprintf(
				"%s\nDeleteGroup()",
				packageName,
			)

			// THEN: errors match expectations.
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf(
						"%s error mismatch\ngot:  %v\nwant: %v",
						prefix, err, tc.wantErr,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// AND: the group, its memberships and its grants are gone.
			if _, err := store.GroupByID(
				t.Context(),
				groupID,
			); !errors.Is(err, ErrNotFound) {
				t.Errorf(
					"%s group should be gone\ngot: %v",
					prefix, err,
				)
			}
			var count int
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM user_groups WHERE group_id = ?;`,
				groupID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"%s memberships should be cascade-deleted\ngot:  %d, err=%v",
					prefix, count, err,
				)
			}
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM group_permissions WHERE group_id = ?;`,
				groupID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"%s grants should be cascade-deleted\ngot:  %d, err=%v",
					prefix, count, err,
				)
			}
		})
	}
}

func TestStore_DeleteGroup__notFound(t *testing.T) {
	// GIVEN: a Store.
	store := testStore(t)

	// WHEN: DeleteGroup is called with an unknown ID.
	err := store.DeleteGroup(t.Context(), "no-such-id")

	prefix := fmt.Sprintf("%s\nDeleteGroup(unknown)", packageName)

	// THEN: ErrNotFound is returned.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}
}

func TestStore_GrantsForUser(t *testing.T) {
	// GIVEN: users spread across groups.
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
		t.Fatalf(
			"%s\nsetup disable failed: %v",
			packageName, err,
		)
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
			userID: disabled.ID,
		},
		{
			name:   "unknown user gets nothing",
			userID: "no-such-id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GrantsForUser is called.
			grants, err := store.GrantsForUser(t.Context(), tc.userID)

			prefix := fmt.Sprintf(
				"%s\nGrantsForUser()",
				packageName,
			)

			// THEN: no error occurs.
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// AND: the union of the user's groups' grants is returned.
			if len(grants) != tc.wantCount {
				t.Errorf(
					"%s grant count mismatch\ngot:  %d (%+v)\nwant: %d",
					prefix, len(grants), grants, tc.wantCount,
				)
			}
			for _, want := range tc.wantHas {
				if !slices.Contains(grants, want) {
					t.Errorf(
						"%s missing grant %+v\ngot: %+v",
						prefix, want, grants,
					)
				}
			}
		})
	}
}

func TestStore_GrantsForUser__errors(t *testing.T) {
	// GIVEN: a Store whose grants are unreadable.
	store := testStore(t)
	user := mustCreateUser(t, store, "member", "", GroupViewer)
	dropTable(t, store, "group_permissions")

	// WHEN: GrantsForUser is called.
	_, err := store.GrantsForUser(t.Context(), user.ID)

	prefix := fmt.Sprintf("%s\nGrantsForUser() with unreadable grants", packageName)

	// THEN: the failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

func TestStore_ServiceScopeRefs(t *testing.T) {
	// GIVEN: a Store with service-scoped and global grants.
	store := testStore(t)
	group := mustCreateGroup(t, store, "scoped",
		globalGrant(rbac.ResourceService, rbac.ActionRead),
		serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "old-name"),
		serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "other-service"),
	)

	prefix := fmt.Sprintf("%s\nRename/DeleteServiceScopeRefs()", packageName)

	// WHEN: a service is renamed.
	if err := store.RenameServiceScopeRefs(
		t.Context(),
		"old-name",
		"new-name",
	); err != nil {
		t.Fatalf(
			"%s rename failed: %v",
			prefix, err,
		)
	}

	// THEN: only the matching service-scoped grant follows the rename.
	got, err := store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf(
			"%s GroupByID failed: %v",
			prefix, err,
		)
	}
	if !slices.Contains(
		got.Grants,
		serviceGrant(rbac.ResourceService, rbac.ActionUpdate, "new-name"),
	) {
		t.Errorf(
			"%s grant should follow the rename\ngot: %+v",
			prefix, got.Grants,
		)
	}
	if !slices.Contains(
		got.Grants,
		serviceGrant(rbac.ResourceVersionRefresh, rbac.ActionExecute, "other-service"),
	) {
		t.Errorf(
			"%s unrelated grant should be untouched\ngot: %+v",
			prefix, got.Grants,
		)
	}

	// WHEN: the service is deleted.
	if err := store.DeleteServiceScopeRefs(t.Context(), "new-name"); err != nil {
		t.Fatalf(
			"%s delete failed: %v",
			prefix, err,
		)
	}

	// THEN: its scoped grant is pruned; global and unrelated grants remain.
	got, err = store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf(
			"%s GroupByID failed: %v",
			prefix, err,
		)
	}
	if len(got.Grants) != 2 {
		t.Errorf(
			"%s grant count mismatch after delete\ngot:  %+v\nwant: 2 grants",
			prefix, got.Grants,
		)
	}

	// AND: both maintenance calls error once the table is unreadable.
	dropTable(t, store, "group_permissions")
	if err := store.RenameServiceScopeRefs(t.Context(), "a", "b"); err == nil {
		t.Errorf("%s rename should error after drop", prefix)
	}
	if err := store.DeleteServiceScopeRefs(t.Context(), "a"); err == nil {
		t.Errorf("%s delete should error after drop", prefix)
	}
}

func TestSetGroupGrants__lookupError(t *testing.T) {
	// GIVEN: a valid grant whose permission row is missing
	// (permissions table emptied behind the catalogue's back).
	store := testStore(t)
	if _, err := store.db.Exec(`DELETE FROM permissions;`); err != nil {
		t.Fatalf(
			"%s\nsetup failed: %v",
			packageName, err,
		)
	}

	// WHEN: a group is created with a catalogue-valid grant.
	_, err := store.CreateGroup(
		t.Context(), "custom", "",
		[]rbac.Grant{
			globalGrant(rbac.ResourceService, rbac.ActionRead),
		},
	)

	prefix := fmt.Sprintf("%s\nCreateGroup() with missing permission rows", packageName)

	// THEN: the lookup failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

func TestPermissionIDsByPair(t *testing.T) {
	// GIVEN: an initialised store, its permissions synced to the catalogue.
	store := testStore(t)
	tx, err := store.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf(
			"%s\nbegin tx: %v",
			packageName, err,
		)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	prefix := fmt.Sprintf("%s\npermissionIDsByPair()", packageName)

	// WHEN: the permission catalogue is loaded into a map.
	ids, err := permissionIDsByPair(t.Context(), tx)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: every catalogue (resource, action) pair maps to an ID.
	for _, rp := range rbac.Catalogue() {
		for _, ap := range rp.Actions {
			perm := rbac.Permission{Resource: rp.Resource, Action: ap.Action}
			if _, ok := ids[perm]; !ok {
				t.Errorf(
					"%s missing id for %s:%s",
					prefix, perm.Resource, perm.Action,
				)
				continue
			}
		}
	}

	// AND: it holds exactly the catalogue's pairs, no more.
	if got, want := len(ids), cataloguePairCount(); got != want {
		t.Errorf(
			"%s map size mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestPermissionIDsByPair__scanError(t *testing.T) {
	// GIVEN: a permissions row whose id cannot scan into an int64.
	store := testStore(t)
	dropTable(t, store, "permissions")
	for _, statement := range []string{
		`CREATE TABLE permissions (id TEXT, resource TEXT, action TEXT);`,
		`INSERT INTO permissions (id, resource, action) VALUES ('not-an-int', 'service', 'read');`,
	} {
		if _, err := store.db.Exec(statement); err != nil {
			t.Fatalf(
				"%s\nsetup failed: %v",
				packageName, err,
			)
		}
	}

	tx, err := store.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf(
			"%s\nbegin tx: %v",
			packageName, err,
		)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	prefix := fmt.Sprintf("%s\npermissionIDsByPair() scan error", packageName)

	// WHEN: the permissions are loaded.
	_, err = permissionIDsByPair(t.Context(), tx)

	// THEN: the scan failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}
