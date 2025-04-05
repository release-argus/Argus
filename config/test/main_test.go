// Copyright [2025] [Argus]
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
	"testing"

	"github.com/release-argus/Argus/config"
)

var packageName = "config.test"

func TestBareConfig(t *testing.T) {
	// GIVEN the config.settings flags.
	strFlags := map[string]struct {
		flag *string
		cfg  *string
	}{
		"log.level":          {flag: config.LogLevel},
		"data.database-file": {flag: config.DataDatabaseFile},
		"web.listen-host":    {flag: config.WebListenHost},
		"web.listen-port":    {flag: config.WebListenPort},
		"web.cert-file":      {flag: config.WebCertFile},
		"web.pkey-file":      {flag: config.WebPKeyFile},
		"web.route-prefix":   {flag: config.WebRoutePrefix},
	}
	for key, value := range strFlags {
		if value.flag == nil {
			t.Errorf("%s\nflag mismatch on %s\nwant: non-nil\ngot:  %v",
				packageName, key, value.flag)
		}
	}
	boolFlags := map[string]struct {
		flag *bool
		cfg  *bool
	}{
		"log.timestamps": {flag: config.LogTimestamps},
	}
	for key, value := range boolFlags {
		if value.flag == nil {
			t.Errorf("%s\nflag mismatch on %s\nwant: non-nil\ngot:  %v",
				packageName, key, value.flag)
		}
	}
	webBasicAuthFlags := map[string]struct {
		flag *string
		cfg  *config.WebSettingsBasicAuth
	}{
		"web.basic-auth.username": {flag: config.WebBasicAuthUsername},
		"web.basic-auth.password": {flag: config.WebBasicAuthPassword},
	}
	for key, value := range webBasicAuthFlags {
		if value.flag == nil {
			t.Errorf("%s\nflag mismatch on %s\nwant: non-nil\ngot:  %v",
				packageName, key, value.flag)
		}
	}

	// WHEN the config is initialised.
	cfg := BareConfig(true)
	strFlags = map[string]struct {
		flag *string
		cfg  *string
	}{
		"log.level":          {cfg: &cfg.Settings.FromFlags.Log.Level},
		"data.database-file": {cfg: &cfg.Settings.FromFlags.Data.DatabaseFile},
		"web.listen-host":    {cfg: &cfg.Settings.FromFlags.Web.ListenHost},
		"web.listen-port":    {cfg: &cfg.Settings.Web.ListenPort},
		"web.cert-file":      {cfg: &cfg.Settings.FromFlags.Web.CertFile},
		"web.pkey-file":      {cfg: &cfg.Settings.FromFlags.Web.KeyFile},
		"web.route-prefix":   {cfg: &cfg.Settings.FromFlags.Web.RoutePrefix},
	}
	boolFlags = map[string]struct {
		flag *bool
		cfg  *bool
	}{
		"log.timestamps": {cfg: cfg.Settings.FromFlags.Log.Timestamps},
	}
	webBasicAuthFlags = map[string]struct {
		flag *string
		cfg  *config.WebSettingsBasicAuth
	}{
		"web.basic-auth.username": {cfg: cfg.Settings.FromFlags.Web.BasicAuth},
		"web.basic-auth.password": {cfg: cfg.Settings.FromFlags.Web.BasicAuth},
	}

	// THEN all flags should be nil.
	for key, value := range strFlags {
		if value.flag != nil {
			t.Errorf("%s\nflag mismatch on %s\nwant: nil\ngot:  %v",
				packageName, key, value.flag)
		}
		if value.cfg == nil {
			t.Errorf("%s\ncfg mismatch on %s\nwant: non-nil\ngot:  %v",
				packageName, key, value.cfg)
		} else if *value.cfg != "" {
			t.Errorf("%s\ncfg mismatch on %s\nwant: empty\ngot:  %q",
				packageName, key, *value.cfg)
		}
	}
	for key, value := range boolFlags {
		if value.flag != nil {
			t.Errorf("%s\nflag mismatch on %s\nwant: nil\ngot:  %v",
				packageName, key, value.flag)
		}
		if value.cfg != nil {
			t.Errorf("%s\ncfg mismatch on %s\nwant: nil\ngot:  %v",
				packageName, key, value.cfg)
		}
	}
	for key, value := range webBasicAuthFlags {
		if value.flag != nil {
			t.Errorf("%s\nflag mismatch on %s\nwant: nil\ngot:  %v",
				packageName, key, value.flag)
		}
		if value.cfg != nil {
			t.Errorf("%s\ncfg mismatch on %s\nwant: nil\ngot:  %v",
				packageName, key, value.cfg)
			t.Error(value.cfg == nil)
		}
	}
}
