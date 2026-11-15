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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestStore_APIToken__roundTrip(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	tokenName := "ci"

	plaintext, token, err := store.CreateAPIToken(
		t.Context(),
		user.ID,
		tokenName,
		nil,
	)
	if err != nil {
		t.Fatalf("Store.CreateAPIToken() failed: %v", err)
	}

	if !strings.HasPrefix(plaintext, apiTokenPrefix) {
		t.Errorf(
			"after create: plaintext prefix mismatch\ngot:  %q\nwant: %q",
			plaintext, apiTokenPrefix,
		)
	}
	if got, want := len(plaintext), len(apiTokenPrefix)+apiTokenBytes*2; got != want {
		t.Errorf("after create: len(plaintext) mismatch\ngot:  %d\nwant: %d", got, want)
	}
	if testErr := test.AssertFields(t, []test.FieldAssertion{
		{Name: "Name", Got: token.Name, Want: tokenName, Mode: test.CompareEqual},
		{Name: "UserID", Got: token.UserID, Want: user.ID, Mode: test.CompareEqual},
		{Name: "Prefix", Got: token.Prefix, Want: plaintext[:apiTokenDisplayLength], Mode: test.CompareEqual},
		{Name: "ExpiresAt", Got: test.StringifyPtr(token.ExpiresAt), Want: "nil", Mode: test.CompareEqual},
		{Name: "LastUsedAt", Got: test.StringifyPtr(token.LastUsedAt), Want: "nil", Mode: test.CompareEqual},
	}, "APIToken"); testErr != nil {
		t.Errorf("after create: %v", testErr)
	}

	// The plaintext is never stored.
	var count int
	if err := store.db.QueryRow(
		`SELECT COUNT(*) FROM api_tokens
		WHERE ? IN (id, user_id, name, token_hash, prefix, created_at, expires_at, last_used_at);`,
		plaintext,
	).Scan(&count); err != nil || count != 0 {
		t.Errorf(
			"after create: api_tokens row count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil",
			count, err,
		)
	}

	got, err := store.APITokenByToken(t.Context(), plaintext)
	if err != nil {
		t.Fatalf("after create: Store.APITokenByToken() failed: %v", err)
	}
	if got.ID != token.ID {
		t.Fatalf("after create: APIToken.ID mismatch\ngot:  %q\nwant: %q", got.ID, token.ID)
	}

	used := timeNow().Add(time.Minute)
	if err := store.TouchAPIToken(t.Context(), token.ID, used); err != nil {
		t.Fatalf("Store.TouchAPIToken() failed: %v", err)
	}
	got, err = store.APITokenByToken(t.Context(), plaintext)
	if err != nil {
		t.Fatalf("after touch: Store.APITokenByToken() failed: %v", err)
	}
	if got.LastUsedAt == nil || !got.LastUsedAt.Equal(used) {
		t.Errorf(
			"after touch: APIToken.LastUsedAt mismatch\ngot:  %v\nwant: %v",
			test.StringifyPtr(got.LastUsedAt), used,
		)
	}

	if err := store.DeleteAPIToken(t.Context(), user.ID, token.ID); err != nil {
		t.Fatalf("Store.DeleteAPIToken() failed: %v", err)
	}
	if _, err := store.APITokenByToken(
		t.Context(),
		plaintext,
	); !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"after delete: Store.APITokenByToken() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrNotFound,
		)
	}
}

func TestStore_CreateAPIToken__perUserCap(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	other := mustCreateUser(t, store, "other", "")
	for i := range maxAPITokensPerUser {
		if _, _, err := store.CreateAPIToken(
			t.Context(), user.ID, fmt.Sprintf("t%d", i), nil,
		); err != nil {
			t.Fatalf("Store.CreateAPIToken(t%d) failed: %v", i, err)
		}
	}

	_, _, err := store.CreateAPIToken(t.Context(), user.ID, "over", nil)

	if !errors.Is(err, ErrTokenLimitReached) {
		t.Errorf("at cap: error type mismatch\ngot:  %v\nwant: %v", err, ErrTokenLimitReached)
	}

	if _, _, err := store.CreateAPIToken(t.Context(), other.ID, "ok", nil); err != nil {
		t.Errorf("other user: Store.CreateAPIToken() failed: %v", err)
	}
}

