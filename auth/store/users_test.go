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

	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestStore_CountUsers(t *testing.T) {
	// GIVEN: a Store with one user.
	store := testStore(t)
	mustCreateUser(t, store, "alpha", "")

	prefix := fmt.Sprintf("%s\nCountUsers()", packageName)

	// WHEN: CountUsers is called.
	count, err := store.CountUsers(t.Context())

	// THEN: it reports one user.
	if err != nil || count != 1 {
		t.Errorf(
			"%s result mismatch\ngot:  %d, err=%v\nwant: 1",
			prefix, count, err,
		)
	}

	// AND: it errors when the table is unreadable.
	dropTable(t, store, "users")
	if _, err := store.CountUsers(t.Context()); err == nil {
		t.Errorf("%s expected an error after dropping users", prefix)
	}
}

func TestStore_Users(t *testing.T) {
	// GIVEN: a Store with two users, one grouped.
	store := testStore(t)
	mustCreateUser(t, store, "beta", "", GroupViewer)
	mustCreateUser(t, store, "alpha", "")

	prefix := fmt.Sprintf("%s\nUsers()", packageName)

	// WHEN: Users is called.
	users, err := store.Users(t.Context())
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: users are returned ordered by username.
	if len(users) != 2 || users[0].Username != "alpha" || users[1].Username != "beta" {
		t.Fatalf(
			"%s ordering mismatch\ngot:  %+v",
			prefix, users,
		)
	}

	// AND: group memberships are attached.
	if len(users[0].Groups) != 0 || !slices.Equal(users[1].Groups, []string{GroupViewer}) {
		t.Errorf(
			"%s groups mismatch\ngot:  alpha=%v beta=%v",
			prefix, users[0].Groups, users[1].Groups,
		)
	}
}

func TestStore_Users__Errors(t *testing.T) {
	// GIVEN: stores in states that break listing.
	tests := []struct {
		name  string
		setup func(t *testing.T, store *Store)
	}{
		{
			name: "users table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "users")
			},
		},
		{
			name: "corrupt user timestamp",
			setup: func(t *testing.T, store *Store) {
				if _, err := store.db.Exec(`
					INSERT INTO users (id, username, created_at, updated_at)
					VALUES ('x', 'x', 'not-a-time', 'not-a-time');`,
				); err != nil {
					t.Fatalf(
						"%s setup failed: %v",
						packageName, err,
					)
				}
			},
		},
		{
			name: "user_groups table missing",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "alpha", "")
				dropTable(t, store, "user_groups")
			},
		},
		{
			name: "NULL membership row",
			setup: func(t *testing.T, store *Store) {
				mustCreateUser(t, store, "alpha", "")
				dropTable(t, store, "user_groups")
				for _, statement := range []string{
					`CREATE TABLE user_groups (user_id TEXT, group_id TEXT);`,
					`INSERT INTO user_groups (user_id, group_id)
						VALUES (NULL, (SELECT id FROM groups WHERE name = 'viewer'));`,
				} {
					if _, err := store.db.Exec(statement); err != nil {
						t.Fatalf(
							"%s setup failed: %v",
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

			// WHEN: Users is called.
			_, err := store.Users(t.Context())

			prefix := fmt.Sprintf("%s\nUsers()", packageName)

			// THEN: the failure is surfaced.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestStore_Users__RowsErr(t *testing.T) {
	// GIVEN: a Store whose row iteration fails.
	store := testStore(t)
	rowsErrHad := rowsErr
	wantErr := errors.New("iteration broke")
	rowsErr = func(_ *sql.Rows) error { return wantErr }
	t.Cleanup(func() { rowsErr = rowsErrHad })

	prefix := fmt.Sprintf("%s\nUsers() with failing iteration", packageName)

	// WHEN: Users is called.
	_, err := store.Users(t.Context())

	// THEN: the iteration failure is surfaced.
	if !errors.Is(err, wantErr) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, wantErr,
		)
	}
}

func TestStore_CreateUser_And_UserByID(t *testing.T) {
	// GIVEN: a Store.
	tests := []struct {
		name         string
		username     string
		passwordHash string
		groups       []string
		presetUser   string // Username created beforehand.
		wantErr      error
	}{
		{
			name:     "user with no groups",
			username: "solo",
		},
		{
			name:     "user with groups",
			username: "grouped",
			groups:   []string{GroupViewer, GroupOperator},
		},
		{
			name:       "duplicate username rejected",
			username:   "taken",
			presetUser: "taken",
			wantErr:    ErrUsernameTaken,
		},
		{
			name:       "duplicate username rejected case-insensitively",
			username:   "TAKEN",
			presetUser: "taken",
			wantErr:    ErrUsernameTaken,
		},
		{
			name:     "unknown group rejected",
			username: "lost",
			groups:   []string{"no-such-group"},
			wantErr:  ErrUnknownGroup,
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

			// WHEN: CreateUser is called.
			user, err := store.CreateUser(
				t.Context(),
				tc.username,
				userDisplayName,
				userEmail,
				tc.passwordHash,
				tc.groups,
			)

			prefix := fmt.Sprintf("%s\nCreateUser()", packageName)

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

			// AND: the user round-trips through UserByID.
			got, err := store.UserByID(t.Context(), user.ID)
			if err != nil {
				t.Fatalf(
					"%s UserByID failed: %v",
					prefix, err,
				)
			}
			if got.Username != tc.username ||
				got.DisplayName != userDisplayName ||
				got.Email != userEmail ||
				!got.Enabled {
				t.Errorf(
					"%s user mismatch\ngot:  %+v",
					prefix, *got,
				)
			}

			// AND: group memberships match (sorted).
			wantGroups := append([]string{}, tc.groups...)
			slices.Sort(wantGroups)
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got.Groups,
				wantGroups,
				func(a, b string) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}

			// AND: timestamps are populated.
			if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
				t.Errorf(
					"%s timestamps should be set\ngot:  created=%v updated=%v",
					prefix, got.CreatedAt, got.UpdatedAt,
				)
			}
		})
	}
}

