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

// Package webhook provides WebHook functionality to services.
package webhook

import (
	"net/http"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util"
)

// Header to use in the HTTP request.
type Header struct {
	Key   string `json:"key" yaml:"key"`     // Header key, e.g. X-Sig.
	Value string `json:"value" yaml:"value"` // Value to give the key.
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// It supports both the canonical form:
//
//	{key: "X", val: "Y"}
//
// and shorthand:
//
//	{X: Y}
//
// The shorthand is converted to the canonical key/val representation.
func (h *Headers) UnmarshalYAML(data []byte) error {
	// try and unmarshal as a Header list.
	var headers []Header
	if err := decode.Unmarshal("yaml", data, &headers); err == nil {
		*h = headers
		return nil
	}

	// Treat it as a map.
	var headersMap map[string]string
	if err := decode.Unmarshal("yaml", data, &headersMap); err != nil {
		return err //nolint:wrapcheck
	}

	// Sort the map keys.
	keys := util.SortedKeys(headersMap)
	*h = make([]Header, 0, len(keys))

	// Convert map to list.
	for _, key := range keys {
		*h = append(*h, Header{Key: key, Value: headersMap[key]})
	}
	return nil
}

// GitHub is the WebHook payload to emulate GitHub.
type GitHub struct {
	Ref    string `json:"ref"`    // "refs/heads/master".
	Before string `json:"before"` // "RandAlphaNumericLower(40)".
	After  string `json:"after"`  // "RandAlphaNumericLower(40)".
}

// setHeaders applies configured custom headers to req.
func (w *WebHook) setHeaders(req *http.Request) {
	var headers Headers
	switch {
	case w.Headers != nil:
		headers = w.Headers
	case w.Main.Headers != nil:
		headers = w.Main.Headers
	case w.Defaults.Headers != nil:
		headers = w.Defaults.Headers
	case w.HardDefaults.Headers != nil:
		headers = w.HardDefaults.Headers
	default:
		return
	}

	svcInfo := w.ServiceStatus.GetServiceInfo()
	for _, header := range headers {
		key := util.EvalEnvVars(header.Key)
		value := util.TemplateString(util.EvalEnvVars(header.Value), svcInfo)
		req.Header[key] = []string{value}
	}
}
