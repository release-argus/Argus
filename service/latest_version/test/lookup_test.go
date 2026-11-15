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

package test

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	lvforgejo "github.com/release-argus/Argus/service/latest_version/types/forgejo"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestMockLookup_ApplyOverrides(t *testing.T) {
	fake := &MockLookup{}

	err := fake.ApplyOverrides("", nil)
	if err != nil {
		t.Errorf("error mismatch\ngot:  %v\nwant: nil", err)
	}
}
func TestMockLookup_ApplyOverrides__error(t *testing.T) {
	wantErr := "TestMockLookup_ApplyOverrides_Error"
	fake := &MockLookup{OverrideErr: wantErr}

	err := fake.ApplyOverrides("", nil)
	e := errfmt.FormatError(err)
	if e != wantErr {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, wantErr)
	}
}
func TestMockLookup_Copy(t *testing.T) {
	fake := &MockLookup{}

	got := fake.Copy(fake.GetStatus())

	if got == nil {
		t.Error("mismatch\ngot:  nil\nwant: non-nil")
	}
}
func TestMockLookup_DecodeSelf(t *testing.T) {
	fake := &MockLookup{}

	err := fake.DecodeSelf("", nil)
	if err != nil {
		t.Errorf("error mismatch\ngot:  %v\nwant: nil", err)
	}
}
func TestMockLookup_Require(t *testing.T) {
	fake := &MockLookup{}
	req := &filter.Require{}

	fake.SetRequire(req)

	if req != fake.Require {
		t.Errorf("MockLookup.Require mismatch\ngot:  %p\nwant: %p", fake.Require, req)
	}
	if got := fake.GetRequire(); got != req {
		t.Errorf("MockLookup.GetRequire() mismatch\ngot:  %p\nwant: %p", got, req)
	}
}
func TestMockLookup_String(t *testing.T) {
	fake := &MockLookup{OverrideErr: "foo"}
	want := "lookup: {}\noverride_err: foo\n"

	got := fake.String("")

	if got != want {
		t.Errorf("MockLookup.String() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestLookupBuilder(t *testing.T) {
	tests := []struct {
		name        string
		typ         string
		wantBuilder func(*testing.T, bool) latestver.Lookup
		wantFatals  []string
	}{
		{
			name:        "valid/forgejo",
			typ:         lvforgejo.Type,
			wantBuilder: testForgejo,
		},
		{
			name:        "valid/github",
			typ:         lvgithub.Type,
			wantBuilder: testGitHub,
		},
		{
			name:        "valid/url",
			typ:         lvweb.Type,
			wantBuilder: testWeb,
		},
		{
			name:       "invalid/unknown type",
			typ:        "something",
			wantFatals: []string{`unsupported lookup type "something"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var fakeT test.FakeT

			got := lookupBuilder(&fakeT, tc.typ)

			gotPC, wantPC := reflect.ValueOf(got).Pointer(), reflect.ValueOf(tc.wantBuilder).Pointer()
			if gotPC != wantPC {
				t.Errorf(
					"lookupBuilder() mismatch\ngot:  %q\nwant: %q",
					runtime.FuncForPC(gotPC).Name(), runtime.FuncForPC(wantPC).Name(),
				)
			}

			if !slices.Equal(fakeT.Fatals, tc.wantFatals) {
				t.Fatalf("FakeT.Fatals mismatch\ngot:  %q\nwant: %q", fakeT.Fatals, tc.wantFatals)
			}
		})
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		name           string
		typ            string
		fail           bool
		wantType       string
		wantServiceURL string
	}{
		{
			name:           "forgejo/passing",
			typ:            lvforgejo.Type,
			wantType:       "*forgejo.Lookup",
			wantServiceURL: test.ValidCertHTTPS + "/forgejo/argus/releases-happy",
		},
		{
			name:           "forgejo/failing, unknown repository",
			typ:            lvforgejo.Type,
			fail:           true,
			wantType:       "*forgejo.Lookup",
			wantServiceURL: test.ValidCertHTTPS + "/forgejo/argus/no-such-repository",
		},
		{
			name:           "github/passing",
			typ:            lvgithub.Type,
			wantType:       "*github.Lookup",
			wantServiceURL: "https://github.com/" + test.ArgusGitHubRepo,
		},
		{
			name:           "github/failing, invalid token",
			typ:            lvgithub.Type,
			fail:           true,
			wantType:       "*github.Lookup",
			wantServiceURL: "https://github.com/" + test.ArgusGitHubRepo,
		},
		{
			name:           "url/passing",
			typ:            lvweb.Type,
			wantType:       "*web.Lookup",
			wantServiceURL: test.LookupBare.URLInvalid + "/1.2.3",
		},
		{
			name:           "url/failing, certificate trust",
			typ:            lvweb.Type,
			fail:           true,
			wantType:       "*web.Lookup",
			wantServiceURL: test.LookupBare.URLInvalid + "/1.2.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lv := Lookup(t, tc.typ, tc.fail)

			if got := fmt.Sprintf("%T", lv); got != tc.wantType {
				t.Fatalf("type mismatch\ngot:  %s\nwant: %s", got, tc.wantType)
			}

			if got := lv.GetType(); got != tc.typ {
				t.Errorf("Lookup.GetType() mismatch\ngot:  %q\nwant: %q", got, tc.typ)
			}

			if got := lv.ServiceURL(); got != tc.wantServiceURL {
				t.Errorf("Lookup.ServiceURL() mismatch\ngot:  %q\nwant: %q", got, tc.wantServiceURL)
			}

			if got, want := lv.GetStatus().ServiceInfo.ID, "TEST_LV"; got != want {
				t.Errorf(
					"Lookup.GetStatus().ServiceInfo.ID mismatch\ngot:  %q\nwant: %q",
					got, want,
				)
			}
		})
	}
}
