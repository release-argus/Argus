// Copyright [2022] [Argus]
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
	"sort"
	"time"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v2"
)

var (
	jLog *util.JLog
)

// Slice mapping of WebHook.
type Slice map[string]*WebHook

// String returns a string representation of the Slice.
func (s *Slice) String() string {
	if s == nil {
		return "<nil>"
	}
	yamlBytes, _ := yaml.Marshal(s)
	return string(yamlBytes)
}

// Slice of Header.
type Headers []Header

// WebHook to send.
type WebHook struct {
	ID                string            `yaml:"-" json:"-"`                                                         // Unique across the Slice
	Type              string            `yaml:"type,omitempty" json:"type,omitempty"`                               // "github"/"url"
	URL               string            `yaml:"url,omitempty" json:"url,omitempty"`                                 // "https://example.com"
	AllowInvalidCerts *bool             `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates.
	CustomHeaders     *Headers          `yaml:"custom_headers,omitempty" json:"custom_headers,omitempty"`           // Custom Headers for the WebHook
	Secret            string            `yaml:"secret,omitempty" json:"secret,omitempty"`                           // "SECRET"
	DesiredStatusCode *int              `yaml:"desired_status_code,omitempty" json:"desired_status_code,omitempty"` // e.g. 202
	Delay             string            `yaml:"delay,omitempty" json:"delay,omitempty"`                             // The delay before sending the WebHook.
	MaxTries          *uint             `yaml:"max_tries,omitempty" json:"max_tries,omitempty"`                     // Number of times to attempt sending the WebHook if the desired status code is not received.
	SilentFails       *bool             `yaml:"silent_fails,omitempty" json:"silent_fails,omitempty"`               // Whether to notify if this WebHook fails MaxTries times.
	Failed            *map[string]*bool `yaml:"-" json:"-"`                                                         // Whether the last send attempt failed
	NextRunnable      time.Time         `yaml:"-" json:"-"`                                                         // Time the WebHook can next be run (for staggering)
	Main              *WebHook          `yaml:"-" json:"-"`                                                         // The Webhook that this Webhook is calling (and may override parts of)
	Defaults          *WebHook          `yaml:"-" json:"-"`                                                         // Default values
	HardDefaults      *WebHook          `yaml:"-" json:"-"`                                                         // Hardcoded default values
	Notifiers         *Notifiers        `yaml:"-" json:"-"`                                                         // The Notify's to notify on failures
	ServiceStatus     *svcstatus.Status `yaml:"-" json:"-"`                                                         // Status of the Service (used for templating vars and Announce channel)
	ParentInterval    *string           `yaml:"-" json:"-"`                                                         // Interval between the parent Service's queries
}

// String returns a string representation of the WebHook.
func (w *WebHook) String() string {
	if w == nil {
		return "<nil>"
	}
	yamlBytes, _ := yaml.Marshal(w)
	return string(yamlBytes)
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
func (h *Headers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// try and unmarshal as a Header list
	var headers []Header
	err := unmarshal(&headers)
	if err != nil {
		// it's not a list, try a map
		var headers map[string]string
		err = unmarshal(&headers)
		if err != nil {
			return err
		}
		// sort the map keys
		keys := make([]string, 0, len(headers))
		for k := range headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// convert map to list
		for _, key := range keys {
			*h = append(*h, Header{Key: key, Value: headers[key]})
		}
		return nil
	}
	*h = headers
	return nil
}
