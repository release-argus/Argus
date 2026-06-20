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
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var packageName = "latestver_github"
var initialEmptyListETag string
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
var testBodyObject []ghtypes.Release

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	SetEmptyListETag(test.GitHubToken(nil))
	initialEmptyListETag = getEmptyListETag()

	// Unmarshal testBody.
	_ = decode.Unmarshal("json", testBody, &testBodyObject)

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// newData returns a new Data.
func newData(
	eTag string,
	releases *[]ghtypes.Release,
) *Data {
	// ETag - https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests.
	if eTag == "" {
		eTag = getEmptyListETag()
	}
	// Releases.
	var releasesDeref []ghtypes.Release
	if releases != nil {
		releasesDeref = *releases
	}

	return &Data{
		eTag:     eTag,
		releases: releasesDeref,
	}
}

func testLookup(t *testing.T, failing bool) *Lookup {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// Options.
	options, _ := opt.Decode(
		"yaml", nil,
		optCfg,
	)
	// Status.
	svcStatus, _ := statustest.New("yaml", nil)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "github-testLookup",
		},
		&dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				WebURL: "https://example.com",
			},
		},
	)

	lookup, _ := Decode(
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
	if failing {
		lookup.AccessToken = "invalid"
	}

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
	hardDefaults.AccessToken = test.GitHubToken(t)
	hardDefaults.Options = optHardDefaults

	defaults.Require.SetDefaults(&hardDefaults.Require)

	return base.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
