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

package store

import (
	"errors"
	"fmt"
	"testing"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestStore_Preferences__roundTrip(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")

	got, err := store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("before set: unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf(
			"before set: Store.PreferencesForUser() mismatch\n"+
				"got:  View=%q, Hide=%q, Timestamps=%q\nwant: nil",
			got.View, got.Hide, got.Timestamps,
		)
	}

	// Out of catalogue order, with a duplicate.
	prefs := DashboardPreferences{
		Hide:       []string{HideInactive, HideUpToDate, HideInactive},
		View:       ViewTable,
		Timestamps: []string{TimestampQueried, TimestampDeployed},
	}
	stored, err := store.SetPreferences(t.Context(), user.ID, prefs)
	if err != nil {
		t.Fatalf("first: Store.SetPreferences() failed: %v", err)
	}

	if testErr := test.AssertSlicesEqual(
		t,
		stored.Hide,
		[]string{HideUpToDate, HideInactive},
		"DashboardPreferences.Hide",
	); testErr != nil {
		t.Errorf("first: stored: %v", testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t,
		stored.Timestamps,
		[]string{TimestampDeployed, TimestampQueried},
		"DashboardPreferences.Timestamps",
	); testErr != nil {
		t.Errorf("first: stored: %v", testErr)
	}

	got, err = store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("first: Store.PreferencesForUser() failed: %v", err)
	}
	if got == nil {
		t.Fatal("first: Store.PreferencesForUser() mismatch\ngot:  nil\nwant: non-nil")
	}

	if testErr := test.AssertSlicesEqual(
		t, got.Hide, []string{HideUpToDate, HideInactive}, "DashboardPreferences.Hide",
	); testErr != nil {
		t.Errorf("first: %v", testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, got.Timestamps, []string{TimestampDeployed, TimestampQueried}, "DashboardPreferences.Timestamps",
	); testErr != nil {
		t.Errorf("first: %v", testErr)
	}
	if got.View != ViewTable {
		t.Errorf(
			"first: DashboardPreferences.View mismatch\ngot:  %q\nwant: %q",
			got.View, ViewTable,
		)
	}

	_, err = store.SetPreferences(t.Context(), user.ID, DashboardPreferences{
		Hide:       []string{},
		View:       ViewGrid,
		Timestamps: []string{TimestampFound},
	})
	if err != nil {
		t.Fatalf("second: Store.SetPreferences() failed: %v", err)
	}
	if rows := test.CountRows(t, store.db, "user_preferences"); rows != 1 {
		t.Errorf("second: user_preferences row count mismatch\ngot:  %d\nwant: 1", rows)
	}
	got, err = store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("second: Store.PreferencesForUser() failed: %v", err)
	}
	if got == nil {
		t.Fatal("second: Store.PreferencesForUser() mismatch\ngot:  nil\nwant: non-nil")
	}

	// An explicitly empty list stays empty, not nil.
	if testErr := test.AssertSlicesEqual(
		t, got.Hide, []string{}, "DashboardPreferences.Hide",
	); testErr != nil {
		t.Errorf("second: %v", testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, got.Timestamps, []string{TimestampFound}, "DashboardPreferences.Timestamps",
	); testErr != nil {
		t.Errorf("second: %v", testErr)
	}

	err = store.DeletePreferences(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("delete: Store.DeletePreferences() failed: %v", err)
	}
	got, err = store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("after delete: Store.PreferencesForUser() failed: %v", err)
	}
	if got != nil {
		t.Errorf(
			"after delete: Store.PreferencesForUser() mismatch\n"+
				"got:  View=%q, Hide=%q, Timestamps=%q\nwant: nil",
			got.View, got.Hide, got.Timestamps,
		)
	}

	if err := store.DeletePreferences(t.Context(), user.ID); err != nil {
		t.Errorf("second delete: Store.DeletePreferences() failed: %v", err)
	}
}

