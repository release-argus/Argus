// Copyright [2025] [Argus]
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

func TestSetMutex(t *testing.T) {
	// GIVEN a ServiceInfo struct and a mutex.
	serviceInfo := &ServiceInfo{}
	mutex := &sync.RWMutex{}

	// WHEN SetMutex is called.
	serviceInfo.SetMutex(mutex)

	// THEN the mutex is set correctly.
	if serviceInfo.mutex != mutex {
		t.Errorf("%s\nmutex mismatch\nwant: %v\ngot:  %v",
			packageName, &mutex, serviceInfo.mutex)
	}
}
func TestGetIcon(t *testing.T) {
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

			// GIVEN a ServiceInfo struct with a mutex and an Icon value.
			mutex := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mutex: mutex,
				Icon:  tc.value,
			}

			// WHEN GetIcon is called.
			result := serviceInfo.GetIcon()

			// THEN the returned value matches the Icon field.
			if result != tc.value {
				t.Errorf("%s\nmismatch\nwant: %s\ngot:  %s",
					packageName, tc.value, result)
			}
		})
	}
}

func TestGetIconLinkTo(t *testing.T) {
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

			// GIVEN a ServiceInfo struct with a mutex and an IconLinkTo values.
			mutex := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mutex:      mutex,
				IconLinkTo: tc.value,
			}

			// WHEN GetIconLinkTo is called.
			result := serviceInfo.GetIconLinkTo()

			// THEN the returned value matches the Icon field.
			if result != tc.value {
				t.Errorf("%s\nmismatch\nwant: %s\ngot:  %s",
					packageName, tc.value, result)
			}
		})
	}
}

func TestGetWebURL(t *testing.T) {
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

			// GIVEN a ServiceInfo struct with a mutex and an WebURL value.
			mutex := &sync.RWMutex{}
			serviceInfo := &ServiceInfo{
				mutex:  mutex,
				WebURL: tc.value,
			}

			// WHEN GetWebURL is called.
			result := serviceInfo.GetWebURL()

			// THEN the returned value matches the Icon field.
			if result != tc.value {
				t.Errorf("%s\nmismatch\nwant: %s\ngot:  %s",
					packageName, tc.value, result)
			}
		})
	}
}
