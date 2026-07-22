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
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestStore_APIToken__roundTrip(t *testing.T) {
	// GIVEN: a Store with a user.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	tokenName := "ci"

	prefix := fmt.Sprintf("%s\nAPI token round-trip", packageName)

	// WHEN: a token is created without expiry.
	plaintext, token, err := store.CreateAPIToken(
		t.Context(),
		user.ID,
		tokenName,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"%s CreateAPIToken failed: %v",
			prefix, err,
		)
	}

	// THEN: the plaintext is prefixed and the record mirrors it.
	if !strings.HasPrefix(plaintext, apiTokenPrefix) {
		t.Errorf(
			"%s token prefix mismatch\ngot: %q\nwant prefix: %q",
			prefix, plaintext, apiTokenPrefix,
		)
	}
	if got, want := len(plaintext), len(apiTokenPrefix)+apiTokenBytes*2; got != want {
		t.Errorf(
			"%s token length mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
	if token.Name != tokenName ||
		token.UserID != user.ID ||
		token.Prefix != plaintext[:apiTokenDisplayLength] ||
		token.ExpiresAt != nil ||
		token.LastUsedAt != nil {
		t.Errorf(
			"%s token record mismatch\ngot: %+v",
			prefix, *token,
		)
	}

	// AND: the plaintext is never stored.
	var count int
	if err := store.db.QueryRow(
		`SELECT COUNT(*) FROM api_tokens WHERE token_hash = ?;`,
		plaintext,
	).Scan(&count); err != nil || count != 0 {
		t.Errorf(
			"%s plaintext must not be a storage key\ngot:  %d rows, err=%v",
			prefix, count, err,
		)
	}

	// AND: the token resolves by plaintext.
	got, err := store.APITokenByToken(t.Context(), plaintext)
	if err != nil || got.ID != token.ID {
		t.Fatalf(
			"%s APITokenByToken mismatch\ngot:  %+v, err=%v",
			prefix, got, err,
		)
	}

	// AND: TouchAPIToken records usage.
	used := timeNow().Add(time.Minute)
	if err := store.TouchAPIToken(
		t.Context(),
		token.ID,
		used,
	); err != nil {
		t.Fatalf(
			"%s TouchAPIToken failed: %v",
			prefix, err,
		)
	}
	got, err = store.APITokenByToken(t.Context(), plaintext)
	if err != nil || got.LastUsedAt == nil || !got.LastUsedAt.Equal(used) {
		t.Errorf(
			"%s last_used_at mismatch\ngot:  %+v, err=%v\nwant: %v",
			prefix, got.LastUsedAt, err, used,
		)
	}

	// AND: DeleteAPIToken removes it.
	if err := store.DeleteAPIToken(
		t.Context(),
		user.ID,
		token.ID,
	); err != nil {
		t.Fatalf(
			"%s DeleteAPIToken failed: %v",
			prefix, err,
		)
	}
	if _, err := store.APITokenByToken(
		t.Context(),
		plaintext,
	); !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s token should be gone\ngot: %v",
			prefix, err,
		)
	}
}

func TestStore_APIToken__expiry(t *testing.T) {
	// GIVEN: a Store with a user.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	expires := timeNow().Add(time.Hour)

	prefix := fmt.Sprintf("%s\nAPI token expiry", packageName)

	// WHEN: a token is created with an expiry.
	plaintext, _, err := store.CreateAPIToken(t.Context(), user.ID, "short-lived", &expires)
	if err != nil {
		t.Fatalf(
			"%s CreateAPIToken failed: %v",
			prefix, err,
		)
	}

	// THEN: the expiry round-trips.
	got, err := store.APITokenByToken(t.Context(), plaintext)
	if err != nil || got.ExpiresAt == nil || !got.ExpiresAt.Equal(expires) {
		t.Errorf(
			"%s expiry mismatch\ngot:  %+v, err=%v\nwant: %v",
			prefix, got.ExpiresAt, err, expires,
		)
	}
}

