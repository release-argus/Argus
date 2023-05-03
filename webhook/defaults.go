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

// SetDefaults for WebHooks.
func (w *WebHookDefaults) SetDefaults() {
	// type
	w.Type = "github"
	// delay
	w.Delay = "0s"
	// allow_invalid_certs
	webhookAllowInvalidCerts := false
	w.AllowInvalidCerts = &webhookAllowInvalidCerts
	// desired_status_code
	webhookDesiredStatusCode := 0
	w.DesiredStatusCode = &webhookDesiredStatusCode
	// max_tries
	webhookMaxTries := uint(3)
	w.MaxTries = &webhookMaxTries
	// silent_fails
	webhookSilentFails := false
	w.SilentFails = &webhookSilentFails
}
