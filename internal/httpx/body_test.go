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

// errReader fails partway through, as a dropped connection would.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

func TestReadBody(t *testing.T) {
	// GIVEN: a response body.
	tests := []struct {
		name     string
		body     string
		reader   func() interface{ Read([]byte) (int, error) }
		wantLen  int
		errRegex string
	}{
		{
			name:     "read/empty",
			body:     "",
			wantLen:  0,
			errRegex: `^$`,
		},
		{
			name:     "read/small document",
			body:     `{"version":"1.2.3"}`,
			wantLen:  19,
			errRegex: `^$`,
		},
		{
			name:     "read/exactly at the limit",
			body:     strings.Repeat("a", MaxBodyBytes),
			wantLen:  MaxBodyBytes,
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

			// WHEN: ReadBody is called with it.
			got, err := ReadBody(reader)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nReadBody() error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}
			if err != nil {
				// AND: nothing is handed back alongside an error.
				if got != nil {
					t.Errorf(
						"%s\nReadBody() returned %d bytes alongside an error\nwant: nil",
						packageName, len(got),
					)
				}
				return
			}

			// AND: the whole body came back when no error returned.
			if g, w := len(got), tc.wantLen; g != w {
				t.Fatalf(
					"%s\nReadBody() length mismatch\ngot:  %d\nwant: %d",
					packageName, g, w,
				)
			}
		})
	}
}
