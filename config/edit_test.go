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
	dvmanual "github.com/release-argus/Argus/service/deployed_version/types/manual"
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
			newService: testServiceURL(t, "test"),
			wantOrder:  []string{"alpha", "bravo", "charlie", "test"},
			added:      true,
			dbMessages: 1,
		},
		{
			name:       "Replace service",
			oldService: "bravo",
			newService: testServiceURL(t, "bravo"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      true,
			dbMessages: 1,
		},
		{
			name:       "Rename service",
			oldService: "bravo",
			newService: testServiceURL(t, "foo"),
			wantOrder:  []string{"alpha", "foo", "charlie"},
			added:      true,
			dbMessages: 2, // 1 for change of ID, 1 for change of versions.
		},
		{
			name:       "Rename to an empty ID",
			oldService: "bravo",
			newService: testServiceURL(t, ""),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		{
			name:       "New service with an empty ID",
			newService: testServiceURL(t, ""),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		{
			name:       "ID already exists",
			newService: testServiceURL(t, "alpha"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		{
			name:       "Name already exists",
			newService: testServiceURL(t, "a"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			added:      false,
			dbMessages: 0,
		},
		{
			name:       "Add to nil service map",
			newService: testServiceURL(t, "test"),
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

func TestConfig_AddService__deployed_version__manual(t *testing.T) {
	// GIVEN: a Service whose `manual` deployed_version carries a pending version.
	tests := map[string]struct {
		version         string
		wantVersion     string
		wantDBMessages  int
		wantDBCellValue string
	}{
		"version is applied and persisted": {
			version:         "1.2.3",
			wantVersion:     "1.2.3",
			wantDBMessages:  1,
			wantDBCellValue: "1.2.3",
		},
		"no version, nothing to apply": {
			version:        "",
			wantVersion:    "",
			wantDBMessages: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing global log state.
			releaseStdout := test.CaptureLog(t, logx.Default())
			t.Cleanup(func() { _ = releaseStdout() })

			file := filepath.Join(t.TempDir(), "config.yml")
			testYAML_Edit(file)
			cfg := testLoadBasic(t, file)
			newService := testServiceManualDV(t, "test", tc.version)

			// WHEN: AddService is called.
			loadMu.RLock()
			err := cfg.AddService("", newService)
			loadMu.RUnlock()

			prefix := fmt.Sprintf(
				"%s\nConfig.AddService(type=manual, version=%q)",
				packageName, tc.version,
			)

			cfg.OrderMu.RLock()
			t.Cleanup(func() {
				cfg.Service["test"].PrepDelete(true)
				cfg.OrderMu.RUnlock()
			})

			// THEN: the Service is added.
			if err != nil {
				t.Fatalf("%s unexpected error: %v", prefix, err)
			}

			// AND: the Service kept its `manual` deployed_version.
			if newService.DeployedVersionLookup == nil {
				t.Fatalf("%s DeployedVersionLookup is nil - it did not decode", prefix)
			}

			// AND: the configured version reaches the Status.
			if got := newService.Status.DeployedVersion(); got != tc.wantVersion {
				t.Errorf("%s mismatch on Status.DeployedVersion()\ngot:  %q\nwant: %q",
					prefix, got, tc.wantVersion,
				)
			}

			// AND: the pending Version is consumed.
			if got := newService.DeployedVersionLookup.(*dvmanual.Lookup).Version; got != "" {
				t.Errorf("%s did not consume the pending version\ngot: %q",
					prefix, got,
				)
			}

			// AND: it is applied silently.
			if got := len(newService.Status.AnnounceChannel); got != 0 {
				t.Errorf("%s AnnounceChannel message count mismatch\ngot:  %d\nwant: 0",
					prefix, got,
				)
			}

			// AND: it is persisted to the DB, so that it survives a restart.
			dbChannel := newService.Status.DatabaseChannel
			if got := len(dbChannel); got != tc.wantDBMessages {
				t.Fatalf("%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantDBMessages,
				)
			}
			var gotCell string
			for range tc.wantDBMessages {
				for _, cell := range (<-dbChannel).Cells {
					if cell.Column == "deployed_version" {
						gotCell = cell.Value
					}
				}
			}
			if gotCell != tc.wantDBCellValue {
				t.Errorf("%s mismatch on the `deployed_version` cell sent to the DB\ngot:  %q\nwant: %q",
					prefix, gotCell, tc.wantDBCellValue,
				)
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
			name: "add/empty name",
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
			name: "add/new name",
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
			name: "add/conflict",
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
			name: "rename/unchanged",
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
			name: "rename/conflict",
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

func TestConfig_serviceIDTaken(t *testing.T) {
	// GIVEN: a Config holding a named and an unnamed service.
	config := &Config{
		Service: map[string]*service.Service{
			"unnamed": {ID: "unnamed"},
			"named":   {ID: "named", Name: "alias"},
		},
	}
	tests := []struct {
		name         string
		oldServiceID string
		newService   *service.Service
		want         bool
	}{
		{
			name:         "create/free id and name",
			oldServiceID: "",
			newService:   &service.Service{ID: "fresh", Name: "novel"},
			want:         false,
		},
		{
			name:         "create/id taken by another id",
			oldServiceID: "",
			newService:   &service.Service{ID: "unnamed"},
			want:         true,
		},
		{
			name:         "create/id taken by another name",
			oldServiceID: "",
			newService:   &service.Service{ID: "alias"},
			want:         true,
		},
		{
			name:         "create/name taken by another name",
			oldServiceID: "",
			newService:   &service.Service{ID: "fresh", Name: "alias"},
			want:         true,
		},
		{
			name:         "create/name taken by another id",
			oldServiceID: "",
			newService:   &service.Service{ID: "fresh", Name: "unnamed"},
			want:         true,
		},
		{
			name:         "edit/keeps its own id",
			oldServiceID: "named",
			newService:   &service.Service{ID: "named", Name: "alias"},
			want:         false,
		},
		{
			name:         "edit/names itself after its own id",
			oldServiceID: "named",
			newService:   &service.Service{ID: "named", Name: "named"},
			want:         false,
		},
		{
			name:         "edit/name taken by another id",
			oldServiceID: "named",
			newService:   &service.Service{ID: "named", Name: "unnamed"},
			want:         true,
		},
		{
			name:         "rename/id taken by another id",
			oldServiceID: "named",
			newService:   &service.Service{ID: "unnamed"},
			want:         true,
		},
		{
			name:         "rename/name taken by another id",
			oldServiceID: "named",
			newService:   &service.Service{ID: "fresh", Name: "unnamed"},
			want:         true,
		},
		{
			name:         "rename/free id and name",
			oldServiceID: "named",
			newService:   &service.Service{ID: "fresh", Name: "novel"},
			want:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: serviceIDTaken is called.
			got := config.serviceIDTaken(tc.oldServiceID, tc.newService)

			// THEN: a collision with any other service ID or name is reported.
			if got != tc.want {
				t.Errorf(
					"%s\nConfig serviceIDTaken(oldID=%q, id=%q, name=%q) result mismatch\ngot:  %t\nwant: %t",
					packageName, tc.oldServiceID, tc.newService.ID, tc.newService.Name,
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
			name:  "new name",
			oldID: "bravo", newID: "foo",
			wantOrder: []string{"alpha", "foo", "charlie"},
			fail:      false,
		},
		{
			name:  "same name",
			oldID: "bravo", newID: "bravo",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			noChange:  true,
			fail:      false,
		},
		{
			name:  "service doesn't exist",
			oldID: "test", newID: "foo",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			fail:      true,
		},
		{
			name:  "existing name",
			oldID: "bravo", newID: "alpha",
			wantOrder: []string{"alpha", "bravo", "charlie"},
			fail:      true,
		},
		{
			name:  "empty name",
			oldID: "bravo", newID: "",
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
			newSVC := testServiceURL(t, tc.newID)

			// WHEN: the service is renamed.
			cfg.renameService(tc.oldID, newSVC)
			logMu.Unlock()
			time.Sleep(time.Second)

			prefix := fmt.Sprintf(
				"%s\nConfig.renameService(oldID=%q, newID=%q)",
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
