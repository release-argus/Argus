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

	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestStore_CountUsers(t *testing.T) {
	store := testStore(t)
	mustCreateUser(t, store, "alpha", "")

	count, err := store.CountUsers(t.Context())
	if err != nil || count != 1 {
		t.Errorf("before DropTable: count mismatch\ngot:  %d, err=%v\nwant: 1, err=nil", count, err)
	}

	test.DropTable(t, store.db, "users")
	_, err = store.CountUsers(t.Context())
	e := errfmt.FormatError(err)
	errRegex := `^count users:\n  SQL logic error: no such table: users`
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("missing table: error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_Users(t *testing.T) {
	store := testStore(t)
	mustCreateUser(t, store, "beta", "", GroupViewer)
	mustCreateUser(t, store, "alpha", "")

	users, err := store.Users(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotNames := make([]string, len(users))
	for i, user := range users {
		gotNames[i] = user.Username
	}
	if want := []string{"alpha", "beta"}; !slices.Equal(gotNames, want) {
		t.Fatalf("usernames mismatch\ngot:  %q\nwant: %q", gotNames, want)
	}

	if len(users[0].Groups) != 0 || !slices.Equal(users[1].Groups, []string{GroupViewer}) {
		t.Errorf(
			"User.Groups mismatch\ngot:  alpha=%q, beta=%q\nwant: alpha=[], beta=[%q]",
			users[0].Groups, users[1].Groups, GroupViewer,
		)
	}
}

func TestStore_Users__errors(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, store *Store)
		errRegex string
	}{
		{
			name: "users table missing",
			errRegex: test.TrimYAML(`
				^list users:
				  SQL logic error: no such table: users`),
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "users")
			},
		},
		{
			name: "corrupt user timestamp",
			errRegex: test.TrimYAML(`
				^scan user:
				  sql: Scan error on column index 5, name "created_at":
						parse stored timestamp "not-a-time"`),
			setup: func(t *testing.T, store *Store) {
				if _, err := store.db.Exec(`
					INSERT INTO users (id, username, password_hash, created_at, updated_at)
					VALUES ('x', 'x', 'x', 'not-a-time', 'not-a-time');`,
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			},
		},
		{
			name: "user_groups table missing",
			errRegex: test.TrimYAML(`
				^list user groups:
				  SQL logic error: no such table: user_groups`),
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "alpha", "")
				test.DropTable(t, store.db, "user_groups")
			},
		},
		{
			name: "NULL membership row",
			errRegex: test.TrimYAML(`
				^scan user group:
				  sql: Scan error on column index 0, name "user_id"`),
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "alpha", "")
				test.DropTable(t, store.db, "user_groups")
				test.ExecSQL(
					t, store.db,
					`CREATE TABLE user_groups (user_id TEXT, group_id TEXT);`,
					`INSERT INTO user_groups (user_id, group_id)
						VALUES (NULL, (SELECT id FROM groups WHERE name = 'viewer'));`,
				)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			tc.setup(t, store)

			_, err := store.Users(t.Context())
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestStore_Users__rowsErr(t *testing.T) {
	store := testStore(t)
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	_, err := store.Users(t.Context())

	if !errors.Is(err, wantErr) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, wantErr)
	}
}

func TestStore_UserByID__notFound(t *testing.T) {
	store := testStore(t)

	_, err := store.UserByID(t.Context(), "no-such-id")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}
}

