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

package base

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/web/metric"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		regex    string
		template string
		want     string
		errRegex string
	}{
		{
			name:     "no template returns last group",
			text:     "version=1.2.3",
			regex:    `version=([0-9.]+)`,
			want:     "1.2.3",
			errRegex: `^$`,
		},
		{
			name:     "no group returns whole match",
			text:     "v1.2.3",
			regex:    `[0-9.]+`,
			want:     "1.2.3",
			errRegex: `^$`,
		},
		{
			name:     "template combines groups",
			text:     "v1-2-3",
			regex:    `v([0-9]+)-([0-9]+)-([0-9]+)`,
			template: "$1.$2.$3",
			want:     "1.2.3",
			errRegex: `^$`,
		},
		{
			name:     "no match",
			text:     "nothing here",
			regex:    `v([0-9]+)`,
			errRegex: `^regex "v\(\[0-9\]\+\)" didn't return any matches on "nothing here"$`,
		},
		{
			name:     "empty result",
			text:     "version= ",
			regex:    `version=(.*)`,
			errRegex: `^regex "version=\(\.\*\)" produced an empty version$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ExtractVersion(tc.text, tc.regex, tc.template)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if got != tc.want {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_QueryWithMetrics(t *testing.T) {
	errQuery := errors.New("query failed")

	tests := []struct {
		name        string
		metrics     bool
		queryErr    error
		wantSuccess float64
		wantFail    float64
	}{
		{
			name:        "success with metrics",
			metrics:     true,
			wantSuccess: 1,
		},
		{
			name:     "failure with metrics",
			metrics:  true,
			queryErr: errQuery,
			wantFail: 1,
		},
		{
			name:     "failure without metrics",
			metrics:  false,
			queryErr: errQuery,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := lookupImpl{
				Lookup: Lookup{
					Status: &status.Status{
						ServiceInfo: serviceinfo.ServiceInfo{
							ID: "TestLookup_QueryWithMetrics__" + tc.name,
						},
					},
				},
			}
			lookup.InitMetrics(&lookup)
			var gotMetrics bool
			query := func(metrics bool, _ logx.LogFrom) error {
				gotMetrics = metrics
				return tc.queryErr
			}

			err := lookup.QueryWithMetrics(&lookup, query, tc.metrics, logx.LogFrom{})

			if !errors.Is(err, tc.queryErr) {
				t.Errorf("error mismatch\ngot:  %v\nwant: %v", err, tc.queryErr)
			}
			if gotMetrics != tc.metrics {
				t.Errorf("metrics argument mismatch\ngot:  %t\nwant: %t", gotMetrics, tc.metrics)
			}
			gotSuccess := testutil.ToFloat64(
				metric.DeployedVersionQueryResultTotal.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(), metric.ActionResultSuccess,
				),
			)
			if gotSuccess != tc.wantSuccess {
				t.Errorf("success count mismatch\ngot:  %v\nwant: %v", gotSuccess, tc.wantSuccess)
			}
			gotFail := testutil.ToFloat64(
				metric.DeployedVersionQueryResultTotal.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(), metric.ActionResultFail,
				),
			)
			if gotFail != tc.wantFail {
				t.Errorf("fail count mismatch\ngot:  %v\nwant: %v", gotFail, tc.wantFail)
			}
			lookup.DeleteMetrics(&lookup)
		})
	}
}

func TestTrack__deleting(t *testing.T) {
	lookup := &lookupImpl{
		Lookup: Lookup{
			Status: &status.Status{},
		},
	}
	lookup.Status.SetDeleting()

	// Returns immediately, without querying or sleeping.
	Track(lookup)
}
