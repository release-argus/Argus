// Copyright [2024] [Argus]
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
	"fmt"
	"strings"
	"time"

	shoutrrr_types "github.com/release-argus/Argus/notify/shoutrrr/types"
	"github.com/release-argus/Argus/util"
)

// ServiceSummary is the Summary of a Service.
type ServiceSummary struct {
	ID                       string  `json:"id,omitempty" yaml:"id,omitempty"`
	Active                   *bool   `json:"active,omitempty" yaml:"active,omitempty"`                             // Active Service?
	Comment                  string  `json:"comment,omitempty" yaml:"comment,omitempty"`                           // Comment on the Service.
	Type                     string  `json:"type,omitempty" yaml:"type,omitempty"`                                 // "github"/"URL".
	WebURL                   string  `json:"url,omitempty" yaml:"url,omitempty"`                                   // URL to provide on the Web UI.
	Icon                     string  `json:"icon,omitempty" yaml:"icon,omitempty"`                                 // Service.Dashboard.Icon / Service.Notify.*.Params.Icon / Service.Notify.*.Defaults.Params.Icon.
	IconLinkTo               string  `json:"icon_link_to,omitempty" yaml:"icon_link_to,omitempty"`                 // URL to redirect Icon clicks to.
	HasDeployedVersionLookup *bool   `json:"has_deployed_version,omitempty" yaml:"has_deployed_version,omitempty"` // Whether this service has a DeployedVersionLookup.
	Command                  int     `json:"command,omitempty" yaml:"command,omitempty"`                           // Amount of Commands to send on a new release.
	WebHook                  int     `json:"webhook,omitempty" yaml:"webhook,omitempty"`                           // Amount of WebHooks to send on a new release.
	Status                   *Status `json:"status,omitempty" yaml:"status,omitempty"`                             // Track the Status of this source (version and regex misses).
}

// String returns a JSON string representation of the ServiceSummary.
func (s *ServiceSummary) String() string {
	if s == nil {
		return ""
	}
	return util.ToJSONString(s)
}

// RemoveUnchanged will nil/clear the fields that haven't changed compared to `other`.
func (s *ServiceSummary) RemoveUnchanged(other *ServiceSummary) {
	if other == nil {
		return
	}

	// ID
	if other.ID == s.ID {
		s.ID = ""
	}
	// Active
	if util.DereferenceOrNilValue(other.Active, true) ==
		util.DereferenceOrNilValue(s.Active, true) {
		s.Active = nil
	} else {
		active := util.DereferenceOrNilValue(s.Active, true)
		s.Active = &active
	}
	// Type
	if other.Type == s.Type {
		s.Type = ""
	}
	// URL
	if other.WebURL == s.WebURL {
		s.WebURL = ""
	}
	// Icon
	if other.Icon == s.Icon {
		s.Icon = ""
	}
	// IconLinkTo
	if other.IconLinkTo == s.IconLinkTo {
		s.IconLinkTo = ""
	}
	// Has DeployedVersionLookup?
	if util.DereferenceOrNilValue(other.HasDeployedVersionLookup, false) ==
		util.DereferenceOrNilValue(s.HasDeployedVersionLookup, false) {
		s.HasDeployedVersionLookup = nil
	}

	// Status
	statusSameCount := 0
	// Status.ApprovedVersion
	if other.Status.ApprovedVersion == s.Status.ApprovedVersion {
		s.Status.ApprovedVersion = ""
		statusSameCount++
	}
	// Status.DeployedVersion
	if other.Status.DeployedVersion == s.Status.DeployedVersion {
		s.Status.DeployedVersion = ""
		s.Status.DeployedVersionTimestamp = ""
		statusSameCount++
	}
	// Status.LatestVersion
	if other.Status.LatestVersion == s.Status.LatestVersion {
		s.Status.LatestVersion = ""
		s.Status.LatestVersionTimestamp = ""
		statusSameCount++
	}
	// nil Status if all fields match.
	if statusSameCount == 3 {
		s.Status = nil
	}
}

