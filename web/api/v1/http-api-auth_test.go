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
	"github.com/release-argus/Argus/config/decode"
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

func TestLoginLimiter_check_and_recordFailure(t *testing.T) {
	// GIVEN: a loginLimiter.
	limiter := newLoginLimiter()
	base := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	now := base
	timeNowHad := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = timeNowHad })

	prefix := fmt.Sprintf("%s\nloginLimiter check/recordFailure", packageName)

	// WHEN: check() is called repeatedly WITHOUT recording
	// THEN: it never trips (read-only; only failures count).
	for range loginLimitAttempts * 2 {
		if !limiter.check("ip") {
			t.Fatalf("%s\ncheck() must not consume the budget", prefix)
		}
	}

	// WHEN: the allowed number of failures are recorded for an IP.
	for range loginLimitAttempts {
		if !limiter.check("ip") {
			t.Fatalf("%s\nfailures within the limit should be allowed", prefix)
		}
		limiter.recordFailure("ip")
	}
	// THEN: the next check is denied.
	if limiter.check("ip") {
		t.Errorf("%s\ncheck past the failure limit should be denied", prefix)
	}
	// AND: another IP still has its own budget.
	if !limiter.check("other-ip") {
		t.Errorf("%s\na fresh IP should have its own budget", prefix)
	}

	// AND: the failure budget resets once the window expires.
	now = base.Add(loginLimitWindow + time.Second)
	if !limiter.check("ip") {
		t.Errorf("%s\nchecks should resume after the window expires", prefix)
	}
}

func TestLoginLimiter__lockIP(t *testing.T) {
	// GIVEN: a loginLimiter and many concurrent failed attempts from one IP.
	limiter := newLoginLimiter()
	base := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	timeNowHad := timeNow
	timeNow = func() time.Time { return base }
	t.Cleanup(func() { timeNow = timeNowHad })

	prefix := fmt.Sprintf("%s\nloginLimiter lockIP", packageName)

	// WHEN: each attempt serialises its check/recordFailure per IP.
	const attempts = 50
	var passed atomic.Int64
	var wg sync.WaitGroup
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer limiter.lockIP("ip")()
			if limiter.check("ip") {
				passed.Add(1)
				limiter.recordFailure("ip")
			}
		}()
	}
	wg.Wait()

	// THEN: exactly the allowed number pass the check (no burst bypass).
	if got := passed.Load(); got != loginLimitAttempts {
		t.Errorf("%s\nexpected exactly %d attempts to pass under contention, got %d",
			prefix,
			loginLimitAttempts,
			got)
	}

	// AND: the window's per-IP lock is fully released once no attempts remain.
	limiter.mu.Lock()
	window := limiter.windows["ip"]
	limiter.mu.Unlock()
	if window == nil {
		t.Fatalf("%s\nwindow holding the failed-attempt count should persist", prefix)
	}
	limiter.mu.Lock()
	refs := window.refs
	limiter.mu.Unlock()
	if refs != 0 {
		t.Errorf("%s\nper-IP lock should be released (refs 0), got %d",
			prefix,
			refs)
	}
}

