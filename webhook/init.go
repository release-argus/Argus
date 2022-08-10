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

package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	metrics "github.com/release-argus/Argus/web/metrics"
)

// Init the Slice metrics and hand out the defaults/notifiers.
func (w *Slice) Init(
	log *utils.JLog,
	serviceStatus *service_status.Status,
	mains *Slice,
	defaults *WebHook,
	hardDefaults *WebHook,
	shoutrrrNotifiers *shoutrrr.Slice,
	parentInterval *string,
) {
	jLog = log
	if w == nil || len(*w) == 0 {
		return
	}
	if mains == nil {
		mains = &Slice{}
	}

	for id := range *w {
		if (*w)[id] == nil {
			(*w)[id] = &WebHook{}
		}
		(*w)[id].ID = id
		(*w)[id].Failed = &serviceStatus.Fails.WebHook
		(*w)[id].Init(
			serviceStatus,
			(*mains)[id],
			defaults,
			hardDefaults,
			shoutrrrNotifiers,
			parentInterval,
		)
	}
}

// Init the WebHook metrics and give the defaults/notifiers.
func (w *WebHook) Init(
	serviceStatus *service_status.Status,
	main *WebHook,
	defaults *WebHook,
	hardDefaults *WebHook,
	shoutrrrNotifiers *shoutrrr.Slice,
	parentInterval *string,
) {
	w.ParentInterval = parentInterval
	w.ServiceStatus = serviceStatus
	w.initMetrics()

	// Give the matching main
	w.Main = main
	if main == nil {
		w.Main = &WebHook{}
	}

	// Give the defaults
	w.Defaults = defaults
	w.HardDefaults = hardDefaults

	// WebHook fail notifiers
	w.Notifiers = &Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// initMetrics, giving them all a starting value.
func (w *WebHook) initMetrics() {
	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterActions(metrics.WebHookMetric, w.ID, *w.ServiceStatus.ServiceID, "", "SUCCESS")
	metrics.InitPrometheusCounterActions(metrics.WebHookMetric, w.ID, *w.ServiceStatus.ServiceID, "", "FAIL")

	// ##########
	// # Gauges #
	// ##########
	metrics.SetPrometheusGaugeWithID(metrics.AckWaiting, *w.ServiceStatus.ServiceID, float64(0))
}

// GetAllowInvalidCerts returns whether invalid HTTPS certs are allowed.
func (w *WebHook) GetAllowInvalidCerts() bool {
	return *utils.GetFirstNonNilPtr(w.AllowInvalidCerts, w.Main.AllowInvalidCerts, w.Defaults.AllowInvalidCerts, w.HardDefaults.AllowInvalidCerts)
}

// GetDelay of the WebHook to use before auto-approving.
func (w *WebHook) GetDelay() string {
	return utils.GetFirstNonDefault(w.Delay, w.Main.Delay, w.Defaults.Delay, w.HardDefaults.Delay)
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

// GetFailStatus of this WebHook.
func (w *WebHook) GetFailStatus() *bool {
	return (*w.Failed)[w.ID]
}

// GetMaxTries allowed for the WebHook.
func (w *WebHook) GetMaxTries() uint {
	return *utils.GetFirstNonNilPtr(w.MaxTries, w.Main.MaxTries, w.Defaults.MaxTries, w.HardDefaults.MaxTries)
}

// GetRequest will return the WebHook http.request ready to be sent.
func (w *WebHook) GetRequest() (req *http.Request) {
	var err error
	switch w.GetType() {
	case "github":
		//#nosec G104 -- Disregard
		//nolint:errcheck // ^
		payload, _ := json.Marshal(GitHub{
			Ref:    "refs/heads/master",
			Before: utils.RandAlphaNumericLower(40),
			After:  utils.RandAlphaNumericLower(40),
		})

		req, err = http.NewRequest(http.MethodPost, w.GetURL(), bytes.NewReader(payload))
		if err != nil {
			return nil
		}
		req.Header.Set("Content-Type", "application/json")

		SetGitHubHeaders(req, payload, w.GetSecret())
	case "gitlab":
		req, err = http.NewRequest(http.MethodPost, w.GetURL(), nil)
		if err != nil {
			return nil
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		SetGitLabParameter(req, w.GetSecret())
	}
	w.SetCustomHeaders(req)
	return
}

// GetType of the WebHook.
func (w *WebHook) GetType() string {
	return utils.GetFirstNonDefault(w.Type, w.Main.Type, w.Defaults.Type, w.HardDefaults.Type)
}

// GetSecret for the WebHook.
func (w *WebHook) GetSecret() string {
	return utils.GetFirstNonDefault(w.Secret, w.Main.Secret, w.Defaults.Secret)
}

// GetSilentFails returns the flag for whether WebHooks should fail silently or notify.
func (w *WebHook) GetSilentFails() bool {
	return *utils.GetFirstNonNilPtr(w.SilentFails, w.Main.SilentFails, w.Defaults.SilentFails, w.HardDefaults.SilentFails)
}

// GetURL of the WebHook.
func (w *WebHook) GetURL() string {
	url := utils.GetFirstNonDefault(w.URL, w.Main.URL, w.Defaults.URL)
	if w.ServiceStatus != nil && url != "" && strings.Contains(url, "{") {
		url = strings.Clone(utils.TemplateString(url, utils.ServiceInfo{LatestVersion: w.ServiceStatus.LatestVersion}))
	}
	return url
}

// IsRunnable will return whether the current time is before NextRunnable
func (w *WebHook) IsRunnable() bool {
	return time.Now().UTC().After(w.NextRunnable)
}

// SetFailStatus of this WebHook.
func (w *WebHook) SetFailStatus(status *bool) {
	(*w.Failed)[w.ID] = status
}

// SetNextRunnable time that the WebHook can be re-run.
//
// addDelay - only used on auto_approved releases
func (w *WebHook) SetNextRunnable(addDelay bool, sending bool) {
	// Different times depending on pass/fail
	// pass
	if !utils.EvalNilPtr(w.GetFailStatus(), true) {
		parentInterval, _ := time.ParseDuration(*w.ParentInterval)
		w.NextRunnable = time.Now().UTC().Add(2 * parentInterval)
		// fail/nil
	} else {
		w.NextRunnable = time.Now().UTC().Add(15 * time.Second)
	}
	// block for delay
	if addDelay {
		w.NextRunnable = w.NextRunnable.Add(w.GetDelayDuration())
	}
	// Block reruns whilst sending
	if sending {
		w.NextRunnable = w.NextRunnable.Add(time.Hour)
		w.NextRunnable = w.NextRunnable.Add(3 * time.Duration(w.GetMaxTries()) * time.Second)
	}
}

// ResetFails of this Slice
func (w *Slice) ResetFails() {
	if w == nil {
		return
	}
	for i := range *w {
		(*w)[i].SetFailStatus(nil)
	}
}
