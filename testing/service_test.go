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

package testing

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
)

func TestServiceTest(t *testing.T) {
	// GIVEN a Config with a Service
	jLog = utils.NewJLog("INFO", false)
	InitJLog(jLog)
	tests := map[string]struct {
		flag          string
		slice         service.Slice
		outputRegex   *string
		panicContains *string
	}{
		"flag is empty": {flag: "",
			outputRegex: stringPtr("^$"),
			slice: service.Slice{
				"argus": {
					ID:       stringPtr("argus"),
					Interval: stringPtr("0s"),
				},
			}},
		"unknown service": {flag: "test",
			panicContains: stringPtr("Service \"test\" could not be found in config.service\nDid you mean one of these?\n  - argus"),
			slice: service.Slice{
				"argus": {
					ID:       stringPtr("argus"),
					Interval: stringPtr("0s"),
				},
			}},
		"github service": {flag: "argus",
			outputRegex: stringPtr("argus, Latest Release - \"[0-9]+\\.[0-9]+\\.[0-9]+\""),
			slice: service.Slice{
				"argus": {
					Type: stringPtr("github"),
					ID:   stringPtr("argus"),
					URL:  stringPtr("release-argus/Argus"),
					URLCommands: &service.URLCommandSlice{
						{Type: "regex", Regex: stringPtr("[0-9.]+$")},
					},
					AllowInvalidCerts:  boolPtr(false),
					SemanticVersioning: boolPtr(true),
					Interval:           stringPtr("0s"),
				},
			}},
		"url service type but github owner/repo url": {flag: "argus",
			outputRegex: stringPtr("This URL looks to be a GitHub repo, but the service's type is url, not github"),
			slice: service.Slice{
				"argus": {
					Type: stringPtr("url"),
					ID:   stringPtr("argus"),
					URL:  stringPtr("release-argus/Argus"),
					URLCommands: &service.URLCommandSlice{
						{Type: "regex", Regex: stringPtr("[0-9.]+$")},
					},
					AllowInvalidCerts:  boolPtr(false),
					SemanticVersioning: boolPtr(true),
					Interval:           stringPtr("0s"),
				},
			}},
		"url service": {flag: "argus",
			outputRegex: stringPtr("Latest Release - \"[0-9]+\\.[0-9]+\\.[0-9]+\""),
			slice: service.Slice{
				"argus": {
					Type: stringPtr("url"),
					ID:   stringPtr("argus"),
					URL:  stringPtr("https://github.com/release-argus/Argus/releases"),
					URLCommands: &service.URLCommandSlice{
						{Type: "regex", Regex: stringPtr("tag/([0-9.]+)\"")},
					},
					AllowInvalidCerts:  boolPtr(false),
					SemanticVersioning: boolPtr(true),
					Interval:           stringPtr("0s"),
				},
			}},
		"service with deployed version lookup": {flag: "argus",
			outputRegex: stringPtr("Latest Release - \"[0-9]+\\.[0-9]+\\.[0-9]+\"\\s.*Deployed version - \"[0-9]+\\.[0-9]+\\.[0-9]+\""),
			slice: service.Slice{
				"argus": {
					Type: stringPtr("url"),
					ID:   stringPtr("argus"),
					URL:  stringPtr("https://github.com/release-argus/Argus/releases"),
					URLCommands: &service.URLCommandSlice{
						{Type: "regex", Regex: stringPtr("tag/([0-9.]+)\"")},
					},
					AllowInvalidCerts:  boolPtr(false),
					SemanticVersioning: boolPtr(true),
					Interval:           stringPtr("0s"),
					DeployedVersionLookup: &service.DeployedVersionLookup{
						URL:               "https://release-argus.io/demo/api/v1/version",
						AllowInvalidCerts: boolPtr(true),
						JSON:              "version",
					},
				},
			}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			jLog.Testing = true
			if tc.panicContains != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					rStr := fmt.Sprint(r)
					if !strings.Contains(rStr, *tc.panicContains) {
						t.Errorf("%s:\nexpected a panic containing %q, not %q",
							name, *tc.panicContains, rStr)
					}
				}()
			}
			if tc.slice[tc.flag] != nil {
				tc.slice[tc.flag].Init(jLog, &service.Service{}, &service.Service{})
				// will do a call for latest_version* and one for deployed_version*
				dbChannel := make(chan db_types.Message, 2)
				tc.slice[tc.flag].DatabaseChannel = &dbChannel
				if tc.slice[tc.flag].DeployedVersionLookup != nil {
					tc.slice[tc.flag].DeployedVersionLookup.Defaults = &service.DeployedVersionLookup{}
					tc.slice[tc.flag].DeployedVersionLookup.HardDefaults = &service.DeployedVersionLookup{}
				}
			}

			// WHEN ServiceTest is called with the test Config
			cfg := config.Config{
				Service: tc.slice,
			}
			ServiceTest(&tc.flag, &cfg, jLog)

			// THEN we get the expected output
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			if tc.outputRegex != nil {
				re := regexp.MustCompile(*tc.outputRegex)
				match := re.MatchString(output)
				if !match {
					t.Errorf("%s:\nwant match on %q\ngot: %q",
						name, *tc.outputRegex, output)
				}
			}
		})
	}
	time.Sleep(100 * time.Millisecond)
}
