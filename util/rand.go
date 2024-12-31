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

// Package util provides utility functions for the Argus project.
package util

import (
	"crypto/rand"
	"math/big"
)

const alphanumericLower = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandAlphaNumericLower returns a random alphanumeric (lowercase) string of length n.
func RandAlphaNumericLower(n int) string {
	return RandString(n, alphanumericLower)
}

const numeric = "0123456789"

// RandNumeric returns a random numeric string of length n.
func RandNumeric(n int) string {
	return RandString(n, numeric)
}

// RandString returns a random string of length n with `alphabet`.
func RandString(n int, alphabet string) string {
	b := make([]byte, n)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[int(n.Int64())]
	}
	return string(b)
}
