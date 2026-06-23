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

package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestDrainChannel(t *testing.T) {
	// GIVEN: a channel with buffer size and values.
	tests := []struct {
		name   string
		buffer int
		values []int
	}{
		{
			name:   "empty buffered channel",
			buffer: 3,
			values: nil,
		},
		{
			name:   "partially-filled buffered channel",
			buffer: 5,
			values: []int{1, 2, 3},
		},
		{
			name:   "full buffered channel",
			buffer: 3,
			values: []int{10, 20, 30},
		},
		{
			name:   "unbuffered channel",
			buffer: 0,
			values: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ch := make(chan int, tc.buffer)
			// Push the values into the channel.
			for _, v := range tc.values {
				ch <- v
			}

			done := make(chan struct{})
			// WHEN: drainChannel is called.
			go func() {
				drainChannel(ch)
				close(done)
			}()

			// THEN: drainChannel must not block (even for unbuffered channels)
			select {
			case <-done:
				// success
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("%s\ndrainChannel blocked", packageName)
			}

			// AND: buffered channels should be fully drained.
			if tc.buffer > 0 {
				if got := len(ch); got != 0 {
					t.Fatalf(
						"%s\ndrainChannel() drain failed\ngot:  %d\nwant: 0",
						packageName, got,
					)
				}
			}
		})
	}
}

func TestDrainAndDebounce(t *testing.T) {
	DebounceDuration = 2 * time.Second
	// GIVEN: a Config.SaveChannel and messages to send/not send.
	tests := []struct {
		name      string
		messages  int
		timeTaken time.Duration
	}{
		{
			name:      "no messages",
			messages:  0,
			timeTaken: DebounceDuration,
		},
		{
			name:      "one message",
			messages:  1,
			timeTaken: 2 * DebounceDuration,
		},
		{
			name:      "two messages",
			messages:  2,
			timeTaken: 2 * DebounceDuration,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := testConfig(t)

			// WHEN: those messages are sent to the channel midway through the wait.
			go func() {
				for tc.messages != 0 {
					time.Sleep(4 * (DebounceDuration / 10))
					config.SaveChannel <- true
					tc.messages--
				}
			}()
			start := time.Now().UTC()
			drainAndDebounce(t.Context(), config.SaveChannel, DebounceDuration)

			// THEN: after `DebounceDuration`, it would have tried to Save.
			elapsed := time.Since(start)
			if elapsed < tc.timeTaken-100*time.Millisecond ||
				elapsed > tc.timeTaken+100*time.Millisecond {
				t.Errorf(
					"%s\nDrainAndDebounce waited %s, but should have waited at least %s",
					packageName, elapsed, tc.timeTaken,
				)
			}
		})
	}
}

