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
	"os"
	"testing"
	"time"
)

func TestWebHook_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		root        *bool
		main        *bool
		dfault      *bool
		hardDefault *bool
		want        bool
	}{
		"root overrides all": {
			want:        true,
			root:        boolPtr(true),
			main:        boolPtr(false),
			dfault:      boolPtr(false),
			hardDefault: boolPtr(false),
		},
		"main overrides default+hardDefault": {
			want:        true,
			main:        boolPtr(true),
			dfault:      boolPtr(false),
			hardDefault: boolPtr(false),
		},
		"default overrides hardDefault": {
			want:        true,
			dfault:      boolPtr(true),
			hardDefault: boolPtr(false),
		},
		"hardDefault is last resort": {
			want:        true,
			hardDefault: boolPtr(true),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.AllowInvalidCerts = tc.root
			webhook.Main.AllowInvalidCerts = tc.main
			webhook.Defaults.AllowInvalidCerts = tc.dfault
			webhook.HardDefaults.AllowInvalidCerts = tc.hardDefault

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
		root        string
		main        string
		dfault      string
		hardDefault string
		want        string
	}{
		"root overrides all": {
			want:        "1s",
			root:        "1s",
			main:        "2s",
			dfault:      "2s",
			hardDefault: "2s",
		},
		"main overrides default+hardDefault": {
			want:        "1s",
			main:        "1s",
			dfault:      "2s",
			hardDefault: "2s",
		},
		"default overrides hardDefault": {
			want:        "1s",
			dfault:      "1s",
			hardDefault: "2s",
		},
		"hardDefault is last resort": {
			want:        "1s",
			hardDefault: "1s",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Delay = tc.root
			webhook.Main.Delay = tc.main
			webhook.Defaults.Delay = tc.dfault
			webhook.HardDefaults.Delay = tc.hardDefault

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
		root        string
		main        string
		dfault      string
		hardDefault string
		want        time.Duration
	}{
		"root overrides all": {
			want:        1 * time.Second,
			root:        "1s",
			main:        "2s",
			dfault:      "2s",
			hardDefault: "2s",
		},
		"main overrides default+hardDefault": {
			want:        1 * time.Second,
			main:        "1s",
			dfault:      "2s",
			hardDefault: "2s",
		},
		"default overrides hardDefault": {
			want:        1 * time.Second,
			dfault:      "1s",
			hardDefault: "2s",
		},
		"hardDefault is last resort": {
			want:        1 * time.Second,
			hardDefault: "1s",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Delay = tc.root
			webhook.Main.Delay = tc.main
			webhook.Defaults.Delay = tc.dfault
			webhook.HardDefaults.Delay = tc.hardDefault

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
		root        *int
		main        *int
		dfault      *int
		hardDefault *int
		want        int
	}{
		"root overrides all": {
			want:        1,
			root:        intPtr(1),
			main:        intPtr(2),
			dfault:      intPtr(2),
			hardDefault: intPtr(2),
		},
		"main overrides default+hardDefault": {
			want:        1,
			main:        intPtr(1),
			dfault:      intPtr(2),
			hardDefault: intPtr(2),
		},
		"default overrides hardDefault": {
			want:        1,
			dfault:      intPtr(1),
			hardDefault: intPtr(2),
		},
		"hardDefault is last resort": {
			want:        1,
			hardDefault: intPtr(1),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.DesiredStatusCode = tc.root
			webhook.Main.DesiredStatusCode = tc.main
			webhook.Defaults.DesiredStatusCode = tc.dfault
			webhook.HardDefaults.DesiredStatusCode = tc.hardDefault

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
		root        *uint
		main        *uint
		dfault      *uint
		hardDefault *uint
		want        uint
	}{
		"root overrides all": {
			want:        uint(1),
			root:        uintPtr(1),
			main:        uintPtr(2),
			dfault:      uintPtr(2),
			hardDefault: uintPtr(2),
		},
		"main overrides default+hardDefault": {
			want:        uint(1),
			main:        uintPtr(1),
			dfault:      uintPtr(2),
			hardDefault: uintPtr(2),
		},
		"default overrides hardDefault": {
			want:        uint(1),
			dfault:      uintPtr(1),
			hardDefault: uintPtr(2),
		},
		"hardDefault is last resort": {
			want:        uint(1),
			hardDefault: uintPtr(1),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.MaxTries = tc.root
			webhook.Main.MaxTries = tc.main
			webhook.Defaults.MaxTries = tc.dfault
			webhook.HardDefaults.MaxTries = tc.hardDefault

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

func TestWebHook_BuildRequest(t *testing.T) {
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Type = tc.webhookType
			webhook.URL = tc.url
			webhook.CustomHeaders = &tc.customHeaders

			// WHEN BuildRequest is called
			req := webhook.BuildRequest()

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
		root        string
		main        string
		dfault      string
		hardDefault string
		want        string
	}{
		"root overrides all": {
			want:        "github",
			root:        "github",
			main:        "url",
			dfault:      "url",
			hardDefault: "url",
		},
		"main overrides default+hardDefault": {
			want:        "github",
			main:        "github",
			dfault:      "url",
			hardDefault: "url",
		},
		"default overrides hardDefault": {
			want:        "github",
			dfault:      "github",
			hardDefault: "url",
		},
		"hardDefault is last resort": {
			want:        "github",
			hardDefault: "github",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.Type = tc.root
			webhook.Main.Type = tc.main
			webhook.Defaults.Type = tc.dfault
			webhook.HardDefaults.Type = tc.hardDefault

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
		env         map[string]string
		root        string
		main        string
		dfault      string
		hardDefault string
		want        string
	}{
		"root overrides all": {
			want:        "argus-secret",
			root:        "argus-secret",
			main:        "unused",
			dfault:      "unused",
			hardDefault: "unused",
		},
		"main overrides default+hardDefault": {
			want:        "argus-secret",
			main:        "argus-secret",
			dfault:      "unused",
			hardDefault: "unused",
		},
		"default overrides hardDefault": {
			want:        "argus-secret",
			dfault:      "argus-secret",
			hardDefault: "unused",
		},
		"hardDefault last resort": {
			want:        "argus-secret",
			hardDefault: "argus-secret",
		},
		"env var is used": {
			want: "argus-secret",
			env:  map[string]string{"TESTWEBHOOK_GETSECRET_ONE": "argus-secret"},
			root: "${TESTWEBHOOK_GETSECRET_ONE}",
		},
		"env var partial is used": {
			want: "argus-secret-two",
			env:  map[string]string{"TESTWEBHOOK_GETSECRET_TWO": "argus-secret"},
			root: "${TESTWEBHOOK_GETSECRET_TWO}-two",
		},
		"empty env var is ignored": {
			want:   "argus-secret",
			env:    map[string]string{"TESTWEBHOOK_GETSECRET_THREE": ""},
			root:   "${TESTWEBHOOK_GETSECRET_THREE}",
			dfault: "argus-secret",
		},
		"unset env var is used": {
			want:   "${TESTWEBHOOK_GETSECRET_UNSET}",
			root:   "${TESTWEBHOOK_GETSECRET_UNSET}",
			dfault: "argus-secret",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			webhook := testWebHook(true, false, false)
			webhook.Secret = tc.root
			webhook.Main.Secret = tc.main
			webhook.Defaults.Secret = tc.dfault
			webhook.HardDefaults.Secret = tc.hardDefault

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
		root        *bool
		main        *bool
		dfault      *bool
		hardDefault *bool
		want        bool
	}{
		"root overrides all": {
			want:        true,
			root:        boolPtr(true),
			main:        boolPtr(false),
			dfault:      boolPtr(false),
			hardDefault: boolPtr(false),
		},
		"main overrides default+hardDefault": {
			want:        true,
			main:        boolPtr(true),
			dfault:      boolPtr(false),
			hardDefault: boolPtr(false),
		},
		"default overrides hardDefault": {
			want:        true,
			dfault:      boolPtr(true),
			hardDefault: boolPtr(false),
		},
		"hardDefault is last resort": {
			want:        true,
			hardDefault: boolPtr(true),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
			webhook.SilentFails = tc.root
			webhook.Main.SilentFails = tc.main
			webhook.Defaults.SilentFails = tc.dfault
			webhook.HardDefaults.SilentFails = tc.hardDefault

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
		env           map[string]string
		root          string
		main          string
		dfault        string
		hardDefault   string
		want          string
		latestVersion string
	}{
		"root overrides all": {
			want:        "https://release-argus.io",
			root:        "https://release-argus.io",
			main:        "https://somewhere.com",
			dfault:      "https://somewhere.com",
			hardDefault: "https://somewhere.com",
		},
		"main overrides default+hardDefault": {
			want:        "https://release-argus.io",
			main:        "https://release-argus.io",
			dfault:      "https://somewhere.com",
			hardDefault: "https://somewhere.com",
		},
		"default is last resort": {
			want:        "https://release-argus.io",
			dfault:      "https://release-argus.io",
			hardDefault: "https://somewhere.com",
		},
		"hardDefault last resort": {
			want:        "https://release-argus.io",
			hardDefault: "https://release-argus.io",
		},
		"uses latest_version": {
			want:          "https://release-argus.io/1.2.3",
			root:          "https://release-argus.io/{{ version }}",
			latestVersion: "1.2.3",
		},
		"empty version when unfound": {
			want: "https://release-argus.io/",
			root: "https://release-argus.io/{{ version }}",
		},
		"env var is used": {
			want: "https://release-argus.io",
			env:  map[string]string{"TESTWEBHOOK_GETURL_ONE": "https://release-argus.io"},
			root: "${TESTWEBHOOK_GETURL_ONE}",
		},
		"env var partial is used": {
			want: "https://release-argus.io",
			env:  map[string]string{"TESTWEBHOOK_GETURL_TWO": "release-argus"},
			root: "https://${TESTWEBHOOK_GETURL_TWO}.io",
		},
		"empty env var is ignored": {
			want:   "https://release-argus.io",
			env:    map[string]string{"TESTWEBHOOK_GETURL_THREE": ""},
			root:   "${TESTWEBHOOK_GETURL_THREE}",
			dfault: "https://release-argus.io",
		},
		"undefined env var is used": {
			want:   "${TESTWEBHOOK_GETURL_UNSET}",
			root:   "${TESTWEBHOOK_GETURL_UNSET}",
			dfault: "https://release-argus.io",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			webhook := testWebHook(true, false, false)
			webhook.URL = tc.root
			webhook.Main.URL = tc.main
			webhook.Defaults.URL = tc.dfault
			webhook.HardDefaults.URL = tc.hardDefault
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			webhook := testWebHook(true, false, false)
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
			gotTime := webhook.NextRunnable()
			if !(minTime.Before(gotTime)) || !(maxTime.After(gotTime)) {
				t.Fatalf("ran at\n%s\nwant between:\n%s and\n%s\ngot\n%s",
					now, minTime, maxTime, gotTime)
			}
		})
	}
}
