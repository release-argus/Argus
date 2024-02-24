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

package shoutrrr

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_GetOption(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		env         map[string]string
		root        *string
		main        *string
		dfault      *string
		hardDefault *string
		wantString  string
	}{
		"root overrides all": {
			wantString:  "this",
			root:        stringPtr("this"),
			dfault:      stringPtr("not_this"),
			hardDefault: stringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			wantString:  "this",
			main:        stringPtr("this"),
			dfault:      stringPtr("not_this"),
			hardDefault: stringPtr("not_this"),
		},
		"default overrides hardDefault": {
			wantString:  "this",
			dfault:      stringPtr("this"),
			hardDefault: stringPtr("not_this"),
		},
		"hardDefault is last resort": {
			wantString:  "this",
			hardDefault: stringPtr("this"),
		},
		"env var is used": {
			wantString: "this",
			env:        map[string]string{"TESTSHOUTRRR_GETOPTION_ONE": "this"},
			root:       stringPtr("${TESTSHOUTRRR_GETOPTION_ONE}"),
		},
		"env var partial is used": {
			wantString: "this",
			env:        map[string]string{"TESTSHOUTRRR_GETOPTION_TWO": "is"},
			root:       stringPtr("th${TESTSHOUTRRR_GETOPTION_TWO}"),
		},
		"empty env var is ignored": {
			wantString: "that",
			env:        map[string]string{"TESTSHOUTRRR_GETOPTION_THREE": ""},
			root:       stringPtr("${TESTSHOUTRRR_GETOPTION_THREE}"),
			dfault:     stringPtr("that"),
		},
		"undefined env var is used": {
			wantString: "${TESTSHOUTRRR_GETOPTION_UNSET}",
			root:       stringPtr("${TESTSHOUTRRR_GETOPTION_UNSET}"),
			dfault:     stringPtr("that"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			if tc.root != nil {
				shoutrrr.Options[key] = *tc.root
			}
			if tc.main != nil {
				shoutrrr.Main.Options[key] = *tc.main
			}
			if tc.dfault != nil {
				shoutrrr.Defaults.Options[key] = *tc.dfault
			}
			if tc.hardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefault
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN GetOption is called
			got := shoutrrr.GetOption(key)

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Fatalf("GetOption:\nwant: %q\ngot:  %q",
					tc.wantString, got)
			}

			// WHEN SetOption is called
			want := got + "-set-test"
			shoutrrr.SetOption(key, want)

			// THEN the Option is set and can be retrieved with a Get
			got = shoutrrr.GetOption(key)
			if got != want {
				t.Fatalf("SetOption:\nwant: %q\ngot:  %q",
					want, got)
			}
		})
	}
}

func TestShoutrrr_GetURLField(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		env         map[string]string
		root        *string
		main        *string
		dfault      *string
		hardDefault *string
		wantString  string
	}{
		"root overrides all": {
			wantString:  "this",
			root:        stringPtr("this"),
			dfault:      stringPtr("not_this"),
			hardDefault: stringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			wantString:  "this",
			main:        stringPtr("this"),
			dfault:      stringPtr("not_this"),
			hardDefault: stringPtr("not_this"),
		},
		"default overrides hardDefault": {
			wantString:  "this",
			dfault:      stringPtr("this"),
			hardDefault: stringPtr("not_this"),
		},
		"hardDefault is last resort": {
			wantString:  "this",
			hardDefault: stringPtr("this"),
		},
		"env var is used": {
			wantString: "this",
			env:        map[string]string{"TESTSHOUTRRR_GETURLFIELD_ONE": "this"},
			root:       stringPtr("${TESTSHOUTRRR_GETURLFIELD_ONE}"),
		},
		"env var partial is used": {
			wantString: "this",
			env:        map[string]string{"TESTSHOUTRRR_GETURLFIELD_TWO": "is"},
			root:       stringPtr("th${TESTSHOUTRRR_GETURLFIELD_TWO}"),
		},
		"empty env var is ignored": {
			wantString: "that",
			env:        map[string]string{"TESTSHOUTRRR_GETURLFIELD_THREE": ""},
			root:       stringPtr("${TESTSHOUTRRR_GETURLFIELD_THREE}"),
			dfault:     stringPtr("that"),
		},
		"undefined env var is used": {
			wantString: "${TESTSHOUTRRR_GETURLFIELD_UNSET}",
			root:       stringPtr("${TESTSHOUTRRR_GETURLFIELD_UNSET}"),
			dfault:     stringPtr("that"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			if tc.root != nil {
				shoutrrr.URLFields[key] = *tc.root
			}
			if tc.main != nil {
				shoutrrr.Main.URLFields[key] = *tc.main
			}
			if tc.dfault != nil {
				shoutrrr.Defaults.URLFields[key] = *tc.dfault
			}
			if tc.hardDefault != nil {
				shoutrrr.HardDefaults.URLFields[key] = *tc.hardDefault
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN GetURLField is called
			got := shoutrrr.GetURLField(key)

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Fatalf("GetURLField:\nwant: %q\ngot:  %q",
					tc.wantString, got)
			}

			// WHEN SetURLField is called
			want := got + "-set-test"
			shoutrrr.SetURLField(key, want)

			// THEN the URLField is set and can be retrieved with a Get
			got = shoutrrr.GetURLField(key)
			if got != want {
				t.Fatalf("SetURLField:\nwant: %q\ngot:  %q",
					want, got)
			}
		})
	}
}

