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

// Package webhook provides Webhook functionality to services.
package webhook

import (
	"fmt"
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

// Config holds root, default, and hard-default Webhook configuration.
type Config struct {
	Root         WebhooksDefaults
	Defaults     *Defaults
	HardDefaults *Defaults
}

// Webhooks is a string map of Webhook.
type Webhooks map[string]*Webhook

// Headers is a list of Header.
type Headers []Header

// Base is the base struct for Webhook.
type Base struct {
	Type              string  `json:"type,omitzero" yaml:"type,omitzero"`                               // "github"/"url".
	URL               string  `json:"url,omitzero" yaml:"url,omitzero"`                                 // "https://example.com".
	AllowInvalidCerts *bool   `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Default - false = Disallows invalid HTTPS certificates.
	CustomHeaders     Headers `json:"custom_headers,omitempty" yaml:"custom_headers,omitempty"`         // Deprecated: Use Headers.
	Headers           Headers `json:"headers,omitempty" yaml:"headers,omitempty"`                       // Custom Headers for the Webhook.
	Secret            string  `json:"secret,omitzero" yaml:"secret,omitzero"`                           // 'SECRET'.
	DesiredStatusCode *uint16 `json:"desired_status_code,omitzero" yaml:"desired_status_code,omitzero"` // e.g. 202.
	Delay             string  `json:"delay,omitzero" yaml:"delay,omitzero"`                             // The delay before sending the Webhook.
	MaxTries          *uint8  `json:"max_tries,omitzero" yaml:"max_tries,omitzero"`                     // Number of times to attempt sending the Webhook until we receive the desired status code.
	SilentFails       *bool   `json:"silent_fails,omitzero" yaml:"silent_fails,omitzero"`               // Whether to notify if this Webhook fails MaxTries times.
}

// WebhooksDefaults is a string map of Defaults.
type WebhooksDefaults map[string]*Defaults

// Defaults are the default values for Webhook.
type Defaults struct {
	Base `json:",embed" yaml:",inline"`
}

// Webhook to send for a new version.
type Webhook struct {
	Base `json:",embed" yaml:",inline"`

	ID string `json:"name,omitzero" yaml:"-"` // Unique across the Webhooks.

	mu     sync.RWMutex         // Mutex for concurrent access.
	Failed *status.FailsWebhook `json:"-" yaml:"-"` // Whether the last send attempt failed.

	Notifiers      Notifiers      `json:"-" yaml:"-"` // The Notifiers to notify on failures.
	ServiceStatus  *status.Status `json:"-" yaml:"-"` // Status of the Service (used for templating vars, and the Announce channel).
	ParentInterval *string        `json:"-" yaml:"-"` // Interval between the parent Service's queries.

	Main         *Defaults `json:"-" yaml:"-"` // The root Webhook (That this Webhook may override parts of).
	Defaults     *Defaults `json:"-" yaml:"-"` // Default values.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hardcoded default values.
}

// Notifiers holds the notifiers used when a Webhook fails.
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
// The map is encoded as a JSON array of Webhook values.
// Entries are sorted by map key to ensure deterministic output.
// The map key corresponds to Webhook.ID and is not separately included in the output.
func (w *Webhooks) MarshalJSON() ([]byte, error) {
	if w == nil {
		return []byte("null"), nil
	}

	keys := util.SortedKeys(*w)
	arr := make([]*Webhook, 0, len(*w))
	for _, key := range keys {
		arr = append(arr, (*w)[key])
	}
	return decode.Marshal("json", arr) //nolint:wrapcheck
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// It supports a JSON array of Webhook entries:
//
//	[
//		{ id: "a", ... },
//		{ id: "b", ... }
//	]
//
// which is converted into a map keyed by each entry's ID field.
// As that key must identify the entry, a name is required, and must be unique.
func (w *Webhooks) UnmarshalJSON(data []byte) error {
	var arr []Webhook
	if err := decode.Unmarshal("json", data, &arr); err != nil {
		return err //nolint:wrapcheck
	}

	webhooks := make(Webhooks, len(arr))
	for i := range arr {
		id := arr[i].ID
		if id == "" || webhooks[id] != nil {
			err := &decode.ErrField{
				Key:   "name",
				Value: id,
			}
			if id != "" {
				err.Description = "must be unique"
			}

			return &decode.ErrKeyField{
				Key: fmt.Sprintf("webhook[%d]", i),
				Err: err,
			}
		}
		webhooks[id] = &arr[i]
	}

	*w = webhooks
	return nil
}

// New returns a new [Webhook].
// TODO: polymorphic types.
func New(
	allowInvalidCerts *bool,
	headers Headers,
	delay string,
	desiredStatusCode *uint16,
	failed *status.FailsWebhook,
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
) *Webhook {
	return &Webhook{
		ID:                id,
		AllowInvalidCerts: allowInvalidCerts,
		Headers:           headers,
		Delay:             delay,
		DesiredStatusCode: desiredStatusCode,
		MaxTries:          maxTries,
		Secret:            secret,
		SilentFails:       silentFails,
		Type:              wType,
		URL:               url,
		Failed:            failed,
		Notifiers:         notifiers,
		ParentInterval:    parentInterval,
		Main:              main,
		Defaults:          defaults,
		HardDefaults:      hardDefaults,
	}
}

// Copy returns a deep copy of the Webhooks map.
func (w *Webhooks) Copy(serviceStatus *status.Status, notifiers Notifiers) Webhooks {
	if w == nil {
		return nil
	}

	newWebhooks := make(Webhooks, len(*w))
	for k, v := range *w {
		newWebhooks[k] = v.Copy(serviceStatus, notifiers)
	}
	return newWebhooks
}

// Copy returns a deep copy of the Webhook.
func (w *Webhook) Copy(serviceStatus *status.Status, notifiers Notifiers) *Webhook {
	if w == nil {
		return nil
	}

	return &Webhook{
		Type:              w.Type,
		URL:               w.URL,
		AllowInvalidCerts: util.ClonePtr(w.AllowInvalidCerts),
		Headers:           util.CopySlice(w.Headers),
		Secret:            w.Secret,
		DesiredStatusCode: util.ClonePtr(w.DesiredStatusCode),
		Delay:             w.Delay,
		MaxTries:          util.ClonePtr(w.MaxTries),
		SilentFails:       util.ClonePtr(w.SilentFails),
		ID:                w.ID,
		Failed:            w.Failed.Copy(),
		Notifiers:         notifiers,
		ServiceStatus:     serviceStatus,
		ParentInterval:    util.ClonePtr(w.ParentInterval),
		Main:              w.Main,
		Defaults:          w.Defaults,
		HardDefaults:      w.HardDefaults,
	}
}

// #########
// # STATE #
// #########

// IsZero implements the yaml.IsZeroer interface.
func (whd WebhooksDefaults) IsZero() bool {
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
func (w *Webhooks) IsZero() bool {
	if w == nil {
		return true
	}

	return len(*w) == 0
}

// IsDefault reports whether all Webhook fields are at their default (zero) values.
func (w *Webhook) IsDefault() bool {
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
func (whd *WebhooksDefaults) String(prefix string) string {
	if whd == nil {
		return ""
	}

	return decode.ToYAMLString(whd, prefix)
}

// String implements fmt.Stringer and returns a YAML representation of the receiver.
func (w *Webhooks) String() string {
	if w == nil {
		return ""
	}
	return decode.ToYAMLString(w, "")
}

// String returns a string representation of the receiver.
func (w *Webhook) String(prefix string) string {
	if w == nil {
		return ""
	}
	return decode.ToYAMLString(w, prefix)
}
