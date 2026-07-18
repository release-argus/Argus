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
	"testing"
)

func TestNewPermissionSet(t *testing.T) {
	// GIVEN: grant slices of varying validity.
	tests := []struct {
		name        string
		grants      []Grant
		wantGlobal  int
		wantService int
		wantTag     int
	}{
		{
			name:   "empty",
			grants: []Grant{},
		},
		{
			name:   "nil",
			grants: nil,
		},
		{
			name: "one grant of each scope",
			grants: []Grant{
				globalGrant(ResourceService, ActionRead),
				serviceGrant(ResourceService, ActionUpdate, "argus"),
				tagGrant(ResourceServiceAction, ActionExecute, "prod"),
			},
			wantGlobal:  1,
			wantService: 1,
			wantTag:     1,
		},
		{
			name: "duplicate grants collapse",
			grants: []Grant{
				globalGrant(ResourceService, ActionRead),
				globalGrant(ResourceService, ActionRead),
				serviceGrant(ResourceService, ActionUpdate, "argus"),
				serviceGrant(ResourceService, ActionUpdate, "argus"),
				tagGrant(ResourceServiceAction, ActionExecute, "prod"),
				tagGrant(ResourceServiceAction, ActionExecute, "prod"),
			},
			wantGlobal:  1,
			wantService: 1,
			wantTag:     1,
		},
		{
			name: "malformed grants are skipped (fail closed)",
			grants: []Grant{
				// Unknown scope type.
				{
					Permission: Permission{Resource: ResourceService, Action: ActionRead},
					Scope:      Scope{Type: ScopeType("environment"), Ref: "prod"},
				},
				// Unknown resource.
				{
					Permission: Permission{Resource: Resource("unknown"), Action: ActionRead},
					Scope:      Scope{Type: ScopeGlobal},
				},
				// Scoped grant without a ref.
				{
					Permission: Permission{Resource: ResourceService, Action: ActionRead},
					Scope:      Scope{Type: ScopeService},
				},
			},
			wantGlobal:  0,
			wantService: 0,
			wantTag:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: NewPermissionSet is called.
			ps := NewPermissionSet(tc.grants)

			prefix := fmt.Sprintf("%s\nNewPermissionSet()", packageName)

			// THEN: only valid grants were indexed, into their scope's index.
			if len(ps.global) != tc.wantGlobal {
				t.Errorf(
					"%s global index size mismatch\ngot:  %d\nwant: %d",
					prefix, len(ps.global), tc.wantGlobal,
				)
			}
			if len(ps.services) != tc.wantService {
				t.Errorf(
					"%s service index size mismatch\ngot:  %d\nwant: %d",
					prefix, len(ps.services), tc.wantService,
				)
			}
			if len(ps.tags) != tc.wantTag {
				t.Errorf(
					"%s tag index size mismatch\ngot:  %d\nwant: %d",
					prefix, len(ps.tags), tc.wantTag,
				)
			}
		})
	}
}

