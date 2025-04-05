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

//go:build unit

package web

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	metric "github.com/release-argus/Argus/web/metric"
)

func TestLookup_Metrics(t *testing.T) {
	// GIVEN a Lookup pointer.
	tests := map[string]struct {
		lookup      *Lookup
		serviceID   string
		wantMetrics bool
	}{
		"non-nil": {
			lookup:      testLookup(false),
			serviceID:   "TestLookup_Metrics",
			wantMetrics: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			// WHEN the Prometheus metrics are initialised with initMetrics.
			hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
			hadG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			tc.lookup.InitMetrics(tc.lookup)

			// THEN they can be collected.
			// counters:
			gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
			wantC := hadC
			if tc.wantMetrics {
				wantC += 2
			}
			if gotC != wantC {
				t.Errorf("%s\nCounter metrics mismatch after InitMetrics()\nwant: %d\ngot:  %d",
					packageName, wantC, gotC)
			}
			// gauges - not initialised.
			gotG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			wantG := hadG
			if gotG != wantG {
				t.Errorf("%s\nGauge metrics mismatch after InitMetrics()\nwant: %d\ngot:  %d",
					packageName, wantG, gotG)
			}
			// But can be added.
			if tc.lookup != nil {
				tc.lookup.QueryMetrics(tc.lookup, nil)
			}
			gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			wantG = hadG
			if tc.wantMetrics {
				wantG += 1
			}
			if gotG != wantG {
				t.Errorf("%s\nGauge metrics mismatch after QueryMetrics()\nwant: %d\ngot:  %d",
					packageName, wantG, gotG)
			}

			// AND they can be deleted.
			tc.lookup.DeleteMetrics(tc.lookup)
			// counters:
			gotC = testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
			if gotC != hadC {
				t.Errorf("%s\nCounter metrics mismatch after DeleteMetrics()\nwant: %d\ngot:  %d",
					packageName, hadC, gotC)
			}
			// gauges:
			gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			if gotG != hadG {
				t.Errorf("%s\nGauge metrics mismatch after DeleteMetrics()\nwant: %d\ngot:  %d",
					packageName, hadG, gotG)
			}
		})
	}
}

func TestLookup_Init(t *testing.T) {
	// GIVEN a Lookup and vars for the Init.
	lookup := testLookup(false)
	defaults := &base.Defaults{}
	hardDefaults := &base.Defaults{}
	status := status.Status{ServiceID: test.StringPtr("TestInit")}
	var options opt.Options

	// WHEN Init is called on it.
	lookup.Init(
		&options,
		&status,
		defaults, hardDefaults)

	// THEN pointers to those vars are handed out to the Lookup:
	// 	Defaults.
	if lookup.Defaults != defaults {
		t.Errorf("%s\nDefaults mismatch\nwant: %v\ngot:  %v",
			packageName, defaults, lookup.Defaults)
	}
	// 	HardDefaults.
	if lookup.HardDefaults != hardDefaults {
		t.Errorf("%s\nHardDefaults mismatch\nwant: %v\ngot:  %v",
			packageName, hardDefaults, lookup.HardDefaults)
	}
	// 	Status.
	if lookup.Status != &status {
		t.Errorf("%s\nStatus mismatch\nwant: %v\ngot:  %v",
			packageName, &status, lookup.Status)
	}
	// 	Options.
	if lookup.Options != &options {
		t.Errorf("%s\nOptions mismatch\nwant: %v\ngot:  %v",
			packageName, &options, lookup.Options)
	}
}
