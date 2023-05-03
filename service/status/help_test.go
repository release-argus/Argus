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

//go:build unit || integration

package svcstatus

import (
	dbtype "github.com/release-argus/Argus/db/types"
)

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}

func testStatus() (status *Status) {
	var (
		announceChannel chan []byte         = make(chan []byte, 2)
		saveChannel     chan bool           = make(chan bool, 5)
		databaseChannel chan dbtype.Message = make(chan dbtype.Message, 5)
	)
	svcStatus := New(
		&announceChannel, &databaseChannel, &saveChannel,
		"", "", "", "", "", "")
	status = svcStatus
	status.ServiceID = stringPtr("test")
	status.WebURL = stringPtr("")
	status.Init(
		0, 0, 0,
		stringPtr("test-service"),
		stringPtr("http://example.com"))
	status.SetApprovedVersion("1.1.1", false)
	status.SetLatestVersion("2.2.2", false)
	status.SetLatestVersionTimestamp("2002-02-02T02:02:02Z")
	status.SetDeployedVersion("0.0.0", false)
	status.SetDeployedVersionTimestamp("2001-01-01T01:01:01Z")
	status.SetLastQueried("2002-02-02T00:00:00Z")
	return status
}
