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
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

func (api *API) wsSendJSON(client *Client, msg api_type.WebSocketMessage, logFrom *util.LogFrom) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, *logFrom, true)
	}
}

func (api *API) wsServiceInit(client *Client) {
	logFrom := util.LogFrom{Primary: "wsServiceInit", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Send the ordering
	msg := api_type.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "SERVICE",
		SubType: "ORDERING",
		Order:   &api.Config.Order}
	api.wsSendJSON(client, msg, &logFrom)

	// Initialise the services
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	for _, key := range api.Config.Order {
		api.announceService(key, client, &logFrom)
	}
}

// wsServiceAction handles approvals/rejections of the latest version of a service.
//
// Required params:
//
// service_data.id - Service ID to approve/deny release.
func (api *API) wsServiceAction(client *Client, payload api_type.WebSocketMessage) {
	logFrom := util.LogFrom{Primary: "wsServiceAction", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	if payload.ServiceData.ID == "" {
		api.Log.Error("service_data.id not provided", logFrom, true)
		return
	}
	if payload.Target == nil {
		api.Log.Error("target for command/webhook not provided", logFrom, true)
		return
	}

	// Check the service exists.
	id := payload.ServiceData.ID
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[id]
	api.Config.OrderMutex.RUnlock()
	if svc == nil {
		api.Log.Error(fmt.Sprintf("%q, service not found", id), logFrom, true)
		return
	}
	// Don't do anything if the service is not active.
	if !svc.Options.GetActive() {
		return
	}

	// SKIP this release
	if *payload.Target == "ARGUS_SKIP" {
		msg := fmt.Sprintf("%q release skip - %q",
			id, payload.ServiceData.Status.LatestVersion)
		api.Log.Info(msg, logFrom, true)
		svc.HandleSkip(payload.ServiceData.Status.LatestVersion)
		return
	}

	if svc.WebHook == nil && svc.Command == nil {
		api.Log.Error(fmt.Sprintf("%q does not have any commands/webhooks to approve", id), logFrom, true)
		return
	}

	// Send the WebHook(s).
	msg := fmt.Sprintf("%s %q Release actioned - %q",
		id,
		svc.Status.LatestVersion(),
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(*payload.Target,
					"ARGUS_ALL", "ALL"),
				"ARGUS_FAILED", "ALL UNSENT/FAILED"),
			"ARGUS_SKIP", "SKIP"),
	)
	api.Log.Info(msg, logFrom, true)
	switch *payload.Target {
	case "ARGUS_ALL", "ARGUS_FAILED":
		go svc.HandleFailedActions()
	default:
		if strings.HasPrefix(*payload.Target, "webhook_") {
			go svc.HandleWebHook(strings.TrimPrefix(*payload.Target, "webhook_"))
		} else {
			go svc.HandleCommand(strings.TrimPrefix(*payload.Target, "command_"))
		}
	}
}

// wsCommand handles getting the Command(s) of a service.
//
// Required params:
//
// service_data.id - Service ID to get the Command(s) of.
func (api *API) wsCommand(client *Client, payload api_type.WebSocketMessage) {
	logFrom := util.LogFrom{Primary: "wsCommand", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	if payload.ServiceData.ID == "" {
		api.Log.Error("service_data.id not provided", logFrom, true)
		return
	}
	id := payload.ServiceData.ID
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[id]
	api.Config.OrderMutex.RUnlock()
	if svc == nil {
		api.Log.Error(fmt.Sprintf("%q, service not found", id), logFrom, true)
		return
	}
	if svc.CommandController == nil {
		return
	}

	// Create and send commandSummary
	commandSummary := make(map[string]*api_type.CommandSummary, len(svc.Command))
	for i := range *svc.CommandController.Command {
		command := (*svc.CommandController.Command)[i].ApplyTemplate(&svc.Status)
		commandSummary[command.String()] = &api_type.CommandSummary{
			Failed:       svc.Status.Fails.Command.Get(i),
			NextRunnable: svc.CommandController.NextRunnable(i)}
	}

	msg := api_type.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "COMMAND",
		SubType: "SUMMARY",
		ServiceData: &api_type.ServiceSummary{
			ID: id},
		CommandData: commandSummary}
	api.wsSendJSON(client, msg, &logFrom)
}

