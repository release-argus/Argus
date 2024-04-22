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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

func TestWebHook_Try(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		url               *string
		allowInvalidCerts bool
		selfSignedCert    bool
		wouldFail         bool
		errRegex          string
		desiredStatusCode int
	}{
		"invalid url": {
			url:      test.StringPtr("invalid://	test"),
			errRegex: "failed to get .?http.request"},
		"fail due to invalid secret": {
			wouldFail: true,
			errRegex:  "WebHook gave [0-9]+, not "},
		"fail due to invalid cert": {
			selfSignedCert: true,
			errRegex:       " x509:"},
		"pass with invalid certs allowed": {
			selfSignedCert:    true,
			errRegex:          "^$",
			allowInvalidCerts: true},
		"pass with valid certs": {
			errRegex:          "^$",
			allowInvalidCerts: true},
		"fail by not getting desired status code": {
			desiredStatusCode: 1,
			errRegex:          "WebHook gave [0-9]+, not ",
			allowInvalidCerts: true},
		"pass by getting desired status code": {
			wouldFail:         true,
			desiredStatusCode: 500,
			errRegex:          "^$",
			allowInvalidCerts: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				webhook := testWebHook(false, tc.selfSignedCert, false)
				if tc.wouldFail {
					webhook = testWebHook(true, tc.selfSignedCert, false)
				}
				if tc.url != nil {
					webhook.URL = *tc.url
				}
				webhook.AllowInvalidCerts = &tc.allowInvalidCerts
				webhook.DesiredStatusCode = &tc.desiredStatusCode

				// WHEN try is called with it
				err := webhook.try(&util.LogFrom{})

				// THEN any err is expected
				e := util.ErrorToString(err)
				re := regexp.MustCompile(tc.errRegex)
				match := re.MatchString(e)
				if !match {
					if strings.Contains(e, "context deadline exceeded") {
						contextDeadlineExceeded = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Errorf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
			}
		})
	}
}

func TestWebHook_Send(t *testing.T) {
	// GIVEN a WebHook
	tests := map[string]struct {
		customHeaders bool
		wouldFail     bool
		useDelay      bool
		delay         string
		stdoutRegex   string
		retries       int
		silentFails   bool
		notifiers     shoutrrr.Slice
		deleting      bool
	}{
		"successful webhook": {
			stdoutRegex: "WebHook received",
		},
		"successful webhook with custom_headers": {
			stdoutRegex:   "WebHook received",
			customHeaders: true,
		},
		"does use delay": {
			useDelay:    true,
			delay:       "3s",
			stdoutRegex: "WebHook received",
		},
		"no delay": {
			useDelay:    true,
			delay:       "0s",
			stdoutRegex: "WebHook received",
		},
		"failing webhook": {
			wouldFail:   true,
			stdoutRegex: `failed \d times to send`,
		},
		"failing webhook with custom_headers": {
			wouldFail:     true,
			customHeaders: true,
			stdoutRegex:   `failed \d times to send`,
		},
		"retries multiple times": {
			wouldFail:   true,
			retries:     2,
			stdoutRegex: `(WebHook gave 500.*){2}WebHook received`,
		},
		"does try notifiers on fail": {
			wouldFail:   true,
			stdoutRegex: `WebHook gave 500.*invalid gotify token`,
			notifiers: shoutrrr.Slice{
				"fail": testNotifier(true, false)},
		},
		"doesn't try notifiers on fail if silentFails": {
			wouldFail:   true,
			silentFails: true,
			stdoutRegex: `WebHook gave 500.*failed \d times to send the WebHook [^-]+-n$`,
			notifiers: shoutrrr.Slice{
				"fail": testNotifier(true, false)},
		},
		"doesn't send if deleting": {
			deleting:    true,
			stdoutRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout

			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				releaseStdout := test.CaptureStdout()
				webhook := testWebHook(tc.wouldFail, false, tc.customHeaders)
				if tc.deleting {
					webhook.ServiceStatus.SetDeleting()
				}
				webhook.Delay = tc.delay
				maxTries := uint(tc.retries + 1)
				webhook.MaxTries = &maxTries
				webhook.SilentFails = &tc.silentFails
				webhook.Notifiers = &Notifiers{Shoutrrr: &tc.notifiers}
				serviceInfo := &util.ServiceInfo{ID: name}
				if tc.retries > 0 {
					go func() {
						fails := testutil.ToFloat64(metric.WebHookMetric.WithLabelValues(
							webhook.ID, "FAIL", serviceInfo.ID))
						for fails < float64(tc.retries) {
							fails = testutil.ToFloat64(metric.WebHookMetric.WithLabelValues(
								webhook.ID, "FAIL", serviceInfo.ID))
							time.Sleep(time.Millisecond * 200)
						}
						t.Logf("Failed %d times", tc.retries)
						webhook.mutex.Lock()
						webhook.Secret = "argus"
						webhook.mutex.Unlock()
					}()
				}

				// WHEN try is called with it
				startAt := time.Now()
				webhook.Send(serviceInfo, tc.useDelay)

				// THEN the logs are expected
				completedAt := time.Now()
				stdout := releaseStdout()
				re := regexp.MustCompile(tc.stdoutRegex)
				stdout = strings.ReplaceAll(stdout, "\n", "-n")
				match := re.MatchString(stdout)
				if !match {
					if strings.Contains(stdout, "context deadline exceeded") {
						contextDeadlineExceeded = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Errorf("match on %q not found in\n%q",
						tc.stdoutRegex, stdout)
				}
				// AND the delay is expected
				if tc.delay != "" {
					delayDuration, _ := time.ParseDuration(tc.delay)
					took := completedAt.Sub(startAt)
					if took < delayDuration {
						t.Errorf("delay %s not used", tc.delay)
					} else if took > delayDuration+2*time.Second {
						t.Errorf("delay %s took too long %s", tc.delay, took)
					}
				}
			}
		})
	}
}

