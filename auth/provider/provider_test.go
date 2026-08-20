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
	"fmt"
	"testing"
)

func TestRegistry_Register_and_Get(t *testing.T) {
	// GIVEN: sequences of registrations.
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

			// WHEN: providers are registered.
			for _, providerName := range tc.register {
				registry.Register(&fakeProvider{name: providerName})
			}

			prefix := fmt.Sprintf(
				"%s\nRegistry(%v)",
				packageName, tc.register,
			)

			// AND: Get returns registered providers by name.
			if tc.getKnown != "" {
				got := registry.Get(tc.getKnown)
				if got == nil || got.Name() != tc.getKnown {
					t.Errorf(
						"%s Get(%q) should return the registered provider\ngot: %v",
						prefix, tc.getKnown, got,
					)
				}
			}

			// AND: Get returns nil for unknown names.
			if tc.getUnknown != "" {
				if got := registry.Get(tc.getUnknown); got != nil {
					t.Errorf(
						"%s Get(%q) should return nil for an unknown provider\ngot: %v",
						prefix, tc.getUnknown, got,
					)
				}
			}
		})
	}
}

func TestRegistry_Register__duplicateNamePanics(t *testing.T) {
	// GIVEN: a registry already holding a provider.
	registry := NewRegistry()
	registry.Register(&fakeProvider{name: "local"})

	// WHEN: a second provider registers under the same name.
	defer func() {
		// THEN: it panics.
		if recover() == nil {
			t.Errorf("%s\nRegistry.Register(duplicate) expected a panic, got none", packageName)
		}
	}()
	registry.Register(&fakeProvider{name: "local"})
}
