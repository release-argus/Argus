// Copyright [2023] [Argus]
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

	"github.com/release-argus/Argus/util"
)

// GetAllowInvalidCerts returns whether invalid HTTPS certs are allowed.
func (w *WebHook) GetAllowInvalidCerts() bool {
	return *util.GetFirstNonNilPtr(
		w.AllowInvalidCerts,
		w.Main.AllowInvalidCerts,
		w.Defaults.AllowInvalidCerts,
		w.HardDefaults.AllowInvalidCerts)
}

// GetDelay of the WebHook to use before auto-approving.
func (w *WebHook) GetDelay() string {
	return util.GetFirstNonDefault(
		w.Delay,
		w.Main.Delay,
		w.Defaults.Delay,
		w.HardDefaults.Delay)
}

// GetDelayDuration before auto-approving this WebHook.
func (w *WebHook) GetDelayDuration() (duration time.Duration) {
	duration, _ = time.ParseDuration(w.GetDelay())
	return duration
}

// GetDesiredStatusCode of the WebHook.
func (w *WebHook) GetDesiredStatusCode() int {
	return *util.GetFirstNonNilPtr(
		w.DesiredStatusCode,
		w.Main.DesiredStatusCode,
		w.Defaults.DesiredStatusCode,
		w.HardDefaults.DesiredStatusCode)
}

// SetNextRunnable time that the WebHook can be re-run.
func (w *WebHook) SetNextRunnable(time *time.Time) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.nextRunnable = *time
}

// SetExecuting will set a time that the WebHook can be re-run.
//
// addDelay - only used on auto_approved releases
func (w *WebHook) SetExecuting(addDelay bool, sending bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Different times depending on pass/fail
	// pass
	if !util.EvalNilPtr(w.Failed.Get(w.ID), true) {
		parentInterval, _ := time.ParseDuration(*w.ParentInterval)
		w.nextRunnable = time.Now().UTC().Add(2 * parentInterval)
		// fail/nil
	} else {
		w.nextRunnable = time.Now().UTC().Add(15 * time.Second)
	}

	// block for delay
	if addDelay {
		w.nextRunnable = w.nextRunnable.Add(w.GetDelayDuration())
	}

	// Block reruns whilst sending
	if sending {
		w.nextRunnable = w.nextRunnable.Add(time.Hour)
		w.nextRunnable = w.nextRunnable.Add(3 * time.Duration(w.GetMaxTries()) * time.Second)
	}
}

// GetMaxTries allowed for the WebHook.
func (w *WebHook) GetMaxTries() uint {
	return *util.GetFirstNonNilPtr(
		w.MaxTries,
		w.Main.MaxTries,
		w.Defaults.MaxTries,
		w.HardDefaults.MaxTries)
}

// NextRunnable returns the time that the WebHook can be re-run.
func (w *WebHook) NextRunnable() time.Time {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.nextRunnable
}

// IsRunnable will return whether the current time is before NextRunnable
func (w *WebHook) IsRunnable() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return time.Now().UTC().After(w.nextRunnable)
}

// GetRequest will return the WebHook http.request ready to be sent.
func (w *WebHook) GetRequest() (req *http.Request) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var err error
	switch w.GetType() {
	case "github":
		//#nosec G104 -- Disregard
		//nolint:errcheck // ^
		payload, _ := json.Marshal(GitHub{
			Ref:    "refs/heads/master",
			Before: util.RandAlphaNumericLower(40),
			After:  util.RandAlphaNumericLower(40),
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
	req.Header.Set("Connection", "close")
	w.setCustomHeaders(req)
	return
}

// GetSecret for the WebHook.
func (w *WebHook) GetSecret() string {
	return util.GetFirstNonDefault(
		w.Secret,
		w.Main.Secret,
		w.Defaults.Secret,
		w.HardDefaults.Secret)
}

// GetSilentFails returns the flag for whether WebHooks should fail silently or notify.
func (w *WebHook) GetSilentFails() bool {
	return *util.GetFirstNonNilPtr(
		w.SilentFails,
		w.Main.SilentFails,
		w.Defaults.SilentFails,
		w.HardDefaults.SilentFails)
}

// GetType of the WebHook.
func (w *WebHook) GetType() string {
	return util.GetFirstNonDefault(
		w.Type,
		w.Main.Type,
		w.Defaults.Type,
		w.HardDefaults.Type)
}

// GetURL of the WebHook.
func (w *WebHook) GetURL() string {
	url := util.GetFirstNonDefault(
		w.URL,
		w.Main.URL,
		w.Defaults.URL,
		w.HardDefaults.URL)

	url = strings.Clone(
		util.TemplateString(
			url,
			util.ServiceInfo{LatestVersion: w.ServiceStatus.GetLatestVersion()}))
	return url
}
