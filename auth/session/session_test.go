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
	token, err := manager.Start(t.Context(), userID, ip, userAgent)
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

	// AND: only the token's hash is stored, with the configured bounds.
	if _, ok := fake.sessions[token]; ok {
		t.Errorf("%s the raw token must never be a storage key", prefix)
	}
	stored, ok := fake.sessions[hashToken(token)]
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
	other, err := manager.Start(t.Context(), userID, ip, userAgent)
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
			_, err := manager.Start(t.Context(), "user-1", "", "")

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

	// AND: a delete failure on drop is swallowed.
	token = mustStartFor(t, manager, "user-a")
	clock.advance(testConfig().IdleTimeout + time.Minute)
	fake.deleteErr = errors.New("delete broke")
	if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"%s expected ErrInvalidSession despite delete failure, got %v",
			prefix, err,
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
	if got := fake.sessions[hashToken(token)].LastSeenAt; !got.Equal(clock.now) {
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

	// THEN: unrevoked sessions still validate.
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

func TestManager_Revoke(t *testing.T) {
	// GIVEN: a live session.
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")

	prefix := fmt.Sprintf("%s\nManager.Revoke()", packageName)

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

	// AND: a store failure is surfaced.
	fake.deleteErr = errors.New("delete broke")
	if err := manager.Revoke(t.Context(), token); err == nil {
		t.Errorf("%s expected an error from a failing store", prefix)
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
	if _, ok := manager.cache[hashToken(expired)]; ok {
		t.Errorf("%s expired session should leave the cache", prefix)
	}
	if _, err := manager.Validate(t.Context(), live); err != nil {
		t.Errorf(
			"%s live session should survive\ngot: %v",
			prefix, err,
		)
	}

	// AND: a store failure is surfaced.
	fake.deleteErr = errors.New("delete broke")
	if err := manager.PruneExpired(t.Context()); err == nil {
		t.Errorf("%s expected an error from a failing store", prefix)
	}
}
