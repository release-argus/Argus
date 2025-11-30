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

// Package types provides the types for the Argus API.
package types

import (
	"strings"
	"time"

	shoutrrr_types "github.com/release-argus/Argus/notify/shoutrrr/types"
	"github.com/release-argus/Argus/util"
)

// ServiceSummary is the Summary of a Service.
type ServiceSummary struct {
	ID                       string    `json:"id,omitempty" yaml:"id,omitempty"`
	Name                     *string   `json:"name,omitempty" yaml:"name,omitempty"`                                 // Name for this Service.
	Active                   *bool     `json:"active,omitempty" yaml:"active,omitempty"`                             // Active Service?
	Comment                  string    `json:"comment,omitempty" yaml:"comment,omitempty"`                           // Comment on the Service.
	Type                     string    `json:"type,omitempty" yaml:"type,omitempty"`                                 // "github"/"URL".
	WebURL                   *string   `json:"url,omitempty" yaml:"url,omitempty"`                                   // URL to provide on the Web UI.
	Icon                     *string   `json:"icon,omitempty" yaml:"icon,omitempty"`                                 // Service.Dashboard.Icon / Service.Notify.*.Params.Icon / Service.Notify.*.Defaults.Params.Icon.
	IconLinkTo               *string   `json:"icon_link_to,omitempty" yaml:"icon_link_to,omitempty"`                 // URL to redirect Icon clicks to.
	HasDeployedVersionLookup *bool     `json:"has_deployed_version,omitempty" yaml:"has_deployed_version,omitempty"` // Whether this service has a DeployedVersionLookup.
	Command                  *int      `json:"command,omitempty" yaml:"command,omitempty"`                           // Amount of Commands to send on a new release.
	WebHook                  *int      `json:"webhook,omitempty" yaml:"webhook,omitempty"`                           // Amount of WebHooks to send on a new release.
	Status                   *Status   `json:"status,omitempty" yaml:"status,omitempty"`                             // Track the Status of this source (version and regex misses).
	Tags                     *[]string `json:"tags,omitempty" yaml:"tags,omitempty"`                                 // Tags for the Service.
}

// String returns a JSON string representation of the ServiceSummary.
func (s *ServiceSummary) String() string {
	if s == nil {
		return ""
	}
	return util.ToJSONString(s)
}

// RemoveUnchanged updates a ServiceSummary by setting unchanged fields to nil or empty values based on a reference
// ServiceSummary.
func (s *ServiceSummary) RemoveUnchanged(oldData *ServiceSummary) {
	if oldData == nil {
		return
	}

	// ID.
	if oldData.ID == s.ID {
		s.ID = ""
	}
	// Name.
	s.Name = nilIfUnchanged(oldData.Name, s.Name)
	// Active.
	if util.DereferenceOrValue(oldData.Active, true) ==
		util.DereferenceOrValue(s.Active, true) {
		s.Active = nil
	} else {
		active := util.DereferenceOrValue(s.Active, true)
		s.Active = &active
	}
	// Type.
	if oldData.Type == s.Type {
		s.Type = ""
	}
	// URL.
	s.WebURL = nilIfUnchanged(oldData.WebURL, s.WebURL)
	// Icon.
	s.Icon = nilIfUnchanged(oldData.Icon, s.Icon)
	// IconLinkTo.
	s.IconLinkTo = nilIfUnchanged(oldData.IconLinkTo, s.IconLinkTo)
	// Has DeployedVersionLookup?
	if util.DereferenceOrValue(oldData.HasDeployedVersionLookup, false) ==
		util.DereferenceOrValue(s.HasDeployedVersionLookup, false) {
		s.HasDeployedVersionLookup = nil
	}

	// Status.
	statusSameCount := 0
	// 	ApprovedVersion.
	if oldData.Status.ApprovedVersion == s.Status.ApprovedVersion {
		s.Status.ApprovedVersion = ""
		statusSameCount++
	}
	// 	DeployedVersion.
	if oldData.Status.DeployedVersion == s.Status.DeployedVersion {
		s.Status.DeployedVersion = ""
		s.Status.DeployedVersionTimestamp = ""
		statusSameCount++
	}
	// 	LatestVersion.
	if oldData.Status.LatestVersion == s.Status.LatestVersion {
		s.Status.LatestVersion = ""
		s.Status.LatestVersionTimestamp = ""
		statusSameCount++
	}
	// nil Status if all fields match.
	if statusSameCount == 3 {
		s.Status = nil
	}

	// Tags - Removed.
	if oldData.Tags != nil && s.Tags == nil {
		var emptyTags []string
		s.Tags = &emptyTags
		// Unchanged.
	} else if oldData.Tags != nil && s.Tags != nil && util.AreSlicesEqual(*oldData.Tags, *s.Tags) {
		s.Tags = nil
	}

	// Command.
	s.Command = nilIfUnchanged(oldData.Command, s.Command)
	// WebHook.
	s.WebHook = nilIfUnchanged(oldData.WebHook, s.WebHook)
}

