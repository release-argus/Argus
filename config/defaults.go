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

	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/conversions"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

// Defaults for the other Structs.
type Defaults struct {
	Service service.Service `yaml:"service,omitempty"`
	Notify  shoutrrr.Slice  `yaml:"notify,omitempty"`
	WebHook webhook.WebHook `yaml:"webhook,omitempty"`
	// TODO: Remove deprecated V
	Gotify *conversions.Gotify `yaml:"gotify,omitempty"`
	Slack  *conversions.Slack  `yaml:"slack,omitempty"`
}

// SetDefaults (last resort vars).
func (d *Defaults) SetDefaults() {
	// Service defaults.
	serviceAutoApprove := false
	d.Service.AutoApprove = &serviceAutoApprove
	serviceAllowInvalidCerts := false
	d.Service.AllowInvalidCerts = &serviceAllowInvalidCerts
	serviceIgnoreMisses := true
	d.Service.IgnoreMisses = &serviceIgnoreMisses
	serviceInterval := "10m"
	d.Service.Interval = &serviceInterval
	usePreRelease := false
	d.Service.UsePreRelease = &usePreRelease
	serviceSemanticVersioning := true
	d.Service.SemanticVersioning = &serviceSemanticVersioning
	// Service DeployedVersionLookup defaults.
	serviceDeployedVersionLookupAllowInvalidCerts := false
	d.Service.DeployedVersionLookup = &service.DeployedVersionLookup{}
	d.Service.DeployedVersionLookup.AllowInvalidCerts = &serviceDeployedVersionLookupAllowInvalidCerts

	notifyDefaultOptions := map[string]string{
		"message":   "{{ service_id }} - {{ version }} released",
		"max_tries": "3",
		"delay":     "0s",
	}

	// Notify defaults.
	d.Notify = make(shoutrrr.Slice)
	d.Notify["discord"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{"username": "Argus"},
	}
	d.Notify["discord"].InitMaps()
	d.Notify["smtp"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		URLFields: &map[string]string{
			"port": "25",
		},
		Params: &types.Params{},
	}
	d.Notify["smtp"].InitMaps()
	d.Notify["googlechat"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{},
	}
	d.Notify["googlechat"].InitMaps()
	d.Notify["gotify"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		URLFields: &map[string]string{
			"port": "443",
		},
		Params: &types.Params{"title": "Argus"},
	}
	d.Notify["gotify"].InitMaps()
	d.Notify["ifttt"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{"title": "Argus"},
	}
	d.Notify["ifttt"].InitMaps()
	d.Notify["join"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{},
	}
	d.Notify["join"].InitMaps()
	d.Notify["mattermost"] = &shoutrrr.Shoutrrr{
		Options: &map[string]string{
			"message":   "<{{ service_url }}|{{ service_id }}> - {{ version }}released{% if web_url %} (<{{ web_url }}|changelog>){% endif %}",
			"max_tries": "3",
			"delay":     "0s",
		},
		URLFields: &map[string]string{
			"port": "443",
		},
		Params: &types.Params{"username": "Argus"},
	}
	d.Notify["mattermost"].InitMaps()
	d.Notify["matrix"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		URLFields: &map[string]string{
			"port": "443",
		},
		Params: &types.Params{},
	}
	d.Notify["matrix"].InitMaps()
	d.Notify["ops_genie"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{},
	}
	d.Notify["ops_genie"].InitMaps()
	d.Notify["pushbullet"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		URLFields: &map[string]string{
			"port": "443",
		},
		Params: &types.Params{"title": "Argus"},
	}
	d.Notify["pushbullet"].InitMaps()
	d.Notify["pushover"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{},
	}
	d.Notify["pushover"].InitMaps()
	d.Notify["rocketchat"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		URLFields: &map[string]string{
			"port": "443",
		},
		Params: &types.Params{},
	}
	d.Notify["rocketchat"].InitMaps()
	d.Notify["slack"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{"botname": "Argus"},
	}
	d.Notify["slack"].InitMaps()
	d.Notify["team"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{},
	}
	d.Notify["team"].InitMaps()
	d.Notify["telegram"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{},
	}
	d.Notify["telegram"].InitMaps()
	d.Notify["zulip_chat"] = &shoutrrr.Shoutrrr{
		Options: &notifyDefaultOptions,
		Params:  &types.Params{},
	}
	d.Notify["zulip_chat"].InitMaps()

	// WebHook defaults.
	webhookDelay := "0s"
	d.WebHook.Delay = &webhookDelay
	webhookMaxTries := uint(3)
	d.WebHook.MaxTries = &webhookMaxTries
	webhookDesiredStatusCode := 0
	d.WebHook.DesiredStatusCode = &webhookDesiredStatusCode
	webhookSilentFails := false
	d.WebHook.SilentFails = &webhookSilentFails
}

// CheckValues are valid.
func (d *Defaults) CheckValues() (errs error) {
	prefix := "  "

	// Service
	serviceErrs := false
	if err := d.Service.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%sservice:\\%w", utils.ErrorToString(errs), prefix, err)
		serviceErrs = true
	}
	if err := d.Service.DeployedVersionLookup.CheckValues(prefix); err != nil {
		customPrefix := ""
		if !serviceErrs {
			customPrefix = fmt.Sprintf("%s%sservice:\\", utils.ErrorToString(errs), prefix)
		}
		errs = fmt.Errorf("%s%s%s  deployed_version:\\%w", utils.ErrorToString(errs), customPrefix, prefix, err)
	}

	// Notify
	for i := range d.Notify {
		// Remove the types since the key is the type
		d.Notify[i].Type = ""
	}
	if err := d.Notify.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%w", utils.ErrorToString(errs), err)
	}

	// WebHook
	if err := d.WebHook.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%w", utils.ErrorToString(errs), err)
	}

	return errs
}

// Print the defaults Strcut.
func (d *Defaults) Print() {
	fmt.Println("defaults:")

	// Service defaults.
	fmt.Println("  service:")
	d.Service.Print("    ")

	// Notify defaults.
	d.Notify.Print("  ")

	// WebHook defaults.
	fmt.Println("  webhook:")
	d.WebHook.Print("    ")
}
