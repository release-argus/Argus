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

package config

import (
	"os"
	"strings"
	"testing"
	"time"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

var TIMEOUT time.Duration = 30 * time.Second

func TestConfig_SaveHandler(t *testing.T) {
	// GIVEN a message is sent to the SaveHandler
	config := testConfig()
	// Disable fatal panics.
	defer func() { _ = recover() }()
	go func() {
		*config.SaveChannel <- true
	}()

	// WHEN the SaveHandler is running for a Config with an inaccessible file
	config.SaveHandler()

	// THEN it should have panic'd after TIMEOUT and not reach this
	time.Sleep(TIMEOUT * time.Second)
	t.Errorf("Save should panic'd on inaccessible file location %q",
		config.File)
}

func TestWaitChannelTimeout(t *testing.T) {
	// GIVEN a Config.SaveChannel and messages to send/not send
	tests := map[string]struct {
		messages  int
		timeTaken time.Duration
	}{
		"no messages": {
			messages:  0,
			timeTaken: TIMEOUT,
		},
		"one message": {
			messages:  1,
			timeTaken: 2 * TIMEOUT,
		},
		"two messages": {
			messages:  2,
			timeTaken: 2 * TIMEOUT,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := testConfig()

			// WHEN those messages are sent to the channel mid-way through the wait
			go func() {
				for tc.messages != 0 {
					time.Sleep(10 * time.Second)
					*config.SaveChannel <- true
					tc.messages--
				}
			}()
			time.Sleep(time.Second)
			start := time.Now().UTC()
			waitChannelTimeout(config.SaveChannel)

			// THEN after `TIMEOUT`, it would have tried to Save
			elapsed := time.Since(start)
			if elapsed < tc.timeTaken-100*time.Millisecond ||
				elapsed > tc.timeTaken+100*time.Millisecond {
				t.Errorf("waitChannelTimeout should have waited atleast %s, but only waited %s",
					tc.timeTaken, elapsed)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	// GIVEN we have a bunch of files that want to be Save'd
	tests := map[string]struct {
		file        func(path string, t *testing.T)
		corrections map[string]string
	}{
		"config_test.yml": {
			file: testYAML_ConfigTest,
			corrections: map[string]string{
				"listen_port: 0\n":         "listen_port: \"0\"\n",
				"semantic_versioning: n\n": "semantic_versioning: false\n",
				"interval: 123\n":          "interval: 123s\n",
				"delay: 2\n":               "delay: 2s\n",
			}},
		"argus.yml": {
			file: testYAML_Argus,
			corrections: map[string]string{
				"listen_port: 0\n": "listen_port: \"0\"\n",
			}},
		"small-config.yml": {
			file: testYAML_SmallConfigTest,
			corrections: map[string]string{
				"settings:\n  data: {}\n  web: {}\n": "",
				"    options: {}\n":                  "",
				"    dashboard: {}\n":                "",
			}},
	}

	for name, tc := range tests {

		// Load here as it could DATA RACE with setting the JLog
		file := name
		tc.file(file, t)
		t.Log(file)
		originalData, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed opening the file for the data we were going to Save\n%s",
				err.Error())
		}
		had := string(originalData)
		config := testLoadBasic(file, t)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we Save it to a new location
			config.File += ".test"
			t.Cleanup(func() { os.Remove(config.File) })
			loadMutex.RLock()
			config.Save()
			loadMutex.RUnlock()

			// THEN it's the same as the original file
			failed := false
			newData, err := os.ReadFile(config.File)
			for from := range tc.corrections {
				had = strings.ReplaceAll(had, from, tc.corrections[from])
			}
			if string(newData) != had {
				failed = true
				t.Errorf("%q is different after Save. Got \n%s\nexpecting:\n%s",
					file, string(newData), had)
			}
			err = os.Remove(config.File)
			if err != nil {
				t.Fatal(err)
			}
			if failed {
				t.Fatal()
			}
			time.Sleep(time.Second)
		})
	}
}

