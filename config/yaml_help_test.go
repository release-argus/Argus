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

package config

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
					url: release-argus/Argus
	`)

	writeFile(path, data, t)
}

func testYAML_config_test(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-config_test.db
			web:
				listen_port: 0
		defaults:
			service:
				options:
					interval: 123
					semantic_versioning: n
				latest_version:
					access_token: ` + os.Getenv("GITHUB_TOKEN") + `
				notify:
					default: {}
				command:
					- - bash
						- /opt/default.sh
				webhook:
					default: {}
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
				delay: 2
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
		service:
			NoDefaults:
				options:
					interval: 10m
				latest_version:
					type: github
					url: release-argus/argusA
					url_commands:
						- type: regex
							regex: v(.*)
					require:
						regex_content: Argus-{{ version }}-linux-amd64
						regex_version: ^[0-9.]+[0-9]$
				notify:
					personal:
						type: mattermost
						options:
							message: overriddenMessage
						url_fields:
							channel: foo
							host: example.io
							token: "123"
						params:
							username: overriddenUsername
				command:
					- - bash
						- /opt/upgrade.sh
				webhook:
					personal:
						type: github
						url: https://awx.example.com/api/v2/job_templates/XX/github/
						secret: YYYY
						desired_status_code: 202
						delay: 3s
						max_tries: 1
			WantDefaults:
				latest_version:
					type: github
					url: release-argus/argusB
					url_commands:
						- type: regex
							regex: v(.*)
					require:
						regex_content: Argus-{{ version }}-linux-amd64
						regex_version: ^[0-9.]+[0-9]$
			Gitea:
				latest_version:
					type: github
					url: go-gitea/giteaC
					url_commands:
						- type: regex
							regex: v(.*)
					require:
						regex_content: gitea-{{ version }}-linux-amd64
						regex_version: ^[0-9.]+[0-9]$
				notify:
					personal:
						type: mattermost
						options:
							delay: 0s
						url_fields:
							host: mattermost.example.com
							port: "443"
							token: ZZZZ
						params:
							icon: https://raw.githubusercontent.com/go-gitea/gitea/main/public/img/logo.png
				webhook:
					personal:
						type: github
						url: https://awx.example.com/api/v2/job_templates/XX/github/
						secret: YYYY
						delay: 0s
			Disabled:
				options:
					active: false
				latest_version:
					type: github
					url: release-argus/argusD
	`)

	writeFile(path, data, t)
}

func testYAML_SomeNilServices(path string, t *testing.T) {
	data := test.TrimYAML(`
		service:
			a:
				latest_version:
					type: github
					url: release-argus/Argus
			b:
			c:
			d:
				latest_version:
					type: github
					url: release-argus/Argus
	`)

	writeFile(path, data, t)
}

func testYAML_SmallConfigTest(path string, t *testing.T) {
	// for the `save.go`
	// if index < 0 {
	// boundary check
	data := test.TrimYAML(`
		settings:
			data: {}
			web: {}
		service:
			some-service:
				options: {}
				latest_version:
					type: github
					url: release-argus/Argus
				dashboard: {}
		`)

	writeFile(path, data, t)
}

func testYAML_Ordering_0(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			web:
				listen_port: 0
			data:
				database_file: 'test-ordering_0.db'
		defaults:
			service:
				latest_version:
					access_token: ` + os.Getenv("GITHUB_TOKEN") + `
				options:
					interval: 123
					semantic_versioning: n

		notify:
			default:
				type: gotify
				options:
					message: mainMessage

		service:
			NoDefaults:
				latest_version:
					type: github
					url: release-argus/argus
					access_token: ghp_other
			WantDefaults:
				latest_version:
					type: github
					url: release-argus/argus
			Disabled:
				options:
					active: false
				latest_version:
					type: github
					url: release-argus/argus
			Gitea:
				latest_version:
					type: github
					url: go-gitea/gitea

		webhook:
			default:
				type: github
				url: https://awx.main.com/api/v2/job_templates/XX/github/
	`)

	writeFile(path, data, t)
}

func testYAML_Ordering_1_no_services(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-ordering_1.db
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
	`)

	writeFile(path, data, t)
}

func testYAML_Ordering_2_obscure_service_names(path string, t *testing.T) {
	data := test.TrimYAML(`
		service:
			"123":
				latest_version:
					type: github
					url: release-argus/argus
			"foo bar":
				latest_version:
					type: github
					url: release-argus/argus
			"foo: bar":
				latest_version:
					type: github
					url: release-argus/argus
			"foo: \"bar\"":
				latest_version:
					type: github
					url: release-argus/argus
			"\"foo: bar\"":
				latest_version:
					type: github
					url: release-argus/argus
			"'foo: bar'":
				latest_version:
					type: github
					url: release-argus/argus
			"\"foo bar\"":
				latest_version:
					type: github
					url: release-argus/argus
			"'foo bar'":
				latest_version:
					type: github
					url: release-argus/argus
			"foo \"bar\"":
				latest_version:
					type: github
					url: release-argus/argus
			"foo: bar, baz":
				latest_version:
					type: github
					url: release-argus/argus
	`)

	writeFile(path, data, t)
}

