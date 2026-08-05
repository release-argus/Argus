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

// Package v1 provides the API for the webserver.
package v1

import (
	"net/http"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// httpConfig handles GET /api/v1/config: returning the active configuration
// with secrets censored.
//
// Response:
//
//	200 OK: JSON object containing the censored configuration.
func (api *API) httpConfig(w http.ResponseWriter, r *http.Request) {
	logFrom := logx.LogFrom{Primary: "httpConfig", Secondary: getIP(r)}

	cfg := &apitype.Config{}

	// Settings.
	cfg.Settings = apitype.Settings{
		Data: apitype.DataSettings{
			DatabaseFile: api.Config.Settings.Data.DatabaseFile,
			Readonly:     api.Config.Settings.Data.Readonly,
		},
		Log: apitype.LogSettings{
			Timestamps: api.Config.Settings.Log.Timestamps,
			Level:      api.Config.Settings.Log.Level,
		},
		Web: apitype.WebSettings{
			ListenHost:     api.Config.Settings.Web.ListenHost,
			ListenPort:     api.Config.Settings.Web.ListenPort,
			CertFile:       api.Config.Settings.Web.CertFile,
			KeyFile:        api.Config.Settings.Web.KeyFile,
			RoutePrefix:    api.Config.Settings.Web.RoutePrefix,
			DisabledRoutes: api.Config.Settings.Web.DisabledRoutes,
			TrustedProxies: api.Config.Settings.Web.TrustedProxies,
		},
		Auth: apitype.AuthSettings{
			Enabled: api.Config.Settings.Auth.Enabled,
			Session: apitype.AuthSessionSettings{
				Lifetime:    api.Config.Settings.Auth.Session.Lifetime,
				IdleTimeout: api.Config.Settings.Auth.Session.IdleTimeout,
			},
			Local: apitype.AuthLocalSettings{
				Enabled: api.Config.Settings.Auth.Local.Enabled,
			},
		},
	}

	// Defaults.Service.LatestVersion.Common.Require.
	serviceLatestVersionRequireDefaults := convertAndCensorLatestVersionRequireDefaults(&api.Config.Defaults.Service.LatestVersion.Common.Require)
	// Defaults.Service.Notify.
	serviceNotifyDefaults := util.SortedKeys(api.Config.Defaults.Service.Notify)
	// Defaults.Service.Command.
	serviceCommandDefaults := convertCommands(api.Config.Defaults.Service.Command)
	// Defaults.Service.WebHook.
	serviceWebHookDefaults := util.SortedKeys(api.Config.Defaults.Service.WebHook)

	// Defaults.Notify.
	notifyDefaults := convertAndCensorNotifiersDefaults(api.Config.Defaults.Notify)
	// Defaults.WebHook.
	webhookDefaults := convertAndCensorWebHookDefaults(api.Config.Defaults.WebHook)

	cfg.Defaults = apitype.Defaults{
		Service: apitype.ServiceDefaults{
			Options: apitype.ServiceOptions{
				Interval:           api.Config.Defaults.Service.Options.Interval,
				SemanticVersioning: api.Config.Defaults.Service.Options.SemanticVersioning,
			},
			DeployedVersionLookup: apitype.DeployedVersionLookupDefaults{
				AllowInvalidCerts: api.Config.Defaults.Service.DeployedVersionLookup.AllowInvalidCerts,
			},
			Dashboard: apitype.DashboardOptions{
				AutoApprove: api.Config.Defaults.Service.Dashboard.AutoApprove,
			},
			LatestVersion: apitype.LatestVersionDefaults{
				Type: api.Config.Defaults.Service.LatestVersion.Type,
				Common: apitype.LatestVersionCommonDefaults{
					Require: serviceLatestVersionRequireDefaults,
				},
				GitHub: apitype.LatestVersionGitHubDefaults{
					AccessToken:   util.ValueUnlessZero(api.Config.Defaults.Service.LatestVersion.GitHub.AccessToken, util.SecretValue),
					UsePreRelease: api.Config.Defaults.Service.LatestVersion.GitHub.UsePreRelease,
				},
				URL: apitype.LatestVersionURLDefaults{
					AllowInvalidCerts: api.Config.Defaults.Service.LatestVersion.URL.AllowInvalidCerts,
				},
			},
			Notify:  serviceNotifyDefaults,
			Command: serviceCommandDefaults,
			WebHook: serviceWebHookDefaults,
		},
		Notify:  notifyDefaults,
		WebHook: webhookDefaults,
	}

	// Notify.
	cfg.Notify = convertAndCensorNotifiersDefaults(api.Config.Notify)

	// WebHook.
	cfg.WebHook = convertAndCensorWebHooksDefaults(api.Config.WebHook)

	// Service.
	api.Config.OrderMu.RLock()
	serviceConfig := make(apitype.Services, len(api.Config.Order))
	if api.Config.Service != nil {
		for _, key := range api.Config.Order {
			svc := api.Config.Service[key]
			serviceConfig[key] = convertAndCensorService(svc)
		}
	}
	cfg.Service = serviceConfig

	cfg.Order = api.Config.Order
	api.Config.OrderMu.RUnlock()

	api.writeJSON(w, cfg, logFrom)
}
