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

package apitype

import (
	"time"

	shoutrrr_types "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/util"
)

// ServiceSummary is the Summary of a Service.
type ServiceSummary struct {
	ID                       string  `json:"id,omitempty"`                   //
	Active                   *bool   `json:"active,omitempty"`               // Active Service?
	Comment                  *string `json:"comment,omitempty"`              // Comment on the Service
	Type                     *string `json:"type,omitempty"`                 // "github"/"URL"
	WebURL                   string  `json:"url,omitempty"`                  // URL to provide on the Web UI
	Icon                     *string `json:"icon,omitempty"`                 // Service.Dashboard.Icon / Service.Notify.*.Params.Icon / Service.Notify.*.Defaults.Params.Icon
	IconLinkTo               *string `json:"icon_link_to,omitempty"`         // URL to redirect Icon clicks to
	HasDeployedVersionLookup *bool   `json:"has_deployed_version,omitempty"` // Whether this service has a DeployedVersionLookup
	Command                  *int    `json:"command,omitempty"`              // Number of Commands to send on a new release
	WebHook                  *int    `json:"webhook,omitempty"`              // Number of WebHooks to send on a new release
	Status                   *Status `json:"status,omitempty"`               // Track the Status of this source (version and regex misses)
}

// String returns a JSON string representation of the ServiceSummary.
func (s *ServiceSummary) String() (str string) {
	if s != nil {
		str = util.ToJSONString(s)
	}
	return
}

