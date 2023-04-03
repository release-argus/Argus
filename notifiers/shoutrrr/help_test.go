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

//go:build unit || integration

package shoutrrr

import (
	"fmt"
	"strings"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func boolPtr(val bool) *bool {
	return &val
}
func intPtr(val int) *int {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func uintPtr(val int) *uint {
	converted := uint(val)
	return &converted
}
func stringifyPointer[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}
func testLogging(level string) {
	jLog = util.NewJLog(level, false)
	jLog.Testing = true
}

func testShoutrrr(failing bool, forService bool, selfSignedCert bool) *Shoutrrr {
	url := "valid.release-argus.io"
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	shoutrrr := &Shoutrrr{
		Type:    "gotify",
		Failed:  nil,
		Options: map[string]string{"max_tries": "1"},
		// trunk-ignore(gitleaks/generic-api-key)
		URLFields: map[string]string{"host": url, "path": "/gotify", "token": "AGE-LlHU89Q56uQ"},
		Params:    map[string]string{},
	}
	if forService {
		shoutrrr.ID = "test"
		shoutrrr.ServiceStatus = &svcstatus.Status{
			ServiceID: stringPtr("service"),
		}
		shoutrrr.ServiceStatus.Fails.Shoutrrr.Init(1)
		shoutrrr.Failed = &shoutrrr.ServiceStatus.Fails.Shoutrrr
		shoutrrr.Main = &Shoutrrr{Options: map[string]string{}, URLFields: map[string]string{}, Params: map[string]string{}}
		shoutrrr.Defaults = &Shoutrrr{Options: map[string]string{}, URLFields: map[string]string{}, Params: map[string]string{}}
		shoutrrr.HardDefaults = &Shoutrrr{Options: map[string]string{}, URLFields: map[string]string{}, Params: map[string]string{}}
	}
	if failing {
		shoutrrr.URLFields["token"] = "invalid"
	}
	return shoutrrr
}
