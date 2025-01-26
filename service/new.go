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

// Package service provides the service functionality for Argus.
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_types "github.com/release-argus/Argus/notify/shoutrrr/types"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

// oldIntIndex to look at for any SecretValues used.
type oldIntIndex struct {
	OldIndex *int `json:"oldIndex,omitempty"`
}

// oldStringIndex to look at for any SecretValues used.
type oldStringIndex struct {
	OldIndex string `json:"oldIndex,omitempty"`
}

// dvSecretRef contains the reference for the DeployedVersionLookup SecretValues.
type dvSecretRef struct {
	Headers []oldIntIndex `json:"headers,omitempty"`
}

// whSecretRef contains the reference for the WebHook SecretValues.
type whSecretRef struct {
	OldIndex      string        `json:"oldIndex,omitempty"`
	CustomHeaders []oldIntIndex `json:"custom_headers,omitempty"`
}

// oldSecretRefs contains the indexes to use for SecretValues.
type oldSecretRefs struct {
	ID                    string                    `json:"id"`
	DeployedVersionLookup dvSecretRef               `json:"deployed_version,omitempty"`
	Notify                map[string]oldStringIndex `json:"notify,omitempty"`
	WebHook               map[string]whSecretRef    `json:"webhook,omitempty"`
}

// FromPayload creates a new/edited Service from a payload.
func FromPayload(
	oldService *Service,
	payload *io.ReadCloser,

	serviceDefaults, serviceHardDefaults *Defaults,

	notifyGlobals *shoutrrr.SliceDefaults,
	notifyDefaults, notifyHardDefaults *shoutrrr.SliceDefaults,

	webhookGlobals *webhook.SliceDefaults,
	webhookDefaults, webhookHardDefaults *webhook.Defaults,

	logFrom logutil.LogFrom,
) (*Service, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(*payload); err != nil {
		return nil, err //nolint:wrapcheck
	}

	// Service.
	newService := &Service{}
	dec1 := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := dec1.Decode(newService); err != nil {
		// Error may result from a nil LatestVersion caused by missing type.
		if newService.LatestVersion == nil && oldService != nil && oldService.LatestVersion != nil {
			// So use the previous Type.
			newService.LatestVersion, _ = latestver.New(
				oldService.LatestVersion.GetType(),
				"yaml", "",
				nil,
				nil,
				&serviceDefaults.LatestVersion, &serviceHardDefaults.LatestVersion)
			dec1 = json.NewDecoder(bytes.NewReader(buf.Bytes()))
			err = dec1.Decode(newService)
		}
		if err != nil {
			logutil.Log.Error(err, logFrom, true)
			logutil.Log.Verbose(
				"Payload: "+buf.String(),
				logFrom, true)
			return nil, err //nolint:wrapcheck
		}
	}

	// SecretRefs.
	dec2 := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	var secretRefs oldSecretRefs
	if err := dec2.Decode(&secretRefs); err != nil {
		logutil.Log.Error(err, logFrom, true)
		return nil, err //nolint:wrapcheck
	}

	// ID + Channels.
	newService.ID = secretRefs.ID
	newService.Status.AnnounceChannel = serviceHardDefaults.Status.AnnounceChannel
	newService.Status.DatabaseChannel = serviceHardDefaults.Status.DatabaseChannel
	newService.Status.SaveChannel = serviceHardDefaults.Status.SaveChannel

	removeDefaults(oldService, newService, serviceDefaults)
	newService.Init(
		serviceDefaults, serviceHardDefaults,
		notifyGlobals, notifyDefaults, notifyHardDefaults,
		webhookGlobals, webhookDefaults, webhookHardDefaults)
	// Turn Active true into nil.
	if newService.Options.GetActive() {
		newService.Options.Active = nil
	}

	// If EDIT, give the secrets from the oldService.
	newService.giveSecrets(oldService, secretRefs)

	return newService, nil
}

// giveSecretsLatestVersion from the `oldLatestVersion`.
func (s *Service) giveSecretsLatestVersion(oldLatestVersion latestver.Lookup) {
	if s == nil || oldLatestVersion == nil {
		return
	}

	// GitHub specific.
	if lv, ok := s.LatestVersion.(*github.Lookup); ok {
		if oldLV, ok := oldLatestVersion.(*github.Lookup); ok {
			// AccessToken
			if lv.AccessToken == util.SecretValue {
				lv.AccessToken = oldLV.AccessToken
			}
		}
	}

	s.LatestVersion.Inherit(oldLatestVersion)
}

