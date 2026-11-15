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

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestManager_Start(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())

	userID, ip, userAgent := "user-1", "127.0.0.1", "go-test"
	token, evicted, err := manager.Start(t.Context(), userID, ip, userAgent)
	if err != nil {
		t.Fatalf("first: unexpected error: %v", err)
	}

	// Hex-encoded, so twice tokenBytes long.
	if got, want := len(token), tokenBytes*2; got != want {
		t.Errorf("first: len(token) mismatch\ngot:  %d\nwant: %d", got, want)
	}

	if evicted != nil {
		t.Errorf("first: evicted mismatch\ngot:  %q\nwant: nil", evicted)
	}

	// Start bumps gen only when it evicts.
	if got := genOf(manager); got != 0 {
		t.Errorf("first: Manager.gen mismatch\ngot:  %d\nwant: 0", got)
	}

	if _, ok := fake.sessions[token]; ok {
		t.Error("first: fakeStore.sessions[token] mismatch\ngot:  present\nwant: absent")
	}
	stored, ok := fake.sessions[auth.HashToken(token)]
	if !ok {
		t.Fatal("first: fakeStore.sessions[HashToken(token)] mismatch\ngot:  absent\nwant: present")
	}
	wantExpiresAt := clock.now.Add(testConfig().Lifetime)
	if testErr := test.AssertFields(t, []test.FieldAssertion{
		{Name: "UserID", Got: stored.UserID, Want: userID, Mode: test.CompareEqual},
		{Name: "IP", Got: stored.IP, Want: ip, Mode: test.CompareEqual},
		{Name: "UserAgent", Got: stored.UserAgent, Want: userAgent, Mode: test.CompareEqual},
		{Name: "CreatedAt", Mode: test.CompareEqual, Got: stored.CreatedAt, Want: clock.now},
		{Name: "ExpiresAt", Mode: test.CompareEqual, Got: stored.ExpiresAt, Want: wantExpiresAt},
		{Name: "LastSeenAt", Mode: test.CompareEqual, Got: stored.LastSeenAt, Want: clock.now},
	}, "Session"); testErr != nil {
		t.Errorf("first: %v", testErr)
	}

	other, _, err := manager.Start(t.Context(), userID, ip, userAgent)
	if err != nil {
		t.Fatalf("second: Manager.Start() failed: %v", err)
	}
	if other == token {
		t.Errorf("second: token mismatch\ngot:  %q\nwant: NOT %q", other, token)
	}
}