func TestStore_APITokensForUser(t *testing.T) {
	// GIVEN: two users with tokens, created at distinct times.
	store := testStore(t)
	alpha := mustCreateUser(t, store, "alpha", "")
	beta := mustCreateUser(t, store, "beta", "")

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	timeNowHad := timeNow
	t.Cleanup(func() { timeNow = timeNowHad })
	tick := 0
	timeNow = func() time.Time {
		tick++
		return base.Add(time.Duration(tick) * time.Second)
	}

	for i, userID := range []string{alpha.ID, alpha.ID, beta.ID} {
		if _, _, err := store.CreateAPIToken(
			t.Context(),
			userID, fmt.Sprintf("token-%d", i), nil); err != nil {
			t.Fatalf(
				"%s\nsetup CreateAPIToken failed: %v",
				packageName, err,
			)
		}
	}

	prefix := fmt.Sprintf("%s\nAPITokensForUser()", packageName)

	// WHEN: alpha's tokens are listed.
	tokens, err := store.APITokensForUser(t.Context(), alpha.ID)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: only alpha's tokens are returned, newest first.
	want := []string{"token-1", "token-0"}
	if len(tokens) != 2 ||
		tokens[0].Name != want[0] || tokens[1].Name != want[1] {
		t.Errorf(
			"%s list mismatch\ngot:  %+v\nwant: [%v]",
			prefix, tokens, want,
		)
	}

	// AND: an unknown user has no tokens.
	if tokens, err := store.APITokensForUser(
		t.Context(),
		"no-such-id",
	); err != nil || len(tokens) != 0 {
		t.Errorf(
			"%s unknown user should list nothing\ngot:  %+v, err=%v",
			prefix, tokens, err,
		)
	}
}

func TestStore_DeleteAPIToken__rails(t *testing.T) {
	// GIVEN: two users, one holding a token.
	store := testStore(t)
	owner := mustCreateUser(t, store, "owner", "")
	other := mustCreateUser(t, store, "other", "")
	_, token, err := store.CreateAPIToken(
		t.Context(),
		owner.ID,
		"ci",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nDeleteAPIToken() rails", packageName)

	// WHEN: another user tries to delete the token.
	err = store.DeleteAPIToken(t.Context(), other.ID, token.ID)

	// THEN: it reads as not found.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s\ncross-user delete\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}

	// AND: unknown IDs read as not found.
	if err := store.DeleteAPIToken(
		t.Context(),
		owner.ID,
		"no-such-id",
	); !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s\nunknown ID\ngot:  %v\nwant: %v",
			prefix, err, ErrNotFound,
		)
	}
}

func TestStore_DeleteUser__cascadesAPITokens(t *testing.T) {
	// GIVEN: a user with a token.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	if _, _, err := store.CreateAPIToken(t.Context(), user.ID, "ci", nil); err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nDeleteUser() cascades api_tokens", packageName)

	// WHEN: the user is deleted.
	if err := store.DeleteUser(t.Context(), user.ID); err != nil {
		t.Fatalf(
			"%s DeleteUser failed: %v",
			prefix, err,
		)
	}

	// THEN: their tokens are gone.
	if got := countRows(t, store, "api_tokens"); got != 0 {
		t.Errorf(
			"%s token count mismatch\ngot:  %d\nwant: 0",
			prefix, got,
		)
	}
}

func TestStore_DeleteExpiredAPITokens(t *testing.T) {
	// GIVEN: a user with expired, live, and non-expiring tokens.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	past := timeNow().Add(-time.Hour)
	future := timeNow().Add(time.Hour)
	for _, tc := range []struct {
		name    string
		expires *time.Time
	}{
		{"expired", &past},
		{"live", &future},
		{"eternal", nil},
	} {
		if _, _, err := store.CreateAPIToken(t.Context(), user.ID, tc.name, tc.expires); err != nil {
			t.Fatalf(
				"%s\nsetup CreateAPIToken(%q) failed: %v",
				packageName, tc.name, err,
			)
		}
	}

	prefix := fmt.Sprintf("%s\nDeleteExpiredAPITokens()", packageName)

	// WHEN: expired tokens are pruned.
	if err := store.DeleteExpiredAPITokens(t.Context()); err != nil {
		t.Fatalf(
			"%s failed: %v",
			prefix, err,
		)
	}

	// THEN: only the expired token is gone.
	tokens, err := store.APITokensForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf(
			"%s APITokensForUser failed: %v",
			prefix, err,
		)
	}
	gotNames := make([]string, len(tokens))
	for i, token := range tokens {
		gotNames[i] = token.Name
	}
	slices.Sort(gotNames)
	if want := []string{"eternal", "live"}; !slices.Equal(gotNames, want) {
		t.Errorf(
			"%s remaining tokens mismatch\ngot:  %v\nwant: %v",
			prefix, gotNames, want,
		)
	}

	// AND: a missing table surfaces the failure.
	dropTable(t, store, "api_tokens")
	err = store.DeleteExpiredAPITokens(t.Context())
	if !util.RegexCheck(`^delete expired API tokens`, errfmt.FormatError(err)) {
		t.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, errfmt.FormatError(err), `^delete expired API tokens`,
		)
	}
}

