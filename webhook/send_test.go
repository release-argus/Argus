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

package webhook

import (
	"testing"
)

func TestCheckWebHookBody(t *testing.T) {
	// GIVEN: a response body.
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "empty body",
			body: "",
			want: true,
		},
		{
			name: "success body",
			body: "success",
			want: true,
		},
		{
			name: "awx invalid secret",
			body: `{"detail":"You do not have permission to perform this action."}`,
			want: false,
		},
		{
			name: "adnanh-webhook - defaults hook fail",
			body: `Hook rules were not satisfied.`,
			want: false,
		},
		{
			name: "case insensitive body message fail",
			body: `hook rULEs wEre nOt SATISFiED.`,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: checkWebHookBody is called on it.
			got := checkWebHookBody(tc.body)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\ncheckWebHookBody(%q) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.body,
					got, tc.want,
				)
			}
		})
	}
}
