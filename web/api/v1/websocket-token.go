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

// Package v1 provides the API for the webserver.
package v1

import (
	"sync"
	"time"

	"github.com/release-argus/Argus/util"
)

// webSocketTokenLength is the number of characters in a minted WebSocket token.
const webSocketTokenLength = 32

// webSocketTokenTTL is how long a minted WebSocket token remains valid if unused.
const webSocketTokenTTL = 15 * time.Second

// webSocketTokenMaxSize is the maximum number of unexpired tokens held with a [webSocketTokenTTL].
const webSocketTokenMaxSize = 100

// webSocketTokenStore issues and validates short-lived, single-use tokens
// used to authenticate the WebSocket handshake.
type webSocketTokenStore struct {
	mu     sync.Mutex
	tokens map[string]time.Time // token -> expiry.
}

// newWebSocketTokenStore creates a new, empty [webSocketTokenStore].
func newWebSocketTokenStore() *webSocketTokenStore {
	return &webSocketTokenStore{
		tokens: make(map[string]time.Time),
	}
}

// New mints a new single-use token, valid for webSocketTokenTTL.
func (s *webSocketTokenStore) New() string {
	token := util.RandAlphaNumericLower(webSocketTokenLength)
	expiry := time.Now().Add(webSocketTokenTTL)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.prune()
	if len(s.tokens) >= webSocketTokenMaxSize {
		s.evictOldest()
	}
	s.tokens[token] = expiry

	return token
}

// Validate reports whether token is a valid, unexpired token, consuming it if so.
func (s *webSocketTokenStore) Validate(token string) bool {
	if token == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.prune()

	expiry, ok := s.tokens[token]
	if !ok {
		return false
	}
	delete(s.tokens, token)

	return time.Now().Before(expiry)
}

// prune removes expired tokens.
//
// Callers must hold s.mu.
func (s *webSocketTokenStore) prune() {
	now := time.Now()
	for token, expiry := range s.tokens {
		if now.After(expiry) {
			delete(s.tokens, token)
		}
	}
}

// evictOldest removes the token with the soonest expiry.
//
// Callers must hold s.mu.
func (s *webSocketTokenStore) evictOldest() {
	var oldestToken string
	var oldestExpiry time.Time

	for token, expiry := range s.tokens {
		if oldestToken == "" || expiry.Before(oldestExpiry) {
			oldestToken = token
			oldestExpiry = expiry
		}
	}

	delete(s.tokens, oldestToken)
}
