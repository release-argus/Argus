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

package docker

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

const packageName = "docker"

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

// plainDefaults returns plain defaults and hardDefaults for testing.
func plainDefaults(t *testing.T) (*Defaults, *Defaults) {
	t.Helper()

	hardDefaults, _ := DecodeDefaults("yaml", nil, nil)
	hardDefaults.Default()
	defaults, _ := DecodeDefaults("yaml", nil, hardDefaults)

	defaults.SetDefaults(hardDefaults)

	return defaults, hardDefaults
}

func getTokenData(t *testing.T, auth RegistryAuth) (token, queryToken string, validUntil time.Time) {
	t.Helper()
	switch a := auth.(type) {
	case *ECRAuth:
		queryToken = a.queryToken
		validUntil = a.validUntil
	case *GHCRAuth:
		token = a.GetToken()
		queryToken = a.queryToken
		validUntil = a.validUntil
	case *HubAuth:
		token = a.GetToken()
		queryToken = a.queryToken
		validUntil = a.validUntil
	case *QuayAuth:
		token = a.GetToken()
	}
	return
}
