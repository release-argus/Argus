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
				AllowInvalidCerts: input.Service.DeployedVersionLookup.AllowInvalidCerts,
				Method:            input.Service.DeployedVersionLookup.Method},
			Dashboard: &apitype.DashboardOptions{
				AutoApprove: input.Service.Dashboard.AutoApprove},
			Command: convertCommands(&input.Service.Command),
			Notify:  input.Service.Notify,
			WebHook: input.Service.WebHook},
		Notify:  *convertAndCensorNotifiersDefaults(&input.Notify),
		WebHook: *convertAndCensorWebHookDefaults(&input.WebHook)}

	return &apiRequire
}

//
// Service.
//

// convertAndCensorService converts Service to API type, censoring any secrets.
func convertAndCensorService(input *service.Service) *apitype.Service {
	if input == nil {
		return nil
	}

	apiService := apitype.Service{}

	// Name.
	if input.MarshalName() {
		apiService.Name = input.Name
	}
	// Comment.
	apiService.Comment = input.Comment

	// Options.
	apiService.Options = &apitype.ServiceOptions{
		Active:             input.Options.Active,
		Interval:           input.Options.Interval,
		SemanticVersioning: input.Options.SemanticVersioning}

	// LatestVersion.
	apiService.LatestVersion = convertAndCensorLatestVersion(input.LatestVersion)
	// DeployedVersionLookup.
	apiService.DeployedVersionLookup = convertAndCensorDeployedVersionLookup(input.DeployedVersionLookup)
	// Notify.
	if !input.NotifyFromDefaults {
		apiService.Notify = convertAndCensorNotifiers(&input.Notify)
	}
	// Command.
	if !input.CommandFromDefaults {
		apiService.Command = convertCommands(&input.Command)
	}
	// WebHook.
	if !input.WebHookFromDefaults {
		apiService.WebHook = convertAndCensorWebHooks(&input.WebHook)
	}

	apiService.Dashboard = &apitype.DashboardOptions{
		AutoApprove: input.Dashboard.AutoApprove,
		Icon:        input.Dashboard.Icon,
		IconLinkTo:  input.Dashboard.IconLinkTo,
		WebURL:      input.Dashboard.WebURL,
		Tags:        input.Dashboard.Tags}

	return &apiService
}

//
// Latest Version.
//

// convertAndCensorLatestVersion converts Lookup to API Type, censoring any secrets.
func convertAndCensorLatestVersion(input latestver.Lookup) *apitype.LatestVersion {
	if input == nil {
		return nil
	}

	switch lv := input.(type) {
	case *github.Lookup:
		return &apitype.LatestVersion{
			Type:          lv.Type,
			URL:           lv.URL,
			AccessToken:   util.ValueUnlessDefault(lv.AccessToken, util.SecretValue),
			UsePreRelease: lv.UsePreRelease,
			URLCommands:   convertURLCommands(&lv.URLCommands),
			Require:       convertAndCensorLatestVersionRequire(lv.Require),
		}
	case *lvweb.Lookup:
		return &apitype.LatestVersion{
			Type:              lv.Type,
			URL:               lv.URL,
			AllowInvalidCerts: lv.AllowInvalidCerts,
			URLCommands:       convertURLCommands(&lv.URLCommands),
			Require:           convertAndCensorLatestVersionRequire(lv.Require),
		}
	}

	return nil
}

// convertAndCensorLatestVersionRequireDefaults converts RequireDefaults to API Type, censoring any secrets.
func convertAndCensorLatestVersionRequireDefaults(input *filter.RequireDefaults) *apitype.LatestVersionRequireDefaults {
	if input == nil {
		return nil
	}

	apiRequire := &apitype.LatestVersionRequireDefaults{
		Docker: apitype.RequireDockerCheckDefaults{
			Type: input.Docker.Type}}

	// Docker:
	//   GHCR.
	if input.Docker.RegistryGHCR != nil {
		apiRequire.Docker.GHCR = &apitype.RequireDockerCheckRegistryDefaults{
			Token: util.ValueUnlessDefault(input.Docker.RegistryGHCR.Token, util.SecretValue)}
	}
	//   Docker Hub.
	if input.Docker.RegistryHub != nil {
		apiRequire.Docker.Hub = &apitype.RequireDockerCheckRegistryDefaultsWithUsername{
			Username: input.Docker.RegistryHub.Username,
			RequireDockerCheckRegistryDefaults: apitype.RequireDockerCheckRegistryDefaults{
				Token: util.ValueUnlessDefault(input.Docker.RegistryHub.Token, util.SecretValue)}}
	}
	//   Quay.
	if input.Docker.RegistryQuay != nil {
		apiRequire.Docker.Quay = &apitype.RequireDockerCheckRegistryDefaults{
			Token: util.ValueUnlessDefault(input.Docker.RegistryQuay.Token, util.SecretValue)}
	}

	return apiRequire
}

// convertAndCensorLatestVersionRequire converts Require to API Type, censoring any secrets.
func convertAndCensorLatestVersionRequire(input *filter.Require) *apitype.LatestVersionRequire {
	if input == nil {
		return nil
	}

	// Docker.
	var docker *apitype.RequireDockerCheck
	if input.Docker != nil {
		docker = &apitype.RequireDockerCheck{
			Type:     input.Docker.Type,
			Image:    input.Docker.Image,
			Tag:      input.Docker.Tag,
			Username: input.Docker.Username,
			Token:    util.ValueUnlessDefault(input.Docker.Token, util.SecretValue)}
	}

	// Require.
	apiRequire := apitype.LatestVersionRequire{
		Command:      input.Command,
		Docker:       docker,
		RegexContent: input.RegexContent,
		RegexVersion: input.RegexVersion}

	return &apiRequire
}

