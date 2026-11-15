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

var secretValueMarshaled string

func TestMain(m *testing.M) {
	logtest.InitLog()

	// decode.Marshal escapes '<secret>' to '\u003csecret\u003e'.
	secretValueMarshaledBytes, err := decode.Marshal("json", util.SecretValue)
	if err != nil {
		fmt.Printf("decode.Marshal() failed: %v\n", err)
		os.Exit(1)
	}
	secretValueMarshaled = string(secretValueMarshaledBytes)

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}
