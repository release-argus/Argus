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
	"net/netip"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/provider"
	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestAPI_EnableAuth(t *testing.T) {
	api := &API{}
	deps := &AuthDeps{}

	api.EnableAuth(deps)

	if api.auth != deps {
		t.Errorf("API.auth mismatch\ngot:  %p\nwant: %p", api.auth, deps)
	}
	if api.loginLimiter == nil {
		t.Error("API.loginLimiter mismatch\ngot:  nil\nwant: non-nil")
	}
}

func TestAuthContextFrom(t *testing.T) {
	authCtx := &auth.Context{
		Identity: auth.Identity{
			Username: "argus",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := authContextFrom(req); got != nil {
		t.Error("without context: authContextFrom() mismatch\ngot:  non-nil\nwant: nil")
	}

	carrying := withAuthCtx(httptest.NewRequest(http.MethodGet, "/", nil), authCtx)
	if got := authContextFrom(carrying); got != authCtx {
		t.Errorf("with context: authContextFrom() mismatch\ngot:  %p\nwant: %p", got, authCtx)
	}
}

func TestAPI_AuthMiddleware(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	tests := []struct {
		name       string
		path       string
		withCookie bool
		wantStatus int
		wantCtx    bool
	}{
		{
			name:       "protected path without credentials",
			path:       "/api/v1/flags",
			withCookie: false,
			wantStatus: http.StatusUnauthorized,
			wantCtx:    false,
		},
		{
			name:       "protected path with a valid session",
			path:       "/api/v1/flags",
			withCookie: true,
			wantStatus: http.StatusOK,
			wantCtx:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotCtx *auth.Context
			handler := api.authMiddleware()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotCtx = authContextFrom(r)
					w.WriteHeader(http.StatusOK)
				}),
			)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.withCookie {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			if (gotCtx != nil) != tc.wantCtx {
				t.Errorf(
					"authContextFrom() != nil mismatch\ngot:  %t\nwant: %t",
					gotCtx != nil, tc.wantCtx,
				)
			}
		})
	}
}

