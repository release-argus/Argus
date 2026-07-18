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

// Package provider defines the authentication provider abstraction.
// A Provider verifies credentials and returns an auth.Identity;
// it knows nothing about users, groups, or permissions.
package provider

import (
	"context"
	"fmt"

	"github.com/release-argus/Argus/auth"
)

// Provider authenticates credentials against a backend.
type Provider interface {
	// Name identifies the provider (e.g. "local").
	Name() string
	// Authenticate verifies the credentials, returning the [auth.Identity] on success.
	// Failures indistinguishable to the caller (unknown user, wrong password,
	// disabled user) return [auth.ErrInvalidCredentials].
	Authenticate(ctx context.Context, username, password string) (*auth.Identity, error)
}

// Registry holds the configured [Provider]s, keyed by name,
// preserving registration order.
type Registry struct {
	providers map[string]Provider
	order     []string
}

// NewRegistry creates an empty [Registry].
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds p to the [Registry]. Panics on a duplicate name.
func (r *Registry) Register(p Provider) {
	name := p.Name()
	if _, exists := r.providers[name]; exists {
		panic(fmt.Sprintf("auth provider %q already registered", name))
	}

	r.providers[name] = p
	r.order = append(r.order, name)
}

// Get returns the [Provider] registered under name (nil if unknown).
func (r *Registry) Get(name string) Provider {
	return r.providers[name]
}

// Names returns the registered [Provider] names, in registration order.
func (r *Registry) Names() []string {
	return append([]string(nil), r.order...)
}
