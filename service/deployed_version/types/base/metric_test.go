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
	// GIVEN a Lookup.
	lookup := Lookup{
		Type: "test"}
	lookup.Status = &status.Status{
		ServiceInfo: serviceinfo.ServiceInfo{
			ID: "TestLookup_Metrics"},
	}

	// WHEN the Prometheus metrics are initialised with initMetrics.
	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	hadG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	lookup.InitMetrics(&lookup)

	// THEN it can be collected.
	// counters:
	gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	wantC := 2
	if (gotC - hadC) != wantC {
		t.Errorf("%s\nCount metrics mismatch after InitMetrics()\nwant: %d\ngot:  %d",
			packageName, wantC, gotC - hadC)
	}
	// gauges:
	gotG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	wantG := 0
	if (gotG - hadG) != wantG {
		t.Errorf("%s\nGauge metrics mismatch after InitMetrics()\nwant: %d\ngot:  %d",
			packageName, wantG, gotG - hadG)
	}
	// But can be added.
	lookup.QueryMetrics(&lookup, nil)
	gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	wantG = 1
	if (gotG - hadG) != wantG {
		t.Errorf("%s\nGauge metrics mismatch after QueryMetrics()\nwant: %d\ngot:  %d",
			packageName, wantG, gotG - hadG)
	}

	// AND it can be deleted.
	lookup.DeleteMetrics(&lookup)
	// counters:
	gotC = testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	if gotC != hadC {
		t.Errorf("%s\nCount metrics mismatch after DeleteMetrics()\nwant: %d\ngot:  %d",
			packageName, hadC, gotC)
	}
	// gauges:
	gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	if gotG != hadG {
		t.Errorf("%s\nGauge metrics mismatch after DeleteMetrics()\nwant: %d\ngot:  %d",
			packageName, hadG, gotG)
	}
}

func TestLookup_QueryMetrics(t *testing.T) {
	type args struct {
		err error
	}
	// GIVEN a Lookup and args for QueryMetrics.
	tests := map[string]struct {
		args     args
		liveness int
	}{
		"success": {
			args: args{
				err: nil},
			liveness: 1,
		},
		"regex error": {
			args: args{
				err: errors.New("no releases were found matching the stuff")},
			liveness: 0,
		},
		"semantic version error": {
			args: args{
				err: errors.New("failed to convert x to a semantic version.")},
			liveness: 0,
		},
		"version less than error": {
			args: args{
				err: errors.New("queried version x is less than the deployed version")},
			liveness: 0,
		},
		"other error": {
			args: args{
				err: errors.New("some other error")},
			liveness: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{
				Type: "test",
				Status: &status.Status{
					ServiceInfo: serviceinfo.ServiceInfo{
						ID: fmt.Sprintf("TestLookup_QueryMetrics__%s", name)}},
			}
			lookup.InitMetrics(&lookup)

			// AND the Prometheus metrics are initialised to 0.
			counterSuccess := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
				lookup.GetServiceID(), lookup.GetType(), "SUCCESS"))
			counterFail := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
				lookup.GetServiceID(), lookup.GetType(), "FAIL"))
			gauge := testutil.ToFloat64(metric.DeployedVersionQueryResultLast.WithLabelValues(
				lookup.GetServiceID(), lookup.GetType()))
			if counterSuccess != 0 || counterFail != 0 || gauge != 0 {
				t.Errorf("%s\nMetrics were not initialised correctly. Got %f, %f, %f",
					packageName, counterSuccess, counterFail, gauge)
			}

			// WHEN QueryMetrics is called.
			lookup.QueryMetrics(&lookup, tc.args.err)

			// THEN the Prometheus metrics are updated.
			wantSuccess := float64(0)
			wantFail := float64(0)
			wantGauge := float64(tc.liveness)
			if tc.args.err == nil {
				wantSuccess = 1
			} else {
				wantFail = 1
			}

			counterSuccess = testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
				lookup.GetServiceID(), lookup.GetType(), "SUCCESS"))
			counterFail = testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
				lookup.GetServiceID(), lookup.GetType(), "FAIL"))
			gauge = testutil.ToFloat64(metric.DeployedVersionQueryResultLast.WithLabelValues(
				lookup.GetServiceID(), lookup.GetType()))
			if counterSuccess != wantSuccess ||
				counterFail != wantFail ||
				gauge != wantGauge {
				t.Errorf("%s\nmetrics mismatch\nwant: %f, %f, %f\ngot:  %f, %f, %f",
					packageName,
					wantSuccess, wantFail, wantGauge,
					counterSuccess, counterFail, gauge)
			}
		})
	}
}
