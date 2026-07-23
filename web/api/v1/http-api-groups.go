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
	"slices"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// callerIsAdmin reports whether the request's authenticated user belongs to the
// admin group.
func callerIsAdmin(r *http.Request) bool {
	authCtx := authContextFrom(r)
	return authCtx != nil && slices.Contains(authCtx.User.Groups, store.GroupAdmin)
}

// httpGroupList handles GET /api/v1/groups: returning a list of all
// [store.Group]s (with member counts and grants).
//
// Response:
//
//	200 OK: JSON array of the groups.
//	500 Internal Server Error: on a store failure.
func (api *API) httpGroupList(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpGroupList", Secondary: getIP(r)}

	groups, err := api.auth.Store.Groups(r.Context())
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("failed to list groups"), http.StatusInternalServerError)
		return
	}

	api.writeJSON(w, groups, logFrom)
}

// httpGroupCreate handles POST /api/v1/groups: creating a new [store.Group]
// and registering it with the DB. Setting grants is admin-only.
//
// Response:
//
//	201 Created: JSON of the created group.
//	400 Bad Request: on a missing name, malformed body, duplicate name, or invalid grant.
//	500 Internal Server Error: on a store failure.
func (api *API) httpGroupCreate(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpGroupCreate", Secondary: getIP(r)}

	var request apitype.GroupCreateRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}
	if request.Name == "" {
		failRequest(&w, errors.New("name is required"), http.StatusBadRequest)
		return
	}

	group, err := api.auth.Store.CreateGroup(
		r.Context(),
		request.Name, request.Description, request.Permissions,
	)
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "create group")
		return
	}

	api.writeJSONStatus(w, http.StatusCreated, group, logFrom)
}

// httpGroupGet handles GET /api/v1/groups/{id}: returning a JSON-encoded
// [store.Group].
//
// Response:
//
//	200 OK: JSON of the group.
//	404 Not Found: on an unknown ID.
//	500 Internal Server Error: on a store failure.
func (api *API) httpGroupGet(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpGroupGet", Secondary: getIP(r)}

	group, err := api.auth.Store.GroupByID(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "get group")
		return
	}

	api.writeJSON(w, group, logFrom)
}

// httpGroupUpdate handles PATCH /api/v1/groups/{id}: a patch to an existing
// [store.Group] (permissions are a replace-set).
//
// Response:
//
//	200 OK: JSON of the updated group.
//	400 Bad Request: on a malformed body, duplicate name, or invalid grant.
//	404 Not Found: on an unknown ID.
//	409 Conflict: when modifying the admin system group.
//	500 Internal Server Error: on a store failure.
func (api *API) httpGroupUpdate(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpGroupUpdate", Secondary: getIP(r)}

	var request apitype.GroupPatchRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}

	groupID := mux.Vars(r)["id"]
	group, err := api.auth.Store.UpdateGroup(r.Context(), groupID,
		store.GroupPatch{
			Name:        request.Name,
			Description: request.Description,
			Grants:      request.Permissions,
		},
	)
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "update group")
		return
	}

	// Grant changes alter the members' permissions.
	// Renames/description edits alter nobody's (grants key on the group ID).
	if request.Permissions != nil {
		api.kickGroupMemberWebSocketClients(r, groupID, logFrom)
	}

	api.writeJSON(w, group, logFrom)
}

// httpGroupDelete handles DELETE /api/v1/groups/{id}: removing a group from the
// DB and all users (kicking the members' WS clients).
//
// Response:
//
//	204 No Content: on success.
//	404 Not Found: on an unknown ID.
//	409 Conflict: when deleting the admin system group.
//	500 Internal Server Error: on a store failure.
func (api *API) httpGroupDelete(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpGroupDelete", Secondary: getIP(r)}
	groupID := mux.Vars(r)["id"]

	// Capture the members before the delete cascades away the membership rows.
	memberIDs, err := api.auth.Store.UserIDsInGroup(r.Context(), groupID)
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "delete group")
		return
	}

	if err := api.auth.Store.DeleteGroup(r.Context(), groupID); err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "delete group")
		return
	}
	api.kickUserWebSocketClients(memberIDs...)

	w.WriteHeader(http.StatusNoContent)
}

// kickGroupMemberWebSocketClients kicks the WebSocket clients of the
// members of the group with groupID.
func (api *API) kickGroupMemberWebSocketClients(r *http.Request, groupID string, logFrom logx.LogFrom) {
	memberIDs, err := api.auth.Store.UserIDsInGroup(r.Context(), groupID)
	if err != nil {
		logx.Error(err, logFrom, true)
		return
	}
	api.kickUserWebSocketClients(memberIDs...)
}

// httpPermissionCatalogue handles GET /api/v1/permissions: returning the
// matrix of valid (resource, action, scope) combinations.
//
// Response:
//
//	200 OK: JSON of the permission catalogue.
func (api *API) httpPermissionCatalogue(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpPermissionCatalogue", Secondary: getIP(r)}

	api.writeJSON(w,
		apitype.PermissionCatalogue{
			Resources: rbac.Catalogue(),
		},
		logFrom,
	)
}
