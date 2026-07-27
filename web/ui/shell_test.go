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

package ui

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// packageName identifies this package in test failures.
const packageName = "web_ui"

func TestRewritten(t *testing.T) {
	// GIVEN: an embedded UI, as a dev build (plain shell),
	// or a release one (gzipped shell).
	const shell = `<!DOCTYPE html><html><head>` +
		baseHrefStart + `<base href="/" />` + baseHrefEnd +
		`</head></html>`

	gzipped := func(t *testing.T, body string) []byte {
		t.Helper()
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	tests := []struct {
		name  string
		files fstest.MapFS
		// The names the rewritten shell should be read back from.
		shellFiles []string
		errRegex   string
	}{
		{
			name: "dev build/plain shell",
			files: fstest.MapFS{
				indexFile:    {Data: []byte(shell)},
				"robots.txt": {Data: []byte("some-other-file")},
			},
			shellFiles: []string{indexFile},
		},
		{
			name: "release build/gzipped shell",
			files: fstest.MapFS{
				indexFileGz:     {Data: gzipped(t, shell)},
				"robots.txt.gz": {Data: gzipped(t, "some-other-file")},
			},
			shellFiles: []string{indexFileGz},
		},
		{
			name: "both encodings present/both are rewritten",
			files: fstest.MapFS{
				indexFile:   {Data: []byte(shell)},
				indexFileGz: {Data: gzipped(t, shell)},
			},
			shellFiles: []string{indexFile, indexFileGz},
		},
		{
			name:     "invalid/no shell to overlay",
			files:    fstest.MapFS{"robots.txt": {Data: []byte("some-other-file")}},
			errRegex: `read index.html`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the UI is rewritten for a route prefix.
			fSys, err := rewritten(tc.files, "/argus")

			prefix := fmt.Sprintf("%s\nRewritten()", packageName)

			// THEN: errors match expectations.
			errRegex := util.ValueOr(tc.errRegex, `^$`)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, errRegex,
				)
			}
			if tc.errRegex != "" {
				return
			}

			entries, err := fSys.ReadDir(".")
			if err != nil {
				t.Fatalf(
					"%s ReadDir: %v",
					prefix, err,
				)
			}

			for _, shellFile := range tc.shellFiles {
				// AND: the shell reads back with its <base href> at the prefix.
				got, err := fs.ReadFile(fSys, shellFile)
				if err != nil {
					t.Fatalf(
						"%s read %q: %v",
						prefix, shellFile, err,
					)
				}
				if strings.HasSuffix(shellFile, ".gz") {
					reader, err := gzip.NewReader(bytes.NewReader(got))
					if err != nil {
						t.Fatalf(
							"%s %q is not gzip: %v",
							prefix, shellFile, err,
						)
					}
					if got, err = io.ReadAll(reader); err != nil {
						t.Fatalf(
							"%s %q does not decompress: %v",
							prefix, shellFile, err,
						)
					}
				}
				want := `<base href="/argus/" />`
				if !bytes.Contains(got, []byte(want)) {
					t.Errorf(
						"%s %q mismatch\nwant to contain: %q\ngot:             %q",
						prefix, shellFile, want, got,
					)
				}

				// AND: it is still listed, since statigz indexes by name.
				if !slices.ContainsFunc(entries, func(entry fs.DirEntry) bool {
					return entry.Name() == shellFile
				}) {
					t.Errorf(
						"%s ReadDir() missing %q\ngot: %v",
						prefix, shellFile, entries,
					)
				}
			}

			// AND: every other file passes through untouched.
			for name, file := range tc.files {
				if slices.Contains(tc.shellFiles, name) {
					continue
				}
				through, err := fs.ReadFile(fSys, name)
				if err != nil {
					t.Fatalf(
						"%s read %q: %v",
						prefix, name, err,
					)
				}
				if !bytes.Equal(through, file.Data) {
					t.Errorf(
						"%s %q was modified\ngot:  %q\nwant: %q",
						prefix, name, through, file.Data,
					)
				}
			}
		})
	}
}

