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
	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvmanual "github.com/release-argus/Argus/service/deployed_version/types/manual"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

//
// Defaults.
//

// convertAndCensorDefaults converts Defaults to API Type and censor secrets.
func convertAndCensorDefaults(input *config.Defaults) *apitype.Defaults {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets.
	apiRequire := apitype.Defaults{
		Service: apitype.ServiceDefaults{
			Options: &apitype.ServiceOptions{
				Interval:           input.Service.Options.Interval,
				SemanticVersioning: input.Service.Options.SemanticVersioning},
			LatestVersion: &apitype.LatestVersionDefaults{
				AccessToken:       util.ValueUnlessDefault(input.Service.LatestVersion.AccessToken, util.SecretValue),
				AllowInvalidCerts: input.Service.LatestVersion.AllowInvalidCerts,
				UsePreRelease:     input.Service.LatestVersion.UsePreRelease,
				Require:           convertAndCensorLatestVersionRequireDefaults(&input.Service.LatestVersion.Require)},
			DeployedVersionLookup: &apitype.DeployedVersionLookup{
				AllowInvalidCerts: input.Service.DeployedVersionLookup.AllowInvalidCerts},
			Dashboard: &apitype.DashboardOptions{
				AutoApprove: input.Service.Dashboard.AutoApprove}},
		Notify:  *convertAndCensorNotifySliceDefaults(&input.Notify),
		WebHook: *convertAndCensorWebHookDefaults(&input.WebHook)}

	return &apiRequire
}

//
// Service.
//

// convertAndCensorService converts Service to API type, censoring any secrets.
func convertAndCensorService(service *service.Service) *apitype.Service {
	if service == nil {
		return nil
	}

	apiService := apitype.Service{}

	// Name.
	if service.MarshalName() {
		apiService.Name = service.Name
	}
	// Comment.
	apiService.Comment = service.Comment

	// Options.
	apiService.Options = &apitype.ServiceOptions{
		Active:             service.Options.Active,
		Interval:           service.Options.Interval,
		SemanticVersioning: service.Options.SemanticVersioning}

	// LatestVersion.
	apiService.LatestVersion = convertAndCensorLatestVersion(service.LatestVersion)
	// DeployedVersionLookup.
	apiService.DeployedVersionLookup = convertAndCensorDeployedVersionLookup(service.DeployedVersionLookup)
	// Notify.
	apiService.Notify = convertAndCensorNotifySlice(&service.Notify)
	// Command.
	apiService.Command = convertCommandSlice(&service.Command)
	// WebHook.
	apiService.WebHook = convertAndCensorWebHookSlice(&service.WebHook)

	apiService.Dashboard = &apitype.DashboardOptions{
		AutoApprove: service.Dashboard.AutoApprove,
		Icon:        service.Dashboard.Icon,
		IconLinkTo:  service.Dashboard.IconLinkTo,
		WebURL:      service.Dashboard.WebURL,
		Tags:        service.Dashboard.Tags}

	return &apiService
}

//
// Latest Version.
//

// convertAndCensorLatestVersion converts Lookup to API Type, censoring any secrets.
func convertAndCensorLatestVersion(lv latestver.Lookup) *apitype.LatestVersion {
	if lv == nil {
		return nil
	}

	switch v := lv.(type) {
	case *github.Lookup:
		return &apitype.LatestVersion{
			Type:          v.Type,
			URL:           v.URL,
			AccessToken:   util.ValueUnlessDefault(v.AccessToken, util.SecretValue),
			UsePreRelease: v.UsePreRelease,
			URLCommands:   convertURLCommandSlice(&v.URLCommands),
			Require:       convertAndCensorLatestVersionRequire(v.Require),
		}
	case *lvweb.Lookup:
		return &apitype.LatestVersion{
			Type:              v.Type,
			URL:               v.URL,
			AllowInvalidCerts: v.AllowInvalidCerts,
			URLCommands:       convertURLCommandSlice(&v.URLCommands),
			Require:           convertAndCensorLatestVersionRequire(v.Require),
		}
	}

	return nil
}

