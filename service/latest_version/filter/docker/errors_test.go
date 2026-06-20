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

package docker

import (
	"testing"
)

func TestTagNotFoundError_Error(t *testing.T) {
	// GIVEN: a TagNotFoundError.
	tests := []struct {
		name  string
		input TagNotFoundError
		want  string
	}{
		{
			name: "basic error",
			input: TagNotFoundError{
				Image: "test/app",
				Tag:   "9001",
			},
			want: "test/app:9001 - tag not found",
		},
		{
			name: "empty fields",
			input: TagNotFoundError{
				Image: "",
				Tag:   "",
			},
			want: ": - tag not found",
		},
		{
			name: "image only",
			input: TagNotFoundError{
				Image: "test/app",
				Tag:   "",
			},
			want: "test/app: - tag not found",
		},
		{
			name: "tag only",
			input: TagNotFoundError{
				Image: "",
				Tag:   "latest",
			},
			want: ":latest - tag not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Error() is called on it.
			got := tc.input.Error()

			// THEN: the expected error message is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nTagNotFoundError.Error() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}
