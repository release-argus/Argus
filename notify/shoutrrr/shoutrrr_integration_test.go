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

package shoutrrr

import (
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/web/metric"
)

func TestShoutrrrs_Send(t *testing.T) {
	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
		errRegex  string
	}{
		{
			name:      "nil map",
			shoutrrrs: nil,
			errRegex:  `^$`,
		},
		{
			name:      "empty map",
			shoutrrrs: &Shoutrrrs{},
			errRegex:  `^$`,
		},
		{
			name: "single shoutrrr/no error",
			shoutrrrs: &Shoutrrrs{
				"single": testShoutrrr(false, false),
			},
			errRegex: `^$`,
		},
		{
			name: "single shoutrrr/with error",
			shoutrrrs: &Shoutrrrs{
				"single": testShoutrrr(true, false),
			},
			errRegex: `^.*invalid gotify token.* x 1$`,
		},
		{
			name: "multiple shoutrrr/mixed results/2",
			shoutrrrs: &Shoutrrrs{
				"passing": testShoutrrr(false, false),
				"failing": testShoutrrr(true, false),
			},
			errRegex: `^.*invalid gotify token.* x 1$`,
		},
		{
			name: "multiple shoutrrr/mixed results/3",
			shoutrrrs: &Shoutrrrs{
				"passing":      testShoutrrr(false, false),
				"failing":      testShoutrrr(true, false),
				"also_failing": testShoutrrr(true, false),
			},
			errRegex: `^(gotify: failed to build request: invalid gotify token.* x 1\s?){2}$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svcInfo := serviceinfo.ServiceInfo{
				ID: tc.name,
			}

			err := tc.shoutrrrs.Send("TestShoutrrrs_Send", tc.name, svcInfo, false)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestShoutrrr_Send(t *testing.T) {
	tests := []struct {
		name                     string
		shoutrrr                 *Shoutrrr
		delay                    string
		useDelay                 bool
		message                  string
		retries                  int
		deleting, noMetrics      bool
		errRegex                 string
		wantSuccesses, wantFails float64
	}{
		{
			name:     "empty",
			shoutrrr: &Shoutrrr{},
			errRegex: `failed to create Shoutrrr sender`,
		},
		{
			name:      "valid, empty message",
			shoutrrr:  testShoutrrr(false, false),
			message:   "",
			errRegex:  `message cannot be empty`,
			wantFails: 1,
		},
		{
			name:          "valid, with message",
			shoutrrr:      testShoutrrr(false, false),
			message:       "__name__",
			errRegex:      `^$`,
			wantSuccesses: 1,
		},
		{
			name:          "valid, with message, with delay",
			shoutrrr:      testShoutrrr(false, false),
			message:       "__name__",
			useDelay:      true,
			delay:         "1s",
			errRegex:      `^$`,
			wantSuccesses: 1,
		},
		{
			name:      "invalid https cert",
			shoutrrr:  testShoutrrr(false, true),
			message:   "__name__",
			errRegex:  `x509`,
			wantFails: 1,
		},
		{
			name:      "failing",
			shoutrrr:  testShoutrrr(true, true),
			message:   "__name__",
			retries:   1,
			errRegex:  `invalid gotify token.* x 2`,
			wantFails: 2,
		},
		{
			name:     "deleting",
			shoutrrr: testShoutrrr(true, true),
			deleting: true,
			errRegex: `^$`, // parseSend ignores a deleting service's errors.
		},
		{
			name:      "no metrics",
			shoutrrr:  testShoutrrr(false, false),
			message:   "__name__",
			errRegex:  `^$`,
			noMetrics: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svcStatus := newTestStatus("TestShoutrrr_Send - "+tc.name, 1)
			tc.shoutrrr.Init(
				svcStatus,
				&Defaults{},
				&Defaults{}, &Defaults{},
			)
			t.Cleanup(tc.shoutrrr.deleteMetrics)
			if tc.shoutrrr.ServiceStatus != nil && tc.deleting {
				tc.shoutrrr.ServiceStatus.SetDeleting()
			}
			if tc.shoutrrr.Options == nil {
				tc.shoutrrr.Options = map[string]string{}
			}
			if tc.delay != "" {
				tc.shoutrrr.Options["delay"] = tc.delay
			}
			tc.shoutrrr.Options["max_tries"] = strconv.Itoa(tc.retries + 1)

			msg := strings.ReplaceAll(tc.message, "__name__", tc.name)
			err := tc.shoutrrr.Send(
				"test",
				msg,
				svcStatus.ServiceInfo,
				tc.useDelay,
				!tc.noMetrics,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			gotMetric := testutil.ToFloat64(
				metric.NotifyResultTotal.WithLabelValues(
					tc.shoutrrr.ID,
					metric.ActionResultSuccess,
					svcStatus.ServiceInfo.ID,
					tc.shoutrrr.GetType(),
				),
			)
			if gotMetric != tc.wantSuccesses {
				t.Errorf(
					"NotifyResultTotal(ActionResultSuccess) mismatch\ngot:  %f\nwant: %f",
					gotMetric, tc.wantSuccesses,
				)
			}

			gotMetric = testutil.ToFloat64(
				metric.NotifyResultTotal.WithLabelValues(
					tc.shoutrrr.ID,
					metric.ActionResultFail,
					svcStatus.ServiceInfo.ID,
					tc.shoutrrr.GetType(),
				),
			)
			if gotMetric != tc.wantFails {
				t.Errorf(
					"NotifyResultTotal(ActionResultFail) mismatch\ngot:  %f\nwant: %f",
					gotMetric, tc.wantFails,
				)
			}
		})
	}
}
