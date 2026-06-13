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

// Package test provides test helpers for the deployed_version package.
package test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

// MockLookup is a configurable deployed version lookup stub for tests.
type MockLookup struct {
	base.Lookup
	OverrideErr string `json:"override_err,omitempty" yaml:"override_err,omitempty"`
}

// ApplyOverrides returns OverrideErr when configured, otherwise nil.
func (f *MockLookup) ApplyOverrides(format string, data []byte) error {
	if f.OverrideErr != "" {
		return errors.New(f.OverrideErr)
	}
	return nil
}
func (f *MockLookup) Copy(*status.Status) base.Interface          { return f }
func (f *MockLookup) DecodeSelf(format string, data []byte) error { return nil }
func (f *MockLookup) GetType() string                             { return "fake" }
func (f *MockLookup) String(prefix string) string                 { return decode.ToYAMLString(f, prefix) }
func (f *MockLookup) Track()                                      {}

// Lookup decodes and validates a deployed version lookup of the given type for tests.
func Lookup(t *testing.T, typ string, fail bool, version string) (dv deployedver.Lookup) {
	dvCfg := PlainDefaultsConfig(t)

	switch typ {
	case "manual":
		dv = testManual(t, version)
	case "url":
		dv = testWeb(t, fail)
	}

	dv.Init(
		dv.GetOptions(),
		dv.GetStatus(),
		dvCfg,
	)
	dv.GetStatus().ServiceInfo.ID = "TEST_DV"

	// Check the values.
	if err := dv.CheckValues(); err != nil {
		t.Fatalf(
			"dvtest.Lookup(type=%q, fail=%t).CheckValues() unexpected error: %v",
			dv, fail, err,
		)
	}

	return dv
}

// testManual builds a manual deployed version lookup for tests.
func testManual(t *testing.T, version string) deployedver.Lookup {
	dvCfg := PlainDefaultsConfig(t)

	svcStatus, _ := statustest.New("yaml", nil)

	dv, _ := deployedver.Decode(
		"yaml", []byte(test.TrimYAML(`
			type: manual
			version: `+version+`
		`)),
		opttest.Options(t),
		svcStatus,
		dvCfg,
	)

	return dv
}

// testWeb builds a URL deployed version lookup for tests.
func testWeb(t *testing.T, fail bool) deployedver.Lookup {
	dvCfg := PlainDefaultsConfig(t)

	svcStatus, _ := statustest.New("yaml", nil)

	dv, _ := deployedver.Decode(
		"yaml", []byte(test.TrimYAML(`
			type: url
			method: GET
			url: `+test.LookupBare["url_invalid"]+`/1.2.3
			allow_invalid_certs: `+fmt.Sprint(!fail)+`
		`)),
		opttest.Options(t),
		svcStatus,
		dvCfg,
	)

	return dv
}
