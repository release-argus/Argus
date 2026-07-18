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

package password

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// legacyHash encodes password with deliberately outdated argon2id parameters.
func legacyHash(
	password string,
	memory, time uint32,
	threads uint8,
	keyLen uint32,
) string {
	salt := []byte("0123456789abcdef")
	key := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory, time, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
}

func TestHash(t *testing.T) {
	// GIVEN: a password.
	plaintext := "s3cret-p4ssw0rd"

	prefix := fmt.Sprintf("%s\nHash(%q)", packageName, plaintext)

	// WHEN: Hash is called twice.
	first, errFirst := Hash(plaintext)
	second, errSecond := Hash(plaintext)

	// THEN: both calls succeed.
	if errFirst != nil || errSecond != nil {
		t.Fatalf(
			"%s unexpected errors\nfirst:  %v\nsecond: %v",
			prefix, errFirst, errSecond,
		)
	}

	// AND: the encoding carries the current parameters.
	wantPrefix := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$",
		argon2.Version, argonMemory, argonTime, argonThreads,
	)
	for _, tc := range []struct {
		name, hash string
	}{
		{"first", first},
		{"second", second},
	} {
		if !strings.HasPrefix(tc.hash, wantPrefix) {
			t.Errorf(
				"%s encoding prefix mismatch (%s hash)\ngot:  %q\nwant: %q",
				prefix, tc.name, tc.hash, wantPrefix,
			)
		}
	}

	// AND: salts are random, so the hashes differ.
	if first == second {
		t.Errorf(
			"%s two hashes of the same password should differ (random salt)\ngot the same: %q",
			prefix, first,
		)
	}

	// AND: the hash verifies against the original password.
	for _, tc := range []struct {
		name, hash string
	}{
		{"first", first},
		{"second", second},
	} {
		match, needsRehash, err := Verify(plaintext, tc.hash)
		if err != nil || !match || needsRehash {
			t.Errorf(
				"%s Verify(%s) mismatch\ngot:  match=%t, needsRehash=%t, err=%v\nwant: match=true, needsRehash=false, err=nil",
				prefix, tc.name, match, needsRehash, err,
			)
		}
	}
}

func TestHash__randRead_error(t *testing.T) {
	// GIVEN: a failing random source.
	wantErr := errors.New("rand broke")
	randReadHad := randRead
	randRead = func(_ []byte) (int, error) { return 0, wantErr }
	t.Cleanup(func() { randRead = randReadHad })

	prefix := fmt.Sprintf("%s\nHash() with failing rand", packageName)

	// WHEN: Hash is called.
	_, err := Hash("anything")

	// THEN: the error is surfaced.
	if !errors.Is(err, wantErr) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, wantErr,
		)
	}
}

func TestHash__tooLong(t *testing.T) {
	// GIVEN: a password over the length cap.
	tooLong := strings.Repeat("a", maxPasswordLength+1)

	prefix := fmt.Sprintf("%s\nHash() with an over-long password", packageName)

	// WHEN: Hash is called.
	_, err := Hash(tooLong)

	// THEN: it is rejected, and Verify treats it as a non-match.
	if !errors.Is(err, ErrPasswordTooLong) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, ErrPasswordTooLong,
		)
	}
	if match, _, err := Verify(tooLong, "$argon2id$anything"); match || err != nil {
		t.Errorf(
			"%s Verify(too long) should be a clean non-match\ngot:  match=%t, err=%v\nwant: match=false, err=nil",
			prefix, match, err,
		)
	}
}

