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
	"fmt"
	"slices"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestCatalogue(t *testing.T) {
	prefix := fmt.Sprintf("%s\nCatalogue()", packageName)

	// GIVEN: the permission catalogue.
	// WHEN: Catalogue is called.
	got := Catalogue()

	// THEN: every known Resource appears exactly once, in stable order.
	wantResources := []Resource{
		ResourceService,
		ResourceServiceOrder,
		ResourceServiceAction,
		ResourceVersionRefresh,
		ResourceNotify,
		ResourceConfig,
	}
	if testErr := test.AssertSlicesEqualFunc(
		t,
		got,
		wantResources,
		func(g ResourcePermissions, w Resource) bool {
			return g.Resource == w
		},
		prefix,
		"",
	); testErr != nil {
		t.Fatal(testErr)
	}

	// AND: every entry has at least one Action, each supporting global scope.
	for _, rp := range got {
		if len(rp.Actions) == 0 {
			t.Errorf(
				"%s %q has no actions",
				prefix, rp.Resource,
			)
		}
		for _, ap := range rp.Actions {
			if !slices.Contains(ap.Scopes, ScopeGlobal) {
				t.Errorf(
					"%s %q:%q does not support global scope",
					prefix, rp.Resource, ap.Action,
				)
			}
		}
	}

	// AND: mutating the returned slice (or its nested scopes) does not affect
	// the source of truth.
	got[0].Actions[0].Action = Action("mutated")
	got[0].Actions[0].Scopes[0] = ScopeType("mutated")
	if catalogue[0].Actions[0].Action == Action("mutated") ||
		catalogue[0].Actions[0].Scopes[0] == ScopeType("mutated") {
		t.Errorf(
			"%s mutating the returned catalogue mutated the source of truth",
			prefix,
		)
	}
}

func TestResourcePermissions(t *testing.T) {
	// GIVEN: a set of Resources to look up in the catalogue.
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
				{ActionCreate, globalOnly},
				{ActionRead, allScopes},
				{ActionUpdate, allScopes},
				{ActionDelete, allScopes},
			},
		},
		{
			name:        "known/service_order (global only)",
			resource:    ResourceServiceOrder,
			wantActions: []ActionPermission{{ActionUpdate, globalOnly}},
		},
		{
			name:        "known/service_action (execute, all scopes)",
			resource:    ResourceServiceAction,
			wantActions: []ActionPermission{{ActionExecute, allScopes}},
		},
		{
			name:        "known/version_refresh (execute, all scopes)",
			resource:    ResourceVersionRefresh,
			wantActions: []ActionPermission{{ActionExecute, allScopes}},
		},
		{
			name:        "known/notify (global only)",
			resource:    ResourceNotify,
			wantActions: []ActionPermission{{ActionExecute, globalOnly}},
		},
		{
			name:        "known/config (read, global only)",
			resource:    ResourceConfig,
			wantActions: []ActionPermission{{ActionRead, globalOnly}},
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

			// WHEN: resourcePermissions is called.
			got := resourcePermissions(tc.resource)

			prefix := fmt.Sprintf(
				"%s\nresourcePermissions(%q)",
				packageName, tc.resource,
			)

			// THEN: an unknown Resource yields nil.
			if tc.wantNil {
				if got != nil {
					t.Fatalf(
						"%s mismatch\ngot:  %+v\nwant: nil",
						prefix, *got,
					)
				}
				return
			}

			// AND: a known Resource yields its catalogue entry.
			if got == nil {
				t.Fatalf(
					"%s mismatch\ngot:  nil\nwant: catalogue entry",
					prefix,
				)
			}
			if got.Resource != tc.resource {
				t.Errorf(
					"%s Resource mismatch\ngot:  %q\nwant: %q",
					prefix, got.Resource, tc.resource,
				)
			}
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got.Actions,
				tc.wantActions,
				func(a, b ActionPermission) bool {
					return a.Action == b.Action && slices.Equal(a.Scopes, b.Scopes)
				},
				prefix,
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestPermission_Valid(t *testing.T) {
	// GIVEN: a set of Permissions.
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

			prefix := fmt.Sprintf(
				"%s\nPermission{%q, %q}.Valid()",
				packageName, tc.permission.Resource, tc.permission.Action,
			)

			// WHEN: Valid is called.
			got := tc.permission.Valid()

			// THEN: the result matches expectations.
			if got != tc.want {
				t.Errorf(
					"%s result mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestGrant_Valid(t *testing.T) {
	// GIVEN: a set of Grants.
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
			name: "invalid/permission not in catalogue",
			grant: Grant{
				Permission: Permission{Resource: ResourceConfig, Action: ActionDelete},
				Scope:      Scope{Type: ScopeGlobal},
			},
			want: false,
		},
		{
			name: "invalid/scope type unsupported by resource (service_order:service-scoped)",
			grant: Grant{
				Permission: Permission{Resource: ResourceServiceOrder, Action: ActionUpdate},
				Scope:      Scope{Type: ScopeService, Ref: "argus"},
			},
			want: false,
		},
		{
			name: "invalid/unknown scope type",
			grant: Grant{
				Permission: Permission{Resource: ResourceService, Action: ActionRead},
				Scope:      Scope{Type: ScopeType("unknown")},
			},
			want: false,
		},
		{
			name: "invalid/global grant with a ref",
			grant: Grant{
				Permission: Permission{Resource: ResourceService, Action: ActionRead},
				Scope:      Scope{Type: ScopeGlobal, Ref: "argus"},
			},
			want: false,
		},
		{
			name: "invalid/service-scoped grant without a ref",
			grant: Grant{
				Permission: Permission{Resource: ResourceService, Action: ActionRead},
				Scope:      Scope{Type: ScopeService},
			},
			want: false,
		},
		{
			name: "invalid/service_tag-scoped grant without a ref",
			grant: Grant{
				Permission: Permission{Resource: ResourceService, Action: ActionRead},
				Scope:      Scope{Type: ScopeServiceTag},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf(
				"%s\nGrant{resource=%q, action=%q, type=%q, ref=%q}.Valid()",
				packageName,
				tc.grant.Resource, tc.grant.Action,
				tc.grant.Scope.Type, tc.grant.Scope.Ref,
			)

			// WHEN: Valid is called.
			got := tc.grant.Valid()

			// THEN: the result matches expectations.
			if got != tc.want {
				t.Errorf(
					"%s result mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want,
				)
			}
		})
	}
}
