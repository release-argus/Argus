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

//go:build unit || integration

package v1

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/store"
)

func TestAPI_Auth__RBAC(t *testing.T) {
	// GIVEN: users in groups of differing power.
	file := "TestAPI_Auth__RBAC.yml"
	api, deps, _ := testAuthServer(t, file)
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	createAuthUser(t, deps, "loner", "loner-password") // No groups.
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	viewerCookie := loginCookie(t, api, "viewer", "viewer-password")
	lonerCookie := loginCookie(t, api, "loner", "loner-password")

	tests := []struct {
		name       string
		method     string
		target     string
		cookie     *http.Cookie
		wantStatus int
	}{
		{
			name:   "no credentials rejected",
			method: http.MethodGet, target: "/api/v1/flags",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "viewer can read flags (config:read)",
			method: http.MethodGet, target: "/api/v1/flags",
			cookie:     viewerCookie,
			wantStatus: http.StatusOK,
		},
		{
			name:   "viewer can read the service order (service:read)",
			method: http.MethodGet, target: "/api/v1/service/order",
			cookie:     viewerCookie,
			wantStatus: http.StatusOK,
		},
		{
			name:   "viewer cannot create users",
			method: http.MethodPost, target: "/api/v1/users",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot list users",
			method: http.MethodGet, target: "/api/v1/users",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot update a user",
			method: http.MethodPatch, target: "/api/v1/users/whatever",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot create groups",
			method: http.MethodPost, target: "/api/v1/groups",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot list groups",
			method: http.MethodGet, target: "/api/v1/groups",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot update a group",
			method: http.MethodPatch, target: "/api/v1/groups/whatever",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot delete a group",
			method: http.MethodDelete, target: "/api/v1/groups/whatever",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot read the permission catalogue",
			method: http.MethodGet, target: "/api/v1/permissions",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "viewer cannot reorder services (service_order:update)",
			method: http.MethodPut, target: "/api/v1/service/order",
			cookie:     viewerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "groupless user can log in but do nothing",
			method: http.MethodGet, target: "/api/v1/flags",
			cookie:     lonerCookie,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "admin can list users",
			method: http.MethodGet, target: "/api/v1/users",
			cookie:     adminCookie,
			wantStatus: http.StatusOK,
		},
		{
			name:   "admin can list groups",
			method: http.MethodGet, target: "/api/v1/groups",
			cookie:     adminCookie,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nRBAC", packageName)

			// WHEN: the route is requested.
			w := serveAuth(api,
				authedRequest(tc.method, tc.target, "", tc.cookie))

			// THEN: the status matches the user's permissions.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, w.Code, w.Body.String(), tc.wantStatus,
				)
			}
		})
	}
}