func TestAPI_AuthLogin(t *testing.T) {
	// GIVEN: an auth-enabled API with an admin and a disabled user.
	file := "TestAPI_AuthLogin.yml"
	api, deps, _ := testAuthServer(t, file)
	disabledUser := createAuthUser(t, deps, "sleeper", "sleeper-password", store.GroupViewer)
	if _, err := deps.Store.UpdateUser(t.Context(), disabledUser.ID,
		store.UserPatch{Enabled: new(false)}); err != nil {
		t.Fatalf(
			"%s\ndisable user: %v",
			packageName, err,
		)
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
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed body",
			body:       `{"username":`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prefix := fmt.Sprintf("%s\nPOST /auth/login", packageName)

			// WHEN: login is attempted.
			w := serveAuth(api,
				httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tc.body)))

			// THEN: the status matches expectations.
			if w.Code != tc.wantStatus {
				t.Fatalf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, w.Code, w.Body.String(), tc.wantStatus,
				)
			}

			// AND: a session cookie is set only on success.
			gotCookie := false
			for _, cookie := range w.Result().Cookies() {
				if cookie.Name == authCookieName && cookie.Value != "" {
					gotCookie = true
					// AND: the cookie is locked down.
					if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
						t.Errorf(
							"%s\ncookie should be HttpOnly + SameSite=Strict\ngot: %+v",
							prefix, cookie,
						)
					}
				}
			}
			if gotCookie != tc.wantCookie {
				t.Errorf(
					"%s\ncookie mismatch\ngot:  %t\nwant: %t",
					prefix, gotCookie, tc.wantCookie,
				)
			}

			// AND: successful responses carry the user and their permissions.
			if tc.wantStatus == http.StatusOK {
				var response apitype.AuthMe
				if err := decode.Unmarshal("json", w.Body.Bytes(), &response); err != nil {
					t.Fatalf(
						"%s\nparse response: %v",
						prefix, err,
					)
				}
				if response.User.Username != "admin" || len(response.Permissions) == 0 {
					t.Errorf(
						"%s\nresponse mismatch\ngot:  user=%q, %d permissions",
						prefix, response.User.Username, len(response.Permissions),
					)
				}
			}
		})
	}
}

func TestAPI_AuthLogin__rateLimit(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_AuthLogin__rateLimit.yml"
	api, _, _ := testAuthServer(t, file)

	prefix := fmt.Sprintf("%s\nPOST /auth/login rate limit", packageName)

	// WHEN: more login attempts arrive than the window allows.
	var lastCode int
	for range loginLimitAttempts + 1 {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
			strings.NewReader(`{"username":"admin","password":"wrong"}`))
		lastCode = serveAuth(api, req).Code
	}

	// THEN: the attempt past the limit is rejected with 429.
	if lastCode != http.StatusTooManyRequests {
		t.Errorf(
			"%s\nstatus mismatch\ngot:  %d\nwant: %d",
			prefix, lastCode, http.StatusTooManyRequests,
		)
	}
}

func TestAPI_AuthLogin__sessionCap(t *testing.T) {
	// GIVEN: an auth-enabled API whose admin is at the session cap,
	// with a WebSocket client on their oldest session.
	file := "TestAPI_AuthLogin__sessionCap.yml"
	api, deps, _ := testAuthServer(t, file)
	oldestCookie := loginCookie(t, api, "admin", "admin-password")
	for range session.MaxSessionsPerUser - 1 {
		_ = loginCookie(t, api, "admin", "admin-password")
	}
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	oldestClient := connect(adminID, session.HashToken(oldestCookie.Value))

	prefix := fmt.Sprintf("%s\nhttpAuthLogin() at the session cap", packageName)

	// WHEN: the user logs in once more.
	newestCookie := loginCookie(t, api, "admin", "admin-password")

	// THEN: the oldest session's WebSocket client is kicked with its eviction.
	if api.hub.hasClient(oldestClient) {
		t.Errorf(
			"%s\nthe evicted session's client should be kicked",
			prefix,
		)
	}

	// AND: the evicted session no longer authenticates; the newest does.
	if w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", oldestCookie),
	); w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nevicted session\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}
	if w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", newestCookie),
	); w.Code != http.StatusOK {
		t.Errorf(
			"%s\nnewest session\ngot:  %d\nwant: 200",
			prefix, w.Code,
		)
	}
}

