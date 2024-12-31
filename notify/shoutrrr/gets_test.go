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

//go:build unit

package shoutrrr

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_GetOption(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		"root overrides all": {
			want:             "this",
			rootValue:        test.StringPtr("this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			want:             "this",
			mainValue:        test.StringPtr("this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"default overrides hardDefault": {
			want:             "this",
			defaultValue:     test.StringPtr("this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"hardDefaultValue is last resort": {
			want:             "this",
			hardDefaultValue: test.StringPtr("this"),
		},
		"env var is used": {
			want:      "this",
			env:       map[string]string{"TEST_SHOUTRRR__GET_OPTION__ONE": "this"},
			rootValue: test.StringPtr("${TEST_SHOUTRRR__GET_OPTION__ONE}"),
		},
		"env var partial is used": {
			want:      "this",
			env:       map[string]string{"TEST_SHOUTRRR__GET_OPTION__TWO": "is"},
			rootValue: test.StringPtr("th${TEST_SHOUTRRR__GET_OPTION__TWO}"),
		},
		"empty env var is ignored": {
			want:         "that",
			env:          map[string]string{"TEST_SHOUTRRR__GET_OPTION__THREE": ""},
			rootValue:    test.StringPtr("${TEST_SHOUTRRR__GET_OPTION__THREE}"),
			defaultValue: test.StringPtr("that"),
		},
		"undefined env var is used": {
			want:         "${TEST_SHOUTRRR__GET_OPTION__UNSET}",
			rootValue:    test.StringPtr("${TEST_SHOUTRRR__GET_OPTION__UNSET}"),
			defaultValue: test.StringPtr("that"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Options[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Options[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Options[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefaultValue
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}

			// WHEN GetOption is called
			got := shoutrrr.GetOption(key)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("GetOption:\nwant: %q\ngot:  %q",
					tc.want, got)
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
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		"root overrides all": {
			want:             "this",
			rootValue:        test.StringPtr("this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			want:             "this",
			mainValue:        test.StringPtr("this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"default overrides hardDefault": {
			want:             "this",
			defaultValue:     test.StringPtr("this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"hardDefaultValue is last resort": {
			want:             "this",
			hardDefaultValue: test.StringPtr("this"),
		},
		"env var is used": {
			want:      "this",
			env:       map[string]string{"TEST_SHOUTRRR__GET_URL_FIELD__ONE": "this"},
			rootValue: test.StringPtr("${TEST_SHOUTRRR__GET_URL_FIELD__ONE}"),
		},
		"env var partial is used": {
			want:      "this",
			env:       map[string]string{"TEST_SHOUTRRR__GET_URL_FIELD__TWO": "is"},
			rootValue: test.StringPtr("th${TEST_SHOUTRRR__GET_URL_FIELD__TWO}"),
		},
		"empty env var is ignored": {
			want:         "that",
			env:          map[string]string{"TEST_SHOUTRRR__GET_URL_FIELD__THREE": ""},
			rootValue:    test.StringPtr("${TEST_SHOUTRRR__GET_URL_FIELD__THREE}"),
			defaultValue: test.StringPtr("that"),
		},
		"undefined env var is used": {
			want:         "${TEST_SHOUTRRR__GET_URL_FIELD__UNSET}",
			rootValue:    test.StringPtr("${TEST_SHOUTRRR__GET_URL_FIELD__UNSET}"),
			defaultValue: test.StringPtr("that"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.URLFields[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.URLFields[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.URLFields[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.URLFields[key] = *tc.hardDefaultValue
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}

			// WHEN GetURLField is called
			got := shoutrrr.GetURLField(key)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("GetURLField:\nwant: %q\ngot:  %q",
					tc.want, got)
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
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		"root overrides all": {
			want:             "this",
			rootValue:        test.StringPtr("this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			want:             "this",
			mainValue:        test.StringPtr("this"),
			defaultValue:     test.StringPtr("not_this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"default overrides hardDefault": {
			want:             "this",
			defaultValue:     test.StringPtr("this"),
			hardDefaultValue: test.StringPtr("not_this"),
		},
		"hardDefaultValue is last resort": {
			want:             "this",
			hardDefaultValue: test.StringPtr("this"),
		},
		"env var is used": {
			want:      "this",
			env:       map[string]string{"TEST_SHOUTRRR__GET_PARAM__ONE": "this"},
			rootValue: test.StringPtr("${TEST_SHOUTRRR__GET_PARAM__ONE}"),
		},
		"env var partial is used": {
			want:      "this",
			env:       map[string]string{"TEST_SHOUTRRR__GET_PARAM__TWO": "is"},
			rootValue: test.StringPtr("th${TEST_SHOUTRRR__GET_PARAM__TWO}"),
		},
		"empty env var is ignored": {
			want:         "that",
			env:          map[string]string{"TEST_SHOUTRRR__GET_PARAM__THREE": ""},
			rootValue:    test.StringPtr("${TEST_SHOUTRRR__GET_PARAM__THREE}"),
			defaultValue: test.StringPtr("that"),
		},
		"undefined env var is used": {
			want:         "${TEST_SHOUTRRR__GET_PARAM__UNSET}",
			rootValue:    test.StringPtr("${TEST_SHOUTRRR__GET_PARAM__UNSET}"),
			defaultValue: test.StringPtr("that"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Params[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Params[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Params[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.hardDefaultValue
			}
			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}

			// WHEN GetParam is called
			got := shoutrrr.GetParam(key)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("GetParam:\nwant: %q\ngot:  %q",
					tc.want, got)
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
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		"root overrides all": {
			want:             "1s",
			rootValue:        test.StringPtr("1s"),
			mainValue:        test.StringPtr("2s"),
			defaultValue:     test.StringPtr("2s"),
			hardDefaultValue: test.StringPtr("2s"),
		},
		"main overrides default and hardDefault": {
			want:             "1s",
			mainValue:        test.StringPtr("1s"),
			defaultValue:     test.StringPtr("2s"),
			hardDefaultValue: test.StringPtr("2s"),
		},
		"default overrides hardDefault": {
			want:             "1s",
			defaultValue:     test.StringPtr("1s"),
			hardDefaultValue: test.StringPtr("2s"),
		},
		"hardDefaultValue is last resort": {
			want:             "1s",
			hardDefaultValue: test.StringPtr("1s"),
		},
		"no delay anywhere defaults to 0s": {want: "0s",
			rootValue:        nil,
			defaultValue:     nil,
			hardDefaultValue: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "delay"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Options[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Options[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Options[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefaultValue
			}

			// WHEN GetDelay is called
			got := shoutrrr.GetDelay()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestShoutrrr_GetDelayDuration(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 time.Duration
	}{
		"root overrides all": {
			want:             1 * time.Second,
			rootValue:        test.StringPtr("1s"),
			mainValue:        test.StringPtr("2s"),
			defaultValue:     test.StringPtr("2s"),
			hardDefaultValue: test.StringPtr("2s"),
		},
		"main overrides default and hardDefault": {
			want:             1 * time.Second,
			mainValue:        test.StringPtr("1s"),
			defaultValue:     test.StringPtr("2s"),
			hardDefaultValue: test.StringPtr("2s"),
		},
		"default overrides hardDefault": {
			want:             1 * time.Second,
			defaultValue:     test.StringPtr("1s"),
			hardDefaultValue: test.StringPtr("2s"),
		},
		"hardDefaultValue is last resort": {
			want:             1 * time.Second,
			hardDefaultValue: test.StringPtr("1s"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "delay"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Options[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Options[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Options[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefaultValue
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
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 int
	}{
		"root overrides all": {
			want:             1,
			rootValue:        test.StringPtr("1"),
			mainValue:        test.StringPtr("2"),
			defaultValue:     test.StringPtr("2"),
			hardDefaultValue: test.StringPtr("2"),
		},
		"main overrides default and hardDefault": {
			want:             1,
			mainValue:        test.StringPtr("1"),
			defaultValue:     test.StringPtr("2"),
			hardDefaultValue: test.StringPtr("2"),
		},
		"default overrides hardDefault": {
			want:             1,
			defaultValue:     test.StringPtr("1"),
			hardDefaultValue: test.StringPtr("2"),
		},
		"hardDefaultValue is last resort": {
			want:             1,
			hardDefaultValue: test.StringPtr("1"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "max_tries"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Options[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Options[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Options[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefaultValue
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
	serviceInfo := util.ServiceInfo{
		ID:            "release-argus/Argus",
		URL:           "https://github.com",
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		"root overrides all": {
			want:             "New version!",
			rootValue:        test.StringPtr("New version!"),
			mainValue:        test.StringPtr("something"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"main overrides default and hardDefault": {
			want:             "New version!",
			mainValue:        test.StringPtr("New version!"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"default overrides hardDefault": {
			want:             "New version!",
			defaultValue:     test.StringPtr("New version!"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"hardDefaultValue is last resort": {
			want:             "New version!",
			hardDefaultValue: test.StringPtr("New version!"),
		},
		"jinja templating": {
			want:             "New version!",
			rootValue:        test.StringPtr("{% if 'a' == 'a' %}New version!{% endif %}"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"jinja vars": {
			want: fmt.Sprintf("%s or %s/%s/releases/tag/%s",
				serviceInfo.WebURL, serviceInfo.URL, serviceInfo.ID, serviceInfo.LatestVersion),
			rootValue:        test.StringPtr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "message"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Options[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Options[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Options[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.hardDefaultValue
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
	serviceInfo := util.ServiceInfo{
		ID:            "release-argus/Argus",
		URL:           "https://github.com",
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := map[string]struct {
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		"root overrides all": {
			want:             "New version!",
			rootValue:        test.StringPtr("New version!"),
			mainValue:        test.StringPtr("something"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"main overrides default and hardDefault": {
			want:             "New version!",
			mainValue:        test.StringPtr("New version!"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"default overrides hardDefault": {
			want:             "New version!",
			defaultValue:     test.StringPtr("New version!"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"hardDefaultValue is last resort": {
			want:             "New version!",
			hardDefaultValue: test.StringPtr("New version!"),
		},
		"jinja templating": {
			want:             "New version!",
			rootValue:        test.StringPtr("{% if 'a' == 'a' %}New version!{% endif %}"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
		"jinja vars": {
			want: fmt.Sprintf("%s or %s/%s/releases/tag/%s",
				serviceInfo.WebURL, serviceInfo.URL, serviceInfo.ID, serviceInfo.LatestVersion),
			rootValue:        test.StringPtr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			defaultValue:     test.StringPtr("something"),
			hardDefaultValue: test.StringPtr("something"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "title"
			shoutrrr := testShoutrrr(false, false)
			if tc.rootValue != nil {
				shoutrrr.Params[key] = *tc.rootValue
			}
			if tc.mainValue != nil {
				shoutrrr.Main.Params[key] = *tc.mainValue
			}
			if tc.defaultValue != nil {
				shoutrrr.Defaults.Params[key] = *tc.defaultValue
			}
			if tc.hardDefaultValue != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.hardDefaultValue
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
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		"root overrides all": {
			want:             "smtp",
			rootValue:        "smtp",
			mainValue:        "other",
			defaultValue:     "other",
			hardDefaultValue: "other",
		},
		"main overrides default and hardDefault": {
			want:             "smtp",
			mainValue:        "smtp",
			defaultValue:     "other",
			hardDefaultValue: "other",
		},
		"default is ignored": { // uses ID
			want:         "test",
			defaultValue: "smtp",
		},
		"hardDefaultValue is ignored": { // uses ID
			want:             "test",
			hardDefaultValue: "smtp",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.rootValue
			shoutrrr.Main.Type = tc.mainValue

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
