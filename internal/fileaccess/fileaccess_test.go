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

package fileaccess

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "nested", "VERSION")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("productversion=\"1.2.3-54321\"\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}
	if err := SetAllowedPaths([]string{root}); err != nil {
		t.Fatalf("SetAllowedPaths() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := SetAllowedPaths(nil); err != nil {
			t.Errorf("SetAllowedPaths(nil) failed: %v", err)
		}
	})

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}
	if want := "productversion=\"1.2.3-54321\"\n"; string(got) != want {
		t.Errorf("Read() mismatch\ngot:  %q\nwant: %q", got, want)
	}

	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(outside) failed: %v", err)
	}
	if _, err := Read(outside); err == nil || !strings.Contains(err.Error(), "outside configured") {
		t.Errorf("Read(outside) error = %v, want outside-root error", err)
	}
}

func TestReadAllowsConfiguredFileOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(path, []byte("version"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(path) failed: %v", err)
	}
	other := filepath.Join(dir, "other")
	if err := os.WriteFile(other, []byte("other"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(other) failed: %v", err)
	}
	if err := SetAllowedPaths([]string{path}); err != nil {
		t.Fatalf("SetAllowedPaths(file) failed: %v", err)
	}
	t.Cleanup(func() {
		if err := SetAllowedPaths(nil); err != nil {
			t.Errorf("SetAllowedPaths(nil) failed: %v", err)
		}
	})

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read(configured file) failed: %v", err)
	}
	if string(got) != "version" {
		t.Errorf("Read(configured file) = %q, want %q", got, "version")
	}
	if _, err := Read(other); err == nil || !strings.Contains(err.Error(), "outside configured") {
		t.Errorf("Read(other) error = %v, want outside-root error", err)
	}
}

func TestReadRejectsUnsafeFiles(t *testing.T) {
	root := t.TempDir()
	if err := SetAllowedPaths([]string{root}); err != nil {
		t.Fatalf("SetAllowedPaths() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := SetAllowedPaths(nil); err != nil {
			t.Errorf("SetAllowedPaths(nil) failed: %v", err)
		}
	})

	tests := []struct {
		name string
		path string
	}{
		{name: "relative path", path: "VERSION"},
		{name: "directory", path: root},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Read(tc.path); err == nil {
				t.Fatal("Read() unexpectedly succeeded")
			}
		})
	}

	largeFile := filepath.Join(root, "large")
	if err := os.WriteFile(largeFile, []byte(strings.Repeat("x", maxFileSize+1)), 0o600); err != nil {
		t.Fatalf("os.WriteFile(large) failed: %v", err)
	}
	if _, err := Read(largeFile); err == nil || !strings.Contains(err.Error(), "limit") {
		t.Errorf("Read(large) error = %v, want size-limit error", err)
	}

	if runtime.GOOS != "windows" {
		outside := filepath.Join(t.TempDir(), "outside")
		if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
			t.Fatalf("os.WriteFile(outside) failed: %v", err)
		}
		link := filepath.Join(root, "escape")
		if err := os.Symlink(outside, link); err != nil {
			t.Fatalf("os.Symlink() failed: %v", err)
		}
		if _, err := Read(link); err == nil {
			t.Error("Read(symlink escape) unexpectedly succeeded")
		}
	}
}

func TestReadRejectsNamedPipesWithoutBlocking(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("named pipes require platform-specific setup")
	}

	root := t.TempDir()
	path := filepath.Join(root, "pipe")
	mkfifo, err := exec.LookPath("mkfifo")
	if err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	if err := exec.Command(mkfifo, path).Run(); err != nil {
		t.Fatalf("creating named pipe failed: %v", err)
	}
	if err := SetAllowedPaths([]string{root}); err != nil {
		t.Fatalf("SetAllowedPaths() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := SetAllowedPaths(nil); err != nil {
			t.Errorf("SetAllowedPaths(nil) failed: %v", err)
		}
	})

	result := make(chan error, 1)
	go func() {
		_, err := Read(path)
		result <- err
	}()

	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "regular file") {
			t.Errorf("Read(named pipe) error = %v, want regular-file error", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Read(named pipe) blocked")
	}
}

func TestSetAllowedPaths__rejectsInvalidPathsAndPreservesConfiguration(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "VERSION")
	if err := os.WriteFile(path, []byte("version"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}
	if err := SetAllowedPaths([]string{root}); err != nil {
		t.Fatalf("SetAllowedPaths(root) failed: %v", err)
	}
	t.Cleanup(func() {
		if err := SetAllowedPaths(nil); err != nil {
			t.Errorf("SetAllowedPaths(nil) failed: %v", err)
		}
	})

	if err := ValidateAllowedPaths([]string{filepath.Join(root, "missing")}); err == nil {
		t.Fatal("ValidateAllowedPaths(missing path) unexpectedly succeeded")
	}
	if err := SetAllowedPaths([]string{filepath.Join(root, "missing")}); err == nil {
		t.Fatal("SetAllowedPaths(missing path) unexpectedly succeeded")
	}
	if _, err := Read(path); err != nil {
		t.Fatalf("Read() after invalid configuration failed: %v", err)
	}
}

func TestRead__symlinkedAllowedPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated privileges")
	}

	root := t.TempDir()
	path := filepath.Join(root, "VERSION")
	if err := os.WriteFile(path, []byte("version"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(path) failed: %v", err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatalf("os.Symlink(directory) failed: %v", err)
	}
	fileAlias := filepath.Join(t.TempDir(), "VERSION")
	if err := os.Symlink(path, fileAlias); err != nil {
		t.Fatalf("os.Symlink(file) failed: %v", err)
	}

	tests := []struct {
		name         string
		allowedPaths []string
		readPaths    []string
		wantEntries  int
	}{
		{
			name:         "directory symlink",
			allowedPaths: []string{alias},
			readPaths:    []string{filepath.Join(alias, "VERSION")},
			wantEntries:  1,
		},
		{
			name:         "file symlink",
			allowedPaths: []string{fileAlias},
			readPaths:    []string{fileAlias},
			wantEntries:  1,
		},
		{
			name:         "directory and its symlink are merged",
			allowedPaths: []string{root, alias, root},
			readPaths:    []string{path, filepath.Join(alias, "VERSION")},
			wantEntries:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: the allowed paths are process-wide.

			if err := SetAllowedPaths(tc.allowedPaths); err != nil {
				t.Fatalf("SetAllowedPaths() failed: %v", err)
			}
			t.Cleanup(func() {
				if err := SetAllowedPaths(nil); err != nil {
					t.Errorf("SetAllowedPaths(nil) failed: %v", err)
				}
			})

			if got := len(configuredPaths.Load().paths); got != tc.wantEntries {
				t.Errorf("configured entry count mismatch\ngot:  %d\nwant: %d", got, tc.wantEntries)
			}
			for _, readPath := range tc.readPaths {
				got, err := Read(readPath)
				if err != nil {
					t.Fatalf("Read(%q) failed: %v", readPath, err)
				}
				if got, want := string(got), "version"; got != want {
					t.Errorf("Read(%q) mismatch\ngot:  %q\nwant: %q", readPath, got, want)
				}
			}
		})
	}
}
