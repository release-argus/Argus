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
	"os/signal"
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
	tests := []struct {
		name           string
		file           func(t *testing.T, path string)
		preStartFunc   func(t *testing.T, baseDir string)
		resetPassword  string // Username for -auth.reset-password.
		serverStarts   bool   // The web server comes up.
		awaitService   string // Wait for this service's first query before shutting down.
		outputContains []string
		outputExcludes []string
		exitCode       *int
	}{
		{
			name: "config fails to load - exits immediately without starting the server",
			file: testYAML_Invalid,
			outputContains: []string{
				"failed to decode",
			},
			outputExcludes: []string{
				"services to monitor",
				"Listening on ",
			},
			exitCode: new(1),
		},
		{
			name: "config with services, db invalid format",
			file: testYAML_Argus,
			preStartFunc: func(t *testing.T, baseDir string) {
				dbFile := filepath.Join(baseDir, "argus.db")
				if err := os.WriteFile(dbFile, []byte("invalid format"), 0644); err != nil {
					t.Fatalf("os.WriteFile() failed: %v", err)
				}
			},
			outputContains: []string{
				"file is not a database",
			},
			exitCode: new(1),
		},
		{
			name:          "auth setup fails - exits without starting the server",
			file:          testYAML_AuthEnabled,
			resetPassword: "ghost",
			outputContains: []string{
				`not found: user "ghost"`,
			},
			outputExcludes: []string{
				"Listening on ",
			},
			exitCode: new(1),
		},
		{
			name:         "config with no services",
			file:         testYAML_NoServices,
			serverStarts: true,
			outputContains: []string{
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
			outputContains: []string{
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
			outputContains: []string{
				"Found 1 services to monitor:",
				"Shutdown complete",
			},
			exitCode: new(0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars, captures stdout and shares the flag vars.

			releaseStdout := test.CaptureLog(t, logx.Default())

			tempDir := t.TempDir()
			file := filepath.Join(tempDir, "config.yml")
			tc.file(t, file)
			resetFlags()
			configFile = &file
			if tc.resetPassword != "" {
				config.AuthResetPassword = new(tc.resetPassword)
				if err := flag.Set("auth.reset-password", tc.resetPassword); err != nil {
					t.Fatalf("flag.Set() failed: %v", err)
				}
				t.Cleanup(func() { _ = flag.Set("auth.reset-password", "") })
			}
			env := map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN": test.GitHubToken(t),
				"ARGUS_DATA_DATABASE_FILE":                         filepath.Join(tempDir, "argus.db"),
				"ARGUS_LOG_LEVEL":                                  "DEBUG",
			}
			port := freePort(t)
			env["ARGUS_WEB_LISTEN_PORT"] = strconv.Itoa(port)
			test.IsolateEnv(t, env)
			if tc.preStartFunc != nil {
				tc.preStartFunc(t, tempDir)
			}

			if tc.awaitService != "" {
				// Process-global, so a repeat run would see the last value.
				metric.LatestVersionQueryResultLast.Reset()
			}

			resultChannel := make(chan int, 1)
			var exitCode *int

			go func() {
				resultChannel <- run()
			}()

			// Covers a wait below failing, or run() not exiting.
			t.Cleanup(func() {
				if exitCode != nil {
					return
				}
				select {
				case <-resultChannel:
					return
				default:
				}
				// Keeps the SIGTERM from killing the test binary should run() return first.
				sink := make(chan os.Signal, 1)
				signal.Notify(sink, syscall.SIGTERM)
				defer signal.Stop(sink)
				_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
				select {
				case <-resultChannel:
				case <-time.After(15 * time.Second):
				}
			})

			// Shut the run down once it is serving, rather than leaking it.
			if tc.serverStarts {
				waitForListener(t, port)
				if tc.awaitService != "" {
					waitForServiceQuery(t, port, tc.awaitService)
				}
				if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
					t.Fatalf("syscall.Kill() failed: %v", err)
				}
			}

			select {
			case code := <-resultChannel:
				exitCode = &code
			case <-time.After(15 * time.Second):
				t.Error("run timed out waiting for exit code")
			}

			stdout := releaseStdout()
			for _, text := range tc.outputContains {
				if !strings.Contains(stdout, text) {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, text)
				}
			}

			for _, text := range tc.outputExcludes {
				if strings.Contains(stdout, text) {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: NOT %q", stdout, text)
				}
			}

			wantCode := test.StringifyPtr(tc.exitCode)
			gotCode := test.StringifyPtr(exitCode)
			if wantCode != gotCode {
				t.Errorf("exit code mismatch\ngot:  %s\nwant: %s", gotCode, wantCode)
			}
		})
	}
}

func TestAwaitShutdown(t *testing.T) {
	tests := []struct {
		name     string
		blocks   bool
		timeout  time.Duration
		wantLog  string
		wantCode int
	}{
		{
			name:     "wait returns in time",
			timeout:  10 * time.Second,
			wantLog:  "Shutdown complete",
			wantCode: 0,
		},
		{
			name:     "wait outlasts the timeout",
			blocks:   true,
			timeout:  time.Millisecond,
			wantLog:  "shutdown timed out",
			wantCode: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())
			release := make(chan struct{})
			t.Cleanup(func() { close(release) })
			wait := func() error {
				if tc.blocks {
					<-release
				}
				return nil
			}

			got := awaitShutdown(wait, tc.timeout)

			stdout := releaseStdout()
			if got != tc.wantCode {
				t.Errorf("exit code mismatch\ngot:  %d\nwant: %d", got, tc.wantCode)
			}
			if !strings.Contains(stdout, tc.wantLog) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.wantLog)
			}
		})
	}
}

// freePort returns a port that is free at the time of the call.
func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() failed: %v", err)
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
	var err error
	for time.Now().Before(deadline) {
		var conn net.Conn
		conn, err = net.DialTimeout("tcp", address, time.Second)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out: net.DialTimeout(%q) failed: %v", address, err)
}

// waitForServiceQuery blocks until the running instance has completed a
// latest-version query for serviceID.
func waitForServiceQuery(t *testing.T, port int, serviceID string) {
	t.Helper()

	want := fmt.Sprintf(`latest_version_query_result_last{id="%s"`, serviceID)
	metricsURL := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)
	deadline := time.Now().Add(15 * time.Second)
	var body string
	for time.Now().Before(deadline) {
		body = fetch(metricsURL)
		if strings.Contains(body, want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	var got []string
	for line := range strings.Lines(body) {
		if strings.HasPrefix(line, "latest_version_query_result_last") {
			got = append(got, strings.TrimSpace(line))
		}
	}
	t.Fatalf("timed out: /metrics mismatch\ngot:  %q\nwant: %q", got, want)
}

// fetch GETs target, returning an empty body on any failure.
func fetch(target string) string {
	//#nosec G107 -- the URL is the test's own listener.
	resp, err := http.Get(target)
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
