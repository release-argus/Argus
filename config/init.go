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

	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
	"gopkg.in/yaml.v3"
)

// Init will hand out the appropriate Defaults.X and HardDefaults.X pointer(s)
func (c *Config) Init() {
	c.HardDefaults.SetDefaults()
	c.Settings.SetDefaults()

	if c.Defaults.Service.DeployedVersionLookup == nil {
		c.Defaults.Service.DeployedVersionLookup = &service.DeployedVersionLookup{}
	}

	jLog.SetTimestamps(*c.Settings.GetLogTimestamps())
	jLog.SetLevel(c.Settings.GetLogLevel())

	for serviceID := range c.Service {
		c.Service[serviceID].Init(
			jLog,
			&c.Defaults.Service,
			&c.HardDefaults.Service,
		)

		c.Service[serviceID].Gotify.Init(
			jLog,
			&serviceID,
			c.Gotify,
			&c.Defaults.Gotify,
			&c.HardDefaults.Gotify,
		)
		c.Service[serviceID].Slack.Init(
			jLog,
			&serviceID,
			c.Service[serviceID].Icon,
			c.Slack,
			&c.Defaults.Slack,
			&c.HardDefaults.Slack,
		)
		c.Service[serviceID].WebHook.Init(
			jLog,
			&serviceID,
			c.WebHook,
			&c.Defaults.WebHook,
			&c.HardDefaults.WebHook,
			c.Service[serviceID].Gotify,
			c.Service[serviceID].Slack,
		)
	}

	// c.Gotify
	if c.Gotify != nil {
		for key := range *c.Gotify {
			(*c.Gotify)[key].Defaults = &c.Defaults.Gotify
			(*c.Gotify)[key].HardDefaults = &c.HardDefaults.Gotify
		}
	}
	// c.Slack
	if c.Slack != nil {
		for key := range *c.Slack {
			(*c.Slack)[key].Defaults = &c.Defaults.Slack
			(*c.Slack)[key].HardDefaults = &c.HardDefaults.Slack
		}
	}
	// c.WebHook
	if c.WebHook != nil {
		for key := range *c.WebHook {
			(*c.WebHook)[key].Defaults = &c.Defaults.WebHook
			(*c.WebHook)[key].HardDefaults = &c.HardDefaults.WebHook
		}
	}
}

// Load `file` as Config.
func (c *Config) Load(file string, flagset *map[string]bool) {
	c.File = file
	logLevel := "WARN"
	jLog = utils.NewJLog(logLevel, false)
	c.Settings.NilUndefinedFlags(flagset)

	//#nosec G304 -- Loading the file asked for by the user
	data, err := os.ReadFile(file)
	msg := fmt.Sprintf("Error reading %q\n%s", file, err)
	jLog.Fatal(msg, utils.LogFrom{}, err != nil)

	err = yaml.Unmarshal(data, c)
	msg = fmt.Sprintf("Unmarshal of %q failed\n%s", file, err)
	jLog.Fatal(msg, utils.LogFrom{}, err != nil)

	c.GetOrder(data)

	saveChannel := make(chan bool, 5)
	c.SaveChannel = &saveChannel

	for key := range c.Service {
		id := key
		c.Service[key].ID = &id
		c.Service[key].SaveChannel = c.SaveChannel
	}
	c.Init()
	c.CheckValues()
}
