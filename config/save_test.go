// Copyright [2025] [Argus]
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
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
	"gopkg.in/yaml.v3"
)

func TestDrainChannel(t *testing.T) {
	// GIVEN a channel with buffer size and values.
	tests := map[string]struct {
		buffer int
		values []int
	}{
		"empty buffered channel": {
			buffer: 3,
			values: nil,
		},
		"partially-filled buffered channel": {
			buffer: 5,
			values: []int{1, 2, 3},
		},
		"full buffered channel": {
			buffer: 3,
			values: []int{10, 20, 30},
		},
		"unbuffered channel": {
			buffer: 0,
			values: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ch := make(chan int, tc.buffer)
			// Push the values into the channel.
			for _, v := range tc.values {
				ch <- v
			}

			done := make(chan struct{})
			// WHEN drainChannel is called.
			go func() {
				drainChannel(ch)
				close(done)
			}()

			// THEN drainChannel must not block (even for unbuffered channels)
			select {
			case <-done:
				// success
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("%s\ndrainChannel blocked",
					packageName)
			}
			// AND buffered channels should be fully drained.
			if tc.buffer > 0 {
				if got := len(ch); got != 0 {
					t.Fatalf("%s\ndrain failed\nwant: %d, got: %d",
						packageName, 0, got)
				}
			}
		})
	}
}

func TestDrainAndDebounce(t *testing.T) {
	DebounceDuration = 2 * time.Second
	// GIVEN a Config.SaveChannel and messages to send/not send.
	tests := map[string]struct {
		messages  int
		timeTaken time.Duration
	}{
		"no messages": {
			messages:  0,
			timeTaken: DebounceDuration,
		},
		"one message": {
			messages:  1,
			timeTaken: 2 * DebounceDuration,
		},
		"two messages": {
			messages:  2,
			timeTaken: 2 * DebounceDuration,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := testConfig()

			// WHEN those messages are sent to the channel mid-way through the wait.
			go func() {
				for tc.messages != 0 {
					time.Sleep(4 * (DebounceDuration / 10))
					config.SaveChannel <- true
					tc.messages--
				}
			}()
			start := time.Now().UTC()
			drainAndDebounce(t.Context(), config.SaveChannel, DebounceDuration)

			// THEN after `DebounceDuration`, it would have tried to Save.
			elapsed := time.Since(start)
			if elapsed < tc.timeTaken-100*time.Millisecond ||
				elapsed > tc.timeTaken+100*time.Millisecond {
				t.Errorf("%s\nshould have waited at least %s, but only waited %s",
					packageName, tc.timeTaken, elapsed)
			}
		})
	}
}

