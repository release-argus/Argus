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

// Package v1 provides the API for the webserver.
package v1

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/gorilla/mux"

	"github.com/release-argus/Argus/internal/logx"
)

// clientIPKey keys the resolved client IP in a request's context.
type clientIPKey struct{}

// clientIPMiddleware resolves each request's client IP once, honouring
// forwarded headers only when the peer is a trusted proxy
// (settings.web.trusted_proxies).
func (api *API) clientIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := remoteAddrIP(r)
		if api.fromTrustedProxy(r) {
			if forwarded := api.forwardedIP(r); forwarded != "" {
				ip = forwarded
			}
		} else if len(api.trustedProxies) == 0 && r.Header.Get("X-Forwarded-For") != "" {
			api.warnUntrustedProxy()
		}

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), clientIPKey{}, ip),
		))
	})
}

// warnUntrustedProxy logs, once, that forwarded headers are arriving while
// settings.web.trusted_proxies is empty. Every control keyed on the client IP
// (rate limiting, logging) then sees the proxy's address for every client.
func (api *API) warnUntrustedProxy() {
	api.untrustedProxyWarning.Do(func() {
		logx.Warn(
			"requests carry X-Forwarded-For but settings.web.trusted_proxies is empty,"+
				" so every client resolves to the proxy's address:"+
				" failed logins from any one client can rate-limit every other."+
				" Set settings.web.trusted_proxies to your proxy's IP/CIDR",
			logx.LogFrom{Primary: "web"}, true)
	})
}

// getIP returns the client IP resolved by [API.clientIPMiddleware],
// falling back to RemoteAddr when the middleware has not run.
func getIP(r *http.Request) string {
	if ip, ok := r.Context().Value(clientIPKey{}).(string); ok {
		return ip
	}
	return remoteAddrIP(r)
}

// forwardedIP returns the real client IP claimed by proxy headers ("" if none).
// Only call for requests from trusted proxies - the headers are spoofable.
//
// X-Forwarded-For is read right-to-left, returning the first address that is
// not itself a trusted proxy. CF-Connecting-Ip / X-Real-Ip are used as a
// fallback.
func (api *API) forwardedIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			addr, err := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if err != nil {
				continue
			}
			addr = addr.Unmap()
			if !api.isTrustedProxy(addr) {
				return addr.String()
			}
		}
	}

	if ip := r.Header.Get("CF-Connecting-Ip"); net.ParseIP(ip) != nil {
		return ip
	}
	if ip := r.Header.Get("X-Real-Ip"); net.ParseIP(ip) != nil {
		return ip
	}

	return ""
}

// remoteAddrIP returns the connection peer's IP ("" if unparseable).
func remoteAddrIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// No port (e.g. hand-built test requests).
		ip = r.RemoteAddr
	}
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}

// fromTrustedProxy reports whether the request's peer is a trusted proxy.
func (api *API) fromTrustedProxy(r *http.Request) bool {
	if len(api.trustedProxies) == 0 {
		return false
	}

	addr, err := netip.ParseAddr(remoteAddrIP(r))
	if err != nil {
		return false
	}
	return api.isTrustedProxy(addr)
}

// isTrustedProxy reports whether addr falls in any configured trusted-proxy range.
// addr is unmapped before matching so IPv4-in-IPv6 literals compare correctly
// against IPv4 prefixes.
func (api *API) isTrustedProxy(addr netip.Addr) bool {
	addr = addr.Unmap()
	for _, prefix := range api.trustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// basicAuthMiddleware handles basic authentication with hashed credentials.
// It rejects unauthorised requests with a 401 and closes the connection.
func (api *API) basicAuthMiddleware() mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if ok {
				// Hash purely to prevent ConstantTimeCompare leaking lengths.
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))

				// Protect from possible timing attacks.
				usernameMatch := ConstantTimeCompare(usernameHash, api.Config.Settings.WebBasicAuthUsernameHash())
				passwordMatch := ConstantTimeCompare(passwordHash, api.Config.Settings.WebBasicAuthPasswordHash())

				if usernameMatch && passwordMatch {
					h.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// loggerMiddleware logs the HTTP request method, client IP, and URL path before
// passing the request to the next handler in the chain.
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request.
		logx.Verbose(
			fmt.Sprintf(
				"%s (%s) %s %v",
				r.Method, getIP(r), r.URL.Path, r.URL.Query(),
			),
			logx.LogFrom{},
			true,
		)

		// Process request.
		next.ServeHTTP(w, r)
	})
}
