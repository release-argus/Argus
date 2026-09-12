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
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestStore_Preferences__roundTrip(t *testing.T) {
	// GIVEN: a Store with a user who has saved no preferences.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")

	prefix := fmt.Sprintf("%s\npreferences round-trip", packageName)

	// WHEN: their preferences are read.
	got, err := store.PreferencesForUser(t.Context(), user.ID)

	// THEN: they have none.
	if err != nil || got != nil {
		t.Fatalf(
			"%s expected no preferences\ngot:  %+v, err=%v\nwant: nil, nil",
			prefix, got, err,
		)
	}

	// WHEN: preferences are saved, out of catalogue order and with a duplicate.
	want := DashboardPreferences{
		Hide:       []string{HideInactive, HideUpToDate, HideInactive},
		View:       ViewTable,
		Timestamps: []string{TimestampQueried, TimestampDeployed},
	}
	stored, err := store.SetPreferences(t.Context(), user.ID, want)
	if err != nil {
		t.Fatalf(
			"%s SetPreferences failed: %v",
			prefix, err,
		)
	}

	// AND: it reports what it stored, so a caller need not re-read.
	if testErr := test.AssertSlicesEqual(
		t, stored.Hide, []string{HideUpToDate, HideInactive}, prefix, "stored hide",
	); testErr != nil {
		t.Error(testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, stored.Timestamps, []string{TimestampDeployed, TimestampQueried}, prefix, "stored timestamps",
	); testErr != nil {
		t.Error(testErr)
	}

	got, err = store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf(
			"%s PreferencesForUser failed: %v",
			prefix, err,
		)
	}

	// THEN: they come back deduplicated, in catalogue order.
	if testErr := test.AssertSlicesEqual(
		t, got.Hide, []string{HideUpToDate, HideInactive}, prefix, "hide",
	); testErr != nil {
		t.Error(testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, got.Timestamps, []string{TimestampDeployed, TimestampQueried}, prefix, "timestamps",
	); testErr != nil {
		t.Error(testErr)
	}
	if got.View != ViewTable {
		t.Errorf(
			"%s view mismatch\ngot:  %q\nwant: %q",
			prefix, got.View, ViewTable,
		)
	}

	// AND: saving again replaces the row rather than adding one.
	if _, err := store.SetPreferences(t.Context(), user.ID, DashboardPreferences{
		Hide:       []string{},
		View:       ViewGrid,
		Timestamps: []string{TimestampFound},
	}); err != nil {
		t.Fatalf(
			"%s second SetPreferences failed: %v",
			prefix, err,
		)
	}
	if got := countRows(t, store, "user_preferences"); got != 1 {
		t.Errorf(
			"%s row count mismatch\ngot:  %d\nwant: 1",
			prefix, got,
		)
	}
	got, err = store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf(
			"%s PreferencesForUser failed: %v",
			prefix, err,
		)
	}

	// AND: an explicitly empty list stays empty rather than reverting to nil.
	if testErr := test.AssertSlicesEqual(
		t, got.Hide, []string{}, prefix, "hide",
	); testErr != nil {
		t.Error(testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, got.Timestamps, []string{TimestampFound}, prefix, "timestamps",
	); testErr != nil {
		t.Error(testErr)
	}

	// AND: DeletePreferences returns them to inheriting everything.
	if err := store.DeletePreferences(t.Context(), user.ID); err != nil {
		t.Fatalf(
			"%s DeletePreferences failed: %v",
			prefix, err,
		)
	}
	got, err = store.PreferencesForUser(t.Context(), user.ID)
	if err != nil || got != nil {
		t.Errorf(
			"%s preferences should be gone\ngot:  %+v, err=%v\nwant: nil, nil",
			prefix, got, err,
		)
	}

	// AND: deleting again is not an error.
	if err := store.DeletePreferences(t.Context(), user.ID); err != nil {
		t.Errorf(
			"%s second DeletePreferences failed: %v",
			prefix, err,
		)
	}
}

func TestStore_Preferences__nilFieldsInherit(t *testing.T) {
	// GIVEN: preferences with some fields set.
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			user := mustCreateUser(t, store, "argus", "")

			prefix := fmt.Sprintf("%s\npreferences with nil fields", packageName)

			// WHEN: they are saved and read back.
			if _, err := store.SetPreferences(t.Context(), user.ID, tc.saved); err != nil {
				t.Fatalf(
					"%s SetPreferences failed: %v",
					prefix, err,
				)
			}
			got, err := store.PreferencesForUser(t.Context(), user.ID)
			if err != nil {
				t.Fatalf(
					"%s PreferencesForUser failed: %v",
					prefix, err,
				)
			}

			// THEN: the fields left unset come back nil, so they keep inheriting.
			if testErr := test.AssertSlicesEqual(
				t, got.Hide, tc.wantHide, prefix, "hide",
			); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertSlicesEqual(
				t, got.Timestamps, tc.wantTimestamps, prefix, "timestamps",
			); testErr != nil {
				t.Error(testErr)
			}
			if got.View != tc.wantView {
				t.Errorf(
					"%s view mismatch\ngot:  %q\nwant: %q",
					prefix, got.View, tc.wantView,
				)
			}
		})
	}
}

