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

package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/test"
)

func TestConfig_AddService(t *testing.T) {
	// GIVEN a service to add/replace/rename and a Config to act on.
	tests := map[string]struct {
		newService *service.Service
		oldService string
		wantOrder  []string
		added      bool
		dbMessages int
		nilMap     bool
	}{
		"New service": {
			newService: testServiceURL("test"),
			wantOrder:  []string{"alpha", "bravo", "charlie", "test"},
			added:      true,
			dbMessages: 1,
		},
		"Replace service": {
			oldService: "bravo",
			newService: testServiceURL("bravo"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      true,
			dbMessages: 1,
		},
		"Rename service": {
			oldService: "bravo",
			newService: testServiceURL("foo"),
			wantOrder:  []string{"alpha", "foo", "charlie"},
			added:      true,
			dbMessages: 2, // 1 for change of id, 1 for change of versions.
		},
		"ID already exists": {
			newService: testServiceURL("alpha"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		"Name already exists": {
			newService: testServiceURL("a"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		"Add to nil service map": {
			newService: testServiceURL("test"),
			wantOrder:  []string{"test"},
			added:      true,
			nilMap:     true,
			dbMessages: 1,
		},
	}
	logMutex := sync.Mutex{}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using sharing global log state.

			file := fmt.Sprintf("TestConfig_AddService_%s.yml", strings.ReplaceAll(name, " ", "_"))
			testYAML_Edit(file, t)
			logMutex.Lock()
			cfg := testLoadBasic(file, t)
			if tc.nilMap {
				cfg.Service = nil
				cfg.Order = []string{}
			}

			// WHEN AddService is called.
			loadMutex.RLock()
			cfg.AddService(tc.oldService, tc.newService)
			loadMutex.RUnlock()
			logMutex.Unlock()

			// THEN the service is
			// 	added/renamed/replaced.
			cfg.OrderMutex.RLock()
			t.Cleanup(func() { cfg.OrderMutex.RUnlock() })
			if tc.added && cfg.Service[tc.newService.ID] != tc.newService {
				t.Fatalf("oldService %q wasn't placed at config[%q]", tc.oldService, tc.newService.ID)
			}
			if !tc.added && cfg.Service[tc.newService.ID] == tc.newService {
				t.Fatalf("config[%q] shouldn't have been added", tc.newService.ID)
			}
			// Added to Order at the correct spot.
			if !test.EqualSlices(cfg.Order, tc.wantOrder) {
				t.Errorf("Order mismatch: got %v, want %v",
					cfg.Order, tc.wantOrder)
			}
			// AND the DatabaseChannel should have a message waiting if the service was added.
			if len(*cfg.HardDefaults.Service.Status.DatabaseChannel) != tc.dbMessages {
				t.Errorf("DatabaseChannel should have %d messages waiting, got %d",
					tc.dbMessages, len(*cfg.HardDefaults.Service.Status.DatabaseChannel))
				for i := 0; i <= len(*cfg.HardDefaults.Service.Status.DatabaseChannel); i++ {
					msg := <-*cfg.HardDefaults.Service.Status.DatabaseChannel
					t.Log(msg)
				}
			}
		})
	}
}

func TestConfig_ServiceWithNameExists(t *testing.T) {
	// GIVEN a Config to act on.
	tests := []struct {
		name         string
		config       *Config
		serviceName  string
		oldServiceID string
		want         bool
	}{
		{
			name: "add - empty name",
			config: &Config{
				Service: map[string]*service.Service{
					"service1": {Name: "a"},
				},
			},
			serviceName:  "",
			oldServiceID: "",
			want:         false,
		},
		{
			name: "add - new name",
			config: &Config{
				Service: map[string]*service.Service{
					"service1": {Name: "a"},
				},
			},
			serviceName:  "b",
			oldServiceID: "",
			want:         false,
		},
		{
			name: "add - conflict",
			config: &Config{
				Service: map[string]*service.Service{
					"service1": {Name: "a"},
				},
			},
			serviceName:  "a",
			oldServiceID: "",
			want:         true,
		},
		{
			name: "rename - unchanged",
			config: &Config{
				Service: map[string]*service.Service{
					"service1": {Name: "a"},
				},
			},
			serviceName:  "a",
			oldServiceID: "service1",
			want:         false,
		},
		{
			name: "rename - conflict",
			config: &Config{
				Service: map[string]*service.Service{
					"service1": {Name: "a"},
					"service2": {Name: "b"},
				},
			},
			serviceName:  "b",
			oldServiceID: "service1",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// WHEN ServiceWithNameExists is called.
			got := tt.config.ServiceWithNameExists(tt.serviceName, tt.oldServiceID)

			// THEN we receive the expected result.
			if got != tt.want {
				t.Errorf("Config.ServiceWithNameExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_RenameService(t *testing.T) {
	// GIVEN a service to rename and a Config to act on.
	tests := map[string]struct {
		oldName, newName string
		wantOrder        []string
		noChange, fail   bool
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
	logMutex := sync.Mutex{}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file := fmt.Sprintf("TestConfig_RenameService_%s.yml", name)
			testYAML_Edit(file, t)
			t.Cleanup(func() { os.Remove(file) })
			logMutex.Lock()
			cfg := testLoadBasic(file, t)
			newSVC := testServiceURL(tc.newName)

			// WHEN the service is renamed.
			cfg.RenameService(tc.oldName, newSVC)
			logMutex.Unlock()
			time.Sleep(time.Second)

			// THEN the order should be as expected.
			cfg.OrderMutex.RLock()
			t.Cleanup(func() { cfg.OrderMutex.RUnlock() })
			if !test.EqualSlices(cfg.Order, tc.wantOrder) {
				t.Errorf("Order mismatch: got %v, want %v",
					cfg.Order, tc.wantOrder)
			}
			// AND the service should be removed if it was renamed.
			if !tc.fail && tc.oldName != tc.newName && cfg.Service[tc.oldName] != nil {
				t.Errorf("%q should have been removed, got %+v", tc.oldName, cfg.Service[tc.oldName])
			}
			// AND the service should be at the address given.
			if !tc.fail && cfg.Service[tc.newName] != newSVC {
				if tc.noChange {
					return
				}
				t.Errorf("%q should be at the given address, got\n%+v", tc.newName, cfg.Service[tc.newName])
			}
			// AND the DatabaseChannel should have a message waiting if it didn't fail.
			want := 0
			if !tc.fail {
				want = 1
			}
			if len(*cfg.HardDefaults.Service.Status.DatabaseChannel) != want {
				t.Errorf("DatabaseChannel should have %d messages waiting, got %d",
					want, len(*cfg.HardDefaults.Service.Status.DatabaseChannel))
				for i := 0; i <= len(*cfg.HardDefaults.Service.Status.DatabaseChannel); i++ {
					msg := <-*cfg.HardDefaults.Service.Status.DatabaseChannel
					t.Log(msg)
				}
			}
		})
	}
}

func TestConfig_DeleteService(t *testing.T) {
	// GIVEN a service to delete and a Config to act on.
	tests := map[string]struct {
		name      string
		wantOrder []string
		dbMessage bool
	}{
		"Delete service": {
			name:      "bravo",
			wantOrder: []string{"alpha", "charlie"},
			dbMessage: true,
		},
		"Delete service that doesn't exist": {
			name:      "test",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			dbMessage: false,
		},
	}
	logMutex := sync.Mutex{}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file := fmt.Sprintf("TestConfig_DeleteService_%s.yml", name)
			testYAML_Edit(file, t)
			logMutex.Lock()
			cfg := testLoadBasic(file, t)

			// WHEN the service is deleted.
			cfg.DeleteService(tc.name)
			logMutex.Unlock()

			// THEN the service was removed.
			cfg.OrderMutex.RLock()
			t.Cleanup(func() { cfg.OrderMutex.RUnlock() })
			if cfg.Service[tc.name] != nil {
				t.Errorf("%q was not removed", tc.name)
			}
			// AND the Order was updated.
			if !test.EqualSlices(cfg.Order, tc.wantOrder) {
				t.Errorf("Order mismatch: got %v, want %v",
					cfg.Order, tc.wantOrder)
			}
			// AND the DatabaseChannel should have a message waiting if the service was deleted.
			want := 0
			if tc.dbMessage {
				want = 1
			}
			if len(*cfg.HardDefaults.Service.Status.DatabaseChannel) != want {
				t.Errorf("DatabaseChannel should have %d messages waiting, got %d",
					want, len(*cfg.HardDefaults.Service.Status.DatabaseChannel))
				for i := 0; i <= len(*cfg.HardDefaults.Service.Status.DatabaseChannel); i++ {
					msg := <-*cfg.HardDefaults.Service.Status.DatabaseChannel
					t.Log(msg)
				}
			}
		})
	}
}
