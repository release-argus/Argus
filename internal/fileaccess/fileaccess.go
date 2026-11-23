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

// Package fileaccess reads files from operator-configured paths.
package fileaccess

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

const maxFileSize = 1 << 20

type allowedPaths struct {
	paths []allowedPath
}

type allowedPath struct {
	path       string
	matchPaths []string
	isFile     bool
}

var configuredPaths atomic.Pointer[allowedPaths]

// ValidateAllowedPaths checks that every path is an existing directory or regular file.
func ValidateAllowedPaths(paths []string) error {
	_, err := resolveAllowedPaths(paths)
	return err
}

// SetAllowedPaths sets the paths that lookups may read.
func SetAllowedPaths(paths []string) error {
	allowed, err := resolveAllowedPaths(paths)
	if err != nil {
		return err
	}
	configuredPaths.Store(&allowedPaths{paths: allowed})
	return nil
}

func resolveAllowedPaths(paths []string) ([]allowedPath, error) {
	allowed := make([]allowedPath, 0, len(paths))
	seen := make(map[string]int, len(paths))
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			return nil, errors.New("must not contain an empty path")
		}

		configured, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", path, err)
		}
		configured = filepath.Clean(configured)
		resolved, err := filepath.EvalSymlinks(configured)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", path, err)
		}

		info, err := os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", path, err)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%q is neither a directory nor a regular file", path)
		}
		matchPaths := []string{configured}
		if resolved != configured {
			matchPaths = append(matchPaths, resolved)
		}
		if index, exists := seen[resolved]; exists {
			allowed[index].matchPaths = appendUnique(allowed[index].matchPaths, matchPaths...)
			continue
		}
		seen[resolved] = len(allowed)
		allowed = append(allowed, allowedPath{
			path:       resolved,
			matchPaths: matchPaths,
			isFile:     !info.IsDir(),
		})
	}

	return allowed, nil
}

func appendUnique(paths []string, additions ...string) []string {
	for _, addition := range additions {
		found := false
		for _, path := range paths {
			if path == addition {
				found = true
				break
			}
		}
		if !found {
			paths = append(paths, addition)
		}
	}
	return paths
}

// Read reads a regular file allowed by the configured paths.
func Read(path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("file path must be absolute")
	}
	path = filepath.Clean(path)

	allowed := configuredPaths.Load()
	if allowed == nil || len(allowed.paths) == 0 {
		return nil, errors.New("local file access is disabled; configure settings.data.file_read_paths")
	}

	var baseDir, relative string
	for _, allowedPath := range allowed.paths {
		if allowedPath.isFile {
			for _, matchPath := range allowedPath.matchPaths {
				if path == matchPath {
					baseDir = filepath.Dir(allowedPath.path)
					relative = filepath.Base(allowedPath.path)
					break
				}
			}
			if baseDir != "" {
				break
			}
			continue
		}

		for _, matchPath := range allowedPath.matchPaths {
			candidate, err := filepath.Rel(matchPath, path)
			if err != nil || candidate == "." || candidate == ".." ||
				strings.HasPrefix(candidate, ".."+string(filepath.Separator)) {
				continue
			}
			baseDir = allowedPath.path
			relative = candidate
			break
		}
		if baseDir != "" {
			break
		}
	}
	if baseDir == "" {
		return nil, errors.New("file path is outside configured settings.data.file_read_paths")
	}

	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return nil, fmt.Errorf("open file access directory: %w", err)
	}
	defer root.Close()

	info, err := root.Stat(relative)
	if err != nil {
		return nil, fmt.Errorf("stat file under allowed path: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("file path must refer to a regular file")
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file exceeds the %d-byte limit", maxFileSize)
	}

	file, err := root.Open(relative)
	if err != nil {
		return nil, fmt.Errorf("open file under allowed path: %w", err)
	}
	defer file.Close()

	info, err = file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("file path must refer to a regular file")
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file exceeds the %d-byte limit", maxFileSize)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if len(data) > maxFileSize {
		return nil, fmt.Errorf("file exceeds the %d-byte limit", maxFileSize)
	}
	return data, nil
}
