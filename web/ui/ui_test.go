// Copyright [2024] [Argus]
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
	// GIVEN there's a static folder with the web ui inside

	// WHEN the FS is retrieved
	fs := GetFS()

	// THEN those files can be accessed
	fileThatExists := "index.html"
	_, errFileShouldExist := fs.Open(fileThatExists)
	fileThatDoesNotExist := fileThatExists + "a"
	_, errDoesNotExist := fs.Open(fileThatDoesNotExist)
	if errFileShouldExist != nil {
		t.Errorf("%q should exist in FS",
			fileThatExists)
	}
	if errDoesNotExist == nil {
		t.Errorf("%q shouldn't exist in FS",
			fileThatDoesNotExist)
	}
}