// Status is the Status of a Service.
type Status struct {
	ApprovedVersion          string `json:"approved_version,omitempty" yaml:"approved_version,omitempty"`                     // The approved version.
	DeployedVersion          string `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"`                     // Track the deployed version of the service from the last successful WebHook.
	DeployedVersionTimestamp string `json:"deployed_version_timestamp,omitempty" yaml:"deployed_version_timestamp,omitempty"` // UTC timestamp that the deployed version changed.
	LatestVersion            string `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`                         // Latest version found from query().
	LatestVersionTimestamp   string `json:"latest_version_timestamp,omitempty" yaml:"latest_version_timestamp,omitempty"`     // UTC timestamp that the latest version changed.
	LastQueried              string `json:"last_queried,omitempty" yaml:"last_queried,omitempty"`                             // UTC timestamp of the last query.
	RegexMissesContent       uint   `json:"regex_misses_content,omitempty" yaml:"regex_misses_content,omitempty"`             // Counter for the amount of regex misses on URL content.
	RegexMissesVersion       uint   `json:"regex_misses_version,omitempty" yaml:"regex_misses_version,omitempty"`             // Counter for the amount of regex misses on version.
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
	Notify  *[]bool `json:"notify,omitempty" yaml:"notify,omitempty"`   // Track whether any of the Slice failed.
	WebHook *[]bool `json:"webhook,omitempty" yaml:"webhook,omitempty"` // Track whether any of the WebHookSlice failed.
}

// StatusFailsSummary is the overall fail status of notifiers/webhooks.
type StatusFailsSummary struct {
	Notify  *bool `json:"notify,omitempty" yaml:"notify,omitempty"`   // Track whether any of the Slice failed.
	WebHook *bool `json:"webhook,omitempty" yaml:"webhook,omitempty"` // Track whether any of the WebHookSlice failed.
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

// Flags defines the runtime flags.
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
	Notify  NotifySlice     `json:"notify,omitempty" yaml:"notify,omitempty"`
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
		builder.WriteString(fmt.Sprintf(`"service":%s`, serviceStr))
	}

	if notifyStr := util.ToJSONString(d.Notify); !util.Contains([]string{"{}", "null"}, notifyStr) {
		if builder.Len() > 1 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(`"notify":%s`, notifyStr))
	}

	if webHookStr := util.ToJSONString(d.WebHook); !util.Contains([]string{"{}", "null"}, webHookStr) {
		if builder.Len() > 1 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(`"webhook":%s`, webHookStr))
	}
	builder.WriteString("}")

	return builder.String()
}

// Config defines the config for Argus.
type Config struct {
	File         string        `json:"-" yaml:"-"`                                             // Path to the config file (-config.file='').
	Settings     *Settings     `json:"settings,omitempty" yaml:"settings,omitempty"`           // Settings for the program.
	HardDefaults *Defaults     `json:"hard_defaults,omitempty" yaml:"hard_defaults,omitempty"` // Hard default values.
	Defaults     *Defaults     `json:"defaults,omitempty" yaml:"defaults,omitempty"`           // Default values.
	Notify       *NotifySlice  `json:"notify,omitempty" yaml:"notify,omitempty"`               // Notify message(s) to send on a new release.
	WebHook      *WebHookSlice `json:"webhook,omitempty" yaml:"webhook,omitempty"`             // WebHook(s) to send on a new release.
	Service      *ServiceSlice `json:"service,omitempty" yaml:"service,omitempty"`             // The service(s) to monitor.
	Order        []string      `json:"order,omitempty" yaml:"order,omitempty"`                 // Ordering for the Service(s) in the WebUI.
}

// Settings contains settings for the program.
type Settings struct {
	Log LogSettings `json:"log,omitempty" yaml:"log,omitempty"`
	Web WebSettings `json:"web,omitempty" yaml:"web,omitempty"`
}

