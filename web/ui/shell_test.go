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
	"io"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// statBlindFS serves its files but cannot stat them, so the shell reads back
// while the overlay finds nothing to replace.
type statBlindFS struct {
	fs.ReadDirFS
}

func (statBlindFS) Stat(string) (fs.FileInfo, error) { return nil, fs.ErrNotExist }

// testShell is an app shell with its <base href> at the root.
const testShell = `<!DOCTYPE html><html><head>` +
	baseHrefStart + `<base href="/" />` + baseHrefEnd +
	`</head></html>`

// gzipFixture returns body gzipped.
func gzipFixture(t *testing.T, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write([]byte(body)); err != nil {
		t.Fatalf("Writer.Write() failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Writer.Close() failed: %v", err)
	}
	return buf.Bytes()
}

func gunzipFixture(t *testing.T, data []byte) []byte {
	t.Helper()
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip.NewReader() failed: %v", err)
	}
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll() failed: %v", err)
	}
	return got
}

func TestRewritten__nothingToOverlay(t *testing.T) {
	// A shell that cannot be stat-ed leaves the overlay nothing to replace.
	fSys := statBlindFS{
		ReadDirFS: fstest.MapFS{indexFile: {Data: []byte(testShell)}},
	}

	got, err := rewritten(fSys, "/argus")
	e := errfmt.FormatError(err)
	errRegex := `no index.html to overlay`
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}

	// Unchanged, so it still serves from the root.
	shell, err := fs.ReadFile(got, indexFile)
	if err != nil {
		t.Fatalf("fs.ReadFile() failed: %v", err)
	}
	if string(shell) != testShell {
		t.Errorf("shell mismatch\ngot:  %q\nwant: %q", shell, testShell)
	}
}

func TestRewritten(t *testing.T) {
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
				indexFile:    {Data: []byte(testShell)},
				"robots.txt": {Data: []byte("some-other-file")},
			},
			shellFiles: []string{indexFile},
			errRegex:   `^$`,
		},
		{
			name: "release build/gzipped shell",
			files: fstest.MapFS{
				indexFileGz:     {Data: gzipFixture(t, testShell)},
				"robots.txt.gz": {Data: gzipFixture(t, "some-other-file")},
			},
			shellFiles: []string{indexFileGz},
			errRegex:   `^$`,
		},
		{
			name: "both encodings present/both are rewritten",
			files: fstest.MapFS{
				indexFile:   {Data: []byte(testShell)},
				indexFileGz: {Data: gzipFixture(t, testShell)},
			},
			shellFiles: []string{indexFile, indexFileGz},
			errRegex:   `^$`,
		},
		{
			name:     "invalid/no shell to read",
			files:    fstest.MapFS{"robots.txt": {Data: []byte("some-other-file")}},
			errRegex: `read index.html`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fSys, err := rewritten(tc.files, "/argus")
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			for name, file := range tc.files {
				if slices.Contains(tc.shellFiles, name) {
					continue
				}
				through, readErr := fs.ReadFile(fSys, name)
				if readErr != nil {
					t.Fatalf("fs.ReadFile(%q) failed: %v", name, readErr)
				}
				if !bytes.Equal(through, file.Data) {
					t.Errorf(
						"fs.ReadFile(%q) mismatch\ngot:  %q\nwant: %q",
						name, through, file.Data,
					)
				}
			}
			if err != nil {
				return
			}

			entries, err := fSys.ReadDir(".")
			if err != nil {
				t.Fatalf("ReadDirFS.ReadDir() failed: %v", err)
			}

			for _, shellFile := range tc.shellFiles {
				t.Run(shellFile, func(t *testing.T) {
					t.Parallel()

					got, err := fs.ReadFile(fSys, shellFile)
					if err != nil {
						t.Fatalf("fs.ReadFile() failed: %v", err)
					}
					if strings.HasSuffix(shellFile, ".gz") {
						got = gunzipFixture(t, got)
					}
					want := `<base href="/argus/" />`
					if !bytes.Contains(got, []byte(want)) {
						t.Errorf("shell mismatch\ngot:  %q\nwant: %q", got, want)
					}

					// statigz indexes by name, so the shell must stay listed.
					if !slices.ContainsFunc(entries, func(entry fs.DirEntry) bool {
						return entry.Name() == shellFile
					}) {
						names := make([]string, len(entries))
						for i, entry := range entries {
							names[i] = entry.Name()
						}
						t.Errorf(
							"ReadDirFS.ReadDir() mismatch\ngot:  %q\nwant: %q",
							names, shellFile,
						)
					}
				})
			}
		})
	}
}

