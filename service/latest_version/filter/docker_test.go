// Copyright [2023] [Argus]
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
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
)

func TestDockerCheck_GetTag(t *testing.T) {
	// GIVEN a DockerCheck and a version
	tests := map[string]struct {
		dockerCheck *DockerCheck
		version     string
		want        string
	}{
		"plain version": {
			dockerCheck: NewDockerCheck(
				"", "", "1.2.3", "", "", "", time.Now(), nil),
			version: "3.2.1",
			want:    "1.2.3"},
		"version template": {
			dockerCheck: NewDockerCheck(
				"", "", "{{ version }}.1", "", "", "", time.Now(), nil),
			version: "3.2",
			want:    "3.2.1"},
	}

	for name, tc := range tests {
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

func TestDockerCheck_getQueryToken(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		hadQueryToken  string
		wantQueryToken *string
		onlyIfEnvToken bool
		errRegex       string
		noToken        bool
	}{
		"no image does nothing": {
			dockerCheck: NewDockerCheck(
				"hub", "", "", "", "", "", time.Now(), nil),
		},
		"DockerHub invalid token": {
			errRegex:       "(.ncorrect username or password|.oo many failed login attempts|.annot log into an organization account|.ncorrect authentication credentials)",
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"argus",
				"",
				"invalid",
				"", time.Now(), nil),
		},
		"DockerHub valid token": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"",
				os.Getenv("DOCKER_USERNAME"),
				os.Getenv("DOCKER_TOKEN"),
				"", time.Now(), nil),
		},
		"GHCR invalid repo": {
			errRegex: "invalid control character",
			dockerCheck: NewDockerCheck(
				"ghcr",
				"	release-argus/argus",
				"", "", "", "", time.Now(), nil),
		},
		"GHCR non-existing repo": {
			errRegex: "invalid repository name",
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus-",
				"", "", "", "", time.Now(), nil),
		},
		"GHCR get default token": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"", "", "", "", time.Now(), nil),
		},
		"GHCR base64 access token": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"",
				"",
				base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN"))),
				"", time.Now(), nil),
		},
		"GHCR plaintext access token": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"",
				"",
				os.Getenv("GH_TOKEN"),
				"", time.Now(), nil),
		},
		"Quay get token": {
			dockerCheck: NewDockerCheck(
				"quay",
				"argus-io/argus",
				"",
				"",
				"foo",
				"", time.Now(), nil),
		},
		"Returns current queryToken if before validUntil": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"",
				"",
				os.Getenv("GITHUB_TOKEN"),
				"", time.Now().UTC().Add(time.Minute),
				nil),
			hadQueryToken: base64.StdEncoding.EncodeToString([]byte("foo")),
			wantQueryToken: stringPtr(
				base64.StdEncoding.EncodeToString([]byte("foo"))),
		},
		"Gets token if past validUntil": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"",
				"",
				os.Getenv("GITHUB_TOKEN"),
				"", time.Now().UTC().Add(-time.Minute),
				nil),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.onlyIfEnvToken && tc.dockerCheck.Token == "" {
				t.Skip("ENV VAR undefined")
			}
			tc.dockerCheck.queryToken = tc.hadQueryToken

			// WHEN getQueryToken is called on it
			queryToken, err := tc.dockerCheck.getQueryToken()

			// THEN the err is what we expect and a queryToken is retrieved when expected
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
				// didn't want a queryToken
				if tc.noToken {
					if queryToken != "" {
						t.Errorf("didn't expect a queryToken. Got %q",
							queryToken)
					}
					// didn't get a queryToken
				} else if queryToken == "" {
					t.Error("didn't get any queryToken")
				}
			}
			// AND the query token is what we expect
			if tc.wantQueryToken != nil &&
				*tc.wantQueryToken != queryToken {
				t.Errorf("want queryToken %q, not %q",
					*tc.wantQueryToken, queryToken)
			}
		})
	}
}

