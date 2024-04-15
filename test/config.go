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

import "github.com/release-argus/Argus/config"

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

func BareConfig() (cfg *config.Config) {
	cfg = &config.Config{
		Settings: config.Settings{
			SettingsBase: config.SettingsBase{
				Web: config.WebSettings{
					RoutePrefix: StringPtr(""),
				}}}}
	NilFlags(cfg)
	cfg.HardDefaults.SetDefaults()
	cfg.Settings.SetDefaults()
	return
}
