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

package util

import (
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

var packageName = "util"

func testServiceInfo() serviceinfo.ServiceInfo {
	return serviceinfo.ServiceInfo{
		ID:   "something",
		Name: "another",
		URL:  "example.com",

		WebURL:     "example.com/other",
		Icon:       "icon.png",
		IconLinkTo: "example.com/link",

		ApprovedVersion: "APPROVED",
		DeployedVersion: "DEPLOYED",
		LatestVersion:   "NEW",
		Tags:            []string{"tag1", "tag2"},
	}
}
