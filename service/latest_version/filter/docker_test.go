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

//go:build unit

package filter

import (
	"encoding/base64"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
)

func TestDockerCheck_GetTag(t *testing.T) {
	// GIVEN a DockerCheck and a version
	testLogging("WARN")
	tests := map[string]struct {
		dockerCheck DockerCheck
		version     string
		want        string
	}{
		"plain version": {
			dockerCheck: DockerCheck{Tag: "1.2.3"},
			version:     "3.2.1",
			want:        "1.2.3"},
		"version template": {
			dockerCheck: DockerCheck{Tag: "{{ version }}.1"},
			version:     "3.2",
			want:        "3.2.1"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// WHEN GetTag is called on it
			got := tc.dockerCheck.GetTag(tc.version)

			// THEN the expected Tag is returned
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestDockerCheck_GetToken(t *testing.T) {
	// GIVEN a DockerCheck
	testLogging("WARN")
	tests := map[string]struct {
		dockerCheck    DockerCheck
		onlyIfEnvToken bool
		errRegex       string
		noToken        bool
	}{
		"no image does nothing": {
			dockerCheck: DockerCheck{Type: "hub"},
		},
		"DockerHub invalid token": {
			errRegex:       "(incorrect username or password|too many failed login attempts|Cannot log into an organization account)",
			onlyIfEnvToken: true,
			dockerCheck: DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: "argus",
				Token:    "invalid"},
		},
		"DockerHub valid token": {
			onlyIfEnvToken: true,
			dockerCheck: DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: os.Getenv("DOCKER_USERNAME"),
				Token:    os.Getenv("DOCKER_TOKEN")},
		},
		"GHCR invalid repo": {
			errRegex: "invalid control character",
			dockerCheck: DockerCheck{
				Type:  "ghcr",
				Image: "	release-argus/argus"},
		},
		"GHCR non-existing repo": {
			errRegex: "invalid repository name",
			dockerCheck: DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus-"},
		},
		"GHCR get default token": {
			dockerCheck: DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus"},
		},
		"GHCR base64 access token": {
			onlyIfEnvToken: true, dockerCheck: DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Token: base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN")))},
		},
		"GHCR plaintext access token": {
			onlyIfEnvToken: true, dockerCheck: DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Token: os.Getenv("GH_TOKEN")},
		},
		"Quay get token": {
			dockerCheck: DockerCheck{
				Type:  "quay",
				Image: "argus-io/argus",
				Token: "foo"},
		},
		"Doesn't get token if before validUntil": {
			onlyIfEnvToken: true,
			noToken:        true,
			dockerCheck: DockerCheck{
				Type:       "ghcr",
				Image:      "release-argus/argus",
				Token:      os.Getenv("GITHUB_TOKEN"),
				validUntil: time.Now().UTC().Add(time.Minute)},
		},
		"Gets token if past validUntil": {
			onlyIfEnvToken: true,
			dockerCheck: DockerCheck{
				Type:       "ghcr",
				Image:      "release-argus/argus",
				Token:      os.Getenv("GITHUB_TOKEN"),
				validUntil: time.Now().UTC().Add(-time.Minute)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}

			// WHEN getToken is called on it
			token, err := tc.dockerCheck.getToken()

			// THEN the err is what we expect and a token is retrieved when expected
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if tc.dockerCheck.Image == "" {
				return
			}
			if tc.errRegex == "^$" {
				// didn't want a token
				if tc.noToken {
					if token != "" {
						t.Error("didn't expect a token")
					}
					// didn't get a token
				} else if token == "" {
					t.Error("didn't get any token")
				}
			}
		})
	}
}

