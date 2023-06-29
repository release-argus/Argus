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

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

// oldIntIndex to look at for any <secret>'s used
type oldIntIndex struct {
	OldIndex *int `json:"oldIndex,omitempty"`
}

// oldStringIndex to look at for any <secret>'s used
type oldStringIndex struct {
	OldIndex *string `json:"oldIndex,omitempty"`
}

// dvSecretRef contains the reference for the DeployedVersionLookup <secret>'s
type dvSecretRef struct {
	Headers []oldIntIndex `json:"headers,omitempty"`
}

// whSecretRef contains the reference for the WebHook <secret>'s
type whSecretRef struct {
	OldIndex      *string       `json:"oldIndex,omitempty"`
	CustomHeaders []oldIntIndex `json:"custom_headers,omitempty"`
}

// oldSecretRefs contains the indexes to use for <secret>'s
type oldSecretRefs struct {
	Name                  string                    `json:"name"`
	DeployedVersionLookup dvSecretRef               `json:"deployed_version,omitempty"`
	Notify                map[string]oldStringIndex `json:"notify,omitempty"`
	WebHook               map[string]whSecretRef    `json:"webhook,omitempty"`
}

// FromPayload will create a new/edited Service from a payload.
func FromPayload(
	oldService *Service,
	payload *io.ReadCloser,

	serviceDefaults *Defaults,
	serviceHardDefaults *Defaults,

	notifyGlobals *shoutrrr.SliceDefaults,
	notifyDefaults *shoutrrr.SliceDefaults,
	notifyHardDefaults *shoutrrr.SliceDefaults,

	webhookGlobals *webhook.SliceDefaults,
	webhookDefaults *webhook.WebHookDefaults,
	webhookHardDefaults *webhook.WebHookDefaults,

	logFrom *util.LogFrom,
) (newService *Service, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(*payload); err != nil {
		return
	}

	// Service
	newService = &Service{}
	dec1 := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	err = dec1.Decode(newService)
	if err != nil {
		jLog.Error(err, *logFrom, true)
		jLog.Verbose(fmt.Sprintf("Payload: %s", buf.String()), *logFrom, true)
		return
	}

	// SecretRefs
	dec2 := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	var secretRefs oldSecretRefs
	err = dec2.Decode(&secretRefs)
	if err != nil {
		jLog.Error(err, *logFrom, true)
		return
	}

	// Name + Channels
	newService.ID = secretRefs.Name
	newService.Status.AnnounceChannel = serviceHardDefaults.Status.AnnounceChannel
	newService.Status.DatabaseChannel = serviceHardDefaults.Status.DatabaseChannel
	newService.Status.SaveChannel = serviceHardDefaults.Status.SaveChannel

	removeDefaults(oldService, newService, serviceDefaults)
	newService.Init(
		serviceDefaults, serviceHardDefaults,
		notifyGlobals, notifyDefaults, notifyHardDefaults,
		webhookGlobals, webhookDefaults, webhookHardDefaults)
	// Turn Active true into nil
	if newService.Options.GetActive() {
		newService.Options.Active = nil
	}

	// If the Docker type/image/tag is empty, remove the Docker requirement
	if newService.LatestVersion.Require != nil && newService.LatestVersion.Require.Docker != nil {
		dockerType := newService.LatestVersion.Require.Docker.Type
		// Remove the Docker requirement if only the type if there's no image or tag
		if newService.LatestVersion.Require.Docker.Image == "" &&
			newService.LatestVersion.Require.Docker.Tag == "" {

			newService.LatestVersion.Require.Docker = nil
			// If that was the only requirement, remove the requirement
			if newService.LatestVersion.Require.String() == "{}\n" {
				newService.LatestVersion.Require = nil
			}
			// If the Docker type is the same as the default, remove the type
		} else if dockerType == util.FirstNonDefault(
			serviceDefaults.LatestVersion.Require.Docker.Type,
			serviceHardDefaults.LatestVersion.Require.Docker.Type) {
			newService.LatestVersion.Require.Docker.Type = ""
		}
	}

	// If EDIT, give the secrets from the oldService
	newService.giveSecrets(oldService, secretRefs)

	return newService, nil
}

