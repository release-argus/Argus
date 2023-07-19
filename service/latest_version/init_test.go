// Copyright [2023] [Argus]
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

package latestver

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	metric "github.com/release-argus/Argus/web/metrics"
)

func TestLookup_Metrics(t *testing.T) {
	// GIVEN a Lookup
	lookup := testLookup(false, false)
	*lookup.Status.ServiceID += "TestLookup_Metrics"

	// WHEN the Prometheus metrics are initialised with initMetrics
	hadC := testutil.CollectAndCount(metric.LatestVersionQueryMetric)
	hadG := testutil.CollectAndCount(metric.LatestVersionQueryLiveness)
	lookup.InitMetrics()

	// THEN it can be collected
	// counters
	gotC := testutil.CollectAndCount(metric.LatestVersionQueryMetric)
	wantC := 2
	if (gotC - hadC) != wantC {
		t.Errorf("%d Counter metrics's were initialised, expecting %d",
			(gotC - hadC), wantC)
	}
	// gauges
	gotG := testutil.CollectAndCount(metric.LatestVersionQueryLiveness)
	wantG := 0
	if (gotG - hadG) != wantG {
		t.Errorf("%d Gauge metrics's were initialised, expecting %d",
			(gotG - hadG), wantG)
	}
	// But can be added
	lookup.queryMetrics(nil)
	gotG = testutil.CollectAndCount(metric.LatestVersionQueryLiveness)
	wantG = 1
	if (gotG - hadG) != wantG {
		t.Errorf("%d Gauge metrics's were initialised, expecting %d",
			(gotG - hadG), wantG)
	}

	// AND it can be deleted
	lookup.DeleteMetrics()
	// counters
	gotC = testutil.CollectAndCount(metric.LatestVersionQueryMetric)
	if gotC != hadC {
		t.Errorf("Counter metrics were not deleted, got %d. expecting %d",
			gotC, hadC)
	}
	// gauges
	gotG = testutil.CollectAndCount(metric.LatestVersionQueryLiveness)
	if gotG != hadG {
		t.Errorf("Gauge metrics were not deleted, got %d. expecting %d",
			gotG, hadG)
	}
}

func TestLookup_Init(t *testing.T) {
	// GIVEN a Lookup and vars for the Init
	lookup := testLookup(false, false)
	var defaults LookupDefaults
	var hardDefaults LookupDefaults
	*lookup.Status.ServiceID += "TestInit"
	status := svcstatus.Status{ServiceID: stringPtr("test")}
	var options opt.Options

	// WHEN Init is called on it
	lookup.Init(
		&defaults, &hardDefaults,
		&status,
		&options)

	// THEN pointers to those vars are handed out to the Lookup
	// defaults
	if lookup.Defaults != &defaults {
		t.Errorf("Defaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&defaults, lookup.Defaults)
	}
	// hardDefaults
	if lookup.HardDefaults != &hardDefaults {
		t.Errorf("HardDefaults were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&hardDefaults, lookup.HardDefaults)
	}
	// status
	if lookup.Status != &status {
		t.Errorf("Status was not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&status, lookup.Status)
	}
	// options
	if lookup.Options != &options {
		t.Errorf("Options were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&options, lookup.Options)
	}
}
