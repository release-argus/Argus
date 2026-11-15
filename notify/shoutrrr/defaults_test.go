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

package shoutrrr

import (
	"fmt"
	"slices"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestShoutrrrsDefaults_Init(t *testing.T) {
	tests := []struct {
		name  string
		dflts ShoutrrrsDefaults
	}{
		{
			name:  "empty",
			dflts: ShoutrrrsDefaults{},
		},
		{
			name: "non-empty",
			dflts: ShoutrrrsDefaults{
				"foo":     &Defaults{},
				"bar":     &Defaults{},
				"charlie": &Defaults{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.dflts.Init()

			for typ, d := range tc.dflts {
				assertMapsInitialised(t, typ, d)
			}
		})
	}
}

func TestShoutrrrsDefaults_Defaults(t *testing.T) {
	var defaults ShoutrrrsDefaults

	defaults.Default()

	for _, typ := range SupportedTypes {
		if defaults[typ] == nil {
			t.Errorf("ShoutrrrsDefaults[%q] mismatch\ngot:  nil\nwant: non-nil", typ)
		}
	}

	for typ := range defaults {
		if !slices.Contains(SupportedTypes, typ) {
			t.Errorf("ShoutrrrsDefaults[%q] mismatch\ngot:  present\nwant: absent", typ)
		}
	}

	for typ, d := range defaults {
		assertMapsInitialised(t, typ, d)
	}
}

func assertMapsInitialised(t *testing.T, typ string, d *Defaults) {
	t.Helper()

	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: d.Options, Mode: test.CompareNotNil},
		{Name: "URLFields", Got: d.URLFields, Mode: test.CompareNotNil},
		{Name: "Params", Got: d.Params, Mode: test.CompareNotNil},
	}
	if testErr := test.AssertFields(t, fieldTests, fmt.Sprintf("ShoutrrrsDefaults[%q]", typ)); testErr != nil {
		t.Error(testErr)
	}
}

func TestNotifyDefaultOptions(t *testing.T) {
	got := notifyDefaultOptions()

	if len(got) == 0 {
		t.Fatal("length mismatch\ngot:  0\nwant: 1+")
	}
	for k, v := range got {
		if v == "" {
			t.Errorf("options[%q] mismatch\ngot:  \"\"\nwant: non-empty", k)
		}
	}
}
