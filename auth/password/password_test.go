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
func legacyHash(password string, memory, time uint32, threads uint8, keyLen uint32) string {
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
			errRegex := util.ValueOr(tc.errRegex, `^$`)
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
