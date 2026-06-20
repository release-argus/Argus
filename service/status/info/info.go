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

// Package info provides information about the service.
package info

import "sync"

// ServiceInfo holds information about a service.
type ServiceInfo struct {
	mu *sync.RWMutex // Mutex for thread-safe access

	ID      string `json:"id,omitempty"`      // Service ID.
	Name    string `json:"name,omitempty"`    // Service Name.
	Comment string `json:"comment,omitempty"` // Service Comment.
	URL     string `json:"url,omitempty"`     // Service URL.

	Icon       string `json:"icon,omitempty"`         // Icon URL.
	IconLinkTo string `json:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
	WebURL     string `json:"web_url,omitempty"`      // Web URL.

	ApprovedVersion string `json:"approved_version,omitempty"` // The version of the Service that has been approved for deployment.
	DeployedVersion string `json:"deployed_version,omitempty"` // The deployed version of the Service.
	LatestVersion   string `json:"latest_version,omitempty"`   // The latest version of the Service.

	Tags []string `json:"tags,omitempty"` // Tags for the Service.
}

// SetMutex sets the mutex pointer for thread-safe access.
func (s *ServiceInfo) SetMutex(mu *sync.RWMutex) {
	s.mu = mu
}

// GetIcon returns the icon URL.
func (s *ServiceInfo) GetIcon() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Icon
}

// GetIconLinkTo returns the URL to redirect Icon clicks to.
func (s *ServiceInfo) GetIconLinkTo() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.IconLinkTo
}

// GetWebURL returns the web URL.
func (s *ServiceInfo) GetWebURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.WebURL
}

// SkipPrefix is the prefix for the ApprovedVersion when skipping a version (e.g. "SKIP_1.2.3").
const SkipPrefix = "SKIP_"

// SkippedVersion returns the version string representing a skip of the given version.
func SkippedVersion(version string) string {
	return SkipPrefix + version
}