func testYAML_Ordering_3_empty_line_after_service_line(path string, t *testing.T) {
	data := test.TrimYAML(`
		service:

			C:
				latest_version:
					type: github
					url: release-argus/argus
					access_token: ghp_other
			B:
				latest_version:
					type: github
					url: release-argus/argus

			A:
				latest_version:
					type: github
					url: release-argus/argus
	`)

	writeFile(path, data, t)
}

func testYAML_Ordering_4_multiple_empty_lines_after_service_line(path string, t *testing.T) {
	data := test.TrimYAML(`
		service:
` + strings.Repeat("\n", 3) + `
			P:
				latest_version:
					type: github
					url: release-argus/argus
` + strings.Repeat("\n", 5) + `
			L:
				latest_version:
					type: github
					url: release-argus/argus
` + strings.Repeat("\n", 2) + `
			S:
				latest_version:
					type: github
					url: release-argus/argus
	`)

	writeFile(path, data, t)
}

func testYAML_Ordering_5_eof_is_service_line(path string, t *testing.T) {
	data := test.TrimYAML(`
	settings:
	data:
	database_file: test-ordering_5.db

	service:`)

	writeFile(path, data, t)
}

func testYAML_Ordering_6_no_services_after_service_line_another_block(path string, t *testing.T) {
	data := test.TrimYAML(`
		service:

		settings:
			data:
				database_file: test-ordering_5.db
	`)

	writeFile(path, data, t)
}

func testYAML_Ordering_7_no_services_after_service_line(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-ordering_5.db
		service:
	`)

	writeFile(path, data, t)
}

func testYAML_LoadDefaults(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-config_test.db
			web:
				listen_port: 0
		defaults:
			service:
				options:
					interval: 123
					semantic_versioning: n
				latest_version:
					access_token: ` + os.Getenv("GITHUB_TOKEN") + `
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
					delay: 2
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
		service:
			NoDefaults:
				options:
					interval: 10m
				latest_version:
					type: github
					url: release-argus/argus
					url_commands:
					- type: regex
						regex: v(.*)
					require:
						regex_content: Argus-{{ version }}-linux-amd64
						regex_version: ^[0-9.]+[0-9]$
				notify:
					personal:
						type: mattermost
						options:
							message: overriddenMessage
						url_fields:
							channel: foo
							host: example.io
							token: "123"
						params:
							username: overriddenUsername
				command:
					- - bash
					- /opt/upgrade.sh
				webhook:
					personal:
						type: github
						url: https://awx.example.com/api/v2/job_templates/XX/github/
						secret: YYYY
						desired_status_code: 202
						delay: 3s
						max_tries: 1
			WantDefaults:
				latest_version:
					type: github
					url: release-argus/argus
					url_commands:
					- type: regex
						regex: v(.*)
					require:
						regex_content: Argus-{{ version }}-linux-amd64
						regex_version: ^[0-9.]+[0-9]$
			notify:
				default: {}
			webhook:
				default: {}
			Gitea:
				latest_version:
					type: github
					url: go-gitea/gitea
					url_commands:
					- type: regex
						regex: v(.*)
					require:
						regex_content: gitea-{{ version }}-linux-amd64
						regex_version: ^[0-9.]+[0-9]$
				notify:
					personal:
						type: mattermost
						options:
							delay: 0s
						url_fields:
							host: mattermost.example.com
							port: "443"
							token: ZZZZ
						params:
							icon: https://raw.githubusercontent.com/go-gitea/gitea/main/public/img/logo.png
				webhook:
					personal:
						type: github
						url: https://awx.example.com/api/v2/job_templates/XX/github/
						secret: YYYY
						delay: 0s
			Disabled:
				options:
					active: false
				latest_version:
					type: github
					url: release-argus/argus
	`)

	writeFile(path, data, t)
}

func testYAML_Edit(path string, t *testing.T) {
	data := test.TrimYAML(`
		settings:
			data:
				database_file: test-config_edit.db
		defaults:
			service:
				latest_version:
					access_token: ` + os.Getenv("GITHUB_TOKEN") + `
		service:
			alpha:
				latest_version:
					type: url
					url: https://valid.release-argus.io/plain
					url_commands:
					- type: regex
						regex: v(.*)
			bravo:
				latest_version:
					type: url
					url: https://valid.release-argus.io/plain
					url_commands:
					- type: regex
						regex: ([0-9.]+)
			charlie:
				latest_version:
					type: url
					url: https://valid.release-argus.io/plain
					url_commands:
					- type: regex
						regex: v?([0-9.]+)
	`)

	writeFile(path, data, t)
}
