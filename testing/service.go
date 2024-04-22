// Copyright [2023] [Argus]
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
	"github.com/release-argus/Argus/util"
)

// ServiceTest will query the service and return the version it finds.
func ServiceTest(
	flag *string,
	cfg *config.Config,
	log *util.JLog,
) {
	// Only if flag has been provided
	if *flag == "" {
		return
	}

	// Log what we are testing
	logFrom := &util.LogFrom{Primary: "Testing", Secondary: *flag}
	log.Info(
		"",
		logFrom,
		true,
	)

	// Check if service exists
	if !util.Contains(cfg.Order, *flag) {
		log.Fatal(
			fmt.Sprintf(
				"Service %q could not be found in config.service\nDid you mean one of these?\n  - %s",
				*flag, strings.Join(cfg.Order, "\n  - "),
			),
			logFrom,
			true,
		)
	}

	// Service we are testing
	service := cfg.Service[*flag]

	// LatestVersion
	_, err := service.LatestVersion.Query(false, logFrom)
	if err != nil {
		helpMsg := ""
		if service.LatestVersion.Type == "url" && strings.Count(service.LatestVersion.URL, "/") == 1 && !strings.HasPrefix(service.LatestVersion.URL, "http") {
			helpMsg = "This URL looks to be a GitHub repo, but the service's type is url, not github. Try using the github service type.\n"
		}
		log.Error(
			fmt.Sprintf(
				"No version matching the conditions specified could be found for %q at %q\n%s",
				*flag,
				service.LatestVersion.ServiceURL(true),
				helpMsg,
			),
			logFrom,
			true,
		)
	}

	// DeployedVersionLookup
	if service.DeployedVersionLookup != nil {
		version, err := service.DeployedVersionLookup.Query(false, logFrom)
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
