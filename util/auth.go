// Copyright [2025] [Argus]
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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// BasicAuth returns the base64-encoded string of the username and password.
func BasicAuth(username, password string) string {
	encode := fmt.Sprintf("%s:%s",
		username, password)
	return base64.StdEncoding.EncodeToString([]byte(encode))
}

// isHashed returns whether the string represents a hashed value.
func isHashed(s string) bool {
	return RegexCheck("^h__[a-f0-9]{64}$", s)
}

// Hash returns the SHA256 hash of the string.
func hash(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

// HashFromString returns the byte slice of the hash string.
func hashFromString(s string) []byte {
	hash, _ := hex.DecodeString(s)
	return hash
}

// GetHash returns the SHA256 hash of the string.
// If already hashed, it converts it to a byte slice.
func GetHash(s string) [32]byte {
	if isHashed(s) {
		hash := hashFromString(s[3:])
		var hash32 [32]byte
		copy(hash32[:], hash)
		return hash32
	}
	return hash(s)
}

// FmtHash returns the formatted hash string.
func FmtHash(h [32]byte) string {
	return fmt.Sprintf("h__%x", h)
}
