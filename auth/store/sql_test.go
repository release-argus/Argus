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
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestTimeFormat__textOrderIsChronological(t *testing.T) {
	tests := []struct {
		name    string
		earlier string
		later   string
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
				t.Fatalf("earlier: parseTime() failed: %v", err)
			}
			later, err := parseTime(tc.later)
			if err != nil {
				t.Fatalf("later: parseTime() failed: %v", err)
			}

			if !earlier.Before(later) {
				t.Fatal("Time.Before() mismatch\ngot:  false\nwant: true")
			}

			// time.RFC3339Nano trims trailing zeros, so its TEXT form would sort 'Z' above a digit.
			gotEarlier := earlier.UTC().Format(TimeFormat)
			gotLater := later.UTC().Format(TimeFormat)

			// ORDER BY and the expiry sweeps rely on TEXT order matching time order.
			if gotEarlier >= gotLater {
				t.Errorf(
					"TimeFormat text order mismatch\n"+
						"got:  earlier=%q, later=%q\nwant: earlier < later",
					gotEarlier, gotLater,
				)
			}
		})
	}
}

func TestParseTime__acceptsTrimmedFractions(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Time
	}{
		{name: "valid/no fraction", value: "2026-01-01T00:00:00Z", want: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{name: "valid/one digit", value: "2026-01-01T00:00:00.5Z", want: time.Date(2026, 1, 1, 0, 0, 0, 500000000, time.UTC)},
		{name: "valid/two digits", value: "2026-01-01T00:00:01.05Z", want: time.Date(2026, 1, 1, 0, 0, 1, 50000000, time.UTC)},
		{name: "valid/nine digits", value: "2026-01-01T00:00:02.123456789Z", want: time.Date(2026, 1, 1, 0, 0, 2, 123456789, time.UTC)},
		{name: "valid/fixed width", value: "2026-01-01T00:00:03.000000000Z", want: time.Date(2026, 1, 1, 0, 0, 3, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseTime(tc.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("mismatch\ngot:  %v\nwant: %v", got, tc.want)
			}
		})
	}
}

func TestTimeText_Scan(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 600000000, time.UTC)
	tests := []struct {
		name     string
		src      any
		want     time.Time
		errRegex string
	}{
		{
			name:     "valid string",
			src:      when.Format(TimeFormat),
			want:     when,
			errRegex: `^$`,
		},
		{
			name:     "time.Time value",
			src:      when,
			want:     when,
			errRegex: `^$`,
		},
		{
			name:     "invalid string",
			src:      "not-a-time",
			errRegex: `^parse stored timestamp "not-a-time"`,
		},
		{
			name:     "unsupported type",
			src:      42,
			errRegex: `^timestamp: unsupported source int$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var dst time.Time
			err := (timeText{&dst}).Scan(tc.src)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if !dst.Equal(tc.want) {
				t.Errorf("time mismatch\ngot:  %v\nwant: %v", dst, tc.want)
			}
		})
	}
}

func TestNullTimeText_Scan(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 600000000, time.UTC)
	tests := []struct {
		name     string
		src      any
		want     *time.Time
		errRegex string
	}{
		{
			name:     "nil leaves the target nil",
			src:      nil,
			want:     nil,
			errRegex: `^$`,
		},
		{
			name:     "valid string",
			src:      when.Format(TimeFormat),
			want:     &when,
			errRegex: `^$`,
		},
		{
			name:     "invalid string",
			src:      "not-a-time",
			errRegex: `^parse stored timestamp "not-a-time"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var dst *time.Time
			err := (nullTimeText{&dst}).Scan(tc.src)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if tc.want == nil {
				if dst != nil {
					t.Errorf("time mismatch\ngot:  %v\nwant: nil", *dst)
				}
				return
			}

			if dst == nil {
				t.Fatalf("time mismatch\ngot:  nil\nwant: %v", *tc.want)
			}
			if !dst.Equal(*tc.want) {
				t.Errorf("time mismatch\ngot:  %v\nwant: %v", *dst, *tc.want)
			}
		})
	}
}

func TestRowUpdate(t *testing.T) {
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
			test.ExecSQL(t, store.db,
				`CREATE TABLE tmp (
					id TEXT PRIMARY KEY, a TEXT, b TEXT, updated_at DATETIME
				);`)
			test.ExecSQL(t, store.db,
				`INSERT INTO tmp (id, a, b, updated_at)
				VALUES ('x', 'A1', 'B1', 'stale');`)

			upd := newRowUpdate()
			for _, s := range tc.sets {
				upd.set(s.col, s.val)
			}
			err := store.inTx(t.Context(), func(tx *sql.Tx) error {
				return upd.exec(t.Context(), tx, "tmp", "x")
			})
			if err != nil {
				t.Fatalf("Store.inTx() failed: %v", err)
			}

			var a, b, updatedAt string
			if err := store.db.QueryRow(
				`SELECT a, b, updated_at FROM tmp WHERE id = 'x';`,
			).Scan(&a, &b, &updatedAt); err != nil {
				t.Fatalf("DB.QueryRow().Scan() failed: %v", err)
			}
			if a != tc.wantA || b != tc.wantB {
				t.Errorf(
					"tmp row mismatch\ngot:  a=%q, b=%q\nwant: a=%q, b=%q",
					a, b,
					tc.wantA, tc.wantB,
				)
			}

			if updatedAt == "stale" {
				t.Errorf("tmp.updated_at mismatch\ngot:  %q\nwant: NOT %q", updatedAt, "stale")
			}
		})
	}
}