func TestFailAuthenticationError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		bodyRegex  string
	}{
		{
			name:       "unauthorised",
			err:        errUnauthorised,
			wantStatus: http.StatusUnauthorized,
			bodyRegex:  `"` + errUnauthorised.Error() + `"`,
		},
		{
			name:       "wrapped unauthorised",
			err:        fmt.Errorf("session: %w", errUnauthorised),
			wantStatus: http.StatusUnauthorized,
			bodyRegex:  `"` + errUnauthorised.Error() + `"`,
		},
		{
			name:       "infrastructure failure",
			err:        errors.New("db exploded"),
			wantStatus: http.StatusInternalServerError,
			bodyRegex:  `"authentication failed"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			failAuthenticationError(w, req, tc.err)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			if !util.RegexCheck(tc.bodyRegex, w.Body.String()) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", w.Body.String(), tc.bodyRegex)
			}

			if strings.Contains(w.Body.String(), "db exploded") {
				t.Errorf("body mismatch\ngot:  %q\nwant: NOT %q", w.Body.String(), "db exploded")
			}
		})
	}
}

func TestAPI_AuthCtxOr401(t *testing.T) {
	api := &API{}
	authCtx := &auth.Context{Identity: auth.Identity{Username: "argus"}}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := api.authCtxOr401(w, req)

	if got != nil || w.Code != http.StatusUnauthorized {
		t.Errorf(
			"without context: mismatch\ngot:  ctx=%p, status=%d\nwant: ctx=nil, status=401",
			got, w.Code,
		)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = withAuthCtx(req, authCtx)
	got = api.authCtxOr401(w, req)

	if got != authCtx || w.Code != http.StatusOK {
		t.Errorf(
			"with context: mismatch\ngot:  ctx=%p, status=%d\nwant: ctx=%p, status=200",
			got, w.Code, authCtx,
		)
	}
}

func TestAPI_Authenticate(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	cookie := loginCookie(t, api, "admin", "admin-password")
	adminCtx := adminContext(t, api, deps)
	token, _ := createAPIToken(t, deps, adminCtx.User.ID, "probe", nil)

	tests := []struct {
		name         string
		cookie       string // "valid"/"stale"/"".
		bearer       string
		wantProvider string
		errRegex     string
	}{
		{
			name:         "valid session cookie",
			cookie:       "valid",
			wantProvider: "local",
			errRegex:     `^$`,
		},
		{
			name:         "stale cookie falls through to a valid Bearer",
			cookie:       "stale",
			bearer:       token,
			wantProvider: "api_token",
			errRegex:     `^$`,
		},
		{
			name:     "stale cookie alone",
			cookie:   "stale",
			errRegex: `^` + errUnauthorised.Error() + `$`,
		},
		{
			name:     "unknown Bearer token",
			bearer:   "argus_" + strings.Repeat("0", 64),
			errRegex: `^` + errUnauthorised.Error() + `$`,
		},
		{
			name:     "no credentials",
			errRegex: `^` + errUnauthorised.Error() + `$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares deps.Sessions, which the parent swaps after the loop.

			req := httptest.NewRequest(http.MethodGet, "/api/v1/flags", nil)
			switch tc.cookie {
			case "valid":
				req.AddCookie(cookie)
			case "stale":
				req.AddCookie(&http.Cookie{Name: authCookieName, Value: "stale"})
			}
			if tc.bearer != "" {
				req.Header.Set("Authorization", "Bearer "+tc.bearer)
			}

			authCtx, err := api.authenticate(req)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if authCtx.Identity.Provider != tc.wantProvider {
				t.Errorf(
					"Context.Identity.Provider mismatch\ngot:  %q\nwant: %q",
					authCtx.Identity.Provider, tc.wantProvider,
				)
			}
		})
	}

	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime:    time.Hour,
			IdleTimeout: time.Hour,
		},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/flags", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: "anything"})
	_, err := api.authenticate(req)
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf(
			"session store failure: error type mismatch\ngot:  %v\nwant: NOT %v",
			err, errUnauthorised,
		)
	}
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
		wantOK bool
	}{
		{
			name:   "Bearer token",
			header: "Bearer argus_abc",
			want:   "argus_abc",
			wantOK: true,
		},
		{
			name:   "scheme is case-insensitive",
			header: "beaRer argus_abc",
			want:   "argus_abc",
			wantOK: true,
		},
		{
			name:   "scheme without a value",
			header: "Bearer ",
			wantOK: false,
		},
		{
			name:   "other scheme",
			header: "Token argus_abc",
			wantOK: false,
		},
		{
			name:   "no header",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			got, ok := bearerToken(req)

			if got != tc.want || ok != tc.wantOK {
				t.Errorf("mismatch\ngot:  %q, %t\nwant: %q, %t", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestAPI_AuthenticateAPIToken(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)

	adminCtx := adminContext(t, api, deps)
	live, _ := createAPIToken(t, deps, adminCtx.User.ID, "live", nil)

	past := timeNow().UTC().Add(-time.Hour)
	expired, _ := createAPIToken(t, deps, adminCtx.User.ID, "expired", &past)

	if _, err := dbConn.Exec(`
		CREATE TRIGGER fail_touch BEFORE UPDATE ON api_tokens
		BEGIN SELECT RAISE(ABORT, 'touch broke'); END;`,
	); err != nil {
		t.Fatalf("create trigger: DB.Exec() failed: %v", err)
	}

	// A failed usage touch does not fail authentication.
	authCtx, err := api.authenticateAPIToken(t.Context(), live)
	if err != nil {
		t.Errorf("touch failure: unexpected error: %v", err)
	} else if got, want := authCtx.Identity.Provider, "api_token"; got != want {
		t.Errorf("touch failure: Context.Identity.Provider mismatch\ngot:  %q\nwant: %q", got, want)
	}
	if _, err := dbConn.Exec(`DROP TRIGGER fail_touch;`); err != nil {
		t.Fatalf("drop trigger: DB.Exec() failed: %v", err)
	}

	authCtx, err = api.authenticateAPIToken(t.Context(), live)
	if err != nil {
		t.Errorf("live token: unexpected error: %v", err)
	} else if authCtx.Identity.Provider != "api_token" || authCtx.User.ID != adminCtx.User.ID {
		t.Errorf(
			"live token: Context mismatch\n"+
				"got:  provider=%q, userID=%q\nwant: provider=%q, userID=%q",
			authCtx.Identity.Provider, authCtx.User.ID,
			"api_token", adminCtx.User.ID,
		)
	}

	tokens, err := deps.Store.APITokensForUser(t.Context(), adminCtx.User.ID)
	if err != nil {
		t.Fatalf("live token: Store.APITokensForUser() failed: %v", err)
	}
	i := slices.IndexFunc(tokens, func(token store.APIToken) bool { return token.Name == "live" })
	if i == -1 {
		t.Fatalf(
			"live token: Store.APITokensForUser() mismatch\n"+
				"got:  %d tokens\nwant: one named \"live\"",
			len(tokens),
		)
	}
	if tokens[i].LastUsedAt == nil {
		t.Error("live token: APIToken.LastUsedAt mismatch\ngot:  nil\nwant: non-nil")
	}

	for _, tc := range []struct {
		name, value string
	}{
		{name: "expired", value: expired},
		{name: "unknown", value: "argus_" + strings.Repeat("0", 64)},
	} {
		t.Run("invalid/"+tc.name+" token", func(t *testing.T) {
			// Serial: shares dbConn, which the parent closes after the loop.

			_, err := api.authenticateAPIToken(t.Context(), tc.value)
			if !errors.Is(err, errUnauthorised) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, errUnauthorised)
			}
		})
	}

	_ = dbConn.Close()
	_, err = api.authenticateAPIToken(t.Context(), live)
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf("store failure: error type mismatch\ngot:  %v\nwant: NOT %v", err, errUnauthorised)
	}
}

func TestAPI_AuthenticateSession(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCtx := adminContext(t, api, deps)
	token, _, err := deps.Sessions.Start(
		t.Context(),
		adminCtx.User.ID,
		"127.0.0.1",
		"go-test",
	)
	if err != nil {
		t.Fatalf("Manager.Start() failed: %v", err)
	}

	authCtx, err := api.authenticateSession(t.Context(), token)
	if err != nil {
		t.Errorf("live session: unexpected error: %v", err)
	} else if authCtx.Identity.Provider != "local" || authCtx.User.ID != adminCtx.User.ID {
		t.Errorf(
			"live session: Context mismatch\n"+
				"got:  provider=%q, userID=%q\nwant: provider=%q, userID=%q",
			authCtx.Identity.Provider, authCtx.User.ID,
			"local", adminCtx.User.ID,
		)
	}

	_, err = api.authenticateSession(t.Context(), "never-issued")
	if !errors.Is(err, errUnauthorised) {
		t.Errorf("unknown token: error type mismatch\ngot:  %v\nwant: %v", err, errUnauthorised)
	}

	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime:    time.Hour,
			IdleTimeout: time.Hour,
		},
	)
	_, err = api.authenticateSession(t.Context(), token)
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf("store failure: error type mismatch\ngot:  %v\nwant: NOT %v", err, errUnauthorised)
	}
}

