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
	"fmt"

	"github.com/hymenaios-io/Hymenaios/notifiers/gotify"
	"github.com/hymenaios-io/Hymenaios/notifiers/slack"
	"github.com/hymenaios-io/Hymenaios/service"
	"github.com/hymenaios-io/Hymenaios/utils"
	"github.com/hymenaios-io/Hymenaios/webhook"
)

// Defaults for the other Structs.
type Defaults struct {
	Service service.Service `yaml:"service"`
	Gotify  gotify.Gotify   `yaml:"gotify"`
	Slack   slack.Slack     `yaml:"slack"`
	WebHook webhook.WebHook `yaml:"webhook"`
}

// SetDefaults (last resort vars).
func (d *Defaults) SetDefaults() {
	// Service defaults.
	serviceAutoApprove := false
	d.Service.AutoApprove = &serviceAutoApprove
	serviceAllowInvalidCerts := false
	d.Service.AllowInvalidCerts = &serviceAllowInvalidCerts
	serviceIgnoreMisses := false
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

	// Gotify defaults.
	gotifyDelay := "0s"
	d.Gotify.Delay = &gotifyDelay
	gotifyNaxTries := uint(3)
	d.Gotify.MaxTries = &gotifyNaxTries
	gotifyTitle := "Hymenaios"
	d.Gotify.Title = &gotifyTitle
	gotifyNessage := "{{ service_id }} - {{ version }} released"
	d.Gotify.Message = &gotifyNessage
	gotifyPriority := 5
	d.Gotify.Priority = &gotifyPriority

	// Slack defaults.
	slackDelay := "0s"
	d.Slack.Delay = &slackDelay
	slackMaxTries := uint(3)
	d.Slack.MaxTries = &slackMaxTries
	slackIconEmoji := ":github:"
	d.Slack.IconEmoji = &slackIconEmoji
	slackUsername := "Hymenaios"
	d.Slack.Username = &slackUsername
	slackMessage := "<{{ service_url }}|{{ service_id }}> - {{ version }}released{% if web_url %} (<{{ web_url }}|changelog>){% endif %}"
	d.Slack.Message = &slackMessage

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

	// Gotify
	if err := d.Gotify.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%sgotify:\\%w", utils.ErrorToString(errs), prefix, err)
	}

	// Slack
	if err := d.Slack.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%sslack:\\%w", utils.ErrorToString(errs), prefix, err)
	}

	// WebHook
	if err := d.WebHook.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%swebhook:%w", utils.ErrorToString(errs), prefix, err)
	}

	return errs
}

// Print the defaults Strcut.
func (d *Defaults) Print() {
	fmt.Println("defaults:")

	// Service defaults.
	fmt.Println("  service:")
	d.Service.Print("    ")

	// Gotify defaults.
	fmt.Println("  gotify:")
	d.Gotify.Print("    ")

	// Slack defaults.
	fmt.Println("  slack:")
	d.Slack.Print("    ")

	// WebHook defaults.
	fmt.Println("  webhook:")
	d.WebHook.Print("    ")
}
