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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestProvider_Name(t *testing.T) {
	// GIVEN: a local Provider.
	provider := New(&mockStore{})

	// WHEN: Name is called.
	got := provider.Name()

	prefix := fmt.Sprintf("%s\nProvider.Name()", packageName)

	// THEN: it returns the package's provider name.
	if got != Name {
		t.Errorf(
			"%s result mismatch\ngot:  %q\nwant: %q",
			prefix, got, Name,
		)
	}
}

func TestProvider_Authenticate(t *testing.T) {
	// GIVEN: user stores in various states.
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
		},
		{
			name: "malformed stored hash errors",
			creds: func(t *testing.T) *auth.Credentials {
				creds := enabledUser(t)
				creds.PasswordHash = "not-a-hash"
				return creds
			},
			password: correctPassword,
			errRegex: `^verify password:`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := &mockStore{creds: tc.creds(t), err: tc.storeErr}
			provider := New(store)

			// WHEN: Authenticate is called.
			identity, err := provider.Authenticate(t.Context(), "argus", tc.password)

			prefix := fmt.Sprintf(
				"%s\nProvider.Authenticate(password=%q)",
				packageName, tc.password,
			)

			// THEN: errors match expectations.
			errRegex := tc.errRegex
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf(
						"%s error mismatch\ngot:  %v\nwant: %v",
						prefix, err, tc.wantErr,
					)
				}
			} else if errRegex == "" {
				errRegex = `^$`
			}
			e := errfmt.FormatError(err)
			if !util.RegexCheck(errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: an Identity is returned only on success, with the user's details.
			if tc.wantIdentity {
				if identity == nil {
					t.Fatalf("%s expected an Identity, got nil", prefix)
				}
				if identity.Provider != wantIdentity.Provider ||
					identity.Subject != wantIdentity.Subject ||
					identity.Username != wantIdentity.Username ||
					identity.DisplayName != wantIdentity.DisplayName ||
					identity.Email != wantIdentity.Email {
					t.Errorf(
						"%s Identity mismatch\ngot:  %+v\nwant: %+v",
						prefix, *identity, wantIdentity,
					)
				}
			} else if identity != nil {
				t.Errorf(
					"%s expected no Identity\ngot: %+v",
					prefix, *identity,
				)
			}

			// AND: no rehash happened in any of these states.
			if store.updateCalls != 0 {
				t.Errorf(
					"%s unexpected UpdatePasswordHash calls: %d",
					prefix, store.updateCalls,
				)
			}
		})
	}
}

func TestProvider_Authenticate__rehash(t *testing.T) {
	// GIVEN: a user whose stored hash uses outdated parameters.
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
			// t.Parallel() - Cannot run in parallel since we're modifying shared vars.

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

			// WHEN: the user authenticates with the correct password.
			identity, err := provider.Authenticate(t.Context(), "argus", "old-params")

			prefix := fmt.Sprintf(
				"%s\nProvider.Authenticate() rehash",
				packageName,
			)

			// THEN: the login succeeds regardless of rehash outcome.
			if err != nil || identity == nil {
				t.Fatalf(
					"%s login should succeed\ngot:  identity=%v, err=%v",
					prefix, identity, err,
				)
			}

			// AND: the hash was (or was not) re-persisted as expected.
			if (store.updateCalls > 0) != tc.wantUpdate {
				t.Fatalf(
					"%s UpdatePasswordHash calls mismatch\ngot:  %d\nwantUpdate: %t",
					prefix, store.updateCalls, tc.wantUpdate,
				)
			}

			// AND: any persisted hash is a current-parameter hash of the password.
			if tc.wantUpdate && tc.updateErr == nil {
				if store.updatedID != "uuid-1" {
					t.Errorf(
						"%s updated userID mismatch\ngot:  %q\nwant: %q",
						prefix, store.updatedID, "uuid-1",
					)
				}
				match, needsRehash, err := password.Verify("old-params", store.updatedHash)
				if err != nil || !match || needsRehash {
					t.Errorf(
						"%s persisted hash should be a current-parameter hash of the password\ngot:  match=%t, needsRehash=%t, err=%v",
						prefix, match, needsRehash, err,
					)
				}
			}
		})
	}
}

func TestProvider_Authenticate__dummy_fallback(t *testing.T) {
	// GIVEN: hashing is broken, so the lazy dummy hash falls back to its constant.
	hashPasswordHad := hashPassword
	hashPassword = func(_ string) (string, error) {
		return "", errors.New("hash broke")
	}
	t.Cleanup(func() { hashPassword = hashPasswordHad })

	provider := New(&mockStore{})

	// WHEN: an unknown user authenticates.
	identity, err := provider.Authenticate(t.Context(), "ghost", "password")

	prefix := fmt.Sprintf("%s\nProvider.Authenticate() dummy fallback", packageName)

	// THEN: the pre-computed fallback hash still yields a rejection.
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf(
			"%s error mismatch\ngot:  %v\nwant: %v",
			prefix, err, auth.ErrInvalidCredentials,
		)
	}
	if identity != nil {
		t.Errorf(
			"%s expected no Identity\ngot: %+v",
			prefix, *identity,
		)
	}
}
