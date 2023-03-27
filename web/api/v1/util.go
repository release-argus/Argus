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
	"encoding/json"
	"net/url"
	"strings"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

// convertAndCensorNotifySlice fromm *shoutrrr.Slice to *api_type.NotifySlice
func convertAndCensorNotifySlice(input *shoutrrr.Slice) *api_type.NotifySlice {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets
	notifies := make(api_type.NotifySlice, len(*input))
	for name := range *input {
		notifies[name] = (&api_type.Notify{
			ID:        name,
			Type:      (*input)[name].Type,
			Options:   (*input)[name].Options,
			URLFields: (*input)[name].URLFields,
			Params:    (*input)[name].Params,
		}).Censor()
	}
	return &notifies
}

// convertAndCensorWebHookSlice from *webhook.Slice to *api_type.WebHookSlice
func convertAndCensorWebHookSlice(input *webhook.Slice) *api_type.WebHookSlice {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets
	webhooks := make(api_type.WebHookSlice, len(*input))
	for name := range *input {
		webhooks[name] = convertWebHookToAPITypeWebHook((*input)[name]).Censor()
	}
	return &webhooks
}

// convertAndCensorDefaults fromm *config.Defaults to *api_type.Defaults
func convertAndCensorDefaults(input *config.Defaults) *api_type.Defaults {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets
	defaults := api_type.Defaults{
		Service: api_type.Service{
			Options: &api_type.ServiceOptions{
				Interval:           input.Service.Options.Interval,
				SemanticVersioning: input.Service.Options.SemanticVersioning,
			},
			LatestVersion: &api_type.LatestVersion{
				AccessToken:       util.DefaultOrValue(input.Service.LatestVersion.AccessToken, "<secret>"),
				AllowInvalidCerts: input.Service.LatestVersion.AllowInvalidCerts,
				UsePreRelease:     input.Service.LatestVersion.UsePreRelease,
			},
			DeployedVersionLookup: &api_type.DeployedVersionLookup{
				AllowInvalidCerts: input.Service.DeployedVersionLookup.AllowInvalidCerts,
			},
			Dashboard: &api_type.DashboardOptions{
				AutoApprove: input.Service.Dashboard.AutoApprove,
			},
		},
		Notify: *convertAndCensorNotifySlice(&input.Service.Notify),
		WebHook: api_type.WebHook{
			Delay:             input.WebHook.Delay,
			MaxTries:          input.WebHook.MaxTries,
			AllowInvalidCerts: input.WebHook.AllowInvalidCerts,
			DesiredStatusCode: input.WebHook.DesiredStatusCode,
			SilentFails:       input.WebHook.SilentFails,
		},
	}

	return &defaults
}

// getParam from a URL query string
func getParam(queryParams *url.Values, param string) *string {
	if queryParams.Has(param) {
		val := queryParams.Get(param)
		return &val
	}
	return nil
}

// announceDelete of a Service to the `s.AnnounceChannel`
// (Broadcast to all WebSocket clients).
func (api *API) announceDelete(serviceID string) {
	payloadData, _ := json.Marshal(api_type.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "DELETE",
		SubType: serviceID,
		Order:   &api.Config.Order,
	})
	*api.Config.HardDefaults.Service.Status.AnnounceChannel <- payloadData
}

// announceEdit of a Service to the `s.AnnounceChannel` if data shown to the user has changed.
// (Broadcast to all WebSocket clients).
func (api *API) announceEdit(old *api_type.ServiceSummary, new *api_type.ServiceSummary) {
	serviceChanged := ""
	if old != nil {
		serviceChanged = old.ID
		new.RemoveUnchanged(old)
	}

	payloadData, _ := json.Marshal(api_type.WebSocketMessage{
		Page:        "APPROVALS",
		Type:        "EDIT",
		SubType:     serviceChanged,
		ServiceData: new,
	})

	// If the service has been changed, the payload will have more than 16 double quotes.
	//       2         4      6      8         10               12             14        16
	// {"page":"SERVICE","type":"EDIT","sub_type":"serviceChanged","service_data":{"status":{}}}
	if strings.Count(string(payloadData), `"`) >= 16 {
		*api.Config.HardDefaults.Service.Status.AnnounceChannel <- payloadData
	}
}

