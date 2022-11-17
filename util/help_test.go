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

package util

func boolPtr(val bool) *bool {
	return &val
}
func intPtr(val int) *int {
	return &val
}
func stringPtr(val string) *string {
	return &val
}

func testServiceInfo() ServiceInfo {
	return ServiceInfo{
		ID:            "something",
		URL:           "example.com",
		WebURL:        "other.com",
		LatestVersion: "NEW",
	}
}
