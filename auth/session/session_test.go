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

package session

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestManager_Start(t *testing.T) {
	// GIVEN: a Manager.
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())

	prefix := fmt.Sprintf("%s\nManager.Start()", packageName)

	// WHEN: a session is started.
	userID, ip, userAgent := "user-1", "127.0.0.1", "go-test"
	token, evicted, err := manager.Start(t.Context(), userID, ip, userAgent)
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: the token has the expected bits of entropy, hex-encoded.
	if len(token) != tokenBytes*2 {
		t.Errorf(
			"%s token length mismatch\ngot:  %d\nwant: %d",
			prefix, len(token), tokenBytes*2,
		)
	}

	// AND: nothing was evicted below the cap.
	if evicted != nil {
		t.Errorf(
			"%s below-cap start should evict nothing\ngot:  %v",
			prefix, evicted,
		)
	}

	// AND: gen stayed put.
	if got := genOf(t, manager); got != 0 {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: 0",
			prefix, got,
		)
	}

	// AND: only the token's hash is stored, with the configured bounds.
	if _, ok := fake.sessions[token]; ok {
		t.Errorf("%s the raw token must never be a storage key", prefix)
	}
	stored, ok := fake.sessions[HashToken(token)]
	if !ok {
		t.Fatalf("%s session not persisted under the token's hash", prefix)
	}
	if stored.UserID != userID ||
		stored.IP != ip ||
		stored.UserAgent != userAgent ||
		!stored.CreatedAt.Equal(clock.now) ||
		!stored.ExpiresAt.Equal(clock.now.Add(testConfig().Lifetime)) ||
		!stored.LastSeenAt.Equal(clock.now) {
		t.Errorf(
			"%s stored session mismatch\ngot: %+v",
			prefix, stored,
		)
	}

	// AND: subsequent sessions don't share a token.
	other, _, err := manager.Start(t.Context(), userID, ip, userAgent)
	if err != nil || other == token {
		t.Errorf(
			"%s tokens must be unique\ngot:  %q twice, err=%v",
			prefix, token, err,
		)
	}
}

func TestManager_Start__errors(t *testing.T) {
	// GIVEN: failing dependencies.
	tests := []struct {
		name      string
		randErr   bool
		insertErr bool
	}{
		{name: "random source fails", randErr: true},
		{name: "store insert fails", insertErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeStore()

			if tc.insertErr {
				fake.insertErr = errors.New("insert broke")
			}
			if tc.randErr {
				randReadHad := randRead
				randRead = func(_ []byte) (int, error) {
					return 0, errors.New("rand broke")
				}
				t.Cleanup(func() { randRead = randReadHad })
			}

			manager := New(fake, testConfig())

			// WHEN: a session is started.
			_, _, err := manager.Start(t.Context(), "user-1", "", "")

			prefix := fmt.Sprintf(
				"%s\nManager.Start()",
				packageName,
			)

			// THEN: the failure is surfaced.
			if err == nil {
				t.Errorf("%s expected an error, got nil", prefix)
			}
		})
	}
}

