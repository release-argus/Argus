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

// Package httpx provides a HTTP client.
package httpx

import (
	"fmt"
	"io"
)

// MaxBodyBytes bounds a response body read into memory.
const MaxBodyBytes = 50 << 20

// ReadBody returns the bytes of body, up to [MaxBodyBytes].
//
// A body over the cap is reported rather than truncated, as a truncated one would
// surface as a parse failure that says nothing about the size. Closing body is left
// to the caller.
func ReadBody(body io.Reader) ([]byte, error) {
	read, err := io.ReadAll(io.LimitReader(body, MaxBodyBytes+1))
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if len(read) > MaxBodyBytes {
		return nil, fmt.Errorf(
			"response body is larger than the %d MiB limit",
			MaxBodyBytes>>20,
		)
	}

	return read, nil
}