// RemoveUnchanged will nil/empty out the fields that haven't changed compared to `other`.
func (s *ServiceSummary) RemoveUnchanged(other *ServiceSummary) {
	if other == nil {
		return
	}

	// ID
	if other.ID == s.ID {
		s.ID = ""
	}
	// Active
	if util.EvalNilPtr(other.Active, true) == util.EvalNilPtr(s.Active, true) {
		s.Active = nil
	} else {
		active := util.EvalNilPtr(s.Active, true)
		s.Active = &active
	}
	// Type
	if util.DefaultIfNil(other.Type) == util.DefaultIfNil(s.Type) {
		s.Type = nil
	}
	// Icon
	if util.DefaultIfNil(other.Icon) == util.DefaultIfNil(s.Icon) {
		s.Icon = nil
	}
	// IconLinkTo
	if util.DefaultIfNil(other.IconLinkTo) == util.DefaultIfNil(s.IconLinkTo) {
		s.IconLinkTo = nil
	}
	// HasDeployedVersionLookup
	if util.EvalNilPtr(other.HasDeployedVersionLookup, false) == util.EvalNilPtr(s.HasDeployedVersionLookup, false) {
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
	// nil Status if all fields are the same
	if statusSameCount == 3 {
		s.Status = nil
	}
}

// Status is the Status of a Service.
type Status struct {
	ApprovedVersion          string `json:"approved_version,omitempty"`           // The version that's been approved
	DeployedVersion          string `json:"deployed_version,omitempty"`           // Track the deployed version of the service from the last successful WebHook
	DeployedVersionTimestamp string `json:"deployed_version_timestamp,omitempty"` // UTC timestamp that the deployed version change was noticed
	LatestVersion            string `json:"latest_version,omitempty"`             // Latest version found from query()
	LatestVersionTimestamp   string `json:"latest_version_timestamp,omitempty"`   // UTC timestamp that the latest version change was noticed
	LastQueried              string `json:"last_queried,omitempty"`               // UTC timestamp that version was last queried/checked
	RegexMissesContent       uint   `json:"regex_misses_content,omitempty"`       // Counter for the number of regex misses on URL content
	RegexMissesVersion       uint   `json:"regex_misses_version,omitempty"`       // Counter for the number of regex misses on version
}

// String returns a JSON string representation of the Status.
func (s *Status) String() (str string) {
	if s != nil {
		str = util.ToJSONString(s)
	}
	return
}

// StatusFails keeps track of whether each of the notifications failed on the last version change.
type StatusFails struct {
	Notify  *[]bool `json:"notify,omitempty"`  // Track whether any of the Slice failed
	WebHook *[]bool `json:"webhook,omitempty"` // Track whether any of the WebHookSlice failed
}

// StatusFailsSummary keeps track of whether any of the notifications failed on the last version change.
type StatusFailsSummary struct {
	Notify  *bool `json:"notify,omitempty"`  // Track whether any of the Slice failed
	WebHook *bool `json:"webhook,omitempty"` // Track whether any of the WebHookSlice failed
}

// WebHookSummary is the summary of a WebHook.
type WebHookSummary struct {
	Failed       *bool     `json:"failed,omitempty"`        // Whether this WebHook failed to send successfully for the LatestVersion
	NextRunnable time.Time `json:"next_runnable,omitempty"` // Time the WebHook can next be run (for staggering)
}

// Info is runtime and build information.
type Info struct {
	Build   BuildInfo   `json:"build,omitempty"`
	Runtime RuntimeInfo `json:"runtime,omitempty"`
}

// BuildInfo is information from build time.
type BuildInfo struct {
	Version   string `json:"version,omitempty"`
	BuildDate string `json:"build_date,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
}

// RuntimeInfo is current runtime information.
type RuntimeInfo struct {
	StartTime      time.Time `json:"start_time,omitempty"`
	CWD            string    `json:"cwd,omitempty"`
	GoRoutineCount int       `json:"goroutines,omitempty"`
	GOMAXPROCS     int       `json:"GOMAXPROCS,omitempty"`
	GoGC           string    `json:"GOGC,omitempty"`
	GoDebug        string    `json:"GODEBUG,omitempty"`
}

// Flags is the runtime flags.
type Flags struct {
	ConfigFile       *string `json:"config.file,omitempty"`
	LogLevel         string  `json:"log.level,omitempty"`
	LogTimestamps    *bool   `json:"log.timestamps,omitempty"`
	DataDatabaseFile *string `json:"data.database-file,omitempty"`
	WebListenHost    string  `json:"web.listen-host,omitempty"`
	WebListenPort    string  `json:"web.listen-port,omitempty"`
	WebCertFile      *string `json:"web.cert-file"`
	WebPKeyFile      *string `json:"web.pkey-file"`
	WebRoutePrefix   string  `json:"web.route-prefix,omitempty"`
}

// Defaults are the global defaults for vars.
type Defaults struct {
	Service ServiceDefaults `json:"service,omitempty"`
	Notify  NotifySlice     `json:"notify,omitempty"`
	WebHook WebHook         `json:"webhook,omitempty"`
}

// String returns a JSON string representation of the Defaults.
func (d *Defaults) String() (str string) {
	if d != nil {
		str = util.ToJSONString(d)
	}
	return
}

// Config is the config for Argus.
type Config struct {
	File         string        `json:"-"`                       // Path to the config file (-config.file='')
	HardDefaults *Defaults     `json:"hard_defaults,omitempty"` // Hard default values
	Defaults     *Defaults     `json:"defaults,omitempty"`      // Default values
	Service      *ServiceSlice `json:"service,omitempty"`       // The service(s) to monitor
	Notify       *NotifySlice  `json:"notify,omitempty"`        // Notify message(s) to send on a new release
	WebHook      *WebHookSlice `json:"webhook,omitempty"`       // WebHook(s) to send on a new release
	Settings     *Settings     `json:"settings,omitempty"`      // Settings for the program
	Order        []string      `json:"order,omitempty"`         // Ordering for the Service(s) in the WebUI
}

// Settings contains settings for the program.
type Settings struct {
	Log LogSettings `json:"log,omitempty"`
	Web WebSettings `json:"web,omitempty"`
}

// LogSettings contains web settings for the program.
type LogSettings struct {
	Timestamps *bool   `json:"timestamps,omitempty"` // Timestamps in CLI output
	Level      *string `json:"level,omitempty"`      // Log level
}

// WebSettings contains web settings for the program.
type WebSettings struct {
	ListenHost  *string `json:"listen_host,omitempty"`  // Web listen host
	ListenPort  *string `json:"listen_port,omitempty"`  // Web listen port
	CertFile    *string `json:"cert_file,omitempty"`    // HTTPS certificate path
	KeyFile     *string `json:"pkey_file,omitempty"`    // HTTPS privkey path
	RoutePrefix *string `json:"route_prefix,omitempty"` // Web endpoint prefix
}

// NotifySlice is a slice mapping of Notify.
type NotifySlice map[string]*Notify

// String returns a JSON string representation of the NotifySlice.
func (slice *NotifySlice) String() (str string) {
	if slice != nil {
		str = util.ToJSONString(slice)
	}
	return
}

// Flatten this NotifySlice into a list
func (slice *NotifySlice) Flatten() *[]Notify {
	if slice == nil {
		return nil
	}

	names := util.SortedKeys(*slice)
	list := make([]Notify, len(names))

	for index, name := range names {
		// Add to list
		list[index] = Notify{
			ID:        name,
			Type:      (*slice)[name].Type,
			Options:   (*slice)[name].Options,
			URLFields: (*slice)[name].URLFields,
			Params:    (*slice)[name].Params,
		}
		list[index].Censor()
	}

	return &list
}

// Censor this NotifySlice for sending externally
func (n *NotifySlice) Censor() *NotifySlice {
	if n == nil {
		return nil
	}

	slice := make(NotifySlice, len(*n))
	for i := range *n {
		slice[i] = (*n)[i].Censor()
	}
	return &slice
}

// Notify is a message w/ destination and from details.
type Notify struct {
	ID        string                `json:"name,omitempty"`       // ID for this Notify sender
	Type      string                `json:"type,omitempty"`       // Notification type, e.g. slack
	Options   map[string]string     `json:"options,omitempty"`    // Options
	URLFields map[string]string     `json:"url_fields,omitempty"` // URL Fields
	Params    shoutrrr_types.Params `json:"params,omitempty"`     // Param props
}

// Censor this Notify for sending over a WebSocket
func (n *Notify) Censor() *Notify {
	if n == nil {
		return nil
	}

	// url_fields
	urlFieldsToCensor := []string{
		"altid",
		"apikey",
		"botkey",
		"password",
		"token",
		"tokena",
		"tokenb",
	}
	n.URLFields = util.CopyMap(n.URLFields)
	for _, field := range urlFieldsToCensor {
		if n.URLFields[field] != "" {
			n.URLFields[field] = "<secret>"
		}
	}

	// params
	paramsToCensor := []string{
		"devices",
	}
	n.Params = util.CopyMap(n.Params)
	for _, param := range paramsToCensor {
		if n.Params[param] != "" {
			n.Params[param] = "<secret>"
		}
	}

	return n
}

type NotifyParams map[string]string

// ServiceSlice is a slice mapping of Service.
type ServiceSlice map[string]*Service

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	Comment               string                 `json:"comment,omitempty"`          // Comment on the Service
	Options               *ServiceOptions        `json:"options,omitempty"`          // Options to give the Service
	LatestVersion         *LatestVersion         `json:"latest_version,omitempty"`   // Latest version lookup for the Service
	Command               *CommandSlice          `json:"command,omitempty"`          // OS Commands to run on new release
	Notify                *NotifySlice           `json:"notify,omitempty"`           // Service-specific Notify vars
	WebHook               *WebHookSlice          `json:"webhook,omitempty"`          // Service-specific WebHook vars
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	Dashboard             *DashboardOptions      `json:"dashboard,omitempty"`        // Dashboard options
	Status                *Status                `json:"status,omitempty"`           // Track the Status of this source (version and regex misses)
}

// ServiceDefaults are default values for a Service.
type ServiceDefaults struct {
	Service `json:",inline"`

	LatestVersion *LatestVersionDefaults `json:"latest_version,omitempty"` // Latest version lookup for the Service
}

// ServiceOptions.
type ServiceOptions struct {
	Active             *bool  `json:"active,omitempty"`              // Active Service?
	Interval           string `json:"interval,omitempty"`            // AhBmCs = Sleep A hours, B minutes and C seconds between queries
	SemanticVersioning *bool  `json:"semantic_versioning,omitempty"` // default - true = Version has to be greater than the previous to trigger alerts/WebHooks
}

// DashboardOptions.
type DashboardOptions struct {
	AutoApprove *bool  `json:"auto_approve,omitempty"` // default - true = Requre approval before actioning new releases
	Icon        string `json:"icon,omitempty"`         // Icon URL to use for messages/Web UI
	IconLinkTo  string `json:"icon_link_to,omitempty"` // URL to redirect Icon clicks to
	WebURL      string `json:"web_url,omitempty"`      // URL to provide on the Web UI
}

// LatestVersion lookup of the service.
type LatestVersion struct {
	Type              string                `json:"type,omitempty"`                // Service Type, github/url
	URL               string                `json:"url,omitempty"`                 // URL to query
	AccessToken       string                `json:"access_token,omitempty"`        // GitHub access token to use
	AllowInvalidCerts *bool                 `json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates
	UsePreRelease     *bool                 `json:"use_prerelease,omitempty"`      // Whether GitHub prereleases should be used
	URLCommands       *URLCommandSlice      `json:"url_commands,omitempty"`        // Commands to filter the release from the URL request
	Require           *LatestVersionRequire `json:"require,omitempty"`             // Requirements for the version to be considered valid
}

// LatestVersionRequireDefaults are default values for a LatestVersion.
type LatestVersionDefaults struct {
	LatestVersion `json:",inline"`

	Require *LatestVersionRequireDefaults `json:"require,omitempty"`
}

// String returns a string representation of the LatestVersionRequireDefaults.
func (slice *LatestVersionRequireDefaults) String() (str string) {
	if slice != nil {
		str = util.ToJSONString(slice)
	}
	return
}

// LatestVersionRequire contains commands, regex etc for the release to be considered valid.
type LatestVersionRequire struct {
	Command      []string            `json:"command,omitempty"`       // Require Command to pass
	Docker       *RequireDockerCheck `json:"docker,omitempty"`        // Docker image tag requirements
	RegexContent string              `json:"regex_content,omitempty"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions
	RegexVersion string              `json:"regex_version,omitempty"` // "v*[0-9.]+" The version found must match this release to trigger new version actions
}

// LatestVersionRequireDefaults for the release to be considered valid.
type LatestVersionRequireDefaults struct {
	Docker RequireDockerCheckDefaults `json:"docker,omitempty"` // Docker repo defaults
}

type RequireDockerCheckRegistryDefaults struct {
	Token string `json:"token,omitempty"` // Token to get the token for the queries
}

type RequireDockerCheckRegistryDefaultsWithUsername struct {
	RequireDockerCheckRegistryDefaults
	Username string `json:"username,omitempty"` // Username to get a new token
}

type RequireDockerCheckDefaults struct {
	Type string                                          `json:"type,omitempty"` // Default DockerCheck Type
	GHCR *RequireDockerCheckRegistryDefaults             `json:"ghcr,omitempty"` // GHCR
	Hub  *RequireDockerCheckRegistryDefaultsWithUsername `json:"hub,omitempty"`  // DockerHub
	Quay *RequireDockerCheckRegistryDefaults             `json:"quay,omitempty"` // Quay
}

type RequireDockerCheck struct {
	Type     string `json:"type,omitempty"`     // Where to check, e.g. hub (DockerHub), GHCR, Quay
	Image    string `json:"image,omitempty"`    // Image to check
	Tag      string `json:"tag,omitempty"`      // Tag to check for
	Username string `json:"username,omitempty"` // Username to get a new token
	Token    string `json:"token,omitempty"`    // Token to get the token for the queries
}

// DeployedVersionLookup of the service.
type DeployedVersionLookup struct {
	URL               string                 `json:"url,omitempty"`                 // URL to query
	AllowInvalidCerts *bool                  `json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates
	BasicAuth         *BasicAuth             `json:"basic_auth,omitempty"`          // Basic Auth for the HTTP(S) request
	Headers           []Header               `json:"headers,omitempty"`             // Headers for the HTTP(S) request
	JSON              string                 `json:"json,omitempty"`                // JSON key to use e.g. version_current
	Regex             string                 `json:"regex,omitempty"`               // Regex to get the DeployedVersion
	HardDefaults      *DeployedVersionLookup `json:"-"`                             // Hardcoded default values
	Defaults          *DeployedVersionLookup `json:"-"`                             // Default values
}

// String returns a JSON string representation of the DeployedVersionLookup.
func (d *DeployedVersionLookup) String() (str string) {
	if d != nil {
		str = util.ToJSONString(d)
	}
	return
}

// BasicAuth to use on the HTTP(s) request.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `json:"key"`   // Header key, e.g. X-Sig
	Value string `json:"value"` // Value to give the key
}

