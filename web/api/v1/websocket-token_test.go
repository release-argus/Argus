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
	store := newWebSocketTokenStore()

	token := store.New()

	if token == "" {
		t.Errorf("token mismatch\ngot:  %q\nwant: non-empty", token)
	}

	if len(token) != webSocketTokenLength {
		t.Errorf("token length mismatch\ngot:  %d\nwant: %d", len(token), webSocketTokenLength)
	}

	other := store.New()
	if token == other {
		t.Errorf("second token mismatch\ngot:  %q\nwant: NOT %q", other, token)
	}
}

func TestWebSocketTokenStore_New_MaxSize(t *testing.T) {
	store := newWebSocketTokenStore()
	base := time.Now().Add(webSocketTokenTTL)

	store.mu.Lock()
	// Distinct expiries, so the soonest-expiring token is unambiguous.
	for i := range webSocketTokenMaxSize {
		store.tokens[fmt.Sprintf("token-%d", i)] = base.Add(time.Duration(i) * time.Millisecond)
	}
	store.mu.Unlock()

	newToken := store.New()

	newestKey := fmt.Sprintf("token-%d", webSocketTokenMaxSize-1)
	store.mu.Lock()
	size := len(store.tokens)
	_, oldestStillPresent := store.tokens["token-0"]
	_, newestStillPresent := store.tokens[newestKey]
	store.mu.Unlock()

	if size != webSocketTokenMaxSize {
		t.Errorf(
			"webSocketTokenStore.tokens length mismatch\ngot:  %d\nwant: %d",
			size, webSocketTokenMaxSize,
		)
	}

	if oldestStillPresent {
		t.Error(
			"webSocketTokenStore.tokens[\"token-0\"] mismatch\n" +
				"got:  present\nwant: absent (evicted)",
		)
	}

	if !newestStillPresent {
		t.Errorf("webSocketTokenStore.tokens[%q] mismatch\ngot:  absent\nwant: present", newestKey)
	}

	if !store.Validate(newToken) {
		t.Error("webSocketTokenStore.Validate(newToken) mismatch\ngot:  false\nwant: true")
	}
}

func TestWebSocketTokenStore_Validate(t *testing.T) {
	tests := []struct {
		name  string
		setup func(store *webSocketTokenStore) string
		want  bool
	}{
		{
			name: "valid, unused token",
			setup: func(store *webSocketTokenStore) string {
				return store.New()
			},
			want: true,
		},
		{
			name: "empty token",
			setup: func(_ *webSocketTokenStore) string {
				return ""
			},
			want: false,
		},
		{
			name: "unknown token",
			setup: func(_ *webSocketTokenStore) string {
				return "unknown-token-that-was-never-issued"
			},
			want: false,
		},
		{
			name: "expired token",
			setup: func(store *webSocketTokenStore) string {
				token := store.New()
				store.mu.Lock()
				store.tokens[token] = time.Now().Add(-time.Second)
				store.mu.Unlock()
				return token
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := newWebSocketTokenStore()
			token := tc.setup(store)

			got := store.Validate(token)

			if got != tc.want {
				t.Errorf(
					"first call: webSocketTokenStore.Validate() mismatch\ngot:  %t\nwant: %t",
					got, tc.want,
				)
			}

			// Tokens are single-use.
			if gotTwo := store.Validate(token); gotTwo {
				t.Error(
					"second call: webSocketTokenStore.Validate() mismatch\ngot:  true\nwant: false",
				)
			}
		})
	}
}

func TestWebSocketTokenStore_Prune(t *testing.T) {
	store := newWebSocketTokenStore()
	expired := store.New()
	valid := store.New()

	store.mu.Lock()
	store.tokens[expired] = time.Now().Add(-time.Second)
	store.mu.Unlock()

	// New prunes expired tokens.
	_ = store.New()

	store.mu.Lock()
	_, expiredStillPresent := store.tokens[expired]
	_, validStillPresent := store.tokens[valid]
	store.mu.Unlock()

	if expiredStillPresent {
		t.Error(
			"expired: webSocketTokenStore.tokens key mismatch\n" +
				"got:  present\nwant: absent (pruned)",
		)
	}

	if !validStillPresent {
		t.Error("valid: webSocketTokenStore.tokens key mismatch\ngot:  absent\nwant: present")
	}
}
