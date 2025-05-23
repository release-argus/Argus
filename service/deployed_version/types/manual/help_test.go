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

package manual

import (
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
)

var packageName = "deployedver_manual"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Exit(exitCode)
}

func testLookup(version string, failing bool) *Lookup {
	// HardDefaults.
	hardDefaults := &base.Defaults{}
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
	status := status.New(
		&announceChannel, &databaseChannel, &saveChannel,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{})
	status.Init(
		0, 0, 0,
		"serviceID", "", "",
		&dashboard.Options{
			WebURL: "https://example.com"})

	// Lookup.
	lookup, _ := New(
		"yaml", test.TrimYAML(`
			version: `+version+`
		`),
		options,
		status,
		defaults, hardDefaults)

	return lookup
}
