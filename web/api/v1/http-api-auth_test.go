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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/provider"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// ghostProvider pretends to be the local provider but authenticates
// subjects that do not exist in the user store.
type ghostProvider struct{}

func (ghostProvider) Name() string { return "local" }

func (ghostProvider) Authenticate(_ context.Context, _, _ string) (*auth.Identity, error) {
	return &auth.Identity{Provider: "local", Subject: "ghost"}, nil
}

// realSubjectProvider authenticates anything as the bootstrap admin.
type realSubjectProvider struct {
	store *store.Store
}

func (realSubjectProvider) Name() string { return "local" }

func (p realSubjectProvider) Authenticate(ctx context.Context, _, _ string) (*auth.Identity, error) {
	creds, err := p.store.LocalCredentials(ctx, "admin")
	if err != nil || creds == nil {
		return nil, auth.ErrInvalidCredentials
	}
	return &auth.Identity{Provider: "local", Subject: creds.UserID}, nil
}

func TestLoginKey(t *testing.T) {
	// Long enough that storing it raw would bloat the bucket map.
	long := strings.Repeat("a", maxAuthBodySize)

	tests := []struct {
		name     string
		ip       string
		username string
		sameAs   string
	}{
		{name: "distinct/shares a prefix", ip: "10.0.0.1", username: "admin"},
		{name: "distinct/longer name sharing that prefix", ip: "10.0.0.1", username: "administrator"},
		{name: "distinct/same name from another address", ip: "10.0.0.2", username: "admin"},
		{name: "distinct/oversized name", ip: "10.0.0.1", username: long},
		{name: "equal/same ip and name", ip: "10.0.0.1", username: "admin", sameAs: "admin"},
		{name: "equal/case variant shares a bucket", ip: "10.0.0.1", username: "ADMIN", sameAs: "admin"},
	}

	seen := map[string]string{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares seen.

			got := loginKey(tc.ip, tc.username)

			if gotLen, want := len(got), len(loginKey(tc.ip, "x")); gotLen != want {
				t.Errorf("len(loginKey()) mismatch\ngot:  %d\nwant: %d", gotLen, want)
			}

			if tc.sameAs != "" {
				if want := loginKey(tc.ip, tc.sameAs); got != want {
					t.Errorf("loginKey() mismatch\ngot:  %q\nwant: %q", got, want)
				}
				return
			}

			if clash, ok := seen[got]; ok {
				t.Errorf(
					"loginKey() mismatch\ngot:  %q\nwant: NOT %q (key of username=%q)",
					got, got, clash,
				)
			}
			seen[got] = tc.username
		})
	}
}

func TestLoginLimiter_LockKey(t *testing.T) {
	limiter := newLoginLimiter()
	base := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	timeNowHad := timeNow
	timeNow = func() time.Time { return base }
	t.Cleanup(func() { timeNow = timeNowHad })

	key := "ip"
	var passed atomic.Int64
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			defer limiter.lockKey(key)()
			if limiter.check(key, loginLimitAttempts) {
				passed.Add(1)
				limiter.recordFailure(key)
			}
		})
	}
	wg.Wait()

	if got := passed.Load(); got != loginLimitAttempts {
		t.Errorf("passed.Load() mismatch\ngot:  %d\nwant: %d", got, loginLimitAttempts)
	}

	limiter.mu.Lock()
	window := limiter.windows[key]
	limiter.mu.Unlock()
	if window == nil {
		t.Fatal("loginLimiter.windows[key] mismatch\ngot:  absent\nwant: present")
	}
	limiter.mu.Lock()
	refs := window.refs
	limiter.mu.Unlock()
	if refs != 0 {
		t.Errorf("loginWindow.refs mismatch\ngot:  %d\nwant: 0", refs)
	}
}

func TestLoginLimiter__expiredWindowsSweptPeriodically(t *testing.T) {
	limiter := newLoginLimiter()
	base := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	now := base
	timeNowHad := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = timeNowHad })

	// The expired window survives a sweep because an attempt still holds it.
	heldKey := "held"
	limiter.recordFailure(heldKey)
	release := limiter.lockKey(heldKey)

	now = base.Add(loginLimitWindow + time.Second)
	limiter.recordFailure("other")
	release()

	now = now.Add(time.Second)
	limiter.recordFailure("later")

	// The scan is skipped, rather than repeated on every call.
	limiter.mu.Lock()
	_, keptEarly := limiter.windows[heldKey]
	limiter.mu.Unlock()
	if !keptEarly {
		t.Error(
			"within the interval: loginLimiter.windows[heldKey] mismatch\n" +
				"got:  absent\nwant: present",
		)
	}

	now = now.Add(loginLimitWindow + time.Second)
	limiter.recordFailure("latest")

	// Dropped, so the map cannot grow unbounded.
	limiter.mu.Lock()
	_, keptLate := limiter.windows[heldKey]
	limiter.mu.Unlock()
	if keptLate {
		t.Error(
			"after a full window: loginLimiter.windows[heldKey] mismatch\n" +
				"got:  present\nwant: absent",
		)
	}
}