func TestStore_UserWithGrants__matchesTheSeparateQueries(t *testing.T) {
	store := testStore(t)

	readGlobal := rbac.Grant{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeGlobal}}
	updateScoped := rbac.Grant{Resource: rbac.ResourceService, Action: rbac.ActionUpdate, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "svc-a"}}
	mustCreateGroup(t, store, "readers", readGlobal)
	mustCreateGroup(t, store, "editors", updateScoped)
	mustCreateGroup(t, store, "empty-group")

	mustCreateUser(t, store, "no-groups", "")
	mustCreateUser(t, store, "one-group", "", "readers")
	mustCreateUser(t, store, "two-groups", "", "readers", "editors")
	mustCreateUser(t, store, "grantless-group", "", "empty-group")

	tests := []struct {
		name       string
		username   string
		wantGroups []string
		wantGrants int
	}{
		{
			name:       "valid/user in no groups",
			username:   "no-groups",
			wantGroups: nil,
			wantGrants: 0,
		},
		{
			name:       "valid/user in one group",
			username:   "one-group",
			wantGroups: []string{"readers"},
			wantGrants: 1,
		},
		{
			name:       "valid/user in two groups",
			username:   "two-groups",
			wantGroups: []string{"editors", "readers"},
			wantGrants: 2,
		},
		{
			name:       "valid/group carrying no grants",
			username:   "grantless-group",
			wantGroups: []string{"empty-group"},
			wantGrants: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users, err := store.Users(t.Context())
			if err != nil {
				t.Fatalf("Store.Users() failed: %v", err)
			}
			var id string
			for _, u := range users {
				if u.Username == tc.username {
					id = u.ID
					break
				}
			}
			if id == "" {
				t.Fatal("Store.Users() user mismatch\ngot:  absent\nwant: present")
			}

			gotUser, gotGrants, err := store.UserWithGrants(t.Context(), id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			wantUser := mustUserByID(t, store, id)
			wantGrants, err := store.GrantsForUser(t.Context(), id)
			if err != nil {
				t.Fatalf("Store.GrantsForUser() failed: %v", err)
			}

			if gotUser.ID != wantUser.ID || gotUser.Username != wantUser.Username {
				t.Errorf(
					"parity: User mismatch\ngot:  ID=%q, Username=%q\nwant: ID=%q, Username=%q",
					gotUser.ID, gotUser.Username,
					wantUser.ID, wantUser.Username,
				)
			}
			if !slices.Equal(gotUser.Groups, wantUser.Groups) {
				t.Errorf(
					"parity: User.Groups mismatch\ngot:  %q\nwant: %q",
					gotUser.Groups, wantUser.Groups,
				)
			}
			if !slices.Equal(gotUser.Groups, tc.wantGroups) {
				t.Errorf("User.Groups mismatch\ngot:  %q\nwant: %q", gotUser.Groups, tc.wantGroups)
			}

			if got, want := len(gotGrants), tc.wantGrants; got != want {
				t.Errorf(
					"len(grants) mismatch\ngot:  %d - %q\nwant: %d",
					got, grantStrings(gotGrants), want,
				)
			}
			if got, want := len(gotGrants), len(wantGrants); got != want {
				t.Errorf("parity: len(grants) mismatch\ngot:  %d\nwant: %d", got, want)
			}
			for _, want := range wantGrants {
				if !slices.Contains(gotGrants, want) {
					t.Errorf(
						"parity: grants mismatch\ngot:  %q\nwant: %q",
						grantStrings(gotGrants), grantString(want),
					)
				}
			}
		})
	}
}

func TestStore_UserWithGrants__errors(t *testing.T) {
	store := testStore(t)
	mustCreateGroup(t, store, "readers", rbac.Grant{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeGlobal}})
	user := mustCreateUser(t, store, "disabled-user", "", "readers")
	disabled := false
	if _, err := store.UpdateUser(t.Context(), user.ID, UserPatch{Enabled: &disabled}); err != nil {
		t.Fatalf("Store.UpdateUser() failed: %v", err)
	}

	_, _, err := store.UserWithGrants(t.Context(), "does-not-exist")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown ID: error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}

	got, grants, err := store.UserWithGrants(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("disabled: unexpected error: %v", err)
	}

	// The row still loads; the caller gates on Enabled.
	if got.Enabled {
		t.Error("disabled: User.Enabled mismatch\ngot:  true\nwant: false")
	}
	if got, want := len(grants), 1; got != want {
		t.Errorf(
			"disabled: len(grants) mismatch\ngot:  %d - %q\nwant: %d",
			got, grantStrings(grants), want,
		)
	}
}

