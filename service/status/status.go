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
	"strconv"
	"strings"
	"sync"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/dashboard"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// statusBase is the base struct for the Status struct.
type statusBase struct {
	AnnounceChannel chan []byte         // Announce to the WebSocket.
	DatabaseChannel chan dbtype.Message // Broadcasts to the Database.
	SaveChannel     chan bool           // Trigger a save of the config.
}

// Defaults are the default values for the Status struct.
type Defaults struct {
	statusBase
}

// NewDefaults returns a new Defaults struct.
func NewDefaults(
	announceChannel chan []byte,
	databaseChannel chan dbtype.Message,
	saveChannel chan bool,
) Defaults {
	return Defaults{
		statusBase: statusBase{
			AnnounceChannel: announceChannel,
			DatabaseChannel: databaseChannel,
			SaveChannel:     saveChannel,
		},
	}
}

// Status holds the current versioning state of the Service (version, and regex misses).
type Status struct {
	statusBase

	ServiceInfo serviceinfo.ServiceInfo // ServiceInfo holds information about the service.
	Dashboard   *dashboard.Options      // Dashboard options for the Service.

	mu                       sync.RWMutex // Lock for the Status.
	deployedVersionTimestamp string       // UTC timestamp of latest DeployedVersion change.
	latestVersionTimestamp   string       // UTC timestamp of latest LatestVersion change.
	lastQueried              string       // UTC timestamp of latest LatestVersion query.
	regexMissesContent       uint         // Counter for the number of regex misses on the URL content.
	regexMissesVersion       uint         // Counter for the number of regex misses on the version.
	Fails                    Fails        // Track the Notify/WebHook fails.
	deleting                 bool         // Flag to indicate undergoing deletion.
}

// New returns a Status populated with version fields and channel references.
func New(
	announceChannel chan []byte,
	databaseChannel chan dbtype.Message,
	saveChannel chan bool,

	av string,
	dv, dvT string,
	lv, lvT string,
	lq string,
	dashboard *dashboard.Options,
) *Status {
	status := &Status{
		statusBase: statusBase{
			AnnounceChannel: announceChannel,
			DatabaseChannel: databaseChannel,
			SaveChannel:     saveChannel,
		},
		ServiceInfo: serviceinfo.ServiceInfo{
			ApprovedVersion: av,
			DeployedVersion: dv,
			LatestVersion:   lv,
		},
		deployedVersionTimestamp: dvT,
		latestVersionTimestamp:   lvT,
		lastQueried:              lq,
		Dashboard:                dashboard,
	}

	status.ServiceInfo.SetMutex(&status.mu)

	return status
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Status) UnmarshalJSON(data []byte) error {
	s.unmarshal()
	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (s *Status) UnmarshalYAML(data []byte) error {
	s.unmarshal()
	return nil
}

// unmarshal wires ServiceInfo to the Status mutex.
func (s *Status) unmarshal() {
	// Set the mutex pointer for ServiceInfo
	s.ServiceInfo.SetMutex(&s.mu)
}

// Copy returns a deep copy of the receiver (with/without channels).
func (s *Status) Copy(withChannels bool) *Status {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	newStatus := New(
		nil, nil, nil,
		s.ServiceInfo.ApprovedVersion,
		s.ServiceInfo.DeployedVersion, s.deployedVersionTimestamp,
		s.ServiceInfo.LatestVersion, s.latestVersionTimestamp,
		s.lastQueried,
		s.Dashboard,
	)

	if withChannels {
		newStatus.AnnounceChannel = s.AnnounceChannel
		newStatus.DatabaseChannel = s.DatabaseChannel
		newStatus.SaveChannel = s.SaveChannel
	}

	newStatus.Init(
		len(s.Fails.Shoutrrr.fails), newStatus.Fails.Command.Length(), newStatus.Fails.WebHook.Length(),
		ServiceInfo{
			ID:         s.ServiceInfo.ID,
			Name:       s.ServiceInfo.Name,
			Comment:    s.ServiceInfo.Comment,
			ServiceURL: s.ServiceInfo.URL,
		},
		s.Dashboard.Copy(),
	)

	return newStatus
}

// String implements fmt.Stringer.
func (s *Status) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fields := []util.Field{
		{Name: "approved_version", Value: s.ServiceInfo.ApprovedVersion},
		{Name: "deployed_version", Value: s.ServiceInfo.DeployedVersion},
		{Name: "deployed_version_timestamp", Value: s.deployedVersionTimestamp},
		{Name: "latest_version", Value: s.ServiceInfo.LatestVersion},
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
				// '<name>: <val>\n'
				builder.WriteString(f.Name)
				builder.WriteString(": ")
				builder.WriteString(v)
				builder.WriteString("\n")
			}
		case uint:
			if v != 0 {
				// '<name>: <val>\n'
				builder.WriteString(f.Name)
				builder.WriteString(": ")
				builder.WriteString(strconv.Itoa(int(v)))
				builder.WriteString("\n")
			}
		case *Fails:
			if fails := v.String("  "); fails != "" {
				// '<name>:\n<failsStr>'
				builder.WriteString(f.Name)
				builder.WriteString(":\n")
				builder.WriteString(fails)
			}
		}
	}

	result := builder.String()
	return result
}