// announceService to the `s.AnnounceChannel` (initial load)
// (Broadcast to all WebSocket clients).
func (api *API) announceService(name string, client *Client, logFrom *util.LogFrom) {
	// Check Service still exists in this ordering
	service := api.Config.Service[name]
	if service == nil || client == nil {
		return
	}

	// Create and send ServiceSummary
	msg := api_type.WebSocketMessage{
		Page:        "APPROVALS",
		Type:        "SERVICE",
		SubType:     "INIT",
		ServiceData: service.Summary(),
	}

	api.wsSendJSON(client, msg, logFrom)
}

func convertDeployedVersionLookupToAPITypeDeployedVersionLookup(dvl *deployedver.Lookup) *api_type.DeployedVersionLookup {
	if dvl == nil {
		return nil
	}
	var headers []api_type.Header
	apiDVL := api_type.DeployedVersionLookup{
		URL:               dvl.URL,
		AllowInvalidCerts: dvl.AllowInvalidCerts,
		Headers:           headers,
		JSON:              dvl.JSON,
		Regex:             dvl.Regex,
	}
	// Basic auth
	if dvl.BasicAuth != nil {
		apiDVL.BasicAuth = &api_type.BasicAuth{
			Username: dvl.BasicAuth.Username,
			Password: "<secret>",
		}
	}
	// Headers
	apiDVL.Headers = make([]api_type.Header, len(dvl.Headers))
	for i := range dvl.Headers {
		apiDVL.Headers[i] = api_type.Header{
			Key:   dvl.Headers[i].Key,
			Value: "<secret>",
		}
	}
	return &apiDVL
}

func convertURLCommandSliceToAPITypeURLCommandSlice(commands *filter.URLCommandSlice) api_type.URLCommandSlice {
	if commands == nil {
		return api_type.URLCommandSlice{}
	}
	apiSlice := make(api_type.URLCommandSlice, len(*commands))
	for index := range *commands {
		apiSlice[index] = api_type.URLCommand{
			Type:  (*commands)[index].Type,
			Regex: (*commands)[index].Regex,
			Index: (*commands)[index].Index,
			Text:  (*commands)[index].Text,
			Old:   (*commands)[index].Old,
			New:   (*commands)[index].New,
		}
	}
	return apiSlice
}

func convertNotifySliceToAPITypeNotifySlice(notifiers *shoutrrr.Slice) *api_type.NotifySlice {
	if notifiers == nil {
		return nil
	}
	apiSlice := make(api_type.NotifySlice, len(*notifiers))
	for index := range *notifiers {
		apiSlice[index] = &api_type.Notify{
			Type:      (*notifiers)[index].Type,
			Options:   (*notifiers)[index].Options,
			URLFields: (*notifiers)[index].URLFields,
			Params:    (*notifiers)[index].Params,
		}
		// Assign as it may be a new pointer as the fields are a map rather than individual pointers/vars
		apiSlice[index] = apiSlice[index].Censor()
	}
	return &apiSlice
}

func convertCommandSliceToAPITypeCommandSlice(commands *command.Slice) *api_type.CommandSlice {
	if commands == nil {
		return nil
	}
	apiSlice := make(api_type.CommandSlice, len(*commands))
	for index := range *commands {
		apiSlice[index] = api_type.Command((*commands)[index])
	}
	return &apiSlice
}

func convertWebHookSliceToAPITypeWebHookSlice(webhooks *webhook.Slice) *api_type.WebHookSlice {
	if webhooks == nil {
		return nil
	}
	apiSlice := make(api_type.WebHookSlice, len(*webhooks))
	for index := range *webhooks {
		apiSlice[index] = convertWebHookToAPITypeWebHook((*webhooks)[index])
	}
	return &apiSlice
}