func TestLoginLimiter_Check_and_RecordFailure(t *testing.T) {
	limiter := newLoginLimiter()
	base := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	now := base
	timeNowHad := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = timeNowHad })

	key := "ip"
	// check() alone never trips: only recorded failures count.
	for i := range loginLimitAttempts * 2 {
		if !limiter.check(key, loginLimitAttempts) {
			t.Fatalf("check %d: loginLimiter.check() mismatch\ngot:  false\nwant: true", i)
		}
	}

	for i := range loginLimitAttempts {
		if !limiter.check(key, loginLimitAttempts) {
			t.Fatalf("attempt %d: loginLimiter.check() mismatch\ngot:  false\nwant: true", i)
		}
		limiter.recordFailure(key)
	}

	if limiter.check(key, loginLimitAttempts) {
		t.Error("past the limit: loginLimiter.check() mismatch\ngot:  true\nwant: false")
	}

	if !limiter.check("other-ip", loginLimitAttempts) {
		t.Error("other IP: loginLimiter.check() mismatch\ngot:  false\nwant: true")
	}

	now = base.Add(loginLimitWindow + time.Second)
	if !limiter.check(key, loginLimitAttempts) {
		t.Error("expired window: loginLimiter.check() mismatch\ngot:  false\nwant: true")
	}

	now = base
	heldKey := "held"
	for range loginLimitAttempts {
		limiter.recordFailure(heldKey)
	}
	if limiter.check(heldKey, loginLimitAttempts) {
		t.Fatal("held key past the limit: loginLimiter.check() mismatch\ngot:  true\nwant: false")
	}

	release := limiter.lockKey(heldKey) // An in-flight attempt holds the key.
	t.Cleanup(release)
	now = base.Add(loginLimitWindow + time.Second)
	// A window held by an in-flight attempt is reset in place, not swept away.
	if !limiter.check(heldKey, loginLimitAttempts) {
		t.Error("expired held window: loginLimiter.check() mismatch\ngot:  false\nwant: true")
	}
}

func TestLoginLimiter_Check_and_RecordFailure__keyedPerUsername(t *testing.T) {
	limiter := newLoginLimiter()
	base := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	timeNowHad := timeNow
	timeNow = func() time.Time { return base }
	t.Cleanup(func() { timeNow = timeNowHad })

	const sharedIP = "10.0.0.1"
	attacker := loginKey(sharedIP, "attacker")
	victim := loginKey(sharedIP, "victim")

	for range loginLimitAttempts {
		limiter.recordFailure(attacker)
	}

	if limiter.check(attacker, loginLimitAttempts) {
		t.Error("attacker: loginLimiter.check() mismatch\ngot:  true\nwant: false")
	}

	// A shared proxy IP must not lock everyone out.
	if !limiter.check(victim, loginLimitAttempts) {
		t.Error("victim: loginLimiter.check() mismatch\ngot:  false\nwant: true")
	}
}

func TestLoginLimiter__perIPCapAcrossUsernames(t *testing.T) {
	limiter := newLoginLimiter()
	base := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	timeNowHad := timeNow
	timeNow = func() time.Time { return base }
	t.Cleanup(func() { timeNow = timeNowHad })

	const sprayIP = "10.0.0.9"

	for i := range loginLimitPerIPAttempts {
		key := loginKey(sprayIP, fmt.Sprintf("user-%d", i))
		limiter.recordFailure(key)
		limiter.recordFailure(sprayIP)
	}

	if limiter.check(sprayIP, loginLimitPerIPAttempts) {
		t.Error("sprayed IP: loginLimiter.check() mismatch\ngot:  true\nwant: false")
	}

	// The per-IP cap is enforced separately, by API.verifyLocalCredentials.
	if !limiter.check(loginKey(sprayIP, "fresh"), loginLimitAttempts) {
		t.Error("fresh username: loginLimiter.check() mismatch\ngot:  false\nwant: true")
	}

	if !limiter.check("10.0.0.1", loginLimitPerIPAttempts) {
		t.Error("other IP: loginLimiter.check() mismatch\ngot:  false\nwant: true")
	}
}

func TestAPI_AuthLogin(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	disabledUser := createAuthUser(t, deps, "sleeper", "sleeper-password", store.GroupViewer)
	if _, err := deps.Store.UpdateUser(t.Context(), disabledUser.ID,
		store.UserPatch{Enabled: new(false)}); err != nil {
		t.Fatalf("Store.UpdateUser() failed: %v", err)
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCookie bool
	}{
		{
			name:       "valid credentials",
			body:       `{"username":"admin","password":"admin-password"}`,
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "wrong password",
			body:       `{"username":"admin","password":"wrong"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unknown user",
			body:       `{"username":"ghost","password":"whatever"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "disabled user with correct password",
			body:       `{"username":"sleeper","password":"sleeper-password"}`,
			wantStatus: http.StatusUnauthorized, // The provider refuses it as bad credentials.
		},
		{
			name:       "malformed body",
			body:       `{"username":`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tc.body))
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Fatalf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			gotCookie := false
			for _, cookie := range w.Result().Cookies() {
				if cookie.Name == authCookieName && cookie.Value != "" {
					gotCookie = true

					if !cookie.HttpOnly {
						t.Error("Cookie.HttpOnly mismatch\ngot:  false\nwant: true")
					}
					if got, want := cookie.SameSite, http.SameSiteStrictMode; got != want {
						t.Errorf(
							"Cookie.SameSite mismatch\n"+
								"got:  %d\nwant: %d (http.SameSiteStrictMode)",
							got, want,
						)
					}
				}
			}
			if gotCookie != tc.wantCookie {
				t.Errorf("session cookie mismatch\ngot:  %t\nwant: %t", gotCookie, tc.wantCookie)
			}

			if tc.wantStatus == http.StatusOK {
				var response apitype.AuthMe
				if err := decode.Unmarshal("json", w.Body.Bytes(), &response); err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}
				wantUsername := "admin"
				if response.User.Username != wantUsername || len(response.Permissions) == 0 {
					t.Errorf(
						"AuthMe mismatch\n"+
							"got:  User.Username=%q, len(Permissions)=%d\n"+
							"want: User.Username=%q, len(Permissions)>0",
						response.User.Username, len(response.Permissions), wantUsername,
					)
				}
			}
		})
	}
}

