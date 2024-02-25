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
	notify := findShoutrrr(*flag, cfg, log, &logFrom)

	err := notify.TestSend()

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
) (notify *shoutrrr.Shoutrrr) {
	// Find in Service.X.Notify.name
	for _, svc := range cfg.Service {
		if svc.Notify != nil && svc.Notify[name] != nil {
			notify = svc.Notify[name]
			break
		}
	}

	// Find in Notify.name
	if notify == nil {
		if cfg.Notify != nil && cfg.Notify[name] != nil {
			hardDefaults := config.Defaults{}
			hardDefaults.SetDefaults()
			emptyShoutrrrs := shoutrrr.ShoutrrrDefaults{}
			emptyShoutrrrs.InitMaps()
			main := cfg.Notify[name]
			notify = shoutrrr.New(
				nil, // failed
				name,
				&main.Options,
				&main.Params,
				main.Type,
				&main.URLFields,
				&emptyShoutrrrs,
				&emptyShoutrrrs,
				&emptyShoutrrrs)
			notify.InitMaps()
			notify.Main.InitMaps()

			notifyType := notify.GetType()
			if cfg.Defaults.Notify[notifyType] != nil {
				notify.Defaults = cfg.Defaults.Notify[notifyType]
			}
			notify.Defaults.InitMaps()
			notify.HardDefaults = hardDefaults.Notify[notifyType]
			notify.HardDefaults.InitMaps()

			serviceID := ""
			notify.ServiceStatus = &svcstatus.Status{ServiceID: &serviceID}
			notify.ServiceStatus.Init(
				1, 0, 0,
				notify.ServiceStatus.ServiceID,
				notify.ServiceStatus.WebURL)
			notify.Failed = &notify.ServiceStatus.Fails.Shoutrrr

			// Check if all values are set
			if err := notify.CheckValues("    "); err != nil {
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
	notify.ServiceStatus = &svcstatus.Status{ServiceID: &serviceID}
	return
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