func TestManager_Start__capEvictsOldest(t *testing.T) {
	// GIVEN: user-b's old session, and user-a already at the session cap.
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	tokenOtherUser := mustStartFor(t, manager, "user-b")
	tokens := make([]string, MaxSessionsPerUser)
	for i := range tokens {
		clock.advance(time.Minute)
		tokens[i] = mustStartFor(t, manager, "user-a")
	}

	prefix := fmt.Sprintf("%s\nManager.Start() at the cap", packageName)
	genBefore := genOf(t, manager)

	// WHEN: one more session is started for the capped user-a.
	clock.advance(time.Minute)
	token, evicted, err := manager.Start(t.Context(), "user-a", "127.0.0.1", "go-test")
	if err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: only their least-recently active session was evicted.
	want := []string{HashToken(tokens[0])}
	if !slices.Equal(evicted, want) {
		t.Fatalf(
			"%s evicted mismatch\ngot:  %v\nwant: %v",
			prefix, evicted, want,
		)
	}

	// AND: gen was bumped once for the eviction, however many it covered.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: the evicted token no longer validates.
	if _, err := manager.Validate(t.Context(), tokens[0]); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s evicted session should be invalid\ngot: %v",
			prefix, err,
		)
	}

	// AND: the remaining and new sessions do, as does user-b's
	// (older than all of user-a's - the cap is per-user).
	for _, token := range append(append([]string{}, tokens[1:]...), token, tokenOtherUser) {
		if _, err := manager.Validate(t.Context(), token); err != nil {
			t.Errorf(
				"%s surviving session should stay valid\ngot: %v",
				prefix, err,
			)
		}
	}

	// AND: a trim failure does not fail the login.
	fake.trimErr = errors.New("trim broke")
	genBefore = genOf(t, manager)
	if _, evicted, err := manager.Start(
		t.Context(),
		"user-a",
		"127.0.0.1",
		"go-test",
	); err != nil || evicted != nil {
		t.Errorf(
			"%s trim failure should not fail Start\ngot:  evicted=%v, err=%v",
			prefix, evicted, err,
		)
	}

	// AND: evicts nothing to bump gen for.
	if got := genOf(t, manager); got != genBefore {
		t.Errorf(
			"%s gen mismatch after a failed trim\ngot:  %d\nwant: %d",
			prefix, got, genBefore,
		)
	}
}

