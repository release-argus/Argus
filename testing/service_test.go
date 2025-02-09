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

package testing

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestServiceTest(t *testing.T) {
	// GIVEN a Config with a Service.
	tests := map[string]struct {
		flag                    string
		slice                   service.Slice
		stdoutRegex, panicRegex *string
	}{
		"flag is empty": {
			flag:        "",
			stdoutRegex: test.StringPtr("^$"),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil),
				}},
		},
		"unknown service": {
			flag:       "test",
			panicRegex: test.StringPtr(`Service "test" could not be found in config.service\sDid you mean one of these\?\s  - argus`),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil),
				}},
		},
		"github service": {
			flag:        "argus",
			stdoutRegex: test.StringPtr(`argus\)?, Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			slice: service.Slice{
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
								"", "", "", "", "", ""),
							nil, nil)
					})}},
		},
		"url service type but github 'owner/repo' url": {
			flag:        "argus",
			stdoutRegex: test.StringPtr("unsupported protocol scheme"),
			slice: service.Slice{
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
								"", "", "", "", "", ""),
							nil, nil)
					})}},
		},
		"url service": {
			flag:        "argus",
			stdoutRegex: test.StringPtr(`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			slice: service.Slice{
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
								"", "", "", "", "", ""),
							nil, nil)
					})}},
		},
		"service with deployed version lookup": {
			flag:        "argus",
			stdoutRegex: test.StringPtr(`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"\s.*Updated to.*\s.*Deployed version - "[0-9]+\.[0-9]+\.[0-9]+"`),
			slice: service.Slice{
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
			releaseStdout := test.CaptureStdout()

			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					releaseStdout()

					// Check the panic message.
					rStr := fmt.Sprint(r)
					if !util.RegexCheck(*tc.panicRegex, rStr) {
						t.Errorf("expected a panic that matched %q\ngot: %q",
							*tc.panicRegex, rStr)
					}
				}()
			}
			if tc.slice[tc.flag] != nil {
				hardDefaults := config.Defaults{}
				hardDefaults.Default()
				tc.slice[tc.flag].ID = tc.flag
				tc.slice[tc.flag].Init(
					&service.Defaults{}, &hardDefaults.Service,
					&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &hardDefaults.Notify,
					&webhook.SliceDefaults{}, &webhook.Defaults{}, &hardDefaults.WebHook)
				// will do a call for latest_version* and one for deployed_version*.
				dbChannel := make(chan dbtype.Message, 4)
				tc.slice[tc.flag].Status.DatabaseChannel = &dbChannel
			}

			// WHEN ServiceTest is called with the test Config.
			order := []string{}
			for i := range tc.slice {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.slice,
				Order:   order,
			}
			ServiceTest(&tc.flag, &cfg)

			// THEN we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf("want match on %q\ngot:\n%s",
						*tc.stdoutRegex, stdout)
				}
			}
		})
	}
	time.Sleep(100 * time.Millisecond)
}
