// Copyright [2022] [Hymenaios]
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
	"github.com/hymenaios-io/Hymenaios/notifiers/gotify"
	"github.com/hymenaios-io/Hymenaios/notifiers/slack"
	"github.com/hymenaios-io/Hymenaios/webhook"
)

// Slice is a slice mapping of Service.
type Slice map[string]*Service

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	ID                 *string          `yaml:"-"`                             // service_name.
	Type               *string          `yaml:"type,omitempty"`                // "github"/"URL"
	URL                *string          `yaml:"url,omitempty"`                 // type:URL - "https://example.com", type:github - "owner/repo" or "https://github.com/owner/repo".
	AllowInvalidCerts  *bool            `yaml:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates.
	AccessToken        *string          `yaml:"access_token,omitempty"`        // GitHub access token to use.
	SemanticVersioning *bool            `yaml:"semantic_versioning,omitempty"` // default - true  = Version has to follow semantic versioning (https://semver.org/) and be greater than the previous to trigger anything.
	Interval           *string          `yaml:"interval,omitempty"`            // AhBmCs = Sleep A hours, B minutes and C seconds between queries.
	URLCommands        *URLCommandSlice `yaml:"url_commands,omitempty"`        // Commands to filter the release from the URL request.
	RegexContent       *string          `yaml:"regex_content,omitempty"`       // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions.
	RegexVersion       *string          `yaml:"regex_version,omitempty"`       // "v*[0-9.]+" The version found must match this release to trigger new version actions.
	UsePreRelease      *bool            `yaml:"use_prerelease,omitempty"`      // Whether the prerelease tag should be used (prereleases are ignored by default).
	WebURL             *string          `yaml:"web_url,omitempty"`             // URL to provide on the Web UI.
	AutoApprove        *bool            `yaml:"auto_approve,omitempty"`        // default - true = Requre approval before sending WebHook(s) for new releases.
	IgnoreMisses       *bool            `yaml:"ignore_misses,omitempty"`       // Ignore URLCommands that fail (e.g. split on text that doesn't exist).
	Icon               *string          `yaml:"icon,omitempty"`                // Icon URL to use for Slack messages/Web UI.
	Gotify             *gotify.Slice    `yaml:"gotify,omitempty"`              // Service-specific Gotify vars.
	Slack              *slack.Slice     `yaml:"slack,omitempty"`               // Service-specific Slack vars.
	WebHook            *webhook.Slice   `yaml:"webhook,omitempty"`             // Service-specific WebHook vars.
	Status             *Status          `yaml:"status,omitempty"`              // Track the Status of this source (version and regex misses).
	HardDefaults       *Service         `yaml:"-"`                             // Hardcoded default values.
	Defaults           *Service         `yaml:"-"`                             // Default values.
	Announce           *chan []byte     `yaml:"-"`                             // Announce to the WebSocket.
	SaveChannel        *chan bool       `yaml:"-"`                             // Channel for triggering a save of the config.
}

// GitHubRelease is the format of a Release on api.github.com/repos/OWNER/REPO/releases.
type GitHubRelease struct {
	URL             string        `yaml:"url"`
	AssetsURL       string        `yaml:"assets_url"`
	UploadURL       string        `yaml:"upload_url"`
	HTMLURL         string        `yaml:"html_url"`
	ID              uint          `yaml:"id"`
	Author          GitHubAuthor  `yaml:"author"`
	NodeID          string        `yaml:"node_id"`
	TagName         string        `yaml:"tag_name"`
	TargetCommitish string        `yaml:"target_commitish"`
	Name            string        `yaml:"name"`
	Draft           bool          `yaml:"draft"`
	PreRelease      bool          `yaml:"prerelease"`
	CreatedAt       string        `yaml:"created_at"`
	PublishedAt     string        `yaml:"published_at"`
	Assets          []GitHubAsset `yaml:"assets"`
}

// GitHubAuthor is the format of an Author on api.github.com/repos/OWNER/REPO/releases.
type GitHubAuthor struct {
	Login             string `yaml:"login"`
	ID                uint   `yaml:"id"`
	NodeID            string `yaml:"node_id"`
	AvatarURL         string `yaml:"avatar_url"`
	GravatarID        string `yaml:"gravatar_id"`
	URL               string `yaml:"url"`
	HTMLURL           string `yaml:"html_url"`
	FollowersURL      string `yaml:"followers_url"`
	FollowingURL      string `yaml:"following_url"`
	GistsURL          string `yaml:"gists_url"`
	StarredURL        string `yaml:"starred_url"`
	SubscriptionsURL  string `yaml:"subscriptions_url"`
	OrganizationsURL  string `yaml:"organizations_url"`
	ReposURL          string `yaml:"repos_url"`
	EventsURL         string `yaml:"events_url"`
	ReceivedEventsURL string `yaml:"received__events_url"`
	Type              string `yaml:"type"`
	SiteAdmin         bool   `yaml:"site_admin"`
}

// GitHubAsset is the format of an Asset on api.github.com/repos/OWNER/REPO/releases.
type GitHubAsset struct {
	URL                string       `yaml:"url"`
	ID                 uint         `yaml:"id"`
	NodeID             string       `yaml:"node_id"`
	Name               string       `yaml:"name"`
	Label              string       `yaml:"label"`
	Uploader           GitHubAuthor `yaml:"uploader"`
	ContentType        string       `yaml:"content_type"`
	State              string       `yaml:"state"`
	Size               uint         `yaml:"size"`
	DownloadCount      uint         `yaml:"download_count"`
	CreatedAt          string       `yaml:"created_at"`
	UpdatedAt          string       `yaml:"updated_at"`
	BrowserDownloadURL string       `yaml:"browser_download_url"`
}
