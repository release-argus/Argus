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

// Package status provides the status functionality to keep track of the approved/deployed/latest versions of a Service.
package status

import (
	"fmt"
	"strings"
	"sync"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// statusBase is the base struct for the Status struct.
type statusBase struct {
	AnnounceChannel *chan []byte         // Announce to the WebSocket.
	DatabaseChannel *chan dbtype.Message // Broadcasts to the Database.
	SaveChannel     *chan bool           // Trigger a save of the config.
}

// Defaults are the default values for the Status struct.
type Defaults struct {
	statusBase
}

// NewDefaults returns a new Defaults struct.
func NewDefaults(
	announceChannel *chan []byte,
	databaseChannel *chan dbtype.Message,
	saveChannel *chan bool,
) Defaults {
	return Defaults{
		statusBase: statusBase{
			AnnounceChannel: announceChannel,
			DatabaseChannel: databaseChannel,
			SaveChannel:     saveChannel}}
}

// Status holds the current versioning state of the Service (version, and regex misses).
type Status struct {
	statusBase

	ServiceID   *string `yaml:"-" json:"-"` // ID of the Service.
	ServiceName *string `yaml:"-" json:"-"` // Name of the Service.
	WebURL      *string `yaml:"-" json:"-"` // Web URL of the Service.

	mutex                    sync.RWMutex // Lock for the Status.
	approvedVersion          string       // The version of the Service that has been approved for deployment.
	deployedVersion          string       // The version of the Service that is deployed.
	deployedVersionTimestamp string       // UTC timestamp of latest DeployedVersion change.
	latestVersion            string       // The latest version of the Service found from query().
	latestVersionTimestamp   string       // UTC timestamp of latest LatestVersion change.
	lastQueried              string       // UTC timestamp of latest LatestVersion query.
	regexMissesContent       uint         // Counter for the amount of regex misses on the URL content.
	regexMissesVersion       uint         // Counter for the amount of regex misses on the version.
	Fails                    Fails        // Track the Notify/WebHook fails.
	deleting                 bool         // Flag to indicate undergoing deletion.
}

// New Status struct.
func New(
	announceChannel *chan []byte,
	databaseChannel *chan dbtype.Message,
	saveChannel *chan bool,

	av string,
	dv, dvT string,
	lv, lvT string,
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

// Copy the Status.
func (s *Status) Copy() *Status {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return New(
		s.AnnounceChannel,
		s.DatabaseChannel,
		s.SaveChannel,
		s.approvedVersion,
		s.deployedVersion,
		s.deployedVersionTimestamp,
		s.latestVersion,
		s.latestVersionTimestamp,
		s.lastQueried)
}

// String returns a string representation of the Status.
func (s *Status) String() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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

	var builder strings.Builder
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			if v != "" {
				builder.WriteString(
					fmt.Sprintf("%s: %v\n",
						f.Name, v))
			}
		case uint:
			if v != 0 {
				builder.WriteString(
					fmt.Sprintf("%s: %d\n",
						f.Name, v))
			}
		case *Fails:
			if fails := v.String("  "); fails != "" {
				builder.WriteString(
					fmt.Sprintf("%s:\n%s",
						f.Name, fails))
			}
		}
	}

	result := builder.String()
	return result
}

// SetAnnounceChannel will set the AnnounceChannel.
func (s *Status) SetAnnounceChannel(channel *chan []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.AnnounceChannel = channel
}

// Init will initialise the Status struct, creating the failure trackers.
func (s *Status) Init(
	shoutrrrs, commands, webhooks int,
	serviceID, serviceName *string,
	webURL *string,
) {
	s.Fails.Shoutrrr.Init(shoutrrrs)
	s.Fails.Command.Init(commands)
	s.Fails.WebHook.Init(webhooks)

	s.ServiceID = serviceID
	s.ServiceName = serviceName

	s.WebURL = webURL
}

// LastQueried time of the LatestVersion.
func (s *Status) LastQueried() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.lastQueried
}