// ServiceInfo holds identifying metadata passed to [Status.Init].
type ServiceInfo struct {
	ID         string
	Name       string
	Comment    string
	ServiceURL string
}

// Init initialises and resets the Status.
// It reinitialises internal failure tracking with the provided capacities,
// copies service metadata into the embedded ServiceInfo, attaches dashboard
// configuration, and refreshes derived fields.
func (s *Status) Init(
	commands, shoutrrrs, webhooks int,
	serviceInfo ServiceInfo,
	dashboard *dashboard.Options,
) {
	s.Fails.Command.Init(commands)
	s.Fails.Shoutrrr.Init(shoutrrrs)
	s.Fails.WebHook.Init(webhooks)

	s.ServiceInfo.ID = serviceInfo.ID
	s.ServiceInfo.Name = util.ValueOr(serviceInfo.Name, serviceInfo.ID)
	s.ServiceInfo.Comment = serviceInfo.Comment
	s.ServiceInfo.URL = serviceInfo.ServiceURL
	s.ServiceInfo.Tags = dashboard.Tags

	s.Dashboard = dashboard
	s.ServiceInfo.SetMutex(&s.mu)
	s.refreshServiceInfo()
}

// SetAnnounceChannel sets the AnnounceChannel.
func (s *Status) SetAnnounceChannel(channel chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.AnnounceChannel = channel
}

// RefreshServiceInfo updates the ServiceInfo struct with the latest values.
// It uses the dashboard options to set the Icon, IconLinkTo, and WebURL fields.
func (s *Status) RefreshServiceInfo() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refreshServiceInfo()
}

// refreshServiceInfo is like [RefreshServiceInfo] but requires the Status mutex to already be held.
func (s *Status) refreshServiceInfo() {
	s.ServiceInfo.Icon = util.TemplateString(
		s.Dashboard.GetIcon(),
		s.ServiceInfo,
	)

	s.ServiceInfo.IconLinkTo = util.TemplateString(
		s.Dashboard.GetIconLinkTo(),
		s.ServiceInfo,
	)

	s.ServiceInfo.WebURL = util.TemplateString(
		s.Dashboard.GetWebURL(),
		s.ServiceInfo,
	)
}

// GetServiceInfo returns a snapshot of the current ServiceInfo.
func (s *Status) GetServiceInfo() serviceinfo.ServiceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ServiceInfo
}

// GetWebURL returns the WebURL field of the ServiceInfo struct.
func (s *Status) GetWebURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ServiceInfo.WebURL
}

// LastQueried returns the timestamp of the most recent LatestVersion query.
func (s *Status) LastQueried() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastQueried
}

// SetLastQueried sets LastQueried to t, or to the current UTC time if t is empty.
func (s *Status) SetLastQueried(t string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t == "" {
		s.lastQueried = time.Now().UTC().Format(time.RFC3339)
	} else {
		s.lastQueried = t
	}
}

// SameVersions reports whether the Status has the same versions as s.
func (s *Status) SameVersions(s2 *Status) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s2.mu.RLock()
	defer s2.mu.RUnlock()

	return s.ServiceInfo.ApprovedVersion == s2.ServiceInfo.ApprovedVersion &&
		s.ServiceInfo.LatestVersion == s2.ServiceInfo.LatestVersion &&
		s.ServiceInfo.DeployedVersion == s2.ServiceInfo.DeployedVersion
}

// ApprovedVersion returns the ApprovedVersion.
func (s *Status) ApprovedVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ServiceInfo.ApprovedVersion
}