// convertAndCensorLatestVersionRequireDefaults converts RequireDefaults to API Type, censoring any secrets.
func convertAndCensorLatestVersionRequireDefaults(require *filter.RequireDefaults) *apitype.LatestVersionRequireDefaults {
	if require == nil {
		return nil
	}

	apiRequire := &apitype.LatestVersionRequireDefaults{
		Docker: apitype.RequireDockerCheckDefaults{
			Type: require.Docker.Type}}

	// Docker:
	//   GHCR.
	if require.Docker.RegistryGHCR != nil {
		apiRequire.Docker.GHCR = &apitype.RequireDockerCheckRegistryDefaults{
			Token: util.ValueUnlessDefault(require.Docker.RegistryGHCR.Token, util.SecretValue)}
	}
	//   Docker Hub.
	if require.Docker.RegistryHub != nil {
		apiRequire.Docker.Hub = &apitype.RequireDockerCheckRegistryDefaultsWithUsername{
			Username: require.Docker.RegistryHub.Username,
			RequireDockerCheckRegistryDefaults: apitype.RequireDockerCheckRegistryDefaults{
				Token: util.ValueUnlessDefault(require.Docker.RegistryHub.Token, util.SecretValue)}}
	}
	//   Quay.
	if require.Docker.RegistryQuay != nil {
		apiRequire.Docker.Quay = &apitype.RequireDockerCheckRegistryDefaults{
			Token: util.ValueUnlessDefault(require.Docker.RegistryQuay.Token, util.SecretValue)}
	}

	return apiRequire
}

// convertAndCensorLatestVersionRequire converts Require to API Type, censoring any secrets.
func convertAndCensorLatestVersionRequire(require *filter.Require) *apitype.LatestVersionRequire {
	if require == nil {
		return nil
	}

	// Docker.
	var docker *apitype.RequireDockerCheck
	if require.Docker != nil {
		docker = &apitype.RequireDockerCheck{
			Type:     require.Docker.Type,
			Image:    require.Docker.Image,
			Tag:      require.Docker.Tag,
			Username: require.Docker.Username,
			Token:    util.ValueUnlessDefault(require.Docker.Token, util.SecretValue)}
	}

	// Require.
	apiRequire := apitype.LatestVersionRequire{
		Command:      require.Command,
		Docker:       docker,
		RegexContent: require.RegexContent,
		RegexVersion: require.RegexVersion}

	return &apiRequire
}

// convertURLCommandSlice converts URLCommandSlice to API Type.
func convertURLCommandSlice(commands *filter.URLCommandSlice) *apitype.URLCommandSlice {
	if commands == nil {
		return nil
	}

	slice := make(apitype.URLCommandSlice, len(*commands))
	for i, cmd := range *commands {
		slice[i] = apitype.URLCommand{
			Type:     cmd.Type,
			Regex:    cmd.Regex,
			Index:    cmd.Index,
			Template: cmd.Template,
			Text:     cmd.Text,
			Old:      cmd.Old,
			New:      cmd.New}
	}

	return &slice
}

//
// Deployed Version.
//

// convertAndCensorDeployedVersionLookup converts Lookup to API Type, censoring any secrets.
func convertAndCensorDeployedVersionLookup(dvl deployedver.Lookup) *apitype.DeployedVersionLookup {
	if dvl == nil {
		return nil
	}

	switch dvl := dvl.(type) {
	case *dvweb.Lookup:
		apiDVL := apitype.DeployedVersionLookup{
			Type:              dvl.Type,
			Method:            dvl.Method,
			URL:               dvl.URL,
			AllowInvalidCerts: dvl.AllowInvalidCerts,
			TargetHeader:      dvl.TargetHeader,
			Headers:           nil,
			Body:              dvl.Body,
			JSON:              dvl.JSON,
			Regex:             dvl.Regex,
			RegexTemplate:     dvl.RegexTemplate}

		// Basic auth.
		if dvl.BasicAuth != nil {
			apiDVL.BasicAuth = &apitype.BasicAuth{
				Username: dvl.BasicAuth.Username,
				Password: util.SecretValue}
		}

		// Headers.
		apiDVL.Headers = make([]apitype.Header, len(dvl.Headers))
		for i := range dvl.Headers {
			apiDVL.Headers[i] = apitype.Header{
				Key:   dvl.Headers[i].Key,
				Value: util.SecretValue}
		}

		return &apiDVL
	case *dvmanual.Lookup:
		return &apitype.DeployedVersionLookup{
			Type:    dvl.Type,
			Version: dvl.Status.DeployedVersion(),
		}
	}

	return nil
}

//
// Notify.
//

