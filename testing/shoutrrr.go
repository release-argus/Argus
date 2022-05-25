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

package testing

import (
	"fmt"
	"os"
	"strings"

	"github.com/release-argus/Argus/config"
	argus_shoutrrr "github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
)

// TestNotify will send a test Shoutrrr message to the Shoutrrr with this flag as its ID.
func TestNotify(flag *string, cfg *config.Config) {
	// Only if flag has been provided
	if *flag == "" {
		return
	}
	logFrom := utils.LogFrom{Primary: "Testing", Secondary: *flag}

	jLog := utils.NewJLog("DEBUG", false)
	jLog.Info(
		"",
		logFrom,
		true,
	)

	// Find the Shoutrrr to test
	shoutrrr := &argus_shoutrrr.Slice{}
	for _, svc := range cfg.Service {
		if svc.Notify != nil && (*svc.Notify)[*flag] != nil {
			(*shoutrrr)["test"] = (*svc.Notify)[*flag]
			break
		}
	}
	if (*shoutrrr)["test"] == nil {
		if cfg.Notify != nil && cfg.Notify[*flag] != nil {
			hardDefaults := config.Defaults{}
			hardDefaults.SetDefaults()
			emptyShoutrrs := argus_shoutrrr.Shoutrrr{}
			emptyShoutrrs.InitMaps()
			(*shoutrrr)["test"] = &argus_shoutrrr.Shoutrrr{
				ID:           flag,
				Main:         cfg.Notify[*flag],
				Defaults:     cfg.Defaults.Notify["telegram"],
				HardDefaults: hardDefaults.Notify["telegram"],
			}
			(*shoutrrr)["test"].InitMaps()
			(*shoutrrr)["test"].Main.InitMaps()
			if (*shoutrrr)["test"].Defaults == nil {
				(*shoutrrr)["test"].Defaults = &emptyShoutrrs
			} else {
				(*shoutrrr)["test"].Defaults.InitMaps()
			}
			if (*shoutrrr)["test"].HardDefaults == nil {
				(*shoutrrr)["test"].HardDefaults = &emptyShoutrrs
			} else {
				(*shoutrrr)["test"].HardDefaults.InitMaps()
			}
			if err := (*shoutrrr)["test"].CheckValues("    "); err != nil {
				msg := fmt.Sprintf("notify:\n  %s:\n%s\n", *flag, strings.ReplaceAll(err.Error(), "\\", "\n"))
				jLog.Fatal(msg, logFrom, true)
			}
		} else {
			var allShoutrrr []string
			if cfg.Notify != nil {
				for key := range cfg.Notify {
					if !utils.Contains(allShoutrrr, key) {
						allShoutrrr = append(allShoutrrr, key)
					}
				}
			}
			if cfg.Service != nil {
				for _, svc := range cfg.Service {
					if svc.Notify != nil {
						for key := range *svc.Notify {
							if !utils.Contains(allShoutrrr, key) {
								allShoutrrr = append(allShoutrrr, key)
							}
						}
					}
				}
			}
			msg := fmt.Sprintf("Shoutrrr %q could not be found in config.notify or in any config.service\nDid you mean one of these?\n  - %s\n", *flag, strings.Join(allShoutrrr, "\n  - "))
			jLog.Fatal(msg, logFrom, true)
		}
	}

	title := (*shoutrrr)["test"].GetTitle(&utils.ServiceInfo{ID: "Test"})
	message := "TEST - " + (*shoutrrr)["test"].GetMessage(&utils.ServiceInfo{ID: "NAME_OF_SERVICE", URL: "QUERY_URL", WebURL: "WEB_URL", LatestVersion: "MAJOR.MINOR.PATCH"})
	err := shoutrrr.Send(title, message, &utils.ServiceInfo{})
	jLog.Info(fmt.Sprintf("Message sent successfully with %q config\n", *flag), logFrom, err == nil)
	jLog.Error(fmt.Sprintf("Message failed to send with %q config\n%s\n", *flag, utils.ErrorToString(err)), logFrom, err != nil)
	os.Exit(0)
}
