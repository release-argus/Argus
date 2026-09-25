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

package command

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/status"
)

func TestString(t *testing.T) {
	// GIVEN: a Lookup.
	lookup := testLookup(t, []string{"printf", "1.2.3"}, "v([0-9.]+)")
	lookup.Status.SetDeployedVersion(
		lookup.Status.DeployedVersion(),
		time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		false,
	)
	lookup.Status.ServiceInfo.ID = t.Name()

	tests := []struct {
		name   string
		lookup *Lookup
		want   string
	}{
		{
			name: "empty",
			lookup: test.Must(t, func() (*Lookup, error) {
				l := Lookup{}
				l.Status = &status.Status{}
				l.Status.ServiceInfo.ID = "empty"
				return &l, nil
			}),
			want: "{}\n",
		},
		{
			name:   "command and regex",
			lookup: lookup,
			want: test.TrimYAML(`
				type: command
				command:
				  - printf
				  - 1.2.3
				regex: v([0-9.]+)
			`),
		},
		{
			name: "command only",
			lookup: test.Must(t, func() (*Lookup, error) {
				l := testLookup(t, []string{"printf", "1.2.3"}, "")
				return l, nil
			}),
			want: test.TrimYAML(`
				type: command
				command:
				  - printf
				  - 1.2.3
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.lookup.String,
				tc.want,
			)
		})
	}
}

func TestLookup_Copy(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name   string
		lookup *Lookup
		status *status.Status
	}{
		{
			name:   "nil",
			lookup: nil,
			status: nil,
		},
		{
			name: "filled",
			lookup: func() *Lookup {
				l := testLookup(t, []string{"sh", "-c", "printf 1.2.3"}, "([0-9.]+)")
				return l
			}(),
			status: test.Must(t, func() (*status.Status, error) {
				return status.New(nil, nil, nil, "", "", "", "", "", "", nil), nil
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wantStr := decode.ToYAMLString(tc.lookup, "")

			prefix := fmt.Sprintf(
				"%s\nLookup.Copy(status=%p)",
				packageName, tc.status,
			)

			// WHEN: Copy() is called on it.
			gotInterface := tc.lookup.Copy(tc.status)

			// THEN: if nil was copied, we get nil.
			if tc.lookup == nil {
				if gotInterface != nil {
					t.Errorf(
						"%s of nil mismatch\ngot:  %v\nwant: nil",
						prefix, gotInterface,
					)
				}
				return
			}

			// AND: the copy is non-nil.
			if gotInterface == nil {
				t.Fatalf("%s got nil want non-nil", prefix)
			}

			// AND: the copy is distinct.
			if gotInterface == tc.lookup {
				t.Fatalf(
					"%s should return a distinct copy\ngot:  %p\nwant: %p",
					prefix, gotInterface, tc.lookup,
				)
			}

			// AND: the type is unchanged.
			got, ok := gotInterface.(*Lookup)
			if !ok {
				t.Fatalf(
					"%s type shouldn't have changed\ngot:  %T\nwant: Lookup",
					prefix, gotInterface,
				)
			}

			// AND: the copy unmarshals the same.
			if gotStr := got.String(""); gotStr != wantStr {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the fields are copied as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: tc.lookup.Type, Mode: test.CompareEqual},
				{Name: "Command", Got: &got.Command, Want: &tc.lookup.Command, Mode: test.CompareDifferentPointer},
				{Name: "Regex", Got: got.Regex, Want: tc.lookup.Regex, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: the copied Command slice is not aliased.
			if len(got.Command) > 0 && len(tc.lookup.Command) > 0 {
				got.Command[0] = "mutated"
				if tc.lookup.Command[0] == got.Command[0] {
					t.Errorf(
						"%s Command slice should be a deep copy, got aliased values",
						prefix,
					)
				}
			}
		})
	}
}