func TestStore_Preferences__deletedWithTheirUser(t *testing.T) {
	// GIVEN: a Store with two users, both holding preferences.
	store := testStore(t)
	target := mustCreateUser(t, store, "target", "")
	other := mustCreateUser(t, store, "other", "")

	prefix := fmt.Sprintf("%s\npreferences cascade", packageName)

	for _, user := range []string{target.ID, other.ID} {
		if _, err := store.SetPreferences(t.Context(), user, DashboardPreferences{
			View: ViewTable,
		}); err != nil {
			t.Fatalf(
				"%s SetPreferences failed: %v",
				prefix, err,
			)
		}
	}

	// WHEN: one user is deleted.
	if err := store.DeleteUser(t.Context(), target.ID); err != nil {
		t.Fatalf(
			"%s DeleteUser failed: %v",
			prefix, err,
		)
	}

	// THEN: their preferences go with them, and nobody else's do.
	if got, err := store.PreferencesForUser(t.Context(), target.ID); err != nil || got != nil {
		t.Errorf(
			"%s preferences outlived their user\ngot:  %+v, err=%v\nwant: nil, nil",
			prefix, got, err,
		)
	}
	if got, err := store.PreferencesForUser(t.Context(), other.ID); err != nil || got == nil {
		t.Errorf(
			"%s the other user's preferences were taken too\ngot:  %+v, err=%v",
			prefix, got, err,
		)
	}
}

