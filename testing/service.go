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
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

// ServiceTest will query the service and return the version it finds.
func ServiceTest(
	flag *string,
	cfg *config.Config,
	log *utils.JLog,
) {
	// Only if flag has been provided
	if *flag == "" {
		return
	}
	logFrom := utils.LogFrom{Primary: "Testing", Secondary: *flag}

	log.Info(
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
		log.Fatal(
			fmt.Sprintf(
				"Service %q could not be found in config.service\nDid you mean one of these?\n  - %s",
				*flag, strings.Join(allService, "\n  - "),
			),
			logFrom,
			true,
		)
	}

	// shouldn't need this as the fatal above prevents it getting here if it is nil, but staticcheck gives a SA5011
	if service != nil {
		service.Status = &service_status.Status{}
	}
	_, err := service.LatestVersion.Query()
	if err != nil {
		log.Error(
			fmt.Sprintf(
				"No version matching the conditions specified could be found for %q at %q",
				*flag,
				service.LatestVersion.GetServiceURL(true),
			),
			logFrom,
			true,
		)
	}

	// DeployedVersionLookup
	if service.DeployedVersionLookup != nil {
		version, err := service.DeployedVersionLookup.Query(
			logFrom,
			service.Options.GetSemanticVersioning())
		log.Info(
			fmt.Sprintf(
				"Deployed version - %q",
				version,
			),
			logFrom,
			err == nil,
		)
	}
	if !log.Testing {
		os.Exit(0)
	}
}
