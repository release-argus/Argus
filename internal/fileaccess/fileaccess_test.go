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
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// setAllowedPaths configures the allowed paths for the test and clears them afterwards.
func setAllowedPaths(t *testing.T, paths ...string) {
	t.Helper()
	if err := SetAllowedPaths(paths); err != nil {
		t.Fatalf("SetAllowedPaths(%q) failed: %v", paths, err)
	}
	t.Cleanup(func() {
		if err := SetAllowedPaths(nil); err != nil {
			t.Errorf("SetAllowedPaths(nil) failed: %v", err)
		}
	})
}

// writeFile writes content to path, creating parent directories.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("os.MkdirAll(%q) failed: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", path, err)
	}
}

// makeFIFO creates a named pipe at path, skipping the test where that is unsupported.
func makeFIFO(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("named pipes require platform-specific setup")
	}
	mkfifo, err := exec.LookPath("mkfifo")
	if err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	if err := exec.Command(mkfifo, path).Run(); err != nil {
		t.Fatalf("creating named pipe failed: %v", err)
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name string
		// setup builds the layout, configures the allowed paths and returns the path to read.
		setup    func(t *testing.T, root string) string
		want     string
		errRegex string
	}{
		{
			name: "file in a configured directory",
			setup: func(t *testing.T, root string) string {
				path := filepath.Join(root, "nested", "VERSION")
				writeFile(t, path, "1.2.3")
				setAllowedPaths(t, root)
				return path
			},
			want:     "1.2.3",
			errRegex: `^$`,
		},
		{
			name: "configured file",
			setup: func(t *testing.T, root string) string {
				path := filepath.Join(root, "VERSION")
				writeFile(t, path, "version")
				setAllowedPaths(t, path)
				return path
			},
			want:     "version",
			errRegex: `^$`,
		},
		{
			name: "sibling of a configured file is refused",
			setup: func(t *testing.T, root string) string {
				path := filepath.Join(root, "VERSION")
				writeFile(t, path, "version")
				other := filepath.Join(root, "other")
				writeFile(t, other, "other")
				setAllowedPaths(t, path)
				return other
			},
			errRegex: `outside configured`,
		},
		{
			name: "file outside the configured directory",
			setup: func(t *testing.T, root string) string {
				path := filepath.Join(root, "allowed")
				writeFile(t, filepath.Join(path, "VERSION"), "version")
				outside := filepath.Join(root, "outside")
				writeFile(t, outside, "secret")
				setAllowedPaths(t, filepath.Join(root, "allowed"))
				return outside
			},
			errRegex: `outside configured`,
		},
		{
			name: "relative path",
			setup: func(t *testing.T, root string) string {
				setAllowedPaths(t, root)
				return "VERSION"
			},
			errRegex: `must be absolute`,
		},
		{
			name: "no paths configured",
			setup: func(t *testing.T, root string) string {
				path := filepath.Join(root, "VERSION")
				writeFile(t, path, "version")
				setAllowedPaths(t)
				return path
			},
			errRegex: `local file access is disabled`,
		},
		{
			name: "the configured directory itself",
			setup: func(t *testing.T, root string) string {
				setAllowedPaths(t, root)
				return root
			},
			errRegex: `outside configured`,
		},
		{
			name: "sub-directory",
			setup: func(t *testing.T, root string) string {
				if err := os.Mkdir(filepath.Join(root, "nested"), 0o750); err != nil {
					t.Fatalf("os.Mkdir() failed: %v", err)
				}
				setAllowedPaths(t, root)
				return filepath.Join(root, "nested")
			},
			errRegex: `must refer to a regular file`,
		},
		{
			name: "file over the size limit",
			setup: func(t *testing.T, root string) string {
				path := filepath.Join(root, "large")
				writeFile(t, path, strings.Repeat("x", maxFileSize+1))
				setAllowedPaths(t, root)
				return path
			},
			errRegex: `exceeds the \d+-byte limit`,
		},
		{
			name: "file at the size limit",
			setup: func(t *testing.T, root string) string {
				path := filepath.Join(root, "limit")
				writeFile(t, path, strings.Repeat("x", maxFileSize))
				setAllowedPaths(t, root)
				return path
			},
			want:     strings.Repeat("x", maxFileSize),
			errRegex: `^$`,
		},
		{
			name: "missing file",
			setup: func(t *testing.T, root string) string {
				setAllowedPaths(t, root)
				return filepath.Join(root, "missing")
			},
			errRegex: `stat file under allowed path`,
		},
		{
			name: "symlink to file outside configured directory is refused",
			setup: func(t *testing.T, root string) string {
				// root/
				// ├── allowed/             # allowed directory
				// │   ├── VERSION
				// │   └── escape -> ../outside
				// └── outside              # file
				outside := filepath.Join(root, "outside")
				writeFile(t, outside, "secret")
				allowed := filepath.Join(root, "allowed")
				link := filepath.Join(allowed, "escape")
				writeFile(t, filepath.Join(allowed, "VERSION"), "version")
				if err := os.Symlink(outside, link); err != nil {
					t.Skipf("os.Symlink() failed: %v", err)
				}
				setAllowedPaths(t, allowed)
				return link
			},
			errRegex: `escapes from parent`,
		},
		{
			name: "directory symlink outside configured directory is refused",
			setup: func(t *testing.T, root string) string {
				// root/
				// ├── dir/                 # allowed directory
				// │   ├── VERSION
				// │   └── current -> ../ext
				// └── ext/
				//     └── VERSION
				dir := filepath.Join(root, "dir")
				writeFile(t, filepath.Join(dir, "VERSION"), "internal")
				writeFile(t, filepath.Join(root, "ext", "VERSION"), "external")
				if err := os.Symlink(filepath.Join("..", "ext"), filepath.Join(dir, "current")); err != nil {
					t.Skipf("os.Symlink() failed: %v", err)
				}
				setAllowedPaths(t, dir)
				return filepath.Join(dir, "current", "VERSION")
			},
			errRegex: `escapes from parent`,
		},
		{
			name: "configured directory removed after configuration",
			setup: func(t *testing.T, root string) string {
				dir := filepath.Join(root, "gone")
				path := filepath.Join(dir, "VERSION")
				writeFile(t, path, "version")
				setAllowedPaths(t, dir)
				if err := os.RemoveAll(dir); err != nil {
					t.Fatalf("os.RemoveAll() failed: %v", err)
				}
				return path
			},
			errRegex: `open file access directory`,
		},
		{
			name: "unreadable file",
			setup: func(t *testing.T, root string) string {
				if runtime.GOOS == "windows" || os.Getuid() == 0 {
					t.Skip("file permissions are not enforced")
				}
				path := filepath.Join(root, "VERSION")
				writeFile(t, path, "version")
				if err := os.Chmod(path, 0o000); err != nil {
					t.Fatalf("os.Chmod() failed: %v", err)
				}
				setAllowedPaths(t, root)
				return path
			},
			errRegex: `permission denied`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: the allowed paths are process-wide.
			path := tc.setup(t, t.TempDir())

			got, err := Read(path)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Read() error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if string(got) != tc.want {
				t.Errorf("Read() mismatch\ngot:  %.20q\nwant: %.20q", got, tc.want)
			}
		})
	}
}

