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
	"sort"
	"strings"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// NotifyTest will send a test Shoutrrr message to the Shoutrrr with this flag as its ID.
func NotifyTest(flag *string, cfg *config.Config) {
	// Only if flag provided.
	if *flag == "" {
		return
	}
	logFrom := logutil.LogFrom{Primary: "Testing", Secondary: *flag}

	// Find the Shoutrrr to test.
	notify := findShoutrrr(*flag, cfg, logFrom)

	err := notify.TestSend("https://example.com/service_url")

	if err == nil {
		logutil.Log.Info(
			fmt.Sprintf("Message sent successfully with %q config\n",
				*flag),
			logFrom,
			true)
	} else {
		logutil.Log.Fatal(
			fmt.Sprintf("Message failed to send with %q config\n%s\n",
				*flag, err),
			logFrom,
			true)
	}

	if !logutil.Log.Testing {
		os.Exit(0)
	}
}

// findShoutrrr with `name` from cfg.Service.Notify || cfg.Notify.
func findShoutrrr(
	name string,
	cfg *config.Config,
	logFrom logutil.LogFrom,
) (notify *shoutrrr.Shoutrrr) {
	// Find in Service.X.Notify.name.
	for _, svc := range cfg.Service {
		if svc.Notify != nil && svc.Notify[name] != nil {
			notify = svc.Notify[name]
			break
		}
	}

	// Find in Notify.name.
	if notify == nil {
		if cfg.Notify != nil && cfg.Notify[name] != nil {
			hardDefaults := config.Defaults{}
			hardDefaults.Default()
			emptyShoutrrrs := shoutrrr.Defaults{}
			emptyShoutrrrs.InitMaps()
			main := cfg.Notify[name]
			notify = shoutrrr.New(
				nil,
				name,
				main.Type,
				main.Options, main.URLFields, main.Params,
				&emptyShoutrrrs,
				&emptyShoutrrrs, &emptyShoutrrrs)
			notify.InitMaps()
			notify.Main.InitMaps()

			notifyType := notify.GetType()
			if cfg.Defaults.Notify[notifyType] != nil {
				notify.Defaults = cfg.Defaults.Notify[notifyType]
			}
			notify.Defaults.InitMaps()
			notify.HardDefaults = hardDefaults.Notify[notifyType]
			notify.HardDefaults.InitMaps()

			serviceID := "service_id"
			serviceName := "service_name"
			notify.ServiceStatus = &status.Status{}
			notify.ServiceStatus.Init(
				1, 0, 0,
				serviceID, serviceName, "",
				&dashboard.Options{})
			notify.Failed = &notify.ServiceStatus.Fails.Shoutrrr

			// Check whether all values set.
			if err := notify.CheckValues("    "); err != nil {
				msg := fmt.Sprintf("notify:\n  %s:\n%s\n",
					name, err)
				logutil.Log.Fatal(msg, logFrom, true)
			}

			// Not found.
		} else {
			all := getAllShoutrrrNames(cfg)
			msg := fmt.Sprintf("Notifier %q could not be found in config.notify or in any config.service\nDid you mean one of these?\n  - %s\n",
				name, strings.Join(all, "\n  - "))
			logutil.Log.Fatal(msg, logFrom, true)
		}
	}

	notify.ServiceStatus.ServiceInfo.ID = "TESTING"
	return
}

// getAllShoutrrrNames returns a list of all unique shoutrrr names.
func getAllShoutrrrNames(cfg *config.Config) (all []string) {
	// All global Shoutrrrs.
	if cfg.Notify != nil {
		all = util.SortedKeys(cfg.Notify)
	}

	// All Shoutrrrs in services.
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
