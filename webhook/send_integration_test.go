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
	"errors"
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

func TestWebhooks_Send(t *testing.T) {
	tests := []struct {
		name                  string
		webhooks              *Webhooks
		stdoutRegex, errRegex string
		useDelay              bool
		delays                map[string]string
		repeat                int
	}{
		{
			name:        "nil map",
			webhooks:    nil,
			stdoutRegex: `^$`,
			errRegex:    `^$`,
		},
		{
			name: "2 successful webhooks",
			webhooks: &Webhooks{
				"pass": testWebhook(false, false),
				"fail": testWebhook(false, false),
			},
			stdoutRegex: `Webhook received.*Webhook received`,
			errRegex:    `^$`,
		},
		{
			name: "successful and failing defaults",
			webhooks: &Webhooks{
				"pass": testWebhook(false, false),
				"fail": testWebhook(true, false),
			},
			stdoutRegex: `Webhook received.*failed \d times to send the Webhook|` +
				`failed \d times to send the Webhook.*Webhook received`,
			errRegex: `failed 1 times to send the Webhook for .* to "fail"$`,
		},
		{
			name: "does apply defaults delay",
			webhooks: &Webhooks{
				"pass": testWebhook(false, false),
				"fail": testWebhook(true, false),
			},
			stdoutRegex: `Webhook received.*failed \d times to send the Webhook`,
			errRegex:    `failed 1 times to send the Webhook for .* to "fail"$`,
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
			// Serial: captures stdout.

		tries:
			for try := 1; ; try++ {
				// Repeat to check delay usage, as map order is random.
				for repeat := tc.repeat; repeat >= 0; repeat-- {
					t.Logf("repeat %d", repeat)
					releaseStdout := test.CaptureLog(t, logx.Default())
					if tc.webhooks != nil {
						for id := range *tc.webhooks {
							(*tc.webhooks)[id].ID = id
						}
						for id := range tc.delays {
							(*tc.webhooks)[id].Delay = tc.delays[id]
						}
					}

					err := tc.webhooks.Send(serviceinfo.ServiceInfo{ID: tc.name}, tc.useDelay)

					stdout := releaseStdout()
					stdout = strings.ReplaceAll(stdout, "\n", "-n")
					if !util.RegexCheck(tc.stdoutRegex, stdout) {
						if try < 3 && strings.Contains(stdout, "context deadline exceeded") {
							time.Sleep(time.Second)
							continue tries
						}
						t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
					}

					e := errfmt.FormatError(err)
					if !util.RegexCheck(tc.errRegex, e) {
						t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
					}
				}
				break
			}
		})
	}
}

