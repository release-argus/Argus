// Copyright [2022] [Argus]
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

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func testLogging(level string) {
	jLog = util.NewJLog(level, false)
	LogInit(jLog)
}

func testLookup(urlType bool, allowInvalidCerts bool) *Lookup {
	announceChannel := make(chan []byte, 24)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	lookup := &Lookup{
		Type:              "github",
		URL:               "release-argus/Argus",
		AllowInvalidCerts: boolPtr(allowInvalidCerts),
		Require:           &filter.Require{},
		Options: &opt.Options{
			SemanticVersioning: boolPtr(true),
			Defaults:           &opt.Options{},
			HardDefaults:       &opt.Options{},
		},
		Status: &svcstatus.Status{
			ServiceID:       stringPtr("test"),
			AnnounceChannel: &announceChannel,
			DatabaseChannel: &databaseChannel,
			SaveChannel:     &saveChannel,
		},
		Defaults:     &Lookup{},
		HardDefaults: &Lookup{},
	}
	if urlType {
		lookup.Type = "url"
		lookup.URL = "https://invalid.release-argus.io/plain"
		lookup.URLCommands = filter.URLCommandSlice{
			{Type: "regex", Regex: stringPtr("v([0-9.]+)")}}
	} else {
		lookup.GitHubData = &GitHubData{}
		lookup.URLCommands = filter.URLCommandSlice{
			{Type: "regex", Regex: stringPtr("([0-9.]+)")}}
		lookup.AccessToken = stringPtr(os.Getenv("GITHUB_TOKEN"))
		lookup.UsePreRelease = boolPtr(false)
	}
	lookup.Status.Init(
		0, 0, 0,
		stringPtr("serviceID"),
		stringPtr("http://example.com"),
	)
	lookup.Require.Status = lookup.Status
	return lookup
}
