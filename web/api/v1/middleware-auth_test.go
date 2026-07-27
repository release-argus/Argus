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
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/provider"
	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestAPI_EnableAuth(t *testing.T) {
	// GIVEN: a bare API and auth dependencies.
	api := &API{}
	deps := &AuthDeps{}

	// WHEN: auth is enabled.
	api.EnableAuth(deps)

	prefix := fmt.Sprintf("%s\nAPI.EnableAuth()", packageName)

	// THEN: the dependencies and login limiter are armed.
	if api.auth != deps {
		t.Errorf("%s\nauth deps not stored", prefix)
	}
	if api.loginLimiter == nil {
		t.Errorf("%s\nlogin limiter not initialised", prefix)
	}
}

func TestAuthContextFrom(t *testing.T) {
	prefix := fmt.Sprintf("%s\nauthContextFrom()", packageName)

	// GIVEN: an auth Context.
	authCtx := &auth.Context{
		Identity: auth.Identity{
			Username: "argus",
		},
	}

	// GIVEN: a request without an auth context.
	plain := httptest.NewRequest(http.MethodGet, "/", nil)
	// WHEN/THEN: no context is returned only.
	if got := authContextFrom(plain); got != nil {
		t.Errorf(
			"%s\ncontext mismatch\ngot:  %+v\nwant: nil",
			prefix, got,
		)
	}

	// GIVEN: a request with an auth context.
	carrying := withAuthCtx(httptest.NewRequest(http.MethodGet, "/", nil), authCtx)
	// WHEN/THEN: the context is returned.
	if got := authContextFrom(carrying); got != authCtx {
		t.Errorf(
			"%s\ncontext mismatch\ngot:  %+v\nwant: %+v",
			prefix, got, authCtx,
		)
	}
}

func TestAPI_AuthMiddleware(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_AuthMiddleware.yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	// AND: requests on various endpoints with/without auth.
	tests := []struct {
		name       string
		path       string
		withCookie bool
		wantStatus int
		wantCtx    bool
	}{
		{
			name:       "open path skips authentication",
			path:       "/api/v1/auth/login",
			withCookie: false,
			wantStatus: http.StatusOK,
			wantCtx:    false,
		},
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

			// WHEN: a request passes through the middleware.
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.withCookie {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			prefix := fmt.Sprintf("%s\nauthMiddleware", packageName)

			// THEN: the status and context match expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, w.Code, w.Body.String(), tc.wantStatus,
				)
			}
			if (gotCtx != nil) != tc.wantCtx {
				t.Errorf(
					"%s\ncontext presence mismatch\ngot:  %t\nwant: %t",
					prefix, gotCtx != nil, tc.wantCtx,
				)
			}
		})
	}
}

func TestFailAuthenticationError(t *testing.T) {
	// GIVEN: authentication failures of both classifications (401/500).
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

			// WHEN: the failure is written.
			w := httptest.NewRecorder()
			failAuthenticationError(w, httptest.NewRequest(http.MethodGet, "/", nil), tc.err)

			prefix := fmt.Sprintf("%s\nfailAuthenticationError()", packageName)

			// THEN: the status and message match expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
			if !util.RegexCheck(tc.bodyRegex, w.Body.String()) {
				t.Errorf(
					"%s\nbody mismatch\ngot:  %q\nwant: %q",
					prefix, w.Body.String(), tc.bodyRegex,
				)
			}

			// AND: internal detail is never leaked.
			if strings.Contains(w.Body.String(), "db exploded") {
				t.Errorf(
					"%s\ninternal error detail leaked\ngot:  %q",
					prefix, w.Body.String(),
				)
			}
		})
	}
}