func TestManager_Start__errors(t *testing.T) {
	tests := []struct {
		name      string
		randErr   bool
		insertErr bool
		errRegex  string
	}{
		{name: "random source fails", randErr: true, errRegex: `^generate session token`},
		{name: "store insert fails", insertErr: true, errRegex: `^persist session`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps randRead.

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

			_, _, err := manager.Start(t.Context(), "user-1", "", "")
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestManager_Start__capEvictsOldest(t *testing.T) {
	userA, userB := "user-a", "user-b"
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	tokenOtherUser := mustStartFor(t, manager, userB)
	tokensA := make([]string, DefaultMaxSessionsPerUser)
	for i := range DefaultMaxSessionsPerUser {
		clock.advance(time.Minute)
		tokensA[i] = mustStartFor(t, manager, userA)
	}

	genBefore := genOf(manager)

	clock.advance(time.Minute)
	ip, userAgent := "127.0.0.1", "go-test"
	token, evicted, err := manager.Start(t.Context(), userA, ip, userAgent)
	if err != nil {
		t.Fatalf("trim success: unexpected error: %v", err)
	}

	want := []string{auth.HashToken(tokensA[0])}
	if !slices.Equal(evicted, want) {
		t.Fatalf("trim success: evicted mismatch\ngot:  %q\nwant: %q", evicted, want)
	}

	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("trim success: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	if _, err := manager.Validate(t.Context(), tokensA[0]); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"evicted: Manager.Validate() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInvalidSession,
		)
	}

	// user-b's older session survives too, so the cap is per-user.
	type survivor struct{ name, token string }
	survivors := []survivor{
		{name: "token", token: token},
		{name: "tokenOtherUser", token: tokenOtherUser},
	}
	for i := 1; i < len(tokensA); i++ {
		survivors = append(survivors, survivor{fmt.Sprintf("tokensA[%d]", i), tokensA[i]})
	}
	for _, s := range survivors {
		t.Run(s.name, func(t *testing.T) {
			// Serial: reads the fake store, whose trim the parent breaks after the loop.

			if _, err := manager.Validate(t.Context(), s.token); err != nil {
				t.Errorf("Manager.Validate() failed: %v", err)
			}
		})
	}

	fake.trimErr = errors.New("trim broke")
	genBefore = genOf(manager)
	_, evicted, err = manager.Start(t.Context(), userA, ip, userAgent)
	if err != nil || evicted != nil {
		t.Errorf(
			"trim failure: Manager.Start() mismatch\n"+
				"got:  evicted=%q, err=%v\nwant: evicted=nil, err=nil",
			evicted, err,
		)
	}

	if got := genOf(manager); got != genBefore {
		t.Errorf("trim failure: Manager.gen mismatch\ngot:  %d\nwant: %d", got, genBefore)
	}
}

func TestManager_Start__configuredCapOverridesDefault(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	cfg := testConfig()
	cfg.MaxPerUser = 2
	manager := New(newFakeStore(), cfg)
	userID, ip, userAgent := "user-a", "127.0.0.1", "go-test"

	first := mustStartFor(t, manager, userID)
	clock.advance(time.Minute)
	second := mustStartFor(t, manager, userID)

	clock.advance(time.Minute)
	third, evicted, err := manager.Start(t.Context(), userID, ip, userAgent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := []string{auth.HashToken(first)}; !slices.Equal(evicted, want) {
		t.Fatalf("evicted mismatch\ngot:  %q\nwant: %q", evicted, want)
	}

	if _, err := manager.Validate(t.Context(), first); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"evicted: Manager.Validate() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInvalidSession,
		)
	}

	for _, s := range []struct{ name, token string }{
		{"second", second},
		{"third", third},
	} {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()

			if _, err := manager.Validate(t.Context(), s.token); err != nil {
				t.Errorf("Manager.Validate() failed: %v", err)
			}
		})
	}
}

func TestManager_Start__revokeDuringInsertDoesNotResurrect(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())

	userID := "user-a"
	// Revokes mid-Start, between the session insert and the cache write.
	revoked := false
	fake.onTrim = func() {
		if revoked {
			return
		}
		revoked = true
		if err := manager.RevokeUser(t.Context(), userID); err != nil {
			t.Errorf("Manager.RevokeUser() failed: %v", err)
		}
	}

	token := mustStartFor(t, manager, userID)

	_, err := manager.Validate(t.Context(), token)
	if !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"Manager.Validate() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInvalidSession,
		)
	}

	_, err = fake.SessionByTokenHash(t.Context(), auth.HashToken(token))
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(
			"fakeStore.SessionByTokenHash() error type mismatch\ngot:  %v\nwant: %v",
			err, store.ErrNotFound,
		)
	}
}

