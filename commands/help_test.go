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

package command

import (
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestMain(m *testing.M) {
	// initialize jLog
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true

	// run other tests
	exitCode := m.Run()

	// exit
	os.Exit(exitCode)
}

func testController(announce *chan []byte) (control *Controller) {
	control = &Controller{}
	svcStatus := svcstatus.New(
		announce, nil, nil,
		"", "", "", "", "", "")
	svcStatus.ServiceID = test.StringPtr("service_id")
	control.Init(
		svcStatus,
		&Slice{{}, {}},
		nil,
		test.StringPtr("14m"),
	)

	return
}

func testShoutrrr(failing bool, selfSignedCert bool) *shoutrrr.Shoutrrr {
	url := "valid.release-argus.io"
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	shoutrrr := shoutrrr.New(
		nil, "",
		&map[string]string{"max_tries": "1"},
		&map[string]string{},
		"gotify",
		// trunk-ignore(gitleaks/generic-api-key)
		&map[string]string{"host": url, "path": "/gotify", "token": "AGE-LlHU89Q56uQ"},
		shoutrrr.NewDefaults(
			"", nil, nil, nil),
		shoutrrr.NewDefaults(
			"", nil, nil, nil),
		shoutrrr.NewDefaults(
			"", nil, nil, nil))
	shoutrrr.Main.InitMaps()
	shoutrrr.Defaults.InitMaps()
	shoutrrr.HardDefaults.InitMaps()

	shoutrrr.ID = "test"
	shoutrrr.ServiceStatus = &svcstatus.Status{
		ServiceID: test.StringPtr("service"),
	}
	shoutrrr.ServiceStatus.Fails.Shoutrrr.Init(1)
	shoutrrr.Failed = &shoutrrr.ServiceStatus.Fails.Shoutrrr

	if failing {
		shoutrrr.URLFields["token"] = "invalid"
	}
	return shoutrrr
}
