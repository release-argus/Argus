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

package apitype

import "encoding/json"

// WebSocketMessage is the message format to send/receive/forward.
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
	InfoData    *Info                      `json:"info_data,omitempty"`
	FlagsData   *Flags                     `json:"flags_data,omitempty"`
	ConfigData  *Config                    `json:"config_data,omitempty"`
}

// String returns a string representation of the WebSocketMessage.
func (w *WebSocketMessage) String() string {
	jsonBytes, _ := json.Marshal(w)
	return string(jsonBytes)
}
