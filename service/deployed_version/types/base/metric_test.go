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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Metrics(t *testing.T) {
	lookup := lookupImpl{
		Type: "test",
	}
	lookup.Status = &status.Status{
		ServiceInfo: serviceinfo.ServiceInfo{
			ID: t.Name(),
		},
	}

	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	hadG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	lookup.InitMetrics(&lookup)

	test.AssertMetricCount(
		t, "Lookup.InitMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, hadC+2,
	)
	test.AssertMetricCount(
		t, "Lookup.InitMetrics() DeployedVersionQueryResultLast",
		metric.DeployedVersionQueryResultLast, hadG,
	)
	lookup.QueryMetrics(&lookup, nil)
	test.AssertMetricCount(
		t, "Lookup.QueryMetrics() DeployedVersionQueryResultLast",
		metric.DeployedVersionQueryResultLast, hadG+1,
	)

	lookup.DeleteMetrics(&lookup)

	test.AssertMetricCount(
		t, "Lookup.DeleteMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, hadC,
	)
	test.AssertMetricCount(
		t, "Lookup.DeleteMetrics() DeployedVersionQueryResultLast",
		metric.DeployedVersionQueryResultLast, hadG,
	)
}

func TestLookup_QueryMetrics(t *testing.T) {
	type args struct {
		err error
	}

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
				Type: "test",
				Status: &status.Status{
					ServiceInfo: serviceinfo.ServiceInfo{
						ID: "TestLookup_QueryMetrics__" + tc.name,
					},
				},
			}
			lookup.InitMetrics(&lookup)

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
					"Lookup.InitMetrics() metrics mismatch\n"+
						"got:  success_count=%f, fail_count=%f, last_query_result=%f\n"+
						"want: success_count=0, fail_count=0, last_query_result=0",
					counterSuccess, counterFail, gauge,
				)
			}

			lookup.QueryMetrics(&lookup, tc.args.err)

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
					"Lookup.QueryMetrics() metrics mismatch\n"+
						"got:  success_count=%f, fail_count=%f, last_query_result=%f\n"+
						"want: success_count=%f, fail_count=%f, last_query_result=%f",
					counterSuccess, counterFail, gauge,
					wantSuccess, wantFail, wantGauge,
				)
			}
		})
	}
}
