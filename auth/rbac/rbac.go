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

// Package rbac provides the role-based access control model for Argus:
// strongly-typed resources, actions and scopes, and the deterministic
// grant-only permission evaluator.
package rbac

import "slices"

// Resource enumerates the protected resource kinds of the API surface.
type Resource string

// Resources.

const (
	ResourceService        Resource = "service"         // Service configs.
	ResourceServiceOrder   Resource = "service_order"   // Dashboard ordering.
	ResourceServiceAction  Resource = "service_action"  // Approve/skip, run webhooks/commands.
	ResourceVersionRefresh Resource = "version_refresh" // On-demand latest/deployed version refresh.
	ResourceNotify         Resource = "notify"          // Notifier testing.
	ResourceConfig         Resource = "config"          // Global config view.
)

// Action enumerates the operations that can be granted on a [Resource].
type Action string

// Actions.
const (
	ActionCreate  Action = "create"
	ActionRead    Action = "read"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionExecute Action = "execute"
)

// ScopeType enumerates how far a [Grant] reaches.
type ScopeType string

// Scope types.

const (
	ScopeGlobal     ScopeType = "global"      // Every service/target.
	ScopeService    ScopeType = "service"     // A single service (Ref = service ID).
	ScopeServiceTag ScopeType = "service_tag" // Services sharing a dashboard tag (Ref = tag).
)

// Scope bounds a [Grant] to a set of targets.
type Scope struct {
	Type ScopeType `json:"type" yaml:"type"`
	Ref  string    `json:"ref,omitzero" yaml:"ref,omitzero"` // Service ID (Empty for global scope).
}

// Permission pairs a [Resource] with an [Action].
type Permission struct {
	Resource Resource `json:"resource" yaml:"resource"`
	Action   Action   `json:"action" yaml:"action"`
}

// Grant is a [Permission] bounded by a [Scope].
type Grant struct {
	Permission `json:",inline" yaml:",inline"`
	Scope      Scope `json:"scope" yaml:"scope"`
}

// ActionPermission pairs an [Action] with the [ScopeType]s it may be granted at.
type ActionPermission struct {
	Action Action      `json:"action" yaml:"action"`
	Scopes []ScopeType `json:"scopes" yaml:"scopes"`
}

// ResourcePermissions describes the valid [ActionPermission]s of a [Resource].
type ResourcePermissions struct {
	Resource Resource           `json:"name" yaml:"name"`
	Actions  []ActionPermission `json:"actions" yaml:"actions"`
}

// allScopes is a list of all [ScopeType]s.
var allScopes = []ScopeType{ScopeGlobal, ScopeService, ScopeServiceTag}

// globalOnly is the scope list for [Action]s that may only be global-scoped.
var globalOnly = []ScopeType{ScopeGlobal}

// catalogue is the single source of truth for valid
// ([Resource], [][ActionPermission]) combinations.
var catalogue = []ResourcePermissions{
	{
		ResourceService, []ActionPermission{
			{ActionCreate, globalOnly},
			{ActionRead, allScopes},
			{ActionUpdate, allScopes},
			{ActionDelete, allScopes},
		},
	},
	{
		ResourceServiceOrder, []ActionPermission{
			{ActionUpdate, globalOnly},
		},
	},
	{
		ResourceServiceAction, []ActionPermission{
			{ActionExecute, allScopes},
		},
	},
	{
		ResourceVersionRefresh, []ActionPermission{
			{ActionExecute, allScopes},
		},
	},
	{
		ResourceNotify, []ActionPermission{
			{ActionExecute, globalOnly},
		},
	},
	{
		ResourceConfig, []ActionPermission{
			{ActionRead, globalOnly},
		},
	},
}

// Catalogue returns a copy of the valid permission matrix ([catalogue]).
func Catalogue() []ResourcePermissions {
	result := make([]ResourcePermissions, len(catalogue))
	for i, rp := range catalogue {
		actions := make([]ActionPermission, len(rp.Actions))
		for j, ap := range rp.Actions {
			actions[j] = ActionPermission{
				Action: ap.Action,
				Scopes: append([]ScopeType(nil), ap.Scopes...),
			}
		}
		result[i] = ResourcePermissions{Resource: rp.Resource, Actions: actions}
	}
	return result
}

// resourcePermissions returns the [catalogue] entry for resource (nil if unknown).
func resourcePermissions(resource Resource) *ResourcePermissions {
	for i := range catalogue {
		if catalogue[i].Resource == resource {
			return &catalogue[i]
		}
	}
	return nil
}

// scopesFor returns the [ScopeType]s valid for action on this resource
// The bool reports whether the action is valid for the resource.
func (rp *ResourcePermissions) scopesFor(action Action) ([]ScopeType, bool) {
	for i := range rp.Actions {
		if rp.Actions[i].Action == action {
			return rp.Actions[i].Scopes, true
		}
	}
	return nil, false
}

// Valid reports whether the Permission ([Resource], [Action] pair) exists in the [catalogue].
func (p Permission) Valid() bool {
	rp := resourcePermissions(p.Resource)
	if rp == nil {
		return false
	}
	_, ok := rp.scopesFor(p.Action)
	return ok
}

// Valid reports whether the [Grant] is well-formed:
// a valid [Permission], a [ScopeType] the [Permission] supports,
// and a [Scope.Ref] that is empty only when the [ScopeType] is global.
func (g Grant) Valid() bool {
	if !g.Permission.Valid() {
		return false
	}

	rp := resourcePermissions(g.Resource)
	scopes, _ := rp.scopesFor(g.Action) // ok guaranteed by Permission.Valid above.
	if !slices.Contains(scopes, g.Scope.Type) {
		return false
	}

	// Global grants carry no Ref; scoped grants require one.
	if g.Scope.Type == ScopeGlobal {
		return g.Scope.Ref == ""
	}
	return g.Scope.Ref != ""
}
