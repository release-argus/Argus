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

package docker

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestOldDockerRegistryDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		val  *oldDockerRegistryDefaults
		want bool
	}{
		{
			name: "nil",
			val:  nil,
			want: true,
		},
		{
			name: "empty",
			val:  &oldDockerRegistryDefaults{},
			want: true,
		},
		{
			name: "non-empty/Username",
			val: &oldDockerRegistryDefaults{
				Username: "a",
			},
			want: false,
		},
		{
			name: "non-empty/Token",
			val: &oldDockerRegistryDefaults{
				Token: "a",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			val: &oldDockerRegistryDefaults{
				Username: "foo",
				Token:    "bar",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.val.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestOldDockerDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		val  *oldDockerDefaults
		want bool
	}{
		{
			name: "nil",
			val:  nil,
			want: true,
		},
		{
			name: "empty",
			val:  &oldDockerDefaults{},
			want: true,
		},
		{
			name: "non-empty/GHCR",
			val: &oldDockerDefaults{
				RegistryGHCR: &oldDockerRegistryDefaults{
					Username: "a",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Hub",
			val: &oldDockerDefaults{
				RegistryHub: &oldDockerRegistryDefaults{
					Username: "a",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Quay",
			val: &oldDockerDefaults{
				RegistryQuay: &oldDockerRegistryDefaults{
					Username: "a",
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			val: &oldDockerDefaults{
				RegistryGHCR: &oldDockerRegistryDefaults{
					Username: "a",
				},
				RegistryHub: &oldDockerRegistryDefaults{
					Username: "a",
				},
				RegistryQuay: &oldDockerRegistryDefaults{
					Username: "a",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.val.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestConvertOldDefaults(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "{}\n",
		},
		{
			name: "empty",
			data: []byte{},
			want: "{}\n",
		},
		{
			name: "GHCR",
			data: []byte(test.TrimJSON(`{
				"ghcr": {
					"username": "a",
					"token": "b"
				}
			}`)),
			want: test.TrimYAML(`
				registry:
					ghcr:
						auth:
							token: b
			`),
		},
		{
			name: "Hub",
			data: []byte(test.TrimJSON(`{
				"hub": {
					"username": "a",
					"token": "b"
				}
			}`)),
			want: test.TrimYAML(`
				registry:
					hub:
						auth:
							username: a
							token: b
			`),
		},
		{
			name: "Quay",
			data: []byte(test.TrimJSON(`{
				"quay": {
					"username": "a",
					"token": "b"
				}
			}`)),
			want: test.TrimYAML(`
				registry:
					quay:
						auth:
							token: b
			`),
		},
		{
			name: "filled",
			data: []byte(test.TrimJSON(`{
				"ghcr": {
					"username": "a",
					"token": "b"
				},
				"hub": {
					"username": "a",
					"token": "b"
				},
				"quay": {
					"username": "a",
					"token": "b"
				}
			}`)),
			want: test.TrimYAML(`
				registry:
					ghcr:
						auth:
							token: b
					hub:
						auth:
							username: a
							token: b
					quay:
						auth:
							token: b
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newFormat := &Defaults{
				Registry: RegistryDefaultsSet{
					GHCR: RegistryDefaultsMap["ghcr"]().(*GHCRRegistryDefaults),
					Hub:  RegistryDefaultsMap["hub"]().(*HubRegistryDefaults),
					Quay: RegistryDefaultsMap["quay"]().(*QuayRegistryDefaults),
				},
			}

			convertOldDefaults("yaml", tc.data, newFormat)

			if got := newFormat.String(""); got != tc.want {
				t.Errorf("Defaults.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
