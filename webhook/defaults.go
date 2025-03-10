// Copyright [2024] [Argus]
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

// Default sets this Defaults to the default values.
func (d *Defaults) Default() {
	// type
	d.Type = "github"
	// delay
	d.Delay = "0s"
	// allow_invalid_certs
	webhookAllowInvalidCerts := false
	d.AllowInvalidCerts = &webhookAllowInvalidCerts
	// desired_status_code
	webhookDesiredStatusCode := uint16(0)
	d.DesiredStatusCode = &webhookDesiredStatusCode
	// max_tries
	webhookMaxTries := uint8(3)
	d.MaxTries = &webhookMaxTries
	// silent_fails
	webhookSilentFails := false
	d.SilentFails = &webhookSilentFails
}
