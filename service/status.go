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

package service

import (
	"time"

	db_types "github.com/release-argus/Argus/db/types"
)

// SetDeployedVersion will set DeployedVersion as well as DeployedVersionTimestamp.
func (s *Service) SetDeployedVersion(version string) {
	s.Status.DeployedVersion = version
	s.Status.DeployedVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
	// Ignore ApprovedVersion if we're on it
	if version == s.Status.ApprovedVersion {
		s.Status.ApprovedVersion = ""
	}

	// Clear the fail status of WebHooks/Commands
	s.WebHook.ResetFails()
	s.CommandController.ResetFails()

	*s.DatabaseChannel <- db_types.Message{
		ServiceID: s.ID,
		Cells: []db_types.Cell{
			{Column: "deployed_version", Value: s.Status.DeployedVersion},
			{Column: "deployed_version_timestamp", Value: s.Status.DeployedVersionTimestamp},
		},
	}
}

// SetLatestVersion will set LatestVersion to `version` and LatestVersionTimestamp to s.LastQueried.
func (s *Service) SetLatestVersion(version string) {
	s.Status.LatestVersion = version
	s.Status.LatestVersionTimestamp = s.Status.LastQueried

	// Clear the fail status of WebHooks/Commands
	s.WebHook.ResetFails()
	s.CommandController.ResetFails()

	*s.DatabaseChannel <- db_types.Message{
		ServiceID: s.ID,
		Cells: []db_types.Cell{
			{Column: "latest_version", Value: s.Status.LatestVersion},
			{Column: "latest_version_timestamp", Value: s.Status.LatestVersionTimestamp},
		},
	}
}
