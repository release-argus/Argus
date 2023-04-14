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
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_GetOption(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		optionRoot        *string
		optionMain        *string
		optionDefault     *string
		optionHardDefault *string
		wantString        string
	}{
		"root overrides all": {
			wantString:        "this",
			optionRoot:        stringPtr("this"),
			optionDefault:     stringPtr("not_this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			wantString:        "this",
			optionRoot:        nil,
			optionMain:        stringPtr("this"),
			optionDefault:     stringPtr("not_this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"default overrides hardDefault": {
			wantString:        "this",
			optionRoot:        nil,
			optionDefault:     stringPtr("this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"hardDefault is last resort": {
			wantString:        "this",
			optionRoot:        nil,
			optionDefault:     nil,
			optionHardDefault: stringPtr("this"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.optionRoot != nil {
				shoutrrr.Options[key] = *tc.optionRoot
			}
			if tc.optionMain != nil {
				shoutrrr.Main.Options[key] = *tc.optionMain
			}
			if tc.optionDefault != nil {
				shoutrrr.Defaults.Options[key] = *tc.optionDefault
			}
			if tc.optionHardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.optionHardDefault
			}

			// WHEN GetOption is called
			got := shoutrrr.GetOption(key)

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Fatalf("GetOption:\nwant: %q\ngot:  %q",
					tc.wantString, got)
			}

			// WHEN GetSelfOption is called
			got = shoutrrr.GetSelfOption(key)

			// THEN the function returns the Option in itself
			if got != util.DefaultIfNil(tc.optionRoot) {
				t.Fatalf("GetSelfOption:\nwant: %q\ngot:  %q",
					util.DefaultIfNil(tc.optionRoot), got)
			}

			// WHEN SetOption is called
			want := got + "-set-test"
			shoutrrr.SetOption(key, want)

			// THEN the Option is set and can be retrieved with a Get
			got = shoutrrr.GetSelfOption(key)
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
		optionRoot        *string
		optionMain        *string
		optionDefault     *string
		optionHardDefault *string
		wantString        string
	}{
		"root overrides all": {
			wantString:        "this",
			optionRoot:        stringPtr("this"),
			optionDefault:     stringPtr("not_this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			wantString:        "this",
			optionRoot:        nil,
			optionMain:        stringPtr("this"),
			optionDefault:     stringPtr("not_this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"default overrides hardDefault": {
			wantString:        "this",
			optionRoot:        nil,
			optionDefault:     stringPtr("this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"hardDefault is last resort": {
			wantString:        "this",
			optionRoot:        nil,
			optionDefault:     nil,
			optionHardDefault: stringPtr("this"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.optionRoot != nil {
				shoutrrr.URLFields[key] = *tc.optionRoot
			}
			if tc.optionMain != nil {
				shoutrrr.Main.URLFields[key] = *tc.optionMain
			}
			if tc.optionDefault != nil {
				shoutrrr.Defaults.URLFields[key] = *tc.optionDefault
			}
			if tc.optionHardDefault != nil {
				shoutrrr.HardDefaults.URLFields[key] = *tc.optionHardDefault
			}

			// WHEN GetURLField is called
			got := shoutrrr.GetURLField(key)

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Fatalf("GetURLField:\nwant: %q\ngot:  %q",
					tc.wantString, got)
			}

			// WHEN GetSelfURLField is called
			got = shoutrrr.GetSelfURLField(key)

			// THEN the function returns the URLField in itself
			if got != util.DefaultIfNil(tc.optionRoot) {
				t.Fatalf("GetSelfURLField:\nwant: %q\ngot:  %q",
					util.DefaultIfNil(tc.optionRoot), got)
			}

			// WHEN SetURLField is called
			want := got + "-set-test"
			shoutrrr.SetURLField(key, want)

			// THEN the URLField is set and can be retrieved with a Get
			got = shoutrrr.GetSelfURLField(key)
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
		optionRoot        *string
		optionMain        *string
		optionDefault     *string
		optionHardDefault *string
		wantString        string
	}{
		"root overrides all": {
			wantString:        "this",
			optionRoot:        stringPtr("this"),
			optionDefault:     stringPtr("not_this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"main overrides default and hardDefault": {
			wantString:        "this",
			optionRoot:        nil,
			optionMain:        stringPtr("this"),
			optionDefault:     stringPtr("not_this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"default overrides hardDefault": {
			wantString:        "this",
			optionRoot:        nil,
			optionDefault:     stringPtr("this"),
			optionHardDefault: stringPtr("not_this"),
		},
		"hardDefault is last resort": {
			wantString:        "this",
			optionRoot:        nil,
			optionDefault:     nil,
			optionHardDefault: stringPtr("this"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.optionRoot != nil {
				shoutrrr.Params[key] = *tc.optionRoot
			}
			if tc.optionMain != nil {
				shoutrrr.Main.Params[key] = *tc.optionMain
			}
			if tc.optionDefault != nil {
				shoutrrr.Defaults.Params[key] = *tc.optionDefault
			}
			if tc.optionHardDefault != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.optionHardDefault
			}

			// WHEN GetParam is called
			got := shoutrrr.GetParam(key)

			// THEN the function returns the correct result
			if got != tc.wantString {
				t.Fatalf("GetParam:\nwant: %q\ngot:  %q",
					tc.wantString, got)
			}

			// WHEN GetSelfParam is called
			got = shoutrrr.GetSelfParam(key)

			// THEN the function returns the Param in itself
			if got != util.DefaultIfNil(tc.optionRoot) {
				t.Fatalf("GetSelfParam:\nwant: %q\ngot:  %q",
					util.DefaultIfNil(tc.optionRoot), got)
			}

			// WHEN SetParam is called
			want := got + "-set-test"
			shoutrrr.SetParam(key, want)

			// THEN the Param is set and can be retrieved with a Get
			got = shoutrrr.GetSelfParam(key)
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
		delayRoot        *string
		delayMain        *string
		delayDefault     *string
		delayHardDefault *string
		wantString       string
	}{
		"root overrides all": {
			wantString:       "1s",
			delayRoot:        stringPtr("1s"),
			delayDefault:     stringPtr("2s"),
			delayHardDefault: stringPtr("2s"),
		},
		"main overrides default and hardDefault": {
			wantString:       "1s",
			delayRoot:        nil,
			delayMain:        stringPtr("1s"),
			delayDefault:     stringPtr("2s"),
			delayHardDefault: stringPtr("2s"),
		},
		"default overrides hardDefault": {
			wantString:       "1s",
			delayRoot:        nil,
			delayDefault:     stringPtr("1s"),
			delayHardDefault: stringPtr("2s"),
		},
		"hardDefault is last resort": {
			wantString:       "1s",
			delayRoot:        nil,
			delayDefault:     nil,
			delayHardDefault: stringPtr("1s"),
		},
		"no delay anywhere defaults to 0s": {wantString: "0s",
			delayRoot:        nil,
			delayDefault:     nil,
			delayHardDefault: nil},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "delay"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.delayRoot != nil {
				shoutrrr.Options[key] = *tc.delayRoot
			}
			if tc.delayMain != nil {
				shoutrrr.Main.Options[key] = *tc.delayMain
			}
			if tc.delayDefault != nil {
				shoutrrr.Defaults.Options[key] = *tc.delayDefault
			}
			if tc.delayHardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.delayHardDefault
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
		delayRoot        *string
		delayMain        *string
		delayDefault     *string
		delayHardDefault *string
		want             time.Duration
	}{
		"root overrides all": {
			want:             1 * time.Second,
			delayRoot:        stringPtr("1s"),
			delayDefault:     stringPtr("2s"),
			delayHardDefault: stringPtr("2s"),
		},
		"main overrides default and hardDefault": {
			want:             1 * time.Second,
			delayRoot:        nil,
			delayMain:        stringPtr("1s"),
			delayDefault:     stringPtr("2s"),
			delayHardDefault: stringPtr("2s"),
		},
		"default overrides hardDefault": {
			want:             1 * time.Second,
			delayRoot:        nil,
			delayDefault:     stringPtr("1s"),
			delayHardDefault: stringPtr("2s"),
		},
		"hardDefault is last resort": {
			want:             1 * time.Second,
			delayRoot:        nil,
			delayDefault:     nil,
			delayHardDefault: stringPtr("1s"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "delay"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.delayRoot != nil {
				shoutrrr.Options[key] = *tc.delayRoot
			}
			if tc.delayMain != nil {
				shoutrrr.Main.Options[key] = *tc.delayMain
			}
			if tc.delayDefault != nil {
				shoutrrr.Defaults.Options[key] = *tc.delayDefault
			}
			if tc.delayHardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.delayHardDefault
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
			maxTriesDefault:     stringPtr("2"),
			maxTriesHardDefault: stringPtr("2"),
		},
		"main overrides default and hardDefault": {
			want:                1,
			maxTriesRoot:        nil,
			maxTriesMain:        stringPtr("1"),
			maxTriesDefault:     stringPtr("2"),
			maxTriesHardDefault: stringPtr("2"),
		},
		"default overrides hardDefault": {
			want:                1,
			maxTriesRoot:        nil,
			maxTriesDefault:     stringPtr("1"),
			maxTriesHardDefault: stringPtr("2"),
		},
		"hardDefault is last resort": {
			want:                1,
			maxTriesRoot:        nil,
			maxTriesDefault:     nil,
			maxTriesHardDefault: stringPtr("1"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "max_tries"
			shoutrrr := testShoutrrr(false, true, false)
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

func TestShoutrrr_GetMessage(t *testing.T) {
	// GIVEN a Shoutrrr
	serviceInfo := &util.ServiceInfo{
		ID:            "release-argus/Argus",
		URL:           "https://github.com",
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := map[string]struct {
		messageRoot        *string
		messageMain        *string
		messageDefault     *string
		messageHardDefault *string
		want               string
	}{
		"root overrides all": {
			want:               "New version!",
			messageRoot:        stringPtr("New version!"),
			messageDefault:     stringPtr("something"),
			messageHardDefault: stringPtr("something"),
		},
		"main overrides default and hardDefault": {
			want:               "New version!",
			messageRoot:        nil,
			messageMain:        stringPtr("New version!"),
			messageDefault:     stringPtr("something"),
			messageHardDefault: stringPtr("something"),
		},
		"default overrides hardDefault": {
			want:               "New version!",
			messageRoot:        nil,
			messageDefault:     stringPtr("New version!"),
			messageHardDefault: stringPtr("something"),
		},
		"hardDefault is last resort": {
			want:               "New version!",
			messageRoot:        nil,
			messageDefault:     nil,
			messageHardDefault: stringPtr("New version!"),
		},
		"jinja templating": {
			want:               "New version!",
			messageRoot:        stringPtr("{% if 'a' == 'a' %}New version!{% endif %}"),
			messageDefault:     stringPtr("something"),
			messageHardDefault: stringPtr("something"),
		},
		"jinja vars": {
			want: fmt.Sprintf("%s or %s/%s/releases/tag/%s",
				serviceInfo.WebURL, serviceInfo.URL, serviceInfo.ID, serviceInfo.LatestVersion),
			messageRoot:        stringPtr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			messageDefault:     stringPtr("something"),
			messageHardDefault: stringPtr("something"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "message"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.messageRoot != nil {
				shoutrrr.Options[key] = *tc.messageRoot
			}
			if tc.messageMain != nil {
				shoutrrr.Main.Options[key] = *tc.messageMain
			}
			if tc.messageDefault != nil {
				shoutrrr.Defaults.Options[key] = *tc.messageDefault
			}
			if tc.messageHardDefault != nil {
				shoutrrr.HardDefaults.Options[key] = *tc.messageHardDefault
			}

			// WHEN GetMessage is called
			got := shoutrrr.GetMessage(serviceInfo)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Fatalf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestShoutrrr_GetTitle(t *testing.T) {
	// GIVEN a Shoutrrr
	serviceInfo := &util.ServiceInfo{
		ID:            "release-argus/Argus",
		URL:           "https://github.com",
		WebURL:        "https://release-argus.io/demo",
		LatestVersion: "0.9.0",
	}
	tests := map[string]struct {
		titleRoot        *string
		titleMain        *string
		titleDefault     *string
		titleHardDefault *string
		want             string
	}{
		"root overrides all": {
			want:             "New version!",
			titleRoot:        stringPtr("New version!"),
			titleDefault:     stringPtr("something"),
			titleHardDefault: stringPtr("something"),
		},
		"main overrides default and hardDefault": {
			want:             "New version!",
			titleRoot:        nil,
			titleMain:        stringPtr("New version!"),
			titleDefault:     stringPtr("something"),
			titleHardDefault: stringPtr("something"),
		},
		"default overrides hardDefault": {
			want:             "New version!",
			titleRoot:        nil,
			titleDefault:     stringPtr("New version!"),
			titleHardDefault: stringPtr("something"),
		},
		"hardDefault is last resort": {
			want:             "New version!",
			titleRoot:        nil,
			titleDefault:     nil,
			titleHardDefault: stringPtr("New version!"),
		},
		"jinja templating": {
			want:             "New version!",
			titleRoot:        stringPtr("{% if 'a' == 'a' %}New version!{% endif %}"),
			titleDefault:     stringPtr("something"),
			titleHardDefault: stringPtr("something"),
		},
		"jinja vars": {
			want: fmt.Sprintf("%s or %s/%s/releases/tag/%s",
				serviceInfo.WebURL, serviceInfo.URL, serviceInfo.ID, serviceInfo.LatestVersion),
			titleRoot:        stringPtr("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			titleDefault:     stringPtr("something"),
			titleHardDefault: stringPtr("something"),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			key := "title"
			shoutrrr := testShoutrrr(false, true, false)
			if tc.titleRoot != nil {
				shoutrrr.Params[key] = *tc.titleRoot
			}
			if tc.titleMain != nil {
				shoutrrr.Main.Params[key] = *tc.titleMain
			}
			if tc.titleDefault != nil {
				shoutrrr.Defaults.Params[key] = *tc.titleDefault
			}
			if tc.titleHardDefault != nil {
				shoutrrr.HardDefaults.Params[key] = *tc.titleHardDefault
			}

			// WHEN GetTitle is called
			got := shoutrrr.GetTitle(serviceInfo)

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
		typeRoot        string
		typeMain        string
		typeDefault     string
		typeHardDefault string
		want            string
	}{
		"root overrides all": {
			want:            "smtp",
			typeRoot:        "smtp",
			typeDefault:     "other",
			typeHardDefault: "other",
		},
		"main overrides default and hardDefault": {
			want:            "smtp",
			typeRoot:        "",
			typeMain:        "smtp",
			typeDefault:     "other",
			typeHardDefault: "other",
		},
		"default is ignored": {
			want:            "",
			typeRoot:        "",
			typeDefault:     "smtp",
			typeHardDefault: "",
		},
		"hardDefault is ignored": {
			want:            "",
			typeRoot:        "",
			typeDefault:     "",
			typeHardDefault: "smtp",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, true, false)
			shoutrrr.Type = tc.typeRoot
			shoutrrr.Main.Type = tc.typeMain
			shoutrrr.Defaults.Type = tc.typeDefault
			shoutrrr.HardDefaults.Type = tc.typeHardDefault

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
