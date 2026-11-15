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

//go:build unit

package v1

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestAPI_HTTPAPITokenList(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)
	createAPIToken(t, deps, authCtx.User.ID, "ci", nil)

	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")

	req := authedRequest(http.MethodGet, "/api/v1/tokens", "", adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"admin list: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var tokens []store.APIToken
	if err := decode.Unmarshal("json", w.Body.Bytes(), &tokens); err != nil {
		t.Fatalf("admin list: decode.Unmarshal() failed: %v", err)
	}
	tokenNames := test.AttributesSorted(tokens, func(token store.APIToken) string { return token.Name })
	if testErr := test.AssertSlicesEqual(
		t, tokenNames, []string{"ci"}, "APIToken.Name",
	); testErr != nil {
		t.Errorf("admin list: %v", testErr)
	}

	req = authedRequest(http.MethodGet, "/api/v1/tokens", "", viewerCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"viewer list: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var viewerTokens []store.APIToken
	if err := decode.Unmarshal("json", w.Body.Bytes(), &viewerTokens); err != nil {
		t.Fatalf("viewer list: decode.Unmarshal() failed: %v", err)
	}
	viewerNames := test.AttributesSorted(viewerTokens, func(token store.APIToken) string { return token.Name })
	// Own-scoped, so the admin's token is not listed.
	if testErr := test.AssertSlicesEqual(
		t, viewerNames, []string{}, "APIToken.Name",
	); testErr != nil {
		t.Errorf("viewer list: %v", testErr)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
	api.httpAPITokenList(w, req)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"without auth context: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	_ = dbConn.Close()
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
	req = withAuthCtx(req, authCtx)
	api.httpAPITokenList(w, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"broken store: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPAPITokenCreate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)

	body := `{"name":"ci","expires_in":"720h"}`
	req := authedRequest(http.MethodPost, "/api/v1/tokens", body, adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"create with expiry: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var created apitype.APITokenCreated
	if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
		t.Fatalf("create with expiry: decode.Unmarshal() failed: %v", err)
	}
	if created.Name != "ci" {
		t.Errorf(
			"create with expiry: APITokenCreated.Name mismatch\ngot:  %q\nwant: %q",
			created.Name, "ci",
		)
	}
	if created.ExpiresAt == nil {
		t.Error("create with expiry: APITokenCreated.ExpiresAt mismatch\ngot:  nil\nwant: non-nil")
	}
	if !strings.HasPrefix(created.Token, "argus_") {
		t.Errorf(
			"create with expiry: APITokenCreated.Token prefix mismatch\ngot:  %q\nwant: %q",
			created.Token, "argus_",
		)
	}
	if created.Prefix == "" {
		t.Error("create with expiry: APITokenCreated.Prefix mismatch\ngot:  \"\"\nwant: non-empty")
	} else if !strings.HasPrefix(created.Token, created.Prefix) {
		t.Errorf(
			"create with expiry: APITokenCreated.Prefix mismatch\ngot:  %q\nwant: %q",
			created.Prefix, created.Token[:min(len(created.Token), len(created.Prefix))],
		)
	}

	if got, want := w.Header().Get("Content-Type"), "application/json; charset=utf-8"; got != want {
		t.Errorf("create with expiry: Content-Type header mismatch\ngot:  %q\nwant: %q", got, want)
	}
	if got, want := w.Header().Get("X-Content-Type-Options"), "nosniff"; got != want {
		t.Errorf(
			"create with expiry: X-Content-Type-Options header mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}

	body = `{"name":"forever"}`
	req = authedRequest(http.MethodPost, "/api/v1/tokens", body, adminCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"no-expiry create: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var forever apitype.APITokenCreated
	if err := decode.Unmarshal("json", w.Body.Bytes(), &forever); err != nil {
		t.Fatalf("no-expiry create: decode.Unmarshal() failed: %v", err)
	}
	if forever.ExpiresAt != nil {
		t.Errorf(
			"no-expiry create: APITokenCreated.ExpiresAt mismatch\ngot:  %v\nwant: nil",
			*forever.ExpiresAt,
		)
	}

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing name",
			body: `{"expires_in":"1h"}`,
		},
		{
			name: "invalid expires_in",
			body: `{"name":"x","expires_in":"soon"}`,
		},
		{
			name: "negative expires_in",
			body: `{"name":"x","expires_in":"-1h"}`,
		},
		{
			name: "malformed body",
			body: `{"name":`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares dbConn, which the parent closes after the loop.

			req := authedRequest(http.MethodPost, "/api/v1/tokens", tc.body, adminCookie)
			w := serveAuth(api, req)

			if got, want := w.Code, http.StatusBadRequest; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}

	var capped *httptest.ResponseRecorder
	for i := range 21 {
		body = fmt.Sprintf(`{"name":"cap-%d"}`, i)
		req = authedRequest(http.MethodPost, "/api/v1/tokens", body, adminCookie)
		capped = serveAuth(api, req)
		if capped.Code != http.StatusCreated {
			break
		}
	}

	if got, want := capped.Code, http.StatusConflict; got != want {
		t.Errorf(
			"over-cap create: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, capped.Body.String(), want,
		)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/tokens", strings.NewReader(`{"name":"x"}`))
	api.httpAPITokenCreate(w, req)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"without auth context: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	_ = dbConn.Close()
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/tokens", strings.NewReader(`{"name":"x"}`))
	req = withAuthCtx(req, authCtx)
	api.httpAPITokenCreate(w, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"broken store: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPAPITokenDelete(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)
	_, adminToken := createAPIToken(t, deps, authCtx.User.ID, "ci", nil)

	viewer1 := createAuthUser(t, deps, "viewer-1", "viewer-password", store.GroupViewer)
	viewer1Cookie := loginCookie(t, api, "viewer-1", "viewer-password")
	_, viewer1Token := createAPIToken(t, deps, viewer1.ID, "ci", nil)

	createAuthUser(t, deps, "viewer-2", "viewer-password", store.GroupViewer)
	viewer2Cookie := loginCookie(t, api, "viewer-2", "viewer-password")

	// 404, not 403: the lookup is scoped to the caller's own tokens.
	req := authedRequest(http.MethodDelete, "/api/v1/tokens/"+adminToken.ID, "", viewer1Cookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf(
			"cross-user (viewer-admin) delete: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodDelete, "/api/v1/tokens/"+viewer1Token.ID, "", viewer2Cookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf(
			"cross-user (viewer-viewer) delete: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	owners := []struct{ name, userID string }{
		{name: "admin", userID: authCtx.User.ID},
		{name: "viewer-1", userID: viewer1.ID},
	}
	for _, owner := range owners {
		t.Run(owner.name, func(t *testing.T) {
			// Serial: shares deps.Store, whose admin token a later step deletes.

			tokens, err := deps.Store.APITokensForUser(t.Context(), owner.userID)
			if err != nil {
				t.Fatalf("after refused deletes: Store.APITokensForUser() failed: %v", err)
			}
			if got, want := len(tokens), 1; got != want {
				t.Errorf(
					"after refused deletes: Store.APITokensForUser() length mismatch\n"+
						"got:  %d\nwant: %d",
					got, want,
				)
			}
		})
	}

	req = authedRequest(http.MethodDelete, "/api/v1/tokens/no-such-id", "", adminCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf(
			"delete unknown token: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodDelete, "/api/v1/tokens/"+adminToken.ID, "", adminCookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"delete own token: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	tokens, err := deps.Store.APITokensForUser(t.Context(), authCtx.User.ID)
	if err != nil {
		t.Fatalf("delete own token: Store.APITokensForUser() failed: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf(
			"delete own token: Store.APITokensForUser() length mismatch\n"+
				"got:  %d\nwant: 0 (deleted)",
			len(tokens),
		)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/x", nil)
	api.httpAPITokenDelete(w, req)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"without auth context: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPAPIToken__bearerCannotManageTokens(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)
	plaintext, token := createAPIToken(t, deps, authCtx.User.ID, "ci", nil)

	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
	}{
		{
			name:       "valid/listing own tokens is allowed",
			method:     http.MethodGet,
			target:     "/api/v1/tokens",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid/minting a replacement is refused",
			method:     http.MethodPost,
			target:     "/api/v1/tokens",
			body:       `{"name":"minted-by-token"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid/revoking itself is refused",
			method:     http.MethodDelete,
			target:     "/api/v1/tokens/" + token.ID,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares token, which the DELETE row revokes.

			req := authedRequest(tc.method, tc.target, tc.body, nil)
			req.Header.Set("Authorization", "Bearer "+plaintext)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"Bearer token request: status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			cookie := loginCookie(t, api, "admin", "admin-password")
			req = authedRequest(tc.method, tc.target, tc.body, cookie)
			w = serveAuth(api, req)

			wantStatus := http.StatusOK
			switch tc.method {
			case http.MethodPost:
				wantStatus = http.StatusCreated
			case http.MethodDelete:
				wantStatus = http.StatusNoContent
			}
			if got, want := w.Code, wantStatus; got != want {
				t.Errorf(
					"browser session request: status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}
