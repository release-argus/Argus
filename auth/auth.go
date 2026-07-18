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

// Package auth provides the core authentication types shared by the
// providers, stores, session management, and the web layer:
// the Identity returned by authentication, the local User it resolves to,
// and the sentinel errors of the authentication flow.
package auth

import (
	"errors"
	"time"

	"github.com/release-argus/Argus/auth/rbac"
)

// ErrInvalidCredentials is returned for every authentication failure.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Identity is what a Provider returns on successful authentication.
// It describes who authenticated at the provider, not a user -
// identity resolution maps it onto a User.
type Identity struct {
	Provider    string // Provider name, e.g. "local".
	Subject     string // Stable identifier at the provider.
	Username    string // Login username.
	DisplayName string // Human-readable name, if known.
	Email       string // Email address, if known.
}

// User is a local user account.
type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitzero"`
	Email       string    `json:"email,omitzero"`
	Enabled     bool      `json:"enabled"`
	Groups      []string  `json:"groups"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Credentials is the stored credential material of a local user,
// as needed by the local authentication provider.
type Credentials struct {
	UserID       string // User's ID.
	Username     string // Canonical (stored) username.
	DisplayName  string // Human-readable name.
	Email        string // Email address.
	PasswordHash string // Encoded argon2id hash.
	Enabled      bool   // Whether the user may log in.
}

// Context is the resolved authentication state of a request:
// the Identity that authenticated, the User it maps to, and the
// User's evaluated permissions.
type Context struct {
	Identity    Identity
	User        User
	Grants      []rbac.Grant // Raw grants (for /auth/me).
	Permissions *rbac.PermissionSet
}
