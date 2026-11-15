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

package manual

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
)

func TestLookup_CheckValues(t *testing.T) {
	tests := []struct {
		name               string
		version            string
		semanticVersioning bool
		errRegex           string
	}{
		{
			name:               "want semantic, valid",
			version:            "1.2.3",
			semanticVersioning: true,
			errRegex:           `^$`,
		},
		{
			name:               "want semantic, fail",
			version:            "1_2_3",
			semanticVersioning: true,
			errRegex:           `^failed to convert "[^"]+" to a semantic version`,
		},
		{
			name:               "want non-semantic, valid",
			version:            "1_2_3",
			semanticVersioning: false,
			errRegex:           `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaultLookup := testLookup(t, "")
			input, err := Decode("yaml", []byte("version: ''"),
				defaultLookup.Options,
				defaultLookup.Status,
				base.DefaultsConfig{
					Soft: defaultLookup.Defaults,
					Hard: defaultLookup.HardDefaults,
				},
			)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}
			defaultLookup.Options.SemanticVersioning = &tc.semanticVersioning
			input.Version = tc.version

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				input.CheckValues,
			)

			if got := input.Status.DeployedVersion(); got != "" {
				t.Errorf("Status.DeployedVersion() mismatch\ngot:  %q\nwant: \"\"", got)
			}

			if got, want := len(input.Status.AnnounceChannel), 0; got != want {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
			if got, want := len(input.Status.DatabaseChannel), 0; got != want {
				t.Errorf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
		})
	}
}