func TestAPI_AuthCtxOr401(t *testing.T) {
	// GIVEN: requests with and without a resolved auth context.
	api := &API{}
	authCtx := &auth.Context{Identity: auth.Identity{Username: "argus"}}

	prefix := fmt.Sprintf("%s\nauthCtxOr401()", packageName)

	// WHEN: the context is absent.
	w := httptest.NewRecorder()
	got := api.authCtxOr401(w, httptest.NewRequest(http.MethodGet, "/", nil))
	// THEN: nil is returned and the request is failed with 401.
	if got != nil || w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nwithout context\ngot:  ctx=%+v, status=%d\nwant: nil, 401",
			prefix, got, w.Code,
		)
	}

	// WHEN: the context is present.
	w = httptest.NewRecorder()
	got = api.authCtxOr401(w, withAuthCtx(httptest.NewRequest(http.MethodGet, "/", nil), authCtx))
	// THEN: it is returned without failing the request.
	if got != authCtx || w.Code != http.StatusOK {
		t.Errorf(
			"%s\nwith context\ngot:  ctx=%+v, status=%d\nwant: the context, 200",
			prefix, got, w.Code,
		)
	}
}

func TestAPI_Authenticate(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_Authenticate.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a session cookie and API token.
	cookie := loginCookie(t, api, "admin", "admin-password")
	adminCtx := adminContext(t, api, deps)
	token, _, err := deps.Store.CreateAPIToken(t.Context(), adminCtx.User.ID, "probe", nil)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}

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
		},
		{
			name:         "stale cookie falls through to a valid Bearer",
			cookie:       "stale",
			bearer:       token,
			wantProvider: "api_token",
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

			// AND: a request to an endpoint requiring auth.
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

			// WHEN: the request is authenticated.
			authCtx, err := api.authenticate(req)

			prefix := fmt.Sprintf("%s\nauthenticate()", packageName)

			// THEN: the error matches expectation.
			errRegex := util.ValueOr(tc.errRegex, `^$`)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s\nerror mismatch\ngot:  %q\nwant: %q",
					prefix, e, errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: the context reflects how it authenticated.
			if authCtx.Identity.Provider != tc.wantProvider {
				t.Errorf(
					"%s\nprovider mismatch\ngot:  %q\nwant: %q",
					prefix, authCtx.Identity.Provider, tc.wantProvider,
				)
			}
		})
	}

	// AND: a session infrastructure failure is not read as unauthorised.
	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime:    time.Hour,
			IdleTimeout: time.Hour,
		},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/flags", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: "anything"})
	if _, err := api.authenticate(req); err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf(
			"%s\nauthenticate() infrastructure failure\ngot:  %v\nwant: a non-unauthorised error",
			packageName, err,
		)
	}
}

func TestBearerToken(t *testing.T) {
	// GIVEN: Authorization headers of varying shapes.
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

			// WHEN: the Bearer token is extracted.
			got, ok := bearerToken(req)

			prefix := fmt.Sprintf("%s\nbearerToken(%q)", packageName, tc.header)

			// THEN: the extraction matches expectations.
			if got != tc.want || ok != tc.wantOK {
				t.Errorf(
					"%s\nmismatch\ngot:  %q, %t\nwant: %q, %t",
					prefix, got, ok, tc.want, tc.wantOK,
				)
			}
		})
	}
}