func TestManager_Expired(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	config := testConfig()

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

			got := manager.expired(session, base.Add(tc.now))

			if got != tc.want {
				t.Errorf("Manager.expired() mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestManager_Evict(t *testing.T) {
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")
	otherToken := mustStartFor(t, manager, "user-b")
	tokenHash, otherTokenHash := auth.HashToken(token), auth.HashToken(otherToken)
	genBefore := genOf(manager)

	manager.evict(t.Context(), tokenHash)

	if _, ok := manager.cache[tokenHash]; ok {
		t.Error(
			"delete success: Manager.cache[tokenHash] mismatch\n" +
				"got:  present\nwant: absent (evicted)",
		)
	}
	if _, ok := fake.sessions[tokenHash]; ok {
		t.Error(
			"delete success: fakeStore.sessions[tokenHash] mismatch\n" +
				"got:  present\nwant: absent (evicted)",
		)
	}

	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("delete success: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	if _, err := manager.Validate(t.Context(), otherToken); err != nil {
		t.Errorf("bystander: Manager.Validate() failed: %v", err)
	}

	fake.deleteErr = errors.New("delete broke")
	manager.evict(t.Context(), otherTokenHash)
	if _, ok := manager.cache[otherTokenHash]; ok {
		t.Error(
			"delete failure: Manager.cache[otherTokenHash] mismatch\n" +
				"got:  present\nwant: absent (evicted)",
		)
	}

	fake.deleteErr = nil
	genBefore = genOf(manager)
	manager.evict(t.Context(), "never-issued")
	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("unknown hash: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestManager_Validate(t *testing.T) {
	userID := "user-a"

	config := testConfig()

	tests := []struct {
		name        string
		token       func(t *testing.T, m *Manager, clock *fakeClock) string
		fromStore   bool // Empty the cache so validation hits the store.
		loadErr     bool
		wantTouches int // Expected last-seen flushes to the store.
		wantErr     error
		errRegex    string
	}{
		{
			name: "valid session/cache hit",
			token: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return mustStartFor(t, m, userID)
			},
			errRegex: `^$`,
		},
		{
			name: "valid session/store hit after restart",
			token: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return mustStartFor(t, m, userID)
			},
			fromStore: true,
			errRegex:  `^$`,
		},
		{
			name:     "empty token",
			token:    func(_ *testing.T, _ *Manager, _ *fakeClock) string { return "" },
			wantErr:  ErrInvalidSession,
			errRegex: `^invalid or expired session$`,
		},
		{
			name: "unknown token",
			token: func(_ *testing.T, _ *Manager, _ *fakeClock) string {
				return "never-issued"
			},
			wantErr:  ErrInvalidSession,
			errRegex: `^invalid or expired session$`,
		},
		{
			name: "absolutely expired session",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.Lifetime + time.Minute)
				return token
			},
			wantErr:  ErrInvalidSession,
			errRegex: `^invalid or expired session$`,
		},
		{
			name: "idle-expired session/cache hit",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.IdleTimeout + time.Minute)
				return token
			},
			wantErr:  ErrInvalidSession,
			errRegex: `^invalid or expired session$`,
		},
		{
			name: "idle-expired session/store hit",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.IdleTimeout + time.Minute)
				return token
			},
			fromStore: true,
			wantErr:   ErrInvalidSession,
			errRegex:  `^invalid or expired session$`,
		},
		{
			name: "stale last-seen/store hit flushes it",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(touchInterval + time.Minute)
				return token
			},
			fromStore:   true,
			wantTouches: 1,
			errRegex:    `^$`,
		},
		{
			name: "activity slides the idle window",
			token: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				// Halfway to idle expiry, activity, then halfway again, so the session
				// is still valid only if the window slid.
				clock.advance(config.IdleTimeout / 2)
				if _, err := m.Validate(t.Context(), token); err != nil {
					t.Fatalf("Manager.Validate() failed: %v", err)
				}
				clock.advance(time.Minute + config.IdleTimeout/2)
				return token
			},
			wantTouches: 2,
			errRegex:    `^$`,
		},
		{
			name: "store load failure is not an auth failure",
			token: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return mustStartFor(t, m, userID)
			},
			fromStore: true,
			loadErr:   true,
			errRegex: test.TrimYAML(`
				^load session:
				  load broke$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps timeNow.

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

			session, err := manager.Validate(t.Context(), token)
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

			if got, want := session.UserID, userID; got != want {
				t.Errorf("Session.UserID mismatch\ngot:  %q\nwant: %q", got, want)
			}

			if got, want := fake.touches, tc.wantTouches; got != want {
				t.Errorf("fakeStore.touches mismatch\ngot:  %d\nwant: %d", got, want)
			}
		})
	}
}

func TestManager_Validate__expiryDeletes(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")
	clock.advance(testConfig().IdleTimeout + time.Minute)

	genBefore := genOf(manager)

	_, err := manager.Validate(t.Context(), token)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("delete success: error type mismatch\ngot:  %v\nwant: %v", err, ErrInvalidSession)
	}

	if sLen, cLen := len(fake.sessions), len(manager.cache); sLen != 0 || cLen != 0 {
		t.Errorf(
			"delete success: len(fakeStore.sessions), len(Manager.cache) mismatch\n"+
				"got:  %d, %d\nwant: 0, 0 (expired)",
			sLen, cLen,
		)
	}

	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("delete success: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	token = mustStartFor(t, manager, "user-a")
	clock.advance(testConfig().IdleTimeout + time.Minute)
	fake.deleteErr = errors.New("delete broke")
	genBefore = genOf(manager)
	_, err = manager.Validate(t.Context(), token)
	if !errors.Is(err, ErrInvalidSession) {
		t.Errorf("delete failure: error type mismatch\ngot:  %v\nwant: %v", err, ErrInvalidSession)
	}

	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("delete failure: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	if fake.deletes != 2 {
		t.Errorf("delete failure: fakeStore.deletes mismatch\ngot:  %d\nwant: 2", fake.deletes)
	}
}

func TestManager_Validate__shortIdleTimeout(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	// Shorter than touchInterval, which the config validator accepts.
	const idleTimeout = time.Minute
	manager := New(fake, Config{Lifetime: time.Hour, IdleTimeout: idleTimeout})
	token := mustStartFor(t, manager, "user-a")

	// touchEvery drops to half the idle timeout, so it never expires.
	for elapsed := 20 * time.Second; elapsed <= 3*idleTimeout; elapsed += 20 * time.Second {
		clock.advance(20 * time.Second)
		_, err := manager.Validate(t.Context(), token)
		if err != nil {
			t.Fatalf("elapsed %s: Manager.Validate() failed: %v", elapsed, err)
		}
	}
}

func TestManager_Validate__touchFlush(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")

	genBefore := genOf(manager)

	clock.advance(touchInterval / 2)
	_, err := manager.Validate(t.Context(), token)
	if err != nil {
		t.Fatalf("within interval: Manager.Validate() failed: %v", err)
	}

	if fake.touches != 0 {
		t.Errorf("within interval: fakeStore.touches mismatch\ngot:  %d\nwant: 0", fake.touches)
	}

	clock.advance(touchInterval)
	_, err = manager.Validate(t.Context(), token)
	if err != nil {
		t.Fatalf("past interval: Manager.Validate() failed: %v", err)
	}

	if fake.touches != 1 {
		t.Errorf("past interval: fakeStore.touches mismatch\ngot:  %d\nwant: 1", fake.touches)
	}
	if got := fake.sessions[auth.HashToken(token)].LastSeenAt; !got.Equal(clock.now) {
		t.Errorf("past interval: Session.LastSeenAt mismatch\ngot:  %v\nwant: %v", got, clock.now)
	}

	fake.touchErr = errors.New("touch broke")
	clock.advance(touchInterval * 2)
	_, err = manager.Validate(t.Context(), token)
	if err != nil {
		t.Errorf("flush failure: Manager.Validate() failed: %v", err)
	}

	// Sliding the window never bumps Manager.gen.
	if got := genOf(manager); got != genBefore {
		t.Errorf("flush failure: Manager.gen mismatch\ngot:  %d\nwant: %d", got, genBefore)
	}
}

func TestManager_Validate__touchFlushFrequentAccess(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	loginTime := clock.now
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")

	for i := range 4 {
		clock.advance(touchInterval / 2)
		_, err := manager.Validate(t.Context(), token)
		if err != nil {
			t.Fatalf("validation %d: Manager.Validate() failed: %v", i, err)
		}
	}

	// Four half touchIntervals span two touchIntervals, so the store flushes twice.
	if fake.touches != 2 {
		t.Errorf("fakeStore.touches mismatch\ngot:  %d\nwant: 2", fake.touches)
	}

	if got := fake.sessions[auth.HashToken(token)].LastSeenAt; !got.After(loginTime) {
		t.Errorf("Session.LastSeenAt mismatch\ngot:  %v\nwant: after %v", got, loginTime)
	}
}

func TestManager_Validate__concurrent(t *testing.T) {
	fake := newFakeStore()
	manager := New(fake, testConfig())
	tokens := make([]string, 8)
	for i := range tokens {
		tokens[i] = mustStartFor(t, manager, fmt.Sprintf("user-%d", i))
	}
	revoked := tokens[len(tokens)-1]

	genBefore := genOf(manager)

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i, token := range tokens {
		wg.Go(func() {
			<-start
			for range 25 {
				_, err := manager.Validate(t.Context(), token)
				if err != nil && (token != revoked || !errors.Is(err, ErrInvalidSession)) {
					t.Errorf("token %d: Manager.Validate() failed: %v", i, err)
					return
				}
			}
		})
	}
	wg.Go(func() {
		<-start
		if err := manager.Revoke(t.Context(), revoked); err != nil {
			t.Errorf("Manager.Revoke() failed: %v", err)
		}
	})
	close(start)
	wg.Wait()

	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	_, err := manager.Validate(t.Context(), tokens[0])
	if err != nil {
		t.Errorf("live: Manager.Validate() failed: %v", err)
	}

	_, err = manager.Validate(t.Context(), revoked)
	if !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"revoked: Manager.Validate() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInvalidSession,
		)
	}
}

func TestManager_Validate__tamperedToken(t *testing.T) {
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

	_, err := manager.Validate(t.Context(), tampered)

	if !errors.Is(err, ErrInvalidSession) {
		t.Errorf("tampered: error type mismatch\ngot:  %v\nwant: %v", err, ErrInvalidSession)
	}

	_, err = manager.Validate(t.Context(), token)
	if err != nil {
		t.Errorf("genuine: Manager.Validate() failed: %v", err)
	}
}

func TestManager_Validate__revokeDuringLoadDoesNotResurrect(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	manager := New(fake, testConfig())

	userID := "user-a"
	token := mustStartFor(t, manager, userID)
	tokenHash := auth.HashToken(token)
	dropFromCache(manager, tokenHash)

	// Revokes between the load and the cache write.
	revoked := false
	fake.onLookup = func() {
		if revoked {
			return
		}
		revoked = true
		if err := manager.Revoke(t.Context(), token); err != nil {
			t.Errorf("Manager.Revoke() failed: %v", err)
		}
	}

	// The row was live when read, so this call may succeed.
	_, _ = manager.Validate(t.Context(), token)

	manager.mu.Lock()
	_, cached := manager.cache[tokenHash]
	manager.mu.Unlock()
	if cached {
		t.Error("Manager.cache[tokenHash] mismatch\ngot:  present\nwant: absent (revoked)")
	}
	if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrInvalidSession)
	}
}

func TestManager_Alive(t *testing.T) {
	userID := "user-a"

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
				return auth.HashToken(mustStartFor(t, m, userID))
			},
			want: true,
		},
		{
			name: "live/store hit after restart",
			tokenHash: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return auth.HashToken(mustStartFor(t, m, userID))
			},
			fromStore: true,
			want:      true,
		},
		{
			name: "revoked session",
			tokenHash: func(t *testing.T, m *Manager, _ *fakeClock) string {
				token := mustStartFor(t, m, userID)
				if err := m.Revoke(t.Context(), token); err != nil {
					t.Fatalf("Manager.Revoke() failed: %v", err)
				}
				return auth.HashToken(token)
			},
			want: false,
		},
		{
			name: "absolutely expired session",
			tokenHash: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.Lifetime + time.Minute)
				return auth.HashToken(token)
			},
			want: false,
		},
		{
			name: "idle-expired session",
			tokenHash: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				clock.advance(config.IdleTimeout + time.Minute)
				return auth.HashToken(token)
			},
			want: false,
		},
		{
			name: "check does not slide the idle window",
			tokenHash: func(t *testing.T, m *Manager, clock *fakeClock) string {
				token := mustStartFor(t, m, userID)
				// A check halfway to idle expiry, then the rest of the window passes, so
				// the session is dead unless the check slid it.
				clock.advance(config.IdleTimeout / 2)
				if !m.Alive(t.Context(), auth.HashToken(token)) {
					t.Fatal("mid-window: Manager.Alive() mismatch\ngot:  false\nwant: true")
				}
				clock.advance(time.Minute + config.IdleTimeout/2)
				return auth.HashToken(token)
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
			name: "store failure counts as alive",
			tokenHash: func(t *testing.T, m *Manager, _ *fakeClock) string {
				return auth.HashToken(mustStartFor(t, m, userID))
			},
			fromStore: true,
			loadErr:   true,
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps timeNow.

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

			got := manager.Alive(t.Context(), tokenHash)

			if got != tc.want {
				t.Errorf("Manager.Alive() mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestManager_Revoke(t *testing.T) {
	fake := newFakeStore()
	manager := New(fake, testConfig())
	token := mustStartFor(t, manager, "user-a")

	genBefore := genOf(manager)

	err := manager.Revoke(t.Context(), token)
	if err != nil {
		t.Fatalf("delete success: unexpected error: %v", err)
	}

	if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf(
			"delete success: Manager.Validate() error type mismatch\ngot:  %v\nwant: %v",
			err, ErrInvalidSession,
		)
	}

	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("delete success: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	fake.deleteErr = errors.New("delete broke")
	genBefore = genOf(manager)
	err = manager.Revoke(t.Context(), token)
	if err == nil {
		t.Error("delete failure: error mismatch\ngot:  nil\nwant: non-nil")
	}

	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("delete failure: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestManager_RevokeUser(t *testing.T) {
	fake := newFakeStore()
	manager := New(fake, testConfig())
	userA, userB := "user-a", "user-b"
	tokenA1 := mustStartFor(t, manager, userA)
	tokenA2 := mustStartFor(t, manager, userA)
	tokenB := mustStartFor(t, manager, userB)

	genBefore := genOf(manager)

	err := manager.RevokeUser(t.Context(), userA)
	if err != nil {
		t.Fatalf("delete success: unexpected error: %v", err)
	}

	for i, token := range []string{tokenA1, tokenA2} {
		if _, err := manager.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
			t.Errorf(
				"user-a token %d: Manager.Validate() error type mismatch\ngot:  %v\nwant: %v",
				i, err, ErrInvalidSession,
			)
		}
	}

	if _, err := manager.Validate(t.Context(), tokenB); err != nil {
		t.Errorf("user-b: Manager.Validate() failed: %v", err)
	}

	// Once for the sweep, not once per session.
	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("delete success: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	fake.deleteErr = errors.New("delete broke")
	err = manager.RevokeUser(t.Context(), userB)
	if err == nil {
		t.Error("delete failure: error mismatch\ngot:  nil\nwant: non-nil")
	}
}

func TestManager_PruneExpired(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	clock.use(t)
	fake := newFakeStore()
	userID := "user-a"
	manager := New(fake, testConfig())
	expired := mustStartFor(t, manager, userID)
	clock.advance(testConfig().Lifetime + time.Minute)
	live := mustStartFor(t, manager, userID)

	genBefore := genOf(manager)

	err := manager.PruneExpired(t.Context())
	if err != nil {
		t.Fatalf("delete success: unexpected error: %v", err)
	}

	if sLen, cLen := len(fake.sessions), len(manager.cache); sLen != 1 || cLen != 1 {
		t.Errorf(
			"delete success: len(fakeStore.sessions), len(Manager.cache) mismatch\n"+
				"got:  %d, %d\nwant: 1, 1",
			sLen, cLen,
		)
	}
	if _, ok := manager.cache[auth.HashToken(expired)]; ok {
		t.Error(
			"delete success: Manager.cache[HashToken(expired)] mismatch\n" +
				"got:  present\nwant: absent (expired)",
		)
	}
	if _, err := manager.Validate(t.Context(), live); err != nil {
		t.Errorf("live: Manager.Validate() failed: %v", err)
	}

	// Once for the sweep, not once per session.
	if got, want := genOf(manager), genBefore+1; got != want {
		t.Errorf("delete success: Manager.gen mismatch\ngot:  %d\nwant: %d", got, want)
	}

	fake.deleteErr = errors.New("delete broke")
	err = manager.PruneExpired(t.Context())
	if err == nil {
		t.Error("delete failure: error mismatch\ngot:  nil\nwant: non-nil")
	}
}
