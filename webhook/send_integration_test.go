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

//go:build integration

package webhook

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/web/metric"
)

func TestWebHooks_Send(t *testing.T) {
	// GIVEN: WebHooks.
	tests := []struct {
		name                        string
		webhooks                    *WebHooks
		stdoutRegex, stdoutRegexAlt string
		notifiers                   shoutrrr.Shoutrrrs
		useDelay                    bool
		delays                      map[string]string
		repeat                      int
	}{
		{
			name:        "nil map",
			webhooks:    nil,
			stdoutRegex: `^$`,
		},
		{
			name: "2 successful webhooks",
			webhooks: &WebHooks{
				"pass": testWebHook(false, false, false),
				"fail": testWebHook(false, false, false),
			},
			stdoutRegex:    `WebHook received.*WebHook received`,
			stdoutRegexAlt: `^$`,
		},
		{
			name: "successful and failing defaults",
			webhooks: &WebHooks{
				"pass": testWebHook(false, false, false),
				"fail": testWebHook(true, false, false),
			},
			stdoutRegex:    `WebHook received.*failed \d times to send the WebHook`,
			stdoutRegexAlt: `failed \d times to send the WebHook.*WebHook received`,
		},
		{
			name: "does apply defaults delay",
			webhooks: &WebHooks{
				"pass": testWebHook(false, false, false),
				"fail": testWebHook(true, false, false),
			},
			stdoutRegex: `WebHook received.*failed \d times to send the WebHook`,
			useDelay:    true,
			delays: map[string]string{
				"fail": "2s",
				"pass": "1ms",
			},
			repeat: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				tc.repeat++ // repeat to check delay usage as map order is random.
				for tc.repeat != 0 {
					releaseStdout := test.CaptureLog(t, logx.Default())
					if tc.webhooks != nil {
						for id := range *tc.webhooks {
							(*tc.webhooks)[id].ID = id
						}
						for id := range tc.delays {
							(*tc.webhooks)[id].Delay = tc.delays[id]
						}
					}

					// WHEN: try is called on it.
					tc.webhooks.Send(serviceinfo.ServiceInfo{ID: tc.name}, tc.useDelay)

					prefix := fmt.Sprintf("%s\nWebHooks.Send()", packageName)

					// THEN: the logs are expected.
					stdout := releaseStdout()
					stdout = strings.ReplaceAll(stdout, "\n", "-n")
					if !util.RegexCheck(tc.stdoutRegex, stdout) {
						if strings.Contains(stdout, "context deadline exceeded") {
							contextDeadlineExceeded = true
							if try != 3 {
								time.Sleep(time.Second)
								continue
							}
						}
						if tc.stdoutRegexAlt != "" {
							if !util.RegexCheck(tc.stdoutRegexAlt, stdout) {
								t.Errorf(
									"%s stdoutAlt mismatch\ngot:  %q\nwant: %q",
									prefix, stdout, tc.stdoutRegexAlt,
								)
							}
							return
						}
						t.Errorf(
							"%s stdout mismatch\ngot:  %q\nwant: %q",
							prefix, stdout, tc.stdoutRegex,
						)
					}
					tc.repeat--
				}
			}
		})
	}
}

