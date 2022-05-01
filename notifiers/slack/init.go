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

package slack

import (
	"encoding/json"
	"time"

	"github.com/hymenaios-io/Hymenaios/utils"
	metrics "github.com/hymenaios-io/Hymenaios/web/metrics"
)

// Init Slice metrics.
func (s *Slice) Init(
	log *utils.JLog,
	serviceID *string,
	serviceIcon *string,
	mains *Slice,
	defaults *Slack,
	hardDefaults *Slack,
) {
	jLog = log
	if s == nil {
		return
	}
	if mains == nil {
		mains = &Slice{}
	}

	for key := range *s {
		id := key
		if (*s)[key] == nil {
			(*s)[key] = &Slack{}
		}
		(*s)[key].ID = &id
		(*s)[key].Init(serviceID, serviceIcon, (*mains)[key], defaults, hardDefaults)
	}
}

// Init the Slack metrics and hand out the defaults.
func (s *Slack) Init(
	serviceID *string,
	serviceIcon *string,
	main *Slack,
	defaults *Slack,
	hardDefaults *Slack,
) {
	s.ServiceIcon = serviceIcon

	s.initMetrics(serviceID)

	if s == nil {
		s = &Slack{}
	}

	// Give the matching main
	(*s).Main = main
	if main == nil {
		s.Main = &Slack{}
	}

	// Give the defaults
	(*s).Defaults = defaults
	(*s).HardDefaults = hardDefaults
}

// initMetrics, giving them all a starting value.
func (s *Slack) initMetrics(serviceID *string) {
	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.SlackMetric, *(*s).ID, *serviceID, "SUCCESS")
	metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.SlackMetric, *(*s).ID, *serviceID, "FAIL")
}

// GetDelay before sending.
func (s *Slack) GetDelay() string {
	return *utils.GetFirstNonNilPtr(s.Delay, s.Main.Delay, s.Defaults.Delay, s.HardDefaults.Delay)
}

// GetDelayDuration before sending.
func (s *Slack) GetDelayDuration() time.Duration {
	d, _ := time.ParseDuration(s.GetDelay())
	return d
}

// GetIcon to use for the Slack message.
//
// URL overrides Emoji (when at the same level)
func (s *Slack) GetIcon() (icon *string, iconType string) {
	// `Service.Slack.IconURL/IconEmoji`
	if s.IconURL != nil {
		icon = s.IconURL
		iconType = "url"
		return
	} else if s.IconEmoji != nil {
		icon = s.IconEmoji
		iconType = "emoji"
		return
	}

	// Service.Icon
	if s.ServiceIcon != nil {
		icon = s.ServiceIcon
		iconType = "url"
		return
	}

	// `Slack.IconURL/IconEmoji`
	if s.Main.IconURL != nil {
		icon = s.Main.IconURL
		iconType = "url"
		return
	} else if s.Main.IconEmoji != nil {
		icon = s.Main.IconEmoji
		iconType = "emoji"
		return
	}

	// `Defaults.Slack.IconURL/IconEmoji`
	if s.Defaults.IconURL != nil {
		icon = s.Defaults.IconURL
		iconType = "url"
		return
	} else if s.Defaults.IconEmoji != nil {
		icon = s.Defaults.IconEmoji
		iconType = "emoji"
		return
	}

	// Hardcoded default emoji
	icon = s.HardDefaults.IconEmoji
	iconType = "emoji"
	return
}

// GetMaxTries allowed for the Slack.
func (s *Slack) GetMaxTries() uint {
	return *utils.GetFirstNonNilPtr(s.MaxTries, s.Main.MaxTries, s.Defaults.MaxTries, s.HardDefaults.MaxTries)
}

// GetMessage returns the Slack message template.
func (s *Slack) GetMessage(message *string, context *utils.ServiceInfo) string {
	msg := *utils.GetFirstNonNilPtr(message, s.Message, s.Main.Message, s.Defaults.Message, s.HardDefaults.Message)
	return utils.TemplateString(msg, *context)
}

// GetPayload will format `message` (or the new release message) for this Service and return the full payload
// to be sent.
func (s *Slack) GetPayload(message string, context *utils.ServiceInfo) []byte {
	// Use 'new release' Slack message (Not a custom message)
	if message == "" {
		message = s.GetMessage(nil, context)
	} else {
		message = s.GetMessage(&message, context)
	}

	payload := Payload{
		Username:  s.GetUsername(),
		IconEmoji: nil,
		IconURL:   nil,
		Text:      message,
	}
	payload.SetIcon(s)

	payloadData, _ := json.Marshal(payload)
	return payloadData
}

// GetURL to send to.
func (s *Slack) GetURL() *string {
	return utils.GetFirstNonNilPtr(s.URL, s.Main.URL, s.Defaults.URL)
}

// GetUsername to send as.
func (s *Slack) GetUsername() string {
	return *utils.GetFirstNonNilPtr(s.Username, s.Main.Username, s.Defaults.Username, s.HardDefaults.Username)
}

// SetIcon to use in this `Payload`.
func (sp *Payload) SetIcon(slack *Slack) {
	icon, iconType := slack.GetIcon()
	if iconType == "emoji" {
		sp.IconEmoji = icon
		return
	}
	sp.IconURL = icon
}
