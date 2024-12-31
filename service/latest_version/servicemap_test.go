// Copyright [2024] [Argus]
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

package latestver

import (
	"testing"

	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
)

func TestServiceMap(t *testing.T) {
	tests := map[string]struct {
		key      string
		expected base.Interface
	}{
		"github": {
			key:      "github",
			expected: &github.Lookup{},
		},
		"web": {
			key:      "web",
			expected: &web.Lookup{},
		},
		"url": {
			key:      "url",
			expected: &web.Lookup{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lookupFunc, exists := ServiceMap[tc.key]
			if !exists {
				t.Fatalf("ServiceMap key %q does not exist", tc.key)
			}

			lookup := lookupFunc()
			if getType(lookup) != getType(tc.expected) {
				t.Errorf("ServiceMap[%q]() = %T, want %T", tc.key, lookup, tc.expected)
			}
		})
	}
}
