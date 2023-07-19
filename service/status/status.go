// Copyright [2023] [Argus]
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

package svcstatus

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// statusBase is the base struct for the Status struct.
type statusBase struct {
	// Announces
	AnnounceChannel *chan []byte         // Announce to the WebSocket
	DatabaseChannel *chan dbtype.Message // Channel for broadcasts to the Database
	SaveChannel     *chan bool           // Channel for triggering a save of the config
}

// StatusDefaults are the default values for the Status struct.
type StatusDefaults struct {
	statusBase
}

// New StatusDefaults struct.
func NewStatusDefaults(
	announceChannel *chan []byte,
	databaseChannel *chan dbtype.Message,
	saveChannel *chan bool,
) StatusDefaults {
	return StatusDefaults{
		statusBase: statusBase{
			AnnounceChannel: announceChannel,
			DatabaseChannel: databaseChannel,
			SaveChannel:     saveChannel}}
}

// Status is the current state of the Service element (version and regex misses).
type Status struct {
	statusBase `yaml:"-" json:"-"`

	ServiceID *string `yaml:"-" json:"-"` // ID of the Service
	WebURL    *string `yaml:"-" json:"-"` // Web URL of the Service

	approvedVersion          string       // The version that's been approved
	deployedVersion          string       // Track the deployed version of the service from the last successful WebHook.
	deployedVersionTimestamp string       // UTC timestamp of DeployedVersion being changed.
	latestVersion            string       // Latest version found from query().
	latestVersionTimestamp   string       // UTC timestamp of LatestVersion being changed.
	lastQueried              string       // UTC timestamp that version was last queried/checked.
	regexMissesContent       uint         // Counter for the number of regex misses on URL content.
	regexMissesVersion       uint         // Counter for the number of regex misses on version.
	Fails                    Fails        // Track the Notify/WebHook fails
	deleting                 bool         // Flag to indicate the service is being deleted
	mutex                    sync.RWMutex // Lock for the Status
}

// New Status struct.
func New(
	announceChannel *chan []byte,
	databaseChannel *chan dbtype.Message,
	saveChannel *chan bool,

	av string,
	dv string,
	dvT string,
	lv string,
	lvT string,
	lq string,
) *Status {
	return &Status{
		statusBase: statusBase{
			AnnounceChannel: announceChannel,
			DatabaseChannel: databaseChannel,
			SaveChannel:     saveChannel},
		approvedVersion:          av,
		deployedVersion:          dv,
		deployedVersionTimestamp: dvT,
		latestVersion:            lv,
		latestVersionTimestamp:   lvT,
		lastQueried:              lq}
}

// String returns a string representation of the Status.
func (s *Status) String() string {
	s.mutex.RLock()
	fields := []util.Field{
		{Name: "approved_version", Value: s.approvedVersion},
		{Name: "deployed_version", Value: s.deployedVersion},
		{Name: "deployed_version_timestamp", Value: s.deployedVersionTimestamp},
		{Name: "latest_version", Value: s.latestVersion},
		{Name: "latest_version_timestamp", Value: s.latestVersionTimestamp},
		{Name: "last_queried", Value: s.lastQueried},
		{Name: "regex_misses_content", Value: s.regexMissesContent},
		{Name: "regex_misses_version", Value: s.regexMissesVersion},
		{Name: "fails", Value: &s.Fails},
	}
	s.mutex.RUnlock()

	var buf bytes.Buffer
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			if v != "" {
				fmt.Fprint(&buf, f.Name, ": ", v, ", ")
			}
		case uint:
			if v != 0 {
				fmt.Fprint(&buf, f.Name, ": ", v, ", ")
			}
		case *Fails:
			if fails := v.String(); fails != "" {
				fmt.Fprint(&buf, f.Name, ": {", fails, "}, ")
			}
		}
	}

	return strings.TrimSuffix(buf.String(), ", ")
}

func (s *Status) SetAnnounceChannel(channel *chan []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.AnnounceChannel = channel
}

