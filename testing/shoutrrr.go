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
	jLog := utils.NewJLog("INFO", false)
	logFrom := utils.LogFrom{Primary: "Testing", Secondary: *flag}

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
		if cfg.Notify != nil && (*cfg.Notify)[*flag] != nil {
			(*shoutrrr)["test"] = (*cfg.Notify)[*flag]
		} else {
			var allShoutrrr []string
			if cfg.Notify != nil {
				for key := range *cfg.Notify {
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
			fmt.Printf("ERROR: Shoutrrr %q could not be found in config.notify or in any config.service\nDid you mean one of these?\n  - %s\n", *flag, strings.Join(allShoutrrr, "\n  - "))
			os.Exit(1)
		}
	}

	err := shoutrrr.Send(
		(*shoutrrr)["test"].GetTitle(&utils.ServiceInfo{ID: "Test"}),
		"TEST - "+(*shoutrrr)["test"].GetMessage(&utils.ServiceInfo{LatestVersion: "MAJOR.MINOR.PATCH"}),
		&utils.ServiceInfo{})
	if err == nil {
		fmt.Printf("INFO: Message sent successfully with %q config\n", *flag)
	}
	os.Exit(0)
}
