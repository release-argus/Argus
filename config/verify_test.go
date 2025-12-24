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
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

func testVerify(t *testing.T) *Config {
	cfg := &Config{}
	cfg.SaveChannel = make(chan bool, 10)

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
		logRegex *string
		ok       bool
		wantSave bool
	}{
		"valid Config": {
			config: testVerify(t),
			ok:     true,
		},
		"invalid Settings": {
			config: &Config{Settings: Settings{
				SettingsBase: SettingsBase{
					Web: WebSettings{
						CertFile: "doesnotexist.pem"}}}},
			errRegex: test.TrimYAML(`
				^settings:
					web:
						cert_file: .*doesnotexist.pem.* no such file.*`),
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
		"valid Config that gets changed is saved": {
			config: test.IgnoreError(t, func() (*Config, error) {
				cfg := testVerify(t)

				newService := &service.Service{
					ID:      "test",
					Comment: "foo_comment",
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
							nil,
							nil,
							nil, nil)
					}),
					WebHook: map[string]*webhook.WebHook{
						"wh": {
							Base: webhook.Base{
								Type:   "github",
								URL:    "example.com",
								Secret: "Argus",
								CustomHeaders: &webhook.Headers{
									webhook.Header{
										Key: "foo", Value: "bar"}},
							}}},
				}
				newService.Init(
					&service.Defaults{}, &service.Defaults{},
					&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{},
					&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
				cfg.Service[t.Name()] = newService

				return cfg, nil
			}),
			logRegex: test.StringPtr(`^DEPRECATED: .*\s$`),
			ok:       true,
			wantSave: true,
		},
		"invalid Config that gets changed is not saved": {
			config: test.IgnoreError(t, func() (*Config, error) {
				cfg := testVerify(t)

				newService := &service.Service{
					ID:      "test",
					Comment: "foo_comment",
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
							nil,
							nil,
							nil, nil)
					}),
					WebHook: map[string]*webhook.WebHook{
						"wh": {
							Base: webhook.Base{
								Type:   "github",
								URL:    "example.com",
								Secret: "Argus",
								CustomHeaders: &webhook.Headers{
									webhook.Header{
										Key: "foo", Value: "bar"}},
							}}},
				}
				newService.Init(
					&service.Defaults{}, &service.Defaults{},
					&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{},
					&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
				cfg.Service[t.Name()] = newService

				badService := &service.Service{
					ID:      "bad",
					Comment: "foo_comment",
					Options: *opt.New(
						nil, "10x", nil, nil, nil),
					LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
						return latestver.New(
							"github",
							"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
							nil,
							nil,
							nil, nil)
					}),
					Notify: shoutrrr.Shoutrrrs{
						"foo": shoutrrr.New(
							nil,
							"", "generic",
							nil,
							map[string]string{
								"host":           "x",
								"secret":         "y",
								"custom_headers": `{"foo": "bar"}`},
							nil,
							nil, nil, nil)},
					WebHook: map[string]*webhook.WebHook{
						"wh": {
							Base: webhook.Base{
								Type:   "github",
								URL:    "example.com",
								Secret: "Argus",
								CustomHeaders: &webhook.Headers{
									webhook.Header{
										Key: "foo", Value: "bar"}},
							}}},
				}
				badService.Init(
					&service.Defaults{}, &service.Defaults{},
					&shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{},
					&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
				cfg.Service["badService"] = badService

				return cfg, nil
			}),
			errRegex: test.TrimYAML(`
				^service:
					badService:
						options:
							interval: "10x" <invalid>.*\s$`),
			logRegex: test.StringPtr(
				test.TrimYAML(`
				^DEPRECATED: .*webhook.custom_headers.*
				DEPRECATED: .*notify.generic.url_fields.custom_headers.*
				DEPRECATED: .*webhook.custom_headers.*
				FATAL: Config could not be parsed.*\s$`)),
			ok:       false,
			wantSave: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseLog := test.CaptureLog(logutil.Log)
			releaseStdout := test.CaptureStdout()

			if tc.config != nil {
				for name, svc := range tc.config.Service {
					svc.ID = name
				}
			}

			resultChannel := make(chan bool, 1)
			// WHEN CheckValues is called on them.
			resultChannel <- tc.config.CheckValues()

			// THEN this call will/won't crash the program.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
			// AND the error line count matches.
			stdout := releaseStdout()
			lines := strings.Split(stdout, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, stdout)
				return
			}
			// AND the error regex matches.
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, stdout)
				return
			}
			// AND the log matches.
			logWant := `^FATAL: Config could not be parsed.*\s$`
			if tc.ok {
				logWant = `^$`
			}
			logWant = util.DereferenceOrValue(tc.logRegex, logWant)
			logOut := releaseLog()
			if !util.RegexCheck(logWant, logOut) {
				t.Errorf("%s\nlog mismatch\nwant: %q\ngot:  %q",
					packageName, logWant, logOut)
			}
			// AND saves are queued as expected.
			saveQueued := len(tc.config.SaveChannel) > 0
			if saveQueued != tc.wantSave {
				t.Errorf("%s\nsave queue mismatch\nwant: %t\ngot:  %t",
					packageName, tc.wantSave, saveQueued)
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
