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

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
)

func TestConfig_AddService(t *testing.T) {
	// GIVEN: a service to add/replace/rename, and a Config to act on.
	tests := []struct {
		name       string
		newService *service.Service
		oldService string
		wantOrder  []string
		added      bool
		dbMessages int
		nilMap     bool
	}{
		{
			name:       "New service",
			newService: testServiceURL("test"),
			wantOrder:  []string{"alpha", "bravo", "charlie", "test"},
			added:      true,
			dbMessages: 1,
		},
		{
			name:       "Replace service",
			oldService: "bravo",
			newService: testServiceURL("bravo"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      true,
			dbMessages: 1,
		},
		{
			name:       "Rename service",
			oldService: "bravo",
			newService: testServiceURL("foo"),
			wantOrder:  []string{"alpha", "foo", "charlie"},
			added:      true,
			dbMessages: 2, // 1 for change of ID, 1 for change of versions.
		},
		{
			name:       "ID already exists",
			newService: testServiceURL("alpha"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		{
			name:       "Name already exists",
			newService: testServiceURL("a"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		{
			name:       "Add to nil service map",
			newService: testServiceURL("test"),
			wantOrder:  []string{"test"},
			added:      true,
			nilMap:     true,
			dbMessages: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using sharing global log state.
			releaseStdout := test.CaptureLog(t, logx.Default())
			t.Cleanup(func() { _ = releaseStdout() })

			file := filepath.Join(t.TempDir(), "config.yml")
			testYAML_Edit(file)
			cfg := testLoadBasic(t, file)
			if tc.nilMap {
				cfg.Service = nil
				cfg.Order = []string{}
			}

			// WHEN: AddService is called.
			loadMu.RLock()
			_ = cfg.AddService(tc.oldService, tc.newService)
			loadMu.RUnlock()

			prefix := fmt.Sprintf(
				"%s\nConfig.AddService(oldID=%q, newService=%q)",
				packageName, tc.oldService, tc.newService.ID,
			)

			// THEN: the service is:
			// 	added/renamed/replaced.
			cfg.OrderMu.RLock()
			t.Cleanup(func() {
				if tc.added {
					cfg.Service[tc.newService.ID].PrepDelete(false)
				}
				cfg.OrderMu.RUnlock()
			})
			if tc.added && cfg.Service[tc.newService.ID] != tc.newService {
				t.Fatalf(
					"%s oldService %q wasn't placed at config[%q]",
					prefix, tc.oldService, tc.newService.ID,
				)
			}
			if !tc.added && cfg.Service[tc.newService.ID] == tc.newService {
				t.Fatalf(
					"%s config[%q] shouldn't have been added",
					prefix, tc.newService.ID,
				)
			}
			// Added to Order at the correct spot.
			if !util.AreSlicesEqual(cfg.Order, tc.wantOrder) {
				t.Errorf(
					"%s Order mismatch (added: %t)\ngot:  %q\nwant: %q",
					prefix, tc.added,
					cfg.Order, tc.wantOrder,
				)
			}

			// AND: the DatabaseChannel should have a message waiting if the service was added.
			if got := len(cfg.HardDefaults.Service.Status.DatabaseChannel); got != tc.dbMessages {
				t.Errorf(
					"%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.dbMessages,
				)
				for i := 0; i < got; i++ {
					msg := <-cfg.HardDefaults.Service.Status.DatabaseChannel
					t.Log(msg)
				}
			}
		})
	}
}

func TestConfig_ServiceWithNameExists(t *testing.T) {
	// GIVEN: a Config to act on.
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ServiceWithNameExists is called.
			got := tc.config.ServiceWithNameExists(tc.serviceName, tc.oldServiceID)

			// THEN: we receive the expected result.
			if got != tc.want {
				t.Errorf(
					"%s\nConfig ServiceWithNameExists(id=%q, oldID=%q) result mismatch\ngot:  %t\nwant: %t",
					packageName, tc.serviceName, tc.oldServiceID,
					got, tc.want,
				)
			}
		})
	}
}

func TestConfig_RenameService(t *testing.T) {
	// GIVEN: a service to rename, and a Config to act on.
	tests := []struct {
		name           string
		oldID, newID   string
		wantOrder      []string
		noChange, fail bool
	}{
		{
			name:  "Rename service",
			oldID: "bravo", newID: "foo",
			wantOrder: []string{"alpha", "foo", "charlie"},
			fail:      false,
		},
		{
			name:  "Rename service to same name",
			oldID: "bravo", newID: "bravo",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			noChange:  true,
			fail:      false,
		},
		{
			name:  "Rename service that doesn't exist",
			oldID: "test", newID: "foo",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			fail:      true,
		},
		{
			name:  "Rename service to existing name",
			oldID: "bravo", newID: "alpha",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			fail:      true,
		},
	}
	logMu := sync.Mutex{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := filepath.Join(t.TempDir(), "config.yml")
			testYAML_Edit(file)
			t.Cleanup(func() { _ = os.Remove(file) })
			logMu.Lock()
			cfg := testLoadBasic(t, file)
			newSVC := testServiceURL(tc.newID)

			// WHEN: the service is renamed.
			cfg.RenameService(tc.oldID, newSVC)
			logMu.Unlock()
			time.Sleep(time.Second)

			prefix := fmt.Sprintf(
				"%s\nConfig.RenameService(oldID=%q, newID=%q)",
				packageName, tc.oldID, tc.newID,
			)

			// THEN: the order should be as expected.
			cfg.OrderMu.RLock()
			t.Cleanup(func() {
				if !tc.fail {
					cfg.Service[tc.newID].PrepDelete(false)
				}
				cfg.OrderMu.RUnlock()
			})
			if !util.AreSlicesEqual(cfg.Order, tc.wantOrder) {
				t.Errorf(
					"%s Order mismatch:\ngot:  %q\nwant: %q",
					prefix, cfg.Order, tc.wantOrder,
				)
			}

			// AND: the service should be removed if it was renamed.
			if !tc.fail && tc.oldID != tc.newID && cfg.Service[tc.oldID] != nil {
				t.Errorf(
					"%s: %q should have been removed, got %+v",
					prefix, tc.oldID, cfg.Service[tc.oldID],
				)
			}

			// AND: the service should be at the address given.
			if !tc.fail && cfg.Service[tc.newID] != newSVC {
				if tc.noChange {
					return
				}
				t.Errorf(
					"%s %q should be at the given address, got\n%+v",
					prefix, tc.newID, cfg.Service[tc.newID],
				)
			}

			// AND: the DatabaseChannel should have a message waiting if it didn't fail.
			want := 0
			if !tc.fail {
				want = 1
			}
			if got := len(cfg.HardDefaults.Service.Status.DatabaseChannel); got != want {
				t.Errorf(
					"%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
				for i := 0; i <= len(cfg.HardDefaults.Service.Status.DatabaseChannel); i++ {
					msg := <-cfg.HardDefaults.Service.Status.DatabaseChannel
					t.Log(msg)
				}
			}
		})
	}
}

