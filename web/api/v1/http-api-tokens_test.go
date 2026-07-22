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
	apitype "github.com/release-argus/Argus/web/api/types"
)

func TestAPI_HTTPAPITokenList(t *testing.T) {
	prefix := fmt.Sprintf("%s\nhttpAPITokenList()", packageName)

	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPAPITokenList.yml"
	api, deps, dbConn := testAuthServer(t, file)
	// AND: a logged-in admin (owning a token).
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)
	if _, _, err := deps.Store.CreateAPIToken(t.Context(), authCtx.User.ID, "ci", nil); err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}
	// AND: a logged-in viewer (owning none).
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")

	// WHEN: the admin lists tokens.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/tokens", "", adminCookie))

	// THEN: their token is returned.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var tokens []store.APIToken
	if err := decode.Unmarshal("json", w.Body.Bytes(), &tokens); err != nil {
		t.Fatalf(
			"%s\nparse list: %v",
			prefix, err,
		)
	}
	if len(tokens) != 1 || tokens[0].Name != "ci" {
		t.Errorf(
			"%s\ntoken list mismatch\ngot: %+v",
			prefix, tokens,
		)
	}

	// AND: the viewer's list is empty (own-scoped).
	w = serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/tokens", "", viewerCookie))
	var viewerTokens []store.APIToken
	if err := decode.Unmarshal("json", w.Body.Bytes(), &viewerTokens); err != nil {
		t.Fatalf(
			"%s\nparse viewer list: %v",
			prefix, err,
		)
	}
	if len(viewerTokens) != 0 {
		t.Errorf(
			"%s\nviewer should have no tokens\ngot: %+v",
			prefix, viewerTokens,
		)
	}

	// WHEN: a direct call arrives without an auth context.
	recorder := httptest.NewRecorder()
	api.httpAPITokenList(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil))

	// THEN: 401.
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nwithout auth context\ngot:  %d\nwant: 401",
			prefix, recorder.Code,
		)
	}

	// WHEN: the store breaks beneath a direct call.
	_ = dbConn.Close()
	recorder = httptest.NewRecorder()
	api.httpAPITokenList(recorder,
		withAuthCtx(httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil), authCtx))

	// THEN: the failure reads as 500.
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nbroken store\ngot:  %d\nwant: 500",
			prefix, recorder.Code,
		)
	}
}

func TestAPI_HTTPAPITokenCreate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPAPITokenCreate.yml"
	api, deps, dbConn := testAuthServer(t, file)
	// AND: a logged-in admin.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)

	prefix := fmt.Sprintf("%s\nhttpAPITokenCreate()", packageName)

	// WHEN: a token is created with an expiry.
	w := serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/tokens",
			`{"name":"ci","expires_in":"720h"}`,
			adminCookie,
		),
	)

	// THEN: the token is created and the plaintext returned once.
	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"%s\ncreate status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var created apitype.APITokenCreated
	if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
		t.Fatalf(
			"%s\nparse create response: %v",
			prefix, err,
		)
	}
	if !strings.HasPrefix(created.Token, "argus_") ||
		created.Name != "ci" ||
		created.ExpiresAt == nil ||
		!strings.HasPrefix(created.Token, created.Prefix) {
		t.Errorf(
			"%s\ncreated token mismatch\ngot: %+v",
			prefix, created,
		)
	}

	// AND: the JSON headers survive the 201 status line.
	if got, want := w.Header().Get("Content-Type"), "application/json; charset=utf-8"; got != want {
		t.Errorf(
			"%s\ncreate Content-Type mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
	if got, want := w.Header().Get("X-Content-Type-Options"), "nosniff"; got != want {
		t.Errorf(
			"%s\ncreate X-Content-Type-Options mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// AND: a token without an expiry never expires.
	w = serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/tokens",
			`{"name":"forever"}`,
			adminCookie,
		),
	)
	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"%s\nno-expiry create\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var forever apitype.APITokenCreated
	if err := decode.Unmarshal("json", w.Body.Bytes(), &forever); err != nil {
		t.Fatalf(
			"%s\nparse no-expiry response: %v",
			prefix, err,
		)
	}
	if forever.ExpiresAt != nil {
		t.Errorf(
			"%s\nno-expiry token should have a nil ExpiresAt\ngot: %+v",
			prefix, forever,
		)
	}

	// WHEN: an invalid create arrives.
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
			w := serveAuth(api,
				authedRequest(http.MethodPost, "/api/v1/tokens",
					tc.body,
					adminCookie,
				),
			)

			// THEN: it is rejected with 400.
			if w.Code != http.StatusBadRequest {
				t.Errorf(
					"%s\n%s\ngot:  %d - %s\nwant: 400",
					prefix, tc.name, w.Code, w.Body.String(),
				)
			}
		})
	}

	// WHEN: a direct call arrives without an auth context.
	recorder := httptest.NewRecorder()
	api.httpAPITokenCreate(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/tokens",
		strings.NewReader(`{"name":"x"}`)))

	// THEN: 401.
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nwithout auth context\ngot:  %d\nwant: 401",
			prefix, recorder.Code,
		)
	}

	// WHEN: the store breaks beneath a direct call.
	_ = dbConn.Close()
	recorder = httptest.NewRecorder()
	api.httpAPITokenCreate(recorder,
		withAuthCtx(httptest.NewRequest(http.MethodPost, "/api/v1/tokens",
			strings.NewReader(`{"name":"x"}`)), authCtx))

	// THEN: the failure reads as 500.
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nbroken store\ngot:  %d\nwant: 500",
			prefix, recorder.Code,
		)
	}
}

func TestAPI_HTTPAPITokenDelete(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPAPITokenDelete.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin (owning a token).
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	authCtx := adminContext(t, api, deps)
	_, token, err := deps.Store.CreateAPIToken(t.Context(), authCtx.User.ID, "ci", nil)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}
	// AND: a viewer (owning none).
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")

	prefix := fmt.Sprintf("%s\nhttpAPITokenDelete()", packageName)

	// WHEN/THEN: the viewer cannot delete the admin's token (own-scoped).
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/tokens/"+token.ID, "", viewerCookie),
	); w.Code != http.StatusNotFound {
		t.Errorf(
			"%s\ncross-user delete\ngot:  %d\nwant: 404",
			prefix, w.Code,
		)
	}

	// AND: unknown IDs 404.
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/tokens/no-such-id", "", adminCookie),
	); w.Code != http.StatusNotFound {
		t.Errorf(
			"%s\ndelete unknown\ngot:  %d\nwant: 404",
			prefix, w.Code,
		)
	}

	// WHEN: the admin revokes their token.
	w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/tokens/"+token.ID, "", adminCookie))

	// THEN: the token is gone.
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"%s\ndelete status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	tokens, err := deps.Store.APITokensForUser(t.Context(), authCtx.User.ID)
	if err != nil {
		t.Fatalf(
			"%s\nAPITokensForUser failed: %v",
			prefix, err,
		)
	}
	if len(tokens) != 0 {
		t.Errorf(
			"%s\ntoken should be deleted\ngot: %+v",
			prefix, tokens,
		)
	}

	// WHEN: a direct call arrives without an auth context.
	recorder := httptest.NewRecorder()
	api.httpAPITokenDelete(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/x", nil))

	// THEN: 401.
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nwithout auth context\ngot:  %d\nwant: 401",
			prefix, recorder.Code,
		)
	}
}