func TestAPI_Auth__noProviders(t *testing.T) {
	// GIVEN: an API whose provider registry is empty.
	file := "TestAPI_Auth__noProviders.yml"
	api, deps, _ := testAuthServer(t, file)
	deps.Providers = provider.NewRegistry()

	prefix := fmt.Sprintf("%s\nempty provider registry", packageName)

	// WHEN: a login arrives.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`))
	w := serveAuth(api, req)

	// THEN: it is rejected as 401.
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"%s\nlogin\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestAPI_Auth__loginEdgeCases(t *testing.T) {
	// GIVEN: an auth-enabled API.
	file := "TestAPI_Auth__loginEdgeCases.yml"
	api, deps, dbConn := testAuthServer(t, file)

	prefix := fmt.Sprintf("%s\nlogin edge cases", packageName)

	// WHEN: the login body exceeds the size cap.
	oversize := `{"username":"` + strings.Repeat("x", maxAuthBodySize) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(oversize))
	// THEN: 400 Bad Request.
	if w := serveAuth(api, req); w.Code != http.StatusBadRequest {
		t.Errorf(
			"%s\noversize body\ngot:  %d\nwant: 400",
			prefix, w.Code,
		)
	}

	// WHEN: the provider authenticates a subject the store cannot resolve.
	deps.Providers = provider.NewRegistry()
	deps.Providers.Register(ghostProvider{})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"ghost","password":"x"}`))
	// THEN: the login reads as invalid credentials, not a server error.
	if w := serveAuth(api, req); w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nghost subject\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}

	// WHEN: grants become unreadable after the provider verified.
	if _, err := dbConn.Exec(`DROP TABLE permissions;`); err != nil {
		t.Fatalf(
			"%s\nsetup drop failed: %v",
			packageName, err,
		)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"ghost","password":"x"}`))
	deps.Providers = provider.NewRegistry()
	deps.Providers.Register(realSubjectProvider{store: deps.Store})
	// THEN: the infrastructure failure reads as 500.
	if w := serveAuth(api, req); w.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\ngrant failure during login\ngot:  %d\nwant: 500",
			prefix, w.Code,
		)
	}
}

