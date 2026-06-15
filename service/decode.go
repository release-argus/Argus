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

package service

import (
	"strconv"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/webhook"
)

// marshalServiceRaw re-encodes a service subtree (overridable for tests).
var marshalServiceRaw = decode.Marshal

// DecodeServices decodes a raw slice into a Services map.
func DecodeServices(
	format string,
	data []byte,
	defaultsCfg DefaultsConfig,
	notifyCfg shoutrrr.Config,
	whCfg webhook.Config,
) (Services, error) {
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}

	// Extract Services.
	var tmp map[string]any
	if err := decode.Unmarshal(format, data, &tmp); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "service",
			Err: err,
		}
	}

	field := make(Services, len(tmp))
	for id, svcRaw := range tmp {
		if svcRaw == nil {
			delete(field, id)
			continue
		}

		// Create each Service.
		raw, err := marshalServiceRaw(format, svcRaw)
		if err != nil {
			return nil, &decode.KeyFieldError{
				Key: "service",
				Err: err,
			}
		}
		svc, err := DecodeService(
			format, raw,
			id,
			defaultsCfg, notifyCfg, whCfg,
		)
		if err != nil {
			return nil, &decode.KeyFieldError{
				Key: "service",
				Err: err,
			}
		}
		field[id] = svc
	}

	return field, nil
}

// DecodeService decodes a raw slice into a Service.
func DecodeService(
	format string,
	data []byte,
	id string,
	defaultsCfg DefaultsConfig,
	notifyCfg shoutrrr.Config,
	whCfg webhook.Config,
) (*Service, error) {
	field := Service{
		ID:           id,
		Defaults:     defaultsCfg.Soft,
		HardDefaults: defaultsCfg.Hard,
	}

	// Unmarshal.
	if err := decode.Unmarshal(format, data, &field); err != nil {
		if id != "" {
			return nil, &decode.KeyFieldError{
				Key: strconv.Quote(field.ID),
				Err: err,
			}
		}
		return nil, err //nolint:wrapcheck
	}

	field.init(
		notifyCfg,
		whCfg,
		defaultsCfg.Hard.Status.AnnounceChannel,
		defaultsCfg.Hard.Status.DatabaseChannel,
		defaultsCfg.Hard.Status.SaveChannel,
	)

	return &field, nil
}

// ApplyOverrides applies format-encoded overrides to a [Service] object.
// If the target is nil, a new [Service] is created.
func ApplyOverrides(
	format string,
	data []byte,
	target *Service,
	id string,
	defaultsCfg DefaultsConfig,
	notifyCfg shoutrrr.Config,
	whCfg webhook.Config,
) (*Service, error) {
	// No overrides.
	if len(data) == 0 {
		return target, nil
	}
	// Remove.
	if decode.IsNull(data) {
		return nil, nil
	}
	// New.
	if target == nil {
		return DecodeService(
			format, data,
			id,
			defaultsCfg, notifyCfg, whCfg,
		)
	}

	// Clear channels as we're changing the Lookup.
	target = target.Copy(false)

	// Apply overrides.
	if err := decode.Unmarshal(format, data, target); err != nil {
		return nil, &decode.KeyFieldError{
			Key: strconv.Quote(id),
			Err: err,
		}
	}

	target.init(notifyCfg, whCfg, nil, nil, nil)

	return target, nil
}