func TestAPI_AuthenticateAPIToken(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_AuthenticateAPIToken.yml"
	api, deps, dbConn := testAuthServer(t, file)
	// AND: a live token.
	adminCtx := adminContext(t, api, deps)
	live, _, err := deps.Store.CreateAPIToken(t.Context(), adminCtx.User.ID, "live", nil)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}
	// AND: an expired token.
	past := timeNow().UTC().Add(-time.Hour)
	expired, _, err := deps.Store.CreateAPIToken(t.Context(), adminCtx.User.ID, "expired", &past)
	if err != nil {
		t.Fatalf(
			"%s\nsetup CreateAPIToken failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\nauthenticateAPIToken()", packageName)

	// WHEN: a live token authenticates while usage recording is broken.
	if _, err := dbConn.Exec(`
		CREATE TRIGGER fail_touch BEFORE UPDATE ON api_tokens
		BEGIN SELECT RAISE(ABORT, 'touch broke'); END;`,
	); err != nil {
		t.Fatalf(
			"%s\nsetup trigger failed: %v",
			prefix, err,
		)
	}
	authCtx, err := api.authenticateAPIToken(t.Context(), live)
	// THEN: the touch failure is only logged - authentication still succeeds.
	if err != nil || authCtx.Identity.Provider != "api_token" {
		t.Errorf(
			"%s\ntouch failure\ngot:  %+v, err=%v\nwant: authenticated",
			prefix, authCtx, err,
		)
	}
	if _, err := dbConn.Exec(`DROP TRIGGER fail_touch;`); err != nil {
		t.Fatalf(
			"%s\ndrop trigger failed: %v",
			prefix, err,
		)
	}

	// WHEN: a live token authenticates.
	authCtx, err = api.authenticateAPIToken(t.Context(), live)
	// THEN: the owning user's context is resolved via the api_token provider.
	if err != nil || authCtx.Identity.Provider != "api_token" ||
		authCtx.User.ID != adminCtx.User.ID {
		t.Errorf(
			"%s\nlive token\ngot:  %+v, err=%v\nwant: admin via api_token",
			prefix, authCtx, err,
		)
	}

	// AND: usage was recorded.
	tokens, err := deps.Store.APITokensForUser(t.Context(), adminCtx.User.ID)
	if err != nil {
		t.Fatalf(
			"%s\nAPITokensForUser failed: %v",
			prefix, err,
		)
	}
	for _, token := range tokens {
		if token.Name == "live" && token.LastUsedAt == nil {
			t.Errorf("%s\nlast_used_at should have been recorded", prefix)
		}
	}

	// AND: expired and unknown tokens read as unauthorised.
	for _, tc := range []struct {
		name, value string
	}{
		{"expired", expired},
		{"unknown", "argus_" + strings.Repeat("0", 64)},
	} {
		if _, err := api.authenticateAPIToken(t.Context(), tc.value); !errors.Is(err, errUnauthorised) {
			t.Errorf(
				"%s\n%s token\ngot:  %v\nwant: %v",
				prefix, tc.name, err, errUnauthorised,
			)
		}
	}

	// AND: a store failure is not read as unauthorised.
	_ = dbConn.Close()
	_, err = api.authenticateAPIToken(t.Context(), live)
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf(
			"%s\nstore failure\ngot:  %v\nwant: a non-unauthorised error",
			prefix, err,
		)
	}
}

func TestAPI_AuthenticateSession(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_AuthenticateSession.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: a live session.
	adminCtx := adminContext(t, api, deps)
	token, _, err := deps.Sessions.Start(
		t.Context(),
		adminCtx.User.ID,
		"127.0.0.1",
		"go-test",
	)
	if err != nil {
		t.Fatalf(
			"%s\nsetup Start failed: %v",
			packageName, err,
		)
	}

	// WHEN: the session token authenticates.
	authCtx, err := api.authenticateSession(t.Context(), token)

	prefix := fmt.Sprintf("%s\nauthenticateSession()", packageName)

	// THEN: the owning user's context is resolved via the local provider.
	if err != nil ||
		authCtx.Identity.Provider != "local" ||
		authCtx.User.ID != adminCtx.User.ID {
		t.Errorf(
			"%s\nlive session\ngot:  %+v, err=%v\nwant: admin via local",
			prefix, authCtx, err,
		)
	}

	// AND: an unknown token reads as unauthorised.
	_, err = api.authenticateSession(t.Context(), "never-issued")
	if !errors.Is(err, errUnauthorised) {
		t.Errorf(
			"%s\nunknown token\ngot:  %v\nwant: %v",
			prefix, err, errUnauthorised,
		)
	}

	// AND: a session-store failure is not read as unauthorised.
	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime:    time.Hour,
			IdleTimeout: time.Hour,
		},
	)
	_, err = api.authenticateSession(t.Context(), token)
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf(
			"%s\nstore failure\ngot:  %v\nwant: a non-unauthorised error",
			prefix, err,
		)
	}
}

