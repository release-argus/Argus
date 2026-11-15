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
	"reflect"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/types/forgejo"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
)

func TestServiceMap(t *testing.T) {
	tests := []struct {
		name     string
		expected Lookup
	}{
		{
			name:     "forgejo",
			expected: &forgejo.Lookup{},
		},
		{
			name:     "github",
			expected: &github.Lookup{},
		},
		{
			name:     "web", // An alias of url.
			expected: &web.Lookup{},
		},
		{
			name:     "url",
			expected: &web.Lookup{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookupFunc, exists := ServiceMap[tc.name]
			if !exists {
				t.Fatal("ServiceMap mismatch\ngot:  absent\nwant: present")
			}

			lookup := lookupFunc()
			if reflect.TypeOf(lookup) != reflect.TypeOf(tc.expected) {
				t.Errorf("ServiceMap constructor mismatch\ngot:  %T\nwant: %T", lookup, tc.expected)
			}
		})
	}
}
