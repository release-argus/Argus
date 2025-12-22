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
	"strings"

	"github.com/release-argus/Argus/config"
	logutil "github.com/release-argus/Argus/util/log"
)

// CommandTest will test the commands given to a Service.
func CommandTest(flag *string, cfg *config.Config) bool {
	// Only if flag provided.
	if *flag == "" {
		return true
	}
	logFrom := logutil.LogFrom{Primary: "Testing", Secondary: *flag}

	// Log the test details.
	logutil.Log.Info(
		"",
		logFrom,
		true)
	service, exist := cfg.Service[*flag]

	if !exist {
		logutil.Log.Fatal(
			fmt.Sprintf("Service %q could not be found in config.service\nDid you mean one of these?\n  - %s",
				*flag, strings.Join(cfg.Order, "\n  - "),
			),
			logFrom)
		return false
	}

	if service.CommandController == nil {
		logutil.Log.Fatal(
			fmt.Sprintf("Service %q does not have any `command` defined",
				*flag),
			logFrom)
		return false
	}

	//#nosec G104 -- Disregard.
	//nolint:errcheck // ^
	service.CommandController.Exec(logFrom)

	return true
}
