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
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

var (
	jLog *utils.JLog
)

// Config for Argus.
type Config struct {
	File         string         `yaml:"-"`                  // Path to the config file (--config.file='').
	Settings     Settings       `yaml:"settings,omitempty"` // Settings for the program.
	HardDefaults Defaults       `yaml:"-"`                  // Hardcoded default values for the various parameters.
	Defaults     Defaults       `yaml:"defaults,omitempty"` // Default values for the various parameters.
	Notify       shoutrrr.Slice `yaml:"notify,omitempty"`   // Shoutrrr message(s) to send on a new release.
	WebHook      webhook.Slice  `yaml:"webhook,omitempty"`  // WebHook(s) to send on a new release.
	Service      service.Slice  `yaml:"service,omitempty"`  // The service(s) to monitor.
	All          []string       `yaml:"-"`                  // Ordered list of all Service(s).
	Order        *[]string      `yaml:"-"`                  // Ordered list of all enabled Service(s).
	SaveChannel  *chan bool     `yaml:"-"`                  // Channel for triggering a save of the config.
}
