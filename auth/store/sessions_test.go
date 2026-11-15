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
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// testSession builds a Session for user with sensible times.
func testSession(tokenHash, userID string, expiresAt time.Time) Session {
	now := timeNow()
	return Session{
		TokenHash:  tokenHash,
		UserID:     userID,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		LastSeenAt: now,
		IP:         "127.0.0.1",
		UserAgent:  "go-test",
	}
}

func TestStore_Sessions__roundTrip(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	expiresAt := timeNow().Add(time.Hour)

	tokenHash := "token-hash-1"
	want := testSession(tokenHash, user.ID, expiresAt)
	if err := store.InsertSession(t.Context(), want); err != nil {
		t.Fatalf("Store.InsertSession() failed: %v", err)
	}
	got, err := store.SessionByTokenHash(t.Context(), tokenHash)
	if err != nil {
		t.Fatalf("after insert: Store.SessionByTokenHash() failed: %v", err)
	}

	if testErr := test.AssertFields(t, []test.FieldAssertion{
		{Name: "TokenHash", Got: got.TokenHash, Want: want.TokenHash, Mode: test.CompareEqual},
		{Name: "UserID", Got: got.UserID, Want: want.UserID, Mode: test.CompareEqual},
		{Name: "CreatedAt", Got: got.CreatedAt, Want: want.CreatedAt, Mode: test.CompareEqual},
		{Name: "ExpiresAt", Got: got.ExpiresAt, Want: want.ExpiresAt, Mode: test.CompareEqual},
		{Name: "LastSeenAt", Got: got.LastSeenAt, Want: want.LastSeenAt, Mode: test.CompareEqual},
		{Name: "IP", Got: got.IP, Want: want.IP, Mode: test.CompareEqual},
		{Name: "UserAgent", Got: got.UserAgent, Want: want.UserAgent, Mode: test.CompareEqual},
	}, "Session"); testErr != nil {
		t.Errorf("after insert: %v", testErr)
	}

	newSeen := timeNow().Add(30 * time.Minute)
	if err := store.TouchSession(t.Context(), tokenHash, newSeen); err != nil {
		t.Fatalf("Store.TouchSession() failed: %v", err)
	}
	got, err = store.SessionByTokenHash(t.Context(), tokenHash)
	if err != nil {
		t.Fatalf("after touch: Store.SessionByTokenHash() failed: %v", err)
	}
	if !got.LastSeenAt.Equal(newSeen) {
		t.Errorf(
			"after touch: Session.LastSeenAt mismatch\ngot:  %v\nwant: %v",
			got.LastSeenAt, newSeen,
		)
	}

	if err := store.DeleteSession(t.Context(), tokenHash); err != nil {
		t.Fatalf("Store.DeleteSession() failed: %v", err)
	}
	if _, err := store.SessionByTokenHash(t.Context(), tokenHash); !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"after delete: Store.SessionByTokenHash() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrNotFound,
		)
	}
}

