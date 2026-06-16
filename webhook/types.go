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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"sync"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// #############
// # CONSTANTS #
// #############

var (
	supportedTypes = []string{"github", "gitlab"}
)

// #########
// # TYPES #
// #########

// Config holds root, default, and hard-default WebHook configuration.
type Config struct {
	Root         WebHooksDefaults
	Defaults     *Defaults
	HardDefaults *Defaults
}

// WebHooks is a string map of WebHook.
type WebHooks map[string]*WebHook

// Headers is a list of Header.
type Headers []Header

// Base is the base struct for WebHook.
type Base struct {
	Type              string  `json:"type,omitempty" yaml:"type,omitempty"`                               // "github"/"url".
	URL               string  `json:"url,omitempty" yaml:"url,omitempty"`                                 // "https://example.com".
	AllowInvalidCerts *bool   `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	CustomHeaders     Headers `json:"custom_headers,omitempty" yaml:"custom_headers,omitempty"`           // Deprecated: Use Headers.
	Headers           Headers `json:"headers,omitempty" yaml:"headers,omitempty"`                         // Custom Headers for the WebHook.
	Secret            string  `json:"secret,omitempty" yaml:"secret,omitempty"`                           // 'SECRET'.
	DesiredStatusCode *uint16 `json:"desired_status_code,omitempty" yaml:"desired_status_code,omitempty"` // e.g. 202.
	Delay             string  `json:"delay,omitempty" yaml:"delay,omitempty"`                             // The delay before sending the WebHook.
	MaxTries          *uint8  `json:"max_tries,omitempty" yaml:"max_tries,omitempty"`                     // Amount of times to attempt sending the WebHook until we receive the desired status code.
	SilentFails       *bool   `json:"silent_fails,omitempty" yaml:"silent_fails,omitempty"`               // Whether to notify if this WebHook fails MaxTries times.
}

// WebHooksDefaults is a string map of Defaults.
type WebHooksDefaults map[string]*Defaults

// Defaults are the default values for WebHook.
type Defaults struct {
	Base `json:",inline" yaml:",inline"`
}

// WebHook to send for a new version.
type WebHook struct {
	Base `json:",inline" yaml:",inline"`

	ID string `json:"name,omitempty" yaml:"-"` // Unique across the WebHooks.

	mu     sync.RWMutex         // Mutex for concurrent access.
	Failed *status.FailsWebHook `json:"-" yaml:"-"` // Whether the last send attempt failed.

	Notifiers      Notifiers      `json:"-" yaml:"-"` // The Notifiers to notify on failures.
	ServiceStatus  *status.Status `json:"-" yaml:"-"` // Status of the Service (used for templating vars, and the Announce channel).
	ParentInterval *string        `json:"-" yaml:"-"` // Interval between the parent Service's queries.

	Main         *Defaults `json:"-" yaml:"-"` // The root Webhook (That this WebHook may override parts of).
	Defaults     *Defaults `json:"-" yaml:"-"` // Default values.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hardcoded default values.
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Shoutrrrs
}

// ############
// # DECODING #
// ############

// DecodeDefaults creates and returns new [Defaults] from format-encoded data.
func DecodeDefaults(format string, data []byte) (*Defaults, error) {
	var field Defaults

	// Unmarshal.
	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &field, nil
}

// MarshalJSON implements the json.Marshaler interface.
//
// The map is encoded as a JSON array of WebHook values.
// Entries are sorted by map key to ensure deterministic output.
// The map key corresponds to WebHook.ID and is not separately included in the output.
func (w *WebHooks) MarshalJSON() ([]byte, error) {
	if w == nil {
		return []byte("null"), nil
	}

	keys := util.SortedKeys(*w)
	arr := make([]*WebHook, 0, len(*w))
	for _, key := range keys {
		arr = append(arr, (*w)[key])
	}
	return decode.Marshal("json", arr) //nolint:wrapcheck
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (w *WebHooks) UnmarshalJSON(data []byte) error {
	var arr []WebHook
	if err := decode.Unmarshal("json", data, &arr); err != nil {
		return err //nolint:wrapcheck
	}
	*w = make(WebHooks, len(arr))

	for i := range arr {
		(*w)[arr[i].ID] = &arr[i]
	}
	return nil
}

// New WebHook.
// TODO: polymorphic types.
func New(
	allowInvalidCerts *bool,
	headers Headers,
	delay string,
	desiredStatusCode *uint16,
	failed *status.FailsWebHook,
	id string,
	maxTries *uint8,
	notifiers Notifiers,
	parentInterval *string,
	secret string,
	silentFails *bool,
	wType string,
	url string,
	main *Defaults,
	defaults, hardDefaults *Defaults,
) *WebHook {
	return &WebHook{
		ID: id,
		Base: Base{
			AllowInvalidCerts: allowInvalidCerts,
			Headers:           headers,
			Delay:             delay,
			DesiredStatusCode: desiredStatusCode,
			MaxTries:          maxTries,
			Secret:            secret,
			SilentFails:       silentFails,
			Type:              wType,
			URL:               url,
		},
		Failed:         failed,
		Notifiers:      notifiers,
		ParentInterval: parentInterval,
		Main:           main,
		Defaults:       defaults,
		HardDefaults:   hardDefaults,
	}
}

// Copy returns a deep copy of the WebHooks map.
func (w *WebHooks) Copy(serviceStatus *status.Status, notifiers Notifiers) WebHooks {
	if w == nil {
		return nil
	}

	newWebHooks := make(WebHooks, len(*w))
	for k, v := range *w {
		newWebHooks[k] = v.Copy(serviceStatus, notifiers)
	}
	return newWebHooks
}

// Copy returns a deep copy of the WebHook.
func (w *WebHook) Copy(serviceStatus *status.Status, notifiers Notifiers) *WebHook {
	if w == nil {
		return nil
	}

	return &WebHook{
		Base: Base{
			Type:              w.Type,
			URL:               w.URL,
			AllowInvalidCerts: util.ClonePtr(w.AllowInvalidCerts),
			Headers:           util.CopySlice(w.Headers),
			Secret:            w.Secret,
			DesiredStatusCode: util.ClonePtr(w.DesiredStatusCode),
			Delay:             w.Delay,
			MaxTries:          util.ClonePtr(w.MaxTries),
			SilentFails:       util.ClonePtr(w.SilentFails),
		},
		ID:             w.ID,
		Failed:         w.Failed.Copy(),
		Notifiers:      notifiers,
		ServiceStatus:  serviceStatus,
		ParentInterval: util.ClonePtr(w.ParentInterval),
		Main:           w.Main,
		Defaults:       w.Defaults,
		HardDefaults:   w.HardDefaults,
	}
}

// #########
// # STATE #
// #########

// IsZero implements the yaml.IsZeroer interface.
func (whd WebHooksDefaults) IsZero() bool {
	for _, v := range whd {
		if !v.IsZero() {
			return false
		}
	}
	return true
}

// IsZero implements the yaml.IsZeroer interface.
func (d *Defaults) IsZero() bool {
	return d == nil || (d.Type == "" && d.URL == "" && d.AllowInvalidCerts == nil &&
		len(d.Headers) == 0 && d.Secret == "" && d.DesiredStatusCode == nil &&
		d.Delay == "" && d.MaxTries == nil && d.SilentFails == nil)
}

// IsZero implements the yaml.IsZeroer interface.
func (w *WebHooks) IsZero() bool {
	if w == nil {
		return true
	}

	return len(*w) == 0
}

// IsDefault checks if the WebHook is empty (i.e. all fields are default values).
func (w *WebHook) IsDefault() bool {
	return w.Type == "" && w.URL == "" && w.AllowInvalidCerts == nil && len(w.Headers) == 0 &&
		w.Secret == "" && w.DesiredStatusCode == nil && w.Delay == "" &&
		w.MaxTries == nil && w.SilentFails == nil
}

// #############
// # STRINGIFY #
// #############

// String returns a string representation of the receiver.
func (d Defaults) String(prefix string) string {
	return decode.ToYAMLString(d, prefix)
}

// String returns a string representation of the receiver.
func (whd *WebHooksDefaults) String(prefix string) string {
	if whd == nil {
		return ""
	}

	return decode.ToYAMLString(whd, prefix)
}

// String implements fmt.Stringer and returns a YAML representation of the receiver.
func (w *WebHooks) String() string {
	if w == nil {
		return ""
	}
	return decode.ToYAMLString(w, "")
}

// String returns a string representation of the receiver.
func (w *WebHook) String(prefix string) string {
	if w == nil {
		return ""
	}
	return decode.ToYAMLString(w, prefix)
}