func TestConfig_SaveHandler(t *testing.T) {
	originalDebounce := DebounceDuration
	DebounceDuration = 2 * time.Second
	t.Cleanup(func() { DebounceDuration = originalDebounce })

	tests := []struct {
		name      string
		cancelCtx bool
	}{
		{
			name:      "normal debounce",
			cancelCtx: false,
		},
		{
			name:      "cancelled",
			cancelCtx: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout and sharing log exitCodeChannel.
			_ = test.CaptureLog(t, logx.Default())

			// GIVEN: a SaveHandler running on a Config with a SaveChannel.
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			config := testConfig(t)

			done := make(chan struct{}, 1)
			go func() {
				config.SaveHandler(ctx)
				done <- struct{}{}
			}()
			now := time.Now().UTC()

			// WHEN: a message is sent to the SaveHandler.
			config.SaveChannel <- true

			// AND: the context optionally cancelled.
			if tc.cancelCtx {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}

			prefix := fmt.Sprintf("%s\nConfig SaveHandler", packageName)

			// THEN: it should exit in the expected time window
			timeout := (2 * DebounceDuration) - (DebounceDuration / 2)

			select {
			case <-done:
				t.Errorf(
					"%s didn't error, expected to on saving before returning",
					prefix,
				)
			case <-logx.ExitCodeChannel():
				elapsed := time.Since(now)
				if tc.cancelCtx {
					if elapsed >= DebounceDuration-500*time.Millisecond {
						t.Errorf(
							"%s waited %s, should have exited <=%s on cancellation",
							prefix, elapsed, DebounceDuration,
						)
					}
				} else if elapsed > timeout {
					t.Errorf(
						"%s did not exit within %s timeout",
						prefix, timeout,
					)
				}
			case <-time.After(timeout):
				t.Errorf(
					"%s did not exit within expected time (%s)",
					prefix, timeout,
				)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	// GIVEN: we have a bunch of files that want to be Saved.
	tests := []struct {
		name        string
		file        func(path string)
		preSaveFunc func(path string)
		openFile    func(string) (io.WriteCloser, error)
		corrections map[string]string
		fatal       bool
		stdoutRegex string
	}{
		{
			name: "config test",
			file: testYAML_config_test,
			corrections: map[string]string{
				"listen_port: 0\n":           "listen_port: '0'\n",
				"semantic_versioning: n\n":   "semantic_versioning: false\n",
				"interval: 123\n":            "interval: 123s\n",
				"delay: 2\n":                 "delay: 2s\n",
				"  EmptyServiceIsDeleted:\n": "",
			},
		},
		{
			name: "Argus",
			file: testYAML_Argus,
			corrections: map[string]string{
				"listen_port: 0\n": "listen_port: '0'\n",
			},
		},
		{
			name: "small config",
			file: testYAML_config_small,
			corrections: map[string]string{
				"settings:\n  data: {}\n  web: {}\n": "",
				"    options: {}\n":                  "",
				"    dashboard: {}\n":                "",
			},
		},
		{
			name: "unreadable file",
			file: testYAML_Argus,
			preSaveFunc: func(path string) {
				if err := os.Chmod(path, 0_222); err != nil {
					t.Fatalf(
						"%s\nFailed to chmod the file to 222: %v",
						packageName, err,
					)
				}
			},
			fatal:       true,
			stdoutRegex: `error opening`,
		},
		{
			name: "unwritable file",
			file: testYAML_Argus,
			preSaveFunc: func(path string) {
				if err := os.Chmod(path, 0_444); err != nil {
					t.Fatalf(
						"%s\nFailed to chmod the file to 444: %v",
						packageName, err,
					)
				}
			},
			stdoutRegex: `error opening`,
		},
		{
			name: "flush error",
			file: testYAML_config_test,
			openFile: func(_ string) (io.WriteCloser, error) {
				return &failWriter{}, nil
			},
			fatal:       true,
			stdoutRegex: `error flushing`,
		},
		{
			name: "write error",
			file: testYAML_config_large,
			openFile: func(_ string) (io.WriteCloser, error) {
				return &failAfterNBytesWriter{remaining: 10240}, nil
			},
			fatal:       true,
			stdoutRegex: `error writing`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			// Write the initial file and load the config from it.
			file := filepath.Join(t.TempDir(), "config.yml")
			tc.file(file)
			t.Log(file)
			originalData, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf(
					"%s\nfailed opening the file - %s",
					packageName, err,
				)
			}
			had := string(originalData)
			cfg := testLoadBasic(t, file)
			if tc.preSaveFunc != nil {
				tc.preSaveFunc(file)
			}
			if tc.openFile != nil {
				original := openSaveFile
				openSaveFile = tc.openFile
				t.Cleanup(func() { openSaveFile = original })
			}

			// WHEN: we Save it to a new location.
			t.Cleanup(func() { _ = os.Remove(cfg.File) })
			saveDone := make(chan struct{})
			go func() {
				loadMu.RLock()
				cfg.Save()
				loadMu.RUnlock()
				close(saveDone)
			}()
			gotExit := gotExitCodeDuringSave(saveDone, tc.fatal)

			prefix := fmt.Sprintf("%s\nConfig.Save()", packageName)

			// THEN: the stdout is expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
			if tc.fatal && !gotExit {
				t.Fatalf("%s didn't Fatal log to ExitCodeChannel", prefix)
			}
			if tc.stdoutRegex != "" {
				return
			}
			if gotExit {
				t.Fatalf("%s unexpected Fatal log to ExitCodeChannel", prefix)
			}

			// AND: it's the same as the original file.
			newData, err := os.ReadFile(cfg.File)
			for from := range tc.corrections {
				had = strings.ReplaceAll(had, from, tc.corrections[from])
			}
			if string(newData) != had {
				t.Errorf(
					"%s shouldn't change data in file\ngot:  %q\nwant: %q",
					prefix, string(newData), had,
				)
			}
		})
	}
}

func TestConfig_Save__encodeError(t *testing.T) {
	// GIVEN: a failing YAML encode.
	original := encodeConfigYAML
	encodeConfigYAML = func(w io.Writer, indent int, c *Config) error {
		return fmt.Errorf("encode failed")
	}
	t.Cleanup(func() { encodeConfigYAML = original })

	// AMD: valid config.
	file := filepath.Join(t.TempDir(), "config.yml")
	testYAML_config_small(file)
	cfg := testLoadBasic(t, file)

	releaseStdout := test.CaptureLog(t, logx.Default())

	// WHEN: Save is called.
	saveDone := make(chan struct{})
	go func() {
		cfg.Save()
		close(saveDone)
	}()
	gotExit := gotExitCodeDuringSave(saveDone, true)

	prefix := fmt.Sprintf("%s\nConfig.Save()", packageName)

	// THEN: the encode error is fatal.
	stdout := releaseStdout()
	if !util.RegexCheck(`error encoding config`, stdout) {
		t.Errorf(
			"%s stdout mismatch\ngot:  %q\nwant: %q",
			prefix, stdout, `error encoding config`,
		)
	}
	if !gotExit {
		t.Fatalf("%s didn't Fatal log to ExitCodeChannel", prefix)
	}
}

// failWriter is an io.WriteCloser whose Write always fails.
// Used to trigger the "error flushing" path in Save() when all data fits in
// bufio's buffer and the error surfaces at the final Flush().
type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated disk full")
}
func (f *failWriter) Close() error { return nil }