// wsWebHook handles getting the WebHook(s) of a service.
//
// Required params:
//
// service_data.id - Service ID to get the WebHook(s) of.
func (api *API) wsWebHook(client *Client, payload api_type.WebSocketMessage) {
	logFrom := util.LogFrom{Primary: "wsWebHook", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	if payload.ServiceData.ID == "" {
		api.Log.Error("service_data.id not provided", logFrom, true)
		return
	}
	id := payload.ServiceData.ID
	api.Config.OrderMutex.RLock()
	svc := api.Config.Service[id]
	api.Config.OrderMutex.RUnlock()
	if svc == nil {
		api.Log.Error(fmt.Sprintf("%q, service not found", id), logFrom, true)
		return
	}
	if svc.WebHook == nil {
		return
	}

	// Create and send webhookSummary
	webhookSummary := make(map[string]*api_type.WebHookSummary, len(svc.WebHook))

	for key := range svc.WebHook {
		webhookSummary[key] = &api_type.WebHookSummary{
			Failed:       svc.Status.Fails.WebHook.Get(key),
			NextRunnable: svc.WebHook[key].NextRunnable(),
		}
	}

	msg := api_type.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "WEBHOOK",
		SubType: "SUMMARY",
		ServiceData: &api_type.ServiceSummary{
			ID: id},
		WebHookData: webhookSummary}
	api.wsSendJSON(client, msg, &logFrom)
}

