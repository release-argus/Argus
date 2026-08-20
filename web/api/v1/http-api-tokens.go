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
	"time"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpAPITokenList handles GET /api/v1/tokens: returning info on the requesting
// user's API tokens (hashes and plaintext are never returned).
//
// Response:
//
//	200 OK: JSON array of the user's tokens.
//	401 Unauthorized: with no authenticated user.
//	500 Internal Server Error: on a store failure.
func (api *API) httpAPITokenList(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAPITokenList", Secondary: getIP(r)}

	authCtx := api.authCtxOr401(w, r)
	if authCtx == nil {
		return
	}

	tokens, err := api.auth.Store.APITokensForUser(r.Context(), authCtx.User.ID)
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("failed to list tokens"), http.StatusInternalServerError)
		return
	}

	api.writeJSON(w, tokens, logFrom)
}

// httpAPITokenCreate handles POST /api/v1/tokens: minting an API token for the
// requesting user. The plaintext token appears in this response only.
//
// Response:
//
//	201 Created: JSON of the token, including its one-time plaintext.
//	400 Bad Request: on a missing name or invalid expires_in.
//	401 Unauthorized: with no authenticated user.
//	500 Internal Server Error: on a store failure.
func (api *API) httpAPITokenCreate(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAPITokenCreate", Secondary: getIP(r)}

	authCtx := api.authCtxOr401(w, r)
	if authCtx == nil {
		return
	}

	var request apitype.APITokenCreateRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}
	if request.Name == "" {
		failRequest(&w, errors.New("name is required"), http.StatusBadRequest)
		return
	}

	// Optional expiry.
	var expiresAt *time.Time
	if request.ExpiresIn != "" {
		duration, err := time.ParseDuration(request.ExpiresIn)
		if err != nil || duration <= 0 {
			failRequest(&w,
				&decode.ErrField{
					Key:         "expires_in",
					Value:       request.ExpiresIn,
					Description: "want a positive duration, e.g. '720h'",
				},
				http.StatusBadRequest,
			)
			return
		}
		expires := timeNow().UTC().Add(duration)
		expiresAt = &expires
	}

	plaintext, token, err := api.auth.Store.CreateAPIToken(
		r.Context(),
		authCtx.User.ID,
		request.Name,
		expiresAt,
	)
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("failed to create token"), http.StatusInternalServerError)
		return
	}

	api.writeJSONStatus(w, http.StatusCreated,
		apitype.APITokenCreated{
			APIToken: *token,
			Token:    plaintext,
		},
		logFrom,
	)
}

// httpAPITokenDelete handles DELETE /api/v1/tokens/{id}: revoking one of the
// requesting user's tokens.
//
// Response:
//
//	204 No Content: on success.
//	401 Unauthorized: with no authenticated user.
//	404 Not Found: when the token is unknown or owned by another user.
//	500 Internal Server Error: on a store failure.
func (api *API) httpAPITokenDelete(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAPITokenDelete", Secondary: getIP(r)}

	authCtx := api.authCtxOr401(w, r)
	if authCtx == nil {
		return
	}

	if err := api.auth.Store.DeleteAPIToken(
		r.Context(),
		authCtx.User.ID,
		mux.Vars(r)["id"],
	); err != nil {
		api.failAuthStoreRequest(w, err, logFrom, "delete token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
