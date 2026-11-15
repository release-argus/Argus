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

//go:build unit

package config

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestConfig_Load(t *testing.T) {
	originalDebounce := DebounceDuration
	DebounceDuration = time.Second
	t.Cleanup(func() { DebounceDuration = originalDebounce })

	tests := []struct {
		name      string
		setupFile func(t *testing.T, path string)
		env       map[string]string
		validate  func(t *testing.T, config *Config, stdout string)
		exitCode  *int
	}{
		{
			name:      "Environment variables loaded",
			setupFile: testYAML_config_test,
			env: map[string]string{
				"TEST_ENV_KEY": "1234",
			},
			validate: func(t *testing.T, config *Config, stdout string) {
				fieldTests := []test.FieldAssertion{
					{Name: "Defaults.Service.Options.Interval", Got: config.Defaults.Service.Options.Interval, Want: "123s", Mode: test.CompareEqual},
					{Name: `Defaults.Notify["slack"].GetParam("title")`, Got: config.Defaults.Notify["slack"].GetParam("title"), Want: "defaultTitle", Mode: test.CompareEqual},
					{Name: "Defaults.Webhook.Delay", Got: config.Defaults.Webhook.Delay, Want: "2s", Mode: test.CompareEqual},
					{Name: `Service["EmptyService"].String("")`, Got: config.Service["EmptyService"].String(""), Want: "", Mode: test.CompareEqual},
				}
				if testErr := test.AssertFields(t, fieldTests, "Config"); testErr != nil {
					t.Fatal(testErr)
				}
			},
		},
		{
			name: "Unreadable .env file is warned about",
			setupFile: func(t *testing.T, path string) {
				testYAML_config_test(t, path)
				envFile := filepath.Join(filepath.Dir(path), ".env")
				if err := os.WriteFile(envFile, nil, 0_000); err != nil {
					t.Fatalf("os.WriteFile() failed: %v", err)
				}
				t.Cleanup(func() { _ = os.Remove(envFile) })
			},
			validate: func(t *testing.T, config *Config, stdout string) {
				wantRegex := `WARNING: failed to open env file "[^"]+\.env"`
				if !util.RegexCheck(wantRegex, stdout) {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, wantRegex)
				}
			},
		},
		{
			name:      "Nil services deleted",
			setupFile: testYAML_SomeNilServices,
			validate: func(t *testing.T, config *Config, stdout string) {
				for name, svc := range config.Service {
					if svc == nil {
						t.Errorf(
							"Config.Service[%q] mismatch\ngot:  present\nwant: absent (deleted)",
							name,
						)
					}
				}
				if got, want := len(config.Service), 2; got != want {
					t.Errorf(
						"len(Config.Service) mismatch\ngot:  %d (%q)\nwant: %d",
						got, util.SortedKeys(config.Service), want,
					)
				}
			},
		},
		{
			name:      "Defaults assigned to services",
			setupFile: testYAML_config_test,
			validate: func(t *testing.T, config *Config, stdout string) {
				if config.Service["WantDefaults"].Options.GetSemanticVersioning() {
					t.Error(
						"Config.Service[\"WantDefaults\"].Options.GetSemanticVersioning() mismatch\n" +
							"got:  true\nwant: false",
					)
				}
			},
		},
		{
			name:      "Nil service map initialises empty map",
			setupFile: testYAML_NilServiceMap,
			validate: func(t *testing.T, config *Config, stdout string) {
				if config.Service == nil {
					t.Error("Config.Service mismatch\ngot:  nil\nwant: non-nil")
				}
			},
		},
		{
			name:      "Missing file loads with defaults",
			setupFile: nil,
			validate: func(t *testing.T, config *Config, stdout string) {
				if config.Service == nil {
					t.Error("Config.Service mismatch\ngot:  nil\nwant: non-nil")
				}

				if config.HardDefaults.Service.LatestVersion.Common.Require.Docker.Tag == "" {
					t.Error(
						"Config.HardDefaults.Service.LatestVersion.Common.Require.Docker.Tag mismatch\n" +
							"got:  \"\"\nwant: NOT \"\"",
					)
				}
				if got, want := config.Defaults.Service.LatestVersion.Common.Require.Docker.Defaults,
					&config.HardDefaults.Service.LatestVersion.Common.Require.Docker; got != want {
					t.Errorf(
						"Config.Defaults.Service.LatestVersion.Common.Require.Docker.Defaults mismatch\n"+
							"got:  %p\nwant: %p",
						got, want,
					)
				}

				// No file to learn indentation from, so the first save uses the default.
				if got := config.Settings.Indentation; got != yaml.DefaultIndentSpaces {
					t.Errorf(
						"Config.Settings.Indentation mismatch\ngot:  %d\nwant: %d",
						got, yaml.DefaultIndentSpaces,
					)
				}
			},
		},
		{
			name:      "Invalid YAML returns exit code",
			setupFile: testYAML_InvalidYAML,
			validate: func(t *testing.T, config *Config, stdout string) {
				wantRegex := `failed to decode "[^"]+"`
				if !util.RegexCheck(wantRegex, stdout) {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, wantRegex)
				}
			},
			exitCode: new(1),
		},
		{
			name: "Config that is unreadable",
			setupFile: func(t *testing.T, path string) {
				testYAML_config_test(t, path)
				if err := os.Chmod(path, 0_222); err != nil {
					t.Fatalf("os.Chmod() failed: %v", err)
				}
			},
			exitCode: new(1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars and captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			g, _ := errgroup.WithContext(t.Context())
			flags := make(map[string]bool)
			file := filepath.Join(testTempDir,
				"TestLoad_"+strings.ReplaceAll(tc.name, " ", "_")+".yml")

			if tc.setupFile != nil {
				tc.setupFile(t, file)
			}

			// Load sets the process-wide logx level from ARGUS_LOG_LEVEL, so keep the suite's DEBUG.
			env := map[string]string{"ARGUS_LOG_LEVEL": "DEBUG"}
			maps.Copy(env, tc.env)
			test.IsolateEnv(t, env)

			var config Config
			config.Load(t.Context(), g, file, &flags)

			stdout := releaseStdout()
			if tc.validate != nil {
				tc.validate(t, &config, stdout)
			}

			exitCodeChannel := logx.ExitCodeChannel()
			var exitCode *int
			var exitMsg string
			select {
			case exitMsg = <-exitCodeChannel:
				exitCode = new(1)
			default:
			}
			got := test.StringifyPtr(exitCode)
			want := test.StringifyPtr(tc.exitCode)
			if got != want {
				t.Errorf(
					"exit code mismatch\ngot:  %s\nwant: %s\nexit message: %q",
					got, want, exitMsg,
				)
			}
		})
	}
}

