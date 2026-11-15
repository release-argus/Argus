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

// Package github provides a github-based lookup type.
package github

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	lvbasetest "github.com/release-argus/Argus/service/latest_version/types/base/test"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	forgetest "github.com/release-argus/Argus/service/latest_version/types/forge/test"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var testBody = []byte(
	test.TrimJSON(`[
		{
			"tag_name":"0.18.0",
			"name":"0.18.0",
			"prerelease":true,
			"published_at":"2024-05-07T13:10:29Z",
			"assets":[
				{"id": 9,"name":"Argus-0.18.0.linux-amd64","created_at":"2024-05-07T13:11:30Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.18.0/Argus-0.18.0.darwin-amd64"},
				{"id": 5,"name":"Argus-0.18.0.linux-arm64","created_at":"2024-05-07T13:11:39Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.18.0/Argus-0.18.0.darwin-arm64"}
			]
		},
		{
			"tag_name":"0.17.4",
			"name":"0.17.4",
			"prerelease":false,
			"published_at":"2024-04-27T10:50:00Z",
			"assets":[
				{"id": 3,"name":"Argus-0.17.4.linux-amd64","created_at":"2024-04-27T10:50:53Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.17.4/Argus-0.17.4.linux-amd64"},
				{"id": 7,"name":"Argus-0.17.4.linux-arm","created_at":"2024-04-27T10:50:59Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.17.4/Argus-0.17.4.linux-arm"}
			]
		}
	]`),
)
var testBodyObject []forgetypes.Release

func TestMain(m *testing.M) {
	logtest.InitLog()

	if err := decode.Unmarshal("json", testBody, &testBodyObject); err != nil {
		fmt.Println("decode.Unmarshal() failed:", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

// newData returns a new Data.
func newData(
	etag string,
	releases *[]forgetypes.Release,
) *Data {
	// ETag - https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests.
	if etag == "" {
		etag = getEmptyListETag()
	}

	var releasesDeref []forgetypes.Release
	if releases != nil {
		releasesDeref = *releases
	}

	return &Data{
		etag:     etag,
		releases: releasesDeref,
	}
}

// releaseEqual reports whether a and b render the same String().
func releaseEqual(a, b forgetypes.Release) bool {
	return a.String() == b.String()
}

func testLookup(t *testing.T) *Lookup {
	t.Helper()

	lvCfg := lvbasetest.PlainDefaultsConfig(t)

	options := opttest.PlainOptions(t, opttest.PlainDefaultsConfig(t))
	svcStatus, _ := statustest.New("yaml", nil)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "github-testLookup",
		},
		&dashboard.Options{
			WebURL: "https://example.com",
		},
	)

	lookup, err := Decode(
		"yaml", []byte(test.TrimYAML(`
			url: `+test.ArgusGitHubRepo+`
			url_commands:
				- type: regex
					regex: '[0-9.]+'
		`)),
		options,
		svcStatus,
		lvCfg,
	)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	typeHardDefaults := &Defaults{}
	typeHardDefaults.Default()
	lookup.SetTypeDefaults(&Defaults{}, typeHardDefaults)

	return lookup
}

// githubEndpoint is one endpoint's reply, shared with the other forges.
type githubEndpoint = forgetest.Endpoint

// githubServer is a fixture standing in for the GitHub API.
type githubServer = forgetest.Server

// newGitHubServer starts a fixture serving endpoints.
func newGitHubServer(t *testing.T, endpoints map[string]githubEndpoint) *githubServer {
	t.Helper()

	return startGitHubServer(t, &githubServer{Endpoints: endpoints})
}

// startGitHubServer starts server, giving it the replies the GitHub API gives.
func startGitHubServer(t *testing.T, server *githubServer) *githubServer {
	t.Helper()

	server.UnauthorizedBody = `{"message":"Bad credentials","documentation_url":"https://docs.github.com/rest"}`
	server.NotFoundBody = `{"message":"Not Found","documentation_url":"https://docs.github.com/rest"}`

	return forgetest.Start(t, server, httptest.NewServer)
}