// giveSecretsLatestVersion from the `oldLatestVersion`
func (s *Service) giveSecretsLatestVersion(oldLatestVersion *latestver.Lookup) {
	// Referencing oldService's AccessToken
	if util.DefaultIfNil(s.LatestVersion.AccessToken) == "<secret>" {
		s.LatestVersion.AccessToken = oldLatestVersion.AccessToken
	}
	// New service has a Require
	if s.LatestVersion.Require != nil {
		// with the Require.Docker referencing the oldService's Docker token
		if oldLatestVersion.Require != nil && oldLatestVersion.Require.Docker != nil &&
			s.LatestVersion.Require.Docker != nil && s.LatestVersion.Require.Docker.Token == "<secret>" {
			s.LatestVersion.Require.Docker.Token = oldLatestVersion.Require.Docker.Token
		}
	}
	// GitHubData
	if s.LatestVersion.Type == "github" && oldLatestVersion.Type == "github" {
		s.LatestVersion.GitHubData = oldLatestVersion.GitHubData
	}
}

// giveSecretsDeployedVersion from the `oldDeployedVersion`
func (s *Service) giveSecretsDeployedVersion(oldDeployedVersion *deployedver.Lookup, secretRefs *dvSecretRef) {
	if s.DeployedVersionLookup == nil || oldDeployedVersion == nil {
		return
	}

	if s.DeployedVersionLookup.BasicAuth != nil &&
		s.DeployedVersionLookup.BasicAuth.Password == "<secret>" &&
		oldDeployedVersion.BasicAuth != nil {
		s.DeployedVersionLookup.BasicAuth.Password = oldDeployedVersion.BasicAuth.Password
	}

	// If we have headers in old and new
	if len(s.DeployedVersionLookup.Headers) != 0 &&
		len(oldDeployedVersion.Headers) != 0 {
		for i := range s.DeployedVersionLookup.Headers {
			// If we're referencing a secret of an existing header
			if s.DeployedVersionLookup.Headers[i].Value == "<secret>" {
				// Don't have a secretRef for this header
				if i >= len(secretRefs.Headers) {
					break
				}
				oldIndex := secretRefs.Headers[i].OldIndex
				// Not a reference to an old Header
				if oldIndex == nil {
					continue
				}

				if *oldIndex < len(oldDeployedVersion.Headers) {
					s.DeployedVersionLookup.Headers[i].Value = oldDeployedVersion.Headers[*oldIndex].Value
				}
			}
		}
	}
}

// giveSecretsNotify from the `oldNotifies`
func (s *Service) giveSecretsNotify(oldNotifies *shoutrrr.Slice, secretRefs *map[string]oldStringIndex) {
	//nolint:typecheck
	if s.Notify == nil || oldNotifies == nil ||
		secretRefs == nil || len(*secretRefs) == 0 {
		return
	}

	for i := range s.Notify {
		// {OldIndex: "disc", Type: "discord", ...} - <secret> is mapped to values in the 'disc' Notify
		// Map <secret> in `i` to this index
		oldIndex := (*secretRefs)[i].OldIndex
		// Not a reference to an old Notify?
		if oldIndex == nil {
			continue
		}
		oldNotify := (*oldNotifies)[*oldIndex]
		// Reference doesn't exist?
		if oldNotify == nil {
			continue
		}

		// url_fields
		urlFieldsPossiblyCensored := []string{
			"altid",
			"apikey",
			"botkey",
			"password",
			"token",
			"tokena",
			"tokenb",
		}
		for _, key := range urlFieldsPossiblyCensored {
			if s.Notify[i].URLFields[key] == "<secret>" && oldNotify.URLFields[key] != "" {
				s.Notify[i].URLFields[key] = oldNotify.URLFields[key]
			}
		}

		// params
		paramsPossiblyCensored := []string{
			"devices",
		}
		for _, key := range paramsPossiblyCensored {
			if s.Notify[i].Params[key] == "<secret>" && oldNotify.Params[key] != "" {
				s.Notify[i].Params[key] = oldNotify.Params[key]
			}
		}
	}
}

