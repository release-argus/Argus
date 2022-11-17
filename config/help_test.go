// Copyright [2022] [Argus]
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

//go:build testing

package config

import (
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/util"
)

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}

func testConfig() Config {
	logLevel := "DEBUG"
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	return Config{
		File:            "/root/inaccessible",
		DatabaseChannel: &databaseChannel,
		SaveChannel:     &saveChannel,
		Settings: Settings{
			Indentation: 4,
			Log: LogSettings{
				Level: &logLevel,
			},
		},
	}
}

func testSettings() Settings {
	logTimestamps := true
	logLevel := "DEBUG"
	dataDatabaseFile := "somewhere.db"
	webListenHost := "test"
	webListenPort := "123"
	webRoutePrefix := "/something"
	webCertFile := "../test/ordering_0.yml"
	webKeyFile := "../test/ordering_1.yml"
	return Settings{
		Log: LogSettings{
			Timestamps: &logTimestamps,
			Level:      &logLevel,
		},
		Data: DataSettings{
			DatabaseFile: &dataDatabaseFile,
		},
		Web: WebSettings{
			ListenHost:  &webListenHost,
			ListenPort:  &webListenPort,
			RoutePrefix: &webRoutePrefix,
			CertFile:    &webCertFile,
			KeyFile:     &webKeyFile,
		},
	}
}

func testLoad(fileOverride string) Config {
	var (
		config     Config
		configFile string = "../test/config_test.yml"
	)
	if fileOverride != "" {
		configFile = fileOverride
	}

	flags := make(map[string]bool)
	config.Load(configFile, &flags, &util.JLog{})

	return config
}