func TestDockerCheckDefaults_getQueryToken(t *testing.T) {
	// GIVEN a DockerCheckDefaults
	tests := map[string]struct {
		dockerCheckDefaults *DockerCheckDefaults
		wantToken           string
		wantValidUntil      time.Time

		queryTokenGHCR string
		validUntilGHCR time.Time
		queryTokenHub  string
		validUntilHub  time.Time
		queryTokenQuay string
		validUntilQuay time.Time

		defaultQueryTokenGHCR string
		defaultValidUntilGHCR time.Time
		defaultQueryTokenHub  string
		defaultValidUntilHub  time.Time
		defaultQueryTokenQuay string
		defaultValidUntilQuay time.Time
	}{
		"nil DockerCheckDefaults": {
			dockerCheckDefaults: nil,
		},
		"empty DockerCheckDefaults": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"", "", "", "", "", nil),
		},
		"GHCR from main": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"ghcr",
				"ghp_x",
				"", "",
				"",
				nil),
			queryTokenGHCR: "foo",
			validUntilGHCR: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantToken:      "foo",
			wantValidUntil: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		"GHCR from default": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"ghcr",
				"", "", "", "",
				NewDockerCheckDefaults(
					"",
					"ghp_y",
					"", "",
					"",
					nil)),
			defaultQueryTokenGHCR: "bar",
			defaultValidUntilGHCR: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			wantToken:             "bar",
			wantValidUntil:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		"Hub from main": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"hub",
				"",
				"hub_a", "",
				"",
				nil),
			queryTokenHub:  "foo",
			validUntilHub:  time.Date(2021, 1, 1, 0, 0, 6, 0, time.UTC),
			wantToken:      "foo",
			wantValidUntil: time.Date(2021, 1, 1, 0, 0, 6, 0, time.UTC),
		},
		"Hub from default": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"hub",
				"", "", "", "",
				NewDockerCheckDefaults(
					"",
					"",
					"hub_b", "",
					"",
					nil)),
			defaultQueryTokenHub: "bam",
			defaultValidUntilHub: time.Date(2022, 1, 1, 0, 3, 0, 0, time.UTC),
			wantToken:            "bam",
			wantValidUntil:       time.Date(2022, 1, 1, 0, 3, 0, 0, time.UTC),
		},
		"Quay from main": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"quay",
				"",
				"", "",
				"quay_1",
				nil),
			queryTokenQuay: "hocus",
			validUntilQuay: time.Date(2021, 1, 1, 2, 0, 6, 0, time.UTC),
			wantToken:      "hocus",
			wantValidUntil: time.Date(2021, 1, 1, 2, 0, 6, 0, time.UTC),
		},
		"Quay from default": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"quay",
				"", "", "", "",
				NewDockerCheckDefaults(
					"",
					"",
					"", "",
					"quay_2",
					nil)),
			defaultQueryTokenQuay: "pocus",
			defaultValidUntilQuay: time.Date(2022, 1, 4, 0, 3, 0, 0, time.UTC),
			wantToken:             "pocus",
			wantValidUntil:        time.Date(2022, 1, 4, 0, 3, 0, 0, time.UTC),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.dockerCheckDefaults != nil {
				if tc.dockerCheckDefaults.RegistryGHCR != nil {
					tc.dockerCheckDefaults.RegistryGHCR.queryToken = tc.queryTokenGHCR
					tc.dockerCheckDefaults.RegistryGHCR.validUntil = tc.validUntilGHCR
				}
				if tc.dockerCheckDefaults.RegistryHub != nil {
					tc.dockerCheckDefaults.RegistryHub.queryToken = tc.queryTokenHub
					tc.dockerCheckDefaults.RegistryHub.validUntil = tc.validUntilHub
				}
				if tc.dockerCheckDefaults.RegistryQuay != nil {
					tc.dockerCheckDefaults.RegistryQuay.queryToken = tc.queryTokenQuay
					tc.dockerCheckDefaults.RegistryQuay.validUntil = tc.validUntilQuay
				}

				if tc.dockerCheckDefaults.defaults != nil {
					if tc.dockerCheckDefaults.defaults.RegistryGHCR != nil {
						tc.dockerCheckDefaults.defaults.RegistryGHCR.queryToken = tc.defaultQueryTokenGHCR
						tc.dockerCheckDefaults.defaults.RegistryGHCR.validUntil = tc.defaultValidUntilGHCR
					}
					if tc.dockerCheckDefaults.defaults.RegistryHub != nil {
						tc.dockerCheckDefaults.defaults.RegistryHub.queryToken = tc.defaultQueryTokenHub
						tc.dockerCheckDefaults.defaults.RegistryHub.validUntil = tc.defaultValidUntilHub
					}
					if tc.dockerCheckDefaults.defaults.RegistryQuay != nil {
						tc.dockerCheckDefaults.defaults.RegistryQuay.queryToken = tc.defaultQueryTokenQuay
						tc.dockerCheckDefaults.defaults.RegistryQuay.validUntil = tc.defaultValidUntilQuay
					}
				}
			}

			// WHEN getQueryToken is called on it
			queryToken, validUntil := tc.dockerCheckDefaults.getQueryToken(
				tc.dockerCheckDefaults.GetType())

			// THEN the query token is what we expect
			if tc.wantToken != queryToken {
				t.Errorf("want queryToken %q, not %q",
					tc.wantToken, queryToken)
			}
			// AND the validUntil is what we expect
			if tc.wantValidUntil != validUntil {
				t.Errorf("want validUntil %q, not %q",
					tc.wantValidUntil, validUntil)
			}
		})
	}
}

func TestDockerCheck_SetQueryToken(t *testing.T) {
	// GIVEN a DockerCheck and a value to set the queryToken/validUntil to
	queryToken := "something"
	validUntil := time.Date(2000, 1, 1, 3, 5, 5, 0, time.UTC)
	tests := map[string]struct {
		dockerCheck      *DockerCheck
		defaultTokenGHCR *string
		setForToken      string
		changeDefaults   bool
	}{
		"nil": {
			dockerCheck: nil,
		},
		"queryToken set in main when Token doesn't match": {
			dockerCheck: NewDockerCheck(
				"hub",
				"", "",
				"", "foo",
				"", time.Time{},
				NewDockerCheckDefaults(
					"",
					"tokenGHCR",
					"tokenHub", "",
					"tokenQuay",
					nil)),
			setForToken: "shazam",
		},
		"token only set in main even if default has the same": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"", "",
				"", "tokenGHCR",
				"", time.Time{},
				NewDockerCheckDefaults(
					"",
					"otherTokenGHCR",
					"tokenHub", "",
					"tokenQuay",
					nil)),
			setForToken:    "tokenGHCR",
			changeDefaults: false,
		},
		"token in defaults - ghcr": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"", "",
				"", "",
				"", time.Time{},
				NewDockerCheckDefaults(
					"",
					"tokenGHCR",
					"tokenHub", "",
					"tokenQuay",
					nil)),
			setForToken:    "tokenGHCR",
			changeDefaults: true,
		},
		"token in defaults - hub": {
			dockerCheck: NewDockerCheck(
				"hub",
				"", "",
				"", "",
				"", time.Time{},
				NewDockerCheckDefaults(
					"",
					"tokenGHCR",
					"tokenHub", "",
					"tokenQuay",
					nil)),
			setForToken:    "tokenHub",
			changeDefaults: true,
		},
		"token in defaults - quay": {
			dockerCheck: NewDockerCheck(
				"quay",
				"", "",
				"", "",
				"", time.Time{},
				NewDockerCheckDefaults(
					"",
					"tokenGHCR",
					"tokenHub", "",
					"tokenQuay",
					nil)),
			setForToken:    "tokenQuay",
			changeDefaults: true,
		},
		"GHCR, NOOP token not given to defaults as repo/image specific": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"", "",
				"", "",
				"", time.Time{},
				NewDockerCheckDefaults(
					"",
					"initGHCR",
					"tokenHub", "",
					"tokenQuay",
					nil)),
			defaultTokenGHCR: stringPtr(""),
			setForToken:      "tokenGHCR",
			changeDefaults:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hadNil := tc.dockerCheck == nil
			hadNilDefaults := tc.dockerCheck == nil || tc.dockerCheck.Defaults == nil
			setForType := tc.dockerCheck.GetType()
			if tc.defaultTokenGHCR != nil {
				tc.dockerCheck.Defaults.RegistryGHCR.Token = *tc.defaultTokenGHCR
			}

			// WHEN SetQueryToken is called on it
			tc.dockerCheck.SetQueryToken(
				&tc.setForToken,
				&queryToken, &validUntil)

			// THEN the query token is what we expect
			if hadNil {
				if tc.dockerCheck != nil {
					t.Errorf("want nil dockerCheck, not %v", tc.dockerCheck)
				}
				return
			}
			if tc.dockerCheck.queryToken != queryToken {
				t.Errorf("want main queryToken to be %q, not %q",
					queryToken, tc.dockerCheck.queryToken)
			}
			if tc.dockerCheck.validUntil != validUntil {
				t.Errorf("want main validUntil to be %q, not %q",
					validUntil, tc.dockerCheck.validUntil)
			}
			if hadNilDefaults {
				if tc.dockerCheck.Defaults != nil {
					t.Errorf("want nil dockerCheck.Defaults, not %v", tc.dockerCheck.Defaults)
				}
				return
			}
			if tc.changeDefaults &&
				setForType == "ghcr" {
				if tc.dockerCheck.Defaults.RegistryGHCR.queryToken != queryToken {
					t.Errorf("want main queryToken for type 'ghcr' to be %q, not %q",
						queryToken, tc.dockerCheck.Defaults.RegistryGHCR.queryToken)
				}
				if tc.dockerCheck.Defaults.RegistryGHCR.validUntil != validUntil {
					t.Errorf("want main validUntil for type 'ghcr' to be %q, not %q",
						validUntil, tc.dockerCheck.Defaults.RegistryGHCR.validUntil)
				}
			} else {
				if tc.dockerCheck.Defaults.RegistryGHCR.queryToken == queryToken {
					t.Errorf("queryToken shouldn't have been set to %q in the defaults for type 'ghcr'",
						tc.dockerCheck.Defaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheck.Defaults.RegistryGHCR.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the defaults for type 'ghcr'",
						tc.dockerCheck.Defaults.RegistryQuay.validUntil)
				}
			}
			if tc.changeDefaults &&
				setForType == "hub" {
				if tc.dockerCheck.Defaults.RegistryHub.queryToken != queryToken {
					t.Errorf("want defaults queryToken for type 'hub' to be %q, not %q",
						queryToken, tc.dockerCheck.Defaults.RegistryHub.queryToken)
				}
				if tc.dockerCheck.Defaults.RegistryHub.validUntil != validUntil {
					t.Errorf("want defaults validUntil for type 'hub' to be %q, not %q",
						validUntil, tc.dockerCheck.Defaults.RegistryHub.validUntil)
				}
			} else {
				if tc.dockerCheck.Defaults.RegistryHub.queryToken == queryToken {
					t.Errorf("queryToken shouldn't have been set to %q in the defaults for type 'hub'",
						tc.dockerCheck.Defaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheck.Defaults.RegistryHub.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the defaults for type 'hub'",
						tc.dockerCheck.Defaults.RegistryQuay.validUntil)
				}
			}
			if tc.changeDefaults &&
				setForType == "quay" {
				if tc.dockerCheck.Defaults.RegistryQuay.queryToken != queryToken {
					t.Errorf("want defaults queryToken for type 'quay' to be %q, not %q",
						queryToken, tc.dockerCheck.Defaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheck.Defaults.RegistryQuay.validUntil != validUntil {
					t.Errorf("want defaults validUntil for type 'quay' to be %q, not %q",
						validUntil, tc.dockerCheck.Defaults.RegistryQuay.validUntil)
				}
			} else {
				if tc.dockerCheck.Defaults.RegistryQuay.queryToken == queryToken {
					t.Errorf("queryToken shouldn't have been set to %q in the defaults for type 'quay'",
						tc.dockerCheck.Defaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheck.Defaults.RegistryQuay.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the defaults for type 'quay'",
						tc.dockerCheck.Defaults.RegistryQuay.validUntil)
				}
			}
		})
	}
}