func TestStore_Preferences__nilFieldsInherit(t *testing.T) {
	tests := []struct {
		name           string
		saved          DashboardPreferences
		wantHide       []string
		wantView       string
		wantTimestamps []string
	}{
		{
			name:     "layout alone",
			saved:    DashboardPreferences{View: ViewTable},
			wantView: ViewTable,
		},
		{
			name:     "filters alone",
			saved:    DashboardPreferences{Hide: []string{HideSkipped}},
			wantHide: []string{HideSkipped},
		},
		{
			name:           "timestamps alone",
			saved:          DashboardPreferences{Timestamps: []string{TimestampFound}},
			wantTimestamps: []string{TimestampFound},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			user := mustCreateUser(t, store, "argus", "")

			_, err := store.SetPreferences(t.Context(), user.ID, tc.saved)
			if err != nil {
				t.Fatalf("Store.SetPreferences() failed: %v", err)
			}
			got, err := store.PreferencesForUser(t.Context(), user.ID)
			if err != nil {
				t.Fatalf("Store.PreferencesForUser() failed: %v", err)
			}
			if got == nil {
				t.Fatal("Store.PreferencesForUser() mismatch\ngot:  nil\nwant: non-nil")
			}

			if testErr := test.AssertSlicesEqual(
				t, got.Hide, tc.wantHide, "DashboardPreferences.Hide",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertSlicesEqual(
				t, got.Timestamps, tc.wantTimestamps, "DashboardPreferences.Timestamps",
			); testErr != nil {
				t.Error(testErr)
			}
			if got.View != tc.wantView {
				t.Errorf(
					"DashboardPreferences.View mismatch\ngot:  %q\nwant: %q",
					got.View, tc.wantView,
				)
			}
		})
	}
}

func TestStore_Preferences__deletedWithTheirUser(t *testing.T) {
	store := testStore(t)
	target := mustCreateUser(t, store, "target", "")
	other := mustCreateUser(t, store, "other", "")

	for _, user := range []*auth.User{target, other} {
		prefs := DashboardPreferences{View: ViewTable}
		if _, err := store.SetPreferences(t.Context(), user.ID, prefs); err != nil {
			t.Fatalf("Store.SetPreferences(%q) failed: %v", user.Username, err)
		}
	}

	err := store.DeleteUser(t.Context(), target.ID)
	if err != nil {
		t.Fatalf("Store.DeleteUser() failed: %v", err)
	}

	got, err := store.PreferencesForUser(t.Context(), target.ID)
	if err != nil {
		t.Fatalf("target: Store.PreferencesForUser() failed: %v", err)
	}
	if got != nil {
		t.Errorf(
			"target: Store.PreferencesForUser() mismatch\n"+
				"got:  View=%q, Hide=%q, Timestamps=%q\nwant: nil",
			got.View, got.Hide, got.Timestamps,
		)
	}
	got, err = store.PreferencesForUser(t.Context(), other.ID)
	if err != nil {
		t.Fatalf("other: Store.PreferencesForUser() failed: %v", err)
	}
	if got == nil {
		t.Error("other: Store.PreferencesForUser() mismatch\ngot:  nil\nwant: non-nil")
	}
}

