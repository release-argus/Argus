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

package types

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/release-argus/Argus/util"
)

var packageName = "api.types"
var secretValueMarshalled string

func TestMain(m *testing.M) {
	// Marshal the secret value '<secret>' -> '\u003csecret\u003e'.
	secretValueMarshalledBytes, _ := json.Marshal(util.SecretValue)
	secretValueMarshalled = string(secretValueMarshalledBytes)

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Exit(exitCode)
}

func testNotify() Notify {
	return Notify{
		URLFields: map[string]string{
			"username":  "a",
			"apikey":    "bb",
			"port":      "ccc",
			"botkey":    "dddd",
			"host":      "eeeee",
			"password":  "ffffff",
			"token":     "ggggggg",
			"tokena":    "hhhhhhhh",
			"tokenb":    "iiiiiiiii",
			"webhookid": "jjjjjjjjjj",
		},
		Params: map[string]string{
			"botname": "a",
			"devices": "bb",
			"color":   "ccc",
			"host":    "dddd",
		},
	}
}
