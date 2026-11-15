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
	"errors"
	"strings"
	"testing"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestProvider_Name(t *testing.T) {
	provider := New(&mockStore{})

	got := provider.Name()

	if got != Name {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, Name)
	}
}

func TestProvider_Authenticate(t *testing.T) {
	correctPassword := "correct-password"

	enabledUser := func(t *testing.T) *auth.Credentials {
		return &auth.Credentials{
			UserID:       "uuid-1",
			Username:     "argus",
			DisplayName:  "Argus",
			Email:        "Argus@example.com",
			PasswordHash: mustHash(t, correctPassword),
			Enabled:      true,
		}
	}
	wantIdentity := auth.Identity{
		Provider:    Name,
		Subject:     "uuid-1",
		Username:    "argus",
		DisplayName: "Argus",
		Email:       "Argus@example.com",
	}

	tests := []struct {
		name         string
		creds        func(t *testing.T) *auth.Credentials
		storeErr     error
		password     string
		wantIdentity bool
		wantErr      error
		errRegex     string
	}{
		{
			name:         "success",
			creds:        enabledUser,
			password:     correctPassword,
			wantIdentity: true,
			errRegex:     `^$`,
		},
		{
			name:     "unknown user",
			creds:    func(_ *testing.T) *auth.Credentials { return nil },
			password: correctPassword,
			wantErr:  auth.ErrInvalidCredentials,
			errRegex: `^invalid credentials$`,
		},
		{
			name:     "store error is propagated, not masked",
			creds:    func(_ *testing.T) *auth.Credentials { return nil },
			storeErr: errors.New("database broke"),
			password: correctPassword,
			errRegex: `^look up user:`,
		},
		{
			name:     "wrong password",
			creds:    enabledUser,
			password: "incorrect",
			wantErr:  auth.ErrInvalidCredentials,
			errRegex: `^invalid credentials$`,
		},
		{
			name: "disabled user with correct password",
			creds: func(t *testing.T) *auth.Credentials {
				creds := enabledUser(t)
				creds.Enabled = false
				return creds
			},
			password: correctPassword,
			wantErr:  auth.ErrInvalidCredentials,
			errRegex: `^invalid credentials$`,
		},
		{
			name: "malformed stored hash is a credential failure",
			creds: func(t *testing.T) *auth.Credentials {
				creds := enabledUser(t)
				creds.PasswordHash = "not-a-hash"
				return creds
			},
			password: correctPassword,
			wantErr:  auth.ErrInvalidCredentials,
			errRegex: `^invalid credentials$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockStore{creds: tc.creds(t), err: tc.storeErr}
			provider := New(store)

			identity, err := provider.Authenticate(t.Context(), "argus", tc.password)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error type mismatch\ngot:  %v\nwant: %v", err, tc.wantErr)
			}

			if tc.wantIdentity {
				if identity == nil {
					t.Fatal("identity mismatch\ngot:  nil\nwant: non-nil")
				}
				if testErr := test.AssertFields(t, []test.FieldAssertion{
					{Name: "Provider", Got: identity.Provider, Want: wantIdentity.Provider, Mode: test.CompareEqual},
					{Name: "Subject", Got: identity.Subject, Want: wantIdentity.Subject, Mode: test.CompareEqual},
					{Name: "Username", Got: identity.Username, Want: wantIdentity.Username, Mode: test.CompareEqual},
					{Name: "DisplayName", Got: identity.DisplayName, Want: wantIdentity.DisplayName, Mode: test.CompareEqual},
					{Name: "Email", Got: identity.Email, Want: wantIdentity.Email, Mode: test.CompareEqual},
				}, "Identity"); testErr != nil {
					t.Error(testErr)
				}
			} else if identity != nil {
				t.Errorf("identity mismatch\ngot:  Subject=%q\nwant: nil", identity.Subject)
			}

			// Every hash here is current, or the login fails.
			if store.updateCalls != 0 {
				t.Errorf("mockStore.updateCalls mismatch\ngot:  %d\nwant: 0", store.updateCalls)
			}
		})
	}
}

func TestProvider_Authenticate__rehash(t *testing.T) {
	tests := []struct {
		name       string
		hashErr    bool  // Make re-hashing fail.
		updateErr  error // Make persisting the new hash fail.
		wantUpdate bool
	}{
		{
			name:       "outdated hash is upgraded",
			wantUpdate: true,
		},
		{
			name:       "failure to persist the new hash does not fail the login",
			updateErr:  errors.New("database broke"),
			wantUpdate: true,
		},
		{
			name:    "failure to derive the new hash does not fail the login",
			hashErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps hashPassword.

			store := &mockStore{
				creds: &auth.Credentials{
					UserID:       "uuid-1",
					Username:     "argus",
					PasswordHash: legacyHash("old-params"),
					Enabled:      true,
				},
				updateErr: tc.updateErr,
			}
			provider := New(store)

			if tc.hashErr {
				hashPasswordHad := hashPassword
				hashPassword = func(_ string) (string, error) {
					return "", errors.New("hash broke")
				}
				t.Cleanup(func() { hashPassword = hashPasswordHad })
			}

			identity, err := provider.Authenticate(t.Context(), "argus", "old-params")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if identity == nil {
				t.Fatal("identity mismatch\ngot:  nil\nwant: non-nil")
			}

			if (store.updateCalls > 0) != tc.wantUpdate {
				t.Fatalf(
					"mockStore.updateCalls > 0 mismatch\ngot:  %t (updateCalls=%d)\nwant: %t",
					store.updateCalls > 0, store.updateCalls, tc.wantUpdate,
				)
			}

			if tc.wantUpdate && tc.updateErr == nil {
				if store.updatedID != "uuid-1" {
					t.Errorf(
						"mockStore.updatedID mismatch\ngot:  %q\nwant: %q",
						store.updatedID, "uuid-1",
					)
				}
				match, needsRehash, err := password.Verify("old-params", store.updatedHash)
				if err != nil || !match || needsRehash {
					t.Errorf(
						"password.Verify() mismatch\n"+
							"got:  match=%t, needsRehash=%t, err=%v\n"+
							"want: match=true, needsRehash=false, err=nil",
						match, needsRehash, err,
					)
				}
			}
		})
	}
}

func TestProvider_Authenticate__verifyError(t *testing.T) {
	// Fails for a reason other than a malformed hash.
	wantErr := errors.New("verify broke")
	verifyPasswordHad := verifyPassword
	verifyPassword = func(_, _ string) (bool, bool, error) {
		return false, false, wantErr
	}
	t.Cleanup(func() { verifyPassword = verifyPasswordHad })

	provider := New(&mockStore{
		creds: &auth.Credentials{
			UserID:       "uuid-1",
			Username:     "argus",
			PasswordHash: mustHash(t, "correct-password"),
			Enabled:      true,
		},
	})

	identity, err := provider.Authenticate(t.Context(), "argus", "correct-password")

	if !errors.Is(err, wantErr) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, wantErr)
	}
	if errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: NOT %v", err, auth.ErrInvalidCredentials)
	}
	if identity != nil {
		t.Errorf("identity mismatch\ngot:  Subject=%q\nwant: nil", identity.Subject)
	}
}

func TestProvider_Authenticate__dummy_fallback(t *testing.T) {
	// Hashing is broken, so the lazy dummy hash falls back to its constant.
	hashPasswordHad := hashPassword
	hashPassword = func(_ string) (string, error) {
		return "", errors.New("hash broke")
	}
	t.Cleanup(func() { hashPassword = hashPasswordHad })

	provider := New(&mockStore{})

	identity, err := provider.Authenticate(t.Context(), "ghost", "password")

	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("error type mismatch\ngot:  %v\nwant: %v", err, auth.ErrInvalidCredentials)
	}
	if identity != nil {
		t.Errorf("identity mismatch\ngot:  Subject=%q\nwant: nil", identity.Subject)
	}
}

func TestDummyHashFallback__matchesCurrentParams(t *testing.T) {
	current, err := password.Hash("timing")
	if err != nil {
		t.Fatalf("password.Hash() failed: %v", err)
	}

	// Equal parameters make an unknown user cost as much to verify as a real one.
	if got, want := argonParams(dummyHashFallback), argonParams(current); got != want {
		t.Errorf("argonParams(dummyHashFallback) mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// argonParams returns the "m=<memory>,t=<time>,p=<threads>" segment of an argon2id encoding.
func argonParams(encoded string) string {
	parts := strings.Split(encoded, "$")
	if len(parts) < 4 {
		return ""
	}
	return parts[3]
}
