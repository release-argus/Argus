// Copyright [2026] [Argus]
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
	"context"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/util"
)

// Load reads and decodes the config file, then starts the save handler.
func (c *Config) Load(
	ctx context.Context,
	g *errgroup.Group,
	file string,
	flagset *map[string]bool,
) bool {
	// Initialise the Log if it hasn't been already.
	logx.Init("ERROR", false)

	// Load the .env file (if it exists).
	envFile := filepath.Join(filepath.Dir(file), ".env")
	err := loadEnvFile(envFile)
	logx.Warn(err, logx.LogFrom{}, err != nil)

	c.File = file
	c.Settings.NilUndefinedFlags(flagset)

	//#nosec G304 -- Loading the file asked for by the user.
	data, err := os.ReadFile(file)
	if err != nil {
		logx.Fatal(
			fmt.Sprintf(
				"Error reading %q\n%s",
				file, err,
			),
			logx.LogFrom{},
		)
		return false
	}

	if err := c.Decode(data); err != nil {
		logx.Fatal(
			fmt.Sprintf(
				"Unmarshal of %q failed\n%s",
				file, err,
			),
			logx.LogFrom{},
		)
		return false
	}

	c.GetOrder(data)

	// Default Empty List ETag as it depends on default access_token.
	accessTokenDefault := util.FirstNonDefaultWithEnv(
		c.Defaults.Service.LatestVersion.AccessToken,
		c.HardDefaults.Service.LatestVersion.AccessToken,
	)
	github.SetEmptyListETag(accessTokenDefault)

	// SaveHandler that listens for calls to save config changes.
	g.Go(func() error {
		c.SaveHandler(ctx)
		return nil
	})

	return c.CheckValues()
}

// InitDefaults initialises configuration state by assigning hard defaults to the defaults.
// It also propagates selected configuration into dependent subsystems such
// as logging and service status channels.
func (c *Config) InitDefaults() {
	c.HardDefaults.Default()
	c.Defaults.SetDefaults(&c.HardDefaults)

	// Settings.
	c.Settings.Default()

	// Save Channel.
	c.HardDefaults.Service.Status.SaveChannel = c.SaveChannel

	// Log.
	logx.SetTimestamps(*c.Settings.LogTimestamps())
	logx.SetLevel(c.Settings.LogLevel())
}
