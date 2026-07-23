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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

func TestAPI_HTTPUserList(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPUserList.yml"
	api, deps, dbConn := testAuthServer(t, file)
	// AND: a logged-in admin and a second user.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	authCtx := adminContext(t, api, deps)

	// WHEN: the admin lists users.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/users", "", adminCookie))

	prefix := fmt.Sprintf("%s\nhttpUserList()", packageName)

	// THEN: every user is returned.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var users []auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &users); err != nil {
		t.Fatalf(
			"%s\nparse list: %v",
			prefix, err,
		)
	}
	if len(users) != 2 { // admin, viewer.
		t.Errorf(
			"%s\nuser count mismatch\ngot:  %d\nwant: 2",
			prefix, len(users),
		)
	}

	// WHEN: the store breaks beneath a direct call.
	_ = dbConn.Close()
	recorder := httptest.NewRecorder()
	api.httpUserList(recorder,
		withAuthCtx(httptest.NewRequest(http.MethodGet, "/api/v1/users", nil), authCtx),
	)

	// THEN: the failure reads as 500.
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nbroken store\ngot:  %d\nwant: 500",
			prefix, recorder.Code,
		)
	}
}

func TestAPI_HTTPUserCreate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPUserCreate.yml"
	api, _, _ := testAuthServer(t, file)
	// AND: a logged-in admin.
	adminCookie := loginCookie(t, api, "admin", "admin-password")

	// WHEN: the admin makes a request to create a user.
	w := serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/users",
			test.TrimJSON(`{
				"username": "fresher",
				"password": "fresher-password",
				"display_name": "New User",
				"groups": ["operator"]
			}`),
			adminCookie,
		),
	)

	prefix := fmt.Sprintf("%s\nhttpUserCreate()", packageName)

	// THEN: the user is created.
	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"%s\ncreate status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var created auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
		t.Fatalf(
			"%s\nparse create response: %v",
			prefix, err,
		)
	}
	if created.Username != "fresher" ||
		len(created.Groups) != 1 || created.Groups[0] != store.GroupOperator {
		t.Errorf(
			"%s\ncreated user mismatch\ngot: %+v",
			prefix, created,
		)
	}

	// AND: the new user can log in.
	_ = loginCookie(t, api, "fresher", "fresher-password")

	// WHEN: an invalid create arrives.
	tests := []struct {
		name string
		body string
	}{
		{
			name: "duplicate username",
			body: `{"username":"fresher","password":"fresher-password"}`,
		},
		{
			name: "unknown group",
			body: `{"username":"grouper","password":"grouper-password","groups":["no-such-group"]}`,
		},
		{
			name: "short password",
			body: `{"username":"shorty","password":"short"}`,
		},
		{
			name: "missing password",
			body: `{"username":"nopass"}`,
		},
		{
			name: "missing username",
			body: `{"password":"has-a-password"}`,
		},
		{
			name: "malformed body",
			body: `{"username":`,
		},
		{
			name: "oversize body",
			body: `{"username":"` + strings.Repeat("x", maxAuthBodySize) + `"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := serveAuth(api,
				authedRequest(http.MethodPost, "/api/v1/users",
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

	// WHEN: password hashing fails.
	hashPasswordHad := hashPassword
	hashPassword = func(_ string) (string, error) {
		return "", errors.New("hash broke")
	}
	t.Cleanup(func() { hashPassword = hashPasswordHad })
	w = serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/users",
			`{"username":"hashless","password":"valid-password"}`,
			adminCookie,
		),
	)

	// THEN: the failure reads as 500.
	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nhash failure\ngot:  %d\nwant: 500",
			prefix, w.Code,
		)
	}
}

func TestAPI_HTTPUserGet(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPUserGet.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin and a target user.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	target := createAuthUser(t, deps, "target", "target-password", store.GroupViewer)

	prefix := fmt.Sprintf("%s\nhttpUserGet()", packageName)

	// WHEN: the admin fetches the user by ID.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/users/"+target.ID, "", adminCookie))

	// THEN: the user is returned.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var got auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &got); err != nil {
		t.Fatalf(
			"%s\nparse response: %v",
			prefix, err,
		)
	}
	if got.Username != "target" {
		t.Errorf(
			"%s\nuser mismatch\ngot: %+v",
			prefix, got,
		)
	}

	// WHEN: an unknown ID is fetched.
	w = serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/users/no-such-id", "", adminCookie))

	// THEN: 404.
	if w.Code != http.StatusNotFound {
		t.Errorf(
			"%s\nget unknown\ngot:  %d\nwant: 404",
			prefix, w.Code,
		)
	}
}

func TestAPI_HTTPUserUpdate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPUserUpdate.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin and logged-in target users.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	rotated := createAuthUser(t, deps, "rotated", "rotated-password", store.GroupOperator)
	rotatedCookie := loginCookie(t, api, "rotated", "rotated-password")
	disabled := createAuthUser(t, deps, "disabled", "disabled-password")
	disabledCookie := loginCookie(t, api, "disabled", "disabled-password")
	unrevoked := createAuthUser(t, deps, "unrevoked", "unrevoked-password")
	authCtx := adminContext(t, api, deps)
	creds, err := deps.Store.LocalCredentials(t.Context(), "admin")
	if err != nil || creds == nil {
		t.Fatalf(
			"%s\nadmin lookup failed: %v",
			packageName, err,
		)
	}

	// WHEN: the admin patches the user's display name and groups.
	w := serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/users/"+rotated.ID,
			`{"display_name":"Renamed","groups":["viewer"]}`,
			adminCookie,
		),
	)

	prefix := fmt.Sprintf("%s\nhttpUserUpdate()", packageName)

	// THEN: the patch is applied.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\npatch status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var patched auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &patched); err != nil {
		t.Fatalf(
			"%s\nparse patch response: %v",
			prefix, err,
		)
	}
	if patched.DisplayName != "Renamed" ||
		len(patched.Groups) != 1 || patched.Groups[0] != store.GroupViewer {
		t.Errorf(
			"%s\npatched user mismatch\ngot: %+v",
			prefix, patched,
		)
	}

	// WHEN: the admin resets the user's password.
	w = serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/users/"+rotated.ID,
			`{"password":"rotated-password-2"}`,
			adminCookie,
		),
	)

	// THEN: the patch succeeds and the user's sessions are revoked.
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\npassword patch\ngot:  %d\nwant: 200",
			prefix, w.Code,
		)
	}
	if w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", rotatedCookie),
	); w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nsessions should be revoked on password change\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}
	_ = loginCookie(t, api, "rotated", "rotated-password-2")

	// WHEN: the admin disables a user.
	w = serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/users/"+disabled.ID,
			`{"enabled":false}`,
			adminCookie,
		),
	)

	// THEN: the patch succeeds and their sessions are revoked.
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\ndisable patch\ngot:  %d\nwant: 200",
			prefix, w.Code,
		)
	}
	if w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", disabledCookie),
	); w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nsessions should be revoked on disable\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}

	// WHEN: an invalid patch arrives.
	tests := []struct {
		name       string
		target     string
		body       string
		wantStatus int
	}{
		{
			name:       "empty password",
			target:     rotated.ID,
			body:       `{"password":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed body",
			target:     rotated.ID,
			body:       `{"display_name":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown user",
			target:     "no-such-id",
			body:       `{"display_name":"x"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "disable the last admin",
			target:     creds.UserID,
			body:       `{"enabled":false}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := serveAuth(api,
				authedRequest(http.MethodPatch, "/api/v1/users/"+tc.target,
					tc.body,
					adminCookie,
				),
			)

			// THEN: it is rejected with the expected status.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\n%s\ngot:  %d - %s\nwant: %d",
					prefix, tc.name, w.Code, w.Body.String(), tc.wantStatus,
				)
			}
		})
	}

	// WHEN: password hashing fails.
	hashPasswordHad := hashPassword
	hashPassword = func(_ string) (string, error) {
		return "", errors.New("hash broke")
	}
	w = serveAuth(api,
		authedRequest(http.MethodPatch, "/api/v1/users/"+rotated.ID,
			`{"password":"valid-password"}`,
			adminCookie,
		),
	)
	hashPassword = hashPasswordHad

	// THEN: the failure reads as 500.
	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nhash failure\ngot:  %d\nwant: 500",
			prefix, w.Code,
		)
	}

	// WHEN: a password is rotated but session revocation fails.
	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+unrevoked.ID,
		strings.NewReader(`{"password":"rotated-password"}`),
	)
	req = mux.SetURLVars(withAuthCtx(req, authCtx), map[string]string{"id": unrevoked.ID})
	recorder := httptest.NewRecorder()
	api.httpUserUpdate(recorder, req)

	// THEN: the update still succeeds (revocation failures are only logged).
	if recorder.Code != http.StatusOK {
		t.Errorf(
			"%s\npatch with failing revoke\ngot:  %d\nwant: 200",
			prefix, recorder.Code,
		)
	}
}

func TestAPI_HTTPUserDelete(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_HTTPUserDelete.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a logged-in admin, viewer, and targets.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	_ = createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")
	victim := createAuthUser(t, deps, "victim", "victim-password")
	victimCookie := loginCookie(t, api, "victim", "victim-password")
	unrevoked := createAuthUser(t, deps, "unrevoked", "unrevoked-password")
	authCtx := adminContext(t, api, deps)
	creds, err := deps.Store.LocalCredentials(t.Context(), "admin")
	if err != nil || creds == nil {
		t.Fatalf(
			"%s\nadmin lookup failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nhttpUserDelete()", packageName)

	// WHEN/THEN: a viewer trying to delete a user 403s.
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/users/"+victim.ID,
			"",
			viewerCookie,
		),
	); w.Code != http.StatusForbidden {
		t.Errorf(
			"%s\nviewer delete\ngot:  %d\nwant: 403",
			prefix, w.Code,
		)
	}

	// AND: deletes of unknown IDs 404.
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/users/no-such-id",
			"",
			adminCookie,
		),
	); w.Code != http.StatusNotFound {
		t.Errorf(
			"%s\ndelete unknown\ngot:  %d\nwant: 404",
			prefix, w.Code,
		)
	}

	// AND: deleting the last admin is rejected with 409.
	if w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/users/"+creds.UserID,
			"",
			adminCookie,
		),
	); w.Code != http.StatusConflict {
		t.Errorf(
			"%s\ndelete last admin\ngot:  %d\nwant: 409",
			prefix, w.Code,
		)
	}

	// WHEN: the admin deletes a user.
	w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/users/"+victim.ID, "", adminCookie))

	// THEN: the user and their sessions are gone.
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"%s\ndelete status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, want, w.Body.String(), want,
		)
	}
	if _, err := deps.Store.UserByID(t.Context(), victim.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf(
			"%s\nuser should be deleted\ngot: %v",
			prefix, err,
		)
	}
	if w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me",
			"",
			victimCookie,
		),
	); w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nsessions should be revoked on delete\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}

	// WHEN: a user is deleted but session revocation fails.
	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+unrevoked.ID, nil)
	req = mux.SetURLVars(withAuthCtx(req, authCtx), map[string]string{"id": unrevoked.ID})
	recorder := httptest.NewRecorder()
	api.httpUserDelete(recorder, req)

	// THEN: the delete still succeeds (revocation failures are only logged).
	if recorder.Code != http.StatusNoContent {
		t.Errorf(
			"%s\ndelete with failing revoke\ngot:  %d\nwant: 204",
			prefix, recorder.Code,
		)
	}
}

func TestAPI_HTTPUserDelete__cascadeFailure(t *testing.T) {
	// GIVEN: an auth-enabled API with a user to delete.
	file := "TestAPI_HTTPUserDelete__cascadeFailure.yml"
	api, deps, dbConn := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	victim := createAuthUser(t, deps, "victim", "victim-password")

	prefix := fmt.Sprintf("%s\nhttpUserDelete() cascade", packageName)

	// AND: a table the delete cascade touches is missing.
	if _, err := dbConn.Exec(`DROP TABLE api_tokens;`); err != nil {
		t.Fatalf(
			"%s\nsetup drop failed: %v",
			prefix, err,
		)
	}

	// WHEN: the admin deletes the user.
	w := serveAuth(api,
		authedRequest(http.MethodDelete, "/api/v1/users/"+victim.ID, "", adminCookie))

	// THEN: the store failure surfaces as 500.
	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\ncascade failure\ngot:  %d\nwant: 500",
			prefix, w.Code,
		)
	}
}
