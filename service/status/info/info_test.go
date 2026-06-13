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
	// GIVEN: a ServiceInfo struct and a mutex.
	serviceInfo := &ServiceInfo{}
	mu := &sync.RWMutex{}

	// WHEN: SetMutex is called.
	serviceInfo.SetMutex(mu)

	// THEN: the mutex is set correctly.
	if serviceInfo.mu != mu {
		t.Errorf(
			"%s\nSetMutex(%p) mu mismatch\ngot:  %p",
			packageName, &mu, serviceInfo.mu,
		)
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

			// GIVEN: a ServiceInfo struct with a mutex and an Icon value.
			mu := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mu:   mu,
				Icon: tc.value,
			}

			// WHEN: GetIcon is called.
			result := serviceInfo.GetIcon()

			// THEN: the returned value matches the Icon field.
			if result != tc.value {
				t.Errorf(
					"%s\nServiceInfo.GetIcon() mismatch\ngot:  %s\nwant: %s",
					packageName, result, tc.value,
				)
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

			// GIVEN: a ServiceInfo struct with a mutex and an IconLinkTo value.
			mu := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mu:         mu,
				IconLinkTo: tc.value,
			}

			// WHEN: GetIconLinkTo is called.
			result := serviceInfo.GetIconLinkTo()

			// THEN: the returned value matches the Icon field.
			if result != tc.value {
				t.Errorf(
					"%s\nServiceInfo.GetIconLinkTo() mismatch\ngot:  %s\nwant: %s",
					packageName, result, tc.value,
				)
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

			// GIVEN: a ServiceInfo struct with a mutex and an WebURL value.
			mu := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mu:     mu,
				WebURL: tc.value,
			}

			// WHEN: GetWebURL is called.
			result := serviceInfo.GetWebURL()

			// THEN: the returned value matches the Icon field.
			if result != tc.value {
				t.Errorf(
					"%s\nServiceInfo.GetWebURL() mismatch\ngot:  %s\nwant: %s",
					packageName, result, tc.value,
				)
			}
		})
	}
}

func TestSkippedVersion(t *testing.T) {
	// GIVEN: a version string.
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
			name:    "already prefixed",
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

			// WHEN: SkippedVersion is called with the version string.
			result := SkippedVersion(tc.version)

			// THEN: the returned string should be the SkipPrefix followed by the version string.
			expected := SkipPrefix + tc.version
			if result != expected {
				t.Fatalf(
					"%s\nSkippedVersion(%q) mismatch\nwant: %q\ngot:  %q",
					packageName, tc.version, expected, result,
				)
			}
		})
	}
}