func TestAPI_VerifyLocalCredentials(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	identity, err := api.verifyLocalCredentials(
		t.Context(),
		"admin",
		"admin-password",
		"ip",
	)
	if err != nil {
		t.Errorf("valid credentials: unexpected error: %v", err)
	} else if got, want := identity.Username, "admin"; got != want {
		t.Errorf("valid credentials: Identity.Username mismatch\ngot:  %q\nwant: %q", got, want)
	}

	for i := range loginLimitAttempts {
		if _, err := api.verifyLocalCredentials(
			t.Context(),
			"admin",
			"wrong",
			"ip",
		); !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatalf(
				"attempt %d: bad credentials: error type mismatch\ngot:  %v\nwant: %v",
				i, err, auth.ErrInvalidCredentials,
			)
		}
	}

	if _, err := api.verifyLocalCredentials(
		t.Context(),
		"admin",
		"admin-password",
		"ip",
	); !errors.Is(err, errTooManyAttempts) {
		t.Errorf(
			"exhausted budget: error type mismatch\ngot:  %v\nwant: %v",
			err, errTooManyAttempts,
		)
	}

	deps.Providers = provider.NewRegistry()
	if _, err := api.verifyLocalCredentials(
		t.Context(),
		"admin",
		"admin-password",
		"other-ip",
	); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf(
			"no providers: error type mismatch\ngot:  %v\nwant: %v",
			err, auth.ErrInvalidCredentials,
		)
	}
}

func TestAPI_ContextForUser(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)
	adminCtx := adminContext(t, api, deps)
	disabled := createAuthUser(t, deps, "disabled", "disabled-password")
	if _, err := deps.Store.UpdateUser(
		t.Context(),
		disabled.ID,
		store.UserPatch{
			Enabled: new(false),
		},
	); err != nil {
		t.Fatalf("Store.UpdateUser() failed: %v", err)
	}

	authCtx, err := api.contextForUser(t.Context(), adminCtx.User.ID, "local")
	if err != nil {
		t.Errorf("enabled user: unexpected error: %v", err)
	} else if authCtx.Identity.Username != "admin" ||
		authCtx.User.ID != adminCtx.User.ID ||
		len(authCtx.Grants) == 0 ||
		authCtx.Permissions == nil {
		t.Errorf(
			"enabled user: Context mismatch\n"+
				"got:  username=%q, userID=%q, %d grants, permissions set=%t\n"+
				"want: username=%q, userID=%q, >0 grants, permissions set=true",
			authCtx.Identity.Username, authCtx.User.ID,
			len(authCtx.Grants), authCtx.Permissions != nil,
			"admin", adminCtx.User.ID,
		)
	}

	for _, tc := range []struct {
		name, id string
	}{
		{name: "disabled", id: disabled.ID},
		{name: "unknown", id: "no-such-id"},
	} {
		t.Run("invalid/"+tc.name+" user", func(t *testing.T) {
			// Serial: shares the users table, which the parent drops after the loop.

			_, err := api.contextForUser(t.Context(), tc.id, "local")
			if !errors.Is(err, errUnauthorised) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, errUnauthorised)
			}
		})
	}

	test.DropTable(t, dbConn, "users")
	_, err = api.contextForUser(t.Context(), adminCtx.User.ID, "local")
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf("store failure: error type mismatch\ngot:  %v\nwant: NOT %v", err, errUnauthorised)
	}
}

