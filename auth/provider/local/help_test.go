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

package local

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"golang.org/x/crypto/argon2"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

var packageName = "auth_provider_local"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// mockStore is an in-memory [UserStore] for a single user.
type mockStore struct {
	creds *auth.Credentials // User returned for lookups.
	err   error             // Error returned from LocalCredentials.

	updateErr   error  // Error returned from UpdatePasswordHash.
	updateCalls int    // Number of UpdatePasswordHash calls.
	updatedID   string // Last userID passed to UpdatePasswordHash.
	updatedHash string // Last hash passed to UpdatePasswordHash.
}

func (m *mockStore) LocalCredentials(_ context.Context, _ string) (*auth.Credentials, error) {
	return m.creds, m.err
}

func (m *mockStore) UpdatePasswordHash(_ context.Context, userID, hash string) error {
	m.updateCalls++
	m.updatedID = userID
	m.updatedHash = hash
	return m.updateErr
}

// mustHash hashes plaintext with current parameters, failing the test on error.
func mustHash(t *testing.T, plaintext string) string {
	t.Helper()

	hash, err := password.Hash(plaintext)
	if err != nil {
		t.Fatalf(
			"%s\nsetup Hash failed: %v",
			packageName, err,
		)
	}
	return hash
}

// legacyHash encodes plaintext with deliberately outdated argon2id
// parameters, to trigger rehash-on-login.
func legacyHash(plaintext string) string {
	salt := []byte("0123456789abcdef")
	key := argon2.IDKey([]byte(plaintext), salt, 1, 8192, 1, 32)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=8192,t=1,p=1$%s$%s",
		argon2.Version,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
}
