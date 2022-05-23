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
	Slack "github.com/release-argus/Argus/notifiers/slack"
	"github.com/release-argus/Argus/utils"
)

// TestSlack will send a test Slack message to the Slack with this flag as its ID.
func TestSlack(flag *string, cfg *config.Config) {
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

	// Find the Slack to test
	slack := &Slack.Slice{}
	for _, svc := range cfg.Service {
		if svc.Slack != nil && (*svc.Slack)[*flag] != nil {
			(*slack)["test"] = (*svc.Slack)[*flag]
			break
		}
	}
	if (*slack)["test"] == nil {
		if (*cfg.Slack)[*flag] != nil {
			(*slack)["test"] = (*cfg.Slack)[*flag]
			(*slack)["test"].ID = flag
			(*slack)["test"].Defaults = &cfg.Defaults.Slack
			(*slack)["test"].HardDefaults = &cfg.Defaults.Slack
			(*slack)["test"].Main = &Slack.Slack{}
		} else {
			var allSlack []string
			for key := range *cfg.Slack {
				if !utils.Contains(allSlack, key) {
					allSlack = append(allSlack, key)
				}
			}
			for _, svc := range cfg.Service {
				if svc.Slack != nil {
					for key := range *svc.Slack {
						if !utils.Contains(allSlack, key) {
							allSlack = append(allSlack, key)
						}
					}
				}
			}
			fmt.Printf("ERROR: Slack %q could not be found in config.slack or in any config.service\nDid you mean one of these?\n  - %s\n", *flag, strings.Join(allSlack, "\n  - "))
			os.Exit(1)
		}
	}

	//#nosec G104 -- Errors will be logged to CL
	//nolint:errcheck // ^
	err := slack.Send("Test message", &utils.ServiceInfo{ID: ""})
	if err == nil {
		fmt.Printf("INFO: Message sent successfully with %q config\n", *flag)
	}
	os.Exit(0)
}
