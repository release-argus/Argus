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

package v1

import (
	"fmt"
	"testing"
	"time"
)

func TestWebSocketTokenStore_New(t *testing.T) {
	// GIVEN: a webSocketTokenStore.
	store := newWebSocketTokenStore()

	// WHEN: New is called.
	token := store.New()

	prefix := fmt.Sprintf("%s\nwebSocketTokenStore.New()", packageName)

	// THEN: a non-empty token is returned.
	if token == "" {
		t.Errorf("%s\ngot an empty token", prefix)
	}

	// AND: the token has the expected length.
	if len(token) != webSocketTokenLength {
		t.Errorf(
			"%s\ntoken length mismatch\ngot:  %d\nwant: %d",
			prefix, len(token), webSocketTokenLength,
		)
	}

	// AND: two consecutive tokens are different.
	other := store.New()
	if token == other {
		t.Errorf(
			"%s\nexpected two different tokens, got the same one twice: %q",
			prefix, token,
		)
	}
}

func TestWebSocketTokenStore_New_MaxSize(t *testing.T) {
	// GIVEN: a store already at webSocketTokenMaxSize, with distinct
	// (unexpired) expiries so the "oldest" (soonest-expiring) token is
	// unambiguous.
	store := newWebSocketTokenStore()
	base := time.Now().Add(webSocketTokenTTL)

	store.mu.Lock()
	for i := range webSocketTokenMaxSize {
		store.tokens[fmt.Sprintf("token-%d", i)] = base.Add(time.Duration(i) * time.Millisecond)
	}
	store.mu.Unlock()

	prefix := fmt.Sprintf("%s\nwebSocketTokenStore.New() (max size)", packageName)

	// WHEN: New is called, which would exceed webSocketTokenMaxSize.
	newToken := store.New()

	// THEN: the store does not grow beyond webSocketTokenMaxSize.
	store.mu.Lock()
	size := len(store.tokens)
	_, oldestStillPresent := store.tokens["token-0"]
	_, newestStillPresent := store.tokens[fmt.Sprintf("token-%d", webSocketTokenMaxSize-1)]
	store.mu.Unlock()

	if size != webSocketTokenMaxSize {
		t.Errorf(
			"%s\nstore size mismatch\ngot:  %d\nwant: %d",
			prefix, size, webSocketTokenMaxSize,
		)
	}

	// AND: the oldest (soonest-expiring) token was evicted to make room.
	if oldestStillPresent {
		t.Errorf("%s\noldest token should have been evicted", prefix)
	}

	// AND: the newest pre-existing token is retained.
	if !newestStillPresent {
		t.Errorf("%s\nnewest pre-existing token should not have been evicted", prefix)
	}

	// AND: the newly-minted token is itself valid.
	if !store.Validate(newToken) {
		t.Errorf("%s\nnewly minted token should be valid", prefix)
	}
}

func TestWebSocketTokenStore_Validate(t *testing.T) {
	// GIVEN: a webSocketTokenStore.
	tests := map[string]struct {
		setup func(store *webSocketTokenStore) string
		want  bool
	}{
		"valid, unused token": {
			setup: func(store *webSocketTokenStore) string {
				return store.New()
			},
			want: true,
		},
		"empty token": {
			setup: func(_ *webSocketTokenStore) string {
				return ""
			},
			want: false,
		},
		"unknown token": {
			setup: func(_ *webSocketTokenStore) string {
				return "unknown-token-that-was-never-issued"
			},
			want: false,
		},
		"expired token": {
			setup: func(store *webSocketTokenStore) string {
				token := store.New()
				// Force the token to have already expired.
				store.mu.Lock()
				store.tokens[token] = time.Now().Add(-time.Second)
				store.mu.Unlock()
				return token
			},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			store := newWebSocketTokenStore()
			token := tc.setup(store)

			prefix := fmt.Sprintf("%s\nwebSocketTokenStore.Validate(%q)", packageName, name)

			// WHEN: Validate is called.
			got := store.Validate(token)

			// THEN: the result matches expectations.
			if got != tc.want {
				t.Errorf(
					"%s\nfirst call result mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want,
				)
			}

			// AND: tokens are single-use - a second call always fails.
			if gotTwo := store.Validate(token); gotTwo {
				t.Errorf(
					"%s\nsecond call (re-use) should always fail (tokens are single-use), got: %t",
					prefix, gotTwo,
				)
			}
		})
	}
}

func TestWebSocketTokenStore_Prune(t *testing.T) {
	// GIVEN: a store with an expired and a valid token.
	store := newWebSocketTokenStore()
	expired := store.New()
	valid := store.New()

	store.mu.Lock()
	store.tokens[expired] = time.Now().Add(-time.Second)
	store.mu.Unlock()

	prefix := fmt.Sprintf("%s\nwebSocketTokenStore.prune()", packageName)

	// WHEN: a New call triggers pruning.
	_ = store.New()

	// THEN: the expired token is removed.
	store.mu.Lock()
	_, expiredStillPresent := store.tokens[expired]
	_, validStillPresent := store.tokens[valid]
	store.mu.Unlock()

	if expiredStillPresent {
		t.Errorf("%s\nexpired token should have been pruned", prefix)
	}

	// AND: the valid token remains.
	if !validStillPresent {
		t.Errorf("%s\nvalid token should not have been pruned", prefix)
	}
}
