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

package v1

import (
	"encoding/json"
	"net/http"

	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

// wsConfig handles getting the config that's in use and sending it as YAML.
func (api *API) httpConfig(w http.ResponseWriter, r *http.Request) {
	logFrom := util.LogFrom{Primary: "httpConfig", Secondary: getIP(r)}
	api.Log.Verbose("-", logFrom, true)

	cfg := &api_type.Config{}

	// Settings
	cfg.Settings = &api_type.Settings{
		Log: api_type.LogSettings{
			Timestamps: api.Config.Settings.Log.Timestamps,
			Level:      api.Config.Settings.Log.Level},
		Web: api_type.WebSettings{
			ListenHost:  api.Config.Settings.Web.ListenHost,
			ListenPort:  api.Config.Settings.Web.ListenPort,
			CertFile:    api.Config.Settings.Web.CertFile,
			KeyFile:     api.Config.Settings.Web.KeyFile,
			RoutePrefix: api.Config.Settings.Web.RoutePrefix}}

	// Defaults
	serviceLatestVersionRequireDefaults := convertAndCensorLatestVersionRequireDefaults(&api.Config.Defaults.Service.LatestVersion.Require)
	var serviceNotifyDefaults map[string]struct{}
	if api.Config.Defaults.Service.Notify != nil {
		serviceNotifyDefaults = make(map[string]struct{}, len(api.Config.Defaults.Service.Notify))
		for notify := range api.Config.Defaults.Service.Notify {
			serviceNotifyDefaults[notify] = struct{}{}
		}
	}
	serviceCommandDefaults := make(api_type.CommandSlice, len(api.Config.Defaults.Service.Command))
	for i, command := range api.Config.Defaults.Service.Command {
		serviceCommandDefaults[i] = make(api_type.Command, len(command))
		copy(serviceCommandDefaults[i], command)
	}

	var serviceWebHookDefaults map[string]struct{}
	if api.Config.Defaults.Service.WebHook != nil {
		serviceWebHookDefaults = make(map[string]struct{}, len(api.Config.Defaults.Service.WebHook))
		for webhook := range api.Config.Defaults.Service.WebHook {
			serviceWebHookDefaults[webhook] = struct{}{}
		}
	}

	notifyDefaults := convertAndCensorNotifySliceDefaults(&api.Config.Defaults.Notify)
	webhookDefaults := convertAndCensorWebHookDefaults(&api.Config.Defaults.WebHook)

	cfg.Defaults = &api_type.Defaults{
		Service: api_type.ServiceDefaults{
			Options: &api_type.ServiceOptions{
				Interval:           api.Config.Defaults.Service.Options.Interval,
				SemanticVersioning: api.Config.Defaults.Service.Options.SemanticVersioning},
			DeployedVersionLookup: &api_type.DeployedVersionLookup{
				AllowInvalidCerts: api.Config.Defaults.Service.DeployedVersionLookup.AllowInvalidCerts},
			Dashboard: &api_type.DashboardOptions{
				AutoApprove: api.Config.Defaults.Service.Dashboard.AutoApprove},
			LatestVersion: &api_type.LatestVersionDefaults{
				AccessToken:       util.DefaultOrValue(api.Config.Defaults.Service.LatestVersion.AccessToken, "<secret>"),
				AllowInvalidCerts: api.Config.Defaults.Service.LatestVersion.AllowInvalidCerts,
				UsePreRelease:     api.Config.Defaults.Service.LatestVersion.UsePreRelease,
				Require:           serviceLatestVersionRequireDefaults},
			Notify:  serviceNotifyDefaults,
			Command: serviceCommandDefaults,
			WebHook: serviceWebHookDefaults},
		Notify:  *notifyDefaults,
		WebHook: *webhookDefaults}

	// Notify
	cfg.Notify = convertAndCensorNotifySliceDefaults(&api.Config.Notify)

	// WebHook
	cfg.WebHook = convertAndCensorWebHookSliceDefaults(&api.Config.WebHook)

	// Service
	api.Config.OrderMutex.RLock()
	serviceConfig := make(api_type.ServiceSlice, len(api.Config.Order))
	if api.Config.Service != nil {
		for _, key := range api.Config.Order {
			service := api.Config.Service[key]
			serviceConfig[key] = convertAndCensorService(service)
		}
	}
	cfg.Service = &serviceConfig

	cfg.Order = api.Config.Order
	api.Config.OrderMutex.RUnlock()

	err := json.NewEncoder(w).Encode(cfg)
	api.Log.Error(err, logFrom, err != nil)
}
