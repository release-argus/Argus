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

package v1

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// hashPassword derives a password hash (overridable for tests).
// see [password.Hash].
var hashPassword = password.Hash

// httpUserList handles GET /api/v1/users: returning a list of all [auth.User]s
// (with group memberships).
//
// Response:
//
//	200 OK: JSON array of the users.
//	500 Internal Server Error: on a store failure.
func (api *API) httpUserList(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpUserList", Secondary: getIP(r)}

	users, err := api.auth.Store.Users(r.Context())
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("failed to list users"), http.StatusInternalServerError)
		return
	}

	api.writeJSON(w, users, logFrom)
}

// httpUserCreate handles POST /api/v1/users: creating a new [auth.User] and
// registering it with the DB.
//
// Response:
//
//	201 Created: JSON of the created user.
//	400 Bad Request: on a missing username, weak password, malformed body, duplicate username, or unknown group.
//	500 Internal Server Error: on a hashing or store failure.
func (api *API) httpUserCreate(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpUserCreate", Secondary: getIP(r)}

	var request apitype.UserCreateRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}
	if request.Username == "" {
		err := &decode.ErrField{
			Key: "username",
		}
		failRequest(&w, err, http.StatusBadRequest)
		return
	}
	if err := validatePassword(request.Password); err != nil {
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	hash, err := hashPassword(request.Password)
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("failed to create user"), http.StatusInternalServerError)
		return
	}

	user, err := api.auth.Store.CreateUser(
		r.Context(),
		request.Username,
		request.DisplayName,
		request.Email,
		hash,
		request.Groups,
	)
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "create user")
		return
	}

	api.writeJSONStatus(w, http.StatusCreated, user, logFrom)
}

// httpUserGet handles GET /api/v1/users/{id}: returning a JSON-encoded [auth.User].
//
// Response:
//
//	200 OK: JSON of the user.
//	404 Not Found: on an unknown ID.
//	500 Internal Server Error: on a store failure.
func (api *API) httpUserGet(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpUserGet", Secondary: getIP(r)}

	user, err := api.auth.Store.UserByID(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "get user")
		return
	}

	api.writeJSON(w, user, logFrom)
}

// httpUserUpdate handles PATCH /api/v1/users/{id}: patching an existing
// [auth.User]. Setting a password or disabling the user revokes their
// sessions (logout everywhere).
//
// Response:
//
//	200 OK: JSON of the updated user.
//	400 Bad Request: on a weak password or malformed body.
//	404 Not Found: on an unknown ID.
//	409 Conflict: when the patch would disable or demote the last admin.
//	500 Internal Server Error: on a hashing or store failure.
func (api *API) httpUserUpdate(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpUserUpdate", Secondary: getIP(r)}
	userID := mux.Vars(r)["id"]

	var request apitype.UserPatchRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}

	patch := store.UserPatch{
		DisplayName: request.DisplayName,
		Email:       request.Email,
		Enabled:     request.Enabled,
		Groups:      request.Groups,
	}
	if request.Password != nil {
		if err := validatePassword(*request.Password); err != nil {
			failRequest(&w, err, http.StatusBadRequest)
			return
		}
		hash, err := hashPassword(*request.Password)
		if err != nil {
			logx.Error(err, logFrom, true)
			failRequest(&w, errors.New("failed to update user"), http.StatusInternalServerError)
			return
		}
		patch.PasswordHash = &hash
	}

	user, err := api.auth.Store.UpdateUser(r.Context(), userID, patch)
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "update user")
		return
	}

	// Password change or disable ends every session of the user.
	if request.Password != nil ||
		(request.Enabled != nil && !*request.Enabled) {
		if err := api.auth.Sessions.RevokeUser(r.Context(), userID); err != nil {
			logx.Error(err, logFrom, true)
		}
	}

	// Password/enabled/group changes affect only this user - kick just their clients.
	if request.Password != nil || request.Enabled != nil || request.Groups != nil {
		api.kickUserWebSocketClients(userID)
	}

	api.writeJSON(w, user, logFrom)
}

// httpUserDelete handles DELETE /api/v1/users/{id}: removing a user
// (and their sessions) from the DB.
//
// Response:
//
//	204 No Content: on success.
//	404 Not Found: on an unknown ID.
//	409 Conflict: when the target is the last admin.
//	500 Internal Server Error: on a store failure.
func (api *API) httpUserDelete(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpUserDelete", Secondary: getIP(r)}
	userID := mux.Vars(r)["id"]

	if err := api.auth.Store.DeleteUser(r.Context(), userID); err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "delete user")
		return
	}

	if err := api.auth.Sessions.RevokeUser(r.Context(), userID); err != nil {
		logx.Error(err, logFrom, true)
	}

	api.kickUserWebSocketClients(userID)

	w.WriteHeader(http.StatusNoContent)
}