func TestAPI__auth__scopedGrants(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	api.Config.OrderMu.Lock()
	api.Config.Service["tagged"] = &service.Service{
		Dashboard: dashboard.Options{Tags: []string{"prod"}},
	}
	// Refreshable by tag, but not readable.
	api.Config.Service["staging-only"] = &service.Service{
		Dashboard: dashboard.Options{Tags: []string{"staging"}},
	}
	api.Config.OrderMu.Unlock()

	createAuthGroup(
		t,
		deps,
		"scoped",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"}},
			{Resource: rbac.ResourceVersionRefresh, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"}},
			{Resource: rbac.ResourceVersionRefresh, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "staging"}},
		},
	)
	createAuthUser(t, deps, "scoped-user", "scoped-password", "scoped")
	cookie := loginCookie(t, api, "scoped-user", "scoped-password")

	tests := []struct {
		name       string
		target     string
		wantStatus int // 403 = denied; anything else = passed the guard.
	}{
		{
			name:       "service-scoped grant allows its service",
			target:     "/api/v1/service/summary?service_id=test",
			wantStatus: http.StatusOK,
		},
		{
			name:       "service-scoped grant denies other services",
			target:     "/api/v1/service/summary?service_id=other",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "tag-scoped grant allows services with the tag",
			target:     "/api/v1/latest_version/refresh?service_id=tagged",
			wantStatus: http.StatusBadRequest, // Lookup is nil.
		},
		{
			name:       "refresh denied without service:read on the target",
			target:     "/api/v1/latest_version/refresh?service_id=staging-only",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "global-only route denies scoped-only users",
			target:     "/api/v1/flags",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := authedRequest(http.MethodGet, tc.target, "", cookie)
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

func TestAPI__auth__GuardAnyScopeOf(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	// The defaults are for authoring a service, so editing one reaches them.
	createAuthGroup(
		t,
		deps,
		"scoped",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceService, Action: rbac.ActionUpdate, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
		},
	)
	createAuthUser(t, deps, "scoped-user", "scoped-password", "scoped")

	createAuthGroup(
		t,
		deps,
		"reader",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
		},
	)
	createAuthUser(t, deps, "reader-user", "reader-password", "reader")

	createAuthUser(t, deps, "loner", "loner-password")

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
	}{
		{
			name:       "any service:update grant passes the guard",
			username:   "scoped-user",
			password:   "scoped-password",
			wantStatus: http.StatusOK,
		},
		{
			name:       "service:read alone is forbidden",
			username:   "reader-user",
			password:   "reader-password",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no grants at all is forbidden",
			username:   "loner",
			password:   "loner-password",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginCookie(t, api, tc.username, tc.password)

			req := authedRequest(http.MethodGet, "/api/v1/service/defaults", "", cookie)
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

func TestAPI__auth__permissionChangesApplyImmediately(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	group := createAuthGroup(
		t,
		deps,
		"readers",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceConfig, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
		},
	)
	createAuthUser(t, deps, "reader", "reader-password", "readers")
	cookie := loginCookie(t, api, "reader", "reader-password")

	req := authedRequest(http.MethodGet, "/api/v1/flags", "", cookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"read before revoking permissions: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	empty := []rbac.Grant{}
	if _, err := deps.Store.UpdateGroup(
		t.Context(),
		group.ID,
		store.GroupPatch{Grants: &empty},
	); err != nil {
		t.Fatalf("Store.UpdateGroup() failed: %v", err)
	}

	req = authedRequest(http.MethodGet, "/api/v1/flags", "", cookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf(
			"read after revoking permissions: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_OriginCheckMiddleware__secFetchSite(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	tests := []struct {
		name          string
		secFetchSite  string
		wantForbidden bool
	}{
		{
			name:          "valid/absent - non-browser client",
			wantForbidden: false,
		},
		{
			name:          "valid/same-origin",
			secFetchSite:  "same-origin",
			wantForbidden: false,
		},
		{
			name:          "valid/none - typed into the address bar",
			secFetchSite:  "none",
			wantForbidden: false,
		},
		{
			name:          "invalid/cross-site with no Origin",
			secFetchSite:  "cross-site",
			wantForbidden: true,
		},
		{
			name:          "invalid/same-site sibling with no Origin",
			secFetchSite:  "same-site",
			wantForbidden: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := authedRequest(http.MethodGet, "/api/v1/flags", "", cookie)
			if tc.secFetchSite != "" {
				req.Header.Set("Sec-Fetch-Site", tc.secFetchSite)
			}

			w := serveAuth(api, req)

			gotForbidden := w.Code == http.StatusForbidden
			if gotForbidden != tc.wantForbidden {
				t.Errorf(
					"forbidden mismatch\ngot:  %t (status %d - %q)\nwant: %t",
					gotForbidden, w.Code, w.Body.String(), tc.wantForbidden,
				)
			}
		})
	}
}

func TestOriginCheckMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		origin        string
		forwardedHost string
		trustPeer     bool
		wantStatus    int
	}{
		{
			name:       "no Origin",
			origin:     "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "null Origin/state-changing is rejected",
			method:     http.MethodPost,
			origin:     "null",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "null Origin/safe method passes",
			method:     http.MethodGet,
			origin:     "null",
			wantStatus: http.StatusOK,
		},
		{
			name:       "matching Origin",
			origin:     "http://example.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "cross-site Origin",
			origin:     "http://evil.example.net",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "malformed Origin",
			origin:     "::not-a-url::",
			wantStatus: http.StatusForbidden,
		},
		{
			name:          "trusted proxy/Origin matching X-Forwarded-Host passes",
			origin:        "https://public.example.net",
			forwardedHost: "public.example.net",
			trustPeer:     true,
			wantStatus:    http.StatusOK,
		},
		{
			name:          "trusted proxy/Origin not matching X-Forwarded-Host is rejected",
			origin:        "http://evil.example.net",
			forwardedHost: "public.example.net",
			trustPeer:     true,
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "untrusted peer/X-Forwarded-Host is ignored",
			origin:        "https://public.example.net",
			forwardedHost: "public.example.net",
			wantStatus:    http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			api := &API{}
			if tc.trustPeer {
				// httptest requests arrive from 192.0.2.0/24.
				api.trustedProxies = []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")}
			}
			handler := api.originCheckMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)

			method := util.ValueOr(tc.method, http.MethodPost)
			req := httptest.NewRequest(method, "http://example.com/api/v1/flags", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.forwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", tc.forwardedHost)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}
}

func TestAPI_Guard(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCtx := adminContext(t, api, deps)
	createAuthGroup(
		t,
		deps,
		"scoped",
		"",
		[]rbac.Grant{
			{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
		},
	)

	scopedUser := createAuthUser(t, deps, "scoped-user", "scoped-password", "scoped")
	scopedCtx, err := api.contextForUser(t.Context(), scopedUser.ID, "local")
	if err != nil {
		t.Fatalf("API.contextForUser() failed: %v", err)
	}

	serviceTarget := func(id string) func(*http.Request) *rbac.Target {
		return func(*http.Request) *rbac.Target {
			return &rbac.Target{ServiceID: id}
		}
	}
	tests := []struct {
		name        string
		disableAuth bool
		authCtx     *auth.Context
		target      func(*http.Request) *rbac.Target
		wantStatus  int
		wantCalled  bool // true if we pass the Guard, false otherwise.
	}{
		{
			name:        "auth disabled passes the handler through",
			disableAuth: true,
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
		{
			name:       "no auth context",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "permission held",
			authCtx:    adminCtx,
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "permission not held",
			authCtx:    scopedCtx,
			target:     serviceTarget("other"),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "scoped permission matches its target",
			authCtx:    scopedCtx,
			target:     serviceTarget("test"),
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guarded := api
			if tc.disableAuth {
				guarded = &API{}
			}

			called := false
			handler := guarded.guard(rbac.ResourceService, rbac.ActionRead, tc.target,
				func(w http.ResponseWriter, _ *http.Request) {
					called = true
					w.WriteHeader(http.StatusOK)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authCtx != nil {
				req = withAuthCtx(req, tc.authCtx)
			}
			w := httptest.NewRecorder()
			handler(w, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			if called != tc.wantCalled {
				t.Errorf("handler called mismatch\ngot:  %t\nwant: %t", called, tc.wantCalled)
			}
		})
	}
}

func TestAPI_RequireAdmin(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	adminCtx := adminContext(t, api, deps)

	viewer := createAuthUser(t, deps, "viewer-user", "viewer-password", store.GroupViewer)
	viewerCtx, err := api.contextForUser(t.Context(), viewer.ID, "local")
	if err != nil {
		t.Fatalf("API.contextForUser() failed: %v", err)
	}

	tests := []struct {
		name        string
		disableAuth bool
		authCtx     *auth.Context
		wantStatus  int
		wantCalled  bool // true if we call the handler being wrapped, false otherwise.
	}{
		{
			name:        "auth disabled/passes",
			disableAuth: true,
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
		{
			name:       "auth/no context",
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "auth/admin",
			authCtx:    adminCtx,
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "auth/non-admin",
			authCtx:    viewerCtx,
			wantStatus: http.StatusForbidden,
			wantCalled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guarded := api
			if tc.disableAuth {
				guarded = &API{}
			}

			called := false
			handler := guarded.requireAdmin(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authCtx != nil {
				req = withAuthCtx(req, tc.authCtx)
			}
			w := httptest.NewRecorder()
			handler(w, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			if called != tc.wantCalled {
				t.Errorf("handler called mismatch\ngot:  %t\nwant: %t", called, tc.wantCalled)
			}
		})
	}
}

func TestAPI_ReadableServices(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	api.Config.OrderMu.Lock()
	api.Config.Service["test"] = &service.Service{}
	api.Config.Service["tagged"] = &service.Service{
		Dashboard: dashboard.Options{Tags: []string{"prod"}},
	}
	api.Config.Order = append(api.Config.Order, "tagged")
	api.Config.OrderMu.Unlock()

	groups := []struct {
		name   string
		grants []rbac.Grant
	}{
		{
			name: "full-scoped",
			grants: []rbac.Grant{
				{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
				{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"}},
			},
		},
		{
			name: "service-scoped",
			grants: []rbac.Grant{
				{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
			},
		},
		{
			name: "tag-scoped",
			grants: []rbac.Grant{
				{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"}},
			},
		},
	}

	for _, group := range groups {
		createAuthGroup(t, deps, group.name, "", group.grants)
	}
	createAuthUser(t, deps, "full-scoped-user", "scoped-password", "full-scoped")
	createAuthUser(t, deps, "service-scoped-user", "scoped-password", "service-scoped")
	createAuthUser(t, deps, "tag-scoped-user", "scoped-password", "tag-scoped")
	createAuthUser(t, deps, "loner", "loner-password")

	tests := []struct {
		name     string
		username string // "" = the bootstrap admin.
		wantNil  bool
		want     map[string]bool
	}{
		{
			name:    "global service:read is unrestricted (nil)",
			wantNil: true,
		},
		{
			name:     "full-scoped grants",
			username: "full-scoped-user",
			want:     map[string]bool{"test": true, "tagged": true},
		},
		{
			name:     "service-scoped grants",
			username: "service-scoped-user",
			want:     map[string]bool{"test": true},
		},
		{
			name:     "tag-scoped grants",
			username: "tag-scoped-user",
			want:     map[string]bool{"tagged": true},
		},
		{
			name:     "no grants allow nothing",
			username: "loner",
			want:     map[string]bool{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			username := util.ValueOr(tc.username, "admin")
			creds, err := deps.Store.LocalCredentials(t.Context(), username)
			if err != nil {
				t.Fatalf("Store.LocalCredentials() failed: %v", err)
			}
			if creds == nil {
				t.Fatal("Store.LocalCredentials() mismatch\ngot:  nil\nwant: non-nil")
			}
			authCtx, err := api.contextForUser(t.Context(), creds.UserID, "local")
			if err != nil {
				t.Fatalf("API.contextForUser() failed: %v", err)
			}

			readable := api.readableServices(authCtx)

			if tc.wantNil {
				if readable != nil {
					t.Errorf(
						"API.readableServices() mismatch\ngot:  %v\nwant: nil (unrestricted)",
						readable,
					)
				}
				return
			}
			if readable == nil {
				t.Fatal("API.readableServices() mismatch\ngot:  nil (unrestricted)\nwant: non-nil")
			}
			if testErr := test.AssertMapEqual(t, readable, tc.want, "API.readableServices()"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestAPI_ActionableServices(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	groups := []struct {
		name   string
		grants []rbac.Grant
	}{
		{
			name: "global-actioner",
			grants: []rbac.Grant{
				{Resource: rbac.ResourceServiceAction, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
			},
		},
		{
			name: "scoped-actioner",
			grants: []rbac.Grant{
				{Resource: rbac.ResourceServiceAction, Action: rbac.ActionExecute, Scope: rbac.Scope{Type: rbac.ScopeService, Ref: "test"}},
			},
		},
		{
			name: "reader",
			grants: []rbac.Grant{
				{Resource: rbac.ResourceService, Action: rbac.ActionRead, Scope: rbac.Scope{Type: rbac.ScopeGlobal}},
			},
		},
	}
	for _, group := range groups {
		createAuthGroup(t, deps, group.name, "", group.grants)
	}
	createAuthUser(t, deps, "global-actioner-user", "actioner-password", "global-actioner")
	createAuthUser(t, deps, "scoped-actioner-user", "actioner-password", "scoped-actioner")
	createAuthUser(t, deps, "reader-user", "reader-password", "reader")

	tests := []struct {
		name     string
		username string // "" = the bootstrap admin.
		wantNil  bool
		want     map[string]bool
	}{
		{
			name:    "admin holds every grant globally (nil)",
			wantNil: true,
		},
		{
			name:     "global execute is unrestricted (nil)",
			username: "global-actioner-user",
			wantNil:  true,
		},
		{
			name:     "service-scoped execute",
			username: "scoped-actioner-user",
			want:     map[string]bool{"test": true},
		},
		{
			name:     "service:read alone permits no action content",
			username: "reader-user",
			want:     map[string]bool{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			username := util.ValueOr(tc.username, "admin")
			creds, err := deps.Store.LocalCredentials(t.Context(), username)
			if err != nil {
				t.Fatalf("Store.LocalCredentials() failed: %v", err)
			}
			if creds == nil {
				t.Fatal("Store.LocalCredentials() mismatch\ngot:  nil\nwant: non-nil")
			}
			authCtx, err := api.contextForUser(t.Context(), creds.UserID, "local")
			if err != nil {
				t.Fatalf("API.contextForUser() failed: %v", err)
			}

			permitted := api.actionableServices(authCtx)

			if tc.wantNil {
				if permitted != nil {
					t.Errorf(
						"API.actionableServices() mismatch\ngot:  %v\nwant: nil (unrestricted)",
						permitted,
					)
				}
				return
			}
			if permitted == nil {
				t.Fatal(
					"API.actionableServices() mismatch\ngot:  nil (unrestricted)\nwant: non-nil",
				)
			}
			if testErr := test.AssertMapEqual(t, permitted, tc.want, "API.actionableServices()"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestAPI_KickWebSocketClients__noAuth(t *testing.T) {
	// kickAll exercises every kick helper against the client's identity.
	kickAll := func(api *API, client *Client) {
		api.kickUserWebSocketClients(client.userID)
		api.kickSessionWebSocketClients(client.sessionHash)
		api.kickRestrictedWebSocketClients()
	}

	file := t.Name() + ".yml"
	apiValue := testAPI(t, file)
	apiNoAuth := &apiValue
	client := &Client{
		userID:           "user-a",
		sessionHash:      "hash-1",
		readableServices: map[string]bool{"svc": true},
	}
	// Without a hub, kicking need only not panic.
	kickAll(apiNoAuth, client)

	apiNoAuth.auth = &AuthDeps{}
	kickAll(apiNoAuth, client)

	apiNoAuth.auth = nil
	apiNoAuth.hub = NewHub()
	go apiNoAuth.hub.Run(t.Context())
	client.hub = apiNoAuth.hub
	client.send = make(chan []byte, 8)
	apiNoAuth.hub.register <- client
	kickAll(apiNoAuth, client)
	if !apiNoAuth.hub.hasClient(client) {
		t.Error("Hub.hasClient() mismatch\ngot:  false\nwant: true")
	}
}

func TestAPI_KickWebSocketClients__auth(t *testing.T) {
	apiWired := &API{auth: &AuthDeps{}}
	newClient := wireHub(t, apiWired)

	userA, userB := "user-a", "user-b"

	target := newClient(userA, "hash-1")
	other := newClient(userB, "hash-2")
	apiWired.kickUserWebSocketClients(userA)

	assertKicked(t, apiWired.hub, []*Client{target}, []*Client{other})

	sessionHash := "hash-3"
	target = newClient(userB, sessionHash)
	apiWired.kickSessionWebSocketClients(sessionHash)

	assertKicked(t, apiWired.hub, []*Client{target}, []*Client{other})

	target = newClient("user-c", "hash-4", map[string]bool{"svc": true})
	apiWired.kickRestrictedWebSocketClients()

	assertKicked(t, apiWired.hub, []*Client{target}, []*Client{other})
}

func TestAPI_ServiceTarget(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	api.Config.OrderMu.Lock()
	api.Config.Service["tagged"] = &service.Service{
		Dashboard: dashboard.Options{Tags: []string{"prod", "web"}},
	}
	api.Config.OrderMu.Unlock()

	tests := []struct {
		name     string
		query    string
		wantNil  bool
		wantID   string
		wantTags int
	}{
		{
			name:    "no service_id yields no target",
			query:   "",
			wantNil: true,
		},
		{
			name:     "unknown service yields a tagless target",
			query:    "service_id=ghost",
			wantID:   "ghost",
			wantTags: 0,
		},
		{
			name:     "known service carries its dashboard tags",
			query:    "service_id=tagged",
			wantID:   "tagged",
			wantTags: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/service/summary?"+tc.query, nil)
			target := api.serviceTarget(req)

			if tc.wantNil {
				if target != nil {
					t.Errorf(
						"API.serviceTarget() mismatch\ngot:  id=%q with %d tags\nwant: nil",
						target.ServiceID, len(target.Tags),
					)
				}
				return
			}
			if target == nil {
				t.Fatal("API.serviceTarget() mismatch\ngot:  nil\nwant: non-nil")
			}
			if target.ServiceID != tc.wantID || len(target.Tags) != tc.wantTags {
				t.Errorf(
					"API.serviceTarget() mismatch\n"+
						"got:  id=%q with %d tags\nwant: id=%q with %d tags",
					target.ServiceID, len(target.Tags),
					tc.wantID, tc.wantTags,
				)
			}
		})
	}
}

func TestAPI_SetupRoutesAPI__openRoutesSkipAuthentication(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		open   bool
	}{
		{
			name: "open/login", method: http.MethodPost,
			body: `{"username":"admin","password":"admin-password"}`,
			path: "/api/v1/auth/login",
			open: true,
		},
		{
			name:   "open/setup state",
			method: http.MethodGet,
			path:   "/api/v1/auth/setup",
			open:   true,
		},
		{
			name:   "open/setup",
			method: http.MethodPost,
			body:   "{}",
			path:   "/api/v1/auth/setup",
			open:   true,
		},
		{
			name:   "open/logout",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
			open:   true,
		},
		{
			name:   "authenticated/me",
			method: http.MethodGet,
			path:   "/api/v1/auth/me",
		},
		{
			name:   "authenticated/flags",
			method: http.MethodGet,
			path:   "/api/v1/flags",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			w := serveAuth(api, req)

			got := w.Code
			if tc.open && got == http.StatusUnauthorized {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: NOT %d",
					got, w.Body.String(), http.StatusUnauthorized,
				)
			}
			if !tc.open && got != http.StatusUnauthorized {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), http.StatusUnauthorized,
				)
			}
		})
	}
}

func TestAPI_SetupRoutesAPI__everyRouteIsGuarded(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	createAuthUser(t, deps, "loner", "loner-password")
	cookie := loginCookie(t, api, "loner", "loner-password")

	// selfServiceRoutes are the routes that answer a grantless session and filter
	// to the caller's own account and grants.
	selfServiceRoutes := map[string]bool{
		"POST   /api/v1/auth/login":          true,
		"GET    /api/v1/auth/setup":          true,
		"POST   /api/v1/auth/setup":          true,
		"POST   /api/v1/auth/logout":         true,
		"GET    /api/v1/auth/me":             true,
		"PATCH  /api/v1/auth/me":             true,
		"PUT    /api/v1/auth/me/preferences": true,
		"DELETE /api/v1/auth/me/preferences": true,
		"GET    /api/v1/tokens":              true,
		"POST   /api/v1/tokens":              true,
		"DELETE /api/v1/tokens/{id}":         true,
		"GET    /api/v1/ws-token":            true,
		"GET    /api/v1/service/order":       true,
	}
	const methodLen = 6

	type route struct {
		path    string
		methods []string
	}
	var routes []route
	if err := api.Router.Walk(
		func(r *mux.Route, _ *mux.Router, _ []*mux.Route) error {
			path, err := r.GetPathTemplate()
			if err != nil || !strings.HasPrefix(path, "/api/v1/") {
				return nil
			}
			methods, err := r.GetMethods()
			if err != nil || len(methods) == 0 {
				methods = []string{http.MethodGet}
			}
			routes = append(routes, route{path: path, methods: methods})
			return nil
		},
	); err != nil {
		t.Fatalf("Router.Walk() failed: %v", err)
	}

	if got := len(routes); got < 20 {
		t.Fatalf("len(routes) mismatch\ngot:  %d\nwant: >= 20", got)
	}

	for _, rt := range routes {
		for _, method := range rt.methods {
			name := fmt.Sprintf("method=%s path=%s", method, rt.path)
			t.Run(name, func(t *testing.T) {
				// Grantless routes are not expected to be guarded.
				selfServiceRouteKey := method + strings.Repeat(" ", methodLen-len(method)) + " " + rt.path
				if selfServiceRoutes[selfServiceRouteKey] {
					t.Skip("grantless route")
				}

				target := strings.ReplaceAll(rt.path, "{id}", "no-such-id")
				req := authedRequest(method, target, "{}", cookie)
				w := serveAuth(api, req)

				if wants := []int{http.StatusUnauthorized, http.StatusForbidden}; !slices.Contains(wants, w.Code) {
					t.Errorf(
						"status code mismatch\ngot:  %d - %q\nwant: %v",
						w.Code, w.Body.String(), wants,
					)
				}
			})
		}
	}
}

// demoGrants are global grants over every resource, so a demo user is refused
// for their group membership rather than for any missing permission.
func demoGrants() []rbac.Grant {
	global := rbac.Scope{Type: rbac.ScopeGlobal}
	grants := make([]rbac.Grant, 0, 16)
	for _, rp := range rbac.Catalogue() {
		for _, ap := range rp.Actions {
			grants = append(grants, rbac.Grant{Permission: rbac.Permission{Resource: rp.Resource, Action: ap.Action}, Scope: global})
		}
	}
	return grants
}

func TestAPI_DemoMiddleware(t *testing.T) {
	// Each refused read makes an outbound request.
	refusedReads := map[string]bool{
		"/api/v1/latest_version/refresh":             true,
		"/api/v1/deployed_version/refresh":           true,
		"/api/v1/latest_version/refresh_uncreated":   true,
		"/api/v1/deployed_version/refresh_uncreated": true,
	}
	openRoutes := map[string]bool{
		"/api/v1/auth/login":  true,
		"/api/v1/auth/logout": true,
		"/api/v1/auth/setup":  true,
	}

	// The allow-list is keyed by full path, so the prefixed API must resolve
	// every route to the same verdict as the bare one.
	servers := []struct {
		name        string
		routePrefix string
	}{
		{
			name: "no route prefix",
		},
		{
			name:        "under a route prefix",
			routePrefix: "/argus",
		},
	}

	for _, srv := range servers {
		t.Run(srv.name, func(t *testing.T) {
			opts := []func(*config.Config){withDemoGroup("demo")}
			if srv.routePrefix != "" {
				opts = append(opts, withRoutePrefix(srv.routePrefix))
			}
			file := strings.ReplaceAll(t.Name(), "/", "_") + ".yml"
			api, deps, _ := testAuthServer(t, file, opts...)
			createAuthGroup(t, deps, "demo", "", demoGrants())
			createAuthUser(t, deps, "demo-user", "demo-password", "demo")
			cookie := loginCookie(t, api, "demo-user", "demo-password")

			prefix := strings.TrimSuffix(api.RoutePrefix, "/")
			if got, want := prefix, srv.routePrefix; got != want {
				t.Fatalf("API.RoutePrefix mismatch\ngot:  %q\nwant: %q", got, want)
			}

			type route struct {
				path    string
				methods []string
			}
			var routes []route
			if err := api.Router.Walk(
				func(r *mux.Route, _ *mux.Router, _ []*mux.Route) error {
					path, err := r.GetPathTemplate()
					if err != nil || !strings.HasPrefix(path, prefix+"/api/v1/") {
						return nil
					}
					methods, err := r.GetMethods()
					if err != nil || len(methods) == 0 {
						methods = []string{http.MethodGet}
					}
					routes = append(routes, route{path: path, methods: methods})
					return nil
				},
			); err != nil {
				t.Fatalf("Router.Walk() failed: %v", err)
			}
			if got := len(routes); got < 20 {
				t.Fatalf("len(routes) mismatch\ngot:  %d\nwant: >= 20", got)
			}

			// Walking every route makes an unlisted new one fail here, not in the demo.
			for _, rt := range routes {
				for _, method := range rt.methods {
					t.Run(method+" "+rt.path, func(t *testing.T) {
						unprefixed := strings.TrimPrefix(rt.path, prefix)
						// No session reaches these, so the guard cannot apply, and
						// calling logout would end this one.
						if openRoutes[unprefixed] {
							t.Skip("served without a session")
						}

						// A demo user may only read, and only reads that cost the instance nothing.
						wantRefused := !safeMethod(method) || refusedReads[unprefixed]

						target := strings.ReplaceAll(rt.path, "{id}", "no-such-id")
						req := authedRequest(method, target, "{}", cookie)
						w := serveAuth(api, req)

						if got := isDemoRefusal(w); got != wantRefused {
							t.Errorf(
								"isDemoRefusal() mismatch\ngot:  %t (status %d - %q)\nwant: %t",
								got, w.Code, w.Body.String(), wantRefused,
							)
						}
					})
				}
			}
		})
	}
}

func TestAPI_DemoMiddleware__disarmed(t *testing.T) {
	tests := []struct {
		name      string
		demoGroup string
		// Member of "demo" group, whether it is the configured one or not.
		asMemberOfDemo bool
	}{
		{
			name:           "the caller is not a member",
			demoGroup:      "demo",
			asMemberOfDemo: false,
		},
		{
			name:           "no group configured",
			demoGroup:      "",
			asMemberOfDemo: true,
		},
		{
			name:           "another group is configured",
			demoGroup:      "not-the-demo-group",
			asMemberOfDemo: true,
		},
	}

	routes := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "service order",
			method: http.MethodPut,
			path:   "/service/order",
			body:   `{"order":["test"]}`,
		},
		{
			name:   "latest_version refresh",
			method: http.MethodGet,
			path:   "/latest_version/refresh?service=test",
		},
		{
			name:   "API token create",
			method: http.MethodPost,
			path:   "/tokens",
			body:   `{"name":"token"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: writes the package-level flag vars via testAuthServer's Config.Load.

			file := strings.ReplaceAll(t.Name(), "/", "_") + ".yml"
			api, deps, _ := testAuthServer(t, file, withDemoGroup(tc.demoGroup))
			createAuthGroup(t, deps, "demo", "", demoGrants())
			createAuthUser(t, deps, "demo-user", "demo-password", "demo")

			cookie := loginCookie(t, api, "admin", "admin-password")
			if tc.asMemberOfDemo {
				cookie = loginCookie(t, api, "demo-user", "demo-password")
			}

			for _, route := range routes {
				t.Run(route.name, func(t *testing.T) {
					req := authedRequest(route.method, apiPath(api, route.path), route.body, cookie)
					w := serveAuth(api, req)

					if isDemoRefusal(w) {
						t.Errorf(
							"isDemoRefusal() mismatch\ngot:  true (status %d - %q)\nwant: false",
							w.Code, w.Body.String(),
						)
					}
				})
			}
		})
	}
}

func TestAPI_DemoMiddleware__logoutStillWorks(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file, withDemoGroup("demo"))
	createAuthGroup(t, deps, "demo", "", demoGrants())
	createAuthUser(t, deps, "demo-user", "demo-password", "demo")
	cookie := loginCookie(t, api, "demo-user", "demo-password")

	// Logout sits outside the demo guard.
	req := authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}

// isDemoRefusal reports whether w is the demo guard's read-only refusal.
func isDemoRefusal(w *httptest.ResponseRecorder) bool {
	return w.Code == http.StatusForbidden &&
		strings.Contains(w.Body.String(), errDemoReadOnly.Error())
}

func TestCurrentRoutePath(t *testing.T) {
	tests := []struct {
		name     string
		register func(*mux.Router, http.Handler)
		want     string
	}{
		{
			name: "path/matched route gives its template",
			register: func(r *mux.Router, h http.Handler) {
				r.Handle("/api/v1/users/{id}", h)
			},
			want: "/api/v1/users/{id}",
		},
		{
			name: "no path/request reached no router",
			want: "",
		},
		{
			name: "no path/route matched on method alone",
			register: func(r *mux.Router, h http.Handler) {
				r.Methods(http.MethodGet).Handler(h)
			},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := "handler not reached"
			capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got = currentRoutePath(r)
			})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)

			if tc.register == nil {
				capture.ServeHTTP(httptest.NewRecorder(), req)
			} else {
				router := mux.NewRouter()
				tc.register(router, capture)
				router.ServeHTTP(httptest.NewRecorder(), req)
			}

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
