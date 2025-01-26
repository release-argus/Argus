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

package deployedver

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

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
			lookup:      testLookup(),
			serviceID:   "TestLookup_Metrics",
			wantMetrics: true,
		},
		"nil": {
			lookup:      nil,
			wantMetrics: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			// WHEN the Prometheus metrics are initialised with initMetrics.
			hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
			hadG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			tc.lookup.InitMetrics()

			// THEN they can be collected.
			// counters:
			gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
			wantC := 2
			if !tc.wantMetrics {
				wantC = hadC
			}
			if (gotC - hadC) != wantC {
				t.Errorf("%d Counter metrics were initialised, expecting %d",
					(gotC - hadC), wantC)
			}
			// gauges - not initialised.
			gotG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			wantG := 0
			if !tc.wantMetrics {
				wantG = hadG
			}
			if (gotG - hadG) != wantG {
				t.Errorf("%d Gauge metrics were initialised, expecting %d",
					(gotG - hadG), wantG)
			}
			// But can be added.
			if tc.lookup != nil {
				tc.lookup.queryMetrics(false)
			}
			gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			wantG = 1
			if !tc.wantMetrics {
				wantG = hadG
			}
			if (gotG - hadG) != wantG {
				t.Errorf("%d Gauge metrics were initialised, expecting %d",
					(gotG - hadG), wantG)
			}

			// AND they can be deleted.
			tc.lookup.DeleteMetrics()
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
		})
	}
}

func TestLookup_Init(t *testing.T) {
	// GIVEN a Lookup and vars for the Init.
	lookup := testLookup()
	defaults := &Defaults{}
	hardDefaults := &Defaults{}
	status := status.Status{ServiceID: test.StringPtr("TestInit")}
	var options opt.Options

	// WHEN Init is called on it.
	lookup.Init(
		&options,
		&status,
		defaults, hardDefaults)

	// THEN pointers to those vars are handed out to the Lookup:
	// 	defaults
	if lookup.Defaults != defaults {
		t.Errorf("Defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			defaults, lookup.Defaults)
	}
	// 	hardDefaults
	if lookup.HardDefaults != hardDefaults {
		t.Errorf("HardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			hardDefaults, lookup.HardDefaults)
	}
	// 	status
	if lookup.Status != &status {
		t.Errorf("Status was not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&status, lookup.Status)
	}
	// 	options
	if lookup.Options != &options {
		t.Errorf("Options were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&options, lookup.Options)
	}

	var nilLookup *Lookup
	nilLookup.Init(
		&options,
		&status,
		defaults, hardDefaults)
}
