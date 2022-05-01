// Copyright [2022] [Hymenaios]
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

	"github.com/hymenaios-io/Hymenaios/utils"
	api_types "github.com/hymenaios-io/Hymenaios/web/api/types"
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
		hasDeployedVersionLookup := service.DeployedVersionLookup != nil

		serviceSummary := api_types.ServiceSummary{
			ID:                       service.ID,
			Type:                     service.Type,
			URL:                      &url,
			Icon:                     (*service).GetIconURL(),
			HasDeployedVersionLookup: &hasDeployedVersionLookup,
			WebHook:                  webhookCount,
			Status: &api_types.Status{
				ApprovedVersion:         service.Status.ApprovedVersion,
				CurrentVersion:          service.Status.CurrentVersion,
				CurrentVersionTimestamp: service.Status.CurrentVersionTimestamp,
				LatestVersion:           service.Status.LatestVersion,
				LatestVersionTimestamp:  service.Status.LatestVersionTimestamp,
				LastQueried:             service.Status.LastQueried,
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

func (api *API) wsDefaults(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsDefaults", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)
	if err := client.conn.WriteJSON(api.Config.Defaults); err != nil {
		api.Log.Error(err, logFrom, true)
		return
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
	if *payload.Target == "HYMENAIOS_SKIP" {
		msg := fmt.Sprintf("%s release skip - %q", *id, *payload.ServiceData.Status.LatestVersion)
		api.Log.Info(msg, logFrom, true)
		api.Config.Service[*id].HandleSkip(payload.ServiceData.Status.LatestVersion)
		return
	}

	// Send the WebHook(s).
	msg := fmt.Sprintf("%s Release approved - %q WebHook", *id,
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(*payload.Target,
					"HYMENAIOS_ALL", "ALL"),
				"HYMENAIOS_FAILED", "ALL UNSENT/FAILED"),
			"HYMENAIOS_SKIP", "ALL"),
	)
	api.Log.Info(msg, logFrom, true)
	switch *payload.Target {
	case "HYMENAIOS_ALL":
		go api.Config.Service[*id].HandleWebHooks(true)
	case "HYMENAIOS_FAILED":
		go api.Config.Service[*id].HandleFailedWebHooks()
	default:
		go api.Config.Service[*id].HandleWebHook(*payload.Target)
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
		api.Log.Error(fmt.Sprintf("%q is not a valid id of a service", *id), logFrom, true)
		return
	}

	// Create and send webhookSummary
	responsePage := "APPROVALS"
	responseType := "WEBHOOK"
	responseSubType := "SUMMARY"
	webhookSummary := make(map[string]*api_types.WebHookSummary)

	if api.Config.Service[*id].WebHook != nil {
		for key := range *api.Config.Service[*id].WebHook {
			webhookSummary[key] = &api_types.WebHookSummary{
				Failed: (*api.Config.Service[*id].WebHook)[key].Failed,
			}
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

// wsStatus handles getting the info of the hymenaios binary status.
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

	gotifyExtras := api_types.Extras{}
	if api.Config.Defaults.Gotify.Extras != nil {
		gotifyExtras = api_types.Extras{
			AndroidAction:      api.Config.Defaults.Gotify.Extras.AndroidAction,
			ClientDisplay:      api.Config.Defaults.Gotify.Extras.ClientDisplay,
			ClientNotification: api.Config.Defaults.Gotify.Extras.ClientNotification,
		}
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
					AccessToken: utils.ValueIfNotNilPtr(
						api.Config.Defaults.Service.AccessToken,
						"<secret>"),
					AllowInvalidCerts: api.Config.Defaults.Service.AllowInvalidCerts,
					DeployedVersionLookup: &api_types.DeployedVersionLookup{
						AllowInvalidCerts: api.Config.Defaults.Service.DeployedVersionLookup.AllowInvalidCerts,
					},
				},
				Gotify: api_types.Gotify{
					Delay:    api.Config.Defaults.Gotify.Delay,
					Message:  api.Config.Defaults.Gotify.Message,
					MaxTries: api.Config.Defaults.Gotify.MaxTries,
					Priority: api.Config.Defaults.Gotify.Priority,
					Token:    api.Config.Defaults.Gotify.Token,
					Title:    api.Config.Defaults.Gotify.Title,
					URL:      api.Config.Defaults.Gotify.URL,
					Extras:   &gotifyExtras,
				},
				Slack: api_types.Slack{
					Delay:     api.Config.Defaults.Slack.Delay,
					IconURL:   api.Config.Defaults.Slack.IconURL,
					IconEmoji: api.Config.Defaults.Slack.IconEmoji,
					MaxTries:  api.Config.Defaults.Slack.MaxTries,
					Message:   api.Config.Defaults.Slack.Message,
					URL:       api.Config.Defaults.Slack.URL,
					Username:  api.Config.Defaults.Slack.Username,
				},
				WebHook: api_types.WebHook{
					Delay:             api.Config.Defaults.WebHook.Delay,
					DesiredStatusCode: api.Config.Defaults.WebHook.DesiredStatusCode,
					MaxTries:          api.Config.Defaults.WebHook.MaxTries,
					Type:              api.Config.Defaults.WebHook.Type,
					Secret:            api.Config.Defaults.WebHook.Secret,
					SilentFails:       api.Config.Defaults.WebHook.SilentFails,
					URL:               api.Config.Defaults.WebHook.URL,
				},
			},
		},
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsConfigGotify handles getting the `gotify` config that's in use.
func (api *API) wsConfigGotify(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsConfigGotify", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "CONFIG"
	responseType := "GOTIFY"
	responseSubType := "INIT"

	gotifyConfig := make(api_types.GotifySlice)
	if api.Config.Gotify != nil {
		for key := range *api.Config.Gotify {
			gotifyConfig[key] = &api_types.Gotify{
				URL:      (*api.Config.Gotify)[key].URL,
				Token:    utils.ValueIfNotNilPtr((*api.Config.Gotify)[key].Token, "<secret>"),
				Title:    (*api.Config.Gotify)[key].Title,
				Message:  (*api.Config.Gotify)[key].Message,
				Priority: (*api.Config.Gotify)[key].Priority,
				Delay:    (*api.Config.Gotify)[key].Delay,
				MaxTries: (*api.Config.Gotify)[key].MaxTries,
			}
			if (*api.Config.Gotify)[key].Extras != nil {
				gotifyConfig[key].Extras = &api_types.Extras{
					AndroidAction:      (*api.Config.Gotify)[key].Extras.AndroidAction,
					ClientDisplay:      (*api.Config.Gotify)[key].Extras.ClientDisplay,
					ClientNotification: (*api.Config.Gotify)[key].Extras.ClientNotification,
				}
			}
		}
	}

	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ConfigData: &api_types.Config{
			Gotify: &gotifyConfig,
		},
	}
	if err := client.conn.WriteJSON(msg); err != nil {
		api.Log.Error(err, logFrom, true)
		return
	}
}

// wsConfigSlack handles getting the `slack` config that's in use.
func (api *API) wsConfigSlack(client *Client) {
	logFrom := utils.LogFrom{Primary: "wsConfigSlack", Secondary: client.ip}
	api.Log.Verbose("-", logFrom, true)

	// Create and send status page data
	responsePage := "CONFIG"
	responseType := "SLACK"
	responseSubType := "INIT"

	slackConfig := make(api_types.SlackSlice)
	if api.Config.Slack != nil {
		for key := range *api.Config.Slack {
			slackConfig[key] = &api_types.Slack{
				URL:       utils.ValueIfNotNilPtr((*api.Config.Slack)[key].URL, "<secret>"),
				IconEmoji: (*api.Config.Slack)[key].IconEmoji,
				IconURL:   (*api.Config.Slack)[key].IconURL,
				Username:  (*api.Config.Slack)[key].Username,
				Message:   (*api.Config.Slack)[key].Message,
				Delay:     (*api.Config.Slack)[key].Delay,
				MaxTries:  (*api.Config.Slack)[key].MaxTries,
			}
		}
	}
	msg := api_types.WebSocketMessage{
		Page:    &responsePage,
		Type:    &responseType,
		SubType: &responseSubType,
		ConfigData: &api_types.Config{
			Slack: &slackConfig,
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
		for key := range *api.Config.WebHook {
			webhookConfig[key] = &api_types.WebHook{
				Type:              (*api.Config.WebHook)[key].Type,
				URL:               (*api.Config.WebHook)[key].URL,
				Secret:            utils.ValueIfNotNilPtr((*api.Config.WebHook)[key].URL, "<secret>"),
				DesiredStatusCode: (*api.Config.WebHook)[key].DesiredStatusCode,
				Delay:             (*api.Config.WebHook)[key].Delay,
				MaxTries:          (*api.Config.WebHook)[key].MaxTries,
				SilentFails:       (*api.Config.WebHook)[key].SilentFails,
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
				Type:               service.Type,
				URL:                service.URL,
				WebURL:             service.WebURL,
				AccessToken:        utils.ValueIfNotNilPtr(service.AccessToken, "<secret>"),
				SemanticVersioning: service.SemanticVersioning,
				RegexContent:       service.RegexContent,
				RegexVersion:       service.RegexVersion,
				UsePreRelease:      service.UsePreRelease,
				AutoApprove:        service.AutoApprove,
				IgnoreMisses:       service.IgnoreMisses,
				AllowInvalidCerts:  service.AllowInvalidCerts,
				Icon:               service.Icon,
				Status: &api_types.Status{
					ApprovedVersion:         service.Status.ApprovedVersion,
					CurrentVersion:          service.Status.CurrentVersion,
					CurrentVersionTimestamp: service.Status.CurrentVersionTimestamp,
					LatestVersion:           service.Status.LatestVersion,
					LatestVersionTimestamp:  service.Status.LatestVersionTimestamp,
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

			// Gotify
			if service.Gotify != nil {
				gotify := make(api_types.GotifySlice, len(*service.Gotify))
				serviceConfig[key].Gotify = &gotify
				for index := range *service.Gotify {
					(*serviceConfig[key].Gotify)[index] = &api_types.Gotify{
						URL:      (*service.Gotify)[index].URL,
						Token:    utils.ValueIfNotNilPtr((*service.Gotify)[index].Token, "<secret>"),
						Title:    (*service.Gotify)[index].Title,
						Message:  (*service.Gotify)[index].Message,
						Priority: (*service.Gotify)[index].Priority,
						Delay:    (*service.Gotify)[index].Delay,
						MaxTries: (*service.Gotify)[index].MaxTries,
					}
					if (*service.Gotify)[index].Delay != nil {
						(*serviceConfig[key].Gotify)[index].Extras = &api_types.Extras{
							AndroidAction:      (*service.Gotify)[index].Extras.AndroidAction,
							ClientDisplay:      (*service.Gotify)[index].Extras.ClientDisplay,
							ClientNotification: (*service.Gotify)[index].Extras.ClientNotification,
						}
					}
				}
			}

			// Slack
			if service.Slack != nil {
				slack := make(api_types.SlackSlice, len(*service.Slack))
				serviceConfig[key].Slack = &slack
				for index := range *service.Slack {
					(*serviceConfig[key].Slack)[index] = &api_types.Slack{
						URL:       utils.ValueIfNotNilPtr((*service.Slack)[index].URL, "<secret>"),
						IconEmoji: (*service.Slack)[index].IconEmoji,
						IconURL:   (*service.Slack)[index].IconURL,
						Username:  (*service.Slack)[index].Username,
						Message:   (*service.Slack)[index].Message,
						Delay:     (*service.Slack)[index].Delay,
						MaxTries:  (*service.Slack)[index].MaxTries,
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
						Secret:            utils.ValueIfNotNilPtr((*service.WebHook)[index].Secret, "<secret>"),
						DesiredStatusCode: (*service.WebHook)[index].DesiredStatusCode,
						Delay:             (*service.WebHook)[index].Delay,
						MaxTries:          (*service.WebHook)[index].MaxTries,
						SilentFails:       (*service.WebHook)[index].SilentFails,
					}
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
