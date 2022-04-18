// Copyright [2022] [Hymenaios]
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
	"github.com/hymenaios-io/Hymenaios/notifiers/gotify"
	"github.com/hymenaios-io/Hymenaios/notifiers/slack"
	"github.com/hymenaios-io/Hymenaios/service"
	"github.com/hymenaios-io/Hymenaios/utils"
	"github.com/hymenaios-io/Hymenaios/webhook"
)

var (
	jLog *utils.JLog
)

// Config for Hymenaios.
type Config struct {
	File         string         `yaml:"-"`        // Path to the config file (--config.file='').
	Settings     Settings       `yaml:"settings"` // Settings for the program.
	HardDefaults Defaults       `yaml:"-"`        // Hardcoded default values for the various parameters.
	Defaults     Defaults       `yaml:"defaults"` // Default values for the various parameters.
	Gotify       *gotify.Slice  `yaml:"gotify"`   // Gotify message(s) to send on a new release.
	Slack        *slack.Slice   `yaml:"slack"`    // Slack message(s) to send on a new release.
	WebHook      *webhook.Slice `yaml:"webhook"`  // WebHook(s) to send on a new release.
	Service      service.Slice  `yaml:"service"`  // The service(s) to monitor.
	Order        []string       `yaml:"order"`    // Ordering for the Service(s) in the WebUI.
	SaveChannel  *chan bool     `yaml:"-"`        // Channel for triggering a save of the config.
}
