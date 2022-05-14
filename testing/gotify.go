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
	Gotify "github.com/release-argus/Argus/notifiers/gotify"
	"github.com/release-argus/Argus/utils"
)

// TestGotify will send a test Gotify message to the Gotify with this flag as its ID.
func TestGotify(flag *string, cfg *config.Config) {
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

	// Find the Gotify to test
	gotify := &Gotify.Slice{}
	for _, svc := range cfg.Service {
		if svc.Gotify != nil && (*svc.Gotify)[*flag] != nil {
			(*gotify)["test"] = (*svc.Gotify)[*flag]
			break
		}
	}
	if (*gotify)["test"] == nil {
		if (*cfg.Gotify)[*flag] != nil {
			(*gotify)["test"] = (*cfg.Gotify)[*flag]
			(*gotify)["test"].ID = flag
			(*gotify)["test"].Defaults = &cfg.Defaults.Gotify
			(*gotify)["test"].Main = &Gotify.Gotify{}
		} else {
			var allGotify []string
			for key := range *cfg.Gotify {
				if !utils.Contains(allGotify, key) {
					allGotify = append(allGotify, key)
				}
			}
			for _, svc := range cfg.Service {
				if svc.Gotify != nil {
					for key := range *svc.Gotify {
						if !utils.Contains(allGotify, key) {
							allGotify = append(allGotify, key)
						}
					}
				}
			}
			fmt.Printf("ERROR: Gotify %q could not be found in config.gotify or in any config.service\nDid you mean one of these?\n  - %s\n", *flag, strings.Join(allGotify, "\n  - "))
			os.Exit(1)
		}
	}

	err := gotify.Send((*gotify)["test"].GetTitle(&utils.ServiceInfo{ID: ""}), "Test message", &utils.ServiceInfo{ID: ""})
	if err == nil {
		fmt.Printf("INFO: Message sent successfully with %q config\n", *flag)
	}
	os.Exit(0)
}
