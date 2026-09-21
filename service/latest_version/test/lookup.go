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

//go:build unit || integration

// Package test provides test helpers for the latest_version package.
package test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	lvforgejo "github.com/release-argus/Argus/service/latest_version/types/forgejo"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

// MockLookup is a configurable latest version lookup stub for tests.
type MockLookup struct {
	base.Lookup
	OverrideErr string `json:"override_err,omitzero" yaml:"override_err,omitzero"`
}

func (f *MockLookup) ApplyOverrides(format string, data []byte) error {
	if f.OverrideErr != "" {
		return errors.New(f.OverrideErr)
	}
	return nil
}
func (f *MockLookup) Copy(*status.Status) base.Interface          { return f }
func (f *MockLookup) DecodeSelf(format string, data []byte) error { return nil }
func (f *MockLookup) GetRequire() *filter.Require                 { return f.Require }
func (f *MockLookup) SetRequire(r *filter.Require)                { f.Require = r }
func (f *MockLookup) String(prefix string) string                 { return decode.ToYAMLString(f, prefix) }

// testingT is the fragments of [testing.T] the type guard needs.
type testingT interface {
	Helper()
	Fatalf(format string, args ...any)
}

// lookupBuilder returns the fixture builder for a lookup type, or nil after
// reporting a type no fixture builds.
func lookupBuilder(t testingT, typ string) func(*testing.T, bool) latestver.Lookup {
	t.Helper()

	switch typ {
	case lvforgejo.Type:
		return testForgejo
	case lvgithub.Type:
		return testGitHub
	case lvweb.Type:
		return testWeb
	}

	t.Fatalf("lvtest.Lookup: unsupported type %q", typ)
	return nil
}

func Lookup(t *testing.T, typ string, fail bool) (lv latestver.Lookup) {
	t.Helper()

	lv = lookupBuilder(t, typ)(t, fail)

	lv.GetStatus().ServiceInfo.ID = "TEST_LV"

	// Check the values.
	if err := lv.CheckValues(); err != nil {
		t.Fatalf(
			"lvtest.Lookup(type=%q, fail=%t).CheckValues() unexpected error: %v",
			typ, fail, err,
		)
	}

	return lv
}

// testForgejo builds a Forgejo latest version lookup for tests.
func testForgejo(t *testing.T, fail bool) latestver.Lookup {
	t.Helper()

	repo := "argus/releases-happy"
	if fail {
		repo = "argus/no-such-repository"
	}

	svcStatus, _ := statustest.New("yaml", nil)

	lv, _ := latestver.Decode(
		"yaml", []byte(test.TrimYAML(`
			type: forgejo
			host: `+test.ValidCertHTTPS+`/forgejo
			url: `+repo+`
		`)),
		opttest.Options(t),
		svcStatus,
		PlainDefaultsConfig(t),
	)

	return lv
}

// testGitHub builds a GitHub latest version lookup for tests.
func testGitHub(t *testing.T, fail bool) latestver.Lookup {
	t.Helper()

	lvCfg := PlainDefaultsConfig(t)
	accessToken := test.GitHubToken(t)
	if fail {
		accessToken = "invalid"
	}

	svcStatus, _ := statustest.New("yaml", nil)

	lv, _ := latestver.Decode(
		"yaml", []byte(test.TrimYAML(`
			type: github
			url: `+test.ArgusGitHubRepo+`
			access_token: `+accessToken+`
		`)),
		opttest.Options(t),
		svcStatus,
		lvCfg,
	)

	return lv
}

// testWeb builds a URL latest version lookup for tests.
func testWeb(t *testing.T, fail bool) latestver.Lookup {
	t.Helper()

	lvCfg := PlainDefaultsConfig(t)

	svcStatus, _ := statustest.New("yaml", nil)

	lv, _ := latestver.Decode(
		"yaml", []byte(test.TrimYAML(`
			type: url
			url: `+test.LookupBare.URLInvalid+`/1.2.3
			allow_invalid_certs: `+fmt.Sprint(!fail)+`
		`)),
		opttest.Options(t),
		svcStatus,
		lvCfg,
	)

	return lv
}
