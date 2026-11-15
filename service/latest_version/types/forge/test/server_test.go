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

package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecordedRequest_String(t *testing.T) {
	tests := []struct {
		name       string
		header     http.Header
		wantMasked bool
		wantAuth   string
	}{
		{
			name:       "a credential is masked",
			header:     http.Header{"Authorization": []string{"token s3cret"}},
			wantMasked: true,
			wantAuth:   "token s3cret",
		},
		{
			name:       "no credential leaves nothing to mask",
			header:     http.Header{"If-None-Match": []string{`"abc"`}},
			wantMasked: false,
			wantAuth:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := RecordedRequest{Path: "/repos/o/r/releases", Header: tc.header}

			got := fmt.Sprintf("%v", request)

			if strings.Contains(got, "s3cret") {
				t.Errorf("RecordedRequest.String() mismatch\ngot:  %q\nwant: NOT %q", got, "s3cret")
			}
			if gotMasked := strings.Contains(got, "<redacted>"); gotMasked != tc.wantMasked {
				t.Errorf(
					"RecordedRequest.String() masking mismatch\n"+
						"got:  masked=%t (%q)\nwant: masked=%t",
					gotMasked, got, tc.wantMasked,
				)
			}

			if auth := request.Header.Get("Authorization"); auth != tc.wantAuth {
				t.Errorf(
					"RecordedRequest.Header[Authorization] mismatch\ngot:  %q\nwant: %q",
					auth, tc.wantAuth,
				)
			}
		})
	}
}

func TestStart(t *testing.T) {
	server := &Server{Endpoints: map[string]Endpoint{}}

	startServing(t, server)

	if got, want := server.EmptyPageBody, "[]"; got != want {
		t.Errorf("Server.EmptyPageBody mismatch\ngot:  %q\nwant: %q", got, want)
	}

	if server.NextPageLink == nil {
		t.Error("Server.NextPageLink mismatch\ngot:  nil\nwant: non-nil")
	}

	if server.URL == "" {
		t.Error("Server.URL mismatch\ngot:  \"\"\nwant: non-empty")
	}
}

