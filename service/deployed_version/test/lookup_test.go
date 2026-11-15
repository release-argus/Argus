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
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
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
	if e := errfmt.FormatError(err); e != wantErr {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, wantErr)
	}
}

func TestMockLookup_Copy(t *testing.T) {
	fake := &MockLookup{}

	got := fake.Copy(fake.Status)

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

func TestMockLookup_GetType(t *testing.T) {
	fake := &MockLookup{}

	got := fake.GetType()

	if want := "fake"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestMockLookup_String(t *testing.T) {
	fake := &MockLookup{}

	got := fake.String("")

	want := "lookup: {}\n"
	if got != want {
		t.Errorf("MockLookup.String() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestMockLookup_Track(t *testing.T) {
	fake := &MockLookup{}

	start := time.Now()
	fake.Track()

	if since := time.Since(start); since > time.Second {
		t.Errorf("MockLookup.Track() duration mismatch\ngot:  %s\nwant: <1s", since)
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		name       string
		typ        string
		fail       bool
		version    string
		wantType   string
		wantFatals []string
	}{
		{
			name:     "manual/passing",
			typ:      "manual",
			version:  "1.2.3",
			wantType: "*manual.Lookup",
		},
		{
			name:     "url/passing",
			typ:      "url",
			wantType: "*web.Lookup",
		},
		{
			name:     "url/failing, certificate trust",
			typ:      "url",
			fail:     true,
			wantType: "*web.Lookup",
		},
		{
			name:       "invalid/unknown type",
			typ:        "something",
			wantFatals: []string{`^unsupported lookup type "something"$`},
		},
		{
			name:       "invalid/manual version not semantic",
			typ:        "manual",
			version:    "not-semver",
			wantFatals: []string{`^Lookup\(type="manual", fail=false\)\.CheckValues\(\) failed: .+`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fakeT := &test.FakeT{Abort: true}

			var dv any
			fakeT.Aborted(func() { dv = Lookup(fakeT, tc.typ, tc.fail, tc.version) })

			if testErr := test.AssertSlicesEqualFunc(
				t, fakeT.Fatals, tc.wantFatals,
				func(got, want string) bool { return util.RegexCheck(want, got) },
				"FakeT.Fatals",
			); testErr != nil {
				t.Fatal(testErr)
			}
			if tc.wantType == "" {
				return
			}

			if got := fmt.Sprintf("%T", dv); got != tc.wantType {
				t.Errorf("type mismatch\ngot:  %q\nwant: %q", got, tc.wantType)
			}
		})
	}
}
