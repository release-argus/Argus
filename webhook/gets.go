// Copyright [2026] [Argus]
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
	"net/http"
	"strings"
	"time"

	"github.com/release-argus/Argus/util"
)

// BuildRequest builds and returns a *http.Request for the WebHook.
func (w *WebHook) BuildRequest() (req *http.Request) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var err error
	switch w.GetType() {
	case "github":
		payload, err := marshalWebhookPayload(
			GitHub{
				Ref:    "refs/heads/master",
				Before: util.RandAlphaNumericLower(40),
				After:  util.RandAlphaNumericLower(40),
			},
		)
		if err != nil {
			return nil
		}

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
	w.setHeaders(req)
	return
}

// GetAllowInvalidCerts resolves whether invalid HTTPS certs are allowed.
func (w *WebHook) GetAllowInvalidCerts() bool {
	return *util.FirstNonNilPtr(
		w.AllowInvalidCerts,
		w.Main.AllowInvalidCerts,
		w.Defaults.AllowInvalidCerts,
		w.HardDefaults.AllowInvalidCerts,
	)
}

// GetDelay resolves the delay to use before auto-approving the WebHook.
func (w *WebHook) GetDelay() string {
	return util.FirstNonDefault(
		w.Delay,
		w.Main.Delay,
		w.Defaults.Delay,
		w.HardDefaults.Delay,
	)
}

// GetDelayDuration resolves the auto-approve delay as a time.Duration.
func (w *WebHook) GetDelayDuration() (duration time.Duration) {
	duration, _ = time.ParseDuration(w.GetDelay())
	return duration
}

// GetDesiredStatusCode resolves the desired status code of WebHook requests.
func (w *WebHook) GetDesiredStatusCode() uint16 {
	return *util.FirstNonNilPtr(
		w.DesiredStatusCode,
		w.Main.DesiredStatusCode,
		w.Defaults.DesiredStatusCode,
		w.HardDefaults.DesiredStatusCode,
	)
}

// DidFail resolves whether the last send of this WebHook failed.
func (w *WebHook) DidFail() *bool {
	return w.Failed.Get(w.ID)
}

// SetFail sets whether the last send attempt failed.
func (w *WebHook) SetFail(state *bool) {
	w.Failed.Set(w.ID, state)
}

// NextRunnable resolves the time this WebHook can next run.
func (w *WebHook) NextRunnable() time.Time {
	return w.Failed.NextRunnable(w.ID)
}

// SetNextRunnable sets when the WebHook may run again.
func (w *WebHook) SetNextRunnable(time time.Time) {
	w.Failed.SetNextRunnable(w.ID, time)
}

// SetExecuting sets the next-runnable time based on the outcome, optionally adding the send delay or blocking for a pending response.
func (w *WebHook) SetExecuting(addDelay bool, received bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var nextRunnable time.Time

	// Different times depending on pass/fail.
	// pass.
	if !util.DerefOr(w.DidFail(), true) {
		parentInterval, _ := time.ParseDuration(*w.ParentInterval)
		nextRunnable = time.Now().UTC().Add(2 * parentInterval)
		// fail/nil.
	} else {
		nextRunnable = time.Now().UTC().Add(15 * time.Second)
	}

	// Block for delay.
	if addDelay {
		nextRunnable = nextRunnable.Add(w.GetDelayDuration())
	}

	// Block reruns whilst waiting for a response.
	if received {
		nextRunnable = nextRunnable.Add(time.Hour)
	}
	w.SetNextRunnable(nextRunnable)
}

// GetMaxTries resolves the maximum number of send attempts allowed.
func (w *WebHook) GetMaxTries() uint8 {
	return *util.FirstNonNilPtr(
		w.MaxTries,
		w.Main.MaxTries,
		w.Defaults.MaxTries,
		w.HardDefaults.MaxTries,
	)
}

// IsRunnable reports whether the WebHook can run now.
func (w *WebHook) IsRunnable() bool {
	return time.Now().UTC().After(w.NextRunnable())
}

// GetSecret resolves the WebHook secret.
func (w *WebHook) GetSecret() string {
	return util.FirstNonDefaultWithEnv(
		w.Secret,
		w.Main.Secret,
		w.Defaults.Secret,
		w.HardDefaults.Secret,
	)
}

// GetSilentFails resolves the flag for whether WebHooks should fail silently or notify.
func (w *WebHook) GetSilentFails() bool {
	return *util.FirstNonNilPtr(
		w.SilentFails,
		w.Main.SilentFails,
		w.Defaults.SilentFails,
		w.HardDefaults.SilentFails,
	)
}

// GetType resolves the type of the WebHook.
func (w *WebHook) GetType() string {
	return util.FirstNonDefault(
		w.Type,
		w.Main.Type,
		w.Defaults.Type,
		w.HardDefaults.Type,
	)
}

// GetURL resolves the URL of the WebHook.
func (w *WebHook) GetURL() (url string) {
	url = strings.Clone(
		util.FirstNonDefaultWithEnv(
			w.URL,
			w.Main.URL,
			w.Defaults.URL,
			w.HardDefaults.URL,
		),
	)

	url = util.TemplateString(url, w.ServiceStatus.GetServiceInfo())
	return
}
