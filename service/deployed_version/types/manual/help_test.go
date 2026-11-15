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

package manual

import (
	"fmt"
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
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

func testLookup(t *testing.T, version string) *Lookup {
	t.Helper()

	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig(t)

	options, err := opt.Decode("yaml", []byte("semantic_versioning: true"), optCfg)
	if err != nil {
		t.Fatalf("opt.Decode() failed: %v", err)
	}

	announceChannel := make(chan []byte, 24)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	svcStatus := status.New(
		announceChannel, databaseChannel, saveChannel,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{},
	)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID:         "serviceID",
			Name:       "serviceName",
			Comment:    "serviceComment",
			ServiceURL: "https://example.com/service/url",
		},
		&dashboard.Options{
			WebURL: "https://example.com",
		},
	)

	lookup, err := Decode(
		"yaml", []byte("version: "+version),
		options,
		svcStatus,
		dvCfg,
	)
	if err != nil {
		t.Fatalf("Decode(version=%q) failed: %v", version, err)
	}

	return lookup
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig() base.DefaultsConfig {
	hardDefaults := &base.Defaults{Options: &opt.Defaults{}}
	hardDefaults.Default()
	hardDefaults.Options.Default()

	return base.DefaultsConfig{
		Soft: &base.Defaults{Options: &opt.Defaults{}},
		Hard: hardDefaults,
	}
}
