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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

var (
	jLog           *util.JLog
	supportedTypes = []string{
		"github", "gitlab"}
)

// Slice mapping of WebHook.
type Slice map[string]*WebHook

// String returns a string representation of the Slice.
func (s *Slice) String() string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, "")
}

// Headers is a list of Header.
type Headers []Header

// Base is the base struct for WebHook.
type Base struct {
	Type              string   `yaml:"type,omitempty" json:"type,omitempty"`                               // "github"/"url".
	URL               string   `yaml:"url,omitempty" json:"url,omitempty"`                                 // "https://example.com".
	AllowInvalidCerts *bool    `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	CustomHeaders     *Headers `yaml:"custom_headers,omitempty" json:"custom_headers,omitempty"`           // Custom Headers for the WebHook.
	Secret            string   `yaml:"secret,omitempty" json:"secret,omitempty"`                           // 'SECRET'.
	DesiredStatusCode *uint16  `yaml:"desired_status_code,omitempty" json:"desired_status_code,omitempty"` // e.g. 202.
	Delay             string   `yaml:"delay,omitempty" json:"delay,omitempty"`                             // The delay before sending the WebHook.
	MaxTries          *uint8   `yaml:"max_tries,omitempty" json:"max_tries,omitempty"`                     // Amount of times to attempt sending the WebHook until we receive the desired status code.
	SilentFails       *bool    `yaml:"silent_fails,omitempty" json:"silent_fails,omitempty"`               // Whether to notify if this WebHook fails MaxTries times.
}

// Defaults are the default values for WebHook.
type Defaults struct {
	Base `yaml:",inline" json:",inline"`
}

// NewDefaults returns a new Defaults.
func NewDefaults(
	allowInvalidCerts *bool,
	customHeaders *Headers,
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
			CustomHeaders:     customHeaders,
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

// SliceDefaults mapping of Defaults.
type SliceDefaults map[string]*Defaults

// String returns a string representation of the SliceDefaults.
func (s *SliceDefaults) String(prefix string) string {
	if s == nil {
		return ""
	}

	keys := util.SortedKeys(*s)
	if len(keys) == 0 {
		return "{}\n"
	}

	var builder strings.Builder
	itemPrefix := prefix + "  "
	for _, k := range keys {
		itemStr := (*s)[k].String(itemPrefix)
		if itemStr != "" {
			delim := "\n"
			if itemStr == "{}\n" {
				delim = " "
			}
			builder.WriteString(fmt.Sprintf("%s%s:%s%s",
				prefix, k, delim, itemStr))
		}
	}

	return builder.String()
}

// WebHook to send for a new version.
type WebHook struct {
	Base `yaml:",inline" json:",inline"`

	ID string `yaml:"-" json:"-"` // Unique across the Slice.

	mutex        sync.RWMutex         // Mutex for concurrent access.
	Failed       *status.FailsWebHook `yaml:"-" json:"-"` // Whether the last send attempt failed.
	nextRunnable time.Time            // Time at which the WebHook can next run (for staggering).

	Notifiers      *Notifiers     `yaml:"-" json:"-"` // The Notifiers to notify on failures.
	ServiceStatus  *status.Status `yaml:"-" json:"-"` // Status of the Service (used for templating vars, and the Announce channel).
	ParentInterval *string        `yaml:"-" json:"-"` // Interval between the parent Service's queries.

	Main         *Defaults `yaml:"-" json:"-"` // The root Webhook (That this WebHook may override parts of).
	Defaults     *Defaults `yaml:"-" json:"-"` // Default values.
	HardDefaults *Defaults `yaml:"-" json:"-"` // Hardcoded default values.
}

// New WebHook.
func New(
	allowInvalidCerts *bool,
	customHeaders *Headers,
	delay string,
	desiredStatusCode *uint16,
	failed *status.FailsWebHook,
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
		Base: Base{
			AllowInvalidCerts: allowInvalidCerts,
			CustomHeaders:     customHeaders,
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
func (w *WebHook) String() string {
	if w == nil {
		return ""
	}
	return util.ToYAMLString(w, "")
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Slice
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `yaml:"key" json:"key"`     // Header key, e.g. X-Sig.
	Value string `yaml:"value" json:"value"` // Value to give the key.
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

	// sort the map keys.
	keys := util.SortedKeys(headersMap)
	*h = make([]Header, 0, len(keys))

	// convert map to list.
	for _, key := range keys {
		*h = append(*h, Header{Key: key, Value: headersMap[key]})
	}
	return nil
}
