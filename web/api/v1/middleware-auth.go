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
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/provider"
	"github.com/release-argus/Argus/auth/provider/local"
	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/internal/logx"
)

// authCookieName is the session cookie's name.
const authCookieName = "argus_session"

// errUnauthorised is the uniform 401 message.
var errUnauthorised = errors.New("unauthorized")

// errTooManyAttempts reports a rate-limited authentication attempt.
var errTooManyAttempts = errors.New("too many login attempts; try again later")

// errForbidden is the uniform 403 message.
var errForbidden = errors.New("forbidden")

// AuthDeps bundles the authentication/authorisation dependencies the API
// needs when auth is enabled.
type AuthDeps struct {
	Store     *store.Store       // Users/groups/grants/sessions.
	Sessions  *session.Manager   // Session lifecycle.
	Providers *provider.Registry // Authentication providers.
}

// EnableAuth arms the API with user/RBAC authentication.
// Must be called before SetupRoutesAPI.
func (api *API) EnableAuth(deps *AuthDeps) {
	api.auth = deps
	api.loginLimiter = newLoginLimiter()
}

// authCtxKey keys the [auth.Context] in a request's context.
type authCtxKey struct{}

// timeNow returns the current time (overridable for tests).
// see [time.Now].
var timeNow = time.Now

// authContextFrom returns the request's resolved [auth.Context] (nil if absent).
func authContextFrom(r *http.Request) *auth.Context {
	authCtx, _ := r.Context().Value(authCtxKey{}).(*auth.Context)
	return authCtx
}

// authMiddleware authenticates every API request via the session cookie,
// falling back to Bearer API tokens for non-browser clients.
func (api *API) authMiddleware() mux.MiddlewareFunc {
	authPrefix := strings.TrimRight(api.RoutePrefix, "/") + "/api/v1/auth/"
	openPaths := map[string]bool{
		authPrefix + "login":  true,
		authPrefix + "setup":  true,
		authPrefix + "logout": true,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if openPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authCtx, err := api.authenticate(r)
			if err != nil {
				failAuthenticationError(w, r, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), authCtxKey{}, authCtx),
			))
		})
	}
}

// failAuthenticationError writes the response for a failed authentication:
//   - 401 for missing/invalid credentials.
//   - 500 (logged) for infrastructure failures.
func failAuthenticationError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, errUnauthorised) {
		failRequest(&w, errUnauthorised, http.StatusUnauthorized)
		return
	}
	logx.Error(err, logx.LogFrom{Primary: "auth", Secondary: getIP(r)}, true)
	failRequest(&w, errors.New("authentication failed"), http.StatusInternalServerError)
}

// authCtxOr401 returns the request's [auth.Context], failing the request with a
// 401 (and returning nil) when absent.
func (api *API) authCtxOr401(w http.ResponseWriter, r *http.Request) *auth.Context {
	authCtx := authContextFrom(r)
	if authCtx == nil {
		failRequest(&w, errUnauthorised, http.StatusUnauthorized)
	}
	return authCtx
}

// authenticate resolves a request's credentials to an [auth.Context].
// Missing/invalid credentials return [errUnauthorised]; infrastructure
// failures return their own error.
func (api *API) authenticate(r *http.Request) (*auth.Context, error) {
	// 1. Session cookie.
	if cookie, err := r.Cookie(authCookieName); err == nil {
		authCtx, err := api.authenticateSession(r.Context(), cookie.Value)
		if err == nil || !errors.Is(err, errUnauthorised) {
			return authCtx, err
		}
	}

	// 2. Bearer API token.
	if token, ok := bearerToken(r); ok {
		return api.authenticateAPIToken(r.Context(), token)
	}

	return nil, errUnauthorised
}

// bearerToken extracts a Bearer token from the 'Authorization' header.
func bearerToken(r *http.Request) (string, bool) {
	const scheme = "Bearer "
	header := r.Header.Get("Authorization")
	if len(header) > len(scheme) && strings.EqualFold(header[:len(scheme)], scheme) {
		return header[len(scheme):], true
	}
	return "", false
}

// authenticateAPIToken resolves a Bearer API token to an [auth.Context].
// The token carries the permissions of its owning user, evaluated at request time.
func (api *API) authenticateAPIToken(ctx context.Context, plaintext string) (*auth.Context, error) {
	token, err := api.auth.Store.APITokenByToken(ctx, plaintext)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errUnauthorised
		}
		return nil, err //nolint:wrapcheck
	}

	now := timeNow()
	if token.ExpiresAt != nil && now.After(*token.ExpiresAt) {
		return nil, errUnauthorised
	}

	// Record usage, at most once per minute; failures are only logged.
	if token.LastUsedAt == nil || now.Sub(*token.LastUsedAt) >= time.Minute {
		if err := api.auth.Store.TouchAPIToken(ctx, token.ID, now); err != nil {
			logx.Error(err, logx.LogFrom{Primary: "auth", Secondary: "token touch"}, true)
		}
	}

	return api.contextForUser(ctx, token.UserID, "api_token")
}

