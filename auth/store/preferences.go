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

	"github.com/release-argus/Argus/util/polymorphic"
)

// The 'hide' filters a user can default to.
const (
	HideUpToDate  = "up_to_date"
	HideUpdatable = "updatable"
	HideSkipped   = "skipped"
	HideInactive  = "inactive"
)

// The dashboard layouts a user can default to.
const (
	ViewGrid  = "grid"
	ViewTable = "table"
)

// The service card timestamps a user can default to showing.
const (
	TimestampDeployed = "deployed"
	TimestampFound    = "found"
	TimestampQueried  = "queried"
)

// The accepted values of each [DashboardPreferences] field, in the order they
// are stored.
var (
	hideNames      = []string{HideUpToDate, HideUpdatable, HideSkipped, HideInactive}
	viewNames      = []string{ViewGrid, ViewTable}
	timestampNames = []string{TimestampDeployed, TimestampFound, TimestampQueried}
)

// DashboardPreferences are a user's saved dashboard preferences.
// A zero field inherits the built-in default; an empty (non-nil) slice is an
// explicit "none".
type DashboardPreferences struct {
	Hide       []string `json:"hide,omitzero"`
	View       string   `json:"view,omitzero"`
	Timestamps []string `json:"timestamps,omitzero"`
}

// CheckValues validates the fields of the receiver,
// wrapping [ErrInvalidPreference] so callers can answer 400.
func (p DashboardPreferences) CheckValues() error {
	var errs []error

	for _, field := range []struct {
		key    string
		values []string
		valid  []string
	}{
		{key: "hide", values: p.Hide, valid: hideNames},
		{key: "timestamps", values: p.Timestamps, valid: timestampNames},
	} {
		var invalid []string
		for _, value := range field.values {
			if !slices.Contains(field.valid, value) {
				invalid = append(invalid, value)
			}
		}
		if invalid == nil {
			continue
		}
		errs = append(errs, polymorphic.ErrInvalidValues{
			Key:     field.key,
			Values:  invalid,
			Allowed: field.valid,
		})
	}

	if p.View != "" && !slices.Contains(viewNames, p.View) {
		errs = append(errs, polymorphic.ErrInvalidType{
			Key:     "view",
			Value:   p.View,
			Allowed: viewNames,
		})
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf(
		"%w: %w",
		ErrInvalidPreference, errors.Join(errs...),
	)
}

// IsZero reports whether no preferences are set (overridden).
func (p DashboardPreferences) IsZero() bool {
	return p.Hide == nil && p.View == "" && p.Timestamps == nil
}

// PreferencesForUser returns userID's saved preferences,
// or (nil, nil) when they have none.
func (s *Store) PreferencesForUser(
	ctx context.Context,
	userID string,
) (*DashboardPreferences, error) {
	var hide, view, timestamps sql.NullString
	if err := s.db.QueryRowContext(ctx, `
		SELECT hide, view, timestamps
		FROM user_preferences
		WHERE user_id = ?;`,
		userID,
	).Scan(&hide, &view, &timestamps); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil //nolint:nilnil
		}
		return nil, fmt.Errorf("load preferences: %w", err)
	}

	return &DashboardPreferences{
		Hide:       canonical(decodeNames(hide), hideNames),
		View:       knownView(view.String),
		Timestamps: canonical(decodeNames(timestamps), timestampNames),
	}, nil
}

// knownView returns view when it is a layout this build knows, else empty.
func knownView(view string) string {
	if slices.Contains(viewNames, view) {
		return view
	}
	return ""
}

// SetPreferences replaces userID's saved preferences, returning them as stored.
// Zero fields are stored as inheriting the built-in default.
// Values outside the catalogues are an error.
func (s *Store) SetPreferences(
	ctx context.Context,
	userID string,
	prefs DashboardPreferences,
) (*DashboardPreferences, error) {
	if err := prefs.CheckValues(); err != nil {
		return nil, err
	}

	// Setting nothing is discarding.
	if prefs.IsZero() {
		return nil, s.DeletePreferences(ctx, userID)
	}

	hide := canonical(prefs.Hide, hideNames)
	timestamps := canonical(prefs.Timestamps, timestampNames)

	now := timeNow().Format(timeFormat)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO user_preferences (
			user_id,
			hide,
			view,
			timestamps,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			hide       = excluded.hide,
			view       = excluded.view,
			timestamps = excluded.timestamps,
			updated_at = excluded.updated_at;`,
		userID,
		encodeNames(hide),
		sql.NullString{String: prefs.View, Valid: prefs.View != ""},
		encodeNames(timestamps),
		now,
		now,
	); err != nil {
		return nil, fmt.Errorf("set preferences: %w", err)
	}

	return &DashboardPreferences{
		Hide:       hide,
		View:       prefs.View,
		Timestamps: timestamps,
	}, nil
}

// DeletePreferences drops userID's saved preferences, returning them to the
// built-in defaults.
func (s *Store) DeletePreferences(ctx context.Context, userID string) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM user_preferences WHERE user_id = ?;`, userID,
	); err != nil {
		return fmt.Errorf("delete preferences: %w", err)
	}
	return nil
}

// canonical filters names down to the catalogue, in catalogue order, for stable
// ordering and deduplication. A nil list stays nil.
func canonical(names []string, catalogue []string) []string {
	if names == nil {
		return nil
	}

	kept := make([]string, 0, len(catalogue))
	for _, name := range catalogue {
		if slices.Contains(names, name) {
			kept = append(kept, name)
		}
	}
	return kept
}

// encodeNames renders a canonical name list for storage:
// NULL when unset, else the names joined.
func encodeNames(names []string) sql.NullString {
	return sql.NullString{String: strings.Join(names, ","), Valid: names != nil}
}

// decodeNames reverses [encodeNames].
func decodeNames(column sql.NullString) []string {
	if !column.Valid {
		return nil
	}
	if column.String == "" {
		return []string{}
	}
	return strings.Split(column.String, ",")
}