func TestAPI_AuthLogin__perIPRateLimit(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)

	// Keep each (IP, username) bucket under its cap while spending the per-IP budget.
	perUser := loginLimitAttempts - 1
	sprayUsers := loginLimitPerIPAttempts / perUser
	for u := range sprayUsers {
		for a := range perUser {
			body := fmt.Sprintf(`{"username":"spray-%d","password":"wrong"}`, u)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
			w := serveAuth(api, req)
			if got, want := w.Code, http.StatusUnauthorized; got != want {
				t.Fatalf(
					"spray user %d: attempt %d: status code mismatch\ngot:  %d - %q\nwant: %d",
					u, a, got, w.Body.String(), want,
				)
			}
		}
	}

	req := httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`),
	)

	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusTooManyRequests; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}

func TestAPI_AuthLogin__rateLimit(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)

	var last *httptest.ResponseRecorder
	for range loginLimitAttempts + 1 {
		req := httptest.NewRequest(
			http.MethodPost, "/api/v1/auth/login",
			strings.NewReader(`{"username":"admin","password":"wrong"}`),
		)
		last = serveAuth(api, req)
	}

	if got, want := last.Code, http.StatusTooManyRequests; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, last.Body.String(), want)
	}
}

func TestAPI_AuthLogin__sessionCap(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	username, password := "admin", "admin-password"
	oldestCookie := loginCookie(t, api, username, password)
	for range session.DefaultMaxSessionsPerUser - 1 {
		_ = loginCookie(t, api, username, password)
	}
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	oldestClient := connect(adminID, auth.HashToken(oldestCookie.Value))

	newestCookie := loginCookie(t, api, username, password)

	if api.hub.hasClient(oldestClient) {
		t.Error("oldest session: Hub.hasClient() mismatch\ngot:  true\nwant: false")
	}

	req := authedRequest(http.MethodGet, "/api/v1/auth/me", "", oldestCookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"evicted session: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", newestCookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf(
			"newest session: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_Auth__noProviders(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	deps.Providers = provider.NewRegistry()

	req := httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`),
	)
	w := serveAuth(api, req)

	// 401, not 500: a missing local provider counts as bad credentials.
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}
}