func TestWebHook_Send(t *testing.T) {
	// GIVEN: a WebHook.
	tests := []struct {
		name                                                string
		headers, wouldFail, useDelay, deleting, silentFails bool
		delay                                               string
		retries                                             uint8
		notifiers                                           shoutrrr.Shoutrrrs
		stdoutRegex                                         string
	}{
		{
			name:        "successful defaults",
			stdoutRegex: `WebHook received`,
		},
		{
			name:        "successful defaults with headers",
			stdoutRegex: `WebHook received`,
			headers:     true,
		},
		{
			name:        "does use delay",
			useDelay:    true,
			delay:       "3s",
			stdoutRegex: `WebHook received`,
		},
		{
			name:        "no delay",
			useDelay:    true,
			delay:       "0s",
			stdoutRegex: `WebHook received`,
		},
		{
			name:        "failing defaults",
			wouldFail:   true,
			stdoutRegex: `failed \d times to send`,
		},
		{
			name:        "failing defaults with headers",
			wouldFail:   true,
			headers:     true,
			stdoutRegex: `failed \d times to send`,
		},
		{
			name:        "retries multiple times",
			wouldFail:   true,
			retries:     2,
			stdoutRegex: `(WebHook gave 500.*){2}WebHook received`,
		},
		{
			name:        "does try notifiers on fail",
			wouldFail:   true,
			stdoutRegex: `WebHook gave 500.*invalid gotify token`,
			notifiers: shoutrrr.Shoutrrrs{
				"fail": shoutrrrtest.Shoutrrr(t, true, false),
			},
		},
		{
			name:        "doesn't try notifiers on fail if silentFails",
			wouldFail:   true,
			silentFails: true,
			stdoutRegex: `WebHook gave 500.*failed \d times to send the WebHook [^-]+-n$`,
			notifiers: shoutrrr.Shoutrrrs{
				"fail": shoutrrrtest.Shoutrrr(t, true, false),
			},
		},
		{
			name:        "doesn't send if deleting",
			deleting:    true,
			stdoutRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded != false {
				try++
				contextDeadlineExceeded = false
				releaseStdout := test.CaptureLog(t, logx.Default())
				webhook := testWebHook(tc.wouldFail, false, tc.headers)
				if tc.deleting {
					webhook.ServiceStatus.SetDeleting()
				}
				webhook.Delay = tc.delay
				webhook.MaxTries = test.Ptr(tc.retries + 1)
				webhook.SilentFails = &tc.silentFails
				webhook.Notifiers = Notifiers{Shoutrrr: &tc.notifiers}
				webhook.ServiceStatus.ServiceInfo.ID = tc.name
				svcInfo := webhook.ServiceStatus.GetServiceInfo()
				if tc.retries > 0 {
					go func() {
						fails := testutil.ToFloat64(
							metric.WebHookResultTotal.WithLabelValues(
								webhook.ID, metric.ActionResultFail, svcInfo.ID,
							),
						)
						for fails < float64(tc.retries) {
							fails = testutil.ToFloat64(
								metric.WebHookResultTotal.WithLabelValues(
									webhook.ID, metric.ActionResultFail, svcInfo.ID,
								),
							)
							time.Sleep(time.Millisecond * 200)
						}
						t.Logf(
							"%s - Failed %d times",
							packageName, tc.retries,
						)
						webhook.mu.Lock()
						webhook.Secret = "argus"
						webhook.mu.Unlock()
					}()
				}

				// WHEN: try is called on it.
				startAt := time.Now()
				webhook.Send(svcInfo, tc.useDelay)

				prefix := fmt.Sprintf("%s\nWebHook.Send()", packageName)

				// THEN: the logs are expected.
				completedAt := time.Now()
				stdout := releaseStdout()
				stdout = strings.ReplaceAll(stdout, "\n", "-n")
				if !util.RegexCheck(tc.stdoutRegex, stdout) {
					if strings.Contains(stdout, "context deadline exceeded") {
						contextDeadlineExceeded = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Errorf(
						"%s stdout mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, tc.stdoutRegex,
					)
				}

				// AND: the delay is expected.
				if tc.delay != "" {
					delayDuration, _ := time.ParseDuration(tc.delay)
					took := completedAt.Sub(startAt)
					if took < delayDuration {
						t.Errorf(
							"%s delay not used\ngot:  %s\nwant: %s+",
							prefix, took, tc.delay,
						)
					} else if took > delayDuration+2*time.Second {
						t.Errorf(
							"%s too much delay\ngot:  %s\nwant: %s",
							prefix, took, tc.delay,
						)
					}
				}
			}
		})
	}
}

