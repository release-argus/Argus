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
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/web/metric"
)

func resetFlags() {
	configFile = new("")
	configCheckFlag = new(false)
	testCommandsFlag = new("")
	testNotifyFlag = new("")
	testServiceFlag = new("")
	config.AuthResetPassword = new("")
}

func TestRun(t *testing.T) {
	// GIVEN: different Configs to test.
	tests := []struct {
		name           string
		file           func(path string)
		preStartFunc   func(baseDir string)
		resetPassword  string // Username for -auth.reset-password.
		serverStarts   bool   // The web server comes up.
		awaitService   string // Wait for this service's first query before shutting down.
		outputContains *[]string
		outputExcludes *[]string
		exitCode       *int
	}{
		{
			name: "config fails to load - exits immediately without starting the server",
			file: testYAML_Invalid,
			outputContains: &[]string{
				"Unmarshal of",
			},
			outputExcludes: &[]string{
				"services to monitor",
				"Listening on ",
			},
			exitCode: new(1),
		},
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
			exitCode: new(1),
		},
		{
			name:          "auth setup fails - exits without starting the server",
			file:          testYAML_AuthEnabled,
			resetPassword: "ghost",
			outputContains: &[]string{
				`not found: user "ghost"`,
			},
			outputExcludes: &[]string{
				"Listening on ",
			},
			exitCode: new(1),
		},
		{
			name:         "config with no services",
			file:         testYAML_NoServices,
			serverStarts: true,
			outputContains: &[]string{
				"Found 0 services to monitor",
				"Listening on ",
				"Shutdown complete",
			},
			exitCode: new(0),
		},
		{
			name:         "config with services",
			file:         testYAML_Argus,
			serverStarts: true,
			awaitService: "SERVICE_NAME",
			outputContains: &[]string{
				"services to monitor:",
				"SERVICE_NAME, Latest Release - ",
				"Listening on ",
				"Shutdown complete",
			},
			exitCode: new(0),
		},
		{
			name:         "config with services and some !active",
			file:         testYAML_Argus_SomeInactive,
			serverStarts: true,
			outputContains: &[]string{
				"Found 1 services to monitor:",
				"Shutdown complete",
			},
			exitCode: new(0),
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
			if tc.resetPassword != "" {
				config.AuthResetPassword = new(tc.resetPassword)
				if err := flag.Set("auth.reset-password", tc.resetPassword); err != nil {
					t.Fatalf(
						"%s\nset auth.reset-password: %v",
						packageName, err,
					)
				}
				t.Cleanup(func() { _ = flag.Set("auth.reset-password", "") })
			}
			env := map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN": test.GitHubToken(t),
				"ARGUS_DATA_DATABASE_FILE":                         filepath.Join(tempDir, "argus.db"),
			}
			port := freePort(t)
			env["ARGUS_WEB_LISTEN_PORT"] = strconv.Itoa(port)
			test.SetEnv(t, env)
			if tc.preStartFunc != nil {
				tc.preStartFunc(tempDir)
			}

			if tc.awaitService != "" {
				// Process-global, so a repeat run would see the last value.
				metric.LatestVersionQueryResultLast.Reset()
			}

			resultChannel := make(chan int)
			// WHEN: run is called.
			go func() {
				resultChannel <- run()
			}()

			// Shut the run down once it is serving, rather than leaking it.
			if tc.serverStarts {
				waitForListener(t, port)
				if tc.awaitService != "" {
					waitForServiceQuery(t, port, tc.awaitService)
				}
				if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
					t.Fatalf(
						"%s\nSIGTERM failed: %v",
						packageName, err,
					)
				}
			}

			var exitCode *int
			select {
			case code := <-resultChannel:
				exitCode = &code
			case <-time.After(15 * time.Second):
				t.Errorf("%s\nrun timed out waiting for exit code", packageName)
			}

			// THEN: the program will have printed everything expected.
			stdout := releaseStdout()
			t.Logf(
				"%s\nstdout: %q",
				packageName, stdout,
			)
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

			// AND: the program will not have printed anything unexpected.
			if tc.outputExcludes != nil {
				for _, text := range *tc.outputExcludes {
					if strings.Contains(stdout, text) {
						t.Errorf(
							"%s\n%q unexpectedly found in stdout:\n%s",
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

// freePort returns a port that is free at the time of the call.
func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf(
			"%s\nreserve a port: %v",
			packageName, err,
		)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

// waitForListener blocks until port accepts connections.
func waitForListener(t *testing.T, port int) {
	t.Helper()

	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf(
		"%s\nnothing listening on %s within the deadline",
		packageName, address,
	)
}

// waitForServiceQuery blocks until the running instance has completed a
// latest-version query for serviceID.
func waitForServiceQuery(t *testing.T, port int, serviceID string) {
	t.Helper()

	want := fmt.Sprintf(`latest_version_query_result_last{id="%s"`, serviceID)
	url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(fetch(t, url), want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf(
		"%s\n%s never reported a query at %s",
		packageName, serviceID, url,
	)
}

// fetch GETs url, returning an empty body on any failure.
func fetch(t *testing.T, url string) string {
	t.Helper()

	//#nosec G107 -- the URL is the test's own listener.
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}
