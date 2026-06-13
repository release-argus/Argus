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

// Package status provides the status functionality to keep track of the approved/deployed/latest versions of a Service.
package status

import (
	"github.com/release-argus/Argus/config/decode"
	apitype "github.com/release-argus/Argus/web/api/types"
)

// AnnounceFirstVersion broadcasts the first retrieved latest version to WebSocket clients.
func (s *Status) AnnounceFirstVersion() {
	var payloadData []byte

	webURL := s.ServiceInfo.GetWebURL()

	payloadData, _ = decode.Marshal(
		"json", apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "VERSION",
			SubType: "INIT",
			ServiceData: &apitype.ServiceSummary{
				ID:     s.ServiceInfo.ID,
				WebURL: &webURL,
				Status: &apitype.Status{
					LatestVersion:          s.LatestVersion(),
					LatestVersionTimestamp: s.LatestVersionTimestamp(),
				},
			},
		},
	)

	s.SendAnnounce(&payloadData)
}

// AnnounceQuery broadcasts a latest version query timestamp to WebSocket clients.
func (s *Status) AnnounceQuery() {
	var payloadData []byte

	payloadData, _ = decode.Marshal(
		"json", apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "VERSION",
			SubType: "QUERY",
			ServiceData: &apitype.ServiceSummary{
				ID: s.ServiceInfo.ID,
				Status: &apitype.Status{
					LastQueried: s.LastQueried(),
				},
			},
		},
	)

	s.SendAnnounce(&payloadData)
}

// AnnounceQueryNewVersion broadcasts a new latest version to WebSocket clients.
func (s *Status) AnnounceQueryNewVersion() {
	var payloadData []byte

	webURL := s.ServiceInfo.GetWebURL()

	// Last query time update OR approval/approved.
	payloadData, _ = decode.Marshal(
		"json", apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "VERSION",
			SubType: "NEW",
			ServiceData: &apitype.ServiceSummary{
				ID:     s.ServiceInfo.ID,
				WebURL: &webURL,
				Status: &apitype.Status{
					LatestVersion:          s.LatestVersion(),
					LatestVersionTimestamp: s.LatestVersionTimestamp(),
				},
			},
		},
	)

	s.SendAnnounce(&payloadData)
}

// AnnounceUpdate broadcasts a deployed version change to WebSocket clients.
func (s *Status) AnnounceUpdate() {
	var payloadData []byte

	// DeployedVersion update.
	payloadData, _ = decode.Marshal(
		"json", apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "VERSION",
			SubType: "UPDATED",
			ServiceData: &apitype.ServiceSummary{
				ID: s.ServiceInfo.ID,
				Status: &apitype.Status{
					DeployedVersion:          s.DeployedVersion(),
					DeployedVersionTimestamp: s.DeployedVersionTimestamp(),
				},
			},
		},
	)

	s.SendAnnounce(&payloadData)
}

// announceApproved broadcasts an approval or skip action to WebSocket clients.
func (s *Status) announceApproved() {
	var payloadData []byte

	// Last query time update OR approval/approved.
	payloadData, _ = decode.Marshal(
		"json", apitype.WebSocketMessage{
			Page:    "APPROVALS",
			Type:    "VERSION",
			SubType: "ACTION",
			ServiceData: &apitype.ServiceSummary{
				ID: s.ServiceInfo.ID,
				Status: &apitype.Status{
					ApprovedVersion: s.ServiceInfo.ApprovedVersion,
				},
			},
		},
	)

	s.SendAnnounce(&payloadData)
}
