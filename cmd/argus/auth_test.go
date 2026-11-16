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
	"github.com/release-argus/Argus/internal/test"
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

// saveAuthFlags snapshots and restores the auth flag vars.
func saveAuthFlags(t *testing.T) {
	t.Helper()

	resetPasswordHad := config.AuthResetPassword
	createAdminHad := config.AuthCreateAdmin
	t.Cleanup(func() {
		config.AuthResetPassword = resetPasswordHad
		config.AuthCreateAdmin = createAdminHad
	})
}

func TestSetupAuth(t *testing.T) {
	// presetHash is the preset user's hash, which a reset must replace.
	const presetHash = "$argon2id$v=19$m=65536,t=2,p=1$c2FsdHNhbHRzYWx0$aGFzaGhhc2hoYXNoaGFzaA"

	tests := []struct {
		name              string
		enabled           bool
		localDisabled     bool
		resetPassword     string // Username for -auth.reset-password.
		createAdmin       string // Username for -auth.create-admin.
		closeDB           bool
		faultPattern      string // Fail statements containing this (post-init).
		presetUser        bool   // A user already exists (setup completed).
		wantDeps          bool
		wantOK            bool
		wantLocalProvider bool
		wantResetExit     bool
		wantAdminExit     bool
		wantFatal         bool
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
			name:      "broken database fails fatally",
			enabled:   true,
			closeDB:   true,
			wantFatal: true,
		},
		{
			name:          "reset-password/flag resets the named user then exits",
			enabled:       true,
			presetUser:    true,
			resetPassword: "existing",
			wantResetExit: true,
		},
		{
			name:          "reset-password/flag for an unknown user fails fatally",
			enabled:       true,
			presetUser:    true,
			resetPassword: "ghost",
			wantFatal:     true,
		},
		{
			name:          "reset-password/store failure fails fatally",
			enabled:       true,
			presetUser:    true,
			resetPassword: "existing",
			faultPattern:  `UPDATE users SET password_hash`,
			wantFatal:     true,
		},
		{
			name:          "create-admin/creates the first administrator then exits",
			enabled:       true,
			createAdmin:   "root",
			wantAdminExit: true,
		},
		{
			name:          "create-admin/trims the username",
			enabled:       true,
			createAdmin:   "  root  ",
			wantAdminExit: true,
		},
		{
			name:        "create-admin/only whitespace fails fatally",
			enabled:     true,
			createAdmin: "   ",
			wantFatal:   true,
		},
		{
			name:        "create-admin/fails fatally once a user exists",
			enabled:     true,
			presetUser:  true,
			createAdmin: "root",
			wantFatal:   true,
		},
		{
			name:         "create-admin/store failure fails fatally",
			enabled:      true,
			createAdmin:  "root",
			faultPattern: `INSERT INTO users`,
			wantFatal:    true,
		},
		{
			name:        "create-admin/auth disabled fails fatally",
			createAdmin: "root",
			wantFatal:   true,
		},
		{
			name:          "reset-password/auth disabled fails fatally",
			resetPassword: "existing",
			wantFatal:     true,
		},
		{
			name:          "create-admin alongside reset-password fails fatally",
			enabled:       true,
			createAdmin:   "root",
			resetPassword: "existing",
			wantFatal:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps exitAfterOneShot and the auth flag vars.

			saveAuthFlags(t)
			config.AuthResetPassword = new(tc.resetPassword)
			config.AuthCreateAdmin = new(tc.createAdmin)

			// Capture the one-shot exit instead of ending the test process.
			var resetExitCode *int
			exitAfterOneShotHad := exitAfterOneShot
			exitAfterOneShot = func(code int) { resetExitCode = &code }
			t.Cleanup(func() { exitAfterOneShot = exitAfterOneShotHad })

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
				db = test.MemoryDB(t)
			}
			if tc.presetUser {
				authStore, err := store.New(t.Context(), db)
				if err != nil {
					t.Fatalf("preset: store.New() failed: %v", err)
				}
				if _, err := authStore.CreateUser(
					t.Context(),
					"existing", "", "",
					presetHash,
					nil,
				); err != nil {
					t.Fatalf("preset: Store.CreateUser() failed: %v", err)
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

			deps, ok := setupAuth(gCtx, g, cfg, db)

			// Stop the prune goroutine (if started) and drain any Fatal.
			cancel()
			_ = g.Wait()
			fatals := 0
			for len(logx.ExitCodeChannel()) > 0 {
				<-logx.ExitCodeChannel()
				fatals++
			}

			if ok != tc.wantOK {
				t.Fatalf("ok mismatch\ngot:  %t\nwant: %t", ok, tc.wantOK)
			}
			if gotFatal := fatals > 0; gotFatal != tc.wantFatal {
				t.Errorf(
					"logx.Fatal() called mismatch\ngot:  %t (%d calls)\nwant: %t",
					gotFatal, fatals, tc.wantFatal,
				)
			}
			if gotDeps := (deps != nil); gotDeps != tc.wantDeps {
				t.Fatalf("deps mismatch\ngot:  non-nil=%t\nwant: non-nil=%t", gotDeps, tc.wantDeps)
			}

			if tc.wantResetExit {
				if resetExitCode == nil || *resetExitCode != 0 {
					t.Errorf(
						"reset: exit code mismatch\ngot:  %s\nwant: 0",
						test.StringifyPtr(resetExitCode),
					)
				}
				verifyStore, err := store.New(t.Context(), db)
				if err != nil {
					t.Fatalf("reset: store.New() failed: %v", err)
				}
				creds, err := verifyStore.LocalCredentials(t.Context(), tc.resetPassword)
				if err != nil {
					t.Fatalf("reset: Store.LocalCredentials() failed: %v", err)
				}
				if creds == nil {
					t.Fatal("reset: Store.LocalCredentials() mismatch\ngot:  nil\nwant: non-nil")
				}

				if creds.PasswordHash == "" || creds.PasswordHash == presetHash {
					t.Errorf(
						"reset: Credentials.PasswordHash mismatch\n"+
							"got:  %q\nwant: non-empty, NOT %q",
						creds.PasswordHash, presetHash,
					)
				}
			}

			if tc.wantAdminExit {
				if resetExitCode == nil || *resetExitCode != 0 {
					t.Errorf(
						"create-admin: exit code mismatch\ngot:  %s\nwant: 0",
						test.StringifyPtr(resetExitCode),
					)
				}
				verifyStore, err := store.New(t.Context(), db)
				if err != nil {
					t.Fatalf("create-admin: store.New() failed: %v", err)
				}

				username := strings.TrimSpace(tc.createAdmin)
				creds, err := verifyStore.LocalCredentials(t.Context(), username)
				if err != nil {
					t.Fatalf("create-admin: Store.LocalCredentials() failed: %v", err)
				}
				if creds == nil {
					t.Fatal(
						"create-admin: Store.LocalCredentials() mismatch\n" +
							"got:  nil\nwant: non-nil",
					)
				}
				if creds.PasswordHash == "" {
					t.Error(
						"create-admin: Credentials.PasswordHash mismatch\n" +
							"got:  \"\"\nwant: non-empty",
					)
				}
				if !creds.Enabled {
					t.Error("create-admin: Credentials.Enabled mismatch\ngot:  false\nwant: true")
				}

				_, grants, err := verifyStore.UserWithGrants(t.Context(), creds.UserID)
				if err != nil {
					t.Fatalf("create-admin: Store.UserWithGrants() failed: %v", err)
				}
				if len(grants) == 0 {
					t.Error("create-admin: len(grants) mismatch\ngot:  0\nwant: > 0")
				}
			}

			if tc.enabled && tc.createAdmin != "" && !tc.wantAdminExit {
				verifyStore, err := store.New(t.Context(), db)
				if err != nil {
					t.Fatalf("refused create-admin: store.New() failed: %v", err)
				}
				username := strings.TrimSpace(tc.createAdmin)
				creds, err := verifyStore.LocalCredentials(t.Context(), username)
				if err != nil {
					t.Fatalf("refused create-admin: Store.LocalCredentials() failed: %v", err)
				}
				if creds != nil {
					t.Errorf(
						"refused create-admin: Store.LocalCredentials() mismatch\n"+
							"got:  Username=%q\nwant: nil",
						creds.Username,
					)
				}
			}

			if !tc.wantDeps {
				return
			}

			if got := deps.Providers.Get(local.Name) != nil; got != tc.wantLocalProvider {
				t.Errorf(
					"Registry.Get(local.Name) != nil mismatch\ngot:  %t\nwant: %t",
					got, tc.wantLocalProvider,
				)
			}

			// The first admin comes from first-run setup, not startup.
			if !tc.presetUser {
				count, err := deps.Store.CountUsers(t.Context())
				if err != nil || count != 0 {
					t.Errorf(
						"Store.CountUsers() mismatch\ngot:  %d, err=%v\nwant: 0, err=nil",
						count, err,
					)
				}
			}
		})
	}
}