func TestAPI_Auth__loginEdgeCases(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, dbConn := testAuthServer(t, file)

	oversize := `{"username":"` + strings.Repeat("x", maxAuthBodySize) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(oversize))

	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf(
			"login body over the cap: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	deps.Providers = provider.NewRegistry()
	deps.Providers.Register(ghostProvider{})
	req = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"ghost","password":"x"}`),
	)

	w = serveAuth(api, req)
	// 401, not 500: API.contextForUser reads a vanished user as unauthorised.
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"login for an unresolvable subject: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	test.DropTable(t, dbConn, "permissions")
	req = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"ghost","password":"x"}`),
	)
	deps.Providers = provider.NewRegistry()
	deps.Providers.Register(realSubjectProvider{store: deps.Store})

	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"grant failure during login: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_Auth__sessionManagerFailures(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	// Swap in a manager whose store always fails (and has an empty cache).
	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)

	req := authedRequest(http.MethodGet, "/api/v1/flags", "", cookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"session validation: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`),
	)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("login: status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	req = authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"logout: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w = serveAuth(api, req)
	// There is no session to revoke.
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Errorf(
			"cookie-less logout: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_AuthSetup(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServerPendingSetup(t, file)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup", nil)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"setup state fetch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var state apitype.SetupState
	if err := decode.Unmarshal("json", w.Body.Bytes(), &state); err != nil {
		t.Fatalf("setup state fetch: decode.Unmarshal() failed: %v", err)
	}
	if !state.SetupRequired {
		t.Error("pending: SetupState.SetupRequired mismatch\ngot:  false\nwant: true")
	}

	if state.Demo != nil {
		t.Errorf(
			"pending: SetupState.Demo mismatch\ngot:  Username=%q\nwant: nil",
			state.Demo.Username,
		)
	}

	api.Config.Settings.Web.Demo = &config.WebSettingsDemo{
		Username: "demo-user",
		Password: "demo-pass",
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup", nil)
	w = serveAuth(api, req)
	state = apitype.SetupState{}
	if err := decode.Unmarshal("json", w.Body.Bytes(), &state); err != nil {
		t.Fatalf("demo state fetch: decode.Unmarshal() failed: %v", err)
	}
	if state.Demo == nil {
		t.Fatal("demo: SetupState.Demo mismatch\ngot:  nil\nwant: non-nil")
	}
	wantDemo := apitype.DemoCredentials{Username: "demo-user", Password: "demo-pass"}
	if *state.Demo != wantDemo {
		t.Errorf(
			"demo: SetupState.Demo mismatch\n"+
				"got:  Username=%q, Password=%q\nwant: Username=%q, Password=%q",
			state.Demo.Username, state.Demo.Password,
			wantDemo.Username, wantDemo.Password,
		)
	}
	api.Config.Settings.Web.Demo = nil

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"request without a session: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	for _, invalid := range []struct {
		name, body string
	}{
		{name: "empty", body: `{}`},
		{name: "no password", body: `{"username":"root"}`},
		{name: "no username", body: `{"password":"setup-password"}`},
		{
			name: "over-long display name",
			body: `{"username":"root","password":"setup-password","display_name":"` +
				strings.Repeat("x", maxFieldLength+1) + `"}`,
		},
	} {
		t.Run("invalid/"+invalid.name, func(t *testing.T) {
			// Serial: shares the pending setup, which the parent completes after the loop.

			req := httptest.NewRequest(
				http.MethodPost, "/api/v1/auth/setup", strings.NewReader(invalid.body),
			)
			w := serveAuth(api, req)
			if got, want := w.Code, http.StatusBadRequest; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
		})
	}

	username, displayName, groups := "root", "Rooty", []string{store.GroupAdmin}
	req = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/setup",
		strings.NewReader(`{"username":"`+username+`","display_name":"`+displayName+`","password":"setup-password"}`),
	)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"first administrator create: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var me apitype.AuthMe
	if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
		t.Fatalf("first administrator create: decode.Unmarshal() failed: %v", err)
	}
	if me.User.Username != username || me.User.DisplayName != displayName {
		t.Errorf(
			"first administrator create: AuthMe.User mismatch\n"+
				"got:  Username=%q, DisplayName=%q\nwant: Username=%q, DisplayName=%q",
			me.User.Username, me.User.DisplayName,
			username, displayName,
		)
	}
	if testErr := test.AssertSlicesEqual(t, me.User.Groups, groups, "AuthMe.User.Groups"); testErr != nil {
		t.Errorf("first administrator create: %v", testErr)
	}
	if len(me.Permissions) == 0 {
		t.Error(
			"first administrator create: AuthMe.Permissions mismatch\ngot:  []\nwant: non-empty",
		)
	}
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == authCookieName {
			cookie = c
		}
	}
	if cookie == nil || cookie.Value == "" {
		t.Fatal("first administrator create: session cookie mismatch\ngot:  absent\nwant: present")
	}

	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf(
			"auth/me with the setup cookie: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup", nil)
	w = serveAuth(api, req)
	state = apitype.SetupState{}
	if err := decode.Unmarshal("json", w.Body.Bytes(), &state); err != nil {
		t.Fatalf("completed state fetch: decode.Unmarshal() failed: %v", err)
	}
	if state.SetupRequired {
		t.Error("complete: SetupState.SetupRequired mismatch\ngot:  true\nwant: false")
	}

	req = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/setup",
		strings.NewReader(`{"username":"root2","password":"another-password"}`),
	)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusConflict; got != want {
		t.Errorf(
			"repeat setup: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	count, err := deps.Store.CountUsers(t.Context())
	if err != nil {
		t.Fatalf("Store.CountUsers() failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Store.CountUsers() mismatch\ngot:  %d\nwant: 1", count)
	}
}

func TestAPI_AuthSetup__concurrent(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServerPendingSetup(t, file)

	var wg sync.WaitGroup
	start := make(chan struct{})
	codes := make([]int, 2)
	for i := range codes {
		wg.Go(func() {
			body := fmt.Sprintf(`{"username":"admin-%d","password":"str0ng-pass"}`, i)
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup", strings.NewReader(body))
			codes[i] = serveAuth(api, req).Code
		})
	}
	close(start)
	wg.Wait()

	slices.Sort(codes)
	if codes[0] != http.StatusCreated || codes[1] != http.StatusConflict {
		t.Errorf("status code mismatch\ngot:  %v\nwant: [201 409]", codes)
	}
}

func TestAPI_AuthSetup__errors(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		closeDB    bool
		dropTable  string // Table dropped before the request.
		wantStatus int
		bodyRegex  string
	}{
		{
			name:       "malformed JSON",
			method:     http.MethodPost,
			body:       `{"username":`,
			wantStatus: http.StatusBadRequest,
			bodyRegex:  `parse request:`,
		},
		{
			name:       "state read failure",
			method:     http.MethodGet,
			closeDB:    true,
			wantStatus: http.StatusInternalServerError,
			bodyRegex:  `failed to read setup state`,
		},
		{
			name:       "create failure",
			method:     http.MethodPost,
			body:       `{"username":"root","password":"setup-password"}`,
			closeDB:    true,
			wantStatus: http.StatusInternalServerError,
			bodyRegex:  `"setup failed"`,
		},
		{
			name:       "login-after-create failure",
			method:     http.MethodPost,
			body:       `{"username":"root","password":"setup-password"}`,
			dropTable:  "permissions", // CreateFirstAdmin survives; the grant load cannot.
			wantStatus: http.StatusInternalServerError,
			bodyRegex:  `setup succeeded but login failed`,
		},
		{
			name:       "session mint failure",
			method:     http.MethodPost,
			body:       `{"username":"root","password":"setup-password"}`,
			dropTable:  "sessions", // Everything up to the session insert survives.
			wantStatus: http.StatusInternalServerError,
			bodyRegex:  `authentication failed`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			file := strings.ReplaceAll(t.Name(), "/", "_") + ".yml"
			api, _, dbConn := testAuthServerPendingSetup(t, file)
			if tc.closeDB {
				_ = dbConn.Close()
			}
			if tc.dropTable != "" {
				test.DropTable(t, dbConn, tc.dropTable)
			}

			req := httptest.NewRequest(tc.method, "/api/v1/auth/setup", strings.NewReader(tc.body))
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			if !util.RegexCheck(tc.bodyRegex, w.Body.String()) {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", w.Body.String(), tc.bodyRegex)
			}
		})
	}
}

