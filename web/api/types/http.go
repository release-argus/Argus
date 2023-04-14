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

package apitype

import "time"

// VersionAPI used in /api/v1/version
type VersionAPI struct {
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

// RefreshAPI used in /api/v1/*_version/refresh
type RefreshAPI struct {
	Version string    `json:"version"`
	Error   string    `json:"error,omitempty"`
	Date    time.Time `json:"timestamp"`
}
