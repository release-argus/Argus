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

//go:build integration

package main

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/auth/provider/local"
	"github.com/release-argus/Argus/auth/store"
	storetest "github.com/release-argus/Argus/auth/store/test"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
)

// testAuthConfig returns a Config with auth enabled and hard defaults set.
func testAuthConfig(enabled bool) *config.Config {
	cfg := &config.Config{}
	cfg.Settings.Auth.Enabled = new(enabled)
	cfg.Settings.HardDefaults.Auth = config.AuthSettings{
		Enabled: new(false),
		Session: config.AuthSessionSettings{
			Lifetime:    "720h",
			IdleTimeout: "168h",
		},
		Local: config.AuthLocalSettings{Enabled: new(true)},
	}
	return cfg
}

// testAuthDB opens a fresh in-memory SQLite database.
func testAuthDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf(
			"%s open test database: %v",
			packageName, err,
		)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// saveAuthFlags snapshots and restores the auth flag vars.
func saveAuthFlags(t *testing.T) {
	t.Helper()

	resetPasswordHad := config.AuthResetPassword
	t.Cleanup(func() {
		config.AuthResetPassword = resetPasswordHad
	})
}

func TestSetupAuth(t *testing.T) {
	// presetHash is the hash the preset user is created with; a reset must
	// replace it, so the assertion below checks the hash actually changed.
	const presetHash = "$argon2id$v=19$m=65536,t=2,p=1$c2FsdHNhbHRzYWx0$aGFzaGhhc2hoYXNoaGFzaA"

	// GIVEN: configs and flag/env states.
	tests := []struct {
		name              string
		enabled           bool
		localDisabled     bool
		resetPassword     string // Username for -auth.reset-password.
		closeDB           bool
		faultPattern      string // Fail statements containing this (post-init).
		presetUser        bool   // A user already exists (setup completed).
		wantDeps          bool
		wantOK            bool
		wantLocalProvider bool
		wantResetExit     bool
	}{
		{
			name:   "auth disabled returns no deps",
			wantOK: true,
		},
		{
			name:              "first start leaves setup pending",
			enabled:           true,
			wantDeps:          true,
			wantOK:            true,
			wantLocalProvider: true,
		},
		{
			name:              "existing users pass straight through",
			enabled:           true,
			presetUser:        true,
			wantDeps:          true,
			wantOK:            true,
			wantLocalProvider: true,
		},
		{
			name:              "local provider disabled registers nothing",
			enabled:           true,
			localDisabled:     true,
			wantDeps:          true,
			wantOK:            true,
			wantLocalProvider: false,
		},
		{
			name:    "broken database fails fatally",
			enabled: true,
			closeDB: true,
		},
		{
			name:          "reset-password flag resets the named user then exits",
			enabled:       true,
			presetUser:    true,
			resetPassword: "existing",
			wantResetExit: true,
		},
		{
			name:          "reset-password for an unknown user fails fatally",
			enabled:       true,
			presetUser:    true,
			resetPassword: "ghost",
		},
		{
			name:          "reset-password store failure fails fatally",
			enabled:       true,
			presetUser:    true,
			resetPassword: "existing",
			faultPattern:  `UPDATE users SET password_hash`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			saveAuthFlags(t)
			config.AuthResetPassword = new(tc.resetPassword)

			// Capture the one-shot reset exit instead of ending the test process.
			var resetExitCode *int
			exitAfterResetHad := exitAfterReset
			exitAfterReset = func(code int) { resetExitCode = &code }
			t.Cleanup(func() { exitAfterReset = exitAfterResetHad })

			cfg := testAuthConfig(tc.enabled)
			if tc.localDisabled {
				cfg.Settings.Auth.Local.Enabled = new(false)
			}

			var (
				db    *sql.DB
				state *storetest.FaultState
			)
			if tc.faultPattern != "" {
				db, state = storetest.FaultDB(t)
			} else {
				db = testAuthDB(t)
			}
			if tc.presetUser {
				authStore, err := store.New(t.Context(), db)
				if err != nil {
					t.Fatalf(
						"%s setup store: %v",
						packageName, err,
					)
				}
				if _, err := authStore.CreateUser(
					t.Context(),
					"existing", "", "",
					presetHash,
					nil,
				); err != nil {
					t.Fatalf(
						"%s setup user: %v",
						packageName, err,
					)
				}
			}
			if tc.closeDB {
				_ = db.Close()
			}
			if state != nil {
				state.Set(tc.faultPattern)
			}

			ctx, cancel := context.WithCancel(t.Context())
			g, gCtx := errgroup.WithContext(ctx)

			prefix := fmt.Sprintf("%s\nsetupAuth", packageName)

			// WHEN: setupAuth runs.
			deps, ok := setupAuth(gCtx, g, cfg, db)

			// Stop the prune goroutine (if started) and drain any Fatal.
			cancel()
			_ = g.Wait()
			for len(logx.ExitCodeChannel()) > 0 {
				<-logx.ExitCodeChannel()
			}

			// THEN: the outcome matches expectations.
			if ok != tc.wantOK {
				t.Fatalf(
					"%s ok mismatch\ngot:  %t\nwant: %t",
					prefix, ok, tc.wantOK,
				)
			}
			if (deps != nil) != tc.wantDeps {
				t.Fatalf(
					"%s deps mismatch\ngot:  %v\nwantDeps: %t",
					prefix, deps, tc.wantDeps,
				)
			}

			// AND: the one-shot reset flow exits after replacing the password,
			// without returning deps (the server never starts).
			if tc.wantResetExit {
				if resetExitCode == nil || *resetExitCode != 0 {
					t.Errorf(
						"%s reset should exit(0)\ngot: %v",
						prefix, resetExitCode,
					)
				}
				verifyStore, err := store.New(t.Context(), db)
				if err != nil {
					t.Fatalf(
						"%s verify store: %v",
						prefix, err,
					)
				}
				creds, err := verifyStore.LocalCredentials(t.Context(), tc.resetPassword)
				if err != nil || creds == nil {
					t.Fatalf(
						"%s reset user lookup failed: %v",
						prefix, err,
					)
				}
				// AND: the reset must replace the preset hash, not leave it untouched.
				if creds.PasswordHash == "" || creds.PasswordHash == presetHash {
					t.Errorf(
						"%s reset should have replaced the password hash\ngot: %q",
						prefix, creds.PasswordHash,
					)
				}
			}

			if !tc.wantDeps {
				return
			}

			// AND: the provider registry matches the local setting.
			if got := deps.Providers.Get(local.Name) != nil; got != tc.wantLocalProvider {
				t.Errorf(
					"%s local provider mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.wantLocalProvider,
				)
			}

			// AND: no account is created at startup - the first admin comes
			// from the UI's first-run setup.
			if !tc.presetUser {
				count, err := deps.Store.CountUsers(t.Context())
				if err != nil || count != 0 {
					t.Errorf(
						"%s no users should exist until first-run setup\ngot:  %d, err=%v",
						prefix, count, err,
					)
				}
			}

		})
	}
}