// SetApprovedVersion sets ApprovedVersion to version.
func (s *Status) SetApprovedVersion(version string, writeToDB bool) {
	s.mu.Lock()

	previousServiceInfo := s.ServiceInfo
	// Do not modify if unchanged, deleting, or latest version is already deployed.
	if previousServiceInfo.ApprovedVersion == version || s.deleting ||
		previousServiceInfo.LatestVersion == previousServiceInfo.DeployedVersion {
		s.mu.Unlock()
		return
	}

	s.ServiceInfo.ApprovedVersion = version
	s.refreshServiceInfo()

	if !writeToDB {
		s.mu.Unlock()
		return
	}

	newServiceInfo := s.ServiceInfo
	s.mu.Unlock()
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Metrics.
	setLatestVersionIsDeployedMetric(newServiceInfo)
	updateUpdatesCurrentMetric(previousServiceInfo, newServiceInfo)

	// Update metrics if acting on the LatestVersion.
	isApproved := newServiceInfo.ApprovedVersion == newServiceInfo.LatestVersion
	isSkipped := newServiceInfo.ApprovedVersion == serviceinfo.SkippedVersion(newServiceInfo.LatestVersion)
	if isApproved || isSkipped {
		value := metric.LatestVersionSkipped // Skipping LatestVersion.
		if isApproved {
			value = metric.LatestVersionApproved // Approving LatestVersion.
		}
		metric.SetPrometheusGauge(
			metric.LatestVersionIsDeployed,
			newServiceInfo.ID, "",
			float64(value),
		)
	}

	// WebSocket.
	s.announceApproved()

	// Database.
	message := dbtype.Message{
		ServiceID: newServiceInfo.ID,
		Cells: []dbtype.Cell{
			{Column: "approved_version", Value: version},
		},
	}
	s.sendDatabase(&message)
}

// DeployedVersion returns the DeployedVersion.
func (s *Status) DeployedVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ServiceInfo.DeployedVersion
}

// SetDeployedVersion sets the DeployedVersion to `version` and DeployedVersionTimestamp to `releaseDate`
// (or now if empty).
func (s *Status) SetDeployedVersion(version, releaseDate string, writeToDB bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	previousServiceInfo := s.ServiceInfo
	// Do not modify if unchanged, or deleting.
	if previousServiceInfo.DeployedVersion == version || s.deleting {
		return
	}

	s.ServiceInfo.DeployedVersion = version
	if releaseDate != "" {
		s.deployedVersionTimestamp = releaseDate
	} else {
		s.deployedVersionTimestamp = time.Now().UTC().Format(time.RFC3339)
	}
	// Reset ApprovedVersion if on it (or previously skipped it).
	if version == previousServiceInfo.ApprovedVersion ||
		serviceinfo.SkippedVersion(version) == previousServiceInfo.ApprovedVersion {
		s.ServiceInfo.ApprovedVersion = ""
	}
	s.refreshServiceInfo()

	if !writeToDB {
		return
	}

	newServiceInfo := s.ServiceInfo

	// Metrics.
	setLatestVersionIsDeployedMetric(newServiceInfo)
	updateUpdatesCurrentMetric(previousServiceInfo, newServiceInfo)

	// Clear the fail status of WebHooks/Commands.
	s.Fails.resetFails()

	// Database.
	message := dbtype.Message{
		ServiceID: newServiceInfo.ID,
		Cells: []dbtype.Cell{
			{Column: "deployed_version", Value: newServiceInfo.DeployedVersion},
			{Column: "deployed_version_timestamp", Value: s.deployedVersionTimestamp},
		},
	}
	s.sendDatabase(&message)
}

// DeployedVersionTimestamp returns the DeployedVersionTimestamp.
func (s *Status) DeployedVersionTimestamp() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.deployedVersionTimestamp
}

// LatestVersion returns the LatestVersion.
func (s *Status) LatestVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ServiceInfo.LatestVersion
}

