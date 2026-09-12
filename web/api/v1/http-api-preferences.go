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
	"net/http"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpAuthMePreferencesUpdate handles PUT /api/v1/auth/me/preferences: the
// signed-in user replacing their dashboard preferences. Fields left out
// of the body go back to inheriting the built-in default; an empty list is an
// explicit "none", e.g. hide nothing.
//
// Response:
//
//	200 OK: JSON of the user, their permission grants and the saved preferences.
//	        Setting nothing discards them, removing the row.
//	400 Bad Request: on a malformed or oversized body, or an unknown value.
//	401 Unauthorized: with no authenticated user.
//	500 Internal Server Error: on a store failure.
func (api *API) httpAuthMePreferencesUpdate(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAuthMePreferencesUpdate", Secondary: getIP(r)}

	authCtx := api.authCtxOr401(w, r)
	if authCtx == nil {
		return
	}

	var request apitype.PreferencesUpdateRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}
	// The store validates, and its ErrInvalidPreference maps to 400.
	stored, err := api.auth.Store.SetPreferences(
		r.Context(),
		authCtx.User.ID,
		store.DashboardPreferences{
			Hide:       request.Hide,
			View:       request.View,
			Timestamps: request.Timestamps,
		},
	)
	if err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "save preferences")
		return
	}

	api.writeJSONStatus(w, http.StatusOK, authMe(authCtx, stored), logFrom)
}

// httpAuthMePreferencesDelete handles DELETE /api/v1/auth/me/preferences: the
// signed-in user discarding their dashboard preferences, returning to the
// built-in defaults.
//
// Response:
//
//	200 OK: JSON of the user and their permission grants, with no preferences.
//	401 Unauthorized: with no authenticated user.
//	500 Internal Server Error: on a store failure.
func (api *API) httpAuthMePreferencesDelete(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAuthMePreferencesDelete", Secondary: getIP(r)}

	authCtx := api.authCtxOr401(w, r)
	if authCtx == nil {
		return
	}

	if err := api.auth.Store.DeletePreferences(r.Context(), authCtx.User.ID); err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "discard preferences")
		return
	}

	api.writeJSONStatus(w, http.StatusOK, authMe(authCtx, nil), logFrom)
}