func TestDockerCheck_CheckToken(t *testing.T) {
	// GIVEN a DockerCheck
	testLogging("WARN")
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"nil DockerCheck does nothing": {
			dockerCheck: nil,
		},
		"DockerHub get default token": {
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "releaseargus/argus"},
		},
		"DockerHub no token for username": {
			errRegex: "token: <required>",
			dockerCheck: &DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: "xxxx"},
		},
		"DockerHub no username for token": {
			errRegex: "username: <required>",
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "releaseargus/argus",
				Token: "xxxx"},
		},
		"DockerHub valid token": {
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: os.Getenv("DOCKER_USERNAME"),
				Token:    os.Getenv("DOCKER_TOKEN")},
		},
		"GHCR get default token": {
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus"},
		},
		"GHCR base64 access token": {
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Token: base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN")))},
		},
		"GHCR plaintext access token": {
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Token: os.Getenv("GITHUB_TOKEN")},
		},
		"Quay have token": {
			dockerCheck: &DockerCheck{
				Type:  "quay",
				Image: "argus-io/argus",
				Token: "foo"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}

			// WHEN checkToken is called on it
			err := tc.dockerCheck.checkToken()

			// THEN the err is what we expect
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestRequire_DockerTagCheck(t *testing.T) {
	// GIVEN a Require
	testLogging("WARN")
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"nil DockerCheck does nothing": {
			dockerCheck: nil,
		},
		"DockerHub with no token, valid tag": {
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "releaseargus/argus",
				Tag:   "{{ version }}"},
		},
		"DockerHub with no token, invalid tag": {
			errRegex: "tag .+ not found",
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "releaseargus/argus",
				Tag:   "0.1.9"},
		},
		"DockerHub valid token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: os.Getenv("DOCKER_USERNAME"),
				Token:    os.Getenv("DOCKER_TOKEN"),
				Tag:      "{{ version }}"},
		},
		"DockerHub valid token, invalid tag": {
			errRegex:       "tag .+ not found",
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: os.Getenv("DOCKER_USERNAME"),
				Token:    os.Getenv("DOCKER_TOKEN"),
				Tag:      "{{ version }}-beta"},
		},
		"GHCR with default token, valid tag": {
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Tag:   "{{ version }}"},
		},
		"GHCR with default token, invalid tag": {
			errRegex: "manifest unknown",
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Tag:   "{{ version }}-beta"},
		},
		"GHCR base64 access token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Token: base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN"))),
				Tag:   "{{ version }}"},
		},
		"GHCR plaintext access token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Token: os.Getenv("GH_TOKEN"),
				Tag:   "{{ version }}"},
		},
		"Quay with no token, valid tag": {
			dockerCheck: &DockerCheck{
				Type:  "quay",
				Image: "argus-io/argus",
				Token: "",
				Tag:   "{{ version }}"},
		},
		"Quay with valid token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:  "quay",
				Image: "argus-io/argus",
				Token: os.Getenv("QUAY_TOKEN"),
				Tag:   "{{ version }}"},
		},
		"Quay with valid token, invalid tag": {
			errRegex:       "tag not found",
			onlyIfEnvToken: true,
			dockerCheck: &DockerCheck{
				Type:  "quay",
				Image: "argus-io/argus",
				Token: os.Getenv("QUAY_TOKEN"),
				Tag:   "{{ version }}-beta"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}
			require := Require{Docker: tc.dockerCheck}

			// WHEN DockerTagCheck is called on it
			err := require.DockerTagCheck("0.9.0")

			// THEN the err is what we expect
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestDockerCheck_RefreshDockerHubToken(t *testing.T) {
	// GIVEN a Require
	testLogging("WARN")
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"valid token": {
			onlyIfEnvToken: true,
			errRegex:       "^$",
			dockerCheck: &DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: os.Getenv("DOCKER_USERNAME"),
				Token:    os.Getenv("DOCKER_TOKEN")},
		},
		"invalid token": {
			onlyIfEnvToken: true,
			errRegex:       "(incorrect username or password|too many failed login attempts|Cannot log into an organization account)",
			dockerCheck: &DockerCheck{
				Type:     "hub",
				Image:    "releaseargus/argus",
				Username: "argus",
				Token:    "dkcr_pat_invalid"},
		},
		"no token": {
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "releaseargus/argus"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}

			// WHEN refreshDockerHubToken is called on it
			err := tc.dockerCheck.refreshDockerHubToken()

			// THEN the err is what we expect
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestDockerCheck_CheckValues(t *testing.T) {
	// GIVEN a DockerCheck
	testLogging("WARN")
	tests := map[string]struct {
		dockerCheck *DockerCheck
		wantImage   string
		errRegex    string
	}{
		"nil DockerCheck": {
			errRegex:    "^$",
			dockerCheck: nil,
		},
		"invalid Type": {
			errRegex: "^-type: .*<invalid>",
			dockerCheck: &DockerCheck{
				Type:  "foo",
				Image: "release-argus/argus",
				Tag:   "1.2.3"},
		},
		"invalid Type and no Image": {
			errRegex: "^-type: .*<invalid>.*-image: .*<required>.*",
			dockerCheck: &DockerCheck{
				Type: "foo",
				Tag:  "1.2.3"},
		},
		"invalid Type and no Image or Tag": {
			errRegex: "^-type: .*<invalid>.*-image: .*<required>.*tag: .*<required>",
			dockerCheck: &DockerCheck{
				Type: "foo"},
		},
		"official docker hub image": {
			errRegex: "^$",
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "release-argus/argus",
				Tag:   "1.2.3"},
		},
		"docker hub type with username but no password": {
			errRegex:  "^-token: <required>",
			wantImage: "library/ubuntu",
			dockerCheck: &DockerCheck{
				Type:     "hub",
				Image:    "ubuntu",
				Tag:      "1.2.3",
				Username: "test"},
		},
		"invalid Image": {
			errRegex: "image: .* <invalid>",
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "	release-argus/argus",
				Tag:   "1.2.3"},
		},
		"invalid Tag templating": {
			errRegex: "tag: .* <invalid>",
			dockerCheck: &DockerCheck{
				Type:  "hub",
				Image: "release-argus/argus",
				Tag:   "{{ version }"},
		},
		"valid Type with image and tag": {
			errRegex: "^$",
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/argus",
				Tag:   "1.2.3"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// WHEN CheckValues is called on it
			err := tc.dockerCheck.CheckValues("-")

			// THEN the err is what we expect
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if tc.wantImage != "" && tc.wantImage != tc.dockerCheck.Image {
				t.Fatalf("expected image to be %q, not: %q",
					tc.wantImage, tc.dockerCheck.Image)
			}
		})
	}
}

