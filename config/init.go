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

// Package config provides the configuration for Argus.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Init will hand out the appropriate Defaults.X and HardDefaults.X pointers.
func (c *Config) Init() {
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

	logutil.Log.SetTimestamps(*c.Settings.LogTimestamps())
	logutil.Log.SetLevel(c.Settings.LogLevel())

	for i, name := range c.Order {
		service := c.Service[name]
		logutil.Log.Debug(
			fmt.Sprintf("%d/%d %s Init",
				i+1, len(c.Service), service.Name),
			logutil.LogFrom{}, true)
		service.Init(
			&c.Defaults.Service, &c.HardDefaults.Service,
			&c.Notify, &c.Defaults.Notify, &c.HardDefaults.Notify,
			&c.WebHook, &c.Defaults.WebHook, &c.HardDefaults.WebHook)
	}
}

// Load `file` as Config.
func (c *Config) Load(file string, flagset *map[string]bool) {
	// Initialise the Log if it hasn't been already.
	logutil.Init("ERROR", false)

	// Load the .env file (if it exists).
	envFile := filepath.Join(filepath.Dir(file), ".env")
	err := loadEnvFile(envFile)
	logutil.Log.Warn(
		err,
		logutil.LogFrom{}, err != nil)

	c.File = file
	c.Settings.NilUndefinedFlags(flagset)

	//#nosec G304 -- Loading the file asked for by the user.
	data, err := os.ReadFile(file)
	logutil.Log.Fatal(
		fmt.Sprintf("Error reading %q\n%s",
			file, err),
		logutil.LogFrom{}, err != nil)

	err = yaml.Unmarshal(data, c)
	logutil.Log.Fatal(
		fmt.Sprintf("Unmarshal of %q failed\n%s",
			file, err),
		logutil.LogFrom{}, err != nil)

	c.GetOrder(data)

	databaseChannel := make(chan dbtype.Message, 32)
	c.DatabaseChannel = &databaseChannel

	saveChannel := make(chan bool, 32)
	c.SaveChannel = &saveChannel

	for _, svc := range c.Service {
		svc.Status = *status.New(
			nil, c.DatabaseChannel, c.SaveChannel,
			"",
			"", "",
			"", "",
			"",
			nil)
	}
	c.HardDefaults.Service.Status.DatabaseChannel = c.DatabaseChannel
	c.HardDefaults.Service.Status.SaveChannel = c.SaveChannel

	// Create the service map if it doesn't exist.
	if c.Service == nil {
		c.Service = make(map[string]*service.Service)
	}

	// Default Empty List ETag as it depends on default access_token.
	accessTokenDefault := util.FirstNonDefaultWithEnv(
		c.Defaults.Service.LatestVersion.AccessToken,
		c.HardDefaults.Service.LatestVersion.AccessToken)
	github.SetEmptyListETag(accessTokenDefault)

	// SaveHandler that listens for calls to save config changes.
	go c.SaveHandler()

	c.Init()
	c.CheckValues()
}
