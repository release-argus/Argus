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

package config

import (
	"os"
	"strings"
	"testing"
)

func writeYAML(path string, data string, t *testing.T) {
	data = strings.TrimPrefix(data, "\n")
	os.WriteFile(path, []byte(data), 0644)
	t.Cleanup(func() { os.Remove(path) })
}

func testYAML_Argus(path string, t *testing.T) {
	data := `
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
`

	writeYAML(path, data, t)
	t.Cleanup(func() { os.Remove(path) })
}

func testYAML_ConfigTest(path string, t *testing.T) {
	data := `
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
      access_token: ghp_default
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
`

	writeYAML(path, data, t)
}

func testYAML_Ordering_0(path string, t *testing.T) {
	data := `
settings:
  web:
   listen_port: 0
  data:
    database_file: 'test-ordering_0.db'
defaults:
  service:
    latest_version:
      access_token: ghp_default
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
`

	writeYAML(path, data, t)
	t.Cleanup(func() { os.Remove(path) })
}

func testYAML_Ordering_1(path string, t *testing.T) {
	data := `
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
            access_token: ghp_default
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
`

	writeYAML(path, data, t)
	t.Cleanup(func() { os.Remove(path) })
}

func testYAML_LoadDefaults(path string, t *testing.T) {
	data := `
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
      access_token: ghp_default
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
`

	writeYAML(path, data, t)
}

func testYAML_Edit(path string, t *testing.T) {
	data := `
settings:
  data:
    database_file: test-config_edit.db
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
`

	writeYAML(path, data, t)
	t.Cleanup(func() { os.Remove(path) })
}