func TestDockerCheckDefaults_setQueryToken(t *testing.T) {
	// GIVEN a DockerCheckDefaults and a value to set the queryToken/validUntil to
	queryToken := "something"
	validUntil := time.Date(2000, 1, 1, 3, 0, 5, 0, time.UTC)
	tests := map[string]struct {
		dockerCheckDefaults *DockerCheckDefaults
		ghcrToken           *string
		ghcrDefaultToken    *string
		setForType          string
		setForToken         string
		changeDefaults      bool
		changeMain          bool
	}{
		"nil": {
			dockerCheckDefaults: nil,
			setForType:          "ghcr",
			setForToken:         "tokenForGHCR",
		},
		"unknown type": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				nil),
			setForType:  "argus",
			setForToken: "tokenForGHCR",
		},
		"unknown token for type": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				nil),
			setForType:  "hub",
			setForToken: "tokenForGHCR",
		},
		"set in main - ghcr": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				nil),
			setForType:  "ghcr",
			setForToken: "tokenForGHCR",
			changeMain:  true,
		},
		"set in main - hub": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				nil),
			setForType:  "hub",
			setForToken: "tokenForHub",
			changeMain:  true,
		},
		"set in main - quay": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				nil),
			setForType:  "quay",
			setForToken: "tokenForQuay",
			changeMain:  true,
		},
		"set in defaults - hub": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				NewDockerCheckDefaults(
					"",
					"defaultTokenForGHCR",
					"defaultTokenForHub", "",
					"defaultTokenForQuay",
					nil)),
			setForType:     "hub",
			setForToken:    "defaultTokenForHub",
			changeDefaults: true,
		},
		"set in defaults - ghcr": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				NewDockerCheckDefaults(
					"",
					"defaultTokenForGHCR",
					"defaultTokenForHub", "",
					"defaultTokenForQuay",
					nil)),
			setForType:     "ghcr",
			setForToken:    "defaultTokenForGHCR",
			changeDefaults: true,
		},
		"ghcr - NOOP not saved in defaults": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"ghcr",
				"initGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				NewDockerCheckDefaults(
					"",
					"initGHCR",
					"defaultTokenForHub", "",
					"defaultTokenForQuay",
					nil)),
			ghcrToken:        stringPtr(""),
			ghcrDefaultToken: stringPtr(""),
			setForType:       "ghcr",
			setForToken:      "",
			changeDefaults:   false,
			changeMain:       false,
		},
		"set in defaults - quay": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"tokenForGHCR",
				"tokenForHub", "",
				"tokenForQuay",
				NewDockerCheckDefaults(
					"",
					"defaultTokenForGHCR",
					"defaultTokenForHub", "",
					"defaultTokenForQuay",
					nil)),
			setForType:     "quay",
			setForToken:    "defaultTokenForQuay",
			changeDefaults: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hadNil := tc.dockerCheckDefaults == nil
			hadNilDefaults := tc.dockerCheckDefaults == nil || tc.dockerCheckDefaults.defaults == nil
			if tc.ghcrToken != nil {
				tc.dockerCheckDefaults.RegistryGHCR.Token = *tc.ghcrToken
			}
			if tc.ghcrDefaultToken != nil {
				tc.dockerCheckDefaults.defaults.RegistryGHCR.Token = *tc.ghcrDefaultToken
			}

			// WHEN setQueryToken is called on it
			tc.dockerCheckDefaults.setQueryToken(
				&tc.setForType, &tc.setForToken,
				&queryToken, &validUntil)

			// THEN the query token isn't set if the DockerCheckDefaults was nil
			if hadNil {
				if tc.dockerCheckDefaults != nil {
					t.Error("didn't expect DockerCheckDefaults to be initialised")
				}
				return
			}
			// THEN the query token is set in the correct main if the Token matched
			// GHCR
			if tc.changeMain &&
				tc.setForType == "ghcr" {
				if tc.dockerCheckDefaults.RegistryGHCR.queryToken != queryToken {
					t.Errorf("want main queryToken for type 'ghcr' to be %q, not %q",
						queryToken, tc.dockerCheckDefaults.RegistryGHCR.queryToken)
				}
				if tc.dockerCheckDefaults.RegistryGHCR.validUntil != validUntil {
					t.Errorf("want main validUntil for type 'ghcr' to be %q, not %q",
						validUntil, tc.dockerCheckDefaults.RegistryGHCR.validUntil)
				}
			} else {
				if tc.dockerCheckDefaults.RegistryGHCR.queryToken == queryToken {
					t.Errorf("token shouldn't have been set to %q in the main for type 'ghcr'",
						tc.dockerCheckDefaults.RegistryGHCR.queryToken)
				}
				if tc.dockerCheckDefaults.RegistryGHCR.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the main for type 'ghcr'",
						tc.dockerCheckDefaults.RegistryGHCR.validUntil)
				}
			}
			// Hub
			if tc.changeMain &&
				tc.setForType == "hub" {
				if tc.dockerCheckDefaults.RegistryHub.queryToken != queryToken {
					t.Errorf("want main queryToken for type 'hub' to be %q, not %q",
						queryToken, tc.dockerCheckDefaults.RegistryHub.queryToken)
				}
				if tc.dockerCheckDefaults.RegistryHub.validUntil != validUntil {
					t.Errorf("want main validUntil for type 'hub' to be %q, not %q",
						validUntil, tc.dockerCheckDefaults.RegistryHub.validUntil)
				}
			} else {
				if tc.dockerCheckDefaults.RegistryHub.queryToken == queryToken {
					t.Errorf("token shouldn't have been set to %q in the main for type 'hub'",
						tc.dockerCheckDefaults.RegistryHub.queryToken)
				}
				if tc.dockerCheckDefaults.RegistryHub.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the main for type 'hub'",
						tc.dockerCheckDefaults.RegistryHub.validUntil)
				}
			}
			// Quay
			if tc.changeMain &&
				tc.setForType == "quay" {
				if tc.dockerCheckDefaults.RegistryQuay.queryToken != queryToken {
					t.Errorf("want ain queryToken for type 'quay' to be %q, not %q",
						queryToken, tc.dockerCheckDefaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheckDefaults.RegistryQuay.validUntil != validUntil {
					t.Errorf("want main validUntil for type 'quay' to be %q, not %q",
						validUntil, tc.dockerCheckDefaults.RegistryQuay.validUntil)
				}
			} else {
				if tc.dockerCheckDefaults.RegistryQuay.queryToken == queryToken {
					t.Errorf("token shouldn't have been set to %q in the main for type 'quay'",
						tc.dockerCheckDefaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheckDefaults.RegistryQuay.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the main for type 'quay'",
						tc.dockerCheckDefaults.RegistryQuay.validUntil)
				}
			}

			// AND the defaults aren't initialised if they were nil
			if hadNilDefaults {
				if tc.dockerCheckDefaults.defaults != nil {
					t.Error("didn't expect defaults to be initialised")
				}
				return
			}

			// AND it's set in the defaults if the Token didn't match any in the main and does in the defaults
			// GHCR
			if tc.changeDefaults &&
				tc.setForType == "ghcr" {
				if tc.dockerCheckDefaults.defaults.RegistryGHCR.queryToken != queryToken {
					t.Errorf("want default queryToken for type 'ghcr' to be %q, not %q",
						queryToken, tc.dockerCheckDefaults.defaults.RegistryGHCR.queryToken)
				}
				if tc.dockerCheckDefaults.defaults.RegistryGHCR.validUntil != validUntil {
					t.Errorf("want default validUntil for type 'ghcr' to be %q, not %q",
						validUntil, tc.dockerCheckDefaults.defaults.RegistryGHCR.validUntil)
				}
			} else {
				if tc.dockerCheckDefaults.defaults.RegistryGHCR.queryToken == queryToken {
					t.Errorf("token shouldn't have been set to %q in the defaults for type 'ghcr'",
						tc.dockerCheckDefaults.defaults.RegistryGHCR.queryToken)
				}
				if tc.dockerCheckDefaults.defaults.RegistryGHCR.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the defaults for type 'hub'",
						tc.dockerCheckDefaults.defaults.RegistryGHCR.validUntil)
				}
			}
			// Hub
			if tc.changeDefaults &&
				tc.setForType == "hub" {
				if tc.dockerCheckDefaults.defaults.RegistryHub.queryToken != queryToken {
					t.Errorf("want default queryToken for type 'hub' to be %q, not %q",
						queryToken, tc.dockerCheckDefaults.defaults.RegistryHub.queryToken)
				}
				if tc.dockerCheckDefaults.defaults.RegistryHub.validUntil != validUntil {
					t.Errorf("want default validUntil for type 'hub' to be %q, not %q",
						validUntil, tc.dockerCheckDefaults.defaults.RegistryHub.validUntil)
				}
			} else {
				if tc.dockerCheckDefaults.defaults.RegistryHub.queryToken == queryToken {
					t.Errorf("token shouldn't have been set to %q in the defaults for type 'hub'",
						tc.dockerCheckDefaults.defaults.RegistryHub.queryToken)
				}
				if tc.dockerCheckDefaults.defaults.RegistryHub.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the defaults for type 'hub'",
						tc.dockerCheckDefaults.defaults.RegistryHub.validUntil)
				}
			}
			// Quay
			if tc.changeDefaults &&
				tc.setForType == "quay" {
				if tc.dockerCheckDefaults.defaults.RegistryQuay.queryToken != queryToken {
					t.Errorf("want default queryToken for type 'quay' to be %q, not %q",
						queryToken, tc.dockerCheckDefaults.defaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheckDefaults.defaults.RegistryQuay.validUntil != validUntil {
					t.Errorf("want default validUntil for type 'quay' to be %q, not %q",
						validUntil, tc.dockerCheckDefaults.defaults.RegistryQuay.validUntil)
				}
			} else {
				if tc.dockerCheckDefaults.defaults.RegistryQuay.queryToken == queryToken {
					t.Errorf("token shouldn't have been set to %q in the defaults for type 'quay'",
						tc.dockerCheckDefaults.defaults.RegistryQuay.queryToken)
				}
				if tc.dockerCheckDefaults.defaults.RegistryQuay.validUntil == validUntil {
					t.Errorf("validUntil shouldn't have been set to %q in the defaults for type 'quay'",
						tc.dockerCheckDefaults.defaults.RegistryQuay.validUntil)
				}
			}
		})
	}
}

