// Copyright [2024] [Argus]
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

//go:build unit || integration

package test

import (
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/test"
)

// NilFlags sets all flags to nil in the given config
func NilFlags(cfg *config.Config) {
	flags := []string{
		"log.level",
		"log.timestamps",
		"data.database-file",
		"web.listen-host",
		"web.listen-port",
		"web.cert-file",
		"web.pkey-file",
		"web.route-prefix",
		"web.basic-auth.username",
		"web.basic-auth.password",
	}
	flagMap := make(map[string]bool, len(flags))
	for i := range flags {
		flagMap[flags[i]] = false
	}
	cfg.Settings.NilUndefinedFlags(&flagMap)
}

// BareConfig returns a minimal config with no flags set
func BareConfig(nilFlags bool) (cfg *config.Config) {
	cfg = &config.Config{
		Settings: config.Settings{
			SettingsBase: config.SettingsBase{
				Web: config.WebSettings{
					RoutePrefix: test.StringPtr(""),
				}}}}

	// NilFlags can be a RACE condition, so use it conditionally
	if nilFlags {
		NilFlags(cfg)
	} else {
		cfg.Settings.FromFlags.Log.Level = nil
		cfg.Settings.FromFlags.Log.Timestamps = nil
		cfg.Settings.FromFlags.Data.DatabaseFile = nil
		cfg.Settings.FromFlags.Web.ListenHost = nil
		cfg.Settings.FromFlags.Web.ListenPort = nil
		cfg.Settings.FromFlags.Web.CertFile = nil
		cfg.Settings.FromFlags.Web.KeyFile = nil
		cfg.Settings.FromFlags.Web.RoutePrefix = nil
		cfg.Settings.FromFlags.Web.BasicAuth = nil
	}

	cfg.HardDefaults.SetDefaults()
	cfg.Settings.SetDefaults()

	return
}