// nilIfUnchanged compares two pointers; returns nil if values are equal, newValue if different, or a zero value if removed.
func nilIfUnchanged[T comparable](oldValue, newValue *T) *T {
	switch {
	// Changed: no old value.
	case oldValue == nil:
		return newValue
	// Removed: new value nil.
	case newValue == nil:
		var zero T
		return &zero
	// Unchanged.
	case *oldValue == *newValue:
		return nil
	}
	// Changed: values differ.
	return newValue
}

// Status is the Status of a Service.
type Status struct {
	ApprovedVersion          string `json:"approved_version,omitempty" yaml:"approved_version,omitempty"`                     // The approved version.
	DeployedVersion          string `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"`                     // Track the deployed version of the service from the last successful WebHook.
	DeployedVersionTimestamp string `json:"deployed_version_timestamp,omitempty" yaml:"deployed_version_timestamp,omitempty"` // UTC timestamp that the deployed version changed.
	LatestVersion            string `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`                         // Latest version found from query().
	LatestVersionTimestamp   string `json:"latest_version_timestamp,omitempty" yaml:"latest_version_timestamp,omitempty"`     // UTC timestamp that the latest version changed.
	LastQueried              string `json:"last_queried,omitempty" yaml:"last_queried,omitempty"`                             // UTC timestamp of the last query.
	RegexMissesContent       uint   `json:"regex_misses_content,omitempty" yaml:"regex_misses_content,omitempty"`             // Counter for the number of regular expression misses on URL content.
	RegexMissesVersion       uint   `json:"regex_misses_version,omitempty" yaml:"regex_misses_version,omitempty"`             // Counter for the number of regular expression misses on version.
}

// String returns a JSON string representation of the Status.
func (s *Status) String() string {
	if s == nil {
		return ""
	}
	return util.ToJSONString(s)
}

// StatusFails is the fail status of each notifier/webhook.
type StatusFails struct {
	Notify  *[]bool `json:"notify,omitempty" yaml:"notify,omitempty"`   // Track whether any of the Notifiers failed.
	WebHook *[]bool `json:"webhook,omitempty" yaml:"webhook,omitempty"` // Track whether any of the WebHooks failed.
}

// StatusFailsSummary is the overall fail status of notifiers/webhooks.
type StatusFailsSummary struct {
	Notify  *bool `json:"notify,omitempty" yaml:"notify,omitempty"`   // Track whether any of the Notifiers failed.
	WebHook *bool `json:"webhook,omitempty" yaml:"webhook,omitempty"` // Track whether any of the WebHooks failed.
}

// ActionSummary is the summary of all Actions for a Service.
type ActionSummary struct {
	Command map[string]CommandSummary `json:"command" yaml:"command"` // Summary of all Commands.
	WebHook map[string]WebHookSummary `json:"webhook" yaml:"webhook"` // Summary of all WebHooks.
}