func TestDockerCheck_getValidToken(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck *DockerCheck
		want        string

		startQueryToken string
		startValidUntil time.Time

		defaultQueryTokenGHCR string
		defaultValidUntilGHCR time.Time
		defaultQueryTokenHub  string
		defaultValidUntilHub  time.Time
		defaultQueryTokenQuay string
		defaultValidUntilQuay time.Time
	}{
		"nil": {
			dockerCheck: nil,
		},
		"main token within valid range": {
			dockerCheck: NewDockerCheck(
				"hub",
				"", "", "", "", "", time.Time{}, nil),
			want:            "foo",
			startQueryToken: "foo",
			startValidUntil: time.Now().Add(time.Minute),
		},
		"main token out of valid range": {
			dockerCheck: NewDockerCheck(
				"hub",
				"", "", "", "", "", time.Time{}, nil),
			want:            "",
			startQueryToken: "",
			startValidUntil: time.Now().Add(-time.Minute),
		},
		"main token out of valid range. get default if that's in range": {
			dockerCheck: NewDockerCheck(
				"hub",
				"", "", "", "", "", time.Time{},
				NewDockerCheckDefaults(
					"", "", "j", "", "", nil)),
			want:                 "bar",
			startQueryToken:      "foo",
			defaultQueryTokenHub: "bar",
			defaultValidUntilHub: time.Now().Add(time.Hour),
			startValidUntil:      time.Now().Add(-time.Minute),
		},
		"main and default tokens out of valid range": {
			dockerCheck: NewDockerCheck(
				"hub",
				"", "", "", "", "", time.Time{},
				NewDockerCheckDefaults(
					"", "", "j", "", "", nil)),
			want:                 "",
			startQueryToken:      "foo",
			defaultQueryTokenHub: "bar",
			defaultValidUntilHub: time.Now().Add(-time.Hour),
			startValidUntil:      time.Now().Add(-time.Minute),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.dockerCheck != nil {
				tc.dockerCheck.queryToken = tc.startQueryToken
				tc.dockerCheck.validUntil = tc.startValidUntil

				if tc.dockerCheck.Defaults != nil {
					if tc.dockerCheck.Defaults.RegistryGHCR != nil {
						tc.dockerCheck.Defaults.RegistryGHCR.queryToken = tc.defaultQueryTokenGHCR
						tc.dockerCheck.Defaults.RegistryGHCR.validUntil = tc.defaultValidUntilGHCR
					}
					if tc.dockerCheck.Defaults.RegistryHub != nil {
						tc.dockerCheck.Defaults.RegistryHub.queryToken = tc.defaultQueryTokenHub
						tc.dockerCheck.Defaults.RegistryHub.validUntil = tc.defaultValidUntilHub
					}
					if tc.dockerCheck.Defaults.RegistryQuay != nil {
						tc.dockerCheck.Defaults.RegistryQuay.queryToken = tc.defaultQueryTokenQuay
						tc.dockerCheck.Defaults.RegistryQuay.validUntil = tc.defaultValidUntilQuay
					}
				}
			}

			// WHEN getValidToken is called on it
			got := tc.dockerCheck.getValidToken(tc.dockerCheck.GetType())

			// THEN the token is what we expect
			if tc.want != got {
				t.Errorf("want %q, not %q",
					tc.want, got)
			}
		})
	}
}

