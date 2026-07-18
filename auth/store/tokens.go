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
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// apiTokenPrefix marks Argus API tokens.
const apiTokenPrefix = "argus_"

// apiTokenBytes is the entropy of an API token (256-bit).
const apiTokenBytes = 32

// apiTokenDisplayLength is how many leading characters of the token are
// stored (and shown) so users can tell their tokens apart.
const apiTokenDisplayLength = len(apiTokenPrefix) + 8

// APIToken is a user-created access token for non-browser API requests
// (the plaintext token is never stored, only its SHA-256).
type APIToken struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"` // Leading characters, for identification.
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`   // nil = no expiry.
	LastUsedAt *time.Time `json:"last_used_at,omitempty"` // nil = never used.
}

// hashAPIToken derives the storage key of an API token.
func hashAPIToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// CreateAPIToken mints an API token for userID, returning the plaintext
// exactly once - it cannot be retrieved again.
func (s *Store) CreateAPIToken(
	ctx context.Context,
	userID, name string,
	expiresAt *time.Time,
) (string, *APIToken, error) {
	raw := make([]byte, apiTokenBytes)
	if _, err := randRead(raw); err != nil {
		return "", nil, fmt.Errorf("generate API token: %w", err)
	}
	plaintext := apiTokenPrefix + hex.EncodeToString(raw)

	token := APIToken{
		ID:        newID(),
		UserID:    userID,
		Name:      name,
		Prefix:    plaintext[:apiTokenDisplayLength],
		CreatedAt: timeNow(),
		ExpiresAt: expiresAt,
	}

	var expires any
	if expiresAt != nil {
		expires = expiresAt.Format(timeFormat)
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO api_tokens (id, user_id, name, token_hash, prefix, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);`,
		token.ID, token.UserID, token.Name,
		hashAPIToken(plaintext), token.Prefix,
		token.CreatedAt.Format(timeFormat), expires,
	); err != nil {
		return "", nil, fmt.Errorf("insert API token: %w", err)
	}

	return plaintext, &token, nil
}

// APITokensForUser lists userID's tokens, newest first.
func (s *Store) APITokensForUser(ctx context.Context, userID string) ([]APIToken, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, name, prefix, created_at, expires_at, last_used_at
		FROM api_tokens
		WHERE user_id = ?
		ORDER BY created_at DESC, id;`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list API tokens: %w", err)
	}
	defer rows.Close()

	return scanAll(rows, "API tokens", scanAPIToken)
}

// APITokenByToken resolves a plaintext token to its record
// ([ErrNotFound] if absent). Expiry is the caller's concern.
func (s *Store) APITokenByToken(ctx context.Context, plaintext string) (*APIToken, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, prefix, created_at, expires_at, last_used_at
		FROM api_tokens
		WHERE token_hash = ?;`,
		hashAPIToken(plaintext),
	)

	token, err := scanAPIToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return token, nil
}

// TouchAPIToken updates the token's last_used_at.
func (s *Store) TouchAPIToken(ctx context.Context, id string, lastUsedAt time.Time) error {
	if _, err := s.db.ExecContext(ctx,
		`UPDATE api_tokens SET last_used_at = ? WHERE id = ?;`,
		lastUsedAt.Format(timeFormat), id,
	); err != nil {
		return fmt.Errorf("touch API token: %w", err)
	}
	return nil
}

// DeleteAPIToken removes userID's token with id
// ([ErrNotFound] when absent or owned by someone else).
func (s *Store) DeleteAPIToken(ctx context.Context, userID, id string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM api_tokens WHERE id = ? AND user_id = ?;`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete API token: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete API token: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteExpiredAPITokens prunes tokens past their expiry.
func (s *Store) DeleteExpiredAPITokens(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM api_tokens WHERE expires_at IS NOT NULL AND expires_at < ?;`,
		timeNow().Format(timeFormat),
	); err != nil {
		return fmt.Errorf("delete expired API tokens: %w", err)
	}
	return nil
}

// scanAPIToken scans an api_tokens row.
func scanAPIToken(row scanner) (*APIToken, error) {
	var (
		token      APIToken
		createdAt  string
		expiresAt  sql.NullString
		lastUsedAt sql.NullString
	)
	if err := row.Scan(
		&token.ID, &token.UserID, &token.Name, &token.Prefix,
		&createdAt, &expiresAt, &lastUsedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err //nolint:wrapcheck
		}
		return nil, fmt.Errorf("scan API token: %w", err)
	}

	var err error
	if token.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		expires, err := parseTime(expiresAt.String)
		if err != nil {
			return nil, err
		}
		token.ExpiresAt = &expires
	}
	if lastUsedAt.Valid {
		lastUsed, err := parseTime(lastUsedAt.String)
		if err != nil {
			return nil, err
		}
		token.LastUsedAt = &lastUsed
	}

	return &token, nil
}