func TestAPI__AuthMe_and_Logout(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	req := authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var me apitype.AuthMe
	if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
		t.Fatalf("auth/me: decode.Unmarshal() failed: %v", err)
	}
	wantUsername := "admin"
	if me.User.Username != wantUsername || len(me.Permissions) == 0 {
		t.Errorf(
			"auth/me: AuthMe mismatch\n"+
				"got:  User.Username=%q, len(Permissions)=%d\n"+
				"want: User.Username=%q, len(Permissions)>0",
			me.User.Username, len(me.Permissions), wantUsername,
		)
	}

	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", nil)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"auth/me without credentials: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie)
	w = serveAuth(api, req)

	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"logout: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	assertSessionCookieCleared(t, w)

	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"auth/me on the revoked session: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_AuthLogout__kicks(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")
	otherCookie := loginCookie(t, api, "admin", "admin-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	logoutClient := connect(adminID, auth.HashToken(cookie.Value))
	otherClient := connect(adminID, auth.HashToken(otherCookie.Value))

	req := authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"first session logout: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	gotLogout, gotOther := api.hub.hasClient(logoutClient), api.hub.hasClient(otherClient)
	if gotLogout || !gotOther {
		t.Errorf(
			"Hub.hasClient() mismatch\ngot:  logout=%t, other=%t\nwant: logout=false, other=true",
			gotLogout, gotOther,
		)
	}
}

