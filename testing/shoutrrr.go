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
	"sort"
	"strings"

	"github.com/release-argus/Argus/config"
	shoutrrr "github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// NotifyTest will send a test Shoutrrr message to the Shoutrrr with this flag as its ID.
func NotifyTest(
	flag *string,
	cfg *config.Config,
	log *util.JLog,
) {
	// Only if flag has been provided
	if *flag == "" {
		return
	}
	shoutrrr.LogInit(log)
	logFrom := util.LogFrom{Primary: "Testing", Secondary: *flag}

	// Find the Shoutrrr to test
	slice := findShoutrrr(*flag, cfg, log, &logFrom)

	title := slice["test"].Title(&util.ServiceInfo{ID: "Test"})
	message := "TEST - " + slice["test"].Message(
		&util.ServiceInfo{
			ID:            "NAME_OF_SERVICE",
			URL:           "QUERY_URL",
			WebURL:        "WEB_URL",
			LatestVersion: "MAJOR.MINOR.PATCH"})
	err := slice.Send(
		title,
		message,
		&util.ServiceInfo{
			ID:            "ID",
			URL:           "URL",
			WebURL:        "WebURL",
			LatestVersion: "MAJOR.MINOR.PATCH"},
		false)

	log.Info(fmt.Sprintf("Message sent successfully with %q config\n", *flag), logFrom, err == nil)
	log.Fatal(fmt.Sprintf("Message failed to send with %q config\n%s\n", *flag, util.ErrorToString(err)), logFrom, err != nil)

	if !log.Testing {
		os.Exit(0)
	}
}

// findShoutrrr with `name` from cfg.Service.Notify || cfg.Notify
func findShoutrrr(
	name string,
	cfg *config.Config,
	log *util.JLog,
	logFrom *util.LogFrom,
) shoutrrr.Slice {
	slice := make(shoutrrr.Slice, 1)
	// Find in Service.X.Notify.name
	for _, svc := range cfg.Service {
		if svc.Notify != nil && svc.Notify[name] != nil {
			slice["test"] = svc.Notify[name]
			break
		}
	}

	// Find in Notify.name
	if slice["test"] == nil {
		if cfg.Notify != nil && cfg.Notify[name] != nil {
			hardDefaults := config.Defaults{}
			hardDefaults.SetDefaults()
			emptyShoutrrs := shoutrrr.ShoutrrrDefaults{}
			emptyShoutrrs.InitMaps()
			main := cfg.Notify[name]
			slice["test"] = shoutrrr.New(
				nil, // failed
				name,
				&main.Options,
				&main.Params,
				main.Type,
				&main.URLFields,
				&emptyShoutrrs,
				&emptyShoutrrs,
				&emptyShoutrrs)
			slice["test"].InitMaps()
			slice["test"].Main.InitMaps()

			notifyType := slice["test"].GetType()
			if cfg.Defaults.Notify[notifyType] != nil {
				slice["test"].Defaults = cfg.Defaults.Notify[notifyType]
			}
			slice["test"].Defaults.InitMaps()
			slice["test"].HardDefaults = hardDefaults.Notify[notifyType]
			slice["test"].HardDefaults.InitMaps()

			serviceID := ""
			slice["test"].ServiceStatus = &svcstatus.Status{ServiceID: &serviceID}
			slice["test"].ServiceStatus.Init(
				1, 0, 0,
				slice["test"].ServiceStatus.ServiceID,
				slice["test"].ServiceStatus.WebURL)
			slice["test"].Failed = &slice["test"].ServiceStatus.Fails.Shoutrrr

			// Check if all values are set
			if err := slice["test"].CheckValues("    "); err != nil {
				msg := fmt.Sprintf("notify:\n  %s:\n%s\n", name, strings.ReplaceAll(err.Error(), "\\", "\n"))
				log.Fatal(msg, *logFrom, true)
			}

			// Not found
		} else {
			all := getAllShoutrrrNames(cfg)
			msg := fmt.Sprintf("Notifier %q could not be found in config.notify or in any config.service\nDid you mean one of these?\n  - %s\n",
				name, strings.Join(all, "\n  - "))
			log.Fatal(msg, *logFrom, true)
		}
	}
	serviceID := "TESTING"
	slice["test"].ServiceStatus = &svcstatus.Status{ServiceID: &serviceID}
	return slice
}

// getAllShoutrrrNames will return a list of all unique shoutrrr names
func getAllShoutrrrNames(cfg *config.Config) (all []string) {
	// All global Shoutrrrs
	if cfg.Notify != nil {
		all = util.SortedKeys(cfg.Notify)
	}

	// All Shoutrrrs in services
	if cfg.Service != nil {
		for _, svc := range cfg.Service {
			for key := range svc.Notify {
				if !util.Contains(all, key) {
					all = append(all, key)
				}
			}
		}
	}
	sort.Strings(all)
	return
}