func TestSetupAuth__prunes(t *testing.T) {
	// GIVEN: an auth database holding an expired session and API token, and a
	// prune loop that fails one, both, or neither of the two prune statements.
	tests := []struct {
		name         string
		faultPattern string   // Fail statements containing this.
		wantPruned   []string // tables.
	}{
		{
			name:       "both prunes run every tick",
			wantPruned: []string{"sessions", "api_tokens"},
		},
		{
			name:         "a session prune failure does not stop the loop",
			faultPattern: `DELETE FROM sessions WHERE expires_at`,
			wantPruned:   []string{"api_tokens"},
		},
		{
			name:         "an API token prune failure does not stop the loop",
			faultPattern: `DELETE FROM api_tokens WHERE expires_at`,
			wantPruned:   []string{"sessions"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're swapping the
			// shared pruneTicker for a channel of our own.

			saveAuthFlags(t)
			config.AuthResetPassword = new("")

			// Drive the loop by hand rather than waiting on a real ticker.
			ticks := make(chan time.Time)
			pruneTickerHad := pruneTicker
			pruneTicker = func() (<-chan time.Time, func()) { return ticks, func() {} }
			t.Cleanup(func() { pruneTicker = pruneTickerHad })

			db, state := storetest.FaultDB(t)
			presetExpiredRows(t, db)
			state.Set(tc.faultPattern)

			ctx, cancel := context.WithCancel(t.Context())
			g, gCtx := errgroup.WithContext(ctx)

			prefix := fmt.Sprintf("%s\nsetupAuth prune loop", packageName)

			// WHEN: setupAuth starts the prune loop.
			if _, ok := setupAuth(gCtx, g, testAuthConfig(true), db); !ok {
				t.Fatalf("%s setup failed", prefix)
			}

			// prunePass returns once a prune loop has finished a pass.
			// The first ends the pass already in flight,
			// and the second ends the one that send started.
			prunePass := func() {
				t.Helper()

				for range 2 {
					select {
					case ticks <- time.Time{}:
					case <-time.After(10 * time.Second):
						t.Fatalf("%s prune loop stopped consuming ticks", prefix)
					}
				}
			}

			// THEN: the surviving prune empties its table, repeatedly.
			for range 5 {
				presetExpiredRows(t, db)
				prunePass()
				for _, table := range tc.wantPruned {
					assertEmpty(t, db, table, prefix)
				}
			}

			// AND: cancelling the context ends the loop.
			cancel()
			state.Set("")
			if err := g.Wait(); err != nil {
				t.Errorf(
					"%s prune loop should exit cleanly on context cancellation\ngot: %v",
					prefix, err,
				)
			}
		})
	}
}

// presetExpiredRows plants an expired session and API token,
// initialising the schema first if needed.
func presetExpiredRows(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := store.New(t.Context(), db); err != nil {
		t.Fatalf(
			"%s setup store: %v",
			packageName, err,
		)
	}
	expired := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	for _, statement := range []string{
		`INSERT OR REPLACE INTO sessions (token_hash, user_id, created_at, expires_at, last_seen_at, ip, user_agent)
			VALUES ('stale', 'nobody', ?, ?, ?, '', '');`,
		`INSERT OR REPLACE INTO api_tokens (id, user_id, name, token_hash, prefix, created_at, expires_at)
			VALUES ('stale', 'nobody', 'ci', 'stale-hash', 'argus_', ?, ?);`,
	} {
		args := slices.Repeat([]any{expired}, strings.Count(statement, "?"))
		if _, err := db.Exec(statement, args...); err != nil {
			t.Fatalf(
				"%s setup expired row: %v",
				packageName, err,
			)
		}
	}
}

// assertEmpty fails when table still holds rows.
func assertEmpty(t *testing.T, db *sql.DB, table, prefix string) {
	t.Helper()

	if got := countRows(t, db, table); got != 0 {
		t.Errorf(
			"%s %s should have been pruned\ngot: %d rows",
			prefix, table, got,
		)
	}
}
