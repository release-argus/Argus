// Copyright [2025] [Argus]
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
	"encoding/json"
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
)

var packageName = "latestver_github"
var initialEmptyListETag string
var testBody = []byte(test.TrimJSON(`
[
	{"tag_name":"0.18.0","name":"0.18.0","prerelease":true,"published_at":"2024-05-07T13:10:29Z",
		"assets":[
			{"id": 9,"name":"Argus-0.18.0.linux-amd64","created_at":"2024-05-07T13:11:30Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.18.0/Argus-0.18.0.darwin-amd64"},
			{"id": 5,"name":"Argus-0.18.0.linux-arm64","created_at":"2024-05-07T13:11:39Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.18.0/Argus-0.18.0.darwin-arm64"}]},
	{"tag_name":"0.17.4","name":"0.17.4","prerelease":false,"published_at":"2024-04-27T10:50:00Z",
		"assets":[
			{"id": 3,"name":"Argus-0.17.4.linux-amd64","created_at":"2024-04-27T10:50:53Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.17.4/Argus-0.17.4.linux-amd64"},
			{"id": 7,"name":"Argus-0.17.4.linux-arm","created_at":"2024-04-27T10:50:59Z","browser_download_url":"https://github.com/release-argus/Argus/releases/download/0.17.4/Argus-0.17.4.linux-arm"}]}]
`))
var testBodyObject []github_types.Release

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	SetEmptyListETag(os.Getenv("GITHUB_TOKEN"))
	initialEmptyListETag = getEmptyListETag()

	// Unmarshal testBody.
	json.Unmarshal(testBody, &testBodyObject)

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Exit(exitCode)
}

// newData returns a new Data.
func newData(
	eTag string,
	releases *[]github_types.Release,
) *Data {
	// ETag - https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests.
	if eTag == "" {
		eTag = getEmptyListETag()
	}
	// Releases.
	var releasesDeref []github_types.Release
	if releases != nil {
		releasesDeref = *releases
	}

	return &Data{
		eTag:     eTag,
		releases: releasesDeref}
}

func testLookup(failing bool) *Lookup {
	// HardDefaults.
	hardDefaults := &base.Defaults{}
	hardDefaults.AccessToken = os.Getenv("GITHUB_TOKEN")
	hardDefaults.Default()
	// Defaults.
	defaults := &base.Defaults{}
	// Options.
	hardDefaultOptions := &opt.Defaults{}
	hardDefaultOptions.Default()
	options := opt.New(
		nil, "", test.BoolPtr(true),
		&opt.Defaults{}, hardDefaultOptions)
	// Status.
	announceChannel := make(chan []byte, 24)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	svcStatus := status.New(
		&announceChannel, &databaseChannel, &saveChannel,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{})
	svcStatus.Init(
		0, 0, 0,
		"serviceID", "", "",
		&dashboard.Options{
			WebURL: "https://example.com"})

	lookup, _ := New(
		"yaml", test.TrimYAML(`
				url: release-argus/Argus
				url_commands:
					- type: regex
						regex: '[0-9.]+'
			`),
		options,
		svcStatus,
		defaults, hardDefaults)
	if failing {
		lookup.AccessToken = "invalid"
	}

	lookup.Init(
		options,
		svcStatus,
		defaults, hardDefaults)

	return lookup
}
