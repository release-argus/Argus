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
	"github.com/release-argus/Argus/utils"
)

// TestCommands will test the commands given to a Service.
func TestCommands(flag *string, cfg *config.Config) {
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
	service := cfg.Service[*flag]

	if service == nil {
		var allService []string
		for key := range cfg.Service {
			if !utils.Contains(allService, key) {
				allService = append(allService, key)
			}
		}
		jLog.Error(
			fmt.Sprintf(
				"Service %q could not be found in config.service\nDid you mean one of these?\n  - %s",
				*flag, strings.Join(allService, "\n  - "),
			),
			logFrom,
			true,
		)
		os.Exit(1)
	}

	jLog.Fatal(
		fmt.Sprintf(
			"Service %q does not have any `commands` defined",
			*flag),
		logFrom,
		service.Command == nil)

	//nolint:errcheck
	(*service.CommandController).Exec(&logFrom)
	os.Exit(0)
}