func TestRewritten__embedded(t *testing.T) {
	// GIVEN: the UI actually embedded in the binary, rather than a fixture.
	routePrefix := "/argus"

	prefix := fmt.Sprintf("%s\nRewritten()", packageName)

	// WHEN: it is rewritten for a route prefix.
	fSys, err := Rewritten(routePrefix)
	if err != nil {
		t.Fatalf(
			"%s error: %v",
			prefix, err,
		)
	}

	// THEN: the shell serves with its <base href> at the prefix.
	var got []byte
	if got, err = fs.ReadFile(fSys, indexFile); err != nil {
		gzipped, gzErr := fs.ReadFile(fSys, indexFileGz)
		if gzErr != nil {
			t.Fatalf(
				"%s no shell to read: %v",
				prefix, errors.Join(err, gzErr),
			)
		}
		reader, gzErr := gzip.NewReader(bytes.NewReader(gzipped))
		if gzErr != nil {
			t.Fatalf(
				"%s %q is not gzip: %v",
				prefix, indexFileGz, gzErr,
			)
		}
		if got, err = io.ReadAll(reader); err != nil {
			t.Fatalf(
				"%s %q does not decompress: %v",
				prefix, indexFileGz, err,
			)
		}
	}
	want := `<base href="` + routePrefix + `/" />`
	if !bytes.Contains(got, []byte(want)) {
		t.Errorf(
			"%s shell mismatch\nwant to contain: %q\ngot:             %q",
			prefix, want, got,
		)
	}

	// AND: an asset beside it is untouched.
	direct, err := fs.ReadFile(GetFS(), "robots.txt")
	if err != nil {
		t.Fatalf(
			"%s read robots.txt from the embed: %v",
			prefix, err,
		)
	}
	through, err := fs.ReadFile(fSys, "robots.txt")
	if err != nil {
		t.Fatalf(
			"%s read robots.txt: %v",
			prefix, err,
		)
	}
	if !bytes.Equal(through, direct) {
		t.Errorf(
			"%s robots.txt was modified\ngot:  %q\nwant: %q",
			prefix, through, direct,
		)
	}
}

func TestIndexHTML(t *testing.T) {
	// GIVEN: an embedded app shell (plain or gzipped) and a route prefix.
	const shell = `<!DOCTYPE html><html><head>` +
		baseHrefStart + `<base href="/" />` + baseHrefEnd +
		`</head></html>`

	gzipped := func(t *testing.T, body string) []byte {
		t.Helper()
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	tests := []struct {
		name        string
		files       fstest.MapFS
		routePrefix string
		want        string
		errRegex    string
	}{
		{
			name:        "root prefix",
			files:       fstest.MapFS{"index.html": {Data: []byte(shell)}},
			routePrefix: "/",
			want:        `<base href="/" />`,
		},
		{
			name:        "route prefix",
			files:       fstest.MapFS{"index.html": {Data: []byte(shell)}},
			routePrefix: "/argus",
			want:        `<base href="/argus/" />`,
		},
		{
			name:        "route prefix/trailing slash is not doubled",
			files:       fstest.MapFS{"index.html": {Data: []byte(shell)}},
			routePrefix: "/argus/",
			want:        `<base href="/argus/" />`,
		},
		{
			name:        "gzipped shell is decompressed",
			files:       fstest.MapFS{"index.html.gz": {Data: gzipped(t, shell)}},
			routePrefix: "/argus",
			want:        `<base href="/argus/" />`,
		},
		{
			name:        "prefix is escaped into the attribute",
			files:       fstest.MapFS{"index.html": {Data: []byte(shell)}},
			routePrefix: `/a"b`,
			want:        `<base href="/a&#34;b/" />`,
		},
		{
			name:        "markers are kept, so the shell can be rewritten again",
			files:       fstest.MapFS{"index.html": {Data: []byte(shell)}},
			routePrefix: "/argus",
			want:        baseHrefStart + `<base href="/argus/" />` + baseHrefEnd,
		},
		{
			name:        "invalid/no markers to rewrite between",
			files:       fstest.MapFS{"index.html": {Data: []byte("<html></html>")}},
			routePrefix: "/argus",
			errRegex:    `no <!--ARGUS_BASE_HREF_START.*to rewrite`,
		},
		{
			name:        "invalid/only the start marker",
			files:       fstest.MapFS{"index.html": {Data: []byte("<head>" + baseHrefStart + "</head>")}},
			routePrefix: "/argus",
			errRegex:    `no <!--ARGUS_BASE_HREF_START.*to rewrite`,
		},
		{
			name:        "invalid/gzipped shell is not gzip",
			files:       fstest.MapFS{indexFileGz: {Data: []byte("not-actually-gzip")}},
			routePrefix: "/argus",
			errRegex:    `decompress index.html.gz`,
		},
		{
			name: "invalid/gzipped shell is truncated",
			files: fstest.MapFS{
				// A valid gzip header, then nothing: the reader opens, the read fails.
				indexFileGz: {Data: []byte{0x1f, 0x8b, 0x08, 0, 0, 0, 0, 0, 0, 0xff}},
			},
			routePrefix: "/argus",
			errRegex:    `decompress index.html.gz`,
		},
		{
			name:        "invalid/no index.html at all",
			files:       fstest.MapFS{},
			routePrefix: "/argus",
			errRegex:    `read index.html`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: indexHTML is called.
			got, err := indexHTML(tc.files, tc.routePrefix)

			// THEN: errors match expectations.
			errRegex := util.ValueOr(tc.errRegex, `^$`)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s\nindexHTML(%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.routePrefix, e, errRegex,
				)
			}
			if tc.errRegex != "" {
				return
			}

			// AND: the base href points at the route prefix.
			if !strings.Contains(string(got), tc.want) {
				t.Errorf(
					"%s\nindexHTML() mismatch\nwant to contain: %q\ngot:             %q",
					packageName, tc.want, got,
				)
			}
		})
	}
}

