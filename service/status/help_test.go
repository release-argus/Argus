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

//go:build testing

package service_status

import (
	db_types "github.com/release-argus/Argus/db/types"
)

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}

func testStatus() Status {
	var (
		announceChannel chan []byte           = make(chan []byte, 2)
		saveChannel     chan bool             = make(chan bool, 5)
		databaseChannel chan db_types.Message = make(chan db_types.Message, 5)
	)
	return Status{
		ServiceID:                stringPtr("test"),
		ApprovedVersion:          "1.1.1",
		LatestVersion:            "2.2.2",
		LatestVersionTimestamp:   "2002-02-02T02:02:02Z",
		DeployedVersion:          "0.0.0",
		DeployedVersionTimestamp: "2001-01-01T01:01:01Z",
		LastQueried:              "2002-02-02T00:00:00Z",
		WebURL:                   stringPtr(""),
		AnnounceChannel:          &announceChannel,
		SaveChannel:              &saveChannel,
		DatabaseChannel:          &databaseChannel,
	}
}
