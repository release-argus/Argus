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

// Package github provides a github-based lookup type.
package github

import (
	"sync"

	"github.com/release-argus/Argus/service/latest_version/types/base"
	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

var (
	emptyListETagMutex sync.RWMutex
	emptyListETag      = `"4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945"`
)

// SetEmptyListETag finds the ETag for an empty list query on the GitHub API
// and sets it to be used as the initial ETag for Data.
func SetEmptyListETag(accessToken string) {
	lookup, _ := New(
		"yaml", "",
		nil,
		nil,
		&base.Defaults{}, &base.Defaults{})
	lookup.URL = "release-argus/.github"
	lookup.AccessToken = accessToken
	lookup.Defaults = &base.Defaults{}
	lookup.HardDefaults = &base.Defaults{}

	// Fallback to /tags to stop the /tags fallback query if on /releases.
	lookup.data.SetTagFallback()
	//#nosec G104 -- Disregard.
	//nolint:errcheck // ^
	lookup.httpRequest(logutil.LogFrom{Primary: "SetEmptyListETag"})

	setEmptyListETag(lookup.data.ETag())
}

// setEmptyListETag sets the ETag for an empty list query on the GitHub API.
func setEmptyListETag(eTag string) {
	emptyListETagMutex.Lock()
	defer emptyListETagMutex.Unlock()

	emptyListETag = eTag
}

// getEmptyListETag returns the ETag for an empty list query on the GitHub API.
func getEmptyListETag() string {
	emptyListETagMutex.RLock()
	defer emptyListETagMutex.RUnlock()

	return emptyListETag
}

// Data contains the information used and retrieved during GitHub requests,
// including the eTag, associated releases, and the usage state of the "/tags" endpoint.
type Data struct {
	mutex       sync.RWMutex           // Mutex to protect the Data.
	eTag        string                 // GitHub ETag for conditional requests https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requestsl.
	releases    []github_types.Release // Store Releases tied to an ETag.
	tagFallback bool                   // Whether we have fallen back to using /tags instead of /releases.
}

// String returns a string representation of the Status.
func (g *Data) String() string {
	if g == nil {
		return ""
	}
	type DataJSON struct {
		ETag     string                 `json:"etag,omitempty"`
		Releases []github_types.Release `json:"releases,omitempty"`
	}

	jsonStruct := DataJSON{
		ETag:     g.ETag(),
		Releases: g.Releases()}

	return util.ToJSONString(jsonStruct)
}

// SetTagFallback will flip the TagFallback bool.
func (g *Data) SetTagFallback() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.tagFallback = !g.tagFallback
}

// TagFallback value of the Data.
func (g *Data) TagFallback() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return g.tagFallback
}

// SetETag of the Data.
func (g *Data) SetETag(etag string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.eTag = etag
}

// ETag value of the Data.
func (g *Data) ETag() string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return g.eTag
}

// SetReleases of the Data.
func (g *Data) SetReleases(releases []github_types.Release) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.releases = releases
}

// Releases stored in the Data.
func (g *Data) Releases() []github_types.Release {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return g.releases
}

// hasReleases returns whether the Data has releases.
func (g *Data) hasReleases() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return len(g.releases) > 0
}

// Copy will return a copy of the ETag, and Releases for the Data.
func (g *Data) Copy() *Data {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return &Data{
		eTag:        g.eTag,
		releases:    g.releases,
		tagFallback: g.tagFallback}
}

// CopyFrom copies the ETag and Releases from the given Data to the provider.
func (g *Data) CopyFrom(from *Data) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	from.mutex.RLock()
	defer from.mutex.RUnlock()

	g.eTag = from.eTag
	g.releases = from.releases
	g.tagFallback = from.tagFallback
}