func TestStore_APIToken__expiry(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	expires := timeNow().Add(time.Hour)

	plaintext, _, err := store.CreateAPIToken(t.Context(), user.ID, "short-lived", &expires)
	if err != nil {
		t.Fatalf("Store.CreateAPIToken() failed: %v", err)
	}

	got, err := store.APITokenByToken(t.Context(), plaintext)
	if err != nil {
		t.Fatalf("Store.APITokenByToken() failed: %v", err)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(expires) {
		t.Errorf(
			"APIToken.ExpiresAt mismatch\ngot:  %v\nwant: %v",
			test.StringifyPtr(got.ExpiresAt), expires,
		)
	}
}

func TestStore_APITokensForUser(t *testing.T) {
	store := testStore(t)
	alpha := mustCreateUser(t, store, "alpha", "")
	beta := mustCreateUser(t, store, "beta", "")

	// Distinct creation times, so newest-first order is observable.
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
			t.Fatalf("Store.CreateAPIToken(token-%d) failed: %v", i, err)
		}
	}

	tokens, err := store.APITokensForUser(t.Context(), alpha.ID)
	if err != nil {
		t.Fatalf("alpha: unexpected error: %v", err)
	}

	gotNames := make([]string, len(tokens))
	for i, token := range tokens {
		gotNames[i] = token.Name
	}
	if want := []string{"token-1", "token-0"}; !slices.Equal(gotNames, want) {
		t.Errorf("alpha: token names mismatch\ngot:  %q\nwant: %q", gotNames, want)
	}

	tokens, err = store.APITokensForUser(t.Context(), "no-such-id")
	if err != nil || len(tokens) != 0 {
		t.Errorf(
			"unknown user: token count mismatch\ngot:  %d, err=%v\nwant: 0, err=nil",
			len(tokens), err,
		)
	}
}

func TestStore_DeleteAPIToken__rails(t *testing.T) {
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
		t.Fatalf("Store.CreateAPIToken() failed: %v", err)
	}

	err = store.DeleteAPIToken(t.Context(), other.ID, token.ID)

	// The delete matches only the caller's tokens.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-user: error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}

	err = store.DeleteAPIToken(t.Context(), owner.ID, "no-such-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown ID: error type mismatch\ngot:  %v\nwant: %v", err, ErrNotFound)
	}
}

