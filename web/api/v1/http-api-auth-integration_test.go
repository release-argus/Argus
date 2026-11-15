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
	"github.com/release-argus/Argus/internal/test"
)

func TestAPI_Auth__RBAC(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	createAuthUser(t, deps, "viewer", "viewer-password", store.GroupViewer)
	createAuthUser(t, deps, "loner", "loner-password")
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

			req := authedRequest(tc.method, tc.target, "", tc.cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_Auth__infrastructureFailures(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, dbConn := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	// Closing the DB makes every store lookup fail.
	_ = dbConn.Close()

	req := authedRequest(http.MethodGet, "/api/v1/flags", "", cookie)
	w := serveAuth(api, req)

	// 500, not an auth failure.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"authenticated request: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`),
	)
	w = serveAuth(api, req)
	// 500, not "invalid credentials".
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("login: status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}

func TestAPI_Auth__disabledMidSession(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)
	createAuthUser(t, deps, "victim", "victim-password", store.GroupViewer)
	cookie := loginCookie(t, api, "victim", "victim-password")
	if _, err := dbConn.Exec(`UPDATE users SET enabled = 0 WHERE username = 'victim';`); err != nil {
		t.Fatalf("DB.Exec() failed: %v", err)
	}

	req := authedRequest(http.MethodGet, "/api/v1/flags", "", cookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}

func TestAPI_Auth__originCheck(t *testing.T) {
	file := t.Name() + ".yml"
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

			req := authedRequest(http.MethodGet, "http://example.com/api/v1/flags", "", cookie)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_Auth__guardDisabled(t *testing.T) {
	file := t.Name() + ".yml"
	apiValue := testAPI(t, file)
	api := &apiValue

	called := false
	handler := api.guard(
		"service",
		"read",
		nil,
		func(_ http.ResponseWriter, _ *http.Request) {
			called = true
		},
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(httptest.NewRecorder(), req)

	if !called {
		t.Error("after guard: handler called mismatch\ngot:  false\nwant: true")
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	w := httptest.NewRecorder()
	api.GuardMetrics(inner).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf(
			"after GuardMetrics: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_Auth__tokenFallbackMetrics(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	adminCtx := adminContext(t, api, deps)
	token, _ := createAPIToken(t, deps, adminCtx.User.ID, "scrape", nil)
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

			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			if tc.basicAuth {
				req.SetBasicAuth("admin", "admin-password")
			}
			w := httptest.NewRecorder()
			guarded.ServeHTTP(w, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_Auth__guardMetricsErrors(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)

	loner := createAuthUser(t, deps, "loner", "loner-password")
	lonerToken, _ := createAPIToken(t, deps, loner.ID, "scrape", nil)

	adminCtx := adminContext(t, api, deps)
	adminToken, _ := createAPIToken(t, deps, adminCtx.User.ID, "scrape", nil)
	guarded := api.GuardMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("metrics"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+lonerToken)
	w := httptest.NewRecorder()
	guarded.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf(
			"no permission: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	_ = dbConn.Close()
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	guarded.ServeHTTP(w, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"broken store: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_Auth__webSocket(t *testing.T) {
	file := t.Name() + ".yml"
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
			req := authedRequest(http.MethodGet, "/ws", "", tc.cookie)
			w := serveAuth(api, req)

			// An authenticated non-upgrade request fails at the WebSocket upgrade instead.
			want := http.StatusBadRequest
			if tc.want401 {
				want = http.StatusUnauthorized
			}
			if got := w.Code; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_Auth__serviceScopeHooks(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	adminCookie := loginCookie(t, api, "admin", "admin-password")
	group := createAuthGroup(t, deps, "scoped", "", []rbac.Grant{
		{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
	})

	// A local upstream so the edited lookup can fetch.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("v1.2.3"))
	}))
	t.Cleanup(upstream.Close)

	payload := fmt.Sprintf(`{
		"id": "test-renamed",
		"latest_version": {
			"type": "url",
			"url": %q,
			"url_commands": [{"type": "regex", "regex": "v([0-9.]+)"}]
		}
	}`, upstream.URL)
	req := authedRequest(
		http.MethodPut, "/api/v1/service/config?service_id=test", payload, adminCookie,
	)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"rename: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	got, err := deps.Store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf("rename: Store.GroupByID() failed: %v", err)
	}
	grantRefs := test.AttributesSorted(got.Grants, func(g rbac.Grant) string { return g.Scope.Ref })
	if testErr := test.AssertSlicesEqual(
		t, grantRefs, []string{"test-renamed"}, "Group.Grants[].Scope.Ref",
	); testErr != nil {
		t.Errorf("rename: %v", testErr)
	}

	req = authedRequest(
		http.MethodDelete, "/api/v1/service/delete?service_id=test-renamed", "", adminCookie,
	)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("prune: status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	got, err = deps.Store.GroupByID(t.Context(), group.ID)
	if err != nil {
		t.Fatalf("prune: Store.GroupByID() failed: %v", err)
	}
	grantRefs = test.AttributesSorted(got.Grants, func(g rbac.Grant) string { return g.Scope.Ref })
	if testErr := test.AssertSlicesEqual(
		t, grantRefs, []string{}, "Group.Grants[].Scope.Ref",
	); testErr != nil {
		t.Errorf("prune: %v", testErr)
	}
}

func TestAPI_Auth__serviceScopeHooks__storeFailure(t *testing.T) {
	// The middleware would hit the broken table first, so the handlers are called directly.
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)
	test.DropTable(t, dbConn, "group_permissions")

	// A local upstream so the edited lookup can fetch.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("v1.2.3"))
	}))
	t.Cleanup(upstream.Close)

	payload := fmt.Sprintf(`{
		"id": "test-renamed",
		"latest_version": {
			"type": "url",
			"url": %q,
			"url_commands": [{"type": "regex", "regex": "v([0-9.]+)"}]
		}
	}`, upstream.URL)
	req := httptest.NewRequest(
		http.MethodPut, "/api/v1/service/config?service_id=test", strings.NewReader(payload),
	)
	req = withAuthCtx(req, authCtx)
	w := httptest.NewRecorder()
	api.httpServiceEdit(w, req, actionEdit)

	// Grants move before the service, so the rename is refused rather than half-applied.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Fatalf(
			"rename: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	// The original ID stays, so no grant is orphaned.
	api.Config.OrderMu.RLock()
	_, renamed := api.Config.Service["test-renamed"]
	_, original := api.Config.Service["test"]
	api.Config.OrderMu.RUnlock()
	if wantRenamed, wantOriginal := false, true; renamed != wantRenamed || original != wantOriginal {
		t.Fatalf(
			"rename: Config.Service keys mismatch\n"+
				"got:  test-renamed=%t, test=%t\nwant: test-renamed=%t, test=%t",
			renamed, original,
			wantRenamed, wantOriginal,
		)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/service/delete?service_id=test", nil)
	req = withAuthCtx(req, authCtx)
	w = httptest.NewRecorder()
	api.httpServiceDelete(w, req)

	// Grants are pruned before the service is removed.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Fatalf(
			"delete: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	// The service survives, so a later service of the same ID cannot inherit leftover grants.
	api.Config.OrderMu.RLock()
	_, survived := api.Config.Service["test"]
	api.Config.OrderMu.RUnlock()
	if !survived {
		t.Fatal("delete: Config.Service[\"test\"] mismatch\ngot:  absent\nwant: present")
	}
}
