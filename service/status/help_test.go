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

//go:build unit || integration

package status

import (
	"sync"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
)

var (
	metricsMutex sync.RWMutex
)

func testStatus() (status *Status) {
	var (
		announceChannel = make(chan []byte, 2)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
	)
	svcStatus := New(
		&announceChannel, &databaseChannel, &saveChannel,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{})
	status = svcStatus
	status.ServiceInfo.ID = "test"
	status.Init(
		0, 0, 0,
		"test-service", "test-service", "https://example.com/service-url",
		&dashboard.Options{
			Icon:       "https://example.com/icon.png",
			IconLinkTo: "https://example.com/icon-link",
			WebURL:     "https://example.com",
			Tags:       []string{"foo", "bar"}})

	status.SetApprovedVersion("1.1.1", false)
	status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	status.SetDeployedVersion("0.0.0", "2001-01-01T01:01:01Z", false)
	status.SetLastQueried("2002-02-02T00:00:00Z")

	return status
}
