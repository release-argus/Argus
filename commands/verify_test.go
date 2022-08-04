// Copyright [2022] [Argus]
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

package command

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		slice *Slice
		lines int
	}{
		"nil Slice":                 {lines: 0, slice: nil},
		"non-nil zero length Slice": {lines: 0, slice: &Slice{}},
		"single arg Command":        {lines: 2, slice: &Slice{{"ls"}}},
		"single multi-arg Command":  {lines: 2, slice: &Slice{{"ls", "-lah", "/root"}}},
		"multiple Commands":         {lines: 4, slice: &Slice{{"ls"}, {"true"}, {"bash", "something.sh"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called
			tc.slice.Print("")

			// THEN it prints the expected number of lines
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
		})
	}
}
