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

// Package test provides a fixture server standing in for a forge's API.
package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Endpoint is one endpoint's reply.
//
// With `Pages` set, the endpoint paginates: page N serves Pages[N-1], and every page but
// the last advertises the next through a Link header.
type Endpoint struct {
	Status  int               // Response status code. Defaults to 200.
	Headers map[string]string // Extra response headers.
	Body    string            // Response body, when not paginating.
	Pages   []string          // One body per page.
	Endless bool              // Serve `Body` on every page, always naming a next page.
	ETag    string            // ETag to serve, and to answer a matching If-None-Match with 304.
}

// RecordedRequest is a request the fixture server received.
type RecordedRequest struct {
	Path   string
	Query  url.Values
	Header http.Header
}

// String implements [fmt.Stringer], masking the credential so that printing a request
// in a failure message cannot carry it into a log. Reading Header is unaffected.
func (r RecordedRequest) String() string {
	header := r.Header
	if header.Get("Authorization") != "" {
		header = r.Header.Clone()
		header.Set("Authorization", "<redacted>")
	}

	return fmt.Sprintf("{%s %v %v}", r.Path, r.Query, header)
}

// Server is a fixture standing in for a forge's API.
type Server struct {
	*httptest.Server

	// Endpoints maps the last path segment ("releases"/"tags") to its reply. An unmapped
	// segment gets [Server.NotFoundBody].
	Endpoints map[string]Endpoint

	// RequireAuth, when set, is the Authorization every request must carry.
	RequireAuth string

	UnauthorizedBody string // Body for a request failing [Server.RequireAuth].
	NotFoundBody     string // Body for an unmapped endpoint.
	EmptyPageBody    string // Body for a page outside the paginated range. Defaults to "[]".

	// NextPageLink builds the Link header naming page `next`. `last` is 0 when no last
	// page is known, as for an endless walk.
	NextPageLink func(r *http.Request, next, last int) string

	mu          sync.Mutex
	received    []RecordedRequest
	notModified int
}

// Start starts `server`, filling in any reply the caller left unset.
//
// `start` is [httptest.NewServer], or [httptest.NewTLSServer] to serve over TLS.
func Start(
	t *testing.T,
	server *Server,
	start func(http.Handler) *httptest.Server,
) *Server {
	t.Helper()

	if server.EmptyPageBody == "" {
		server.EmptyPageBody = "[]"
	}
	if server.NextPageLink == nil {
		server.NextPageLink = RelativeNextPageLink
	}

	server.Server = start(http.HandlerFunc(server.serve))
	t.Cleanup(server.Close)

	return server
}

// RelativeNextPageLink names the next, and any known last, page relative to the request.
func RelativeNextPageLink(r *http.Request, next, last int) string {
	if last <= 0 {
		return fmt.Sprintf(`<%s?page=%d>; rel="next"`, r.URL.Path, next)
	}

	return fmt.Sprintf(
		`<%s?page=%d>; rel="next", <%s?page=%d>; rel="last"`,
		r.URL.Path, next, r.URL.Path, last,
	)
}

// serve replies to a request for one of the configured endpoints.
func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.received = append(s.received, RecordedRequest{
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Header: r.Header.Clone(),
	})
	s.mu.Unlock()

	if s.RequireAuth != "" && r.Header.Get("Authorization") != s.RequireAuth {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(s.UnauthorizedBody))
		return
	}

	endpoint, known := s.Endpoints[path.Base(r.URL.Path)]
	if !known {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(s.NotFoundBody))
		return
	}

	// Serve the ETag, answering a weakly-matching If-None-Match with 304.
	if endpoint.ETag != "" {
		w.Header().Set("ETag", endpoint.ETag)
		if StrongETag(r.Header.Get("If-None-Match")) == StrongETag(endpoint.ETag) {
			s.mu.Lock()
			s.notModified++
			s.mu.Unlock()

			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	body := endpoint.Body
	if endpoint.Endless || len(endpoint.Pages) > 0 {
		page := 1
		if raw := r.URL.Query().Get("page"); raw != "" {
			page, _ = strconv.Atoi(raw)
		}

		switch {
		case endpoint.Endless:
			w.Header().Set("Link", s.NextPageLink(r, page+1, 0))
		case page < 1 || page > len(endpoint.Pages):
			body = s.EmptyPageBody
		default:
			body = endpoint.Pages[page-1]
			if page < len(endpoint.Pages) {
				w.Header().Set("Link", s.NextPageLink(r, page+1, len(endpoint.Pages)))
			}
		}
	}

	for key, value := range endpoint.Headers {
		w.Header().Set(key, value)
	}
	if endpoint.Status != 0 {
		w.WriteHeader(endpoint.Status)
	}
	_, _ = w.Write([]byte(body))
}

// Requests returns the requests the server received.
func (s *Server) Requests() []RecordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]RecordedRequest(nil), s.received...)
}

// NotModifiedCount returns how many requests the server answered with a 304.
func (s *Server) NotModifiedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.notModified
}

// StrongETag strips the weak-validator prefix, so the weak and strong spelling of one
// ETag compare equal.
func StrongETag(etag string) string {
	return strings.TrimPrefix(etag, `W/`)
}
