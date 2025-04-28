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

//go:build integration

package shoutrrr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metric"
)

func TestShoutrrr_Send(t *testing.T) {
	// GIVEN a Shoutrrr to try and send.
	tests := map[string]struct {
		shoutrrr                           *Shoutrrr
		delay                              string
		useDelay                           bool
		message                            string
		retries                            int
		deleting, noMetrics, expectMetrics bool
		errRegex                           string
	}{
		"empty": {
			shoutrrr: &Shoutrrr{},
			errRegex: `failed to create Shoutrrr sender`,
		},
		"valid, empty message": {
			shoutrrr:      testShoutrrr(false, false),
			message:       "",
			errRegex:      `Field 'message' is required`,
			expectMetrics: true,
		},
		"valid, with message": {
			shoutrrr:      testShoutrrr(false, false),
			message:       "__name__",
			errRegex:      `^$`,
			expectMetrics: true,
		},
		"valid, with message, with delay": {
			shoutrrr:      testShoutrrr(false, false),
			message:       "__name__",
			useDelay:      true,
			delay:         "1s",
			errRegex:      `^$`,
			expectMetrics: true,
		},
		"invalid https cert": {
			shoutrrr:      testShoutrrr(false, true),
			errRegex:      `x509`,
			expectMetrics: true,
		},
		"failing": {
			shoutrrr:      testShoutrrr(true, true),
			retries:       1,
			errRegex:      `invalid gotify token .* x 2`,
			expectMetrics: true,
		},
		"deleting": {
			shoutrrr: testShoutrrr(true, true),
			deleting: true,
			errRegex: "",
		},
		"no metrics": {
			shoutrrr:  testShoutrrr(false, false),
			message:   "__name__",
			errRegex:  `^$`,
			noMetrics: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svcStatus := status.Status{}
			serviceID := "TestShoutrrr_Send - " + name
			svcStatus.Init(
				1, 0, 0,
				serviceID, "", "",
				&dashboard.Options{})
			tc.shoutrrr.Init(
				&svcStatus,
				&Defaults{},
				&Defaults{}, &Defaults{})
			t.Cleanup(func() { tc.shoutrrr.deleteMetrics() })
			if tc.shoutrrr.ServiceStatus != nil && tc.deleting {
				tc.shoutrrr.ServiceStatus.SetDeleting()
			}
			if tc.shoutrrr.Options == nil {
				tc.shoutrrr.Options = map[string]string{}
			}
			if tc.delay != "" {
				tc.shoutrrr.Options["delay"] = tc.delay
			}
			tc.shoutrrr.Options["max_tries"] = fmt.Sprint(tc.retries + 1)

			// WHEN send attempted.
			msg := strings.ReplaceAll(tc.message, "__name__", name)
			err := tc.shoutrrr.Send(
				"test",
				msg,
				svcStatus.ServiceInfo,
				tc.useDelay,
				!tc.noMetrics)

			// THEN any error should match the expected regex.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q\n",
					packageName, tc.errRegex, e)
			}
			// AND SUCCESS metrics are recorded as expected.
			var want float64 = 0
			if tc.errRegex == "^$" && tc.expectMetrics {
				want = 1
			}
			gotMetric := testutil.ToFloat64(metric.NotifyResultTotal.WithLabelValues(
				tc.shoutrrr.ID,
				"SUCCESS",
				svcStatus.ServiceInfo.ID,
				tc.shoutrrr.GetType()))
			if gotMetric != want {
				t.Errorf("%s\nwant: %f success metrics\ngot:  %f",
					packageName, want, gotMetric)
			}
			// AND FAILURE metrics are recorded as expected.
			want = 0
			if tc.errRegex != "^$" && tc.expectMetrics {
				want = float64(tc.shoutrrr.GetMaxTries())
			}
			gotMetric = testutil.ToFloat64(metric.NotifyResultTotal.WithLabelValues(
				tc.shoutrrr.ID,
				"FAIL",
				svcStatus.ServiceInfo.ID,
				tc.shoutrrr.GetType()))
			if gotMetric != want {
				t.Errorf("%s\nwant: %f failure metrics\ngot:  %f",
					packageName, want, gotMetric)
			}
		})
	}
}

func TestSlice_Send(t *testing.T) {
	// GIVEN a Slice of Shoutrrr.
	tests := map[string]struct {
		slice    *Slice
		useDelay bool
		errRegex string
	}{
		"nil slice": {
			slice:    nil,
			errRegex: `^$`,
		},
		"empty slice": {
			slice:    &Slice{},
			errRegex: `^$`,
		},
		"single shoutrrr, no error": {
			slice: &Slice{
				"single": testShoutrrr(false, false)},
			errRegex: `^$`,
		},
		"single shoutrrr, with error": {
			slice: &Slice{
				"single": testShoutrrr(true, false)},
			errRegex: `^invalid .* x 1$`,
		},
		"multiple shoutrrr, mixed results": {
			slice: &Slice{
				"passing": testShoutrrr(false, false),
				"failing": testShoutrrr(true, false)},
			errRegex: `^invalid .* x 1$`,
		},
		"multiple shoutrrr, mixed results - more": {
			slice: &Slice{
				"passing":      testShoutrrr(false, false),
				"failing":      testShoutrrr(true, false),
				"also_failing": testShoutrrr(true, false)},
			errRegex: `^(invalid .* x 1\s?){2}$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			serviceInfo := serviceinfo.ServiceInfo{
				ID: name}

			// WHEN Send is called.
			err := tc.slice.Send("TestSlice_Send", name, serviceInfo, tc.useDelay)

			// THEN the expected error state is returned.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}