// authenticateSession resolves a session token to an [auth.Context].
func (api *API) authenticateSession(ctx context.Context, token string) (*auth.Context, error) {
	sess, err := api.auth.Sessions.Validate(ctx, token)
	if err != nil {
		if errors.Is(err, session.ErrInvalidSession) {
			return nil, errUnauthorised
		}
		return nil, err //nolint:wrapcheck
	}

	return api.contextForUser(ctx, sess.UserID, local.Name)
}

// verifyLocalCredentials authenticates username/password via the local
// provider, spending the failure rate-limiter's budget for ip.
// Bad credentials return [auth.ErrInvalidCredentials]; rate-limited
// attempts return [errTooManyAttempts] without any verification work.
func (api *API) verifyLocalCredentials(ctx context.Context, username, password, ip string) (*auth.Identity, error) {
	// Serialise this IP's attempts so check/recordFailure run atomically and a
	// concurrent burst cannot slip past the limit (distinct IPs stay concurrent).
	defer api.loginLimiter.lockIP(ip)()

	if !api.loginLimiter.check(ip) {
		return nil, errTooManyAttempts
	}

	localProvider := api.auth.Providers.Get(local.Name)
	if localProvider == nil {
		return nil, auth.ErrInvalidCredentials
	}

	identity, err := localProvider.Authenticate(ctx, username, password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			api.loginLimiter.recordFailure(ip)
		}
		return nil, err //nolint:wrapcheck
	}
	return identity, nil
}

// contextForUser builds the [auth.Context] of userID: the user record,
// its raw grants, and the evaluated permission set. Disabled or vanished
// users read as unauthorised.
//
// Runs on every authenticated request (grants are re-read, never cached)
// so a disable or grant change takes effect immediately.
func (api *API) contextForUser(ctx context.Context, userID, providerName string) (*auth.Context, error) {
	user, err := api.auth.Store.UserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errUnauthorised
		}
		return nil, err //nolint:wrapcheck
	}
	if !user.Enabled {
		return nil, errUnauthorised
	}

	grants, err := api.auth.Store.GrantsForUser(ctx, userID)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &auth.Context{
		Identity: auth.Identity{
			Provider:    providerName,
			Subject:     user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Email:       user.Email,
		},
		User:        *user,
		Grants:      grants,
		Permissions: rbac.NewPermissionSet(grants),
	}, nil
}

