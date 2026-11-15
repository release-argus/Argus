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

package test

import (
	"regexp"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/release-argus/Argus/util"
)

func TestDropTable(t *testing.T) {
	tests := []struct {
		name  string
		table string
		other string
	}{
		{name: "first table", table: "alpha", other: "beta"},
		{name: "second table", table: "beta", other: "alpha"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := MemoryDB(t)
			for _, table := range []string{"alpha", "beta"} {
				stmt := `CREATE TABLE ` + table + ` (id TEXT);`
				if _, err := db.ExecContext(t.Context(), stmt); err != nil {
					t.Fatalf("DB.ExecContext(%q) failed: %v", stmt, err)
				}
			}

			DropTable(t, db, tc.table)

			_, err := db.ExecContext(
				t.Context(),
				`INSERT INTO `+tc.table+` (id) VALUES ('x');`,
			)
			wantErr := "no such table"
			if err == nil || !strings.Contains(err.Error(), wantErr) {
				t.Errorf(
					"dropped: DB.ExecContext() error mismatch\ngot:  %v\nwant: %q",
					err, wantErr,
				)
			}

			if _, err := db.ExecContext(
				t.Context(),
				`INSERT INTO `+tc.other+` (id) VALUES ('x');`,
			); err != nil {
				t.Errorf("other: DB.ExecContext() error mismatch\ngot:  %v\nwant: nil", err)
			}
		})
	}
}

func TestDropTable__unknownTable(t *testing.T) {
	t.Parallel()

	db := MemoryDB(t)
	fakeT := &FakeT{Abort: true}

	aborted := fakeT.Aborted(func() {
		DropTable(fakeT, db, "missing")
	})

	if !aborted {
		t.Error("FakeT.Aborted() mismatch\ngot:  false\nwant: true")
	}
	errRegex := `^DB\.ExecContext\("DROP TABLE missing;"\) failed: .*no such table: missing`
	if fatals := strings.Join(fakeT.Fatals, "\n"); !util.RegexCheck(errRegex, fatals) {
		t.Errorf("FakeT.Fatals mismatch\ngot:  %q\nwant: %q", fatals, errRegex)
	}
}

func TestExecSQL(t *testing.T) {
	tests := []struct {
		name       string
		statements []string
		wantRows   int
		wantFatal  string
	}{
		{
			name: "runs every statement in order",
			statements: []string{
				`CREATE TABLE things (id TEXT);`,
				`INSERT INTO things (id) VALUES ('a');`,
				`INSERT INTO things (id) VALUES ('b');`,
			},
			wantRows: 2,
		},
		{
			name: "stops at the first failing statement",
			statements: []string{
				`CREATE TABLE things (id TEXT);`,
				`INSERT INTO missing (id) VALUES ('a');`,
				`INSERT INTO things (id) VALUES ('b');`,
			},
			wantRows:  0,
			wantFatal: `^statement 1: DB.ExecContext\("INSERT INTO missing [^"]*"\) failed: .*no such table: missing`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := MemoryDB(t)
			fakeT := &FakeT{Abort: true}

			aborted := fakeT.Aborted(func() { ExecSQL(fakeT, db, tc.statements...) })

			wantFatals := 0
			if tc.wantFatal != "" {
				wantFatals = 1
			}
			if aborted != (wantFatals == 1) || len(fakeT.Fatals) != wantFatals {
				t.Fatalf(
					"FakeT mismatch\ngot:  aborted=%t, Fatals=%q\nwant: aborted=%t, %d Fatals",
					aborted, fakeT.Fatals,
					wantFatals == 1, wantFatals,
				)
			}
			if wantFatals == 1 && !regexp.MustCompile(tc.wantFatal).MatchString(fakeT.Fatals[0]) {
				t.Errorf(
					"FakeT.Fatals[0] mismatch\ngot:  %q\nwant: %q",
					fakeT.Fatals[0], tc.wantFatal,
				)
			}

			if got := CountRows(t, db, "things"); got != tc.wantRows {
				t.Errorf("CountRows() mismatch\ngot:  %d\nwant: %d", got, tc.wantRows)
			}
		})
	}
}

func TestCountRows(t *testing.T) {
	tests := []struct {
		name        string
		rows        int
		table       string
		want        int
		errRegex    string
		wantAborted bool
	}{
		{
			name:     "valid/empty table",
			table:    "things",
			want:     0,
			errRegex: `^$`,
		},
		{
			name:     "valid/populated table",
			rows:     3,
			table:    "things",
			want:     3,
			errRegex: `^$`,
		},
		{
			name:        "invalid/unknown table",
			table:       "missing",
			want:        -1,
			errRegex:    `^DB\.QueryRowContext\("SELECT COUNT\(\*\) FROM missing;"\) failed: .*no such table: missing`,
			wantAborted: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := MemoryDB(t)
			ExecSQL(t, db, `CREATE TABLE things (id INTEGER);`)
			for i := range tc.rows {
				if _, err := db.ExecContext(
					t.Context(), `INSERT INTO things (id) VALUES (?);`, i,
				); err != nil {
					t.Fatalf("row %d: DB.ExecContext() failed: %v", i, err)
				}
			}
			fakeT := &FakeT{Abort: true}

			// Stays -1 if CountRows aborts.
			got := -1
			aborted := fakeT.Aborted(func() {
				got = CountRows(fakeT, db, tc.table)
			})

			if got != tc.want {
				t.Errorf("count mismatch\ngot:  %d\nwant: %d", got, tc.want)
			}

			if fatals := strings.Join(fakeT.Fatals, "\n"); !util.RegexCheck(tc.errRegex, fatals) {
				t.Errorf("FakeT.Fatals mismatch\ngot:  %q\nwant: %q", fatals, tc.errRegex)
			}

			if aborted != tc.wantAborted {
				t.Errorf("FakeT.Aborted() mismatch\ngot:  %t\nwant: %t", aborted, tc.wantAborted)
			}
		})
	}
}

func TestMemoryDB(t *testing.T) {
	t.Parallel()

	db := MemoryDB(t)

	ExecSQL(t, db, `CREATE TABLE things (id TEXT);`, `INSERT INTO things (id) VALUES ('a');`)
	if got := CountRows(t, db, "things"); got != 1 {
		t.Errorf("CountRows() mismatch\ngot:  %d\nwant: 1", got)
	}
	if got := db.Stats().MaxOpenConnections; got != 1 {
		t.Errorf("DB.Stats().MaxOpenConnections mismatch\ngot:  %d\nwant: 1", got)
	}
}
