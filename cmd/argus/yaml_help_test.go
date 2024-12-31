// Copyright [2024] [Argus]
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

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
)

func writeFile(path string, data string, t *testing.T) {
	data = strings.TrimPrefix(data, "\n")
	os.WriteFile(path, []byte(data), 0644)
	t.Cleanup(func() { os.Remove(path) })
}

func testYAML_NoServices(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-no_services.db
			web:
				listen_port: "0"
		defaults:
			service:
				options:
					interval: 123s
					semantic_versioning: false
				latest_version:
					access_token: ` + os.Getenv("GITHUB_TOKEN") + `
				deployed_version: {}
			notify:
				gotify:
					url_fields:
						host: foo.bar
						token: anonymous
				slack:
					params:
						host: defaultHost
						title: defaultTitle
						username: defaultUsername
			webhook:
				desired_status_code: 0
				delay: 2s
				max_tries: 3
				silent_fails: false
		notify:
			default:
				type: gotify
				options:
					message: mainMessage
				params:
					username: mainUsername
		webhook:
			default:
				type: github
				url: https://awx.main.com/api/v2/job_templates/XX/github/
				secret: YYYYmain
				desired_status_code: 202
				delay: 3s
				max_tries: 1
	`)

	writeFile(path, data, t)
}

func testYAML_Argus(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-argus.db
			web:
				listen_port: 0
		service:
			release-argus/Argus:
				latest_version:
					type: github
					url: release-argus/argus
	`)

	writeFile(path, data, t)
}

func testYAML_Argus_SomeInactive(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-argus-some-inactive.db
			web:
				listen_port: 0
		service:
			release-argus/Argus:
				latest_version:
					type: github
					url: release-argus/argus
			release-argus/Argus-Not-Active:
				options:
					active: false
				latest_version:
					type: github
					url: release-argus/argus
	`)

	writeFile(path, data, t)
}