func TestStore_CreateUser_And_UserByID(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		groups     []string
		presetUser string // Username created beforehand.
		wantErr    error
		errRegex   string
	}{
		{
			name:     "user with no groups",
			username: "solo",
			errRegex: `^$`,
		},
		{
			name:     "user with groups",
			username: "grouped",
			groups:   []string{GroupViewer, GroupOperator},
			errRegex: `^$`,
		},
		{
			name:       "duplicate username rejected",
			username:   "taken",
			presetUser: "taken",
			wantErr:    ErrUsernameTaken,
			errRegex:   `^username already taken$`,
		},
		{
			name:       "duplicate username rejected case-insensitively",
			username:   "TAKEN", // users.username is COLLATE NOCASE.
			presetUser: "taken",
			wantErr:    ErrUsernameTaken,
			errRegex:   `^username already taken$`,
		},
		{
			name:     "unknown group rejected",
			username: "lost",
			groups:   []string{"no-such-group"},
			wantErr:  ErrUnknownGroup,
			errRegex: `^unknown group: "no-such-group"$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			if tc.presetUser != "" {
				mustCreateUser(t, store, tc.presetUser, "")
			}

			userDisplayName := "Display"
			userEmail := "user@example.com"

			user, err := store.CreateUser(
				t.Context(),
				tc.username,
				userDisplayName,
				userEmail,
				testPasswordHash,
				tc.groups,
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

			got := mustUserByID(t, store, user.ID)
			if testErr := test.AssertFields(t, []test.FieldAssertion{
				{Name: "Username", Got: got.Username, Want: tc.username, Mode: test.CompareEqual},
				{Name: "DisplayName", Got: got.DisplayName, Want: userDisplayName, Mode: test.CompareEqual},
				{Name: "Email", Got: got.Email, Want: userEmail, Mode: test.CompareEqual},
				{Name: "Enabled", Got: got.Enabled, Want: true, Mode: test.CompareEqual},
			}, "User"); testErr != nil {
				t.Error(testErr)
			}

			var wantGroups []string
			if len(tc.groups) > 0 {
				wantGroups = append(wantGroups, tc.groups...)
				slices.Sort(wantGroups)
			}
			if testErr := test.AssertSlicesEqual(t, got.Groups, wantGroups, "User.Groups"); testErr != nil {
				t.Fatal(testErr)
			}

			if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
				t.Errorf(
					"User timestamps mismatch\n"+
						"got:  CreatedAt=%v, UpdatedAt=%v\n"+
						"want: CreatedAt=non-zero, UpdatedAt=non-zero",
					got.CreatedAt, got.UpdatedAt,
				)
			}
		})
	}
}

func TestStore_CreateUser__requiresPassword(t *testing.T) {
	store := testStore(t)

	// Every account authenticates locally, so needs a password.
	_, err := store.CreateUser(
		t.Context(), "nopass", "", "", "", nil,
	)

	if !errors.Is(err, ErrNoPassword) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrNoPassword)
	}
}

func TestStore_CreateUser__duplicateGroupNames(t *testing.T) {
	store := testStore(t)

	user, err := store.CreateUser(
		t.Context(),
		"argus",
		"",
		"",
		testPasswordHash,
		[]string{GroupViewer, GroupViewer},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(user.Groups) != 1 || user.Groups[0] != GroupViewer {
		t.Errorf("User.Groups mismatch\ngot:  %q\nwant: %q", user.Groups, []string{GroupViewer})
	}

	grants, err := store.GrantsForUser(t.Context(), user.ID)
	if got, want := len(grants), seedGrantCount(t, GroupViewer); err != nil || got != want {
		t.Errorf(
			"len(Store.GrantsForUser()) mismatch\ngot:  %d, err=%v\nwant: %d, err=nil",
			got, err, want,
		)
	}
}

func TestStore_UpdateUser(t *testing.T) {
	tests := []struct {
		name         string
		targetGroups []string // Target user's initial groups.
		otherAdmin   bool     // A second enabled admin exists.
		patch        UserPatch
		wantErr      error
		errRegex     string
		check        func(t *testing.T, store *Store, userID string)
	}{
		{
			name:     "update display name",
			patch:    UserPatch{DisplayName: new("Renamed")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				if user.DisplayName != "Renamed" {
					t.Errorf(
						"User.DisplayName mismatch\ngot:  %q\nwant: %q",
						user.DisplayName, "Renamed",
					)
				}
			},
		},
		{
			name:     "update email",
			patch:    UserPatch{Email: new("new@example.com")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				if user.Email != "new@example.com" {
					t.Errorf(
						"User.Email mismatch\ngot:  %q\nwant: %q",
						user.Email, "new@example.com",
					)
				}
			},
		},
		{
			name:         "replace groups",
			targetGroups: []string{GroupViewer},
			patch:        UserPatch{Groups: new([]string{GroupOperator})},
			errRegex:     `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				if !slices.Equal(user.Groups, []string{GroupOperator}) {
					t.Errorf(
						"User.Groups mismatch\ngot:  %q\nwant: %q",
						user.Groups, []string{GroupOperator},
					)
				}
			},
		},
		{
			name:         "clear groups",
			targetGroups: []string{GroupViewer},
			patch:        UserPatch{Groups: new([]string{})},
			errRegex:     `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				if len(user.Groups) != 0 {
					t.Errorf("User.Groups mismatch\ngot:  %q\nwant: []", user.Groups)
				}
			},
		},
		{
			name:     "set password hash",
			patch:    UserPatch{PasswordHash: new("$argon2id$new")},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				creds := mustLocalCredentials(t, store, user.Username)
				if creds.PasswordHash != "$argon2id$new" {
					t.Errorf(
						"Credentials.PasswordHash mismatch\ngot:  %q\nwant: %q",
						creds.PasswordHash, "$argon2id$new",
					)
				}
			},
		},
		{
			name:     "unknown group rejected",
			patch:    UserPatch{Groups: new([]string{"no-such-group"})},
			wantErr:  ErrUnknownGroup,
			errRegex: `^unknown group: "no-such-group"$`,
		},
		{
			name:     "disable non-admin",
			patch:    UserPatch{Enabled: new(false)},
			errRegex: `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				if user.Enabled {
					t.Error("User.Enabled mismatch\ngot:  true\nwant: false")
				}
			},
		},
		{
			name:         "disable last admin rejected",
			targetGroups: []string{GroupAdmin},
			patch:        UserPatch{Enabled: new(false)},
			wantErr:      ErrLastAdmin,
			errRegex:     `^cannot delete, disable, or demote the last enabled admin$`,
		},
		{
			name:         "demote last admin rejected",
			targetGroups: []string{GroupAdmin},
			patch:        UserPatch{Groups: new([]string{GroupViewer})},
			wantErr:      ErrLastAdmin,
			errRegex:     `^cannot delete, disable, or demote the last enabled admin$`,
		},
		{
			name:         "disable admin with another admin present",
			targetGroups: []string{GroupAdmin},
			otherAdmin:   true,
			patch:        UserPatch{Enabled: new(false)},
			errRegex:     `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				if user.Enabled {
					t.Error("User.Enabled mismatch\ngot:  true\nwant: false")
				}
			},
		},
		{
			name:     "empty password hash rejected",
			patch:    UserPatch{PasswordHash: new("")},
			wantErr:  ErrNoPassword,
			errRegex: `^a password is required$`,
		},
		{
			name:         "demote admin with another admin present",
			targetGroups: []string{GroupAdmin},
			otherAdmin:   true,
			patch:        UserPatch{Groups: new([]string{})},
			errRegex:     `^$`,
			check: func(t *testing.T, store *Store, userID string) {
				user := mustUserByID(t, store, userID)
				if len(user.Groups) != 0 {
					t.Errorf("User.Groups mismatch\ngot:  %q\nwant: []", user.Groups)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			target := mustCreateUser(
				t,
				store,
				"target",
				"",
				tc.targetGroups...,
			)
			if tc.otherAdmin {
				mustCreateUser(
					t,
					store,
					"other-admin",
					"",
					GroupAdmin,
				)
			}

			_, err := store.UpdateUser(t.Context(), target.ID, tc.patch)
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
				tc.check(t, store, target.ID)
			}
		})
	}
}

func TestStore_UpdateUser__notFound(t *testing.T) {
	store := testStore(t)

	_, err := store.UpdateUser(t.Context(), "no-such-id", UserPatch{})

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}
}

func TestStore_DeleteUser(t *testing.T) {
	tests := []struct {
		name         string
		targetGroups []string
		disabled     bool // Target disabled before the delete.
		otherAdmin   bool
		wantErr      error
		errRegex     string
	}{
		{
			name:     "plain user",
			errRegex: `^$`,
		},
		{
			name:         "grouped user",
			targetGroups: []string{GroupViewer},
			errRegex:     `^$`,
		},
		{
			name:         "last admin rejected",
			targetGroups: []string{GroupAdmin},
			wantErr:      ErrLastAdmin,
			errRegex:     `^cannot delete, disable, or demote the last enabled admin$`,
		},
		{
			name:         "admin with another admin present",
			targetGroups: []string{GroupAdmin},
			otherAdmin:   true,
			errRegex:     `^$`,
		},
		{
			name:         "disabled admin (not the last enabled one)",
			targetGroups: []string{GroupAdmin},
			disabled:     true,
			otherAdmin:   true,
			errRegex:     `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			if tc.otherAdmin {
				mustCreateUser(
					t,
					store,
					"other-admin",
					"",
					GroupAdmin,
				)
			}
			target, err := store.CreateUser(
				t.Context(),
				"target",
				"Target",
				"",
				testPasswordHash,
				tc.targetGroups,
			)
			if err != nil {
				t.Fatalf("Store.CreateUser() failed: %v", err)
			}
			if tc.disabled {
				// Direct update to skip the rails.
				if _, err := store.db.Exec(
					`UPDATE users SET enabled = 0 WHERE id = ?;`, target.ID,
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			}
			// A session and membership that must be cascade-deleted.
			if err := store.InsertSession(t.Context(), Session{
				TokenHash: "hash-" + target.ID, UserID: target.ID,
				CreatedAt: timeNow(), ExpiresAt: timeNow(), LastSeenAt: timeNow(),
			}); err != nil {
				t.Fatalf("Store.InsertSession() failed: %v", err)
			}

			err = store.DeleteUser(t.Context(), target.ID)
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

			if _, err := store.UserByID(
				t.Context(),
				target.ID,
			); !errors.Is(err, ErrNotFound) {
				t.Errorf(
					"Store.UserByID() error type mismatch\ngot:  %v\nwant: %v",
					err, ErrNotFound,
				)
			}
			var count int
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM sessions WHERE user_id = ?;`,
				target.ID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"sessions row count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil (deleted)",
					count, err,
				)
			}
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM user_groups WHERE user_id = ?;`,
				target.ID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"user_groups row count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil (deleted)",
					count, err,
				)
			}
		})
	}
}