func TestDockerCheck_CopyQueryToken(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck *DockerCheck

		startQueryToken string
		startValidUntil time.Time

		wantQueryToken string
		wantValidUntil time.Time
	}{
		"nil": {
			dockerCheck: nil},
		"have token + validUntil": {
			dockerCheck:     &DockerCheck{},
			startQueryToken: "foo",
			startValidUntil: time.Date(2020, 6, 4, 0, 3, 0, 0, time.UTC),
			wantQueryToken:  "foo",
			wantValidUntil:  time.Date(2020, 6, 4, 0, 3, 0, 0, time.UTC)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.dockerCheck != nil {
				tc.dockerCheck.queryToken = tc.startQueryToken
				tc.dockerCheck.validUntil = tc.startValidUntil
			}

			// WHEN CopyQueryToken is called on it
			gotQueryToken, gotValidUntil := tc.dockerCheck.CopyQueryToken()

			// THEN the token is what we expect
			if tc.wantQueryToken != gotQueryToken {
				t.Errorf("want %q, not %q",
					tc.wantQueryToken, gotQueryToken)
			}
			if tc.wantValidUntil != gotValidUntil {
				t.Errorf("want %q, not %q",
					tc.wantValidUntil, gotValidUntil)
			}
		})
	}
}

func TestDockerCheck_getUsername(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck *DockerCheck
		want        string
	}{
		"nil": {
			dockerCheck: nil,
		},
		"empty": {
			dockerCheck: &DockerCheck{},
		},
		"username from main": {
			dockerCheck: NewDockerCheck(
				"",
				"", "",
				"aUser", "",
				"", time.Time{},
				nil),
			want: "aUser",
		},
		"username from default": {
			dockerCheck: NewDockerCheck(
				"",
				"", "",
				"", "",
				"", time.Time{},
				NewDockerCheckDefaults(
					"",
					"",
					"", "anotherUser",
					"",
					nil)),
			want: "anotherUser",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN getUsername is called on it
			got := tc.dockerCheck.getUsername()

			// THEN the username is what we expect
			if tc.want != got {
				t.Errorf("want %q, not %q",
					tc.want, got)
			}
		})
	}
}

func TestDockerCheckDefaults_getUsername(t *testing.T) {
	// GIVEN a DockerCheckDefaults
	tests := map[string]struct {
		dockerCheckDefaults *DockerCheckDefaults
		want                string
	}{
		"nil": {
			dockerCheckDefaults: nil,
		},
		"empty": {
			dockerCheckDefaults: &DockerCheckDefaults{},
		},
		"username in base": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"",
				"", "daUser",
				"",
				nil),
			want: "daUser",
		},
		"username in defaults": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"",
				"", "",
				"",
				NewDockerCheckDefaults(
					"",
					"",
					"", "daOtherUser",
					"",
					nil)),
			want: "daOtherUser",
		},
		"base takes priority": {
			dockerCheckDefaults: NewDockerCheckDefaults(
				"",
				"",
				"", "daUser",
				"",
				NewDockerCheckDefaults(
					"",
					"",
					"", "daOtherUser",
					"",
					nil)),
			want: "daUser",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN getUsername is called on it
			got := tc.dockerCheckDefaults.getUsername()

			// THEN the username is what we expect
			if tc.want != got {
				t.Errorf("want %q, not %q",
					tc.want, got)
			}
		})
	}
}

