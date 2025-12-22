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

//go:build integration

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
	logutil "github.com/release-argus/Argus/util/log"
)

func resetFlags() {
	configFile = test.StringPtr("")
	configCheckFlag = test.BoolPtr(false)
	testCommandsFlag = test.StringPtr("")
	testNotifyFlag = test.StringPtr("")
	testServiceFlag = test.StringPtr("")
}

func TestRun(t *testing.T) {
	// Log.
	logtest.InitLog()

	// GIVEN different Configs to test.
	tests := map[string]struct {
		file           func(path string)
		preStartFunc   func(baseDir string)
		outputContains *[]string
		exitCode       *int
	}{
		"config with no services": {
			file: testYAML_NoServices,
			outputContains: &[]string{
				"Found 0 services to monitor",
				"Listening on "}},
		"config with services, db invalid format": {
			file: testYAML_Argus,
			preStartFunc: func(baseDir string) {
				// Create an invalid database file.
				dbFile := filepath.Join(baseDir, "argus.db")
				_ = os.WriteFile(dbFile, []byte("invalid format"), 0644)
			},
			outputContains: &[]string{
				"file is not a database"},
			exitCode: test.IntPtr(1)},
		"config with services": {
			file: testYAML_Argus,
			outputContains: &[]string{
				"services to monitor:",
				"release-argus/Argus, Latest Release - ",
				"Listening on "}},
		"config with services and some !active": {
			file: testYAML_Argus_SomeInactive,
			outputContains: &[]string{
				"Found 1 services to monitor:"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout and sharing log exitCodeChannel.
			releaseStdout := test.CaptureLog(logutil.Log)

			tempDir := t.TempDir()
			file := filepath.Join(tempDir, "config.yml")
			tc.file(file)
			resetFlags()
			configFile = &file
			accessToken := os.Getenv("GITHUB_TOKEN")
			_ = os.Setenv("ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN", accessToken)
			// Add tempDir to database file path.
			_ = os.Setenv("ARGUS_DATA_DATABASE_FILE", filepath.Join(tempDir, "argus.db"))
			t.Cleanup(func() { _ = os.Unsetenv("ARGUS_DATA_DATABASE_FILE") })
			if tc.preStartFunc != nil {
				tc.preStartFunc(tempDir)
			}

			exitCodeChannel := make(chan int)
			// WHEN run is called.
			go func() {
				exitCodeChannel <- run()
			}()

			var exitCode *int
			select {
			case code := <-exitCodeChannel:
				exitCode = &code
			case <-time.After(3 * time.Second):
				// Cancel after 3 seconds.
				logutil.Log.Fatal("--TestRun--", logutil.LogFrom{})
				time.Sleep(time.Second)
			}

			// THEN the program will have printed everything expected.
			stdout := releaseStdout()
			if tc.outputContains != nil {
				for _, text := range *tc.outputContains {
					if !strings.Contains(stdout, text) {
						t.Errorf("%s\n%q couldn't be found in stdout:\n%s",
							packageName, text, stdout)
					}
				}
			}
			// AND the exit code is as expected.
			wantCode := test.StringifyPtr(tc.exitCode)
			gotCode := test.StringifyPtr(exitCode)
			if wantCode != gotCode {
				t.Errorf("%s\nexit code mismatch\nwant: %s\ngot:  %s",
					packageName, wantCode, gotCode)
			}
		})
	}
}
