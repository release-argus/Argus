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

package config

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func testVerify(t *testing.T) *Config {
	cfg := &Config{}
	cfg.Order = []string{"test"}

	cfg.Defaults = Defaults{}
	cfg.Defaults.Default()
	cfg.HardDefaults = Defaults{}

	cfg.Notify = shoutrrr.ShoutrrrsDefaults{
		"test": shoutrrr.NewDefaults(
			"discord",
			cfg.Defaults.Notify["discord"].Options,
			cfg.Defaults.Notify["discord"].URLFields,
			cfg.Defaults.Notify["discord"].Params)}

	cfg.WebHook = webhook.WebHooksDefaults{
		"test": &cfg.Defaults.WebHook}

	serviceID := "test"
	cfg.Service = service.Services{
		serviceID: &service.Service{
			ID: serviceID,
			LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"github",
					"yaml", test.TrimYAML(`
							url: release-argus/argus
					`),
					nil,
					nil,
					&cfg.Defaults.Service.LatestVersion, &cfg.HardDefaults.Service.LatestVersion)
			})}}

	return cfg
}

func TestConfig_CheckValues(t *testing.T) {
	// GIVEN variations of Config to test.
	tests := map[string]struct {
		config   *Config
		errRegex string
		noPanic  bool
	}{
		"valid Config": {
			config:  testVerify(t),
			noPanic: true,
		},
		"invalid Defaults": {
			config: &Config{
				Defaults: Defaults{
					Service: service.Defaults{
						Options: *opt.NewDefaults("1x", nil)}}},
			errRegex: test.TrimYAML(`
				^defaults:
					service:
						options:
							interval: "[^"]+" <invalid>`),
		},
		"invalid Notify": {
			config: &Config{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"test": shoutrrr.NewDefaults(
						"discord",
						map[string]string{
							"delay": "2x"},
						nil, nil)}},
			errRegex: test.TrimYAML(`
				^notify:
					test:
						options:
							delay: "[^"]+" <invalid>`),
		},
		"invalid WebHook": {
			config: &Config{
				WebHook: webhook.WebHooksDefaults{
					"test": webhook.NewDefaults(
						nil, nil, "3x", nil, nil, "", nil, "", "")}},
			errRegex: test.TrimYAML(`
				^webhook:
					test:
						delay: "3x" <invalid>`),
		},
		"invalid Service": {
			config: &Config{
				Service: service.Services{
					"test": &service.Service{
						Options: *opt.New(
							nil, "4x", nil,
							nil, nil)}}},
			errRegex: test.TrimYAML(`
				^service:
					test:
						options:
							interval: "4x" <invalid>.*`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			if tc.config != nil {
				for name, svc := range tc.config.Service {
					svc.ID = name
				}
			}
			// Switch Fatal to panic and disable this panic.
			if !tc.noPanic {
				defer func() {
					recover()
					stdout := releaseStdout()

					lines := strings.Split(stdout, "\n")
					wantLines := strings.Count(tc.errRegex, "\n")
					if wantLines > len(lines) {
						t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
							packageName, wantLines, tc.errRegex, len(lines), lines, stdout)
					}
					if !util.RegexCheck(tc.errRegex, stdout) {
						t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
							packageName, tc.errRegex, stdout)
						return
					}
				}()
			}

			// WHEN CheckValues is called on them.
			tc.config.CheckValues()

			// THEN this call will/wont crash the program.
			stdout := releaseStdout()
			lines := strings.Split(stdout, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, stdout)
				return
			}
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, stdout)
				return
			}
		})
	}
}

func TestConfig_Print(t *testing.T) {
	// GIVEN a Config and print flags of true and false.
	config := testVerify(t)
	tests := map[string]struct {
		flag  bool
		lines int
	}{
		"flag on":  {flag: true, lines: 198 + len(config.Defaults.Notify)},
		"flag off": {flag: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			// WHEN Print is called with these flags.
			config.Print(&tc.flag)

			// THEN config is printed only when the flag is true.
			stdout := releaseStdout()
			got := strings.Count(stdout, "\n")
			if got != tc.lines {
				t.Errorf("%s\nPrint with %s mismatch\nwant: %d lines\ngot:  %d\nstdout:\n%s\n\nstdout: %q",
					packageName, name, tc.lines, got, stdout, stdout)
			}
		})
	}
}