// WebHookSummary is the summary of a WebHook.
type WebHookSummary struct {
	Failed       *bool     `json:"failed,omitempty" yaml:"failed,omitempty"`               // Whether this WebHook failed to send successfully for the LatestVersion.
	NextRunnable time.Time `json:"next_runnable,omitempty" yaml:"next_runnable,omitempty"` // Time the WebHook can next run (for staggering).
}

// BuildInfo is information from build time.
type BuildInfo struct {
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
	BuildDate string `json:"build_date,omitempty" yaml:"build_date,omitempty"`
	GoVersion string `json:"go_version,omitempty" yaml:"go_version,omitempty"`
}

// RuntimeInfo defines current runtime information.
type RuntimeInfo struct {
	StartTime      time.Time `json:"start_time,omitempty" yaml:"start_time,omitempty"`
	CWD            string    `json:"cwd,omitempty" yaml:"cwd,omitempty"`
	GoRoutineCount int       `json:"goroutines,omitempty" yaml:"goroutines,omitempty"`
	GOMAXPROCS     int       `json:"GOMAXPROCS,omitempty" yaml:"GOMAXPROCS,omitempty"`
	GoGC           string    `json:"GOGC,omitempty" yaml:"GOGC,omitempty"`
	GoDebug        string    `json:"GODEBUG,omitempty" yaml:"GODEBUG,omitempty"`
}

// Flags define the runtime flags.
type Flags struct {
	ConfigFile       string `json:"config.file,omitempty" yaml:"config.file,omitempty"`
	LogLevel         string `json:"log.level,omitempty" yaml:"log.level,omitempty"`
	LogTimestamps    *bool  `json:"log.timestamps,omitempty" yaml:"log.timestamps,omitempty"`
	DataDatabaseFile string `json:"data.database-file,omitempty" yaml:"data.database-file,omitempty"`
	WebListenHost    string `json:"web.listen-host,omitempty" yaml:"web.listen-host,omitempty"`
	WebListenPort    string `json:"web.listen-port,omitempty" yaml:"web.listen-port,omitempty"`
	WebCertFile      string `json:"web.cert-file" yaml:"web.cert-file"`
	WebPKeyFile      string `json:"web.pkey-file" yaml:"web.pkey-file"`
	WebRoutePrefix   string `json:"web.route-prefix,omitempty" yaml:"web.route-prefix,omitempty"`
}

// Defaults are the global defaults for vars.
type Defaults struct {
	Service ServiceDefaults `json:"service,omitempty" yaml:"service,omitempty"`
	Notify  Notifiers       `json:"notify,omitempty" yaml:"notify,omitempty"`
	WebHook WebHook         `json:"webhook,omitempty" yaml:"webhook,omitempty"`
}

// String returns a JSON string representation of the Defaults.
func (d *Defaults) String() string {
	if d == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("{")

	if serviceStr := util.ToJSONString(d.Service); !util.Contains([]string{"{}", "null"}, serviceStr) {
		builder.WriteString(`"service":` + serviceStr)
	}

	if notifyStr := util.ToJSONString(d.Notify); !util.Contains([]string{"{}", "null"}, notifyStr) {
		if builder.Len() > 1 {
			builder.WriteString(",")
		}
		builder.WriteString(`"notify":` + notifyStr)
	}

	if webHookStr := util.ToJSONString(d.WebHook); !util.Contains([]string{"{}", "null"}, webHookStr) {
		if builder.Len() > 1 {
			builder.WriteString(",")
		}
		builder.WriteString(`"webhook":` + webHookStr)
	}
	builder.WriteString("}")

	return builder.String()
}

// Config defines the config for Argus.
type Config struct {
	File         string     `json:"-" yaml:"-"`                                             // Path to the config file (-config.file='').
	Settings     *Settings  `json:"settings,omitempty" yaml:"settings,omitempty"`           // Settings for the program.
	HardDefaults *Defaults  `json:"hard_defaults,omitempty" yaml:"hard_defaults,omitempty"` // Hard default values.
	Defaults     *Defaults  `json:"defaults,omitempty" yaml:"defaults,omitempty"`           // Default values.
	Notify       *Notifiers `json:"notify,omitempty" yaml:"notify,omitempty"`               // Notify messages to send on a new release.
	WebHook      *WebHooks  `json:"webhook,omitempty" yaml:"webhook,omitempty"`             // WebHooks to send on a new release.
	Service      *Services  `json:"service,omitempty" yaml:"service,omitempty"`             // The services to monitor.
	Order        []string   `json:"order,omitempty" yaml:"order,omitempty"`                 // Ordering for the Services in the WebUI.
}

