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

package test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{name: "leading newline trimmed", data: "\nkey: value\n", want: "key: value\n"},
		{name: "only one leading newline trimmed", data: "\n\nkey: value", want: "\nkey: value"},
		{name: "no leading newline", data: "key: value", want: "key: value"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "file.yml")

			WriteFile(t, path, tc.data)

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("os.ReadFile() failed: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("file content mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestWriteFile__fails(t *testing.T) {
	var fakeT FakeT
	fakeT.Abort = true
	path := filepath.Join(t.TempDir(), "missing-dir", "file.yml")

	aborted := fakeT.Aborted(func() { WriteFile(&fakeT, path, "x") })

	if !aborted || len(fakeT.Fatals) != 1 {
		t.Errorf(
			"FakeT mismatch\ngot:  aborted=%t, fatals=%q\nwant: aborted=true, fatals=1",
			aborted, fakeT.Fatals,
		)
	}
}