// giveSecretsDeployedVersion from the `oldDeployedVersion`.
func (s *Service) giveSecretsDeployedVersion(oldDeployedVersion *deployedver.Lookup, secretRefs *dvSecretRef) {
	if s.DeployedVersionLookup == nil || oldDeployedVersion == nil {
		return
	}

	if s.DeployedVersionLookup.BasicAuth != nil &&
		s.DeployedVersionLookup.BasicAuth.Password == util.SecretValue &&
		oldDeployedVersion.BasicAuth != nil {
		s.DeployedVersionLookup.BasicAuth.Password = oldDeployedVersion.BasicAuth.Password
	}

	// If we have headers in old and new.
	if len(s.DeployedVersionLookup.Headers) != 0 &&
		len(oldDeployedVersion.Headers) != 0 {
		for i := range s.DeployedVersionLookup.Headers {
			// If referencing a secret of an existing header.
			if s.DeployedVersionLookup.Headers[i].Value == util.SecretValue {
				// Don't have a secretRef for this header.
				if i >= len(secretRefs.Headers) {
					break
				}
				oldIndex := secretRefs.Headers[i].OldIndex
				// Not a reference to an old Header.
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

// giveSecretsNotify from the `oldNotifies`.
func (s *Service) giveSecretsNotify(oldNotifies shoutrrr.Slice, secretRefs map[string]oldStringIndex) {
	//nolint:typecheck
	if len(s.Notify) == 0 || len(oldNotifies) == 0 ||
		len(secretRefs) == 0 {
		return
	}

	for i, notify := range s.Notify {
		// {OldIndex: "disc", Type: "discord", ...} - SecretValue maps to values in the 'disc' Notify.
		// Map SecretValue in `i` to this index.
		oldIndex := secretRefs[i].OldIndex
		// Not a reference to an old Notify?
		if oldIndex == "" {
			continue
		}
		oldNotify := oldNotifies[oldIndex]
		// Reference doesn't exist?
		if oldNotify == nil {
			continue
		}

		// url_fields
		util.CopySecretValues(oldNotify.URLFields, notify.URLFields, shoutrrr_types.CensorableURLFields[:])
		// params
		util.CopySecretValues(oldNotify.Params, notify.Params, shoutrrr_types.CensorableParams[:])
	}
}

// giveSecretsWebHook from the `oldWebHooks`.
func (s *Service) giveSecretsWebHook(oldWebHooks webhook.Slice, secretRefs map[string]whSecretRef) {
	//nolint:typecheck
	if s.WebHook == nil || oldWebHooks == nil ||
		len(secretRefs) == 0 {
		return
	}

	for i, wh := range s.WebHook {
		// {OldIndex: "update", Type: "github", ...} - SecretValue maps values in the 'update' WebHook.
		// Map SecretValue in `i` to this index.
		oldIndex := secretRefs[i].OldIndex
		// Not a reference to an old WebHook?
		if oldIndex == "" {
			continue
		}
		// Reference doesn't exist?.
		oldWebHook := oldWebHooks[oldIndex]
		if oldWebHook == nil {
			continue
		}
		whSecretRefs := secretRefs[i]

		// secret.
		if wh.Secret == util.SecretValue {
			wh.Secret = oldWebHook.Secret
		}

		// custom_headers.
		// Check we have headers in old and new.
		if wh.CustomHeaders != nil && oldWebHook.CustomHeaders != nil ||
			len(whSecretRefs.CustomHeaders) != 0 {
			for headerIndex := range *wh.CustomHeaders {
				// Skip if out of range,
				// or not referencing a secret of an existing header.
				if headerIndex >= len(whSecretRefs.CustomHeaders) ||
					(*wh.CustomHeaders)[headerIndex].Value != util.SecretValue {
					continue
				}

				// Index for this headers secret in the old Service.
				// Map SecretValue in `i.hI` to this index.
				oldHeaderIndex := whSecretRefs.CustomHeaders[headerIndex].OldIndex
				// New header, or not referencing a previous secret.
				if oldHeaderIndex == nil || len(*oldWebHook.CustomHeaders) <= *oldHeaderIndex {
					continue
				}

				// Set the new header value to the old one.
				(*wh.CustomHeaders)[headerIndex].Value = (*oldWebHook.CustomHeaders)[*oldHeaderIndex].Value
			}
		}

		// failed
		if oldWebHook.String() == wh.String() {
			wh.Failed.Set(i, oldWebHook.Failed.Get(oldWebHook.ID))
		}
		// next_runnable
		wh.SetNextRunnable(oldWebHook.NextRunnable())
	}
}

// giveSecrets replaces `SecretValue` in this Service with the corresponding value from oldService,
// using secretRefs to locate secrets in maps/lists.
func (s *Service) giveSecrets(oldService *Service, secretRefs oldSecretRefs) {
	if oldService == nil {
		return
	}

	// Latest Version.
	s.giveSecretsLatestVersion(oldService.LatestVersion)
	// Deployed Version.
	s.giveSecretsDeployedVersion(oldService.DeployedVersionLookup, &secretRefs.DeployedVersionLookup)
	// Notify.
	s.giveSecretsNotify(oldService.Notify, secretRefs.Notify)
	// WebHook.
	s.giveSecretsWebHook(oldService.WebHook, secretRefs.WebHook)
	// Command.
	s.CommandController.CopyFailsFrom(oldService.CommandController)

	// Keep LatestVersion if the LatestVersion Lookup is unchanged.
	if s.LatestVersion.IsEqual(s.LatestVersion, oldService.LatestVersion) {
		s.Status.SetApprovedVersion(oldService.Status.ApprovedVersion(), false)
		s.Status.SetLatestVersion(oldService.Status.LatestVersion(), oldService.Status.LatestVersionTimestamp(), false)
		s.Status.SetLastQueried(oldService.Status.LastQueried())
	}
	// Keep DeployedVersion if the DeployedVersionLookup is unchanged.
	if s.DeployedVersionLookup.IsEqual(oldService.DeployedVersionLookup) &&
		oldService.Options.SemanticVersioning == s.Options.SemanticVersioning {
		s.Status.SetDeployedVersion(oldService.Status.DeployedVersion(), oldService.Status.DeployedVersionTimestamp(), false)
	}
}

// CheckFetches verifies that, if set, the LatestVersion and DeployedVersion can be retrieved.
func (s *Service) CheckFetches() error {
	// Don't check if the Service is inactive.
	if !s.Options.GetActive() {
		return nil
	}

	// nil the channels, so we don't make any updates.
	announceChannel := s.Status.AnnounceChannel
	databaseChannel := s.Status.DatabaseChannel
	s.Status.AnnounceChannel = nil
	s.Status.DatabaseChannel = nil
	// Restore on exit.
	defer func() {
		s.Status.AnnounceChannel = announceChannel
		s.Status.DatabaseChannel = databaseChannel
	}()

	logFrom := logutil.LogFrom{Primary: s.ID, Secondary: "CheckFetches"}

	// Fetch latest version.
	{
		// Erase DeployedVersion so that 'require' is checked.
		deployedVersion := s.Status.DeployedVersion()
		s.Status.SetDeployedVersion("", "", false)

		if _, err := s.LatestVersion.Query(
			false,
			logFrom); err != nil {
			return fmt.Errorf("latest_version - %w", err)
		}
		s.Status.SetDeployedVersion(deployedVersion, "", false)
	}

	// Fetch deployed version.
	if s.DeployedVersionLookup != nil {
		version, err := s.DeployedVersionLookup.Query(
			false,
			logFrom)
		if err != nil {
			return fmt.Errorf("deployed_version - %w", err)
		}
		s.Status.SetDeployedVersion(version, "", false)
	}

	return nil
}

// removeDefaults removes Notify/Command/WebHook values from the `newService` that match defaults.
func removeDefaults(oldService, newService *Service, d *Defaults) {
	// Defaults in use.
	notifyDefaults, commandDefaults, webhookDefaults := oldService.UsingDefaults()
	// No defaults in use.
	if !notifyDefaults && !commandDefaults && !webhookDefaults {
		return
	}

	// Notify.
	if notifyDefaults {
		defaultNotifiers := util.SortedKeys(d.Notify)
		usingNotifiers := util.SortedKeys(newService.Notify)
		// If the length differs, defaults not in use.
		if len(defaultNotifiers) != len(usingNotifiers) {
			notifyDefaults = false
		} else {
			// Check whether the Notifier(s) have changed.
			for i, notify := range usingNotifiers {
				// Name has changed or now has values that override the defaults.
				if defaultNotifiers[i] != notify ||
					newService.Notify[notify].String("") != fmt.Sprintf("type: %s\n", newService.Notify[notify].Type) {
					notifyDefaults = false
					break
				}
			}
		}
		// If using defaults, remove them.
		if notifyDefaults {
			newService.Notify = nil
		}
	}

	// Command.
	if commandDefaults {
		if len(newService.Command) != len(d.Command) {
			commandDefaults = false
		} else {
			// Check whether the Command(s) have changed.
			for i, command := range d.Command {
				if newService.Command[i].FormattedString() != command.FormattedString() {
					commandDefaults = false
					break
				}
			}
		}
		// If using defaults, remove them.
		if commandDefaults {
			newService.Command = nil
		}
	}

	// WebHook.
	if webhookDefaults {
		defaultWebHooks := util.SortedKeys(d.WebHook)
		usingWebHooks := util.SortedKeys(newService.WebHook)
		// If the length differs, defaults not in use.
		if len(defaultWebHooks) != len(usingWebHooks) {
			webhookDefaults = false
		} else {
			// Check whether the WebHook(s) have changed.
			for i, wh := range usingWebHooks {
				// Name has changed or now has values that override the defaults.
				if defaultWebHooks[i] != wh || newService.WebHook[wh].String() != fmt.Sprintf("type: %s\n", newService.WebHook[wh].Type) {
					webhookDefaults = false
					break
				}
			}
		}
		// If using defaults, remove them.
		if webhookDefaults {
			newService.WebHook = nil
		}
	}
}
