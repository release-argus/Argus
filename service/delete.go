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

// Package service provides the service functionality for Argus.
package service

import (
	dbtype "github.com/release-argus/Argus/db/types"
)

// PrepDelete removes all channels and sets the deleting flag to prepare a service for deletion.
func (s *Service) PrepDelete(removeFromDB bool) {
	s.Status.SetDeleting()

	// Set the channels to nil to prevent the service from triggering further events.
	s.Status.AnnounceChannel = nil
	s.Status.DatabaseChannel = nil
	s.Status.SaveChannel = nil

	// Delete the row for this service in the database.
	if removeFromDB {
		*s.HardDefaults.Status.DatabaseChannel <- dbtype.Message{
			ServiceID: s.ID,
			Delete:    true}
	}

	s.deleteMetrics()
}