func TestStore_UserByID__NotFound(t *testing.T) {
	// GIVEN: a Store.
	store := testStore(t)

	prefix := fmt.Sprintf("%s\nUserByID(unknown)", packageName)

	// WHEN: UserByID is called with an unknown ID.
	_, err := store.UserByID(t.Context(), "no-such-id")

	// THEN: ErrNotFound is returned.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}
}

func TestStore_CreateUser__DuplicateGroupNames(t *testing.T) {
	// GIVEN: a create request naming the same group twice.
	store := testStore(t)

	prefix := fmt.Sprintf("%s\nCreateUser() with duplicate group names", packageName)

	// WHEN: the user is created.
	user, err := store.CreateUser(t.Context(),
		"argus",
		"",
		"",
		"",
		[]string{GroupViewer, GroupViewer},
	)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: the membership is stored once.
	if len(user.Groups) != 1 || user.Groups[0] != GroupViewer {
		t.Errorf(
			"%s memberships should dedupe\ngot:  %v\nwant: [%s]",
			prefix, user.Groups, GroupViewer,
		)
	}

	// AND: the grants are not multiplied.
	grants, err := store.GrantsForUser(t.Context(), user.ID)
	if got, want := len(grants), 2; err != nil || got != want { // viewer: service:read + config:read.
		t.Errorf(
			"%s grant count mismatch\ngot:  %d, err=%v\nwant: 2",
			prefix, len(grants), err,
		)
	}
}

