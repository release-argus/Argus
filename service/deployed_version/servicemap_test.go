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

package deployedver

import (
	"testing"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
)

func TestServiceMap(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected base.Interface
	}{
		{
			name:     "web",
			key:      "web",
			expected: &web.Lookup{},
		},
		{
			name:     "url",
			key:      "url",
			expected: &web.Lookup{},
		},
		{
			name:     "manual",
			key:      "manual",
			expected: &manual.Lookup{},
		},
		{
			name:     "unknown",
			key:      "foo",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookupFunc, exists := ServiceMap[tc.key]

			if !exists {
				if tc.expected == nil {
					return
				}
				t.Fatal("ServiceMap key mismatch\ngot:  absent\nwant: present")
			}

			lookup := lookupFunc()
			if getType(lookup) != getType(tc.expected) {
				t.Errorf("type mismatch\ngot:  %T\nwant: %T", lookup, tc.expected)
			}
		})
	}
}
