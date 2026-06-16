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
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestMockLookup_ApplyOverrides(t *testing.T) {
	// GIVEN: a MockLookup.
	fake := &MockLookup{}
	// WHEN: ApplyOverrides is called with empty format and data.
	err := fake.ApplyOverrides("", nil)
	// THEN: no error is returned.
	if err != nil {
		t.Errorf("MockLookup.ApplyOverrides(format=\"\", data=nil) error mismatch\ngot:  %v\nwant: nil", err)
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
		t.Errorf("MockLookup.DecodeSelf(format=\"\", data=nil) error mismatch\ngot:  %v\nwant: nil", err)
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