func TestAPI_Auth__sessionManagerFailures(t *testing.T) {
	// GIVEN: an API whose session manager persists to a broken store.
	file := "TestAPI_Auth__sessionManagerFailures.yml"
	api, deps, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	// Swap in a manager whose store always fails (and has an empty cache).
	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)

	prefix := fmt.Sprintf("%s\nsession manager failures", packageName)

	// WHEN: a cookie is validated against the broken manager.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/flags", "", cookie))
	// THEN: the infrastructure failure reads as 500.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"%s\nvalidate failure\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// WHEN: a login tries to mint a session via the broken manager.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"username":"admin","password":"admin-password"}`))
	w = serveAuth(api, req)
	// THEN: session minting failure reads as 500.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"%s\nmint failure\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// WHEN: logout hits the broken manager.
	w = serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie))
	// THEN: the revoke failure reads as 500.
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf(
			"%s\nrevoke failure\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: a cookieless logout still succeeds - there is no session to revoke.
	w = serveAuth(api,
		httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Errorf(
			"%s\ncookieless logout\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestAPI_AuthSetup(t *testing.T) {
	// GIVEN: an auth-enabled API with no users yet (first-run setup pending).
	file := "TestAPI_AuthSetup.yml"
	api, deps, _ := testAuthServerPendingSetup(t, file)

	prefix := fmt.Sprintf("%s\nfirst-run setup", packageName)

	// WHEN: the setup state is fetched (unauthenticated).
	w := serveAuth(api,
		httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup", nil))

	// THEN: setup is reported as pending.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\nstate status mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
	var state apitype.SetupState
	if err := decode.Unmarshal("json", w.Body.Bytes(), &state); err != nil {
		t.Fatalf(
			"%s\nparse state: %v",
			prefix, err,
		)
	}
	if !state.SetupRequired {
		t.Errorf("%s\nsetup_required should be true before setup", prefix)
	}

	// AND: other endpoints still require authentication.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	if w := serveAuth(api, req); w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nusers without a session\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}

	// AND: missing fields are rejected with 400.
	for _, body := range []string{
		`{}`,
		`{"username":"root"}`,
		`{"password":"setup-password"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup", strings.NewReader(body))
		if w := serveAuth(api, req); w.Code != http.StatusBadRequest {
			t.Errorf(
				"%s\nsetup with %s\ngot:  %d\nwant: 400",
				prefix, body, w.Code,
			)
		}
	}

	// WHEN: the first administrator is created.
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup",
		strings.NewReader(`{"username":"root","display_name":"Rooty","password":"setup-password"}`))
	w = serveAuth(api, req)

	// THEN: the account is created, in the admin group, and logged in.
	if got, want := w.Code, http.StatusCreated; got != want {
		t.Fatalf(
			"%s\nsetup status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var me apitype.AuthMe
	if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
		t.Fatalf(
			"%s\nparse setup response: %v",
			prefix, err,
		)
	}
	if me.User.Username != "root" ||
		me.User.DisplayName != "Rooty" ||
		len(me.User.Groups) != 1 ||
		me.User.Groups[0] != store.GroupAdmin {
		t.Errorf(
			"%s\nsetup user mismatch\ngot: %+v",
			prefix, me.User,
		)
	}
	if len(me.Permissions) == 0 {
		t.Errorf("%s\nadmin permissions should not be empty", prefix)
	}
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == authCookieName {
			cookie = c
		}
	}
	if cookie == nil || cookie.Value == "" {
		t.Fatalf("%s\nsetup should set the session cookie", prefix)
	}

	// AND: the minted session authenticates /auth/me.
	req = authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie)
	if w := serveAuth(api, req); w.Code != http.StatusOK {
		t.Errorf(
			"%s\nme with setup session\ngot:  %d\nwant: 200",
			prefix, w.Code,
		)
	}

	// AND: setup now reports complete.
	w = serveAuth(api,
		httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup", nil))
	if err := decode.Unmarshal("json", w.Body.Bytes(), &state); err != nil {
		t.Fatalf(
			"%s\nparse state: %v",
			prefix, err,
		)
	}
	if state.SetupRequired {
		t.Errorf("%s\nsetup_required should be false after setup", prefix)
	}
	// AND: repeats are rejected with 409.
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup",
		strings.NewReader(`{"username":"root2","password":"another-password"}`))
	if w := serveAuth(api, req); w.Code != http.StatusConflict {
		t.Errorf(
			"%s\nrepeat setup\ngot:  %d\nwant: 409",
			prefix, w.Code,
		)
	}

	// AND: exactly one user exists.
	if count, err := deps.Store.CountUsers(t.Context()); err != nil || count != 1 {
		t.Errorf(
			"%s\nuser count mismatch\ngot:  %d, err=%v\nwant: 1",
			prefix, count, err,
		)
	}
}

func TestAPI_AuthSetup__concurrent(t *testing.T) {
	// GIVEN: first-run setup pending.
	file := "TestAPI_AuthSetup__concurrent.yml"
	api, _, _ := testAuthServerPendingSetup(t, file)

	prefix := fmt.Sprintf("%s\nconcurrent /auth/setup", packageName)

	// WHEN: two setup requests race.
	var wg sync.WaitGroup
	start := make(chan struct{})
	codes := make([]int, 2)
	for i := range codes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := fmt.Sprintf(`{"username":"admin-%d","password":"str0ng-pass"}`, i)
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup",
				strings.NewReader(body))
			codes[i] = serveAuth(api, req).Code
		}()
	}
	close(start)
	wg.Wait()

	// THEN: exactly one succeeds and the other conflicts.
	slices.Sort(codes)
	if codes[0] != http.StatusCreated || codes[1] != http.StatusConflict {
		t.Errorf(
			"%s\nstatus mismatch\ngot:  %v\nwant: [201 409]",
			prefix, codes,
		)
	}
}

func TestAPI_AuthSetup__errors(t *testing.T) {
	// GIVEN: setup requests that fail.
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
		},
		{
			name:       "state read failure",
			method:     http.MethodGet,
			closeDB:    true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "create failure",
			method:     http.MethodPost,
			body:       `{"username":"root","password":"setup-password"}`,
			closeDB:    true,
			wantStatus: http.StatusInternalServerError,
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
			file := fmt.Sprintf("TestAPI_AuthSetup__errors_%s.yml",
				strings.ReplaceAll(tc.name, " ", "_"))
			api, _, dbConn := testAuthServerPendingSetup(t, file)
			if tc.closeDB {
				_ = dbConn.Close()
			}
			if tc.dropTable != "" {
				if _, err := dbConn.Exec("DROP TABLE " + tc.dropTable + ";"); err != nil {
					t.Fatalf(
						"%s\nsetup drop failed: %v",
						packageName, err,
					)
				}
			}

			prefix := fmt.Sprintf("%s\nsetup", packageName)

			// WHEN: the request is made.
			w := serveAuth(api,
				httptest.NewRequest(tc.method, "/api/v1/auth/setup",
					strings.NewReader(tc.body)))

			// THEN: the status matches expectations.
			if w.Code != tc.wantStatus {
				t.Errorf(
					"%s\nstatus mismatch\ngot:  %d - %s\nwant: %d",
					prefix, w.Code, w.Body.String(), tc.wantStatus,
				)
			}

			// AND: the failure surfaces the expected message.
			if !util.RegexCheck(tc.bodyRegex, w.Body.String()) {
				t.Errorf(
					"%s\nbody mismatch\ngot:  %q\nwant: %q",
					prefix, w.Body.String(), tc.bodyRegex,
				)
			}
		})
	}
}

func TestAPI__AuthMe_and_Logout(t *testing.T) {
	// GIVEN: a logged-in admin.
	file := "TestAPI_AuthMe_And_Logout.yml"
	api, _, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")

	prefix := fmt.Sprintf("%s\n/auth/me + /auth/logout", packageName)

	// WHEN: /auth/me is requested with the session cookie.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie))

	// THEN: the user and permissions are returned.
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"%s\n/auth/me status mismatch\ngot:  %d - %s\nwant: %d",
			prefix, got, w.Body.String(), want,
		)
	}
	var me apitype.AuthMe
	if err := decode.Unmarshal("json", w.Body.Bytes(), &me); err != nil {
		t.Fatalf(
			"%s\nparse /auth/me: %v",
			prefix, err,
		)
	}
	if me.User.Username != "admin" || len(me.Permissions) == 0 {
		t.Errorf(
			"%s\n/auth/me mismatch\ngot:  user=%q, %d permissions",
			prefix, me.User.Username, len(me.Permissions),
		)
	}

	// AND: /auth/me without credentials is 401.
	w = serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\nunauthenticated /auth/me\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}

	// WHEN: the session is logged out.
	w = serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie))

	// THEN: the logout succeeds and clears the cookie.
	if got, want := w.Code, http.StatusNoContent; got != want {
		t.Fatalf(
			"%s\nlogout status mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == authCookieName && c.MaxAge >= 0 {
			t.Errorf(
				"%s\nlogout should clear the cookie\ngot: %+v",
				prefix, c,
			)
		}
	}

	// AND: the old session no longer authenticates.
	w = serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie))
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"%s\nrevoked session status mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestAPI_AuthLogout__kicks(t *testing.T) {
	// GIVEN: an auth-enabled API with a WebSocket client on each of a user's
	// two sessions.
	file := "TestAPI_AuthLogout__kicks.yml"
	api, deps, _ := testAuthServer(t, file)
	cookie := loginCookie(t, api, "admin", "admin-password")
	otherCookie := loginCookie(t, api, "admin", "admin-password")
	adminID := adminContext(t, api, deps).User.ID
	connect := wireHub(t, api)
	logoutClient := connect(adminID, session.HashToken(cookie.Value))
	otherClient := connect(adminID, session.HashToken(otherCookie.Value))

	prefix := fmt.Sprintf("%s\nhttpAuthLogout() kicks", packageName)

	// WHEN: the first session logs out.
	w := serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie))
	if w.Code != http.StatusNoContent {
		t.Fatalf(
			"%s\nlogout\ngot:  %d - %s\nwant: 204",
			prefix, w.Code, w.Body.String(),
		)
	}

	// THEN: only that session's WebSocket client is kicked - the same user's
	// other session stays connected.
	if api.hub.hasClient(logoutClient) || !api.hub.hasClient(otherClient) {
		t.Errorf(
			"%s\nlogout should kick exactly the revoked session's clients",
			prefix,
		)
	}
}

func TestAPI_Auth__sessionExpired(t *testing.T) {
	// GIVEN: a server whose sessions are already past their lifetime.
	file := "TestAPI_Auth__sessionExpired.yml"
	api, deps, _ := testAuthServer(t, file)
	deps.Sessions = session.New(
		deps.Store,
		session.Config{
			Lifetime:    -time.Hour,
			IdleTimeout: time.Hour,
		},
	)
	cookie := loginCookie(t, api, "admin", "admin-password")

	prefix := fmt.Sprintf("%s\nexpired session", packageName)

	// WHEN: the expired session cookie is presented.
	w := serveAuth(api,
		authedRequest(http.MethodGet, "/api/v1/auth/me", "", cookie))

	// THEN: the request is 401.
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf(
			"%s\nstatus mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: logout with the expired cookie still succeeds and clears it.
	w = serveAuth(api,
		authedRequest(http.MethodPost, "/api/v1/auth/logout", "", cookie))
	if w.Code != http.StatusNoContent {
		t.Errorf(
			"%s\nlogout with an expired session\ngot:  %d\nwant: 204",
			prefix, w.Code,
		)
	}
}

func TestAPI_Auth__directHandlerBranches(t *testing.T) {
	// GIVEN: handlers invoked directly with crafted contexts.
	file := "TestAPI_Auth__directHandlerBranches.yml"
	api, deps, _ := testAuthServer(t, file)
	authCtx := adminContext(t, api, deps)

	prefix := fmt.Sprintf("%s\ndirect handler branches", packageName)

	// WHEN: /auth/me runs without an auth context.
	w := httptest.NewRecorder()
	api.httpAuthMe(w, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))
	// THEN: 401.
	if w.Code != http.StatusUnauthorized {
		t.Errorf(
			"%s\n/auth/me without context\ngot:  %d\nwant: 401",
			prefix, w.Code,
		)
	}

	// WHEN: logout's revoke fails (broken session manager, cookie present).
	deps.Sessions = session.New(
		failingSessionStore{},
		session.Config{
			Lifetime: time.Hour, IdleTimeout: time.Hour,
		},
	)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: "some-token"})
	w = httptest.NewRecorder()
	api.httpAuthLogout(w, withAuthCtx(req, authCtx))
	// THEN: 500.
	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"%s\nlogout revoke failure\ngot:  %d\nwant: 500",
			prefix, w.Code,
		)
	}
}

func TestAPI_Auth__sessionCookie(t *testing.T) {
	// GIVEN: APIs with differing prefix/TLS settings.
	file := "TestAPI_Auth__sessionCookie.yml"
	api, _, _ := testAuthServer(t, file)

	prefix := fmt.Sprintf("%s\nsessionCookie", packageName)

	// WHEN: the route prefix is empty.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	prefixHad := api.RoutePrefix
	api.RoutePrefix = ""
	cookie := api.sessionCookie(req, "token", 60)
	api.RoutePrefix = prefixHad
	// THEN: the path falls back to "/".
	if cookie.Path != "/" {
		t.Errorf(
			"%s\npath fallback\ngot:  %q\nwant: \"/\"",
			prefix, cookie.Path,
		)
	}
	if cookie.Secure {
		t.Errorf("%s\nSecure should be off without TLS", prefix)
	}

	// WHEN: TLS is configured.
	api.Config.Settings.Web.CertFile = "cert.pem"
	cookie = api.sessionCookie(req, "token", 60)
	api.Config.Settings.Web.CertFile = ""
	// THEN: the cookie is Secure.
	if !cookie.Secure {
		t.Errorf("%s\nSecure should be on with TLS", prefix)
	}

	// WHEN: a proxy reports the client connection as HTTPS.
	req.Header.Set("X-Forwarded-Proto", "https")
	cookie = api.sessionCookie(req, "token", 60)
	// THEN: the cookie is Secure.
	if !cookie.Secure {
		t.Errorf("%s\nSecure should be on when X-Forwarded-Proto is https", prefix)
	}

	// WHEN: the header arrives without any configured trusted proxies.
	trustedHad := api.trustedProxies
	api.trustedProxies = nil
	cookie = api.sessionCookie(req, "token", 60)
	api.trustedProxies = trustedHad
	// THEN: Secure is still set - trusting X-Forwarded-Proto for the Secure flag
	// is fail-safe (it can only withhold the cookie over HTTP, never leak it);
	// trusted-proxy gating is reserved for client-IP resolution.
	if !cookie.Secure {
		t.Errorf("%s\nSecure should track X-Forwarded-Proto regardless of proxy trust (fail-safe)", prefix)
	}
}
