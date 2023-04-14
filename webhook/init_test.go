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

//go:build unit

package webhook

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	metric "github.com/release-argus/Argus/web/metrics"
)

func TestSlice_Metrics(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice  *Slice
		ignore *int
	}{
		"nil": {
			slice: nil},
		"empty": {
			slice: &Slice{}},
		"with one": {
			slice: &Slice{
				"foo": &WebHook{
					Main: &WebHook{}}}},
		"no Main, no metrics": {
			slice: &Slice{
				"foo": &WebHook{}},
			ignore: intPtr(1)},
		"multiple": {
			slice: &Slice{
				"bish": &WebHook{
					Main: &WebHook{}},
				"bash": &WebHook{
					Main: &WebHook{}},
				"bosh": &WebHook{
					Main: &WebHook{}}}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

			if tc.slice != nil {
				for name, s := range *tc.slice {
					s.ID = name
					s.ServiceStatus = &svcstatus.Status{ServiceID: stringPtr(name + "-service")}
				}
			}

			// WHEN the Prometheus metrics are initialised with initMetrics
			had := testutil.CollectAndCount(metric.WebHookMetric)
			tc.slice.InitMetrics()

			// THEN it can be counted
			got := testutil.CollectAndCount(metric.WebHookMetric)
			want := had
			if tc.slice != nil {
				want += 2 * len(*tc.slice)
			}
			if tc.ignore != nil {
				want -= 2 * *tc.ignore
			}
			if got != want {
				t.Errorf("got %d metrics, expecting %d",
					got, want)
			}

			// AND the metrics can be deleted
			tc.slice.DeleteMetrics()
			got = testutil.CollectAndCount(metric.WebHookMetric)
			if got != had {
				t.Errorf("deleted metrics but got %d, expecting %d",
					got, want)
			}
		})
	}
}

func TestWebHook_Metrics(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		forService bool
	}{
		"for service": {forService: true},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, tc.forService, false, false)
			if tc.forService {
				webhook.ID = name + "TestInitMetrics"
				webhook.ServiceStatus.ServiceID = stringPtr(name + "TestInitMetrics")
			}

			// WHEN the Prometheus metrics are initialised with initMetrics
			hadC := testutil.CollectAndCount(metric.WebHookMetric)
			webhook.initMetrics()

			// THEN it can be collected
			// counters
			gotC := testutil.CollectAndCount(metric.WebHookMetric)
			wantC := 0
			if tc.forService {
				wantC = 2
			}
			if (gotC - hadC) != wantC {
				t.Errorf("%d Counter metrics's were initialised, expecting %d",
					(gotC - hadC), wantC)
			}

			// AND it can be deleted
			webhook.deleteMetrics()
			gotC = testutil.CollectAndCount(metric.WebHookMetric)
			if gotC != hadC {
				t.Errorf("Counter metrics were not deleted, still have %d. expexting %d",
					gotC, hadC)
			}
		})
	}
}

