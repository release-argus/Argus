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

package testing

import (
	"fmt"
	"os"
	"strings"

	"github.com/hymenaios-io/Hymenaios/config"
	"github.com/hymenaios-io/Hymenaios/utils"
)

// TestService will query the service and return the version it finds.
func TestService(flag *string, cfg *config.Config) {
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

	service.Status.CurrentVersion = nil
	service.Status.LatestVersion = nil
	_, err := service.Query()
	if err != nil {
		helpMsg := ""
		if *service.Type == "url" && strings.Count(*service.URL, "/") == 1 && !strings.HasPrefix(*service.URL, "http") {
			helpMsg = "\nThis URL looks to be a GitHub repo, but the service's type is url, not github. Try using the github service type."
		}
		jLog.Error(
			fmt.Sprintf(
				"No version matching the conditions specified could be found for %q at %q%s",
				*flag,
				service.GetServiceURL(true),
				helpMsg,
			),
			logFrom,
			true,
		)
	}

	// DeployedVersionLookup
	if service.DeployedVersionLookup != nil {
		version, err := service.DeployedVersionLookup.Query(
			logFrom,
			service.GetSemanticVersioning())
		if err != nil {
			jLog.Error(
				err,
				logFrom,
				true,
			)
		} else {
			jLog.Info(
				fmt.Sprintf(
					"Deployed version - %q",
					version,
				),
				logFrom,
				true,
			)
		}
	}
	os.Exit(0)
}
