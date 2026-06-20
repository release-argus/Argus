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

// Package base provides the base struct for latest_version lookups.
//go:build integration

package base

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Metrics(t *testing.T) {
	// GIVEN: a Lookup.
	lookup := Lookup{
		Type: "test",
	}
	lookup.Status = &status.Status{}
	lookup.Status.ServiceInfo.ID = "TestLookup_Metrics"

	// WHEN: the Prometheus metrics are initialised with initMetrics.
	hadC := testutil.CollectAndCount(metric.LatestVersionQueryResultTotal)
	hadG := testutil.CollectAndCount(metric.LatestVersionQueryResultLast)
	lookup.InitMetrics(&lookup)

	// THEN: it can be collected.
	// counters:
	gotC := testutil.CollectAndCount(metric.LatestVersionQueryResultTotal)
	wantC := 2 + hadC
	if gotC != wantC {
		t.Errorf(
			"%s\nLookup.InitMetrics() Counter metrics mismatch\ngot:  %d\nwant: %d",
			packageName, gotC, wantC,
		)
	}
	// gauges:
	gotG := testutil.CollectAndCount(metric.LatestVersionQueryResultLast)
	wantG := 0 + hadG
	if gotG != wantG {
		t.Errorf(
			"%s\nLookup.InitMetrics() Gauge metrics mismatch\ngot:  %d\nwant: %d",
			packageName, gotG, wantG,
		)
	}
	// But can be added.
	lookup.QueryMetrics(&lookup, nil)
	gotG = testutil.CollectAndCount(metric.LatestVersionQueryResultLast)
	wantG = 1 + hadG
	if gotG != wantG {
		t.Errorf(
			"%s\nLookup.QueryMetrics() Gauge metrics mismatch\ngot:  %d\nwant: %d",
			packageName, gotG, wantG,
		)
	}

	// AND: it can be deleted.
	lookup.DeleteMetrics(&lookup)
	// counters:
	gotC = testutil.CollectAndCount(metric.LatestVersionQueryResultTotal)
	if gotC != hadC {
		t.Errorf(
			"%s\nLookup.DeleteMetrics() Counter metrics mismatch\ngot:  %d\nwant: %d",
			packageName, gotC, hadC,
		)
	}
	// gauges:
	gotG = testutil.CollectAndCount(metric.LatestVersionQueryResultLast)
	if gotG != hadG {
		t.Errorf(
			"%s\nLookup.DeleteMetrics() Gauge metrics mismatch\ngot:  %d\nwant: %d",
			packageName, gotG, hadG,
		)
	}
}

func TestLookup_QueryMetrics(t *testing.T) {
	type args struct {
		err error
	}
	// GIVEN: a Lookup and args for QueryMetrics.
	tests := []struct {
		name     string
		args     args
		liveness metric.LatestVersionQueryResult
	}{
		{
			name: "success",
			args: args{
				err: nil,
			},
			liveness: metric.LatestVersionQueryResultSuccess,
		},
		{
			name: "regex error",
			args: args{
				err: errors.New("no releases were found matching the stuff"),
			},
			liveness: metric.LatestVersionQueryResultNoMatch,
		},
		{
			name: "semantic version error",
			args: args{
				err: errors.New("failed to convert \"x\" to a semantic version."),
			},
			liveness: metric.LatestVersionQueryResultSemanticVersionFail,
		},
		{
			name: "other error",
			args: args{
				err: errors.New("some other error"),
			},
			liveness: metric.LatestVersionQueryResultFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{
				Type:   "test",
				Status: &status.Status{},
			}
			serviceID := "TestLookup_QueryMetrics__" + tc.name
			lookup.Status.ServiceInfo.ID = serviceID
			lookup.InitMetrics(&lookup)
			t.Cleanup(func() {
				lookup.DeleteMetrics(&lookup)
			})

			// AND: the Prometheus metrics are initialised to 0.
			counterSuccess := testutil.ToFloat64(
				metric.LatestVersionQueryResultTotal.WithLabelValues(
					serviceID, lookup.GetType(), metric.ActionResultSuccess,
				),
			)
			counterFail := testutil.ToFloat64(
				metric.LatestVersionQueryResultTotal.WithLabelValues(
					serviceID, lookup.GetType(), metric.ActionResultFail,
				),
			)
			gauge := testutil.ToFloat64(
				metric.LatestVersionQueryResultLast.WithLabelValues(
					serviceID, lookup.GetType(),
				),
			)
			if counterSuccess != 0 || counterFail != 0 || gauge != 0 {
				t.Errorf(
					"%s\nLookup.InitMetrics() did not initialise metrics correctly\n"+
						"Got success_count=%f, fail_count=%f, last_query_result=%f",
					packageName, counterSuccess, counterFail, gauge,
				)
			}

			// WHEN: QueryMetrics is called.
			lookup.QueryMetrics(&lookup, tc.args.err)

			// THEN: the Prometheus metrics are updated.
			wantSuccess := float64(0)
			wantFail := float64(0)
			wantGauge := float64(tc.liveness)
			if tc.args.err == nil {
				wantSuccess = 1
			} else {
				wantFail = 1
			}

			counterSuccess = testutil.ToFloat64(
				metric.LatestVersionQueryResultTotal.WithLabelValues(
					serviceID, lookup.GetType(), metric.ActionResultSuccess,
				),
			)
			counterFail = testutil.ToFloat64(
				metric.LatestVersionQueryResultTotal.WithLabelValues(
					serviceID, lookup.GetType(), metric.ActionResultFail,
				),
			)
			gauge = testutil.ToFloat64(
				metric.LatestVersionQueryResultLast.WithLabelValues(
					serviceID, lookup.GetType(),
				),
			)
			if counterSuccess != wantSuccess ||
				counterFail != wantFail ||
				gauge != wantGauge {
				t.Errorf(
					"%s\nMetrics mismatch after Lookup QueryMetrics(%v)\n"+
						"got:  success_count=%f, fail_count=%f, last_query_result=%f\n"+
						"want: success_count=%f, fail_count=%f, last_query_result=%f",
					packageName, tc.args.err,
					counterSuccess, counterFail, gauge,
					wantSuccess, wantFail, wantGauge,
				)
			}
		})
	}
}