func TestNotifiers_SendWithNotifier(t *testing.T) {
	// GIVEN: Notifiers.
	tests := []struct {
		name              string
		shoutrrrNotifiers *shoutrrr.Shoutrrrs
		errRegex          string
	}{
		{
			name:     "nil",
			errRegex: `^$`,
		},
		{
			name:     "successful",
			errRegex: `^$`,
			shoutrrrNotifiers: &shoutrrr.Shoutrrrs{
				"pass": shoutrrrtest.Shoutrrr(t, false, false),
			},
		},
		{
			name:     "failing",
			errRegex: `invalid gotify token`,
			shoutrrrNotifiers: &shoutrrr.Shoutrrrs{
				"fail": shoutrrrtest.Shoutrrr(t, true, false),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifiers := Notifiers{Shoutrrr: tc.shoutrrrNotifiers}

			// WHEN: Send is called with them.
			err := notifiers.Send(
				"TestNotifiersSendWithNotifier",
				tc.name,
				serviceinfo.ServiceInfo{ID: tc.name},
			)

			// THEN: decode is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nNotifiers.Send() error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}
		})
	}
}

func TestWebHook_Try(t *testing.T) {
	tests := []struct {
		name string

		url               string
		secret            string
		whType            string
		headers           Headers
		allowInvalidCerts *bool
		desiredStatusCode *uint16

		errRegex    string
		stdoutRegex string
	}{
		{
			name:        "success on github-style endpoint",
			url:         test.WebHookGitHub["url_valid"],
			secret:      test.WebHookGitHub["secret_pass"],
			whType:      "github",
			errRegex:    `^$`,
			stdoutRegex: `WebHook received`,
		},
		{
			name:     "failure on wrong secret",
			url:      test.WebHookGitHub["url_valid"],
			secret:   test.WebHookGitHub["secret_fail"],
			whType:   "github",
			errRegex: `WebHook gave 500`,
		},
		{
			name:   "success with header auth",
			url:    test.LookupWithHeaderAuth["url_valid"],
			secret: test.WebHookGitHub["secret_pass"],
			whType: "github",
			headers: Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_pass"],
				},
			},
			errRegex:    `^$`,
			stdoutRegex: `WebHook received`,
		},
		{
			name:   "failure with invalid header auth",
			url:    test.LookupWithHeaderAuth["url_valid"],
			secret: test.WebHookGitHub["secret_pass"],
			whType: "github",
			headers: Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_fail"],
				},
			},
			errRegex: `(?s)WebHook gave 200, not 2XX.*Hook rules were not satisfied`,
		},
		{
			name:     "rejects invalid TLS certificate",
			url:      test.WebHookGitHub["url_invalid"],
			secret:   test.WebHookGitHub["secret_pass"],
			whType:   "github",
			errRegex: `(?i)certificate`,
		},
		{
			name:              "allows invalid TLS certificate when configured",
			url:               test.WebHookGitHub["url_invalid"],
			secret:            test.WebHookGitHub["secret_pass"],
			whType:            "github",
			allowInvalidCerts: test.Ptr(true),
			errRegex:          `^$`,
			stdoutRegex:       `WebHook received`,
		},
		{
			name:        "unsupported webhook type",
			url:         test.WebHookGitHub["url_valid"],
			secret:      test.WebHookGitHub["secret_pass"],
			whType:      "url",
			errRegex:    `failed to get \*http.request for WebHook`,
			stdoutRegex: `failed to get \*http.request for WebHook`,
		},
		{
			name:              "wrong desired status code",
			url:               test.WebHookGitHub["url_valid"],
			secret:            test.WebHookGitHub["secret_pass"],
			whType:            "github",
			desiredStatusCode: test.Ptr(uint16(404)),
			errRegex:          `WebHook gave 200, not 404`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			try := 0
			contextDeadlineExceeded := true
			for contextDeadlineExceeded {
				try++
				contextDeadlineExceeded = false

				releaseStdout := test.CaptureLog(t, logx.Default())

				webhook := testWebHook(false, false, false)
				webhook.URL = tc.url
				webhook.Secret = tc.secret
				webhook.Type = tc.whType
				if tc.headers != nil {
					webhook.Headers = tc.headers
				}
				if tc.allowInvalidCerts != nil {
					webhook.AllowInvalidCerts = tc.allowInvalidCerts
				}
				if tc.desiredStatusCode != nil {
					webhook.DesiredStatusCode = tc.desiredStatusCode
				}

				logFrom := logx.LogFrom{
					Primary:   webhook.ID,
					Secondary: webhook.ServiceStatus.ServiceInfo.ID,
				}

				// WHEN: try is called.
				err := webhook.try(logFrom)

				stdout := releaseStdout()
				prefix := fmt.Sprintf("%s\nWebHook.try()", packageName)

				// THEN: the error matches.
				gotErr := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, gotErr) {
					if strings.Contains(gotErr, "context deadline exceeded") ||
						strings.Contains(stdout, "context deadline exceeded") {
						contextDeadlineExceeded = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Errorf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, gotErr, tc.errRegex,
					)
				}

				// AND: stdout matches.
				if tc.stdoutRegex != "" && !util.RegexCheck(tc.stdoutRegex, stdout) {
					if strings.Contains(stdout, "context deadline exceeded") {
						contextDeadlineExceeded = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Errorf(
						"%s stdout mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, tc.stdoutRegex,
					)
				}
			}
		})
	}
}

