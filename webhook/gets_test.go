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

//go:build unit

package webhook

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

func TestWebHook_BuildRequest(t *testing.T) {
	// GIVEN: a WebHook and a HTTP Request.
	tests := []struct {
		name        string
		webhookType string
		url         string
		headers     Headers
		wantNil     bool
	}{
		{
			name:        "valid github type",
			webhookType: "github",
			url:         test.ArgusGitHubRepo,
		},
		{
			name:        "catch invalid github request",
			webhookType: "github",
			url:         "release-argus	/	Argus",
			wantNil:     true,
		},
		{
			name:        "valid gitlab type",
			webhookType: "gitlab",
			url:         "https://release-argus.io",
		},
		{
			name:        "catch invalid gitlab request",
			webhookType: "gitlab",
			url:         "release-argus	/	Argus",
			wantNil:     true,
		},
		{
			name:        "sets custom headers for github",
			webhookType: "github",
			url:         test.ArgusGitHubRepo,
			headers: Headers{
				{Key: "X-Foo", Value: "bar"},
			},
		},
		{
			name:        "sets custom headers for gitlab",
			webhookType: "gitlab",
			url:         "https://release-argus.io",
			headers: Headers{
				{Key: "X-Foo", Value: "bar"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Type = tc.webhookType
			webhook.URL = tc.url
			webhook.Headers = tc.headers

			// WHEN: BuildRequest is called.
			req := webhook.BuildRequest()

			prefix := fmt.Sprintf("%s\nWebHook.BuildRequest()", packageName)

			// THEN: the function returns the correct result.
			if tc.wantNil {
				if req != nil {
					t.Fatalf(
						"%s expected request to fail with url %q",
						packageName, tc.url,
					)
				}
				return
			}
			var fieldTests []test.FieldAssertion
			switch tc.webhookType {
			case "github":
				// Payload.
				body, _ := io.ReadAll(req.Body)
				var payload GitHub
				_ = decode.Unmarshal("json", body, &payload)
				fieldTests = []test.FieldAssertion{
					{Name: "payload.Ref", Got: payload.Ref, Want: "refs/heads/master", Target: "request.Body", Mode: test.CompareEqual},
					{Name: "Header['content-type']", Got: req.Header["Content-Type"][0], Want: "application/json", Mode: test.CompareEqual},
					{Name: "Header['X-Github-Event']", Got: req.Header["X-Github-Event"][0], Want: "push", Mode: test.CompareEqual},
				}
			case "gitlab":
				fieldTests = []test.FieldAssertion{
					{Name: "Header['Content-Type']", Got: req.Header["Content-Type"][0], Want: "application/x-www-form-urlencoded", Mode: test.CompareEqual},
				}
			}
			if err := test.AssertFields(t, fieldTests, prefix, "request"); err != nil {
				t.Fatal(err)
			}
			// Custom Headers.
			for _, header := range tc.headers {
				if len(req.Header[header.Key]) == 0 {
					t.Fatalf(
						"%s didn't set Headers[%q]\ngot: %+v",
						packageName, header.Key,
						req.Header,
					)
				}
				if got := req.Header[header.Key][0]; got != header.Value {
					t.Fatalf(
						"%s Headers[%q] mismatch\ngot:  %q (%+v)\nwant: %q",
						packageName, header.Key,
						got, req.Header,
						header.Value,
					)
				}
			}
		})
	}
}

func TestWebHook_BuildRequest__MarshalError(t *testing.T) {
	// GIVEN: a failing marshal function.
	original := marshalWebhookPayload
	marshalWebhookPayload = func(v any) ([]byte, error) {
		return nil, fmt.Errorf("marshal failed")
	}
	t.Cleanup(func() { marshalWebhookPayload = original })

	// AND: a WebHook.
	webhook := testWebHook(true, false, false)
	webhook.Type = "github"
	webhook.URL = test.ArgusGitHubRepo

	// WHEN: BuildRequest is called.
	req := webhook.BuildRequest()

	prefix := fmt.Sprintf("%s\nWebHook.BuildRequest(marshal error)", packageName)

	// THEN: the request is nil.
	if req != nil {
		t.Errorf(
			"%s expected nil request\ngot:  %+v\nwant: nil",
			prefix, req,
		)
	}
}

func TestWebHook_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *bool
		want                                                 bool
	}{
		{
			name:             "root overrides all",
			want:             true,
			rootValue:        test.Ptr(true),
			mainValue:        test.Ptr(false),
			defaultValue:     test.Ptr(false),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "main overrides default+hardDefault",
			want:             true,
			mainValue:        test.Ptr(true),
			defaultValue:     test.Ptr(false),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "default overrides hardDefault",
			want:             true,
			defaultValue:     test.Ptr(true),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             true,
			hardDefaultValue: test.Ptr(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.AllowInvalidCerts = tc.rootValue
			webhook.Main.AllowInvalidCerts = tc.mainValue
			webhook.Defaults.AllowInvalidCerts = tc.defaultValue
			webhook.HardDefaults.AllowInvalidCerts = tc.hardDefaultValue

			// WHEN: GetAllowInvalidCerts is called.
			got := webhook.GetAllowInvalidCerts()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetAllowInvalidCerts() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_GetDelay(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "1s",
			rootValue:        "1s",
			mainValue:        "2s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		{
			name:             "main overrides default+hardDefault",
			want:             "1s",
			mainValue:        "1s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		{
			name:             "default overrides hardDefault",
			want:             "1s",
			defaultValue:     "1s",
			hardDefaultValue: "2s",
		},
		{
			name:             "hardDefault is last resort",
			want:             "1s",
			hardDefaultValue: "1s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Delay = tc.rootValue
			webhook.Main.Delay = tc.mainValue
			webhook.Defaults.Delay = tc.defaultValue
			webhook.HardDefaults.Delay = tc.hardDefaultValue

			// WHEN: GetDelay is called.
			got := webhook.GetDelay()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetDelay() value mismatch\ngot:  %s\nwant: %s",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_GetDelayDuration(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 time.Duration
	}{
		{
			name:             "root overrides all",
			want:             1 * time.Second,
			rootValue:        "1s",
			mainValue:        "2s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		{
			name:             "main overrides default+hardDefault",
			want:             1 * time.Second,
			mainValue:        "1s",
			defaultValue:     "2s",
			hardDefaultValue: "2s",
		},
		{
			name:             "default overrides hardDefault",
			want:             1 * time.Second,
			defaultValue:     "1s",
			hardDefaultValue: "2s",
		},
		{
			name:             "hardDefault is last resort",
			want:             1 * time.Second,
			hardDefaultValue: "1s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Delay = tc.rootValue
			webhook.Main.Delay = tc.mainValue
			webhook.Defaults.Delay = tc.defaultValue
			webhook.HardDefaults.Delay = tc.hardDefaultValue

			// WHEN: GetDelayDuration is called.
			got := webhook.GetDelayDuration()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetDelayDuration() value mismatch\ngot:  %s\nwant: %s",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_GetDesiredStatusCode(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *uint16
		want                                                 uint16
	}{
		{
			name:             "root overrides all",
			want:             1,
			rootValue:        test.Ptr[uint16](1),
			mainValue:        test.Ptr[uint16](2),
			defaultValue:     test.Ptr[uint16](2),
			hardDefaultValue: test.Ptr[uint16](2),
		},
		{
			name:             "main overrides default+hardDefault",
			want:             1,
			mainValue:        test.Ptr[uint16](1),
			defaultValue:     test.Ptr[uint16](2),
			hardDefaultValue: test.Ptr[uint16](2),
		},
		{
			name:             "default overrides hardDefault",
			want:             1,
			defaultValue:     test.Ptr[uint16](1),
			hardDefaultValue: test.Ptr[uint16](2),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             1,
			hardDefaultValue: test.Ptr[uint16](1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.DesiredStatusCode = tc.rootValue
			webhook.Main.DesiredStatusCode = tc.mainValue
			webhook.Defaults.DesiredStatusCode = tc.defaultValue
			webhook.HardDefaults.DesiredStatusCode = tc.hardDefaultValue

			// WHEN: GetDesiredStatusCode is called.
			got := webhook.GetDesiredStatusCode()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetDesiredStatusCode() value mismatch\ngot:  %d\nwant: %d",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_SetAndGetFail(t *testing.T) {
	whCfg := plainConfig(t)
	// GIVEN: an initial state, and a new state for the WebHook's failure state.
	tests := []struct {
		name         string
		initialState *bool
		newState     *bool
	}{
		{
			name:         "initial nil, set true",
			initialState: nil,
			newState:     test.Ptr(true),
		},
		{
			name:         "initial false, set true",
			initialState: test.Ptr(false),
			newState:     test.Ptr(true),
		},
		{
			name:         "initial true, set false",
			initialState: test.Ptr(true),
			newState:     test.Ptr(false),
		},
		{
			name:         "initial nil, set nil",
			initialState: nil,
			newState:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a WebHook with this initial state.
			svcStatus := status.Status{}
			svcStatus.Init(
				0, 0, 1,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			wh := &WebHook{ID: t.Name()}
			wh.init(
				&svcStatus,
				nil, whCfg, nil,
				nil,
			)
			wh.SetFail(tc.initialState)

			// WHEN: we call SetFail with the new state.
			wh.SetFail(tc.newState)

			// THEN: DidFail returns the expected value.
			didFail := wh.DidFail()
			want := test.StringifyPtr(tc.newState)
			if got := test.StringifyPtr(didFail); got != want {
				t.Errorf(
					"%s\nWebHook.DidFail() value mismatch\ngot:  %s\nwant: %s",
					packageName, got, want,
				)
			}
		})
	}
}

func TestWebHook_SetAndGetNextRunnable(t *testing.T) {
	whCfg := plainConfig(t)
	// GIVEN: an initial state, and a new state for the WebHook's next runnable time.
	tests := []struct {
		name         string
		initialState time.Time
		newState     time.Time
	}{
		{
			name:         "date changed",
			initialState: time.Time{},
			newState:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "date not changed",
			initialState: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			newState:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a WebHook with this initial state.
			svcStatus := status.Status{}
			svcStatus.Init(
				0, 0, 1,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			wh := &WebHook{ID: t.Name()}
			wh.init(
				&svcStatus,
				nil, whCfg, nil,
				nil,
			)
			wh.SetNextRunnable(tc.initialState)

			// WHEN: we call SetNextRunnable with the new state.
			wh.SetNextRunnable(tc.newState)

			// THEN: NextRunnable returns the expected value.
			got := wh.NextRunnable()
			if got != tc.newState {
				t.Errorf(
					"%s\nWebHook.NextRunnable() value mismatch\ngot:  %s\nwant: %s",
					packageName, got, tc.newState,
				)
			}
		})
	}
}

func TestWebHook_SetExecuting(t *testing.T) {
	// GIVEN: a WebHook in different fail states.
	tests := []struct {
		name           string
		failed         *bool
		timeDifference time.Duration
		addDelay       bool
		delay          string
		sending        bool
		maxTries       uint8
	}{
		{
			name:           "sending does delay by 1h15s",
			timeDifference: time.Hour + 15*time.Second,
			failed:         nil,
			sending:        true,
		},
		{
			name:           "sending with delay does delay by delay+1h15s",
			timeDifference: time.Hour + 30*time.Minute + 15*time.Second,
			failed:         nil,
			sending:        true,
			addDelay:       true,
			delay:          "30m",
		},
		{
			name:           "sending with maxTries 10 and delay does delay by delay+1h+15s",
			timeDifference: time.Hour + 30*time.Minute + 15*time.Second,
			failed:         nil,
			sending:        true,
			addDelay:       true,
			delay:          "30m",
			maxTries:       10,
		},
		{
			name:           "not tried (failed=nil) does delay by 15s",
			timeDifference: 15 * time.Second,
			failed:         nil,
		},
		{
			name:           "failed (failed=true) does delay by 15s",
			timeDifference: 15 * time.Second,
			failed:         test.Ptr(true),
		},
		{
			name:           "success (failed=false) does delay by 2*Interval",
			timeDifference: 24 * time.Minute,
			failed:         test.Ptr(false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Failed.Set(webhook.ID, tc.failed)
			webhook.Delay = tc.delay
			webhook.MaxTries = test.Ptr(tc.maxTries)

			// WHEN: SetExecuting is run.
			webhook.SetExecuting(tc.addDelay, tc.sending)

			// THEN: the correct response is received,
			// and next runnable is within expected range.
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := webhook.NextRunnable()
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf(
					"%s\nWebHook.SetExecuting() ran at:\n%s\ngot\n%s\nwant between:\n%s and\n%s",
					packageName,
					now,
					gotTime,
					minTime, maxTime,
				)
			}
		})
	}
}

func TestWebHook_GetMaxTries(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *uint8
		want                                                 uint8
	}{
		{
			name:             "root overrides all",
			want:             uint8(1),
			rootValue:        test.Ptr[uint8](1),
			mainValue:        test.Ptr[uint8](2),
			defaultValue:     test.Ptr[uint8](2),
			hardDefaultValue: test.Ptr[uint8](2),
		},
		{
			name:             "main overrides default+hardDefault",
			want:             uint8(1),
			mainValue:        test.Ptr[uint8](1),
			defaultValue:     test.Ptr[uint8](2),
			hardDefaultValue: test.Ptr[uint8](2),
		},
		{
			name:             "default overrides hardDefault",
			want:             uint8(1),
			defaultValue:     test.Ptr[uint8](1),
			hardDefaultValue: test.Ptr[uint8](2),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             uint8(1),
			hardDefaultValue: test.Ptr[uint8](1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.MaxTries = tc.rootValue
			webhook.Main.MaxTries = tc.mainValue
			webhook.Defaults.MaxTries = tc.defaultValue
			webhook.HardDefaults.MaxTries = tc.hardDefaultValue

			// WHEN: GetMaxTries is called.
			got := webhook.GetMaxTries()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetMaxTries() value mismatch\ngot:  %d\nwant: %d",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_IsRunnable(t *testing.T) {
	// GIVEN: a WebHook with a NextRunnable time.
	tests := []struct {
		name         string
		nextRunnable time.Time
		want         bool
	}{
		{
			name: "default time is runnable",
			want: true,
		},
		{
			name:         "nextRunnable now is runnable",
			want:         true,
			nextRunnable: time.Now().UTC(),
		},
		{
			name:         "nextRunnable in the future isn't runnable",
			want:         false,
			nextRunnable: time.Now().UTC().Add(time.Minute),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Failed.SetNextRunnable(webhook.ID, tc.nextRunnable)
			time.Sleep(time.Nanosecond)

			// WHEN: GetIsRunnable is called.
			got := webhook.IsRunnable()

			// THEN: the function returns whether the defaults is runnable now.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.IsRunnable() mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_GetSecret(t *testing.T) {
	// GIVEN: a WebHook with Secret in various locations.
	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "argus-secret",
			rootValue:        "argus-secret",
			mainValue:        "unused",
			defaultValue:     "unused",
			hardDefaultValue: "unused",
		},
		{
			name:             "main overrides default+hardDefault",
			want:             "argus-secret",
			mainValue:        "argus-secret",
			defaultValue:     "unused",
			hardDefaultValue: "unused",
		},
		{
			name:             "default overrides hardDefault",
			want:             "argus-secret",
			defaultValue:     "argus-secret",
			hardDefaultValue: "unused",
		},
		{
			name:             "hardDefaultValue last resort",
			want:             "argus-secret",
			hardDefaultValue: "argus-secret",
		},
		{
			name: "env var is used",
			want: "argus-secret",
			env: map[string]string{
				"TEST_WEBHOOK__GET_SECRET__ONE": "argus-secret",
			},
			rootValue: "${TEST_WEBHOOK__GET_SECRET__ONE}",
		},
		{
			want: "argus-secret",
			env: map[string]string{
				"TEST_WEBHOOK__GET_SECRET__ONE": "argus-secret",
			},
			rootValue: "${TEST_WEBHOOK__GET_SECRET__ONE}",
		},
		{
			name: "env var partial is used",
			want: "argus-secret-two",
			env: map[string]string{
				"TEST_WEBHOOK__GET_SECRET__TWO": "argus-secret",
			},
			rootValue: "${TEST_WEBHOOK__GET_SECRET__TWO}-two",
		},
		{
			name: "empty env var is ignored",
			want: "argus-secret",
			env: map[string]string{
				"TEST_WEBHOOK__GET_SECRET__THREE": "",
			},
			rootValue:    "${TEST_WEBHOOK__GET_SECRET__THREE}",
			defaultValue: "argus-secret",
		},
		{
			name:         "unset env var is used",
			want:         "${TEST_WEBHOOK__GET_SECRET__UNSET}",
			rootValue:    "${TEST_WEBHOOK__GET_SECRET__UNSET}",
			defaultValue: "argus-secret",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)
			webhook := testWebHook(true, false, false)
			webhook.Secret = tc.rootValue
			webhook.Main.Secret = tc.mainValue
			webhook.Defaults.Secret = tc.defaultValue
			webhook.HardDefaults.Secret = tc.hardDefaultValue

			// WHEN: GetSecret is called.
			got := webhook.GetSecret()

			// THEN: the function returns the correct secret.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetSecret() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_GetSilentFails(t *testing.T) {
	// GIVEN: a WebHook with SilentFails in various locations.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *bool
		want                                                 bool
	}{
		{
			name:             "root overrides all",
			want:             true,
			rootValue:        test.Ptr(true),
			mainValue:        test.Ptr(false),
			defaultValue:     test.Ptr(false),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "main overrides default+hardDefault",
			want:             true,
			mainValue:        test.Ptr(true),
			defaultValue:     test.Ptr(false),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "default overrides hardDefault",
			want:             true,
			defaultValue:     test.Ptr(true),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             true,
			hardDefaultValue: test.Ptr(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.SilentFails = tc.rootValue
			webhook.Main.SilentFails = tc.mainValue
			webhook.Defaults.SilentFails = tc.defaultValue
			webhook.HardDefaults.SilentFails = tc.hardDefaultValue

			// WHEN: GetSilentFails is called.
			got := webhook.GetSilentFails()

			// THEN: the function returns the correct boolean.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetSilentFails() mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_GetType(t *testing.T) {
	// GIVEN: a WebHook with Type in various locations.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "github",
			rootValue:        "github",
			mainValue:        "url",
			defaultValue:     "url",
			hardDefaultValue: "url",
		},
		{
			name:             "main overrides default+hardDefault",
			want:             "github",
			mainValue:        "github",
			defaultValue:     "url",
			hardDefaultValue: "url",
		},
		{
			name:             "default overrides hardDefault",
			want:             "github",
			defaultValue:     "github",
			hardDefaultValue: "url",
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "github",
			hardDefaultValue: "github",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Type = tc.rootValue
			webhook.Main.Type = tc.mainValue
			webhook.Defaults.Type = tc.defaultValue
			webhook.HardDefaults.Type = tc.hardDefaultValue

			// WHEN: GetType is called.
			got := webhook.GetType()

			// THEN: the function returns the correct type.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetType() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestWebHook_GetURL(t *testing.T) {
	// GIVEN: a WebHook with urls in various locations.
	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
		latestVersion                                        string
	}{
		{
			name:             "root overrides all",
			want:             "https://release-argus.io",
			rootValue:        "https://release-argus.io",
			mainValue:        "https://example.com",
			defaultValue:     "https://example.com",
			hardDefaultValue: "https://example.com",
		},
		{
			name:             "main overrides default+hardDefault",
			want:             "https://release-argus.io",
			mainValue:        "https://release-argus.io",
			defaultValue:     "https://example.com",
			hardDefaultValue: "https://example.com",
		},
		{
			name:             "default is last resort",
			want:             "https://release-argus.io",
			defaultValue:     "https://release-argus.io",
			hardDefaultValue: "https://example.com",
		},
		{
			name:             "hardDefaultValue last resort",
			want:             "https://release-argus.io",
			hardDefaultValue: "https://release-argus.io",
		},
		{
			name:          "uses latest_version",
			want:          "https://release-argus.io/1.2.3",
			rootValue:     "https://release-argus.io/{{ version }}",
			latestVersion: "1.2.3",
		},
		{
			name:      "empty version when not found",
			want:      "https://release-argus.io/",
			rootValue: "https://release-argus.io/{{ version }}",
		},
		{
			name: "env var is used",
			want: "https://release-argus.io",
			env: map[string]string{
				"TEST_WEBHOOK__GET_URL__ONE": "https://release-argus.io",
			},
			rootValue: "${TEST_WEBHOOK__GET_URL__ONE}",
		},
		{
			name: "env var partial is used",
			want: "https://release-argus.io",
			env: map[string]string{
				"TEST_WEBHOOK__GET_URL__TWO": "release-argus",
			},
			rootValue: "https://${TEST_WEBHOOK__GET_URL__TWO}.io",
		},
		{
			name: "empty env var is ignored",
			want: "https://release-argus.io",
			env: map[string]string{
				"TEST_WEBHOOK__GET_URL__THREE": "",
			},
			rootValue:    "${TEST_WEBHOOK__GET_URL__THREE}",
			defaultValue: "https://release-argus.io",
		},
		{
			name:         "undefined env var is used",
			want:         "${TEST_WEBHOOK__GET_URL__UNSET}",
			rootValue:    "${TEST_WEBHOOK__GET_URL__UNSET}",
			defaultValue: "https://release-argus.io",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)
			webhook := testWebHook(true, false, false)
			webhook.URL = tc.rootValue
			webhook.Main.URL = tc.mainValue
			webhook.Defaults.URL = tc.defaultValue
			webhook.HardDefaults.URL = tc.hardDefaultValue
			webhook.ServiceStatus.SetLatestVersion(tc.latestVersion, "", false)

			// WHEN: GetURL is called.
			got := webhook.GetURL()

			// THEN: the function returns the url.
			if got != tc.want {
				t.Errorf(
					"%s\nWebHook.GetURL() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}
