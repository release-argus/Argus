// Copyright [2025] [Argus]
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

// Package webhook provides WebHook functionality to services.
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
func (wh *WebHook) GetAllowInvalidCerts() bool {
	return *util.FirstNonNilPtr(
		wh.AllowInvalidCerts,
		wh.Main.AllowInvalidCerts,
		wh.Defaults.AllowInvalidCerts,
		wh.HardDefaults.AllowInvalidCerts)
}

// GetDelay of the WebHook to use before auto-approving.
func (wh *WebHook) GetDelay() string {
	return util.FirstNonDefault(
		wh.Delay,
		wh.Main.Delay,
		wh.Defaults.Delay,
		wh.HardDefaults.Delay)
}

// GetDelayDuration before auto-approving this WebHook.
func (wh *WebHook) GetDelayDuration() (duration time.Duration) {
	duration, _ = time.ParseDuration(wh.GetDelay())
	return duration
}

// GetDesiredStatusCode of the WebHook.
func (wh *WebHook) GetDesiredStatusCode() uint16 {
	return *util.FirstNonNilPtr(
		wh.DesiredStatusCode,
		wh.Main.DesiredStatusCode,
		wh.Defaults.DesiredStatusCode,
		wh.HardDefaults.DesiredStatusCode)
}

// DidFail returns whether the last send of this WebHook failed.
func (wh *WebHook) DidFail() *bool {
	return wh.Failed.Get(wh.ID)
}

// SetFailed will set the 'Fail' status of this WebHook.
func (wh *WebHook) SetFail(state *bool) {
	wh.Failed.Set(wh.ID, state)
}

// NextRunnable returns the time the WebHook can next run.
func (wh *WebHook) NextRunnable() time.Time {
	return wh.Failed.NextRunnable(wh.ID)
}

// SetNextRunnable time the WebHook can next run.
func (wh *WebHook) SetNextRunnable(time time.Time) {
	wh.Failed.SetNextRunnable(wh.ID, time)
}

// SetExecuting will set the time the WebHook can next run.
//
// Parameters:
//
//	addDelay: only used on auto_approved releases.
//	received: whether the WebHook has received a response.
func (wh *WebHook) SetExecuting(addDelay bool, received bool) {
	wh.mutex.Lock()
	defer wh.mutex.Unlock()

	var nextRunnable time.Time

	// Different times depending on pass/fail.
	// pass.
	if !util.DereferenceOrValue(wh.DidFail(), true) {
		parentInterval, _ := time.ParseDuration(*wh.ParentInterval)
		nextRunnable = time.Now().UTC().Add(2 * parentInterval)
		// fail/nil.
	} else {
		nextRunnable = time.Now().UTC().Add(15 * time.Second)
	}

	// Block for delay.
	if addDelay {
		nextRunnable = nextRunnable.Add(wh.GetDelayDuration())
	}

	// Block reruns whilst waiting for a response.
	if received {
		nextRunnable = nextRunnable.Add(time.Hour)
	}
	wh.SetNextRunnable(nextRunnable)
}

// GetMaxTries allowed for the WebHook.
func (wh *WebHook) GetMaxTries() uint8 {
	return *util.FirstNonNilPtr(
		wh.MaxTries,
		wh.Main.MaxTries,
		wh.Defaults.MaxTries,
		wh.HardDefaults.MaxTries)
}

// IsRunnable returns whether the current time is before NextRunnable.
func (wh *WebHook) IsRunnable() bool {
	return time.Now().UTC().After(wh.NextRunnable())
}

// BuildRequest returns the WebHook http.request ready to be sent.
func (wh *WebHook) BuildRequest() (req *http.Request) {
	wh.mutex.RLock()
	defer wh.mutex.RUnlock()

	var err error
	switch wh.GetType() {
	case "github":
		//#nosec G104 -- Disregard.
		//nolint:errcheck // ^
		payload, _ := json.Marshal(GitHub{
			Ref:    "refs/heads/master",
			Before: util.RandAlphaNumericLower(40),
			After:  util.RandAlphaNumericLower(40),
		})

		req, err = http.NewRequest(http.MethodPost,
			wh.GetURL(),
			bytes.NewReader(payload))
		if err != nil {
			return nil
		}
		req.Header.Set("Content-Type", "application/json")

		SetGitHubHeaders(req, payload, wh.GetSecret())
	case "gitlab":
		req, err = http.NewRequest(http.MethodPost,
			wh.GetURL(),
			nil)
		if err != nil {
			return nil
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		SetGitLabParameter(req, wh.GetSecret())
	}
	req.Header.Set("Connection", "close")
	wh.setHeaders(req)
	return
}

// GetSecret for the WebHook.
func (wh *WebHook) GetSecret() string {
	return util.FirstNonDefaultWithEnv(
		wh.Secret,
		wh.Main.Secret,
		wh.Defaults.Secret,
		wh.HardDefaults.Secret)
}

// GetSilentFails returns the flag for whether WebHooks should fail silently or notify.
func (wh *WebHook) GetSilentFails() bool {
	return *util.FirstNonNilPtr(
		wh.SilentFails,
		wh.Main.SilentFails,
		wh.Defaults.SilentFails,
		wh.HardDefaults.SilentFails)
}

// GetType of the WebHook.
func (wh *WebHook) GetType() string {
	return util.FirstNonDefault(
		wh.Type,
		wh.Main.Type,
		wh.Defaults.Type,
		wh.HardDefaults.Type)
}

// GetURL of the WebHook.
func (wh *WebHook) GetURL() (url string) {
	url = strings.Clone(
		util.FirstNonDefaultWithEnv(
			wh.URL,
			wh.Main.URL,
			wh.Defaults.URL,
			wh.HardDefaults.URL))

	url = util.TemplateString(
		url,
		wh.ServiceStatus.GetServiceInfo())
	return
}