func TestDockerCheckDefaults_SetDefaults(t *testing.T) {
	// GIVEN a DockerCheck and a set of defaults
	definedDefaults := &DockerCheckDefaults{}
	tests := map[string]struct {
		base     *DockerCheckDefaults
		defaults *DockerCheckDefaults
		want     *DockerCheckDefaults
	}{
		"nil base and defaults does nothing": {
			base: nil},
		"nil base with defaults does nothing": {
			base:     nil,
			defaults: definedDefaults},
		"base with nil defaults does nothing": {
			base: &DockerCheckDefaults{}},
		"base with defaults does set": {
			base:     &DockerCheckDefaults{},
			defaults: definedDefaults,
			want:     definedDefaults},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN SetDefaults is called on it
			tc.base.SetDefaults(tc.defaults)

			// THEN the defaults are what we expect
			if tc.want == nil {
				if tc.base != nil && tc.base.defaults != nil {
					t.Errorf("want nil, not %v",
						tc.base.defaults)
				}
				return
			}
			if tc.want != tc.base.defaults {
				t.Errorf("want %v, not %v",
					tc.want, tc.base.defaults)
			}
		})
	}
}

func TestDockerCheck_CheckToken(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"nil DockerCheck does nothing": {
			dockerCheck: nil,
		},
		"DockerHub get default token": {
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"", "", "", "", time.Now(), nil),
		},
		"DockerHub no token for username": {
			errRegex: "token: <required>",
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"",
				"xxxx",
				"", "", time.Now(), nil),
		},
		"DockerHub no username for token": {
			errRegex: "username: <required>",
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"",
				"",
				"xxxx",
				"", time.Now(), nil),
		},
		"DockerHub valid token": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"",
				os.Getenv("DOCKER_USERNAME"),
				os.Getenv("DOCKER_TOKEN"),
				"", time.Now(), nil),
		},
		"GHCR get default token": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"", "", "", "", time.Now(), nil),
		},
		"GHCR base64 access token": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"", "",
				base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN"))),
				"", time.Now(), nil),
		},
		"GHCR plaintext access token": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"", "",
				os.Getenv("GITHUB_TOKEN"),
				"", time.Now(), nil),
		},
		"Quay have token": {
			dockerCheck: NewDockerCheck(
				"quay",
				"argus-io/argus",
				"", "",
				"foo",
				"", time.Now(), nil),
		},
	}

	for name, tc := range tests {
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
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"nil DockerCheck does nothing": {
			dockerCheck: nil,
		},
		"DockerHub with no token, valid tag": {
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"{{ version }}",
				"", "", "", time.Now(), nil),
		},
		"DockerHub with no token, invalid tag": {
			errRegex: "tag .+ not found",
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"0.1.9",
				"", "", "", time.Now(), nil),
		},
		"DockerHub valid token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"{{ version }}",
				os.Getenv("DOCKER_USERNAME"),
				os.Getenv("DOCKER_TOKEN"),
				"", time.Now(), nil),
		},
		"DockerHub valid token, invalid tag": {
			errRegex:       "tag .+ not found",
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"{{ version }}-beta",
				os.Getenv("DOCKER_USERNAME"),
				os.Getenv("DOCKER_TOKEN"),
				"", time.Now(), nil),
		},
		"GHCR with default token, valid tag": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"{{ version }}",
				"", "", "", time.Now(), nil),
		},
		"GHCR with default token, invalid tag": {
			errRegex: "manifest unknown",
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"{{ version }}-beta",
				"", "", "", time.Now(), nil),
		},
		"GHCR base64 access token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"{{ version }}",
				"",
				base64.StdEncoding.EncodeToString([]byte(os.Getenv("GITHUB_TOKEN"))),
				"", time.Now(), nil),
		},
		"GHCR plaintext access token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"{{ version }}",
				"",
				os.Getenv("GH_TOKEN"),
				"", time.Now(), nil),
		},
		"Quay with no token, valid tag": {
			dockerCheck: NewDockerCheck(
				"quay",
				"argus-io/argus",
				"{{ version }}",
				"", "", "", time.Now(), nil),
		},
		"Quay with valid token, valid tag": {
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"quay",
				"argus-io/argus",
				"{{ version }}",
				"",
				os.Getenv("QUAY_TOKEN"),
				"", time.Now(), nil),
		},
		"Quay with valid token, invalid tag": {
			errRegex:       "tag not found",
			onlyIfEnvToken: true,
			dockerCheck: NewDockerCheck(
				"quay",
				"argus-io/argus",
				"{{ version }}-beta",
				"",
				os.Getenv("QUAY_TOKEN"),
				"", time.Now(), nil),
		},
	}

	for name, tc := range tests {
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
	tests := map[string]struct {
		dockerCheck    *DockerCheck
		onlyIfEnvToken bool
		errRegex       string
	}{
		"valid token": {
			onlyIfEnvToken: true,
			errRegex:       "^$",
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"",
				os.Getenv("DOCKER_USERNAME"),
				os.Getenv("DOCKER_TOKEN"),
				"", time.Now(), nil),
		},
		"invalid token": {
			onlyIfEnvToken: true,
			errRegex:       "(incorrect username or password|too many failed login attempts|Cannot log into an organization account)",
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"",
				"argus",
				"dkcr_pat_invalid",
				"", time.Now(), nil),
		},
		"no token": {
			dockerCheck: NewDockerCheck(
				"hub",
				"releaseargus/argus",
				"", "", "", "", time.Now(), nil),
		},
	}

	for name, tc := range tests {
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

func TestDockerCheckDefaults_CheckValues(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck *DockerCheckDefaults
		errRegex    string
	}{
		"nil DockerCheck": {
			errRegex:    "^$",
			dockerCheck: nil,
		},
		"valid Type - ghcr": {
			errRegex: "^$",
			dockerCheck: NewDockerCheckDefaults(
				"ghcr",
				"", "", "", "", nil),
		},
		"valid Type - hub": {
			errRegex: "^$",
			dockerCheck: NewDockerCheckDefaults(
				"hub",
				"", "", "", "", nil),
		},
		"valid Type - quay": {
			errRegex: "^$",
			dockerCheck: NewDockerCheckDefaults(
				"quay",
				"", "", "", "", nil),
		},
		"invalid Type": {
			errRegex: "^-type: .*<invalid>",
			dockerCheck: NewDockerCheckDefaults(
				"foo",
				"", "", "", "", nil),
		},
	}

	for name, tc := range tests {
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
		})
	}
}

