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

// Package session manages login sessions: opaque 256-bit tokens whose
// SHA-256 is persisted (with a write-through in-memory cache), a sliding
// idle window under an absolute lifetime cap, and revocation.
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/internal/logx"
)

// tokenBytes is the entropy of a session token (256-bit).
const tokenBytes = 32

// MaxSessionsPerUser caps a user's concurrent sessions; starting a session
// past the cap evicts their least-recently active one.
const MaxSessionsPerUser = 10

// touchInterval is how often a session's last-seen time is flushed to the
// database (activity within the interval only updates the cache).
const touchInterval = 2 * time.Minute

// ErrInvalidSession is returned for tokens that are unknown, expired,
// or idle past their window - indistinguishable to the caller.
var ErrInvalidSession = errors.New("invalid or expired session")

// timeNow returns the current UTC time (overridable for tests).
// see [time.Now] followed by [time.Time.UTC].
var timeNow = func() time.Time { return time.Now().UTC() }

// randRead fills b with cryptographically random bytes (overridable for tests).
// see [rand.Read].
var randRead = rand.Read

// Config bounds session validity.
type Config struct {
	Lifetime    time.Duration // Absolute cap from creation.
	IdleTimeout time.Duration // Sliding window from last activity.
}

// Store is the slice of the auth store the [Manager] needs.
type Store interface {
	// InsertSession persists a new session.
	InsertSession(ctx context.Context, session store.Session) error
	// SessionByTokenHash returns the session with tokenHash.
	SessionByTokenHash(ctx context.Context, tokenHash string) (*store.Session, error)
	// TouchSession updates the session's last_seen_at.
	TouchSession(ctx context.Context, tokenHash string, lastSeenAt time.Time) error
	// DeleteSession removes the session with tokenHash.
	DeleteSession(ctx context.Context, tokenHash string) error
	// DeleteSessionsForUser removes every session of userID.
	DeleteSessionsForUser(ctx context.Context, userID string) error
	// DeleteExpiredSessions prunes sessions past their absolute expiry.
	DeleteExpiredSessions(ctx context.Context, now time.Time) error
	// TrimSessionsForUser deletes all but the keep most-recently active sessions of userID.
	TrimSessionsForUser(ctx context.Context, userID string, keep int) ([]string, error)
}

// Manager mints, validates, and revokes sessions.
type Manager struct {
	store  Store
	config Config

	mu    sync.Mutex
	cache map[string]store.Session // tokenHash -> session.
	// Bumped on mint/eviction/revoke; guards [Manager.Validate]'s cache-MISS
	// write-back against resurrecting a concurrently-revoked session.
	gen uint64
}

// New creates a Manager persisting via sessionStore, bounded by config.
func New(sessionStore Store, config Config) *Manager {
	return &Manager{
		store:  sessionStore,
		config: config,
		cache:  make(map[string]store.Session),
	}
}

// HashToken derives the storage key of a session token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Start mints a session for userID, returning the opaque token to set as
// the cookie value, plus the token hashes of any sessions evicted to keep
// the user within [MaxSessionsPerUser]. Only the token's hash is stored.
func (m *Manager) Start(ctx context.Context, userID, ip, userAgent string) (token string, evicted []string, err error) {
	raw := make([]byte, tokenBytes)
	if _, err := randRead(raw); err != nil {
		return "", nil, fmt.Errorf("generate session token: %w", err)
	}
	token = hex.EncodeToString(raw)

	now := timeNow()
	session := store.Session{
		TokenHash:  HashToken(token),
		UserID:     userID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(m.config.Lifetime),
		LastSeenAt: now,
		IP:         ip,
		UserAgent:  userAgent,
	}

	if err := m.store.InsertSession(ctx, session); err != nil {
		return "", nil, fmt.Errorf("persist session: %w", err)
	}

	// Enforce the session cap; failures are only logged (retried on the next login).
	evicted, err = m.store.TrimSessionsForUser(ctx, userID, MaxSessionsPerUser)
	if err != nil {
		logx.Error(err, logx.LogFrom{Primary: "session", Secondary: "trim"}, true)
	}

	m.mu.Lock()
	m.cache[session.TokenHash] = session
	for _, tokenHash := range evicted {
		delete(m.cache, tokenHash)
	}
	if len(evicted) > 0 {
		m.gen++
	}
	m.mu.Unlock()

	return token, evicted, nil
}

// expired reports whether session is past its absolute cap or its idle window at now.
func (m *Manager) expired(session store.Session, now time.Time) bool {
	return now.After(session.ExpiresAt) ||
		now.After(session.LastSeenAt.Add(m.config.IdleTimeout))
}