// Init initialises the Status vars when more than the default value is needed.
func (s *Status) Init(
	shoutrrrs int,
	commands int,
	webhooks int,
	serviceID *string,
	webURL *string,
) {
	s.Fails.Shoutrrr.Init(shoutrrrs)
	s.Fails.Command.Init(commands)
	s.Fails.WebHook.Init(webhooks)

	s.ServiceID = serviceID
	s.WebURL = webURL
}

// LastQueried time of the LatestVersion.
func (s *Status) LastQueried() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastQueried
}

// SetLastQueried will update LastQueried to `t`, or now if `t` is empty.
func (s *Status) SetLastQueried(t string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if t == "" {
		s.lastQueried = time.Now().UTC().Format(time.RFC3339)
	} else {
		s.lastQueried = t
	}
}

// ApprovedVersion returns the ApprovedVersion.
func (s *Status) ApprovedVersion() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.approvedVersion
}

// SetApprovedVersion.
func (s *Status) SetApprovedVersion(version string, writeToDB bool) {
	s.mutex.Lock()
	{
		s.approvedVersion = version
	}
	s.mutex.Unlock()

	if writeToDB {
		s.mutex.RLock()
		// Update metrics if we're acting on the latest version
		if strings.HasSuffix(s.approvedVersion, s.latestVersion) {
			value := float64(3) // Skipping latest version
			if s.approvedVersion == s.latestVersion {
				value = 2 // Approving latest version
			}
			if s.ServiceID != nil {
				metric.SetPrometheusGauge(metric.LatestVersionIsDeployed,
					*s.ServiceID,
					value)
			}
		}

		// WebSocket
		s.announceApproved()
		// Database
		message := dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "approved_version", Value: version}}}
		s.sendDatabase(&message)
		s.mutex.RUnlock()
	}
}

// DeployedVersion returns the DeployedVersion.
func (s *Status) DeployedVersion() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.deployedVersion
}

// SetDeployedVersion will set DeployedVersion as well as DeployedVersionTimestamp.
func (s *Status) SetDeployedVersion(version string, writeToDB bool) {
	s.mutex.Lock()
	{
		s.deployedVersion = version
		s.deployedVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
		// Reset ApprovedVersion if we're on it
		if version == s.approvedVersion {
			s.approvedVersion = ""
		}
	}
	s.mutex.Unlock()

	// Write to the database if we're not deleting and have a channel
	if writeToDB {
		s.mutex.RLock()
		s.setLatestVersionIsDeployedMetric()

		// Clear the fail status of WebHooks/Commands
		s.Fails.resetFails()

		message := dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "deployed_version", Value: s.deployedVersion},
				{Column: "deployed_version_timestamp", Value: s.deployedVersionTimestamp}}}
		s.sendDatabase(&message)
		s.mutex.RUnlock()
	}
}

// DeployedVersionTimestamp returns the DeployedVersionTimestamp.
func (s *Status) DeployedVersionTimestamp() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.deployedVersionTimestamp
}

// SetDeployedVersionTimeestamp will set DeployedVersionTimestamp to `timestamp`.
func (s *Status) SetDeployedVersionTimestamp(timestamp string) {
	s.mutex.Lock()
	{
		s.deployedVersionTimestamp = timestamp
	}
	s.mutex.Unlock()
}

// LatestVersion returns the latest version.
func (s *Status) LatestVersion() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.latestVersion
}

// SetLatestVersion will set LatestVersion to `version` and LatestVersionTimestamp to s.LastQueried.
func (s *Status) SetLatestVersion(version string, writeToDB bool) {
	s.mutex.Lock()
	{
		s.latestVersion = version
		s.latestVersionTimestamp = s.lastQueried
	}
	s.mutex.Unlock()

	// Write to the database if we're not deleting and have a channel
	if writeToDB {
		s.mutex.RLock()
		s.setLatestVersionIsDeployedMetric()

		// Clear the fail status of WebHooks/Commands
		s.Fails.resetFails()

		message := dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "latest_version", Value: s.latestVersion},
				{Column: "latest_version_timestamp", Value: s.latestVersionTimestamp}}}
		s.sendDatabase(&message)
		s.mutex.RUnlock()
	}
}

