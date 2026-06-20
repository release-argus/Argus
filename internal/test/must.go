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

//go:build unit || integration

package test

import (
	"testing"
)

// Must calls the given function and returns the result, panicking if any error is returned.
func Must[T any](t *testing.T, fn func() (T, error)) T {
	if t != nil {
		t.Helper()
	}

	result, err := fn()
	if err != nil {
		panic("unexpected error: " + err.Error())
	}

	return result
}
