// Copyright [2023] [Argus]
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

package latestver

import (
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

// Unsure why Go tests give a different result than the compiled binary
var initialEmptyListETag string

func TestMain(m *testing.M) {
	// initialize jLog
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true
	LogInit(jLog)
	FindEmptyListETag(os.Getenv("GITHUB_TOKEN"))
	initialEmptyListETag = getEmptyListETag()

	// run other tests
	exitCode := m.Run()

	// exit
	os.Exit(exitCode)
}

func testLookup(urlType bool, allowInvalidCerts bool) *Lookup {
	announceChannel := make(chan []byte, 24)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	svcStatus := svcstatus.New(
		&announceChannel, &databaseChannel, &saveChannel,
		"", "", "", "", "", "")
	lookup := New(
		test.StringPtr(os.Getenv("GITHUB_TOKEN")),
		test.BoolPtr(allowInvalidCerts),
		nil,
		opt.New(
			nil, "", test.BoolPtr(true),
			&opt.OptionsDefaults{},
			opt.NewDefaults(
				"0s", test.BoolPtr(true))),
		&filter.Require{}, nil,
		"github",
		"release-argus/Argus",
		nil,
		nil,
		&LookupDefaults{},
		&LookupDefaults{})
	lookup.Status = svcStatus
	if urlType {
		lookup.Type = "url"
		lookup.URL = "https://invalid.release-argus.io/plain"
		lookup.URLCommands = filter.URLCommandSlice{
			{Type: "regex", Regex: test.StringPtr("ver([0-9.]+)")}}
	} else {
		lookup.GitHubData = NewGitHubData("", nil)
		lookup.URLCommands = filter.URLCommandSlice{
			{Type: "regex", Regex: test.StringPtr("([0-9.]+)")}}
		lookup.AccessToken = test.StringPtr(os.Getenv("GITHUB_TOKEN"))
		lookup.UsePreRelease = test.BoolPtr(false)
	}
	lookup.Status.Init(
		0, 0, 0,
		test.StringPtr("serviceID"),
		test.StringPtr("http://example.com"),
	)
	lookup.Require.Status = lookup.Status
	lookup.Defaults = &LookupDefaults{}
	lookup.HardDefaults = &LookupDefaults{}
	return lookup
}