// evict drops tokenHash from the cache and store, bumping gen so a concurrent
// cache-MISS write-back cannot resurrect it. The caller must not hold m.mu.
func (m *Manager) evict(ctx context.Context, tokenHash string) error {
	m.mu.Lock()
	delete(m.cache, tokenHash)
	m.gen++
	m.mu.Unlock()
	m.deleteRow(ctx, tokenHash)
	return ErrInvalidSession
}

// Validate resolves token to its live session, sliding the idle window.
// Unknown, absolutely-expired, and idle-expired tokens all return [ErrInvalidSession]
// (expired rows are deleted on detection).
func (m *Manager) Validate(ctx context.Context, token string) (*store.Session, error) {
	if token == "" {
		return nil, ErrInvalidSession
	}
	tokenHash := HashToken(token)
	now := timeNow()

	m.mu.Lock()
	session, cached := m.cache[tokenHash]
	if cached {
		if m.expired(session, now) {
			m.mu.Unlock()
			return nil, m.evict(ctx, tokenHash)
		}
		flush := now.Sub(session.LastSeenAt) >= touchInterval
		session.LastSeenAt = now
		m.cache[tokenHash] = session
		m.mu.Unlock()
		if flush {
			if err := m.store.TouchSession(ctx, tokenHash, now); err != nil {
				logx.Error(err, logx.LogFrom{Primary: "session", Secondary: "touch"}, true)
			}
		}
		return &session, nil
	}
	gen := m.gen
	m.mu.Unlock()

	// Cache miss: load from the store.
	fromStore, err := m.store.SessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, fmt.Errorf("load session: %w", err)
	}
	session = *fromStore

	// Expiry: absolute cap/idle window.
	if m.expired(session, now) {
		return nil, m.evict(ctx, tokenHash)
	}

	// Slide the idle window: cache immediately, database at most once per [touchInterval].
	flush := now.Sub(session.LastSeenAt) >= touchInterval
	session.LastSeenAt = now
	m.mu.Lock()
	// Only populate the cache if no revoke/eviction landed during the load.
	if m.gen == gen {
		m.cache[tokenHash] = session
	}
	m.mu.Unlock()
	if flush {
		if err := m.store.TouchSession(ctx, tokenHash, now); err != nil {
			logx.Error(err, logx.LogFrom{Primary: "session", Secondary: "touch"}, true)
		}
	}

	return &session, nil
}

// Alive reports whether tokenHash still maps to a live session, without
// sliding the idle window. Only a missing row reads as dead - store
// failures read as alive so an outage doesn't disconnect every client.
func (m *Manager) Alive(ctx context.Context, tokenHash string) bool {
	m.mu.Lock()
	session, cached := m.cache[tokenHash]
	m.mu.Unlock()

	if !cached {
		fromStore, err := m.store.SessionByTokenHash(ctx, tokenHash)
		if err != nil {
			return !errors.Is(err, store.ErrNotFound)
		}
		session = *fromStore
	}

	return !m.expired(session, timeNow())
}

// Revoke ends the session of token (logout).
func (m *Manager) Revoke(ctx context.Context, token string) error {
	tokenHash := HashToken(token)

	m.mu.Lock()
	delete(m.cache, tokenHash)
	m.gen++
	m.mu.Unlock()

	if err := m.store.DeleteSession(ctx, tokenHash); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeUser ends every session of userID
// (password reset, disable, delete - logout everywhere).
func (m *Manager) RevokeUser(ctx context.Context, userID string) error {
	m.mu.Lock()
	for tokenHash, session := range m.cache {
		if session.UserID == userID {
			delete(m.cache, tokenHash)
		}
	}
	m.gen++
	m.mu.Unlock()

	if err := m.store.DeleteSessionsForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke user's sessions: %w", err)
	}
	return nil
}

// PruneExpired removes sessions past their absolute expiry.
func (m *Manager) PruneExpired(ctx context.Context) error {
	now := timeNow()

	m.mu.Lock()
	for tokenHash, session := range m.cache {
		if now.After(session.ExpiresAt) {
			delete(m.cache, tokenHash)
		}
	}
	m.gen++
	m.mu.Unlock()

	if err := m.store.DeleteExpiredSessions(ctx, now); err != nil {
		return fmt.Errorf("prune expired sessions: %w", err)
	}
	return nil
}

// deleteRow removes tokenHash's database row; failures are only logged.
func (m *Manager) deleteRow(ctx context.Context, tokenHash string) {
	if err := m.store.DeleteSession(ctx, tokenHash); err != nil {
		logx.Error(err, logx.LogFrom{Primary: "session", Secondary: "drop"}, true)
	}
}
