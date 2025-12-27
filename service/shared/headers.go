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

// Package shared provides shared functionality for Latest Version and Deployed Version lookups.
package shared

import "github.com/release-argus/Argus/util"

// Header to use in a HTTP request.
type Header struct {
	Key   string `json:"key" yaml:"key"`     // Header key, e.g. X-Sig.
	Value string `json:"value" yaml:"value"` // Value to give the key.
}

type Headers []Header

// InheritSecrets copies secret values from otherHeaders that are referenced in secretRefs.
func (h *Headers) InheritSecrets(otherHeaders Headers, secretRefs []OldIntIndex) {
	// If we don't have headers in both locations.
	if len(*h) == 0 && len(otherHeaders) == 0 {
		return
	}

	for i := range *h {
		// If referencing a secret of an existing header.
		if (*h)[i].Value == util.SecretValue {
			// Don't have a secretRef for this header.
			if i >= len(secretRefs) {
				break
			}
			oldIndex := secretRefs[i].OldIndex
			// Not a reference to an old Header.
			if oldIndex == nil {
				continue
			}

			if *oldIndex < len(otherHeaders) {
				(*h)[i].Value = otherHeaders[*oldIndex].Value
			}
		}
	}
}
