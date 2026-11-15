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
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service"
	dvmanual "github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestConfig_AddService(t *testing.T) {
	tests := []struct {
		name       string
		newService *service.Service
		oldService string
		oldIsCopy  bool // oldService is first replaced by one built like newService.
		wantOrder  []string
		errRegex   string
		dbMessages int
		saves      int
		nilMap     bool
	}{
		{
			name:       "New service/no old service",
			newService: testServiceURL(t, "test"),
			wantOrder:  []string{"alpha", "bravo", "charlie", "test"},
			errRegex:   `^$`,
			dbMessages: 1,
			saves:      1, // For the added service.
		},
		{
			name:       "New service/old ID not held",
			oldService: "unknown",
			newService: testServiceURL(t, "test"),
			wantOrder:  []string{"alpha", "bravo", "charlie", "test"},
			errRegex:   `^$`,
			dbMessages: 1,
			saves:      1, // For the added service.
		},
		{
			name:       "Replace service/old ID held",
			oldService: "bravo",
			newService: testServiceURL(t, "bravo"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			errRegex:   `^$`,
			dbMessages: 1,
			saves:      1, // For the updated service.
		},
		{
			name:       "Replace service/unchanged",
			oldService: "bravo",
			oldIsCopy:  true,
			newService: testServiceURL(t, "bravo"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			errRegex:   `^$`,
			dbMessages: 0,
			saves:      0,
		},
		{
			name:       "Rename service/other fields changed",
			oldService: "bravo",
			newService: testServiceURL(t, "foo"),
			wantOrder:  []string{"alpha", "foo", "charlie"},
			errRegex:   `^$`,
			dbMessages: 2, // 1 for change of ID, 1 for change of versions.
			saves:      1,
		},
		{
			name:       "Rename service/otherwise unchanged",
			oldService: "bravo",
			oldIsCopy:  true,
			newService: testServiceURL(t, "foo"),
			wantOrder:  []string{"alpha", "foo", "charlie"},
			errRegex:   `^$`,
			dbMessages: 1, // For the change of ID.
			saves:      1,
		},
		{
			name:       "Rename to an empty ID",
			oldService: "bravo",
			newService: testServiceURL(t, ""),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			errRegex:   `^id: <required>$`,
			dbMessages: 0,
		},
		{
			name:       "New service with an empty ID",
			newService: testServiceURL(t, ""),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			errRegex:   `^id: <required>$`,
			dbMessages: 0,
		},
		{
			name:       "ID already exists",
			newService: testServiceURL(t, "alpha"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			errRegex:   `^service "alpha" already exists$`,
			dbMessages: 0,
		},
		{
			name:       "Name already exists",
			newService: testServiceURL(t, "a"),
			wantOrder:  []string{"alpha", "bravo", "charlie"},
			errRegex:   `^service "a" already exists$`,
			dbMessages: 0,
		},
		{
			name:       "Add to nil service map",
			newService: testServiceURL(t, "test"),
			wantOrder:  []string{"test"},
			errRegex:   `^$`,
			nilMap:     true,
			dbMessages: 1,
			saves:      1, // For the added service.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())
			t.Cleanup(func() { _ = releaseStdout() })

			file := filepath.Join(t.TempDir(), "config.yml")
			testYAML_Edit(t, file)
			cfg := testLoadBasic(t, file)
			if tc.nilMap {
				cfg.Service = nil
				cfg.Order = []string{}
			}
			if tc.oldIsCopy {
				cfg.Service[tc.oldService] = testServiceURL(t, tc.oldService)
			}

			err := func() error {
				loadMu.RLock()
				defer loadMu.RUnlock()
				return cfg.AddService(tc.oldService, tc.newService)
			}()

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			added := tc.errRegex == `^$`

			cfg.OrderMu.RLock()
			svc := cfg.Service[tc.newService.ID]
			t.Cleanup(func() {
				if added && svc != nil {
					svc.PrepDelete(false)
				}
				cfg.OrderMu.RUnlock()
			})
			if added && svc != tc.newService {
				t.Fatalf(
					"Config.Service[newService.ID] mismatch\ngot:  %p\nwant: %p",
					svc, tc.newService,
				)
			}
			if !added && svc == tc.newService {
				t.Fatalf(
					"Config.Service[newService.ID] mismatch\ngot:  %p\nwant: NOT %p",
					svc, tc.newService,
				)
			}
			if !util.AreSlicesEqual(cfg.Order, tc.wantOrder) {
				t.Errorf("Config.Order mismatch\ngot:  %q\nwant: %q", cfg.Order, tc.wantOrder)
			}

			test.AssertChannelLen(
				t, "Status.DatabaseChannel", cfg.HardDefaults.Service.Status.DatabaseChannel, tc.dbMessages,
			)

			if got := len(cfg.SaveChannel); got != tc.saves {
				t.Errorf(
					"Config.SaveChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.saves,
				)
			}
		})
	}
}

func TestConfig_AddService__deployed_version__manual(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		wantVersion     string
		wantDBMessages  int
		wantDBCellValue string
	}{
		{
			name:            "version is applied and persisted",
			version:         "1.2.3",
			wantVersion:     "1.2.3",
			wantDBMessages:  1,
			wantDBCellValue: "1.2.3",
		},
		{
			name:           "no version, nothing to apply",
			version:        "",
			wantVersion:    "",
			wantDBMessages: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())
			t.Cleanup(func() { _ = releaseStdout() })

			file := filepath.Join(t.TempDir(), "config.yml")
			testYAML_Edit(t, file)
			cfg := testLoadBasic(t, file)
			newService := testServiceManualDV(t, "test", tc.version)

			err := func() error {
				loadMu.RLock()
				defer loadMu.RUnlock()
				return cfg.AddService("", newService)
			}()

			cfg.OrderMu.RLock()
			t.Cleanup(func() {
				if svc := cfg.Service["test"]; svc != nil {
					svc.PrepDelete(true)
				}
				cfg.OrderMu.RUnlock()
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if newService.DeployedVersionLookup == nil {
				t.Fatal("Service.DeployedVersionLookup mismatch\ngot:  nil\nwant: non-nil")
			}

			if got := newService.Status.DeployedVersion(); got != tc.wantVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.wantVersion,
				)
			}

			lookup, ok := newService.DeployedVersionLookup.(*dvmanual.Lookup)
			if !ok {
				t.Fatalf(
					"Service.DeployedVersionLookup type mismatch\ngot:  %T\nwant: *manual.Lookup",
					newService.DeployedVersionLookup,
				)
			}
			if got := lookup.Version; got != "" {
				t.Errorf("Lookup.Version mismatch\ngot:  %q\nwant: \"\" (consumed)", got)
			}

			if got := len(newService.Status.AnnounceChannel); got != 0 {
				t.Errorf("Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: 0", got)
			}

			dbChannel := newService.Status.DatabaseChannel
			if got := len(dbChannel); got != tc.wantDBMessages {
				t.Fatalf(
					"Status.DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					got, tc.wantDBMessages,
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
				t.Errorf(
					"DB `deployed_version` cell mismatch\ngot:  %q\nwant: %q",
					gotCell, tc.wantDBCellValue,
				)
			}
		})
	}
}

