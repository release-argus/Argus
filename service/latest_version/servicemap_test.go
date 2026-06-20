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

package latestver

import (
	"testing"

	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
)

func TestServiceMap(t *testing.T) {
	tests := []struct {
		key      string
		expected Lookup
	}{
		{
			key:      "github",
			expected: &github.Lookup{},
		},
		{
			key:      "web",
			expected: &web.Lookup{},
		},
		{
			key:      "url",
			expected: &web.Lookup{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			lookupFunc, exists := ServiceMap[tc.key]
			if !exists {
				t.Fatalf(
					"%s\nServiceMap[%q] does not exist",
					packageName, tc.key,
				)
			}

			lookup := lookupFunc()
			if getType(lookup) != getType(tc.expected) {
				t.Errorf(
					"%s\nServiceMap[%q]() mismatch\ngot:  %T\nwant: %T",
					packageName, tc.key,
					lookup, tc.expected,
				)
			}
		})
	}
}

func getType(lookup Lookup) string {
	switch lookup.(type) {
	case *github.Lookup:
		return "github"
	case *web.Lookup:
		return "url"
	}
	return "unknown"
}
