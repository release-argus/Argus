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

package rbac

import (
	"slices"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestCatalogue(t *testing.T) {
	got := Catalogue()

	wantResources := []Resource{
		ResourceService,
		ResourceServiceOrder,
		ResourceServiceAction,
		ResourceVersionRefresh,
		ResourceNotify,
		ResourceConfig,
		ResourceMetric,
	}
	if testErr := test.AssertSlicesEqualFunc(
		t,
		got,
		wantResources,
		func(g ResourcePermissions, w Resource) bool {
			return g.Resource == w
		},
		"",
	); testErr != nil {
		t.Fatal(testErr)
	}

	for _, rp := range got {
		if len(rp.Actions) == 0 {
			t.Errorf(
				"ResourcePermissions[%q].Actions mismatch\ngot:  []\nwant: non-empty",
				rp.Resource,
			)
		}
		for _, ap := range rp.Actions {
			if !slices.Contains(ap.Scopes, ScopeGlobal) {
				t.Errorf(
					"ResourcePermissions[%q].Actions[%q].Scopes mismatch\ngot:  %q\nwant: %q",
					rp.Resource, ap.Action,
					ap.Scopes, ScopeGlobal,
				)
			}
		}
	}

	got[0].Actions[0].Action = Action("mutated")
	got[0].Actions[0].Scopes[0] = ScopeType("mutated")
	original := catalogue[0].Actions[0]
	if original.Action == Action("mutated") ||
		original.Scopes[0] == ScopeType("mutated") {
		t.Errorf(
			"catalogue[0].Actions[0] mismatch\ngot:  Action=%q, Scopes[0]=%q\nwant: NOT %q",
			original.Action, original.Scopes[0], "mutated",
		)
	}
}

func TestResourcePermissions(t *testing.T) {
	wantAllScopes := []ScopeType{ScopeGlobal, ScopeService, ScopeServiceTag}
	wantGlobalOnly := []ScopeType{ScopeGlobal}
	tests := []struct {
		name        string
		resource    Resource
		wantNil     bool
		wantActions []ActionPermission
	}{
		{
			name:     "known/service (create is global-only, rest all scopes)",
			resource: ResourceService,
			wantActions: []ActionPermission{
				{ActionCreate, wantGlobalOnly},
				{ActionRead, wantAllScopes},
				{ActionUpdate, wantAllScopes},
				{ActionDelete, wantAllScopes},
			},
		},
		{
			name:        "known/service_order (global only)",
			resource:    ResourceServiceOrder,
			wantActions: []ActionPermission{{ActionUpdate, wantGlobalOnly}},
		},
		{
			name:        "known/service_action (execute, all scopes)",
			resource:    ResourceServiceAction,
			wantActions: []ActionPermission{{ActionExecute, wantAllScopes}},
		},
		{
			name:        "known/version_refresh (execute, all scopes)",
			resource:    ResourceVersionRefresh,
			wantActions: []ActionPermission{{ActionExecute, wantAllScopes}},
		},
		{
			name:        "known/notify (global only)",
			resource:    ResourceNotify,
			wantActions: []ActionPermission{{ActionExecute, wantGlobalOnly}},
		},
		{
			name:        "known/config (read, global only)",
			resource:    ResourceConfig,
			wantActions: []ActionPermission{{ActionRead, wantGlobalOnly}},
		},
		{
			name:        "known/metric (read, global only)",
			resource:    ResourceMetric,
			wantActions: []ActionPermission{{ActionRead, wantGlobalOnly}},
		},
		{
			name:     "unknown/not in catalogue",
			resource: Resource("unknown"),
			wantNil:  true,
		},
		{
			name:     "unknown/empty",
			resource: Resource(""),
			wantNil:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resourcePermissions(tc.resource)

			if tc.wantNil {
				if got != nil {
					t.Fatalf(
						"resourcePermissions() mismatch\ngot:  Resource=%q\nwant: nil",
						got.Resource,
					)
				}
				return
			}

			if got == nil {
				t.Fatal("resourcePermissions() mismatch\ngot:  nil\nwant: non-nil")
			}
			if got.Resource != tc.resource {
				t.Errorf(
					"ResourcePermissions.Resource mismatch\ngot:  %q\nwant: %q",
					got.Resource, tc.resource,
				)
			}
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got.Actions,
				tc.wantActions,
				func(a, b ActionPermission) bool {
					return a.Action == b.Action && slices.Equal(a.Scopes, b.Scopes)
				},
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestPermission_Valid(t *testing.T) {
	tests := []struct {
		name       string
		permission Permission
		want       bool
	}{
		{
			name:       "valid/service:read",
			permission: Permission{Resource: ResourceService, Action: ActionRead},
			want:       true,
		},
		{
			name:       "valid/service:delete",
			permission: Permission{Resource: ResourceService, Action: ActionDelete},
			want:       true,
		},
		{
			name:       "valid/version_refresh:execute",
			permission: Permission{Resource: ResourceVersionRefresh, Action: ActionExecute},
			want:       true,
		},
		{
			name:       "invalid/unknown resource",
			permission: Permission{Resource: Resource("unknown"), Action: ActionRead},
			want:       false,
		},
		{
			name:       "invalid/action not in resource's matrix (config:update)",
			permission: Permission{Resource: ResourceConfig, Action: ActionUpdate},
			want:       false,
		},
		{
			name:       "invalid/action not in resource's matrix (version_refresh:read)",
			permission: Permission{Resource: ResourceVersionRefresh, Action: ActionRead},
			want:       false,
		},
		{
			name:       "invalid/unknown action",
			permission: Permission{Resource: ResourceService, Action: Action("unknown")},
			want:       false,
		},
		{
			name:       "invalid/empty",
			permission: Permission{},
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.permission.Valid()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestGrant_Valid(t *testing.T) {
	tests := []struct {
		name  string
		grant Grant
		want  bool
	}{
		{
			name:  "valid/global grant with empty ref",
			grant: globalGrant(ResourceService, ActionRead),
			want:  true,
		},
		{
			name:  "valid/service-scoped grant with ref",
			grant: serviceGrant(ResourceService, ActionUpdate, "argus"),
			want:  true,
		},
		{
			name:  "valid/service_tag-scoped grant with ref",
			grant: tagGrant(ResourceServiceAction, ActionExecute, "prod"),
			want:  true,
		},
		{
			name:  "valid/global service:create",
			grant: globalGrant(ResourceService, ActionCreate),
			want:  true,
		},
		{
			name:  "invalid/service-scoped create",
			grant: serviceGrant(ResourceService, ActionCreate, "argus"),
			want:  false,
		},
		{
			name:  "invalid/service_tag-scoped create",
			grant: tagGrant(ResourceService, ActionCreate, "prod"),
			want:  false,
		},
		{
			name:  "invalid/permission not in catalogue",
			grant: Grant{Resource: ResourceConfig, Action: ActionDelete, Scope: Scope{Type: ScopeGlobal}},
			want:  false,
		},
		{
			name:  "invalid/scope type unsupported by resource (service_order:service-scoped)",
			grant: Grant{Resource: ResourceServiceOrder, Action: ActionUpdate, Scope: Scope{Type: ScopeService, Ref: "argus"}},
			want:  false,
		},
		{
			name:  "invalid/unknown scope type",
			grant: Grant{Resource: ResourceService, Action: ActionRead, Scope: Scope{Type: ScopeType("unknown")}},
			want:  false,
		},
		{
			name:  "invalid/global grant with a ref",
			grant: Grant{Resource: ResourceService, Action: ActionRead, Scope: Scope{Type: ScopeGlobal, Ref: "argus"}},
			want:  false,
		},
		{
			name:  "invalid/service-scoped grant without a ref",
			grant: Grant{Resource: ResourceService, Action: ActionRead, Scope: Scope{Type: ScopeService}},
			want:  false,
		},
		{
			name:  "invalid/service_tag-scoped grant without a ref",
			grant: Grant{Resource: ResourceService, Action: ActionRead, Scope: Scope{Type: ScopeServiceTag}},
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.grant.Valid()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}
