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
	"testing"
)

func TestGetFS(t *testing.T) {
	fs := GetFS()

	f, err := fs.Open("index.html")
	if err != nil {
		t.Errorf("exists: ReadDirFS.Open() failed: %v", err)
	} else {
		_ = f.Close()
	}

	if f, err := fs.Open("index.htmla"); err == nil {
		_ = f.Close()
		t.Error("missing: ReadDirFS.Open() error mismatch\ngot:  nil\nwant: non-nil")
	}
}
