// Copyright [2022] [Argus]
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

package webhook

import (
	"net/http"

	"github.com/release-argus/Argus/util"
)

// GitHub is the WebHook payload to emulate GitHub.
type GitHub struct {
	Ref    string `json:"ref"`    // "refs/heads/master"
	Before string `json:"before"` // "RandAlphaNumericLower(40)"
	After  string `json:"after"`  // "RandAlphaNumericLower(40)"
}

// setCustomHeaders of the req.
func (w *WebHook) setCustomHeaders(req *http.Request) {
	var customHeaders *Headers
	// trunk-ignore(golangci-lint/gocritic)
	if w.CustomHeaders != nil {
		customHeaders = w.CustomHeaders
	} else if w.Main.CustomHeaders != nil {
		customHeaders = w.Main.CustomHeaders
	} else if w.Defaults.CustomHeaders != nil {
		customHeaders = w.Defaults.CustomHeaders
	}
	if customHeaders == nil {
		return
	}

	serviceInfo := util.ServiceInfo{
		ID:            *w.ServiceStatus.ServiceID,
		LatestVersion: w.ServiceStatus.GetLatestVersion()}
	for _, header := range *customHeaders {
		value := util.TemplateString(header.Value, serviceInfo)
		req.Header[header.Key] = []string{value}
	}
}
