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
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// authHardDefaults returns Settings hard defaults matching Settings.Default().
func authHardDefaults() SettingsBase {
	return SettingsBase{
		Auth: AuthSettings{
			Enabled: new(false),
			Session: AuthSessionSettings{
				Lifetime:    "720h",
				IdleTimeout: "168h",
			},
			Local: AuthLocalSettings{
				Enabled: new(true),
			},
		},
	}
}

func TestAuthSettings_IsZero(t *testing.T) {
	// GIVEN: AuthSettings in various states.
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

			prefix := fmt.Sprintf("%s\nAuthSettings.IsZero()", packageName)

			// WHEN: IsZero is called.
			got := tc.auth.IsZero()

			// THEN: the result matches expectations.
			if got != tc.want {
				t.Errorf(
					"%s result mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestAuthSettings_CheckValues(t *testing.T) {
	// GIVEN: AuthSettings with session durations of varying validity.
	tests := []struct {
		name     string
		auth     AuthSettings
		errRegex string // Empty = no error wanted.
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

			prefix := fmt.Sprintf("%s\nAuthSettings.CheckValues()", packageName)

			// WHEN: CheckValues is called.
			err := tc.auth.CheckValues()

			// THEN: errors match expectations.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %v\nwant: %q",
					prefix, err, tc.errRegex,
				)
			}
		})
	}
}

func TestSettings_Default_Auth(t *testing.T) {
	// GIVEN: fresh Settings.
	loadMu.Lock()
	defer loadMu.Unlock()
	settings := Settings{}

	prefix := fmt.Sprintf("%s\nSettings.Default() auth", packageName)

	// WHEN: Default is applied.
	if ok := settings.Default(); !ok {
		t.Fatalf("%s Default() failed", prefix)
	}

	// THEN: the auth hard defaults are populated.
	if settings.HardDefaults.Auth.Enabled == nil || *settings.HardDefaults.Auth.Enabled {
		t.Errorf(
			"%s auth should be disabled by default\ngot: %v",
			prefix, settings.HardDefaults.Auth.Enabled,
		)
	}
	if settings.HardDefaults.Auth.Session.Lifetime != "720h" ||
		settings.HardDefaults.Auth.Session.IdleTimeout != "168h" {
		t.Errorf(
			"%s session bounds mismatch\ngot:  %+v",
			prefix, settings.HardDefaults.Auth.Session,
		)
	}
	if settings.HardDefaults.Auth.Local.Enabled == nil || !*settings.HardDefaults.Auth.Local.Enabled {
		t.Errorf(
			"%s local provider should default to enabled\ngot: %v",
			prefix, settings.HardDefaults.Auth.Local.Enabled,
		)
	}
}

func TestSettings_CheckValues_Auth(t *testing.T) {
	// GIVEN: Settings combining auth with basic auth across layers.
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{
				SettingsBase: SettingsBase{Auth: tc.auth},
				HardDefaults: authHardDefaults(),
			}
			settings.Web.BasicAuth = tc.yamlBasic
			settings.FromFlags.Web.BasicAuth = tc.flagBasic
			settings.HardDefaults.Web.BasicAuth = tc.hardBasic

			prefix := fmt.Sprintf("%s\nSettings.CheckValues()", packageName)

			// WHEN: CheckValues is called.
			err := settings.CheckValues()

			// THEN: errors match expectations.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %v\nwant: %q",
					prefix, err, tc.errRegex,
				)
			}
		})
	}
}

func TestSettings_AuthAccessors(t *testing.T) {
	// GIVEN: Settings with and without explicit auth values.
	tests := []struct {
		name            string
		auth            AuthSettings
		wantEnabled     bool
		wantLocal       bool
		wantLifetime    time.Duration
		wantIdleTimeout time.Duration
	}{
		{
			name:            "hard defaults",
			wantEnabled:     false,
			wantLocal:       true,
			wantLifetime:    720 * time.Hour,
			wantIdleTimeout: 168 * time.Hour,
		},
		{
			name: "explicit values",
			auth: AuthSettings{
				Enabled: new(true),
				Session: AuthSessionSettings{
					Lifetime:    "48h",
					IdleTimeout: "2h30m",
				},
				Local: AuthLocalSettings{Enabled: new(false)},
			},
			wantEnabled:     true,
			wantLocal:       false,
			wantLifetime:    48 * time.Hour,
			wantIdleTimeout: 2*time.Hour + 30*time.Minute,
		},
		{
			name: "invalid duration falls back to the hard default",
			auth: AuthSettings{
				Session: AuthSessionSettings{Lifetime: "not-a-duration"},
			},
			wantEnabled:     false,
			wantLocal:       true,
			wantLifetime:    720 * time.Hour,
			wantIdleTimeout: 168 * time.Hour,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{
				SettingsBase: SettingsBase{Auth: tc.auth},
				HardDefaults: authHardDefaults(),
			}

			prefix := fmt.Sprintf("%s\nSettings auth accessors", packageName)

			// WHEN/THEN: each accessor resolves the layered value.
			if got := settings.AuthEnabled(); got != tc.wantEnabled {
				t.Errorf(
					"%s AuthEnabled() mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.wantEnabled,
				)
			}
			if got := settings.AuthLocalEnabled(); got != tc.wantLocal {
				t.Errorf(
					"%s AuthLocalEnabled() mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.wantLocal,
				)
			}
			if got := settings.AuthSessionLifetime(); got != tc.wantLifetime {
				t.Errorf(
					"%s AuthSessionLifetime() mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.wantLifetime,
				)
			}
			if got := settings.AuthSessionIdleTimeout(); got != tc.wantIdleTimeout {
				t.Errorf(
					"%s AuthSessionIdleTimeout() mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.wantIdleTimeout,
				)
			}
		})
	}
}
