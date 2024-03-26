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

package shoutrrr

import (
	"fmt"
	"os"
	"strings"
	"testing"

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
func TestMain(m *testing.M) {
	// initialize jLog
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true
	LogInit(jLog)

	// run other tests
	exitCode := m.Run()

	// exit
	os.Exit(exitCode)
}

func testShoutrrrDefaults(failing bool, selfSignedCert bool) *ShoutrrrDefaults {
	url := "valid.release-argus.io"
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	shoutrrr := NewDefaults(
		"gotify",
		&map[string]string{"max_tries": "1"},
		&map[string]string{},
		// trunk-ignore(gitleaks/generic-api-key)
		&map[string]string{"host": url, "path": "/gotify", "token": "AGE-LlHU89Q56uQ"},
	)
	if failing {
		shoutrrr.URLFields["token"] = "invalid"
	}
	return shoutrrr
}

func testShoutrrr(failing bool, selfSignedCert bool) *Shoutrrr {
	url := "valid.release-argus.io"
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	shoutrrr := New(
		nil, "",
		&map[string]string{"max_tries": "1"},
		&map[string]string{},
		"gotify",
		// trunk-ignore(gitleaks/generic-api-key)
		&map[string]string{"host": url, "path": "/gotify", "token": "AGE-LlHU89Q56uQ"},
		NewDefaults(
			"", nil, nil, nil),
		NewDefaults(
			"", nil, nil, nil),
		NewDefaults(
			"", nil, nil, nil))
	shoutrrr.Main.InitMaps()
	shoutrrr.Defaults.InitMaps()
	shoutrrr.HardDefaults.InitMaps()

	shoutrrr.ID = "test"
	shoutrrr.ServiceStatus = &svcstatus.Status{
		ServiceID: stringPtr("service"),
	}
	shoutrrr.ServiceStatus.Fails.Shoutrrr.Init(1)
	shoutrrr.Failed = &shoutrrr.ServiceStatus.Fails.Shoutrrr

	if failing {
		shoutrrr.URLFields["token"] = "invalid"
	}
	return shoutrrr
}

func trimJSON(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, ": ", ":")
	return str
}
