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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import (
	"errors"
	"fmt"

	"github.com/release-argus/Argus/notify/shoutrrr/types"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// TestPayload is the payload for testing a Notify at httpNotifyTest.
type TestPayload struct {
	ServiceID         string                       `json:"service_id"`
	ServiceIDPrevious string                       `json:"service_id_previous"`
	ServiceName       string                       `json:"service_name"`
	Name              string                       `json:"name"`
	NamePrevious      string                       `json:"name_previous"`
	Type              string                       `json:"type,omitempty"`
	Options           util.MapStringStringOmitNull `json:"options"`
	URLFields         util.MapStringStringOmitNull `json:"url_fields"`
	Params            util.MapStringStringOmitNull `json:"params"`
	ServiceURL        string                       `json:"service_url"`
	WebURL            string                       `json:"web_url"`
}

// FromPayload will create a Shoutrrr from a payload.
// Replacing any undefined values with that of the previous Notify.
//
//	Returns the Notify, the ServiceURL, and any errors encountered.
func FromPayload(
	payload TestPayload,
	serviceNotify *Shoutrrr, serviceStatus *status.Status,
	mains ShoutrrrsDefaults,
	defaults, hardDefaults ShoutrrrsDefaults,
) (*Shoutrrr, error) {
	// No `name` or `name_previous`.
	if payload.NamePrevious == "" && payload.Name == "" {
		return nil, errors.New("name and/or name_previous are required")
	}

	name := util.FirstNonDefault(payload.Name, payload.NamePrevious)

	// Original Notifier?
	var original *Shoutrrr
	if serviceNotify != nil {
		original = serviceNotify
		// Copy that previous Notify Type if not set.
		payload.Type = util.FirstNonDefault(payload.Type, serviceNotify.Type)
	}

	// Get the Type, Main, Defaults, and HardDefaults for this Notify.
	nType, main, typeDefaults, typeHardDefaults, err := resolveDefaults(
		name, payload.Type,
		mains[name], defaults, hardDefaults)
	if err != nil {
		return nil, err
	}

	// Merge the payload with the original.
	util.InitMap(&payload.Options)
	util.InitMap(&payload.URLFields)
	util.InitMap(&payload.Params)
	if original != nil {
		payload.Options = util.MergeMaps(original.Options, payload.Options, []string{})
		payload.URLFields = util.MergeMaps(original.URLFields, payload.URLFields, types.CensorableURLFields[:])
		payload.Params = util.MergeMaps(original.Params, payload.Params, types.CensorableParams[:])
	}

	// Create the Notify.
	s := New(
		nil,
		payload.Name,
		nType,
		payload.Options, payload.URLFields, payload.Params,
		main,
		typeDefaults, typeHardDefaults)
	s.ServiceStatus = serviceStatus
	s.Failed = &s.ServiceStatus.Fails.Shoutrrr

	// Check the final Notify.
	errs, _ := s.CheckValues("")
	if errs != nil {
		return nil, errs
	}
	return s, nil
}

// resolveDefaults resolves the default values for a given name and/or type.
// It returns the resolved type, main defaults, type-specific defaults, type-specific hard defaults, and an error if the type is invalid.
func resolveDefaults(
	name, nType string,
	main *Defaults,
	defaults, hardDefaults ShoutrrrsDefaults,
) (string, *Defaults, *Defaults, *Defaults, error) {
	// If a Main doesn't exist with this Name.
	if main == nil {
		// Type should be already set, or in the Name.
		nType = util.FirstNonDefault(nType, name)
		main = defaults[nType]
	} else {
		// Type should be already set, in the Main, or in the Name.
		nType = util.FirstNonDefault(nType, main.Type, name)
	}

	// Have Type, so set the Defaults.
	typeDefaults := defaults[nType]

	// If a Hard Default doesn't exist with this Type, then this Type is invalid.
	typeHardDefaults := hardDefaults[nType]
	if typeHardDefaults == nil {
		err := fmt.Errorf("invalid type %q", nType)
		return nType, nil, nil, nil, err
	}

	// Check whether Defaults exist for this Type.
	if typeDefaults == nil {
		typeDefaults = typeHardDefaults
		// Main may be nil if it was set to Default.
		if main == nil {
			main = typeHardDefaults
		}
	}
	return nType, main, typeDefaults, typeHardDefaults, nil
}