func TestStart_doesNotOverrideWhatTheCallerSet(t *testing.T) {
	wantEmptyPage := "[]\n"
	server := &Server{
		Endpoints:     map[string]Endpoint{},
		EmptyPageBody: wantEmptyPage,
		NextPageLink: func(_ *http.Request, next, _ int) string {
			return fmt.Sprintf("page=%d", next)
		},
	}

	startServing(t, server)

	if got := server.EmptyPageBody; got != wantEmptyPage {
		t.Errorf("Server.EmptyPageBody mismatch\ngot:  %q\nwant: %q", got, wantEmptyPage)
	}

	if got, want := server.NextPageLink(nil, 4, 9), "page=4"; got != want {
		t.Errorf("Server.NextPageLink() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestRelativeNextPageLink(t *testing.T) {
	tests := []struct {
		name       string
		next, last int
		want       string
	}{
		{
			name: "a known last page is named alongside the next",
			next: 2,
			last: 5,
			want: `</repos/o/r/releases?page=2>; rel="next", </repos/o/r/releases?page=5>; rel="last"`,
		},
		{
			name: "no known last page names only the next",
			next: 7,
			last: 0,
			want: `</repos/o/r/releases?page=7>; rel="next"`,
		},
		{
			name: "a negative last page names only the next",
			next: 3,
			last: -1,
			want: `</repos/o/r/releases?page=3>; rel="next"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, "/repos/o/r/releases", nil)

			got := RelativeNextPageLink(request, tc.next, tc.last)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestServer_Serve(t *testing.T) {
	const (
		releasesBody = `[{"tag_name":"1.2.3"}]`
		notFoundBody = `{"message":"nothing here"}`
		refusedBody  = `{"message":"no scope"}`
	)

	tests := []struct {
		name        string
		endpoints   map[string]Endpoint
		requireAuth string
		path        string
		header      http.Header
		wantStatus  int
		wantBody    string
		wantHeaders map[string]string
	}{
		{
			name:       "invalid/an unmapped endpoint gets the not-found body",
			endpoints:  map[string]Endpoint{},
			path:       "/repos/o/r/releases",
			wantStatus: http.StatusNotFound,
			wantBody:   notFoundBody,
		},
		{
			name: "valid/a mapped endpoint serves its body",
			endpoints: map[string]Endpoint{
				"releases": {Body: releasesBody},
			},
			path:       "/repos/o/r/releases",
			wantStatus: http.StatusOK,
			wantBody:   releasesBody,
		},
		{
			name: "valid/the endpoint is chosen by the last path segment",
			endpoints: map[string]Endpoint{
				"releases": {Body: releasesBody},
				"tags":     {Body: `[{"name":"1.0.0"}]`},
			},
			path:       "/repos/o/r/tags",
			wantStatus: http.StatusOK,
			wantBody:   `[{"name":"1.0.0"}]`,
		},
		{
			name: "invalid/a request missing the required Authorization is refused",
			endpoints: map[string]Endpoint{
				"releases": {Body: releasesBody},
			},
			requireAuth: "token wanted",
			path:        "/repos/o/r/releases",
			wantStatus:  http.StatusUnauthorized,
			wantBody:    refusedBody,
		},
		{
			name: "valid/a request carrying the required Authorization is served",
			endpoints: map[string]Endpoint{
				"releases": {Body: releasesBody},
			},
			requireAuth: "token wanted",
			path:        "/repos/o/r/releases",
			header:      http.Header{"Authorization": []string{"token wanted"}},
			wantStatus:  http.StatusOK,
			wantBody:    releasesBody,
		},
		{
			name: "valid/a strong ETag asked with its own value is answered 304",
			endpoints: map[string]Endpoint{
				"releases": {Body: releasesBody, ETag: `"abc"`},
			},
			path:        "/repos/o/r/releases",
			header:      http.Header{"If-None-Match": []string{`"abc"`}},
			wantStatus:  http.StatusNotModified,
			wantBody:    "",
			wantHeaders: map[string]string{"ETag": `"abc"`},
		},
		{
			name: "valid/a weak served ETag matches the stripped value it is asked with",
			endpoints: map[string]Endpoint{
				"releases": {Body: releasesBody, ETag: `W/"abc"`},
			},
			path:       "/repos/o/r/releases",
			header:     http.Header{"If-None-Match": []string{`"abc"`}},
			wantStatus: http.StatusNotModified,
			wantBody:   "",
		},
		{
			name:       "valid/a strong served ETag matches the weak value it is asked with",
			endpoints:  map[string]Endpoint{"releases": {Body: releasesBody, ETag: `"abc"`}},
			path:       "/repos/o/r/releases",
			header:     http.Header{"If-None-Match": []string{`W/"abc"`}},
			wantStatus: http.StatusNotModified,
			wantBody:   "",
		},
		{
			name:        "valid/a different ETag serves the body, and the ETag to ask with next",
			endpoints:   map[string]Endpoint{"releases": {Body: releasesBody, ETag: `W/"abc"`}},
			path:        "/repos/o/r/releases",
			header:      http.Header{"If-None-Match": []string{`"stale"`}},
			wantStatus:  http.StatusOK,
			wantBody:    releasesBody,
			wantHeaders: map[string]string{"ETag": `W/"abc"`},
		},
		{
			name:       "valid/an unconditional request to an ETag endpoint serves the body",
			endpoints:  map[string]Endpoint{"releases": {Body: releasesBody, ETag: `W/"abc"`}},
			path:       "/repos/o/r/releases",
			wantStatus: http.StatusOK,
			wantBody:   releasesBody,
		},
		{
			name: "valid/page one of a paginated endpoint names the next and last pages",
			endpoints: map[string]Endpoint{
				"releases": {Pages: []string{"page-1", "page-2", "page-3"}},
			},
			path:       "/repos/o/r/releases",
			wantStatus: http.StatusOK,
			wantBody:   "page-1",
			wantHeaders: map[string]string{
				"Link": `</repos/o/r/releases?page=2>; rel="next", </repos/o/r/releases?page=3>; rel="last"`,
			},
		},
		{
			name: "valid/a numbered page serves that page",
			endpoints: map[string]Endpoint{
				"releases": {Pages: []string{"page-1", "page-2", "page-3"}},
			},
			path:       "/repos/o/r/releases?page=2",
			wantStatus: http.StatusOK,
			wantBody:   "page-2",
		},
		{
			name: "valid/the last page names no next page",
			endpoints: map[string]Endpoint{
				"releases": {Pages: []string{"page-1", "page-2"}},
			},
			path:        "/repos/o/r/releases?page=2",
			wantStatus:  http.StatusOK,
			wantBody:    "page-2",
			wantHeaders: map[string]string{"Link": ""},
		},
		{
			name: "valid/a page past the last gets the empty-page body",
			endpoints: map[string]Endpoint{
				"releases": {Pages: []string{"page-1"}},
			},
			path:       "/repos/o/r/releases?page=9",
			wantStatus: http.StatusOK,
			wantBody:   "[]",
		},
		{
			name: "valid/page zero gets the empty-page body",
			endpoints: map[string]Endpoint{
				"releases": {Pages: []string{"page-1"}},
			},
			path:       "/repos/o/r/releases?page=0",
			wantStatus: http.StatusOK,
			wantBody:   "[]",
		},
		{
			name: "valid/an unparsable page number gets the empty-page body",
			endpoints: map[string]Endpoint{
				"releases": {Pages: []string{"page-1"}},
			},
			path:       "/repos/o/r/releases?page=later",
			wantStatus: http.StatusOK,
			wantBody:   "[]",
		},
		{
			name: "valid/an endless endpoint serves its body and always names a next page",
			endpoints: map[string]Endpoint{
				"releases": {Body: releasesBody, Endless: true},
			},
			path:       "/repos/o/r/releases?page=40",
			wantStatus: http.StatusOK,
			wantBody:   releasesBody,
			wantHeaders: map[string]string{
				"Link": `</repos/o/r/releases?page=41>; rel="next"`,
			},
		},
		{
			name: "valid/a configured status code and headers are applied",
			endpoints: map[string]Endpoint{
				"releases": {
					Status:  http.StatusTeapot,
					Headers: map[string]string{"X-Rate-Limit": "0"},
					Body:    releasesBody,
				},
			},
			path:        "/repos/o/r/releases",
			wantStatus:  http.StatusTeapot,
			wantBody:    releasesBody,
			wantHeaders: map[string]string{"X-Rate-Limit": "0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := startServing(t, &Server{
				Endpoints:        tc.endpoints,
				RequireAuth:      tc.requireAuth,
				NotFoundBody:     notFoundBody,
				UnauthorizedBody: refusedBody,
			})

			response, body := get(t, server, tc.path, tc.header)

			if got, want := response.StatusCode, tc.wantStatus; got != want {
				t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, body, want)
			}

			if body != tc.wantBody {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", body, tc.wantBody)
			}

			for key, want := range tc.wantHeaders {
				if got := response.Header.Get(key); got != want {
					t.Errorf("Response.Header[%q] mismatch\ngot:  %q\nwant: %q", key, got, want)
				}
			}
		})
	}
}

func TestServer_Requests(t *testing.T) {
	server := startServing(t, &Server{
		Endpoints: map[string]Endpoint{"releases": {Body: "[]"}},
	})
	header := http.Header{"If-None-Match": []string{`"abc"`}}
	get(t, server, "/repos/o/r/releases?page=2&limit=50", header)

	requests := server.Requests()

	if got, want := len(requests), 1; got != want {
		t.Fatalf("before mutation: len(Server.Requests()) mismatch\ngot:  %d\nwant: %d", got, want)
	}

	if got, want := requests[0].Path, "/repos/o/r/releases"; got != want {
		t.Errorf("before mutation: RecordedRequest.Path mismatch\ngot:  %q\nwant: %q", got, want)
	}
	if got, want := requests[0].Query.Get("page"), "2"; got != want {
		t.Errorf(
			"before mutation: RecordedRequest.Query[page] mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
	if got, want := requests[0].Header.Get("If-None-Match"), `"abc"`; got != want {
		t.Errorf(
			"before mutation: RecordedRequest.Header[If-None-Match] mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}

	requests[0].Path = "mutated"
	if got, want := server.Requests()[0].Path, "/repos/o/r/releases"; got != want {
		t.Errorf(
			"after mutation: Server.Requests()[0].Path mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
}

func TestServer_NotModifiedCount(t *testing.T) {
	server := startServing(t, &Server{
		Endpoints: map[string]Endpoint{"releases": {Body: "[]", ETag: `W/"abc"`}},
	})
	conditional := http.Header{"If-None-Match": []string{`"abc"`}}

	if got, want := server.NotModifiedCount(), 0; got != want {
		t.Errorf(
			"before any request: Server.NotModifiedCount() mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}

	get(t, server, "/repos/o/r/releases", nil)

	if got, want := server.NotModifiedCount(), 0; got != want {
		t.Errorf("after a 200: Server.NotModifiedCount() mismatch\ngot:  %d\nwant: %d", got, want)
	}

	get(t, server, "/repos/o/r/releases", conditional)
	get(t, server, "/repos/o/r/releases", conditional)

	if got, want := server.NotModifiedCount(), 2; got != want {
		t.Errorf(
			"after two 304s: Server.NotModifiedCount() mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}
}

func TestStrongETag(t *testing.T) {
	tests := []struct {
		name string
		eTag string
		want string
	}{
		{
			name: "a weak validator loses its prefix",
			eTag: `W/"abc"`,
			want: `"abc"`,
		},
		{
			name: "a strong validator is unchanged",
			eTag: `"abc"`,
			want: `"abc"`,
		},
		{
			name: "an empty ETag stays empty",
			eTag: "",
			want: "",
		},
		{
			name: "only a leading prefix is stripped",
			eTag: `"W/abc"`,
			want: `"W/abc"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := StrongETag(tc.eTag)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