func TestDashboardPreferences_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		prefs    DashboardPreferences
		errRegex string
	}{
		{
			name:     "valid/everything unset",
			prefs:    DashboardPreferences{},
			errRegex: `^$`,
		},
		{
			name: "valid/every field set",
			prefs: DashboardPreferences{
				Hide: []string{
					HideUpToDate, HideUpdatable, HideSkipped, HideInactive,
				},
				View: ViewGrid,
				Timestamps: []string{
					TimestampDeployed, TimestampFound, TimestampQueried,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "valid/explicitly empty lists",
			prefs: DashboardPreferences{
				Hide:       []string{},
				Timestamps: []string{},
			},
			errRegex: `^$`,
		},
		{
			name:  "invalid/unknown hide name",
			prefs: DashboardPreferences{Hide: []string{HideSkipped, "upToDate"}},
			errRegex: test.TrimYAML(`
				^invalid preference value:
				  hide: "upToDate" <invalid> \(.*$`),
		},
		{
			name:  "invalid/unknown view",
			prefs: DashboardPreferences{View: "list"},
			errRegex: test.TrimYAML(`
				^invalid preference value:
				  view: "list" <invalid> \(.*$`),
		},
		{
			name:  "invalid/unknown timestamp name",
			prefs: DashboardPreferences{Timestamps: []string{"last_queried"}},
			errRegex: test.TrimYAML(`
				^invalid preference value:
				  timestamps: "last_queried" <invalid> \(.*$`),
		},
		{
			name: "invalid/every field reported",
			prefs: DashboardPreferences{
				Hide:       []string{"nope"},
				View:       "nope",
				Timestamps: []string{"nope"},
			},
			errRegex: test.TrimYAML(`
				^invalid preference value:
					hide: "nope" <invalid> \(supported values = \['up_to_date', 'updatable', 'skipped', 'inactive'\]\)
					timestamps: "nope" <invalid> \(supported values = \['deployed', 'found', 'queried'\]\)
					view: "nope" <invalid> \(supported values = \['grid', 'table'\]\)`,
			),
		},
		{
			name: "invalid/at the cap, every value is named",
			prefs: DashboardPreferences{
				Hide: []string{
					"nope1", "nope2", HideSkipped, "nope3", "nope4", "nope5",
				},
			},
			errRegex: test.TrimYAML(`
				^invalid preference value:
				  hide: "nope1", "nope2", "nope3", "nope4", "nope5" <invalid> \(.*$`),
		},
		{
			name: "invalid/over the cap, the rest are counted",
			prefs: DashboardPreferences{
				Hide: []string{
					"nope1", "nope2", "nope3", "nope4", "nope5", "nope6", "nope7",
				},
				Timestamps: []string{
					"nope1", "nope2", "nope3", "nope4", "nope5", "nope6",
				},
			},
			errRegex: test.TrimYAML(`
				^invalid preference value:
					hide: "nope1", "nope2", "nope3", "nope4", "nope5", and 2 more <invalid> \(.*
					timestamps: "nope1", "nope2", "nope3", "nope4", "nope5", and 1 more <invalid> \(.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.prefs.CheckValues()
			if err != nil && !errors.Is(err, ErrInvalidPreference) {
				t.Fatalf("error type mismatch\ngot:  %v\nwant: %v", err, ErrInvalidPreference)
			}
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestStore_PreferencesForUser__canonicalisesOnRead(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	test.ExecSQL(t, store.db, fmt.Sprintf(`
		INSERT INTO user_preferences
			(user_id, hide, view, timestamps, created_at, updated_at)
		VALUES
			('%s', '%s,retired_hide', 'compact', '%s,retired_timestamp', '%s', '%s');`,
		user.ID, HideSkipped, TimestampFound, timeNow().Format(timeFormat), timeNow().Format(timeFormat),
	))

	got, err := store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("Store.PreferencesForUser() failed: %v", err)
	}
	if got == nil {
		t.Fatal("Store.PreferencesForUser() mismatch\ngot:  nil\nwant: non-nil")
	}

	if testErr := test.AssertSlicesEqual(
		t, got.Hide, []string{HideSkipped}, "DashboardPreferences.Hide",
	); testErr != nil {
		t.Error(testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, got.Timestamps, []string{TimestampFound}, "DashboardPreferences.Timestamps",
	); testErr != nil {
		t.Error(testErr)
	}
	if got.View != "" {
		t.Errorf("DashboardPreferences.View mismatch\ngot:  %q\nwant: %q", got.View, "")
	}
}

func TestStore_SetPreferences__rejectsUnknownValues(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")

	invalidHide := "upToDate"
	prefs := DashboardPreferences{
		Hide: []string{HideSkipped, invalidHide},
	}
	_, err := store.SetPreferences(t.Context(), user.ID, prefs)
	e := errfmt.FormatError(err)
	errRegex := fmt.Sprintf(`^invalid preference value:\n  hide: %q <invalid> \(.*$`, invalidHide)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}

	if got := test.CountRows(t, store.db, "user_preferences"); got != 0 {
		t.Errorf("user_preferences row count mismatch\ngot:  %d\nwant: 0", got)
	}
}

func TestStore_SetPreferences__zeroSetDiscards(t *testing.T) {
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")

	_, err := store.SetPreferences(t.Context(), user.ID, DashboardPreferences{View: ViewTable})
	if err != nil {
		t.Fatalf("seed: Store.SetPreferences() failed: %v", err)
	}

	stored, err := store.SetPreferences(t.Context(), user.ID, DashboardPreferences{})
	if err != nil {
		t.Fatalf("empty set: Store.SetPreferences() failed: %v", err)
	}

	if stored != nil {
		t.Errorf(
			"Store.SetPreferences() mismatch\ngot:  View=%q, Hide=%q, Timestamps=%q\nwant: nil",
			stored.View, stored.Hide, stored.Timestamps,
		)
	}

	if got := test.CountRows(t, store.db, "user_preferences"); got != 0 {
		t.Errorf("user_preferences row count mismatch\ngot:  %d\nwant: 0", got)
	}
	got, err := store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("Store.PreferencesForUser() failed: %v", err)
	}
	if got != nil {
		t.Errorf(
			"Store.PreferencesForUser() mismatch\ngot:  View=%q, Hide=%q, Timestamps=%q\nwant: nil",
			got.View, got.Hide, got.Timestamps,
		)
	}
}
