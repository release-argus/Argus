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

// Package service provides the service functionality for Argus.
package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_types "github.com/release-argus/Argus/notify/shoutrrr/types"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

// oldSecretRefs contains the indexes to use for SecretValues.
type oldSecretRefs struct {
	ID                    string                           `json:"id"`
	LatestVersion         shared.VSecretRef                `json:"latest_version,omitempty"`
	DeployedVersionLookup shared.VSecretRef                `json:"deployed_version,omitempty"`
	Notify                map[string]shared.OldStringIndex `json:"notify,omitempty"`
	WebHook               map[string]shared.WHSecretRef    `json:"webhook,omitempty"`
}

type oldSecretRefsIncoming struct {
	ID                    string                  `json:"id"`
	LatestVersion         shared.VSecretRef       `json:"latest_version,omitempty"`
	DeployedVersionLookup shared.VSecretRef       `json:"deployed_version,omitempty"`
	Notify                []shared.OldStringIndex `json:"notify,omitempty"`
	WebHook               []shared.WHSecretRef    `json:"webhook,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler for oldSecretRefs.
//
// The Web API represents Notify and WebHook entries as lists,
// but internally they are stored as maps keyed by each entry's ID field.
func (o *oldSecretRefs) UnmarshalJSON(data []byte) error {
	var aux oldSecretRefsIncoming
	if err := decode.Unmarshal("json", data, &aux); err != nil {
		return err //nolint:wrapcheck
	}

	o.ID = aux.ID
	o.LatestVersion = aux.LatestVersion
	o.DeployedVersionLookup = aux.DeployedVersionLookup

	// Convert Notify array -> map
	if aux.Notify != nil {
		o.Notify = make(map[string]shared.OldStringIndex, len(aux.Notify))
		for _, n := range aux.Notify {
			o.Notify[n.Name] = n
		}
	}

	// Convert WebHook array -> map
	if aux.WebHook != nil {
		o.WebHook = make(map[string]shared.WHSecretRef, len(aux.WebHook))
		for _, wh := range aux.WebHook {
			o.WebHook[wh.Name] = wh
		}
	}

	return nil
}

// decodeServiceFromPayload decodes a Service from JSON. (overridable for tests).
var decodeServiceFromPayload = DecodeService

// FromPayload creates a new/edited Service from a payload.
func FromPayload(
	previous *Service,
	payload *io.ReadCloser,

	defaultsCfg DefaultsConfig,
	notifyCfg shoutrrr.Config,
	whCfg webhook.Config,

	logFrom logx.LogFrom,
) (*Service, error) {
	var raw []byte
	raw, err := io.ReadAll(*payload)
	if err != nil {
		logx.Error(err, logFrom, true)
		return nil, err //nolint:wrapcheck
	}

	// SecretRefs.
	var secretRefs oldSecretRefs
	if err := decode.Unmarshal("json", raw, &secretRefs); err != nil {
		err = fmt.Errorf("unmarshal service payload: %w", err)
		logx.Error(err, logFrom, true)
		return nil, err //nolint:wrapcheck
	}

	field, err := decodeServiceFromPayload(
		"json", raw,
		secretRefs.ID,
		defaultsCfg, notifyCfg, whCfg,
	)
	if err != nil {
		logx.Error(err, logFrom, true)
		logx.Debug(
			fmt.Sprintf("Payload: %s", raw),
			logFrom,
			true,
		)
		return nil, err
	}
	if field == nil {
		return nil, errors.New("no service created from payload")
	}

	// Channels.
	field.Status.AnnounceChannel = defaultsCfg.Hard.Status.AnnounceChannel
	field.Status.DatabaseChannel = defaultsCfg.Hard.Status.DatabaseChannel
	field.Status.SaveChannel = defaultsCfg.Hard.Status.SaveChannel

	// If EDIT, give the secrets from the previous.
	field.giveSecrets(previous, secretRefs)

	// Turn Active true into nil.
	if field.Options.GetActive() {
		field.Options.Active = nil
	}

	if err, _ := field.CheckValues(); err != nil {
		return nil, err
	}
	return field, nil
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

	logFrom := logx.LogFrom{Primary: s.ID, Secondary: "CheckFetches"}

	// Fetch latest version.
	if s.LatestVersion != nil {
		// Erase DeployedVersion so that 'require' is checked.
		deployedVersion := s.Status.DeployedVersion()
		s.Status.SetDeployedVersion("", "", false)

		if _, err := s.LatestVersion.Query(
			false,
			logFrom,
		); err != nil {
			return fmt.Errorf("latest_version fetches failed: %w", err)
		}
		s.Status.SetDeployedVersion(deployedVersion, "", false)
	}

	// Fetch deployed version.
	if s.DeployedVersionLookup != nil {
		if err := s.DeployedVersionLookup.Query(
			false,
			logFrom,
		); err != nil {
			return fmt.Errorf("deployed_version fetches failed: %w", err)
		}
	}

	s.Status.SetLastQueried("")
	return nil
}

// giveSecrets replaces `SecretValue` in this Service with the corresponding value from oldService,
// using secretRefs to locate secrets in maps/lists.
func (s *Service) giveSecrets(oldService *Service, secretRefs oldSecretRefs) {
	if oldService == nil {
		return
	}

	// Latest Version.
	s.giveSecretsLatestVersion(oldService.LatestVersion, &secretRefs.LatestVersion)
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

// giveSecretsLatestVersion from the `oldLatestVersion`.
func (s *Service) giveSecretsLatestVersion(oldLatestVersion latestver.Lookup, secretRefs *shared.VSecretRef) {
	if s == nil || s.LatestVersion == nil || oldLatestVersion == nil {
		return
	}

	s.LatestVersion.InheritSecrets(oldLatestVersion, secretRefs)
}

// giveSecretsDeployedVersion from the `oldDeployedVersion`.
func (s *Service) giveSecretsDeployedVersion(oldDeployedVersion deployedver.Lookup, secretRefs *shared.VSecretRef) {
	if s.DeployedVersionLookup == nil || oldDeployedVersion == nil {
		return
	}

	s.DeployedVersionLookup.InheritSecrets(oldDeployedVersion, secretRefs)
}

// giveSecretsNotify from the `oldNotifies`.
func (s *Service) giveSecretsNotify(oldNotifies shoutrrr.Shoutrrrs, secretRefs map[string]shared.OldStringIndex) {
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
		util.RestoreMaskedValues(oldNotify.URLFields, notify.URLFields, shoutrrr_types.CensorableURLFields)
		// params
		util.RestoreMaskedValues(oldNotify.Params, notify.Params, shoutrrr_types.CensorableParams)
	}
}

// giveSecretsWebHook from the `oldWebHooks`.
func (s *Service) giveSecretsWebHook(oldWebHooks webhook.WebHooks, secretRefs map[string]shared.WHSecretRef) {
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

		// headers.
		// Check we have headers in old and new.
		if wh.Headers != nil && oldWebHook.Headers != nil ||
			len(whSecretRefs.Headers) != 0 {
			for headerIndex := range wh.Headers {
				// Skip if out of range,
				// or not referencing a secret of an existing header.
				if headerIndex >= len(whSecretRefs.Headers) ||
					wh.Headers[headerIndex].Value != util.SecretValue {
					continue
				}

				// Index for this headers secret in the old Service.
				// Map SecretValue in `i.hI` to this index.
				oldHeaderIndex := whSecretRefs.Headers[headerIndex].OldIndex
				// Decode header, or not referencing a previous secret.
				if oldHeaderIndex == nil || len(oldWebHook.Headers) <= *oldHeaderIndex {
					continue
				}

				// Set the new header value to the old one.
				wh.Headers[headerIndex].Value = oldWebHook.Headers[*oldHeaderIndex].Value
			}
		}

		// failed
		if oldWebHook.String("") == wh.String("") {
			wh.SetFail(oldWebHook.DidFail())
		}
		// next_runnable
		wh.SetNextRunnable(oldWebHook.NextRunnable())
	}
}
