// Copyright [2026] [Argus]
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
	"time"

	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

func testController(announceChannel chan []byte) *Controller {
	svcStatus := status.New(
		announceChannel, nil, nil,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{},
	)
	svcStatus.ServiceInfo.ID = "service_id"

	return NewController(
		svcStatus,
		Commands{{}, {}},
		nil,
		new("14m"),
	)
}

// testRunnableController returns a Controller of three Commands, varied Failed and NextRunnable.
func testRunnableController(serviceID string) (*Controller, []time.Time) {
	controller := NewController(
		&status.Status{
			ServiceInfo: serviceinfo.ServiceInfo{
				ID: serviceID,
			},
		},
		Commands{
			{"date", "+%m-%d-%Y"},
			{"true"},
			{"false"},
		},
		nil,
		new("11m"),
	)
	controller.Failed.Set(1, false)
	controller.Failed.Set(2, true)
	nextRunnables := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute),
		time.Now().UTC().Add(time.Minute),
	}
	for i := range nextRunnables {
		controller.SetNextRunnable(i, nextRunnables[i])
	}

	return controller, nextRunnables
}
