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

// Package local implements the "local" authentication provider,
// verifying credentials against Argus's own user store.
package local

import (
	"context"
	"fmt"
	"sync"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/internal/logx"
)

// Name is the provider name of the local provider.
const Name = "local"

// hashPassword derives a password hash (overridable for tests).
// See [password.Hash].
var hashPassword = password.Hash

// verifyPassword checks a password against a hash (overridable for tests).
// See [password.Verify].
var verifyPassword = password.Verify

// UserStore is the slice of the user store the local provider needs.
type UserStore interface {
	// LocalCredentials returns the credential material for username.
	LocalCredentials(ctx context.Context, username string) (*auth.Credentials, error)
	// UpdatePasswordHash replaces the stored hash for userID
	// (used for transparent parameter upgrades on login).
	UpdatePasswordHash(ctx context.Context, userID, hash string) error
}

// Provider authenticates against the local user store.
type Provider struct {
	store UserStore

	dummyOnce sync.Once
	dummyHash string // Hash verified against for unknown users, to equalise timing.
}

// New creates a local Provider backed by store.
func New(store UserStore) *Provider {
	return &Provider{store: store}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return Name
}

// Authenticate verifies username/password against the user store.
// Unknown users, wrong passwords, disabled users, and users without a
// password all return [auth.ErrInvalidCredentials].
func (p *Provider) Authenticate(ctx context.Context, username, password string) (*auth.Identity, error) {
	creds, err := p.store.LocalCredentials(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("look up user: %w", err)
	}

	// Verify against a dummy hash for unknown/password-less users so response
	// timing does not reveal whether the account exists.
	hash := p.dummy()
	realHash := false
	if creds != nil && creds.PasswordHash != "" {
		hash = creds.PasswordHash
		realHash = true
	}

	match, needsRehash, err := verifyPassword(password, hash)
	if err != nil {
		return nil, fmt.Errorf("verify password: %w", err)
	}

	if !realHash || !match || !creds.Enabled {
		return nil, auth.ErrInvalidCredentials
	}

	// Upgrade hashes derived with outdated parameters.
	if needsRehash {
		p.rehash(ctx, creds.UserID, password)
	}

	return &auth.Identity{
			Provider:    Name,
			Subject:     creds.UserID,
			Username:    creds.Username,
			DisplayName: creds.DisplayName,
			Email:       creds.Email,
		},
		nil
}

// rehash re-hashes password with current parameters and stores it.
func (p *Provider) rehash(ctx context.Context, userID, password string) {
	logFrom := logx.LogFrom{Primary: "auth", Secondary: "rehash"}

	newHash, err := hashPassword(password)
	if err != nil {
		logx.Error(err, logFrom, true)
		return
	}
	if err := p.store.UpdatePasswordHash(ctx, userID, newHash); err != nil {
		logx.Error(err, logFrom, true)
	}
}

// dummy lazily derives the dummy hash used to equalise verification timing.
func (p *Provider) dummy() string {
	p.dummyOnce.Do(func() {
		// Random throwaway password; only the derivation cost matters.
		hash, err := hashPassword("argus-dummy-timing-equalisation")
		if err != nil {
			// Fall back to a pre-computed encoding of the same password.
			hash = "$argon2id$v=19$m=19456,t=2,p=1" +
				"$YXJndXMtZHVtbXktc2FsdA" +
				"$KDiYCbFMevkzOp1i2Cm+r5GnLpDb8pRU3O7eyIfCTGw"
		}
		p.dummyHash = hash
	})
	return p.dummyHash
}
