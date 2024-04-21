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

//go:build unit

package testing

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/webhook"
)

func TestServiceTest(t *testing.T) {
	// GIVEN a Config with a Service
	tests := map[string]struct {
		flag        string
		slice       service.Slice
		stdoutRegex *string
		panicRegex  *string
	}{
		"flag is empty": {
			flag:        "",
			stdoutRegex: stringPtr("^$"),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil),
				},
			},
		},
		"unknown service": {
			flag:       "test",
			panicRegex: stringPtr(`Service "test" could not be found in config.service\sDid you mean one of these\?\s  - argus`),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil),
				},
			},
		},
		"github service": {
			flag:        "argus",
			stdoutRegex: stringPtr(`argus\)?, Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			slice: service.Slice{
				"argus": {
					LatestVersion: *latestver.New(
						stringPtr(os.Getenv("GITHUB_TOKEN")),
						boolPtr(false),
						nil,
						opt.New(
							nil, "0s", nil,
							nil, nil),
						nil, nil,
						"github",
						"release-argus/Argus",
						&filter.URLCommandSlice{
							{Type: "regex", Regex: stringPtr("[0-9.]+$")}},
						nil, nil, nil)},
			},
		},
		"url service type but github owner/repo url": {
			flag:        "argus",
			stdoutRegex: stringPtr("This URL looks to be a GitHub repo, but the service's type is url, not github"),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					LatestVersion: *latestver.New(
						nil,
						boolPtr(false),
						nil,
						opt.New(
							nil, "0s", nil,
							nil, nil),
						nil, nil,
						"url",
						"release-argus/Argus",
						&filter.URLCommandSlice{
							{Type: "regex", Regex: stringPtr("[0-9.]+$")}},
						nil, nil, nil)}},
		},
		"url service": {
			flag:        "argus",
			stdoutRegex: stringPtr(`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					LatestVersion: *latestver.New(
						nil,
						boolPtr(false),
						nil,
						opt.New(
							nil, "0s", nil,
							nil, nil),
						nil, nil,
						"url",
						"https://github.com/release-argus/Argus/releases",
						&filter.URLCommandSlice{
							{Type: "regex", Regex: stringPtr(`tag/([0-9.]+)"`)}},
						nil, nil, nil)}},
		},
		"service with deployed version lookup": {
			flag:        "argus",
			stdoutRegex: stringPtr(`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"\s.*Deployed version - "[0-9]+\.[0-9]+\.[0-9]+"`),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					LatestVersion: *latestver.New(
						nil,
						boolPtr(false),
						nil, nil, nil, nil,
						"url",
						"https://github.com/release-argus/Argus/releases",
						&filter.URLCommandSlice{
							{Type: "regex", Regex: stringPtr(`tag/([0-9.]+)"`)}},
						nil, nil, nil),
					DeployedVersionLookup: deployedver.New(
						boolPtr(true),
						nil, nil,
						"version",
						nil, "", nil,
						&svcstatus.Status{},
						"https://release-argus.io/demo/api/v1/version",
						nil, nil),
					Options: *opt.New(
						nil, "0s", boolPtr(true),
						nil, nil)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					releaseStdout()

					// Check the panic message
					rStr := fmt.Sprint(r)
					re := regexp.MustCompile(*tc.panicRegex)
					match := re.MatchString(rStr)
					if !match {
						t.Errorf("expected a panic that matched %q\ngot: %q",
							*tc.panicRegex, rStr)
					}
				}()
			}
			if tc.slice[tc.flag] != nil {
				defaults := config.Defaults{}
				defaults.SetDefaults()
				tc.slice[tc.flag].ID = tc.flag
				tc.slice[tc.flag].Init(
					&service.Defaults{}, &defaults.Service,
					&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &defaults.Notify,
					&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &defaults.WebHook)
				// will do a call for latest_version* and one for deployed_version*
				dbChannel := make(chan dbtype.Message, 4)
				tc.slice[tc.flag].Status.DatabaseChannel = &dbChannel
				if tc.slice[tc.flag].DeployedVersionLookup != nil {
					tc.slice[tc.flag].DeployedVersionLookup.Defaults = &deployedver.LookupDefaults{}
					tc.slice[tc.flag].DeployedVersionLookup.HardDefaults = &deployedver.LookupDefaults{}
				}
			}

			// WHEN ServiceTest is called with the test Config
			order := []string{}
			for i := range tc.slice {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.slice,
				Order:   order,
			}
			ServiceTest(&tc.flag, &cfg, jLog)

			// THEN we get the expected stdout
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				re := regexp.MustCompile(*tc.stdoutRegex)
				match := re.MatchString(stdout)
				if !match {
					t.Errorf("want match on %q\ngot:\n%s",
						*tc.stdoutRegex, stdout)
				}
			}
		})
	}
	time.Sleep(100 * time.Millisecond)
}
