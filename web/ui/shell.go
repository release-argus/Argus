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

package ui

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"slices"
	"strings"
	"time"
)

// The app shell, as embedded: plain from a dev build, gzipped from a release.
const (
	indexFile   = "index.html"
	indexFileGz = indexFile + ".gz"
)

// react-app/index.html delimits its <base href> with these markers, so the
// rewrite does not depend on the tag's formatting.
const (
	baseHrefStart = "<!--ARGUS_BASE_HREF_START-->"
	baseHrefEnd   = "<!--ARGUS_BASE_HREF_END-->"
)

// Rewritten returns the embedded UI with the app shell's <base href> pointing
// at routePrefix, so the relative asset URLs inside it resolve there whatever
// the depth of the route being served.
//
// On any failure the UI is returned as embedded, alongside the error: the shell
// then still serves, just only correctly from the root.
func Rewritten(routePrefix string) (fs.ReadDirFS, error) {
	return rewritten(GetFS(), routePrefix)
}

// rewritten is [Rewritten] over any FS, for testing against a fixture.
func rewritten(fSys fs.ReadDirFS, routePrefix string) (fs.ReadDirFS, error) {
	shell, err := indexHTML(fSys, routePrefix)
	if err != nil {
		return fSys, err
	}

	// Release builds embed the shell gzipped and dev builds plain, so replace
	// whichever is actually there.
	overlay := make(map[string][]byte, 1)
	if _, err := fs.Stat(fSys, indexFile); err == nil {
		overlay[indexFile] = shell
	}
	if _, err := fs.Stat(fSys, indexFileGz); err == nil {
		gzipped, err := gzipBytes(shell)
		if err != nil {
			return fSys, err
		}
		overlay[indexFileGz] = gzipped
	}
	if len(overlay) == 0 {
		return fSys, fmt.Errorf("no %s to overlay", indexFile)
	}

	return overlaidFS{ReadDirFS: fSys, overlay: overlay}, nil
}

// indexHTML reads the app shell from fSys (gzipped or plain) and rewrites its
// <base href> to routePrefix.
func indexHTML(fSys fs.FS, routePrefix string) ([]byte, error) {
	raw, err := fs.ReadFile(fSys, indexFile)
	if err != nil {
		gzipped, gzErr := fs.ReadFile(fSys, indexFileGz)
		if gzErr != nil {
			return nil, fmt.Errorf("read %s: %w", indexFile, errors.Join(err, gzErr))
		}
		reader, gzErr := gzip.NewReader(bytes.NewReader(gzipped))
		if gzErr != nil {
			return nil, fmt.Errorf("decompress %s: %w", indexFileGz, gzErr)
		}
		defer reader.Close()
		if raw, err = io.ReadAll(reader); err != nil {
			return nil, fmt.Errorf("decompress %s: %w", indexFileGz, err)
		}
	}

	start := bytes.Index(raw, []byte(baseHrefStart))
	end := bytes.Index(raw, []byte(baseHrefEnd))
	if start < 0 || end < start {
		return raw, fmt.Errorf(
			"%s has no %s...%s to rewrite",
			indexFile, baseHrefStart, baseHrefEnd)
	}

	base := strings.TrimRight(routePrefix, "/") + "/"
	rewritten := fmt.Appendf(nil,
		`%s<base href="%s" />%s`,
		baseHrefStart, html.EscapeString(base), baseHrefEnd)

	return slices.Concat(raw[:start], rewritten, raw[end+len(baseHrefEnd):]), nil
}

// gzipBytes compresses data at the highest level.
func gzipBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("compress %s: %w", indexFile, err)
	}
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("compress %s: %w", indexFile, err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("compress %s: %w", indexFile, err)
	}

	return buf.Bytes(), nil
}

// overlaidFS serves overlay's entries in place of the underlying file. Only
// [fs.ReadDirFS.Open] is replaced: statigz indexes with ReadDir for the names
// and Open for the bytes, taking each file's size and hash from what it reads.
type overlaidFS struct {
	fs.ReadDirFS
	overlay map[string][]byte
}

func (f overlaidFS) Open(name string) (fs.File, error) {
	if content, ok := f.overlay[name]; ok {
		return &memFile{name: name, Reader: bytes.NewReader(content)}, nil
	}

	return f.ReadDirFS.Open(name) //nolint:wrapcheck
}

// memFile is an [fs.File] over a byte slice. The embedded [bytes.Reader] also
// makes it an [io.ReadSeeker], which is what range requests are served from.
type memFile struct {
	*bytes.Reader
	name string
}

func (f *memFile) Stat() (fs.FileInfo, error) {
	return memFileInfo{name: f.name, size: f.Size()}, nil
}

func (f *memFile) Close() error { return nil }

// memFileInfo is the [fs.FileInfo] of a [memFile].
type memFileInfo struct {
	name string
	size int64
}

func (i memFileInfo) Name() string       { return i.name }
func (i memFileInfo) Size() int64        { return i.size }
func (i memFileInfo) Mode() fs.FileMode  { return 0o444 }
func (i memFileInfo) ModTime() time.Time { return time.Time{} }
func (i memFileInfo) IsDir() bool        { return false }
func (i memFileInfo) Sys() any           { return nil }