func TestAPI_VerifyLocalCredentials(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_VerifyLocalCredentials.yml"
	api, deps, _ := testAuthServer(t, file)

	prefix := fmt.Sprintf("%s\nverifyLocalCredentials()", packageName)

	// WHEN: valid credentials verify.
	identity, err := api.verifyLocalCredentials(
		t.Context(),
		"admin",
		"admin-password",
		"ip",
	)

	// THEN: the identity is resolved.
	if err != nil || identity.Username != "admin" {
		t.Errorf(
			"%s\nvalid credentials\ngot:  %+v, err=%v\nwant: admin",
			prefix, identity, err,
		)
	}

	// AND: bad credentials return ErrInvalidCredentials, spending the budget.
	for range loginLimitAttempts {
		if _, err := api.verifyLocalCredentials(
			t.Context(),
			"admin",
			"wrong",
			"ip",
		); !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatalf(
				"%s\nbad credentials\ngot:  %v\nwant: %v",
				prefix, err, auth.ErrInvalidCredentials,
			)
		}
	}

	// AND: an exhausted budget rejects even valid credentials, unverified.
	if _, err := api.verifyLocalCredentials(
		t.Context(),
		"admin",
		"admin-password",
		"ip",
	); !errors.Is(err, errTooManyAttempts) {
		t.Errorf(
			"%s\nexhausted budget\ngot:  %v\nwant: %v",
			prefix, err, errTooManyAttempts,
		)
	}

	// AND: an empty provider registry reads as invalid credentials.
	deps.Providers = provider.NewRegistry()
	if _, err := api.verifyLocalCredentials(
		t.Context(),
		"admin",
		"admin-password",
		"other-ip",
	); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf(
			"%s\nno providers\ngot:  %v\nwant: %v",
			prefix, err, auth.ErrInvalidCredentials,
		)
	}
}

func TestAPI_ContextForUser(t *testing.T) {
	// GIVEN: an auth-enabled API with enabled and disabled users.
	file := "TestAPI_ContextForUser.yml"
	api, deps, dbConn := testAuthServer(t, file)
	adminCtx := adminContext(t, api, deps)
	sleeper := createAuthUser(t, deps, "sleeper", "sleeper-password")
	if _, err := deps.Store.UpdateUser(
		t.Context(),
		sleeper.ID,
		store.UserPatch{
			Enabled: new(false),
		},
	); err != nil {
		t.Fatalf(
			"%s\ndisable user: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\ncontextForUser()", packageName)

	// WHEN: an enabled user's context is built.
	authCtx, err := api.contextForUser(t.Context(), adminCtx.User.ID, "local")
	// THEN: it carries the identity, user, grants, and permissions.
	if err != nil ||
		authCtx.Identity.Username != "admin" ||
		authCtx.User.ID != adminCtx.User.ID ||
		len(authCtx.Grants) == 0 ||
		authCtx.Permissions == nil {
		t.Errorf(
			"%s\nenabled user\ngot:  %+v, err=%v",
			prefix, authCtx, err,
		)
	}

	// AND: disabled and unknown users read as unauthorised.
	for _, tc := range []struct {
		name, id string
	}{
		{"disabled", sleeper.ID},
		{"unknown", "no-such-id"},
	} {
		_, err := api.contextForUser(t.Context(), tc.id, "local")
		if !errors.Is(err, errUnauthorised) {
			t.Errorf(
				"%s\n%s user\ngot:  %v\nwant: %v",
				prefix, tc.name, err, errUnauthorised,
			)
		}
	}

	// AND: a grant-load failure is not read as unauthorised.
	if _, err := dbConn.Exec(`DROP TABLE permissions;`); err != nil {
		t.Fatalf(
			"%s\nsetup drop failed: %v",
			packageName, err,
		)
	}
	_, err = api.contextForUser(t.Context(), adminCtx.User.ID, "local")
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf(
			"%s\ngrant-load failure\ngot:  %v\nwant: a non-unauthorised error",
			prefix, err,
		)
	}

	// AND: a user-load failure is not read as unauthorised.
	if _, err := dbConn.Exec(`DROP TABLE users;`); err != nil {
		t.Fatalf(
			"%s\nsetup drop failed: %v",
			packageName, err,
		)
	}
	_, err = api.contextForUser(t.Context(), adminCtx.User.ID, "local")
	if err == nil || errors.Is(err, errUnauthorised) {
		t.Errorf(
			"%s\nuser-load failure\ngot:  %v\nwant: a non-unauthorised error",
			prefix, err,
		)
	}
}

