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

//go:build unit

package shoutrrr

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_GetType(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "smtp",
			rootValue:        "smtp",
			mainValue:        "other",
			defaultValue:     "other",
			hardDefaultValue: "other",
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "smtp",
			mainValue:        "smtp",
			defaultValue:     "other",
			hardDefaultValue: "other",
		},
		{
			name:         "default is ignored", // uses ID.
			want:         "test",
			defaultValue: "smtp",
		},
		{
			name:             "hardDefaultValue is ignored", // uses ID.
			want:             "test",
			hardDefaultValue: "smtp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.rootValue
			shoutrrr.Main.Type = tc.mainValue

			// WHEN: GetType is called.
			got := shoutrrr.GetType()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr GetType() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrr_Title(t *testing.T) {
	// GIVEN: a Shoutrrr.
	svcInfo := serviceinfo.ServiceInfo{
		ID:            test.ArgusGitHubRepo,
		URL:           "https://github.com",
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "New version!",
			rootValue:        test.Ptr("New version!"),
			mainValue:        test.Ptr("something"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "New version!",
			mainValue:        test.Ptr("New version!"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "New version!",
			defaultValue:     test.Ptr("New version!"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "New version!",
			hardDefaultValue: test.Ptr("New version!"),
		},
		{
			name:             "django templating",
			want:             "New version!",
			rootValue:        test.Ptr("{% if 'a' == 'a' %}New version!{% endif %}"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name: "django vars",
			want: fmt.Sprintf(
				"%s or %s/%s/releases/tag/%s",
				svcInfo.WebURL, svcInfo.URL, svcInfo.ID, svcInfo.LatestVersion,
			),
			rootValue:        test.Ptr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			// WHEN: Title is called.
			got := shoutrrr.Title(svcInfo)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr Title(%+v) mismatch\ngot:  %q\nwant: %q",
					packageName, svcInfo, got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrr_Message(t *testing.T) {
	// GIVEN: a Shoutrrr.
	svcInfo := serviceinfo.ServiceInfo{
		ID:            test.ArgusGitHubRepo,
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "New version!",
			rootValue:        test.Ptr("New version!"),
			mainValue:        test.Ptr("something"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "New version!",
			mainValue:        test.Ptr("New version!"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "New version!",
			defaultValue:     test.Ptr("New version!"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "New version!",
			hardDefaultValue: test.Ptr("New version!"),
		},
		{
			name:             "django templating",
			want:             "New version!",
			rootValue:        test.Ptr("{% if 'a' == 'a' %}New version!{% endif %}"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
		{
			name: "django vars",
			want: fmt.Sprintf(
				"%s or %s/%s/releases/tag/%s",
				svcInfo.WebURL, svcInfo.URL, svcInfo.ID, svcInfo.LatestVersion,
			),
			rootValue:        test.Ptr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			defaultValue:     test.Ptr("something"),
			hardDefaultValue: test.Ptr("something"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			// WHEN: Message is called.
			got := shoutrrr.Message(svcInfo)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr Message(%+v) mismatch\ngot:  %q\nwant: %q",
					packageName, svcInfo, got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrr_GetOption(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "this",
			rootValue:        test.Ptr("this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "this",
			mainValue:        test.Ptr("this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "this",
			defaultValue:     test.Ptr("this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "this",
			hardDefaultValue: test.Ptr("this"),
		},
		{
			name: "env var is used",
			want: "this",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_OPTION__ONE": "this",
			},
			rootValue: test.Ptr("${TEST_SHOUTRRR__GET_OPTION__ONE}"),
		},
		{
			name: "env var partial is used",
			want: "this",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_OPTION__TWO": "is",
			},
			rootValue: test.Ptr("th${TEST_SHOUTRRR__GET_OPTION__TWO}"),
		},
		{
			name: "empty env var is ignored",
			want: "that",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_OPTION__THREE": "",
			},
			rootValue:    test.Ptr("${TEST_SHOUTRRR__GET_OPTION__THREE}"),
			defaultValue: test.Ptr("that"),
		},
		{
			name:         "undefined env var is used",
			want:         "${TEST_SHOUTRRR__GET_OPTION__UNSET}",
			rootValue:    test.Ptr("${TEST_SHOUTRRR__GET_OPTION__UNSET}"),
			defaultValue: test.Ptr("that"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			test.SetEnv(t, tc.env)

			// WHEN: GetOption is called.
			got := shoutrrr.GetOption(key)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr GetOption(%q) mismatch:\ngot:  %q\nwant: %q",
					packageName, key,
					got, tc.want,
				)
			}

			// WHEN: setOption is called.
			want := got + "-set-test"
			shoutrrr.setOption(key, want)

			// THEN: the Option is set and can be retrieved with a Get.
			got = shoutrrr.GetOption(key)
			if got != want {
				t.Fatalf(
					"%s\nShoutrrr GetOption(%q) mismatch after setOption(%q):\ngot:  %q\nwant: %q",
					packageName, key, want,
					got, want,
				)
			}
		})
	}
}
func TestShoutrrr_GetURLField(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "this",
			rootValue:        test.Ptr("this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "this",
			mainValue:        test.Ptr("this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "this",
			defaultValue:     test.Ptr("this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "this",
			hardDefaultValue: test.Ptr("this"),
		},
		{
			name: "env var is used",
			want: "this",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_URL_FIELD__ONE": "this",
			},
			rootValue: test.Ptr("${TEST_SHOUTRRR__GET_URL_FIELD__ONE}"),
		},
		{
			name: "env var partial is used",
			want: "this",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_URL_FIELD__TWO": "is",
			},
			rootValue: test.Ptr("th${TEST_SHOUTRRR__GET_URL_FIELD__TWO}"),
		},
		{
			name: "empty env var is ignored",
			want: "that",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_URL_FIELD__THREE": "",
			},
			rootValue:    test.Ptr("${TEST_SHOUTRRR__GET_URL_FIELD__THREE}"),
			defaultValue: test.Ptr("that"),
		},
		{
			name:         "undefined env var is used",
			want:         "${TEST_SHOUTRRR__GET_URL_FIELD__UNSET}",
			rootValue:    test.Ptr("${TEST_SHOUTRRR__GET_URL_FIELD__UNSET}"),
			defaultValue: test.Ptr("that"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			test.SetEnv(t, tc.env)

			// WHEN: GetURLField is called.
			got := shoutrrr.GetURLField(key)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr,GetURLField(%q) mismatch before setURLField():\ngot:  %q\nwant: %q",
					packageName, key,
					got, tc.want,
				)
			}

			// WHEN: setURLField is called.
			want := got + "-set-test"
			shoutrrr.setURLField(key, want)

			// THEN: the URLField is set and can be retrieved with a Get.
			got = shoutrrr.GetURLField(key)
			if got != want {
				t.Fatalf(
					"%s\nShoutrrr.GetURLField(%q) mismatch after setURLField(%q):\ngot:  %q\nwant: %q",
					packageName, key, want,
					got, want,
				)
			}
		})
	}
}
func TestShoutrrr_GetParam(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "this",
			rootValue:        test.Ptr("this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "this",
			mainValue:        test.Ptr("this"),
			defaultValue:     test.Ptr("not_this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "this",
			defaultValue:     test.Ptr("this"),
			hardDefaultValue: test.Ptr("not_this"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "this",
			hardDefaultValue: test.Ptr("this"),
		},
		{
			name: "env var is used",
			want: "this",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_PARAM__ONE": "this",
			},
			rootValue: test.Ptr("${TEST_SHOUTRRR__GET_PARAM__ONE}"),
		},
		{
			name: "env var partial is used",
			want: "this",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_PARAM__TWO": "is",
			},
			rootValue: test.Ptr("th${TEST_SHOUTRRR__GET_PARAM__TWO}"),
		},
		{
			name: "empty env var is ignored",
			want: "that",
			env: map[string]string{
				"TEST_SHOUTRRR__GET_PARAM__THREE": "",
			},
			rootValue:    test.Ptr("${TEST_SHOUTRRR__GET_PARAM__THREE}"),
			defaultValue: test.Ptr("that"),
		},
		{
			name:         "undefined env var is used",
			want:         "${TEST_SHOUTRRR__GET_PARAM__UNSET}",
			rootValue:    test.Ptr("${TEST_SHOUTRRR__GET_PARAM__UNSET}"),
			defaultValue: test.Ptr("that"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			test.SetEnv(t, tc.env)

			// WHEN: GetParam is called.
			got := shoutrrr.GetParam(key)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.GetParam(%q) mismatch before setParam():\ngot:  %q\nwant: %q",
					packageName, key,
					got, tc.want,
				)
			}

			// WHEN: setParam is called.
			want := got + "-set-test"
			shoutrrr.setParam(key, want)

			// THEN: the Param is set and can be retrieved with a Get.
			got = shoutrrr.GetParam(key)
			if got != want {
				t.Fatalf(
					"%s\nShoutrrr GetParam(%q) mismatch after setParam(%q):\ngot:  %q\nwant: %q",
					packageName, key, want,
					got, want,
				)
			}
		})
	}
}

func TestShoutrrr_GetDelay(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "1s",
			rootValue:        test.Ptr("1s"),
			mainValue:        test.Ptr("2s"),
			defaultValue:     test.Ptr("2s"),
			hardDefaultValue: test.Ptr("2s"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "1s",
			mainValue:        test.Ptr("1s"),
			defaultValue:     test.Ptr("2s"),
			hardDefaultValue: test.Ptr("2s"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "1s",
			defaultValue:     test.Ptr("1s"),
			hardDefaultValue: test.Ptr("2s"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "1s",
			hardDefaultValue: test.Ptr("1s"),
		},
		{
			name:             "no delay anywhere defaults to 0s",
			want:             "0s",
			rootValue:        nil,
			defaultValue:     nil,
			hardDefaultValue: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			// WHEN: GetDelay is called.
			got := shoutrrr.GetDelay()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.GetDelay() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}
func TestShoutrrr_GetDelayDuration(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 time.Duration
	}{
		{
			name:             "root overrides all",
			want:             1 * time.Second,
			rootValue:        test.Ptr("1s"),
			mainValue:        test.Ptr("2s"),
			defaultValue:     test.Ptr("2s"),
			hardDefaultValue: test.Ptr("2s"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             1 * time.Second,
			mainValue:        test.Ptr("1s"),
			defaultValue:     test.Ptr("2s"),
			hardDefaultValue: test.Ptr("2s"),
		},
		{
			name:             "default overrides hardDefault",
			want:             1 * time.Second,
			defaultValue:     test.Ptr("1s"),
			hardDefaultValue: test.Ptr("2s"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             1 * time.Second,
			hardDefaultValue: test.Ptr("1s"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			// WHEN: GetDelay is called.
			got := shoutrrr.GetDelayDuration()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.GetDelayDuration() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestShoutrrr_GetMaxTries(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 int
	}{
		{
			name:             "root overrides all",
			want:             1,
			rootValue:        test.Ptr("1"),
			mainValue:        test.Ptr("2"),
			defaultValue:     test.Ptr("2"),
			hardDefaultValue: test.Ptr("2"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             1,
			mainValue:        test.Ptr("1"),
			defaultValue:     test.Ptr("2"),
			hardDefaultValue: test.Ptr("2"),
		},
		{
			name:             "default overrides hardDefault",
			want:             1,
			defaultValue:     test.Ptr("1"),
			hardDefaultValue: test.Ptr("2"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             1,
			hardDefaultValue: test.Ptr("1"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			// WHEN: GetMaxTries is called.
			got := shoutrrr.GetMaxTries()

			// THEN: the function returns the correct result.
			if int(got) != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.GetMaxTries() value mismatch\ngot:  %d\nwant: %d",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestBase_X_Option(t *testing.T) {
	// GIVEN: a map[string]string.
	values := map[string]string{
		"a":  "foo",
		"b":  "bar",
		"hi": "bye",
	}

	// AND: a set of keys to query for.
	tests := []struct {
		name      string
		key, want string
	}{
		{
			name: "unknown key",
			key:  "foo",
			want: "",
		},
		{
			name: "value 1",
			key:  "a",
			want: "foo",
		},
		{
			name: "value 2",
			key:  "b",
			want: "bar",
		},
		{
			name: "value 3",
			key:  "hi",
			want: "bye",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Base with this Options map.
			base := Base{
				Options: util.CopyMap(values),
			}

			// WHEN: GetOption is called.
			got := base.getOption(tc.key)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.GetOption(%q) mismatch before setOption():\ngot:  %q\nwant: %q",
					packageName, tc.key,
					got, tc.want,
				)
			}

			// WHEN: setOption is called.
			want := got + "-set-test"
			base.setOption(tc.key, want)

			// THEN: the Option is set and can be retrieved with a Get.
			got = base.getOption(tc.key)
			if got != want {
				t.Fatalf(
					"%s\nShoutrrr.GetOption(%q) mismatch after setOption(%q):\ngot:  %q\nwant: %q",
					packageName, tc.key, want,
					got, want,
				)
			}
		})
	}
}
func TestBase_X_URLField(t *testing.T) {
	// GIVEN: a map[string]string.
	values := map[string]string{
		"a":  "foo",
		"b":  "bar",
		"hi": "bye",
	}

	// AND: a set of keys to query for.
	tests := []struct {
		name      string
		key, want string
	}{
		{
			name: "unknown key",
			key:  "foo",
			want: "",
		},
		{
			name: "value 1",
			key:  "a",
			want: "foo",
		},
		{
			name: "value 2",
			key:  "b",
			want: "bar",
		},
		{
			name: "value 3",
			key:  "hi",
			want: "bye",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Base with this URLFields map.
			base := Base{
				URLFields: util.CopyMap(values),
			}

			// WHEN: GetURLField is called.
			got := base.getURLField(tc.key)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.GetURLField(%q) mismatch before setURLField():\ngot:  %q\nwant: %q",
					packageName, tc.key,
					got, tc.want,
				)
			}

			// WHEN: setURLField is called.
			want := got + "-set-test"
			base.setURLField(tc.key, want)

			// THEN: the Option is set and can be retrieved with a Get.
			got = base.getURLField(tc.key)
			if got != want {
				t.Fatalf(
					"%s\nShoutrrr.GetURLField(%q) mismatch after setURLField(%q):\ngot:  %q\nwant: %q",
					packageName, tc.key, want,
					got, want,
				)
			}
		})
	}
}
func TestBase_X_Param(t *testing.T) {
	// GIVEN: a map[string]string.
	values := map[string]string{
		"a":  "foo",
		"b":  "bar",
		"hi": "bye",
	}

	// AND: a set of keys to query for.
	tests := []struct {
		name      string
		key, want string
	}{
		{
			name: "unknown key",
			key:  "foo",
			want: "",
		},
		{
			name: "value 1",
			key:  "a",
			want: "foo",
		},
		{
			name: "value 2",
			key:  "b",
			want: "bar",
		},
		{
			name: "value 3",
			key:  "hi",
			want: "bye",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Base with this Params map.
			base := Base{
				Params: util.CopyMap(values),
			}

			// WHEN: GetParam is called.
			got := base.GetParam(tc.key)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Fatalf(
					"%s\nShoutrrr.GetParam(%q) mismatch before setParam():\ngot:  %q\nwant: %q",
					packageName, tc.key,
					got, tc.want,
				)
			}

			// WHEN: setParam is called.
			want := got + "-set-test"
			base.setParam(tc.key, want)

			// THEN: the Option is set and can be retrieved with a Get.
			got = base.GetParam(tc.key)
			if got != want {
				t.Fatalf(
					"%s\nShoutrrr.GetParam(%q) mismatch after setParam(%q):\ngot:  %q\nwant: %q",
					packageName, tc.key, want,
					got, want,
				)
			}
		})
	}
}
