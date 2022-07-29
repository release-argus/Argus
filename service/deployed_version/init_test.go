// Copyright [2022] [Argus]
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

package deployed_version

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

func TestInitMetrics(t *testing.T) {
	// GIVEN a Lookup
	lookup := testDeployedVersion()

	// WHEN the Prometheus metrics are initialised with initMetrics
	hadC := testutil.CollectAndCount(metrics.DeployedVersionQueryMetric)
	lookup.initMetrics()

	// THEN it can be collected
	// counters
	gotC := testutil.CollectAndCount(metrics.DeployedVersionQueryMetric)
	wantC := 2
	if (gotC - hadC) != wantC {
		t.Errorf("%d Counter metrics's were initialised, expecting %d",
			(gotC - hadC), wantC)
	}
}

func TestInit(t *testing.T) {
	// GIVEN a Lookup and vars for the Init
	lookup := testDeployedVersion()
	log := utils.NewJLog("WARN", false)
	var defaults *Lookup = &Lookup{}
	var hardDefaults *Lookup = &Lookup{}
	status := service_status.Status{ServiceID: stringPtr("TestInit")}
	var options options.Options

	// WHEN Init is called on it
	hadC := testutil.CollectAndCount(metrics.DeployedVersionQueryMetric)
	lookup.Init(log, &defaults, &hardDefaults, &status, &options)

	// THEN pointers to those vars are handed out to the Lookup
	// log
	if jLog != log {
		t.Errorf("JLog was not initialised from the Init\n want: %v\ngot:  %v",
			log, jLog)
	}
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
	if lookup.options != &options {
		t.Errorf("Options were not handed to the Lookup correctly\n want: %v\ngot:  %v",
			&options, lookup.options)
	}
	// initMetrics - counters
	gotC := testutil.CollectAndCount(metrics.DeployedVersionQueryMetric)
	wantC := 2
	if (gotC - hadC) != wantC {
		t.Errorf("%d Counter metrics's were initialised, expecting %d\ngot=%d, had=%d",
			(gotC - hadC), wantC, gotC, hadC)
	}

	var nilLookup *Lookup
	nilLookup.Init(log, &defaults, &hardDefaults, &status, &options)
	if nilLookup != nil {
		t.Error("Init on nil shouldn't have initialised the Lookup")
	}
}