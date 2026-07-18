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

// Package password hashes and verifies user passwords with argon2id,
// using the standard "$argon2id$..." encoded string format so parameters
// can be raised later and detected via rehash-on-login.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Current argon2id parameters.
const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 64 * 1024 // KiB.
	argonThreads uint8  = 1
	saltLength   uint32 = 16
	keyLength    uint32 = 32
)

// maxPasswordLength caps the plaintext password length.
const maxPasswordLength = 1024

// ErrMalformedHash is returned when an encoded hash cannot be parsed.
var ErrMalformedHash = errors.New("malformed password hash")

// ErrPasswordTooLong is returned when a password exceeds [maxPasswordLength].
var ErrPasswordTooLong = errors.New("password too long")

// randRead fills b with cryptographically random bytes (overridable for tests).
// see [rand.Read].
var randRead = rand.Read

// Hash derives an argon2id hash of password, encoded with its parameters.
func Hash(password string) (string, error) {
	if len(password) > maxPasswordLength {
		return "", ErrPasswordTooLong
	}

	salt := make([]byte, saltLength)
	if _, err := randRead(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	key := argon2.IDKey(
		[]byte(password), salt,
		argonTime, argonMemory, argonThreads, keyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify reports whether password matches the encoded hash, and whether the
// hash was derived with outdated parameters (needsRehash).
func Verify(password, encoded string) (match, needsRehash bool, err error) {
	if len(password) > maxPasswordLength {
		return false, false, nil
	}

	version, memory, time, threads, salt, key, err := decodeHash(encoded)
	if err != nil {
		return false, false, err
	}

	computed := argon2.IDKey(
		[]byte(password), salt,
		time, memory, threads, uint32(len(key)),
	)

	match = subtle.ConstantTimeCompare(key, computed) == 1
	needsRehash = version != argon2.Version ||
		memory != argonMemory ||
		time != argonTime ||
		threads != argonThreads ||
		uint32(len(key)) != keyLength

	return match, needsRehash, nil
}

// decodeHash parses a "$argon2id$v=<version>$m=<memory>,t=<time>,p=<threads>$salt$key" string.
func decodeHash(encoded string) (
	version int,
	memory, time uint32, threads uint8,
	salt, key []byte,
	err error,
) {
	fail := func(err error) (int, uint32, uint32, uint8, []byte, []byte, error) {
		return 0, 0, 0, 0, nil, nil, err
	}

	parts := strings.Split(encoded, "$")
	// Format.
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return fail(fmt.Errorf("%w: %q", ErrMalformedHash, encoded))
	}

	// 1. Version.
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return fail(fmt.Errorf("%w: version %q", ErrMalformedHash, parts[2]))
	}
	// 2. Memory, Time, Threads.
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return fail(fmt.Errorf("%w: parameters %q", ErrMalformedHash, parts[3]))
	}
	// 3. Salt.
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fail(fmt.Errorf("%w: salt: %w", ErrMalformedHash, err))
	}
	// 4. Key.
	key, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fail(fmt.Errorf("%w: key: %w", ErrMalformedHash, err))
	}

	return version, memory, time, threads, salt, key, nil
}
