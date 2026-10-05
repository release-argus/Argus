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

// Package github provides a github-based lookup type.
package github

import (
	"sync"
	"sync/atomic"

	"github.com/release-argus/Argus/config/decode"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
)

var (
	emptyListETag  atomic.Pointer[string]
	defaultPerPage = 30
)

func init() {
	setEmptyListETag("")
}

// Data contains the information used and retrieved during GitHub requests,
// including the ETag, associated releases, and the usage state of the "/tags" endpoint.
type Data struct {
	mu          sync.RWMutex         // Mutex to protect the Data.
	etag        string               // GitHub ETag for conditional requests https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests.
	perPage     int                  // Number of releases per page.
	releases    []forgetypes.Release // Store Releases tied to an ETag.
	tagFallback bool                 // Whether we have fallen back to using /tags instead of /releases.
}

// DataJSON is the JSON representation of cached GitHub release data.
type DataJSON struct {
	ETag     string               `json:"etag,omitzero"`
	Releases []forgetypes.Release `json:"releases,omitempty"`
}

// String implements fmt.Stringer and returns a JSON representation.
func (g *Data) String() string {
	if g == nil {
		return ""
	}

	jsonStruct := DataJSON{
		ETag:     g.ETag(),
		Releases: g.Releases(),
	}

	return decode.ToJSONString(jsonStruct)
}

// SetTagFallback toggles the TagFallback flag.
func (g *Data) SetTagFallback() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.tagFallback = !g.tagFallback
}

// TagFallback value of the Data.
func (g *Data) TagFallback() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.tagFallback
}

// SetETag of the Data.
func (g *Data) SetETag(etag string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.etag = etag
}

// ETag value of the Data.
func (g *Data) ETag() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.etag
}

// SetPerPage of the Data.
func (g *Data) SetPerPage(foundOnPage int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.perPage = util.ValueOr(g.perPage, defaultPerPage) * foundOnPage
}

// ResetPerPage of the Data.
func (g *Data) ResetPerPage() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.perPage = 0
}

// PerPage value of the Data.
func (g *Data) PerPage() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.perPage
}

// SetReleases of the Data.
func (g *Data) SetReleases(releases []forgetypes.Release) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.releases = releases
}

// Releases stored in the Data.
func (g *Data) Releases() []forgetypes.Release {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.releases
}

// hasReleases reports whether the Data has any releases.
func (g *Data) hasReleases() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.releases) > 0
}

// Copy returns a deep copy of the receiver.
func (g *Data) Copy() *Data {
	g.mu.Lock()
	defer g.mu.Unlock()

	return &Data{
		etag:        g.etag,
		perPage:     g.perPage,
		releases:    g.releases,
		tagFallback: g.tagFallback,
	}
}

// CopyFrom copies the ETag and Releases from the given Data to the provider.
func (g *Data) CopyFrom(from *Data) {
	g.mu.Lock()
	defer g.mu.Unlock()
	from.mu.RLock()
	defer from.mu.RUnlock()

	g.etag = from.etag
	g.releases = from.releases
	g.tagFallback = from.tagFallback
}

// setEmptyListETag sets the ETag for an empty list query on the GitHub API.
func setEmptyListETag(etag string) {
	emptyListETag.Store(&etag)
}

// getEmptyListETag returns the ETag for an empty list query on the GitHub API.
func getEmptyListETag() string {
	return *emptyListETag.Load()
}
