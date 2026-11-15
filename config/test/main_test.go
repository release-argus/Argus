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
	"testing"

	"github.com/release-argus/Argus/config"
)

func TestBareConfig(t *testing.T) {
	strFlags := []struct {
		name string
		flag *string
	}{
		{name: "log.level", flag: config.LogLevel},
		{name: "data.database-file", flag: config.DataDatabaseFile},
		{name: "web.listen-host", flag: config.WebListenHost},
		{name: "web.listen-port", flag: config.WebListenPort},
		{name: "web.cert-file", flag: config.WebCertFile},
		{name: "web.pkey-file", flag: config.WebPKeyFile},
		{name: "web.route-prefix", flag: config.WebRoutePrefix},
		{name: "web.basic-auth.username", flag: config.WebBasicAuthUsername},
		{name: "web.basic-auth.password", flag: config.WebBasicAuthPassword},
	}
	for _, f := range strFlags {
		if f.flag == nil {
			t.Errorf("before: flag %q: value mismatch\ngot:  nil\nwant: non-nil", f.name)
		}
	}
	boolFlags := []struct {
		name string
		flag *bool
	}{
		{name: "data.readonly", flag: config.DataReadonly},
		{name: "log.timestamps", flag: config.LogTimestamps},
	}
	for _, f := range boolFlags {
		if f.flag == nil {
			t.Errorf("before: flag %q: value mismatch\ngot:  nil\nwant: non-nil", f.name)
		}
	}

	cfg := BareConfig(true)

	strAfter := []struct {
		name string
		flag *string
		cfg  string
	}{
		{name: "log.level", flag: config.LogLevel, cfg: cfg.Settings.FromFlags.Log.Level},
		{name: "data.database-file", flag: config.DataDatabaseFile, cfg: cfg.Settings.FromFlags.Data.DatabaseFile},
		{name: "web.listen-host", flag: config.WebListenHost, cfg: cfg.Settings.FromFlags.Web.ListenHost},
		{name: "web.listen-port", flag: config.WebListenPort, cfg: cfg.Settings.FromFlags.Web.ListenPort},
		{name: "web.cert-file", flag: config.WebCertFile, cfg: cfg.Settings.FromFlags.Web.CertFile},
		{name: "web.pkey-file", flag: config.WebPKeyFile, cfg: cfg.Settings.FromFlags.Web.KeyFile},
		{name: "web.route-prefix", flag: config.WebRoutePrefix, cfg: cfg.Settings.FromFlags.Web.RoutePrefix},
	}
	for _, f := range strAfter {
		if f.flag != nil {
			t.Errorf(
				"after BareConfig: flag %q: value mismatch\ngot:  %q\nwant: nil",
				f.name, *f.flag,
			)
		}
		if f.cfg != "" {
			t.Errorf(
				"after BareConfig: flag %q: Settings.FromFlags mismatch\ngot:  %q\nwant: \"\"",
				f.name, f.cfg,
			)
		}
	}
	boolAfter := []struct {
		name string
		flag *bool
		cfg  *bool
	}{
		{name: "data.readonly", flag: config.DataReadonly, cfg: cfg.Settings.FromFlags.Data.Readonly},
		{name: "log.timestamps", flag: config.LogTimestamps, cfg: cfg.Settings.FromFlags.Log.Timestamps},
	}
	for _, f := range boolAfter {
		if f.flag != nil {
			t.Errorf(
				"after BareConfig: flag %q: value mismatch\ngot:  %t\nwant: nil",
				f.name, *f.flag,
			)
		}
		if f.cfg != nil {
			t.Errorf(
				"after BareConfig: flag %q: Settings.FromFlags mismatch\ngot:  %t\nwant: nil",
				f.name, *f.cfg,
			)
		}
	}
	for _, f := range []struct {
		name string
		flag *string
	}{
		{name: "web.basic-auth.username", flag: config.WebBasicAuthUsername},
		{name: "web.basic-auth.password", flag: config.WebBasicAuthPassword},
	} {
		if f.flag != nil {
			t.Errorf(
				"after BareConfig: flag %q: value mismatch\ngot:  %q\nwant: nil",
				f.name, *f.flag,
			)
		}
	}
	if cfg.Settings.FromFlags.Web.BasicAuth != nil {
		t.Error(
			"after BareConfig: Settings.FromFlags.Web.BasicAuth mismatch\n" +
				"got:  non-nil\nwant: nil",
		)
	}
}