func TestStore_DeleteUser__notFound(t *testing.T) {
	store := testStore(t)

	err := store.DeleteUser(t.Context(), "no-such-id")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}
}

func TestStore_LocalCredentials(t *testing.T) {
	store := testStore(t)
	knownUsername := "argus"
	knownHash := "$argon2id$hash"
	mustCreateUser(
		t,
		store,
		knownUsername,
		knownHash,
		GroupViewer,
	)

	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{name: "exact username", username: "argus", want: true},
		{name: "case-insensitive username", username: "ARGUS", want: true},
		{name: "unknown username", username: "ghost", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares the users table, which the parent drops after the loop.

			creds, err := store.LocalCredentials(t.Context(), tc.username)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.want {
				if creds == nil {
					t.Fatal("credentials mismatch\ngot:  nil\nwant: non-nil")
				}
				if testErr := test.AssertFields(t, []test.FieldAssertion{
					{Name: "Username", Got: creds.Username, Want: knownUsername, Mode: test.CompareEqual},
					{Name: "PasswordHash", Got: creds.PasswordHash, Want: knownHash, Mode: test.CompareEqual},
					{Name: "Enabled", Got: creds.Enabled, Want: true, Mode: test.CompareEqual},
				}, "Credentials"); testErr != nil {
					t.Error(testErr)
				}
			} else if creds != nil {
				t.Errorf("credentials mismatch\ngot:  Username=%q\nwant: nil", creds.Username)
			}
		})
	}

	test.DropTable(t, store.db, "users")
	_, err := store.LocalCredentials(t.Context(), "argus")
	e := errfmt.FormatError(err)
	errRegex := `^load credentials:\n  SQL logic error: no such table: users`
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("missing table: error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_UpdatePasswordHash(t *testing.T) {
	const (
		username = "argus"
		oldHash  = "$argon2id$old"
		newHash  = "$argon2id$new"
	)

	tests := []struct {
		name       string
		hash       string
		dropUsers  bool // Drop the users table first.
		wantErr    error
		errRegex   string
		wantStored string // Hash held afterwards.
	}{
		{
			name:       "valid/replaces the stored hash",
			hash:       newHash,
			wantStored: newHash,
			errRegex:   `^$`,
		},
		{
			name:       "invalid/empty hash keeps the old one",
			hash:       "",
			wantErr:    ErrNoPassword,
			wantStored: oldHash,
			errRegex:   `^a password is required$`,
		},
		{
			name:      "invalid/users table unreadable",
			hash:      newHash,
			dropUsers: true,
			errRegex:  `^update password hash`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			user := mustCreateUser(t, store, username, oldHash)
			if tc.dropUsers {
				test.DropTable(t, store.db, "users")
			}

			err := store.UpdatePasswordHash(t.Context(), user.ID, tc.hash)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error type mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}

			if tc.dropUsers {
				return
			}
			creds := mustLocalCredentials(t, store, username)
			if creds.PasswordHash != tc.wantStored {
				t.Errorf(
					"Credentials.PasswordHash mismatch\ngot:  %q\nwant: %q",
					creds.PasswordHash, tc.wantStored,
				)
			}
		})
	}
}