// failAfterNBytesWriter is an io.WriteCloser that succeeds for the first
// `remaining` bytes then fails. Used to trigger the "error writing" path in
// Save() when the config is large enough to overflow bufio's buffer mid-loop.
type failAfterNBytesWriter struct{ remaining int }

func (f *failAfterNBytesWriter) Write(p []byte) (int, error) {
	if f.remaining <= 0 {
		return 0, fmt.Errorf("simulated disk full")
	}
	if len(p) > f.remaining {
		n := f.remaining
		f.remaining = 0
		return n, fmt.Errorf("simulated disk full")
	}
	f.remaining -= len(p)
	return len(p), nil
}
func (f *failAfterNBytesWriter) Close() error { return nil }

func gotExitCodeDuringSave(saveDone <-chan struct{}, expectFatal bool) bool {
	gotExit := false
	for {
		select {
		case <-logx.ExitCodeChannel():
			gotExit = true
		case _, open := <-saveDone:
			if !open {
				select {
				case <-logx.ExitCodeChannel():
					gotExit = true
				default:
				}
				if !gotExit && expectFatal {
					select {
					case <-logx.ExitCodeChannel():
						gotExit = true
					case <-time.After(200 * time.Millisecond):
					}
				}
				return gotExit
			}
		}
	}
}

