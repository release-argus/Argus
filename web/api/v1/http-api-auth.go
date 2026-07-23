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

package v1

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/provider/local"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// maxAuthBodySize bounds auth request bodies.
const maxAuthBodySize = 64 * 1024

// Login rate limiting: fixed windows keyed purely on client IP.
const (
	// loginLimitWindow is the duration over which failed login attempts are counted.
	loginLimitWindow = 5 * time.Minute
	// loginLimitAttempts is the maximum number of failed login attempts allowed within [loginLimitWindow].
	loginLimitAttempts = 5
)

// loginLimiter is an in-memory fixed-window rate limiter for failed
// login attempts.
type loginLimiter struct {
	mu      sync.Mutex
	windows map[string]*loginWindow
}

// loginWindow counts failed attempts within a [loginLimitWindow].
type loginWindow struct {
	start    time.Time
	attempts int
}

// newLoginLimiter creates an empty [loginLimiter].
func newLoginLimiter() *loginLimiter {
	return &loginLimiter{windows: make(map[string]*loginWindow)}
}

// sweepExpired drops windows past the [loginLimitWindow] window so the map cannot grow
// unbounded. The caller must hold l.mu.
func (l *loginLimiter) sweepExpired(now time.Time) {
	for k, window := range l.windows {
		if now.Sub(window.start) >= loginLimitWindow {
			delete(l.windows, k)
		}
	}
}

// check reports whether ip is within the [loginLimitAttempts] limit, without recording an attempt
// (and prunes expired windows).
func (l *loginLimiter) check(ip string) bool {
	now := timeNow()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweepExpired(now)
	window := l.windows[ip]
	return window == nil || window.attempts < loginLimitAttempts
}

// recordFailure records one failed attempt for ip.
func (l *loginLimiter) recordFailure(ip string) {
	now := timeNow()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweepExpired(now)
	if window := l.windows[ip]; window != nil {
		window.attempts++
		return
	}
	l.windows[ip] = &loginWindow{start: now, attempts: 1}
}

// httpAuthLogin handles POST /api/v1/auth/login: verifying credentials via
// the local provider, minting a session, and setting the session cookie.
//
// Response:
//
//	200 OK: with the user and their permission grants (session cookie set).
//	400 Bad Request: on a malformed or oversized body.
//	401 Unauthorized: uniformly for any credential failure.
//	429 Too Many Requests: when the client IP is rate-limited.
//	500 Internal Server Error: on an unexpected internal failure.
func (api *API) httpAuthLogin(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAuthLogin", Secondary: getIP(r)}

	// Decode the credentials.
	var request apitype.LoginRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}

	// Authenticate via the local provider (failed attempts are rate limited).
	identity, err := api.verifyLocalCredentials(r.Context(), request.Username, request.Password, getIP(r))
	if err != nil {
		switch {
		case errors.Is(err, errTooManyAttempts):
			failRequest(&w, errTooManyAttempts, http.StatusTooManyRequests)
		case errors.Is(err, auth.ErrInvalidCredentials):
			failRequest(&w, auth.ErrInvalidCredentials, http.StatusUnauthorized)
		default:
			logx.Error(err, logFrom, true)
			failRequest(&w, errors.New("authentication failed"), http.StatusInternalServerError)
		}
		return
	}

	// Resolve the Identity to its user and permissions.
	authCtx, err := api.contextForUser(r.Context(), identity.Subject, identity.Provider)
	if err != nil {
		if errors.Is(err, errUnauthorised) {
			failRequest(&w, auth.ErrInvalidCredentials, http.StatusUnauthorized)
			return
		}
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("authentication failed"), http.StatusInternalServerError)
		return
	}

	// Mint the session and return the user.
	if !api.startSession(w, r, authCtx, logFrom) {
		return
	}
	api.writeAuthMe(w, http.StatusOK, authCtx, logFrom)
}

// httpAuthSetupState handles GET /api/v1/auth/setup: first-run setup state
// (creating the first administrator), reporting whether or not setup is pending.
//
// Response:
//
//	200 OK: JSON reporting whether first-run setup is required.
//	500 Internal Server Error: on a store failure.
func (api *API) httpAuthSetupState(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAuthSetupState", Secondary: getIP(r)}

	count, err := api.auth.Store.CountUsers(r.Context())
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("failed to read setup state"), http.StatusInternalServerError)
		return
	}

	api.writeJSON(w, apitype.SetupState{SetupRequired: count == 0}, logFrom)
}