func TestStore_InsertSession__error(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	session := testSession("duplicate", user.ID, timeNow().Add(time.Hour))
	if err := store.InsertSession(t.Context(), session); err != nil {
		t.Fatalf("Store.InsertSession() failed: %v", err)
	}

	err := store.InsertSession(t.Context(), session)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^insert session:
			constraint failed: UNIQUE constraint failed: sessions\.token_hash`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_SessionByTokenHash__errors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, store *Store)
		tokenHash string
		wantErr   error
		errRegex  string
	}{
		{
			name:      "unknown token",
			setup:     func(_ *testing.T, _ *Store) {},
			tokenHash: "never-issued",
			wantErr:   ErrNotFound,
			errRegex:  `^not found$`,
		},
		{
			name: "corrupt timestamps",
			setup: func(t *testing.T, store *Store) {
				user := mustCreateUser(t, store, "argus", "")
				if _, err := store.db.Exec(`
					INSERT INTO sessions (token_hash, user_id, created_at, expires_at, last_seen_at)
					VALUES ('corrupt', ?, 'not-a-time', 'not-a-time', 'not-a-time');`,
					user.ID,
				); err != nil {
					t.Fatalf("DB.Exec() failed: %v", err)
				}
			},
			tokenHash: "corrupt",
			errRegex: test.TrimYAML(`
				^load session:
					sql: Scan error on column index 2, name "created_at":
						parse stored timestamp "not-a-time"`),
		},
		{
			name: "sessions table missing",
			setup: func(t *testing.T, store *Store) {
				test.DropTable(t, store.db, "sessions")
			},
			tokenHash: "anything",
			errRegex:  `^load session`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			tc.setup(t, store)

			_, err := store.SessionByTokenHash(t.Context(), tc.tokenHash)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}
		})
	}
}

func TestStore_TouchSession__error(t *testing.T) {
	store := testStore(t)
	test.DropTable(t, store.db, "sessions")

	err := store.TouchSession(t.Context(), "any", timeNow())
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^touch session:
			SQL logic error: no such table: sessions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_DeleteSession__error(t *testing.T) {
	store := testStore(t)
	test.DropTable(t, store.db, "sessions")

	err := store.DeleteSession(t.Context(), "any")
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^delete session:
			SQL logic error: no such table: sessions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_DeleteSessionsForUser(t *testing.T) {
	store := testStore(t)
	alpha := mustCreateUser(t, store, "alpha", "")
	beta := mustCreateUser(t, store, "beta", "")
	expiresAt := timeNow().Add(time.Hour)
	for i, userID := range []string{alpha.ID, alpha.ID, beta.ID} {
		if err := store.InsertSession(
			t.Context(),
			testSession(fmt.Sprintf("hash-%d", i), userID, expiresAt),
		); err != nil {
			t.Fatalf("Store.InsertSession(hash-%d) failed: %v", i, err)
		}
	}

	err := store.DeleteSessionsForUser(t.Context(), alpha.ID)
	if err != nil {
		t.Fatalf("before DropTable: unexpected error: %v", err)
	}

	if got := test.CountRows(t, store.db, "sessions"); got != 1 {
		t.Errorf("before DropTable: sessions row count mismatch\ngot:  %d\nwant: 1", got)
	}
	if _, err := store.SessionByTokenHash(t.Context(), "hash-2"); err != nil {
		t.Errorf("beta: Store.SessionByTokenHash() failed: %v", err)
	}

	test.DropTable(t, store.db, "sessions")
	err = store.DeleteSessionsForUser(t.Context(), alpha.ID)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^delete user's sessions:
			SQL logic error: no such table: sessions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("missing table: error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_DeleteExpiredSessions(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	now := timeNow()
	for _, session := range []struct {
		hash      string
		expiresAt time.Time
	}{
		{hash: "expired-1", expiresAt: now.Add(-time.Hour)},
		{hash: "expired-2", expiresAt: now.Add(-time.Minute)},
		{hash: "live", expiresAt: now.Add(time.Hour)},
	} {
		if err := store.InsertSession(
			t.Context(),
			testSession(session.hash, user.ID, session.expiresAt),
		); err != nil {
			t.Fatalf("Store.InsertSession(%q) failed: %v", session.hash, err)
		}
	}

	err := store.DeleteExpiredSessions(t.Context(), now)
	if err != nil {
		t.Fatalf("before DropTable: unexpected error: %v", err)
	}

	if got := test.CountRows(t, store.db, "sessions"); got != 1 {
		t.Errorf("before DropTable: sessions row count mismatch\ngot:  %d\nwant: 1", got)
	}
	if _, err := store.SessionByTokenHash(t.Context(), "live"); err != nil {
		t.Errorf("live: Store.SessionByTokenHash() failed: %v", err)
	}

	test.DropTable(t, store.db, "sessions")
	err = store.DeleteExpiredSessions(t.Context(), now)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^delete expired sessions:
			SQL logic error: no such table: sessions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("missing table: error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestStore_TrimSessionsForUser(t *testing.T) {
	store := testStore(t)
	alpha := mustCreateUser(t, store, "alpha", "")
	beta := mustCreateUser(t, store, "beta", "")
	now := timeNow()
	insert := func(tokenHash, userID string, lastSeenAt time.Time) {
		t.Helper()
		session := testSession(tokenHash, userID, now.Add(time.Hour))
		session.LastSeenAt = lastSeenAt
		if err := store.InsertSession(t.Context(), session); err != nil {
			t.Fatalf("Store.InsertSession(%q) failed: %v", tokenHash, err)
		}
	}
	insert("beta-oldest", beta.ID, now.Add(-3*time.Hour))
	insert("alpha-old", alpha.ID, now.Add(-2*time.Hour))
	insert("alpha-mid", alpha.ID, now.Add(-time.Hour))
	insert("alpha-new", alpha.ID, now)

	evicted, err := store.TrimSessionsForUser(t.Context(), alpha.ID, 2)
	if err != nil {
		t.Fatalf("over cap: unexpected error: %v", err)
	}

	if want := []string{"alpha-old"}; !slices.Equal(evicted, want) {
		t.Errorf("over cap: evicted mismatch\ngot:  %q\nwant: %q", evicted, want)
	}

	// The cap is per-user, so beta-oldest survives.
	for _, tokenHash := range []string{"alpha-mid", "alpha-new", "beta-oldest"} {
		t.Run(tokenHash, func(t *testing.T) {
			// Serial: reads the sessions table, which the parent drops after the loop.

			if _, err := store.SessionByTokenHash(t.Context(), tokenHash); err != nil {
				t.Errorf("Store.SessionByTokenHash() failed: %v", err)
			}
		})
	}

	evicted, err = store.TrimSessionsForUser(t.Context(), alpha.ID, 2)
	if err != nil || evicted != nil {
		t.Errorf("at cap: evicted mismatch\ngot:  %q, err=%v\nwant: nil, err=nil", evicted, err)
	}

	test.DropTable(t, store.db, "sessions")
	_, err = store.TrimSessionsForUser(t.Context(), alpha.ID, 2)
	e := errfmt.FormatError(err)
	errRegex := test.TrimYAML(`
		^list sessions over cap:
			SQL logic error: no such table: sessions`)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("missing table: error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}
