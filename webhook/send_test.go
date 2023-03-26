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
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/util"
)

func TestWebHook_Try(t *testing.T) {
	// GIVEN a WebHook
	testLogging("WARN")
	tests := map[string]struct {
		url               *string
		allowInvalidCerts bool
		selfSignedCert    bool
		wouldFail         bool
		errRegex          string
		desiredStatusCode int
	}{
		"invalid url": {
			url:      stringPtr("invalid://	test"),
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				webhook := testWebHook(false, true, tc.selfSignedCert, false)
				if tc.wouldFail {
					webhook = testWebHook(true, true, tc.selfSignedCert, false)
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
	testLogging("INFO")
	tests := map[string]struct {
		customHeaders bool
		wouldFail     bool
		useDelay      bool
		delay         string
		stdoutRegex   string
		tries         int
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
		"does use delay webhook": {
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
		"tries multiple times": {
			wouldFail:   true,
			tries:       2,
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
			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				stdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				webhook := testWebHook(tc.wouldFail, true, false, tc.customHeaders)
				webhook.ServiceStatus.Deleting = tc.deleting
				webhook.Delay = tc.delay
				maxTries := uint(tc.tries + 1)
				webhook.MaxTries = &maxTries
				webhook.SilentFails = &tc.silentFails
				webhook.Notifiers = &Notifiers{Shoutrrr: &tc.notifiers}
				if tc.tries > 0 {
					go func() {
						time.Sleep(time.Duration(6*(tc.tries-1))*time.Second + time.Second)
						webhook.Secret = "argus"
					}()
				}

				// WHEN try is called with it
				webhook.Send(util.ServiceInfo{}, tc.useDelay)

				// THEN the logs are expected
				w.Close()
				out, _ := io.ReadAll(r)
				os.Stdout = stdout
				output := string(out)
				re := regexp.MustCompile(tc.stdoutRegex)
				output = strings.ReplaceAll(output, "\n", "-n")
				match := re.MatchString(output)
				if !match {
					if strings.Contains(output, "context deadline exceeded") {
						contextDeadlineExceeded = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Errorf("match on %q not found in\n%q",
						tc.stdoutRegex, output)
				}
			}
		})
	}
}

func TestSlice_Send(t *testing.T) {
	// GIVEN a Slice
	testLogging("INFO")
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
				"pass": testWebHook(false, true, false, false),
				"fail": testWebHook(true, true, false, false)},
			stdoutRegex:    `WebHook received.*failed \d times to send the WebHook`,
			stdoutRegexAlt: `failed \d times to send the WebHook.*WebHook received`},
		"does apply webhook delay": {
			slice: &Slice{
				"pass": testWebHook(false, true, false, false),
				"fail": testWebHook(true, true, false, false)},
			stdoutRegex: `WebHook received.*failed \d times to send the WebHook`,
			useDelay:    true,
			delays: map[string]string{
				"fail": "2s",
				"pass": "1ms"},
			repeat: 5},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				tc.repeat++ // repeat to check delay usage as map order is random
				for tc.repeat != 0 {
					stdout := os.Stdout
					r, w, _ := os.Pipe()
					os.Stdout = w
					if tc.slice != nil {
						for id := range *tc.slice {
							(*tc.slice)[id].ID = id
						}
						for id := range tc.delays {
							(*tc.slice)[id].Delay = tc.delays[id]
						}
					}

					// WHEN try is called with it
					tc.slice.Send(util.ServiceInfo{}, tc.useDelay)

					// THEN the logs are expected
					w.Close()
					out, _ := io.ReadAll(r)
					os.Stdout = stdout
					output := string(out)
					output = strings.ReplaceAll(output, "\n", "-n")
					re := regexp.MustCompile(tc.stdoutRegex)
					match := re.MatchString(output)
					if !match {
						if strings.Contains(output, "context deadline exceeded") {
							contextDeadlineExceeded = true
							if try != 3 {
								time.Sleep(time.Second)
								continue
							}
						}
						if tc.stdoutRegexAlt != "" {
							re = regexp.MustCompile(tc.stdoutRegexAlt)
							match = re.MatchString(output)
							if !match {
								t.Errorf("match on %q not found in\n%q",
									tc.stdoutRegexAlt, output)
							}
							return
						}
						t.Errorf("match on %q not found in\n%q",
							tc.stdoutRegex, output)
					}
					tc.repeat--
				}
			}
		})
	}
}

func TestNotifiers_SendWithNotifier(t *testing.T) {
	// GIVEN Notifiers
	testLogging("INFO")
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			notifiers := Notifiers{Shoutrrr: tc.shoutrrrNotifiers}

			// WHEN Send is called with them
			err := notifiers.Send("TestNotifiersSendWithNotifier", name, &util.ServiceInfo{})

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
		name, tc := name, tc
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
