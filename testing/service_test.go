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

//go:build unit

// Package testing provides utilities for CLI-based testing.
package testing

import (
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

func TestServiceTest(t *testing.T) {
	// GIVEN a Config with a Service.
	tests := map[string]struct {
		flag        string
		services    service.Services
		stdoutRegex *string
		ok          bool
	}{
		"flag is empty": {
			flag:        "",
			ok:          true,
			stdoutRegex: test.StringPtr("^$"),
			services: service.Services{
				"argus": {
					ID: "argus",
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil),
				}},
		},
		"unknown service": {
			flag:        "test",
			ok:          false,
			stdoutRegex: test.StringPtr(`Service "test" could not be found in config.service\sDid you mean one of these\?\s  - argus`),
			services: service.Services{
				"argus": {
					ID: "argus",
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil),
				}},
		},
		"github service": {
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.StringPtr(`argus\)?, Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			services: service.Services{
				"argus": {
					ID: "argus",
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
								url: release-argus/Argus
								access_token: `+os.Getenv("GITHUB_TOKEN")+`
								url_commands:
									- type: regex
										regex: '[0-9.]+$'
							`),
							opt.New(
								nil, "0s", nil,
								nil, nil),
							status.New(
								nil, nil, nil,
								"",
								"", "",
								"", "",
								"",
								&dashboard.Options{}),
							nil, nil)
					})}},
		},
		"url service type but github 'owner/repo' url": {
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.StringPtr("unsupported protocol scheme"),
			services: service.Services{
				"argus": {
					ID: "argus",
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"url",
							"yaml", test.TrimYAML(`
								url: release-argus/Argus
								url_commands:
									- type: regex
										regex: '[0-9.]+$'
							`),
							opt.New(
								nil, "0s", nil,
								nil, nil),
							status.New(
								nil, nil, nil,
								"",
								"", "",
								"", "",
								"",
								&dashboard.Options{}),
							nil, nil)
					})}},
		},
		"url service": {
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.StringPtr(`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			services: service.Services{
				"argus": {
					ID: "argus",
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"url",
							"yaml", test.TrimYAML(`
								url: https://github.com/release-argus/Argus/releases
								url_commands:
									- type: regex
										regex: 'tag/([0-9.]+)"'
							`),
							opt.New(
								nil, "0s", nil,
								nil, nil),
							status.New(
								nil, nil, nil,
								"",
								"", "",
								"", "",
								"",
								&dashboard.Options{}),
							nil, nil)
					})}},
		},
		"service with deployed version lookup": {
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.StringPtr(`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"\s.*Updated to.*\s.*Deployed version - "[0-9]+\.[0-9]+\.[0-9]+"`),
			services: service.Services{
				"argus": {
					ID: "argus",
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"url",
							"yaml", test.TrimYAML(`
								url: https://github.com/release-argus/Argus/releases
								url_commands:
									- type: regex
										regex: tag/([0-9.]+)"
							`),
							nil,
							nil,
							nil, nil)
					}),
					DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
						return deployedver.New(
							"url",
							"yaml", test.TrimYAML(`
								method: GET
								url: `+test.LookupJSON["url_valid"]+`
								json: foo.bar.version
							`),
							opt.New(
								nil, "0s", nil,
								nil, nil),
							nil,
							nil, nil)
					}),
					Options: *opt.New(
						nil, "0s", test.BoolPtr(true),
						nil, nil)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			if tc.services[tc.flag] != nil {
				hardDefaults := config.Defaults{}
				hardDefaults.Default()
				tc.services[tc.flag].ID = tc.flag
				tc.services[tc.flag].Init(
					&service.Defaults{}, &hardDefaults.Service,
					&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &hardDefaults.Notify,
					&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &hardDefaults.WebHook)
				// will do a call for latest_version* and one for deployed_version*.
				dbChannel := make(chan dbtype.Message, 4)
				tc.services[tc.flag].Status.DatabaseChannel = dbChannel
			}

			resultChannel := make(chan bool, 1)
			// WHEN ServiceTest is called with the test Config.
			order := []string{}
			for i := range tc.services {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.services,
				Order:   order,
			}
			resultChannel <- ServiceTest(&tc.flag, &cfg)
			// WHEN the db is initialised with it.

			// THEN we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %s",
						packageName, *tc.stdoutRegex, stdout)
				}
			}
			// AND it succeeds/fails as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
		})
	}
	time.Sleep(100 * time.Millisecond)
}