func TestIndexHTML__embedded(t *testing.T) {
	// GIVEN: the index.html actually embedded in the binary, rather than a fixture.

	// WHEN: it is rewritten for a route prefix.
	got, err := indexHTML(GetFS(), "/argus")

	// THEN: the rewrite finds its markers.
	if err != nil {
		t.Fatalf("%s\nindexHTML(embedded) error: %v", packageName, err)
	}

	// AND: the base href points at the prefix.
	want := `<base href="/argus/" />`
	if !bytes.Contains(got, []byte(want)) {
		t.Errorf(
			"%s\nindexHTML(embedded) mismatch\nwant to contain: %q\ngot:             %q",
			packageName, want, got,
		)
	}
}

func TestMemFileInfo(t *testing.T) {
	// GIVEN: a shell served out of the overlay.
	const body = "<html>the shell</html>"
	fSys := overlaidFS{
		ReadDirFS: fstest.MapFS{indexFile: {Data: []byte("the embedded one")}},
		overlay:   map[string][]byte{indexFile: []byte(body)},
	}

	// WHEN: its [fs.FileInfo] is read.
	file, err := fSys.Open(indexFile)
	if err != nil {
		t.Fatalf("%s\noverlaidFS.Open() error: %v", packageName, err)
	}
	t.Cleanup(func() { file.Close() })
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("%s\nmemFile.Stat() error: %v", packageName, err)
	}

	// THEN: it describes the overlaid content, not the file underneath.
	if got := info.Name(); got != indexFile {
		t.Errorf("%s\nmemFileInfo.Name() mismatch\ngot:  %q\nwant: %q",
			packageName, got, indexFile)
	}
	if got, want := info.Size(), int64(len(body)); got != want {
		t.Errorf("%s\nmemFileInfo.Size() mismatch\ngot:  %d\nwant: %d",
			packageName, got, want)
	}
	if got := info.Mode(); got != 0o444 {
		t.Errorf("%s\nmemFileInfo.Mode() mismatch\ngot:  %v\nwant: %v",
			packageName, got, fs.FileMode(0o444))
	}
	// Zero, so statigz falls back to its own hash for caching rather than a
	// modtime that would change on every restart.
	if got := info.ModTime(); !got.IsZero() {
		t.Errorf("%s\nmemFileInfo.ModTime() mismatch\ngot:  %v\nwant: the zero time",
			packageName, got)
	}
	if info.IsDir() {
		t.Errorf("%s\nmemFileInfo.IsDir() mismatch\ngot:  true\nwant: false",
			packageName)
	}
	if got := info.Sys(); got != nil {
		t.Errorf("%s\nmemFileInfo.Sys() mismatch\ngot:  %v\nwant: nil",
			packageName, got)
	}
}

func TestMemFile__seek(t *testing.T) {
	// GIVEN: a shell served out of the overlay - range requests are served by
	// seeking it.
	const body = "0123456789"
	fSys := overlaidFS{
		ReadDirFS: fstest.MapFS{},
		overlay:   map[string][]byte{indexFile: []byte(body)},
	}
	file, err := fSys.Open(indexFile)
	if err != nil {
		t.Fatalf(
			"%s\noverlaidFS.Open() error: %v",
			packageName, err,
		)
	}
	t.Cleanup(func() { file.Close() })

	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		t.Fatalf("%s\nmemFile is not an io.ReadSeeker", packageName)
	}

	// WHEN: it is seeked into and read from.
	if _, err := seeker.Seek(4, io.SeekStart); err != nil {
		t.Fatalf(
			"%s\nmemFile.Seek() error: %v",
			packageName, err,
		)
	}
	got, err := io.ReadAll(seeker)
	if err != nil {
		t.Fatalf(
			"%s\nmemFile read after seek: %v",
			packageName, err,
		)
	}

	// THEN: the remainder is returned.
	if want := body[4:]; string(got) != want {
		t.Errorf(
			"%s\nmemFile read after seek mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}
