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
	// GIVEN: a Store with a user.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	expiresAt := timeNow().Add(time.Hour)

	prefix := fmt.Sprintf("%s\nsession round-trip", packageName)

	// WHEN: a session is inserted and read back.
	want := testSession("token-hash-1", user.ID, expiresAt)
	if err := store.InsertSession(t.Context(), want); err != nil {
		t.Fatalf(
			"%s InsertSession failed: %v",
			prefix, err,
		)
	}
	got, err := store.SessionByTokenHash(t.Context(), "token-hash-1")
	if err != nil {
		t.Fatalf(
			"%s SessionByTokenHash failed: %v",
			prefix, err,
		)
	}

	// THEN: every field round-trips.
	if got.TokenHash != want.TokenHash ||
		got.UserID != want.UserID ||
		!got.CreatedAt.Equal(want.CreatedAt) ||
		!got.ExpiresAt.Equal(want.ExpiresAt) ||
		!got.LastSeenAt.Equal(want.LastSeenAt) ||
		got.IP != want.IP ||
		got.UserAgent != want.UserAgent {
		t.Errorf(
			"%s session mismatch\ngot:  %+v\nwant: %+v",
			prefix, *got, want,
		)
	}

	// AND: TouchSession moves last_seen_at.
	newSeen := timeNow().Add(30 * time.Minute)
	if err := store.TouchSession(t.Context(), "token-hash-1", newSeen); err != nil {
		t.Fatalf(
			"%s TouchSession failed: %v",
			prefix, err,
		)
	}
	got, err = store.SessionByTokenHash(t.Context(), "token-hash-1")
	if err != nil || !got.LastSeenAt.Equal(newSeen) {
		t.Errorf(
			"%s last_seen_at mismatch\ngot:  %v, err=%v\nwant: %v",
			prefix, got.LastSeenAt, err, newSeen,
		)
	}

	// AND: DeleteSession removes it.
	if err := store.DeleteSession(t.Context(), "token-hash-1"); err != nil {
		t.Fatalf(
			"%s DeleteSession failed: %v",
			prefix, err,
		)
	}
	if _, err := store.SessionByTokenHash(t.Context(), "token-hash-1"); !errors.Is(err, ErrNotFound) {
		t.Errorf(
			"%s session should be gone\ngot: %v",
			prefix, err,
		)
	}
}

func TestStore_InsertSession__error(t *testing.T) {
	// GIVEN: a Store with a session token hash.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	session := testSession("duplicate", user.ID, timeNow().Add(time.Hour))
	if err := store.InsertSession(t.Context(), session); err != nil {
		t.Fatalf(
			"%s\nsetup InsertSession failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nInsertSession(duplicate)", packageName)

	// WHEN: the same token hash is inserted again.
	err := store.InsertSession(t.Context(), session)

	// THEN: the constraint violation is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

func TestStore_SessionByTokenHash__errors(t *testing.T) {
	// GIVEN: stores in various broken states.
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
					t.Fatalf(
						"%s\nsetup failed: %v",
						packageName, err,
					)
				}
			},
			tokenHash: "corrupt",
			errRegex:  `^parse stored timestamp`,
		},
		{
			name: "sessions table missing",
			setup: func(t *testing.T, store *Store) {
				dropTable(t, store, "sessions")
			},
			tokenHash: "anything",
			errRegex:  `^load session`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := testStore(t)
			tc.setup(t, store)

			// WHEN: SessionByTokenHash is called.
			_, err := store.SessionByTokenHash(t.Context(), tc.tokenHash)

			prefix := fmt.Sprintf(
				"%s\nSessionByTokenHash(%q)",
				packageName, tc.tokenHash,
			)

			// THEN: the failure is surfaced.
			errRegex := util.ValueOr(tc.errRegex, `^$`)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf(
						"%s error mismatch\ngot:  %v\nwant: %v",
						prefix, err, tc.wantErr,
					)
				}
			} else {
				e := errfmt.FormatError(err)
				if !util.RegexCheck(errRegex, e) {
					t.Errorf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, e, errRegex,
					)
				}
			}
		})
	}
}

