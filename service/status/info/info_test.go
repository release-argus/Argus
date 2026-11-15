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

//go:build unit || integration

package info

import (
	"sync"
	"testing"
)

func TestServiceInfo_SetMutex(t *testing.T) {
	serviceInfo := &ServiceInfo{}
	mu := &sync.RWMutex{}

	serviceInfo.SetMutex(mu)

	if serviceInfo.mu != mu {
		t.Errorf("ServiceInfo.mu mismatch\ngot:  %p\nwant: %p", serviceInfo.mu, mu)
	}
}

func TestServiceInfo_GetIcon(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "Valid Icon URL",
			value: "https://example.com/icon1.png",
		},
		{
			name:  "Another Valid Icon URL",
			value: "https://example.com/icon2.png",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mu := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mu:   mu,
				Icon: tc.value,
			}

			result := serviceInfo.GetIcon()

			if result != tc.value {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", result, tc.value)
			}
		})
	}
}

func TestServiceInfo_GetIconLinkTo(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "URL",
			value: "https://example.com/icon-link",
		},
		{
			name:  "empty",
			value: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mu := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mu:         mu,
				IconLinkTo: tc.value,
			}

			result := serviceInfo.GetIconLinkTo()

			if result != tc.value {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", result, tc.value)
			}
		})
	}
}

func TestServiceInfo_GetWebURL(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "URL",
			value: "https://example.com/web",
		},
		{
			name:  "empty",
			value: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mu := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mu:     mu,
				WebURL: tc.value,
			}

			result := serviceInfo.GetWebURL()

			if result != tc.value {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", result, tc.value)
			}
		})
	}
}

func TestSkippedVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "semver",
			version: "1.2.3",
		},
		{
			name:    "empty string",
			version: "",
		},
		{
			name:    "already prefixed gets a second prefix",
			version: SkipPrefix + "1.2.3",
		},
		{
			name:    "non-semver string",
			version: "latest",
		},
		{
			name:    "with whitespace",
			version: " 1.2.3 ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := SkippedVersion(tc.version)

			expected := SkipPrefix + tc.version
			if result != expected {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", result, expected)
			}
		})
	}
}
