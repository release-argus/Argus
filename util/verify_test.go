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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckFileReadable(t *testing.T) {
	tmpDir := t.TempDir() // isolated temp dir for testing.

	// Prepare some files and directories.
	existingFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Directory instead of file.
	existingDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(existingDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Relative path that doesn’t exist.
	relativeNonExistent := "nonexistent.txt"

	tests := []struct {
		name         string
		path         string
		expectError  bool
		containsPath string // substring expected in the error message (for relative paths).
	}{
		{
			name:        "empty path",
			path:        "",
			expectError: false,
		},
		{
			name:        "existing absolute file",
			path:        existingFile,
			expectError: false,
		},
		{
			name:         "existing directory",
			path:         existingDir,
			expectError:  true,
			containsPath: existingDir,
		},
		{
			name:        "non-existent absolute file",
			path:        filepath.Join(tmpDir, "missing.txt"),
			expectError: true,
		},
		{
			name:         "non-existent relative file",
			path:         relativeNonExistent,
			expectError:  true,
			containsPath: relativeNonExistent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// GIVEN: a path.
			path := tc.path

			// WHEN: CheckFileReadable is called.
			err := CheckFileReadable(path)

			prefix := fmt.Sprintf(
				"%s\nCheckFileReadable(%q)",
				packageName, path,
			)

			// THEN: the error behavior should match expectations.
			if tc.expectError && err == nil || !tc.expectError && err != nil {
				want := "nil"
				if tc.expectError {
					want = "error"
				}
				t.Fatalf(
					"%s error mismatch\ngot:  %v\nwant: %v",
					prefix, err, want,
				)
			}

			// AND: if a substring is expected in the error, check it.
			if tc.containsPath != "" && err != nil {
				if !errors.Is(err, os.ErrNotExist) && !strings.Contains(err.Error(), tc.containsPath) {
					t.Fatalf(
						"%s error mismatch\ngot:  %v\nwant contains(%q)",
						prefix, err, tc.containsPath,
					)
				}
			}
		})
	}
}

func TestCheckFileReadable__statError(t *testing.T) {
	// GIVEN: a failing stat call.
	original := fileStat
	customErr := fmt.Errorf("stat failed")
	fileStat = func(f *os.File) (os.FileInfo, error) {
		return nil, customErr
	}
	t.Cleanup(func() { fileStat = original })

	// AND: a readable file.
	existingFile := filepath.Join(t.TempDir(), "existing.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// WHEN: CheckFileReadable is called.
	err := CheckFileReadable(existingFile)

	prefix := fmt.Sprintf(
		"%s\nCheckFileReadable(%q)",
		packageName, existingFile,
	)

	// THEN: the stat error is returned.
	if err == nil || err.Error() != customErr.Error() {
		t.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, err, customErr,
		)
	}
}
