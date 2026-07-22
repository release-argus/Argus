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
	"testing"

	"golang.org/x/sync/errgroup"

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
	// GIVEN: configs and flag/env states.
	tests := []struct {
		name          string
		enabled       bool
		localDisabled bool
		resetPassword string // Username for -auth.reset-password.
		closeDB       bool
		faultPattern  string // Fail statements containing this (post-init).
		presetUser    bool   // A user already exists (setup completed).
		wantDeps      bool
		wantOK        bool
		wantProviders int
	}{
		{
			name:   "auth disabled returns no deps",
			wantOK: true,
		},
		{
			name:          "first start leaves setup pending",
			enabled:       true,
			wantDeps:      true,
			wantOK:        true,
			wantProviders: 1,
		},
		{
			name:          "existing users pass straight through",
			enabled:       true,
			presetUser:    true,
			wantDeps:      true,
			wantOK:        true,
			wantProviders: 1,
		},
		{
			name:          "local provider disabled registers nothing",
			enabled:       true,
			localDisabled: true,
			wantDeps:      true,
			wantOK:        true,
			wantProviders: 0,
		},
		{
			name:    "broken database fails fatally",
			enabled: true,
			closeDB: true,
		},
		{
			name:          "reset-password flag resets the named user",
			enabled:       true,
			presetUser:    true,
			resetPassword: "existing",
			wantDeps:      true,
			wantOK:        true,
			wantProviders: 1,
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
					"existing", "", "", "", nil,
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
			if !tc.wantDeps {
				return
			}

			// AND: the provider registry matches the local setting.
			if got := len(deps.Providers.Names()); got != tc.wantProviders {
				t.Errorf(
					"%s provider count mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wantProviders,
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

			// AND: the named user's (generated) password is set on reset.
			if tc.resetPassword != "" {
				creds, err := deps.Store.LocalCredentials(t.Context(), tc.resetPassword)
				if err != nil || creds == nil {
					t.Fatalf(
						"%s reset user lookup failed: %v",
						prefix, err,
					)
				}
				// The preset user starts with no password hash.
				if creds.PasswordHash == "" {
					t.Errorf(
						"%s reset should have set a password hash",
						prefix,
					)
				}
			}
		})
	}
}
