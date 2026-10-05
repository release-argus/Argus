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
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	forgetest "github.com/release-argus/Argus/service/latest_version/types/forge/test"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
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

// preReleaseTagsBody is a /tags response whose newest tag is a pre-release, which only
// its name says.
var preReleaseTagsBody = test.TrimJSON(`[
	{"name":"0.13.0-rc1","id":"8a1c0d4e2f6b9c3a5d7e1f0b2c4a6e8d0f2b4c6a"},
	{"name":"0.12.1","id":"6e56b5ebad3fb05036b1ff68a6b47b80f5859c7c"}
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
type forgeEndpoint = forgetest.Endpoint

// forgeServer is a fixture standing in for a Forgejo instance.
type forgeServer = forgetest.Server

// newForgeServer starts a fixture serving `endpoints`.
func newForgeServer(t *testing.T, endpoints map[string]forgeEndpoint) *forgeServer {
	t.Helper()

	return startForgeServer(t, &forgeServer{Endpoints: endpoints}, httptest.NewServer)
}

// startForgeServer starts `server`, giving it the replies a Forgejo instance gives.
func startForgeServer(
	t *testing.T,
	server *forgeServer,
	start func(http.Handler) *httptest.Server,
) *forgeServer {
	t.Helper()

	server.UnauthorizedBody = `{"message":"token does not have at least one of required scope(s)"}`
	server.NotFoundBody = `{"message":"The target couldn't be found.","url":"https://forge.example.com/api/swagger","errors":[]}`
	server.EmptyPageBody = emptyListNewline
	server.NextPageLink = forgeNextPageLink

	return forgetest.Start(t, server, start)
}

// forgeNextPageLink returns a Link header for a paginated list, as a Forgejo instance would.
func forgeNextPageLink(_ *http.Request, next, last int) string {
	const root = "https://root-url.example.com/api/v1/repos/o/r/x"
	if last <= 0 {
		return fmt.Sprintf(`<%s?limit=50&page=%d>; rel="next"`, root, next)
	}

	return fmt.Sprintf(
		`<%s?limit=50&page=%d>; rel="next",`+`<%s?limit=50&page=%d>; rel="last"`,
		root, next, root, last,
	)
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
	svcStatus, _ := statustest.New("yaml", []byte(`id: forgejo-testLookup`))

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
	accessToken string,
) *http.Response {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		t.Fatalf(
			"%s\nfailed to build the request: %v",
			packageName, err,
		)
	}
	if accessToken != "" {
		request.Header.Set("Authorization", "token "+accessToken)
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

// setHostDefaults layers host-keyed defaults onto a [Lookup].
func setHostDefaults(lookup *Lookup, defaults, hardDefaults map[string]HostDefaults) {
	typeDefaults, typeHardDefaults := lookup.GetTypeDefaults()
	typeDefaults.Host = defaults
	typeHardDefaults.Host = hardDefaults
}

// hostPlaceholder stands in for a fixture's URL, which is only known once it has started.
const hostPlaceholder = "HOST"

// hostKeyedAt rewrites the [hostPlaceholder] in each entry's URL to `url`, so a
// fixture can name an instance whose address is only known once it has started.
func hostKeyedAt(hosts map[string]HostDefaults, url string) map[string]HostDefaults {
	if hosts == nil {
		return nil
	}

	keyed := make(map[string]HostDefaults, len(hosts))
	for name, entry := range hosts {
		entry.URL = strings.Replace(entry.URL, hostPlaceholder, url, 1)
		keyed[name] = entry
	}

	return keyed
}
