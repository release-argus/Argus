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
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

// Defaults for the other Structs.
type Defaults struct {
	Service service.Service `yaml:"service,omitempty"`
	Notify  shoutrrr.Slice  `yaml:"notify,omitempty"`
	WebHook webhook.WebHook `yaml:"webhook,omitempty"`
}

// SetDefaults (last resort vars).
func (d *Defaults) SetDefaults() {
	// Service defaults.
	serviceSemanticVersioning := true
	d.Service.Options = opt.Options{
		Interval:           "10m",
		SemanticVersioning: &serviceSemanticVersioning}
	serviceLatestVersionAllowInvalidCerts := false
	usePreRelease := false
	d.Service.LatestVersion = latestver.Lookup{
		AllowInvalidCerts: &serviceLatestVersionAllowInvalidCerts,
		UsePreRelease:     &usePreRelease}
	serviceDeployedVersionLookupAllowInvalidCerts := false
	d.Service.DeployedVersionLookup = &deployedver.Lookup{
		AllowInvalidCerts: &serviceDeployedVersionLookupAllowInvalidCerts}
	serviceAutoApprove := false
	d.Service.Dashboard = service.DashboardOptions{
		AutoApprove: &serviceAutoApprove}

	notifyDefaultOptions := map[string]string{
		"message":   "{{ service_id }} - {{ version }} released",
		"max_tries": "3",
		"delay":     "0s"}

	// Notify defaults.
	d.Notify = make(shoutrrr.Slice)
	d.Notify["discord"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"username": "Argus"}}
	d.Notify["smtp"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "25"},
		Params: types.Params{}}
	d.Notify["googlechat"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["gotify"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{
			"title":    "Argus",
			"priority": "0"}}
	d.Notify["ifttt"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"title":             "Argus",
			"usemessageasvalue": "2",
			"usetitleasvalue":   "0"}}
	d.Notify["join"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["mattermost"] = &shoutrrr.Shoutrrr{
		Options: map[string]string{
			"message":   "<{{ service_url }}|{{ service_id }}> - {{ version }} released{% if web_url %} (<{{ web_url }}|changelog>){% endif %}",
			"max_tries": "3",
			"delay":     "0s"},
		URLFields: map[string]string{
			"username": "Argus",
			"port":     "443"}}
	d.Notify["matrix"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{}}
	d.Notify["opsgenie"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["pushbullet"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{
			"title": "Argus"}}
	d.Notify["pushover"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["rocketchat"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{}}
	d.Notify["slack"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"botname": "Argus"}}
	d.Notify["team"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["telegram"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"notification": "yes",
			"preview":      "yes"}}
	d.Notify["zulip"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	// InitMaps
	for _, notify := range d.Notify {
		notify.InitMaps()
	}

	// WebHook defaults.
	d.WebHook.Type = "github"
	d.WebHook.Delay = "0s"
	webhookMaxTries := uint(3)
	d.WebHook.MaxTries = &webhookMaxTries
	webhookAllowInvalidCerts := false
	d.WebHook.AllowInvalidCerts = &webhookAllowInvalidCerts
	webhookDesiredStatusCode := 0
	d.WebHook.DesiredStatusCode = &webhookDesiredStatusCode
	webhookSilentFails := false
	d.WebHook.SilentFails = &webhookSilentFails
}

// CheckValues are valid.
func (d *Defaults) CheckValues() (errs error) {
	prefix := "  "

	// Service
	if err := d.Service.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%sservice:\\%w",
			util.ErrorToString(errs), prefix, err)
	}

	// Notify
	for i := range d.Notify {
		// Remove the types since the key is the type
		d.Notify[i].Type = ""
	}
	if err := d.Notify.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), err)
	}

	// WebHook
	if err := d.WebHook.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%swebhook:\\%w",
			util.ErrorToString(errs), prefix, err)
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