func TestConfig_SaveHandler(t *testing.T) {
	DebounceDuration = 2 * time.Second
	tests := map[string]struct {
		cancelCtx bool
	}{
		"normal debounce": {
			cancelCtx: false,
		},
		"cancelled": {
			cancelCtx: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// GIVEN a SaveHandler running on a Config with a SaveChannel.
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			config := testConfig()

			done := make(chan struct{}, 1)
			go func() {
				config.SaveHandler(ctx)
				done <- struct{}{}
			}()

			now := time.Now().UTC()
			// WHEN a message is sent to the SaveHandler.
			config.SaveChannel <- true
			// AND the context optionally cancelled.
			if tc.cancelCtx {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}

			// THEN it should exit in the expected time window
			timeout := DebounceDuration + 500*time.Millisecond
			if tc.cancelCtx {
				timeout = 1 * time.Second
			}

			select {
			case <-done:
				t.Errorf("%s\nSaveHandler expected to error on saving before returning, but didn't error first.",
					packageName)
			case <-logutil.ExitCodeChannel():
				elapsed := time.Since(now)
				if tc.cancelCtx {
					if elapsed >= DebounceDuration-500*time.Millisecond {
						t.Errorf("%sSaveHandler should have exited quickly on cancellation, but waited %s",
							packageName, elapsed)
					}
				} else if elapsed > timeout {
					t.Errorf("%s\nSaveHandler should have exited within expected time (%s)",
						packageName, timeout)
				}
			case <-time.After(timeout):
				t.Errorf("%s\nSaveHandler did not exit within expected time (%s)",
					packageName, timeout)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	// GIVEN we have a bunch of files that want to be Saved.
	tests := map[string]struct {
		file        func(path string)
		preSaveFunc func(path string)
		corrections map[string]string
		stdoutRegex string
	}{
		"config test": {
			file: testYAML_config_test,
			corrections: map[string]string{
				"listen_port: 0\n":           "listen_port: \"0\"\n",
				"semantic_versioning: n\n":   "semantic_versioning: false\n",
				"interval: 123\n":            "interval: 123s\n",
				"delay: 2\n":                 "delay: 2s\n",
				"  EmptyServiceIsDeleted:\n": "",
			},
		},
		"Argus": {
			file: testYAML_Argus,
			corrections: map[string]string{
				"listen_port: 0\n": "listen_port: \"0\"\n",
			},
		},
		"small config": {
			file: testYAML_SmallConfigTest,
			corrections: map[string]string{
				"settings:\n  data: {}\n  web: {}\n": "",
				"    options: {}\n":                  "",
				"    dashboard: {}\n":                "",
			},
		},
		"unreadable file": {
			file: testYAML_Argus,
			preSaveFunc: func(path string) {
				if err := os.Chmod(path, 0_222); err != nil {
					t.Fatalf("%s\nFailed to chmod the file to 222: %v",
						packageName, err)
				}
			},
			stdoutRegex: `error opening`,
		},
		"unwritable file": {
			file: testYAML_Argus,
			preSaveFunc: func(path string) {
				if err := os.Chmod(path, 0_444); err != nil {
					t.Fatalf("%s\nFailed to chmod the file to 444: %v",
						packageName, err)
				}
			},
			stdoutRegex: `error opening`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			// Write the initial file and load the config from it.
			file := filepath.Join(t.TempDir(), "config.yml")
			tc.file(file)
			t.Log(file)
			originalData, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("%s\nFailed opening the file for the data we were going to Save\n%s",
					packageName, err)
			}
			had := string(originalData)
			cfg := testLoadBasic(t, file)
			if tc.preSaveFunc != nil {
				tc.preSaveFunc(file)
			}

			// WHEN we Save it to a new location.
			t.Cleanup(func() { _ = os.Remove(cfg.File) })
			loadMutex.RLock()
			cfg.Save()
			loadMutex.RUnlock()

			// THEN the stdout is expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
					packageName, tc.stdoutRegex, stdout)
			}
			if tc.stdoutRegex != "" {
				drainAndDebounce(t.Context(), logutil.ExitCodeChannel(), 100*time.Millisecond)
				return
			}
			// AND it's the same as the original file.
			newData, err := os.ReadFile(cfg.File)
			for from := range tc.corrections {
				had = strings.ReplaceAll(had, from, tc.corrections[from])
			}
			if string(newData) != had {
				t.Errorf("%s\n%q is different after Save()\ngot:\n%s\nwant:\n%s\n\nwant: %q\ngot:  %q",
					packageName, file, string(newData), had, had, string(newData))
			}
			err = os.Remove(cfg.File)
			if err != nil {
				t.Errorf("%s\n%v",
					packageName, err)
			}
			time.Sleep(time.Second)
		})
	}
}