// Settings contain settings for the program.
type Settings struct {
	Log LogSettings `json:"log,omitempty" yaml:"log,omitempty"`
	Web WebSettings `json:"web,omitempty" yaml:"web,omitempty"`
}

// LogSettings contains web settings for the program.
type LogSettings struct {
	Timestamps *bool  `json:"timestamps,omitempty" yaml:"timestamps,omitempty"` // Timestamps in command-line tool output.
	Level      string `json:"level,omitempty" yaml:"level,omitempty"`           // Log level.
}

// WebSettings contains web settings for the program.
type WebSettings struct {
	ListenHost  string `json:"listen_host,omitempty" yaml:"listen_host,omitempty"`   // Web listen host.
	ListenPort  string `json:"listen_port,omitempty" yaml:"listen_port,omitempty"`   // Web listen port.
	CertFile    string `json:"cert_file,omitempty" yaml:"cert_file,omitempty"`       // HTTPS certificate path.
	KeyFile     string `json:"pkey_file,omitempty" yaml:"pkey_file,omitempty"`       // HTTPS privkey path.
	RoutePrefix string `json:"route_prefix,omitempty" yaml:"route_prefix,omitempty"` // Web endpoint prefix.
}

// Notifiers is a string map of Notify.
type Notifiers map[string]*Notify

// String returns a JSON string representation of the Notifiers.
func (n *Notifiers) String() string {
	if n == nil {
		return ""
	}
	return util.ToJSONString(n)
}

// Flatten these Notifiers into a list.
func (n *Notifiers) Flatten() []Notify {
	if n == nil {
		return nil
	}

	names := util.SortedKeys(*n)
	list := make([]Notify, len(names))

	for index, name := range names {
		element := (*n)[name]
		// Add to list.
		list[index] = Notify{
			ID:        name,
			Type:      element.Type,
			Options:   element.Options,
			URLFields: element.URLFields,
			Params:    element.Params,
		}
		list[index].Censor()
	}

	return list
}

// Censor these Notifiers for sending externally.
func (n *Notifiers) Censor() *Notifiers {
	if n == nil {
		return nil
	}

	notifiers := make(Notifiers, len(*n))
	for i, notify := range *n {
		notifiers[i] = notify.Censor()
	}
	return &notifiers
}

// Notify is a message notifier source.
type Notify struct {
	ID        string            `json:"name,omitempty" yaml:"name,omitempty"`             // ID for this Notify sender.
	Type      string            `json:"type,omitempty" yaml:"type,omitempty"`             // Notification type, e.g. slack.
	Options   map[string]string `json:"options,omitempty" yaml:"options,omitempty"`       // Options.
	URLFields map[string]string `json:"url_fields,omitempty" yaml:"url_fields,omitempty"` // URL Fields.
	Params    map[string]string `json:"params,omitempty" yaml:"params,omitempty"`         // Param props.
}

// Censor this Notify for sending over a WebSocket.
func (n *Notify) Censor() *Notify {
	if n == nil {
		return nil
	}

	// url_fields.
	n.URLFields = util.CopyMap(n.URLFields)
	for _, field := range shoutrrr_types.CensorableURLFields {
		if n.URLFields[field] != "" {
			n.URLFields[field] = util.SecretValue
		}
	}

	// params.
	n.Params = util.CopyMap(n.Params)
	for _, param := range shoutrrr_types.CensorableParams {
		if n.Params[param] != "" {
			n.Params[param] = util.SecretValue
		}
	}

	return n
}