// convertAndCensorNotifySliceDefaults converts Slice to NotifySlice, censoring any secrets.
func convertAndCensorNotifySliceDefaults(input *shoutrrr.SliceDefaults) *apitype.NotifySlice {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets.
	slice := make(apitype.NotifySlice, len(*input))
	for name, notify := range *input {
		n := &apitype.Notify{
			Type:      notify.Type,
			Options:   notify.Options,
			URLFields: notify.URLFields,
			Params:    notify.Params}
		// Censor and add to slice.
		n.Censor()
		slice[name] = n
	}

	return &slice
}

// convertAndCensorNotifySlice converts Slice to API Type, censoring any secrets.
func convertAndCensorNotifySlice(input *shoutrrr.Slice) *apitype.NotifySlice {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets.
	slice := make(apitype.NotifySlice, len(*input))
	for name, notify := range *input {
		n := &apitype.Notify{
			Type:      notify.Type,
			Options:   notify.Options,
			URLFields: notify.URLFields,
			Params:    notify.Params}
		// Censor and add to slice.
		n.Censor()
		slice[name] = n
	}

	return &slice
}

//
// Command.
//

// convertCommandSlice converts Slice to API type.
func convertCommandSlice(commands *command.Slice) *apitype.CommandSlice {
	if commands == nil {
		return nil
	}

	slice := make(apitype.CommandSlice, len(*commands))
	for index, cmd := range *commands {
		slice[index] = apitype.Command(cmd)
	}

	return &slice
}

//
// WebHook.
//

// convertWebHookHeaders converts WebHook Headers to API type.
func convertWebHookHeaders(headers *webhook.Headers) *[]apitype.Header {
	if headers == nil {
		return nil
	}

	apiHeaders := make([]apitype.Header, len(*headers))
	for index, header := range *headers {
		apiHeaders[index] = apitype.Header(header)
	}

	return &apiHeaders
}

// convertAndCensorWebHookSliceDefaults converts SliceDefaults to API Type, censoring any secrets.
func convertAndCensorWebHookSliceDefaults(input *webhook.SliceDefaults) *apitype.WebHookSlice {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets.
	slice := make(apitype.WebHookSlice, len(*input))
	for name, wh := range *input {
		slice[name] = convertAndCensorWebHookDefaults(wh)
	}

	return &slice
}

// convertAndCensorWebHookDefaults converts Defaults to API type, censoring any secrets.
func convertAndCensorWebHookDefaults(webhook *webhook.Defaults) *apitype.WebHook {
	if webhook == nil {
		return nil
	}

	apiElement := &apitype.WebHook{
		Type:              webhook.Type,
		URL:               webhook.URL,
		AllowInvalidCerts: webhook.AllowInvalidCerts,
		Secret:            util.ValueUnlessDefault(webhook.Secret, util.SecretValue),
		CustomHeaders:     convertWebHookHeaders(webhook.CustomHeaders),
		DesiredStatusCode: webhook.DesiredStatusCode,
		Delay:             webhook.Delay,
		MaxTries:          webhook.MaxTries,
		SilentFails:       webhook.SilentFails}
	apiElement.Censor()

	return apiElement
}

// convertAndCensorWebHookSlice converts Slice to API Type, censoring any secrets.
func convertAndCensorWebHookSlice(webhooks *webhook.Slice) *apitype.WebHookSlice {
	if webhooks == nil {
		return nil
	}

	slice := make(apitype.WebHookSlice, len(*webhooks))
	for index, wh := range *webhooks {
		slice[index] = convertAndCensorWebHook(wh)
	}

	return &slice
}

// convertAndCensorWebHook converts WebHook to API type, censoring any secrets.
func convertAndCensorWebHook(webhook *webhook.WebHook) *apitype.WebHook {
	if webhook == nil {
		return nil
	}

	apiElement := &apitype.WebHook{
		Type:              webhook.Type,
		URL:               webhook.URL,
		AllowInvalidCerts: webhook.AllowInvalidCerts,
		Secret:            util.ValueUnlessDefault(webhook.Secret, util.SecretValue),
		CustomHeaders:     convertWebHookHeaders(webhook.CustomHeaders),
		DesiredStatusCode: webhook.DesiredStatusCode,
		Delay:             webhook.Delay,
		MaxTries:          webhook.MaxTries,
		SilentFails:       webhook.SilentFails}
	apiElement.Censor()

	return apiElement
}
