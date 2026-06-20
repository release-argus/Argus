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

// Package base provides the base struct for deployed_version lookups.
//go:build unit

package base

import (
	"errors"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Metrics(t *testing.T) {
	// GIVEN: a Lookup.
	lookup := lookupImpl{
		Lookup: Lookup{
			Type: "test",
		},
	}
	lookup.Status = &status.Status{
		ServiceInfo: serviceinfo.ServiceInfo{
			ID: "TestLookup_Metrics",
		},
	}

	// WHEN: the Lookup's Prometheus metrics are initialised with InitMetrics.
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	hadG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	lookup.InitMetrics(&lookup)

	// THEN: it can be collected.
	// counters:
	gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	wantC := 2
	if resC := gotC - hadC; resC != wantC {
		t.Errorf(
			"%s\nLookup.InitMetrics() Counter metrics (DeployedVersionQueryResultTotal) mismatch\ngot:  %d\nwant: %d",
			packageName, resC, wantC,
		)
	}
	// gauges:
	gotG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	wantG := 0
	if resG := gotG - hadG; resG != wantG {
		t.Errorf(
			"%s\nLookup.InitMetrics() Gauge metrics (DeployedVersionQueryResultLast) mismatch\ngot:  %d\nwant: %d",
			packageName, resG, wantG,
		)
	}
	// But can be added.
	lookup.QueryMetrics(&lookup, nil)
	gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	wantG = 1
	if resG := gotG - hadG; resG != wantG {
		t.Errorf(
			"%s\nLookup.QueryMetrics() Gauge metrics (DeployedVersionQueryResultLast) mismatch\ngot:  %d\nwant: %d",
			packageName, resG, wantG,
		)
	}

	// WHEN: DeleteMetrics is called on the Lookup.
	lookup.DeleteMetrics(&lookup)

	// counters:
	gotC = testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	if gotC != hadC {
		t.Errorf(
			"%s\nLookup.DeleteMetrics() Counter metrics (DeployedVersionQueryResultTotal) mismatch\ngot:  %d\nwant: %d",
			packageName, gotC, hadC,
		)
	}
	// gauges:
	gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	if gotG != hadG {
		t.Errorf(
			"%s\nLookup.DeleteMetrics() Gauge metrics (DeployedVersionQueryResultLast) mismatch\ngot:  %d\nwant: %d",
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
		liveness metric.DeployedVersionQueryResult
	}{
		{
			name: "success",
			args: args{
				err: nil,
			},
			liveness: metric.DeployedVersionQueryResultSuccess,
		},
		{
			name: "regex error",
			args: args{
				err: errors.New("no releases were found matching the stuff"),
			},
			liveness: metric.DeployedVersionQueryResultFailed,
		},
		{
			name: "semantic version error",
			args: args{
				err: errors.New("failed to convert x to a semantic version."),
			},
			liveness: metric.DeployedVersionQueryResultFailed,
		},
		{
			name: "version less than error",
			args: args{
				err: errors.New("queried version x is less than the deployed version"),
			},
			liveness: metric.DeployedVersionQueryResultFailed,
		},
		{
			name: "other error",
			args: args{
				err: errors.New("some other error"),
			},
			liveness: metric.DeployedVersionQueryResultFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := lookupImpl{
				Lookup: Lookup{
					Type: "test",
					Status: &status.Status{
						ServiceInfo: serviceinfo.ServiceInfo{
							ID: fmt.Sprintf("TestLookup_QueryMetrics__%s", tc.name),
						},
					},
				},
			}
			lookup.InitMetrics(&lookup)

			// AND: the Prometheus metrics are initialised to 0.
			counterSuccess := testutil.ToFloat64(
				metric.DeployedVersionQueryResultTotal.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(), metric.ActionResultSuccess,
				),
			)
			counterFail := testutil.ToFloat64(
				metric.DeployedVersionQueryResultTotal.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(), metric.ActionResultFail,
				),
			)
			gauge := testutil.ToFloat64(
				metric.DeployedVersionQueryResultLast.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(),
				),
			)
			if counterSuccess != 0 || counterFail != 0 || gauge != 0 {
				t.Errorf(
					"%s\nLookup.InitMetrics() Metrics were not initialised correctly\n"+
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
				metric.DeployedVersionQueryResultTotal.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(), metric.ActionResultSuccess,
				),
			)
			counterFail = testutil.ToFloat64(
				metric.DeployedVersionQueryResultTotal.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(), metric.ActionResultFail,
				),
			)
			gauge = testutil.ToFloat64(
				metric.DeployedVersionQueryResultLast.WithLabelValues(
					lookup.GetServiceID(), lookup.GetType(),
				),
			)
			if counterSuccess != wantSuccess ||
				counterFail != wantFail ||
				gauge != wantGauge {
				t.Errorf(
					"%s\nLookup QueryMetrics(%v) Metrics mismatch\n"+
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
