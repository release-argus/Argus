// Copyright [2022] [Argus]
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

package config

import (
	"testing"
	"time"

	"github.com/release-argus/Argus/service"
)

func TestConfig_RenameService(t *testing.T) {
	// GIVEN a service to rename and a Config to act on
	tests := map[string]struct {
		oldName   string
		newName   string
		wantOrder []string
		noChange  bool
		fail      bool
	}{
		"Rename service": {
			oldName: "bravo", newName: "foo",
			wantOrder: []string{"alpha", "foo", "charlie"},
			fail:      false,
		},
		"Rename service to same name": {
			oldName: "bravo", newName: "bravo",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			noChange:  true,
			fail:      false,
		},
		"Rename service that doesn't exist": {
			oldName: "test", newName: "foo",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			fail:      true,
		},
		"Rename service to existing name": {
			oldName: "bravo", newName: "alpha",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			fail:      true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := testConfigEdit()
			newSVC := testServiceURL(tc.newName)

			// WHEN the service is renamed
			config.RenameService(tc.oldName, newSVC)
			time.Sleep(time.Second)

			// THEN the order should be as expected
			if len(config.Order) != len(tc.wantOrder) {
				t.Errorf("Order length mismatch: got %d, want %d", len(config.Order), len(tc.wantOrder))
			}
			for i, service := range config.Order {
				if service != tc.wantOrder[i] {
					t.Fatalf("Order mismatch at index %d: got %s, want %s\ngot:  %v\nwant: %v",
						i, service, tc.wantOrder[i], config.Order, tc.wantOrder)
				}
			}
			// AND the service should be removed if it was renamed
			if !tc.fail && tc.oldName != tc.newName && config.Service[tc.oldName] != nil {
				t.Errorf("%q should have been removed, got %+v", tc.oldName, config.Service[tc.oldName])
			}
			// AND the service should be at the address given
			if !tc.fail && config.Service[tc.newName] != newSVC {
				if tc.noChange {
					return
				}
				t.Errorf("%q should be at the given address, got\n%+v", tc.newName, config.Service[tc.newName])
			}
		})
	}
}

func TestConfig_DeleteService(t *testing.T) {
	// GIVEN a service to delete and a Config to act on
	tests := map[string]struct {
		name      string
		wantOrder []string
	}{
		"Delete service": {
			name:      "bravo",
			wantOrder: []string{"alpha", "charlie"},
		},
		"Delete service that doesn't exist": {
			name:      "test",
			wantOrder: []string{"alpha", "bravo", "charlie"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := testConfigEdit()

			// WHEN the service is deleted
			config.DeleteService(tc.name)

			// THEN the service was removed
			if config.Service[tc.name] != nil {
				t.Errorf("%q was not removed", tc.name)
			}
			// AND the order was updated
			if len(config.Order) != len(tc.wantOrder) {
				t.Errorf("Order length mismatch: got %d, want %d",
					len(config.Order), len(tc.wantOrder))
			}
			for i, service := range config.Order {
				if service != tc.wantOrder[i] {
					t.Fatalf("Order mismatch at index %d: got %s, want %s\ngot:  %v\nwant: %v",
						i, service, tc.wantOrder[i], config.Order, tc.wantOrder)
				}
			}
		})
	}
}

func TestConfig_AddService(t *testing.T) {
	// GIVEN a service to add/replace/rename and a Config to act on
	tests := map[string]struct {
		newService *service.Service
		oldService string
		wantOrder  []string
		added      bool
		nilMap     bool
	}{
		"New service": {
			newService: testServiceURL("test"),
			wantOrder:  []string{"alpha", "bravo", "charlie", "test"},
			added:      true,
		},
		"Replace service": {
			oldService: "bravo",
			newService: testServiceURL("bravo"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      true,
		},
		"Rename service": {
			oldService: "bravo",
			newService: testServiceURL("foo"),
			wantOrder:  []string{"alpha", "foo", "charlie"},
			added:      true,
		},
		"Add service that already exists": {
			newService: testServiceURL("alpha"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
		},
		"nil service map": {
			newService: testServiceURL("test"),
			wantOrder:  []string{"test"},
			added:      true,
			nilMap:     true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			config := testConfigEdit()
			if tc.nilMap {
				config.Service = nil
				config.Order = []string{}
			}

			// WEHN AddService is called
			config.AddService(tc.oldService, tc.newService)

			// THEN the service is
			// added/renamed/replaced
			if tc.added && config.Service[tc.newService.ID] != tc.newService {
				t.Fatalf("oldService %q wasn't placed at config[%q]", tc.oldService, tc.newService.ID)
			}
			if !tc.added && config.Service[tc.newService.ID] == tc.newService {
				t.Fatalf("config[%q] shouldn't have been added", tc.newService.ID)
			}
			// Added to Order
			if len(config.Order) != len(tc.wantOrder) {
				t.Errorf("Order length mismatch: got %d, want %d\nwant: %v\ngot: %v",
					len(config.Order), len(tc.wantOrder), config.Order, tc.wantOrder)
			}
			// In the correct spot
			for i, service := range config.Order {
				if service != tc.wantOrder[i] {
					t.Fatalf("Order mismatch at index %d: got %s, want %s\ngot:  %v\nwant: %v",
						i, service, tc.wantOrder[i], config.Order, tc.wantOrder)
				}
			}
		})
	}
}
