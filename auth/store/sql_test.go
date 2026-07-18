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
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestRowUpdate(t *testing.T) {
	// GIVEN: a row with two columns and a timestamp.
	type colValue struct {
		col string
		val any
	}
	tests := []struct {
		name  string
		sets  []colValue
		wantA string
		wantB string
	}{
		{
			name:  "no columns still bumps updated_at",
			sets:  nil,
			wantA: "A1",
			wantB: "B1",
		},
		{
			name:  "single column",
			sets:  []colValue{{"a", "A2"}},
			wantA: "A2",
			wantB: "B1",
		},
		{
			name:  "multiple columns",
			sets:  []colValue{{"a", "A2"}, {"b", "B2"}},
			wantA: "A2",
			wantB: "B2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := testStore(t)
			mustExec(t, store,
				`CREATE TABLE tmp (
					id TEXT PRIMARY KEY, a TEXT, b TEXT, updated_at DATETIME
				);`)
			mustExec(t, store,
				`INSERT INTO tmp (id, a, b, updated_at)
				VALUES ('x', 'A1', 'B1', 'stale');`)

			prefix := fmt.Sprintf("%s\nrowUpdate", packageName)

			// WHEN: the accumulated columns are written to the row.
			upd := newRowUpdate()
			for _, s := range tc.sets {
				upd.set(s.col, s.val)
			}
			if err := store.inTx(t.Context(), func(tx *sql.Tx) error {
				return upd.exec(t.Context(), tx, "tmp", "x")
			}); err != nil {
				t.Fatalf(
					"%s exec failed: %v",
					prefix, err,
				)
			}

			// THEN: the set columns take their new values, the rest are unchanged.
			var a, b, updatedAt string
			if err := store.db.QueryRow(
				`SELECT a, b, updated_at FROM tmp WHERE id = 'x';`,
			).Scan(&a, &b, &updatedAt); err != nil {
				t.Fatalf(
					"%s read-back failed: %v",
					prefix, err,
				)
			}
			if a != tc.wantA || b != tc.wantB {
				t.Errorf(
					"%s column mismatch\ngot:  a=%q b=%q\nwant: a=%q b=%q",
					prefix,
					a, b,
					tc.wantA, tc.wantB,
				)
			}

			// AND: updated_at is always bumped, even with no columns set.
			if updatedAt == "stale" {
				t.Errorf("%s updated_at should have been bumped", prefix)
			}
		})
	}
}

func TestTimeText_Scan(t *testing.T) {
	// GIVEN: a timestamp and the sources a driver might hand a scanner.
	when := time.Date(2026, 1, 2, 3, 4, 5, 600000000, time.UTC)
	tests := []struct {
		name    string
		src     any
		want    time.Time
		wantErr bool
	}{
		{
			name: "valid string",
			src:  when.Format(timeFormat),
			want: when,
		},
		{
			name: "time.Time value",
			src:  when,
			want: when,
		},
		{
			name:    "invalid string",
			src:     "not-a-time",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			src:     42,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf(
				"%s\ntimeText.Scan(%v)",
				packageName, tc.src,
			)

			// WHEN: the source is scanned.
			var dst time.Time
			err := (timeText{&dst}).Scan(tc.src)

			// THEN: an error surfaces for bad input.
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s expected an error, got nil", prefix)
				}
				return
			}
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// AND: clean values are returned.
			if !dst.Equal(tc.want) {
				t.Errorf(
					"%s value mismatch\ngot:  %v\nwant: %v",
					prefix, dst, tc.want,
				)
			}
		})
	}
}

