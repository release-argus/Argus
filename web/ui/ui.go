// Copyright [2022] [Hymenaios]
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
	"embed"
	"io/fs"
)

//go:embed static
var EmbedFS embed.FS

// GetFS returns the embedded React files.
func GetFS() fs.FS {
	fsys, err := fs.Sub(EmbedFS, "static")
	if err != nil {
		panic(err)
	}

	return fsys
}