// SetLatestVersion sets the LatestVersion to `version`, and LatestVersionTimestamp to `releaseDate`
// (or now if empty).
func (s *Status) SetLatestVersion(version, releaseDate string, writeToDB bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	previousServiceInfo := s.ServiceInfo
	// Do not modify if unchanged, or deleting.
	if previousServiceInfo.LatestVersion == version || s.deleting {
		return
	}

	s.ServiceInfo.LatestVersion = version
	if releaseDate != "" {
		s.latestVersionTimestamp = releaseDate
	} else {
		s.latestVersionTimestamp = s.lastQueried
	}
	s.refreshServiceInfo()

	if !writeToDB {
		return
	}

	newServiceInfo := s.ServiceInfo

	// Metrics.
	setLatestVersionIsDeployedMetric(newServiceInfo)
	updateUpdatesCurrentMetric(previousServiceInfo, newServiceInfo)

	// Clear the fail status of WebHooks/Commands.
	s.Fails.resetFails()

	// Database.
	message := dbtype.Message{
		ServiceID: newServiceInfo.ID,
		Cells: []dbtype.Cell{
			{Column: "latest_version", Value: newServiceInfo.LatestVersion},
			{Column: "latest_version_timestamp", Value: s.latestVersionTimestamp},
		},
	}
	s.sendDatabase(&message)
}

// LatestVersionTimestamp returns the timestamp of the LatestVersion.
func (s *Status) LatestVersionTimestamp() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.latestVersionTimestamp
}

// RegexMissContent increments the count of RegEx misses on content.
func (s *Status) RegexMissContent() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.regexMissesContent++
}

// RegexMissesContent returns the count of RegEx misses on content.
func (s *Status) RegexMissesContent() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.regexMissesContent
}

// RegexMissVersion increments the count of RegEx misses on version.
func (s *Status) RegexMissVersion() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.regexMissesVersion++
}

// RegexMissesVersion returns the count of RegEx misses on version.
func (s *Status) RegexMissesVersion() uint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.regexMissesVersion
}

// ResetRegexMisses resets the content and version regex miss counters to zero.
func (s *Status) ResetRegexMisses() {
	s.mu.Lock()
	{
		s.regexMissesContent = 0
		s.regexMissesVersion = 0
	}
	s.mu.Unlock()
}

// SetDeleting marks the service as undergoing deletion.
func (s *Status) SetDeleting() {
	s.mu.Lock()
	{
		s.deleting = true
	}
	s.mu.Unlock()
}

// Deleting returns whether the Service is undergoing deletion.
func (s *Status) Deleting() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.deleting
}

// SendAnnounce sends payload to the AnnounceChannel if the service is not being deleted.
func (s *Status) SendAnnounce(payload []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.deleting || s.AnnounceChannel == nil {
		return
	}

	s.AnnounceChannel <- payload
}

// sendDatabase enqueues a database update when the service is not being deleted.
func (s *Status) sendDatabase(payload *dbtype.Message) {
	if s.deleting || s.DatabaseChannel == nil {
		return
	}

	s.DatabaseChannel <- *payload
}

// SendSave sends a save request to the SaveChannel if the service is not being deleted.
func (s *Status) SendSave() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.deleting || s.SaveChannel == nil {
		return
	}

	s.SaveChannel <- true
}

// setLatestVersionIsDeployedMetric sets the Prometheus metric for whether the LatestVersion is deployed.
func setLatestVersionIsDeployedMetric(serviceInfo serviceinfo.ServiceInfo) {
	metric.SetPrometheusGauge(
		metric.LatestVersionIsDeployed,
		serviceInfo.ID, "",
		float64(metric.GetVersionDeployedState(serviceInfo)),
	)
}

// updateUpdatesCurrentMetric adjusts the UpdatesCurrent metric when the deployment state changes.
func updateUpdatesCurrentMetric(previousServiceInfo, newServiceInfo serviceinfo.ServiceInfo) {
	previousValue := metric.GetVersionDeployedState(previousServiceInfo)
	newValue := metric.GetVersionDeployedState(newServiceInfo)
	// No change.
	if previousValue == newValue {
		return
	}

	metric.SetUpdatesCurrent(-1, previousValue)
	metric.SetUpdatesCurrent(1, newValue)
}

// InitMetrics registers status-derived Prometheus metrics for the service.
func (s *Status) InitMetrics() {
	serviceInfo := s.GetServiceInfo()

	setLatestVersionIsDeployedMetric(serviceInfo)
	metric.SetUpdatesCurrent(1, metric.GetVersionDeployedState(serviceInfo))
}

// DeleteMetrics removes status-derived Prometheus metrics for the service.
func (s *Status) DeleteMetrics() {
	metric.DeletePrometheusGauge(
		metric.LatestVersionIsDeployed,
		s.ServiceInfo.ID, "",
	)
	metric.SetUpdatesCurrent(-1, metric.GetVersionDeployedState(s.GetServiceInfo()))
}