func TestSetupAuth__prunes(t *testing.T) {
	tests := []struct {
		name         string
		faultPattern string   // Fail statements containing this.
		wantPruned   []string // Table names.
		wantKept     []string // Table names.
	}{
		{
			name:       "both prunes run every tick",
			wantPruned: []string{"sessions", "api_tokens"},
		},
		{
			name:         "a session prune failure does not stop the loop",
			faultPattern: `DELETE FROM sessions WHERE expires_at`,
			wantPruned:   []string{"api_tokens"},
			wantKept:     []string{"sessions"},
		},
		{
			name:         "an API token prune failure does not stop the loop",
			faultPattern: `DELETE FROM api_tokens WHERE expires_at`,
			wantPruned:   []string{"sessions"},
			wantKept:     []string{"api_tokens"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps pruneTicker.

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

			_, ok := setupAuth(gCtx, g, testAuthConfig(true), db)
			if !ok {
				t.Fatal("setupAuth() failed")
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
						t.Fatal("prune loop stopped consuming ticks")
					}
				}
			}

			for i := range 5 {
				t.Logf("pass %d", i)
				presetExpiredRows(t, db)
				prunePass()
				for _, table := range tc.wantPruned {
					assertRowCount(t, db, table, 0)
				}
				for _, table := range tc.wantKept {
					assertRowCount(t, db, table, 1)
				}
			}

			cancel()
			state.Set("")
			if err := g.Wait(); err != nil {
				t.Errorf("after cancel: Group.Wait() failed: %v", err)
			}
		})
	}
}

// presetExpiredRows plants an expired session and API token,
// initialising the schema first if needed.
func presetExpiredRows(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := store.New(t.Context(), db); err != nil {
		t.Fatalf("store.New() failed: %v", err)
	}
	expired := time.Now().UTC().Add(-time.Hour).Format(store.TimeFormat)
	for i, statement := range []string{
		`INSERT OR REPLACE INTO sessions (token_hash, user_id, created_at, expires_at, last_seen_at, ip, user_agent)
			VALUES ('stale', 'nobody', ?, ?, ?, '', '');`,
		`INSERT OR REPLACE INTO api_tokens (id, user_id, name, token_hash, prefix, created_at, expires_at)
			VALUES ('stale', 'nobody', 'ci', 'stale-hash', 'argus_', ?, ?);`,
	} {
		args := slices.Repeat([]any{expired}, strings.Count(statement, "?"))
		if _, err := db.Exec(statement, args...); err != nil {
			t.Fatalf("statement %d: DB.Exec() failed: %v", i, err)
		}
	}
}

// assertRowCount fails when table does not hold want rows.
func assertRowCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()

	if got := test.CountRows(t, db, table); got != want {
		t.Errorf("%s row count mismatch\ngot:  %d\nwant: %d", table, got, want)
	}
}
