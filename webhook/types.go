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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	supportedTypes = []string{
		"github", "gitlab"}
)

// WebHooks is a string map of WebHook.
type WebHooks map[string]*WebHook

// String returns a string representation of the WebHooks.
func (wh *WebHooks) String() string {
	if wh == nil {
		return ""
	}
	return util.ToYAMLString(wh, "")
}

// Headers is a list of Header.
type Headers []Header

// Base is the base struct for WebHook.
type Base struct {
	Type              string   `json:"type,omitempty" yaml:"type,omitempty"`                               // "github"/"url".
	URL               string   `json:"url,omitempty" yaml:"url,omitempty"`                                 // "https://example.com".
	AllowInvalidCerts *bool    `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	CustomHeaders     *Headers `json:"custom_headers,omitempty" yaml:"custom_headers,omitempty"`           // Deprecated: Use Headers.
	Headers           *Headers `json:"headers,omitempty" yaml:"headers,omitempty"`                         // Custom Headers for the WebHook.
	Secret            string   `json:"secret,omitempty" yaml:"secret,omitempty"`                           // 'SECRET'.
	DesiredStatusCode *uint16  `json:"desired_status_code,omitempty" yaml:"desired_status_code,omitempty"` // e.g. 202.
	Delay             string   `json:"delay,omitempty" yaml:"delay,omitempty"`                             // The delay before sending the WebHook.
	MaxTries          *uint8   `json:"max_tries,omitempty" yaml:"max_tries,omitempty"`                     // Amount of times to attempt sending the WebHook until we receive the desired status code.
	SilentFails       *bool    `json:"silent_fails,omitempty" yaml:"silent_fails,omitempty"`               // Whether to notify if this WebHook fails MaxTries times.
}

// Defaults are the default values for WebHook.
type Defaults struct {
	Base `json:",inline" yaml:",inline"`
}

// NewDefaults returns a new Defaults.
func NewDefaults(
	allowInvalidCerts *bool,
	headers *Headers,
	delay string,
	desiredStatusCode *uint16,
	maxTries *uint8,
	secret string,
	silentFails *bool,
	wType string,
	url string,
) *Defaults {
	return &Defaults{
		Base: Base{
			AllowInvalidCerts: allowInvalidCerts,
			Headers:           headers,
			Delay:             delay,
			DesiredStatusCode: desiredStatusCode,
			MaxTries:          maxTries,
			Secret:            secret,
			SilentFails:       silentFails,
			Type:              wType,
			URL:               url}}
}

// String returns a string representation of the Defaults.
func (d *Defaults) String(prefix string) string {
	if d == nil {
		return ""
	}
	return util.ToYAMLString(d, prefix)
}

// WebHooksDefaults is a string map of Defaults.
type WebHooksDefaults map[string]*Defaults

// String returns a string representation of the WebHooksDefaults.
func (whd *WebHooksDefaults) String(prefix string) string {
	if whd == nil {
		return ""
	}

	keys := util.SortedKeys(*whd)
	if len(keys) == 0 {
		return "{}\n"
	}

	var builder strings.Builder
	itemPrefix := prefix + "  "
	for _, k := range keys {
		itemStr := (*whd)[k].String(itemPrefix)
		if itemStr != "" {
			delim := "\n"
			if itemStr == "{}\n" {
				delim = " "
			}
			_, _ = fmt.Fprintf(&builder, "%s%s:%s%s",
				prefix, k, delim, itemStr)
		}
	}

	return builder.String()
}

// WebHook to send for a new version.
type WebHook struct {
	Base `json:",inline" yaml:",inline"`

	ID string `json:"name,omitempty" yaml:"-"` // Unique across the WebHooks.

	mutex  sync.RWMutex         // Mutex for concurrent access.
	Failed *status.FailsWebHook `json:"-" yaml:"-"` // Whether the last send attempt failed.

	Notifiers      *Notifiers     `json:"-" yaml:"-"` // The Notifiers to notify on failures.
	ServiceStatus  *status.Status `json:"-" yaml:"-"` // Status of the Service (used for templating vars, and the Announce channel).
	ParentInterval *string        `json:"-" yaml:"-"` // Interval between the parent Service's queries.

	Main         *Defaults `json:"-" yaml:"-"` // The root Webhook (That this WebHook may override parts of).
	Defaults     *Defaults `json:"-" yaml:"-"` // Default values.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hardcoded default values.
}

func (wh *WebHooks) UnmarshalJSON(data []byte) error {
	var arr []WebHook
	if err := json.Unmarshal(data, &arr); err != nil {
		return err //nolint:wrapcheck
	}
	*wh = make(WebHooks, len(arr))

	for i := range arr {
		(*wh)[arr[i].ID] = &arr[i]
	}
	return nil
}

func (wh *WebHooks) MarshalJSON() ([]byte, error) {
	if wh == nil {
		return []byte("null"), nil
	}

	keys := util.SortedKeys(*wh)
	arr := make([]*WebHook, 0, len(*wh))
	for _, key := range keys {
		arr = append(arr, (*wh)[key])
	}
	return json.Marshal(arr) //nolint:wrapcheck
}

// New WebHook.
func New(
	allowInvalidCerts *bool,
	headers *Headers,
	delay string,
	desiredStatusCode *uint16,
	failed *status.FailsWebHook,
	id string,
	maxTries *uint8,
	notifiers *Notifiers,
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
			URL:               url},
		Failed:         failed,
		Notifiers:      notifiers,
		ParentInterval: parentInterval,
		Main:           main,
		Defaults:       defaults,
		HardDefaults:   hardDefaults}
}

// String returns a string representation of the WebHook.
func (wh *WebHook) String() string {
	if wh == nil {
		return ""
	}
	return util.ToYAMLString(wh, "")
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Shoutrrrs
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `json:"key" yaml:"key"`     // Header key, e.g. X-Sig.
	Value string `json:"value" yaml:"value"` // Value to give the key.
}

// UnmarshalYAML and convert map[string]string to {key: "X", val: "Y"}.
func (h *Headers) UnmarshalYAML(value *yaml.Node) error {
	// try and unmarshal as a Header list.
	var headers []Header
	if err := value.Decode(&headers); err == nil {
		*h = headers
		return nil
	}

	// Treat it as a map?
	var headersMap map[string]string
	if err := value.Decode(&headersMap); err != nil {
		return err //nolint:wrapcheck
	}

	// Sort the map keys.
	keys := util.SortedKeys(headersMap)
	*h = make([]Header, 0, len(keys))

	// Convert map to list.
	for _, key := range keys {
		*h = append(*h, Header{Key: key, Value: headersMap[key]})
	}
	return nil
}
