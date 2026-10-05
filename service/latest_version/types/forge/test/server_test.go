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

func TestStart(t *testing.T) {
	// GIVEN: a Server left with no replies configured.
	server := &Server{Endpoints: map[string]Endpoint{}}

	// WHEN: it is started.
	startServing(t, server)

	prefix := packageName + "\nStart()"

	// THEN: the empty-page body falls back to the compact encoding.
	if got, want := server.EmptyPageBody, "[]"; got != want {
		t.Errorf(
			"%s EmptyPageBody mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// AND: a Link builder is supplied.
	if server.NextPageLink == nil {
		t.Errorf("%s left NextPageLink nil", prefix)
	}

	// AND: it is listening.
	if server.URL == "" {
		t.Errorf("%s did not set a URL", prefix)
	}
}

func TestStart_doesNotOverrideWhatTheCallerSet(t *testing.T) {
	// GIVEN: a Server whose replies the caller configured.
	wantEmptyPage := "[]\n"
	server := &Server{
		Endpoints:     map[string]Endpoint{},
		EmptyPageBody: wantEmptyPage,
		NextPageLink: func(_ *http.Request, next, _ int) string {
			return fmt.Sprintf("page=%d", next)
		},
	}

	// WHEN: it is started.
	startServing(t, server)

	prefix := packageName + "\nStart()"

	// THEN: the caller's empty-page body survives.
	if got := server.EmptyPageBody; got != wantEmptyPage {
		t.Errorf(
			"%s overwrote EmptyPageBody\ngot:  %q\nwant: %q",
			prefix, got, wantEmptyPage,
		)
	}

	// AND: so does the caller's Link builder.
	if got, want := server.NextPageLink(nil, 4, 9), "page=4"; got != want {
		t.Errorf(
			"%s overwrote NextPageLink\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
}

func TestRelativeNextPageLink(t *testing.T) {
	// GIVEN: a request, and the pages to name.
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

			// WHEN: the Link header is built.
			got := RelativeNextPageLink(request, tc.next, tc.last)

			// THEN: it names the pages as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRelativeNextPageLink(next=%d, last=%d) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.next, tc.last,
					got, tc.want,
				)
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

	// GIVEN: a fixture serving the configured endpoints.
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

			// WHEN: the endpoint is requested.
			response, body := get(t, server, tc.path, tc.header)

			prefix := fmt.Sprintf(
				"%s\nServer.serve(%q)",
				packageName, tc.path,
			)

			// THEN: the status code is as expected.
			if got := response.StatusCode; got != tc.wantStatus {
				t.Errorf(
					"%s status code mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantStatus,
				)
			}

			// AND: the body is as expected.
			if body != tc.wantBody {
				t.Errorf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, body, tc.wantBody,
				)
			}

			// AND: the response headers are as expected.
			for key, want := range tc.wantHeaders {
				if got := response.Header.Get(key); got != want {
					t.Errorf(
						"%s %s header mismatch\ngot:  %q\nwant: %q",
						prefix, key, got, want,
					)
				}
			}
		})
	}
}

func TestServer_Requests(t *testing.T) {
	// GIVEN: a fixture that has answered a request.
	server := startServing(t, &Server{
		Endpoints: map[string]Endpoint{"releases": {Body: "[]"}},
	})
	header := http.Header{"If-None-Match": []string{`"abc"`}}
	get(t, server, "/repos/o/r/releases?page=2&limit=50", header)

	prefix := packageName + "\nServer.Requests()"

	// WHEN: the requests are read.
	requests := server.Requests()

	// THEN: the one request was recorded.
	if got, want := len(requests), 1; got != want {
		t.Fatalf(
			"%s count mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}

	// AND: its path, query and headers were all captured.
	if got, want := requests[0].Path, "/repos/o/r/releases"; got != want {
		t.Errorf(
			"%s Path mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
	if got, want := requests[0].Query.Get("page"), "2"; got != want {
		t.Errorf(
			"%s Query[page] mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
	if got, want := requests[0].Header.Get("If-None-Match"), `"abc"`; got != want {
		t.Errorf(
			"%s Header[If-None-Match] mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// AND: the slice is a copy, so a caller cannot disturb the record.
	requests[0].Path = "mutated"
	if got, want := server.Requests()[0].Path, "/repos/o/r/releases"; got != want {
		t.Errorf(
			"%s handed out the backing slice\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
}

func TestRecordedRequest_String(t *testing.T) {
	// GIVEN: requests carrying a credential, or not.
	tests := []struct {
		name       string
		header     http.Header
		wantMasked bool
	}{
		{
			name:       "a credential is masked",
			header:     http.Header{"Authorization": []string{"token s3cret"}},
			wantMasked: true,
		},
		{
			name:       "no credential leaves nothing to mask",
			header:     http.Header{"If-None-Match": []string{`"abc"`}},
			wantMasked: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := RecordedRequest{Path: "/repos/o/r/releases", Header: tc.header}

			// WHEN: it is printed the way a failure message does.
			got := fmt.Sprintf("%v", request)

			prefix := fmt.Sprintf("%s\nRecordedRequest.String()", packageName)

			// THEN: the credential never appears.
			if strings.Contains(got, "s3cret") {
				t.Errorf(
					"%s leaked the credential\ngot:  %s",
					prefix, got,
				)
			}
			if gotMasked := strings.Contains(got, "<redacted>"); gotMasked != tc.wantMasked {
				t.Errorf(
					"%s masking mismatch\ngot:  %s\nwant: masked=%t",
					prefix, got, tc.wantMasked,
				)
			}

			// AND: reading the header still gives the real value.
			if g, w := request.Header.Get("Authorization"), tc.header.Get("Authorization"); g != w {
				t.Errorf(
					"%s must not alter the recorded header\ngot:  %q\nwant: %q",
					prefix, g, w,
				)
			}
		})
	}
}

func TestServer_NotModifiedCount(t *testing.T) {
	// GIVEN: a fixture serving an endpoint with an ETag.
	server := startServing(t, &Server{
		Endpoints: map[string]Endpoint{"releases": {Body: "[]", ETag: `W/"abc"`}},
	})
	conditional := http.Header{"If-None-Match": []string{`"abc"`}}

	prefix := packageName + "\nServer.NotModifiedCount()"

	// THEN: it starts at zero.
	if got := server.NotModifiedCount(); got != 0 {
		t.Errorf(
			"%s should start at 0\ngot:  %d",
			prefix, got,
		)
	}

	// WHEN: a request that does not match the ETag is made.
	get(t, server, "/repos/o/r/releases", nil)

	// THEN: it is not counted.
	if got := server.NotModifiedCount(); got != 0 {
		t.Errorf(
			"%s counted a 200\ngot:  %d\nwant: 0",
			prefix, got,
		)
	}

	// WHEN: two conditional requests that do match are made.
	get(t, server, "/repos/o/r/releases", conditional)
	get(t, server, "/repos/o/r/releases", conditional)

	// THEN: both are counted.
	if got, want := server.NotModifiedCount(), 2; got != want {
		t.Errorf(
			"%s mismatch\ngot:  %d\nwant: %d",
			prefix, got, want,
		)
	}
}

func TestStrongETag(t *testing.T) {
	// GIVEN: an ETag in either spelling.
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

			// WHEN: it is normalised.
			got := StrongETag(tc.eTag)

			// THEN: the value compares as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nStrongETag(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.eTag,
					got, tc.want,
				)
			}
		})
	}
}
