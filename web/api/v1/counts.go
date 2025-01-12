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

// Package v1 provides the API for the webserver.
package v1

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	promclient "github.com/prometheus/client_model/go"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

type CountsResponse struct {
	ServiceCount            int `json:"service_count"`
	UpdatesCurrentAvailable int `json:"updates_available"`
	UpdatesCurrentSkipped   int `json:"updates_skipped"`
}

func (api *API) httpCounts(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpServiceSummary", Secondary: getIP(r)}
	jLog.Verbose("-", logFrom, true)

	resp := CountsResponse{}

	// Get service count from metric.
	serviceCountMetric := metric.ServiceCountCurrent

	// Get updates counts from metrics.
	availableMetric := metric.UpdatesCurrent.WithLabelValues("AVAILABLE")
	skippedMetric := metric.UpdatesCurrent.WithLabelValues("SKIPPED")

	// Channel to extract values.
	metricCh := make(chan prometheus.Metric, 1)
	defer close(metricCh)

	// Service count.
	serviceCountMetric.Collect(metricCh)
	if m := <-metricCh; m != nil {
		var parser promclient.Metric
		_ = m.Write(&parser)
		resp.ServiceCount = int(*parser.Gauge.Value)
	}

	// Available updates.
	availableMetric.Collect(metricCh)
	if m := <-metricCh; m != nil {
		var parser promclient.Metric
		_ = m.Write(&parser)
		resp.UpdatesCurrentAvailable = int(*parser.Gauge.Value)
	}

	// Skipped updates.
	skippedMetric.Collect(metricCh)
	if m := <-metricCh; m != nil {
		var parser promclient.Metric
		_ = m.Write(&parser)
		resp.UpdatesCurrentSkipped = int(*parser.Gauge.Value)
	}

	api.writeJSON(w, resp, logFrom)
}
