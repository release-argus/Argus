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
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestShoutrrrsDefaults_Init(t *testing.T) {
	// GIVEN: ShoutrrrsDefaults.
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
				"foo": NewDefaults(
					"discord",
					nil, nil, nil,
				),
				"bar": NewDefaults(
					"discord",
					nil, nil, nil,
				),
				"charlie": NewDefaults(
					"discord",
					nil, nil, nil,
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Init() is called on it.
			tc.dflts.Init()

			// THEN: all options, url_fields, and params are non-nil.
			for typ, d := range tc.dflts {
				if d.Options == nil || d.URLFields == nil || d.Params == nil {
					t.Errorf(
						"%s didn't initialise maps for %q (%+v)",
						packageName, typ, d,
					)
				}
			}
		})
	}
}

func TestShoutrrrsDefaults_Defaults(t *testing.T) {
	// GIVEN: ShoutrrrsDefaults.
	var defaults ShoutrrrsDefaults

	// WHEN: Default() is called on it.
	defaults.Default()

	prefix := fmt.Sprintf("%s\nShoutrrrsDefaults.Default()", packageName)

	// THEN: the defaults is given keys of all shoutrrr types.
	for _, typ := range supportedTypes {
		if defaults[typ] == nil {
			t.Errorf(
				"%s didn't set defaults for %q",
				prefix, typ,
			)
		}
	}

	// AND: no unexpected types are initialised.
	for typ := range defaults {
		if !util.Contains(supportedTypes, typ) {
			t.Errorf(
				"%s initialised an unexpected notify type: %q",
				prefix, typ,
			)
		}
	}

	// AND: all options, url_fields, and params are non-nil.
	for typ, d := range defaults {
		if d.Options == nil || d.URLFields == nil || d.Params == nil {
			t.Errorf(
				"%s didn't initialise the options/url_fields/params for %q correctly:\ngot: %+v",
				prefix, typ, d,
			)
		}
	}
}

func TestNotifyDefaultOptions(t *testing.T) {
	// WHEN: notifyDefaultOptions() is called.
	got := notifyDefaultOptions()

	prefix := fmt.Sprintf("%s\nnotifyDefaultOptions()", packageName)

	// THEN: we get a map[string]string with some keys/values.
	if len(got) == 0 {
		t.Fatalf("%s gave 0 keys. want 1+", prefix)
	}
	for k, v := range got {
		if v == "" {
			t.Errorf(
				"%s gave empty value for key %q. want non-empty",
				prefix, k,
			)
		}
	}
}