func TestVerify(t *testing.T) {
	// GIVEN: hashes of a known password in various states.
	plaintext := "correct horse battery staple"
	current, err := Hash(plaintext)
	if err != nil {
		t.Fatalf(
			"%s\nsetup Hash failed: %v",
			packageName, err,
		)
	}

	tests := []struct {
		name            string
		password        string
		encoded         string
		wantMatch       bool
		wantNeedsRehash bool
		errRegex        string
	}{
		{
			name:     "empty password",
			password: "",
			encoded:  current,
		},
		{
			name:     "wrong password",
			password: "incorrect",
			encoded:  current,
		},
		{
			name:      "correct password/current params",
			password:  plaintext,
			encoded:   current,
			wantMatch: true,
		},
		{
			name:     "correct password/outdated memory param",
			password: plaintext,
			encoded: legacyHash(
				plaintext,
				argonMemory/2,
				argonTime,
				argonThreads,
				keyLength,
			),
			wantMatch:       true,
			wantNeedsRehash: true,
		},
		{
			name:     "correct password/outdated time param",
			password: plaintext,
			encoded: legacyHash(
				plaintext,
				argonMemory,
				argonTime+1,
				argonThreads,
				keyLength,
			),
			wantMatch:       true,
			wantNeedsRehash: true,
		},
		{
			name:     "correct password/outdated thread param",
			password: plaintext,
			encoded: legacyHash(plaintext,
				argonMemory,
				argonTime,
				argonThreads+1,
				keyLength,
			),
			wantMatch:       true,
			wantNeedsRehash: true,
		},
		{
			name:     "correct password/outdated key length",
			password: plaintext,
			encoded: legacyHash(plaintext,
				argonMemory,
				argonTime,
				argonThreads,
				keyLength/2,
			),
			wantMatch:       true,
			wantNeedsRehash: true,
		},
		{
			name:     "wrong password/outdated params",
			password: "incorrect",
			encoded: legacyHash(
				plaintext,
				argonMemory/2,
				argonTime,
				argonThreads,
				keyLength,
			),
			wantNeedsRehash: true,
		},
		{
			name:     "malformed/empty",
			password: plaintext,
			encoded:  "",
			errRegex: `^malformed password hash: ""`,
		},
		{
			name:     "malformed/not a hash",
			password: plaintext,
			encoded:  "not-a-hash",
			errRegex: `^malformed password hash: "not-a-hash"`,
		},
		{
			name:     "malformed/wrong algorithm",
			password: plaintext,
			encoded:  "$bcrypt$v=19$m=19456,t=1,p=2$c2FsdA$a2V5",
			errRegex: `^malformed password hash: "\$bcrypt`,
		},
		{
			name:     "malformed/bad version",
			password: plaintext,
			encoded:  "$argon2id$vX$m=19456,t=1,p=2$c2FsdA$a2V5",
			errRegex: `^malformed password hash: version "vX"`,
		},
		{
			name:     "malformed/bad parameters",
			password: plaintext,
			encoded:  "$argon2id$v=19$mem=19456$c2FsdA$a2V5",
			errRegex: `^malformed password hash: parameters "mem=19456"`,
		},
		{
			name:     "malformed/bad salt base64",
			password: plaintext,
			encoded:  "$argon2id$v=19$m=19456,t=1,p=2$!!!$a2V5",
			errRegex: `^malformed password hash\n\s*illegal base64 data`,
		},
		{
			name:     "malformed/bad key base64",
			password: plaintext,
			encoded:  "$argon2id$v=19$m=19456,t=1,p=2$c2FsdA$!!!",
			errRegex: `^malformed password hash\n\s*illegal base64 data`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Verify is called.
			match, needsRehash, err := Verify(tc.password, tc.encoded)

			prefix := fmt.Sprintf(
				"%s\nVerify(password=%q, encoded=%q)",
				packageName, tc.password, tc.encoded,
			)

			// THEN: errors match expectations.
			errRegex := tc.errRegex
			if errRegex == "" {
				errRegex = `^$`
			}
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, errRegex,
				)
			}
			if err != nil && !errors.Is(err, ErrMalformedHash) {
				t.Errorf(
					"%s error should wrap ErrMalformedHash\ngot: %v",
					prefix, err,
				)
			}

			// AND: match matches expectations.
			if match != tc.wantMatch {
				t.Errorf(
					"%s match mismatch\ngot:  %t\nwant: %t",
					prefix, match, tc.wantMatch,
				)
			}

			// AND: needsRehash matches expectations.
			if needsRehash != tc.wantNeedsRehash {
				t.Errorf(
					"%s needsRehash mismatch\ngot:  %t\nwant: %t",
					prefix, needsRehash, tc.wantNeedsRehash,
				)
			}
		})
	}
}

// encodeParts assembles an argon2id encoding from already-encoded salt/key fields.
func encodeParts(
	version int,
	memory, time uint32,
	threads uint8,
	salt, key string,
) string {
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		version, memory, time, threads, salt, key,
	)
}