func TestRemoveSection(t *testing.T) {
	// GIVEN a file as a string and a section to remove from it
	file := `
  foo:
    latest_version:
      type: url
      url: https://example.com
    notify:
      bish: {}
      bash: {}
      bosh: {}
    command:
      - ['echo' '"hello"']
      - ['ls', '-lah']`
	tests := map[string]struct {
		section     string
		indentation int
		aStart      int
		aEnd        int
		bStart      int
		bEnd        int
	}{
		"remove latest_version": {
			section:     "latest_version",
			indentation: 2,
			aStart:      0,
			aEnd:        1,
			bStart:      4,
			bEnd:        11,
		},
		"remove notify": {
			section:     "notify",
			indentation: 2,
			aStart:      0,
			aEnd:        4,
			bStart:      8,
			bEnd:        11,
		},
		"remove command": {
			section:     "command",
			indentation: 2,
			aStart:      0,
			aEnd:        8,
			bStart:      8,
			bEnd:        8,
		},
		"remove root": {
			section:     "foo",
			indentation: 1,
			aStart:      0,
			aEnd:        0,
			bStart:      0,
			bEnd:        0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lines := strings.Split(file, "\n")[1:]
			want := make([]string, (tc.aEnd-tc.aStart)+(tc.bEnd-tc.bStart))
			for i := 0; i < (tc.aEnd - tc.aStart); i++ {
				want[i] = lines[tc.aStart+i]
			}
			for i := 0; i < (tc.bEnd - tc.bStart); i++ {
				want[i+(tc.aEnd-tc.aStart)] = lines[tc.bStart+i]
			}

			// WHEN we remove that section
			removeSection(tc.section, &lines, uint8(2), tc.indentation)

			// THEN it's removed
			if len(want) != len(lines) {
				t.Fatalf("want %d lines\n%v\ngot %d\n%v",
					len(want), want, len(lines), lines)
			}
			for i := range want {
				if want[i] != lines[i] {
					t.Errorf("%d: want %q, got %q",
						i, want[i], lines[i])
				}
			}
		})
	}
}

