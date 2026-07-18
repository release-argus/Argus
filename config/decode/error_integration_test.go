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

package decode

import (
	"errors"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestErrField__formatError(t *testing.T) {
	// GIVEN: a ErrKeyField.
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "single error",
			err: &ErrKeyField{
				Key: "type",
			},
			want: "type: <nil>",
		},
		{
			name: "single wrap",
			err: &ErrKeyField{
				Key: "another_one",
				Err: &ErrKeyField{
					Key: "and_another_one",
				},
			},
			want: test.TrimYAML(`
				another_one:
					and_another_one: <nil>`,
			),
		},
		{
			name: "double ErrKeyField wrap",
			err: &ErrKeyField{
				Key: "another_one",
				Err: &ErrKeyField{
					Key: "and_another_one",
					Err: &ErrKeyField{
						Key: "and_another_one",
					},
				},
			},
			want: test.TrimYAML(`
				another_one:
					and_another_one:
						and_another_one: <nil>`,
			),
		},
		{
			name: "ErrKeyField wrapping ErrField",
			err: &ErrKeyField{
				Key: "hello",
				Err: &ErrField{
					Key:         "type",
					Value:       "foo",
					Description: "description goes here",
				},
			},
			want: test.TrimYAML(`
				hello:
				  type: "foo" <invalid> (description goes here)`,
			),
		},
		{
			name: "ErrKeyField wrapping joined ErrField's",
			err: &ErrKeyField{
				Key: "hello",
				Err: errors.Join(
					&ErrField{
						Key:         "type",
						Value:       "mysql",
						Description: "description goes here",
					},
					&ErrField{
						Key:         "field",
						Value:       "val",
						Description: "DESCRIPTION",
					},
					&ErrField{
						Key: "other_field",
					},
				),
			},
			want: test.TrimYAML(`
				hello:
				  type: "mysql" <invalid> (description goes here)
				  field: "val" <invalid> (DESCRIPTION)
				  other_field: <required>`,
			),
		},
		{
			name: "joined ErrKeyField's wrapping joined ErrField's",
			err: errors.Join(
				&ErrKeyField{
					Key: "hello",
					Err: errors.Join(
						&ErrField{
							Key:         "field",
							Value:       "val",
							Description: "DESCRIPTION",
						},
						&ErrField{
							Key: "other_field",
						},
					),
				},
				&ErrKeyField{
					Key: "there",
					Err: errors.Join(
						&ErrField{
							Key: "foo",
						},
					),
				},
			),
			want: test.TrimYAML(`
				hello:
				  field: "val" <invalid> (DESCRIPTION)
				  other_field: <required>
				there:
				  foo: <required>`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: FormatError is called with it.
			got := errfmt.FormatError(tc.err)

			// THEN: the error is formatted as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nFormatError mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}
