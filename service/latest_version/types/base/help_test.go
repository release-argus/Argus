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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

var packageName = "lvbase"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

type testLookup struct {
	Lookup `yaml:",inline" json:",inline"`
}

// decodeTestLookup returns a new testLookup from a string in a given format (json/yaml).
func decodeTestLookup(
	t *testing.T,
	format string, // "json" | "yaml"
	data []byte,
	options *opt.Options,
	svcStatus *status.Status,
	cfg DefaultsConfig,
) (*testLookup, error) {
	t.Helper()

	field := testLookup{
		Lookup: Lookup{
			Defaults:     cfg.Soft,
			HardDefaults: cfg.Hard,
		},
	}

	// Unmarshal static fields.
	if err := decode.Unmarshal(format, []byte(data), &field); err != nil {
		return nil, err
	}

	// Require.
	if err := UnmarshalRequire(
		format, []byte(data),
		&field,
		svcStatus,
		&cfg.Soft.Require,
	); err != nil {
		return nil, err
	}

	field.Init(
		options,
		svcStatus,
		cfg,
	)

	return &field, nil
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig(t *testing.T) DefaultsConfig {
	t.Helper()

	optDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults.Default()

	defaults, _ := DecodeDefaults("yaml", nil)
	defaults.Options = optDefaults
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	hardDefaults.AccessToken = test.GitHubToken(t)
	hardDefaults.Options = optHardDefaults

	defaults.Require.SetDefaults(&hardDefaults.Require)

	return DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