// giveSecretsWebHook from the `oldWebHooks`
func (s *Service) giveSecretsWebHook(oldWebHooks *webhook.Slice, secretRefs *map[string]whSecretRef) {
	//nolint:typecheck
	if s.WebHook == nil || oldWebHooks == nil ||
		secretRefs == nil || len(*secretRefs) == 0 {
		return
	}

	for i := range s.WebHook {
		// {OldIndex: "update", Type: "github", ...} - <secret> is mapped to values in the 'update' WebHook
		// Map <secret> in `i` to this index
		oldIndex := (*secretRefs)[i].OldIndex
		// Not a reference to an old WebHook?
		if oldIndex == nil {
			continue
		}
		// Reference doesn't exist?
		oldWebHook := (*oldWebHooks)[*oldIndex]
		if oldWebHook == nil {
			continue
		}

		// secret
		if s.WebHook[i].Secret == "<secret>" {
			s.WebHook[i].Secret = oldWebHook.Secret
		}

		// custom_headers
		// Check we have headers in old and new
		if s.WebHook[i].CustomHeaders != nil && oldWebHook.CustomHeaders != nil ||
			len((*secretRefs)[i].CustomHeaders) != 0 {
			for hI := range *s.WebHook[i].CustomHeaders {
				// Skip if we're out of range or
				// not referencing a secret of an existing header
				if hI >= len((*secretRefs)[i].CustomHeaders) ||
					(*s.WebHook[i].CustomHeaders)[hI].Value != "<secret>" {
					continue
				}

				// Index for this headers secret in the old Service
				// Map <secret> in `i.hI` to this index
				oldHeaderIndex := (*secretRefs)[i].CustomHeaders[hI].OldIndex
				// New header, or not referencing a previous secret
				if oldHeaderIndex == nil || len(*oldWebHook.CustomHeaders) <= *oldHeaderIndex {
					continue
				}

				// Set the new header value to the old one
				(*s.WebHook[i].CustomHeaders)[hI].Value = (*oldWebHook.CustomHeaders)[*oldHeaderIndex].Value
			}
		}

		// failed
		if oldWebHook.String() == s.WebHook[i].String() {
			s.WebHook[i].Failed.Set(i, oldWebHook.Failed.Get(oldWebHook.ID))
		}
		// next_runnable
		nextRunnable := oldWebHook.NextRunnable()
		s.WebHook[i].SetNextRunnable(&nextRunnable)
	}
}

// giveSecrets will replace <secret> in this Service with that of the oldService and uses secretRefs to find
// secrets inside maps/lists
func (s *Service) giveSecrets(oldService *Service, secretRefs oldSecretRefs) {
	if oldService == nil {
		return
	}

	// Latest Version
	s.giveSecretsLatestVersion(&oldService.LatestVersion)
	// Deployed Version
	s.giveSecretsDeployedVersion(oldService.DeployedVersionLookup, &secretRefs.DeployedVersionLookup)
	// Notify
	s.giveSecretsNotify(&oldService.Notify, &secretRefs.Notify)
	// WebHook
	s.giveSecretsWebHook(&oldService.WebHook, &secretRefs.WebHook)
	// Command
	s.CommandController.CopyFailsFrom(oldService.CommandController)

	// Keep LatestVersion if the LatestVersion lookup is unchanged
	if s.LatestVersion.IsEqual(&oldService.LatestVersion) {
		s.Status.SetApprovedVersion(oldService.Status.ApprovedVersion(), false)
		s.Status.SetLatestVersion(oldService.Status.LatestVersion(), false)
		s.Status.SetLatestVersionTimestamp(oldService.Status.LatestVersionTimestamp())
		s.Status.SetLastQueried(oldService.Status.LastQueried())
	}
	// Keep DeployedVersion if the DeployedVersionLookup is unchanged
	if s.DeployedVersionLookup.IsEqual(oldService.DeployedVersionLookup) &&
		oldService.Options.SemanticVersioning == s.Options.SemanticVersioning {
		s.Status.SetDeployedVersion(oldService.Status.DeployedVersion(), false)
		s.Status.SetDeployedVersionTimestamp(oldService.Status.DeployedVersionTimestamp())
	}
}