// wsStatus handles getting the info of the Argus binary status.
func (api *API) wsStatus(client *Client) {
	logFrom := util.LogFrom{Primary: "wsStatus", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	info := api_type.Info{
		Build: api_type.BuildInfo{
			Version:   util.Version,
			BuildDate: util.BuildDate,
			GoVersion: util.GoVersion},
		Runtime: api_type.RuntimeInfo{
			StartTime:      util.StartTime,
			CWD:            util.CWD,
			GoRoutineCount: runtime.NumGoroutine(),
			GOMAXPROCS:     runtime.GOMAXPROCS(0),
			GoGC:           os.Getenv("GOGC"),
			GoDebug:        os.Getenv("GODEBUG")}}

	msg := api_type.WebSocketMessage{
		Page:     "RUNTIME_BUILD",
		Type:     "INIT",
		InfoData: &info}
	api.wsSendJSON(client, msg, &logFrom)
}

// wsFlags handles getting the values of the flags that can be used with the binary.
func (api *API) wsFlags(client *Client) {
	logFrom := util.LogFrom{Primary: "wsFlags", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	msg := api_type.WebSocketMessage{
		Page: "FLAGS",
		Type: "INIT",
		FlagsData: &api_type.Flags{
			ConfigFile:       &api.Config.File,
			LogLevel:         api.Config.Settings.LogLevel(),
			LogTimestamps:    api.Config.Settings.LogTimestamps(),
			DataDatabaseFile: api.Config.Settings.DataDatabaseFile(),
			WebListenHost:    api.Config.Settings.WebListenHost(),
			WebListenPort:    api.Config.Settings.WebListenPort(),
			WebCertFile:      api.Config.Settings.WebCertFile(),
			WebPKeyFile:      api.Config.Settings.WebKeyFile(),
			WebRoutePrefix:   api.Config.Settings.WebRoutePrefix()}}
	api.wsSendJSON(client, msg, &logFrom)
}

// wsConfigSettings handles getting the `settings` config that's in use.
func (api *API) wsConfigSettings(client *Client) {
	logFrom := util.LogFrom{Primary: "wsConfigSettings", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	msg := api_type.WebSocketMessage{
		Page:    "CONFIG",
		Type:    "SETTINGS",
		SubType: "INIT",
		ConfigData: &api_type.Config{
			Settings: &api_type.Settings{
				Log: api_type.LogSettings{
					Timestamps: api.Config.Settings.Log.Timestamps,
					Level:      api.Config.Settings.Log.Level},
				Web: api_type.WebSettings{
					ListenHost:  api.Config.Settings.Web.ListenHost,
					ListenPort:  api.Config.Settings.Web.ListenPort,
					CertFile:    api.Config.Settings.Web.CertFile,
					KeyFile:     api.Config.Settings.Web.KeyFile,
					RoutePrefix: api.Config.Settings.Web.RoutePrefix}}}}
	api.wsSendJSON(client, msg, &logFrom)
}

// wsConfigDefaults handles getting the `defaults` config that's in use.
func (api *API) wsConfigDefaults(client *Client) {
	logFrom := util.LogFrom{Primary: "wsConfigDefaults", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	latestVersionRequireDefaults := convertAndCensorLatestVersionRequireDefaults(&api.Config.Defaults.Service.LatestVersion.Require)
	notifyDefaults := convertAndCensorNotifySliceDefaults(&api.Config.Defaults.Notify)
	webhookDefaults := convertAndCensorWebHookDefaults(&api.Config.Defaults.WebHook)

	msg := api_type.WebSocketMessage{
		Page:    "CONFIG",
		Type:    "DEFAULTS",
		SubType: "INIT",
		ConfigData: &api_type.Config{
			Defaults: &api_type.Defaults{
				Service: api_type.ServiceDefaults{
					Service: api_type.Service{
						Options: &api_type.ServiceOptions{
							Interval:           api.Config.Defaults.Service.Options.Interval,
							SemanticVersioning: api.Config.Defaults.Service.Options.SemanticVersioning},
						DeployedVersionLookup: &api_type.DeployedVersionLookup{
							AllowInvalidCerts: api.Config.Defaults.Service.DeployedVersionLookup.AllowInvalidCerts},
						Dashboard: &api_type.DashboardOptions{
							AutoApprove: api.Config.Defaults.Service.Dashboard.AutoApprove}},
					LatestVersion: &api_type.LatestVersionDefaults{
						LatestVersion: api_type.LatestVersion{
							AccessToken:       util.DefaultOrValue(api.Config.Defaults.Service.LatestVersion.AccessToken, "<secret>"),
							AllowInvalidCerts: api.Config.Defaults.Service.LatestVersion.AllowInvalidCerts,
							UsePreRelease:     api.Config.Defaults.Service.LatestVersion.UsePreRelease},
						Require: latestVersionRequireDefaults}},
				Notify:  *notifyDefaults,
				WebHook: *webhookDefaults}}}

	msg.ConfigData.Defaults.Notify = *msg.ConfigData.Defaults.Notify.Censor()
	api.wsSendJSON(client, msg, &logFrom)
}

// wsConfigNotify handles getting the `notify` config that's in use.
func (api *API) wsConfigNotify(client *Client) {
	logFrom := util.LogFrom{Primary: "wsConfigNotify", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	msg := api_type.WebSocketMessage{
		Page:    "CONFIG",
		Type:    "NOTIFY",
		SubType: "INIT",
		ConfigData: &api_type.Config{
			Notify: convertAndCensorNotifySliceDefaults(&api.Config.Notify)}}
	api.wsSendJSON(client, msg, &logFrom)
}

// wsConfigWebHook handles getting the `webhook` config that's in use.
func (api *API) wsConfigWebHook(client *Client) {
	logFrom := util.LogFrom{Primary: "wsConfigWebHook", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	msg := api_type.WebSocketMessage{
		Page:    "CONFIG",
		Type:    "WEBHOOK",
		SubType: "INIT",
		ConfigData: &api_type.Config{
			WebHook: convertAndCensorWebHookSliceDefaults(&api.Config.WebHook)}}
	api.wsSendJSON(client, msg, &logFrom)
}

// wsConfigService handles getting the `service` config that's in use.
func (api *API) wsConfigService(client *Client) {
	logFrom := util.LogFrom{Primary: "wsConfigService", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	api.Config.OrderMutex.RLock()
	defer api.Config.OrderMutex.RUnlock()
	serviceConfig := make(api_type.ServiceSlice, len(api.Config.Order))
	if api.Config.Service != nil {
		for _, key := range api.Config.Order {
			service := api.Config.Service[key]
			serviceConfig[key] = convertAndCensorService(service)
		}
	}

	msg := api_type.WebSocketMessage{
		Page:    "CONFIG",
		Type:    "SERVICE",
		SubType: "INIT",
		ConfigData: &api_type.Config{
			Service: &serviceConfig,
			Order:   api.Config.Order}}
	api.wsSendJSON(client, msg, &logFrom)
}
