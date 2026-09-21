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

package shoutrrr

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

var packageName = "shoutrrr"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

func testDefaults(failing bool, selfSignedCert bool) *Defaults {
	host := test.NotifyGotify.HostValid
	if selfSignedCert {
		host = test.NotifyGotify.HostInvalid
	}
	shoutrrr := NewDefaults(
		"gotify",
		map[string]string{
			"max_tries": "1",
		},
		map[string]string{
			"host":  host,
			"path":  test.NotifyGotify.Path,
			"token": test.NotifyGotify.TokenPass,
		},
		map[string]string{},
	)
	if failing {
		shoutrrr.URLFields["token"] = test.NotifyGotify.TokenMalformed
	}
	return shoutrrr
}

func testShoutrrr(failing bool, selfSignedCert bool) *Shoutrrr {
	host := test.NotifyGotify.HostValid
	if selfSignedCert {
		host = test.NotifyGotify.HostInvalid
	}
	shoutrrr := New(
		nil, "",
		"gotify",
		map[string]string{
			"max_tries": "1",
		},
		map[string]string{
			"host":  host,
			"path":  test.NotifyGotify.Path,
			"token": test.NotifyGotify.TokenPass,
		},
		make(map[string]string),
		NewDefaults(
			"",
			make(map[string]string),
			make(map[string]string),
			make(map[string]string),
		),
		NewDefaults(
			"",
			make(map[string]string),
			make(map[string]string),
			make(map[string]string),
		),
		NewDefaults(
			"",
			make(map[string]string),
			make(map[string]string),
			make(map[string]string),
		),
	)
	shoutrrr.Main.InitMaps()
	shoutrrr.Defaults.InitMaps()
	shoutrrr.HardDefaults.InitMaps()

	shoutrrr.ID = "test"
	shoutrrr.ServiceStatus = &status.Status{
		ServiceInfo: serviceinfo.ServiceInfo{
			ID: "service",
		},
	}
	shoutrrr.ServiceStatus.Fails.Shoutrrr.Init(1)
	shoutrrr.Failed = &shoutrrr.ServiceStatus.Fails.Shoutrrr

	if failing {
		shoutrrr.URLFields["token"] = test.NotifyGotify.TokenMalformed
	}
	return shoutrrr
}

func plainConfig(t *testing.T) Config {
	t.Helper()

	defaults := ShoutrrrsDefaults{}
	hardDefaults := ShoutrrrsDefaults{}
	hardDefaults.Default()

	return Config{
		Root:         ShoutrrrsDefaults{},
		Defaults:     defaults,
		HardDefaults: hardDefaults,
	}
}
