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

	"github.com/release-argus/Argus/config"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/web/metric"
)

type UpdateDetails struct {
	ServiceName     string `json:"service_name"`
	DeployedVersion string `json:"deployed_version"`
	LatestVersion   string `json:"latest_version"`
	LastChecked     string `json:"last_checked"`
	AutoApprove     bool   `json:"auto_approve"`
	Approved        bool   `json:"approved"`
	Skipped         bool   `json:"skipped"`
}

func getUpdateDetails(cfg *config.Config, length int) []UpdateDetails {
	updateDetails := make([]UpdateDetails, 0, length)
	for _, id := range cfg.Order {
		svc := cfg.Service[id]
		svcInfo := svc.Status.GetServiceInfo()
		// Skip services that have the latest version deployed.
		if svcInfo.DeployedVersion == svcInfo.LatestVersion {
			continue
		}

		updateApproved := svcInfo.ApprovedVersion == svcInfo.LatestVersion
		updateSkipped := !updateApproved && svcInfo.ApprovedVersion == "SKIP_"+svcInfo.LatestVersion
		updateDetails = append(updateDetails, UpdateDetails{
			ServiceName:     id,
			DeployedVersion: svcInfo.DeployedVersion,
			LatestVersion:   svcInfo.LatestVersion,
			LastChecked:     svc.Status.LastQueried(),
			AutoApprove:     svc.Dashboard.GetAutoApprove(),
			Approved:        updateApproved,
			Skipped:         updateSkipped,
		})
	}

	return updateDetails
}

type CountsResponse struct {
	ServiceCount            int             `json:"service_count"`
	ServiceCountActive      int             `json:"service_count_active"`
	ServiceCountInactive    int             `json:"service_count_inactive"`
	UpdatesCurrentAvailable int             `json:"updates_available"`
	UpdatesCurrentSkipped   int             `json:"updates_skipped"`
	UpdateDetails           []UpdateDetails `json:"update_details,omitempty"`
}

func getGaugeValue(g prometheus.Gauge) float64 {
	var m promclient.Metric
	_ = g.Write(&m)
	return m.GetGauge().GetValue()
}

func (api *API) httpCounts(w http.ResponseWriter, r *http.Request) {
	logFrom := logutil.LogFrom{Primary: "httpServiceSummary", Secondary: getIP(r)}

	resp := CountsResponse{}

	// Get service count from metric.
	serviceCountActiveMetric := metric.ServiceCountCurrent.WithLabelValues(metric.ServiceStateActive)
	serviceCountInactiveMetric := metric.ServiceCountCurrent.WithLabelValues(metric.ServiceStateInactive)
	resp.ServiceCountActive = int(getGaugeValue(serviceCountActiveMetric))
	resp.ServiceCountInactive = int(getGaugeValue(serviceCountInactiveMetric))
	resp.ServiceCount = resp.ServiceCountActive + resp.ServiceCountInactive

	// Get updates counts from metrics.
	availableMetric := metric.UpdatesCurrent.WithLabelValues("AVAILABLE")
	skippedMetric := metric.UpdatesCurrent.WithLabelValues("SKIPPED")
	resp.UpdatesCurrentAvailable = int(getGaugeValue(availableMetric))
	resp.UpdatesCurrentSkipped = int(getGaugeValue(skippedMetric))

	// Get update details from services.
	updatesCount := resp.UpdatesCurrentAvailable + resp.UpdatesCurrentSkipped
	if updatesCount > 0 {
		api.Config.OrderMutex.RLock()
		defer api.Config.OrderMutex.RUnlock()
		resp.UpdateDetails = getUpdateDetails(api.Config, updatesCount)
	}

	api.writeJSON(w, resp, logFrom)
}