// URLCommandSlice is a slice of URLCommand to be used to filter version from the URL Content.
type URLCommandSlice []URLCommand

// String returns a string representation of the URLCommandSlice.
func (slice *URLCommandSlice) String() (str string) {
	if slice != nil {
		str = util.ToJSONString(slice)
	}
	return
}

// URLCommand is a command to be ran to filter version from the URL body.
type URLCommand struct {
	Type  string  `json:"type,omitempty"`  // regex/replace/split
	Regex *string `json:"regex,omitempty"` // regex: regexp.MustCompile(Regex)
	Index int     `json:"index,omitempty"` // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index]
	Text  *string `json:"text,omitempty"`  // split:       strings.Split(tgtString, "Text")
	New   *string `json:"new,omitempty"`   // replace:     strings.ReplaceAll(tgtString, "Old", "New")
	Old   *string `json:"old,omitempty"`   // replace:     strings.ReplaceAll(tgtString, "Old", "New")
}

type Command []string
type CommandSlice []Command

// WebHookSlice is a slice mapping of WebHook.
type WebHookSlice map[string]*WebHook

// String returns a string representation of the WebHookSlice.
func (slice *WebHookSlice) String() (str string) {
	if slice != nil {
		str = util.ToJSONString(slice)
	}
	return
}

