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
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/webhook"
)

func testVerify() (cfg *Config) {
	cfg = &Config{}
	cfg.Order = []string{"test"}
	cfg.Defaults = Defaults{}
	cfg.Defaults.SetDefaults()
	cfg.Notify = shoutrrr.SliceDefaults{
		"test": shoutrrr.NewDefaults(
			"discord",
			&cfg.Defaults.Notify["discord"].Options,
			&cfg.Defaults.Notify["discord"].Params,
			&cfg.Defaults.Notify["discord"].URLFields)}
	cfg.WebHook = webhook.SliceDefaults{
		"test": &cfg.Defaults.WebHook,
	}
	serviceID := "test"
	cfg.Service = service.Slice{
		serviceID: &service.Service{
			ID: serviceID,
			LatestVersion: latestver.Lookup{
				Type: "github",
				URL:  "release-argus/argus",
			},
		},
	}
	return
}

func TestConfig_CheckValues(t *testing.T) {
	// GIVEN variations of Config to test
	tests := map[string]struct {
		config   *Config
		errRegex []string
		noPanic  bool
	}{
		"valid Config": {
			config:  testVerify(),
			noPanic: true,
		},
		"invalid Defaults": {
			config: &Config{
				Defaults: Defaults{
					Service: service.Defaults{
						Options: *opt.NewDefaults("1x", nil)}}},
			errRegex: []string{
				`^defaults:$`,
				`^  service:$`,
				`^    options:$`,
				`^      interval: "[^"]+" <invalid>`},
		},
		"invalid Notify": {
			config: &Config{
				Notify: shoutrrr.SliceDefaults{
					"test": shoutrrr.NewDefaults(
						"discord",
						&map[string]string{
							"delay": "2x"},
						nil, nil)}},
			errRegex: []string{
				`^notify:$`,
				`^  test:$`,
				`^    options:$`,
				`^      delay: "[^"]+" <invalid>`},
		},
		"invalid WebHook": {
			config: &Config{
				WebHook: webhook.SliceDefaults{
					"test": webhook.NewDefaults(
						nil, nil, "3x", nil, nil, "", nil, "", "")}},
			errRegex: []string{
				`^webhook:$`,
				`^  test:$`,
				`^    delay: "3x" <invalid>`},
		},
		"invalid Service": {
			config: &Config{
				Service: service.Slice{
					"test": &service.Service{
						Options: *opt.New(
							nil, "4x", nil,
							nil, nil)}}},
			errRegex: []string{
				`^service:$`,
				`^  test:$`,
				`^    options:$`,
				`^      interval: "4x" <invalid>`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			if tc.config != nil {
				for name, svc := range tc.config.Service {
					svc.ID = name
				}
			}
			// Switch Fatal to panic and disable this panic.
			if !tc.noPanic {
				defer func() {
					_ = recover()

					w.Close()
					out, _ := io.ReadAll(r)
					os.Stdout = stdout
					output := string(out)
					lines := strings.Split(output, "\n")
					if len(tc.errRegex) == 0 {
						t.Fatalf("want 0 errors, not %d:\n%v",
							len(lines), lines)
					}
					if len(tc.errRegex) > len(lines) {
						t.Fatalf("want %d errors:\n['%s']\ngot %d errors:\n%v\noutput: %q",
							len(tc.errRegex), strings.Join(tc.errRegex, `'  '`), len(lines), lines, output)
					}
					for i := range tc.errRegex {
						re := regexp.MustCompile(tc.errRegex[i])
						match := re.MatchString(lines[i])
						if !match {
							t.Errorf("want match for: %q\ngot:  %q",
								tc.errRegex[i], output)
							return
						}
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
			lines := strings.Split(output, `\n`)
			if len(tc.errRegex) > len(lines) {
				t.Errorf("want %d errors:\n%v\ngot %d errors:\n%v",
					len(tc.errRegex), tc.errRegex, len(lines), lines)
				return
			}
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				match := re.MatchString(lines[i])
				if !match {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], output)
					return
				}
			}
		})
	}
}

func TestConfig_Print(t *testing.T) {
	// GIVEN a Config and print flags of true and false
	config := testVerify()
	tests := map[string]struct {
		flag  bool
		lines int
	}{
		"flag on":  {flag: true, lines: 174 + len(config.Defaults.Notify)},
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
			if got != tc.lines {
				t.Errorf("Print with %s wants %d lines but got %d\n%s",
					name, tc.lines, got, string(out))
			}
		})
	}
}