func TestConfig_ReorderYAML(t *testing.T) {
	// GIVEN: a YAML to sort with a certain order of services.
	tests := []struct {
		name  string
		order []string
		lines []string
		want  []string
	}{
		{
			name:  "empty input",
			order: nil,
			lines: nil,
			want:  nil,
		},
		{
			name:  "single service, no empty maps",
			order: nil,
			lines: []string{
				"service:",
				"  alpha:",
				"    name: Alpha",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
			},
			want: []string{
				"service:",
				"  alpha:",
				"    name: Alpha",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
			},
		},
		{
			name:  "don't remove empty notify or webhook under service",
			order: nil,
			lines: []string{
				"defaults:",
				"  notify:",
				"    discord:",
				"      url_fields:",
				"        token: test",
				"        webhookid: foo",
				"  webhook:",
				"    url: https://example.com",
				"    secret: foo",
				"service:",
				"  alpha:",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
				"    notify:",
				"      discord: {}",
				"    webhook:",
				"      URL: {}",
			},
			want: []string{
				"defaults:",
				"  notify:",
				"    discord:",
				"      url_fields:",
				"        token: test",
				"        webhookid: foo",
				"  webhook:",
				"    url: https://example.com",
				"    secret: foo",
				"service:",
				"  alpha:",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
				"    notify:",
				"      discord: {}",
				"    webhook:",
				"      URL: {}",
			},
		},
		{
			name:  "reorder services according to order",
			order: []string{"beta", "alpha"},
			lines: []string{
				"service:",
				"  alpha:",
				"    name: Alpha",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
				"  beta:",
				"    name: Beta",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
			},
			want: []string{
				"service:",
				"  beta:",
				"    name: Beta",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
				"  alpha:",
				"    name: Alpha",
				"    latest_version:",
				"      type: github",
				"      url: " + test.ArgusGitHubRepo,
			},
		},
		{
			name:  "nested empty maps removed recursively",
			order: []string{"beta", "Alpha"},
			lines: []string{
				"defaults:",
				"  notify:",
				"    discord:",
				"      url_fields:",
				"        token: test",
				"        webhookid: foo",
				"service:",
				"  Alpha:",
				"    notify:",
				"      discord: {}",
				"    options: {}",
				"    latest_version:",
				"    deployed_version:",
				"      url: example.com",
				"      headers: []",
				"  beta:",
				"    unknown: {}",
				"    deployed_version:",
				"      url: example.com",
				"      headers: []",
				"      basic_auth: {}",
			},
			want: []string{
				"defaults:",
				"  notify:",
				"    discord:",
				"      url_fields:",
				"        token: test",
				"        webhookid: foo",
				"service:",
				"  beta:",
				"    deployed_version:",
				"      url: example.com",
				"  Alpha:",
				"    notify:",
				"      discord: {}",
				"    deployed_version:",
				"      url: example.com",
			},
		},
		{
			name:  "deep nesting beyond initial parentKey capacity",
			order: []string{"alpha"},
			lines: []string{
				"service:",
				"  alpha:",
				"    latest_version:",
				"      type: url",
				"      url: https://example.com",
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
				"        url_fields:",
				"          host: example.com",
			},
			want: []string{
				"service:",
				"  alpha:",
				"    latest_version:",
				"      type: url",
				"      url: https://example.com",
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
				"        url_fields:",
				"          host: example.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg Config
			yamlStr := strings.Join(tc.lines, "\n")
			if err := cfg.Decode([]byte(yamlStr)); err != nil {
				t.Fatalf("%s\n%s", packageName, errfmt.FormatError(err))
			}
			cfg.Order = tc.order
			cfg.Settings.Indentation = 2

			// WHEN: reorderYAML is called.
			got := cfg.reorderYAML(tc.lines)

			// THEN: the resulting lines match the expected output.
			gotStr := strings.Join(got, "\n")
			wantStr := strings.Join(tc.want, "\n")
			if gotStr != wantStr {
				t.Errorf(
					"%s\nConfig.reorderYAML() value mismatch\ngot:  %q\nwant: %q",
					packageName, gotStr, wantStr,
				)
			}
		})
	}
}