// convertURLCommands converts URLCommands to API Type.
func convertURLCommands(input *filter.URLCommands) *apitype.URLCommands {
	if input == nil {
		return nil
	}

	urlCommands := make(apitype.URLCommands, len(*input))
	for i, cmd := range *input {
		urlCommands[i] = apitype.URLCommand{
			Type:     cmd.Type,
			Regex:    cmd.Regex,
			Index:    cmd.Index,
			Template: cmd.Template,
			Text:     cmd.Text,
			Old:      cmd.Old,
			New:      cmd.New}
	}

	return &urlCommands
}

//
// Deployed Version.
//

// convertAndCensorDeployedVersionLookup converts Lookup to API Type, censoring any secrets.
func convertAndCensorDeployedVersionLookup(input deployedver.Lookup) *apitype.DeployedVersionLookup {
	if input == nil {
		return nil
	}

	switch dvl := input.(type) {
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

// convertAndCensorNotifiersDefaults converts Shoutrrrs to Notifiers, censoring any secrets.
func convertAndCensorNotifiersDefaults(input *shoutrrr.ShoutrrrsDefaults) *apitype.Notifiers {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets.
	notifiers := make(apitype.Notifiers, len(*input))
	for name, notify := range *input {
		n := &apitype.Notify{
			Type:      notify.Type,
			Options:   notify.Options,
			URLFields: notify.URLFields,
			Params:    notify.Params}
		// Censor and add to mappint.
		n.Censor()
		notifiers[name] = n
	}

	return &notifiers
}

// convertAndCensorNotifiers converts Shoutrrrs to API Type, censoring any secrets.
func convertAndCensorNotifiers(input *shoutrrr.Shoutrrrs) *apitype.Notifiers {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets.
	notifiers := make(apitype.Notifiers, len(*input))
	for name, notify := range *input {
		n := &apitype.Notify{
			Type:      notify.Type,
			Options:   notify.Options,
			URLFields: notify.URLFields,
			Params:    notify.Params}
		// Censor and add to mapping.
		n.Censor()
		notifiers[name] = n
	}

	return &notifiers
}

//
// Command.
//

// convertCommands converts Commands to API type.
func convertCommands(input *command.Commands) *apitype.Commands {
	if input == nil {
		return nil
	}

	commands := make(apitype.Commands, len(*input))
	for index, cmd := range *input {
		commands[index] = apitype.Command(cmd)
	}

	return &commands
}

//
// WebHook.
//

// convertWebHookHeaders converts WebHook Headers to API type.
func convertWebHookHeaders(input *webhook.Headers) *[]apitype.Header {
	if input == nil {
		return nil
	}

	apiHeaders := make([]apitype.Header, len(*input))
	for index, header := range *input {
		apiHeaders[index] = apitype.Header(header)
	}

	return &apiHeaders
}

// convertAndCensorWebHooksDefaults converts WebHooksDefaults to API Type, censoring any secrets.
func convertAndCensorWebHooksDefaults(input *webhook.WebHooksDefaults) *apitype.WebHooks {
	if input == nil {
		return nil
	}

	// Convert to API Type, censoring secrets.
	webhooks := make(apitype.WebHooks, len(*input))
	for name, wh := range *input {
		webhooks[name] = convertAndCensorWebHookDefaults(wh)
	}

	return &webhooks
}

// convertAndCensorWebHookDefaults converts Defaults to API type, censoring any secrets.
func convertAndCensorWebHookDefaults(input *webhook.Defaults) *apitype.WebHook {
	if input == nil {
		return nil
	}

	apiElement := &apitype.WebHook{
		Type:              input.Type,
		URL:               input.URL,
		AllowInvalidCerts: input.AllowInvalidCerts,
		Secret:            util.ValueUnlessDefault(input.Secret, util.SecretValue),
		CustomHeaders:     convertWebHookHeaders(input.CustomHeaders),
		DesiredStatusCode: input.DesiredStatusCode,
		Delay:             input.Delay,
		MaxTries:          input.MaxTries,
		SilentFails:       input.SilentFails}
	apiElement.Censor()

	return apiElement
}

// convertAndCensorWebHooks converts WebHooks to API Type, censoring any secrets.
func convertAndCensorWebHooks(input *webhook.WebHooks) *apitype.WebHooks {
	if input == nil {
		return nil
	}

	webhooks := make(apitype.WebHooks, len(*input))
	for index, wh := range *input {
		webhooks[index] = convertAndCensorWebHook(wh)
	}

	return &webhooks
}

// convertAndCensorWebHook converts WebHook to API type, censoring any secrets.
func convertAndCensorWebHook(input *webhook.WebHook) *apitype.WebHook {
	if input == nil {
		return nil
	}

	apiElement := &apitype.WebHook{
		Type:              input.Type,
		URL:               input.URL,
		AllowInvalidCerts: input.AllowInvalidCerts,
		Secret:            util.ValueUnlessDefault(input.Secret, util.SecretValue),
		CustomHeaders:     convertWebHookHeaders(input.CustomHeaders),
		DesiredStatusCode: input.DesiredStatusCode,
		Delay:             input.Delay,
		MaxTries:          input.MaxTries,
		SilentFails:       input.SilentFails}
	apiElement.Censor()

	return apiElement
}