func TestDockerCheck_Print(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck *DockerCheck
		lines       int
		dontRender  string
	}{
		"nil DockerCheck": {
			lines:       0,
			dockerCheck: nil,
		},
		"empty DockerCheck": {
			lines:       1,
			dockerCheck: &DockerCheck{},
		},
		"Type": {
			lines: 2,
			dockerCheck: &DockerCheck{
				Type: "ghcr"},
		},
		"Type, Image": {
			lines: 3,
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/Argus"},
		},
		"Type, Image, Tag": {
			lines: 4,
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "release-argus/Argus",
				Tag:   "{{ version }}"},
		},
		"Type, Image, Tag, Username": {
			lines: 5,
			dockerCheck: &DockerCheck{
				Type:     "ghcr",
				Image:    "release-argus/Argus",
				Tag:      "{{ version }}",
				Username: "Test"},
		},
		"Type, Image, Tag, Username, Token": {
			lines: 6,
			dockerCheck: &DockerCheck{
				Type:     "ghcr",
				Image:    "release-argus/Argus",
				Tag:      "{{ version }}",
				Username: "Test",
				Token:    "SECRET_TOKEN"},
			dontRender: "SECRET_TOKEN",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.dockerCheck.Print("")

			// THEN it prints the expected number of lines
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, string(out))
			}
			if tc.dontRender != "" && strings.Contains(string(out), tc.dontRender) {
				t.Errorf("Print shouldn't have printed %q\n%s",
					tc.dontRender, string(out))
			}
		})
	}
}

func TestDockerCheck_String(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		docker *DockerCheck
		want   string
	}{
		"nil": {
			docker: nil,
			want:   "<nil>",
		},
		"empty": {
			docker: &DockerCheck{},
			want: `
type: ""
image: ""
tag: ""
`,
		},
		"filled": {
			docker: &DockerCheck{
				Type:       "ghcr",
				Image:      "release-argus/Argus",
				Tag:        "{{ version }}",
				Username:   "test",
				Token:      "SECRET_TOKEN",
				token:      "won't be printed",
				validUntil: time.Now(),
			},
			want: `
type: ghcr
image: release-argus/Argus
tag: '{{ version }}'
username: test
token: SECRET_TOKEN
`,
		},
		"quotes otherwise invalid yaml strings": {
			docker: &DockerCheck{
				Username: ">123"},
			want: `
type: ""
image: ""
tag: ""
username: '>123'
`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// WHEN the DockerCheck is stringified with String
			got := tc.docker.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
