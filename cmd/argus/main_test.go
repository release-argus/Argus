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

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
)

func resetFlags() {
	configFile = test.Ptr("")
	configCheckFlag = test.Ptr(false)
	testCommandsFlag = test.Ptr("")
	testNotifyFlag = test.Ptr("")
	testServiceFlag = test.Ptr("")
}

func TestRun(t *testing.T) {
	// GIVEN: different Configs to test.
	tests := []struct {
		name           string
		file           func(path string)
		preStartFunc   func(baseDir string)
		outputContains *[]string
		exitCode       *int
	}{
		{
			name: "config with services, db invalid format",
			file: testYAML_Argus,
			preStartFunc: func(baseDir string) {
				// Create an invalid database file.
				dbFile := filepath.Join(baseDir, "argus.db")
				_ = os.WriteFile(dbFile, []byte("invalid format"), 0644)
			},
			outputContains: &[]string{
				"file is not a database",
			},
			exitCode: test.Ptr(1),
		},
		{
			name: "config with no services",
			file: testYAML_NoServices,
			outputContains: &[]string{
				"Found 0 services to monitor",
				"Listening on ",
			},
		},
		{
			name: "config with services",
			file: testYAML_Argus,
			outputContains: &[]string{
				"services to monitor:",
				"SERVICE_NAME, Latest Release - ",
				"Listening on ",
			},
		},
		{
			name: "config with services and some !active",
			file: testYAML_Argus_SomeInactive,
			outputContains: &[]string{
				"Found 1 services to monitor:",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout and sharing log resultChannel.
			releaseStdout := test.CaptureLog(t, logx.Default())

			tempDir := t.TempDir()
			file := filepath.Join(tempDir, "config.yml")
			tc.file(file)
			resetFlags()
			configFile = &file
			env := map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN": test.GitHubToken(t),
				"ARGUS_DATA_DATABASE_FILE":                  filepath.Join(tempDir, "argus.db"),
			}
			test.SetEnv(t, env)
			if tc.preStartFunc != nil {
				tc.preStartFunc(tempDir)
			}

			resultChannel := make(chan int)
			// WHEN: run is called.
			go func() {
				resultChannel <- run()
			}()

			var exitCode *int
			select {
			case code := <-resultChannel:
				exitCode = &code
			case <-time.After(6 * time.Second):
				if tc.exitCode != nil {
					t.Logf("%s\nrun timed out waiting for exit code", packageName)
				}
			}

			// THEN: the program will have printed everything expected.
			stdout := releaseStdout()
			t.Logf("%s\nstdout: %q", packageName, stdout)
			if tc.outputContains != nil {
				for _, text := range *tc.outputContains {
					if !strings.Contains(stdout, text) {
						t.Errorf(
							"%s\n%q couldn't be found in stdout:\n%s",
							packageName, text, stdout,
						)
					}
				}
			}

			// AND: the exit code is as expected.
			wantCode := test.StringifyPtr(tc.exitCode)
			gotCode := test.StringifyPtr(exitCode)
			if wantCode != gotCode {
				t.Errorf(
					"%s\nexit code mismatch\ngot:  %s\nwant: %s",
					packageName, gotCode, wantCode,
				)
			}
		})
	}
}