// httpAuthSetup handles POST /api/v1/auth/setup: creates the first
// administrator account (only while no users exist) and logs it in.
//
// Response:
//
//	201 Created: with the user and their grants (session cookie set).
//	400 Bad Request: on a validation fail.
//	409 Conflict: once setup has completed.
//	500 Internal Server Error: on an unexpected internal failure.
func (api *API) httpAuthSetup(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAuthSetup", Secondary: getIP(r)}

	// Decode the account details.
	var request apitype.SetupRequest
	if !api.decodeAuthBody(w, r, &request) {
		return
	}
	if request.Username == "" {
		failRequest(&w, errors.New("username is required"), http.StatusBadRequest)
		return
	}
	if err := validatePassword(request.Password); err != nil {
		failRequest(&w, err, http.StatusBadRequest)
		return
	}

	// Create the administrator.
	user, err := api.auth.Store.CreateFirstAdmin(
		r.Context(),
		request.Username,
		request.DisplayName,
		request.Password,
	)
	if err != nil {
		if errors.Is(err, store.ErrSetupComplete) {
			failRequest(&w, store.ErrSetupComplete, http.StatusConflict)
			return
		}
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("setup failed"), http.StatusInternalServerError)
		return
	}

	// Log the new administrator straight in.
	authCtx, err := api.contextForUser(r.Context(), user.ID, local.Name)
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("setup succeeded but login failed"), http.StatusInternalServerError)
		return
	}
	if !api.startSession(w, r, authCtx, logFrom) {
		return
	}

	api.writeAuthMe(w, http.StatusCreated, authCtx, logFrom)
}

// httpAuthLogout handles POST /api/v1/auth/logout: revokes the session and
// clears the cookie.
//
// Response:
//
//	204 No Content: on success (including when no session cookie is present).
//	500 Internal Server Error: on a session-revocation failure.
func (api *API) httpAuthLogout(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAuthLogout", Secondary: getIP(r)}

	if cookie, err := r.Cookie(authCookieName); err == nil {
		if err := api.auth.Sessions.Revoke(r.Context(), cookie.Value); err != nil {
			logx.Error(err, logFrom, true)
			failRequest(&w, errors.New("logout failed"), http.StatusInternalServerError)
			return
		}
		// Kick any WebSocket clients still connected under this session.
		api.kickSessionWebSocketClients(session.HashToken(cookie.Value))
	}

	http.SetCookie(w, api.sessionCookie(r, "", -1))
	w.WriteHeader(http.StatusNoContent)
}

// httpAuthMe handles GET /api/v1/auth/me: the authenticated user and their
// effective permission grants.
//
// Response:
//
//	200 OK: JSON of the user and their permission grants.
//	401 Unauthorized: with no authenticated user.
func (api *API) httpAuthMe(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpAuthMe", Secondary: getIP(r)}

	authCtx := api.authCtxOr401(w, r)
	if authCtx == nil {
		return
	}

	api.writeAuthMe(w, http.StatusOK, authCtx, logFrom)
}

// startSession mints a session for authCtx.User and sets the session cookie,
// reporting false on failure. WebSocket clients of any sessions evicted by
// the per-user cap are kicked.
func (api *API) startSession(w http.ResponseWriter, r *http.Request, authCtx *auth.Context, logFrom logx.LogFrom) bool {
	token, evicted, err := api.auth.Sessions.Start(r.Context(), authCtx.User.ID, getIP(r), r.UserAgent())
	if err != nil {
		logx.Error(err, logFrom, true)
		failRequest(&w, errors.New("authentication failed"), http.StatusInternalServerError)
		return false
	}
	api.kickSessionWebSocketClients(evicted...)
	http.SetCookie(w, api.sessionCookie(r, token, int(api.Config.Settings.AuthSessionLifetime().Seconds())))
	return true
}

// writeAuthMe writes the authenticated user and their grants - the body shared
// by /auth/login, /auth/setup and /auth/me.
func (api *API) writeAuthMe(
	w http.ResponseWriter,
	status int,
	authCtx *auth.Context,
	logFrom logx.LogFrom,
) {
	api.writeJSONStatus(w, status,
		apitype.AuthMe{User: authCtx.User, Permissions: authCtx.Grants},
		logFrom,
	)
}

// sessionCookie builds the session cookie (maxAge < 0 clears it).
// Secure when Argus terminates TLS itself, or when a trusted proxy
// reports the client connection as HTTPS.
func (api *API) sessionCookie(r *http.Request, token string, maxAge int) *http.Cookie {
	path := api.RoutePrefix
	if path == "" {
		path = "/"
	}
	secure := api.Config.Settings.WebCertFile() != "" ||
		(api.fromTrustedProxy(r) &&
			strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"))
	return &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
}
