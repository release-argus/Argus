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

package httpx

import (
	"errors"
	"strings"
	"testing"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// errReader fails on the first read, as a dropped connection would.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

func TestReadBody(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		reader   func() interface{ Read([]byte) (int, error) }
		errRegex string
	}{
		{
			name:     "read/empty",
			body:     "",
			errRegex: `^$`,
		},
		{
			name:     "read/small document",
			body:     `{"version":"1.2.3"}`,
			errRegex: `^$`,
		},
		{
			name:     "read/exactly at the limit",
			body:     strings.Repeat("a", MaxBodyBytes),
			errRegex: `^$`,
		},
		{
			name:     "refused/one byte over the limit is reported, not truncated",
			body:     strings.Repeat("a", MaxBodyBytes+1),
			errRegex: `^response body is larger than the 50 MiB limit$`,
		},
		{
			name:     "refused/far over the limit",
			body:     strings.Repeat("a", MaxBodyBytes*2),
			errRegex: `^response body is larger than the 50 MiB limit$`,
		},
		{
			name: "error/the read itself fails",
			reader: func() interface{ Read([]byte) (int, error) } {
				return errReader{err: errors.New("connection reset")}
			},
			errRegex: `^connection reset$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var reader interface{ Read([]byte) (int, error) } = strings.NewReader(tc.body)
			if tc.reader != nil {
				reader = tc.reader()
			}

			got, err := ReadBody(reader)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				if got != nil {
					t.Errorf("body mismatch\ngot:  %d bytes\nwant: nil", len(got))
				}
				return
			}

			if string(got) != tc.body {
				t.Fatalf(
					"body mismatch\ngot:  %d bytes, %.40q\nwant: %d bytes, %.40q",
					len(got), got,
					len(tc.body), tc.body,
				)
			}
		})
	}
}