func TestDecodeHash(t *testing.T) {
	// GIVEN: an encoded argon2id hash string.
	b64 := base64.RawStdEncoding.EncodeToString
	salt := bytes.Repeat([]byte("S"), int(saltLength))
	key := bytes.Repeat([]byte("K"), int(keyLength))
	minSalt := bytes.Repeat([]byte("S"), int(saltLength/2))
	minKey := bytes.Repeat([]byte("K"), int(keyLength/2))
	shortSalt := bytes.Repeat([]byte("S"), int(saltLength/2)-1)
	shortKey := bytes.Repeat([]byte("K"), int(keyLength/2)-1)
	longSalt := bytes.Repeat([]byte("S"), int(saltLength)*2)
	longKey := bytes.Repeat([]byte("K"), int(keyLength)*2)

	tests := []struct {
		name        string
		encoded     string
		wantVersion int
		wantMemory  uint32
		wantTime    uint32
		wantThreads uint8
		wantSalt    []byte
		wantKey     []byte
		errRegex    string
	}{
		{
			name:     "invalid/empty string",
			encoded:  "",
			errRegex: `^malformed password hash: ""`,
		},
		{
			name:     "invalid/no separators",
			encoded:  "not-a-hash",
			errRegex: `^malformed password hash: "not-a-hash"`,
		},
		{
			name:     "invalid/too few parts",
			encoded:  "$argon2id$v=19$m=65536,t=2,p=1$" + b64(salt),
			errRegex: `^malformed password hash: "\$argon2id`,
		},
		{
			name: "invalid/too many parts",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(salt), b64(key),
			) + "$extra",
			errRegex: `^malformed password hash: "\$argon2id`,
		},
		{
			name: "valid/current parameters",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(salt), b64(key),
			),
			wantVersion: argon2.Version,
			wantMemory:  argonMemory,
			wantTime:    argonTime,
			wantThreads: argonThreads,
			wantSalt:    salt,
			wantKey:     key,
		},
		{
			name: "valid/older version",
			encoded: encodeParts(
				16, argonMemory, argonTime, argonThreads, b64(salt), b64(key),
			),
			wantVersion: 16,
			wantMemory:  argonMemory,
			wantTime:    argonTime,
			wantThreads: argonThreads,
			wantSalt:    salt,
			wantKey:     key,
		},
		{
			name: "valid/zeroed parameters are not validated",
			encoded: encodeParts(
				0, 0, 0, 0, b64(salt), b64(key),
			),
			wantSalt: salt,
			wantKey:  key,
		},
		{
			name: "invalid/salt empty",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, "", b64(key),
			),
			errRegex: `^malformed password hash: salt/key too short`,
		},
		{
			name: "invalid/salt not base64",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, "!!!", b64(key),
			),
			errRegex: `^malformed password hash\n\s*illegal base64 data`,
		},
		{
			name: "invalid/salt too short",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(shortSalt), b64(key),
			),
			errRegex: `^malformed password hash: salt/key too short`,
		},
		{
			name: "invalid/key empty",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(salt), "",
			),
			errRegex: `^malformed password hash: salt/key too short`,
		},
		{
			name: "invalid/key not base64",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(salt), "!!!",
			),
			errRegex: `^malformed password hash\n\s*illegal base64 data`,
		},
		{
			name: "invalid/key too short",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(salt), b64(shortKey),
			),
			errRegex: `^malformed password hash: salt/key too short`,
		},
		{
			name: "valid/salt and key at minimum length",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(minSalt), b64(minKey),
			),
			wantVersion: argon2.Version,
			wantMemory:  argonMemory,
			wantTime:    argonTime,
			wantThreads: argonThreads,
			wantSalt:    minSalt,
			wantKey:     minKey,
		},
		{
			name: "valid/salt and key over the expected length",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, argonThreads, b64(longSalt), b64(longKey),
			),
			wantVersion: argon2.Version,
			wantMemory:  argonMemory,
			wantTime:    argonTime,
			wantThreads: argonThreads,
			wantSalt:    longSalt,
			wantKey:     longKey,
		},
		{
			name: "valid/maximum threads",
			encoded: encodeParts(
				argon2.Version, argonMemory, argonTime, 255, b64(salt), b64(key),
			),
			wantVersion: argon2.Version,
			wantMemory:  argonMemory,
			wantTime:    argonTime,
			wantThreads: 255,
			wantSalt:    salt,
			wantKey:     key,
		},
		{
			name:     "invalid/threads overflows uint8",
			encoded:  "$argon2id$v=19$m=65536,t=2,p=256$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: parameters "m=65536,t=2,p=256"`,
		},
		{
			name:     "invalid/missing leading separator",
			encoded:  "argon2id$v=19$m=65536,t=2,p=1$" + b64(salt) + "$" + b64(key) + "$",
			errRegex: `^malformed password hash: "argon2id`,
		},
		{
			name:     "invalid/wrong algorithm",
			encoded:  "$argon2i$v=19$m=65536,t=2,p=1$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: "\$argon2i\$`,
		},
		{
			name:     "invalid/version value not numeric",
			encoded:  "$argon2id$v=abc$m=65536,t=2,p=1$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: version "v=abc"`,
		},
		{
			name:     "invalid/version empty",
			encoded:  "$argon2id$$m=65536,t=2,p=1$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: version ""`,
		},
		{
			name:     "invalid/parameters mislabelled",
			encoded:  "$argon2id$v=19$mem=65536$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: parameters "mem=65536"`,
		},
		{
			name:     "invalid/parameters missing threads",
			encoded:  "$argon2id$v=19$m=65536,t=2$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: parameters "m=65536,t=2"`,
		},
		{
			name:     "invalid/parameters out of order",
			encoded:  "$argon2id$v=19$t=2,m=65536,p=1$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: parameters "t=2,m=65536,p=1"`,
		},
		{
			name:     "invalid/negative memory",
			encoded:  "$argon2id$v=19$m=-1,t=2,p=1$" + b64(salt) + "$" + b64(key),
			errRegex: `^malformed password hash: parameters "m=-1,t=2,p=1"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: decodeHash is called.
			version, memory, time, threads, gotSalt, gotKey, err := decodeHash(tc.encoded)

			prefix := fmt.Sprintf(
				"%s\ndecodeHash(%q)",
				packageName, tc.encoded,
			)

			// THEN: errors match expectations.
			errRegex := tc.errRegex
			if errRegex == "" {
				errRegex = `^$`
			}
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, errRegex,
				)
			}

			// AND: all errors wrap [ErrMalformedHash].
			if err != nil && !errors.Is(err, ErrMalformedHash) {
				t.Errorf(
					"%s error should wrap ErrMalformedHash\ngot: %v",
					prefix, err,
				)
			}

			// AND: the parsed parameters match expectations (all zeroed on error).
			if version != tc.wantVersion {
				t.Errorf(
					"%s version mismatch\ngot:  %d\nwant: %d",
					prefix, version, tc.wantVersion,
				)
			}
			if memory != tc.wantMemory {
				t.Errorf(
					"%s memory mismatch\ngot:  %d\nwant: %d",
					prefix, memory, tc.wantMemory,
				)
			}
			if time != tc.wantTime {
				t.Errorf(
					"%s time mismatch\ngot:  %d\nwant: %d",
					prefix, time, tc.wantTime,
				)
			}
			if threads != tc.wantThreads {
				t.Errorf(
					"%s threads mismatch\ngot:  %d\nwant: %d",
					prefix, threads, tc.wantThreads,
				)
			}
			if !bytes.Equal(gotSalt, tc.wantSalt) {
				t.Errorf(
					"%s salt mismatch\ngot:  %v\nwant: %v",
					prefix, gotSalt, tc.wantSalt,
				)
			}
			if !bytes.Equal(gotKey, tc.wantKey) {
				t.Errorf(
					"%s key mismatch\ngot:  %v\nwant: %v",
					prefix, gotKey, tc.wantKey,
				)
			}

			// AND: no salt/key is returned alongside an error.
			if err != nil && (gotSalt != nil || gotKey != nil) {
				t.Errorf(
					"%s salt/key should be nil on error\ngot: salt=%v, key=%v",
					prefix, gotSalt, gotKey,
				)
			}
		})
	}
}

func TestDecodeHash__hashRoundTrip(t *testing.T) {
	// GIVEN: an encoding produced by Hash.
	encoded, err := Hash("round-trip")
	if err != nil {
		t.Fatalf(
			"%s\nsetup Hash failed: %v",
			packageName, err,
		)
	}

	prefix := fmt.Sprintf("%s\ndecodeHash(%q)", packageName, encoded)

	// WHEN: decodeHash is called on it.
	version, memory, time, threads, salt, key, err := decodeHash(encoded)

	// THEN: it parses back to the parameters Hash used.
	if err != nil {
		t.Fatalf(
			"%s unexpected error\ngot: %v",
			prefix, err,
		)
	}
	if version != argon2.Version ||
		memory != argonMemory ||
		time != argonTime ||
		threads != argonThreads {
		t.Errorf(
			"%s parameter mismatch\ngot:  v=%d, m=%d, t=%d, p=%d\nwant: v=%d, m=%d, t=%d, p=%d",
			prefix,
			version, memory, time, threads,
			argon2.Version, argonMemory, argonTime, argonThreads,
		)
	}

	// AND: the salt and key are the configured lengths.
	if uint32(len(salt)) != saltLength || uint32(len(key)) != keyLength {
		t.Errorf(
			"%s length mismatch\ngot:  len(salt)=%d, len(key)=%d\nwant: len(salt)=%d, len(key)=%d",
			prefix, len(salt), len(key), saltLength, keyLength,
		)
	}
}
