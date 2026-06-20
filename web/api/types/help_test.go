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

package types

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/util"
)

var packageName = "api.types"
var secretValueMarshalled string

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Marshal the secret value '<secret>' -> '\u003csecret\u003e'.
	secretValueMarshalledBytes, _ := decode.Marshal("json", util.SecretValue)
	secretValueMarshalled = string(secretValueMarshalledBytes)

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}
