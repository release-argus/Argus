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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var packageName = "latestver_forgejo"

// releasesBody is a /releases response in the shape a forge returns, trimmed to the
// fields Argus reads. 0.17.4 is the highest stable semantic version, and was published before 0.18.0-rc1.
var releasesBody = test.TrimJSON(`[
	{
		"tag_name":"0.16.9",
		"name":"0.16.9",
		"prerelease":false,
		"published_at":"2000-01-01T00:00:00Z",
		"assets":[
			{"id":9,"name":"Argus-0.16.9.linux-amd64","created_at":"2000-01-01T00:01:00Z","browser_download_url":"https://forge.example.com/owner/repo/releases/download/0.16.9/Argus-0.16.9.linux-amd64"}
		]
	},
	{
		"tag_name":"0.17.4",
		"name":"0.17.4",
		"prerelease":false,
		"published_at":"2000-01-02T00:00:00Z",
		"assets":[
			{"id":3,"name":"Argus-0.17.4.linux-amd64","created_at":"2000-01-02T00:01:00Z","browser_download_url":"https://forge.example.com/owner/repo/releases/download/0.17.4/Argus-0.17.4.linux-amd64"}
		]
	},
	{
		"tag_name":"0.18.0-rc1",
		"name":"0.18.0-rc1",
		"prerelease":true,
		"published_at":"2000-01-02T12:00:00Z",
		"assets":[]
	}
]`)

// tagsBody is a /tags response. Tags carry no "tag_name", "prerelease" or "published_at".
var tagsBody = test.TrimJSON(`[
	{"name":"0.12.1","id":"6e56b5ebad3fb05036b1ff68a6b47b80f5859c7c"},
	{"name":"0.12.0","id":"1fa4a03ee2fb2a2cc9dbbbbd6a9e2b3d7d2c93f0"}
]`)

// Empty-list encodings - some instances append a newline, others do not.
const (
	emptyListNewline = "[]\n"
	emptyListCompact = "[]"
)

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// forgeEndpoint is one endpoint's reply.
//
// With `pages` set, the endpoint paginates: page N serves pages[N-1], and every page but the
// last advertises the next through a Link header.
type forgeEndpoint struct {
	status  int               // Response status code. Defaults to 200.
	headers map[string]string // Extra response headers.
	body    string            // Response body, when not paginating.
	pages   []string          // One body per page.
	endless bool              // Serve `body` on every page, always naming a next page.
}

// recordedRequest is a request the fixture server received.
type recordedRequest struct {
	path   string
	query  url.Values
	header http.Header
}

// forgeServer is a fixture standing in for a Forgejo instance.
type forgeServer struct {
	*httptest.Server

	// endpoints maps the last path segment ("releases"/"tags") to its reply. An unmapped
	// segment gets the 404 a real instance gives for a missing repository, or for one with
	// the feature disabled.
	endpoints map[string]forgeEndpoint

	mu       sync.Mutex
	received []recordedRequest
}

// newForgeServer starts a fixture serving `endpoints`.
func newForgeServer(t *testing.T, endpoints map[string]forgeEndpoint) *forgeServer {
	t.Helper()

	return startForgeServer(t, endpoints, httptest.NewServer)
}

// newForgeServerTLS is [newForgeServer] over HTTPS, presenting a certificate no client trusts.
func newForgeServerTLS(t *testing.T, endpoints map[string]forgeEndpoint) *forgeServer {
	t.Helper()

	return startForgeServer(t, endpoints, httptest.NewTLSServer)
}

// startForgeServer starts a fixture serving `endpoints`.
func startForgeServer(
	t *testing.T,
	endpoints map[string]forgeEndpoint,
	start func(http.Handler) *httptest.Server,
) *forgeServer {
	t.Helper()

	server := &forgeServer{endpoints: endpoints}
	server.Server = start(http.HandlerFunc(server.serve))
	t.Cleanup(server.Close)

	return server
}

// serve replies to a request for one of the configured endpoints.
func (s *forgeServer) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.received = append(s.received, recordedRequest{
		path:   r.URL.Path,
		query:  r.URL.Query(),
		header: r.Header.Clone(),
	})
	s.mu.Unlock()

	endpoint, known := s.endpoints[path.Base(r.URL.Path)]
	if !known {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"The target couldn't be found.","url":"https://forge.example.com/api/swagger","errors":[]}`))
		return
	}

	body := endpoint.body
	if endpoint.endless || len(endpoint.pages) > 0 {
		page := 1
		if raw := r.URL.Query().Get("page"); raw != "" {
			page, _ = strconv.Atoi(raw)
		}

		switch {
		case endpoint.endless:
			w.Header().Set("Link", fmt.Sprintf(
				`<https://root-url.example.com/api/v1/repos/o/r/x?limit=50&page=%d>; rel="next"`,
				page+1,
			))
		case page < 1 || page > len(endpoint.pages):
			body = emptyListNewline
		default:
			body = endpoint.pages[page-1]
			if page < len(endpoint.pages) {
				w.Header().Set("Link", fmt.Sprintf(
					`<https://root-url.example.com/api/v1/repos/o/r/x?limit=50&page=%d>; rel="next",`+
						`<https://root-url.example.com/api/v1/repos/o/r/x?limit=50&page=%d>; rel="last"`,
					page+1, len(endpoint.pages),
				))
			}
		}
	}

	for key, value := range endpoint.headers {
		w.Header().Set(key, value)
	}
	if endpoint.status != 0 {
		w.WriteHeader(endpoint.status)
	}
	_, _ = w.Write([]byte(body))
}

// requests returns the requests the server received.
func (s *forgeServer) requests() []recordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]recordedRequest(nil), s.received...)
}

// testLookup returns a Lookup decoded from `lookupYAML`.
func testLookup(t *testing.T, lookupYAML string) *Lookup {
	t.Helper()

	// Options.
	options, _ := opt.Decode(
		"yaml", nil,
		opttest.PlainDefaultsConfig(t),
	)
	// Status.
	svcStatus, _ := statustest.New("yaml", nil)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "forgejo-testLookup",
		},
		&dashboard.Options{
			WebURL: "https://example.com",
		},
	)

	lookupCfg := plainDefaultsConfig(t)
	lookup, err := Decode(
		"yaml", []byte(test.TrimYAML(lookupYAML)),
		options,
		svcStatus,
		lookupCfg,
	)
	if err != nil {
		t.Fatalf(
			"%s\nfailed to decode Lookup: %v",
			packageName, err,
		)
	}
	lookup.Init(options, svcStatus, lookupCfg)

	typeHardDefaults := &Defaults{}
	typeHardDefaults.Default()
	lookup.SetTypeDefaults(&Defaults{}, typeHardDefaults)

	return lookup
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig(t *testing.T) base.DefaultsConfig {
	t.Helper()

	optDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults.Default()

	defaults, _ := base.DecodeDefaults("yaml", nil)
	defaults.Options = optDefaults
	hardDefaults, _ := base.DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	hardDefaults.Options = optHardDefaults

	defaults.Require.SetDefaults(&hardDefaults.Require)

	return base.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}

// newResponse builds a response to a GET request for `address` for the
// status-map tests.
func newResponse(
	t *testing.T,
	address string,
	status int,
	headers map[string]string,
	body string,
) *http.Response {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		t.Fatalf(
			"%s\nfailed to build the request: %v",
			packageName, err,
		)
	}

	response := &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
	for key, value := range headers {
		response.Header.Set(key, value)
	}

	return response
}