func TestAPI__auth_scopedGrants(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_Auth__ScopedGrants.yml"
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

	// AND: a user whose only grants are scoped to one service and two tags.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"scoped",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
				Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
			},
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
				Scope:      rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"},
			},
			{
				Permission: rbac.Permission{Resource: rbac.ResourceVersionRefresh, Action: rbac.ActionExecute},
				Scope:      rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"},
			},
			{
				Permission: rbac.Permission{Resource: rbac.ResourceVersionRefresh, Action: rbac.ActionExecute},
				Scope:      rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "staging"},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate scoped group: %v",
			packageName, err,
		)
	}
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

			// WHEN: the route is requested.
			w := serveAuth(api,
				authedRequest(http.MethodGet, tc.target, "", cookie))

			prefix := fmt.Sprintf("%s\nscoped grants", packageName)

			// THEN: the guard's verdict matches expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
		})
	}
}

func TestAPI__auth__GuardAnyScope(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_Auth__GuardAnyScope.yml"
	api, deps, _ := testAuthServer(t, file)

	// AND: a user whose only grant is a single service-scoped service:read.
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"scoped",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
				Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate scoped group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "scoped-user", "scoped-password", "scoped")
	// AND: a user with no grants.
	createAuthUser(t, deps, "loner", "loner-password")

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
	}{
		{
			name:       "any service:read grant passes the guard",
			username:   "scoped-user",
			password:   "scoped-password",
			wantStatus: http.StatusOK,
		},
		{
			name:       "no service:read grant is forbidden",
			username:   "loner",
			password:   "loner-password",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginCookie(t, api, tc.username, tc.password)

			// WHEN: the guardAnyScope-protected /service/defaults route is requested.
			w := serveAuth(api,
				authedRequest(http.MethodGet, "/api/v1/service/defaults", "", cookie))

			prefix := fmt.Sprintf("%s\nguardAnyScope", packageName)

			// THEN: the guard's verdict matches expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, w.Code, w.Body.String(), tc.wantStatus,
				)
			}
		})
	}
}

func TestAPI__auth_permissionChangesApplyImmediately(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI__auth_permissionChangesApplyImmediately.yml"
	api, deps, _ := testAuthServer(t, file)

	// AND: a logged-in user in a group with grants.
	group, err := deps.Store.CreateGroup(
		t.Context(),
		"readers",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceConfig, Action: rbac.ActionRead},
				Scope:      rbac.Scope{Type: rbac.ScopeGlobal},
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"%s\ncreate group: %v",
			packageName, err,
		)
	}
	createAuthUser(t, deps, "reader", "reader-password", "readers")
	cookie := loginCookie(t, api, "reader", "reader-password")

	prefix := fmt.Sprintf("%s\nlive grant changes", packageName)

	// WHEN: the user reads flags.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/flags", "", cookie))

	// THEN: they are returned successfully.
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\ninitial read\ngot:  %d\nwant: 200",
			prefix, w.Code,
		)
	}

	// WHEN: the group's grants are emptied.
	empty := []rbac.Grant{}
	if _, err := deps.Store.UpdateGroup(
		t.Context(),
		group.ID,
		store.GroupPatch{Grants: &empty},
	); err != nil {
		t.Fatalf(
			"%s\nupdate group: %v",
			packageName, err,
		)
	}

	// THEN: the same request is now denied.
	w = serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/flags", "", cookie))
	if w.Code != http.StatusForbidden {
		t.Errorf(
			"%s\nread after revoke\ngot:  %d\nwant: 403",
			prefix, w.Code,
		)
	}
}

