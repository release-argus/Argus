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
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

// oldSecretRefs contains the indexes to use for SecretValues.
type oldSecretRefs struct {
	ID                    string                           `json:"id"`
	DeployedVersionLookup shared.DVSecretRef               `json:"deployed_version,omitempty"`
	Notify                map[string]shared.OldStringIndex `json:"notify,omitempty"`
	WebHook               map[string]shared.WHSecretRef    `json:"webhook,omitempty"`
}

// UnmarshalJSON converts certain arrays into maps.
func (o *oldSecretRefs) UnmarshalJSON(data []byte) error {
	aux := struct {
		ID                    string                  `json:"id"`
		DeployedVersionLookup shared.DVSecretRef      `json:"deployed_version,omitempty"`
		Notify                []shared.OldStringIndex `json:"notify,omitempty"`
		WebHook               []shared.WHSecretRef    `json:"webhook,omitempty"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err //nolint:wrapcheck
	}

	o.ID = aux.ID
	o.DeployedVersionLookup = aux.DeployedVersionLookup

	// Convert Notify array -> map
	if aux.Notify != nil {
		o.Notify = make(map[string]shared.OldStringIndex, len(aux.Notify))
		for _, n := range aux.Notify {
			o.Notify[n.OldIndex] = n
		}
	}

	// Convert WebHook array -> map
	if aux.WebHook != nil {
		o.WebHook = make(map[string]shared.WHSecretRef, len(aux.WebHook))
		for _, wh := range aux.WebHook {
			o.WebHook[wh.OldIndex] = wh
		}
	}

	return nil
}

// FromPayload creates a new/edited Service from a payload.
func FromPayload(
	oldService *Service,
	payload *io.ReadCloser,

	serviceDefaults, serviceHardDefaults *Defaults,

	notifyGlobals *shoutrrr.ShoutrrrsDefaults,
	notifyDefaults, notifyHardDefaults *shoutrrr.ShoutrrrsDefaults,

	webhookGlobals *webhook.WebHooksDefaults,
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
func (s *Service) giveSecretsDeployedVersion(oldDeployedVersion deployedver.Lookup, secretRefs *shared.DVSecretRef) {
	if s.DeployedVersionLookup == nil || oldDeployedVersion == nil {
		return
	}

	s.DeployedVersionLookup.InheritSecrets(oldDeployedVersion, secretRefs)
}

// giveSecretsNotify from the `oldNotifies`.
func (s *Service) giveSecretsNotify(oldNotifies shoutrrr.Shoutrrrs, secretRefs map[string]shared.OldStringIndex) {
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
func (s *Service) giveSecretsWebHook(oldWebHooks webhook.WebHooks, secretRefs map[string]shared.WHSecretRef) {
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
	if s.LatestVersion != nil {
		s.giveSecretsLatestVersion(oldService.LatestVersion)
	}
	// Deployed Version.
	s.giveSecretsDeployedVersion(oldService.DeployedVersionLookup, &secretRefs.DeployedVersionLookup)
	// Notify.
	s.giveSecretsNotify(oldService.Notify, secretRefs.Notify)
	// WebHook.
	s.giveSecretsWebHook(oldService.WebHook, secretRefs.WebHook)
	// Command.
	s.CommandController.CopyFailsFrom(oldService.CommandController)

	// Keep LatestVersion if the LatestVersion Lookup is unchanged.
	if latestver.IsEqual(s.LatestVersion, oldService.LatestVersion) {
		s.Status.SetApprovedVersion(oldService.Status.ApprovedVersion(), false)
		s.Status.SetLatestVersion(oldService.Status.LatestVersion(), oldService.Status.LatestVersionTimestamp(), false)
		s.Status.SetLastQueried(oldService.Status.LastQueried())
	}
	// Keep DeployedVersion if the DeployedVersionLookup is unchanged.
	if deployedver.IsEqual(s.DeployedVersionLookup, oldService.DeployedVersionLookup) &&
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
		if err := s.DeployedVersionLookup.Query(
			false,
			logFrom); err != nil {
			return fmt.Errorf("deployed_version - %w", err)
		}
	}

	s.Status.SetLastQueried("")
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
			// Check whether the Notifiers have changed.
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
			// Check whether the Commands have changed.
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
			// Check whether the WebHooks have changed.
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
