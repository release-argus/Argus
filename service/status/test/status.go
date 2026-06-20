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

// Package test provides test helpers for the status package.
package test

import (
	"fmt"

	"github.com/release-argus/Argus/config/decode"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/service/status/info"
)

type statusTimestamps struct {
	DeployedVersionTimestamp string `json:"deployed_version_timestamp" yaml:"deployed_version_timestamp"`
	LatestVersionTimestamp   string `json:"latest_version_timestamp" yaml:"latest_version_timestamp"`
	LastQueried              string `json:"last_queried" yaml:"last_queried"`
}

// New decodes format-encoded status fields into a Status with test channels.
func New(format string, data []byte) (*status.Status, error) {
	var field status.Status

	// Channels.
	var (
		announceChannel = make(chan []byte, 8)
		databaseChannel = make(chan dbtype.Message, 8)
		saveChannel     = make(chan bool, 8)
	)
	field.AnnounceChannel = announceChannel
	field.DatabaseChannel = databaseChannel
	field.SaveChannel = saveChannel

	field.Init(
		0, 0, 0,
		status.ServiceInfo{},
		&dashboard.Options{},
	)

	// Versions.
	serviceInfo := info.ServiceInfo{}
	if err := decode.Unmarshal(format, data, &serviceInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ServiceInfo: %w", err)
	}

	// Timestamps.
	timestamps := statusTimestamps{}
	if err := decode.Unmarshal(format, data, &timestamps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal timestamps: %w", err)
	}
	if serviceInfo.DeployedVersion != "" {
		field.SetDeployedVersion(serviceInfo.DeployedVersion, timestamps.DeployedVersionTimestamp, false)
		field.ServiceInfo.DeployedVersion = serviceInfo.DeployedVersion
	}
	if serviceInfo.LatestVersion != "" {
		field.SetLatestVersion(serviceInfo.LatestVersion, timestamps.LatestVersionTimestamp, false)
		field.ServiceInfo.LatestVersion = serviceInfo.LatestVersion
	}
	if serviceInfo.ApprovedVersion != "" {
		field.SetApprovedVersion(serviceInfo.ApprovedVersion, false)
		field.ServiceInfo.ApprovedVersion = serviceInfo.ApprovedVersion
	}
	if timestamps.LastQueried != "" {
		field.SetLastQueried(timestamps.LastQueried)
	}

	return &field, nil
}
