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

package gotify

import (
	"encoding/json"
	"time"

	"github.com/release-argus/Argus/utils"
	metrics "github.com/release-argus/Argus/web/metrics"
)

// Init the Slice metrics.
func (g *Slice) Init(
	log *utils.JLog,
	serviceID *string,
	mains *Slice,
	defaults *Gotify,
	hardDefaults *Gotify,
) {
	jLog = log
	if g == nil {
		return
	}
	if mains == nil {
		mains = &Slice{}
	}

	for key := range *g {
		id := key
		if (*g)[key] == nil {
			(*g)[key] = &Gotify{}
		}
		(*g)[key].ID = &id
		(*g)[key].Init(serviceID, (*mains)[key], defaults, hardDefaults)
	}
}

// Init the Gotify metrics and hand out the defaults.
func (g *Gotify) Init(
	serviceID *string,
	main *Gotify,
	defaults *Gotify,
	hardDefaults *Gotify,
) {
	g.initMetrics(serviceID)

	if g == nil {
		g = &Gotify{}
		g.Extras = &Extras{}
	}

	// Give the matching main
	(*g).Main = main
	if main == nil {
		g.Main = &Gotify{}
	}

	// Give the defaults
	(*g).Defaults = defaults
	(*g).HardDefaults = hardDefaults
}

// initMetrics, giving them all a starting value.
func (g *Gotify) initMetrics(serviceID *string) {
	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.GotifyMetric, *(*g).ID, *serviceID, "SUCCESS")
	metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.GotifyMetric, *(*g).ID, *serviceID, "FAIL")
}

// GetDelay before sending.
func (g *Gotify) GetDelay() string {
	return *utils.GetFirstNonNilPtr(g.Delay, g.Main.Delay, g.Defaults.Delay, g.HardDefaults.Delay)
}

// GetDelayDuration before sending.
func (g *Gotify) GetDelayDuration() time.Duration {
	d, _ := time.ParseDuration(g.GetDelay())
	return d
}

// GetExtraAndroidAction to use with the Gotification.
func (g *Gotify) GetExtraAndroidAction() *string {
	if g.Extras != nil && g.Extras.AndroidAction != nil {
		return g.Extras.AndroidAction
	}
	if g.Main.Extras != nil && g.Main.Extras.AndroidAction != nil {
		return g.Main.Extras.AndroidAction
	}
	if g.Defaults.Extras != nil && g.Defaults.Extras.AndroidAction != nil {
		return g.Defaults.Extras.AndroidAction
	}
	return nil
}

// GetExtraClientDisplay to use with the Gotification.
func (g *Gotify) GetExtraClientDisplay() *string {
	if g.Extras != nil && g.Extras.ClientDisplay != nil {
		return g.Extras.ClientDisplay
	}
	if g.Main.Extras != nil && g.Main.Extras.ClientDisplay != nil {
		return g.Main.Extras.ClientDisplay
	}
	if g.Defaults.Extras != nil && g.Defaults.Extras.ClientDisplay != nil {
		return g.Defaults.Extras.ClientDisplay
	}
	return nil
}

// GetExtraClientNotification to use with the Gotification.
func (g *Gotify) GetExtraClientNotification() *string {
	if g.Extras != nil && g.Extras.ClientNotification != nil {
		return g.Extras.ClientNotification
	}
	if g.Main.Extras != nil && g.Main.Extras.ClientNotification != nil {
		return g.Main.Extras.ClientNotification
	}
	if g.Defaults.Extras != nil && g.Defaults.Extras.ClientNotification != nil {
		return g.Defaults.Extras.ClientNotification
	}
	return nil
}

// GetMaxTries allowed for the Gotification.
func (g *Gotify) GetMaxTries() uint {
	return *utils.GetFirstNonNilPtr(g.MaxTries, g.Main.MaxTries, g.Defaults.MaxTries, g.HardDefaults.MaxTries)
}

// GetMessage of the Gotification after the context is applied and template evaluated.
func (g *Gotify) GetMessage(context *utils.ServiceInfo) string {
	message := *utils.GetFirstNonNilPtr(g.Message, g.Main.Message, g.Defaults.Message, g.HardDefaults.Message)
	return utils.TemplateString(message, *context)
}

// GetPayload will format `message` (or the new release message) for the Service and return the full payload
// to be sent.
func (g *Gotify) GetPayload(
	title string,
	message string,
	serviceInfo *utils.ServiceInfo,
) []byte {
	// Use 'new release' Gotify message (Not a custom message)
	if message == "" {
		message = (*g).GetMessage(serviceInfo)
	}

	if title == "" {
		title = (*g).GetTitle(serviceInfo)
	}

	payload := Payload{
		Message:  message,
		Priority: g.GetPriority(),
		Title:    title,
		Extras:   map[string]interface{}{},
	}

	payload.HandleExtras(g, serviceInfo)

	payloadData, _ := json.Marshal(payload)
	return payloadData
}

// GetPriority to give the message.
func (g *Gotify) GetPriority() int {
	return *utils.GetFirstNonNilPtr(g.Priority, g.Main.Priority, g.Defaults.Priority, g.HardDefaults.Priority)
}

// GetToken to send with.
func (g *Gotify) GetToken() *string {
	return utils.GetFirstNonNilPtr(g.Token, g.Main.Token, g.Defaults.Token)
}

// GetTitle of the Gotification after the context is applied and template evaluated.
func (g *Gotify) GetTitle(serviceInfo *utils.ServiceInfo) string {
	title := *utils.GetFirstNonNilPtr(g.Title, g.Main.Title, g.Defaults.Title, g.HardDefaults.Title)
	return utils.TemplateString(title, *serviceInfo)
}

// GetURL to send to.
func (g *Gotify) GetURL() *string {
	return utils.GetFirstNonNilPtr(g.URL, g.Main.URL, g.Defaults.URL)
}
