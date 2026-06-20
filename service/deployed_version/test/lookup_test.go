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
	"time"

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
	got := fake.Copy(fake.Status)
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
func TestMockLookup_GetType(t *testing.T) {
	// GIVEN: a MockLookup.
	fake := &MockLookup{}
	// WHEN: GetType is called.
	got := fake.GetType()
	// THEN: the returned value is "fake".
	if got != "fake" {
		t.Errorf("MockLookup.GetType() value mismatch\ngot:  %q\nwant: \"fake\"", got)
	}
}
func TestMockLookup_String(t *testing.T) {
	// GIVEN: a MockLookup.
	fake := &MockLookup{}
	// WHEN: String is called with empty prefix.
	got := fake.String("")
	// THEN: the returned value is "MockLookup".
	want := "lookup: {}\n"
	if got != want {
		t.Errorf(
			"MockLookup.String() value mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
}
func TestMockLookup_Track(t *testing.T) {
	// GIVEN: a MockLookup.
	fake := &MockLookup{}
	// WHEN: Track is called.
	start := time.Now()
	fake.Track()
	// THEN: no panic occurs and the method completes successfully.
	if since := time.Since(start); since > time.Second {
		t.Errorf("MockLookup.Track() didn't complete within 1 second\ntook: %v", since)
	}
}
