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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestConfig_Load(t *testing.T) {
	prefix := fmt.Sprintf("%q\nConfig.Load()", packageName)
	DebounceDuration = time.Second
	// GIVEN: a test file to load.
	tests := []struct {
		name        string
		setupFile   func(path string)
		env         map[string]string
		validate    func(t *testing.T, config *Config, stdout string)
		exitCode    *int
		stdoutRegex string
	}{
		{
			name:      "Environment variables loaded",
			setupFile: testYAML_config_test,
			env: map[string]string{
				"TEST_ENV_KEY": "1234",
			},
			validate: func(t *testing.T, config *Config, stdout string) {
				fieldTests := []test.FieldAssertion{
					{
						Name: "Service.Options.Interval",
						Got:  config.Defaults.Service.Options.Interval,
						Want: "123s",
						Mode: test.CompareEqual,
					},
					{
						Name: "Notify.slack.title",
						Got:  config.Defaults.Notify["slack"].GetParam("title"),
						Want: "defaultTitle",
						Mode: test.CompareEqual,
					},
					{
						Name: "WebHook.Delay",
						Got:  config.Defaults.WebHook.Delay,
						Want: "2s",
						Mode: test.CompareEqual,
					},
					{
						Name: "EmptyService.String('')",
						Got:  config.Service["EmptyService"].String(""),
						Want: "",
						Mode: test.CompareEqual,
					},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "Config"); err != nil {
					t.Fatal(err)
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
							"%s should didn't delete nil Service %q",
							prefix, name,
						)
					}
				}
				if len(config.Service) != 2 {
					t.Errorf(
						"%s got %d services. Expected 2",
						prefix, len(config.Service),
					)
				}
			},
		},
		{
			name:      "Defaults assigned to services",
			setupFile: testYAML_config_test,
			validate: func(t *testing.T, config *Config, stdout string) {
				got := config.Service["WantDefaults"].Options.GetSemanticVersioning()
				if got != false {
					t.Errorf(
						"%s Service.X.Options.SemanticVersioning value mismatch\ngot:  %t\nwant: false",
						prefix, *config.Service["WantDefaults"].Options.SemanticVersioning,
					)
				}
			},
		},
		{
			name:      "Nil service map initialises empty map",
			setupFile: testYAML_NilServiceMap,
			validate: func(t *testing.T, config *Config, stdout string) {
				if config.Service == nil {
					t.Errorf("%s got a nil config.Service, want non-nil", prefix)
				}
			},
		},
		{
			name:      "Invalid YAML returns exit code",
			setupFile: testYAML_InvalidYAML,
			validate: func(t *testing.T, config *Config, stdout string) {
				wantRegex := `Unmarshal of "[^"]+" failed`
				if !util.RegexCheck(wantRegex, stdout) {
					t.Errorf(
						"%s stdout mismatch:\ngot:  %q\nwant: %q",
						packageName, stdout, wantRegex,
					)
				}
			},
			exitCode: test.Ptr(1),
		},
		{
			name: "Config that is unreadable",
			setupFile: func(path string) {
				testYAML_config_test(path)
				_ = os.Chmod(path, 0_222)
			},
			exitCode: test.Ptr(1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout and sharing log exitCodeChannel.
			releaseStdout := test.CaptureLog(t, logx.Default())

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
			test.SetEnv(t, tc.env)

			// WHEN: we Load this config.
			var config Config
			config.Load(t.Context(), g, file, &flags)

			stdout := releaseStdout()
			// Per-test validation.
			if tc.validate != nil {
				tc.validate(t, &config, stdout)
			}

			// THEN: the exit code is as expected.
			exitCodeChannel := logx.ExitCodeChannel()
			var exitCode *int
			select {
			case msg := <-exitCodeChannel:
				t.Logf("%s\n%s", packageName, msg)
				exitCode = test.Ptr(1)
			default:
			}
			got := test.StringifyPtr(tc.exitCode)
			want := test.StringifyPtr(exitCode)
			if got != want {
				t.Errorf(
					"%s unexpected exit code\ngot:  %s\nwant: %s",
					packageName, got, want,
				)
			}
		})
	}
}

func TestConfig_InitDefaults(t *testing.T) {
	// GIVEN: a Config.= with defined (empty) Defaults.
	var cfg Config
	defaults, _ := DecodeDefaults("yaml", nil)
	cfg.Defaults = *defaults
	var hardDefaults Defaults
	cfg.HardDefaults = hardDefaults

	// WHEN: we InitDefaults().
	cfg.InitDefaults()

	prefix := fmt.Sprintf("%s\nConfig InitDefaults()", packageName)

	// THEN: HardDefaults...Docker.Type has a value.
	if cfg.HardDefaults.Service.LatestVersion.Require.Docker.Type == "" {
		t.Fatalf("%s got HardDefaults.Docker.Require=''. want value", packageName)
	}

	// AND: Defaults inherit from HardDefaults.
	if cfg.Defaults.Service.LatestVersion.Require.Docker.Defaults !=
		&cfg.HardDefaults.Service.LatestVersion.Require.Docker {
		t.Fatalf("%s got Defaults....Docker.Defaults != HardDefaults...Docker.Defaults", packageName)
	}

	// AND: Options pointers are shared correctly.
	fieldTests := []test.FieldAssertion{
		{
			Name: "Defaults: Service.LatestVersion.Options -> Service.Options",
			Got:  cfg.Defaults.Service.LatestVersion.Options,
			Want: &cfg.Defaults.Service.Options,
			Mode: test.CompareSamePointer,
		},
		{
			Name: "Defaults: Service.DeployedVersionLookup.Options -> Service.Options",
			Got:  cfg.Defaults.Service.DeployedVersionLookup.Options,
			Want: &cfg.Defaults.Service.Options,
			Mode: test.CompareSamePointer,
		},
		{
			Name: "HardDefaults: Service.LatestVersion.Options -> Service.Options",
			Got:  cfg.HardDefaults.Service.LatestVersion.Options,
			Want: &cfg.HardDefaults.Service.Options,
			Mode: test.CompareSamePointer,
		},
		{
			Name: "HardDefaults: Service.DeployedVersionLookup.Options -> Service.Options",
			Got:  cfg.HardDefaults.Service.DeployedVersionLookup.Options,
			Want: &cfg.HardDefaults.Service.Options,
			Mode: test.CompareSamePointer,
		},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "Config"); err != nil {
		t.Fatal(err)
	}

	// AND: SaveChannel propagated to HardDefaults.Status.
	if cfg.HardDefaults.Service.Status.SaveChannel != cfg.SaveChannel {
		t.Fatalf(
			"%s\nConfig InitDefaults() wanted SaveChannel to be propagated to HardDefaults.Status",
			packageName,
		)
	}
}
