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

	"github.com/release-argus/Argus/internal/test"
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
	plaintext := "s3cret-p4ssw0rd"

	first, errFirst := Hash(plaintext)
	second, errSecond := Hash(plaintext)

	if errFirst != nil {
		t.Fatalf("first: Hash() failed: %v", errFirst)
	}
	if errSecond != nil {
		t.Fatalf("second: Hash() failed: %v", errSecond)
	}

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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !strings.HasPrefix(tc.hash, wantPrefix) {
				t.Errorf("prefix mismatch\ngot:  %q\nwant: %q", tc.hash, wantPrefix)
			}

			match, needsRehash, err := Verify(plaintext, tc.hash)
			if err != nil || !match || needsRehash {
				t.Errorf(
					"Verify() mismatch\n"+
						"got:  match=%t, needsRehash=%t, err=%v\n"+
						"want: match=true, needsRehash=false, err=nil",
					match, needsRehash, err,
				)
			}
		})
	}

	// Salts are random, so the hashes differ.
	if first == second {
		t.Errorf("second: hash mismatch\ngot:  %q\nwant: NOT %q", second, first)
	}
}

func TestHash__randRead_error(t *testing.T) {
	wantErr := errors.New("rand broke")
	randReadHad := randRead
	randRead = func(_ []byte) (int, error) { return 0, wantErr }
	t.Cleanup(func() { randRead = randReadHad })

	_, err := Hash("anything")

	if !errors.Is(err, wantErr) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, wantErr)
	}
}

func TestHash__tooLong(t *testing.T) {
	tooLong := strings.Repeat("a", maxPasswordLength+1)

	_, err := Hash(tooLong)

	if !errors.Is(err, ErrPasswordTooLong) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrPasswordTooLong)
	}
	if match, _, err := Verify(tooLong, "$argon2id$anything"); match || err != nil {
		t.Errorf(
			"Verify(too-long) mismatch\ngot:  match=%t, err=%v\nwant: match=false, err=nil",
			match, err,
		)
	}
}

func TestVerify(t *testing.T) {
	plaintext := "correct horse battery staple"
	current, err := Hash(plaintext)
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
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
			errRegex: `^$`,
		},
		{
			name:     "wrong password",
			password: "incorrect",
			encoded:  current,
			errRegex: `^$`,
		},
		{
			name:      "correct password/current params",
			password:  plaintext,
			encoded:   current,
			wantMatch: true,
			errRegex:  `^$`,
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
			errRegex:        `^$`,
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
			errRegex:        `^$`,
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
			errRegex:        `^$`,
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
			errRegex:        `^$`,
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
			errRegex:        `^$`,
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
			errRegex: test.TrimYAML(`
				^malformed password hash:
					salt: illegal base64 data`),
		},
		{
			name:     "malformed/bad key base64",
			password: plaintext,
			encoded:  "$argon2id$v=19$m=19456,t=1,p=2$c2FsdA$!!!",
			errRegex: test.TrimYAML(`
				^malformed password hash:
					key: illegal base64 data`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			match, needsRehash, err := Verify(tc.password, tc.encoded)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil && !errors.Is(err, ErrMalformedHash) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrMalformedHash)
			}

			if match != tc.wantMatch {
				t.Errorf("match mismatch\ngot:  %t\nwant: %t", match, tc.wantMatch)
			}

			if needsRehash != tc.wantNeedsRehash {
				t.Errorf(
					"needsRehash mismatch\ngot:  %t\nwant: %t",
					needsRehash, tc.wantNeedsRehash,
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
			errRegex:    `^$`,
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
			errRegex:    `^$`,
		},
		{
			name: "valid/zeroed parameters are not validated",
			encoded: encodeParts(
				0, 0, 0, 0, b64(salt), b64(key),
			),
			wantSalt: salt,
			wantKey:  key,
			errRegex: `^$`,
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
			errRegex: test.TrimYAML(`
				^malformed password hash:
					salt: illegal base64 data`),
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
			errRegex: test.TrimYAML(`
				^malformed password hash:
					key: illegal base64 data`),
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
			errRegex:    `^$`,
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
			errRegex:    `^$`,
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
			errRegex:    `^$`,
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

			version, memory, time, threads, gotSalt, gotKey, err := decodeHash(tc.encoded)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if err != nil && !errors.Is(err, ErrMalformedHash) {
				t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, ErrMalformedHash)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "version", Got: version, Want: tc.wantVersion, Mode: test.CompareEqual},
				{Name: "memory", Got: memory, Want: tc.wantMemory, Mode: test.CompareEqual},
				{Name: "time", Got: time, Want: tc.wantTime, Mode: test.CompareEqual},
				{Name: "threads", Got: threads, Want: tc.wantThreads, Mode: test.CompareEqual},
				{Name: "salt", Got: gotSalt, Want: tc.wantSalt, Mode: test.CompareEqual},
				{Name: "key", Got: gotKey, Want: tc.wantKey, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, ""); testErr != nil {
				t.Error(testErr)
			}

			if err != nil && (gotSalt != nil || gotKey != nil) {
				t.Errorf(
					"salt/key mismatch\ngot:  salt=%#v, key=%#v\nwant: salt=nil, key=nil",
					gotSalt, gotKey,
				)
			}
		})
	}
}

func TestDecodeHash__hashRoundTrip(t *testing.T) {
	encoded, err := Hash("round-trip")
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	version, memory, time, threads, salt, key, err := decodeHash(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != argon2.Version ||
		memory != argonMemory ||
		time != argonTime ||
		threads != argonThreads {
		t.Errorf(
			"parameter mismatch\ngot:  v=%d, m=%d, t=%d, p=%d\nwant: v=%d, m=%d, t=%d, p=%d",
			version, memory, time, threads,
			argon2.Version, argonMemory, argonTime, argonThreads,
		)
	}

	if uint32(len(salt)) != saltLength || uint32(len(key)) != keyLength {
		t.Errorf(
			"length mismatch\ngot:  len(salt)=%d, len(key)=%d\nwant: len(salt)=%d, len(key)=%d",
			len(salt), len(key),
			saltLength, keyLength,
		)
	}
}