// LatestVersionTimestamp returns the timestamp of the latest version.
func (s *Status) LatestVersionTimestamp() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.latestVersionTimestamp
}

// SetLatestVersionTimeestamp will set LatestVersionTimestamp to `timestamp`.
func (s *Status) SetLatestVersionTimestamp(timestamp string) {
	s.mutex.Lock()
	{
		s.latestVersionTimestamp = timestamp
	}
	s.mutex.Unlock()
}

// RegexMissContent will increment the count of RegEx misses on content.
func (s *Status) RegexMissContent() {
	s.mutex.Lock()
	{
		s.regexMissesContent++
	}
	s.mutex.Unlock()
}

// RegexMissesContent will return the number of RegEx misses on content.
func (s *Status) RegexMissesContent() uint {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.regexMissesContent
}

// RegexMissVersion will increment the count of RegEx misses on version.
func (s *Status) RegexMissVersion() {
	s.mutex.Lock()
	{
		s.regexMissesVersion++
	}
	s.mutex.Unlock()
}

// RegexMissesVersion will return the number of RegEx misses on version.
func (s *Status) RegexMissesVersion() uint {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.regexMissesVersion
}

// ResetRegexMisses (the counters for RegEx misses).
func (s *Status) ResetRegexMisses() {
	s.mutex.Lock()
	{
		s.regexMissesContent = 0
		s.regexMissesVersion = 0
	}
	s.mutex.Unlock()
}

// SetDeleting will set the Service to be deleted.
func (s *Status) SetDeleting() {
	s.mutex.Lock()
	{
		s.deleting = true
	}
	s.mutex.Unlock()
}

// Deleting returns true if the Service is being deleted.
func (s *Status) Deleting() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.deleting
}

// SendAnnounce payload to the AnnounceChannel.
func (s *Status) SendAnnounce(payload *[]byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.deleting || s.AnnounceChannel == nil {
		return
	}

	*s.AnnounceChannel <- *payload
}

// sendDatabase payload to the DatabaseChannel.
func (s *Status) sendDatabase(payload *dbtype.Message) {
	if s.deleting || s.DatabaseChannel == nil {
		return
	}

	*s.DatabaseChannel <- *payload
}

// SendSave request to the SaveChannel.
func (s *Status) SendSave() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.deleting || s.SaveChannel == nil {
		return
	}

	*s.SaveChannel <- true
}

// GetWebURL returns the Web URL.
func (s *Status) GetWebURL() string {
	if util.DefaultIfNil(s.WebURL) == "" {
		return ""
	}

	return util.TemplateString(
		*s.WebURL,
		util.ServiceInfo{LatestVersion: s.LatestVersion()})
}

// setLatestVersionIsDeployedMetric will set the metric for whether the latest version is currently deployed.
func (s *Status) setLatestVersionIsDeployedMetric() {
	if s.ServiceID == nil {
		return
	}

	value := float64(0) // Not deployed
	if s.latestVersion == s.deployedVersion {
		value = 1 // Is deployed

		// Latest version isn't deployed, but has been approved/skipped, so carry that over
	} else if strings.HasSuffix(s.approvedVersion, s.latestVersion) {
		value = 3 // Latest version was skipped
		if s.approvedVersion == s.latestVersion {
			value = 2 // Latest version was approved
		}
	}
	metric.SetPrometheusGauge(metric.LatestVersionIsDeployed,
		*s.ServiceID,
		value)
}

// InitMetrics for the Status.
func (s *Status) InitMetrics() {
	if s == nil || s.ServiceID == nil {
		return
	}

	s.setLatestVersionIsDeployedMetric()
}

// DeleteMetrics of the Status.
func (s *Status) DeleteMetrics() {
	if s == nil || s.ServiceID == nil {
		return
	}

	metric.DeletePrometheusGauge(metric.LatestVersionIsDeployed,
		*s.ServiceID)
}