func TestOriginCheckMiddleware(t *testing.T) {
	// GIVEN: requests with varying Origin headers against host example.com.
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

			// WHEN: a request carries the Origin header.
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

			prefix := fmt.Sprintf("%s\noriginCheckMiddleware()", packageName)

			// THEN: the verdict matches expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
		})
	}
}

func TestAPI_Guard(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_Guard.yml"
	api, deps, _ := testAuthServer(t, file)
	// AND: an admin.
	adminCtx := adminContext(t, api, deps)
	if _, err := deps.Store.CreateGroup(
		t.Context(),
		"scoped",
		"",
		[]rbac.Grant{
			{
				Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
				Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
			},
		},
	); err != nil {
		t.Fatalf(
			"%s\ncreate scoped group: %v",
			packageName, err,
		)
	}
	// AND: a user scoped to one service.
	scopedUser := createAuthUser(t, deps, "scoped-user", "scoped-password", "scoped")
	scopedCtx, err := api.contextForUser(t.Context(), scopedUser.ID, "local")
	if err != nil {
		t.Fatalf(
			"%s\nscoped context: %v",
			packageName, err,
		)
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

			testAPI := api
			if tc.disableAuth {
				testAPI = &API{}
			}

			called := false
			handler := testAPI.guard(rbac.ResourceService, rbac.ActionRead, tc.target,
				func(w http.ResponseWriter, _ *http.Request) {
					called = true
					w.WriteHeader(http.StatusOK)
				},
			)

			// WHEN: a request hits the guarded handler.
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authCtx != nil {
				req = withAuthCtx(req, tc.authCtx)
			}
			w := httptest.NewRecorder()
			handler(w, req)

			prefix := fmt.Sprintf("%s\nguard()", packageName)

			// THEN: the response code and handler invocation match expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
			if called != tc.wantCalled {
				t.Errorf(
					"%s\nhandler called mismatch\ngot:  %t\nwant: %t",
					prefix, called, tc.wantCalled,
				)
			}
		})
	}
}

func TestAPI_RequireAdmin(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_RequireAdmin.yml"
	api, deps, _ := testAuthServer(t, file)

	// AND: an admin.
	adminCtx := adminContext(t, api, deps)
	// AND: a viewer.
	viewer := createAuthUser(t, deps, "viewer-user", "viewer-password", store.GroupViewer)
	viewerCtx, err := api.contextForUser(t.Context(), viewer.ID, "local")
	if err != nil {
		t.Fatalf(
			"%s\nviewer context: %v",
			packageName, err,
		)
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

			testAPI := api
			if tc.disableAuth {
				testAPI = &API{}
			}

			called := false
			handler := testAPI.requireAdmin(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			// WHEN: a request hits the guarded handler.
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authCtx != nil {
				req = withAuthCtx(req, tc.authCtx)
			}
			w := httptest.NewRecorder()
			handler(w, req)

			prefix := fmt.Sprintf("%s\nrequireAdmin()", packageName)

			// THEN: the response code and handler invocation match expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d\nwant: %d",
					prefix, w.Code, tc.wantStatus,
				)
			}
			if called != tc.wantCalled {
				t.Errorf(
					"%s\nhandler called mismatch\ngot:  %t\nwant: %t",
					prefix, called, tc.wantCalled,
				)
			}
		})
	}
}