func TestManager_Expired(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	config := testConfig()

	// GIVEN: sessions at various points in their lifetime
	// (all offsets are from base).
	tests := []struct {
		name       string
		expiresAt  time.Duration
		lastSeenAt time.Duration
		now        time.Duration
		want       bool
	}{
		{
			name:       "live/well within both windows",
			expiresAt:  config.Lifetime,
			lastSeenAt: 0,
			now:        time.Hour,
			want:       false,
		},
		{
			name:       "live/exactly at the absolute expiry",
			expiresAt:  config.Lifetime,
			lastSeenAt: config.Lifetime,
			now:        config.Lifetime,
			want:       false,
		},
		{
			name:       "expired/a moment past the absolute expiry",
			expiresAt:  config.Lifetime,
			lastSeenAt: config.Lifetime,
			now:        config.Lifetime + time.Nanosecond,
			want:       true,
		},
		{
			name:      "live/exactly at the idle deadline",
			expiresAt: config.Lifetime,
			now:       config.IdleTimeout,
			want:      false,
		},
		{
			name:      "expired/a moment past the idle deadline",
			expiresAt: config.Lifetime,
			now:       config.IdleTimeout + time.Nanosecond,
			want:      true,
		},
		{
			name:      "expired/past both windows",
			expiresAt: time.Hour,
			now:       2 * config.Lifetime,
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			manager := New(newFakeStore(), config)
			session := store.Session{
				ExpiresAt:  base.Add(tc.expiresAt),
				LastSeenAt: base.Add(tc.lastSeenAt),
			}

			// WHEN: the session is checked for expiry.
			got := manager.expired(session, base.Add(tc.now))

			prefix := fmt.Sprintf(
				"%s Manager.expired(expiresAt=%s, lastSeenAt=%s, now=%s)",
				packageName, tc.expiresAt, tc.lastSeenAt, tc.now,
			)

			// THEN: it reads as expected.
			if got != tc.want {
				t.Errorf(
					"%s expiry mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestManager_Evict(t *testing.T) {
	// GIVEN: two live sessions.
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")
	other := mustStartFor(t, manager, "user-b")
	tokenHash, otherHash := HashToken(token), HashToken(other)
	genBefore := genOf(t, manager)

	prefix := fmt.Sprintf("%s\nManager.evict()", packageName)

	// WHEN: one of them is evicted.
	err := manager.evict(t.Context(), tokenHash)

	// THEN: the caller is told the session is invalid.
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrInvalidSession,
		)
	}

	// AND: it left both the cache and the store.
	if _, ok := manager.cache[tokenHash]; ok {
		t.Errorf("%s evicted session should leave the cache", prefix)
	}
	if _, ok := fake.sessions[tokenHash]; ok {
		t.Errorf("%s evicted session should leave the store", prefix)
	}

	// AND: gen was bumped.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: the other session is untouched.
	if _, err := manager.Validate(t.Context(), other); err != nil {
		t.Errorf(
			"%s bystander session should stay valid\ngot: %v",
			prefix, err,
		)
	}

	// AND: a store delete failure is swallowed, and still drops the cache entry.
	fake.deleteErr = errors.New("delete broke")
	if err := manager.evict(t.Context(), otherHash); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s error mismatch despite delete failure\ngot:  %v\nwant: %v",
			prefix, err, ErrInvalidSession,
		)
	}
	if _, ok := manager.cache[otherHash]; ok {
		t.Errorf("%s cache entry should go even when the store delete fails", prefix)
	}

	// AND: evicting an unknown hash does nothing beyond bumping gen.
	fake.deleteErr = nil
	genBefore = genOf(t, manager)
	if err := manager.evict(t.Context(), "never-issued"); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s error mismatch for an unknown hash\ngot:  %v\nwant: %v",
			prefix, err, ErrInvalidSession,
		)
	}
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch for an unknown hash\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestManager_Validate(t *testing.T) {
	userID := "user-a"

	// GIVEN: sessions in various lifecycle states.
	config := testConfig()

	tests := []struct {
		name      string
		token     func(t *testing.T, m *Manager, clock *fakeClock) string
		fromStore bool // Empty the cache so validation hits the store.
		loadErr   bool
		wantErr   error
		errRegex  string
	}{
		{
			name: "valid session/cache hit",
			token: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return mustStartFor(t, m, userID)
			},
		},
		{
			name: "valid session/store hit after restart",
			token: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return mustStartFor(t, m, userID)
			},
			fromStore: true,
		},
		{
			name:    "empty token",
			token:   func(_ *testing.T, _ *Manager, _ *fakeClock) string { return "" },
			wantErr: ErrInvalidSession,
		},
		{
			name: "unknown token",
			token: func(_ *testing.T, _ *Manager, _ *fakeClock) string {
				return "never-issued"
			},
			wantErr: ErrInvalidSession,
		},
		{
			name: "absolutely expired session",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.Lifetime + time.Minute)
				return token
			},
			wantErr: ErrInvalidSession,
		},
		{
			name: "idle-expired session",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.IdleTimeout + time.Minute)
				return token
			},
			wantErr: ErrInvalidSession,
		},
		{
			name: "activity slides the idle window",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				// Halfway to idle expiry, then activity, then halfway again:
				// still valid only if the window slid.
				clock.advance(config.IdleTimeout / 2)
				if _, err := m.Validate(t.Context(), token); err != nil {
					t.Fatalf(
						"%s\nmid-window validate failed: %v",
						packageName, err,
					)
				}
				clock.advance(time.Minute + config.IdleTimeout/2)
				return token
			},
		},
		{
			name: "store load failure is not an auth failure",
			token: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return mustStartFor(t, m, userID)
			},
			fromStore: true,
			loadErr:   true,
			errRegex:  `^load session`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
			clock.use(t)
			fake := newFakeStore()
			manager := New(fake, config)

			token := tc.token(t, manager, clock)
			if tc.fromStore {
				manager.cache = make(map[string]store.Session)
			}
			if tc.loadErr {
				fake.loadErr = errors.New("load broke")
			}

			// WHEN: the token is validated.
			session, err := manager.Validate(t.Context(), token)

			prefix := fmt.Sprintf(
				"%s Manager.Validate(%q)",
				packageName, token,
			)

			// THEN: errors match expectations.
			errRegex := util.ValueOr(tc.errRegex, `^$`)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf(
						"%s error mismatch\ngot:  %v\nwant: %v",
						prefix, err, tc.wantErr,
					)
				}
				return
			}
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: the session belongs to the user who started it.
			if got, want := session.UserID, userID; got != want {
				t.Errorf(
					"%s user mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}

func TestManager_Validate__expiryDeletes(t *testing.T) {
	// GIVEN: an idle-expired session.
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")
	clock.advance(testConfig().IdleTimeout + time.Minute)

	prefix := fmt.Sprintf("%s\nManager.Validate() expiry cleanup", packageName)
	genBefore := genOf(t, manager)

	// WHEN: the expired token is validated.
	if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf(
			"%s expected ErrInvalidSession, got %v",
			prefix, err,
		)
	}

	// THEN: the session row was deleted on detection.
	if len(fake.sessions) != 0 || len(manager.cache) != 0 {
		t.Errorf(
			"%s expired session should be dropped\ngot:  store=%d cache=%d",
			prefix, len(fake.sessions), len(manager.cache),
		)
	}

	// AND: the eviction bumped gen.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: a delete failure on drop is swallowed.
	token = mustStartFor(t, manager, "user-a")
	clock.advance(testConfig().IdleTimeout + time.Minute)
	fake.deleteErr = errors.New("delete broke")
	genBefore = genOf(t, manager)
	if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s expected ErrInvalidSession despite delete failure, got %v",
			prefix, err,
		)
	}

	// AND: still bumped gen.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch after a failed delete\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestManager_Validate__touchFlush(t *testing.T) {
	// GIVEN: a valid session and a clock.
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")

	prefix := fmt.Sprintf("%s\nManager.Validate() touch flushing", packageName)
	genBefore := genOf(t, manager)

	// WHEN: the session is validated within the touch interval.
	clock.advance(touchInterval / 2)
	if _, err := manager.Validate(t.Context(), token); err != nil {
		t.Fatalf(
			"%s validate failed: %v",
			prefix, err,
		)
	}

	// THEN: the database was not touched (cache only).
	if fake.touches != 0 {
		t.Errorf(
			"%s no flush expected within touchInterval\ngot:  %d touches",
			prefix, fake.touches,
		)
	}

	// WHEN: validated again past the touch interval.
	clock.advance(touchInterval)
	if _, err := manager.Validate(t.Context(), token); err != nil {
		t.Fatalf(
			"%s validate failed: %v",
			prefix, err,
		)
	}

	// THEN: the last-seen time was flushed once.
	if fake.touches != 1 {
		t.Errorf(
			"%s flush count mismatch\ngot:  %d\nwant: 1",
			prefix, fake.touches,
		)
	}
	if got := fake.sessions[HashToken(token)].LastSeenAt; !got.Equal(clock.now) {
		t.Errorf(
			"%s flushed last-seen mismatch\ngot:  %v\nwant: %v",
			prefix, got, clock.now,
		)
	}

	// AND: a flush failure does not fail validation.
	fake.touchErr = errors.New("touch broke")
	clock.advance(touchInterval * 2)
	if _, err := manager.Validate(t.Context(), token); err != nil {
		t.Errorf(
			"%s validation should survive a flush failure\ngot: %v",
			prefix, err,
		)
	}

	// AND: sliding the window never bumps gen.
	if got := genOf(t, manager); got != genBefore {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, genBefore,
		)
	}
}

