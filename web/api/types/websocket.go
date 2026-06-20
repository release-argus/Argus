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

// Package types provides the types for the Argus API.
package types

import "github.com/release-argus/Argus/config/decode"

// WebSocketMessage is the message format to send/receive.
type WebSocketMessage struct {
	Version     *int                       `json:"version,omitempty"`
	Page        string                     `json:"page"`
	Type        string                     `json:"type"`
	SubType     string                     `json:"sub_type,omitempty"`
	Target      *string                    `json:"target,omitempty"`
	Order       *[]string                  `json:"order,omitempty"`
	ServiceData *ServiceSummary            `json:"service_data,omitempty"`
	CommandData map[string]*CommandSummary `json:"command_data,omitempty"`
	WebHookData map[string]*WebHookSummary `json:"webhook_data,omitempty"`
}

// String implements fmt.Stringer and returns a JSON representation.
func (w *WebSocketMessage) String() string {
	return decode.ToJSONString(w)
}
