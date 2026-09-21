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
	"slices"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	lvforgejo "github.com/release-argus/Argus/service/latest_version/types/forgejo"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestMockLookup_ApplyOverrides(t *testing.T) {
	// GIVEN: a MockLookup.
	fake := &MockLookup{}
	// WHEN: ApplyOverrides is called with empty format and data.
	err := fake.ApplyOverrides("", nil)
	// THEN: no error is returned.
	if err != nil {
		t.Errorf(
			"MockLookup.ApplyOverrides(format=\"\", data=nil) error mismatch\ngot:  %v\nwant: nil",
			err,
		)
	}
}
func TestMockLookup_ApplyOverrides__error(t *testing.T) {
	// GIVEN: a MockLookup with an error override.
	wantErr := "TestMockLookup_ApplyOverrides_Error"
	fake := &MockLookup{OverrideErr: wantErr}
	// WHEN: ApplyOverrides is called with empty format and data.
	err := fake.ApplyOverrides("", nil)
	e := errfmt.FormatError(err)
	if e != wantErr {
		t.Errorf(
			"MockLookup.ApplyOverrides(format=\"\", data=nil) error mismatch with OverrideErr set\ngot:  %q\nwant: %q",
			e, wantErr,
		)
	}
}
func TestMockLookup_Copy(t *testing.T) {
	// GIVEN: a MockLookup.
	fake := &MockLookup{}
	// WHEN: Copy is called with true.
	got := fake.Copy(fake.GetStatus())
	// THEN: the returned value is always non-nil.
	if got == nil {
		t.Errorf("MockLookup.Copy() result mismatch\ngot:  %v\nwant: non-nil",
			got,
		)
	}
}
func TestMockLookup_DecodeSelf(t *testing.T) {
	// GIVEN: a MockLookup.
	fake := &MockLookup{}
	// WHEN: DecodeSelf is called with empty format and data.
	err := fake.DecodeSelf("", nil)
	// THEN: no error is returned.
	if err != nil {
		t.Errorf(
			"MockLookup.DecodeSelf(format=\"\", data=nil) error mismatch\ngot:  %v\nwant: nil",
			err,
		)
	}
}
func TestMockLookup_Require(t *testing.T) {
	// GIVEN: a MockLookup and a Require.
	fake := &MockLookup{}
	req := &filter.Require{}
	// WHEN: SetRequire is called with a Require.
	fake.SetRequire(req)
	// THEN: the Require is set and can be retrieved.
	if req != fake.Require {
		t.Errorf(
			"MockLookup.SetRequire(%p) .Require pointer mismatch\ngot:  %p\nwant: %p",
			req, fake.Require, req,
		)
	}
	if got := fake.GetRequire(); got != req {
		t.Errorf(
			"MockLookup.GetRequire() pointer mismatch\ngot:  %p\nwant: %p",
			got, req,
		)
	}
}
func TestMockLookup_String(t *testing.T) {
	// GIVEN: a MockLookup with an error override.
	fake := &MockLookup{OverrideErr: "foo"}
	want := "lookup: {}\noverride_err: foo\n"
	// WHEN: String is called with an empty prefix.
	got := fake.String("")
	// THEN: the returned value is the expected string.
	if got != want {
		t.Errorf(
			"MockLookup.String() value mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
}

func TestLookupBuilder(t *testing.T) {
	// GIVEN: a lookup type name.
	tests := []struct {
		name      string
		typ       string
		wantBuild bool
		wantErrs  []string
	}{
		{
			name:      "valid/forgejo",
			typ:       lvforgejo.Type,
			wantBuild: true,
		},
		{
			name:      "valid/github",
			typ:       lvgithub.Type,
			wantBuild: true,
		},
		{
			name:      "valid/url",
			typ:       lvweb.Type,
			wantBuild: true,
		},
		{
			name:     "invalid/unknown type",
			typ:      "something",
			wantErrs: []string{`lvtest.Lookup: unsupported type "something"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var fakeT test.FakeT

			// WHEN: lookupBuilder is asked for that type's builder.
			got := lookupBuilder(&fakeT, tc.typ)

			prefix := fmt.Sprintf(
				"%s\nlookupBuilder(type=%q)",
				packageName, tc.typ,
			)

			// THEN: a builder comes back only for a known type.
			if gotBuild := got != nil; gotBuild != tc.wantBuild {
				t.Errorf(
					"%s builder mismatch\ngot:  %t\nwant: %t",
					prefix, gotBuild, tc.wantBuild,
				)
			}

			// AND: only unsupported types are reported.
			if !slices.Equal(fakeT.Errors, tc.wantErrs) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, fakeT.Errors, tc.wantErrs,
				)
			}
		})
	}
}

func TestLookup(t *testing.T) {
	// GIVEN: a lookup type the fixture builds, and whether it should be built to fail.
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

			// WHEN: Lookup is called for that type.
			lv := Lookup(t, tc.typ, tc.fail)

			prefix := fmt.Sprintf(
				"%s\nLookup(type=%q, fail=%t)",
				packageName, tc.typ, tc.fail,
			)

			// THEN: the lookup is of the concrete type that name addresses.
			if got := fmt.Sprintf("%T", lv); got != tc.wantType {
				t.Fatalf(
					"%s type mismatch\ngot:  %s\nwant: %s",
					prefix, got, tc.wantType,
				)
			}

			// AND: it reports the type it was asked for.
			if got := lv.GetType(); got != tc.typ {
				t.Errorf(
					"%s GetType() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.typ,
				)
			}

			// AND: it addresses the source the fixture describes.
			if got := lv.ServiceURL(); got != tc.wantServiceURL {
				t.Errorf(
					"%s ServiceURL() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantServiceURL,
				)
			}

			// AND: the status carries the fixture's service ID, for tests that key on it.
			if got, want := lv.GetStatus().ServiceInfo.ID, "TEST_LV"; got != want {
				t.Errorf(
					"%s service ID mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}
