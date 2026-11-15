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
	"errors"
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
			for _, v := range tc.values {
				ch <- v
			}

			done := make(chan struct{})

			go func() {
				drainChannel(ch)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timed out waiting for drainChannel() to return")
			}

			if tc.buffer > 0 {
				if got, want := len(ch), 0; got != want {
					t.Fatalf("len(ch) mismatch\ngot:  %d\nwant: %d", got, want)
				}
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
			// Serial: captures stdout.

			_ = test.CaptureLog(t, logx.Default())

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			config := testConfig(t)

			done := make(chan struct{}, 1)
			go func() {
				config.SaveHandler(ctx)
				done <- struct{}{}
			}()
			now := time.Now().UTC()

			config.SaveChannel <- true

			if tc.cancelCtx {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}

			timeout := (2 * DebounceDuration) - (DebounceDuration / 2)

			// testConfig cannot be saved, so the handler exits fatally.
			select {
			case <-done:
				t.Error("SaveHandler() returned before logx.ExitCodeChannel() received a message")
			case <-logx.ExitCodeChannel():
				elapsed := time.Since(now)
				if tc.cancelCtx {
					if maxWait := DebounceDuration - 500*time.Millisecond; elapsed >= maxWait {
						t.Errorf("elapsed mismatch\ngot:  %s\nwant: < %s", elapsed, maxWait)
					}
				} else if elapsed > timeout {
					t.Errorf("elapsed mismatch\ngot:  %s\nwant: <= %s", elapsed, timeout)
				}
			case <-time.After(timeout):
				t.Error("timed out waiting for logx.ExitCodeChannel()")
			}
		})
	}
}

func TestDrainAndDebounce(t *testing.T) {
	originalDebounce := DebounceDuration
	DebounceDuration = 2 * time.Second
	t.Cleanup(func() { DebounceDuration = originalDebounce })

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

			go func() {
				for tc.messages != 0 {
					time.Sleep(4 * (DebounceDuration / 10))
					config.SaveChannel <- true
					tc.messages--
				}
			}()
			start := time.Now().UTC()
			drainAndDebounce(t.Context(), config.SaveChannel, DebounceDuration)

			// One debounce, plus one more if messages arrived during the first.
			elapsed := time.Since(start)
			if elapsed < tc.timeTaken-100*time.Millisecond ||
				elapsed > tc.timeTaken+100*time.Millisecond {
				t.Errorf("elapsed mismatch\ngot:  %s\nwant: %s (±100ms)", elapsed, tc.timeTaken)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	tests := []struct {
		name        string
		file        func(t *testing.T, path string)
		preSaveFunc func(t *testing.T, path string)
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
			name: "Forgejo hosts are unchanged by save",
			file: testYAML_config_forgejo,
			corrections: map[string]string{
				"settings:\n  data: {}\n  web: {}\n": "",
				"    options: {}\n":                  "",
				"    dashboard: {}\n":                "",
			},
		},
		{
			name: "Forgejo host defaults survive a save",
			file: testYAML_config_forgejo_defaults,
			corrections: map[string]string{
				"settings:\n  data: {}\n  web: {}\n": "",
				"    options: {}\n":                  "",
				"    dashboard: {}\n":                "",
			},
		},
		{
			name: "settings.auth",
			file: testYAML_config_auth,
			corrections: map[string]string{
				"    options: {}\n":   "",
				"    dashboard: {}\n": "",
			},
		},
		{
			name: "unreadable file",
			file: testYAML_Argus,
			preSaveFunc: func(t *testing.T, path string) {
				if err := os.Chmod(path, 0_222); err != nil {
					t.Fatalf("os.Chmod() failed: %v", err)
				}
			},
			fatal:       true,
			stdoutRegex: `failed to open`,
		},
		{
			name: "unwritable file",
			file: testYAML_Argus,
			preSaveFunc: func(t *testing.T, path string) {
				if err := os.Chmod(path, 0_444); err != nil {
					t.Fatalf("os.Chmod() failed: %v", err)
				}
			},
			fatal:       true,
			stdoutRegex: `failed to open`,
		},
		{
			name: "flush error",
			file: testYAML_config_test,
			openFile: func(_ string) (io.WriteCloser, error) {
				return &failWriter{}, nil
			},
			fatal:       true,
			stdoutRegex: `failed to flush`,
		},
		{
			name: "write error",
			file: testYAML_config_large,
			openFile: func(_ string) (io.WriteCloser, error) {
				return &failAfterNBytesWriter{remaining: 10240}, nil
			},
			fatal:       true,
			stdoutRegex: `failed to write`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout and swaps openSaveFile.

			releaseStdout := test.CaptureLog(t, logx.Default())

			file := filepath.Join(t.TempDir(), "config.yml")
			tc.file(t, file)
			originalData, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("original: os.ReadFile() failed: %v", err)
			}
			had := string(originalData)
			cfg := testLoadBasic(t, file)
			if tc.preSaveFunc != nil {
				tc.preSaveFunc(t, file)
			}
			if tc.openFile != nil {
				original := openSaveFile
				openSaveFile = tc.openFile
				t.Cleanup(func() { openSaveFile = original })
			}

			t.Cleanup(func() { _ = os.Remove(cfg.File) })
			saveDone := make(chan struct{})
			go func() {
				loadMu.RLock()
				cfg.Save()
				loadMu.RUnlock()
				close(saveDone)
			}()
			gotExit := gotExitCodeDuringSave(saveDone, tc.fatal)

			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
			}
			if tc.fatal && !gotExit {
				t.Fatal("logx.ExitCodeChannel() message mismatch\ngot:  absent\nwant: present")
			}
			if !tc.fatal && gotExit {
				t.Fatal("logx.ExitCodeChannel() message mismatch\ngot:  present\nwant: absent")
			}
			if tc.stdoutRegex != "" {
				return
			}

			newData, err := os.ReadFile(cfg.File)
			if err != nil {
				t.Fatalf("saved: os.ReadFile() failed: %v", err)
			}
			for from := range tc.corrections {
				had = strings.ReplaceAll(had, from, tc.corrections[from])
			}
			if string(newData) != had {
				t.Errorf("saved file mismatch\ngot:  %q\nwant: %q", string(newData), had)
			}
		})
	}
}