func TestAPI_AllowedServices(t *testing.T) {
	// GIVEN: an auth-enabled API with tagged services
	file := "TestAPI_AllowedServices.yml"
	api, deps, _ := testAuthServer(t, file)
	api.Config.OrderMu.Lock()
	api.Config.Service["test"] = &service.Service{}
	api.Config.Service["tagged"] = &service.Service{
		Dashboard: dashboard.Options{Tags: []string{"prod"}},
	}
	api.Config.Order = append(api.Config.Order, "tagged")
	api.Config.OrderMu.Unlock()

	// AND: users of differing grant scopes.
	groups := []struct {
		name   string
		grants []rbac.Grant
	}{
		{
			name: "full-scoped",
			grants: []rbac.Grant{
				{
					Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
					Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
				},
				{
					Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
					Scope:      rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"},
				},
			},
		},
		{
			name: "service-scoped",
			grants: []rbac.Grant{
				{
					Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
					Scope:      rbac.Scope{Type: rbac.ScopeService, Ref: "test"},
				},
			},
		},
		{
			name: "tag-scoped",
			grants: []rbac.Grant{
				{
					Permission: rbac.Permission{Resource: rbac.ResourceService, Action: rbac.ActionRead},
					Scope:      rbac.Scope{Type: rbac.ScopeServiceTag, Ref: "prod"},
				},
			},
		},
	}

	for _, group := range groups {
		if _, err := deps.Store.CreateGroup(
			t.Context(),
			group.name,
			"",
			group.grants,
		); err != nil {
			t.Fatalf(
				"%s\ncreate group %q: %v",
				packageName, group.name, err,
			)
		}
	}
	createAuthUser(t, deps, "full-scoped-user", "scoped-password", "full-scoped")
	createAuthUser(t, deps, "service-scoped-user", "scoped-password", "service-scoped")
	createAuthUser(t, deps, "tag-scoped-user", "scoped-password", "tag-scoped")
	createAuthUser(t, deps, "loner", "loner-password")

	tests := []struct {
		name     string
		username string // "" = the bootstrap admin.
		wantNil  bool
		want     []string
	}{
		{
			name:    "global service:read is unrestricted (nil)",
			wantNil: true,
		},
		{
			name:     "full-scoped grants",
			username: "full-scoped-user",
			want:     []string{"test", "tagged"},
		},
		{
			name:     "service-scoped grants",
			username: "service-scoped-user",
			want:     []string{"test"},
		},
		{
			name:     "tag-scoped grants",
			username: "tag-scoped-user",
			want:     []string{"tagged"},
		},
		{
			name:     "no grants allow nothing",
			username: "loner",
			want:     []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			username := util.ValueOr(tc.username, "admin")
			creds, err := deps.Store.LocalCredentials(t.Context(), username)
			if err != nil || creds == nil {
				t.Fatalf(
					"%s\nlookup %q failed: %v",
					packageName, username, err,
				)
			}
			authCtx, err := api.contextForUser(t.Context(), creds.UserID, "local")
			if err != nil {
				t.Fatalf(
					"%s\ncontext failed: %v",
					packageName, err,
				)
			}

			// WHEN: the permitted-service set is derived.
			allowed := api.allowedServices(authCtx)

			prefix := fmt.Sprintf("%s\nallowedServices()", packageName)

			// THEN: it matches expectations.
			if tc.wantNil {
				if allowed != nil {
					t.Errorf(
						"%s\nexpected nil (unrestricted)\ngot: %v",
						prefix, allowed,
					)
				}
				return
			}
			if allowed == nil || len(allowed) != len(tc.want) {
				t.Fatalf(
					"%s\nset mismatch\ngot:  %v\nwant: %v",
					prefix, allowed, tc.want,
				)
			}
			for _, serviceID := range tc.want {
				if !allowed[serviceID] {
					t.Errorf(
						"%s\nmissing %q\ngot: %v",
						prefix, serviceID, allowed,
					)
				}
			}
		})
	}
}

