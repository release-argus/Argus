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

// Package util provides utility functions for the Argus project.
package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppendCheckError adds a formatted error to the slice if checkErr exists.
// The message includes the prefix and label.
func AppendCheckError(errs *[]error, prefix, label string, checkErr error) {
	if checkErr != nil {
		*errs = append(*errs,
			fmt.Errorf("%s%s:\n%w",
				prefix, label, checkErr))
	}
}

// ErrorToString converts an error to a string.
// nil == "".
func ErrorToString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

// CheckFileReadable checks if the file at the given path is readable.
//
// It returns an error if the file cannot be opened/read.
// If the file path is empty, it returns nil.
func CheckFileReadable(path string) error {
	if path == "" {
		return nil
	}

	f, err := os.Open(path)
	// Failed to open.
	if err != nil {
		if !filepath.IsAbs(path) {
			execPath, _ := os.Executable()
			err = errors.New(strings.Replace(
				err.Error(),
				fmt.Sprintf(" %s:", path),
				fmt.Sprintf(" %s/%s:", execPath, path),
				1,
			))
		}
		return err
	}
	defer f.Close()

	if info, e := f.Stat(); e != nil {
		return e //nolint:wrapcheck
	} else if info.IsDir() {
		return fmt.Errorf("path %q is a directory, not a file", path)
	}

	return nil
}