func TestShoutrrr_GetParam(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		env         map[string]string
		root        *string
		main        *string
		dfault      *string
		hardDefault *string
		wantString  string
	}{
		"root overrides all": {
			wantString:  "this",
			root:        stringPtr("this"),
			dfault:      stringPtr("not_this"),
			hardDefault: stringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			wantString:  "this",
			main:        stringPtr("this"),
			dfault:      stringPtr("not_this"),
			hardDefault: stringPtr("not_this"),
		},
		"default overrides hardDefault": {
			wantString:  "this",
			dfault:      stringPtr("this"),
			hardDefault: stringPtr("not_this"),
		},
		"hardDefault is last resort": {
			wantString:  "this",
			hardDefault: stringPtr("this"),
		},
		"env var is used": {
			wantString: "this",
			env:        map[string]string{"TESTSHOUTRRR_GETPARAM_ONE": "this"},
			root:       stringPtr("${TESTSHOUTRRR_GETPARAM_ONE}"),
		},
		"env var partial is used": {
			wantString: "this",
			env:        map[string]string{"TESTSHOUTRRR_GETPARAM_TWO": "is"},
			root:       stringPtr("th${TESTSHOUTRRR_GETPARAM_TWO}"),
		},
		"empty env var is ignored": {
			wantString: "that",
			env:        map[string]string{"TESTSHOUTRRR_GETPARAM_THREE": ""},
			root:       stringPtr("${TESTSHOUTRRR_GETPARAM_THREE}"),
			dfault:     stringPtr("that"),
		},
		"undefined env var is used": {
			wantString: "${TESTSHOUTRRR_GETPARAM_UNSET}",
			root:       stringPtr("${TESTSHOUTRRR_GETPARAM_UNSET}"),
			dfault:     stringPtr("that"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			if tc.root != nil {
				shoutrrr.Params[key] = *tc.root
			}
			if tc.main != nil {
				shoutrrr.Main.Params[key] = *tc.main
			}
			if tc.dfault != nil {
				shoutrrr.Defaults.Params[key] = *tc.dfault
			}
			if tc.hardDefault != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.hardDefault
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN GetParam is called
			got := shoutrrr.GetParam(key)

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Fatalf("GetParam:\nwant: %q\ngot:  %q",
					tc.wantString, got)
			}

			// WHEN SetParam is called
			want := got + "-set-test"
			shoutrrr.SetParam(key, want)

			// THEN the Param is set and can be retrieved with a Get
			got = shoutrrr.GetParam(key)
			if got != want {
				t.Fatalf("SetParam:\nwant: %q\ngot:  %q",
					want, got)
			}
		})
	}
}