// NotifyParams is a map of Notify parameters.
type NotifyParams map[string]string

// Services is a string map of Service.
type Services map[string]*Service

// Service defines a software source to track and where/what to notify.
type Service struct {
	Name                  string                 `json:"name,omitempty" yaml:"name,omitempty"`                         // Name for this Service.
	Comment               string                 `json:"comment,omitempty" yaml:"comment,omitempty"`                   // Comment on the Service.
	Options               *ServiceOptions        `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         *LatestVersion         `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Latest version lookup for the Service.
	Command               *Commands              `json:"command,omitempty" yaml:"command,omitempty"`                   // OS Commands to run on new release.
	Notify                *Notifiers             `json:"notify,omitempty" yaml:"notify,omitempty"`                     // Service-specific Notify vars.
	WebHook               *WebHooks              `json:"webhook,omitempty" yaml:"webhook,omitempty"`                   // Service-specific WebHook vars.
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version.
	Dashboard             *DashboardOptions      `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`               // Dashboard options.
	Status                *Status                `json:"status,omitempty" yaml:"status,omitempty"`                     // Track the Status of this source (version and regex misses).
}

// String returns a string representation of the Service.
func (r *Service) String() string {
	if r == nil {
		return ""
	}
	return util.ToJSONString(r)
}

// ServiceDefaults defines default values for a Service.
type ServiceDefaults struct {
	Comment               string                 `json:"comment,omitempty" yaml:"comment,omitempty"`                   // Comment on the Service.
	Options               *ServiceOptions        `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         *LatestVersionDefaults `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Latest version lookup for the Service.
	Notify                map[string]struct{}    `json:"notify,omitempty" yaml:"notify,omitempty"`                     // Service-specific Notify vars.
	Command               *Commands              `json:"command,omitempty" yaml:"command,omitempty"`                   // OS Commands to run on new release.
	WebHook               map[string]struct{}    `json:"webhook,omitempty" yaml:"webhook,omitempty"`                   // Service-specific WebHook vars.
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version.
	Dashboard             *DashboardOptions      `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`               // Dashboard options.
	Status                *Status                `json:"status,omitempty" yaml:"status,omitempty"`
}

// ServiceOptions defines configuration options for a service.
type ServiceOptions struct {
	Active             *bool  `json:"active,omitempty" yaml:"active,omitempty"`                           // Active Service?.
	Interval           string `json:"interval,omitempty" yaml:"interval,omitempty"`                       // AhBmCs = Sleep A hours, B minutes and C seconds between queries.
	SemanticVersioning *bool  `json:"semantic_versioning,omitempty" yaml:"semantic_versioning,omitempty"` // Default - true = Version must exceed the previous version to trigger alerts/Commands/WebHooks.
}

// DashboardOptions defines configuration options for a service on the Web UI dashboard.
type DashboardOptions struct {
	AutoApprove *bool    `json:"auto_approve,omitempty" yaml:"auto_approve,omitempty"` // Default - true = Require approval before actioning new releases.
	Icon        string   `json:"icon,omitempty" yaml:"icon,omitempty"`                 // Icon URL to use for messages/Web UI.
	IconLinkTo  string   `json:"icon_link_to,omitempty" yaml:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
	WebURL      string   `json:"web_url,omitempty" yaml:"web_url,omitempty"`           // URL to provide on the Web UI.
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`                 // Tags for the Service.
}

