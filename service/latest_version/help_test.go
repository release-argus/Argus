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

package latestver

import (
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	github "github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
)

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Exit(exitCode)
}

func testLookup(lookupType string, failing bool) Lookup {
	lookup, _ := New(
		lookupType,
		"yaml", "",
		nil,
		nil,
		nil, nil)

	// HardDefaults.
	hardDefaults := &base.Defaults{}
	hardDefaults.Default()
	// Defaults.
	defaults := &base.Defaults{}
	// Options.
	hardDefaultOptions := &opt.Defaults{}
	hardDefaultOptions.Default()
	options := opt.New(
		nil, "5m", test.BoolPtr(false),
		&opt.Defaults{}, hardDefaultOptions)
	// Status.
	announceChannel := make(chan []byte, 24)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	svcStatus := status.New(
		&announceChannel, &databaseChannel, &saveChannel,
		"", "", "", "", "", "")
	svcStatus.Init(
		0, 0, 0,
		test.StringPtr("serviceID"), nil,
		test.StringPtr("http://example.com"),
	)

	switch l := lookup.(type) {
	case *github.Lookup:
		l.URL = "release-argus/Argus"
		l.URLCommands = filter.URLCommandSlice{
			{Type: "regex", Regex: `([0-9.]+)`}}
		l.AccessToken = os.Getenv("GITHUB_TOKEN")
		if failing {
			l.AccessToken = "invalid"
		}
		l.UsePreRelease = test.BoolPtr(false)
		l.Init(
			options,
			svcStatus,
			defaults, hardDefaults)

	case *web.Lookup:
		l.URL = test.LookupPlain["url_invalid"]
		l.AllowInvalidCerts = test.BoolPtr(true)
		if failing {
			*l.AllowInvalidCerts = false
		}
		l.URLCommands = filter.URLCommandSlice{
			{Type: "regex", Regex: `ver([0-9.]+)`}}
		l.Init(
			options,
			svcStatus,
			defaults, hardDefaults)
	}

	return lookup
}

func getType(lookup Lookup) string {
	switch lookup.(type) {
	case *github.Lookup:
		return "github"
	case *web.Lookup:
		return "url"
	}
	return "unknown"
}
