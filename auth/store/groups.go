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

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/release-argus/Argus/auth/rbac"
)

// Group is a named set of users carrying permission grants.
type Group struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	System      bool         `json:"system"`             // Protected group with modification rails (admin).
	SeedKey     string       `json:"seed_key,omitempty"` // Immutable seed identity; empty for user-created groups.
	Members     int          `json:"members"`
	Grants      []rbac.Grant `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// GroupPatch is a partial update of a Group; nil fields stay unchanged.
type GroupPatch struct {
	Name        *string       // New name (rejected for system groups).
	Description *string       // New description.
	Grants      *[]rbac.Grant // Replace-set of grants (rejected for admin).
}

// CreateGroup creates a group with the given grants.
func (s *Store) CreateGroup(
	ctx context.Context,
	name, description string,
	grants []rbac.Grant,
) (*Group, error) {
	id := newID()

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := rejectDuplicateGroupName(ctx, tx, name, ""); err != nil {
			return err
		}

		now := timeNow().Format(timeFormat)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO groups (id, name, description, system, created_at, updated_at)
			VALUES (?, ?, ?, 0, ?, ?);`,
			id, name, description, now, now,
		); err != nil {
			if isUniqueViolation(err) {
				return ErrGroupNameTaken
			}
			return fmt.Errorf("insert group: %w", err)
		}

		return setGroupGrants(ctx, tx, id, grants)
	}); err != nil {
		return nil, err
	}

	return s.GroupByID(ctx, id)
}

// Groups lists every group, ordered by name, with member counts and grants.
func (s *Store) Groups(ctx context.Context) ([]Group, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.description, g.system, COALESCE(g.seed_key, ''),
			g.created_at, g.updated_at, COUNT(ug.user_id)
		FROM groups g
		LEFT JOIN user_groups ug ON ug.group_id = g.id
		GROUP BY g.id
		ORDER BY g.name;`,
	)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	groups, err := scanAll(rows, "groups", scanGroup)
	if err != nil {
		return nil, err
	}
	index := make(map[string]int, len(groups)) // Group ID -> index in groups.
	for i, group := range groups {
		index[group.ID] = i
	}

	// Attach grants.
	if err := s.queryGrants(ctx, `
		SELECT gp.group_id, p.resource, p.action, gp.scope_type, gp.scope_ref
		FROM group_permissions gp
		JOIN permissions p ON p.id = gp.permission_id
		ORDER BY p.resource, p.action, gp.scope_type, gp.scope_ref;`,
		func(groupID string, grant rbac.Grant) {
			if i, ok := index[groupID]; ok {
				groups[i].Grants = append(groups[i].Grants, grant)
			}
		}); err != nil {
		return nil, fmt.Errorf("attach grants: %w", err)
	}

	return groups, nil
}

// GroupByID returns the group with id ([ErrNotFound] if absent).
func (s *Store) GroupByID(ctx context.Context, id string) (*Group, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT g.id, g.name, g.description, g.system, COALESCE(g.seed_key, ''),
			g.created_at, g.updated_at, COUNT(ug.user_id)
		FROM groups g
		LEFT JOIN user_groups ug ON ug.group_id = g.id
		WHERE g.id = ?
		GROUP BY g.id;`,
		id,
	)

	group, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := s.queryGrants(ctx, `
		SELECT gp.group_id, p.resource, p.action, gp.scope_type, gp.scope_ref
		FROM group_permissions gp
		JOIN permissions p ON p.id = gp.permission_id
		WHERE gp.group_id = ?
		ORDER BY p.resource, p.action, gp.scope_type, gp.scope_ref;`,
		func(_ string, grant rbac.Grant) {
			group.Grants = append(group.Grants, grant)
		}, id); err != nil {
		return nil, fmt.Errorf("list group's grants: %w", err)
	}

	return group, nil
}

