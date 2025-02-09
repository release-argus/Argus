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

// Package testing provides utilities for CLI-based testing.
package testing

import (
	"fmt"
	"os"
	"strings"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// ServiceTest queries the service and returns the found version.
func ServiceTest(flag *string, cfg *config.Config) {
	// Only if flag provided.
	if *flag == "" {
		return
	}

	// Log the test details.
	logFrom := logutil.LogFrom{Primary: "Testing", Secondary: *flag}
	logutil.Log.Info(
		"",
		logFrom,
		true,
	)

	// Check service exists.
	if !util.Contains(cfg.Order, *flag) {
		logutil.Log.Fatal(
			fmt.Sprintf(
				"Service %q could not be found in config.service\nDid you mean one of these?\n  - %s",
				*flag, strings.Join(cfg.Order, "\n  - "),
			),
			logFrom,
			true,
		)
	}

	// Service to test.
	service := cfg.Service[*flag]

	// LatestVersion.
	_, err := service.LatestVersion.Query(false, logFrom)
	if err != nil {
		logutil.Log.Error(
			fmt.Sprintf(
				"No version matching the conditions specified could be found for %q at %q",
				*flag,
				service.LatestVersion.ServiceURL(true),
			),
			logFrom,
			true,
		)
	}

	// DeployedVersionLookup.
	if service.DeployedVersionLookup != nil {
		if err := service.DeployedVersionLookup.Query(false, logFrom); err == nil {
			logutil.Log.Info(
				fmt.Sprintf(
					"Deployed version - %q",
					service.Status.DeployedVersion()),
				logFrom,
				true,
			)
		}
	}

	if !logutil.Log.Testing {
		os.Exit(0)
	}
}
