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

//go:build unit

package webhook

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
)

func TestWebHook_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *bool
		want                                                 bool
	}{
		"root overrides all": {
			want:             true,
			rootValue:        test.BoolPtr(true),
			mainValue:        test.BoolPtr(false),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false),
		},
		"main overrides default+hardDefault": {
			want:             true,
			mainValue:        test.BoolPtr(true),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false),
		},
		"default overrides hardDefault": {
			want:             true,
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(false),
		},
		"hardDefaultValue is last resort": {
			want:             true,
			hardDefaultValue: test.BoolPtr(true),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.AllowInvalidCerts = tc.rootValue
			webhook.Main.AllowInvalidCerts = tc.mainValue
			webhook.Defaults.AllowInvalidCerts = tc.defaultValue
			webhook.HardDefaults.AllowInvalidCerts = tc.hardDefaultValue

			// WHEN GetAllowInvalidCerts is called.
			got := webhook.GetAllowInvalidCerts()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_GetDelay(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		"root overrides all": {
			want:             "1s",
			rootValue:        "1s",
			mainValue:        "2s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		"main overrides default+hardDefault": {
			want:             "1s",
			mainValue:        "1s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		"default overrides hardDefault": {
			want:             "1s",
			defaultValue:     "1s",
			hardDefaultValue: "2s",
		},
		"hardDefault is last resort": {
			want:             "1s",
			hardDefaultValue: "1s",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Delay = tc.rootValue
			webhook.Main.Delay = tc.mainValue
			webhook.Defaults.Delay = tc.defaultValue
			webhook.HardDefaults.Delay = tc.hardDefaultValue

			// WHEN GetDelay is called.
			got := webhook.GetDelay()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %s\ngot:  %s",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_GetDelayDuration(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 time.Duration
	}{
		"root overrides all": {
			want:             1 * time.Second,
			rootValue:        "1s",
			mainValue:        "2s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		"main overrides default+hardDefault": {
			want:             1 * time.Second,
			mainValue:        "1s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		"default overrides hardDefault": {
			want:             1 * time.Second,
			defaultValue:     "1s",
			hardDefaultValue: "2s",
		},
		"hardDefault is last resort": {
			want:             1 * time.Second,
			hardDefaultValue: "1s",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Delay = tc.rootValue
			webhook.Main.Delay = tc.mainValue
			webhook.Defaults.Delay = tc.defaultValue
			webhook.HardDefaults.Delay = tc.hardDefaultValue

			// WHEN GetDelayDuration is called.
			got := webhook.GetDelayDuration()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %s\ngot:  %s",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_GetDesiredStatusCode(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *uint16
		want                                                 uint16
	}{
		"root overrides all": {
			want:             1,
			rootValue:        test.UInt16Ptr(1),
			mainValue:        test.UInt16Ptr(2),
			defaultValue:     test.UInt16Ptr(2),
			hardDefaultValue: test.UInt16Ptr(2),
		},
		"main overrides default+hardDefault": {
			want:             1,
			mainValue:        test.UInt16Ptr(1),
			defaultValue:     test.UInt16Ptr(2),
			hardDefaultValue: test.UInt16Ptr(2),
		},
		"default overrides hardDefault": {
			want:             1,
			defaultValue:     test.UInt16Ptr(1),
			hardDefaultValue: test.UInt16Ptr(2),
		},
		"hardDefaultValue is last resort": {
			want:             1,
			hardDefaultValue: test.UInt16Ptr(1),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.DesiredStatusCode = tc.rootValue
			webhook.Main.DesiredStatusCode = tc.mainValue
			webhook.Defaults.DesiredStatusCode = tc.defaultValue
			webhook.HardDefaults.DesiredStatusCode = tc.hardDefaultValue

			// WHEN GetDesiredStatusCode is called.
			got := webhook.GetDesiredStatusCode()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %d\ngot:  %d",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_SetAndGetFail(t *testing.T) {
	// GIVEN an initial state, and a new state for the WebHook's failure state.
	tests := map[string]struct {
		initialState *bool
		newState     *bool
	}{
		"initial nil, set true": {
			initialState: nil,
			newState:     test.BoolPtr(true),
		},
		"initial false, set true": {
			initialState: test.BoolPtr(false),
			newState:     test.BoolPtr(true),
		},
		"initial true, set false": {
			initialState: test.BoolPtr(true),
			newState:     test.BoolPtr(false),
		},
		"initial nil, set nil": {
			initialState: nil,
			newState:     nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// AND a WebHook with this initial state.
			svcStatus := status.Status{}
			svcStatus.Init(
				0, 0, 1,
				packageName, name, "",
				&dashboard.Options{})
			wh := &WebHook{ID: t.Name()}
			wh.Init(
				&svcStatus,
				nil, nil, nil,
				nil,
				nil)
			wh.SetFail(tc.initialState)

			// WHEN we call SetFail with the new state.
			wh.SetFail(tc.newState)

			// THEN DidFail returns the expected value.
			got := wh.DidFail()
			gotStr := test.StringifyPtr(got)
			wantStr := test.StringifyPtr(tc.newState)
			if gotStr != wantStr {
				t.Errorf("%s\nFail mismatch\nwant: %s\ngot:  %s",
					packageName, wantStr, gotStr)
			}
		})
	}
}

func TestWebHook_SetAndGetNextRunnable(t *testing.T) {
	// GIVEN an initial state, and a new state for the WebHook's next runnable time.
	tests := map[string]struct {
		initialState time.Time
		newState     time.Time
	}{
		"date changed": {
			initialState: time.Time{},
			newState:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		"date not changed": {
			initialState: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			newState:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// AND a WebHook with this initial state.
			svcStatus := status.Status{}
			svcStatus.Init(
				0, 0, 1,
				packageName, name, "",
				&dashboard.Options{})
			wh := &WebHook{ID: t.Name()}
			wh.Init(
				&svcStatus,
				nil, nil, nil,
				nil,
				nil)
			wh.SetNextRunnable(tc.initialState)

			// WHEN we call SetNextRunnable with the new state.
			wh.SetNextRunnable(tc.newState)

			// THEN NextRunnable returns the expected value.
			got := wh.NextRunnable()
			if got != tc.newState {
				t.Errorf("%s\nFail mismatch\nwant: %s\ngot:  %s",
					packageName, tc.newState, got)
			}
		})
	}
}

func TestWebHook_GetMaxTries(t *testing.T) {
	// GIVEN a WebHook.
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *uint8
		want                                                 uint8
	}{
		"root overrides all": {
			want:             uint8(1),
			rootValue:        test.UInt8Ptr(1),
			mainValue:        test.UInt8Ptr(2),
			defaultValue:     test.UInt8Ptr(2),
			hardDefaultValue: test.UInt8Ptr(2),
		},
		"main overrides default+hardDefault": {
			want:             uint8(1),
			mainValue:        test.UInt8Ptr(1),
			defaultValue:     test.UInt8Ptr(2),
			hardDefaultValue: test.UInt8Ptr(2),
		},
		"default overrides hardDefault": {
			want:             uint8(1),
			defaultValue:     test.UInt8Ptr(1),
			hardDefaultValue: test.UInt8Ptr(2),
		},
		"hardDefaultValue is last resort": {
			want:             uint8(1),
			hardDefaultValue: test.UInt8Ptr(1),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.MaxTries = tc.rootValue
			webhook.Main.MaxTries = tc.mainValue
			webhook.Defaults.MaxTries = tc.defaultValue
			webhook.HardDefaults.MaxTries = tc.hardDefaultValue

			// WHEN GetMaxTries is called.
			got := webhook.GetMaxTries()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %d\ngot:  %d",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_BuildRequest(t *testing.T) {
	// GIVEN a WebHook and a HTTP Request.
	tests := map[string]struct {
		webhookType   string
		url           string
		customHeaders Headers
		wantNil       bool
	}{
		"valid github type": {
			webhookType: "github",
			url:         "release-argus/Argus",
		},
		"catch invalid github request": {
			webhookType: "github",
			url:         "release-argus	/	Argus",
			wantNil:     true,
		},
		"valid gitlab type": {
			webhookType: "gitlab",
			url:         "https://release-argus.io",
		},
		"catch invalid gitlab request": {
			webhookType: "gitlab",
			url:         "release-argus	/	Argus",
			wantNil:     true,
		},
		"sets custom headers for github": {
			webhookType: "github",
			url:         "release-argus/Argus",
			customHeaders: Headers{
				{Key: "X-Foo", Value: "bar"}},
		},
		"sets custom headers for gitlab": {
			webhookType: "gitlab",
			url:         "https://release-argus.io",
			customHeaders: Headers{
				{Key: "X-Foo", Value: "bar"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Type = tc.webhookType
			webhook.URL = tc.url
			webhook.CustomHeaders = &tc.customHeaders

			// WHEN BuildRequest is called.
			req := webhook.BuildRequest()

			// THEN the function returns the correct result.
			if tc.wantNil {
				if req != nil {
					t.Fatalf("%s\nexpected request to fail with url %q",
						packageName, tc.url)
				}
				return
			}
			switch tc.webhookType {
			case "github":
				// Payload.
				body, _ := io.ReadAll(req.Body)
				var payload GitHub
				_ = json.Unmarshal(body, &payload)
				want := "refs/heads/master"
				if payload.Ref != want {
					t.Errorf("%s\npayload mismatch\nwant: %q\ngot:  %q (%+v)",
						packageName, want, payload.Ref, payload)
				}
				// Content-Type.
				want = "application/json"
				if req.Header["Content-Type"][0] != want {
					t.Errorf("%s\nContent-Type mismatch\nwant: %q\ngot:  %q (%+v)",
						packageName, want, req.Header["Content-Type"][0], req.Header)
				}
				// X-Github-Event.
				want = "push"
				if req.Header["X-Github-Event"][0] != want {
					t.Errorf("%s\nGitHub headers weren't set?\nwant: %q in 'X-Github-Event'\ngot:  %v",
						packageName, want, req.Header["X-Github-Event"])
				}
			case "gitlab":
				// Content-Type.
				want := "application/x-www-form-urlencoded"
				if req.Header["Content-Type"][0] != want {
					t.Errorf("%s\nContent-Type mismatch\nwant: %q\ngot:  %q (%+v)",
						packageName, want, req.Header["Content-Type"][0], req.Header)
				}
			}
			// Custom Headers.
			for _, header := range tc.customHeaders {
				if len(req.Header[header.Key]) == 0 {
					t.Fatalf("%s\nCustom Headers not set\n%+v",
						packageName, req.Header)
				}
				if req.Header[header.Key][0] != header.Value {
					t.Fatalf("%s\nCustom Headers mismatch\nwant: %q\ngot:  %q (%+v)",
						packageName, header.Value, req.Header[header.Key][0], req.Header)
				}
			}
		})
	}
}

func TestWebHook_GetType(t *testing.T) {
	// GIVEN a WebHook with Type in various locations.
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		"root overrides all": {
			want:             "github",
			rootValue:        "github",
			mainValue:        "url",
			defaultValue:     "url",
			hardDefaultValue: "url",
		},
		"main overrides default+hardDefault": {
			want:             "github",
			mainValue:        "github",
			defaultValue:     "url",
			hardDefaultValue: "url",
		},
		"default overrides hardDefault": {
			want:             "github",
			defaultValue:     "github",
			hardDefaultValue: "url",
		},
		"hardDefaultValue is last resort": {
			want:             "github",
			hardDefaultValue: "github",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Type = tc.rootValue
			webhook.Main.Type = tc.mainValue
			webhook.Defaults.Type = tc.defaultValue
			webhook.HardDefaults.Type = tc.hardDefaultValue

			// WHEN GetType is called.
			got := webhook.GetType()

			// THEN the function returns the correct type.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_GetSecret(t *testing.T) {
	// GIVEN a WebHook with Secret in various locations.
	tests := map[string]struct {
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		"root overrides all": {
			want:             "argus-secret",
			rootValue:        "argus-secret",
			mainValue:        "unused",
			defaultValue:     "unused",
			hardDefaultValue: "unused",
		},
		"main overrides default+hardDefault": {
			want:             "argus-secret",
			mainValue:        "argus-secret",
			defaultValue:     "unused",
			hardDefaultValue: "unused",
		},
		"default overrides hardDefault": {
			want:             "argus-secret",
			defaultValue:     "argus-secret",
			hardDefaultValue: "unused",
		},
		"hardDefaultValue last resort": {
			want:             "argus-secret",
			hardDefaultValue: "argus-secret",
		},
		"env var is used": {
			want:      "argus-secret",
			env:       map[string]string{"TEST_WEBHOOK__GET_SECRET__ONE": "argus-secret"},
			rootValue: "${TEST_WEBHOOK__GET_SECRET__ONE}",
		},
		"env var partial is used": {
			want:      "argus-secret-two",
			env:       map[string]string{"TEST_WEBHOOK__GET_SECRET__TWO": "argus-secret"},
			rootValue: "${TEST_WEBHOOK__GET_SECRET__TWO}-two",
		},
		"empty env var is ignored": {
			want:         "argus-secret",
			env:          map[string]string{"TEST_WEBHOOK__GET_SECRET__THREE": ""},
			rootValue:    "${TEST_WEBHOOK__GET_SECRET__THREE}",
			defaultValue: "argus-secret",
		},
		"unset env var is used": {
			want:         "${TEST_WEBHOOK__GET_SECRET__UNSET}",
			rootValue:    "${TEST_WEBHOOK__GET_SECRET__UNSET}",
			defaultValue: "argus-secret",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}
			webhook := testWebHook(true, false, false)
			webhook.Secret = tc.rootValue
			webhook.Main.Secret = tc.mainValue
			webhook.Defaults.Secret = tc.defaultValue
			webhook.HardDefaults.Secret = tc.hardDefaultValue

			// WHEN GetSecret is called.
			got := webhook.GetSecret()

			// THEN the function returns the correct secret.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_GetSilentFails(t *testing.T) {
	// GIVEN a WebHook with SilentFails in various locations.
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *bool
		want                                                 bool
	}{
		"root overrides all": {
			want:             true,
			rootValue:        test.BoolPtr(true),
			mainValue:        test.BoolPtr(false),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false),
		},
		"main overrides default+hardDefault": {
			want:             true,
			mainValue:        test.BoolPtr(true),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false),
		},
		"default overrides hardDefault": {
			want:             true,
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(false),
		},
		"hardDefaultValue is last resort": {
			want:             true,
			hardDefaultValue: test.BoolPtr(true),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.SilentFails = tc.rootValue
			webhook.Main.SilentFails = tc.mainValue
			webhook.Defaults.SilentFails = tc.defaultValue
			webhook.HardDefaults.SilentFails = tc.hardDefaultValue

			// WHEN GetSilentFails is called.
			got := webhook.GetSilentFails()

			// THEN the function returns the correct boolean.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_GetURL(t *testing.T) {
	// GIVEN a WebHook with urls in various locations.
	tests := map[string]struct {
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
		latestVersion                                        string
	}{
		"root overrides all": {
			want:             "https://release-argus.io",
			rootValue:        "https://release-argus.io",
			mainValue:        "https://example.com",
			defaultValue:     "https://example.com",
			hardDefaultValue: "https://example.com",
		},
		"main overrides default+hardDefault": {
			want:             "https://release-argus.io",
			mainValue:        "https://release-argus.io",
			defaultValue:     "https://example.com",
			hardDefaultValue: "https://example.com",
		},
		"default is last resort": {
			want:             "https://release-argus.io",
			defaultValue:     "https://release-argus.io",
			hardDefaultValue: "https://example.com",
		},
		"hardDefaultValue last resort": {
			want:             "https://release-argus.io",
			hardDefaultValue: "https://release-argus.io",
		},
		"uses latest_version": {
			want:          "https://release-argus.io/1.2.3",
			rootValue:     "https://release-argus.io/{{ version }}",
			latestVersion: "1.2.3",
		},
		"empty version when not found": {
			want:      "https://release-argus.io/",
			rootValue: "https://release-argus.io/{{ version }}",
		},
		"env var is used": {
			want:      "https://release-argus.io",
			env:       map[string]string{"TEST_WEBHOOK__GET_URL__ONE": "https://release-argus.io"},
			rootValue: "${TEST_WEBHOOK__GET_URL__ONE}",
		},
		"env var partial is used": {
			want:      "https://release-argus.io",
			env:       map[string]string{"TEST_WEBHOOK__GET_URL__TWO": "release-argus"},
			rootValue: "https://${TEST_WEBHOOK__GET_URL__TWO}.io",
		},
		"empty env var is ignored": {
			want:         "https://release-argus.io",
			env:          map[string]string{"TEST_WEBHOOK__GET_URL__THREE": ""},
			rootValue:    "${TEST_WEBHOOK__GET_URL__THREE}",
			defaultValue: "https://release-argus.io",
		},
		"undefined env var is used": {
			want:         "${TEST_WEBHOOK__GET_URL__UNSET}",
			rootValue:    "${TEST_WEBHOOK__GET_URL__UNSET}",
			defaultValue: "https://release-argus.io",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}
			webhook := testWebHook(true, false, false)
			webhook.URL = tc.rootValue
			webhook.Main.URL = tc.mainValue
			webhook.Defaults.URL = tc.defaultValue
			webhook.HardDefaults.URL = tc.hardDefaultValue
			webhook.ServiceStatus.SetLatestVersion(tc.latestVersion, "", false)

			// WHEN GetURL is called.
			got := webhook.GetURL()

			// THEN the function returns the url.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_GetIsRunnable(t *testing.T) {
	// GIVEN a WebHook with a NextRunnable time.
	tests := map[string]struct {
		nextRunnable time.Time
		want         bool
	}{
		"default time is runnable": {
			want: true},
		"nextRunnable now is runnable": {
			want: true, nextRunnable: time.Now().UTC()},
		"nextRunnable in the future isn't runnable": {
			want: false, nextRunnable: time.Now().UTC().Add(time.Minute)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Failed.SetNextRunnable(webhook.ID, tc.nextRunnable)
			time.Sleep(time.Nanosecond)

			// WHEN GetIsRunnable is called.
			got := webhook.IsRunnable()

			// THEN the function returns whether the webhook is runnable now.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestWebHook_SetExecuting(t *testing.T) {
	// GIVEN a WebHook in different fail states.
	tests := map[string]struct {
		failed         *bool
		timeDifference time.Duration
		addDelay       bool
		delay          string
		sending        bool
		maxTries       int
	}{
		"sending does delay by 1h15s": {
			timeDifference: time.Hour + 15*time.Second,
			failed:         nil,
			sending:        true,
		},
		"sending with delay does delay by delay+1h15s": {
			timeDifference: time.Hour + 30*time.Minute + 15*time.Second,
			failed:         nil,
			sending:        true,
			addDelay:       true,
			delay:          "30m",
		},
		"sending with maxTries 10 and delay does delay by delay+1h+15s": {
			timeDifference: time.Hour + 30*time.Minute + 15*time.Second,
			failed:         nil,
			sending:        true,
			addDelay:       true,
			delay:          "30m",
			maxTries:       10,
		},
		"not tried (failed=nil) does delay by 15s": {
			timeDifference: 15 * time.Second,
			failed:         nil,
		},
		"failed (failed=true) does delay by 15s": {
			timeDifference: 15 * time.Second,
			failed:         test.BoolPtr(true),
		},
		"success (failed=false) does delay by 2*Interval": {
			timeDifference: 24 * time.Minute,
			failed:         test.BoolPtr(false),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Failed.Set(webhook.ID, tc.failed)
			webhook.Delay = tc.delay
			webhook.MaxTries = test.UInt8Ptr(tc.maxTries)

			// WHEN SetExecuting is run.
			webhook.SetExecuting(tc.addDelay, tc.sending)

			// THEN the correct response is received,
			// and next runnable is within expected range.
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := webhook.NextRunnable()
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf("%s\nran at\n%s\nwant between:\n%s and\n%s\ngot\n%s",
					packageName,
					now,
					minTime, maxTime,
					gotTime)
			}
		})
	}
}
