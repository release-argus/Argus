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

		c.Service[serviceID].Notify.Init(
			jLog,
			&serviceID,
			&c.Notify,
			&c.Defaults.Notify,
			&c.HardDefaults.Notify,
		)

		c.Service[serviceID].WebHook.Init(
			jLog,
			&serviceID,
			&c.WebHook,
			&c.Defaults.WebHook,
			&c.HardDefaults.WebHook,
			c.Service[serviceID].Notify,
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

	// Handle deprecations
	c.convertCurrentVersionToDeployedVersion()
	c.convertDeprecatedSlackAndGotify()
	c.convertDeprecatedURLCommands()

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
