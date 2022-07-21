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

/*
Argus monitors GitHub and/or other URLs for version changes.
On a version change, send notification(s) and/or webhook(s).
*/
package main

import (
	"flag"
	"fmt"

	"github.com/release-argus/Argus/db"
	argus_testing "github.com/release-argus/Argus/testing"

	cfg "github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web"
)

var (
	jLog             utils.JLog
	config           cfg.Config
	configFile       = flag.String("config.file", "config.yml", "Argus configuration file path.")
	configCheckFlag  = flag.Bool("config.check", false, "Print the fully-parsed config.")
	testCommandsFlag = flag.String("test.commands", "", "Put the name of the Service to test the `commands` of.")
	testNotifyFlag   = flag.String("test.notify", "", "Put the name of the Notify service to send a test message.")
	testServiceFlag  = flag.String("test.service", "", "Put the name of the Service to test the version query.")
)

// main loads the config and then calls service.Track to monitor
// each Service of the config for version changes and acts on
// them as defined. It also sets up the Web UI and SaveHandler.
func main() {
	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	config.Load(*configFile, &flagset, &jLog)
	jLog.SetTimestamps(*config.Settings.GetLogTimestamps())
	jLog.SetLevel(config.Settings.GetLogLevel())

	// config.check
	config.Print(configCheckFlag)
	// test.*
	argus_testing.CommandTest(testCommandsFlag, &config, &jLog)
	argus_testing.NotifyTest(testNotifyFlag, &config, &jLog)
	argus_testing.ServiceTest(testServiceFlag, &config, &jLog)

	// config.Service.Init()
	serviceCount := len(*config.Order)
	jLog.Fatal("No services to monitor were found.", utils.LogFrom{}, serviceCount == 0)

	// INFO or above
	if jLog.Level > 1 {
		msg := fmt.Sprintf("Found %d services to monitor:", serviceCount)
		jLog.Info(msg, utils.LogFrom{}, true)

		for _, key := range *config.Order {
			fmt.Printf("  - %s\n", config.Service[key].ID)
		}
	}

	go db.Run(&config, &jLog)

	// Track all targets for changes in version and act on any found changes.
	go (&config).Service.Track(config.Order)

	// SaveHandler that listens for calls to save config changes.
	go (&config).SaveHandler()

	// Web server
	web.Run(&config, &jLog)
}