func TestConfig_serviceIDTaken(t *testing.T) {
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

			got := config.serviceIDTaken(tc.oldServiceID, tc.newService)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestConfig_ServiceWithNameExists(t *testing.T) {
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

			got := tc.config.ServiceWithNameExists(tc.serviceName, tc.oldServiceID)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestConfig_RenameService(t *testing.T) {
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
			testYAML_Edit(t, file)
			t.Cleanup(func() { _ = os.Remove(file) })
			newSVC := testServiceURL(t, tc.newID)

			cfg := func() *Config {
				logMu.Lock()
				defer logMu.Unlock()
				cfg := testLoadBasic(t, file)
				cfg.renameService(tc.oldID, newSVC)
				return cfg
			}()

			cfg.OrderMu.RLock()
			t.Cleanup(func() {
				if !tc.fail {
					cfg.Service[tc.newID].PrepDelete(false)
				}
				cfg.OrderMu.RUnlock()
			})
			if !util.AreSlicesEqual(cfg.Order, tc.wantOrder) {
				t.Errorf("Config.Order mismatch\ngot:  %q\nwant: %q", cfg.Order, tc.wantOrder)
			}

			renamed := !tc.fail && !tc.noChange
			if got := cfg.Service[tc.oldID]; renamed && got != nil {
				t.Error("Config.Service[oldID] mismatch\ngot:  present\nwant: absent (renamed)")
			}

			if got := cfg.Service[tc.newID]; renamed && got != newSVC {
				t.Errorf("Config.Service[newID] mismatch\ngot:  %p\nwant: %p", got, newSVC)
			} else if !renamed && got == newSVC {
				t.Errorf("Config.Service[newID] mismatch\ngot:  %p\nwant: NOT %p", got, newSVC)
			}

			want := 0
			if renamed {
				want = 1
			}
			test.AssertChannelLen(
				t, "Status.DatabaseChannel", cfg.HardDefaults.Service.Status.DatabaseChannel, want,
			)
		})
	}
}

func TestConfig_DeleteService(t *testing.T) {
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
			testYAML_Edit(t, file)

			cfg := func() *Config {
				logMu.Lock()
				defer logMu.Unlock()
				cfg := testLoadBasic(t, file)
				cfg.DeleteService(tc.id)
				return cfg
			}()

			cfg.OrderMu.RLock()
			t.Cleanup(cfg.OrderMu.RUnlock)
			if got := cfg.Service[tc.id]; got != nil {
				t.Error("Config.Service[id] mismatch\ngot:  present\nwant: absent (deleted)")
			}

			if !util.AreSlicesEqual(cfg.Order, tc.wantOrder) {
				t.Errorf("Config.Order mismatch\ngot:  %q\nwant: %q", cfg.Order, tc.wantOrder)
			}

			want := 0
			if tc.dbMessage {
				want = 1
			}
			test.AssertChannelLen(
				t, "Status.DatabaseChannel", cfg.HardDefaults.Service.Status.DatabaseChannel, want,
			)
		})
	}
}
