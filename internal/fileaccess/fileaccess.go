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
	path   string
	isFile bool
}

var configuredPaths atomic.Pointer[allowedPaths]

// ConfigureAllowedPaths validates and sets the paths that lookups may read.
func ConfigureAllowedPaths(paths []string) error {
	allowed := make([]allowedPath, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			return errors.New("must not contain an empty path")
		}

		absolute, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve %q: %w", path, err)
		}
		absolute, err = filepath.EvalSymlinks(absolute)
		if err != nil {
			return fmt.Errorf("resolve %q: %w", path, err)
		}

		info, err := os.Stat(absolute)
		if err != nil {
			return fmt.Errorf("stat %q: %w", path, err)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("%q is neither a directory nor a regular file", path)
		}
		if _, exists := seen[absolute]; exists {
			continue
		}
		seen[absolute] = struct{}{}
		allowed = append(allowed, allowedPath{path: absolute, isFile: !info.IsDir()})
	}

	configuredPaths.Store(&allowedPaths{paths: allowed})
	return nil
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
			if path != allowedPath.path {
				continue
			}
			baseDir = filepath.Dir(allowedPath.path)
			relative = filepath.Base(allowedPath.path)
			break
		}

		candidate, err := filepath.Rel(allowedPath.path, path)
		if err != nil || candidate == "." || candidate == ".." ||
			strings.HasPrefix(candidate, ".."+string(filepath.Separator)) {
			continue
		}
		baseDir = allowedPath.path
		relative = candidate
		break
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
