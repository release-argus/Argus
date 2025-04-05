// Copyright [2025] [Argus]
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

package deployedver

import (
	"testing"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
)

func TestServiceMap(t *testing.T) {
	// GIVEN a service type string.
	tests := map[string]struct {
		key      string
		expected base.Interface
	}{
		"web": {
			key:      "web",
			expected: &web.Lookup{},
		},
		"url": {
			key:      "url",
			expected: &web.Lookup{},
		},
		"manual": {
			key:      "manual",
			expected: &manual.Lookup{},
		},
		"unknown": {
			key:      "foo",
			expected: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN the service type is looked up in the ServiceMap.
			lookupFunc, exists := ServiceMap[tc.key]

			// THEN a type is returned.
			if !exists {
				// If the expected value is nil, then the key should not exist.
				if tc.expected == nil {
					return
				}
				t.Fatalf("%s\nServiceMap key %q does not exist",
					packageName, tc.key)
			}
			// And the returned type is of the expected type.
			lookup := lookupFunc()
			if getType(lookup) != getType(tc.expected) {
				t.Errorf("%s\nServiceMap[%q]() = %T, want %T",
					packageName, tc.key, lookup, tc.expected)
			}
		})
	}
}