func TestDockerCheck_CheckValues(t *testing.T) {
	// GIVEN a DockerCheck
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
			dockerCheck: NewDockerCheck(
				"foo",
				"release-argus/argus",
				"1.2.3",
				"", "", "", time.Now(), nil),
		},
		"invalid Type and no Image": {
			errRegex: "^-type: .*<invalid>.*-image: .*<required>.*",
			dockerCheck: NewDockerCheck(
				"foo",
				"",
				"1.2.3",
				"", "", "", time.Now(), nil),
		},
		"invalid Type and no Image or Tag": {
			errRegex: "^-type: .*<invalid>.*-image: .*<required>.*tag: .*<required>",
			dockerCheck: NewDockerCheck(
				"foo",
				"", "", "", "", "", time.Now(), nil),
		},
		"official docker hub image": {
			errRegex: "^$",
			dockerCheck: NewDockerCheck(
				"hub",
				"release-argus/argus",
				"1.2.3",
				"", "", "", time.Now(), nil),
		},
		"type from defaults": {
			errRegex: "^$",
			dockerCheck: NewDockerCheck(
				"",
				"release-argus/argus",
				"1.2.3",
				"", "", "", time.Now(),
				&DockerCheckDefaults{Type: "hub"}),
		},
		"invalid type from defaults": { // defaults are validated separately
			errRegex: "^$",
			dockerCheck: NewDockerCheck(
				"",
				"release-argus/argus",
				"1.2.3",
				"", "", "", time.Now(),
				&DockerCheckDefaults{Type: "foo"}),
		},
		"no type from defaults": { // defaults are validated separately
			errRegex: "^-type: .*<required>",
			dockerCheck: NewDockerCheck(
				"",
				"release-argus/argus",
				"1.2.3",
				"", "", "", time.Now(),
				&DockerCheckDefaults{Type: ""}),
		},
		"image with period in name": {
			errRegex: "^$",
			dockerCheck: &DockerCheck{
				Type:  "ghcr",
				Image: "test/image.io",
				Tag:   "1.2.3"},
		},
		"docker hub type with username but no password": {
			errRegex:  "^-token: <required>",
			wantImage: "library/ubuntu",
			dockerCheck: NewDockerCheck(
				"hub",
				"ubuntu",
				"1.2.3",
				"test",
				"", "", time.Now(), nil),
		},
		"invalid Image": {
			errRegex: "image: .* <invalid>",
			dockerCheck: NewDockerCheck(
				"hub",
				"	release-argus/argus",
				"1.2.3",
				"", "", "", time.Now(), nil),
		},
		"invalid Tag templating": {
			errRegex: "tag: .* <invalid>",
			dockerCheck: NewDockerCheck(
				"hub",
				"release-argus/argus",
				"{{ version }",
				"", "", "", time.Now(), nil),
		},
		"valid Type with image and tag": {
			errRegex: "^$",
			dockerCheck: NewDockerCheck(
				"ghcr",
				"release-argus/argus",
				"1.2.3",
				"", "", "", time.Now(), nil),
		},
	}

	for name, tc := range tests {
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

func TestDockerCheckDefaults_GetType(t *testing.T) {
	// GIVEN a DockerCheckDefaults
	tests := map[string]struct {
		defaults *DockerCheckDefaults
		want     string
	}{
		"nil DockerCheckDefaults": {
			defaults: nil,
			want:     "",
		},
		"empty DockerCheckDefaults": {
			defaults: NewDockerCheckDefaults(
				"",
				"", "", "", "", nil),
			want: "",
		},
		"Type from base": {
			defaults: NewDockerCheckDefaults(
				"ghcr",
				"", "", "", "", nil),
			want: "ghcr",
		},
		"Type from defaults": {
			defaults: NewDockerCheckDefaults(
				"",
				"", "", "", "",
				NewDockerCheckDefaults(
					"hub",
					"", "", "", "", nil)),
			want: "hub",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetType is called on it
			got := tc.defaults.GetType()

			// THEN the result is what we expect
			if got != tc.want {
				t.Fatalf("expected %q, got: %q", tc.want, got)
			}
		})
	}
}

func TestDockerCheck_GetType(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck *DockerCheck
		want        string
	}{
		"nil DockerCheck": {
			dockerCheck: nil,
			want:        "",
		},
		"empty DockerCheck": {
			dockerCheck: NewDockerCheck(
				"", "", "", "", "", "", time.Now(), nil),
			want: "",
		},
		"Type from base": {
			dockerCheck: NewDockerCheck(
				"ghcr", "", "", "", "", "", time.Now(), nil),
			want: "ghcr",
		},
		"Type from defaults": {
			dockerCheck: NewDockerCheck(
				"", "", "", "", "", "", time.Now(),
				&DockerCheckDefaults{Type: "hub"}),
			want: "hub",
		},
		"Type from defaults with no type": {
			dockerCheck: NewDockerCheck(
				"", "", "", "", "", "", time.Now(),
				&DockerCheckDefaults{Type: ""}),
			want: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetType is called on it
			got := tc.dockerCheck.GetType()

			// THEN the result is what we expect
			if got != tc.want {
				t.Fatalf("expected %q, got: %q", tc.want, got)
			}
		})
	}
}

