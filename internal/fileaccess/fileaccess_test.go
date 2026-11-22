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
	if err := os.WriteFile(path, []byte("productversion=\"7.2.2-72806\"\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}
	if err := ConfigureAllowedPaths([]string{root}); err != nil {
		t.Fatalf("ConfigureAllowedPaths() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := ConfigureAllowedPaths(nil); err != nil {
			t.Errorf("ConfigureAllowedPaths(nil) failed: %v", err)
		}
	})

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}
	if want := "productversion=\"7.2.2-72806\"\n"; string(got) != want {
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
	if err := ConfigureAllowedPaths([]string{path}); err != nil {
		t.Fatalf("ConfigureAllowedPaths(file) failed: %v", err)
	}
	t.Cleanup(func() {
		if err := ConfigureAllowedPaths(nil); err != nil {
			t.Errorf("ConfigureAllowedPaths(nil) failed: %v", err)
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
	if err := ConfigureAllowedPaths([]string{root}); err != nil {
		t.Fatalf("ConfigureAllowedPaths() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := ConfigureAllowedPaths(nil); err != nil {
			t.Errorf("ConfigureAllowedPaths(nil) failed: %v", err)
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
	if err := ConfigureAllowedPaths([]string{root}); err != nil {
		t.Fatalf("ConfigureAllowedPaths() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := ConfigureAllowedPaths(nil); err != nil {
			t.Errorf("ConfigureAllowedPaths(nil) failed: %v", err)
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

func TestConfigureAllowedPaths__rejectsInvalidPathsAndPreservesConfiguration(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "VERSION")
	if err := os.WriteFile(path, []byte("version"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}
	if err := ConfigureAllowedPaths([]string{root}); err != nil {
		t.Fatalf("ConfigureAllowedPaths(root) failed: %v", err)
	}
	t.Cleanup(func() {
		if err := ConfigureAllowedPaths(nil); err != nil {
			t.Errorf("ConfigureAllowedPaths(nil) failed: %v", err)
		}
	})

	if err := ConfigureAllowedPaths([]string{filepath.Join(root, "missing")}); err == nil {
		t.Fatal("ConfigureAllowedPaths(missing path) unexpectedly succeeded")
	}
	if _, err := Read(path); err != nil {
		t.Fatalf("Read() after invalid configuration failed: %v", err)
	}
}