func TestPermissionSet_Allowed(t *testing.T) {
	// GIVEN: PermissionSets built from grants, and evaluation targets.
	tests := []struct {
		name     string
		grants   []Grant
		resource Resource
		action   Action
		target   *Target
		want     bool
	}{
		{
			name:     "global grant allows, no target",
			grants:   []Grant{globalGrant(ResourceService, ActionRead)},
			resource: ResourceService, action: ActionRead,
			target: nil,
			want:   true,
		},
		{
			name:     "global grant allows any target",
			grants:   []Grant{globalGrant(ResourceService, ActionRead)},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "anything"},
			want:   true,
		},
		{
			name:     "no grants denies",
			grants:   nil,
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "argus"},
			want:   false,
		},
		{
			name:     "grant for different action denies",
			grants:   []Grant{globalGrant(ResourceService, ActionRead)},
			resource: ResourceService, action: ActionDelete,
			target: nil,
			want:   false,
		},
		{
			name:     "grant for different resource denies",
			grants:   []Grant{globalGrant(ResourceService, ActionRead)},
			resource: ResourceConfig, action: ActionRead,
			target: nil,
			want:   false,
		},
		{
			name:     "service-scoped grant allows matching service",
			grants:   []Grant{serviceGrant(ResourceService, ActionUpdate, "argus")},
			resource: ResourceService, action: ActionUpdate,
			target: &Target{ServiceID: "argus"},
			want:   true,
		},
		{
			name:     "service-scoped grant denies other service",
			grants:   []Grant{serviceGrant(ResourceService, ActionUpdate, "argus")},
			resource: ResourceService, action: ActionUpdate,
			target: &Target{ServiceID: "other"},
			want:   false,
		},
		{
			name:     "service-scoped grant denies nil target",
			grants:   []Grant{serviceGrant(ResourceService, ActionUpdate, "argus")},
			resource: ResourceService, action: ActionUpdate,
			target: nil,
			want:   false,
		},
		{
			name:     "service-scoped grant denies target without service ID",
			grants:   []Grant{serviceGrant(ResourceService, ActionUpdate, "argus")},
			resource: ResourceService, action: ActionUpdate,
			target: &Target{},
			want:   false,
		},
		{
			name:     "tag-scoped grant allows service with matching tag",
			grants:   []Grant{tagGrant(ResourceServiceAction, ActionExecute, "prod")},
			resource: ResourceServiceAction, action: ActionExecute,
			target: &Target{ServiceID: "argus", Tags: []string{"web", "prod"}},
			want:   true,
		},
		{
			name:     "tag-scoped grant denies service without matching tag",
			grants:   []Grant{tagGrant(ResourceServiceAction, ActionExecute, "prod")},
			resource: ResourceServiceAction, action: ActionExecute,
			target: &Target{ServiceID: "argus", Tags: []string{"web", "staging"}},
			want:   false,
		},
		{
			name:     "tag-scoped grant denies service with no tags",
			grants:   []Grant{tagGrant(ResourceServiceAction, ActionExecute, "prod")},
			resource: ResourceServiceAction, action: ActionExecute,
			target: &Target{ServiceID: "argus"},
			want:   false,
		},
		{
			name: "multiple groups merge/union allows from either",
			grants: []Grant{
				globalGrant(ResourceService, ActionRead),
				serviceGrant(ResourceService, ActionUpdate, "argus"),
			},
			resource: ResourceService, action: ActionUpdate,
			target: &Target{ServiceID: "argus"},
			want:   true,
		},
		{
			name: "multiple groups merge/neither grants the pair",
			grants: []Grant{
				globalGrant(ResourceService, ActionRead),
				serviceGrant(ResourceService, ActionUpdate, "argus"),
			},
			resource: ResourceService, action: ActionDelete,
			target: &Target{ServiceID: "argus"},
			want:   false,
		},
		{
			name: "global and scoped grant for same permission/global wins for other services",
			grants: []Grant{
				serviceGrant(ResourceService, ActionRead, "argus"),
				globalGrant(ResourceService, ActionRead),
			},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "other"},
			want:   true,
		},
		{
			name: "malformed grant (unknown scope type) fails closed",
			grants: []Grant{
				{Permission: Permission{Resource: ResourceService, Action: ActionRead},
					Scope: Scope{Type: ScopeType("environment"), Ref: "prod"}},
			},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "argus", Tags: []string{"prod"}},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps := NewPermissionSet(tc.grants)

			// WHEN: Allowed is called.
			got := ps.Allowed(tc.resource, tc.action, tc.target)

			prefix := fmt.Sprintf(
				"%s\nPermissionSet.Allowed(resource=%q, action=%q, target=%q)",
				packageName, tc.resource, tc.action, tc.target,
			)

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

func TestPermissionSet_Allowed__nil(t *testing.T) {
	// GIVEN: a nil PermissionSet.
	var ps *PermissionSet

	// WHEN: Allowed is called on it.
	got := ps.Allowed(ResourceService, ActionRead, nil)

	prefix := fmt.Sprintf("%s\n(*PermissionSet)(nil).Allowed()", packageName)

	// THEN: it denies rather than panicking.
	if got {
		t.Errorf(
			"%s nil PermissionSet should deny\ngot:  %t\nwant: false",
			prefix, got,
		)
	}
}

