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

package config

import (
	"testing"
	"time"

	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// authHardDefaults returns the auth hard defaults Settings.Default() sets.
func authHardDefaults() SettingsBase {
	base := SettingsBase{
		Auth: AuthSettings{
			Enabled: new(false),
			Local: AuthLocalSettings{
				Enabled: new(true),
			},
		},
	}
	base.Auth.Session.Default()

	return base
}

func TestAuthSettings_IsZero(t *testing.T) {
	tests := []struct {
		name string
		auth AuthSettings
		want bool
	}{
		{
			name: "empty",
			auth: AuthSettings{},
			want: true,
		},
		{
			name: "enabled set",
			auth: AuthSettings{Enabled: new(true)},
		},
		{
			name: "session lifetime set",
			auth: AuthSettings{Session: AuthSessionSettings{Lifetime: "1h"}},
		},
		{
			name: "session idle_timeout set",
			auth: AuthSettings{Session: AuthSessionSettings{IdleTimeout: "1h"}},
		},
		{
			name: "local enabled set",
			auth: AuthSettings{Local: AuthLocalSettings{Enabled: new(false)}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.auth.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestAuthSettings_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		auth     AuthSettings
		errRegex string
	}{
		{
			name:     "empty",
			auth:     AuthSettings{},
			errRegex: `^$`,
		},
		{
			name: "valid durations",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					Lifetime:    "720h",
					IdleTimeout: "30m",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "valid/positive max_per_user",
			auth: AuthSettings{
				Session: AuthSessionSettings{MaxPerUser: new(5)},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid/zero max_per_user",
			auth: AuthSettings{
				Session: AuthSessionSettings{MaxPerUser: new(0)},
			},
			errRegex: `max_per_user: "0" <invalid>.*$`,
		},
		{
			name: "invalid/negative max_per_user",
			auth: AuthSettings{
				Session: AuthSessionSettings{MaxPerUser: new(-3)},
			},
			errRegex: "max_per_user",
		},
		{
			name: "invalid lifetime",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					Lifetime: "one-month",
				},
			},
			errRegex: "lifetime",
		},
		{
			name: "invalid idle_timeout",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					IdleTimeout: "0x2h",
				},
			},
			errRegex: "idle_timeout",
		},
		{
			name: "both invalid",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					Lifetime:    "x",
					IdleTimeout: "y",
				},
			},
			errRegex: `lifetime[\s\S]*idle_timeout`,
		},
		{
			name: "invalid/zero lifetime",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					Lifetime: "0s"},
			},
			errRegex: `lifetime: "0s" <invalid> \(must be a positive duration \(e\.g\. '720h'\)\)$`,
		},
		{
			name: "invalid/zero idle_timeout",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					IdleTimeout: "0",
				},
			},
			errRegex: `idle_timeout: "0" <invalid> \(must be a positive duration \(e\.g\. '720h'\)\)$`,
		},
		{
			name: "invalid/negative lifetime",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					Lifetime: "-5h",
				},
			},
			errRegex: `lifetime: "-5h" <invalid> \(must be a positive duration \(e\.g\. '720h'\)\)$`,
		},
		{
			name: "invalid/negative idle_timeout",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					IdleTimeout: "-1m",
				},
			},
			errRegex: `idle_timeout: "-1m" <invalid> \(must be a positive duration \(e\.g\. '720h'\)\)$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.auth.CheckValues()
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestAuthSessionSettings_Default(t *testing.T) {
	var settings AuthSessionSettings
	if !settings.IsZero() {
		t.Fatal("before: AuthSessionSettings.IsZero() mismatch\ngot:  false\nwant: true")
	}

	settings.Default()

	if got, want := settings.Lifetime, "720h"; got != want {
		t.Errorf(
			"after Default: AuthSessionSettings.Lifetime mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
	if got, want := settings.IdleTimeout, "168h"; got != want {
		t.Errorf(
			"after Default: AuthSessionSettings.IdleTimeout mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
	if settings.MaxPerUser == nil {
		t.Fatal("after Default: AuthSessionSettings.MaxPerUser mismatch\ngot:  nil\nwant: non-nil")
	}
	if got, want := *settings.MaxPerUser, 10; got != want {
		t.Errorf(
			"after Default: AuthSessionSettings.MaxPerUser mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}

	lifetime, err := time.ParseDuration(settings.Lifetime)
	if err != nil {
		t.Fatalf("time.ParseDuration(%q) failed: %v", settings.Lifetime, err)
	}
	idle, err := time.ParseDuration(settings.IdleTimeout)
	if err != nil {
		t.Fatalf("time.ParseDuration(%q) failed: %v", settings.IdleTimeout, err)
	}
	// CheckValues rejects an idle timeout past the lifetime, so the defaults keep it inside.
	if idle > lifetime {
		t.Errorf(
			"after Default: AuthSessionSettings.IdleTimeout mismatch\ngot:  %s\nwant: at most %s",
			idle, lifetime,
		)
	}

	if settings.IsZero() {
		t.Error("after Default: AuthSessionSettings.IsZero() mismatch\ngot:  true\nwant: false")
	}
}

func TestSettings_CheckValues__auth(t *testing.T) {
	tests := []struct {
		name      string
		auth      AuthSettings
		yamlBasic *WebSettingsBasicAuth
		flagBasic *WebSettingsBasicAuth
		hardBasic *WebSettingsBasicAuth
		errRegex  string
	}{
		{
			name: "auth disabled with basic auth is fine",
			yamlBasic: &WebSettingsBasicAuth{
				Username: "u",
				Password: "p",
			},
			errRegex: `^$`,
		},
		{
			name: "auth enabled alone is fine",
			auth: AuthSettings{
				Enabled: new(true),
			},
			errRegex: `^$`,
		},
		{
			name: "auth enabled with YAML basic auth conflicts",
			auth: AuthSettings{
				Enabled: new(true),
			},
			yamlBasic: &WebSettingsBasicAuth{
				Username: "u",
				Password: "p",
			},
			errRegex: "web.basic_auth",
		},
		{
			name: "auth enabled with flag basic auth conflicts",
			auth: AuthSettings{
				Enabled: new(true),
			},
			flagBasic: &WebSettingsBasicAuth{
				Username: "u",
				Password: "p",
			},
			errRegex: "web.basic_auth",
		},
		{
			name: "auth enabled with env (hard-default layer) basic auth conflicts",
			auth: AuthSettings{
				Enabled: new(true),
			},
			hardBasic: &WebSettingsBasicAuth{
				Username: "u",
				Password: "p",
			},
			errRegex: "web.basic_auth",
		},
		{
			name: "auth enabled with every provider disabled errors",
			auth: AuthSettings{
				Enabled: new(true),
				Local: AuthLocalSettings{
					Enabled: new(false),
				},
			},
			errRegex: "provider",
		},
		{
			name: "invalid session duration surfaces through Settings",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					Lifetime: "x",
				},
			},
			errRegex: "lifetime",
		},
		{
			name: "auth enabled with an idle window inside the lifetime is fine",
			auth: AuthSettings{
				Enabled: new(true),
				Session: AuthSessionSettings{
					Lifetime:    "48h",
					IdleTimeout: "24h",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "auth enabled with an idle window past the lifetime errors",
			auth: AuthSettings{
				Enabled: new(true),
				Session: AuthSessionSettings{
					Lifetime:    "24h",
					IdleTimeout: "48h",
				},
			},
			errRegex: "idle_timeout",
		},
		{
			name: "auth enabled with an idle window past the default lifetime errors",
			auth: AuthSettings{
				Enabled: new(true),
				Session: AuthSessionSettings{
					IdleTimeout: "1000h",
				},
			},
			errRegex: "idle_timeout",
		},
		{
			name: "auth disabled does not check the session bounds",
			auth: AuthSettings{
				Session: AuthSessionSettings{
					IdleTimeout: "1000h",
				},
			},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{
				Auth:         tc.auth,
				HardDefaults: authHardDefaults(),
			}
			settings.Web.BasicAuth = tc.yamlBasic
			settings.FromFlags.Web.BasicAuth = tc.flagBasic
			settings.HardDefaults.Web.BasicAuth = tc.hardBasic

			err := settings.CheckValues()
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestSettings_Default__auth(t *testing.T) {
	loadMu.Lock()
	defer loadMu.Unlock()
	settings := Settings{}

	if !settings.Default() {
		t.Fatal("ok mismatch\ngot:  false\nwant: true")
	}

	if got, want := test.StringifyPtr(settings.HardDefaults.Auth.Enabled), "false"; got != want {
		t.Errorf("Settings.HardDefaults.Auth.Enabled mismatch\ngot:  %s\nwant: %s", got, want)
	}
	if got, want := settings.HardDefaults.Auth.Session.Lifetime, "720h"; got != want {
		t.Errorf(
			"Settings.HardDefaults.Auth.Session.Lifetime mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
	if got, want := settings.HardDefaults.Auth.Session.IdleTimeout, "168h"; got != want {
		t.Errorf(
			"Settings.HardDefaults.Auth.Session.IdleTimeout mismatch\ngot:  %q\nwant: %q",
			got, want,
		)
	}
	if got, want := test.StringifyPtr(settings.HardDefaults.Auth.Local.Enabled), "true"; got != want {
		t.Errorf("Settings.HardDefaults.Auth.Local.Enabled mismatch\ngot:  %s\nwant: %s", got, want)
	}
}

func TestSettings__authAccessors(t *testing.T) {
	tests := []struct {
		name                    string
		auth                    AuthSettings
		wantEnabled             bool
		wantLocalEnabled        bool
		wantSessionLifetime     time.Duration
		wantSessionIdleTimeout  time.Duration
		wantSessionMaxPerUser   int
		wantSessionSecureCookie *bool
	}{
		{
			name:                   "hard defaults",
			wantEnabled:            false,
			wantLocalEnabled:       true,
			wantSessionLifetime:    720 * time.Hour,
			wantSessionIdleTimeout: 168 * time.Hour,
			wantSessionMaxPerUser:  session.DefaultMaxSessionsPerUser,
		},
		{
			name: "explicit values",
			auth: AuthSettings{
				Enabled: new(true),
				Session: AuthSessionSettings{
					Lifetime:     "48h",
					IdleTimeout:  "2h30m",
					MaxPerUser:   new(3),
					SecureCookie: new(true),
				},
				Local: AuthLocalSettings{Enabled: new(false)},
			},
			wantEnabled:             true,
			wantLocalEnabled:        false,
			wantSessionLifetime:     48 * time.Hour,
			wantSessionIdleTimeout:  2*time.Hour + 30*time.Minute,
			wantSessionMaxPerUser:   3,
			wantSessionSecureCookie: new(true),
		},
		{
			name: "invalid duration falls back to the hard default",
			auth: AuthSettings{
				Session: AuthSessionSettings{Lifetime: "not-a-duration"},
			},
			wantEnabled:            false,
			wantLocalEnabled:       true,
			wantSessionLifetime:    720 * time.Hour,
			wantSessionIdleTimeout: 168 * time.Hour,
			wantSessionMaxPerUser:  session.DefaultMaxSessionsPerUser,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{
				Auth:         tc.auth,
				HardDefaults: authHardDefaults(),
			}

			if got := settings.AuthEnabled(); got != tc.wantEnabled {
				t.Errorf("Settings.AuthEnabled() mismatch\ngot:  %t\nwant: %t", got, tc.wantEnabled)
			}
			if got := settings.AuthLocalEnabled(); got != tc.wantLocalEnabled {
				t.Errorf(
					"Settings.AuthLocalEnabled() mismatch\ngot:  %t\nwant: %t",
					got, tc.wantLocalEnabled,
				)
			}
			if got := settings.AuthSessionLifetime(); got != tc.wantSessionLifetime {
				t.Errorf(
					"Settings.AuthSessionLifetime() mismatch\ngot:  %v\nwant: %v",
					got, tc.wantSessionLifetime,
				)
			}
			if got := settings.AuthSessionIdleTimeout(); got != tc.wantSessionIdleTimeout {
				t.Errorf(
					"Settings.AuthSessionIdleTimeout() mismatch\ngot:  %v\nwant: %v",
					got, tc.wantSessionIdleTimeout,
				)
			}
			if got := settings.AuthSessionMaxPerUser(); got != tc.wantSessionMaxPerUser {
				t.Errorf(
					"Settings.AuthSessionMaxPerUser() mismatch\ngot:  %d\nwant: %d",
					got, tc.wantSessionMaxPerUser,
				)
			}
			gotSecure := test.StringifyPtr(settings.AuthSessionSecureCookie())
			if wantSecure := test.StringifyPtr(tc.wantSessionSecureCookie); gotSecure != wantSecure {
				t.Errorf(
					"Settings.AuthSessionSecureCookie() mismatch\ngot:  %s\nwant: %s",
					gotSecure, wantSecure,
				)
			}
		})
	}
}
