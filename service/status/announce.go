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

package service_status

import (
	"encoding/json"

	api_types "github.com/release-argus/Argus/web/api/types"
)

// AnnounceFirstVersion of a Service to the `s.AnnounceChannel`
// (Broadcast to all WebSocket clients).
func (s *Status) AnnounceFirstVersion() {
	var payloadData []byte

	serviceWebURL := s.GetWebURL()
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "VERSION",
		SubType: "INIT",
		ServiceData: &api_types.ServiceSummary{
			ID:  *s.ServiceID,
			URL: &serviceWebURL,
			Status: &api_types.Status{
				LatestVersion:          s.LatestVersion,
				LatestVersionTimestamp: s.LatestVersionTimestamp,
			},
		},
	})

	if s.AnnounceChannel != nil {
		*s.AnnounceChannel <- payloadData
	}
}

// AnnounceQuery to the `s.AnnounceChannel`
// (Broadcast to all WebSocket clients).
func (s *Status) AnnounceQuery() {
	var payloadData []byte

	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "VERSION",
		SubType: "QUERY",
		ServiceData: &api_types.ServiceSummary{
			ID: *s.ServiceID,
			Status: &api_types.Status{
				LastQueried: s.LastQueried,
			},
		},
	})

	if s.AnnounceChannel != nil {
		*s.AnnounceChannel <- payloadData
	}
}

// AnnounceQueryNewVersion to the `s.AnnounceChannel`
// (Broadcast to all WebSocket clients).
func (s *Status) AnnounceQueryNewVersion() {
	var payloadData []byte

	// Last query time update OR approvel/approved
	serviceWebURL := s.GetWebURL()
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "VERSION",
		SubType: "NEW",
		ServiceData: &api_types.ServiceSummary{
			ID:  *s.ServiceID,
			URL: &serviceWebURL,
			Status: &api_types.Status{
				LatestVersion:          s.LatestVersion,
				LatestVersionTimestamp: s.LatestVersionTimestamp,
			},
		},
	})

	if s.AnnounceChannel != nil {
		*s.AnnounceChannel <- payloadData
	}
}

// AnnounceUpdate being applied to the `s.AnnounceChannel`
// (Broadcast to all WebSocket clients).
func (s *Status) AnnounceUpdate() {
	var payloadData []byte

	// DeployedVersion update
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "VERSION",
		SubType: "UPDATED",
		ServiceData: &api_types.ServiceSummary{
			ID: *s.ServiceID,
			Status: &api_types.Status{
				DeployedVersion:          s.DeployedVersion,
				DeployedVersionTimestamp: s.DeployedVersionTimestamp,
			},
		},
	})

	if s.AnnounceChannel != nil {
		*s.AnnounceChannel <- payloadData
	}
}

// AnnounceAction on an update (skip/approve) to the `s.AnnounceChannel`
// (Broadcast to all WebSocket clients).
func (s *Status) AnnounceApproved() {
	var payloadData []byte

	// Last query time update OR approvel/approved
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    "APPROVALS",
		Type:    "VERSION",
		SubType: "ACTION",
		ServiceData: &api_types.ServiceSummary{
			ID: *s.ServiceID,
			Status: &api_types.Status{
				ApprovedVersion: s.ApprovedVersion,
			},
		},
	})

	if s.AnnounceChannel != nil {
		*s.AnnounceChannel <- payloadData
	}
}