// CheckFetches will check that, if set, the LatestVersion and DeployedVersion can be fetched
func (s *Service) CheckFetches() (err error) {
	// Don't check if the service is inactive
	if !s.Options.GetActive() {
		return
	}

	// nil the channels so we don't make any updates
	announceChannel := s.Status.AnnounceChannel
	databaseChannel := s.Status.DatabaseChannel
	s.Status.AnnounceChannel = nil
	s.Status.DatabaseChannel = nil
	// Restore on exit
	defer func() {
		s.Status.AnnounceChannel = announceChannel
		s.Status.DatabaseChannel = databaseChannel
	}()

	logFrom := util.LogFrom{Primary: s.ID, Secondary: "CheckFetches"}

	// Fetch latest version
	{
		// Erase DeployedVersion so that 'require' is checked
		deployedVersion := s.Status.DeployedVersion()
		s.Status.SetDeployedVersion("", false)

		_, err = s.LatestVersion.Query(
			false,
			&logFrom)
		if err != nil {
			err = fmt.Errorf("latest_version - %w", err)
			return
		}
		s.Status.SetDeployedVersion(deployedVersion, false)
	}

	// Fetch deployed version
	if s.DeployedVersionLookup != nil {
		var version string
		version, err = s.DeployedVersionLookup.Query(
			false,
			&logFrom)
		if err != nil {
			err = fmt.Errorf("deployed_version - %w", err)
			return
		}
		s.Status.SetDeployedVersion(version, false)
	}

	return
}

func removeDefaults(oldService *Service, newService *Service, d *Defaults) {
	notifyDefaults, commandDefaults, webhookDefaults := oldService.UsingDefaults()
	if !notifyDefaults && !commandDefaults && !webhookDefaults {
		return
	}

	// Notify
	if notifyDefaults {
		defaultNotifys := util.SortedKeys(d.Notify)
		usingNotifys := util.SortedKeys(newService.Notify)
		// If the length is different, then we're not using defaults
		if len(defaultNotifys) != len(usingNotifys) {
			notifyDefaults = false
		} else {
			// Check that the keys are the same
			for i, notify := range usingNotifys {
				if defaultNotifys[i] != notify || newService.Notify[notify].String("") != oldService.Notify[notify].String("") {
					notifyDefaults = false
					break
				}
			}
		}
		// If we're using defaults, then remove them
		if notifyDefaults {
			newService.Notify = nil
		}
	}

	// Command
	if commandDefaults {
		if len(newService.Command) != len(d.Command) {
			commandDefaults = false
		} else {
			// Check that the commands are the defaults
			for i, command := range d.Command {
				if newService.Command[i].FormattedString() != command.FormattedString() {
					commandDefaults = false
					break
				}
			}
		}
		// If we're using defaults, then remove them
		if commandDefaults {
			newService.Command = nil
		}
	}

	// WebHook
	if webhookDefaults {
		defaultWebHooks := util.SortedKeys(d.WebHook)
		usingWebHooks := util.SortedKeys(newService.WebHook)
		// If the length is different, then we're not using defaults
		if len(defaultWebHooks) != len(usingWebHooks) {
			webhookDefaults = false
		} else {
			// Check that the keys are the same
			for i, webhook := range usingWebHooks {
				if defaultWebHooks[i] != webhook || newService.WebHook[webhook].String() != oldService.WebHook[webhook].String() {
					webhookDefaults = false
					break
				}
			}
		}
		// If we're using defaults, then remove them
		if webhookDefaults {
			newService.WebHook = nil
		}
	}
}
