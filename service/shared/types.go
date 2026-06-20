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

// Package shared provides shared functionality for Latest Version and Deployed Version lookups.
package shared

// OldIntIndex is an integer index reference used when restoring SecretValues.
type OldIntIndex struct {
	OldIndex *int `json:"old_index,omitempty"`
}

// OldStringIndex is a named string index reference used when restoring SecretValues.
type OldStringIndex struct {
	Name     string `json:"name,omitempty"`
	OldIndex string `json:"old_index,omitempty"`
}

// VSecretRef contains the reference for the Headers SecretValues.
type VSecretRef struct {
	Headers []OldIntIndex `json:"headers,omitempty"`
}

// WHSecretRef contains the reference for the WebHook SecretValues.
type WHSecretRef struct {
	Name     string        `json:"name,omitempty"`
	OldIndex string        `json:"old_index,omitempty"`
	Headers  []OldIntIndex `json:"headers,omitempty"`
}