// Flatten the WebHookSlice into a list.
func (slice *WebHookSlice) Flatten() *[]*WebHook {
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

	return &list
}

// WebHook is a WebHook to send.
type WebHook struct {
	ServiceID         string    `json:"-"`                             // ID of the service this WebHook is attached to
	ID                string    `json:"name,omitempty"`                // Name of this WebHook
	Type              *string   `json:"type,omitempty"`                // "github"/"url"
	URL               *string   `json:"url,omitempty"`                 // "https://example.com"
	AllowInvalidCerts *bool     `json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates.
	Secret            *string   `json:"secret,omitempty"`              // "SECRET"
	CustomHeaders     *[]Header `json:"custom_headers,omitempty"`      // Custom Headers for the WebHook
	DesiredStatusCode *int      `json:"desired_status_code,omitempty"` // e.g. 202
	Delay             string    `json:"delay,omitempty"`               // The delay before sending the WebHook
	MaxTries          *uint     `json:"max_tries,omitempty"`           // Number of times to attempt sending the WebHook if the desired status code is not received
	SilentFails       *bool     `json:"silent_fails,omitempty"`        // Whether to notify if this WebHook fails MaxTries times
}

// String returns a string representation of the WebHook.
func (w *WebHook) String() (str string) {
	if w != nil {
		str = util.ToJSONString(w)
	}
	return
}

// Censor this WebHook for sending to the web client
func (w *WebHook) Censor() {
	if w == nil {
		return
	}

	// Secret
	if w.Secret != nil {
		secret := "<secret>"
		w.Secret = &secret
	}

	// Headers
	if w.CustomHeaders != nil {
		for i := range *w.CustomHeaders {
			(*w.CustomHeaders)[i].Value = "<secret>"
		}
	}
}

// Notifiers are the notifiers to use when a WebHook fails.
type Notifiers struct {
	Notify *NotifySlice // Service.Notify
}

// CommandSummary is the summary of a Command.
type CommandSummary struct {
	Failed       *bool     `json:"failed,omitempty"`        // Whether this WebHook failed to send successfully for the LatestVersion
	NextRunnable time.Time `json:"next_runnable,omitempty"` // Time the Command can next be run (for staggering)
}

// CommandStateUpdate will give an update of the current state of the Command
// @ index
type CommandStatusUpdate struct {
	Command string `json:"command"` // Index of the Command
	Failed  bool   `json:"failed"`  // Whether the last attempt of this command failed
}

// ################################################################################
// #                                     EDIT                                     #
// ################################################################################

// CommandEdit for JSON react-hook-form
type CommandEdit struct {
	Arg string `json:"arg,omitempty"` // Command argument
}

// ServiceEdit wants NotifySlice to be a list to prevent it being marshalled with indicies
type ServiceEdit struct {
	Comment               string                 `json:"comment,omitempty"`          // Comment on the Service
	Options               *ServiceOptions        `json:"options,omitempty"`          // Options to give the Service
	LatestVersion         *LatestVersion         `json:"latest_version,omitempty"`   // Latest version lookup for the Service
	Command               *CommandSlice          `json:"command,omitempty"`          // OS Commands to run on new release
	Notify                *[]Notify              `json:"notify,omitempty"`           // Service-specific Notify vars
	WebHook               *[]*WebHook            `json:"webhook,omitempty"`          // Service-specific WebHook vars
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	Dashboard             *DashboardOptions      `json:"dashboard,omitempty"`        // Dashboard options
	Status                *Status                `json:"status,omitempty"`           // Track the Status of this source (version and regex misses)
}