// UpdateGroup applies patch to the group with id, enforcing the system rails:
// system groups cannot be renamed, and admin grants cannot be changed.
func (s *Store) UpdateGroup(
	ctx context.Context,
	id string,
	patch GroupPatch,
) (*Group, error) {
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		var (
			name   string
			system bool
		)
		if err := tx.QueryRowContext(ctx,
			`SELECT name, system FROM groups WHERE id = ?;`, id,
		).Scan(&name, &system); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("load group: %w", err)
		}

		// Rails.
		if patch.Name != nil && *patch.Name != name && system {
			return fmt.Errorf("%w: cannot rename a system group", ErrSystemGroup)
		}
		if patch.Grants != nil && system && name == GroupAdmin {
			return fmt.Errorf("%w: admin grants are fixed", ErrSystemGroup)
		}

		upd := newRowUpdate()
		if patch.Name != nil && *patch.Name != name {
			if err := rejectDuplicateGroupName(ctx, tx, *patch.Name, id); err != nil {
				return err
			}
			upd.set("name", *patch.Name)
		}
		if patch.Description != nil {
			upd.set("description", *patch.Description)
		}
		if err := upd.exec(ctx, tx, "groups", id); err != nil {
			return fmt.Errorf("update group: %w", err)
		}

		if patch.Grants != nil {
			if err := setGroupGrants(ctx, tx, id, *patch.Grants); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return s.GroupByID(ctx, id)
}

// DeleteGroup removes the group with id and its memberships/grants.
// System groups cannot be deleted.
func (s *Store) DeleteGroup(ctx context.Context, id string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		var system bool
		if err := tx.QueryRowContext(ctx,
			`SELECT system FROM groups WHERE id = ?;`, id,
		).Scan(&system); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("load group: %w", err)
		}
		if system {
			return fmt.Errorf("%w: cannot delete a system group", ErrSystemGroup)
		}

		// Explicit cascade.
		for _, statement := range []string{
			`DELETE FROM user_groups WHERE group_id = ?;`,
			`DELETE FROM group_permissions WHERE group_id = ?;`,
			`DELETE FROM groups WHERE id = ?;`,
		} {
			if _, err := tx.ExecContext(ctx, statement, id); err != nil {
				return fmt.Errorf("delete group: %w", err)
			}
		}

		return nil
	})
}

// UserIDsInGroup returns the IDs of the users belonging to the group with id.
func (s *Store) UserIDsInGroup(ctx context.Context, id string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT user_id
		FROM user_groups
		WHERE group_id = ?
		ORDER BY user_id;`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("list group's users: %w", err)
	}
	defer rows.Close()

	return scanColumn[string](rows, "group's users")
}

// GrantsForUser returns the union of the grants of every group userID
// belongs to (Disabled users get no grants).
func (s *Store) GrantsForUser(ctx context.Context, userID string) ([]rbac.Grant, error) {
	grants := []rbac.Grant{}
	if err := s.queryGrants(ctx, `
		SELECT gp.group_id, p.resource, p.action, gp.scope_type, gp.scope_ref
		FROM user_groups ug
		JOIN users u              ON u.id = ug.user_id
		JOIN group_permissions gp ON gp.group_id = ug.group_id
		JOIN permissions p        ON p.id = gp.permission_id
		WHERE ug.user_id = ? AND u.enabled = 1
		ORDER BY p.resource, p.action, gp.scope_type, gp.scope_ref;`,
		func(_ string, grant rbac.Grant) {
			grants = append(grants, grant)
		}, userID); err != nil {
		return nil, fmt.Errorf("list user's grants: %w", err)
	}

	return grants, nil
}

// scopedGrantKey identifies one grant row within a scope.
type scopedGrantKey struct {
	groupID      string
	permissionID int64
}

// ServiceScopeMove records what a rename repointed, so [Store.UndoServiceScopeMove]
// can put it back without disturbing grants the target already held.
type ServiceScopeMove struct {
	from, to string
	moved    []scopedGrantKey // Repointed by the rename.
	merged   []scopedGrantKey // Absorbed into a row the target already held.
}

// From returns the service ID the grants were moved from.
func (m *ServiceScopeMove) From() string { return m.from }

// To returns the service ID the grants were moved to.
func (m *ServiceScopeMove) To() string { return m.to }

// scopedGrantKeys returns the grants of a service scope, keyed within it.
func scopedGrantKeys(
	ctx context.Context,
	tx *sql.Tx,
	serviceID string,
) ([]scopedGrantKey, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT group_id, permission_id
		FROM group_permissions
		WHERE scope_type = ? AND scope_ref = ?;`,
		string(rbac.ScopeService), serviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list service-scoped grants: %w", err)
	}
	defer rows.Close()

	var keys []scopedGrantKey
	for rows.Next() {
		var key scopedGrantKey
		if err := rows.Scan(&key.groupID, &key.permissionID); err != nil {
			return nil, fmt.Errorf("scan service-scoped grant: %w", err)
		}
		keys = append(keys, key)
	}
	if err := rowsErr(rows); err != nil {
		return nil, fmt.Errorf("iterate service-scoped grants: %w", err)
	}

	return keys, nil
}