func TestConfig_ReorderYAML(t *testing.T) {
	// GIVEN a YAML to sort with a certain order of services.
	tests := map[string]struct {
		order []string
		lines []string
		want  []string
	}{
		"empty input": {
			order: nil,
			lines: nil,
			want:  nil,
		},
		"single service, no empty maps": {
			order: nil,
			lines: []string{
				"service:",
				"  alpha:",
				"    name: Alpha",
			},
			want: []string{
				"service:",
				"  alpha:",
				"    name: Alpha",
			},
		},
		"don't remove empty notify/webhook under service": {
			order: nil,
			lines: []string{
				"service:",
				"  alpha:",
				"    notify:",
				"      DISCORD: {}",
				"    webhook:",
				"      URL: {}",
			},
			want: []string{
				"service:",
				"  alpha:",
				"    notify:",
				"      DISCORD: {}",
				"    webhook:",
				"      URL: {}",
			},
		},
		"reorder services according to order": {
			order: []string{"beta", "alpha"},
			lines: []string{
				"service:",
				"  alpha:",
				"    name: Alpha",
				"  beta:",
				"    name: Beta",
			},
			want: []string{
				"service:",
				"  beta:",
				"    name: Beta",
				"  alpha:",
				"    name: Alpha",
			},
		},
		"nested empty maps removed recursively": {
			order: []string{"beta", "Alpha"},
			lines: []string{
				"service:",
				"  Alpha:",
				"    notify:",
				"      DISCORD: {}",
				"    options: {}",
				"    latest_version:",
				"    deployed_version:",
				"      url: example.com",
				"      headers: []",
				"  beta:",
				"    unknown: {}",
				"    deployed_version:",
				"      headers: []",
				"      basic_auth: {}",
			},
			want: []string{
				"service:",
				"  Alpha:",
				"    notify:",
				"      DISCORD: {}",
				"    deployed_version:",
				"      url: example.com",
			},
		},
		"deep nesting beyond initial parentKey capacity": {
			order: []string{"alpha"},
			lines: []string{
				"service:",
				"  alpha:",
				"    notify:",
				"      discord:",
				"        type: generic",
				"        options:",
				"          message: >-",
				"            {",
				"              \"username\": \"Test\",",
				"              \"embeds\": [",
				"                {",
				"                  \"title\": \"Release\",",
				"                  \"fields\": [",
				"                    {",
				"                      \"name\": \"Link\",",
				"                      \"value\": \"url\"",
				"                    }",
				"                  ]",
				"                }",
				"              ]",
				"            }",
			},
			want: []string{
				"service:",
				"  alpha:",
				"    notify:",
				"      discord:",
				"        type: generic",
				"        options:",
				"          message: >-",
				"            {",
				"              \"username\": \"Test\",",
				"              \"embeds\": [",
				"                {",
				"                  \"title\": \"Release\",",
				"                  \"fields\": [",
				"                    {",
				"                      \"name\": \"Link\",",
				"                      \"value\": \"url\"",
				"                    }",
				"                  ]",
				"                }",
				"              ]",
				"            }",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var cfg Config
			yamlStr := strings.Join(tc.lines, "\n")
			if err := yaml.Unmarshal([]byte(yamlStr), &cfg); err != nil {
				t.Fatalf("%s\nFailed to unmarshal YAML:\n%q\nerror:\n%s",
					packageName, yamlStr, err.Error())
			}
			cfg.Order = tc.order
			cfg.Settings.Indentation = 2
			//c := &Config{
			//	Service: service.Services{},
			//	Settings: Settings{
			//		Indentation: 2},
			//	Order: tc.order,
			//}

			// WHEN reorderYAML is called.
			got := cfg.reorderYAML(tc.lines)

			// THEN the resulting lines match the expected output.
			gotStr := strings.Join(got, "\n")
			wantStr := strings.Join(tc.want, "\n")
			if gotStr != wantStr {
				t.Errorf("%sreorderYAML() mismatch\ngot:\n%s\nwant:\n%s\n\nwant: %q\ngot:  %q",
					packageName, gotStr, wantStr, wantStr, gotStr)
			}
		})
	}
}

func TestRemoveSection(t *testing.T) {
	// GIVEN a file as a string and a section to remove from it.
	file := test.TrimYAML(`
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
				- ['ls', '-lah']`)
	tests := map[string]struct {
		section      string
		indentation  int
		aStart, aEnd int
		bStart, bEnd int
	}{
		"remove latest_version": {
			section:     "latest_version",
			indentation: 1,
			aStart:      0,
			aEnd:        1,
			bStart:      4,
			bEnd:        11,
		},
		"remove notify": {
			section:     "notify",
			indentation: 1,
			aStart:      0,
			aEnd:        4,
			bStart:      8,
			bEnd:        11,
		},
		"remove command": {
			section:     "command",
			indentation: 1,
			aStart:      0,
			aEnd:        8,
			bStart:      8,
			bEnd:        8,
		},
		"remove root": {
			section:     "foo",
			indentation: 0,
			aStart:      0,
			aEnd:        0,
			bStart:      0,
			bEnd:        0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lines := strings.Split(file, "\n")
			want := make([]string, (tc.aEnd-tc.aStart)+(tc.bEnd-tc.bStart))
			for i := 0; i < (tc.aEnd - tc.aStart); i++ {
				want[i] = lines[tc.aStart+i]
			}
			for i := 0; i < (tc.bEnd - tc.bStart); i++ {
				want[i+(tc.aEnd-tc.aStart)] = lines[tc.bStart+i]
			}

			// WHEN we remove that section.
			removeSection(tc.section, &lines, uint8(2), tc.indentation)

			// THEN it's removed.
			if len(lines) != len(want) {
				t.Fatalf("%s\nwant %d lines\n%v\ngot %d\n%v",
					packageName, len(want), want, len(lines), lines)
			}
			for i := range want {
				if lines[i] != want[i] {
					t.Errorf("%s\n%d: want %q, got %q",
						packageName, i, want[i], lines[i])
				}
			}
		})
	}
}

