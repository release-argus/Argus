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

package util

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/release-argus/Argus/util/errfmt"
)

func TestCheckFileReadable(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	existingDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(existingDir, 0o755); err != nil {
		t.Fatalf("os.Mkdir() failed: %v", err)
	}

	relativeNonExistent := "nonexistent.txt"
	missingFile := filepath.Join(tmpDir, "missing.txt")

	tests := []struct {
		name     string
		path     string
		errRegex string
	}{
		{
			name:     "empty path",
			path:     "",
			errRegex: `^$`,
		},
		{
			name:     "existing absolute file",
			path:     existingFile,
			errRegex: `^$`,
		},
		{
			name:     "existing directory",
			path:     existingDir,
			errRegex: `^path "` + regexp.QuoteMeta(existingDir) + `" is a directory, not a file$`,
		},
		{
			name: "non-existent absolute file",
			path: missingFile,
			errRegex: `^open ` + regexp.QuoteMeta(missingFile) +
				`:\n  no such file or directory$`,
		},
		{
			name:     "non-existent relative file",
			path:     relativeNonExistent,
			errRegex: `^open .+/` + regexp.QuoteMeta(relativeNonExistent) + `: no such file or directory$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckFileReadable(tc.path)
			e := errfmt.FormatError(err)
			if !RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestCheckFileReadable__statError(t *testing.T) {
	original := fileStat
	customErr := errors.New("stat failed")
	fileStat = func(f *os.File) (os.FileInfo, error) {
		return nil, customErr
	}
	t.Cleanup(func() { fileStat = original })

	existingFile := filepath.Join(t.TempDir(), "existing.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	err := CheckFileReadable(existingFile)
	if err == nil || err.Error() != customErr.Error() {
		t.Errorf("error mismatch\ngot:  %v\nwant: %v", err, customErr)
	}
}
