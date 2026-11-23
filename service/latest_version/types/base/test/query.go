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

//go:build unit || integration

package test

import (
	"github.com/prometheus/client_golang/prometheus/testutil"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/web/metric"
)

// AssertQueryNoSideEffects runs a writeToDB=false Query of lookup from startVersion, and asserts
// it finds wantVersion without announcing, writing to the DB, or recording metrics.
func AssertQueryNoSideEffects(t test.TB, lookup base.Interface, startVersion, wantVersion string) {
	t.Helper()

	status := lookup.GetStatus()
	status.ServiceInfo.ID = t.Name()
	announceChannel := make(chan []byte, 8)
	databaseChannel := make(chan dbtype.Message, 8)
	status.AnnounceChannel = announceChannel
	status.DatabaseChannel = databaseChannel
	status.SetLatestVersion(startVersion, "", false)
	// Distinct from every state, so a write is visible.
	sentinel := 99.0
	gauge := metric.LatestVersionIsDeployed.WithLabelValues(t.Name())
	gauge.Set(sentinel)
	t.Cleanup(func() { metric.DeletePrometheusGauge(metric.LatestVersionIsDeployed, t.Name(), "") })

	if _, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()}); err != nil {
		t.Fatalf("Lookup.Query() failed: %s", errfmt.FormatError(err))
	}

	if testErr := test.AssertFields(t, []test.FieldAssertion{
		{Name: "Status.LatestVersion()", Got: status.LatestVersion(), Want: wantVersion, Mode: test.CompareEqual},
		{Name: "len(Status.AnnounceChannel)", Got: len(announceChannel), Want: 0, Mode: test.CompareEqual},
		{Name: "len(Status.DatabaseChannel)", Got: len(databaseChannel), Want: 0, Mode: test.CompareEqual},
		{Name: "metric.LatestVersionIsDeployed", Got: testutil.ToFloat64(gauge), Want: sentinel, Mode: test.CompareEqual},
		{Name: "metric.LatestVersionQueryResultTotal", Got: testutil.ToFloat64(metric.LatestVersionQueryResultTotal.WithLabelValues(t.Name(), lookup.GetType(), metric.ActionResultSuccess)), Want: 0.0, Mode: test.CompareEqual},
	}, ""); testErr != nil {
		t.Errorf("%v", testErr)
	}
}
