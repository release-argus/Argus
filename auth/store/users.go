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

// CreateUser creates a user with the given group memberships.
// passwordHash may be empty for external-only users.
func (s *Store) CreateUser(
	ctx context.Context,
	username, displayName, email, passwordHash string,
	groups []string,
) (*auth.User, error) {
	id := newID()

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		// Reject duplicate usernames (case-insensitive).
		var existing int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM users WHERE username = ?;`,
			username,
		).Scan(&existing); err != nil {
			return fmt.Errorf("check username: %w", err)
		}
		if existing > 0 {
			return ErrUsernameTaken
		}

		now := timeNow().Format(timeFormat)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, username, display_name, email, password_hash, enabled, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 1, ?, ?);`,
			id, username, displayName, email,
			sqlNullString(passwordHash), now, now,
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
func (s *Store) UpdateUser(ctx context.Context, id string, patch UserPatch) (*auth.User, error) {
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
		sets := []string{"updated_at = ?"}
		args := []any{timeNow().Format(timeFormat)}
		if patch.DisplayName != nil {
			sets = append(sets, "display_name = ?")
			args = append(args, *patch.DisplayName)
		}
		if patch.Email != nil {
			sets = append(sets, "email = ?")
			args = append(args, *patch.Email)
		}
		if patch.Enabled != nil {
			sets = append(sets, "enabled = ?")
			args = append(args, *patch.Enabled)
		}
		if patch.PasswordHash != nil {
			sets = append(sets, "password_hash = ?")
			args = append(args, sqlNullString(*patch.PasswordHash))
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET `+strings.Join(sets, ", ")+` WHERE id = ?;`,
			append(args, id)...,
		); err != nil {
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
// (memberships, sessions, external identities, API tokens),
// enforcing the rails.
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
func (s *Store) LocalCredentials(ctx context.Context, username string) (*auth.Credentials, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, email, COALESCE(password_hash, ''), enabled
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
func (s *Store) CreateFirstAdmin(ctx context.Context, username, displayName, password string) (*auth.User, error) {
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
		// First user only.
		var count int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM users;`,
		).Scan(&count); err != nil {
			return fmt.Errorf("count users: %w", err)
		}
		if count > 0 {
			return ErrSetupComplete
		}

		now := timeNow().Format(timeFormat)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, username, display_name, email, password_hash, enabled, created_at, updated_at)
			VALUES (?, ?, ?, '', ?, 1, ?, ?);`,
			id, username, displayName, hash, now, now,
		); err != nil {
			return fmt.Errorf("insert user: %w", err)
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

// isEnabledAdmin reports whether userID is an enabled member of admin.
func isEnabledAdmin(ctx context.Context, tx *sql.Tx, userID string) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_groups ug
		JOIN users u  ON u.id = ug.user_id
		JOIN groups g ON g.id = ug.group_id
		WHERE g.name = ? AND g.system = 1
			AND u.id = ?   AND u.enabled = 1;`,
		GroupAdmin, userID,
	).Scan(&count); err != nil {
		return false, fmt.Errorf("check admin membership: %w", err)
	}

	return count > 0, nil
}

// otherEnabledAdmins counts enabled admin members besides userID.
func otherEnabledAdmins(ctx context.Context, tx *sql.Tx, userID string) (int, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_groups ug
		JOIN users u  ON u.id = ug.user_id
		JOIN groups g ON g.id = ug.group_id
		WHERE g.name = ? AND g.system = 1
			AND u.id != ?  AND u.enabled = 1;`,
		GroupAdmin, userID,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count other admins: %w", err)
	}

	return count, nil
}

// rejectIfLastAdmin returns [ErrLastAdmin] when userID is the last enabled admin.
func rejectIfLastAdmin(ctx context.Context, tx *sql.Tx, userID string) error {
	isAdmin, err := isEnabledAdmin(ctx, tx, userID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return nil
	}

	others, err := otherEnabledAdmins(ctx, tx, userID)
	if err != nil {
		return err
	}
	if others == 0 {
		return ErrLastAdmin
	}

	return nil
}

// scanner abstracts [sql.Row]/[sql.Rows] for [scanUser].
type scanner interface {
	Scan(dest ...any) error
}

// scanUser scans a users row (without groups).
func scanUser(row scanner) (*auth.User, error) {
	var (
		user      auth.User
		createdAt string
		updatedAt string
	)
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Email,
		&user.Enabled,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err //nolint:wrapcheck
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	var err error
	if user.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if user.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}

	return &user, nil
}

// sqlNullString maps "" to NULL.
func sqlNullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