func TestWebHook_Init(t *testing.T) {
	// GIVEN a WebHook and vars for the Init
	webhook := testWebHook(true, true, false, false)
	var notifiers shoutrrr.Slice
	var main WebHook
	var defaults WebHook
	var hardDefaults WebHook
	status := svcstatus.Status{ServiceID: stringPtr("TestInit")}
	status.Init(
		0, 0, 1,
		stringPtr("TestInit"),
		stringPtr("https://example.com"))

	// WHEN Init is called on it
	webhook.Init(
		&status,
		&main, &defaults, &hardDefaults,
		&notifiers,
		webhook.ParentInterval)
	webhook.ID = "TestInit"

	// THEN pointers to those vars are handed out to the WebHook
	// main
	if webhook.Main != &main {
		t.Errorf("Main was not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&main, webhook.Main)
	}
	// defaults
	if webhook.Defaults != &defaults {
		t.Errorf("Defaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&defaults, webhook.Defaults)
	}
	// hardDefaults
	if webhook.HardDefaults != &hardDefaults {
		t.Errorf("HardDefaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&hardDefaults, webhook.HardDefaults)
	}
	// status
	if webhook.ServiceStatus != &status {
		t.Errorf("Status was not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&status, webhook.ServiceStatus)
	}
	// options
	if webhook.Notifiers.Shoutrrr != &notifiers {
		t.Errorf("Notifiers were not handed to the WebHook correctly\n want: %v\ngot:  %v",
			&notifiers, webhook.Notifiers.Shoutrrr)
	}
}

func TestSlice_Init(t *testing.T) {
	// GIVEN a Slice and vars for the Init
	testLogging("DEBUG")
	var notifiers shoutrrr.Slice
	tests := map[string]struct {
		slice        *Slice
		nilSlice     bool
		mains        *Slice
		defaults     *WebHook
		hardDefaults *WebHook
	}{
		"nil slice": {
			slice: nil, nilSlice: true,
		},
		"empty slice": {
			slice: &Slice{},
		},
		"no mains": {
			slice: &Slice{
				"fail": testWebHook(true, true, false, false),
				"pass": testWebHook(false, true, false, false)},
		},
		"slice with nil element and matching main": {
			slice: &Slice{
				"fail": nil},
			mains: &Slice{
				"fail": testWebHook(false, true, false, false)},
		},
		"have matching mains": {
			slice: &Slice{
				"fail": testWebHook(true, true, false, false),
				"pass": testWebHook(false, true, false, false)},
			mains: &Slice{
				"fail": testWebHook(false, true, false, false),
				"pass": testWebHook(true, true, false, false),
			},
		},
		"some matching mains": {
			slice: &Slice{
				"fail": testWebHook(true, true, false, false),
				"pass": testWebHook(false, true, false, false)},
			mains: &Slice{
				"other": testWebHook(false, true, false, false),
				"pass":  testWebHook(true, true, false, false)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !tc.nilSlice {
				for i := range *tc.slice {
					if (*tc.slice)[i] != nil {
						(*tc.slice)[i].ID = name + i
					}
				}
			}
			serviceStatus := svcstatus.Status{ServiceID: &name}
			mainCount := 0
			if tc.mains != nil {
				mainCount = len(*tc.mains)
			}
			serviceStatus.Init(
				0, 0, mainCount,
				&name,
				nil)
			if tc.mains != nil {
				for i := range *tc.mains {
					(*tc.mains)[i].ID = ""
					(*tc.mains)[i].Main = nil
					(*tc.mains)[i].Defaults = nil
					(*tc.mains)[i].HardDefaults = nil
				}
			}
			parentInterval := "10s"

			// WHEN Init is called on it
			tc.slice.Init(
				&serviceStatus,
				tc.mains, tc.defaults, tc.hardDefaults,
				&notifiers,
				&parentInterval)

			// THEN pointers to those vars are handed out to the WebHook
			if tc.nilSlice {
				if tc.slice != nil {
					t.Fatalf("expecting the Slice to be nil, not %v",
						*tc.slice)
				}
				return
			}
			for _, webhook := range *tc.slice {
				// main
				if webhook.Main == nil {
					t.Errorf("Main of the WebHook was not initialised. got: %v",
						webhook.Main)
				} else if tc.mains != nil && (*tc.mains)[webhook.ID] != nil && webhook.Main != (*tc.mains)[webhook.ID] {
					t.Errorf("Main were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						(*tc.mains)[webhook.ID], webhook.Main)
				}
				// defaults
				if webhook.Defaults != tc.defaults {
					t.Errorf("Defaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&tc.defaults, webhook.Defaults)
				}
				// hardDefaults
				if webhook.HardDefaults != tc.hardDefaults {
					t.Errorf("HardDefaults were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&tc.hardDefaults, webhook.HardDefaults)
				}
				// status
				if webhook.ServiceStatus != &serviceStatus {
					t.Errorf("Status was not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&serviceStatus, webhook.ServiceStatus)
				}
				// notifiers
				if webhook.Notifiers.Shoutrrr != &notifiers {
					t.Errorf("Notifiers were not handed to the WebHook correctly\n want: %v\ngot:  %v",
						&notifiers, webhook.Notifiers.Shoutrrr)
				}
			}
		})
	}
}

func TestWebHook_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		allowInvalidCertsRoot        *bool
		allowInvalidCertsMain        *bool
		allowInvalidCertsDefault     *bool
		allowInvalidCertsHardDefault *bool
		want                         bool
	}{
		"root overrides all": {
			want:                         true,
			allowInvalidCertsRoot:        boolPtr(true),
			allowInvalidCertsMain:        boolPtr(false),
			allowInvalidCertsDefault:     boolPtr(false),
			allowInvalidCertsHardDefault: boolPtr(false),
		},
		"main overrides default+hardDefault": {
			want:                         true,
			allowInvalidCertsRoot:        nil,
			allowInvalidCertsMain:        boolPtr(true),
			allowInvalidCertsDefault:     boolPtr(false),
			allowInvalidCertsHardDefault: boolPtr(false),
		},
		"default overrides hardDefault": {
			want:                         true,
			allowInvalidCertsRoot:        nil,
			allowInvalidCertsMain:        nil,
			allowInvalidCertsDefault:     boolPtr(true),
			allowInvalidCertsHardDefault: boolPtr(false),
		},
		"hardDefault is last resort": {
			want:                         true,
			allowInvalidCertsRoot:        nil,
			allowInvalidCertsMain:        nil,
			allowInvalidCertsDefault:     nil,
			allowInvalidCertsHardDefault: boolPtr(true),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.AllowInvalidCerts = tc.allowInvalidCertsRoot
			webhook.Main.AllowInvalidCerts = tc.allowInvalidCertsMain
			webhook.Defaults.AllowInvalidCerts = tc.allowInvalidCertsDefault
			webhook.HardDefaults.AllowInvalidCerts = tc.allowInvalidCertsHardDefault

			// WHEN GetAllowInvalidCerts is called
			got := webhook.GetAllowInvalidCerts()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetDelay(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		delayRoot        string
		delayMain        string
		delayDefault     string
		delayHardDefault string
		want             string
	}{
		"root overrides all": {
			want:             "1s",
			delayRoot:        "1s",
			delayMain:        "2s",
			delayDefault:     "2s",
			delayHardDefault: "2s",
		},
		"main overrides default+hardDefault": {
			want:             "1s",
			delayRoot:        "",
			delayMain:        "1s",
			delayDefault:     "2s",
			delayHardDefault: "2s",
		},
		"default overrides hardDefault": {
			want:             "1s",
			delayRoot:        "",
			delayMain:        "",
			delayDefault:     "1s",
			delayHardDefault: "2s",
		},
		"hardDefault is last resort": {
			want:             "1s",
			delayRoot:        "",
			delayMain:        "",
			delayDefault:     "",
			delayHardDefault: "1s",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.Delay = tc.delayRoot
			webhook.Main.Delay = tc.delayMain
			webhook.Defaults.Delay = tc.delayDefault
			webhook.HardDefaults.Delay = tc.delayHardDefault

			// WHEN GetDelay is called
			got := webhook.GetDelay()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %s\ngot:  %s",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetDelayDuration(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		delayRoot        string
		delayMain        string
		delayDefault     string
		delayHardDefault string
		want             time.Duration
	}{
		"root overrides all": {
			want:             1 * time.Second,
			delayRoot:        "1s",
			delayMain:        "2s",
			delayDefault:     "2s",
			delayHardDefault: "2s",
		},
		"main overrides default+hardDefault": {
			want:             1 * time.Second,
			delayRoot:        "",
			delayMain:        "1s",
			delayDefault:     "2s",
			delayHardDefault: "2s",
		},
		"default overrides hardDefault": {
			want:             1 * time.Second,
			delayRoot:        "",
			delayMain:        "",
			delayDefault:     "1s",
			delayHardDefault: "2s",
		},
		"hardDefault is last resort": {
			want:             1 * time.Second,
			delayRoot:        "",
			delayMain:        "",
			delayDefault:     "",
			delayHardDefault: "1s",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.Delay = tc.delayRoot
			webhook.Main.Delay = tc.delayMain
			webhook.Defaults.Delay = tc.delayDefault
			webhook.HardDefaults.Delay = tc.delayHardDefault

			// WHEN GetDelayDuration is called
			got := webhook.GetDelayDuration()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %s\ngot:  %s",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetDesiredStatusCode(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		desiredStatusCodeRoot        *int
		desiredStatusCodeMain        *int
		desiredStatusCodeDefault     *int
		desiredStatusCodeHardDefault *int
		want                         int
	}{
		"root overrides all": {
			want:                         1,
			desiredStatusCodeRoot:        intPtr(1),
			desiredStatusCodeMain:        intPtr(2),
			desiredStatusCodeDefault:     intPtr(2),
			desiredStatusCodeHardDefault: intPtr(2),
		},
		"main overrides default+hardDefault": {
			want:                         1,
			desiredStatusCodeRoot:        nil,
			desiredStatusCodeMain:        intPtr(1),
			desiredStatusCodeDefault:     intPtr(2),
			desiredStatusCodeHardDefault: intPtr(2),
		},
		"default overrides hardDefault": {
			want:                         1,
			desiredStatusCodeRoot:        nil,
			desiredStatusCodeMain:        nil,
			desiredStatusCodeDefault:     intPtr(1),
			desiredStatusCodeHardDefault: intPtr(2),
		},
		"hardDefault is last resort": {
			want:                         1,
			desiredStatusCodeRoot:        nil,
			desiredStatusCodeMain:        nil,
			desiredStatusCodeDefault:     nil,
			desiredStatusCodeHardDefault: intPtr(1),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.DesiredStatusCode = tc.desiredStatusCodeRoot
			webhook.Main.DesiredStatusCode = tc.desiredStatusCodeMain
			webhook.Defaults.DesiredStatusCode = tc.desiredStatusCodeDefault
			webhook.HardDefaults.DesiredStatusCode = tc.desiredStatusCodeHardDefault

			// WHEN GetDesiredStatusCode is called
			got := webhook.GetDesiredStatusCode()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %d\ngot:  %d",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetMaxTries(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		maxTriesRoot        *uint
		maxTriesMain        *uint
		maxTriesDefault     *uint
		maxTriesHardDefault *uint
		want                uint
	}{
		"root overrides all": {
			want:                uint(1),
			maxTriesRoot:        uintPtr(1),
			maxTriesMain:        uintPtr(2),
			maxTriesDefault:     uintPtr(2),
			maxTriesHardDefault: uintPtr(2),
		},
		"main overrides default+hardDefault": {
			want:                uint(1),
			maxTriesRoot:        nil,
			maxTriesMain:        uintPtr(1),
			maxTriesDefault:     uintPtr(2),
			maxTriesHardDefault: uintPtr(2),
		},
		"default overrides hardDefault": {
			want:                uint(1),
			maxTriesRoot:        nil,
			maxTriesMain:        nil,
			maxTriesDefault:     uintPtr(1),
			maxTriesHardDefault: uintPtr(2),
		},
		"hardDefault is last resort": {
			want:                uint(1),
			maxTriesRoot:        nil,
			maxTriesMain:        nil,
			maxTriesDefault:     nil,
			maxTriesHardDefault: uintPtr(1),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.MaxTries = tc.maxTriesRoot
			webhook.Main.MaxTries = tc.maxTriesMain
			webhook.Defaults.MaxTries = tc.maxTriesDefault
			webhook.HardDefaults.MaxTries = tc.maxTriesHardDefault

			// WHEN GetMaxTries is called
			got := webhook.GetMaxTries()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %d\ngot:  %d",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetRequest(t *testing.T) {
	// GIVEN a WebHook and a HTTP Request
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.Type = tc.webhookType
			webhook.URL = tc.url
			webhook.CustomHeaders = &tc.customHeaders

			// WHEN GetRequest is called
			req := webhook.GetRequest()

			// THEN the function returns the correct result
			if tc.wantNil {
				if req != nil {
					t.Fatalf("expected request to fail with url %q",
						tc.url)
				}
				return
			}
			switch tc.webhookType {
			case "github":
				// Payload
				body, _ := io.ReadAll(req.Body)
				var payload GitHub
				json.Unmarshal(body, &payload)
				want := "refs/heads/master"
				if payload.Ref != want {
					t.Errorf("didn't get %q in the payload\n%v",
						want, payload)
				}
				// Content-Type
				want = "application/json"
				if req.Header["Content-Type"][0] != want {
					t.Errorf("didn't get %q in the Content-Type\n%v",
						want, req.Header["Content-Type"])
				}
				// X-Github-Event
				want = "push"
				if req.Header["X-Github-Event"][0] != want {
					t.Errorf("GitHub headers weren't set? Didn't get %q in the X-Github-Event\n%v",
						want, req.Header["X-Github-Event"])
				}
			case "gitlab":
				// Content-Type
				want := "application/x-www-form-urlencoded"
				if req.Header["Content-Type"][0] != want {
					t.Errorf("didn't get %q in the Content-Type\n%v",
						want, req.Header["Content-Type"])
				}
			}
			// Custom Headers
			for _, header := range tc.customHeaders {
				if len(req.Header[header.Key]) == 0 {
					t.Fatalf("Custom Headers not set\n%v",
						req.Header)
				}
				if req.Header[header.Key][0] != header.Value {
					t.Fatalf("Custom Headers not set correctly\nwant %q to be %q, not %q\n%v",
						header, header.Value, req.Header[header.Key][0], req.Header)
				}
			}
		})
	}
}

func TestWebHook_GetType(t *testing.T) {
	// GIVEN a WebHook with Type in various locations
	tests := map[string]struct {
		typeRoot        string
		typeMain        string
		typeDefault     string
		typeHardDefault string
		want            string
	}{
		"root overrides all": {
			want:            "github",
			typeRoot:        "github",
			typeMain:        "url",
			typeDefault:     "url",
			typeHardDefault: "url",
		},
		"main overrides default+hardDefault": {
			want:            "github",
			typeRoot:        "",
			typeMain:        "github",
			typeDefault:     "url",
			typeHardDefault: "url",
		},
		"default overrides hardDefault": {
			want:            "github",
			typeRoot:        "",
			typeMain:        "",
			typeDefault:     "github",
			typeHardDefault: "url",
		},
		"hardDefault is last resort": {
			want:            "github",
			typeRoot:        "",
			typeMain:        "",
			typeDefault:     "",
			typeHardDefault: "github",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.Type = tc.typeRoot
			webhook.Main.Type = tc.typeMain
			webhook.Defaults.Type = tc.typeDefault
			webhook.HardDefaults.Type = tc.typeHardDefault

			// WHEN GetType is called
			got := webhook.GetType()

			// THEN the function returns the correct type
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetSecret(t *testing.T) {
	// GIVEN a WebHook with Secret in various locations
	tests := map[string]struct {
		secretRoot        string
		secretMain        string
		secretDefault     string
		secretHardDefault string
		want              string
	}{
		"root overrides all": {
			want:              "argus-secret",
			secretRoot:        "argus-secret",
			secretMain:        "unused",
			secretDefault:     "unused",
			secretHardDefault: "unused",
		},
		"main overrides default+hardDefault": {
			want:              "argus-secret",
			secretRoot:        "",
			secretMain:        "argus-secret",
			secretDefault:     "unused",
			secretHardDefault: "unused",
		},
		"default overrides hardDefault": {
			want:              "argus-secret",
			secretRoot:        "",
			secretMain:        "",
			secretDefault:     "argus-secret",
			secretHardDefault: "unused",
		},
		"hardDefault isn't used": {
			want:              "",
			secretRoot:        "",
			secretMain:        "",
			secretDefault:     "",
			secretHardDefault: "argus-secret",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.Secret = tc.secretRoot
			webhook.Main.Secret = tc.secretMain
			webhook.Defaults.Secret = tc.secretDefault
			webhook.HardDefaults.Secret = tc.secretHardDefault

			// WHEN GetSecret is called
			got := webhook.GetSecret()

			// THEN the function returns the correct secret
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetSilentFails(t *testing.T) {
	// GIVEN a WebHook with SilentFails in various locations
	tests := map[string]struct {
		silentFailsRoot        *bool
		silentFailsMain        *bool
		silentFailsDefault     *bool
		silentFailsHardDefault *bool
		want                   bool
	}{
		"root overrides all": {
			want:                   true,
			silentFailsRoot:        boolPtr(true),
			silentFailsMain:        boolPtr(false),
			silentFailsDefault:     boolPtr(false),
			silentFailsHardDefault: boolPtr(false),
		},
		"main overrides default+hardDefault": {
			want:                   true,
			silentFailsRoot:        nil,
			silentFailsMain:        boolPtr(true),
			silentFailsDefault:     boolPtr(false),
			silentFailsHardDefault: boolPtr(false),
		},
		"default overrides hardDefault": {
			want:                   true,
			silentFailsRoot:        nil,
			silentFailsMain:        nil,
			silentFailsDefault:     boolPtr(true),
			silentFailsHardDefault: boolPtr(false),
		},
		"hardDefault is last resort": {
			want:                   true,
			silentFailsRoot:        nil,
			silentFailsMain:        nil,
			silentFailsDefault:     nil,
			silentFailsHardDefault: boolPtr(true),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.SilentFails = tc.silentFailsRoot
			webhook.Main.SilentFails = tc.silentFailsMain
			webhook.Defaults.SilentFails = tc.silentFailsDefault
			webhook.HardDefaults.SilentFails = tc.silentFailsHardDefault

			// WHEN GetSilentFails is called
			got := webhook.GetSilentFails()

			// THEN the function returns the correct boolean
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetURL(t *testing.T) {
	// GIVEN a WebHook with urls in various locations
	tests := map[string]struct {
		urlRoot        string
		urlMain        string
		urlDefault     string
		urlHardDefault string
		want           string
		latestVersion  string
	}{
		"root overrides all": {
			want:           "https://release-argus.io",
			urlRoot:        "https://release-argus.io",
			urlMain:        "https://somewhere.com",
			urlDefault:     "https://somewhere.com",
			urlHardDefault: "https://somewhere.com",
		},
		"main overrides default+hardDefault": {
			want:           "https://release-argus.io",
			urlRoot:        "",
			urlMain:        "https://release-argus.io",
			urlDefault:     "https://somewhere.com",
			urlHardDefault: "https://somewhere.com",
		},
		"default is last resort": {
			want:           "https://release-argus.io",
			urlRoot:        "",
			urlMain:        "",
			urlDefault:     "https://release-argus.io",
			urlHardDefault: "https://somewhere.com",
		},
		"hardDefault isn't used": {
			want:           "",
			urlRoot:        "",
			urlMain:        "",
			urlDefault:     "",
			urlHardDefault: "https://release-argus.io",
		},
		"uses latest_version": {
			want:           "https://release-argus.io/1.2.3",
			urlRoot:        "https://release-argus.io/{{ version }}",
			urlMain:        "",
			urlDefault:     "",
			urlHardDefault: "",
			latestVersion:  "1.2.3",
		},
		"empty version when unfound": {
			want:           "https://release-argus.io/",
			urlRoot:        "https://release-argus.io/{{ version }}",
			urlMain:        "",
			urlDefault:     "",
			urlHardDefault: "",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.URL = tc.urlRoot
			webhook.Main.URL = tc.urlMain
			webhook.Defaults.URL = tc.urlDefault
			webhook.HardDefaults.URL = tc.urlHardDefault
			webhook.ServiceStatus.SetLatestVersion(tc.latestVersion, false)

			// WHEN GetURL is called
			got := webhook.GetURL()

			// THEN the function returns the url
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_GetIsRunnable(t *testing.T) {
	// GIVEN a WebHook with a NextRunnable time
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.SetNextRunnable(&tc.nextRunnable)
			time.Sleep(time.Nanosecond)

			// WHEN GetIsRunnable is called
			got := webhook.IsRunnable()

			// THEN the function returns whether the webhook is runnable now
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestWebHook_SetExecuting(t *testing.T) {
	// GIVEN a WebHook in different fail states
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
		"sending with maxTries 10 and delay does delay by 3*maxTries+delay+1h": {
			timeDifference: time.Hour + 30*time.Minute + 30*time.Second + 15*time.Second,
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
			failed:         boolPtr(true),
		},
		"success (failed=false) does delay by 2*Interval": {
			timeDifference: 24 * time.Minute,
			failed:         boolPtr(false),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, true, false, false)
			webhook.Failed.Set(webhook.ID, tc.failed)
			webhook.Delay = tc.delay
			maxTries := uint(tc.maxTries)
			webhook.MaxTries = &maxTries

			// WHEN SetExecuting is run
			webhook.SetExecuting(tc.addDelay, tc.sending)

			// THEN the correct response is received
			// next runnable is within expectred range
			now := time.Now().UTC()
			minTime := now.Add(tc.timeDifference - time.Second)
			maxTime := now.Add(tc.timeDifference + time.Second)
			gotTime := webhook.GetNextRunnable()
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf("ran at\n%s\nwant between:\n%s and\n%s\ngot\n%s",
					now, minTime, maxTime, gotTime)
			}
		})
	}
}