func TestAPI_AuthMeUpdate__password(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")
	otherCookie := loginCookie(t, api, "admin", "admin-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	changingClient := connect(adminID, auth.HashToken(cookie.Value))
	otherClient := connect(adminID, auth.HashToken(otherCookie.Value))

	req := authedRequest(
		http.MethodPatch, "/api/v1/auth/me",
		test.TrimJSON(`{
			"current_password":"admin-password",
			"new_password":"new-admin-password"
		}`),
		cookie,
	)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"password change: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	var me apitype.AuthMe
	if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
		t.Fatalf("password change: decode.Unmarshal() failed: %v", err)
	}
	if got, want := me.User.Username, "admin"; got != want {
		t.Errorf("password change: AuthMe.User.Username mismatch\ngot:  %q\nwant: %q", got, want)
	}

	var fresh *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == authCookieName {
			fresh = c
		}
	}
	if fresh == nil || fresh.Value == "" {
		t.Fatal("password change: session cookie mismatch\ngot:  absent\nwant: present")
	}
	if fresh.Value == cookie.Value {
		t.Fatal("password change: Cookie.Value mismatch\ngot:  unchanged\nwant: rotated")
	}

	gotChanging, gotOther := api.hub.hasClient(changingClient), api.hub.hasClient(otherClient)
	if gotChanging || gotOther {
		t.Errorf(
			"password change: Hub.hasClient() mismatch\n"+
				"got:  changing=%t, other=%t\nwant: changing=false, other=false",
			gotChanging, gotOther,
		)
	}

	for _, stale := range []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "changing session", cookie: cookie},
		{name: "other session", cookie: otherCookie},
	} {
		t.Run("stale/"+stale.name, func(t *testing.T) {
			// Serial: checks sessions before the parent logs in again after the loop.

			staleReq := authedRequest(http.MethodGet, "/api/v1/auth/me", "", stale.cookie)
			staleW := serveAuth(api, staleReq)
			if got, want := staleW.Code, http.StatusUnauthorized; got != want {
				t.Errorf(
					"auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
					got, staleW.Body.String(), want,
				)
			}
		})
	}

	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", fresh)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf(
			"auth/me on the fresh session: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(test.TrimJSON(`{
			"username":"admin",
			"password":"admin-password"
		}`)),
	)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"login with the old password: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	loginCookie(t, api, "admin", "new-admin-password")
}

func TestAPI_AuthMeUpdate__profile(t *testing.T) {
	const (
		password = "profile-password"
		wasName  = "Before"
		wasEmail = "before@example.com"
		nowName  = "After"
		nowEmail = "after@example.com"
	)
	tests := []struct {
		name            string
		body            string
		wantDisplayName string
		wantEmail       string
	}{
		{
			name: "valid/display name only",
			body: test.TrimJSON(`{
				"current_password":"profile-password",
				"display_name":"After"
			}`),
			wantDisplayName: nowName,
			wantEmail:       wasEmail,
		},
		{
			name: "valid/email only",
			body: test.TrimJSON(`{
				"current_password":"profile-password",
				"email":"after@example.com"
			}`),
			wantDisplayName: wasName,
			wantEmail:       nowEmail,
		},
		{
			name: "valid/both at once",
			body: test.TrimJSON(`{
				"current_password":"profile-password",
				"display_name":"After","email":"after@example.com"
			}`),
			wantDisplayName: nowName,
			wantEmail:       nowEmail,
		},
		{
			name: "valid/clearing a field",
			body: test.TrimJSON(`{
				"current_password":"profile-password",
				"display_name":"","email":""
			}`),
			wantDisplayName: "",
			wantEmail:       "",
		},
		{
			name: "valid/no fields, just the password check",
			body: test.TrimJSON(`{
				"current_password":"profile-password"
			}`),
			wantDisplayName: wasName,
			wantEmail:       wasEmail,
		},
	}

	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			username := fmt.Sprintf("profile-user-%d", i)
			user := createAuthUser(t, deps, username, password)
			seedName, seedEmail := wasName, wasEmail
			if _, err := deps.Store.UpdateUser(t.Context(),
				user.ID,
				store.UserPatch{DisplayName: &seedName, Email: &seedEmail},
			); err != nil {
				t.Fatalf("Store.UpdateUser() failed: %v", err)
			}
			cookie := loginCookie(t, api, username, password)

			req := authedRequest(http.MethodPatch, "/api/v1/auth/me", tc.body, cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, http.StatusOK; got != want {
				t.Fatalf(
					"patch auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			var me apitype.AuthMe
			if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
				t.Fatalf("patch auth/me: decode.Unmarshal() failed: %v", err)
			}
			if me.User.DisplayName != tc.wantDisplayName ||
				me.User.Email != tc.wantEmail {
				t.Errorf(
					"patch auth/me: AuthMe.User mismatch\n"+
						"got:  DisplayName=%q, Email=%q\nwant: DisplayName=%q, Email=%q",
					me.User.DisplayName, me.User.Email,
					tc.wantDisplayName, tc.wantEmail,
				)
			}

			// No password changed, so no session to revoke.
			for _, c := range w.Result().Cookies() {
				if c.Name == authCookieName {
					t.Error("patch auth/me: session cookie mismatch\ngot:  present\nwant: absent")
				}
			}

			req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
			w = serveAuth(api, req)
			if got, want := w.Code, http.StatusOK; got != want {
				t.Fatalf(
					"get auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}
			me = apitype.AuthMe{}
			if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
				t.Fatalf("get auth/me: decode.Unmarshal() failed: %v", err)
			}
			if me.User.DisplayName != tc.wantDisplayName ||
				me.User.Email != tc.wantEmail {
				t.Errorf(
					"get auth/me: AuthMe.User mismatch\n"+
						"got:  DisplayName=%q, Email=%q\nwant: DisplayName=%q, Email=%q",
					me.User.DisplayName, me.User.Email,
					tc.wantDisplayName, tc.wantEmail,
				)
			}

			// The original password still logs in: a profile patch leaves it alone.
			loginCookie(t, api, username, password)
		})
	}
}

func TestAPI_AuthMeUpdate__refusals(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	tests := []struct {
		name        string
		body        string
		wantStatus  int
		wantMessage string // Checked when set: the UI maps errors by message.
	}{
		{
			name:       "invalid/malformed body",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid/over-long display name",
			body: test.TrimJSON(`{
				"current_password":"admin-password",
				"display_name":"` + strings.Repeat("x", maxFieldLength+1) + `"
			}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid/new password too short",
			body: test.TrimJSON(`{
				"current_password":"admin-password",
				"new_password":"short"
			}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid/no current password",
			body: test.TrimJSON(`{
				"display_name":"Nice Try"
			}`),
			// 400, not 401: the session is valid, and a 401 would log the caller out of the app.
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid/wrong current password",
			body: test.TrimJSON(`{
				"current_password":"not-my-password",
				"display_name":"Nice Try"
			}`),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "current password is incorrect",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := authedRequest(http.MethodPatch, "/api/v1/auth/me", tc.body, cookie)
			w := serveAuth(api, req)

			if got, want := w.Code, tc.wantStatus; got != want {
				t.Errorf(
					"refusal: status code mismatch\ngot:  %d - %q\nwant: %d",
					got, w.Body.String(), want,
				)
			}

			if tc.wantMessage != "" {
				var body struct {
					Message string `json:"message"`
				}
				if err := decode.Unmarshal("json", w.Body.Bytes(), &body); err != nil {
					t.Fatalf("refusal: decode.Unmarshal() failed: %v", err)
				}
				if body.Message != tc.wantMessage {
					t.Errorf(
						"refusal: message mismatch\ngot:  %q\nwant: %q",
						body.Message, tc.wantMessage,
					)
				}
			}

			req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
			w = serveAuth(api, req)
			var me apitype.AuthMe
			if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
				t.Fatalf("get auth/me: decode.Unmarshal() failed: %v", err)
			}
			if got, want := me.User.DisplayName, "Administrator"; got != want {
				t.Errorf(
					"get auth/me: AuthMe.User.DisplayName mismatch\ngot:  %q\nwant: %q",
					got, want,
				)
			}

			// The original password still logs in: the refused update left it alone.
			loginCookie(t, api, "admin", "admin-password")
		})
	}

	req := httptest.NewRequest(
		http.MethodPatch, "/api/v1/auth/me",
		strings.NewReader(test.TrimJSON(`{
			"current_password": "admin-password",
			"display_name": "Nice Try"
		}`)),
	)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"unauthenticated patch: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_AuthMeUpdate__failures(t *testing.T) {
	t.Run("invalid/no auth context", func(t *testing.T) {
		// Serial: writes the package-level flag vars via testAuthServer's Config.Load.

		w := httptest.NewRecorder()
		api, _, _ := testAuthServer(t, "TestAPI_AuthMeUpdate__failures_noCtx.yml")
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/auth/me", strings.NewReader(`{}`))
		api.httpAuthMeUpdate(w, req)

		if got, want := w.Code, http.StatusUnauthorized; got != want {
			t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
		}
	})

	t.Run("invalid/rate-limited", func(t *testing.T) {
		// Serial: writes the package-level flag vars via testAuthServer's Config.Load.

		api, _, _ := testAuthServer(t, "TestAPI_AuthMeUpdate__failures_limit.yml")
		cookie := loginCookie(t, api, "admin", "admin-password")
		wrong := test.TrimJSON(`{
			"current_password":"not-my-password",
			"display_name":"Nope"
		}`)

		for i := range loginLimitAttempts {
			req := authedRequest(http.MethodPatch, "/api/v1/auth/me", wrong, cookie)
			w := serveAuth(api, req)
			if got, want := w.Code, http.StatusBadRequest; got != want {
				t.Fatalf(
					"attempt %d: status code mismatch\ngot:  %d - %q\nwant: %d",
					i, got, w.Body.String(), want,
				)
			}
		}

		req := authedRequest(http.MethodPatch, "/api/v1/auth/me", wrong, cookie)
		w := serveAuth(api, req)

		if got, want := w.Code, http.StatusTooManyRequests; got != want {
			t.Errorf(
				"past rate-limit: status code mismatch\ngot:  %d - %q\nwant: %d",
				got, w.Body.String(), want,
			)
		}
	})

	t.Run("invalid/credential lookup breaks", func(t *testing.T) {
		// Serial: writes the package-level flag vars via testAuthServer's Config.Load.

		api, deps, dbConn := testAuthServer(t, "TestAPI_AuthMeUpdate__failures_lookup.yml")
		authCtx := adminContext(t, api, deps)
		_ = dbConn.Close()

		req := httptest.NewRequest(
			http.MethodPatch, "/api/v1/auth/me",
			strings.NewReader(`{"current_password":"admin-password"}`),
		)
		w := httptest.NewRecorder()
		api.httpAuthMeUpdate(w, withAuthCtx(req, authCtx))

		// 500, not 400: the store failed, not the password.
		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
		}
	})

	t.Run("invalid/hashing breaks", func(t *testing.T) {
		// Serial: swaps hashPassword.

		api, _, _ := testAuthServer(t, "TestAPI_AuthMeUpdate__failures_hash.yml")
		cookie := loginCookie(t, api, "admin", "admin-password")
		hashPasswordHad := hashPassword
		t.Cleanup(func() { hashPassword = hashPasswordHad })
		hashPassword = func(_ string) (string, error) {
			return "", errors.New("hash broke")
		}

		req := authedRequest(
			http.MethodPatch, "/api/v1/auth/me",
			test.TrimJSON(`{
				"current_password":"admin-password",
				"new_password":"new-admin-password"
			}`),
			cookie,
		)
		w := serveAuth(api, req)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
		}
	})

	t.Run("invalid/the account vanished from the DB", func(t *testing.T) {
		// Serial: writes the package-level flag vars via testAuthServer's Config.Load.

		api, deps, _ := testAuthServer(t, "TestAPI_AuthMeUpdate__failures_gone.yml")
		authCtx := *adminContext(t, api, deps)
		authCtx.User.ID = "no-such-id"

		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPatch, "/api/v1/auth/me",
			strings.NewReader(test.TrimJSON(`{
				"current_password":"admin-password",
				"display_name":"Ghost"
			}`)),
		)
		api.httpAuthMeUpdate(w, withAuthCtx(req, &authCtx))

		if got, want := w.Code, http.StatusNotFound; got != want {
			t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
		}
	})

	t.Run("invalid/sessions break on a password change", func(t *testing.T) {
		api, deps, _ := testAuthServer(t, "TestAPI_AuthMeUpdate__failures_sessions.yml")
		authCtx := adminContext(t, api, deps)

		// A password change revokes and re-mints sessions.
		deps.Sessions = session.New(
			failingSessionStore{},
			session.Config{Lifetime: time.Hour, IdleTimeout: time.Hour},
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPatch, "/api/v1/auth/me",
			strings.NewReader(test.TrimJSON(`{
				"current_password":"admin-password",
				"new_password":"new-admin-password"
			}`)),
		)
		api.httpAuthMeUpdate(w, withAuthCtx(req, authCtx))

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
		}
	})

	t.Run("invalid/grants unreadable after the update", func(t *testing.T) {
		api, deps, dbConn := testAuthServer(t, "TestAPI_AuthMeUpdate__failures_grants.yml")
		authCtx := adminContext(t, api, deps)
		// The update itself succeeds; only re-reading the account fails.
		test.DropTable(t, dbConn, "permissions")

		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPatch, "/api/v1/auth/me",
			strings.NewReader(test.TrimJSON(`{
				"current_password":"admin-password",
				"display_name":"Renamed"
			}`)),
		)
		api.httpAuthMeUpdate(w, withAuthCtx(req, authCtx))

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
		}
	})
}

func TestAPI_AuthMeUpdate__bearerRefused(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)
	plaintext, _ := createAPIToken(t, deps, authCtx.User.ID, "ci", nil)

	req := authedRequest(http.MethodPatch, "/api/v1/auth/me", test.TrimJSON(`{
		"current_password":"admin-password",
		"new_password":"new-admin-password"
	}`), nil)
	req.Header.Set("Authorization", "Bearer "+plaintext)
	w := serveAuth(api, req)

	// The route takes only a session.
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, w.Body.String(), want)
	}

	// The original password still logs in: the refused change was not applied.
	loginCookie(t, api, "admin", "admin-password")
}

func TestAPI_Auth__sessionExpired(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	deps.Sessions = session.New(
		deps.Store,
		session.Config{
			Lifetime:    -time.Hour,
			IdleTimeout: time.Hour,
		},
	)
	cookie := loginCookie(t, api, "admin", "admin-password")

	req := authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
	w := serveAuth(api, req)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	req = authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie)
	w = serveAuth(api, req)
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Errorf(
			"logout: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
	assertSessionCookieCleared(t, w)
}

func TestAPI_Auth__directHandlerBranches(t *testing.T) {
	file := t.Name() + ".yml"
	api, deps, _ := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	api.httpAuthMe(w, req)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"auth/me: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}

	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: "some-token"})
	w = httptest.NewRecorder()
	api.httpAuthLogout(w, withAuthCtx(req, authCtx))

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"logout: status code mismatch\ngot:  %d - %q\nwant: %d",
			got, w.Body.String(), want,
		)
	}
}

func TestAPI_Auth__sessionCookie(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	prefixHad := api.RoutePrefix
	api.RoutePrefix = ""

	cookie := api.sessionCookie(req, "token", 60)
	api.RoutePrefix = prefixHad

	if got, want := cookie.Path, "/"; got != want {
		t.Errorf("no TLS or prefix: Cookie.Path mismatch\ngot:  %q\nwant: %q", got, want)
	}
	if cookie.Secure {
		t.Error("no TLS or prefix: Cookie.Secure mismatch\ngot:  true\nwant: false")
	}

	api.Config.Settings.Web.CertFile = "cert.pem"

	cookie = api.sessionCookie(req, "token", 60)
	api.Config.Settings.Web.CertFile = ""

	if !cookie.Secure {
		t.Error("TLS: Cookie.Secure mismatch\ngot:  false\nwant: true")
	}

	trustedHad := api.trustedProxies
	api.trustedProxies = nil
	req.Header.Set("X-Forwarded-Proto", "https")
	cookie = api.sessionCookie(req, "token", 60)
	api.trustedProxies = trustedHad

	// X-Forwarded-Proto is honoured without trust.
	if !cookie.Secure {
		t.Error("untrusted proxy HTTPS: Cookie.Secure mismatch\ngot:  false\nwant: true")
	}
}

func TestAPI_Auth__sessionCookie__secureCookieOverride(t *testing.T) {
	file := t.Name() + ".yml"
	api, _, _ := testAuthServer(t, file)
	secureOn, secureOff := true, false

	tests := []struct {
		name                 string
		secureCookieOverride *bool
		forwardedProto       string
		wantSecure           bool
	}{
		{
			name:                 "unset/plain HTTP is Secure=false",
			secureCookieOverride: nil,
			wantSecure:           false,
		},
		{
			name:                 "unset/proxied https is Secure=true",
			secureCookieOverride: nil,
			forwardedProto:       "https",
			wantSecure:           true,
		},
		{
			name:                 "forced on/plain http is Secure=true",
			secureCookieOverride: &secureOn,
			forwardedProto:       "",
			wantSecure:           true,
		},
		{
			name:                 "forced on/proxied https is Secure=true",
			secureCookieOverride: &secureOn,
			forwardedProto:       "https",
			wantSecure:           true,
		},
		{
			name:                 "forced off/proxied https is Secure=false",
			secureCookieOverride: &secureOff,
			forwardedProto:       "https",
			wantSecure:           false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares api.Config.Settings.Auth.Session.

			api.Config.Settings.Auth.Session.SecureCookie = tc.secureCookieOverride
			t.Cleanup(func() { api.Config.Settings.Auth.Session.SecureCookie = nil })

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
			if tc.forwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", tc.forwardedProto)
			}

			cookie := api.sessionCookie(req, "token", 60)

			if got, want := cookie.Secure, tc.wantSecure; got != want {
				t.Errorf("Cookie.Secure mismatch\ngot:  %t\nwant: %t", got, want)
			}
		})
	}
}