func TestStore_CreateFirstAdmin(t *testing.T) {
	tests := []struct {
		name       string
		presetUser bool
		hashErr    bool
		raceUser   bool
		wantErr    error
		errRegex   string
	}{
		{
			name:     "empty database creates the admin",
			errRegex: `^$`,
		},
		{
			name:       "any existing user means setup is complete",
			presetUser: true,
			wantErr:    ErrSetupComplete,
			errRegex:   `^setup already completed$`,
		},
		{
			name:     "hash failure",
			hashErr:  true,
			errRegex: `^hash password`,
		},
		{
			name:     "another setup completed between the count and the insert",
			raceUser: true,
			wantErr:  ErrSetupComplete, // The guarded INSERT affects no rows.
			errRegex: `^setup already completed$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps hashPassword.

			store := testStore(t)
			if tc.presetUser {
				mustCreateUser(t, store, "existing", "")
			}
			if tc.hashErr {
				hashPasswordHad := hashPassword
				hashPassword = func(_ string) (string, error) {
					return "", errors.New("hash broke")
				}
				t.Cleanup(func() { hashPassword = hashPasswordHad })
			}
			if tc.raceUser {
				hashPasswordHad := hashPassword
				hashPassword = func(plaintext string) (string, error) {
					mustCreateUser(t, store, "racer", "")
					return hashPasswordHad(plaintext)
				}
				t.Cleanup(func() { hashPassword = hashPasswordHad })
			}

			user, err := store.CreateFirstAdmin(
				t.Context(),
				"root",
				"Rooty",
				"setup-password",
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("first call: error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("first call: error type mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if testErr := test.AssertFields(t, []test.FieldAssertion{
				{Name: "Username", Got: user.Username, Want: "root", Mode: test.CompareEqual},
				{Name: "DisplayName", Got: user.DisplayName, Want: "Rooty", Mode: test.CompareEqual},
				{Name: "Enabled", Got: user.Enabled, Want: true, Mode: test.CompareEqual},
			}, "User"); testErr != nil {
				t.Errorf("first call: %v", testErr)
			}
			if !slices.Contains(user.Groups, GroupAdmin) {
				t.Errorf(
					"first call: User.Groups mismatch\ngot:  %q\nwant: %q",
					user.Groups, GroupAdmin,
				)
			}
			creds := mustLocalCredentials(t, store, "root")
			match, _, err := password.Verify("setup-password", creds.PasswordHash)
			if err != nil || !match {
				t.Errorf(
					"first call: password.Verify() mismatch\n"+
						"got:  match=%t, err=%v\nwant: match=true, err=nil",
					match, err,
				)
			}

			_, err = store.CreateFirstAdmin(
				t.Context(),
				"root2",
				"",
				"another-password",
			)
			if !errors.Is(err, ErrSetupComplete) {
				t.Errorf(
					"second call: error type mismatch\ngot:  %v\nwant: %v",
					err, ErrSetupComplete,
				)
			}
		})
	}
}

func TestStore_ResetUserPassword(t *testing.T) {
	tests := []struct {
		name     string
		username string // Username to reset ("target" exists).
		disabled bool
		hashErr  bool
		wantErr  error
		errRegex string
	}{
		{
			name:     "existing user",
			username: "target",
			errRegex: `^$`,
		},
		{
			name:     "case-insensitive username",
			username: "TARGET",
			errRegex: `^$`,
		},
		{
			name:     "disabled user stays disabled",
			username: "target",
			disabled: true,
			errRegex: `^$`,
		},
		{
			name:     "unknown user",
			username: "ghost",
			wantErr:  ErrNotFound,
			errRegex: `^not found: user "ghost"$`,
		},
		{
			name:     "hash failure",
			username: "target",
			hashErr:  true,
			errRegex: `^hash password`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps hashPassword.

			store := testStore(t)
			user := mustCreateUser(t, store, "target", "$argon2id$old")
			if tc.disabled {
				if _, err := store.db.Exec(
					`UPDATE users SET enabled = 0 WHERE id = ?;`,
					user.ID,
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			}
			// A session that must be revoked by the reset.
			if err := store.InsertSession(
				t.Context(),
				Session{
					TokenHash: "hash-1", UserID: user.ID,
					CreatedAt: timeNow(), ExpiresAt: timeNow(), LastSeenAt: timeNow(),
				},
			); err != nil {
				t.Fatalf("Store.InsertSession() failed: %v", err)
			}
			if tc.hashErr {
				hashPasswordHad := hashPassword
				hashPassword = func(_ string) (string, error) {
					return "", errors.New("hash broke")
				}
				t.Cleanup(func() { hashPassword = hashPasswordHad })
			}

			err := store.ResetUserPassword(t.Context(), tc.username, "recovered-password")
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

			creds := mustLocalCredentials(t, store, "target")
			match, _, err := password.Verify("recovered-password", creds.PasswordHash)
			if err != nil || !match {
				t.Errorf(
					"password.Verify() mismatch\ngot:  match=%t, err=%v\nwant: match=true, err=nil",
					match, err,
				)
			}

			if creds.Enabled == tc.disabled {
				t.Errorf(
					"Credentials.Enabled mismatch\ngot:  %t\nwant: %t",
					creds.Enabled, !tc.disabled,
				)
			}

			var count int
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM sessions WHERE user_id = ?;`,
				user.ID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"sessions row count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil (revoked)",
					count, err,
				)
			}
		})
	}
}
