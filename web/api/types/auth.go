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

// AuthMe is the shared response of the /auth/me-shaped endpoints - /auth/login,
// /auth/setup, GET and PATCH /auth/me, and PUT and DELETE
// /auth/me/preferences: the authenticated user, their effective permission
// grants, and any dashboard preferences they have saved.
type AuthMe struct {
	User        auth.User                   `json:"user"`
	Permissions []rbac.Grant                `json:"permissions"`
	Preferences *store.DashboardPreferences `json:"preferences,omitzero"`
}

// SetupState is the response of GET /api/v1/auth/setup: the pre-login state
// of the instance - whether the first-run setup (creating the first
// administrator) is still pending, and any demo credentials to prefill.
type SetupState struct {
	SetupRequired bool             `json:"setup_required"`
	Demo          *DemoCredentials `json:"demo,omitzero"`
}

// DemoCredentials are the credentials a public demo instance prefills its
// login form with.
type DemoCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SetupRequest is the body of POST /api/v1/auth/setup:
// the first administrator's account details.
type SetupRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitzero"`
	Password    string `json:"password"`
}

// AccountUpdateRequest is the body of PATCH /api/v1/auth/me: the signed-in
// user changing their own account. CurrentPassword is always required;
// nil fields stay unchanged.
type AccountUpdateRequest struct {
	CurrentPassword string  `json:"current_password"`
	DisplayName     *string `json:"display_name,omitzero"`
	Email           *string `json:"email,omitzero"`
	NewPassword     *string `json:"new_password,omitzero"`
}

// PreferencesUpdateRequest is the body of PUT /api/v1/auth/me/preferences: the
// signed-in user's dashboard preferences. An omitted field inherits the
// built-in default; an empty list is an explicit "none", e.g. hide nothing.
type PreferencesUpdateRequest struct {
	Hide       []string `json:"hide,omitzero"`
	View       string   `json:"view,omitzero"`
	Timestamps []string `json:"timestamps,omitzero"`
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