// LatestVersion lookup of the service.
type LatestVersion struct {
	Type              string                `json:"type,omitempty" yaml:"type,omitempty"`                               // Service Type, github/url.
	URL               string                `json:"url,omitempty" yaml:"url,omitempty"`                                 // URL to query.
	AccessToken       string                `json:"access_token,omitempty" yaml:"access_token,omitempty"`               // GitHub access token to use.
	AllowInvalidCerts *bool                 `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	UsePreRelease     *bool                 `json:"use_prerelease,omitempty" yaml:"use_prerelease,omitempty"`           // Whether to use GitHub prereleases.
	URLCommands       *URLCommands          `json:"url_commands,omitempty" yaml:"url_commands,omitempty"`               // Commands to filter the release from the URL request.
	Require           *LatestVersionRequire `json:"require,omitempty" yaml:"require,omitempty"`                         // Requirements before treating a release as valid.
}

// String returns a string representation of the LatestVersion.
func (r *LatestVersion) String() string {
	if r == nil {
		return ""
	}
	return util.ToJSONString(r)
}

// LatestVersionDefaults are default values for a LatestVersion.
type LatestVersionDefaults struct {
	URL               string                        `json:"url,omitempty" yaml:"url,omitempty"`                                 // URL to query.
	AccessToken       string                        `json:"access_token,omitempty" yaml:"access_token,omitempty"`               // GitHub access token to use.
	AllowInvalidCerts *bool                         `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	UsePreRelease     *bool                         `json:"use_prerelease,omitempty" yaml:"use_prerelease,omitempty"`           // Whether to use GitHub prereleases.
	Require           *LatestVersionRequireDefaults `json:"require,omitempty" yaml:"require,omitempty"`
}

// String returns a string representation of the LatestVersionRequireDefaults.
func (r *LatestVersionRequireDefaults) String() string {
	if r == nil {
		return ""
	}

	str := util.ToJSONString(r)
	str = strings.Replace(str, `"docker":{}`, "", 1)

	return str
}

// LatestVersionRequire contains commands, regex, etc. that must pass before considering a release valid.
type LatestVersionRequire struct {
	Command      []string            `json:"command,omitempty" yaml:"command,omitempty"`             // Require Command to pass.
	Docker       *RequireDockerCheck `json:"docker,omitempty" yaml:"docker,omitempty"`               // Docker image tag requirements.
	RegexContent string              `json:"regex_content,omitempty" yaml:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions.
	RegexVersion string              `json:"regex_version,omitempty" yaml:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions/.
}

// String returns a string representation of the LatestVersionRequire.
func (r *LatestVersionRequire) String() string {
	if r == nil {
		return ""
	}
	return util.ToJSONString(r)
}

// LatestVersionRequireDefaults define the default requirements before considering a release valid.
type LatestVersionRequireDefaults struct {
	Docker RequireDockerCheckDefaults `json:"docker,omitempty" yaml:"docker,omitempty"` // Docker repo defaults.
}

// RequireDockerCheckRegistryDefaults are default values for a RequireDockerCheckRegistry.
type RequireDockerCheckRegistryDefaults struct {
	Token string `json:"token,omitempty" yaml:"token,omitempty"` // Token to get the token for the queries.
}

// RequireDockerCheckRegistryDefaultsWithUsername are RequireDockerCheckRegistryDefaults with a Username.
type RequireDockerCheckRegistryDefaultsWithUsername struct {
	RequireDockerCheckRegistryDefaults
	Username string `json:"username,omitempty" yaml:"username,omitempty"` // Username to get a new token.
}

// RequireDockerCheckDefaults are default values for a RequireDockerCheck.
type RequireDockerCheckDefaults struct {
	Type string                                          `json:"type,omitempty" yaml:"type,omitempty"` // Default DockerCheck Type.
	GHCR *RequireDockerCheckRegistryDefaults             `json:"ghcr,omitempty" yaml:"ghcr,omitempty"` // GHCR.
	Hub  *RequireDockerCheckRegistryDefaultsWithUsername `json:"hub,omitempty" yaml:"hub,omitempty"`   // DockerHub.
	Quay *RequireDockerCheckRegistryDefaults             `json:"quay,omitempty" yaml:"quay,omitempty"` // Quay.
}