// originCheckMiddleware rejects requests whose Origin header disagrees with
// the request Host - CSRF. Trusted proxies may override the host they front
// via X-Forwarded-Host.
func (api *API) originCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch origin := r.Header.Get("Origin"); origin {
		case "":
			// No Origin header (same-origin, or non-browser client).
		case "null":
			// Opaque origin (sandboxed iframe, data:/file: context) - untrusted
			// for state-changing requests.
			if !safeMethod(r.Method) {
				failRequest(&w, errForbidden, http.StatusForbidden)
				return
			}
		default:
			host := r.Host
			if api.fromTrustedProxy(r) {
				if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
					host = forwardedHost
				}
			}
			if parsed, err := url.Parse(origin); err != nil || parsed.Host != host {
				failRequest(&w, errForbidden, http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// safeMethod reports whether method is read-only (no CSRF risk).
func safeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// authGuard enforces authentication before calling handler and delegates the
// authorisation decision to allow. It returns 401 for unauthenticated requests
// and 403 for requests rejected by allow.
// When auth is disabled, handler is returned unchanged.
func (api *API) authGuard(
	allow func(authCtx *auth.Context, r *http.Request) bool,
	handler http.HandlerFunc,
) http.HandlerFunc {
	if api.auth == nil {
		return handler
	}

	return func(w http.ResponseWriter, r *http.Request) {
		authCtx := authContextFrom(r)
		if authCtx == nil {
			failRequest(&w, errUnauthorised, http.StatusUnauthorized)
			return
		}
		if !allow(authCtx, r) {
			failRequest(&w, errForbidden, http.StatusForbidden)
			return
		}

		handler(w, r)
	}
}

// guard wraps handler with a permission check: the authenticated user must
// hold (resource, action) for the request's target. A nil target func means
// the operation has no per-service target (global grants only).
// When auth is disabled, handler is returned unchanged.
func (api *API) guard(
	resource rbac.Resource,
	action rbac.Action,
	target func(r *http.Request) *rbac.Target,
	handler http.HandlerFunc,
) http.HandlerFunc {
	return api.authGuard(func(authCtx *auth.Context, r *http.Request) bool {
		var tgt *rbac.Target
		if target != nil {
			tgt = target(r)
		}
		return authCtx.Permissions.Allowed(resource, action, tgt)
	}, handler)
}

// guardReadable wraps handler with a permission check for (resource, action) on
// the request's service target, and additionally requires `service:read` on that
// same target.
// When auth is disabled, handler is returned unchanged.
func (api *API) guardReadable(
	resource rbac.Resource,
	action rbac.Action,
	target func(r *http.Request) *rbac.Target,
	handler http.HandlerFunc,
) http.HandlerFunc {
	return api.authGuard(func(authCtx *auth.Context, r *http.Request) bool {
		tgt := target(r)
		return authCtx.Permissions.Allowed(resource, action, tgt) &&
			authCtx.Permissions.Allowed(rbac.ResourceService, rbac.ActionRead, tgt)
	}, handler)
}

// guardAnyScope wraps handler with a permission check satisfied by a grant of
// (resource, action) under any scope. For endpoints that serve no per-service
// data, which a scope-limited grant would otherwise be unable to reach.
// When auth is disabled, handler is returned unchanged.
func (api *API) guardAnyScope(
	resource rbac.Resource,
	action rbac.Action,
	handler http.HandlerFunc,
) http.HandlerFunc {
	return api.authGuard(func(authCtx *auth.Context, _ *http.Request) bool {
		return authCtx.Permissions.AllowedAnyScope(resource, action)
	}, handler)
}

// requireAdmin wraps handler so only members of the admin group may reach it.
// When auth is disabled, handler is returned unchanged.
func (api *API) requireAdmin(handler http.HandlerFunc) http.HandlerFunc {
	return api.authGuard(func(authCtx *auth.Context, _ *http.Request) bool {
		return slices.Contains(authCtx.User.Groups, store.GroupAdmin)
	}, handler)
}

// allowedServices returns the set of service IDs whose broadcasts the user
// may receive, or nil when unrestricted (global service:read).
// Evaluated at the WebSocket handshake and on each GET /service/order request.
func (api *API) allowedServices(authCtx *auth.Context) map[string]bool {
	// Global read: unrestricted.
	if authCtx.Permissions.Allowed(rbac.ResourceService, rbac.ActionRead, nil) {
		return nil
	}

	api.Config.OrderMu.RLock()
	defer api.Config.OrderMu.RUnlock()
	allowed := make(map[string]bool, len(api.Config.Order))
	for _, serviceID := range api.Config.Order {
		target := rbac.Target{ServiceID: serviceID}
		if svc := api.Config.Service[serviceID]; svc != nil {
			target.Tags = svc.Dashboard.Tags
		}
		if authCtx.Permissions.Allowed(rbac.ResourceService, rbac.ActionRead, &target) {
			allowed[serviceID] = true
		}
	}
	return allowed
}

// kickUserWebSocketClients is a nil-safe wrapper around [Hub.KickUserClients]
// that kicks all the WebSocket clients of userIDs.
func (api *API) kickUserWebSocketClients(userIDs ...string) {
	if api.auth == nil || api.hub == nil {
		return
	}
	api.hub.KickUserClients(userIDs...)
}

// kickSessionWebSocketClients is a nil-safe wrapper around [Hub.KickSessionClients]
// that disconnects the WebSocket clients connected under tokenHashes.
func (api *API) kickSessionWebSocketClients(tokenHashes ...string) {
	if api.auth == nil || api.hub == nil {
		return
	}
	api.hub.KickSessionClients(tokenHashes...)
}

// kickRestrictedWebSocketClients is a nil-safe wrapper around [Hub.KickRestrictedClients]
// that kicks all service-restricted WebSocket clients.
func (api *API) kickRestrictedWebSocketClients() {
	if api.auth == nil || api.hub == nil {
		return
	}
	api.hub.KickRestrictedClients()
}

// serviceTarget extracts the [rbac.Target] from the request's service_id
// query parameter, attaching the service's dashboard tags for
// service_tag-scope evaluation.
func (api *API) serviceTarget(r *http.Request) *rbac.Target {
	serviceID := r.URL.Query().Get("service_id")
	if serviceID == "" {
		return nil
	}

	target := rbac.Target{ServiceID: serviceID}

	api.Config.OrderMu.RLock()
	if svc := api.Config.Service[serviceID]; svc != nil {
		target.Tags = append([]string(nil), svc.Dashboard.Tags...)
	}
	api.Config.OrderMu.RUnlock()

	return &target
}
