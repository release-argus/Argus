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

package shared

// OldIntIndex to look at for any SecretValues used.
type OldIntIndex struct {
	OldIndex *int `json:"oldIndex,omitempty"`
}

// OldStringIndex to look at for any SecretValues used.
type OldStringIndex struct {
	OldIndex string `json:"oldIndex,omitempty"`
}

// DVSecretRef contains the reference for the DeployedVersionLookup SecretValues.
type DVSecretRef struct {
	Headers []OldIntIndex `json:"headers,omitempty"`
}

// WHSecretRef contains the reference for the WebHook SecretValues.
type WHSecretRef struct {
	OldIndex      string        `json:"oldIndex,omitempty"`
	CustomHeaders []OldIntIndex `json:"custom_headers,omitempty"`
}
