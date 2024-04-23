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

//go:build integration

package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func resetFlags() {
	configFile = test.StringPtr("")
	configCheckFlag = test.BoolPtr(false)
	testCommandsFlag = test.StringPtr("")
	testNotifyFlag = test.StringPtr("")
	testServiceFlag = test.StringPtr("")
}

func TestTheMain(t *testing.T) {
	// GIVEN different Config's to test
	jLog = *util.NewJLog("WARN", false)
	jLog.Testing = true
	tests := map[string]struct {
		file           func(path string, t *testing.T)
		outputContains *[]string
		db             string
	}{
		"config with no services": {
			file: testYAML_NoServices,
			db:   "test-no_services.db",
			outputContains: &[]string{
				"Found 0 services to monitor",
				"Listening on "}},
		"config with services": {
			file: testYAML_Argus,
			db:   "test-argus.db",
			outputContains: &[]string{
				"services to monitor:",
				"release-argus/Argus, Latest Release - ",
				"Listening on "}},
		"config with services and some !active": {
			file: testYAML_Argus_SomeInactive,
			db:   "test-argus-some-inactive.db",
			outputContains: &[]string{
				"Found 1 services to monitor:"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			file := fmt.Sprintf("%s.yml", name)
			os.Remove(tc.db)
			tc.file(file, t)
			defer os.Remove(tc.db)
			resetFlags()
			configFile = &file
			accessToken := os.Getenv("GITHUB_TOKEN")
			os.Setenv("ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN", accessToken)

			// WHEN Main is called
			go main()
			time.Sleep(3 * time.Second)

			// THEN the program will have printed everything expected
			stdout := releaseStdout()
			if tc.outputContains != nil {
				for _, text := range *tc.outputContains {
					if !strings.Contains(stdout, text) {
						t.Errorf("%q couldn't be found in stdout:\n%s",
							text, stdout)
					}
				}
			}
		})
	}
}
