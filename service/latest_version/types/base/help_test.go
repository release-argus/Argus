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
	logtest "github.com/release-argus/Argus/internal/test/log"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

type testLookup struct {
	Lookup `yaml:",inline" json:",embed"`
}

// decodeTestLookup returns a new testLookup from a string in a given format (json/yaml).
func decodeTestLookup(
	format string,
	data []byte,
	options *opt.Options,
	svcStatus *status.Status,
	cfg DefaultsConfig,
) (*testLookup, error) {
	field := testLookup{
		Defaults:     cfg.Soft,
		HardDefaults: cfg.Hard,
	}

	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, err
	}

	if err := UnmarshalRequire(
		format, data,
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
func plainDefaultsConfig() DefaultsConfig {
	optCfg := opttest.PlainDefaultsConfig()

	defaults, _ := DecodeDefaults("yaml", nil)
	defaults.Options = optCfg.Soft
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	hardDefaults.Options = optCfg.Hard

	defaults.Require.SetDefaults(&hardDefaults.Require)

	return DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
