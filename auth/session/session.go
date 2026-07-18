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
	InsertSession(ctx context.Context, session store.Session) error                   // InsertSession persists a new session.
	SessionByTokenHash(ctx context.Context, tokenHash string) (*store.Session, error) // SessionByTokenHash returns the session with tokenHash.
	TouchSession(ctx context.Context, tokenHash string, lastSeenAt time.Time) error   // TouchSession updates the session's last_seen_at.
	DeleteSession(ctx context.Context, tokenHash string) error                        // DeleteSession removes the session with tokenHash.
	DeleteSessionsForUser(ctx context.Context, userID string) error                   // DeleteSessionsForUser removes every session of userID.
	DeleteExpiredSessions(ctx context.Context, now time.Time) error                   // DeleteExpiredSessions prunes sessions past their absolute expiry.
}

// Manager mints, validates, and revokes sessions.
type Manager struct {
	store  Store
	config Config

	mu    sync.Mutex
	cache map[string]store.Session // tokenHash -> session.
}

// New creates a Manager persisting via sessionStore, bounded by config.
func New(sessionStore Store, config Config) *Manager {
	return &Manager{
		store:  sessionStore,
		config: config,
		cache:  make(map[string]store.Session),
	}
}

// hashToken derives the storage key of a token.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Start mints a session for userID, returning the opaque token to set as
// the cookie value. Only the token's hash is stored.
func (m *Manager) Start(ctx context.Context, userID, ip, userAgent string) (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := randRead(raw); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	token := hex.EncodeToString(raw)

	now := timeNow()
	session := store.Session{
		TokenHash:  hashToken(token),
		UserID:     userID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(m.config.Lifetime),
		LastSeenAt: now,
		IP:         ip,
		UserAgent:  userAgent,
	}

	if err := m.store.InsertSession(ctx, session); err != nil {
		return "", fmt.Errorf("persist session: %w", err)
	}

	m.mu.Lock()
	m.cache[session.TokenHash] = session
	m.mu.Unlock()

	return token, nil
}

// Validate resolves token to its live session, sliding the idle window.
// Unknown, absolutely-expired, and idle-expired tokens all return [ErrInvalidSession]
// (expired rows are deleted on detection).
func (m *Manager) Validate(ctx context.Context, token string) (*store.Session, error) {
	if token == "" {
		return nil, ErrInvalidSession
	}
	tokenHash := hashToken(token)

	m.mu.Lock()
	session, cached := m.cache[tokenHash]
	m.mu.Unlock()

	// Cache miss.
	if !cached {
		fromStore, err := m.store.SessionByTokenHash(ctx, tokenHash)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, ErrInvalidSession
			}
			return nil, fmt.Errorf("load session: %w", err)
		}
		session = *fromStore
	}

	// Expiry: absolute cap, then idle window.
	now := timeNow()
	if now.After(session.ExpiresAt) ||
		now.After(session.LastSeenAt.Add(m.config.IdleTimeout)) {
		m.mu.Lock()
		delete(m.cache, tokenHash)
		m.mu.Unlock()
		m.deleteRow(ctx, tokenHash)
		return nil, ErrInvalidSession
	}

	// Slide the idle window: cache immediately, database at most once per touchInterval.
	flush := now.Sub(session.LastSeenAt) >= touchInterval
	session.LastSeenAt = now
	m.mu.Lock()
	m.cache[tokenHash] = session
	m.mu.Unlock()
	if flush {
		if err := m.store.TouchSession(ctx, tokenHash, now); err != nil {
			logx.Error(err, logx.LogFrom{Primary: "session", Secondary: "touch"}, true)
		}
	}

	return &session, nil
}

// Revoke ends the session of token (logout).
func (m *Manager) Revoke(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	m.mu.Lock()
	delete(m.cache, tokenHash)
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
