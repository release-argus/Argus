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

package provider

import (
	"testing"
)

func TestRegistry_Register_and_Get(t *testing.T) {
	tests := []struct {
		name       string
		register   []string
		getKnown   string
		getUnknown string
	}{
		{
			name:       "empty registry",
			register:   []string{},
			getUnknown: "local",
		},
		{
			name:       "single provider",
			register:   []string{"local"},
			getKnown:   "local",
			getUnknown: "unregistered",
		},
		{
			name:     "multiple providers",
			register: []string{"local", "provider-a", "provider-b"},
			getKnown: "provider-a",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			registry := NewRegistry()

			for _, providerName := range tc.register {
				registry.Register(&fakeProvider{name: providerName})
			}

			if tc.getKnown != "" {
				got := registry.Get(tc.getKnown)
				if got == nil {
					t.Error("Registry.Get() mismatch\ngot:  nil\nwant: non-nil")
				} else if got.Name() != tc.getKnown {
					t.Errorf(
						"Registry.Get().Name() mismatch\ngot:  %q\nwant: %q",
						got.Name(), tc.getKnown,
					)
				}
			}

			if tc.getUnknown != "" {
				if got := registry.Get(tc.getUnknown); got != nil {
					t.Errorf(
						"Registry.Get(unknown) mismatch\ngot:  Name()=%q\nwant: nil",
						got.Name(),
					)
				}
			}
		})
	}
}

func TestRegistry_Register__duplicateNamePanics(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakeProvider{name: "local"})

	defer func() {
		if recover() == nil {
			t.Error("recover() mismatch\ngot:  nil\nwant: non-nil")
		}
	}()
	registry.Register(&fakeProvider{name: "local"})
}
