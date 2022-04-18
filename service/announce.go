// Copyright [2022] [Hymenaios]
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

package service

import (
	"encoding/json"

	api_types "github.com/hymenaios-io/Hymenaios/web/api/types"
)

// AnnounceFirstVersion of a Service to the `s.Announce` channel
// (Broadcast to all WebSocket clients).
func (s *Service) AnnounceFirstVersion() {
	var payloadData []byte

	wsPage := "APPROVALS"
	wsType := "VERSION"
	wsSubType := "INIT"
	serviceWebURL := s.GetWebURL()
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    &wsPage,
		Type:    &wsType,
		SubType: &wsSubType,
		ServiceData: &api_types.ServiceSummary{
			ID:  s.ID,
			URL: &serviceWebURL,
			Status: &api_types.Status{
				LatestVersion:          s.Status.LatestVersion,
				LatestVersionTimestamp: s.Status.LatestVersionTimestamp,
			},
		},
	})

	if s.Announce != nil {
		*s.Announce <- payloadData
	}
}

// AnnounceQuery to the `s.Announce` channel
// (Broadcast to all WebSocket clients).
func (s *Service) AnnounceQuery() {
	var payloadData []byte

	wsPage := "APPROVALS"
	wsType := "VERSION"
	wsSubType := "QUERY"
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    &wsPage,
		Type:    &wsType,
		SubType: &wsSubType,
		ServiceData: &api_types.ServiceSummary{
			ID: s.ID,
			Status: &api_types.Status{
				LastQueried: s.Status.LastQueried,
			},
		},
	})

	if s.Announce != nil {
		*s.Announce <- payloadData
	}
}

// AnnounceQueryNewVersion to the `s.Announce` channel
// (Broadcast to all WebSocket clients).
func (s *Service) AnnounceQueryNewVersion() {
	var payloadData []byte

	// Last query time update OR approvel/approved
	wsPage := "APPROVALS"
	wsType := "VERSION"
	wsSubType := "NEW"
	serviceWebURL := s.GetWebURL()
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    &wsPage,
		Type:    &wsType,
		SubType: &wsSubType,
		ServiceData: &api_types.ServiceSummary{
			ID:  s.ID,
			URL: &serviceWebURL,
			Status: &api_types.Status{
				LatestVersion:          s.Status.LatestVersion,
				LatestVersionTimestamp: s.Status.LatestVersionTimestamp,
			},
		},
	})

	if s.Announce != nil {
		*s.Announce <- payloadData
	}
}

// AnnounceUpdate being applied to the `s.Announce` channel
// (Broadcast to all WebSocket clients).
func (s *Service) AnnounceUpdate() {
	var payloadData []byte

	// CurrentVersion update
	wsPage := "APPROVALS"
	wsType := "VERSION"
	wsSubType := "UPDATED"
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    &wsPage,
		Type:    &wsType,
		SubType: &wsSubType,
		ServiceData: &api_types.ServiceSummary{
			ID: s.ID,
			Status: &api_types.Status{
				CurrentVersion:          s.Status.CurrentVersion,
				CurrentVersionTimestamp: s.Status.CurrentVersionTimestamp,
			},
		},
	})

	if s.Announce != nil {
		*s.Announce <- payloadData
	}
}

// AnnounceSkip of an update `s.Announce` channel
// (Broadcast to all WebSocket clients).
func (s *Service) AnnounceSkip() {
	var payloadData []byte

	// Last query time update OR approvel/approved
	wsPage := "APPROVALS"
	wsType := "VERSION"
	wsSubType := "SKIPPED"
	payloadData, _ = json.Marshal(api_types.WebSocketMessage{
		Page:    &wsPage,
		Type:    &wsType,
		SubType: &wsSubType,
		ServiceData: &api_types.ServiceSummary{
			ID: s.ID,
			Status: &api_types.Status{
				ApprovedVersion: s.Status.ApprovedVersion,
			},
		},
	})

	if s.Announce != nil {
		*s.Announce <- payloadData
	}
}
