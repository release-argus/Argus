// Copyright [2022] [Argus]
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

package web

import (
	"os"
	"strings"
)

func writeYAML(path string, data string) {
	data = strings.TrimPrefix(data, "\n")
	os.WriteFile(path, []byte(data), 0644)
}

func testYAML_Argus(path string) {
	data := `
settings:
  data:
    database_file: test-web.db
`

	writeYAML(path, data)
}