func TestStore_APIToken__errors(t *testing.T) {
	// GIVEN: stores in states that break each token operation.
	tests := []struct {
		name    string
		randErr bool
		setup   func(t *testing.T, store *Store)
		invoke  func(t *testing.T, store *Store) error
	}{
		{
			name:    "create/random source fails",
			randErr: true,
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.CreateAPIToken(t.Context(), "uid", "x", nil)
				return err
			},
		},
		{
			name: "create/table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.CreateAPIToken(t.Context(), "uid", "x", nil)
				return err
			},
		},
		{
			name: "list/table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), "uid")
				return err
			},
		},
		{
			name: "by-token/table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokenByToken(t.Context(), "argus_x")
				return err
			},
		},
		{
			name: "touch/table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				return store.TouchAPIToken(t.Context(), "id", timeNow())
			},
		},
		{
			name: "delete/table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				return store.DeleteAPIToken(t.Context(), "uid", "id")
			},
		},
		{
			name: "scan/corrupt created_at",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "argus", "")
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, created_at)
					VALUES ('x', '%s', 'x', 'hash', 'argus_', 'not-a-time');`, user.ID))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), mustUserID(t, store, "argus"))
				return err
			},
		},
		{
			name: "scan/corrupt expires_at",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "argus", "")
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, created_at, expires_at)
					VALUES ('x', '%s', 'x', 'hash', 'argus_', '%s', 'not-a-time');`,
					user.ID, timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), mustUserID(t, store, "argus"))
				return err
			},
		},
		{
			name: "scan/corrupt last_used_at",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "argus", "")
				mustExec(t, store, fmt.Sprintf(`
					INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, created_at, last_used_at)
					VALUES ('x', '%s', 'x', 'hash', 'argus_', '%s', 'not-a-time');`,
					user.ID, timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), mustUserID(t, store, "argus"))
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
			if tc.randErr {
				randReadHad := randRead
				randRead = func(_ []byte) (int, error) {
					return 0, errors.New("rand broke")
				}
				t.Cleanup(func() { randRead = randReadHad })
			}

			// WHEN: the operation runs.
			err := tc.invoke(t, store)

			prefix := fmt.Sprintf("%s\nAPI token errors", packageName)

			// THEN: the failure is surfaced.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestStore_APIToken__rowsErr_and_rowsAffected(t *testing.T) {
	// GIVEN: a fault-injecting store with a user and token.
	store, state := testFaultStore(t)
	user := mustCreateUser(t, store, "argus", "")
	_, token, err := store.CreateAPIToken(
		t.Context(),
		user.ID,
		"ci",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nAPI token fault injection", packageName)

	// WHEN: list iteration fails.
	overrideRowsErrOnCall(t, 1)
	if _, err := store.APITokensForUser(t.Context(), user.ID); err == nil {
		t.Errorf("%s list iteration failure should surface", prefix)
	}

	// AND: RowsAffected fails on delete.
	state.SetRowsAffected(`DELETE FROM api_tokens WHERE id = ?`)
	err = store.DeleteAPIToken(t.Context(), user.ID, token.ID)
	state.Set("")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s RowsAffected failure should surface as an error\ngot: %v",
			prefix, err,
		)
	}
}

// mustUserID resolves a username to its ID.
func mustUserID(t *testing.T, store *Store, username string) string {
	t.Helper()

	creds, err := store.LocalCredentials(t.Context(), username)
	if err != nil || creds == nil {
		t.Fatalf(
			"%s\nlookup %q failed: %v",
			packageName, username, err,
		)
	}
	return creds.UserID
}