// RequireDockerCheck points to a Docker repository for a release to qualify as valid.
type RequireDockerCheck struct {
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`         // Where to check, e.g. hub (DockerHub), GHCR, Quay.
	Image    string `json:"image,omitempty" yaml:"image,omitempty"`       // Image to check.
	Tag      string `json:"tag,omitempty" yaml:"tag,omitempty"`           // Tag to check for.
	Username string `json:"username,omitempty" yaml:"username,omitempty"` // Username to get a new token.
	Token    string `json:"token,omitempty" yaml:"token,omitempty"`       // Token to get the token for the queries.
}

// DeployedVersionLookup of the service.
type DeployedVersionLookup struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"` // Service Type, url/manual.

	// manual
	Version string `json:"version,omitempty" yaml:"version,omitempty"` // Deployed version.

	// url
	Method            string                 `json:"method,omitempty" yaml:"method,omitempty"`                           // HTTP method.
	URL               string                 `json:"url,omitempty" yaml:"url,omitempty"`                                 // URL to query.
	AllowInvalidCerts *bool                  `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	TargetHeader      string                 `json:"target_header,omitempty" yaml:"target_header,omitempty"`             // Header to target for the version.
	BasicAuth         *BasicAuth             `json:"basic_auth,omitempty" yaml:"basic_auth,omitempty"`                   // Basic Auth credentials.
	Headers           []Header               `json:"headers,omitempty" yaml:"headers,omitempty"`                         // Request Headers.
	Body              string                 `json:"body,omitempty" yaml:"body,omitempty"`                               // Request Body.
	JSON              string                 `json:"json,omitempty" yaml:"json,omitempty"`                               // JSON key to use e.g. version_current.
	Regex             string                 `json:"regex,omitempty" yaml:"regex,omitempty"`                             // Regex for the version.
	RegexTemplate     string                 `json:"regex_template,omitempty" yaml:"regex_template,omitempty"`           // Template to apply to the RegEx match.
	HardDefaults      *DeployedVersionLookup `json:"-" yaml:"-"`                                                         // Hardcoded default values.
	Defaults          *DeployedVersionLookup `json:"-" yaml:"-"`                                                         // Default values.
}

// String returns a JSON string representation of the DeployedVersionLookup.
func (d *DeployedVersionLookup) String() string {
	if d == nil {
		return ""
	}
	return util.ToJSONString(d)
}

// BasicAuth to use on the HTTP(S) request.
type BasicAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `json:"key" yaml:"key"`     // Header key, e.g. X-Sig.
	Value string `json:"value" yaml:"value"` // Value to give the key.
}

// URLCommands is a slice of URLCommand to filter versions from the URL Content.
type URLCommands []URLCommand

// String returns a string representation of the URLCommands.
func (slice *URLCommands) String() string {
	if slice == nil {
		return ""
	}
	return util.ToJSONString(slice)
}

// URLCommand is a command to run to filter versions from the URL body.
type URLCommand struct {
	Type     string  `json:"type,omitempty" yaml:"type,omitempty"`         // regex/replace/split.
	Regex    string  `json:"regex,omitempty" yaml:"regex,omitempty"`       // regex: regexp.MustCompile(Regex).
	Index    *int    `json:"index,omitempty" yaml:"index,omitempty"`       // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index].
	Template string  `json:"template,omitempty" yaml:"template,omitempty"` // regex: template.
	Text     string  `json:"text,omitempty" yaml:"text,omitempty"`         // split:       strings.Split(tgtString, "Text").
	New      *string `json:"new,omitempty" yaml:"new,omitempty"`           // replace:     strings.ReplaceAll(tgtString, "Old", "New").
	Old      string  `json:"old,omitempty" yaml:"old,omitempty"`           // replace:     strings.ReplaceAll(tgtString, "Old", "New").
}

// Command is a command to run.
type Command []string

// Commands is a slice of Command.
type Commands []Command

// WebHooks is a slice map of WebHook.
type WebHooks map[string]*WebHook

// String returns a string representation of the WebHooks.
func (slice *WebHooks) String() string {
	if slice == nil {
		return ""
	}
	return util.ToJSONString(slice)
}

// Flatten these WebHooks into a list.
func (slice *WebHooks) Flatten() []*WebHook {
	if slice == nil {
		return nil
	}

	names := util.SortedKeys(*slice)
	list := make([]*WebHook, len(names))

	for index, name := range names {
		list[index] = (*slice)[name]
		list[index].Censor()
		list[index].ID = name
	}

	return list
}

// WebHook is a WebHook to send.
type WebHook struct {
	ServiceID         string    `json:"-" yaml:"-"`                                                         // ID of the service this WebHook belongs to.
	ID                string    `json:"name,omitempty" yaml:"name,omitempty"`                               // Name of this WebHook.
	Type              string    `json:"type,omitempty" yaml:"type,omitempty"`                               // "github"/"url".
	URL               string    `json:"url,omitempty" yaml:"url,omitempty"`                                 // "https://example.com".
	AllowInvalidCerts *bool     `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	Secret            string    `json:"secret,omitempty" yaml:"secret,omitempty"`                           // "SECRET".
	CustomHeaders     *[]Header `json:"custom_headers,omitempty" yaml:"custom_headers,omitempty"`           // Custom Headers for the WebHook.
	DesiredStatusCode *uint16   `json:"desired_status_code,omitempty" yaml:"desired_status_code,omitempty"` // e.g. 202.
	Delay             string    `json:"delay,omitempty" yaml:"delay,omitempty"`                             // The delay before sending the WebHook.
	MaxTries          *uint8    `json:"max_tries,omitempty" yaml:"max_tries,omitempty"`                     // Number of times to send the WebHook until we receive the desired status code.
	SilentFails       *bool     `json:"silent_fails,omitempty" yaml:"silent_fails,omitempty"`               // Whether to notify if this WebHook fails MaxTries times.
}