func TestShoutrrr_GetDelay(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		root        *string
		main        *string
		dfault      *string
		hardDefault *string
		wantString  string
	}{
		"root overrides all": {
			wantString:  "1s",
			root:        stringPtr("1s"),
			main:        stringPtr("2s"),
			dfault:      stringPtr("2s"),
			hardDefault: stringPtr("2s"),
		},
		"main overrides default and hardDefault": {
			wantString:  "1s",
			main:        stringPtr("1s"),
			dfault:      stringPtr("2s"),
			hardDefault: stringPtr("2s"),
		},
		"default overrides hardDefault": {
			wantString:  "1s",
			dfault:      stringPtr("1s"),
			hardDefault: stringPtr("2s"),
		},
		"hardDefault is last resort": {
			wantString:  "1s",
			hardDefault: stringPtr("1s"),
		},
		"no delay anywhere defaults to 0s": {wantString: "0s",
			root:        nil,
			dfault:      nil,
			hardDefault: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "delay"
			shoutrrr := testShoutrrr(false, false)
			if tc.root != nil {
				shoutrrr.Options[key] = *tc.root
			}
			if tc.main != nil {
				shoutrrr.Main.Options[key] = *tc.main
			}
			if tc.dfault != nil {
				shoutrrr.Defaults.Options[key] = *tc.dfault
			}
			if tc.hardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefault
			}

			// WHEN GetDelay is called
			got := shoutrrr.GetDelay()

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Fatalf("want: %q\ngot:  %q",
					tc.wantString, got)
			}
		})
	}
}

func TestShoutrrr_GetDelayDuration(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		root        *string
		main        *string
		dfault      *string
		hardDefault *string
		want        time.Duration
	}{
		"root overrides all": {
			want:        1 * time.Second,
			root:        stringPtr("1s"),
			main:        stringPtr("2s"),
			dfault:      stringPtr("2s"),
			hardDefault: stringPtr("2s"),
		},
		"main overrides default and hardDefault": {
			want:        1 * time.Second,
			main:        stringPtr("1s"),
			dfault:      stringPtr("2s"),
			hardDefault: stringPtr("2s"),
		},
		"default overrides hardDefault": {
			want:        1 * time.Second,
			dfault:      stringPtr("1s"),
			hardDefault: stringPtr("2s"),
		},
		"hardDefault is last resort": {
			want:        1 * time.Second,
			hardDefault: stringPtr("1s"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "delay"
			shoutrrr := testShoutrrr(false, false)
			if tc.root != nil {
				shoutrrr.Options[key] = *tc.root
			}
			if tc.main != nil {
				shoutrrr.Main.Options[key] = *tc.main
			}
			if tc.dfault != nil {
				shoutrrr.Defaults.Options[key] = *tc.dfault
			}
			if tc.hardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefault
			}

			// WHEN GetDelay is called
			got := shoutrrr.GetDelayDuration()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestShoutrrr_GetMaxTries(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		maxTriesRoot        *string
		maxTriesMain        *string
		maxTriesDefault     *string
		maxTriesHardDefault *string
		want                int
	}{
		"root overrides all": {
			want:                1,
			maxTriesRoot:        stringPtr("1"),
			maxTriesMain:        stringPtr("2"),
			maxTriesDefault:     stringPtr("2"),
			maxTriesHardDefault: stringPtr("2"),
		},
		"main overrides default and hardDefault": {
			want:                1,
			maxTriesMain:        stringPtr("1"),
			maxTriesDefault:     stringPtr("2"),
			maxTriesHardDefault: stringPtr("2"),
		},
		"default overrides hardDefault": {
			want:                1,
			maxTriesDefault:     stringPtr("1"),
			maxTriesHardDefault: stringPtr("2"),
		},
		"hardDefault is last resort": {
			want:                1,
			maxTriesHardDefault: stringPtr("1"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "max_tries"
			shoutrrr := testShoutrrr(false, false)
			if tc.maxTriesRoot != nil {
				shoutrrr.Options[key] = *tc.maxTriesRoot
			}
			if tc.maxTriesMain != nil {
				shoutrrr.Main.Options[key] = *tc.maxTriesMain
			}
			if tc.maxTriesDefault != nil {
				shoutrrr.Defaults.Options[key] = *tc.maxTriesDefault
			}
			if tc.maxTriesHardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.maxTriesHardDefault
			}

			// WHEN GetMaxTries is called
			got := shoutrrr.GetMaxTries()

			// THEN the function returns the correct result
			if int(got) != tc.want {
				t.Fatalf("want: %d\ngot:  %d",
					tc.want, got)
			}
		})
	}
}

