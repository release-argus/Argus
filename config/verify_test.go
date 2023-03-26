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

//go:build unit

package config

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func testVerify() *Config {
	order := []string{"test"}
	defaults := Defaults{}
	defaults.SetDefaults()
	notify := shoutrrr.Slice{
		"test": defaults.Notify["discord"],
	}
	webhook := webhook.Slice{
		"test": &defaults.WebHook,
	}
	serviceID := "test"
	service := service.Slice{
		serviceID: &service.Service{
			ID: serviceID,
			LatestVersion: latestver.Lookup{
				Type: "github",
				URL:  "release-argus/argus",
			},
		},
	}
	return &Config{
		Order:    order,
		Defaults: defaults,
		Notify:   notify,
		WebHook:  webhook,
		Service:  service,
	}
}

func TestConfig_CheckValues(t *testing.T) {
	// GIVEN variations of Config to test
	jLog = util.NewJLog("WARN", false)
	tests := map[string]struct {
		config      *Config
		errContains string
		noPanic     bool
	}{
		"valid Config": {
			config:      testVerify(),
			errContains: "",
			noPanic:     true,
		},
		"invalid Defaults": {
			config: &Config{
				Defaults: Defaults{
					Service: service.Service{
						Options: opt.Options{
							Interval: "1x"}}}},
			errContains: `interval: "1x" <invalid>`,
		},
		"invalid Notify": {
			config: &Config{
				Notify: shoutrrr.Slice{
					"test": &shoutrrr.Shoutrrr{
						Options: map[string]string{
							"delay": "2x",
						}}}},
			errContains: `delay: "2x" <invalid>`,
		},
		"invalid WebHook": {
			config: &Config{
				WebHook: webhook.Slice{
					"test": &webhook.WebHook{
						Delay: "3x",
					}}},
			errContains: `delay: "3x" <invalid>`,
		},
		"invalid Service": {
			config: &Config{
				Service: service.Slice{
					"test": &service.Service{
						Options: opt.Options{
							Interval: "4x"},
					}}},
			errContains: `interval: "4x" <invalid>`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			// Switch Fatal to panic and disable this panic.
			if !tc.noPanic {
				jLog.Testing = true
				defer func() {
					_ = recover()

					w.Close()
					out, _ := io.ReadAll(r)
					os.Stdout = stdout
					output := string(out)
					if !strings.Contains(output, tc.errContains) {
						t.Fatalf("%s should have panic'd with %q, not:\n%s",
							name, tc.errContains, output)
					}

				}()
			}

			// WHEN CheckValues is called on them
			tc.config.CheckValues()

			// THEN this call will/wont crash the program
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			if !strings.Contains(output, tc.errContains) {
				t.Fatalf("%s should have panic'd with %q. stdout:\n%s",
					name, tc.errContains, output)
			}
		})
	}
}

func TestConfig_Print(t *testing.T) {
	// GIVEN a Config and print flags of true and false
	jLog = util.NewJLog("WARN", false)
	jLog.Testing = true
	config := testVerify()
	tests := map[string]struct {
		flag        bool
		wantedLines int
	}{
		"flag on":  {flag: true, wantedLines: 153},
		"flag off": {flag: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called with these flags
			config.Print(&tc.flag)

			// THEN config is printed onlt when the flag is true
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.wantedLines {
				t.Errorf("Print with %s wants %d lines but got %d\n%s",
					name, tc.wantedLines, got, string(out))
			}
		})
	}
}