func TestPermissionSet_Allowed__determinism(t *testing.T) {
	// GIVEN: the same grants in different orders, with duplicates.
	base := []Grant{
		globalGrant(ResourceService, ActionRead),
		serviceGrant(ResourceService, ActionUpdate, "argus"),
		tagGrant(ResourceServiceAction, ActionExecute, "prod"),
		serviceGrant(ResourceVersionRefresh, ActionExecute, "other"),
	}
	reversed := make([]Grant, len(base))
	for i, grant := range base {
		reversed[len(base)-1-i] = grant
	}
	withDuplicates := append(append([]Grant{}, base...), base...)

	checks := []struct {
		resource Resource
		action   Action
		target   *Target
	}{
		{ResourceService, ActionRead, nil},
		{ResourceService, ActionUpdate, &Target{ServiceID: "argus"}},
		{ResourceService, ActionUpdate, &Target{ServiceID: "other"}},
		{ResourceServiceAction, ActionExecute, &Target{ServiceID: "x", Tags: []string{"prod"}}},
		{ResourceVersionRefresh, ActionExecute, &Target{ServiceID: "other"}},
		{ResourceVersionRefresh, ActionExecute, &Target{ServiceID: "argus"}},
	}

	psBase := NewPermissionSet(base)
	psReversed := NewPermissionSet(reversed)
	psDuplicated := NewPermissionSet(withDuplicates)

	prefix := fmt.Sprintf("%s\nPermissionSet.Allowed() determinism", packageName)

	// WHEN: the same evaluations run against each ordering.
	for _, check := range checks {
		want := psBase.Allowed(check.resource, check.action, check.target)

		// THEN: grant order and duplication never change the outcome.
		if got := psReversed.Allowed(check.resource, check.action, check.target); got != want {
			t.Errorf(
				"%s reversed grant order changed the outcome for (%q, %q)\ngot:  %t\nwant: %t",
				prefix, check.resource, check.action, got, want,
			)
		}
		if got := psDuplicated.Allowed(check.resource, check.action, check.target); got != want {
			t.Errorf(
				"%s duplicated grants changed the outcome for (%q, %q)\ngot:  %t\nwant: %t",
				prefix, check.resource, check.action, got, want,
			)
		}
	}
}

// TestPermissionSet__allScopeTypesEnforced guards against a new [ScopeType]
// being added to [allScopes] without a matching case in [NewPermissionSet] and
// [PermissionSet.Allowed], which would silently make grants at that scope unenforceable.
func TestPermissionSet__allScopeTypesEnforced(t *testing.T) {
	// GIVEN: a resource/action valid at every scope.
	const resource, action = ResourceService, ActionRead

	for _, scope := range allScopes {
		t.Run(string(scope), func(t *testing.T) {
			// AND: a grant at this scope, and a target it should match.
			grant := Grant{
				Permission: Permission{Resource: resource, Action: action},
				Scope:      Scope{Type: scope},
			}
			target := &Target{ServiceID: "svc-1", Tags: []string{"prod"}}
			switch scope {
			case ScopeService:
				grant.Scope.Ref = target.ServiceID
			case ScopeServiceTag:
				grant.Scope.Ref = target.Tags[0]
			}

			// WHEN: the grant is indexed and evaluated against a matching target.
			// THEN: the scope is enforced.
			if !NewPermissionSet([]Grant{grant}).Allowed(resource, action, target) {
				t.Errorf(
					"%s\nscope %q is in allScopes but not enforced by NewPermissionSet/Allowed",
					packageName, scope,
				)
			}
		})
	}
}

