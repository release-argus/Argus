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

package v1

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/release-argus/Argus/utils"
	api_types "github.com/release-argus/Argus/web/api/types"
)

func (api *API) wsService(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsService", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	responsePage := "APPROVALS"
	responseType := "SERVICE"

	// Send the ordering
	responseSubType := "ORDERING"
	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		Order:   &api.Config.Order,
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}

	// Initialise the services
	responseSubType = "INIT"
	for _, key := range api.Config.Order {
		// Check Service still exists in this ordering
		if api.Config.Service[key] == nil {
			continue
		}

		service := api.Config.Service[key]
		url := service.GetServiceURL(false)
		webhookCount := 0
		if service.WebHook != nil {
			webhookCount = len(*service.WebHook)
		}
		commandCount := 0
		if service.Command != nil {
			commandCount = len(*service.Command)
		}
		hasDeployedVersionLookup := service.DeployedVersionLookup != nil

		serviceSummary := api_types.ServiceSummary{
			Active:                   service.Active,
			ID:                       service.ID,
			Type:                     service.Type,
			URL:                      &url,
			Icon:                     (*service).GetIconURL(),
			HasDeployedVersionLookup: &hasDeployedVersionLookup,
			Command:                  commandCount,
			WebHook:                  webhookCount,
			Status: &api_types.Status{
				ApprovedVersion:          service.Status.ApprovedVersion,
				DeployedVersion:          service.Status.DeployedVersion,
				DeployedVersionTimestamp: service.Status.DeployedVersionTimestamp,
				LatestVersion:            service.Status.LatestVersion,
				LatestVersionTimestamp:   service.Status.LatestVersionTimestamp,
				LastQueried:              service.Status.LastQueried,
			},
		}

		// Create and send ServiceSummary
		msg := api_types.WebSocketMessage{
			Page:        &responsePage,
			Type:        &responseType,
			SubType:     &responseSubType,
			ServiceData: &serviceSummary,
		}
		if err := client.conn.WriteJSON(msg); err != nil {
			api.Log.Error(err, logFrom, true)
			return
		}
	}
}