// RenameServiceScopeRefs points service-scoped grants at a service's new ID,
// returning what it moved. Grants the new ID already held are absorbed, since
// after the rename both IDs name the same service.
func (s *Store) RenameServiceScopeRefs(
	ctx context.Context, oldServiceID, newServiceID string,
) (*ServiceScopeMove, error) {
	move := &ServiceScopeMove{from: oldServiceID, to: newServiceID}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		moved, err := scopedGrantKeys(ctx, tx, oldServiceID)
		if err != nil {
			return err
		}
		held, err := scopedGrantKeys(ctx, tx, newServiceID)
		if err != nil {
			return err
		}

		heldKeys := make(map[scopedGrantKey]bool, len(held))
		for _, key := range held {
			heldKeys[key] = true
		}
		for _, key := range moved {
			if heldKeys[key] {
				move.merged = append(move.merged, key)
			}
		}
		move.moved = moved

		if _, err := tx.ExecContext(ctx, `
			UPDATE OR REPLACE group_permissions
			SET scope_ref = ?
			WHERE scope_type = ? AND scope_ref = ?;`,
			newServiceID,
			string(rbac.ScopeService), oldServiceID,
		); err != nil {
			return fmt.Errorf("rename service scope refs: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return move, nil
}

// UndoServiceScopeMove reverses a [Store.RenameServiceScopeRefs], restoring the
// rows it absorbed. Only the rows that move recorded are touched, so grants the
// target held beforehand keep their scope.
func (s *Store) UndoServiceScopeMove(ctx context.Context, move *ServiceScopeMove) error {
	if move == nil || len(move.moved) == 0 {
		return nil
	}

	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, key := range move.moved {
			if _, err := tx.ExecContext(ctx, `
				UPDATE group_permissions
				SET scope_ref = ?
				WHERE scope_type = ? AND scope_ref = ?
					AND group_id = ? AND permission_id = ?;`,
				move.from,
				string(rbac.ScopeService), move.to,
				key.groupID, key.permissionID,
			); err != nil {
				return fmt.Errorf("undo service scope move: %w", err)
			}
		}
		// Absorbed rows were deleted by the rename, so re-create them.
		for _, key := range move.merged {
			if _, err := tx.ExecContext(ctx, `
				INSERT OR IGNORE INTO group_permissions
					(group_id, permission_id, scope_type, scope_ref)
				VALUES (?, ?, ?, ?);`,
				key.groupID, key.permissionID, string(rbac.ScopeService), move.to,
			); err != nil {
				return fmt.Errorf("restore absorbed service scope grants: %w", err)
			}
		}
		return nil
	})
}

// DeleteServiceScopeRefs prunes service-scoped grants for a deleted service.
func (s *Store) DeleteServiceScopeRefs(ctx context.Context, serviceID string) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM group_permissions WHERE scope_type = ? AND scope_ref = ?;`,
		string(rbac.ScopeService), serviceID,
	); err != nil {
		return fmt.Errorf("delete service scope refs: %w", err)
	}
	return nil
}

// groupIDBySeedKey returns the ID of the group with the given immutable
// seed_key ([ErrNotFound] if absent).
func groupIDBySeedKey(ctx context.Context, tx *sql.Tx, key string) (string, error) {
	var id string
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM groups WHERE seed_key = ?;`, key,
	).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("look up seed group %q: %w", key, err)
	}
	return id, nil
}

// rejectDuplicateGroupName returns [ErrGroupNameTaken] when another group
// already holds name (case-insensitive). excludeID is the group being renamed,
// so it does not collide with itself (pass "" when creating).
func rejectDuplicateGroupName(
	ctx context.Context,
	tx *sql.Tx,
	name, excludeID string,
) error {
	existingID, err := groupIDByName(ctx, tx, name)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil
	case err != nil:
		return err
	case existingID == excludeID:
		return nil
	default:
		return ErrGroupNameTaken
	}
}

