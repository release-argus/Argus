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

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session is a persisted login session (the token itself is never stored,
// only its hash).
type Session struct {
	TokenHash  string    // SHA-256 hash of the opaque session token.
	UserID     string    // Owning user.
	CreatedAt  time.Time // When the session was minted.
	ExpiresAt  time.Time // Absolute expiry cap.
	LastSeenAt time.Time // Basis of the sliding idle window.
	IP         string    // Client IP at login.
	UserAgent  string    // Client user agent at login.
}

// InsertSession persists a new session.
func (s *Store) InsertSession(ctx context.Context, session Session) error {
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (token_hash, user_id, created_at, expires_at, last_seen_at, ip, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?);`,
		session.TokenHash, session.UserID,
		session.CreatedAt.Format(timeFormat),
		session.ExpiresAt.Format(timeFormat),
		session.LastSeenAt.Format(timeFormat),
		session.IP, session.UserAgent,
	); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// SessionByTokenHash returns the [Session] with tokenHash ([ErrNotFound] if absent).
func (s *Store) SessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT token_hash, user_id, created_at, expires_at, last_seen_at, ip, user_agent
		FROM sessions
		WHERE token_hash = ?;`,
		tokenHash,
	)

	var (
		session    Session
		createdAt  string
		expiresAt  string
		lastSeenAt string
	)
	if err := row.Scan(
		&session.TokenHash, &session.UserID,
		&createdAt, &expiresAt, &lastSeenAt,
		&session.IP, &session.UserAgent,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load session: %w", err)
	}

	var err error
	if session.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if session.ExpiresAt, err = parseTime(expiresAt); err != nil {
		return nil, err
	}
	if session.LastSeenAt, err = parseTime(lastSeenAt); err != nil {
		return nil, err
	}

	return &session, nil
}

// TouchSession updates the session's last_seen_at.
func (s *Store) TouchSession(ctx context.Context, tokenHash string, lastSeenAt time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET last_seen_at = ? WHERE token_hash = ?;`,
		lastSeenAt.Format(timeFormat), tokenHash,
	); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

// DeleteSession removes the session with tokenHash.
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE token_hash = ?;`, tokenHash,
	); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteSessionsForUser removes every session of userID
// (logout-everywhere: password reset, disable, delete).
func (s *Store) DeleteSessionsForUser(ctx context.Context, userID string) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = ?;`, userID,
	); err != nil {
		return fmt.Errorf("delete user's sessions: %w", err)
	}
	return nil
}

// DeleteExpiredSessions prunes sessions past their absolute expiry.
func (s *Store) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE expires_at < ?;`, now.Format(timeFormat),
	); err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}
