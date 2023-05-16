// Copyright [2023] [Argus]
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

package webhook

import (
	"fmt"
	"sync"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog           *util.JLog
	supportedTypes = []string{
		"github", "gitlab"}
)

// Slice mapping of WebHook.
type Slice map[string]*WebHook

// String returns a string representation of the Slice.
func (s *Slice) String() (str string) {
	if s != nil {
		str = util.ToYAMLString(s, "")
	}
	return
}

// Slice of Header.
type Headers []Header

// WebHookBase is the base struct for WebHook.
type WebHookBase struct {
	Type              string   `yaml:"type,omitempty" json:"type,omitempty"`                               // "github"/"url"
	URL               string   `yaml:"url,omitempty" json:"url,omitempty"`                                 // "https://example.com"
	AllowInvalidCerts *bool    `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates.
	CustomHeaders     *Headers `yaml:"custom_headers,omitempty" json:"custom_headers,omitempty"`           // Custom Headers for the WebHook
	Secret            string   `yaml:"secret,omitempty" json:"secret,omitempty"`                           // "SECRET"
	DesiredStatusCode *int     `yaml:"desired_status_code,omitempty" json:"desired_status_code,omitempty"` // e.g. 202
	Delay             string   `yaml:"delay,omitempty" json:"delay,omitempty"`                             // The delay before sending the WebHook.
	MaxTries          *uint    `yaml:"max_tries,omitempty" json:"max_tries,omitempty"`                     // Number of times to attempt sending the WebHook if the desired status code is not received.
	SilentFails       *bool    `yaml:"silent_fails,omitempty" json:"silent_fails,omitempty"`               // Whether to notify if this WebHook fails MaxTries times.
}

// WebHookDefaults are the default values for WebHook.
type WebHookDefaults struct {
	WebHookBase `yaml:",inline" json:",inline"`
}

// NewDefaults returns a new WebHookDefaults.
func NewDefaults(
	allowInvalidCerts *bool,
	customHeaders *Headers,
	delay string,
	desiredStatusCode *int,
	maxTries *uint,
	secret string,
	silentFails *bool,
	wType string,
	url string,
) *WebHookDefaults {
	return &WebHookDefaults{
		WebHookBase: WebHookBase{
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

// String returns a string representation of the WebHookDefaults.
func (w *WebHookDefaults) String(prefix string) (str string) {
	if w != nil {
		str = util.ToYAMLString(w, prefix)
	}
	return
}

// SliceDefaults mapping of WebHookDefaults.
type SliceDefaults map[string]*WebHookDefaults

// String returns a string representation of the SliceDefaults.
func (s *SliceDefaults) String(prefix string) (str string) {
	if s == nil {
		return ""
	}

	keys := util.SortedKeys(*s)
	if len(keys) == 0 {
		return "{}\n"
	}

	for _, k := range keys {
		itemStr := (*s)[k].String(prefix + "  ")
		if itemStr != "" {
			delim := "\n"
			if itemStr == "{}\n" {
				delim = " "
			}
			str += fmt.Sprintf("%s%s:%s%s",
				prefix, k, delim, itemStr)
		}
	}

	return
}

// WebHook to send.
type WebHook struct {
	WebHookBase `yaml:",inline" json:",inline"`

	ID string `yaml:"-" json:"-"` // Unique across the Slice

	mutex          sync.RWMutex            `yaml:"-" json:"-"` // Mutex for concurrent access.
	Failed         *svcstatus.FailsWebHook `yaml:"-" json:"-"` // Whether the last send attempt failed
	nextRunnable   time.Time               `yaml:"-" json:"-"` // Time the WebHook can next be run (for staggering)
	Notifiers      *Notifiers              `yaml:"-" json:"-"` // The Notify's to notify on failures
	ServiceStatus  *svcstatus.Status       `yaml:"-" json:"-"` // Status of the Service (used for templating vars and Announce channel)
	ParentInterval *string                 `yaml:"-" json:"-"` // Interval between the parent Service's queries

	Main         *WebHookDefaults `yaml:"-" json:"-"` // The Webhook that this Webhook is calling (and may override parts of)
	Defaults     *WebHookDefaults `yaml:"-" json:"-"` // Default values
	HardDefaults *WebHookDefaults `yaml:"-" json:"-"` // Hardcoded default values
}

// New WebHook.
func New(
	allowInvalidCerts *bool,
	customHeaders *Headers,
	delay string,
	desiredStatusCode *int,
	failed *svcstatus.FailsWebHook,
	maxTries *uint,
	notifiers *Notifiers,
	parentInterval *string,
	secret string,
	silentFails *bool,
	wType string,
	url string,
	main *WebHookDefaults,
	defaults *WebHookDefaults,
	hardDefaults *WebHookDefaults,
) *WebHook {
	return &WebHook{
		WebHookBase: WebHookBase{
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
func (w *WebHook) String() (str string) {
	if w != nil {
		str = util.ToYAMLString(w, "")
	}
	return
}

// Notifiers to use when their WebHook fails.
type Notifiers struct {
	Shoutrrr *shoutrrr.Slice // Shoutrrr
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `yaml:"key" json:"key"`     // Header key, e.g. X-Sig
	Value string `yaml:"value" json:"value"` // Value to give the key
}

// UnmarshalYAML, converting map[string]string to {key: "X", val: "Y"}
func (h *Headers) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	// try and unmarshal as a Header list
	var headers []Header
	err = unmarshal(&headers)
	if err != nil {
		// it's not a list, try a map
		var headers map[string]string
		err = unmarshal(&headers)
		if err != nil {
			return
		}
		// sort the map keys
		keys := util.SortedKeys(headers)

		// convert map to list
		for _, key := range keys {
			*h = append(*h, Header{Key: key, Value: headers[key]})
		}
		return
	}
	*h = headers
	return
}
