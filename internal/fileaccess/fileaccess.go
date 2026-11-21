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

// Package fileaccess reads files from operator-configured roots.
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

type allowedRoots struct {
	paths []allowedRoot
}

type allowedRoot struct {
	path   string
	isFile bool
}

var configuredRoots atomic.Pointer[allowedRoots]

// ConfigureAllowedPaths validates and sets the paths that lookups may read.
func ConfigureAllowedPaths(paths []string) error {
	roots := make([]allowedRoot, 0, len(paths))
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
		roots = append(roots, allowedRoot{path: absolute, isFile: !info.IsDir()})
	}

	configuredRoots.Store(&allowedRoots{paths: roots})
	return nil
}

// Read reads a regular file beneath one of the configured roots.
func Read(path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("file path must be absolute")
	}
	path = filepath.Clean(path)

	roots := configuredRoots.Load()
	if roots == nil || len(roots.paths) == 0 {
		return nil, errors.New("local file access is disabled; configure settings.data.file_read_roots")
	}

	var rootPath, relative string
	for _, root := range roots.paths {
		if root.isFile {
			if path != root.path {
				continue
			}
			rootPath = filepath.Dir(root.path)
			relative = filepath.Base(root.path)
			break
		}

		candidate, err := filepath.Rel(root.path, path)
		if err != nil || candidate == "." || candidate == ".." ||
			strings.HasPrefix(candidate, ".."+string(filepath.Separator)) {
			continue
		}
		rootPath = root.path
		relative = candidate
		break
	}
	if rootPath == "" {
		return nil, errors.New("file path is outside configured settings.data.file_read_roots")
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, fmt.Errorf("open allowed root: %w", err)
	}
	defer root.Close()

	info, err := root.Stat(relative)
	if err != nil {
		return nil, fmt.Errorf("stat file beneath allowed root: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("file path must refer to a regular file")
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file exceeds the %d-byte limit", maxFileSize)
	}

	file, err := root.Open(relative)
	if err != nil {
		return nil, fmt.Errorf("open file beneath allowed root: %w", err)
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
