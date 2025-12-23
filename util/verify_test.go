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

package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendCheckError(t *testing.T) {
	tests := map[string]struct {
		prefix, label string
		checkErr      error
		errs, want    []error
	}{
		"nil checkErr": {
			errs:     []error{},
			prefix:   "prefix",
			label:    "label",
			checkErr: nil,
			want:     []error{},
		},
		"non-nil checkErr": {
			errs:     []error{},
			prefix:   "prefix_",
			label:    "label",
			checkErr: fmt.Errorf("an error occurred"),
			want: []error{
				fmt.Errorf("%slabel:\nan error occurred",
					"prefix_"),
			},
		},
		"existing errors with non-nil checkErr": {
			errs: []error{
				fmt.Errorf("existing error"),
			},
			prefix:   "prefix_",
			label:    "label",
			checkErr: fmt.Errorf("an error occurred"),
			want: []error{
				fmt.Errorf("existing error"),
				fmt.Errorf("%slabel:\nan error occurred",
					"prefix_"),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			errs := tc.errs
			AppendCheckError(&errs, tc.prefix, tc.label, tc.checkErr)
			if len(errs) != len(tc.want) {
				t.Fatalf("%s\nerror mismatch\nwant: %d errors, got %d\nerrors: %q",
					packageName, len(tc.want), len(errs), errs)
			}
			for i, err := range errs {
				if err.Error() != tc.want[i].Error() {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want[i], err)
				}
			}
		})
	}
}

func TestErrorToString(t *testing.T) {
	// GIVEN a bunch of comparables.
	tests := map[string]struct {
		err  error
		want string
	}{
		"nil error": {
			err: nil, want: ""},
		"non-nil error": {
			err: fmt.Errorf("test error"), want: "test error"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ErrorToString is called.
			got := ErrorToString(tc.err)

			// THEN the var is printed when it should be.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestCheckFileReadable(t *testing.T) {
	tmpDir := t.TempDir() // isolated temp dir for testing

	// Prepare some files and directories
	existingFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Directory instead of file
	existingDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(existingDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Relative path that doesn’t exist
	relativeNonExistent := "nonexistent.txt"

	tests := map[string]struct {
		path         string
		expectError  bool
		containsPath string // substring expected in the error message (for relative paths)
	}{
		"empty path": {
			path:        "",
			expectError: false,
		},
		"existing absolute file": {
			path:        existingFile,
			expectError: false,
		},
		"existing directory": {
			path:         existingDir,
			expectError:  true,
			containsPath: existingDir,
		},
		"non-existent absolute file": {
			path:        filepath.Join(tmpDir, "missing.txt"),
			expectError: true,
		},
		"non-existent relative file": {
			path:         relativeNonExistent,
			expectError:  true,
			containsPath: relativeNonExistent,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN a path
			path := tc.path

			// WHEN CheckFileReadable is called
			err := CheckFileReadable(path)

			// THEN the error behavior should match expectations
			if tc.expectError && err == nil || !tc.expectError && err != nil {
				want := "nil"
				if tc.expectError {
					want = "error"
				}
				t.Fatalf("%s\nerror mismatch\nwant: %v\ngot:  %v",
					packageName, want, err)
			}

			// AND if a substring is expected in the error, check it
			if tc.containsPath != "" && err != nil {
				if !errors.Is(err, os.ErrNotExist) && !strings.Contains(err.Error(), tc.containsPath) {
					t.Fatalf("%s\nexpected error message to contain %q, got: %v",
						packageName, tc.containsPath, err)
				}
			}
		})
	}
}