func TestManager_Validate__concurrent(t *testing.T) {
	// GIVEN: live sessions for several users.
	fake := newFakeStore()
	manager := New(fake, testConfig())
	tokens := make([]string, 8)
	for i := range tokens {
		tokens[i] = mustStartFor(t, manager, fmt.Sprintf("user-%d", i))
	}
	revoked := tokens[len(tokens)-1]

	prefix := fmt.Sprintf("%s\nManager.Validate() concurrently", packageName)
	genBefore := genOf(t, manager)

	// WHEN: the sessions are validated concurrently while one is revoked.
	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, token := range tokens {
		wg.Go(func() {
			<-start
			for range 25 {
				_, _ = manager.Validate(t.Context(), token)
			}
		})
	}
	wg.Go(func() {
		<-start
		_ = manager.Revoke(t.Context(), revoked)
	})
	close(start)
	wg.Wait()

	// THEN: gen moved exactly once.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: unrevoked sessions still validate.
	if _, err := manager.Validate(t.Context(), tokens[0]); err != nil {
		t.Errorf(
			"%s live session should stay valid\ngot: %v",
			prefix, err,
		)
	}

	// AND: the revoked one does not.
	if _, err := manager.Validate(t.Context(), revoked); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s revoked session should be invalid\ngot: %v",
			prefix, err,
		)
	}
}