func TestRead__namedPipe(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "pipe")
	makeFIFO(t, path)
	setAllowedPaths(t, root)

	result := make(chan error, 1)
	go func() {
		_, err := Read(path)
		result <- err
	}()

	select {
	case err := <-result:
		e := errfmt.FormatError(err)
		if errRegex := `regular file`; !util.RegexCheck(errRegex, e) {
			t.Errorf("Read() error mismatch\ngot:  %q\nwant: %q", e, errRegex)
		}
	case <-time.After(time.Second):
		t.Fatal("Read(named pipe) blocked")
	}
}

// fakeInfo is a [fs.FileInfo] with a fixed mode and size.
type fakeInfo struct {
	fs.FileInfo
	mode fs.FileMode
	size int64
}

func (f fakeInfo) Mode() fs.FileMode { return f.mode }
func (f fakeInfo) Size() int64       { return f.size }

func TestCheckFileInfo(t *testing.T) {
	tests := []struct {
		name string
		info fakeInfo
		want error
	}{
		{name: "empty regular file", info: fakeInfo{}, want: nil},
		{name: "regular file at the limit", info: fakeInfo{size: maxFileSize}, want: nil},
		{name: "regular file over the limit", info: fakeInfo{size: maxFileSize + 1}, want: errTooLarge},
		{name: "directory", info: fakeInfo{mode: fs.ModeDir}, want: errNotRegular},
		{name: "named pipe", info: fakeInfo{mode: fs.ModeNamedPipe}, want: errNotRegular},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkFileInfo(tc.info); !errors.Is(got, tc.want) {
				t.Errorf("checkFileInfo() mismatch\ngot:  %v\nwant: %v", got, tc.want)
			}
		})
	}
}