func TestAPI_KickWebSocketClients(t *testing.T) {
	prefix := fmt.Sprintf("%s\nkick*WebSocketClients()", packageName)

	// kickAll exercises every kick helper against the client's identity.
	kickAll := func(api *API, client *Client) {
		api.kickUserWebSocketClients(client.userID)
		api.kickSessionWebSocketClients(client.sessionHash)
		api.kickRestrictedWebSocketClients()
	}

	// GIVEN: an API without auth/hub wiring.
	file := "TestAPI_KickWebSocketClients.yml"
	apiValue := testAPI(t, file)
	apiNoAuth := &apiValue
	client := &Client{
		userID:          "user-a",
		sessionHash:     "hash-1",
		allowedServices: map[string]bool{"svc": true},
	}
	// WHEN/THEN: no auth -> no-op.
	kickAll(apiNoAuth, client)
	// AND: API without a hub -> no-op too.
	apiNoAuth.auth = &AuthDeps{}
	kickAll(apiNoAuth, client)
	// AND: API without auth (but a hub) -> no-op too.
	apiNoAuth.auth = nil
	apiNoAuth.hub = NewHub()
	go apiNoAuth.hub.Run()
	client.hub = apiNoAuth.hub
	client.send = make(chan []byte, 8)
	apiNoAuth.hub.register <- client
	kickAll(apiNoAuth, client)
	if !apiNoAuth.hub.hasClient(client) {
		t.Errorf("%s\nno-auth API should not kick clients", prefix)
	}

	// GIVEN: API with auth and hub wired, with connected clients.
	apiWired := &API{auth: &AuthDeps{}}
	newClient := wireHub(t, apiWired)

	// WHEN: a user's clients are kicked.
	target := newClient("user-a", "hash-1")
	other := newClient("user-b", "hash-2")
	apiWired.kickUserWebSocketClients("user-a")
	// THEN: only that user's client is kicked.
	if apiWired.hub.hasClient(target) || !apiWired.hub.hasClient(other) {
		t.Errorf("%s\nwired API should kick exactly user-a's clients", prefix)
	}

	// WHEN: a session's clients are kicked.
	target = newClient("user-b", "hash-3")
	apiWired.kickSessionWebSocketClients("hash-3")
	// THEN: only that session's client is kicked.
	if apiWired.hub.hasClient(target) || !apiWired.hub.hasClient(other) {
		t.Errorf("%s\nwired API should kick exactly hash-3's clients", prefix)
	}

	// WHEN: the restricted clients are kicked.
	target = newClient("user-c", "hash-4", map[string]bool{"svc": true})
	apiWired.kickRestrictedWebSocketClients()
	// THEN: only the restricted client is kicked.
	if apiWired.hub.hasClient(target) || !apiWired.hub.hasClient(other) {
		t.Errorf("%s\nwired API should kick exactly the restricted clients", prefix)
	}
}

func TestAPI_ServiceTarget(t *testing.T) {
	// GIVEN: an auth-enabled API whose config holds a tagged service.
	file := "TestAPI_ServiceTarget.yml"
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

			// WHEN: the target is extracted from a request.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/service/summary?"+tc.query, nil)
			target := api.serviceTarget(req)

			prefix := fmt.Sprintf(
				"%s\nserviceTarget(%s)",
				packageName, tc.query,
			)

			// THEN: the target matches expectations.
			if tc.wantNil {
				if target != nil {
					t.Errorf(
						"%s\nexpected nil target\ngot: %+v",
						prefix, target,
					)
				}
				return
			}
			if target == nil || target.ServiceID != tc.wantID || len(target.Tags) != tc.wantTags {
				t.Errorf(
					"%s\ntarget mismatch\ngot:  %+v\nwant: id=%q with %d tags",
					prefix, target, tc.wantID, tc.wantTags,
				)
			}
		})
	}
}