func TestStore_UpdateUser(t *testing.T) {
	// GIVEN: a Store with users in various admin states.
	tests := []struct {
		name         string
		targetGroups []string // Target user's initial groups.
		otherAdmin   bool     // A second enabled admin exists.
		patch        UserPatch
		wantErr      error
		check        func(t *testing.T, store *Store, prefix string, userID string)
	}{
		{
			name:  "update display name",
			patch: UserPatch{DisplayName: test.Ptr("Renamed")},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				user, _ := store.UserByID(t.Context(), userID)
				if user.DisplayName != "Renamed" {
					t.Errorf(
						"%s display name not updated: %+v",
						prefix, *user,
					)
				}
			},
		},
		{
			name:  "update email",
			patch: UserPatch{Email: test.Ptr("new@example.com")},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				user, _ := store.UserByID(t.Context(), userID)
				if user.Email != "new@example.com" {
					t.Errorf(
						"%s email not updated: %+v",
						prefix, *user,
					)
				}
			},
		},
		{
			name:         "replace groups",
			targetGroups: []string{GroupViewer},
			patch:        UserPatch{Groups: test.Ptr([]string{GroupOperator})},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				user, _ := store.UserByID(t.Context(), userID)
				if !slices.Equal(user.Groups, []string{GroupOperator}) {
					t.Errorf(
						"%s groups not replaced: %v",
						prefix, user.Groups,
					)
				}
			},
		},
		{
			name:         "clear groups",
			targetGroups: []string{GroupViewer},
			patch:        UserPatch{Groups: test.Ptr([]string{})},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				user, _ := store.UserByID(t.Context(), userID)
				if len(user.Groups) != 0 {
					t.Errorf(
						"%s groups not cleared: %v",
						prefix, user.Groups,
					)
				}
			},
		},
		{
			name:  "set password hash",
			patch: UserPatch{PasswordHash: test.Ptr("$argon2id$new")},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				creds, _ := store.LocalCredentials(t.Context(), "target")
				if creds.PasswordHash != "$argon2id$new" {
					t.Errorf(
						"%s password hash not updated: %q",
						prefix, creds.PasswordHash,
					)
				}
			},
		},
		{
			name:    "unknown group rejected",
			patch:   UserPatch{Groups: test.Ptr([]string{"no-such-group"})},
			wantErr: ErrUnknownGroup,
		},
		{
			name:  "disable non-admin",
			patch: UserPatch{Enabled: test.Ptr(false)},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				user, _ := store.UserByID(t.Context(), userID)
				if user.Enabled {
					t.Errorf("%s user should be disabled", prefix)
				}
			},
		},
		{
			name:         "disable last admin rejected",
			targetGroups: []string{GroupAdmin},
			patch:        UserPatch{Enabled: test.Ptr(false)},
			wantErr:      ErrLastAdmin,
		},
		{
			name:         "demote last admin rejected",
			targetGroups: []string{GroupAdmin},
			patch:        UserPatch{Groups: test.Ptr([]string{GroupViewer})},
			wantErr:      ErrLastAdmin,
		},
		{
			name:         "disable admin with another admin present",
			targetGroups: []string{GroupAdmin},
			otherAdmin:   true,
			patch:        UserPatch{Enabled: test.Ptr(false)},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				user, _ := store.UserByID(t.Context(), userID)
				if user.Enabled {
					t.Errorf("%s user should be disabled", prefix)
				}
			},
		},
		{
			name:         "demote admin with another admin present",
			targetGroups: []string{GroupAdmin},
			otherAdmin:   true,
			patch:        UserPatch{Groups: test.Ptr([]string{})},
			check: func(t *testing.T, store *Store, prefix, userID string) {
				user, _ := store.UserByID(t.Context(), userID)
				if len(user.Groups) != 0 {
					t.Errorf(
						"%s user should be demoted: %v",
						prefix, user.Groups,
					)
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

			// WHEN: UpdateUser is called.
			_, err := store.UpdateUser(t.Context(), target.ID, tc.patch)

			prefix := fmt.Sprintf("%s\nUpdateUser()", packageName)

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
				tc.check(t, store, prefix, target.ID)
			}
		})
	}
}

func TestStore_UpdateUser__NotFound(t *testing.T) {
	// GIVEN: a Store.
	store := testStore(t)

	// WHEN: UpdateUser is called with an unknown ID.
	_, err := store.UpdateUser(t.Context(), "no-such-id", UserPatch{})

	prefix := fmt.Sprintf("%s\nUpdateUser(unknown)", packageName)

	// THEN: ErrNotFound is returned.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}
}