func TestRewritten__embedded(t *testing.T) {
	routePrefix := "/argus"

	fSys, err := Rewritten(routePrefix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []byte
	if got, err = fs.ReadFile(fSys, indexFile); err != nil {
		gzipped, gzErr := fs.ReadFile(fSys, indexFileGz)
		if gzErr != nil {
			t.Fatalf("fs.ReadFile() failed: %v", errors.Join(err, gzErr))
		}
		got = gunzipFixture(t, gzipped)
	}
	want := `<base href="` + routePrefix + `/" />`
	if !bytes.Contains(got, []byte(want)) {
		t.Errorf("shell mismatch\ngot:  %q\nwant: %q", got, want)
	}

	direct, err := fs.ReadFile(GetFS(), "robots.txt")
	if err != nil {
		t.Fatalf("direct: fs.ReadFile() failed: %v", err)
	}
	through, err := fs.ReadFile(fSys, "robots.txt")
	if err != nil {
		t.Fatalf("through: fs.ReadFile() failed: %v", err)
	}
	if !bytes.Equal(through, direct) {
		t.Errorf("robots.txt mismatch\ngot:  %q\nwant: %q", through, direct)
	}
}

func TestIndexHTML(t *testing.T) {
	tests := []struct {
		name        string
		files       fstest.MapFS
		routePrefix string
		want        string
		errRegex    string
	}{
		{
			name:        "root prefix",
			files:       fstest.MapFS{indexFile: {Data: []byte(testShell)}},
			routePrefix: "/",
			want:        `<base href="/" />`,
			errRegex:    `^$`,
		},
		{
			name:        "route prefix",
			files:       fstest.MapFS{indexFile: {Data: []byte(testShell)}},
			routePrefix: "/argus",
			want:        `<base href="/argus/" />`,
			errRegex:    `^$`,
		},
		{
			name:        "route prefix/trailing slash is not doubled",
			files:       fstest.MapFS{indexFile: {Data: []byte(testShell)}},
			routePrefix: "/argus/",
			want:        `<base href="/argus/" />`,
			errRegex:    `^$`,
		},
		{
			name:        "gzipped shell is decompressed",
			files:       fstest.MapFS{indexFileGz: {Data: gzipFixture(t, testShell)}},
			routePrefix: "/argus",
			want:        `<base href="/argus/" />`,
			errRegex:    `^$`,
		},
		{
			name:        "prefix is escaped into the attribute",
			files:       fstest.MapFS{indexFile: {Data: []byte(testShell)}},
			routePrefix: `/a"b`,
			want:        `<base href="/a&#34;b/" />`,
			errRegex:    `^$`,
		},
		{
			name:        "markers are kept, so the shell can be rewritten again",
			files:       fstest.MapFS{indexFile: {Data: []byte(testShell)}},
			routePrefix: "/argus",
			want:        baseHrefStart + `<base href="/argus/" />` + baseHrefEnd,
			errRegex:    `^$`,
		},
		{
			name:        "invalid/no markers to rewrite between",
			files:       fstest.MapFS{indexFile: {Data: []byte("<html></html>")}},
			routePrefix: "/argus",
			errRegex:    `no <!--ARGUS_BASE_HREF_START.*to rewrite`,
		},
		{
			name:        "invalid/only the start marker",
			files:       fstest.MapFS{indexFile: {Data: []byte("<head>" + baseHrefStart + "</head>")}},
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

			got, err := indexHTML(tc.files, tc.routePrefix)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if !strings.Contains(string(got), tc.want) {
				t.Errorf("shell mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestIndexHTML__embedded(t *testing.T) {
	got, err := indexHTML(GetFS(), "/argus")

	// Fails if react-app/index.html loses the markers.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `<base href="/argus/" />`
	if !bytes.Contains(got, []byte(want)) {
		t.Errorf("shell mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestGzipBytes(t *testing.T) {
	gzipped, err := gzipBytes([]byte(testShell))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := gunzipFixture(t, gzipped)
	if string(got) != testShell {
		t.Errorf("shell mismatch\ngot:  %q\nwant: %q", got, testShell)
	}

	// gzip writes XFL 2 only for BestCompression.
	if xfl, want := gzipped[8], byte(2); xfl != want {
		t.Errorf("gzipped[8] mismatch\ngot:  %d\nwant: %d", xfl, want)
	}
}

func TestMemFile__seek(t *testing.T) {
	const body = "0123456789"
	fSys := overlaidFS{
		ReadDirFS: fstest.MapFS{},
		overlay:   map[string][]byte{indexFile: []byte(body)},
	}
	file, err := fSys.Open(indexFile)
	if err != nil {
		t.Fatalf("overlaidFS.Open() failed: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })

	// Range requests seek into the overlay.
	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		t.Fatalf("overlaidFS.Open() type mismatch\ngot:  %T\nwant: io.ReadSeeker", file)
	}

	offset := int64(4)
	_, err = seeker.Seek(offset, io.SeekStart)
	if err != nil {
		t.Fatalf("ReadSeeker.Seek() failed: %v", err)
	}
	got, err := io.ReadAll(seeker)
	if err != nil {
		t.Fatalf("io.ReadAll() failed: %v", err)
	}

	if want := body[offset:]; string(got) != want {
		t.Errorf("io.ReadAll() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestMemFileInfo(t *testing.T) {
	// The embedded file differs in size, so Size() tells the two apart.
	const body = "<html>the shell</html>"
	fSys := overlaidFS{
		ReadDirFS: fstest.MapFS{indexFile: {Data: []byte("the embedded one")}},
		overlay:   map[string][]byte{indexFile: []byte(body)},
	}

	file, err := fSys.Open(indexFile)
	if err != nil {
		t.Fatalf("overlaidFS.Open() failed: %v", err)
	}
	t.Cleanup(func() { _ = file.Close() })
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("memFile.Stat() failed: %v", err)
	}

	if got := info.Name(); got != indexFile {
		t.Errorf("memFileInfo.Name() mismatch\ngot:  %q\nwant: %q", got, indexFile)
	}
	if got, want := info.Size(), int64(len(body)); got != want {
		t.Errorf("memFileInfo.Size() mismatch\ngot:  %d\nwant: %d", got, want)
	}
	if got := info.Mode(); got != 0o444 {
		t.Errorf("memFileInfo.Mode() mismatch\ngot:  %v\nwant: %v", got, fs.FileMode(0o444))
	}
	// statigz ETags on its own content hash, so it never reads ModTime.
	if got := info.ModTime(); !got.IsZero() {
		t.Errorf("memFileInfo.ModTime() mismatch\ngot:  %v\nwant: %v", got, time.Time{})
	}
	if info.IsDir() {
		t.Error("memFileInfo.IsDir() mismatch\ngot:  true\nwant: false")
	}
	if got := info.Sys(); got != nil {
		t.Errorf("memFileInfo.Sys() mismatch\ngot:  %v\nwant: nil", got)
	}
}
