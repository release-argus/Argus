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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// AssertMetricCount checks that c collects want metrics.
// subject names what is counted, e.g. "Lookup.InitMetrics() LatestVersionQueryResultTotal".
func AssertMetricCount(t TB, subject string, c prometheus.Collector, want int) {
	t.Helper()

	if got := testutil.CollectAndCount(c); got != want {
		t.Errorf("%s count mismatch\ngot:  %d\nwant: %d", subject, got, want)
	}
}
