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

package deployedver

import (
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func stringPtr(val string) *string {
	return &val
}
func boolPtr(val bool) *bool {
	return &val
}
func testLogging() {
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true
}

func testLookup() *Lookup {
	return &Lookup{
		URL:               "https://invalid.release-argus.io/json",
		AllowInvalidCerts: boolPtr(true),
		JSON:              "version",
		Options: &opt.Options{
			SemanticVersioning: boolPtr(true),
			Defaults:           &opt.Options{},
			HardDefaults:       &opt.Options{},
		},
		Status:       &svcstatus.Status{ServiceID: stringPtr("test")},
		Defaults:     &Lookup{},
		HardDefaults: &Lookup{},
	}
}
