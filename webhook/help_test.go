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

//go:build testing

package webhook

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}

func testWebHookSuccessful() WebHook {
	whID := "test"
	whType := "github"
	whURL := "https://httpbin.org/anything"
	whSecret := "secret"
	whAllowInvalidCerts := false
	whDesiredStatusCode := 0
	whDelay := "0s"
	whSilentFails := true
	whMaxTries := uint(1)
	parentInterval := "12m"
	return WebHook{
		ID:                &whID,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		Delay:             &whDelay,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		Main:              &WebHook{},
		Defaults:          &WebHook{},
		HardDefaults:      &WebHook{},
		ParentInterval:    &parentInterval,
	}
}

func testWebHookFailing() WebHook {
	whID := "test"
	whType := "github"
	whURL := "https://httpbin.org/hidden-basic-auth/:user/:passwd"
	whSecret := "secret"
	whAllowInvalidCerts := false
	whDesiredStatusCode := 0
	whSilentFails := true
	whMaxTries := uint(1)
	parentInterval := "12m"
	return WebHook{
		ID:                &whID,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		Main:              &WebHook{},
		Defaults:          &WebHook{},
		HardDefaults:      &WebHook{},
		Notifiers:         &Notifiers{},
		ParentInterval:    &parentInterval,
	}
}