func TestStore_DeleteUser__cascadesAPITokens(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	if _, _, err := store.CreateAPIToken(t.Context(), user.ID, "ci", nil); err != nil {
		t.Fatalf("Store.CreateAPIToken() failed: %v", err)
	}

	if err := store.DeleteUser(t.Context(), user.ID); err != nil {
		t.Fatalf("Store.DeleteUser() failed: %v", err)
	}

	if got, want := test.CountRows(t, store.db, "api_tokens"), 0; got != want {
		t.Errorf("api_tokens row count mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestStore_DeleteExpiredAPITokens(t *testing.T) {
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
			t.Fatalf("Store.CreateAPIToken(%q) failed: %v", tc.name, err)
		}
	}

	err := store.DeleteExpiredAPITokens(t.Context())
	if err != nil {
		t.Fatalf("before DropTable: Store.DeleteExpiredAPITokens() failed: %v", err)
	}

	tokens, err := store.APITokensForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("before DropTable: Store.APITokensForUser() failed: %v", err)
	}
	gotNames := test.AttributesSorted(tokens, func(token APIToken) string { return token.Name })
	if want := []string{"eternal", "live"}; !slices.Equal(gotNames, want) {
		t.Errorf("before DropTable: token names mismatch\ngot:  %q\nwant: %q", gotNames, want)
	}

	test.DropTable(t, store.db, "api_tokens")
	err = store.DeleteExpiredAPITokens(t.Context())
	e := errfmt.FormatError(err)
	errRegex := `^delete expired API tokens`
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("missing table: error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_APIToken__errors(t *testing.T) {
	tests := []struct {
		name     string
		randErr  bool
		setup    func(t *testing.T, store *Store)
		invoke   func(t *testing.T, store *Store) error
		errRegex string
	}{
		{
			name:    "create/random source fails",
			randErr: true,
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.CreateAPIToken(t.Context(), "uid", "x", nil)
				return err
			},
			errRegex: test.TrimYAML(`
				^generate API token:
					rand broke$`),
		},
		{
			name: "create/table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				_, _, err := store.CreateAPIToken(t.Context(), "uid", "x", nil)
				return err
			},
			errRegex: test.TrimYAML(`
				^insert API token:
					SQL logic error: no such table: api_tokens`),
		},
		{
			name: "list/table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), "uid")
				return err
			},
			errRegex: test.TrimYAML(`
				^list API tokens:
					SQL logic error: no such table: api_tokens`),
		},
		{
			name: "by-token/table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokenByToken(t.Context(), "argus_x")
				return err
			},
			errRegex: test.TrimYAML(`
				^scan API token:
					SQL logic error: no such table: api_tokens`),
		},
		{
			name: "touch/table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				return store.TouchAPIToken(t.Context(), "id", timeNow())
			},
			errRegex: test.TrimYAML(`
				^touch API token:
					SQL logic error: no such table: api_tokens`),
		},
		{
			name: "delete/table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "api_tokens")
			},
			invoke: func(t *testing.T, store *Store) error {
				return store.DeleteAPIToken(t.Context(), "uid", "id")
			},
			errRegex: test.TrimYAML(`
				^delete API token:
					SQL logic error: no such table: api_tokens`),
		},
		{
			name: "scan/corrupt created_at",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "argus", "")
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, created_at)
					VALUES ('x', '%s', 'x', 'hash', 'argus_', 'not-a-time');`, user.ID))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), mustUserID(t, store, "argus"))
				return err
			},
			errRegex: test.TrimYAML(`
				^scan API token:
					.*name "created_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "scan/corrupt expires_at",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "argus", "")
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, created_at, expires_at)
					VALUES ('x', '%s', 'x', 'hash', 'argus_', '%s', 'not-a-time');`,
					user.ID, timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), mustUserID(t, store, "argus"))
				return err
			},
			errRegex: test.TrimYAML(`
				^scan API token:
					.*name "expires_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "scan/corrupt last_used_at",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "argus", "")
				test.ExecSQL(t, store.db, fmt.Sprintf(`
					INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, created_at, last_used_at)
					VALUES ('x', '%s', 'x', 'hash', 'argus_', '%s', 'not-a-time');`,
					user.ID, timeNow().Format(timeFormat)))
			},
			invoke: func(t *testing.T, store *Store) error {
				_, err := store.APITokensForUser(t.Context(), mustUserID(t, store, "argus"))
				return err
			},
			errRegex: test.TrimYAML(`
				^scan API token:
					.*name "last_used_at":
						parse stored timestamp "not-a-time"`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps randRead.

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

			err := tc.invoke(t, store)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestStore_APIToken__rowsErr_and_rowsAffected(t *testing.T) {
	store, state := testFaultStore(t)
	user := mustCreateUser(t, store, "argus", "")
	_, token, err := store.CreateAPIToken(
		t.Context(),
		user.ID,
		"ci",
		nil,
	)
	if err != nil {
		t.Fatalf("Store.CreateAPIToken() failed: %v", err)
	}

	overrideRowsErrOnCall(t, 1)
	_, err = store.APITokensForUser(t.Context(), user.ID)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^iterate API tokens:
			injected fault$`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("Store.APITokensForUser() error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}

	state.SetRowsAffected(`DELETE FROM api_tokens WHERE id = ?`)
	err = store.DeleteAPIToken(t.Context(), user.ID, token.ID)
	state.Set("")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf(
			"Store.DeleteAPIToken() error mismatch\ngot:  %v\nwant: non-nil, NOT %v",
			err, ErrNotFound,
		)
	}
}

// mustUserID resolves a username to its ID.
func mustUserID(t *testing.T, store *Store, username string) string {
	t.Helper()

	return mustLocalCredentials(t, store, username).UserID
}
