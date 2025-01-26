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

/*
Argus monitors GitHub and/or other URLs for version changes.
On a version change, send notification(s) and/or webhook(s).
*/
package main

import (
	"flag"
	"fmt"

	_ "modernc.org/sqlite"

	cfg "github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/db"
	"github.com/release-argus/Argus/testing"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/web"
)

var (
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

	var config cfg.Config
	config.Load(*configFile, &flagset)
	logutil.Log.SetTimestamps(*config.Settings.LogTimestamps())
	logutil.Log.SetLevel(config.Settings.LogLevel())

	// config.check
	config.Print(configCheckFlag)
	// test.*
	testing.CommandTest(testCommandsFlag, &config)
	testing.NotifyTest(testNotifyFlag, &config)
	testing.ServiceTest(testServiceFlag, &config)

	// Count of active services to monitor (if log level INFO or above).
	if logutil.Log.Level > 1 {
		// Count active services.
		serviceCount := len(config.Order)
		for _, key := range config.Order {
			if !config.Service[key].Options.GetActive() {
				serviceCount--
			}
		}

		// Log active count.
		msg := fmt.Sprintf("Found %d services to monitor:", serviceCount)
		logutil.Log.Info(msg, logutil.LogFrom{}, true)

		// Log names of active services.
		for _, key := range config.Order {
			if config.Service[key].Options.GetActive() {
				fmt.Printf("  - %s\n", config.Service[key].Name)
			}
		}
	}

	// Setup DB and last known service versions.
	db.Run(&config)

	// Track all targets for changes in version and act on any found changes.
	go config.Service.Track(&config.Order, &config.OrderMutex)

	// Web server.
	web.Run(&config)
}