func TestDashboardPreferences_CheckValues(t *testing.T) {
	// GIVEN: preferences to validate.
	tests := []struct {
		name     string
		prefs    DashboardPreferences
		errRegex string
	}{
		{
			name:  "valid/everything unset",
			prefs: DashboardPreferences{},
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
		},
		{
			name: "valid/explicitly empty lists",
			prefs: DashboardPreferences{
				Hide:       []string{},
				Timestamps: []string{},
			},
		},
		{
			name:     "invalid/unknown hide name",
			prefs:    DashboardPreferences{Hide: []string{HideSkipped, "upToDate"}},
			errRegex: `^hide: "upToDate" <invalid> \(.*$`,
		},
		{
			name:     "invalid/unknown view",
			prefs:    DashboardPreferences{View: "list"},
			errRegex: `^view: "list" <invalid> \(.*$`,
		},
		{
			name:     "invalid/unknown timestamp name",
			prefs:    DashboardPreferences{Timestamps: []string{"last_queried"}},
			errRegex: `^timestamps: "last_queried" <invalid> \(.*$`,
		},
		{
			name: "invalid/every field reported",
			prefs: DashboardPreferences{
				Hide:       []string{"nope"},
				View:       "nope",
				Timestamps: []string{"nope"},
			},
			errRegex: test.TrimYAML(`
				^hide: "nope" <invalid> \(supported values = \['up_to_date', 'updatable', 'skipped', 'inactive'\]\)
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
			errRegex: `^hide: "nope1", "nope2", "nope3", "nope4", "nope5" <invalid> \(.*$`,
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
				hide: "nope1", "nope2", "nope3", "nope4", "nope5", and 2 more <invalid> \(.*
				timestamps: "nope1", "nope2", "nope3", "nope4", "nope5", and 1 more <invalid> \(.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prefix := fmt.Sprintf("%s\nCheckValues()", packageName)

			// WHEN: they are validated.
			err := tc.prefs.CheckValues()

			// THEN: it errors when expected.
			if err == nil && tc.errRegex == "" {
				return
			}

			// AND: it errors as expected.
			if !errors.Is(err, ErrInvalidPreference) {
				t.Fatalf(
					"%s error doesn't wrap the expected error\ngot:  %v\nwant: %v",
					prefix, err, ErrInvalidPreference,
				)
			}
			e := errfmt.FormatError(err)
			e = strings.TrimPrefix(e, ErrInvalidPreference.Error()+"\n")
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\nwant: %v\ngot: %v",
					prefix, tc.errRegex, e,
				)
			}
		})
	}
}

func TestStore_SetPreferences__rejectsUnknownValues(t *testing.T) {
	// GIVEN: a Store with a user.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")

	// WHEN: a name outside the catalogue is saved.
	invalidHide := "upToDate"
	prefs := DashboardPreferences{
		Hide: []string{HideSkipped, invalidHide},
	}
	stored, err := store.SetPreferences(t.Context(), user.ID, prefs)

	prefix := fmt.Sprintf(
		"%s\nSetPreferences(prefs=%+v) unknown value",
		packageName, prefs,
	)

	// THEN: it is refused with an error.
	if err == nil {
		t.Fatalf(
			"%s expected an error\ngot: %+v",
			prefix, stored,
		)
	}
	if !strings.Contains(err.Error(), invalidHide) {
		t.Errorf(
			"%s error should name the offending value\ngot: %v",
			prefix, err,
		)
	}

	// AND: nothing was written.
	if got := countRows(t, store, "user_preferences"); got != 0 {
		t.Errorf(
			"%s row written despite the error\ngot:  %d\nwant: 0",
			prefix, got,
		)
	}
}

func TestStore_SetPreferences__zeroSetDiscards(t *testing.T) {
	// GIVEN: a Store with a user who has saved preferences.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")

	prefix := fmt.Sprintf("%s\nSetPreferences() zero set", packageName)

	prefs := DashboardPreferences{
		View: ViewTable,
	}
	if _, err := store.SetPreferences(t.Context(), user.ID, prefs); err != nil {
		t.Fatalf(
			"%s setup SetPreferences(prefs=%+v) failed: %v",
			prefix, prefs, err,
		)
	}

	// WHEN: a set with nothing in it is saved.
	prefs = DashboardPreferences{}
	stored, err := store.SetPreferences(t.Context(), user.ID, prefs)
	if err != nil {
		t.Fatalf(
			"%s SetPreferences(prefs=%+v) failed: %v",
			prefix, prefs, err,
		)
	}

	// THEN: it reports nothing saved.
	if stored != nil {
		t.Errorf(
			"%s expected no preferences\ngot: %+v",
			prefix, *stored,
		)
	}

	// AND: the row is gone, so a read agrees.
	if got := countRows(t, store, "user_preferences"); got != 0 {
		t.Errorf(
			"%s row survived a zero set\ngot:  %d\nwant: 0",
			prefix, got,
		)
	}
	if got, err := store.PreferencesForUser(t.Context(), user.ID); err != nil || got != nil {
		t.Errorf(
			"%s read should find none\ngot:  %+v, err=%v",
			prefix, got, err,
		)
	}
}

func TestStore_PreferencesForUser__canonicalisesOnRead(t *testing.T) {
	// GIVEN: a Store holding a row written from an earlier build's catalogues.
	store := testStore(t)
	user := mustCreateUser(t, store, "argus", "")
	mustExec(t, store, fmt.Sprintf(`
		INSERT INTO user_preferences
			(user_id, hide, view, timestamps, created_at, updated_at)
		VALUES
			('%s', '%s,retired_hide', 'compact', '%s,retired_timestamp', '%s', '%s');`,
		user.ID, HideSkipped, TimestampFound, timeNow().Format(timeFormat), timeNow().Format(timeFormat),
	))

	prefix := fmt.Sprintf("%s\nPreferencesForUser() canonicalise", packageName)

	// WHEN: they are read.
	got, err := store.PreferencesForUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf(
			"%s PreferencesForUser failed: %v",
			prefix, err,
		)
	}

	// THEN: values this build no longer knows are dropped.
	if testErr := test.AssertSlicesEqual(
		t, got.Hide, []string{HideSkipped}, prefix, "hide",
	); testErr != nil {
		t.Error(testErr)
	}
	if testErr := test.AssertSlicesEqual(
		t, got.Timestamps, []string{TimestampFound}, prefix, "timestamps",
	); testErr != nil {
		t.Error(testErr)
	}
	if got.View != "" {
		t.Errorf(
			"%s unknown view should be blanked\ngot: %q",
			prefix, got.View,
		)
	}
}