func TestSlice_Send(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice          *Slice
		stdoutRegex    string
		stdoutRegexAlt string
		notifiers      shoutrrr.Slice
		useDelay       bool
		delays         map[string]string
		repeat         int
	}{
		"nil slice": {
			slice:       nil,
			stdoutRegex: `^$`},
		"successful and failing webhook": {
			slice: &Slice{
				"pass": testWebHook(false, false, false),
				"fail": testWebHook(true, false, false)},
			stdoutRegex:    `WebHook received.*failed \d times to send the WebHook`,
			stdoutRegexAlt: `failed \d times to send the WebHook.*WebHook received`},
		"does apply webhook delay": {
			slice: &Slice{
				"pass": testWebHook(false, false, false),
				"fail": testWebHook(true, false, false)},
			stdoutRegex: `WebHook received.*failed \d times to send the WebHook`,
			useDelay:    true,
			delays: map[string]string{
				"fail": "2s",
				"pass": "1ms"},
			repeat: 5},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout

			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				tc.repeat++ // repeat to check delay usage as map order is random
				for tc.repeat != 0 {
					releaseStdout := test.CaptureStdout()
					if tc.slice != nil {
						for id := range *tc.slice {
							(*tc.slice)[id].ID = id
						}
						for id := range tc.delays {
							(*tc.slice)[id].Delay = tc.delays[id]
						}
					}

					// WHEN try is called with it
					tc.slice.Send(&util.ServiceInfo{ID: name}, tc.useDelay)

					// THEN the logs are expected
					stdout := releaseStdout()
					stdout = strings.ReplaceAll(stdout, "\n", "-n")
					re := regexp.MustCompile(tc.stdoutRegex)
					match := re.MatchString(stdout)
					if !match {
						if strings.Contains(stdout, "context deadline exceeded") {
							contextDeadlineExceeded = true
							if try != 3 {
								time.Sleep(time.Second)
								continue
							}
						}
						if tc.stdoutRegexAlt != "" {
							re = regexp.MustCompile(tc.stdoutRegexAlt)
							match = re.MatchString(stdout)
							if !match {
								t.Errorf("match on %q not found in\n%q",
									tc.stdoutRegexAlt, stdout)
							}
							return
						}
						t.Errorf("match on %q not found in\n%q",
							tc.stdoutRegex, stdout)
					}
					tc.repeat--
				}
			}
		})
	}
}

func TestNotifiers_SendWithNotifier(t *testing.T) {
	// GIVEN Notifiers
	tests := map[string]struct {
		shoutrrrNotifiers *shoutrrr.Slice
		errRegex          string
	}{
		"nill Notifiers": {
			errRegex: "^$"},
		"successful notifier": {
			errRegex: "^$",
			shoutrrrNotifiers: &shoutrrr.Slice{
				"pass": testNotifier(false, false)}},
		"failing notifier": {
			errRegex: "invalid gotify token",
			shoutrrrNotifiers: &shoutrrr.Slice{
				"fail": testNotifier(true, false)}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			notifiers := Notifiers{Shoutrrr: tc.shoutrrrNotifiers}

			// WHEN Send is called with them
			err := notifiers.Send("TestNotifiersSendWithNotifier", name, &util.ServiceInfo{ID: name})

			// THEN err is as expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("match on %q not found in\n%q",
					tc.errRegex, e)
			}
		})
	}
}

func TestCheckWebHookBody(t *testing.T) {
	// GIVEN a response body
	tests := map[string]struct {
		body string
		want bool
	}{
		"empty body": {
			body: "",
			want: true},
		"success body": {
			body: "success",
			want: true},
		"awx invalid secret": {
			body: `{"detail":"You do not have permission to perform this action."}`,
			want: false},
		"adnanh/webhook hook fail": {
			body: `Hook rules were not satisfied.`,
			want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN checkWebHookBody is called on it
			got := checkWebHookBody(tc.body)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}
