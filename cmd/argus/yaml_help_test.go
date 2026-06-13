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

package main

import (
	"os"
	"strings"

	"github.com/release-argus/Argus/internal/test"
)

func writeFile(path string, data string) {
	data = strings.TrimPrefix(data, "\n")
	_ = os.WriteFile(path, []byte(data), 0644)
}

func testYAML_NoServices(path string) {
	data := test.TrimYAML(`
		settings:
			web:
				listen_port: "0"

		defaults:
			service:
				options:
					interval: 123s
					semantic_versioning: false
				latest_version:
					access_token: ` + test.GitHubToken(nil) + `
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
				url: https://awx.main.example.com/api/v2/job_templates/XX/github/
				secret: YYYYmain
				desired_status_code: 202
				delay: 3s
				max_tries: 1
	`)

	writeFile(path, data)
}

func testYAML_Argus(path string) {
	data := test.TrimYAML(`
		defaults:
			service:
				latest_version:
					access_token: ` + test.GitHubToken(nil) + `
		settings:
			web:
				listen_port: 0
		service:
			SERVICE_NAME:
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
	`)

	writeFile(path, data)
}

func testYAML_Argus_SomeInactive(path string) {
	data := test.TrimYAML(`
		defaults:
			service:
				latest_version:
					access_token: ` + test.GitHubToken(nil) + `
		settings:
				web:
						listen_port: 0
		service:
				SERVICE_NAME:
						latest_version:
								type: github
								url: ` + test.ArgusGitHubRepo + `
				SERVICE_NAME_NOT_ACTIVE:
						options:
								active: false
						latest_version:
								type: github
								url: ` + test.ArgusGitHubRepo + `
		#     SERVICE_NAME_COMMENTED_OUT:
		#         options:
		#             active: false
		#         latest_version:
		#             type: github
		#             url: ` + test.ArgusGitHubRepo + `
				# SERVICE_NAME_COMMENTED_OUT_INDENT:
				#     options:
				#         active: false
				#     latest_version:
				#         type: github
				#         url: ` + test.ArgusGitHubRepo + `
	`)

	writeFile(path, data)
}