func TestConfig_DeleteService(t *testing.T) {
	// GIVEN: a service to delete, and a Config to act on.
	tests := []struct {
		name      string
		id        string
		wantOrder []string
		dbMessage bool
	}{
		{
			name:      "Delete service",
			id:        "bravo",
			wantOrder: []string{"alpha", "charlie"},
			dbMessage: true,
		},
		{
			name:      "Delete service that doesn't exist",
			id:        "test",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			dbMessage: false,
		},
	}
	logMu := sync.Mutex{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := filepath.Join(t.TempDir(), "config.yml")
			testYAML_Edit(file)
			logMu.Lock()
			cfg := testLoadBasic(t, file)

			// WHEN: the service is deleted.
			cfg.DeleteService(tc.id)
			logMu.Unlock()

			prefix := fmt.Sprintf(
				"%s\nConfig.DeleteService(%q)",
				packageName, tc.name,
			)

			// THEN: the service was removed.
			cfg.OrderMu.RLock()
			t.Cleanup(cfg.OrderMu.RUnlock)
			if got := cfg.Service[tc.name]; got != nil {
				t.Errorf(
					"%s service was not removed\ngot:  %p\nwant: nil",
					prefix, got,
				)
			}

			// AND: the Order was updated.
			if !util.AreSlicesEqual(cfg.Order, tc.wantOrder) {
				t.Errorf(
					"%s Order mismatch:\ngot:  %q\nwant: %q",
					prefix, cfg.Order, tc.wantOrder,
				)
			}

			// AND: the DatabaseChannel should have a message waiting if the service was deleted.
			want := 0
			if tc.dbMessage {
				want = 1
			}
			if got := len(cfg.HardDefaults.Service.Status.DatabaseChannel); got != want {
				t.Errorf(
					"%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
				for i := 0; i <= len(cfg.HardDefaults.Service.Status.DatabaseChannel); i++ {
					msg := <-cfg.HardDefaults.Service.Status.DatabaseChannel
					t.Log(msg)
				}
			}
		})
	}
}