func TestNullTimeText_Scan(t *testing.T) {
	// GIVEN: a timestamp and the nullable sources a driver might hand a scanner.
	when := time.Date(2026, 1, 2, 3, 4, 5, 600000000, time.UTC)
	tests := []struct {
		name    string
		src     any
		want    *time.Time
		wantErr bool
	}{
		{
			name: "nil leaves the target nil",
			src:  nil,
			want: nil,
		},
		{
			name: "valid string",
			src:  when.Format(timeFormat),
			want: &when,
		},
		{
			name:    "invalid string",
			src:     "not-a-time",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nnullTimeText.Scan(%v)", packageName, tc.src)

			// WHEN: the source is scanned.
			var dst *time.Time
			err := (nullTimeText{&dst}).Scan(tc.src)

			// THEN: bad input errors.
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s expected an error, got nil", prefix)
				}
				return
			}
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %v",
					prefix, err,
				)
			}

			// AND: NULL leaves nil.
			if tc.want == nil {
				if dst != nil {
					t.Errorf(
						"%s want nil, got %v",
						prefix, dst,
					)
				}
				return
			}

			// AND: clean values are returned.
			if dst == nil || !dst.Equal(*tc.want) {
				t.Errorf(
					"%s mismatch\ngot:  %v\nwant: %v",
					prefix, dst, tc.want,
				)
			}
		})
	}
}

func TestTimeFormat__textOrderIsChronological(t *testing.T) {
	// GIVEN: pairs of timestamps whose fractional parts differ in digit count,
	// where one is a prefix of the other. RFC3339Nano trims trailing zeros, so
	// its TEXT form sorts 'Z' above a digit and reverses these pairs.
	tests := []struct {
		name     string
		earlier  string
		later    string
		wantSwap bool
	}{
		{
			name:    "prefix/tenths before hundredths",
			earlier: "2026-01-01T00:00:00.1Z",
			later:   "2026-01-01T00:00:00.12Z",
		},
		{
			name:    "prefix/half second before half plus a nanosecond slice",
			earlier: "2026-01-01T00:00:00.5Z",
			later:   "2026-01-01T00:00:00.50000001Z",
		},
		{
			name:    "prefix/whole second before a fraction of it",
			earlier: "2026-01-01T00:00:00Z",
			later:   "2026-01-01T00:00:00.000000001Z",
		},
		{
			name:    "distinct/differing seconds",
			earlier: "2026-01-01T00:00:00.5Z",
			later:   "2026-01-01T00:00:01.05Z",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			earlier, err := parseTime(tc.earlier)
			if err != nil {
				t.Fatalf(
					"%s\nparseTime(%q): %v",
					packageName, tc.earlier, err,
				)
			}
			later, err := parseTime(tc.later)
			if err != nil {
				t.Fatalf(
					"%s\nparseTime(%q): %v",
					packageName, tc.later, err,
				)
			}

			prefix := fmt.Sprintf("%s\ntimeFormat TEXT ordering", packageName)

			// AND: the pair really is chronologically ordered.
			if !earlier.Before(later) {
				t.Fatalf(
					"%s\ntest case is not chronological\n%q is not before %q",
					prefix, tc.earlier, tc.later,
				)
			}

			// WHEN: both are written with timeFormat.
			gotEarlier := earlier.UTC().Format(timeFormat)
			gotLater := later.UTC().Format(timeFormat)

			// THEN: their TEXT order matches their chronological order, which
			// ORDER BY last_seen_at and the `expires_at < ?` sweeps depend on.
			if !(gotEarlier < gotLater) {
				t.Errorf(
					"%s\nstored order mismatch\ngot:  %q is not < %q\nwant: chronological order preserved",
					prefix, gotEarlier, gotLater,
				)
			}
		})
	}
}

func TestParseTime__acceptsTrimmedFractions(t *testing.T) {
	// GIVEN: timestamps as either layout may have written them.
	tests := []struct {
		name  string
		value string
	}{
		{name: "valid/no fraction", value: "2026-01-01T00:00:00Z"},
		{name: "valid/one digit", value: "2026-01-01T00:00:00.5Z"},
		{name: "valid/two digits", value: "2026-01-01T00:00:01.05Z"},
		{name: "valid/nine digits", value: "2026-01-01T00:00:02.123456789Z"},
		{name: "valid/fixed width", value: "2026-01-01T00:00:03.000000000Z"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN: the stored value is parsed.
			got, err := parseTime(tc.value)

			// THEN: it loads, whatever its fractional width.
			if err != nil {
				t.Fatalf(
					"%s\nparseTime(%q)\nunexpected error: %v",
					packageName, tc.value, err,
				)
			}
			if got.IsZero() {
				t.Errorf(
					"%s\nparseTime(%q)\ngot a zero time",
					packageName, tc.value,
				)
			}
		})
	}
}
