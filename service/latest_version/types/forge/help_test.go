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

// Package forge provides the release handling shared by git forge lookup types.
package forge

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var packageName = "latestver_forge"

// testBody is a releases response in the shape a supported forge returns.
var testBody = []byte(
	test.TrimJSON(`[
		{
			"tag_name":"0.18.0",
			"name":"0.18.0",
			"prerelease":true,
			"published_at":"2000-01-02T03:04:05Z",
			"assets":[
				{"id": 9,"name":"Argus-0.18.0.linux-amd64","created_at":"2001-01-02T03:04:05Z","browser_download_url":"https://forge.example.com/owner/repo/releases/download/0.18.0/Argus-0.18.0.linux-amd64"},
				{"id": 5,"name":"Argus-0.18.0.linux-arm64","created_at":"2002-01-02T03:04:05Z","browser_download_url":"https://forge.example.com/owner/repo/releases/download/0.18.0/Argus-0.18.0.linux-arm64"}
			]
		},
		{
			"tag_name":"0.17.4",
			"name":"0.17.4",
			"prerelease":false,
			"published_at":"1998-07-06T05:04:03Z",
			"assets":[
				{"id": 3,"name":"Argus-0.17.4.linux-amd64","created_at":"1999-08-07T06:05:04Z","browser_download_url":"https://forge.example.com/owner/repo/releases/download/0.17.4/Argus-0.17.4.linux-amd64"}
			]
		}
	]`),
)

var testBodyObject []forgetypes.Release

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

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

// testRequire returns the Require decoded from requireYAML, or nil when it is empty.
func testRequire(t *testing.T, requireYAML string) *filter.Require {
	t.Helper()

	if requireYAML == "" {
		return nil
	}

	svcStatus, _ := statustest.New("yaml", nil)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "forge-testRequire",
		},
		&dashboard.Options{},
	)

	defaults, _ := base.DecodeDefaults("yaml", nil)
	hardDefaults, _ := base.DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	defaults.Require.SetDefaults(&hardDefaults.Require)

	require, err := filter.Decode(
		"yaml", []byte(requireYAML),
		svcStatus,
		&defaults.Require,
	)
	if err != nil {
		t.Fatalf(
			"%s\nfailed to decode Require: %v",
			packageName, err,
		)
	}
	require.Init(svcStatus, &defaults.Require)

	return require
}