func TestAPI_Auth__infrastructureFailures(t *testing.T) {
	// GIVEN: a logged-in admin whose auth store then breaks.
	file := "TestAPI_Auth__infrastructureFailures.yml"
	api, _, dbConn := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	// Break the store: every lookup now fails
	_ = dbConn.Close()

	prefix := fmt.Sprintf("%s\ninfrastructure failure", packageName)

	// WHEN: an authenticated request arrives.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/flags", "", cookie))

	// THEN: it reads as a 500, not as an auth failure.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}

	// AND: login reports 500 too (not "invalid credentials").
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`))
	if w := serveAuth(api, req); w.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nlogin during outage\ngot:  %d\nwant: 500",
			prefix, w.Code,
		)
	}
}

func TestAPI_Auth__disabledMidSession(t *testing.T) {
	// GIVEN: a logged-in user who is then disabled behind the API's back.
	file := "TestAPI_Auth__disabledMidSession.yml"
	api, deps, dbConn := testAuthServer(t, file)
	createAuthUser(t, deps, "victim", "victim-password", store.GroupViewer)
	cookie := loginCookie(t, api, "victim", "victim-password")
	if _, err := dbConn.Exec(`UPDATE users SET enabled = 0 WHERE username = 'victim';`); err != nil {
		t.Fatalf(
			"%s\nsetup disable failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\ndisabled mid-session", packageName)

	// WHEN: the still-live session is used.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/flags", "", cookie))

	// THEN: the disabled user reads as unauthorised despite the valid session.
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"%s\nstatus mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestAPI_Auth__originCheck(t *testing.T) {
	// GIVEN: a logged-in admin.
	file := "TestAPI_Auth__originCheck.yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	tests := []struct {
		name       string
		origin     string
		wantStatus int
	}{
		{name: "no Origin header passes", origin: "", wantStatus: http.StatusOK},
		{name: "matching Origin passes", origin: "http://example.com", wantStatus: http.StatusOK},
		{name: "cross-site Origin is rejected", origin: "http://evil.example.net", wantStatus: http.StatusForbidden},
		{name: "malformed Origin is rejected", origin: "::not-a-url::", wantStatus: http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nOrigin check", packageName)

			// WHEN: an authenticated request carries the Origin header.
			req := authedRequest(http.MethodGet, "http://example.com/api/v1/flags", "", cookie)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := serveAuth(api, req)

			// THEN: cross-site origins are rejected.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
		})
	}
}

func TestAPI_Auth__guardDisabled(t *testing.T) {
	// GIVEN: an API without auth enabled.
	file := "TestAPI_Auth__guardDisabled.yml"
	apiValue := testAPI(t, file)
	api := &apiValue

	prefix := fmt.Sprintf("%s\nguard() with auth disabled", packageName)

	// WHEN: a handler is guarded.
	called := false
	handler := api.guard(
		"service",
		"read",
		nil,
		func(_ http.ResponseWriter, _ *http.Request) {
			called = true
		},
	)
	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// THEN: the handler runs unwrapped.
	if !called {
		t.Errorf("%s\nhandler should have been called directly", prefix)
	}

	// AND: GuardMetrics is a pass-through.
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	w := httptest.NewRecorder()
	api.GuardMetrics(inner).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if w.Code != http.StatusTeapot {
		t.Errorf(
			"%s\nGuardMetrics should pass through when auth is disabled\ngot:  %d",
			prefix, w.Code,
		)
	}
}

func TestAPI_Auth__tokenFallback_Metrics(t *testing.T) {
	// GIVEN: an auth-enabled API, a guarded metrics handler, and an API token.
	file := "TestAPI_Auth__tokenFallback_Metrics.yml"
	api, deps, _ := testAuthServer(t, file)
	adminCtx := adminContext(t, api, deps)
	token, _, err := deps.Store.CreateAPIToken(t.Context(), adminCtx.User.ID, "scrape", nil)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("metrics"))
	})
	guarded := api.GuardMetrics(inner)

	tests := []struct {
		name       string
		header     string
		basicAuth  bool
		wantStatus int
	}{
		{
			name:       "valid Bearer token passes",
			header:     "Bearer " + token,
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown Bearer token rejected",
			header:     "Bearer argus_" + strings.Repeat("0", 64),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Basic credentials are not an authentication method",
			basicAuth:  true,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "no credentials rejected",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nGuardMetrics", packageName)

			// WHEN: /metrics is scraped.
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			if tc.basicAuth {
				req.SetBasicAuth("admin", "admin-password")
			}
			w := httptest.NewRecorder()
			guarded.ServeHTTP(w, req)

			// THEN: the status matches expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, w.Code, w.Body.String(), tc.wantStatus,
				)
			}
		})
	}
}

func TestAPI_Auth_GuardMetricsErrors(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_Auth_GuardMetricsErrors.yml"
	api, deps, dbConn := testAuthServer(t, file)
	// AND: a user with no grants.
	loner := createAuthUser(t, deps, "loner", "loner-password")
	lonerToken, _, err := deps.Store.CreateAPIToken(t.Context(), loner.ID, "scrape", nil)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}
	// AND: an admin.
	adminCtx := adminContext(t, api, deps)
	adminToken, _, err := deps.Store.CreateAPIToken(t.Context(), adminCtx.User.ID, "scrape", nil)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}
	guarded := api.GuardMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("metrics"))
	}))

	prefix := fmt.Sprintf("%s\nGuardMetrics errors", packageName)

	// WHEN: an authenticated user without config:read scrapes.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+lonerToken)
	w := httptest.NewRecorder()
	guarded.ServeHTTP(w, req)
	// THEN: 403.
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf(
			"%s\nno permission\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// WHEN: the store breaks beneath the scrape from an authentication admin.
	_ = dbConn.Close()
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	guarded.ServeHTTP(w, req)
	// THEN: 500.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"%s\ninfrastructure failure\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestAPI_Auth__webSocket(t *testing.T) {
	// GIVEN: an auth-enabled API with a logged-in admin.
	file := "TestAPI_Auth__webSocket.yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	tests := []struct {
		name    string
		cookie  *http.Cookie
		want401 bool
	}{
		{name: "no cookie is rejected", want401: true},
		{
			name:    "invalid cookie is rejected",
			cookie:  &http.Cookie{Name: authCookieName, Value: "not-a-session"},
			want401: true,
		},
		{name: "valid cookie reaches the WS", cookie: cookie},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prefix := fmt.Sprintf("%s\n/ws", packageName)

			// WHEN: the /ws route is requested.
			w := serveAuth(api,
				authedRequest(http.MethodGet, "/ws", "", tc.cookie))

			// THEN: only authenticated handshakes get past the gate
			if tc.want401 {
				if got, want := w.Code, http.StatusUnauthorized; got != want {
					t.Errorf(
						"%s\nstatus mismatch\ngot:  %d\nwant: %d",
						prefix, got, want,
					)
				}
			}
			// AND: an authenticated non-upgrade request fails later, as 400.
			if !tc.want401 && w.Code == http.StatusUnauthorized {
				t.Errorf(
					"%s\nauthenticated handshake should pass the auth gate\ngot: %d",
					prefix, w.Code,
				)
			}
		})
	}
}

func TestAPI_Auth_serviceScopeHooks(t *testing.T) {
	// GIVEN: an auth-enabled API whose config holds the fixture service
	// "test", and a group with a grant scoped to it.
	file := "TestAPI_Auth_serviceScopeHooks.yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	group, err := deps.Store.CreateGroup(t.Context(), "scoped", "", []rbac.Grant{
		{Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
			Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
	})
	if err != nil {
		t.Fatalf(
			"%s\ncreate scoped group: %v",
			packageName, err,
		)
	}

	// AND: a local upstream so the edited lookup can fetch.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("v1.2.3"))
	}))
	t.Cleanup(upstream.Close)

	prefix := fmt.Sprintf("%s\nservice scope hooks", packageName)

	// WHEN: the service is renamed through the edit API.
	payload := fmt.Sprintf(`{
		"id": "test-renamed",
		"latest_version": {
			"type": "url",
			"url": %q,
			"url_commands": [{"type": "regex", "regex": "v([0-9.]+)"}]
		}
	}`, upstream.URL)
	w := serveAuth(api,
		authedRequest(http.MethodPut,
			"/api/v1/service/config?service_id=test", payload, adminCookie))

	// THEN: the edit succeeds.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nedit status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}

	// AND: the service-scoped grant followed the rename.
	got, err := deps.Store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf(
			"%s\nGroupByID failed: %v",
			prefix, err,
		)
	}
	if len(got.Grants) != 1 || got.Grants[0].Scope.Ref != "test-renamed" {
		t.Errorf(
			"%s\ngrant should follow the rename\ngot: %+v",
			prefix, got.Grants,
		)
	}

	// WHEN: the renamed service is deleted through the API.
	w = serveAuth(api,
		authedRequest(http.MethodDelete,
			"/api/v1/service/delete?service_id=test-renamed", "", adminCookie))

	// THEN: the delete succeeds.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\ndelete status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}

	// AND: the scoped grant was pruned.
	got, err = deps.Store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf(
			"%s\nGroupByID failed: %v",
			prefix, err,
		)
	}
	if len(got.Grants) != 0 {
		t.Errorf(
			"%s\nscoped grants should be pruned on delete\ngot: %+v",
			prefix, got.Grants,
		)
	}
}

func TestAPI_Auth__serviceScopeHooks_StoreFailure(t *testing.T) {
	// GIVEN: an auth-enabled API whose grants table is broken, invoked
	// directly (bypassing the middleware, which would fail first).
	file := "TestAPI_Auth__serviceScopeHooks_StoreFailure.yml"
	api, deps, dbConn := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)
	if _, err := dbConn.Exec(`DROP TABLE group_permissions;`); err != nil {
		t.Fatalf(
			"%s\nsetup drop failed: %v",
			packageName, err,
		)
	}

	// AND: a local upstream so the edited lookup can fetch.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("v1.2.3"))
	}))
	t.Cleanup(upstream.Close)

	prefix := fmt.Sprintf("%s\nscope hooks with a broken store", packageName)

	// WHEN: the fixture service is renamed.
	payload := fmt.Sprintf(`{
		"id": "test-renamed",
		"latest_version": {
			"type": "url",
			"url": %q,
			"url_commands": [{"type": "regex", "regex": "v([0-9.]+)"}]
		}
	}`, upstream.URL)
	req := withAuthCtx(httptest.NewRequest(http.MethodPut,
		"/api/v1/service/config?service_id=test", strings.NewReader(payload)), authCtx)
	w := httptest.NewRecorder()
	api.httpServiceEdit(w, req)

	// THEN: the edit still succeeds - grant maintenance never blocks the edit.
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\nedit should survive a hook failure\ngot:  %d - %s",
			prefix, w.Code, w.Body.String(),
		)
	}

	// WHEN: the renamed service is deleted.
	req = withAuthCtx(httptest.NewRequest(http.MethodDelete,
		"/api/v1/service/delete?service_id=test-renamed", nil), authCtx)
	w = httptest.NewRecorder()
	api.httpServiceDelete(w, req)

	// THEN: the delete still succeeds too.
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\ndelete should survive a hook failure\ngot:  %d - %s",
			prefix, w.Code, w.Body.String(),
		)
	}
}