func TestManager_Validate__tamperedToken(t *testing.T) {
	// GIVEN: a live session and a tampered copy of its token.
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")
	// Flip the final character.
	last := token[len(token)-1]
	flipped := byte('0')
	if last == '0' {
		flipped = '1'
	}
	tampered := token[:len(token)-1] + string(flipped)

	// WHEN: the tampered token is validated.
	_, err := manager.Validate(t.Context(), tampered)

	prefix := fmt.Sprintf("%s\nManager.Validate() tampered token", packageName)

	// THEN: it is rejected as an invalid session.
	if !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrInvalidSession,
		)
	}

	// AND: the genuine token still validates.
	if _, err := manager.Validate(t.Context(), token); err != nil {
		t.Errorf(
			"%s genuine token should stay valid\ngot: %v",
			prefix, err,
		)
	}
}

func TestManager_Alive(t *testing.T) {
	userID := "user-a"

	// GIVEN: sessions in various lifecycle states.
	config := testConfig()

	tests := []struct {
		name      string
		tokenHash func(t *testing.T, m *Manager, clock *fakeClock) string
		fromStore bool // Empty the cache so the check hits the store.
		loadErr   bool
		want      bool
	}{
		{
			name: "live/cache hit",
			tokenHash: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return HashToken(mustStartFor(t, m, userID))
			},
			want: true,
		},
		{
			name: "live/store hit after restart",
			tokenHash: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return HashToken(mustStartFor(t, m, userID))
			},
			fromStore: true,
			want:      true,
		},
		{
			name: "revoked session",
			tokenHash: func(t *testing.T, m *Manager, _ *fakeClock) string {
				token := mustStartFor(t, m, userID)
				if err := m.Revoke(t.Context(), token); err != nil {
					t.Fatalf(
						"%s\nsetup Revoke failed: %v",
						packageName, err,
					)
				}
				return HashToken(token)
			},
			want: false,
		},
		{
			name: "absolutely expired session",
			tokenHash: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.Lifetime + time.Minute)
				return HashToken(token)
			},
			want: false,
		},
		{
			name: "idle-expired session",
			tokenHash: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.IdleTimeout + time.Minute)
				return HashToken(token)
			},
			want: false,
		},
		{
			name: "check does not slide the idle window",
			tokenHash: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				// Halfway to idle expiry, a check, then past the rest of the
				// window: dead only if the check did not slide it.
				clock.advance(config.IdleTimeout / 2)
				if !m.Alive(t.Context(), HashToken(token)) {
					t.Fatalf(
						"%s\nmid-window check should read alive",
						packageName,
					)
				}
				clock.advance(time.Minute + config.IdleTimeout/2)
				return HashToken(token)
			},
			want: false,
		},
		{
			name: "unknown hash",
			tokenHash: func(_ *testing.T, _ *Manager, _ *fakeClock) string {
				return "never-issued"
			},
			want: false,
		},
		{
			name: "store failure reads as alive",
			tokenHash: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return HashToken(mustStartFor(t, m, userID))
			},
			fromStore: true,
			loadErr:   true,
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

			clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
			clock.use(t)
			fake := newFakeStore()
			manager := New(fake, config)

			tokenHash := tc.tokenHash(t, manager, clock)
			if tc.fromStore {
				manager.cache = make(map[string]store.Session)
			}
			if tc.loadErr {
				fake.loadErr = errors.New("load broke")
			}

			// WHEN: the session's liveness is checked.
			got := manager.Alive(t.Context(), tokenHash)

			prefix := fmt.Sprintf(
				"%s Manager.Alive(%q)",
				packageName, tokenHash,
			)

			// THEN: it reads as expected.
			if got != tc.want {
				t.Errorf(
					"%s liveness mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestManager_Revoke(t *testing.T) {
	// GIVEN: a live session.
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")

	prefix := fmt.Sprintf("%s\nManager.Revoke()", packageName)
	genBefore := genOf(t, manager)

	// WHEN: it is revoked.
	if err := manager.Revoke(t.Context(), token); err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: the token no longer validates.
	if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s revoked token should be invalid\ngot: %v",
			prefix, err,
		)
	}

	// AND: gen was bumped.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: a store failure is surfaced,.
	fake.deleteErr = errors.New("delete broke")
	genBefore = genOf(t, manager)
	if err := manager.Revoke(t.Context(), token); err == nil {
		t.Errorf("%s expected an error from a failing store", prefix)
	}
	// AND: gen is bumped.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch after a failed delete\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestManager_RevokeUser(t *testing.T) {
	// GIVEN: sessions of two users.
	fake := newFakeStore()
	manager := New(fake, testConfig())
	tokenA1 := mustStartFor(t, manager, "user-a")
	tokenA2 := mustStartFor(t, manager, "user-a")
	tokenB := mustStartFor(t, manager, "user-b")

	prefix := fmt.Sprintf("%s\nManager.RevokeUser()", packageName)
	genBefore := genOf(t, manager)

	// WHEN: user-a's sessions are revoked.
	if err := manager.RevokeUser(t.Context(), "user-a"); err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: user-a's tokens are removed.
	for _, token := range []string{tokenA1, tokenA2} {
		if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
			t.Errorf(
				"%s user-a token should be invalid\ngot: %v",
				prefix, err,
			)
		}
	}

	// AND: user-b's survive.
	if _, err := manager.Validate(t.Context(), tokenB); err != nil {
		t.Errorf(
			"%s user-b token should survive\ngot: %v",
			prefix, err,
		)
	}

	// AND: gen was bumped once for the sweep, not once per session.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: a store failure is surfaced.
	fake.deleteErr = errors.New("delete broke")
	if err := manager.RevokeUser(t.Context(), "user-b"); err == nil {
		t.Errorf("%s expected an error from a failing store", prefix)
	}
}

