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
	"maps"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

func TestShoutrrr_GetType(t *testing.T) {
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue string
		nilMain                                              bool
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
			name:         "default is ignored",
			want:         "test", // GetType falls back to the ID.
			defaultValue: "smtp",
		},
		{
			name:             "hardDefaultValue is ignored",
			want:             "test", // GetType falls back to the ID.
			hardDefaultValue: "smtp",
		},
		{
			name:    "nil Main",
			want:    "test", // GetType falls back to the ID.
			nilMain: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.Type = tc.rootValue
			shoutrrr.Main.Type = tc.mainValue
			shoutrrr.Defaults.Type = tc.defaultValue
			shoutrrr.HardDefaults.Type = tc.hardDefaultValue
			if tc.nilMain {
				shoutrrr.Main = nil
			}

			got := shoutrrr.GetType()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestShoutrrr_Title(t *testing.T) {
	testTemplatedLayers(
		t,
		"title", func(b *Base) MapStringStringOmitNull { return b.Params },
		(*Shoutrrr).Title,
	)
}

func TestShoutrrr_Message(t *testing.T) {
	testTemplatedLayers(
		t,
		"message", func(b *Base) MapStringStringOmitNull { return b.Options },
		(*Shoutrrr).Message,
	)
}

// testTemplatedLayers checks that a templated getter reads the first layer that sets key,
// root -> main -> defaults -> hardDefaults, and renders it with the ServiceInfo.
func testTemplatedLayers(
	t *testing.T,
	key string, layer func(*Base) MapStringStringOmitNull,
	call func(*Shoutrrr, serviceinfo.ServiceInfo) string,
) {
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
			rootValue:        new("New version!"),
			mainValue:        new("something"),
			defaultValue:     new("something"),
			hardDefaultValue: new("something"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "New version!",
			mainValue:        new("New version!"),
			defaultValue:     new("something"),
			hardDefaultValue: new("something"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "New version!",
			defaultValue:     new("New version!"),
			hardDefaultValue: new("something"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "New version!",
			hardDefaultValue: new("New version!"),
		},
		{
			name:             "django templating",
			want:             "New version!",
			rootValue:        new("{% if 'a' == 'a' %}New version!{% endif %}"),
			defaultValue:     new("something"),
			hardDefaultValue: new("something"),
		},
		{
			name: "django vars",
			want: fmt.Sprintf(
				"%s or %s/%s/releases/tag/%s",
				svcInfo.WebURL, svcInfo.URL, svcInfo.ID, svcInfo.LatestVersion,
			),
			rootValue:        new("{{ web_url }} or {{ service_url }}/{{ service_id }}/releases/tag/{{ version }}"),
			defaultValue:     new("something"),
			hardDefaultValue: new("something"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			setLayers(
				shoutrrr, key, layer,
				tc.rootValue, tc.mainValue, tc.defaultValue, tc.hardDefaultValue,
			)

			got := call(shoutrrr, svcInfo)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestShoutrrr_GetOption(t *testing.T) {
	testGetSetLayers(
		t,
		"OPTION", func(b *Base) MapStringStringOmitNull { return b.Options },
		(*Shoutrrr).GetOption, (*Shoutrrr).setOption,
	)
}

func TestShoutrrr_GetURLField(t *testing.T) {
	testGetSetLayers(
		t,
		"URL_FIELD", func(b *Base) MapStringStringOmitNull { return b.URLFields },
		(*Shoutrrr).GetURLField, (*Shoutrrr).setURLField,
	)
}

func TestShoutrrr_GetParam(t *testing.T) {
	testGetSetLayers(
		t,
		"PARAM", func(b *Base) MapStringStringOmitNull { return b.Params },
		(*Shoutrrr).GetParam, (*Shoutrrr).setParam,
	)
}

// testGetSetLayers checks that get reads the first layer that sets a key,
// root -> main -> defaults -> hardDefaults, expanding env vars, and that set writes the root.
func testGetSetLayers(
	t *testing.T,
	envKey string, layer func(*Base) MapStringStringOmitNull,
	get func(*Shoutrrr, string) string, set func(*Shoutrrr, string, string),
) {
	envVar := func(suffix string) string { return "TEST_SHOUTRRR__GET_" + envKey + "__" + suffix }

	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "this",
			rootValue:        new("this"),
			defaultValue:     new("not_this"),
			hardDefaultValue: new("not_this"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "this",
			mainValue:        new("this"),
			defaultValue:     new("not_this"),
			hardDefaultValue: new("not_this"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "this",
			defaultValue:     new("this"),
			hardDefaultValue: new("not_this"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "this",
			hardDefaultValue: new("this"),
		},
		{
			name:      "env var is used",
			want:      "this",
			env:       map[string]string{envVar("ONE"): "this"},
			rootValue: new("${" + envVar("ONE") + "}"),
		},
		{
			name:      "env var partial is used",
			want:      "this",
			env:       map[string]string{envVar("TWO"): "is"},
			rootValue: new("th${" + envVar("TWO") + "}"),
		},
		{
			name:         "empty env var is ignored",
			want:         "that",
			env:          map[string]string{envVar("THREE"): ""},
			rootValue:    new("${" + envVar("THREE") + "}"),
			defaultValue: new("that"),
		},
		{
			name:         "undefined env var is used",
			want:         "${" + envVar("UNSET") + "}",
			rootValue:    new("${" + envVar("UNSET") + "}"),
			defaultValue: new("that"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key := "test"
			shoutrrr := testShoutrrr(false, false)
			setLayers(
				shoutrrr, key, layer,
				tc.rootValue, tc.mainValue, tc.defaultValue, tc.hardDefaultValue,
			)
			test.IsolateEnv(t, tc.env)

			got := get(shoutrrr, key)

			if got != tc.want {
				t.Fatalf("before set: get() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			want := got + "-set-test"
			set(shoutrrr, key, want)

			// set writes the root layer.
			got = get(shoutrrr, key)
			if got != want {
				t.Fatalf("after set: get() mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

// setLayers sets key in each layer of the Shoutrrr whose value is non-nil.
func setLayers(
	shoutrrr *Shoutrrr,
	key string, layer func(*Base) MapStringStringOmitNull,
	rootValue, mainValue, defaultValue, hardDefaultValue *string,
) {
	for base, value := range map[*Base]*string{
		&shoutrrr.Base:              rootValue,
		&shoutrrr.Main.Base:         mainValue,
		&shoutrrr.Defaults.Base:     defaultValue,
		&shoutrrr.HardDefaults.Base: hardDefaultValue,
	} {
		if value != nil {
			layer(base)[key] = *value
		}
	}
}

func TestShoutrrr_GetDelay(t *testing.T) {
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 string
	}{
		{
			name:             "root overrides all",
			want:             "1s",
			rootValue:        new("1s"),
			mainValue:        new("2s"),
			defaultValue:     new("2s"),
			hardDefaultValue: new("2s"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             "1s",
			mainValue:        new("1s"),
			defaultValue:     new("2s"),
			hardDefaultValue: new("2s"),
		},
		{
			name:             "default overrides hardDefault",
			want:             "1s",
			defaultValue:     new("1s"),
			hardDefaultValue: new("2s"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             "1s",
			hardDefaultValue: new("1s"),
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

			shoutrrr := testShoutrrr(false, false)
			setLayers(
				shoutrrr, "delay", func(b *Base) MapStringStringOmitNull { return b.Options },
				tc.rootValue, tc.mainValue, tc.defaultValue, tc.hardDefaultValue,
			)

			got := shoutrrr.GetDelay()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestShoutrrr_GetDelayDuration(t *testing.T) {
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 time.Duration
	}{
		{
			name:             "root overrides all",
			want:             1 * time.Second,
			rootValue:        new("1s"),
			mainValue:        new("2s"),
			defaultValue:     new("2s"),
			hardDefaultValue: new("2s"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             1 * time.Second,
			mainValue:        new("1s"),
			defaultValue:     new("2s"),
			hardDefaultValue: new("2s"),
		},
		{
			name:             "default overrides hardDefault",
			want:             1 * time.Second,
			defaultValue:     new("1s"),
			hardDefaultValue: new("2s"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             1 * time.Second,
			hardDefaultValue: new("1s"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			setLayers(
				shoutrrr, "delay", func(b *Base) MapStringStringOmitNull { return b.Options },
				tc.rootValue, tc.mainValue, tc.defaultValue, tc.hardDefaultValue,
			)

			got := shoutrrr.GetDelayDuration()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %s\nwant: %s", got, tc.want)
			}
		})
	}
}

func TestShoutrrr_GetMaxTries(t *testing.T) {
	tests := []struct {
		name                                                 string
		rootValue, mainValue, defaultValue, hardDefaultValue *string
		want                                                 int
	}{
		{
			name:             "root overrides all",
			want:             1,
			rootValue:        new("1"),
			mainValue:        new("2"),
			defaultValue:     new("2"),
			hardDefaultValue: new("2"),
		},
		{
			name:             "main overrides default and hardDefault",
			want:             1,
			mainValue:        new("1"),
			defaultValue:     new("2"),
			hardDefaultValue: new("2"),
		},
		{
			name:             "default overrides hardDefault",
			want:             1,
			defaultValue:     new("1"),
			hardDefaultValue: new("2"),
		},
		{
			name:             "hardDefaultValue is last resort",
			want:             1,
			hardDefaultValue: new("1"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			setLayers(
				shoutrrr, "max_tries", func(b *Base) MapStringStringOmitNull { return b.Options },
				tc.rootValue, tc.mainValue, tc.defaultValue, tc.hardDefaultValue,
			)

			got := shoutrrr.GetMaxTries()

			if int(got) != tc.want {
				t.Fatalf("mismatch\ngot:  %d\nwant: %d", got, tc.want)
			}
		})
	}
}

func TestBase_GetSetOption(t *testing.T) {
	testBaseAccessors(
		t,
		func(b *Base) MapStringStringOmitNull { return b.Options },
		(*Base).getOption, (*Base).setOption, nil,
	)
}

func TestBase_GetSetDeleteURLField(t *testing.T) {
	testBaseAccessors(
		t,
		func(b *Base) MapStringStringOmitNull { return b.URLFields },
		(*Base).getURLField, (*Base).setURLField, (*Base).deleteURLField,
	)
}

func TestBase_GetSetDeleteParam(t *testing.T) {
	testBaseAccessors(
		t,
		func(b *Base) MapStringStringOmitNull { return b.Params },
		(*Base).GetParam, (*Base).setParam, (*Base).deleteParam,
	)
}

// testBaseAccessors checks that get reads key from field, that set overwrites it,
// and that del, when non-nil, removes it.
func testBaseAccessors(
	t *testing.T,
	field func(*Base) MapStringStringOmitNull,
	get func(*Base, string) string, set func(*Base, string, string), del func(*Base, string),
) {
	values := map[string]string{
		"a":  "foo",
		"b":  "bar",
		"hi": "bye",
	}
	tests := []struct {
		name      string
		key, want string
	}{
		{name: "unknown key", key: "foo", want: ""},
		{name: "value 1", key: "a", want: "foo"},
		{name: "value 2", key: "b", want: "bar"},
		{name: "value 3", key: "hi", want: "bye"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base := Base{}
			base.InitMaps()
			maps.Copy(field(&base), values)

			got := get(&base, tc.key)

			if got != tc.want {
				t.Fatalf("before set: get() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			want := got + "-set-test"
			set(&base, tc.key, want)

			got = get(&base, tc.key)
			if got != want {
				t.Fatalf("after set: get() mismatch\ngot:  %q\nwant: %q", got, want)
			}

			if del == nil {
				return
			}
			del(&base, tc.key)

			if _, found := field(&base)[tc.key]; found {
				t.Fatalf("after del: field()[%q] mismatch\ngot:  present\nwant: absent", tc.key)
			}
		})
	}
}