func TestWebHook_ParseTry(t *testing.T) {
	// GIVEN: a WebHook and parseTry inputs.
	tests := []struct {
		name             string
		tryErr           error
		wantSuccessDelta float64
		wantFailDelta    float64
		stdoutRegex      string
	}{
		{
			name:             "success increments SUCCESS and clears fail",
			tryErr:           nil,
			wantSuccessDelta: 1,
			wantFailDelta:    0,
			stdoutRegex:      `^$`,
		},
		{
			name:             "failure increments FAIL and logs",
			tryErr:           fmt.Errorf("boom"),
			wantSuccessDelta: 0,
			wantFailDelta:    1,
			stdoutRegex:      `boom`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			webhook := testWebHook(false, false, false)
			webhook.ID = tc.name
			svcID := webhook.ServiceStatus.ServiceInfo.ID
			webhook.initMetrics()

			hadS := testutil.ToFloat64(
				metric.WebHookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultSuccess, svcID,
				),
			)
			hadF := testutil.ToFloat64(
				metric.WebHookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultFail, svcID,
				),
			)

			logFrom := logx.LogFrom{Primary: webhook.ID, Secondary: svcID}

			// WHEN: parseTry is called.
			webhook.parseTry(tc.tryErr, svcID, logFrom)

			stdout := releaseStdout()
			prefix := fmt.Sprintf(
				"%s\nWebHook.parseTry(err=%v, serviceID=%q)",
				packageName, tc.tryErr, svcID,
			)

			// THEN: Prometheus counters change as expected.
			gotS := testutil.ToFloat64(
				metric.WebHookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultSuccess, svcID,
				),
			)
			gotF := testutil.ToFloat64(
				metric.WebHookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultFail, svcID,
				),
			)
			if resS := gotS - hadS; resS != tc.wantSuccessDelta {
				t.Errorf(
					"%s SUCCESS metric delta mismatch:\ngot:  %v\nwant: %v",
					prefix, resS, tc.wantSuccessDelta,
				)
			}
			if resF := gotF - hadF; resF != tc.wantFailDelta {
				t.Errorf(
					"%s FAIL metric delta mismatch:\ngot:  %v\nwant: %v",
					prefix, resF, tc.wantFailDelta,
				)
			}

			// AND: stdout matches the expected regex.
			if !util.RegexCheck(tc.stdoutRegex, strings.ReplaceAll(stdout, "\n", "-n")) {
				t.Errorf(
					"%s stdout mismatch:\ngot:  %q\nwant regex: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}

			// AND: on success, Failed is set to false.
			gotFail := test.StringifyPtr(webhook.DidFail())
			if tc.tryErr == nil && gotFail != "false" {
				t.Errorf(
					"%s .DidFail() state was not set to false:\ngot:  %s\nwant: false",
					prefix, gotFail,
				)
			}
		})
	}
}
