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
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestNewPermissionSet(t *testing.T) {
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
				{Resource: ResourceService, Action: ActionRead, Scope: Scope{Type: ScopeType("environment"), Ref: "prod"}},
				// Unknown resource.
				{Resource: Resource("unknown"), Action: ActionRead, Scope: Scope{Type: ScopeGlobal}},
				// Scoped grant without a ref.
				{Resource: ResourceService, Action: ActionRead, Scope: Scope{Type: ScopeService}},
			},
			wantGlobal:  0,
			wantService: 0,
			wantTag:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ps := NewPermissionSet(tc.grants)

			if testErr := test.AssertFields(t, []test.FieldAssertion{
				{Name: "len(PermissionSet.global)", Got: len(ps.global), Want: tc.wantGlobal, Mode: test.CompareEqual},
				{Name: "len(PermissionSet.services)", Got: len(ps.services), Want: tc.wantService, Mode: test.CompareEqual},
				{Name: "len(PermissionSet.tags)", Got: len(ps.tags), Want: tc.wantTag, Mode: test.CompareEqual},
			}, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestPermissionSet_Allowed(t *testing.T) {
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
				{Resource: ResourceService, Action: ActionRead, Scope: Scope{Type: ScopeType("environment"), Ref: "prod"}},
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

			got := ps.Allowed(tc.resource, tc.action, tc.target)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestPermissionSet_Allowed__nil(t *testing.T) {
	var ps *PermissionSet

	got := ps.Allowed(ResourceService, ActionRead, nil)

	if got {
		t.Errorf("mismatch\ngot:  %t\nwant: false", got)
	}
}

func TestPermissionSet_Allowed__determinism(t *testing.T) {
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
		name     string
		resource Resource
		action   Action
		target   *Target
	}{
		{"service:read/no target", ResourceService, ActionRead, nil},
		{"service:update/argus", ResourceService, ActionUpdate, &Target{ServiceID: "argus"}},
		{"service:update/other", ResourceService, ActionUpdate, &Target{ServiceID: "other"}},
		{"service_action:execute/prod tag", ResourceServiceAction, ActionExecute, &Target{ServiceID: "x", Tags: []string{"prod"}}},
		{"version_refresh:execute/other", ResourceVersionRefresh, ActionExecute, &Target{ServiceID: "other"}},
		{"version_refresh:execute/argus", ResourceVersionRefresh, ActionExecute, &Target{ServiceID: "argus"}},
	}

	psBase := NewPermissionSet(base)
	psReversed := NewPermissionSet(reversed)
	psDuplicated := NewPermissionSet(withDuplicates)

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()

			want := psBase.Allowed(check.resource, check.action, check.target)

			if got := psReversed.Allowed(check.resource, check.action, check.target); got != want {
				t.Errorf(
					"reversed: PermissionSet.Allowed() mismatch\ngot:  %t\nwant: %t",
					got, want,
				)
			}
			if got := psDuplicated.Allowed(check.resource, check.action, check.target); got != want {
				t.Errorf(
					"duplicated: PermissionSet.Allowed() mismatch\ngot:  %t\nwant: %t",
					got, want,
				)
			}
		})
	}
}

// TestPermissionSet_Allowed__allScopeTypesEnforced guards against a new [ScopeType]
// being added to [allScopes] without a matching case in [NewPermissionSet] and
// [PermissionSet.Allowed], which would silently make grants at that scope unenforceable.
func TestPermissionSet_Allowed__allScopeTypesEnforced(t *testing.T) {
	const resource, action = ResourceService, ActionRead

	for _, scope := range allScopes {
		t.Run(string(scope), func(t *testing.T) {
			grant := Grant{Resource: resource, Action: action, Scope: Scope{Type: scope}}
			target := &Target{ServiceID: "svc-1", Tags: []string{"prod"}}
			switch scope {
			case ScopeService:
				grant.Scope.Ref = target.ServiceID
			case ScopeServiceTag:
				grant.Scope.Ref = target.Tags[0]
			}

			if !NewPermissionSet([]Grant{grant}).Allowed(resource, action, target) {
				t.Error("PermissionSet.Allowed() mismatch\ngot:  false\nwant: true")
			}
		})
	}
}

func TestPermissionSet_Allowed__boundaryConditions(t *testing.T) {
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

			got := ps.Allowed(tc.resource, tc.action, tc.target)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestPermissionSet_AllowedAnyScope(t *testing.T) {
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

			got := ps.AllowedAnyScope(tc.resource, tc.action)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestPermissionSet_AllowedAnyScope__nil(t *testing.T) {
	var ps *PermissionSet

	got := ps.AllowedAnyScope(ResourceService, ActionRead)

	if got {
		t.Errorf("mismatch\ngot:  %t\nwant: false", got)
	}
}
