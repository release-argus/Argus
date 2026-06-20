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

//go:build integration

// Package testing provides utilities for CLI-based testing.
package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestServiceTest(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: a Config with a Service.
	tests := []struct {
		name        string
		flag        string
		services    service.Services
		stdoutRegex *string
		ok          bool
	}{
		{
			name:        "empty flag",
			flag:        "",
			ok:          true,
			stdoutRegex: test.Ptr("^$"),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							latest_version:
						`+lvtest.Lookup(t, "url", false).String("    ")+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "unknown service",
			flag:        "test",
			ok:          false,
			stdoutRegex: test.Ptr(`Service "test" could not be found in config.service\sDid you mean one of these\?\s  - argus`),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							latest_version:
						`+lvtest.Lookup(t, "url", false).String("    ")+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "github service",
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.Ptr(`argus\)?, Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							latest_version:
						`+lvtest.Lookup(t, "github", false).String("    ")+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "url service type but github owner-repo url",
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.Ptr("unsupported protocol scheme"),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							latest_version:
								type: url
								url: `+test.ArgusGitHubRepo+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "url service",
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.Ptr(`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"`),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							latest_version:
						`+lvtest.Lookup(t, "url", false).String("    ")+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "service with deployed version lookup",
			flag: "argus",
			ok:   true,
			stdoutRegex: test.Ptr(
				`Latest Release - "[0-9]+\.[0-9]+\.[0-9]+"\s` +
					`.*Updated to.*\s` +
					`.*Deployed version - "[0-9]+\.[0-9]+\.[0-9]+"`,
			),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							latest_version:
						`+lvtest.Lookup(t, "github", false).String("    ")+`
							deployed_version:
						`+dvtest.Lookup(t, "url", false, "").String("    ")+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			if tc.services[tc.flag] != nil {
				hardDefaults := config.Defaults{}
				hardDefaults.Default()
				// will do a call for latest_version* and one for deployed_version*.
				dbChannel := make(chan dbtype.Message, 4)
				tc.services[tc.flag].Status.DatabaseChannel = dbChannel
			}

			resultChannel := make(chan bool, 1)
			// WHEN: ServiceTest is called with the test Config.
			var order []string
			for i := range tc.services {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.services,
				Order:   order,
			}
			resultChannel <- ServiceTest(&tc.flag, &cfg)
			// WHEN: the DB is initialised with it.

			prefix := fmt.Sprintf(
				"%s\nServiceTest(%q)",
				packageName, tc.flag,
			)

			// THEN: it succeeds/fails as expected.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				nil,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			// THEN: we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf(
						"%s\nerror mismatch\ngot:  %q\nwant: %s",
						prefix, stdout, *tc.stdoutRegex,
					)
				}
			}
		})
	}
	time.Sleep(100 * time.Millisecond)
}