func TestWebhook_Send(t *testing.T) {
	tests := []struct {
		name                                                string
		headers, wouldFail, useDelay, deleting, silentFails bool
		delay                                               string
		retries                                             uint8
		notifiers                                           shoutrrr.Shoutrrrs
		stdoutRegex, errRegex                               string
	}{
		{
			name:        "successful defaults",
			stdoutRegex: `Webhook received`,
			errRegex:    `^$`,
		},
		{
			name:        "successful defaults with headers",
			stdoutRegex: `Webhook received`,
			errRegex:    `^$`,
			headers:     true,
		},
		{
			name:        "does use delay",
			useDelay:    true,
			delay:       "3s",
			stdoutRegex: `Webhook received`,
			errRegex:    `^$`,
		},
		{
			name:        "no delay",
			useDelay:    true,
			delay:       "0s",
			stdoutRegex: `Webhook received`,
			errRegex:    `^$`,
		},
		{
			name:        "failing defaults",
			wouldFail:   true,
			stdoutRegex: `failed \d times to send`,
			errRegex:    `failed 1 times to send the Webhook for .* to ".*"$`,
		},
		{
			name:        "failing defaults with headers",
			wouldFail:   true,
			headers:     true,
			stdoutRegex: `failed \d times to send`,
			errRegex:    `failed 1 times to send the Webhook for .* to ".*"$`,
		},
		{
			name:        "retries multiple times",
			wouldFail:   true,
			retries:     2,
			stdoutRegex: `(Webhook gave 500.*){2}Webhook received`,
			errRegex:    `^$`,
		},
		{
			name:        "does try notifiers on fail",
			wouldFail:   true,
			stdoutRegex: `Webhook gave 500.*\(401\)`,
			errRegex:    `failed 1 times to send the Webhook for .* to ".*"$`,
			notifiers: shoutrrr.Shoutrrrs{
				"fail": shoutrrrtest.Shoutrrr(true, false),
			},
		},
		{
			name:        "doesn't try notifiers on fail if silentFails",
			wouldFail:   true,
			silentFails: true,
			stdoutRegex: `Webhook gave 500.*failed \d times to send the Webhook [^-]+-n$`,
			errRegex:    `failed 1 times to send the Webhook for .* to ".*"$`,
			notifiers: shoutrrr.Shoutrrrs{
				"fail": shoutrrrtest.Shoutrrr(true, false),
			},
		},
		{
			name:        "doesn't send if deleting",
			deleting:    true,
			stdoutRegex: `^$`,
			errRegex:    `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			for try := 1; ; try++ {
				releaseStdout := test.CaptureLog(t, logx.Default())
				webhook := testWebhook(tc.wouldFail, tc.headers)
				if tc.deleting {
					webhook.ServiceStatus.SetDeleting()
				}
				webhook.Delay = tc.delay
				webhook.MaxTries = new(tc.retries + 1)
				webhook.SilentFails = &tc.silentFails
				webhook.Notifiers = Notifiers{Shoutrrr: &tc.notifiers}
				webhook.ServiceStatus.ServiceInfo.ID = tc.name
				svcInfo := webhook.ServiceStatus.GetServiceInfo()
				done := make(chan struct{})
				// Fixes the Secret once the expected failures are seen, so a retry can succeed.
				if tc.retries > 0 {
					go func() {
						for testutil.ToFloat64(
							metric.WebhookResultTotal.WithLabelValues(
								webhook.ID, metric.ActionResultFail, svcInfo.ID,
							),
						) < float64(tc.retries) {
							select {
							case <-done:
								return
							case <-time.After(200 * time.Millisecond):
							}
						}
						webhook.mu.Lock()
						webhook.Secret = "argus"
						webhook.mu.Unlock()
					}()
				}

				startAt := time.Now()
				err := webhook.Send(svcInfo, tc.useDelay)
				close(done)

				completedAt := time.Now()
				stdout := releaseStdout()
				stdout = strings.ReplaceAll(stdout, "\n", "-n")
				if !util.RegexCheck(tc.stdoutRegex, stdout) {
					if try < 3 && strings.Contains(stdout, "context deadline exceeded") {
						time.Sleep(time.Second)
						continue
					}
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
				}

				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
				}

				if tc.delay != "" {
					delayDuration, err := time.ParseDuration(tc.delay)
					if err != nil {
						t.Fatalf("time.ParseDuration() failed: %v", err)
					}
					took := completedAt.Sub(startAt)
					if took < delayDuration {
						t.Errorf(
							"Webhook.Send() duration mismatch\ngot:  %s\nwant: >= %s",
							took, tc.delay,
						)
					} else if took > delayDuration+2*time.Second {
						t.Errorf(
							"Webhook.Send() duration mismatch\ngot:  %s\nwant: <= %s+2s",
							took, tc.delay,
						)
					}
				}
				break
			}
		})
	}
}

func TestNotifiers_Send(t *testing.T) {
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
				"pass": shoutrrrtest.Shoutrrr(false, false),
			},
		},
		{
			name:     "failing",
			errRegex: `\(401\)`,
			shoutrrrNotifiers: &shoutrrr.Shoutrrrs{
				"fail": shoutrrrtest.Shoutrrr(true, false),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifiers := Notifiers{Shoutrrr: tc.shoutrrrNotifiers}

			err := notifiers.Send(
				"title",
				tc.name,
				serviceinfo.ServiceInfo{ID: tc.name},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestWebhook_Try(t *testing.T) {
	headerAuthURL := test.HeaderAuthServer(t, "1.2.3")
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
			url:         test.WebhookGitHub.URLValid,
			secret:      test.WebhookGitHub.SecretPass,
			whType:      "github",
			errRegex:    `^$`,
			stdoutRegex: `Webhook received`,
		},
		{
			name:     "failure on wrong secret",
			url:      test.WebhookGitHub.URLValid,
			secret:   test.WebhookGitHub.SecretFail,
			whType:   "github",
			errRegex: `Webhook gave 500`,
		},
		{
			name:   "success with header auth",
			url:    headerAuthURL,
			secret: test.WebhookGitHub.SecretPass,
			whType: "github",
			headers: Headers{
				{
					Key:   test.LookupWithHeaderAuth.HeaderKey,
					Value: test.LookupWithHeaderAuth.HeaderValuePass,
				},
			},
			errRegex:    `^$`,
			stdoutRegex: `Webhook received`,
		},
		{
			name:   "failure with invalid header auth",
			url:    headerAuthURL,
			secret: test.WebhookGitHub.SecretPass,
			whType: "github",
			headers: Headers{
				{
					Key:   test.LookupWithHeaderAuth.HeaderKey,
					Value: test.LookupWithHeaderAuth.HeaderValueFail,
				},
			},
			errRegex: `(?s)Webhook gave 200, not 2XX.*Hook rules were not satisfied`,
		},
		{
			name:     "rejects invalid TLS certificate",
			url:      test.WebhookGitHub.URLInvalid,
			secret:   test.WebhookGitHub.SecretPass,
			whType:   "github",
			errRegex: `(?i)certificate`,
		},
		{
			name:              "allows invalid TLS certificate when configured",
			url:               test.WebhookGitHub.URLInvalid,
			secret:            test.WebhookGitHub.SecretPass,
			whType:            "github",
			allowInvalidCerts: new(true),
			errRegex:          `^$`,
			stdoutRegex:       `Webhook received`,
		},
		{
			name:        "unsupported webhook type",
			url:         test.WebhookGitHub.URLValid,
			secret:      test.WebhookGitHub.SecretPass,
			whType:      "url",
			errRegex:    `failed to get \*http.request for Webhook`,
			stdoutRegex: `failed to get \*http.request for Webhook`,
		},
		{
			name:              "wrong desired status code",
			url:               test.WebhookGitHub.URLValid,
			secret:            test.WebhookGitHub.SecretPass,
			whType:            "github",
			desiredStatusCode: new(uint16(404)),
			errRegex:          `Webhook gave 200, not 404`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			for try := 1; ; try++ {
				releaseStdout := test.CaptureLog(t, logx.Default())

				webhook := testWebhook(false, false)
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

				err := webhook.try(logFrom)

				stdout := releaseStdout()

				gotErr := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, gotErr) {
					if try < 3 && (strings.Contains(gotErr, "context deadline exceeded") ||
						strings.Contains(stdout, "context deadline exceeded")) {
						time.Sleep(time.Second)
						continue
					}
					t.Errorf("error mismatch\ngot:  %q\nwant: %q", gotErr, tc.errRegex)
				}

				if tc.stdoutRegex != "" && !util.RegexCheck(tc.stdoutRegex, stdout) {
					if try < 3 && strings.Contains(stdout, "context deadline exceeded") {
						time.Sleep(time.Second)
						continue
					}
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
				}
				break
			}
		})
	}
}

func TestWebhook_ParseTry(t *testing.T) {
	tests := []struct {
		name             string
		tryErr           error
		wantSuccessDelta float64
		wantFailDelta    float64
		wantDidFail      string
		stdoutRegex      string
	}{
		{
			name:             "success increments SUCCESS and clears fail",
			tryErr:           nil,
			wantSuccessDelta: 1,
			wantFailDelta:    0,
			wantDidFail:      "false",
			stdoutRegex:      `^$`,
		},
		{
			name:             "failure increments FAIL and logs",
			tryErr:           errors.New("boom"),
			wantSuccessDelta: 0,
			wantFailDelta:    1,
			wantDidFail:      "true",
			stdoutRegex:      `boom`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			webhook := testWebhook(false, false)
			webhook.ID = tc.name
			svcID := webhook.ServiceStatus.ServiceInfo.ID
			webhook.initMetrics()
			webhook.SetFail(new(true))

			hadS := testutil.ToFloat64(
				metric.WebhookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultSuccess, svcID,
				),
			)
			hadF := testutil.ToFloat64(
				metric.WebhookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultFail, svcID,
				),
			)

			logFrom := logx.LogFrom{Primary: webhook.ID, Secondary: svcID}

			webhook.parseTry(tc.tryErr, svcID, logFrom)

			stdout := releaseStdout()

			gotS := testutil.ToFloat64(
				metric.WebhookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultSuccess, svcID,
				),
			)
			gotF := testutil.ToFloat64(
				metric.WebhookResultTotal.WithLabelValues(
					webhook.ID, metric.ActionResultFail, svcID,
				),
			)
			if resS := gotS - hadS; resS != tc.wantSuccessDelta {
				t.Errorf(
					"WebhookResultTotal SUCCESS delta mismatch\ngot:  %v\nwant: %v",
					resS, tc.wantSuccessDelta,
				)
			}
			if resF := gotF - hadF; resF != tc.wantFailDelta {
				t.Errorf(
					"WebhookResultTotal FAIL delta mismatch\ngot:  %v\nwant: %v",
					resF, tc.wantFailDelta,
				)
			}

			if !util.RegexCheck(tc.stdoutRegex, strings.ReplaceAll(stdout, "\n", "-n")) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
			}

			if gotFail := test.StringifyPtr(webhook.DidFail()); gotFail != tc.wantDidFail {
				t.Errorf("Webhook.DidFail() mismatch\ngot:  %s\nwant: %s", gotFail, tc.wantDidFail)
			}
		})
	}
}
