// Copyright [2024] [Argus]
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

// Package config provides the configuration for Argus.
package config

import (
	"fmt"
	"os"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

// LogInit for Argus.
func LogInit(log *util.JLog) {
	if jLog != nil {
		return
	}

	jLog = log
	service.LogInit(jLog)
}

// Init will hand out the appropriate Defaults.X and HardDefaults.X pointer(s).
func (c *Config) Init(setLog bool) {
	c.OrderMutex.RLock()
	defer c.OrderMutex.RUnlock()

	c.HardDefaults.Default()
	// Give the HardDefaults to the Defaults.
	c.Defaults.Service.LatestVersion.Require.Docker.SetDefaults(
		&c.HardDefaults.Service.LatestVersion.Require.Docker)

	c.Settings.Default()

	// Options.
	c.HardDefaults.Service.LatestVersion.Options = &c.HardDefaults.Service.Options
	c.HardDefaults.Service.DeployedVersionLookup.Options = &c.HardDefaults.Service.Options
	c.Defaults.Service.LatestVersion.Options = &c.Defaults.Service.Options
	c.Defaults.Service.DeployedVersionLookup.Options = &c.Defaults.Service.Options

	c.HardDefaults.Service.Status.SaveChannel = c.SaveChannel

	if setLog {
		jLog.SetTimestamps(*c.Settings.LogTimestamps())
		jLog.SetLevel(c.Settings.LogLevel())
	}

	for i, name := range c.Order {
		jLog.Debug(
			fmt.Sprintf("%d/%d %s Init",
				i+1, len(c.Service), name),
			util.LogFrom{}, true)
		c.Service[name].Init(
			&c.Defaults.Service, &c.HardDefaults.Service,
			&c.Notify, &c.Defaults.Notify, &c.HardDefaults.Notify,
			&c.WebHook, &c.Defaults.WebHook, &c.HardDefaults.WebHook)
	}
}

// Load `file` as Config.
func (c *Config) Load(file string, flagset *map[string]bool, log *util.JLog) {
	c.File = file
	// Give the log to the other packages.
	if log != nil {
		LogInit(log)
	}
	c.Settings.NilUndefinedFlags(flagset)

	//#nosec G304 -- Loading the file asked for by the user.
	data, err := os.ReadFile(file)
	jLog.Fatal(
		fmt.Sprintf("Error reading %q\n%s",
			file, err),
		util.LogFrom{}, err != nil)

	err = yaml.Unmarshal(data, c)
	jLog.Fatal(
		fmt.Sprintf("Unmarshal of %q failed\n%s",
			file, err),
		util.LogFrom{}, err != nil)

	c.GetOrder(data)

	databaseChannel := make(chan dbtype.Message, 32)
	c.DatabaseChannel = &databaseChannel

	saveChannel := make(chan bool, 32)
	c.SaveChannel = &saveChannel

	var toDelete []string
	for key, svc := range c.Service {
		// Remove the service if nil.
		if svc == nil {
			toDelete = append(toDelete, key)
			continue
		}

		svc.ID = key
		svc.Status = *status.New(
			nil, c.DatabaseChannel, c.SaveChannel,
			"", "", "", "", "", "")
	}
	c.HardDefaults.Service.Status.DatabaseChannel = c.DatabaseChannel
	c.HardDefaults.Service.Status.SaveChannel = c.SaveChannel

	// Create the service map if it doesn't exist.
	if c.Service == nil {
		c.Service = make(map[string]*service.Service)
	} else if len(toDelete) > 0 {
		for _, key := range toDelete {
			c.Order = util.RemoveElement(c.Order, key)
			delete(c.Service, key)
		}
	}

	// Default Empty List ETag as it depends on default access_token.
	accessTokenDefault := util.FirstNonDefaultWithEnv(
		c.Defaults.Service.LatestVersion.AccessToken,
		c.HardDefaults.Service.LatestVersion.AccessToken)
	github.SetEmptyListETag(accessTokenDefault)

	// SaveHandler that listens for calls to save config changes.
	go c.SaveHandler()

	c.Init(log != nil)
	c.CheckValues()
}