func TestDockerCheckDefaults_getToken(t *testing.T) {
	// GIVEN a DockerCheckDefaults
	tests := map[string]struct {
		defaults *DockerCheckDefaults
		seekType string
		want     string
	}{
		"nil DockerCheckDefaults": {
			defaults: nil,
			want:     "",
		},
		"empty DockerCheckDefaults": {
			defaults: NewDockerCheckDefaults(
				"", "", "", "", "", nil),
			want: "",
		},
		"TokenRaw from base - ghcr": {
			defaults: NewDockerCheckDefaults(
				"",
				"ghcrToken",
				"hubToken", "",
				"quayToken",
				nil),
			seekType: "ghcr",
			want:     "ghcrToken",
		},
		"TokenRaw from base - hub": {
			defaults: NewDockerCheckDefaults(
				"",
				"ghcrToken",
				"hubToken", "",
				"quayToken",
				nil),
			seekType: "hub",
			want:     "hubToken",
		},
		"TokenRaw from base - quay": {
			defaults: NewDockerCheckDefaults(
				"",
				"ghcrToken",
				"hubToken", "",
				"quayToken",
				nil),
			seekType: "quay",
			want:     "quayToken",
		},
		"TokenRaw from defaults - ghcr": {
			defaults: NewDockerCheckDefaults(
				"", "", "", "", "",
				NewDockerCheckDefaults(
					"",
					"ghcrToken",
					"hubToken", "",
					"quayToken",
					nil)),
			seekType: "ghcr",
			want:     "ghcrToken",
		},
		"TokenRaw from defaults - hub": {
			defaults: NewDockerCheckDefaults(
				"", "", "", "", "",
				NewDockerCheckDefaults(
					"",
					"ghcrToken",
					"hubToken", "",
					"quayToken",
					nil)),
			seekType: "hub",
			want:     "hubToken",
		},
		"TokenRaw from defaults - quay": {
			defaults: NewDockerCheckDefaults(
				"", "", "", "", "",
				NewDockerCheckDefaults(
					"",
					"ghcrToken",
					"hubToken", "",
					"quayToken",
					nil)),
			seekType: "quay",
			want:     "quayToken",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN getToken is called on it
			got := tc.defaults.getToken(tc.seekType)

			// THEN the result is what we expect
			if got != tc.want {
				t.Fatalf("expected %q, got: %q", tc.want, got)
			}
		})
	}
}

func TestDockerCheck_getToken(t *testing.T) {
	// GIVEN a DockerCheck
	tests := map[string]struct {
		dockerCheck *DockerCheck
		want        string
	}{
		"nil": {
			dockerCheck: nil,
			want:        "",
		},
		"empty": {
			dockerCheck: NewDockerCheck(
				"", "", "", "", "", "", time.Now(), nil),
			want: "",
		},
		"TokenRaw from base": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"", "",
				"", "baseToken",
				"", time.Now(),
				NewDockerCheckDefaults(
					"ghcr",
					"ghcrToken",
					"hubToken", "",
					"quayToken",
					nil)),
			want: "baseToken",
		},
		"TokenRaw from defaults": {
			dockerCheck: NewDockerCheck(
				"ghcr",
				"", "",
				"", "",
				"", time.Now(),
				NewDockerCheckDefaults(
					"hub",
					"ghcrToken",
					"hubToken", "",
					"quayToken",
					nil)),
			want: "ghcrToken",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN getToken is called on it
			got := tc.dockerCheck.getToken()

			// THEN the result is what we expect
			if got != tc.want {
				t.Fatalf("expected %q, got: %q", tc.want, got)
			}
		})
	}
}

func TestDockerCheckHub_Print(t *testing.T) {
	// GIVEN a DockerCheckHub
	tests := map[string]struct {
		dockerCheckHub *DockerCheckHub
		want           string
	}{
		"nil DockerCheckHub": {
			dockerCheckHub: nil,
			want:           ""},
		"no username": {
			dockerCheckHub: &DockerCheckHub{
				DockerCheckRegistryBase: DockerCheckRegistryBase{
					Token:      "SECRET_TOKEN",
					queryToken: "something",
					validUntil: time.Now()}},
			want: `
token: SECRET_TOKEN`},
		"filled": {
			dockerCheckHub: &DockerCheckHub{
				DockerCheckRegistryBase: DockerCheckRegistryBase{
					Token:      "SECRET_TOKEN",
					queryToken: "something",
					validUntil: time.Now()},
				Username: "admin"},
			want: `
token: SECRET_TOKEN
username: admin`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// different prefixes
			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN String is called
				got := tc.dockerCheckHub.String(prefix)

				// THEN the result is what we expect
				if got != want {
					t.Fatalf("(prefix=%q), want:\n%q\ngot\n%q",
						prefix, want, got)
				}
			}
		})
	}
}

func TestDockerCheckDefaults_Print(t *testing.T) {
	// GIVEN a DockerCheckDefaults
	tests := map[string]struct {
		dockerCheckDefaults *DockerCheckDefaults
		want                string
	}{
		"nil": {
			dockerCheckDefaults: nil,
			want:                ""},
		"empty": {
			dockerCheckDefaults: &DockerCheckDefaults{},
			want:                ""},
		"Type only": {
			dockerCheckDefaults: &DockerCheckDefaults{
				Type: "ghcr"},
			want: `
type: ghcr`},
		"Filled": {
			dockerCheckDefaults: &DockerCheckDefaults{
				Type: "hub",
				RegistryGHCR: &DockerCheckGHCR{
					DockerCheckRegistryBase{
						Token:      "SECRET_TOKEN",
						queryToken: "something",
						validUntil: time.Now()}},
				RegistryHub: &DockerCheckHub{
					DockerCheckRegistryBase: DockerCheckRegistryBase{
						Token:      "ANOTHER_TOKEN",
						queryToken: "something",
						validUntil: time.Now()},
					Username: "admin"},
				RegistryQuay: &DockerCheckQuay{
					DockerCheckRegistryBase{
						Token:      "HOW_MANY_TOKENS",
						queryToken: "something",
						validUntil: time.Now()}}},
			want: `
type: hub
ghcr:
    token: SECRET_TOKEN
hub:
    token: ANOTHER_TOKEN
    username: admin
quay:
    token: HOW_MANY_TOKENS`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// different prefixes
			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN String is called
				got := tc.dockerCheckDefaults.String(prefix)

				// THEN the result is what we expect
				if got != want {
					t.Fatalf("(prefix=%q), expected\n%q\ngot:\n%q",
						prefix, want, got)
				}
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
			want:   "",
		},
		"empty": {
			docker: NewDockerCheck(
				"", "", "", "", "", "", time.Now(), nil),
			want: "{}",
		},
		"filled": {
			docker: NewDockerCheck(
				"ghcr",
				"release-argus/Argus",
				"{{ version }}",
				"test",
				"SECRET_TOKEN",
				"", time.Now(), nil),
			want: `
type: ghcr
username: test
token: SECRET_TOKEN
image: release-argus/Argus
tag: '{{ version }}'`,
		},
		"quotes otherwise invalid yaml strings": {
			docker: NewDockerCheck(
				"", "", "",
				">123",
				"", "", time.Now(), nil),
			want: `
username: '>123'`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// different prefixes
			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the DockerCheck is stringified with String
				got := tc.docker.String(prefix)

				// THEN the result is as expected
				if got != want {
					t.Errorf("(prefix=%q) got:\n%q\nwant:\n%q",
						prefix, got, want)
				}
			}
		})
	}
}
