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
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Init(t *testing.T) {
	dvCfg := plainDefaultsConfig()

	l := testLookup(t)
	svcStatus := &status.Status{}
	svcStatus.ServiceInfo.ID = "TestInit"
	options := &opt.Options{}

	l.Init(
		options,
		svcStatus,
		dvCfg,
	)

	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: l.Options, Want: options, Mode: test.CompareSamePointer},
		{Name: "Status", Got: l.Status, Want: svcStatus, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: l.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: l.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
		t.Fatal(testErr)
	}
}

func TestLookup_Metrics(t *testing.T) {
	lookup := testLookup(t)

	hadC := testutil.CollectAndCount(metric.DeployedVersionQueryResultTotal)
	hadG := testutil.CollectAndCount(metric.DeployedVersionQueryResultLast)
	lookup.InitMetrics(lookup)

	test.AssertMetricCount(
		t, "Lookup.InitMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, hadC+2,
	)
	test.AssertMetricCount(
		t, "Lookup.InitMetrics() DeployedVersionQueryResultLast",
		metric.DeployedVersionQueryResultLast, hadG,
	)

	lookup.QueryMetrics(lookup, nil)

	test.AssertMetricCount(
		t, "Lookup.QueryMetrics() DeployedVersionQueryResultLast",
		metric.DeployedVersionQueryResultLast, hadG+1,
	)

	lookup.DeleteMetrics(lookup)

	test.AssertMetricCount(
		t, "Lookup.DeleteMetrics() DeployedVersionQueryResultTotal",
		metric.DeployedVersionQueryResultTotal, hadC,
	)
	test.AssertMetricCount(
		t, "Lookup.DeleteMetrics() DeployedVersionQueryResultLast",
		metric.DeployedVersionQueryResultLast, hadG,
	)
}
