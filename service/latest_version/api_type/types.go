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

package types

import (
	"github.com/coreos/go-semver/semver"
)

// Release is the format of a Release on api.github.com/repos/OWNER/REPO/releases.
type Release struct {
	URL             string          `json:"url"`
	AssetsURL       string          `json:"assets_url"`
	UploadURL       string          `json:"upload_url"`
	HTMLURL         string          `json:"html_url"`
	ID              uint            `json:"id"`
	Author          Author          `json:"author"`
	NodeID          string          `json:"node_id"`
	SemanticVersion *semver.Version `json:"-"`
	TagName         string          `json:"tag_name"`
	TargetCommitish string          `json:"target_commitish"`
	Name            string          `json:"name"`
	Draft           bool            `json:"draft"`
	PreRelease      bool            `json:"prerelease"`
	CreatedAt       string          `json:"created_at"`
	PublishedAt     string          `json:"published_at"`
	Assets          []Asset         `json:"assets"`
}

// Author is the format of an Author on api.github.com/repos/OWNER/REPO/releases.
type Author struct {
	Login             string `json:"login"`
	ID                uint   `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received__events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

// Asset is the format of an Asset on api.github.com/repos/OWNER/REPO/releases.
type Asset struct {
	URL                string `json:"url"`
	ID                 uint   `json:"id"`
	NodeID             string `json:"node_id"`
	Name               string `json:"name"`
	Label              string `json:"label"`
	Uploader           Author `json:"uploader"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               uint   `json:"size"`
	DownloadCount      uint   `json:"download_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	BrowserDownloadURL string `json:"browser_download_url"`
}
