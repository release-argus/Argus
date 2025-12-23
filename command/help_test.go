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

package command

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
	logutil "github.com/release-argus/Argus/util/log"
)

var packageName = "command"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logutil.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty",
			packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

func testController(announceChannel chan []byte) (control *Controller) {
	control = &Controller{}
	svcStatus := status.New(
		announceChannel, nil, nil,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{})
	svcStatus.ServiceInfo.ID = "service_id"
	control.Init(
		svcStatus,
		&Commands{{}, {}},
		nil,
		test.StringPtr("14m"),
	)

	return
}
