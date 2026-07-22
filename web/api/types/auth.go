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

package types

import (
	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/store"
)

// LoginRequest is the body of POST /api/v1/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthMe is the response of /auth/login, /auth/setup and /auth/me:
// the authenticated user and their effective permission grants.
type AuthMe struct {
	User        auth.User    `json:"user"`
	Permissions []rbac.Grant `json:"permissions"`
}

// SetupState is the response of GET /api/v1/auth/setup: whether the
// first-run setup (creating the first administrator) is still pending.
type SetupState struct {
	SetupRequired bool `json:"setup_required"`
}

// SetupRequest is the body of POST /api/v1/auth/setup:
// the first administrator's account details.
type SetupRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitzero"`
	Password    string `json:"password"`
}

// UserCreateRequest is the body of POST /api/v1/users.
type UserCreateRequest struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	DisplayName string   `json:"display_name,omitzero"`
	Email       string   `json:"email,omitzero"`
	Groups      []string `json:"groups,omitzero"`
}

// UserPatchRequest is the body of PATCH /api/v1/users/{id};
// nil fields stay unchanged.
type UserPatchRequest struct {
	DisplayName *string   `json:"display_name,omitzero"`
	Email       *string   `json:"email,omitzero"`
	Enabled     *bool     `json:"enabled,omitzero"`
	Groups      *[]string `json:"groups,omitzero"`
	Password    *string   `json:"password,omitzero"`
}

// GroupCreateRequest is the body of POST /api/v1/groups.
type GroupCreateRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitzero"`
	Permissions []rbac.Grant `json:"permissions,omitzero"`
}

// GroupPatchRequest is the body of PATCH /api/v1/groups/{id};
// nil fields stay unchanged (Permissions is a replace-set).
type GroupPatchRequest struct {
	Name        *string       `json:"name,omitzero"`
	Description *string       `json:"description,omitzero"`
	Permissions *[]rbac.Grant `json:"permissions,omitzero"`
}

// APITokenCreateRequest is the body of POST /api/v1/tokens.
type APITokenCreateRequest struct {
	Name      string `json:"name"`
	ExpiresIn string `json:"expires_in,omitzero"` // Optional duration, e.g. "720h".
}

// APITokenCreated is the response of POST /api/v1/tokens.
// Token is the plaintext - shown here and never again.
type APITokenCreated struct {
	store.APIToken
	Token string `json:"token"`
}

// PermissionCatalogue is the response of GET /api/v1/permissions:
// the valid (resource, action, scope) matrix, defined in code.
type PermissionCatalogue struct {
	Resources []rbac.ResourcePermissions `json:"resources"`
}
