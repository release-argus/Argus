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

package filters

import (
	"encoding/base64"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/release-argus/Argus/utils"
)

func TestDockerCheckGetTag(t *testing.T) {
	// GIVEN a DockerCheck and a version
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		dockerCheck DockerCheck
		version     string
		want        string
	}{
		"plain version":    {dockerCheck: DockerCheck{Tag: "1.2.3"}, version: "3.2.1", want: "1.2.3"},
		"version template": {dockerCheck: DockerCheck{Tag: "{{ version }}.1"}, version: "3.2", want: "3.2.1"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

func TestDockerCheckGetToken(t *testing.T) {
	// GIVEN a DockerCheck
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		dockerCheck    DockerCheck
		onlyIfEnvToken bool
		errRegex       string
		noToken        bool
	}{
		"no image does nothing": {dockerCheck: DockerCheck{Type: "hub"}},
		"DockerHub invalid repo": {onlyIfEnvToken: true, errRegex: "invalid control character", dockerCheck: DockerCheck{Type: "hub", Image: "	releaseargus/argus",
			Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN")}},
		"DockerHub get default token": {dockerCheck: DockerCheck{Type: "hub", Image: "releaseargus/argus"}},
		"DockerHub invalid token": {errRegex: "(incorrect username or password|too many failed login attempts)", onlyIfEnvToken: true, dockerCheck: DockerCheck{
			Type: "hub", Image: "releaseargus/argus", Username: "argus", Token: "invalid"}},
		"DockerHub valid token": {onlyIfEnvToken: true, dockerCheck: DockerCheck{
			Type: "hub", Image: "releaseargus/argus", Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN")}},
		"GHCR invalid repo":      {errRegex: "invalid control character", dockerCheck: DockerCheck{Type: "ghcr", Image: "	release-argus/argus"}},
		"GHCR non-existing repo": {errRegex: "invalid repository name", dockerCheck: DockerCheck{Type: "ghcr", Image: "release-argus/argus-"}},
		"GHCR get default token": {dockerCheck: DockerCheck{Type: "ghcr", Image: "release-argus/argus"}},
		"GHCR base64 access token": {onlyIfEnvToken: true, dockerCheck: DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN")))}},
		"GHCR plaintext access token": {onlyIfEnvToken: true, dockerCheck: DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: os.Getenv("GITHUB_TOKEN")}},
		"Quay get token": {dockerCheck: DockerCheck{
			Type: "quay", Image: "argus-io/argus", Token: "foo"}},
		"Doesn't get token if before validUntil": {onlyIfEnvToken: true, noToken: true, dockerCheck: DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: os.Getenv("GITHUB_TOKEN"), validUntil: time.Now().UTC().Add(time.Minute)}},
		"Gets token if past validUntil": {onlyIfEnvToken: true, dockerCheck: DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: os.Getenv("GITHUB_TOKEN"), validUntil: time.Now().UTC().Add(-time.Minute)}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}

			// WHEN GetToken is called on it
			token, err := tc.dockerCheck.GetToken()

			// THEN the err is what we expect and a token is retrieved when expected
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			e := utils.ErrorToString(err)
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

func TestDockerCheckCheckToken(t *testing.T) {
	// GIVEN a DockerCheck
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"nil DockerCheck does nothing": {dockerCheck: nil},
		"DockerHub invalid repo": {onlyIfEnvToken: true, errRegex: "token: .* <invalid> .*invalid control character", dockerCheck: &DockerCheck{Type: "hub", Image: "	releaseargus/argus",
			Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN")}},
		"DockerHub get default token": {dockerCheck: &DockerCheck{Type: "hub", Image: "releaseargus/argus"}},
		"DockerHub invalid token": {errRegex: "token: .* <invalid> .*(incorrect username or password|too many failed login attempts)", onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "hub", Image: "releaseargus/argus", Username: "argus", Token: "invalid"}},
		"DockerHub valid token": {onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "hub", Image: "releaseargus/argus", Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN")}},
		"GHCR get default token": {dockerCheck: &DockerCheck{Type: "ghcr", Image: "release-argus/argus"}},
		"GHCR base64 access token": {onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN")))}},
		"GHCR plaintext access token": {onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: os.Getenv("GITHUB_TOKEN")}},
		"Quay have token": {dockerCheck: &DockerCheck{
			Type: "quay", Image: "argus-io/argus", Token: "foo"}},
		"Quay no token": {errRegex: "token: <required>", dockerCheck: &DockerCheck{
			Type: "quay", Image: "argus-io/argus", Token: ""}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}

			// WHEN CheckToken is called on it
			err := tc.dockerCheck.CheckToken()

			// THEN the err is what we expect
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestRequireDockerTagCheck(t *testing.T) {
	// GIVEN a Require
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"nil DockerCheck does nothing": {dockerCheck: nil},
		"DockerHub valid token, invalid repo": {onlyIfEnvToken: true, errRegex: "invalid control character", dockerCheck: &DockerCheck{Type: "hub", Image: "	releaseargus/argus",
			Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN")}},
		"DockerHub with default token, valid tag": {dockerCheck: &DockerCheck{Type: "hub", Image: "releaseargus/argus", Tag: "{{ version }}"}},
		"DockerHub with default token, invalid tag": {errRegex: "tag .+ not found",
			dockerCheck: &DockerCheck{Type: "hub", Image: "releaseargus/argus", Tag: "0.1.9"}},
		"DockerHub valid token, valid tag": {onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "hub", Image: "releaseargus/argus", Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN"), Tag: "{{ version }}"}},
		"DockerHub valid token, invalid tag": {errRegex: "tag .+ not found", onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "hub", Image: "releaseargus/argus", Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN"), Tag: "{{ version }}-beta"}},
		"GHCR with default token, valid tag":   {dockerCheck: &DockerCheck{Type: "ghcr", Image: "release-argus/argus", Tag: "{{ version }}"}},
		"GHCR with default token, invalid tag": {errRegex: "manifest unknown", dockerCheck: &DockerCheck{Type: "ghcr", Image: "release-argus/argus", Tag: "{{ version }}-beta"}},
		"GHCR base64 access token, valid tag": {onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN"))), Tag: "{{ version }}"}},
		"GHCR plaintext access token, valid tag": {onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "ghcr", Image: "release-argus/argus", Token: os.Getenv("GITHUB_TOKEN"), Tag: "{{ version }}"}},
		"Quay with valid token, valid tag": {onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "quay", Image: "argus-io/argus", Token: os.Getenv("QUAY_TOKEN"), Tag: "{{ version }}"}},
		"Quay with valid token, invalid tag": {errRegex: "tag not found", onlyIfEnvToken: true, dockerCheck: &DockerCheck{
			Type: "quay", Image: "argus-io/argus", Token: os.Getenv("QUAY_TOKEN"), Tag: "{{ version }}-beta"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestDockerCheckRefreshDockerHubToken(t *testing.T) {
	// GIVEN a Require
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"DockerHub valid token": {onlyIfEnvToken: true, errRegex: "^$", dockerCheck: &DockerCheck{Type: "hub", Image: "releaseargus/argus",
			Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN")}},
		"DockerHub valid token, invalid repo": {onlyIfEnvToken: true, errRegex: "invalid control character", dockerCheck: &DockerCheck{Type: "hub", Image: "	releaseargus/argus",
			Username: os.Getenv("DOCKER_USERNAME"), Token: os.Getenv("DOCKER_TOKEN")}},
		"DockerHub invalid token": {onlyIfEnvToken: true, errRegex: "(incorrect username or password|too many failed login attempts)", dockerCheck: &DockerCheck{Type: "hub", Image: "releaseargus/argus",
			Username: "argus", Token: "dkcr_pat_invalid"}},
		"DockerHub with default token": {dockerCheck: &DockerCheck{Type: "hub", Image: "releaseargus/argus"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}

			// WHEN refreshDockerHubToken is called on it
			err := tc.dockerCheck.refreshDockerHubToken()

			// THEN the err is what we expect
			if tc.errRegex == "" {
				tc.errRegex = "^$"
			}
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestDockerCheckCheckValues(t *testing.T) {
	// GIVEN a DockerCheck
	jLog = utils.NewJLog("WARN", false)
	tests := map[string]struct {
		dockerCheck *DockerCheck
		errRegex    string
	}{
		"nil DockerCheck": {errRegex: "^$", dockerCheck: nil},
		"invalid Type": {errRegex: "^-type: .*<invalid>",
			dockerCheck: &DockerCheck{Type: "foo", Image: "release-argus/argus", Tag: "1.2.3"}},
		"invalid Type and no Image": {errRegex: "^-type: .*<invalid>.*-image: .*<required>.*",
			dockerCheck: &DockerCheck{Type: "foo", Tag: "1.2.3"}},
		"invalid Type and no Image or Tag": {errRegex: "^-type: .*<invalid>.*-image: .*<required>.*tag: .*<required>",
			dockerCheck: &DockerCheck{Type: "foo"}},
		"invalid Image": {errRegex: "image: .* <invalid>",
			dockerCheck: &DockerCheck{Type: "hub", Image: "	release-argus/argus", Tag: "1.2.3"}},
		"invalid Tag templating": {errRegex: "tag: .* <invalid>",
			dockerCheck: &DockerCheck{Type: "hub", Image: "release-argus/argus", Tag: "{{ version }"}},
		"valid Type with image and tag": {errRegex: "^$",
			dockerCheck: &DockerCheck{Type: "ghcr", Image: "release-argus/argus", Tag: "1.2.3"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called on it
			err := tc.dockerCheck.CheckValues("-")

			// THEN the err is what we expect
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}
