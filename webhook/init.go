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

package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hymenaios-io/Hymenaios/notifiers/gotify"
	"github.com/hymenaios-io/Hymenaios/notifiers/slack"
	"github.com/hymenaios-io/Hymenaios/utils"
	metrics "github.com/hymenaios-io/Hymenaios/web/metrics"
)

// Init the Slice metrics and hand out the defaults/notifiers.
func (w *Slice) Init(
	log *utils.JLog,
	serviceID *string,
	main *Slice,
	defaults *WebHook,
	hardDefaults *WebHook,
	gotifyNotifiers *gotify.Slice,
	slackNotifiers *slack.Slice,
) {
	jLog = log
	if w == nil {
		return
	}

	for key := range *w {
		id := key
		if (*w)[key] == nil {
			(*w)[key] = &WebHook{}
		}
		(*w)[key].ID = &id
		(*w)[key].Init(
			serviceID,
			(*main)[key],
			defaults,
			hardDefaults,
			gotifyNotifiers,
			slackNotifiers,
		)
	}
}

// Init the WebHook metrics and give the defaults/notifiers.
func (w *WebHook) Init(
	serviceID *string,
	main *WebHook,
	defaults *WebHook,
	hardDefaults *WebHook,
	gotifyNotifiers *gotify.Slice,
	slackNotifiers *slack.Slice,
) {
	if w == nil {
		return
	}

	w.initMetrics(*serviceID)

	if w == nil {
		w = &WebHook{}
	}

	// Give the matchinw main
	(*w).Main = main
	if main == nil {
		w.Main = &WebHook{}
	}

	// Give the defaults
	(*w).Defaults = defaults
	(*w).HardDefaults = hardDefaults

	// WebHook fail notifiers
	(*w).Notifiers = &Notifiers{
		Gotify: gotifyNotifiers,
		Slack:  slackNotifiers,
	}
}

// initMetrics, giving them all a starting value.
func (w *WebHook) initMetrics(serviceID string) {
	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.WebHookMetric, *(*w).ID, serviceID, "SUCCESS")
	metrics.InitPrometheusCounterWithIDAndServiceIDAndResult(metrics.WebHookMetric, *(*w).ID, serviceID, "FAIL")

	// ##########
	// # Gauges #
	// ##########
	metrics.SetPrometheusGaugeWithID(metrics.AckWaiting, serviceID, float64(0))
}

// GetDelay of the WebHook to use before auto-approving.
func (w *WebHook) GetDelay() string {
	return *utils.GetFirstNonNilPtr(w.Delay, w.Main.Delay, w.Defaults.Delay, w.HardDefaults.Delay)
}

// GetDelayDuration before auto-approving this WebHook.
func (w *WebHook) GetDelayDuration() time.Duration {
	d, _ := time.ParseDuration(w.GetDelay())
	return d
}

// GetDesiredStatusCode of the WebHook.
func (w *WebHook) GetDesiredStatusCode() int {
	return *utils.GetFirstNonNilPtr(w.DesiredStatusCode, w.Main.DesiredStatusCode, w.Defaults.DesiredStatusCode, w.HardDefaults.DesiredStatusCode)
}

// GetMaxTries allowed for the WebHook.
func (w *WebHook) GetMaxTries() uint {
	return *utils.GetFirstNonNilPtr(w.MaxTries, w.Main.MaxTries, w.Defaults.MaxTries, w.HardDefaults.MaxTries)
}

// GetRequest will return the WebHook http.request ready to be sent.
func (w *WebHook) GetRequest() (req *http.Request) {
	// GitHub style payload.
	if w.GetType() == "github" {
		payload, err := json.Marshal(GitHub{Ref: "refs/heads/master", Before: utils.RandAlphaNumericLower(40), After: utils.RandAlphaNumericLower(40)})
		if err != nil {
			return
		}
		req, err = http.NewRequest(http.MethodPost, *w.GetURL(), bytes.NewReader(payload))
		if err != nil {
			return nil
		}
		req.Header.Set("Content-Type", "application/json")

		req = SetGitHubHeaders(req, payload, *w.GetSecret())
	}
	return
}

// GetType of the WebHook.
func (w *WebHook) GetType() string {
	return *utils.GetFirstNonNilPtr(w.Type, w.Main.Type, w.Defaults.Type)
}

// GetSecret for the WebHook.
func (w *WebHook) GetSecret() *string {
	return utils.GetFirstNonNilPtr(w.Secret, w.Main.Secret, w.Defaults.Secret)
}

// GetSilentFails returns the flag for whether WebHooks should fail silently or notify.
func (w *WebHook) GetSilentFails() bool {
	return *utils.GetFirstNonNilPtr(w.SilentFails, w.Main.SilentFails, w.Defaults.SilentFails, w.HardDefaults.SilentFails)
}

// GetURL of the WebHook.
func (w *WebHook) GetURL() *string {
	return utils.GetFirstNonNilPtr(w.URL, w.Main.URL, w.Defaults.URL)
}