type fakeOpenedFile struct {
	reader  io.Reader
	info    fs.FileInfo
	statErr error
}

func (f fakeOpenedFile) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (f fakeOpenedFile) Stat() (fs.FileInfo, error) {
	return f.info, f.statErr
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestReadOpenedFile(t *testing.T) {
	statErr := errors.New("stat failed")
	readErr := errors.New("read failed")
	tests := []struct {
		name string
		file fakeOpenedFile
		want string
		err  error
	}{
		{
			name: "stat error",
			file: fakeOpenedFile{
				statErr: statErr,
			},
			err: statErr,
		},
		{
			name: "file changed to non-regular",
			file: fakeOpenedFile{
				info: fakeInfo{mode: fs.ModeDir},
			},
			err: errNotRegular,
		},
		{
			name: "file changed to over the size limit",
			file: fakeOpenedFile{
				info: fakeInfo{size: maxFileSize + 1},
			},
			err: errTooLarge,
		},
		{
			name: "read error",
			file: fakeOpenedFile{
				info:   fakeInfo{size: 1},
				reader: failingReader{err: readErr},
			},
			err: readErr,
		},
		{
			name: "file grows over the size limit after stat",
			file: fakeOpenedFile{
				info:   fakeInfo{size: maxFileSize},
				reader: strings.NewReader(strings.Repeat("x", maxFileSize+1)),
			},
			err: errTooLarge,
		},
		{
			name: "file read",
			file: fakeOpenedFile{
				info:   fakeInfo{size: 3},
				reader: strings.NewReader("1.2"),
			},
			want: "1.2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readOpenedFile(tc.file)
			if !errors.Is(err, tc.err) {
				t.Errorf("readOpenedFile() error mismatch\ngot:  %v\nwant: %v", err, tc.err)
			}
			if string(got) != tc.want {
				t.Errorf("readOpenedFile() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestResolveAllowedPaths__operationErrors(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "VERSION")
	writeFile(t, path, "1.2.3")

	absErr := errors.New("absolute path failed")
	statErr := errors.New("stat failed")
	parentResolutionErr := errors.New("parent resolution failed")
	tests := []struct {
		name      string
		configure func(*pathResolutionOps)
		wantErr   error
	}{
		{
			name: "absolute path error",
			configure: func(ops *pathResolutionOps) {
				ops.abs = func(string) (string, error) {
					return "", absErr
				}
			},
			wantErr: absErr,
		},
		{
			name: "stat error after symlink resolution",
			configure: func(ops *pathResolutionOps) {
				ops.stat = func(string) (os.FileInfo, error) {
					return nil, statErr
				}
			},
			wantErr: statErr,
		},
		{
			name: "configured file parent resolution error",
			configure: func(ops *pathResolutionOps) {
				evalSymlinks := ops.evalSymlinks
				calls := 0
				ops.evalSymlinks = func(path string) (string, error) {
					calls++
					if calls == 2 {
						return "", parentResolutionErr
					}
					return evalSymlinks(path)
				}
			},
			wantErr: parentResolutionErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ops := defaultPathResolutionOps()
			tc.configure(&ops)

			_, err := resolveAllowedPathsWithOps([]string{path}, ops)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("resolveAllowedPathsWithOps() error mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}
		})
	}
}

func TestLocateFile__parentResolutionError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "VERSION")

	_, err := locateFile(path, path, filepath.EvalSymlinks)

	if err == nil || !strings.Contains(err.Error(), "resolve directory") {
		t.Errorf("locateFile() error mismatch\ngot:  %v\nwant error containing %q", err, "resolve directory")
	}
}

func TestSetAllowedPaths__invalidPaths(t *testing.T) {
	tests := []struct {
		name string
		// paths returns the invalid configuration to attempt.
		paths    func(t *testing.T, root string) []string
		errRegex string
	}{
		{
			name:  "missing path",
			paths: func(t *testing.T, root string) []string { return []string{filepath.Join(root, "missing")} },
			errRegex: test.TrimYAML(`
				^resolve "[^"]+":
					lstat .*
						(no such file|cannot find).*$`),
		},
		{
			name:     "empty path",
			paths:    func(t *testing.T, root string) []string { return []string{root, ""} },
			errRegex: `must not contain an empty path`,
		},
		{
			name:     "blank path",
			paths:    func(t *testing.T, root string) []string { return []string{"  "} },
			errRegex: `must not contain an empty path`,
		},
		{
			name: "named pipe",
			paths: func(t *testing.T, root string) []string {
				path := filepath.Join(root, "pipe")
				makeFIFO(t, path)
				return []string{path}
			},
			errRegex: `neither a directory nor a regular file`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: the allowed paths are process-wide.
			root := t.TempDir()
			path := filepath.Join(root, "VERSION")
			writeFile(t, path, "version")
			setAllowedPaths(t, root)

			invalid := tc.paths(t, root)

			err := ValidateAllowedPaths(invalid)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("ValidateAllowedPaths() error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			err = SetAllowedPaths(invalid)
			e = errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("SetAllowedPaths() error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			// The previous configuration is kept.
			got, err := Read(path)
			if err != nil {
				t.Fatalf("Read() after invalid configuration failed: %v", err)
			}
			if string(got) != "version" {
				t.Errorf("Read() mismatch\ngot:  %q\nwant: %q", got, "version")
			}
		})
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

			setAllowedPaths(t, tc.allowedPaths...)

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

func TestRead__retargetedSymlinks(t *testing.T) {
	// retarget atomically repoints link to target, as kubelet and release tools do.
	retarget := func(t *testing.T, link, target string) {
		t.Helper()
		tmp := link + ".tmp"
		if err := os.Symlink(target, tmp); err != nil {
			t.Skipf("os.Symlink(%q) failed: %v", target, err)
		}
		if err := os.Rename(tmp, link); err != nil {
			t.Fatalf("os.Rename(%q) failed: %v", tmp, err)
		}
	}

	versions := []string{"1.0.0", "2.0.0"}
	tests := []struct {
		name string
		// setup builds the initial layout and returns the allowed path, the path read,
		// and a func that retargets the symlink and prunes the old target.
		setup func(t *testing.T, root string) (allowed, read string, swap func())
	}{
		{
			name: "file entry in a Kubernetes ConfigMap mount",
			setup: func(t *testing.T, root string) (string, string, func()) {
				// root/
				// ├── ..ts1/VERSION ("1.0.0")
				// ├── ..ts2/VERSION ("2.0.0")
				// ├── ..data -> ..ts1
				// └── VERSION -> ..data/VERSION
				writeFile(t, filepath.Join(root, "..ts1", "VERSION"), versions[0])
				writeFile(t, filepath.Join(root, "..ts2", "VERSION"), versions[1])
				link := filepath.Join(root, "..data")
				retarget(t, link, "..ts1")
				if err := os.Symlink(filepath.Join("..data", "VERSION"), filepath.Join(root, "VERSION")); err != nil {
					t.Skipf("os.Symlink(VERSION) failed: %v", err)
				}

				path := filepath.Join(root, "VERSION")
				return path, path, func() {
					retarget(t, link, "..ts2")
					if err := os.RemoveAll(filepath.Join(root, "..ts1")); err != nil {
						t.Fatalf("os.RemoveAll(..ts1) failed: %v", err)
					}
				}
			},
		},
		{
			name: "directory entry that is a release symlink",
			setup: func(t *testing.T, root string) (string, string, func()) {
				// root/
				// ├── releases/
				// │   ├── v1/VERSION ("1.0.0")
				// │   └── v2/VERSION ("2.0.0")
				// └── current -> releases/v1
				writeFile(t, filepath.Join(root, "releases", "v1", "VERSION"), versions[0])
				writeFile(t, filepath.Join(root, "releases", "v2", "VERSION"), versions[1])
				current := filepath.Join(root, "current")
				retarget(t, current, filepath.Join("releases", "v1"))

				return current, filepath.Join(current, "VERSION"), func() {
					retarget(t, current, filepath.Join("releases", "v2"))
					if err := os.RemoveAll(filepath.Join(root, "releases", "v1")); err != nil {
						t.Fatalf("os.RemoveAll(v1) failed: %v", err)
					}
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: the allowed paths are process-wide.
			allowed, read, swap := tc.setup(t, t.TempDir())
			setAllowedPaths(t, allowed)

			for i, want := range versions {
				if i == 1 {
					swap()
				}
				got, err := Read(read)
				if err != nil {
					t.Fatalf("Read(%q) wanting %q failed: %v", read, want, err)
				}
				if string(got) != want {
					t.Errorf("Read(%q) mismatch\ngot:  %q\nwant: %q", read, got, want)
				}
			}
		})
	}
}

func TestRead__fileEntrySwappedForSymlink(t *testing.T) {
	tests := []struct {
		name string
		// target returns the symlink target that replaces the allowed file.
		target   func(root string) string
		want     string
		errRegex string
	}{
		{
			name:     "target within the file's directory is followed",
			target:   func(root string) string { return "VERSION.new" },
			want:     "2.0.0",
			errRegex: `^$`,
		},
		{
			name:     "target outside the file's directory is refused",
			target:   func(root string) string { return filepath.Join(root, "secret", "VERSION") },
			errRegex: `escapes from parent`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: the allowed paths are process-wide.
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "allowed", "VERSION"), "1.0.0")
			writeFile(t, filepath.Join(root, "secret", "VERSION"), "SECRET")
			if err := os.WriteFile(filepath.Join(root, "allowed", "VERSION.new"), []byte("2.0.0"), 0o600); err != nil {
				t.Fatalf("os.WriteFile(VERSION.new) failed: %v", err)
			}
			path := filepath.Join(root, "allowed", "VERSION")
			setAllowedPaths(t, path)

			if err := os.Remove(path); err != nil {
				t.Fatalf("os.Remove(%q) failed: %v", path, err)
			}
			if err := os.Symlink(tc.target(root), path); err != nil {
				t.Fatalf("os.Symlink(%q) failed: %v", path, err)
			}

			got, err := Read(path)

			e := ""
			if err != nil {
				e = err.Error()
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Read() error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if string(got) != tc.want {
				t.Errorf("Read() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