func TestStore_DeleteUser(t *testing.T) {
	// GIVEN: a Store with users in various states.
	tests := []struct {
		name         string
		targetGroups []string
		disabled     bool // Target disabled before the delete.
		otherAdmin   bool
		wantErr      error
	}{
		{
			name: "plain user",
		},
		{
			name:         "grouped user",
			targetGroups: []string{GroupViewer},
		},
		{
			name:         "last admin rejected",
			targetGroups: []string{GroupAdmin},
			wantErr:      ErrLastAdmin,
		},
		{
			name:         "admin with another admin present",
			targetGroups: []string{GroupAdmin},
			otherAdmin:   true,
		},
		{
			name:         "disabled admin (not the last enabled one)",
			targetGroups: []string{GroupAdmin},
			disabled:     true,
			otherAdmin:   true,
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
				"",
				tc.targetGroups,
			)
			if err != nil {
				t.Fatalf(
					"%s\nsetup CreateUser failed: %v",
					packageName, err,
				)
			}
			if tc.disabled {
				// Direct update to skip the rails.
				if _, err := store.db.Exec(
					`UPDATE users SET enabled = 0 WHERE id = ?;`, target.ID,
				); err != nil {
					t.Fatalf(
						"%s\nsetup disable failed: %v",
						packageName, err,
					)
				}
			}
			// A session and membership that must be cascade-deleted.
			if err := store.InsertSession(t.Context(), Session{
				TokenHash: "hash-" + target.ID, UserID: target.ID,
				CreatedAt: timeNow(), ExpiresAt: timeNow(), LastSeenAt: timeNow(),
			}); err != nil {
				t.Fatalf(
					"%s\nsetup InsertSession failed: %v",
					packageName, err,
				)
			}

			// WHEN: DeleteUser is called.
			err = store.DeleteUser(t.Context(), target.ID)

			prefix := fmt.Sprintf("%s\nDeleteUser()", packageName)

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

			// AND: the user and their attachments are gone.
			if _, err := store.UserByID(
				t.Context(),
				target.ID,
			); !errors.Is(err, ErrNotFound) {
				t.Errorf(
					"%s user should be gone\ngot: %v",
					prefix, err,
				)
			}
			var count int
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM sessions WHERE user_id = ?;`,
				target.ID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"%s sessions should be cascade-deleted\ngot:  %d, err=%v",
					prefix, count, err,
				)
			}
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM user_groups WHERE user_id = ?;`,
				target.ID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"%s memberships should be cascade-deleted\ngot:  %d, err=%v",
					prefix, count, err,
				)
			}
		})
	}
}

func TestStore_DeleteUser__NotFound(t *testing.T) {
	// GIVEN: a Store.
	store := testStore(t)

	// WHEN: DeleteUser is called with an unknown ID.
	err := store.DeleteUser(t.Context(), "no-such-id")

	prefix := fmt.Sprintf("%s\nDeleteUser(unknown)", packageName)

	// THEN: ErrNotFound is returned.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}
}

func TestStore_LocalCredentials(t *testing.T) {
	// GIVEN: a Store with a user holding a password hash.
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
			// WHEN: LocalCredentials is called.
			creds, err := store.LocalCredentials(t.Context(), tc.username)

			prefix := fmt.Sprintf(
				"%s\nLocalCredentials(%q)",
				packageName, tc.username,
			)

			// THEN: no error occurs.
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// AND: known users return their material; unknown return nil.
			if tc.want {
				if creds == nil ||
					creds.Username != knownUsername ||
					creds.PasswordHash != knownHash ||
					!creds.Enabled {
					t.Errorf(
						"%s credentials mismatch\ngot: %+v",
						prefix, creds,
					)
				}
			} else if creds != nil {
				t.Errorf(
					"%s expected nil credentials\ngot: %+v",
					prefix, creds,
				)
			}
		})
	}

	// AND: it errors when the table is unreadable.
	dropTable(t, store, "users")
	if _, err := store.LocalCredentials(t.Context(), "argus"); err == nil {
		t.Errorf(
			"%s\nexpected an error after dropping users",
			packageName,
		)
	}
}

func TestStore_UpdatePasswordHash(t *testing.T) {
	// GIVEN: a Store with a user.
	store := testStore(t)
	username := "argus"
	oldHash := "$argon2id$old"
	user := mustCreateUser(
		t,
		store,
		username,
		oldHash,
	)

	// WHEN: UpdatePasswordHash is called.
	newHash := "$argon2id$new"
	err := store.UpdatePasswordHash(
		t.Context(),
		user.ID,
		newHash,
	)

	prefix := fmt.Sprintf(
		"%s\nUpdatePasswordHash()",
		packageName,
	)

	// THEN: the stored hash changes.
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}
	creds, err := store.LocalCredentials(t.Context(), username)
	if err != nil || creds.PasswordHash != newHash {
		t.Errorf(
			"%s hash mismatch\ngot:  %q, err=%v\nwant: %q",
			prefix, creds.PasswordHash, err, newHash,
		)
	}

	// AND: it errors when the table is unreadable.
	dropTable(t, store, "users")
	if err := store.UpdatePasswordHash(
		t.Context(),
		user.ID,
		"x",
	); err == nil {
		t.Errorf("%s expected an error after dropping users", prefix)
	}
}

