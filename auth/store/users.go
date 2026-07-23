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
	"slices"
	"strings"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/auth/rbac"
)

// hashPassword derives a password hash (overridable for tests).
// see [password.Hash].
var hashPassword = password.Hash

// UserPatch is a partial update of a User; nil fields stay unchanged.
type UserPatch struct {
	DisplayName  *string   // New display name.
	Email        *string   // New email address.
	Enabled      *bool     // Enable/disable login.
	Groups       *[]string // Replace-set of group names.
	PasswordHash *string   // New (already hashed) password.
}

// CountUsers returns the number of user accounts.
func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users;`,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// Users lists every user, ordered by username, with group memberships.
func (s *Store) Users(ctx context.Context) ([]auth.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, username, display_name, email, enabled, created_at, updated_at
		FROM users
		ORDER BY username;`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users, err := scanAll(rows, "users", scanUser)
	if err != nil {
		return nil, err
	}

	// Attach group memberships.
	memberships, err := s.db.QueryContext(ctx, `
		SELECT ug.user_id, g.name
		FROM user_groups ug
		JOIN groups g ON g.id = ug.group_id
		ORDER BY g.name;`,
	)
	if err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}
	defer memberships.Close()

	groupsByUser := map[string][]string{}
	for memberships.Next() {
		var userID, groupName string
		if err := memberships.Scan(&userID, &groupName); err != nil {
			return nil, fmt.Errorf("scan user group: %w", err)
		}
		groupsByUser[userID] = append(groupsByUser[userID], groupName)
	}
	if err := rowsErr(memberships); err != nil {
		return nil, fmt.Errorf("iterate user groups: %w", err)
	}

	for i := range users {
		users[i].Groups = groupsByUser[users[i].ID]
	}

	return users, nil
}

// UserByID returns the user with id ([ErrNotFound] if absent).
func (s *Store) UserByID(ctx context.Context, id string) (*auth.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, email, enabled, created_at, updated_at
		FROM users
		WHERE id = ?;`,
		id,
	)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	groups, err := s.userGroupNames(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Groups = groups

	return user, nil
}

// UserWithGrants returns the user, their group names, and every grant those
// groups carry, in a single query. Grants come back regardless of
// [auth.User.Enabled], so callers gating on it must check the flag themselves.
func (s *Store) UserWithGrants(
	ctx context.Context,
	id string,
) (*auth.User, []rbac.Grant, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.display_name, u.email, u.enabled, u.created_at, u.updated_at,
		       g.name,
		       p.resource, p.action, gp.scope_type, gp.scope_ref
		FROM users u
		LEFT JOIN user_groups ug       ON ug.user_id  = u.id
		LEFT JOIN groups g             ON g.id        = ug.group_id
		LEFT JOIN group_permissions gp ON gp.group_id = ug.group_id
		LEFT JOIN permissions p        ON p.id        = gp.permission_id
		WHERE u.id = ?
		ORDER BY g.name, p.resource, p.action, gp.scope_type, gp.scope_ref;`,
		id,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("query user with grants: %w", err)
	}
	defer rows.Close()

	var (
		user   *auth.User
		groups []string
		grants []rbac.Grant
	)
	for rows.Next() {
		var (
			row                                              auth.User
			groupName, resource, action, scopeType, scopeRef sql.NullString
		)
		if err := rows.Scan(
			&row.ID, &row.Username, &row.DisplayName, &row.Email, &row.Enabled,
			timeText{&row.CreatedAt}, timeText{&row.UpdatedAt},
			&groupName,
			&resource, &action, &scopeType, &scopeRef,
		); err != nil {
			return nil, nil, fmt.Errorf("scan user with grants: %w", err)
		}

		// The user's columns repeat on every joined row.
		if user == nil {
			user = &row
		}
		// Ordered by name, so a group only ever repeats consecutively.
		if groupName.Valid && (len(groups) == 0 || groups[len(groups)-1] != groupName.String) {
			groups = append(groups, groupName.String)
		}
		// NULL where the group carries no permissions.
		if resource.Valid {
			grants = append(grants, rbac.Grant{
				Permission: rbac.Permission{
					Resource: rbac.Resource(resource.String),
					Action:   rbac.Action(action.String),
				},
				Scope: rbac.Scope{
					Type: rbac.ScopeType(scopeType.String),
					Ref:  scopeRef.String,
				},
			})
		}
	}
	if err := rowsErr(rows); err != nil {
		return nil, nil, fmt.Errorf("iterate user with grants: %w", err)
	}
	if user == nil {
		return nil, nil, ErrNotFound
	}
	user.Groups = groups

	return user, grants, nil
}

