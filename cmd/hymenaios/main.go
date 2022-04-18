// Copyright [2022] [Hymenaios]
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
Hymenaios monitors GitHub and/or other URLs for version changes.
On a version change, send Slack message(s) and/or webhook(s).
main.go uses track.go for the goroutines that call query.go
and then, on a version change, will call slack.go and webhook.go.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	hymenaios_testing "github.com/hymenaios-io/Hymenaios/testing"

	cfg "github.com/hymenaios-io/Hymenaios/config"
	"github.com/hymenaios-io/Hymenaios/utils"
	"github.com/hymenaios-io/Hymenaios/web"
)

var (
	jLog utils.JLog
)

// main loads the config and then calls service.Track to monitor
// each Service of the config for version changes and acts on
// them as defined. It also sets up the Web UI and SaveHandler.
func main() {
	var (
		config          cfg.Config
		configFile      = flag.String("config.file", "config.yml", "Hymenaios configuration file path.")
		configCheckFlag = flag.Bool("config.check", false, "Print the fully-parsed config.")
		testGotifyFlag  = flag.String("test.gotify", "", "Put the name of the Gotify service to send a test message.")
		testSlackFlag   = flag.String("test.slack", "", "Put the name of the Slack service to send a test message.")
		testServiceFlag = flag.String("test.service", "", "Put the name of the Service to test the version query.")
	)

	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	config.Load(*configFile, &flagset)
	jLog.SetTimestamps(*config.Settings.GetLogTimestamps())
	jLog.SetLevel(config.Settings.GetLogLevel())

	// config.check
	config.Print(configCheckFlag)
	// test.*
	hymenaios_testing.TestGotify(testGotifyFlag, &config)
	hymenaios_testing.TestSlack(testSlackFlag, &config)
	hymenaios_testing.TestService(testServiceFlag, &config)

	// config.Service.Init()
	serviceCount := len(config.Service)
	if serviceCount == 0 {
		jLog.Warn("No services to monitor were found.", utils.LogFrom{}, true)
		os.Exit(0)
	}

	// INFO or above
	if jLog.Level > 1 {
		msg := fmt.Sprintf("Found %d services to monitor:", serviceCount)
		jLog.Info(msg, utils.LogFrom{}, true)

		for _, key := range config.Order {
			fmt.Printf("  - %s\n", *config.Service[key].ID)
		}
	}

	// Track all targets for changes in version and act on any found changes.
	go (&config).Service.Track(&config.Order)

	// SaveHandler that listens for calls to save config changes.
	go (&config).SaveHandler(configFile)

	// Web server
	web.Main(&config, &jLog)
}