// SetLastQueried will update LastQueried to `t` (or now if `t` empty).
func (s *Status) SetLastQueried(t string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if t == "" {
		s.lastQueried = time.Now().UTC().Format(time.RFC3339)
	} else {
		s.lastQueried = t
	}
}

// SameVersions returns whether the Status has the same versions as `s`.
func (s *Status) SameVersions(s2 *Status) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s2.mutex.RLock()
	defer s2.mutex.RUnlock()

	return s.approvedVersion == s2.approvedVersion &&
		s.latestVersion == s2.latestVersion &&
		s.deployedVersion == s2.deployedVersion
}

// ApprovedVersion returns the ApprovedVersion.
func (s *Status) ApprovedVersion() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.approvedVersion
}

// SetApprovedVersion will set .ApprovedVersion to `version`.
func (s *Status) SetApprovedVersion(version string, writeToDB bool) {
	s.mutex.Lock()
	// Do not modify if unchanged, or deleting.
	if s.approvedVersion == version || s.deleting {
		s.mutex.Unlock()
		return
	}

	previousApprovedVersion := s.approvedVersion
	s.approvedVersion = version
	s.mutex.Unlock()

	if writeToDB {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		s.setLatestVersionIsDeployedMetric()
		s.updateUpdatesCurrent(previousApprovedVersion, s.latestVersion, s.deployedVersion)

		// Update metrics if acting on the LatestVersion.
		if strings.HasSuffix(s.approvedVersion, s.latestVersion) {
			value := float64(3) // Skipping LatestVersion.
			if s.approvedVersion == s.latestVersion {
				value = 2 // Approving LatestVersion.
			}
			metric.SetPrometheusGauge(metric.LatestVersionIsDeployed,
				*s.ServiceID, "",
				value)
		}

		// WebSocket.
		s.announceApproved()
		// Database.
		message := dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "approved_version", Value: version}}}
		s.sendDatabase(&message)
	}
}

// DeployedVersion returns the DeployedVersion.
func (s *Status) DeployedVersion() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.deployedVersion
}

// SetDeployedVersion sets the DeployedVersion to `version` and DeployedVersionTimestamp to `releaseDate`
// (or now if empty).
func (s *Status) SetDeployedVersion(version, releaseDate string, writeToDB bool) {
	s.mutex.Lock()
	// Do not modify if unchanged, or deleting.
	if s.deployedVersion == version || s.deleting {
		s.mutex.Unlock()
		return
	}

	previousDeployedVersion := s.deployedVersion
	s.deployedVersion = version
	if releaseDate != "" {
		s.deployedVersionTimestamp = releaseDate
	} else {
		s.deployedVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
	}
	// Reset ApprovedVersion if on it.
	if version == s.approvedVersion {
		s.approvedVersion = ""
	}
	s.mutex.Unlock()

	// Write to the database if not deleting and have a channel.
	if writeToDB {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		s.setLatestVersionIsDeployedMetric()
		s.updateUpdatesCurrent(s.approvedVersion, s.latestVersion, previousDeployedVersion)

		// Clear the fail status of WebHooks/Commands.
		s.Fails.resetFails()

		message := dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "deployed_version", Value: s.deployedVersion},
				{Column: "deployed_version_timestamp", Value: s.deployedVersionTimestamp}}}
		s.sendDatabase(&message)
	}
}

// DeployedVersionTimestamp returns the DeployedVersionTimestamp.
func (s *Status) DeployedVersionTimestamp() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.deployedVersionTimestamp
}

// LatestVersion returns the LatestVersion.
func (s *Status) LatestVersion() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.latestVersion
}