// CreateUser creates a user with the given group memberships.
func (s *Store) CreateUser(
	ctx context.Context,
	username, displayName, email, passwordHash string,
	groups []string,
) (*auth.User, error) {
	if passwordHash == "" {
		return nil, ErrNoPassword
	}
	id := newID()

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		now := timeNow().Format(timeFormat)

		if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (
			id,
			username,
			display_name,
			email,
			password_hash,
			enabled,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?);`,
			id,
			username,
			displayName,
			email,
			passwordHash,
			now,
			now,
		); err != nil {
			if isUniqueViolation(err) {
				return ErrUsernameTaken
			}
			return fmt.Errorf("insert user: %w", err)
		}

		return setUserGroups(ctx, tx, id, groups)
	}); err != nil {
		return nil, err
	}

	return s.UserByID(ctx, id)
}

// UpdateUser applies patch to the user with id, enforcing the last-admin rail.
func (s *Store) UpdateUser(
	ctx context.Context,
	id string,
	patch UserPatch,
) (*auth.User, error) {
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		var enabled bool
		if err := tx.QueryRowContext(ctx,
			`SELECT enabled FROM users WHERE id = ?;`, id,
		).Scan(&enabled); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("load user: %w", err)
		}

		// Last-admin rail: an update may not disable, or remove from admin,
		// the last enabled admin.
		disabling := patch.Enabled != nil && !*patch.Enabled && enabled
		demoting := patch.Groups != nil && !slices.ContainsFunc(*patch.Groups,
			func(group string) bool { return strings.EqualFold(group, GroupAdmin) })
		if disabling || demoting {
			if err := rejectIfLastAdmin(ctx, tx, id); err != nil {
				return err
			}
		}

		// Apply the field updates in one statement, always touching updated_at.
		upd := newRowUpdate()
		if patch.DisplayName != nil {
			upd.set("display_name", *patch.DisplayName)
		}
		if patch.Email != nil {
			upd.set("email", *patch.Email)
		}
		if patch.Enabled != nil {
			upd.set("enabled", *patch.Enabled)
		}
		if patch.PasswordHash != nil {
			if *patch.PasswordHash == "" {
				return ErrNoPassword
			}
			upd.set("password_hash", *patch.PasswordHash)
		}
		if err := upd.exec(ctx, tx, "users", id); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		if patch.Groups != nil {
			if err := setUserGroups(ctx, tx, id, *patch.Groups); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return s.UserByID(ctx, id)
}

// DeleteUser removes the user with id and everything hanging off it
// (memberships, sessions, API tokens), enforcing the rails.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		var enabled bool
		if err := tx.QueryRowContext(ctx,
			`SELECT enabled FROM users WHERE id = ?;`, id,
		).Scan(&enabled); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("load user: %w", err)
		}

		if enabled {
			if err := rejectIfLastAdmin(ctx, tx, id); err != nil {
				return err
			}
		}

		// Explicit cascade.
		for _, statement := range []string{
			`DELETE FROM user_groups WHERE user_id = ?;`,
			`DELETE FROM sessions WHERE user_id = ?;`,
			`DELETE FROM api_tokens WHERE user_id = ?;`,
			`DELETE FROM users WHERE id = ?;`,
		} {
			if _, err := tx.ExecContext(ctx, statement, id); err != nil {
				return fmt.Errorf("delete user: %w", err)
			}
		}

		return nil
	})
}

// LocalCredentials returns the credential material for username,
// or (nil, nil) when no such user exists.
func (s *Store) LocalCredentials(
	ctx context.Context,
	username string,
) (*auth.Credentials, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, email, password_hash, enabled
		FROM users
		WHERE username = ?;`,
		username,
	)

	creds := auth.Credentials{}
	if err := row.Scan(
		&creds.UserID, &creds.Username, &creds.DisplayName, &creds.Email,
		&creds.PasswordHash, &creds.Enabled,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil //nolint:nilnil
		}
		return nil, fmt.Errorf("load credentials: %w", err)
	}

	return &creds, nil
}

