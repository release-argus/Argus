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

package rbac

// Target identifies what a request acts on, for scope evaluation.
// A nil/zero Target means the operation has no per-service target
// (e.g. creating a service) and only global grants can allow it.
type Target struct {
	ServiceID string   // ID of the targeted service.
	Tags      []string // Dashboard tags of the targeted service (service_tag scope).
}

// PermissionSet is the union of a user's grants across all their groups,
// indexed for evaluation.
type PermissionSet struct {
	global   map[Permission]bool            // Permissions granted globally.
	services map[Permission]map[string]bool // Permission -> service IDs granted.
	tags     map[Permission]map[string]bool // Permission -> dashboard tags granted.
}

// NewPermissionSet indexes grants for evaluation.
// Malformed grants (per [Grant.Valid]) are skipped.
func NewPermissionSet(grants []Grant) *PermissionSet {
	ps := &PermissionSet{
		global:   make(map[Permission]bool),
		services: make(map[Permission]map[string]bool),
		tags:     make(map[Permission]map[string]bool),
	}

	for _, grant := range grants {
		if !grant.Valid() {
			continue
		}

		switch grant.Scope.Type {
		case ScopeGlobal:
			ps.global[grant.Permission] = true
		case ScopeService:
			if ps.services[grant.Permission] == nil {
				ps.services[grant.Permission] = make(map[string]bool, 1)
			}
			ps.services[grant.Permission][grant.Scope.Ref] = true
		case ScopeServiceTag:
			if ps.tags[grant.Permission] == nil {
				ps.tags[grant.Permission] = make(map[string]bool, 1)
			}
			ps.tags[grant.Permission][grant.Scope.Ref] = true
		}
	}

	return ps
}

// Allowed reports whether the [PermissionSet] grants action on resource for target.
func (ps *PermissionSet) Allowed(resource Resource, action Action, target *Target) bool {
	if ps == nil {
		return false
	}

	permission := Permission{Resource: resource, Action: action}

	// 1. Global.
	if ps.global[permission] {
		return true
	}

	if target == nil || target.ServiceID == "" {
		return false
	}

	// 2. Service.
	if ps.services[permission][target.ServiceID] {
		return true
	}

	// 3. Service Group (tags).
	grantedTags := ps.tags[permission]
	for _, tag := range target.Tags {
		if grantedTags[tag] {
			return true
		}
	}

	return false
}

// AllowedAnyScope reports whether the [PermissionSet] grants action on resource
// under any scope. For endpoints that serve no per-service data, so holding the
// permission for a single service is enough to see it.
func (ps *PermissionSet) AllowedAnyScope(resource Resource, action Action) bool {
	if ps == nil {
		return false
	}

	permission := Permission{Resource: resource, Action: action}

	return ps.global[permission] ||
		len(ps.services[permission]) > 0 ||
		len(ps.tags[permission]) > 0
}
