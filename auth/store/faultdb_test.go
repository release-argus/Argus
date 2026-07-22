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

	storetest "github.com/release-argus/Argus/auth/store/test"
)

// errInjected is the error produced by the fault-injecting test driver.
var errInjected = storetest.ErrInjected

// testFaultDB opens an in-memory database whose statements can be failed
// selectively via the returned FaultState.
func testFaultDB(t *testing.T) (*sql.DB, *storetest.FaultState) {
	t.Helper()
	return storetest.FaultDB(t)
}

// testFaultStore initialises a Store (fault disarmed) on a fault-injecting DB.
func testFaultStore(t *testing.T) (*Store, *storetest.FaultState) {
	t.Helper()

	db, state := testFaultDB(t)
	store, err := New(t.Context(), db)
	if err != nil {
		t.Fatalf(
			"%s\nNew() on fault DB failed: %v",
			packageName, err,
		)
	}
	return store, state
}

// overrideRowsErrOnCall makes rowsErr fail on its nth call.
func overrideRowsErrOnCall(t *testing.T, n int) {
	t.Helper()

	rowsErrHad := rowsErr
	calls := 0
	rowsErr = func(rows *sql.Rows) error {
		calls++
		if calls == n {
			return errInjected
		}
		return rowsErrHad(rows)
	}
	t.Cleanup(func() { rowsErr = rowsErrHad })
}