// UpdatePasswordHash replaces userID's stored password hash.
func (s *Store) UpdatePasswordHash(ctx context.Context, userID, hash string) error {
	if hash == "" {
		return ErrNoPassword
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?;`,
		hash, timeNow().Format(timeFormat), userID,
	); err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	return nil
}

// CreateFirstAdmin creates the first administrator account (in the admin group)
// from the first-run setup flow. It only succeeds while no users exist
// ([ErrSetupComplete] once any do).
func (s *Store) CreateFirstAdmin(
	ctx context.Context,
	username, displayName, password string,
) (*auth.User, error) {
	if count, err := s.CountUsers(ctx); err != nil {
		return nil, err
	} else if count > 0 {
		return nil, ErrSetupComplete
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	id := newID()
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		now := timeNow().Format(timeFormat)
		// First user only.
		res, err := tx.ExecContext(ctx, `
			INSERT INTO users (
				id,
				username,
				display_name,
				email,
				password_hash,
				enabled,
				created_at,
				updated_at
			)
			SELECT ?, ?, ?, '', ?, 1, ?, ?
			WHERE NOT EXISTS (
				SELECT 1
				FROM users
			);`,
			id, username, displayName, hash, now, now,
		)
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
		if n, err := res.RowsAffected(); err != nil {
			return fmt.Errorf("insert user: %w", err)
		} else if n == 0 {
			return ErrSetupComplete
		}

		return setUserGroups(ctx, tx, id, []string{GroupAdmin})
	}); err != nil {
		return nil, err
	}

	return s.UserByID(ctx, id)
}

// ResetUserPassword sets a new password for username and revokes their sessions
// ([ErrNotFound] for unknown usernames).
func (s *Store) ResetUserPassword(ctx context.Context, username, password string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.inTx(ctx, func(tx *sql.Tx) error {
		var id string
		if err := tx.QueryRowContext(ctx,
			`SELECT id FROM users WHERE username = ?;`,
			username,
		).Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf(
					"%w: user %q",
					ErrNotFound, username,
				)
			}
			return fmt.Errorf("load user: %w", err)
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?;`,
			hash, timeNow().Format(timeFormat), id,
		); err != nil {
			return fmt.Errorf("reset password: %w", err)
		}

		// Logout everywhere.
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM sessions WHERE user_id = ?;`, id,
		); err != nil {
			return fmt.Errorf("revoke sessions: %w", err)
		}

		return nil
	})
}

// userGroupNames returns the names of the groups userID belongs to, sorted.
func (s *Store) userGroupNames(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT g.name
		FROM user_groups ug
		JOIN groups g ON g.id = ug.group_id
		WHERE ug.user_id = ?
		ORDER BY g.name;`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list user's groups: %w", err)
	}
	defer rows.Close()

	return scanColumn[string](rows, "user's groups")
}

// setUserGroups replaces userID's memberships with the named groups.
func setUserGroups(ctx context.Context, tx *sql.Tx, userID string, groups []string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM user_groups WHERE user_id = ?;`, userID,
	); err != nil {
		return fmt.Errorf("clear user groups: %w", err)
	}

	for _, groupName := range groups {
		groupID, err := groupIDByName(ctx, tx, groupName)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w: %q", ErrUnknownGroup, groupName)
			}
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO user_groups (user_id, group_id) VALUES (?, ?);`,
			userID, groupID,
		); err != nil {
			return fmt.Errorf("add user to group %q: %w", groupName, err)
		}
	}

	return nil
}

// enabledAdminStatus returns the number of enabled admin members and whether
// userID is one of them.
func enabledAdminStatus(
	ctx context.Context,
	tx *sql.Tx,
	userID string,
) (total int, isMember bool, err error) {
	var member int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(u.id = ?), 0)
		FROM user_groups ug
		JOIN users u  ON u.id = ug.user_id
		JOIN groups g ON g.id = ug.group_id
		WHERE g.name = ? AND g.system = 1 AND u.enabled = 1;`,
		userID, GroupAdmin,
	).Scan(&total, &member); err != nil {
		return 0, false, fmt.Errorf("check admin membership: %w", err)
	}

	return total, member > 0, nil
}

// rejectIfLastAdmin returns [ErrLastAdmin] when userID is the last enabled admin.
func rejectIfLastAdmin(ctx context.Context, tx *sql.Tx, userID string) error {
	total, isMember, err := enabledAdminStatus(ctx, tx, userID)
	if err != nil {
		return err
	}
	if isMember && total == 1 {
		return ErrLastAdmin
	}

	return nil
}

// scanUser scans a users row (without groups).
func scanUser(row scanner) (*auth.User, error) {
	var user auth.User
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Email,
		&user.Enabled,
		timeText{&user.CreatedAt},
		timeText{&user.UpdatedAt},
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err //nolint:wrapcheck
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &user, nil
}
