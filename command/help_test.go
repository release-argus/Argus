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
	"os"
	"testing"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
)

var packageName = "command"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Exit(exitCode)
}

func testController(announce *chan []byte) (control *Controller) {
	control = &Controller{}
	svcStatus := status.New(
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