func TestManager_PruneExpired(t *testing.T) {
	// GIVEN: one live and one absolutely-expired session.
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	expired := mustStartFor(t, manager, "user-a")
	clock.advance(testConfig().Lifetime + time.Minute)
	live := mustStartFor(t, manager, "user-a")

	prefix := fmt.Sprintf("%s\nManager.PruneExpired()", packageName)
	genBefore := genOf(t, manager)

	// WHEN: expired sessions are pruned.
	if err := manager.PruneExpired(t.Context()); err != nil {
		t.Fatalf(
			"%s unexpected error: %v",
			prefix, err,
		)
	}

	// THEN: only the live session remains, in cache and store.
	if len(fake.sessions) != 1 || len(manager.cache) != 1 {
		t.Errorf(
			"%s session counts mismatch\ngot:  store=%d cache=%d\nwant: 1/1",
			prefix, len(fake.sessions), len(manager.cache),
		)
	}
	if _, ok := manager.cache[HashToken(expired)]; ok {
		t.Errorf("%s expired session should leave the cache", prefix)
	}
	if _, err := manager.Validate(t.Context(), live); err != nil {
		t.Errorf(
			"%s live session should survive\ngot: %v",
			prefix, err,
		)
	}

	// AND: gen was bumped once for the sweep, not once per session.
	if got, want := genOf(t, manager), genBefore+1; got != want {
		t.Errorf(
			"%s gen mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: a store failure is surfaced.
	fake.deleteErr = errors.New("delete broke")
	if err := manager.PruneExpired(t.Context()); err == nil {
		t.Errorf("%s expected an error from a failing store", prefix)
	}
}
