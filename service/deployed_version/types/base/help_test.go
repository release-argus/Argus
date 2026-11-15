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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
)

type testLookup struct {
	Lookup `yaml:",inline" json:",embed"`
}

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig() DefaultsConfig {
	hardDefaults := &Defaults{Options: &opt.Defaults{}}
	hardDefaults.Default()
	hardDefaults.Options.Default()

	return DefaultsConfig{
		Soft: &Defaults{Options: &opt.Defaults{}},
		Hard: hardDefaults,
	}
}

type lookupImpl struct {
	Lookup `json:",embed" yaml:",inline"`
}

func (l *lookupImpl) ApplyOverrides(string, []byte) error { return nil }
func (l *lookupImpl) CheckValues() error                  { return nil }
func (l *lookupImpl) Copy(*status.Status) Interface       { return &lookupImpl{} }
func (l *lookupImpl) DecodeSelf(string, []byte) error     { return nil }
func (l *lookupImpl) GetType() string                     { return "test" }
func (l *lookupImpl) String(prefix string) string         { return decode.ToYAMLString(l, prefix) }
func (l *lookupImpl) Track()                              {}