// groupIDByName returns the ID of the group named name ([ErrNotFound] if absent).
func groupIDByName(ctx context.Context, tx *sql.Tx, name string) (string, error) {
	var id string
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM groups WHERE name = ? COLLATE NOCASE;`, name,
	).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("look up group %q: %w", name, err)
	}
	return id, nil
}

// queryGrants runs query and invokes fn for each scanned (group_id, grant) row.
func (s *Store) queryGrants(
	ctx context.Context,
	query string,
	fn func(groupID string, grant rbac.Grant),
	args ...any,
) error {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err //nolint:wrapcheck // Callers add context.
	}
	defer rows.Close()
	return collectGrants(rows, fn)
}

// collectGrants scans (group_id, grant) rows, invoking fn for each.
func collectGrants(rows *sql.Rows, fn func(groupID string, grant rbac.Grant)) error {
	for rows.Next() {
		var (
			groupID string
			grant   rbac.Grant
		)
		if err := scanGrant(rows, &groupID, &grant); err != nil {
			return err
		}
		fn(groupID, grant)
	}
	return rowsErr(rows)
}

// scanGrant scans a (group_id, resource, action, scope_type, scope_ref) row.
func scanGrant(row scanner, groupID *string, grant *rbac.Grant) error {
	var (
		resource  string
		action    string
		scopeType string
	)
	if err := row.Scan(groupID, &resource, &action, &scopeType, &grant.Scope.Ref); err != nil {
		return fmt.Errorf("scan grant: %w", err)
	}
	grant.Resource = rbac.Resource(resource)
	grant.Action = rbac.Action(action)
	grant.Scope.Type = rbac.ScopeType(scopeType)
	return nil
}

// scanGroup scans a groups row with a member count (without grants).
func scanGroup(row scanner) (*Group, error) {
	group := Group{Grants: []rbac.Grant{}}
	if err := row.Scan(
		&group.ID, &group.Name, &group.Description, &group.System, &group.SeedKey,
		timeText{&group.CreatedAt}, timeText{&group.UpdatedAt}, &group.Members,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err //nolint:wrapcheck // Callers translate to [ErrNotFound].
		}
		return nil, fmt.Errorf("scan group: %w", err)
	}

	return &group, nil
}

// setGroupGrants replaces groupID's grants, validating each against the
// permission catalogue ([ErrInvalidGrant] on malformed/unknown grants).
func setGroupGrants(
	ctx context.Context,
	tx *sql.Tx,
	groupID string,
	grants []rbac.Grant,
) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM group_permissions WHERE group_id = ?;`, groupID,
	); err != nil {
		return fmt.Errorf("clear group grants: %w", err)
	}

	return insertGroupGrants(ctx, tx, groupID, grants)
}

// insertGroupGrants adds grants to groupID (keeping existing ones), validating
// each against the permission catalogue ([ErrInvalidGrant] on malformed/unknown
// grants). Duplicates of existing grants are ignored.
func insertGroupGrants(
	ctx context.Context,
	tx *sql.Tx,
	groupID string,
	grants []rbac.Grant,
) error {
	if len(grants) == 0 {
		return nil
	}

	permissionIDs, err := permissionIDsByPair(ctx, tx)
	if err != nil {
		return err
	}

	for _, grant := range grants {
		if !grant.Valid() {
			return fmt.Errorf(
				"%w: %s:%s @ %s/%s",
				ErrInvalidGrant,
				grant.Resource, grant.Action, grant.Scope.Type, grant.Scope.Ref,
			)
		}

		permissionID, ok := permissionIDs[grant.Permission]
		if !ok {
			return fmt.Errorf(
				"look up permission %s:%s: %w",
				grant.Resource, grant.Action, ErrNotFound,
			)
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO group_permissions
				(group_id, permission_id, scope_type, scope_ref)
			VALUES (?, ?, ?, ?);`,
			groupID, permissionID, string(grant.Scope.Type), grant.Scope.Ref,
		); err != nil {
			return fmt.Errorf("insert grant: %w", err)
		}
	}

	return nil
}

// permissionIDsByPair loads the permission catalogue keyed by [rbac.Permission] ([rbac.Resource], [rbac.Action]).
func permissionIDsByPair(
	ctx context.Context,
	tx *sql.Tx,
) (map[rbac.Permission]int64, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, resource, action FROM permissions;`)
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}
	defer rows.Close()

	ids := map[rbac.Permission]int64{}
	for rows.Next() {
		var (
			id       int64
			resource string
			action   string
		)
		if err := rows.Scan(&id, &resource, &action); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		ids[rbac.Permission{
			Resource: rbac.Resource(resource),
			Action:   rbac.Action(action),
		}] = id
	}
	if err := rowsErr(rows); err != nil {
		return nil, fmt.Errorf("iterate permissions: %w", err)
	}

	return ids, nil
}