func TestPermissionSet_Allowed__boundaryConditions(t *testing.T) {
	// GIVEN: grants covering precedence, duplication, case sensitivity, and action isolation.
	tests := []struct {
		name     string
		grants   []Grant
		resource Resource
		action   Action
		target   *Target
		want     bool
	}{
		{
			name: "conflicting groups and broad grant wins over a narrower one (union)",
			grants: []Grant{
				serviceGrant(ResourceService, ActionRead, "argus"),
				globalGrant(ResourceService, ActionRead),
			},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "unrelated"},
			want:   true,
		},
		{
			name: "identical grants from multiple groups collapse",
			grants: []Grant{
				serviceGrant(ResourceService, ActionRead, "argus"),
				serviceGrant(ResourceService, ActionRead, "argus"),
			},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "argus"},
			want:   true,
		},
		{
			name:     "service IDs match case-sensitively",
			grants:   []Grant{serviceGrant(ResourceService, ActionRead, "Argus")},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "argus"},
			want:   false,
		},
		{
			name:     "tags match case-sensitively",
			grants:   []Grant{tagGrant(ResourceService, ActionRead, "Prod")},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "argus", Tags: []string{"prod"}},
			want:   false,
		},
		{
			name:     "empty tag in the target never matches",
			grants:   []Grant{tagGrant(ResourceService, ActionRead, "prod")},
			resource: ResourceService, action: ActionRead,
			target: &Target{ServiceID: "argus", Tags: []string{""}},
			want:   false,
		},
		{
			name:     "scoped grant for one action does not leak to another",
			grants:   []Grant{serviceGrant(ResourceService, ActionUpdate, "argus")},
			resource: ResourceService, action: ActionDelete,
			target: &Target{ServiceID: "argus"},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps := NewPermissionSet(tc.grants)

			// WHEN: Allowed is called.
			got := ps.Allowed(tc.resource, tc.action, tc.target)

			prefix := fmt.Sprintf(
				"%s\nPermissionSet.Allowed(%q, %q) edge case",
				packageName, tc.resource, tc.action,
			)

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

func TestPermissionSet_AllowedAnyScope(t *testing.T) {
	// GIVEN: PermissionSets built from grants of differing scopes.
	tests := []struct {
		name     string
		grants   []Grant
		resource Resource
		action   Action
		want     bool
	}{
		{
			name:     "valid/global grant allows",
			grants:   []Grant{globalGrant(ResourceService, ActionRead)},
			resource: ResourceService, action: ActionRead,
			want: true,
		},
		{
			name:     "valid/service-scoped grant allows without a target",
			grants:   []Grant{serviceGrant(ResourceService, ActionRead, "argus")},
			resource: ResourceService, action: ActionRead,
			want: true,
		},
		{
			name:     "valid/tag-scoped grant allows without a target",
			grants:   []Grant{tagGrant(ResourceService, ActionRead, "prod")},
			resource: ResourceService, action: ActionRead,
			want: true,
		},
		{
			name:     "invalid/no grants denies",
			grants:   nil,
			resource: ResourceService, action: ActionRead,
			want: false,
		},
		{
			name:     "invalid/grant for a different action denies",
			grants:   []Grant{serviceGrant(ResourceService, ActionRead, "argus")},
			resource: ResourceService, action: ActionDelete,
			want: false,
		},
		{
			name:     "invalid/grant for a different resource denies",
			grants:   []Grant{serviceGrant(ResourceService, ActionRead, "argus")},
			resource: ResourceConfig, action: ActionRead,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps := NewPermissionSet(tc.grants)

			// WHEN: AllowedAnyScope is called.
			got := ps.AllowedAnyScope(tc.resource, tc.action)

			prefix := fmt.Sprintf(
				"%s\nPermissionSet.AllowedAnyScope(%q, %q)",
				packageName, tc.resource, tc.action,
			)

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

func TestPermissionSet_AllowedAnyScope__nil(t *testing.T) {
	// GIVEN: a nil PermissionSet.
	var ps *PermissionSet

	// WHEN: AllowedAnyScope is called on it.
	got := ps.AllowedAnyScope(ResourceService, ActionRead)

	prefix := fmt.Sprintf("%s\n(*PermissionSet)(nil).AllowedAnyScope()", packageName)

	// THEN: it denies rather than panicking.
	if got {
		t.Errorf(
			"%s nil PermissionSet should deny\ngot:  %t\nwant: false",
			prefix, got,
		)
	}
}