// SetLatestVersion sets the LatestVersion to `version`, and LatestVersionTimestamp to `releaseDate`
// (or now if empty).
func (s *Status) SetLatestVersion(version, releaseDate string, writeToDB bool) {
	s.mutex.Lock()
	// Do not modify if unchanged, or deleting.
	if s.latestVersion == version || s.deleting {
		s.mutex.Unlock()
		return
	}

	previousLatestVersion := s.latestVersion
	s.latestVersion = version
	if releaseDate != "" {
		s.latestVersionTimestamp = releaseDate
	} else {
		s.latestVersionTimestamp = s.lastQueried
	}
	s.mutex.Unlock()

	// Write to the database if not deleting, and have a channel.
	if writeToDB {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		s.updateUpdatesCurrent(s.approvedVersion, previousLatestVersion, s.deployedVersion)
		s.setLatestVersionIsDeployedMetric()

		// Clear the fail status of WebHooks/Commands.
		s.Fails.resetFails()

		message := dbtype.Message{
			ServiceID: *s.ServiceID,
			Cells: []dbtype.Cell{
				{Column: "latest_version", Value: s.latestVersion},
				{Column: "latest_version_timestamp", Value: s.latestVersionTimestamp}}}
		s.sendDatabase(&message)
	}
}

// LatestVersionTimestamp returns the timestamp of the LatestVersion.
func (s *Status) LatestVersionTimestamp() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.latestVersionTimestamp
}

// RegexMissContent increments the count of RegEx misses on content.
func (s *Status) RegexMissContent() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.regexMissesContent++
}

// RegexMissesContent returns the count of RegEx misses on content.
func (s *Status) RegexMissesContent() uint {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.regexMissesContent
}

// RegexMissVersion increments the count of RegEx misses on version.
func (s *Status) RegexMissVersion() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.regexMissesVersion++
}

// RegexMissesVersion returns the count of RegEx misses on version.
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

// SetDeleting will set `deleting` flag.
func (s *Status) SetDeleting() {
	s.mutex.Lock()
	{
		s.deleting = true
	}
	s.mutex.Unlock()
}

// Deleting returns whether the Service is undergoing deletion.
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
func (s *Status) GetWebURL() *string {
	if util.DereferenceOrDefault(s.WebURL) == "" {
		return nil
	}

	webURL := util.TemplateString(
		*s.WebURL,
		util.ServiceInfo{LatestVersion: s.LatestVersion()})
	return &webURL
}

// setLatestVersionIsDeployedMetric sets the Prometheus metric for whether the LatestVersion is deployed.
func (s *Status) setLatestVersionIsDeployedMetric() {
	metric.SetPrometheusGauge(metric.LatestVersionIsDeployed,
		*s.ServiceID, "",
		metric.GetVersionDeployedState(s.approvedVersion, s.latestVersion, s.deployedVersion))
}

// updateUpdatesCurrent updates the Prometheus metric `UpdatesCurrent`
// to reflect changes in the deployment state of the LatestVersion.
// It compares the previous deployment state with the current state and adjusts the metric accordingly.
// If the deployment state hasn't changed, no updates are made.
func (s *Status) updateUpdatesCurrent(previousApprovedVersion, previousLatestVersion, previousDeployedVersion string) {
	previousValue := metric.GetVersionDeployedState(previousApprovedVersion, previousLatestVersion, previousDeployedVersion)
	newValue := metric.GetVersionDeployedState(s.approvedVersion, s.latestVersion, s.deployedVersion)
	// No change.
	if previousValue == newValue {
		return
	}

	metric.SetUpdatesCurrent(-1, previousValue)
	metric.SetUpdatesCurrent(1, newValue)
}

// InitMetrics for the Status.
func (s *Status) InitMetrics() {
	s.setLatestVersionIsDeployedMetric()
	metric.SetUpdatesCurrent(1,
		metric.GetVersionDeployedState(s.approvedVersion, s.latestVersion, s.deployedVersion))
}

// DeleteMetrics of the Status.
func (s *Status) DeleteMetrics() {
	metric.DeletePrometheusGauge(metric.LatestVersionIsDeployed,
		*s.ServiceID, "")
	metric.SetUpdatesCurrent(-1,
		metric.GetVersionDeployedState(s.approvedVersion, s.latestVersion, s.deployedVersion))
}
