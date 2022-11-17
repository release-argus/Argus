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
	"github.com/release-argus/Argus/util"
)

// CommandTest will test the commands given to a Service.
func CommandTest(
	flag *string,
	cfg *config.Config,
	log *util.JLog,
) {
	// Only if flag has been provided
	if *flag == "" {
		return
	}
	logFrom := util.LogFrom{Primary: "Testing", Secondary: *flag}

	log.Info(
		"",
		logFrom,
		true,
	)
	service, exist := cfg.Service[*flag]

	if !exist {
		log.Fatal(
			fmt.Sprintf(
				"Service %q could not be found in config.service\nDid you mean one of these?\n  - %s",
				*flag, strings.Join(*cfg.Order, "\n  - "),
			),
			logFrom,
			true,
		)
	}

	log.Fatal(
		fmt.Sprintf(
			"Service %q does not have any `command` defined",
			*flag),
		logFrom,
		service.CommandController == nil)

	//nolint:errcheck
	service.CommandController.Exec(&logFrom)
	if !log.Testing {
		os.Exit(0)
	}
}
