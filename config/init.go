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

package config

import (
	"fmt"
	"os"

	command "github.com/release-argus/Argus/commands"
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/deployed_version"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"gopkg.in/yaml.v3"
)

// Init will hand out the appropriate Defaults.X and HardDefaults.X pointer(s)
func (c *Config) Init() {
	c.HardDefaults.SetDefaults()
	c.Settings.SetDefaults()

	if c.Defaults.Service.DeployedVersionLookup == nil {
		c.Defaults.Service.DeployedVersionLookup = &deployed_version.Lookup{}
	}

	jLog.SetTimestamps(*c.Settings.GetLogTimestamps())
	jLog.SetLevel(c.Settings.GetLogLevel())

	for serviceID, service := range c.Service {
		service.Status.Init(len(
			service.Notify),
			len(service.Command),
			len(service.WebHook),
			&service.ID,
			&service.Dashboard.WebURL)

		service.Init(
			jLog,
			&c.Defaults.Service,
			&c.HardDefaults.Service,
		)

		service.Notify.Init(
			jLog,
			&serviceID,
			&service.Status,
			&c.Notify,
			&c.Defaults.Notify,
			&c.HardDefaults.Notify,
		)

		service.WebHook.Init(
			jLog,
			&serviceID,
			&service.Status,
			&c.WebHook,
			&c.Defaults.WebHook,
			&c.HardDefaults.WebHook,
			&service.Notify,
			service.Options.GetIntervalPointer(),
		)

		if service.Command != nil {
			service.CommandController = &command.Controller{}
		}
		service.CommandController.Init(
			jLog,
			&serviceID,
			&service.Status,
			&service.Command,
			&service.Notify,
			service.Options.GetIntervalPointer(),
		)
	}

	// c.Notify
	if c.Notify != nil {
		for key := range c.Notify {
			// DefaultIfNil to handle testing. CheckValues will pick up on this nil
			c.Notify[key].Defaults = c.Defaults.Notify[c.Notify[key].Type]
			c.Notify[key].HardDefaults = c.HardDefaults.Notify[c.Notify[key].Type]
		}
	}
	// c.WebHook
	if c.WebHook != nil {
		for key := range c.WebHook {
			c.WebHook[key].Defaults = &c.Defaults.WebHook
			c.WebHook[key].HardDefaults = &c.HardDefaults.WebHook
		}
	}
}

// Load `file` as Config.
func (c *Config) Load(file string, flagset *map[string]bool, log *utils.JLog) {
	c.File = file
	jLog = log
	c.Settings.NilUndefinedFlags(flagset)

	//#nosec G304 -- Loading the file asked for by the user
	data, err := os.ReadFile(file)
	msg := fmt.Sprintf("Error reading %q\n%s", file, err)
	jLog.Fatal(msg, utils.LogFrom{}, err != nil)

	err = yaml.Unmarshal(data, c)
	msg = fmt.Sprintf("Unmarshal of %q failed\n%s", file, err)
	jLog.Fatal(msg, utils.LogFrom{}, err != nil)

	c.GetOrder(data)

	databaseChannel := make(chan db_types.Message, 8)
	c.DatabaseChannel = &databaseChannel

	saveChannel := make(chan bool, 4)
	c.SaveChannel = &saveChannel

	for key := range c.Service {
		c.Service[key].ID = key
		c.Service[key].Status = service_status.Status{
			DatabaseChannel: c.DatabaseChannel,
			SaveChannel:     c.SaveChannel,
		}
	}
	c.Init()
	c.CheckValues()
}