func TestRemoveAllServiceDefaults(t *testing.T) {
	// GIVEN a file as a []string and services that may/may not be using defaults in it
	tests := map[string]struct {
		lines                  string
		services               *service.Slice
		currentOrderIndexStart []int
		currentOrderIndexEnd   []int
		serviceDefaults        service.Defaults
		rootNotify             shoutrrr.SliceDefaults
		rootWebHook            webhook.SliceDefaults
		want                   string
	}{
		"empty": {
			lines:    "",
			services: &service.Slice{},
			want:     "",
		},
		"service using defaults": {
			lines: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
`,
			want: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
`,
			services: &service.Slice{
				"alpha": &service.Service{}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Slice{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.SliceDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.SliceDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexStart: []int{0, 24},
			currentOrderIndexEnd:   []int{0, 35},
		},
		"service overriding defaults": {
			lines: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
		notify:
			foo:
				options:
					message: 123
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
`,
			want: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
		notify:
			foo:
				options:
					message: 123
			bar: {}
`,
			services: &service.Slice{
				"alpha": &service.Service{
					Notify: shoutrrr.Slice{
						"foo": shoutrrr.New(
							nil, "foo", nil, nil, "gotify", nil, nil, nil, nil),
						"bar": shoutrrr.New(
							nil, "bar", nil, nil, "gotify", nil, nil, nil, nil)}}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Slice{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.SliceDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.SliceDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexStart: []int{0, 24},
			currentOrderIndexEnd:   []int{0, 37},
		},
		"service not using defaults": {
			lines: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
		notify:
			boop:
				type: gotify
			zoop:
				type: slack
		command:
			- ["ls", "-lah"]
		webhook:
			bang:
				type: gitlab
			crash:
				type: github
		`,
			want: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
		notify:
			boop:
				type: gotify
			zoop:
				type: slack
		command:
			- ["ls", "-lah"]
		webhook:
			bang:
				type: gitlab
			crash:
				type: github
		`,
			services: &service.Slice{
				"alpha": &service.Service{
					Notify: shoutrrr.Slice{
						"boop": shoutrrr.New(
							nil, "boop", nil, nil, "gotify", nil, nil, nil, nil),
						"zoop": shoutrrr.New(
							nil, "zoop", nil, nil, "slack", nil, nil, nil, nil)},
					Command: command.Slice{
						{"ls", "-lah"}},
					WebHook: webhook.Slice{
						"bang": webhook.New(
							nil, nil, "", nil, nil, nil, nil, nil, "", nil, "gitlab", "", nil, nil, nil),
						"crash": webhook.New(
							nil, nil, "", nil, nil, nil, nil, nil, "", nil, "github", "", nil, nil, nil)},
				}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Slice{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.SliceDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.SliceDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexStart: []int{0, 24},
			currentOrderIndexEnd:   []int{0, 34},
		},
		"service using defaults, service overriding defaults and service not using defaults": {
			lines: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
	bravo:
		latest_version:
			url: release-argus/Argus
		notify:
			foo: {}
			bar: {}
		command:
			- ["ls", "-lah"]
		webhook:
			bash:
				type: gitlab
			bish:
				type: github
			bosh:
				type: gitlab
	charlie:
		latest_version:
			url: release-argus/Argus
		notify:
			boop:
				type: gotify
			zoop:
				type: slack
		command:
			- ["ls", "-lah", "/tmp"]
		webhook:
			bang:
				type: gitlab
			crash:
				type: github
`,
			want: `
defaults:
	service:
		notify:
			foo: {}
			bar: {}
		command:
			- ["echo", "hello"]
		webhook:
			bash: {}
			bish: {}
			bosh: {}
notify:
	foo:
		type: gotify
	bar:
		type: discord
webhook:
	bash:
		type: github
	bish:
		type: gitlab
	bosh:
		type: github
service:
	alpha:
		latest_version:
			url: release-argus/Argus
	bravo:
		latest_version:
			url: release-argus/Argus
		command:
			- ["ls", "-lah"]
		webhook:
			bash:
				type: gitlab
			bish:
				type: github
			bosh:
				type: gitlab
	charlie:
		latest_version:
			url: release-argus/Argus
		notify:
			boop:
				type: gotify
			zoop:
				type: slack
		command:
			- ["ls", "-lah", "/tmp"]
		webhook:
			bang:
				type: gitlab
			crash:
				type: github
`,
			services: &service.Slice{
				"alpha": &service.Service{},
				"bravo": &service.Service{
					Command: command.Slice{
						{"ls", "-lah"}},
					WebHook: webhook.Slice{
						"bash": webhook.New(
							nil, nil, "", nil, nil, nil, nil, nil, "", nil, "gitlab", "", nil, nil, nil),
						"bish": webhook.New(
							nil, nil, "", nil, nil, nil, nil, nil, "", nil, "github", "", nil, nil, nil),
						"bosh": webhook.New(
							nil, nil, "", nil, nil, nil, nil, nil, "", nil, "gitlab", "", nil, nil, nil)}},
				"charlie": &service.Service{
					Notify: shoutrrr.Slice{
						"boop": shoutrrr.New(
							nil, "boop", nil, nil, "gotify", nil, nil, nil, nil),
						"zoop": shoutrrr.New(
							nil, "zoop", nil, nil, "slack", nil, nil, nil, nil)},
					Command: command.Slice{
						{"ls", "-lah", "/tmp"}},
					WebHook: webhook.Slice{
						"bang": webhook.New(
							nil, nil, "", nil, nil, nil, nil, nil, "", nil, "gitlab", "", nil, nil, nil),
						"crash": webhook.New(
							nil, nil, "", nil, nil, nil, nil, nil, "", nil, "github", "", nil, nil, nil)}}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Slice{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.SliceDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.SliceDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexStart: []int{0, 24, 36, 51},
			currentOrderIndexEnd:   []int{0, 35, 50, 65},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			indentation := 4
			tc.lines = strings.TrimLeft(tc.lines, "\n")
			tc.lines = strings.ReplaceAll(tc.lines, "\t", strings.Repeat(" ", indentation))
			lines := strings.Split(tc.lines, "\n")
			currentOrder := util.SortedKeys(*tc.services)
			tc.want = strings.TrimLeft(tc.want, "\n")
			tc.want = strings.ReplaceAll(tc.want, "\t", strings.Repeat(" ", indentation))
			want := strings.Split(tc.want, "\n")
			// Init the Services with the defaults
			for _, s := range *tc.services {
				s.Init(
					&tc.serviceDefaults, &service.Defaults{},
					&tc.rootNotify, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
					&tc.rootWebHook, &webhook.WebHookDefaults{}, &webhook.WebHookDefaults{})
			}

			// WHEN we remove all the service defaults
			removeAllServiceDefaults(
				&lines,
				uint8(indentation),
				tc.services,
				&currentOrder,
				&tc.currentOrderIndexStart,
				&tc.currentOrderIndexEnd)

			// THEN they're removed
			if len(want) != len(lines) {
				t.Fatalf("want %d lines, got %d\nwant:\n%v\n---\ngot:\n%v",
					len(want), len(lines), want, lines)
			}
			failed := false
			for i := range want {
				if want[i] != lines[i] {
					failed = true
					t.Errorf("%d: tc.want %q, got %q",
						i, want[i], lines[i])
				}
			}
			if failed {
				t.Logf("want:\n%v\n\n---\ngot:\n%v", want, lines)
			}
		})
	}
}