func TestStore_CreateFirstAdmin(t *testing.T) {
	// GIVEN: Stores with and without existing users.
	tests := []struct {
		name       string
		presetUser bool
		hashErr    bool
		wantErr    error
		errRegex   string
	}{
		{
			name: "empty database creates the admin",
		},
		{
			name:       "any existing user means setup is complete",
			presetUser: true,
			wantErr:    ErrSetupComplete,
		},
		{
			name:     "hash failure",
			hashErr:  true,
			errRegex: `^hash password`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			// WHEN: CreateFirstAdmin is called.
			user, err := store.CreateFirstAdmin(
				t.Context(), "root", "Rooty", "setup-password")

			prefix := fmt.Sprintf("%s\nCreateFirstAdmin()", packageName)

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
			if tc.errRegex != "" {
				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, e, tc.errRegex,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf("%s unexpected error: %v", prefix, err)
			}

			// AND: the admin is an enabled admin member whose password verifies.
			if user.Username != "root" || user.DisplayName != "Rooty" ||
				!user.Enabled ||
				!slices.Contains(user.Groups, GroupAdmin) {
				t.Errorf(
					"%s admin mismatch\ngot: %+v",
					prefix, user,
				)
			}
			creds, err := store.LocalCredentials(t.Context(), "root")
			if err != nil || creds == nil {
				t.Fatalf(
					"%s admin lookup failed: %v",
					prefix, err,
				)
			}
			match, _, err := password.Verify("setup-password", creds.PasswordHash)
			if err != nil || !match {
				t.Errorf(
					"%s admin password should verify\ngot:  match=%t, err=%v",
					prefix, match, err,
				)
			}

			// AND: a second call is rejected - setup happens exactly once.
			if _, err := store.CreateFirstAdmin(
				t.Context(), "root2", "", "another-password",
			); !errors.Is(err, ErrSetupComplete) {
				t.Errorf(
					"%s second call error mismatch\ngot:  %v\nwant: %v",
					prefix, err, ErrSetupComplete,
				)
			}
		})
	}
}

func TestStore_ResetUserPassword(t *testing.T) {
	// GIVEN: users in various states.
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
		},
		{
			name:     "case-insensitive username",
			username: "TARGET",
		},
		{
			name:     "disabled user stays disabled",
			username: "target",
			disabled: true,
		},
		{
			name:     "unknown user",
			username: "ghost",
			wantErr:  ErrNotFound,
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
			store := testStore(t)
			user := mustCreateUser(t, store, "target", "$argon2id$old")
			if tc.disabled {
				if _, err := store.db.Exec(
					`UPDATE users SET enabled = 0 WHERE id = ?;`,
					user.ID,
				); err != nil {
					t.Fatalf(
						"%s\nsetup disable failed: %v",
						packageName, err,
					)
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
				t.Fatalf(
					"%s\nsetup InsertSession failed: %v",
					packageName, err,
				)
			}
			if tc.hashErr {
				hashPasswordHad := hashPassword
				hashPassword = func(_ string) (string, error) {
					return "", errors.New("hash broke")
				}
				t.Cleanup(func() { hashPassword = hashPasswordHad })
			}

			// WHEN: the password is reset.
			err := store.ResetUserPassword(t.Context(), tc.username, "recovered-password")

			prefix := fmt.Sprintf("%s\nResetUserPassword()", packageName)

			// THEN: errors match expectations.
			errRegex := tc.errRegex
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf(
						"%s error mismatch\ngot:  %v\nwant: %v",
						prefix, err, tc.wantErr,
					)
				}
				return
			} else if tc.errRegex == "" {
				errRegex = `^$`
			}
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %v\nwant: %v",
					prefix, err, errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: the new password verifies.
			creds, err := store.LocalCredentials(t.Context(), "target")
			if err != nil || creds == nil {
				t.Fatalf(
					"%s lookup failed: %v",
					prefix, err,
				)
			}
			match, _, err := password.Verify("recovered-password", creds.PasswordHash)
			if err != nil || !match {
				t.Errorf(
					"%s new password should verify\ngot:  match=%t, err=%v",
					prefix, match, err,
				)
			}

			// AND: the enabled state is untouched.
			if creds.Enabled == tc.disabled {
				t.Errorf(
					"%s enabled state should be untouched\ngot:  %t\nwant: %t",
					prefix, creds.Enabled, !tc.disabled,
				)
			}

			// AND: their sessions are revoked.
			var count int
			if err := store.db.QueryRow(
				`SELECT COUNT(*) FROM sessions WHERE user_id = ?;`,
				user.ID,
			).Scan(&count); err != nil || count != 0 {
				t.Errorf(
					"%s sessions should be revoked\ngot:  %d, err=%v",
					prefix, count, err,
				)
			}
		})
	}
}
