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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"golang.org/x/sync/errgroup"
)

func TestConfig_Load(t *testing.T) {
	DebounceDuration = time.Second
	// GIVEN a test file to load.
	tests := []struct {
		name        string
		setupFile   func(path string)
		envVars     map[string]string
		validate    func(t *testing.T, config *Config, stdout string)
		exitCode    *int
		stdoutRegex string
	}{
		{
			name:      "Environment variables loaded",
			setupFile: testYAML_config_test,
			envVars: map[string]string{
				"TEST_ENV_KEY": "1234",
			},
			validate: func(t *testing.T, config *Config, stdout string) {
				checks := map[string]struct {
					want, got string
				}{
					"Service.Options.Interval": {want: "123s", got: config.Defaults.Service.Options.Interval},
					"Notify.slack.title":       {want: "defaultTitle", got: config.Defaults.Notify["slack"].GetParam("title")},
					"WebHook.Delay":            {want: "2s", got: config.Defaults.WebHook.Delay},
					"EmptyService.String('')":  {want: "", got: config.Service["EmptyService"].String("")},
				}
				for name, check := range checks {
					if check.got != check.want {
						t.Errorf("%s\nmismatch on %q\nwant: %q\ngot:  %q",
							packageName, name, check.want, check.got)
					}
				}
				if config.Defaults.Service.Options.Interval != "123s" {
					t.Errorf("Expected interval 123s, got %s", config.Defaults.Service.Options.Interval)
				}
				if config.Defaults.Notify["slack"].GetParam("title") != "defaultTitle" {
					t.Errorf("Expected defaultTitle, got %s", config.Defaults.Notify["slack"].GetParam("title"))
				}
				if config.Defaults.WebHook.Delay != "2s" {
					t.Errorf("Expected WebHook.Delay 2s, got %s", config.Defaults.WebHook.Delay)
				}
				if config.Service["EmptyService"].String("") != "" {
					t.Errorf("Expected EmptyService to be deleted or empty, got %s", config.Service["EmptyService"].String(""))
				}
			},
		},
		{
			name:      "Nil services deleted",
			setupFile: testYAML_SomeNilServices,
			validate: func(t *testing.T, config *Config, stdout string) {
				for name, svc := range config.Service {
					if svc == nil {
						t.Errorf("%s\nService %q is nil",
							packageName, name)
					}
				}
				if len(config.Service) != 2 {
					t.Errorf("%s\nExpected 2 services, got %d",
						packageName, len(config.Service))
				}
			},
		},
		{
			name:      "Defaults assigned to services",
			setupFile: testYAML_config_test,
			validate: func(t *testing.T, config *Config, stdout string) {
				want := false
				got := config.Service["WantDefaults"].Options.GetSemanticVersioning()
				if got != want {
					t.Errorf("%s\nService.X.Options.SemanticVersioning mismatch\nwant: %v\ngot:  %t",
						packageName, want, *config.Service["WantDefaults"].Options.SemanticVersioning)
				}
			},
		},
		{
			name:      "Nil service map initialises empty map",
			setupFile: testYAML_NilServiceMap,
			validate: func(t *testing.T, config *Config, stdout string) {
				if config.Service == nil {
					t.Errorf("%s\nconfig.Service is nil after Load, want non-nil",
						packageName)
				}
			},
		},
		{
			name:      "Invalid YAML returns exit code",
			setupFile: testYAML_InvalidYAML,
			validate: func(t *testing.T, config *Config, stdout string) {
				wantRegex := `Unmarshal of "[^"]+" failed`
				if !util.RegexCheck(wantRegex, stdout) {
					t.Errorf("%s\nstdout mismatch:\nwant: %q\ngot:  %q",
						packageName, wantRegex, stdout)
				}
			},
			exitCode: test.IntPtr(1),
		},
		{
			name: "Config that is unreadable",
			setupFile: func(path string) {
				testYAML_config_test(path)
				_ = os.Chmod(path, 0_222)
			},
			exitCode: test.IntPtr(1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout and sharing log exitCodeChannel.
			releaseStdout := test.CaptureLog(logutil.Log)

			g, _ := errgroup.WithContext(t.Context())
			flags := make(map[string]bool)
			file := filepath.Join(t.TempDir(), "config.yml")

			// Setup YAML file
			if tc.setupFile != nil {
				tc.setupFile(file)
			}
			t.Cleanup(func() {
				// Give time for save before TempDir clean-up.
				time.Sleep(2 * DebounceDuration)
			})

			// Set environment variables
			for k, v := range tc.envVars {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}

			// WHEN we Load this config.
			var config Config
			config.Load(t.Context(), g, file, &flags)

			stdout := releaseStdout()
			// Per-test validation.
			if tc.validate != nil {
				tc.validate(t, &config, stdout)
			}

			// THEN the exit code is as expected.
			exitCodeChannel := logutil.ExitCodeChannel()
			var exitCode *int
			select {
			case msg := <-exitCodeChannel:
				t.Logf("%s\n%s", packageName, msg)
				exitCode = test.IntPtr(1)
			default:
			}
			wantExitCode := test.StringifyPtr(tc.exitCode)
			gotExitCode := test.StringifyPtr(exitCode)
			if gotExitCode != wantExitCode {
				t.Errorf("%s\nunexpected exit code\nwant: %s\ngot:  %s",
					packageName, wantExitCode, gotExitCode)
			}
		})
	}
}
