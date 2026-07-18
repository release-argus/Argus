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
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

var packageName = "auth_session"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// fakeClock pins timeNow to a mutable instant.
type fakeClock struct {
	now time.Time
}

// use makes the package clock read from c until the test ends.
func (c *fakeClock) use(t *testing.T) {
	t.Helper()

	timeNowHad := timeNow
	timeNow = func() time.Time { return c.now }
	t.Cleanup(func() { timeNow = timeNowHad })
}

// advance moves the clock forward.
func (c *fakeClock) advance(d time.Duration) {
	c.now = c.now.Add(d)
}

// fakeStore is an in-memory [store.Session] store with fault switches.
type fakeStore struct {
	mu       sync.Mutex
	sessions map[string]store.Session

	insertErr error
	loadErr   error
	touchErr  error
	deleteErr error

	touches int // TouchSession calls.
	deletes int // DeleteSession calls.
}

func newFakeStore() *fakeStore {
	return &fakeStore{sessions: make(map[string]store.Session)}
}

func (f *fakeStore) InsertSession(_ context.Context, session store.Session) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.insertErr != nil {
		return f.insertErr
	}
	f.sessions[session.TokenHash] = session
	return nil
}

func (f *fakeStore) SessionByTokenHash(_ context.Context, tokenHash string) (*store.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.loadErr != nil {
		return nil, f.loadErr
	}
	session, ok := f.sessions[tokenHash]
	if !ok {
		return nil, store.ErrNotFound
	}
	return &session, nil
}

func (f *fakeStore) TouchSession(_ context.Context, tokenHash string, lastSeenAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.touches++
	if f.touchErr != nil {
		return f.touchErr
	}
	session := f.sessions[tokenHash]
	session.LastSeenAt = lastSeenAt
	f.sessions[tokenHash] = session
	return nil
}

func (f *fakeStore) DeleteSession(_ context.Context, tokenHash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.deletes++
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.sessions, tokenHash)
	return nil
}

func (f *fakeStore) DeleteSessionsForUser(_ context.Context, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.deleteErr != nil {
		return f.deleteErr
	}
	for tokenHash, session := range f.sessions {
		if session.UserID == userID {
			delete(f.sessions, tokenHash)
		}
	}
	return nil
}

func (f *fakeStore) DeleteExpiredSessions(_ context.Context, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.deleteErr != nil {
		return f.deleteErr
	}
	for tokenHash, session := range f.sessions {
		if session.ExpiresAt.Before(now) {
			delete(f.sessions, tokenHash)
		}
	}
	return nil
}

// testConfig is a week of idle inside a month of lifetime.
func testConfig() Config {
	return Config{
		Lifetime:    30 * 24 * time.Hour,
		IdleTimeout: 7 * 24 * time.Hour,
	}
}

// mustStartFor starts a session for userID, failing the test on error.
func mustStartFor(t *testing.T, manager *Manager, userID string) string {
	t.Helper()

	token, err := manager.Start(t.Context(), userID, "127.0.0.1", "go-test")
	if err != nil {
		t.Fatalf(
			"%s Start(%q) failed: %v",
			packageName, userID, err,
		)
	}
	return token
}