func TestShoutrrr_Message(t *testing.T) {
	// GIVEN a Shoutrrr
	serviceInfo := &util.ServiceInfo{
		ID:            "release-argus/Argus",
		URL:           "https://github.com",
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := map[string]struct {
		root        *string
		main        *string
		dfault      *string
		hardDefault *string
		want        string
	}{
		"root overrides all": {
			want:        "New version!",
			root:        stringPtr("New version!"),
			main:        stringPtr("something"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
		"main overrides default and hardDefault": {
			want:        "New version!",
			main:        stringPtr("New version!"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
		"default overrides hardDefault": {
			want:        "New version!",
			dfault:      stringPtr("New version!"),
			hardDefault: stringPtr("something"),
		},
		"hardDefault is last resort": {
			want:        "New version!",
			hardDefault: stringPtr("New version!"),
		},
		"jinja templating": {
			want:        "New version!",
			root:        stringPtr("{% if 'a' == 'a' %}New version!{% endif %}"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
		"jinja vars": {
			want: fmt.Sprintf("%s or %s/%s/releases/tag/%s",
				serviceInfo.WebURL, serviceInfo.URL, serviceInfo.ID, serviceInfo.LatestVersion),
			root:        stringPtr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "message"
			shoutrrr := testShoutrrr(false, false)
			if tc.root != nil {
				shoutrrr.Options[key] = *tc.root
			}
			if tc.main != nil {
				shoutrrr.Main.Options[key] = *tc.main
			}
			if tc.dfault != nil {
				shoutrrr.Defaults.Options[key] = *tc.dfault
			}
			if tc.hardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefault
			}

			// WHEN Message is called
			got := shoutrrr.Message(serviceInfo)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestShoutrrr_Title(t *testing.T) {
	// GIVEN a Shoutrrr
	serviceInfo := &util.ServiceInfo{
		ID:            "release-argus/Argus",
		URL:           "https://github.com",
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := map[string]struct {
		root        *string
		main        *string
		dfault      *string
		hardDefault *string
		want        string
	}{
		"root overrides all": {
			want:        "New version!",
			root:        stringPtr("New version!"),
			main:        stringPtr("something"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
		"main overrides default and hardDefault": {
			want:        "New version!",
			main:        stringPtr("New version!"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
		"default overrides hardDefault": {
			want:        "New version!",
			dfault:      stringPtr("New version!"),
			hardDefault: stringPtr("something"),
		},
		"hardDefault is last resort": {
			want:        "New version!",
			hardDefault: stringPtr("New version!"),
		},
		"jinja templating": {
			want:        "New version!",
			root:        stringPtr("{% if 'a' == 'a' %}New version!{% endif %}"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
		"jinja vars": {
			want: fmt.Sprintf("%s or %s/%s/releases/tag/%s",
				serviceInfo.WebURL, serviceInfo.URL, serviceInfo.ID, serviceInfo.LatestVersion),
			root:        stringPtr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			dfault:      stringPtr("something"),
			hardDefault: stringPtr("something"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "title"
			shoutrrr := testShoutrrr(false, false)
			if tc.root != nil {
				shoutrrr.Params[key] = *tc.root
			}
			if tc.main != nil {
				shoutrrr.Main.Params[key] = *tc.main
			}
			if tc.dfault != nil {
				shoutrrr.Defaults.Params[key] = *tc.dfault
			}
			if tc.hardDefault != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.hardDefault
			}

			// WHEN Title is called
			got := shoutrrr.Title(serviceInfo)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestShoutrrr_GetType(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		root        string
		main        string
		dfault      string
		hardDefault string
		want        string
	}{
		"root overrides all": {
			want:        "smtp",
			root:        "smtp",
			main:        "other",
			dfault:      "other",
			hardDefault: "other",
		},
		"main overrides default and hardDefault": {
			want:        "smtp",
			main:        "smtp",
			dfault:      "other",
			hardDefault: "other",
		},
		"default is ignored": { // uses ID
			want:   "test",
			dfault: "smtp",
		},
		"hardDefault is ignored": { // uses ID
			want:        "test",
			hardDefault: "smtp",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.root
			shoutrrr.Main.Type = tc.main

			// WHEN GetType is called
			got := shoutrrr.GetType()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}