func TestRemoveAllServiceDefaults(t *testing.T) {
	// GIVEN a file as a []string and services that may/may not be using defaults in it.
	tests := map[string]struct {
		lines                  string
		services               *service.Services
		currentOrderIndexStart []int
		currentOrderIndexEnd   []int
		serviceDefaults        service.Defaults
		rootNotify             shoutrrr.ShoutrrrsDefaults
		rootWebHook            webhook.WebHooksDefaults
		want                   string
	}{
		"empty": {
			lines:    "",
			services: &service.Services{},
			want:     "",
		},
		"service using defaults": {
			lines: test.TrimYAML(`
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
			`),
			want: test.TrimYAML(`
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
			`),
			services: &service.Services{
				"alpha": &service.Service{}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Commands{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.ShoutrrrsDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.WebHooksDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexStart: []int{24},
			currentOrderIndexEnd:   []int{35},
		},
		"service overriding defaults": {
			lines: test.TrimYAML(`
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
			`),
			want: test.TrimYAML(`
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
			`),
			services: &service.Services{
				"alpha": &service.Service{
					Notify: shoutrrr.Shoutrrrs{
						"foo": shoutrrr.New(
							nil,
							"foo", "gotify",
							nil, nil, nil,
							nil, nil, nil),
						"bar": shoutrrr.New(
							nil,
							"bar", "gotify",
							nil, nil, nil,
							nil, nil, nil)}}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Commands{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.ShoutrrrsDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.WebHooksDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexStart: []int{24},
			currentOrderIndexEnd:   []int{37},
		},
		"service not using defaults": {
			lines: test.TrimYAML(`
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
							bop:
								type: gotify
							top:
								type: slack
						command:
							- ["ls", "-lah"]
						webhook:
							bang:
								type: gitlab
							crash:
								type: github
			`),
			want: test.TrimYAML(`
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
							bop:
								type: gotify
							top:
								type: slack
						command:
							- ["ls", "-lah"]
						webhook:
							bang:
								type: gitlab
							crash:
								type: github
			`),
			services: &service.Services{
				"alpha": &service.Service{
					Notify: shoutrrr.Shoutrrrs{
						"bop": shoutrrr.New(
							nil,
							"bop", "gotify",
							nil, nil, nil,
							nil, nil, nil),
						"top": shoutrrr.New(
							nil,
							"top", "slack",
							nil, nil, nil,
							nil, nil, nil)},
					Command: command.Commands{
						{"ls", "-lah"}},
					WebHook: webhook.WebHooks{
						"bang": webhook.New(
							nil, nil,
							"",
							nil, nil,
							"bang",
							nil, nil, nil,
							"", nil,
							"gitlab", "",
							nil, nil, nil),
						"crash": webhook.New(
							nil, nil,
							"",
							nil, nil,
							"crash",
							nil, nil, nil,
							"",
							nil,
							"github", "",
							nil, nil, nil)},
				}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Commands{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.ShoutrrrsDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.WebHooksDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexStart: []int{24},
			currentOrderIndexEnd:   []int{34},
		},
		"service using defaults, service overriding defaults and service not using defaults": {
			lines: test.TrimYAML(`
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
							bop:
								type: gotify
							top:
								type: slack
						command:
							- ["ls", "-lah", "/tmp"]
						webhook:
							bang:
								type: gitlab
							crash:
								type: github
			`),
			want: test.TrimYAML(`
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
							bop:
								type: gotify
							top:
								type: slack
						command:
							- ["ls", "-lah", "/tmp"]
						webhook:
							bang:
								type: gitlab
							crash:
								type: github
			`),
			services: &service.Services{
				"alpha": &service.Service{},
				"bravo": &service.Service{
					Command: command.Commands{
						{"ls", "-lah"}},
					WebHook: webhook.WebHooks{
						"bash": webhook.New(
							nil, nil,
							"",
							nil, nil,
							"bash",
							nil, nil, nil,
							"",
							nil,
							"gitlab", "",
							nil, nil, nil),
						"bish": webhook.New(
							nil, nil,
							"", nil, nil,
							"bish",
							nil, nil, nil,
							"",
							nil,
							"github", "",
							nil, nil, nil),
						"bosh": webhook.New(
							nil, nil,
							"",
							nil, nil,
							"bosh",
							nil, nil, nil,
							"",
							nil,
							"gitlab", "",
							nil, nil, nil)}},
				"charlie": &service.Service{
					Notify: shoutrrr.Shoutrrrs{
						"bop": shoutrrr.New(
							nil,
							"bop", "gotify",
							nil, nil, nil,
							nil, nil, nil),
						"top": shoutrrr.New(
							nil,
							"top", "slack",
							nil, nil, nil,
							nil, nil, nil)},
					Command: command.Commands{
						{"ls", "-lah", "/tmp"}},
					WebHook: webhook.WebHooks{
						"bang": webhook.New(
							nil, nil,
							"",
							nil, nil,
							"bang",
							nil, nil, nil,
							"",
							nil,
							"gitlab", "",
							nil, nil, nil),
						"crash": webhook.New(
							nil, nil,
							"",
							nil, nil,
							"crash",
							nil, nil, nil,
							"",
							nil,
							"github", "",
							nil, nil, nil)}}},
			serviceDefaults: service.Defaults{
				Notify: map[string]struct{}{
					"foo": {},
					"bar": {}},
				Command: command.Commands{
					{"echo", "hello"}},
				WebHook: map[string]struct{}{
					"bash": {},
					"bish": {},
					"bosh": {}}},
			rootNotify: shoutrrr.ShoutrrrsDefaults{
				"foo": shoutrrr.NewDefaults(
					"gotify", nil, nil, nil),
				"bar": shoutrrr.NewDefaults(
					"discord", nil, nil, nil)},
			rootWebHook: webhook.WebHooksDefaults{
				"bash": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", ""),
				"bish": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "gitlab", ""),
				"bosh": webhook.NewDefaults(
					nil, nil, "", nil, nil, "", nil, "github", "")},
			currentOrderIndexEnd:   []int{35, 50, 65},
			currentOrderIndexStart: []int{24, 36, 51},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.lines = strings.TrimLeft(tc.lines, "\n")
			indentationRegex := regexp.MustCompile(`(\n\s+)`)
			indentationStr := indentationRegex.FindString(tc.lines)
			indentation := strings.Count(indentationStr, " ")
			tc.lines = strings.ReplaceAll(tc.lines, "\t", strings.Repeat(" ", indentation))
			lines := strings.Split(tc.lines, "\n")

			currentOrder := util.SortedKeys(*tc.services)
			tc.want = strings.TrimLeft(tc.want, "\n")
			tc.want = strings.ReplaceAll(tc.want, "\t", strings.Repeat(" ", indentation))
			want := strings.Split(tc.want, "\n")
			// Init the Services with the defaults.
			for _, s := range *tc.services {
				s.Init(
					&tc.serviceDefaults, &service.Defaults{},
					&tc.rootNotify, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrr.ShoutrrrsDefaults{},
					&tc.rootWebHook, &webhook.Defaults{}, &webhook.Defaults{})
			}

			// WHEN we remove all the service defaults.
			removeAllServiceDefaults(
				&lines,
				uint8(indentation),
				tc.services,
				&currentOrder,
				&tc.currentOrderIndexStart,
				&tc.currentOrderIndexEnd)

			// THEN they're removed.
			if len(lines) != len(want) {
				t.Fatalf("%s\nwant: %d lines\ngot:  %d lines\nwant: %v\n---\ngot:  %v",
					packageName, len(want), len(lines), want, lines)
			}
			failed := false
			for i := range want {
				if lines[i] != want[i] {
					failed = true
					t.Errorf("%s\nline %d: tc.want %q, got %q",
						packageName, i, want[i], lines[i])
				}
			}
			if failed {
				t.Logf("%s\nwant:\n%v\n\n---\ngot:\n%v",
					packageName, want, lines)
			}
		})
	}
}
