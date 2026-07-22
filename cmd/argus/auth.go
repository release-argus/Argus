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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/auth/provider"
	"github.com/release-argus/Argus/auth/provider/local"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	v1 "github.com/release-argus/Argus/web/api/v1"
)

// sessionPruneInterval is how often abandoned sessions are pruned.
const sessionPruneInterval = 6 * time.Hour

// setupAuth initialises the auth subsystem when auth is enabled:
// the store (migrations, permission sync, seeded groups), password recovery,
// the session manager (with periodic pruning), and the provider registry.
// The first administrator is created through the UI's first-run setup page.
// Returns (nil, true) when auth is disabled and (nil, false) after a fatal error.
func setupAuth(ctx context.Context, g *errgroup.Group, cfg *config.Config, dbHandle *sql.DB) (*v1.AuthDeps, bool) {
	if !cfg.Settings.AuthEnabled() {
		return nil, true
	}

	logFrom := logx.LogFrom{Primary: "auth"}

	authStore, err := store.New(ctx, dbHandle)
	if err != nil {
		logx.Fatal(err, logFrom)
		return nil, false
	}

	// -auth.reset-password <username>: lockout recovery - set a generated
	// password on the named user and revoke their sessions.
	if config.AuthResetPassword != nil && *config.AuthResetPassword != "" {
		username := *config.AuthResetPassword
		password := util.RandAlphaNumericLower(24)
		if err := authStore.ResetUserPassword(ctx, username, password); err != nil {
			logx.Fatal(err, logFrom)
			return nil, false
		}
		fmt.Printf(
			"Password for user %q has been reset to: %q\n",
			username, password,
		)
	}

	// Sessions, with periodic pruning of abandoned/expired rows.
	sessions := session.New(
		authStore,
		session.Config{
			Lifetime:    cfg.Settings.AuthSessionLifetime(),
			IdleTimeout: cfg.Settings.AuthSessionIdleTimeout(),
		},
	)
	g.Go(func() error {
		ticker := time.NewTicker(sessionPruneInterval)
		defer ticker.Stop()
		for {
			if err := sessions.PruneExpired(ctx); err != nil {
				logx.Error(err, logFrom, true)
			}
			if err := authStore.DeleteExpiredAPITokens(ctx); err != nil {
				logx.Error(err, logFrom, true)
			}
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	})

	// Providers.
	registry := provider.NewRegistry()
	if cfg.Settings.AuthLocalEnabled() {
		registry.Register(local.New(authStore))
	}

	return &v1.AuthDeps{
		Store:     authStore,
		Sessions:  sessions,
		Providers: registry,
	}, true
}