// reindentLines converts the leading whitespace of lines from one indent size
// to another, preserving any partial offset (e.g. the 2-space alignment of
// sequence continuation lines).
func reindentLines(lines []string, from, to int) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		spaces := len(line) - len(strings.TrimLeft(line, " "))
		levels := spaces / from
		offset := spaces % from
		out[i] = strings.Repeat(" ", levels*to+offset) + line[spaces:]
	}
	return out
}

func TestConfig_Save__indentation(t *testing.T) {
	tests := []struct {
		name        string
		indentation uint8
	}{
		{name: "2 spaces", indentation: 2},
		{name: "4 spaces", indentation: 4},
		{name: "6 spaces", indentation: 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			file := filepath.Join(t.TempDir(), "config.yml")
			testYAML_config_indent4(t, file)
			hadData, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("original: os.ReadFile() failed: %v", err)
			}
			cfg := testLoadBasic(t, file)
			hadIndentation := cfg.Settings.Indentation

			cfg.Settings.Indentation = tc.indentation

			saveDone := make(chan struct{})
			go func() {
				loadMu.RLock()
				cfg.Save()
				loadMu.RUnlock()
				close(saveDone)
			}()
			gotExit := gotExitCodeDuringSave(saveDone, false)

			releaseStdout()

			if gotExit {
				t.Fatal("logx.ExitCodeChannel() message mismatch\ngot:  present\nwant: absent")
			}

			newData, err := os.ReadFile(cfg.File)
			if err != nil {
				t.Fatalf("saved: os.ReadFile() failed: %v", err)
			}
			gotLines := strings.Split(string(newData), "\n")
			wantLines := reindentLines(
				strings.Split(string(hadData), "\n"),
				int(hadIndentation), int(tc.indentation),
			)
			if testErr := test.AssertSlicesEqualFunc(
				t,
				gotLines,
				wantLines,
				func(a, b string) bool { return a == b },
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}

			var cfg2 Config
			if err := cfg2.Decode(newData); err != nil {
				t.Fatalf("Config.Decode() failed: %s", errfmt.FormatError(err))
			}
			svc := cfg2.Service["awesome"]
			if svc == nil {
				t.Fatal("Config.Service[\"awesome\"] mismatch\ngot:  nil\nwant: non-nil")
			}
			if svc.LatestVersion == nil {
				t.Error("Service.LatestVersion mismatch\ngot:  nil\nwant: non-nil")
			}
			if svc.Options.SemanticVersioning == nil {
				t.Error("Service.Options.SemanticVersioning mismatch\ngot:  nil\nwant: non-nil")
			}
			if svc.Dashboard.Icon == "" {
				t.Error("Service.Dashboard.Icon mismatch\ngot:  \"\"\nwant: NOT \"\"")
			}
			if got, want := len(svc.Command), 1; got != want {
				t.Errorf("len(Service.Command) mismatch\ngot:  %d\nwant: %d", got, want)
			}
		})
	}
}