func TestStore_TouchSession__error(t *testing.T) {
	// GIVEN: a Store whose sessions table is unreadable.
	store := testStore(t)
	dropTable(t, store, "sessions")

	// WHEN: TouchSession is called.
	err := store.TouchSession(t.Context(), "any", timeNow())

	prefix := fmt.Sprintf("%s\nTouchSession() with unreadable sessions", packageName)

	// THEN: the failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

func TestStore_DeleteSession__error(t *testing.T) {
	// GIVEN: a Store whose sessions table is unreadable.
	store := testStore(t)
	dropTable(t, store, "sessions")

	prefix := fmt.Sprintf("%s\nDeleteSession() with unreadable sessions", packageName)

	// WHEN: DeleteSession is called.
	err := store.DeleteSession(t.Context(), "any")

	// THEN: the failure is surfaced.
	if err == nil {
		t.Errorf("%s expected an error, got nil", prefix)
	}
}

func TestStore_DeleteSessionsForUser(t *testing.T) {
	// GIVEN: two users with sessions.
	store := testStore(t)
	alpha := mustCreateUser(t, store, "alpha", "")
	beta := mustCreateUser(t, store, "beta", "")
	expiresAt := timeNow().Add(time.Hour)
	for i, userID := range []string{alpha.ID, alpha.ID, beta.ID} {
		if err := store.InsertSession(
			t.Context(),
			testSession(fmt.Sprintf("hash-%d", i), userID, expiresAt),
		); err != nil {
			t.Fatalf(
				"%s\nsetup InsertSession failed: %v",
				packageName, err,
			)
		}
	}

	prefix := fmt.Sprintf("%s\nDeleteSessionsForUser()", packageName)

	// WHEN: alpha's sessions are deleted.
	if err := store.DeleteSessionsForUser(t.Context(), alpha.ID); err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: only beta's session remains.
	if got := countRows(t, store, "sessions"); got != 1 {
		t.Errorf(
			"%s session count mismatch\ngot:  %d\nwant: 1",
			prefix, got,
		)
	}
	if _, err := store.SessionByTokenHash(t.Context(), "hash-2"); err != nil {
		t.Errorf(
			"%s beta's session should survive\ngot: %v",
			prefix, err,
		)
	}

	// AND: it errors when the table is unreadable.
	dropTable(t, store, "sessions")
	if err := store.DeleteSessionsForUser(t.Context(), alpha.ID); err == nil {
		t.Errorf("%s expected an error after dropping sessions", prefix)
	}
}

func TestStore_DeleteExpiredSessions(t *testing.T) {
	// GIVEN: sessions either side of the expiry cutoff.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	now := timeNow()
	for hash, expiresAt := range map[string]time.Time{
		"expired-1": now.Add(-time.Hour),
		"expired-2": now.Add(-time.Minute),
		"live":      now.Add(time.Hour),
	} {
		if err := store.InsertSession(
			t.Context(),
			testSession(hash, user.ID, expiresAt),
		); err != nil {
			t.Fatalf(
				"%s\nsetup InsertSession failed: %v",
				packageName, err,
			)
		}
	}

	prefix := fmt.Sprintf("%s\nDeleteExpiredSessions()", packageName)

	// WHEN: expired sessions are pruned.
	if err := store.DeleteExpiredSessions(t.Context(), now); err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: only the live session remains.
	if got := countRows(t, store, "sessions"); got != 1 {
		t.Errorf(
			"%s session count mismatch\ngot:  %d\nwant: 1",
			prefix, got,
		)
	}
	if _, err := store.SessionByTokenHash(t.Context(), "live"); err != nil {
		t.Errorf(
			"%s live session should survive\ngot: %v",
			prefix, err,
		)
	}

	// AND: it errors when the table is unreadable.
	dropTable(t, store, "sessions")
	if err := store.DeleteExpiredSessions(t.Context(), now); err == nil {
		t.Errorf("%s expected an error after dropping sessions", prefix)
	}
}

func TestStore_TrimSessionsForUser(t *testing.T) {
	// GIVEN: a user with sessions of varying activity,
	// and another user's session older than all of them.
	store := testStore(t)
	alpha := mustCreateUser(t, store, "alpha", "")
	beta := mustCreateUser(t, store, "beta", "")
	now := timeNow()
	insert := func(tokenHash, userID string, lastSeenAt time.Time) {
		t.Helper()
		session := testSession(tokenHash, userID, now.Add(time.Hour))
		session.LastSeenAt = lastSeenAt
		if err := store.InsertSession(t.Context(), session); err != nil {
			t.Fatalf(
				"%s\nsetup InsertSession failed: %v",
				packageName, err,
			)
		}
	}
	insert("beta-oldest", beta.ID, now.Add(-3*time.Hour))
	insert("alpha-old", alpha.ID, now.Add(-2*time.Hour))
	insert("alpha-mid", alpha.ID, now.Add(-time.Hour))
	insert("alpha-new", alpha.ID, now)

	prefix := fmt.Sprintf("%s\nTrimSessionsForUser()", packageName)

	// WHEN: alpha (3-sessions) is trimmed to their 2 most-recently active sessions.
	evicted, err := store.TrimSessionsForUser(t.Context(), alpha.ID, 2)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: only alpha's least-recently active session was evicted.
	if want := []string{"alpha-old"}; !slices.Equal(evicted, want) {
		t.Errorf(
			"%s evicted mismatch\ngot:  %v\nwant: %v",
			prefix, evicted, want,
		)
	}

	// AND: the survivors and beta's session (the cap is per-user) remain.
	for _, tokenHash := range []string{"alpha-mid", "alpha-new", "beta-oldest"} {
		if _, err := store.SessionByTokenHash(t.Context(), tokenHash); err != nil {
			t.Errorf(
				"%s %q should survive\ngot: %v",
				prefix, tokenHash, err,
			)
		}
	}

	// AND: a below-cap trim evicts nothing.
	evicted, err = store.TrimSessionsForUser(t.Context(), alpha.ID, 2)
	if err != nil || evicted != nil {
		t.Errorf(
			"%s below-cap trim should evict nothing\ngot:  %v, err=%v",
			prefix, evicted, err,
		)
	}

	// AND: it errors when the table is unreadable.
	dropTable(t, store, "sessions")
	if _, err := store.TrimSessionsForUser(t.Context(), alpha.ID, 2); err == nil {
		t.Errorf("%s expected an error after dropping sessions", prefix)
	}
}