func TestConfig_InitDefaults(t *testing.T) {
	var cfg Config
	defaults, _ := DecodeDefaults("yaml", nil)
	cfg.Defaults = *defaults
	var hardDefaults Defaults
	cfg.HardDefaults = hardDefaults
	cfg.SaveChannel = make(chan bool, 1)

	cfg.InitDefaults()

	fieldTests := []test.FieldAssertion{
		{Name: "Defaults.Service.LatestVersion.Common.Require.Docker.Defaults", Got: cfg.Defaults.Service.LatestVersion.Common.Require.Docker.Defaults, Want: &cfg.HardDefaults.Service.LatestVersion.Common.Require.Docker, Mode: test.CompareSamePointer},
		{Name: "Defaults.Service.LatestVersion.Common.Options", Got: cfg.Defaults.Service.LatestVersion.Common.Options, Want: &cfg.Defaults.Service.Options, Mode: test.CompareSamePointer},
		{Name: "Defaults.Service.DeployedVersionLookup.Options", Got: cfg.Defaults.Service.DeployedVersionLookup.Options, Want: &cfg.Defaults.Service.Options, Mode: test.CompareSamePointer},
		{Name: "HardDefaults.Service.LatestVersion.Common.Require.Docker.Type", Got: cfg.HardDefaults.Service.LatestVersion.Common.Require.Docker.Type, Want: "", Mode: test.CompareNotEqual},
		{Name: "HardDefaults.Service.LatestVersion.Common.Options", Got: cfg.HardDefaults.Service.LatestVersion.Common.Options, Want: &cfg.HardDefaults.Service.Options, Mode: test.CompareSamePointer},
		{Name: "HardDefaults.Service.DeployedVersionLookup.Options", Got: cfg.HardDefaults.Service.DeployedVersionLookup.Options, Want: &cfg.HardDefaults.Service.Options, Mode: test.CompareSamePointer},
		{Name: "HardDefaults.Service.Status.SaveChannel", Got: cfg.HardDefaults.Service.Status.SaveChannel, Want: cfg.SaveChannel, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Config"); testErr != nil {
		t.Fatal(testErr)
	}
}
