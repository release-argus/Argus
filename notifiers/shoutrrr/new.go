// Copyright [2024] [Argus]
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

package shoutrrr

import (
	"fmt"

	shoutrrr_vars "github.com/release-argus/Argus/notifiers/shoutrrr/types"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// TestPayload is the payload for testing a Notify at httpNotifyTest.
type TestPayload struct {
	ServiceName         string            `json:"service_name"`
	ServiceNamePrevious string            `json:"service_name_previous"`
	Name                string            `json:"name"`
	NamePrevious        string            `json:"name_previous"`
	Type                string            `json:"type,omitempty"`
	Options             map[string]string `json:"options"`
	URLFields           map[string]string `json:"url_fields"`
	Params              map[string]string `json:"params"`
	ServiceURL          string            `json:"service_url"`
	WebURL              string            `json:"web_url"`
}

// FromPayload will create a Shoutrrr from a payload.
// Replacing any undefined values with that of the previous Notify.
func FromPayload(
	payload *TestPayload,
	serviceNotify *Shoutrrr,
	mains SliceDefaults,
	defaults SliceDefaults,
	hardDefaults SliceDefaults,
) (s *Shoutrrr, serviceURL string, err error) {
	// No `name` or `name_previous`
	if payload.NamePrevious == "" && payload.Name == "" {
		err = fmt.Errorf("name and/or name_previous are required")
		return
	}

	name := util.FirstNonDefault(payload.Name, payload.NamePrevious)

	// Original Notifier?
	original := &Shoutrrr{}
	if serviceNotify != nil {
		original = serviceNotify
		// Copy that previous Notify Type if not set
		payload.Type = util.FirstNonDefault(payload.Type, serviceNotify.Type)
	}

	// Get the Type, Main, Defaults, and HardDefaults for this Notify
	nType, main, dfault, hardDefault, err := sortDefaults(
		name, payload.Type,
		mains[name], defaults, hardDefaults)
	if err != nil {
		return
	}

	// Merge the payload with the original
	util.InitMap(&payload.Options)
	payload.Options = util.MergeMaps(original.Options, payload.Options, []string{})
	util.InitMap(&payload.URLFields)
	payload.URLFields = util.MergeMaps(original.URLFields, payload.URLFields, shoutrrr_vars.CensorableURLFields[:])
	util.InitMap(&payload.Params)
	payload.Params = util.MergeMaps(original.Params, payload.Params, shoutrrr_vars.CensorableParams[:])

	// Create the Notify
	s = New(
		nil,
		payload.Name,
		&payload.Options,
		&payload.Params,
		nType,
		&payload.URLFields,
		main,
		dfault,
		hardDefault)
	s.ServiceStatus = &svcstatus.Status{}
	s.ServiceStatus.Init(
		1, 0, 0,
		&payload.ServiceName,
		&payload.WebURL,
	)
	s.Failed = &s.ServiceStatus.Fails.Shoutrrr
	serviceURL = payload.ServiceURL

	// Check the final Notify
	err = s.CheckValues("")
	return
}

// sortDefaults will retrieve the Main and Defaults/HardDefaults for the Notify with this Name and Type.
//
// Returns the (Type, Main, Defaults, HardDefaults, error if the Type is invalid).
func sortDefaults(
	name string,
	nType string,
	main *ShoutrrrDefaults,
	defaults SliceDefaults,
	hardDefaults SliceDefaults,
) (string, *ShoutrrrDefaults, *ShoutrrrDefaults, *ShoutrrrDefaults, error) {
	// If a Main doesn't exist with this Name
	if main == nil {
		// Type should be already set, or in the Name.
		nType = util.FirstNonDefault(nType, name)
		main = defaults[nType]
	} else {
		// Type should be already set, in the Main, or in the Name.
		nType = util.FirstNonDefault(nType, main.Type, name)
	}

	// Have Type, so set the Defaults
	dfault := defaults[nType]

	// If a Hard Default doesn't exist with this Type, then this Type is invalid.
	hardDefault := hardDefaults[nType]
	if hardDefault == nil {
		err := fmt.Errorf("invalid type %q", nType)
		return nType, nil, nil, nil, err
	}

	// Check whether there are Defaults for this Type
	if dfault == nil {
		dfault = hardDefault
		// Main may be nil if it was set to Default
		if main == nil {
			main = hardDefault
		}
	}
	return nType, main, dfault, hardDefault, nil
}