func convertWebHookToAPITypeWebHook(webhook *webhook.WebHook) (apiElement *api_type.WebHook) {
	if webhook == nil {
		return
	}

	var customHeaders []api_type.Header
	// Convert CustomHeaders
	if webhook.CustomHeaders != nil {
		customHeaders = make([]api_type.Header, len(*webhook.CustomHeaders))
		for index, header := range *webhook.CustomHeaders {
			customHeaders[index] = api_type.Header{
				Key:   header.Key,
				Value: header.Value,
			}
		}
	}

	apiElement = (&api_type.WebHook{
		Type:              util.StringToPointer(webhook.Type),
		URL:               util.StringToPointer(webhook.URL),
		AllowInvalidCerts: webhook.AllowInvalidCerts,
		Secret: util.ValueIfNotNil(
			util.StringToPointer(webhook.Secret), "<secret>"),
		CustomHeaders:     customHeaders,
		DesiredStatusCode: webhook.DesiredStatusCode,
		Delay:             webhook.Delay,
		MaxTries:          webhook.MaxTries,
		SilentFails:       webhook.SilentFails,
	}).Censor()
	return
}

func convertServiceToAPITypeService(service *service.Service) (apiService *api_type.Service) {
	apiService = &api_type.Service{}

	apiService.Comment = service.Comment

	apiService.Options = &api_type.ServiceOptions{
		Active:             service.Options.Active,
		Interval:           service.Options.Interval,
		SemanticVersioning: service.Options.SemanticVersioning,
	}

	urlCommands := convertURLCommandSliceToAPITypeURLCommandSlice(&service.LatestVersion.URLCommands)
	apiService.LatestVersion = &api_type.LatestVersion{
		Type:              service.LatestVersion.Type,
		URL:               service.LatestVersion.URL,
		AccessToken:       util.DefaultOrValue(service.LatestVersion.AccessToken, "<secret>"),
		AllowInvalidCerts: service.LatestVersion.AllowInvalidCerts,
		UsePreRelease:     service.LatestVersion.UsePreRelease,
		URLCommands:       &urlCommands,
	}
	if service.LatestVersion.Require != nil {
		var docker *api_type.RequireDockerCheck
		if service.LatestVersion.Require.Docker != nil {
			docker = &api_type.RequireDockerCheck{
				Type:     service.LatestVersion.Require.Docker.Type,
				Image:    service.LatestVersion.Require.Docker.Image,
				Tag:      service.LatestVersion.Require.Docker.Tag,
				Username: service.LatestVersion.Require.Docker.Username,
				Token:    util.ValueIfNotDefault(service.LatestVersion.Require.Docker.Token, "<secret>"),
			}
		}
		apiService.LatestVersion.Require = &api_type.LatestVersionRequire{
			Command:      service.LatestVersion.Require.Command,
			Docker:       docker,
			RegexContent: service.LatestVersion.Require.RegexContent,
			RegexVersion: service.LatestVersion.Require.RegexVersion,
		}
	}

	// DeployedVersionLookup
	apiService.DeployedVersionLookup = convertDeployedVersionLookupToAPITypeDeployedVersionLookup(service.DeployedVersionLookup)
	// Notify
	apiService.Notify = convertNotifySliceToAPITypeNotifySlice(&service.Notify)
	// Command
	apiService.Command = convertCommandSliceToAPITypeCommandSlice(&service.Command)
	// WebHook
	apiService.WebHook = convertWebHookSliceToAPITypeWebHookSlice(&service.WebHook)

	apiService.Dashboard = &api_type.DashboardOptions{
		AutoApprove: service.Dashboard.AutoApprove,
		Icon:        service.Dashboard.Icon,
		IconLinkTo:  service.Dashboard.IconLinkTo,
		WebURL:      service.Dashboard.WebURL,
	}
	return
}