func TestConfig_Save__encodeError(t *testing.T) {
	original := encodeConfigYAML
	encodeConfigYAML = func(w io.Writer, indent int, c *Config) error {
		return errors.New("encode failed")
	}
	t.Cleanup(func() { encodeConfigYAML = original })

	file := filepath.Join(t.TempDir(), "config.yml")
	testYAML_config_small(t, file)
	cfg := testLoadBasic(t, file)

	releaseStdout := test.CaptureLog(t, logx.Default())

	saveDone := make(chan struct{})
	go func() {
		cfg.Save()
		close(saveDone)
	}()
	gotExit := gotExitCodeDuringSave(saveDone, true)

	stdout := releaseStdout()
	wantRegex := `failed to encode config`
	if !util.RegexCheck(wantRegex, stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, wantRegex)
	}
	if !gotExit {
		t.Fatal("logx.ExitCodeChannel() message mismatch\ngot:  absent\nwant: present")
	}
}

func TestConfig_Save__readonly(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.yml")
	testYAML_config_small(t, file)
	cfg := testLoadBasic(t, file)
	cfg.Settings.Data.Readonly = new(true)

	opens := 0
	original := openSaveFile
	openSaveFile = func(name string) (io.WriteCloser, error) {
		opens++
		return original(name)
	}
	t.Cleanup(func() { openSaveFile = original })

	releaseStdout := test.CaptureLog(t, logx.Default())

	loadMu.RLock()
	got := cfg.Save()
	loadMu.RUnlock()

	stdout := releaseStdout()

	if opens != 0 {
		t.Errorf("openSaveFile() call count mismatch\ngot:  %d\nwant: 0", opens)
	}

	if !got {
		t.Error("ok mismatch\ngot:  false\nwant: true")
	}

	wantRegex := `Readonly mode: not saving`
	if !util.RegexCheck(wantRegex, stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, wantRegex)
	}
}

// failWriter is an io.WriteCloser whose Write always fails.
// Used to trigger the "failed to flush" path in Save() when all data fits in
// bufio's buffer and the error surfaces at the final Flush().
type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("simulated disk full")
}
func (f *failWriter) Close() error { return nil }

// failAfterNBytesWriter is an io.WriteCloser that succeeds until remaining
// bytes are written, then fails. Used to trigger the "failed to write" path in
// Save() when the config is large enough to overflow bufio's buffer mid-loop.
type failAfterNBytesWriter struct{ remaining int }

func (f *failAfterNBytesWriter) Write(p []byte) (int, error) {
	if f.remaining <= 0 {
		return 0, errors.New("simulated disk full")
	}
	if len(p) > f.remaining {
		n := f.remaining
		f.remaining = 0
		return n, errors.New("simulated disk full")
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
		{
			name:  "deep nesting/an indentation jump of 10+ levels",
			order: []string{"alpha"},
			lines: []string{
				"service:",
				"  alpha:",
				"    notify:",
				"      discord:",
				"        options:",
				"          message: |-",
				"            a",
				"                                            b",
			},
			want: []string{
				"service:",
				"  alpha:",
				"    notify:",
				"      discord:",
				"        options:",
				"          message: |-",
				"            a",
				"                                            b",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg Config
			yamlStr := strings.Join(tc.lines, "\n")
			if err := cfg.Decode([]byte(yamlStr)); err != nil {
				t.Fatalf("Config.Decode() failed: %s", errfmt.FormatError(err))
			}
			cfg.Order = tc.order
			cfg.Settings.Indentation = 2

			got := cfg.reorderYAML(tc.lines)

			gotStr := strings.Join(got, "\n")
			wantStr := strings.Join(tc.want, "\n")
			if gotStr != wantStr {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}
		})
	}
}
