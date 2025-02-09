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
	"github.com/release-argus/Argus/test"
	metric "github.com/release-argus/Argus/web/metric"
)

func TestLookup_Metrics(t *testing.T) {
	// GIVEN a Lookup.
	lookup := Lookup{
		Type: "test"}
	lookup.Status = &status.Status{
		ServiceID: test.StringPtr("TestLookup_Metrics"),
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
		t.Errorf("%d Counter metrics were initialised, expecting %d",
			(gotC - hadC), wantC)
	}
	// gauges:
	gotG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	wantG := 0
	if (gotG - hadG) != wantG {
		t.Errorf("%d Gauge metrics were initialised, expecting %d",
			(gotG - hadG), wantG)
	}
	// But can be added.
	lookup.QueryMetrics(&lookup, nil)
	gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	wantG = 1
	if (gotG - hadG) != wantG {
		t.Errorf("%d Gauge metrics were initialised, expecting %d",
			(gotG - hadG), wantG)
	}

	// AND it can be deleted.
	lookup.DeleteMetrics(&lookup)
	// counters:
	gotC = testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	if gotC != hadC {
		t.Errorf("Counter metrics were not deleted, got %d. expecting %d",
			gotC, hadC)
	}
	// gauges:
	gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	if gotG != hadG {
		t.Errorf("Gauge metrics were not deleted, got %d. expecting %d",
			gotG, hadG)
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
				err: errors.New("failed converting x to a semantic version.")},
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
					ServiceID: test.StringPtr(
						fmt.Sprintf("TestLookup_QueryMetrics__%s", name))}}
			lookup.InitMetrics(&lookup)

			// AND the Prometheus metrics are initialised to 0.
			counterSuccess := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
				*lookup.Status.ServiceID, lookup.GetType(), "SUCCESS"))
			counterFail := testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
				*lookup.Status.ServiceID, lookup.GetType(), "FAIL"))
			gauge := testutil.ToFloat64(metric.DeployedVersionQueryResultLast.WithLabelValues(
				*lookup.Status.ServiceID, lookup.GetType()))
			if counterSuccess != 0 || counterFail != 0 || gauge != 0 {
				t.Errorf("Metrics were not initialised correctly. Got %f, %f, %f",
					counterSuccess, counterFail, gauge)
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
				*lookup.Status.ServiceID, lookup.GetType(), "SUCCESS"))
			counterFail = testutil.ToFloat64(metric.DeployedVersionQueryResultTotal.WithLabelValues(
				*lookup.Status.ServiceID, lookup.GetType(), "FAIL"))
			gauge = testutil.ToFloat64(metric.DeployedVersionQueryResultLast.WithLabelValues(
				*lookup.Status.ServiceID, lookup.GetType()))
			if counterSuccess != wantSuccess || counterFail != wantFail || gauge != wantGauge {
				t.Errorf("Metrics were not updated correctly.\nGot:  %f, %f, %f\nWant: %f, %f, %f",
					counterSuccess, counterFail, gauge,
					wantSuccess, wantFail, wantGauge)
			}
		})
	}
}