// LogSettings contains web settings for the program.
type LogSettings struct {
	Timestamps *bool  `json:"timestamps,omitempty" yaml:"timestamps,omitempty"` // Timestamps in CLI output.
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

// NotifySlice is a slice map of Notify.
type NotifySlice map[string]*Notify

// String returns a JSON string representation of the NotifySlice.
func (n *NotifySlice) String() string {
	if n == nil {
		return ""
	}
	return util.ToJSONString(n)
}

// Flatten this NotifySlice into a list.
func (n *NotifySlice) Flatten() []Notify {
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

// Censor this NotifySlice for sending externally.
func (n *NotifySlice) Censor() *NotifySlice {
	if n == nil {
		return nil
	}

	slice := make(NotifySlice, len(*n))
	for i, notify := range *n {
		slice[i] = notify.Censor()
	}
	return &slice
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

	// url_fields
	n.URLFields = util.CopyMap(n.URLFields)
	for _, field := range shoutrrr_types.CensorableURLFields {
		if n.URLFields[field] != "" {
			n.URLFields[field] = util.SecretValue
		}
	}

	// params
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

// ServiceSlice is a slice map of Service.
type ServiceSlice map[string]*Service

// Service defines a software source to track and where/what to notify.
type Service struct {
	Comment               string                 `json:"comment,omitempty" yaml:"comment,omitempty"`                   // Comment on the Service.
	Options               *ServiceOptions        `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         *LatestVersion         `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Latest version lookup for the Service.
	Command               *CommandSlice          `json:"command,omitempty" yaml:"command,omitempty"`                   // OS Commands to run on new release.
	Notify                *NotifySlice           `json:"notify,omitempty" yaml:"notify,omitempty"`                     // Service-specific Notify vars.
	WebHook               *WebHookSlice          `json:"webhook,omitempty" yaml:"webhook,omitempty"`                   // Service-specific WebHook vars.
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
	Command               CommandSlice           `json:"command,omitempty" yaml:"command,omitempty"`                   // OS Commands to run on new release.
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
	AutoApprove *bool  `json:"auto_approve,omitempty" yaml:"auto_approve,omitempty"` // Default - true = Require approval before actioning new releases.
	Icon        string `json:"icon,omitempty" yaml:"icon,omitempty"`                 // Icon URL to use for messages/Web UI.
	IconLinkTo  string `json:"icon_link_to,omitempty" yaml:"icon_link_to,omitempty"` // URL to redirect Icon clicks to.
	WebURL      string `json:"web_url,omitempty" yaml:"web_url,omitempty"`           // URL to provide on the Web UI.
}

// LatestVersion lookup of the service.
type LatestVersion struct {
	Type              string                `json:"type,omitempty" yaml:"type,omitempty"`                               // Service Type, github/url.
	URL               string                `json:"url,omitempty" yaml:"url,omitempty"`                                 // URL to query.
	AccessToken       string                `json:"access_token,omitempty" yaml:"access_token,omitempty"`               // GitHub access token to use.
	AllowInvalidCerts *bool                 `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	UsePreRelease     *bool                 `json:"use_prerelease,omitempty" yaml:"use_prerelease,omitempty"`           // Whether to use GitHub prereleases.
	URLCommands       *URLCommandSlice      `json:"url_commands,omitempty" yaml:"url_commands,omitempty"`               // Commands to filter the release from the URL request.
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
	Type              string                        `json:"type,omitempty" yaml:"type,omitempty"`                               // Service Type, github/url.
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

// RequireDockerCheckRegistryDefaults are a default value(s) for a RequireDockerCheckRegistry.
type RequireDockerCheckRegistryDefaults struct {
	Token string `json:"token,omitempty" yaml:"token,omitempty"` // Token to get the token for the queries.
}

// RequireDockerCheckRegistryDefaultsWithUsername is a RequireDockerCheckRegistryDefaults with a Username.
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
	Method            string                 `json:"method,omitempty" yaml:"method,omitempty"`                           // HTTP method.
	URL               string                 `json:"url,omitempty" yaml:"url,omitempty"`                                 // URL to query.
	AllowInvalidCerts *bool                  `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
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

// BasicAuth to use on the HTTP(s) request.
type BasicAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `json:"key" yaml:"key"`     // Header key, e.g. X-Sig.
	Value string `json:"value" yaml:"value"` // Value to give the key.
}

// URLCommandSlice is a slice of URLCommand to filter version(s) from the URL Content.
type URLCommandSlice []URLCommand

// String returns a string representation of the URLCommandSlice.
func (slice *URLCommandSlice) String() string {
	if slice == nil {
		return ""
	}
	return util.ToJSONString(slice)
}

// URLCommand is a command to run to filter version(s) from the URL body.
type URLCommand struct {
	Type     string  `json:"type,omitempty" yaml:"type,omitempty"`         // regex/replace/split.
	Regex    string  `json:"regex,omitempty" yaml:"regex,omitempty"`       // regex: regexp.MustCompile(Regex).
	Index    *int    `json:"index,omitempty" yaml:"index,omitempty"`       // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index].
	Template string  `yaml:"template,omitempty" json:"template,omitempty"` // regex: template.
	Text     string  `json:"text,omitempty" yaml:"text,omitempty"`         // split:       strings.Split(tgtString, "Text").
	New      *string `json:"new,omitempty" yaml:"new,omitempty"`           // replace:     strings.ReplaceAll(tgtString, "Old", "New").
	Old      string  `json:"old,omitempty" yaml:"old,omitempty"`           // replace:     strings.ReplaceAll(tgtString, "Old", "New").
}

// Command is a command to run.
type Command []string

// CommandSlice is a slice of Command.
type CommandSlice []Command

// WebHookSlice is a slice map of WebHook.
type WebHookSlice map[string]*WebHook

// String returns a string representation of the WebHookSlice.
func (slice *WebHookSlice) String() string {
	if slice == nil {
		return ""
	}
	return util.ToJSONString(slice)
}

// Flatten the WebHookSlice into a list.
func (slice *WebHookSlice) Flatten() []*WebHook {
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
	MaxTries          *uint8    `json:"max_tries,omitempty" yaml:"max_tries,omitempty"`                     // Amount of times to send the WebHook until we receive the desired status code.
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

	// Secret
	if w.Secret != "" {
		w.Secret = util.SecretValue
	}

	// Headers
	if w.CustomHeaders != nil {
		for i := range *w.CustomHeaders {
			(*w.CustomHeaders)[i].Value = util.SecretValue
		}
	}
}

// Notifiers represent the notifiers to use when a WebHook fails.
type Notifiers struct {
	Notify *NotifySlice
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

// CommandEdit for JSON react-hook-form
type CommandEdit struct {
	Arg string `json:"arg,omitempty" yaml:"arg,omitempty"` // Command argument.
}

// ServiceEdit is a Service in API format.
type ServiceEdit struct {
	Comment               string                 `json:"comment,omitempty" yaml:"comment,omitempty"`                   // Comment on the Service.
	Options               *ServiceOptions        `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         *LatestVersion         `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Latest version lookup for the Service.
	Command               *CommandSlice          `json:"command,omitempty" yaml:"command,omitempty"`                   // OS Commands to run on new release.
	Notify                []Notify               `json:"notify,omitempty" yaml:"notify,omitempty"`                     // Service-specific Notify vars.
	WebHook               []*WebHook             `json:"webhook,omitempty" yaml:"webhook,omitempty"`                   // Service-specific WebHook vars.
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version.
	Dashboard             *DashboardOptions      `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`               // Dashboard options.
	Status                *Status                `json:"status,omitempty" yaml:"status,omitempty"`                     // Track the Status of this source (version and regex misses).
}
