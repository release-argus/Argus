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
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
)

// maxFileSize is the largest file, in bytes, that Read returns.
const maxFileSize = 1 << 20 // 1MiB.

var (
	// errNotRegular is returned when a path refers to anything other than a regular file.
	errNotRegular = errors.New("file path must refer to a regular file")
	// errTooLarge is returned when a file exceeds [maxFileSize].
	errTooLarge = fmt.Errorf("file exceeds the %d-byte limit", maxFileSize)
)

// allowedPaths is the set of entries that Read may serve files from.
type allowedPaths struct {
	paths []allowedPath
}

// allowedPath is a single configured directory or file entry.
type allowedPath struct {
	matchPaths []string
	isFile     bool
	// file is where a file entry is opened from. When the file resolves inside its configured
	// directory, it is the configured path, so [os.Root] follows symlink swaps within that
	// directory (e.g. a Kubernetes ConfigMap update) but refuses escapes. Otherwise, it pins
	// the location the file resolved to at configuration time.
	file string
}

// configuredPaths holds the active configuration, or nil until one is set.
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

type pathResolutionOps struct {
	abs          func(string) (string, error)
	evalSymlinks func(string) (string, error)
	stat         func(string) (os.FileInfo, error)
}

func defaultPathResolutionOps() pathResolutionOps {
	return pathResolutionOps{
		abs:          filepath.Abs,
		evalSymlinks: filepath.EvalSymlinks,
		stat:         os.Stat,
	}
}

// resolveAllowedPaths validates paths and returns one entry per distinct resolved path.
// Paths that resolve to the same location (e.g. a directory and a symlink to it) are merged.
func resolveAllowedPaths(paths []string) ([]allowedPath, error) {
	return resolveAllowedPathsWithOps(paths, defaultPathResolutionOps())
}

func resolveAllowedPathsWithOps(paths []string, ops pathResolutionOps) ([]allowedPath, error) {
	allowed := make([]allowedPath, 0, len(paths))
	seen := make(map[string]int, len(paths))
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			return nil, errors.New("must not contain an empty path")
		}

		configured, err := ops.abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", path, err)
		}
		configured = filepath.Clean(configured)
		resolved, err := ops.evalSymlinks(configured)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", path, err)
		}

		info, err := ops.stat(resolved)
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
		entry := allowedPath{
			matchPaths: matchPaths,
			isFile:     !info.IsDir(),
		}
		if entry.isFile {
			entry.file, err = locateFile(configured, resolved, ops.evalSymlinks)
			if err != nil {
				return nil, fmt.Errorf("resolve %q: %w", path, err)
			}
		}
		allowed = append(allowed, entry)
	}

	return allowed, nil
}

// locateFile returns the path to open a configured file entry from.
// See [allowedPath.file].
func locateFile(configured, resolved string, evalSymlinks func(string) (string, error)) (string, error) {
	resolvedDir, err := evalSymlinks(filepath.Dir(configured))
	if err != nil {
		return "", fmt.Errorf("resolve directory: %w", err)
	}
	// Check whether the resolved file remains within the resolved directory.
	if rel, err := filepath.Rel(resolvedDir, resolved); err == nil &&
		rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return configured, nil
	}
	return resolved, nil
}

// appendUnique appends each addition to paths unless it is already present.
func appendUnique(paths []string, additions ...string) []string {
	for _, addition := range additions {
		if !slices.Contains(paths, addition) {
			paths = append(paths, addition)
		}
	}
	return paths
}

// findTarget returns the root directory, and the path relative to it, that the first
// entry matching path serves it from. ok is false when no entry matches.
func findTarget(entries []allowedPath, path string) (baseDir, relative string, ok bool) {
	for _, entry := range entries {
		if entry.isFile {
			if slices.Contains(entry.matchPaths, path) {
				return filepath.Dir(entry.file), filepath.Base(entry.file), true
			}
			continue
		}

		// Directory entries are opened at the configured path, so a retargeted directory
		// symlink (e.g. a `current` release link) is followed.
		for _, matchPath := range entry.matchPaths {
			relative, err := filepath.Rel(matchPath, path)
			if err != nil || relative == "." || relative == ".." ||
				strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				continue
			}
			return matchPath, relative, true
		}
	}
	return "", "", false
}

// checkFileInfo returns an error unless info describes a regular file within the [maxFileSize]
// limit.
func checkFileInfo(info fs.FileInfo) error {
	if !info.Mode().IsRegular() {
		return errNotRegular
	}
	if info.Size() > maxFileSize {
		return errTooLarge
	}
	return nil
}

// Read reads a regular file allowed by the configured paths.
// The first entry that matches path is used.
func Read(path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("file path must be absolute")
	}
	path = filepath.Clean(path)

	allowed := configuredPaths.Load()
	if allowed == nil || len(allowed.paths) == 0 {
		return nil, errors.New("local file access is disabled; configure settings.data.file_read_paths")
	}

	baseDir, relative, ok := findTarget(allowed.paths, path)
	if !ok {
		return nil, errors.New("file path is outside configured settings.data.file_read_paths")
	}
	return readUnder(baseDir, relative)
}

// readUnder reads relative from within baseDir, refusing anything that is not a
// regular file or that leaves baseDir.
func readUnder(baseDir, relative string) ([]byte, error) {
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return nil, fmt.Errorf("open file access directory: %w", err)
	}
	defer root.Close()

	// Stat before opening to reject non-regular files before opening them.
	info, err := root.Stat(relative)
	if err != nil {
		return nil, fmt.Errorf("stat file under allowed path: %w", err)
	}
	if err := checkFileInfo(info); err != nil {
		return nil, err
	}

	file, err := root.Open(relative)
	if err != nil {
		return nil, fmt.Errorf("open file under allowed path: %w", err)
	}
	defer file.Close()

	return readOpenedFile(file)
}

type openedFile interface {
	io.Reader
	Stat() (fs.FileInfo, error)
}

func readOpenedFile(file openedFile) ([]byte, error) {
	// Stat again, as the file may have been swapped since the first check.
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if err := checkFileInfo(info); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if len(data) > maxFileSize {
		return nil, errTooLarge
	}
	return data, nil
}