// String returns a string representation of the WebHook.
func (w *WebHook) String() string {
	if w == nil {
		return ""
	}
	return util.ToJSONString(w)
}

// Censor this WebHook for sending to the web client.
func (w *WebHook) Censor() {
	if w == nil {
		return
	}

	// Secret.
	if w.Secret != "" {
		w.Secret = util.SecretValue
	}

	// Headers.
	if w.CustomHeaders != nil {
		for i := range *w.CustomHeaders {
			(*w.CustomHeaders)[i].Value = util.SecretValue
		}
	}
}

// CommandSummary holds the summary of a Command.
type CommandSummary struct {
	Failed       *bool     `json:"failed,omitempty" yaml:"failed,omitempty"`               // Whether the last run failed.
	NextRunnable time.Time `json:"next_runnable,omitempty" yaml:"next_runnable,omitempty"` // Time at which the Command can next run (for staggering).
}

// CommandStatusUpdate holds an update of the current state of the Command.
// @ index.
type CommandStatusUpdate struct {
	Command string `json:"command" yaml:"command"` // Index of the Command.
	Failed  bool   `json:"failed" yaml:"failed"`   // Whether the last attempt of this command failed.
}

// ################################################################################
// #                                     EDIT                                     #
// ################################################################################

// CommandEdit for JSON react-hook-form.
type CommandEdit struct {
	Arg string `json:"arg,omitempty" yaml:"arg,omitempty"` // Command argument.
}

// ServiceEdit is a Service in API format.
type ServiceEdit struct {
	Name                  string                 `json:"name,omitempty" yaml:"name,omitempty"`                         // Name of the Service.
	Comment               string                 `json:"comment,omitempty" yaml:"comment,omitempty"`                   // Comment on the Service.
	Options               *ServiceOptions        `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         *LatestVersion         `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Latest version lookup for the Service.
	Command               *Commands              `json:"command,omitempty" yaml:"command,omitempty"`                   // OS Commands to run on new release.
	Notify                []Notify               `json:"notify,omitempty" yaml:"notify,omitempty"`                     // Service-specific Notify vars.
	WebHook               []*WebHook             `json:"webhook,omitempty" yaml:"webhook,omitempty"`                   // Service-specific WebHook vars.
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version.
	Dashboard             *DashboardOptions      `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`               // Dashboard options.
	Status                *Status                `json:"status,omitempty" yaml:"status,omitempty"`                     // Track the Status of this source (version and regex misses).
}
