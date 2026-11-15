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
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestAPI_HTTPUserList(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	authCtx := adminContext(t, api, deps)

	req := authedRequest(http.MethodGet, "/api/v1/users", "", adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"admin listing users: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var users []auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &users); err != nil {
		t.Fatalf("admin listing users: decode.Unmarshal() failed: %v", err)
	}
	usernames := test.AttributesSorted(users, func(u auth.User) string { return u.Username })
	if testErr := test.AssertSlicesEqual(
		t, usernames, []string{"admin", "viewer"}, "User.Username",
	); testErr != nil {
		t.Errorf("admin listing users: %v", testErr)
	}

	_ = dbConn.Close()
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req = withAuthCtx(req, authCtx)
	api.httpUserList(w, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"broken store: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPUserCreate(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")

	req := authedRequest(
		http.MethodPost, "/api/v1/users",
		test.TrimJSON(`{
			"username": "fresher",
			"password": "fresher-password",
			"display_name": "New User",
			"groups": ["operator"]
		}`),
		adminCookie,
	)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"admin creating user: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var created auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
		t.Fatalf("admin creating user: decode.Unmarshal() failed: %v", err)
	}
	wantUsername, wantGroups := "fresher", []string{store.GroupOperator}
	if created.Username != wantUsername || !slices.Equal(created.Groups, wantGroups) {
		t.Errorf(
			"admin creating user: User mismatch\n"+
				"got:  Username=%q, Groups=%q\nwant: Username=%q, Groups=%q",
			created.Username, created.Groups,
			wantUsername, wantGroups,
		)
	}

	_ = loginCookie(t, api, "fresher", "fresher-password")

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "duplicate username",
			body:       `{"username":"fresher","password":"fresher-password"}`,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "unknown group",
			body:       `{"username":"grouper","password":"grouper-password","groups":["no-such-group"]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short password",
			body:       `{"username":"shorty","password":"short"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			body:       `{"username":"nopass"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing username",
			body:       `{"password":"has-a-password"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed body",
			body:       `{"username":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "oversize body",
			body:       `{"username":"` + strings.Repeat("x", maxAuthBodySize) + `"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares hashPassword, which the parent swaps after the loop.

			req := authedRequest(http.MethodPost, "/api/v1/users", tc.body, adminCookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}

	hashPasswordHad := hashPassword
	hashPassword = func(_ string) (string, error) {
		return "", errors.New("hash broke")
	}
	t.Cleanup(func() { hashPassword = hashPasswordHad })
	req = authedRequest(
		http.MethodPost, "/api/v1/users",
		`{"username":"hashless","password":"valid-password"}`, adminCookie,
	)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"hash failure: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPUserCreate__fieldBounds(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")

	tooLong := strings.Repeat("x", maxFieldLength+1)
	atLimit := strings.Repeat("y", maxFieldLength)
	atLimit2 := strings.Repeat("z", maxFieldLength)

	tests := []struct {
		name         string
		body         string
		wantStatus   int
		wantUsername string // The username the account is stored under ("" to skip).
	}{
		{
			name:         "valid/surrounding whitespace is trimmed",
			body:         `{"username":"  spaced  ","password":"a-good-password"}`,
			wantStatus:   http.StatusCreated,
			wantUsername: "spaced",
		},
		{
			name:         "valid/username at the limit",
			body:         `{"username":"` + atLimit + `","password":"a-good-password"}`,
			wantStatus:   http.StatusCreated,
			wantUsername: atLimit,
		},
		{
			name:         "valid/username at the limit after whitespace trimmed",
			body:         `{"username":"` + atLimit2 + ` ","password":"a-good-password"}`,
			wantStatus:   http.StatusCreated,
			wantUsername: atLimit2,
		},
		{
			name:       "invalid/username over the limit",
			body:       `{"username":"` + tooLong + `","password":"a-good-password"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid/whitespace-only username",
			body:       `{"username":"   ","password":"a-good-password"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid/display_name over the limit",
			body:       `{"username":"dn","password":"a-good-password","display_name":"` + tooLong + `"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid/email over the limit",
			body:       `{"username":"em","password":"a-good-password","email":"` + tooLong + `"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := authedRequest(http.MethodPost, "/api/v1/users", tc.body, adminCookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Fatalf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			if tc.wantUsername == "" {
				return
			}

			var created auth.User
			if err := decode.Unmarshal("json", w.Body.Bytes(), &created); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			if got, want := created.Username, tc.wantUsername; got != want {
				t.Errorf("User.Username mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

func TestAPI_HTTPUserGet(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	target := createAuthUser(t, deps, "target", "target-password", store.GroupViewer)

	req := authedRequest(http.MethodGet, "/api/v1/users/"+target.ID, "", adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"get user by ID: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var got auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &got); err != nil {
		t.Fatalf("get user by ID: decode.Unmarshal() failed: %v", err)
	}
	if want := "target"; got.Username != want {
		t.Errorf("get user by ID: User.Username mismatch\ngot:  %q\nwant: %q", got.Username, want)
	}

	req = authedRequest(http.MethodGet, "/api/v1/users/no-such-id", "", adminCookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf(
			"get unknown user ID: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_HTTPUserUpdate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	rotated := createAuthUser(t, deps, "rotated", "rotated-password", store.GroupOperator)
	rotatedCookie := loginCookie(t, api, "rotated", "rotated-password")
	disabled := createAuthUser(t, deps, "disabled", "disabled-password")
	disabledCookie := loginCookie(t, api, "disabled", "disabled-password")
	unrevoked := createAuthUser(t, deps, "unrevoked", "unrevoked-password")
	authCtx := adminContext(t, api, deps)

	req := authedRequest(
		http.MethodPatch, "/api/v1/users/"+rotated.ID,
		`{"display_name":"Renamed","groups":["viewer"]}`, adminCookie,
	)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"admin patching user: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var patched auth.User
	if err := decode.Unmarshal("json", w.Body.Bytes(), &patched); err != nil {
		t.Fatalf("admin patching user: decode.Unmarshal() failed: %v", err)
	}
	wantDisplayName, wantGroups := "Renamed", []string{store.GroupViewer}
	if patched.DisplayName != wantDisplayName || !slices.Equal(patched.Groups, wantGroups) {
		t.Errorf(
			"admin patching user: User mismatch\n"+
				"got:  DisplayName=%q, Groups=%q\nwant: DisplayName=%q, Groups=%q",
			patched.DisplayName, patched.Groups,
			wantDisplayName, wantGroups,
		)
	}

	req = authedRequest(
		http.MethodPatch, "/api/v1/users/"+rotated.ID,
		`{"password":"rotated-password-2"}`, adminCookie,
	)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"password patch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", rotatedCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"password patch: GET /auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	_ = loginCookie(t, api, "rotated", "rotated-password-2")

	req = authedRequest(
		http.MethodPatch, "/api/v1/users/"+disabled.ID, `{"enabled":false}`, adminCookie,
	)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"disable patch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", disabledCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"disable patch: GET /auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

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
			target:     authCtx.User.ID,
			body:       `{"enabled":false}`,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "display_name over the limit",
			target:     rotated.ID,
			body:       `{"display_name":"` + strings.Repeat("x", maxFieldLength+1) + `"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares deps.Sessions, which the parent swaps after the loop.

			req := authedRequest(http.MethodPatch, "/api/v1/users/"+tc.target, tc.body, adminCookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}

	hashPasswordHad := hashPassword
	t.Cleanup(func() { hashPassword = hashPasswordHad })
	hashPassword = func(_ string) (string, error) {
		return "", errors.New("hash broke")
	}
	req = authedRequest(
		http.MethodPatch, "/api/v1/users/"+rotated.ID, `{"password":"valid-password"}`, adminCookie,
	)
	w = serveAuth(api, req)
	hashPassword = hashPasswordHad

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"hash failure: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)
	req = httptest.NewRequest(
		http.MethodPatch, "/api/v1/users/"+unrevoked.ID,
		strings.NewReader(`{"password":"rotated-password"}`),
	)
	req = mux.SetURLVars(withAuthCtx(req, authCtx), map[string]string{"id": unrevoked.ID})
	w = httptest.NewRecorder()
	releaseStdout := test.CaptureLog(t, logx.Default())
	api.httpUserUpdate(w, req)
	stdout := releaseStdout()

	// Revocation failures are only logged.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf(
			"patch with failing revoke: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	wantStdout := `^ERROR: httpUserUpdate \(.*\), revoke user's sessions:\s+` + errSessionStore.Error()
	if !util.RegexCheck(wantStdout, stdout) {
		t.Errorf(
			"patch with failing revoke: stdout mismatch\ngot:  %q\nwant: %q",
			stdout, wantStdout,
		)
	}

	creds, err := deps.Store.LocalCredentials(t.Context(), "unrevoked")
	if err != nil {
		t.Fatalf("patch with failing revoke: Store.LocalCredentials() failed: %v", err)
	}
	if creds == nil {
		t.Fatal(
			"patch with failing revoke: Store.LocalCredentials() mismatch\n" +
				"got:  nil\nwant: non-nil",
		)
	}
	match, _, err := password.Verify("rotated-password", creds.PasswordHash)
	if err != nil || !match {
		t.Errorf(
			"patch with failing revoke: password.Verify() mismatch\n"+
				"got:  match=%t, err=%v\nwant: match=true, err=nil",
			match, err,
		)
	}
}

func TestAPI_HTTPUserUpdate__kicks(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	member := createAuthUser(t, deps, "member", "member-password", store.GroupOperator)
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	adminClient := connect(adminID, "admin-session")
	memberClient := connect(member.ID, "member-session")

	req := authedRequest(
		http.MethodPatch, "/api/v1/users/"+member.ID, `{"display_name":"Renamed"}`, adminCookie,
	)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"display-name patch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	// No grants changed, so nobody is kicked.
	memberKept := api.hub.hasClient(memberClient)
	adminKept := api.hub.hasClient(adminClient)
	if !memberKept || !adminKept {
		t.Errorf(
			"display-name patch: Hub.hasClient() mismatch\n"+
				"got:  member=%t, admin=%t\nwant: member=true, admin=true",
			memberKept, adminKept,
		)
	}

	req = authedRequest(
		http.MethodPatch, "/api/v1/users/"+member.ID, `{"groups":["viewer"]}`, adminCookie,
	)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"groups patch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	memberKept = api.hub.hasClient(memberClient)
	adminKept = api.hub.hasClient(adminClient)
	if memberKept || !adminKept {
		t.Errorf(
			"groups patch: Hub.hasClient() mismatch\n"+
				"got:  member=%t, admin=%t\nwant: member=false, admin=true",
			memberKept, adminKept,
		)
	}
}

func TestAPI_HTTPUserDelete(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCookie := loginCookie(t, api, "admin", "admin-password")
	_ = createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")
	victim := createAuthUser(t, deps, "victim", "victim-password")
	victimCookie := loginCookie(t, api, "victim", "victim-password")
	unrevoked := createAuthUser(t, deps, "unrevoked", "unrevoked-password")
	authCtx := adminContext(t, api, deps)

	req := authedRequest(http.MethodDelete, "/api/v1/users/"+victim.ID, "", viewerCookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf(
			"viewer deleting user: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodDelete, "/api/v1/users/no-such-id", "", adminCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf(
			"admin deleting unknown user ID: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodDelete, "/api/v1/users/"+authCtx.User.ID, "", adminCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusConflict; got != want {
		t.Errorf(
			"admin deleting last admin: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodDelete, "/api/v1/users/"+victim.ID, "", adminCookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"admin deleting user: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	if _, err := deps.Store.UserByID(t.Context(), victim.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf(
			"admin deleting user: Store.UserByID() error type mismatch\n"+
				"got:  %v\nwant: %v (deleted)",
			err, store.ErrNotFound,
		)
	}
	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", victimCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"admin deleting user: GET /auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+unrevoked.ID, nil)
	req = mux.SetURLVars(withAuthCtx(req, authCtx), map[string]string{"id": unrevoked.ID})
	w = httptest.NewRecorder()
	releaseStdout := test.CaptureLog(t, logx.Default())
	api.httpUserDelete(w, req)
	stdout := releaseStdout()

	// Revocation failures are only logged.
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Errorf(
			"delete with failing revoke: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	wantStdout := `^ERROR: httpUserDelete \(.*\), revoke user's sessions:\s+` + errSessionStore.Error()
	if !util.RegexCheck(wantStdout, stdout) {
		t.Errorf(
			"delete with failing revoke: stdout mismatch\ngot:  %q\nwant: %q",
			stdout, wantStdout,
		)
	}

	_, err := deps.Store.UserByID(t.Context(), unrevoked.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(
			"delete with failing revoke: Store.UserByID() error type mismatch\n"+
				"got:  %v\nwant: %v (deleted)",
			err, store.ErrNotFound,
		)
	}
}

func TestAPI_HTTPUserDelete__kicks(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	doomed := createAuthUser(t, deps, "doomed", "doomed-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	adminClient := connect(adminID, "admin-session")
	doomedClient := connect(doomed.ID, "doomed-session")

	req := authedRequest(http.MethodDelete, "/api/v1/users/"+doomed.ID, "", adminCookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	doomedKept := api.hub.hasClient(doomedClient)
	adminKept := api.hub.hasClient(adminClient)
	if doomedKept || !adminKept {
		t.Errorf(
			"Hub.hasClient() mismatch\ngot:  doomed=%t, admin=%t\nwant: doomed=false, admin=true",
			doomedKept, adminKept,
		)
	}
}

func TestAPI_HTTPUserDelete__cascadeFailure(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	victim := createAuthUser(t, deps, "victim", "victim-password")

	// A table the delete cascade touches.
	test.DropTable(t, dbConn, "api_tokens")

	req := authedRequest(http.MethodDelete, "/api/v1/users/"+victim.ID, "", adminCookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}