// wsServiceAction handles approvals/rejections of the latest version of a service.
//
// Required params:
//
// service_data.id - Service ID to approve/deny release.
func (api *API) wsServiceAction(client *Client, payload api_types.WebSocketMessage) {
	logFrom := utils.LogFrom{Primary: "wsServiceAction", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	if payload.ServiceData.ID == nil {
		api.Log.Error("service_data.id not provided", logFrom, true)
		return
	}
	if payload.Target == nil {
		api.Log.Error("target for webhooks not provided", logFrom, true)
		return
	}
	id := payload.ServiceData.ID
	if api.Config.Service[*id] == nil {
		api.Log.Error(fmt.Sprintf("%q is not a valid service_id", *id), logFrom, true)
		return
	}

	// SKIP this release
	if *payload.Target == "ARGUS_SKIP" {
		msg := fmt.Sprintf("%s release skip - %q", *id, payload.ServiceData.Status.LatestVersion)
		api.Log.Info(msg, logFrom, true)
		api.Config.Service[*id].HandleSkip(payload.ServiceData.Status.LatestVersion)
		return
	}

	if api.Config.Service[*id].WebHook == nil && api.Config.Service[*id].Command == nil {
		api.Log.Error(fmt.Sprintf("%q does not have any webhooks/commands to approve", *id), logFrom, true)
		return
	}

	// Send the WebHook(s).
	msg := fmt.Sprintf("%s %q Release actioned - %q",
		*id,
		api.Config.Service[*id].Status.LatestVersion,
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
		go api.Config.Service[*id].HandleFailedActions()
	default:
		if strings.HasPrefix(*payload.Target, "webhook_") {
			go api.Config.Service[*id].HandleWebHook(strings.TrimPrefix(*payload.Target, "webhook_"))
		} else {
			go api.Config.Service[*id].HandleCommand(strings.TrimPrefix(*payload.Target, "command_"))
		}
	}
}

// wsCommand handles getting the Command(s) of a service.
//
// Required params:
//
// service_data.id - Service ID to get the Command(s) of.
func (api *API) wsCommand(client *Client, payload api_types.WebSocketMessage) {
	logFrom := utils.LogFrom{Primary: "wsCommand", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	if payload.ServiceData.ID == nil {
		api.Log.Error("service_data.id not provided", logFrom, true)
		return
	}
	id := payload.ServiceData.ID
	if api.Config.Service[*id] == nil {
		api.Log.Error(fmt.Sprintf("%q, service not found", *id), logFrom, true)
		return
	}
	if api.Config.Service[*id].CommandController == nil {
		return
	}

	// Create and send commandSummary
	responsePage := "APPROVALS"
	responseType := "COMMAND"
	responseSubType := "SUMMARY"

	commandSummary := make(map[string]*api_types.CommandSummary, len(*api.Config.Service[*id].Command))
	for key := range *api.Config.Service[*id].CommandController.Command {
		command := strings.Join((*api.Config.Service[*id].CommandController.Command)[key], " ")
		commandSummary[command] = &api_types.CommandSummary{
			Failed: api.Config.Service[*id].CommandController.Failed[key],
		}
	}

	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ServiceData: &api_types.ServiceSummary{
			ID: id,
		},
		CommandData: commandSummary,
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsWebHook handles getting the WebHook(s) of a service.
//
// Required params:
//
// service_data.id - Service ID to get the WebHook(s) of.
func (api *API) wsWebHook(client *Client, payload api_types.WebSocketMessage) {
	logFrom := utils.LogFrom{Primary: "wsWebHook", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	if payload.ServiceData.ID == nil {
		api.Log.Error("service_data.id not provided", logFrom, true)
		return
	}
	id := payload.ServiceData.ID
	if api.Config.Service[*id] == nil {
		api.Log.Error(fmt.Sprintf("%q, service not found", *id), logFrom, true)
		return
	}
	if api.Config.Service[*id].WebHook == nil {
		return
	}

	// Create and send webhookSummary
	responsePage := "APPROVALS"
	responseType := "WEBHOOK"
	responseSubType := "SUMMARY"
	webhookSummary := make(map[string]*api_types.WebHookSummary, len(*api.Config.Service[*id].WebHook))

	for key := range *api.Config.Service[*id].WebHook {
		webhookSummary[key] = &api_types.WebHookSummary{
			Failed: (*api.Config.Service[*id].WebHook)[key].Failed,
		}
	}

	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ServiceData: &api_types.ServiceSummary{
			ID: id,
		},
		WebHookData: webhookSummary,
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsStatus handles getting the info of the Argus binary status.
func (api *API) wsStatus(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsStatus", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "RUNTIME_BUILD"
	responseType := "INIT"
	info := api_types.Info{
		Build: api_types.BuildInfo{
			Version:   utils.Version,
			BuildDate: utils.BuildDate,
			GoVersion: utils.GoVersion,
		},
		Runtime: api_types.RuntimeInfo{
			StartTime:      utils.StartTime,
			CWD:            utils.CWD,
			GoRoutineCount: runtime.NumGoroutine(),
			GOMAXPROCS:     runtime.GOMAXPROCS(0),
			GoGC:           os.Getenv("GOGC"),
			GoDebug:        os.Getenv("GODEBUG"),
		},
	}

	msg := api_types.WebSocketMessage{
		Page:     &responsePage,
		Type:     &responseType,
		InfoData: &info,
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsFlags handles getting the values of the flags that can be used with the binary.
func (api *API) wsFlags(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsFlags", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "FLAGS"
	responseType := "INIT"
	msg := api_types.WebSocketMessage{
		Page: &responsePage,
		Type: &responseType,
		FlagsData: &api_types.Flags{
			ConfigFile:     &api.Config.File,
			LogLevel:       api.Config.Settings.GetLogLevel(),
			LogTimestamps:  api.Config.Settings.GetLogTimestamps(),
			WebListenHost:  api.Config.Settings.GetWebListenHost(),
			WebListenPort:  api.Config.Settings.GetWebListenPort(),
			WebCertFile:    api.Config.Settings.GetWebCertFile(),
			WebPKeyFile:    api.Config.Settings.GetWebKeyFile(),
			WebRoutePrefix: api.Config.Settings.GetWebRoutePrefix(),
		},
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsConfigSettings handles getting the `settings` config that's in use.
func (api *API) wsConfigSettings(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsConfigSettings", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "CONFIG"
	responseType := "SETTINGS"
	responseSubType := "INIT"

	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ConfigData: &api_types.Config{
			Settings: &api_types.Settings{
				Log: api_types.LogSettings{
					Timestamps: api.Config.Settings.Log.Timestamps,
					Level:      api.Config.Settings.Log.Level,
				},
				Web: api_types.WebSettings{
					ListenHost:  api.Config.Settings.Web.ListenHost,
					ListenPort:  api.Config.Settings.Web.ListenPort,
					CertFile:    api.Config.Settings.Web.CertFile,
					KeyFile:     api.Config.Settings.Web.KeyFile,
					RoutePrefix: api.Config.Settings.Web.RoutePrefix,
				},
			},
		},
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsConfigDefaults handles getting the `defaults` config that's in use.
func (api *API) wsConfigDefaults(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsConfigDefaults", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "CONFIG"
	responseType := "DEFAULTS"
	responseSubType := "INIT"

	notifyDefaults := make(api_types.NotifySlice)
	// key == type for defaults
	for key := range api.Config.Defaults.Notify {
		dflt := api_types.Notify{
			Type:      api.Config.Defaults.Notify[key].Type,
			Options:   api.Config.Defaults.Notify[key].Options,
			URLFields: api.Config.Defaults.Notify[key].URLFields,
			Params:    api.Config.Defaults.Notify[key].Params,
		}
		dflt = *dflt.Censor()
		notifyDefaults[key] = &dflt
	}

	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ConfigData: &api_types.Config{
			Defaults: &api_types.Defaults{
				Service: api_types.Service{
					Interval:           api.Config.Defaults.Service.Interval,
					SemanticVersioning: api.Config.Defaults.Service.SemanticVersioning,
					RegexContent:       api.Config.Defaults.Service.RegexContent,
					RegexVersion:       api.Config.Defaults.Service.RegexVersion,
					UsePreRelease:      api.Config.Defaults.Service.UsePreRelease,
					AutoApprove:        api.Config.Defaults.Service.AutoApprove,
					IgnoreMisses:       api.Config.Defaults.Service.IgnoreMisses,
					AccessToken:        utils.ValueIfNotNil(api.Config.Defaults.Service.AccessToken, "<secret>"),
					AllowInvalidCerts:  api.Config.Defaults.Service.AllowInvalidCerts,
					DeployedVersionLookup: &api_types.DeployedVersionLookup{
						AllowInvalidCerts: api.Config.Defaults.Service.DeployedVersionLookup.AllowInvalidCerts,
					},
				},
				Notify: notifyDefaults,
				WebHook: api_types.WebHook{
					Type:              api.Config.Defaults.WebHook.Type,
					Delay:             api.Config.Defaults.WebHook.Delay,
					DesiredStatusCode: api.Config.Defaults.WebHook.DesiredStatusCode,
					MaxTries:          api.Config.Defaults.WebHook.MaxTries,
					Secret:            api.Config.Defaults.WebHook.Secret,
					SilentFails:       api.Config.Defaults.WebHook.SilentFails,
					URL:               api.Config.Defaults.WebHook.URL,
				},
			},
		},
	}

	msg.ConfigData.Defaults.Notify = *msg.ConfigData.Defaults.Notify.Censor()
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsConfigNotify handles getting the `notify` config that's in use.
func (api *API) wsConfigNotify(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsConfigNotify", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "CONFIG"
	responseType := "NOTIFY"
	responseSubType := "INIT"

	notifyConfig := make(api_types.NotifySlice)
	if api.Config.Notify != nil {
		for key := range api.Config.Notify {
			notifyConfig[key] = &api_types.Notify{
				Type:      api.Config.Notify[key].Type,
				Options:   api.Config.Notify[key].Options,
				URLFields: api.Config.Notify[key].URLFields,
				Params:    api.Config.Notify[key].Params,
			}
			notifyConfig[key] = notifyConfig[key].Censor()
		}
	}
	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ConfigData: &api_types.Config{
			Notify: &notifyConfig,
		},
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsConfigWebHook handles getting the `webhook` config that's in use.
func (api *API) wsConfigWebHook(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsConfigWebHook", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "CONFIG"
	responseType := "WEBHOOK"
	responseSubType := "INIT"

	webhookConfig := make(api_types.WebHookSlice)
	if api.Config.WebHook != nil {
		secretVar := "<secret>"
		for key := range api.Config.WebHook {
			webhookConfig[key] = &api_types.WebHook{
				Type:              api.Config.WebHook[key].Type,
				URL:               api.Config.WebHook[key].URL,
				Secret:            utils.ValueIfNotDefault(api.Config.WebHook[key].URL, &secretVar),
				DesiredStatusCode: api.Config.WebHook[key].DesiredStatusCode,
				Delay:             api.Config.WebHook[key].Delay,
				MaxTries:          api.Config.WebHook[key].MaxTries,
				SilentFails:       api.Config.WebHook[key].SilentFails,
			}
		}
	}

	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ConfigData: &api_types.Config{
			WebHook: &webhookConfig,
		},
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsConfigService handles getting the `service` config that's in use.
func (api *API) wsConfigService(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsConfigService", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "CONFIG"
	responseType := "SERVICE"
	responseSubType := "INIT"

	serviceConfig := make(api_types.ServiceSlice)
	if api.Config.Service != nil {
		for _, key := range api.Config.Order {
			service := api.Config.Service[key]

			serviceConfig[key] = &api_types.Service{
				Active:             service.Active,
				Type:               service.Type,
				URL:                service.URL,
				WebURL:             service.WebURL,
				AccessToken:        utils.ValueIfNotNil(service.AccessToken, "<secret>"),
				SemanticVersioning: service.SemanticVersioning,
				RegexContent:       service.RegexContent,
				RegexVersion:       service.RegexVersion,
				UsePreRelease:      service.UsePreRelease,
				AutoApprove:        service.AutoApprove,
				IgnoreMisses:       service.IgnoreMisses,
				AllowInvalidCerts:  service.AllowInvalidCerts,
				Icon:               service.Icon,
				Status: &api_types.Status{
					ApprovedVersion:          service.Status.ApprovedVersion,
					DeployedVersion:          service.Status.DeployedVersion,
					DeployedVersionTimestamp: service.Status.DeployedVersionTimestamp,
					LatestVersion:            service.Status.LatestVersion,
					LatestVersionTimestamp:   service.Status.LatestVersionTimestamp,
				},
			}

			// DeployedVersionLookup
			if service.DeployedVersionLookup != nil {
				deployedVersionLookup := api_types.DeployedVersionLookup{}
				// URL
				if service.DeployedVersionLookup.URL != "" {
					deployedVersionLookup.URL = service.DeployedVersionLookup.URL
				}
				deployedVersionLookup.AllowInvalidCerts = service.DeployedVersionLookup.AllowInvalidCerts
				if service.DeployedVersionLookup.BasicAuth != nil {
					deployedVersionLookup.BasicAuth = &api_types.BasicAuth{
						Username: service.DeployedVersionLookup.BasicAuth.Username,
						Password: "<secret>",
					}
				}
				var headers []api_types.Header
				for _, header := range service.DeployedVersionLookup.Headers {
					headers = append(
						headers,
						api_types.Header{
							Key:   header.Key,
							Value: "<secret>",
						},
					)
				}
				deployedVersionLookup.Headers = headers
				if service.DeployedVersionLookup.JSON != "" {
					deployedVersionLookup.JSON = service.DeployedVersionLookup.JSON
				}
				if service.DeployedVersionLookup.Regex != "" {
					deployedVersionLookup.Regex = service.DeployedVersionLookup.Regex
				}
				serviceConfig[key].DeployedVersionLookup = &deployedVersionLookup
			}

			// URL Commands
			if service.URLCommands != nil {
				urlCommands := make(api_types.URLCommandSlice, len(*service.URLCommands))
				serviceConfig[key].URLCommands = &urlCommands
				for index := range *service.URLCommands {
					(*serviceConfig[key].URLCommands)[index] = api_types.URLCommand{
						Type:         (*service.URLCommands)[index].Type,
						Regex:        (*service.URLCommands)[index].Regex,
						Index:        (*service.URLCommands)[index].Index,
						Text:         (*service.URLCommands)[index].Text,
						Old:          (*service.URLCommands)[index].Old,
						New:          (*service.URLCommands)[index].New,
						IgnoreMisses: (*service.URLCommands)[index].IgnoreMisses,
					}
				}
			}

			// Notify
			if service.Notify != nil {
				notify := make(api_types.NotifySlice, len(*service.Notify))
				serviceConfig[key].Notify = &notify
				for index := range *service.Notify {
					if (*service.Notify)[index] == nil {
						(*serviceConfig[key].Notify)[index] = &api_types.Notify{}
					} else {
						(*serviceConfig[key].Notify)[index] = &api_types.Notify{
							Type:      (*service.Notify)[index].Type,
							Options:   (*service.Notify)[index].Options,
							URLFields: (*service.Notify)[index].URLFields,
							Params:    (*service.Notify)[index].Params,
						}
						// May be a new pointer as the fields are a map rather than individual pointers/vars
						(*serviceConfig[key].Notify)[index] = (*serviceConfig[key].Notify)[index].Censor()
					}
				}
			}

			// WebHook
			if service.WebHook != nil {
				webhook := make(api_types.WebHookSlice, len(*service.WebHook))
				serviceConfig[key].WebHook = &webhook
				for index := range *service.WebHook {
					(*serviceConfig[key].WebHook)[index] = &api_types.WebHook{
						Type:              (*service.WebHook)[index].Type,
						URL:               (*service.WebHook)[index].URL,
						Secret:            utils.ValueIfNotNil((*service.WebHook)[index].Secret, "<secret>"),
						CustomHeaders:     (*service.WebHook)[index].CustomHeaders,
						DesiredStatusCode: (*service.WebHook)[index].DesiredStatusCode,
						Delay:             (*service.WebHook)[index].Delay,
						MaxTries:          (*service.WebHook)[index].MaxTries,
						SilentFails:       (*service.WebHook)[index].SilentFails,
					}
				}
			}
			// Command
			if service.Command != nil {
				command := make(api_types.CommandSlice, len(*service.Command))
				serviceConfig[key].Command = &command
				for index := range *service.Command {
					copy((*serviceConfig[key].Command)[index], (*service.Command)[index])
				}
			}
		}
	}

	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ConfigData: &api_types.Config{
			Service: &serviceConfig,
			Order:   api.Config.Order,
		},
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}
