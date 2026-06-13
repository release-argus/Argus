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

package web

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/internal/test"

	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Init(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)

	// GIVEN: a Lookup and vars for the Init.
	l := testLookup(t, false)
	svcStatus := &status.Status{}
	svcStatus.ServiceInfo.ID = "TestInit"
	options := &opt.Options{}

	// WHEN: Init is called on it.
	l.Init(
		options,
		svcStatus,
		dvCfg,
	)

	prefix := fmt.Sprintf(
		"%s\nLookup.Init(options=%p, status=%p, defaults=%v)",
		packageName, options, &svcStatus, dvCfg,
	)

	// THEN: pointers to those vars are handed out to the Lookup.
	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: l.Options, Want: options, Mode: test.CompareSamePointer},
		{Name: "Status", Got: l.Status, Want: svcStatus, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: l.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: l.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
		t.Fatal(err)
	}
}

func TestLookup_Metrics(t *testing.T) {
	// GIVEN: a Lookup pointer.
	tests := []struct {
		name        string
		lookup      *Lookup
		serviceID   string
		wantMetrics bool
	}{
		{
			name:        "non-nil",
			lookup:      testLookup(t, false),
			serviceID:   "TestLookup_Metrics",
			wantMetrics: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			// WHEN: the Prometheus metrics are initialised with initMetrics.
			hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
			hadG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			tc.lookup.InitMetrics(tc.lookup)

			// THEN: they can be collected.
			// counters:
			gotC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
			wantC := hadC
			if tc.wantMetrics {
				wantC += 2
			}
			if gotC != wantC {
				t.Errorf(
					"%s\nLookup.InitMetrics() Counter metrics mismatch after\ngot:  %d\nwant: %d",
					packageName, gotC, wantC,
				)
			}
			// gauges - not initialised.
			gotG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
			wantG := hadG
			if gotG != wantG {
				t.Errorf(
					"%s\nLookup.InitMetrics() Gauge metrics mismatch\ngot:  %d\nwant: %d",
					packageName, gotG, wantG,
				)
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
				t.Errorf(
					"%s\nLookup.QueryMetrics() Gauge metrics mismatch\ngot:  %d\nwant: %d",
					packageName, gotG, wantG,
				)
			}

			// AND: they can be deleted.
			tc.lookup.DeleteMetrics(tc.lookup)
			// counters:
			if gotC = testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal); gotC != hadC {
				t.Errorf(
					"%s\nLookup.DeleteMetrics() Counter metrics mismatch\ngot:  %d\nwant: %d",
					packageName, gotC, hadC,
				)
			}
			// gauges:
			if gotG = testutil.CollectAndCount(metric.DeployedVersionQueryResultLast); gotG != hadG {
				t.Errorf(
					"%s\nLookup.DeleteMetrics() Gauge metrics mismatch\ngot:  %d\nwant: %d",
					packageName, gotG, hadG,
				)
			}
		})
	}
}
