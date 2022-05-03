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

package types

import (
	"time"
)

// ServiceSummary is the Summary of a Service.
type ServiceSummary struct {
	ID                       *string `json:"id"`
	Type                     *string `json:"type,omitempty"`                 // "github"/"URL"
	URL                      *string `json:"url,omitempty"`                  // type:URL - "https://example.com", type:github - "owner/repo" or "https://github.com/owner/repo".
	Icon                     *string `json:"icon,omitempty"`                 // Service.Slack.IconURL / Slack.IconURL / Defaults.Slack.IconURL
	HasDeployedVersionLookup *bool   `json:"has_deployed_version,omitempty"` // Whether this service has a DeployedVersionLookup.
	WebHook                  int     `json:"webhook,omitempty"`              // Whether there are WebHook(s) to send on a new release.
	Status                   *Status `json:"status,omitempty"`               // Track the Status of this source (version and regex misses).
}

// Status is the Status of a Service.
type Status struct {
	ApprovedVersion         string `json:"approved_version,omitempty"`          // The version that's been approved
	CurrentVersion          string `json:"current_version,omitempty"`           // Track the current version of the service from the last successful WebHook.
	CurrentVersionTimestamp string `json:"current_version_timestamp,omitempty"` // UTC timestamp that the current version change was noticed.
	LatestVersion           string `json:"latest_version,omitempty"`            // Latest version found from query().
	LatestVersionTimestamp  string `json:"latest_version_timestamp,omitempty"`  // UTC timestamp that the latest version change was noticed.
	LastQueried             string `json:"last_queried,omitempty"`              // UTC timestamp that version was last queried/checked.
	RegexMissesContent      uint   `json:"regex_misses_content,omitempty"`      // Counter for the number of regex misses on URL content.
	RegexMissesVersion      uint   `json:"regex_misses_version,omitempty"`      // Counter for the number of regex misses on version.
}

// StatusFails keeps track of whether each of the notifications failed on the last version change.
type StatusFails struct {
	Gotify  *[]bool `json:"gotify,omitempty"`  // Track whether any of the Slice failed.
	Slack   *[]bool `json:"slack,omitempty"`   // Track whether any of the Slice failed.
	WebHook *[]bool `json:"webhook,omitempty"` // Track whether any of the WebHookSlice failed.
}

// StatusFailsSummary keeps track of whether any of the notifications failed on the last version change.
type StatusFailsSummary struct {
	Gotify  *bool `json:"gotify,omitempty"`  // Track whether any of the Slice failed.
	Slack   *bool `json:"slack,omitempty"`   // Track whether any of the Slice failed.
	WebHook *bool `json:"webhook,omitempty"` // Track whether any of the WebHookSlice failed.
}

// WebHookSummary is the summary of a WebHook.
type WebHookSummary struct {
	Failed *bool `json:"failed,omitempty"` // Whether this WebHook failed to send successfully for the LatestVersion.
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
	ConfigFile     *string `json:"config.file,omitempty"`
	LogLevel       string  `json:"log.level,omitempty"`
	LogTimestamps  *bool   `json:"log.timestamps,omitempty"`
	WebListenHost  string  `json:"web.listen-host,omitempty"`
	WebListenPort  string  `json:"web.listen-port,omitempty"`
	WebCertFile    *string `json:"web.cert-file"`
	WebPKeyFile    *string `json:"web.pkey-file"`
	WebRoutePrefix string  `json:"web.route-prefix,omitempty"`
}

// Defaults is the global default for vars.
type Defaults struct {
	Service Service `json:"service,omitempty"`
	Gotify  Gotify  `json:"gotify,omitempty"`
	Slack   Slack   `json:"slack,omitempty"`
	WebHook WebHook `json:"webhook,omitempty"`
}

// Config is the config for Hymenaios.
type Config struct {
	File     string        `json:"-"`                  // Path to the config file (-config.file='')
	Defaults *Defaults     `json:"defaults,omitempty"` // Default values for the various parameters.
	Service  *ServiceSlice `json:"service,omitempty"`  // The service(s) to monitor.
	Gotify   *GotifySlice  `json:"gotify,omitempty"`   // Gotify message(s) to send on a new release.
	Slack    *SlackSlice   `json:"slack,omitempty"`    // Slack message(s) to send on a new release.
	WebHook  *WebHookSlice `json:"webhook,omitempty"`  // WebHook(s) to send on a new release.
	Settings *Settings     `json:"settings,omitempty"` // Settings for the program
	Order    []string      `json:"order,omitempty"`    // Ordering for the Service(s) in the WebUI
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

// GotifySlice is a slice mapping of Gotify.
type GotifySlice map[string]*Gotify

// Gotify is a Gotify message w/ destination and from details.
type Gotify struct {
	URL      *string `json:"url,omitempty"`       // "https://example.com
	Token    *string `json:"token,omitempty"`     // apptoken
	Title    *string `json:"title,omitempty"`     // "{{ service_id }} - {{ version }} released"
	Message  *string `json:"message,omitempty"`   // "Hymenaios"
	Extras   *Extras `json:"extras,omitempty"`    // Message extras
	Priority *int    `json:"priority,omitempty"`  // <1 = Min, 1-3 = Low, 4-7 = Med, >7 = High
	Delay    *string `json:"delay,omitempty"`     // The delay before sending the Gotify message.
	MaxTries *uint   `json:"max_tries,omitempty"` // Number of times to attempt sending the Gotify message if a 200 is not received.
}

// Extras are the message extras (https://gotify.net/docs/msgextras) for the Gotify messages.
type Extras struct {
	AndroidAction      *string `json:"android_action,omitempty"`      // URL to open on notification delivery
	ClientDisplay      *string `json:"client_display,omitempty"`      // Render message in 'text/plain' or 'text/markdown'
	ClientNotification *string `json:"client_notification,omitempty"` // URL to open on notification click
}

// ServiceSlice is a slice mapping of Service.
type ServiceSlice map[string]*Service

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	Type                  *string                `json:"type,omitempty"`                // "github"/"URL"
	URL                   *string                `json:"url,omitempty"`                 // type:URL - "https://example.com", type:github - "owner/repo" or "https://github.com/owner/repo".
	WebURL                *string                `json:"web_url,omitempty"`             // URL to provide on the Web UI
	URLCommands           *URLCommandSlice       `json:"url_commands,omitempty"`        // Commands to filter the release from the URL request.
	Interval              *string                `json:"interval,omitempty"`            // AhBmCs = Sleep A hours, B minutes and C seconds between queries.
	SemanticVersioning    *bool                  `json:"semantic_versioning,omitempty"` // default - true  = Version has to be greater than the previous to trigger Slack(s)/WebHook(s).
	RegexContent          *string                `json:"regex_content,omitempty"`       // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions.
	RegexVersion          *string                `json:"regex_version,omitempty"`       // "v*[0-9.]+" The version found must match this release to trigger new version actions.
	UsePreRelease         *bool                  `json:"use_prerelease,omitempty"`      // Whether GitHub prereleases should be used
	AutoApprove           *bool                  `json:"auto_approve,omitempty"`        // default - true = Requre approval before sending WebHook(s) for new releases
	IgnoreMisses          *bool                  `json:"ignore_misses,omitempty"`       // Ignore URLCommands that fail (e.g. split on text that doesn't exist)
	AccessToken           *string                `json:"access_token,omitempty"`        // GitHub access token to use.
	AllowInvalidCerts     *bool                  `json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates.
	Icon                  *string                `json:"icon,omitempty"`                // Icon URL to use for Slack messages/Web UI
	Gotify                *GotifySlice           `json:"gotify,omitempty"`              // Service-specific Gotify vars.
	Slack                 *SlackSlice            `json:"slack,omitempty"`               // Service-specific Slack vars.
	WebHook               *WebHookSlice          `json:"webhook,omitempty"`             // Service-specific WebHook vars.
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitempty"`    // Var to scrape the Service's current deployed version.
	Status                *Status                `json:"status,omitempty"`              // Track the Status of this source (version and regex misses).
}

// DeployedVersionLookup of the service.
type DeployedVersionLookup struct {
	URL               string                 `json:"url,omitempty"`                 // URL to query.
	AllowInvalidCerts *bool                  `json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates.
	BasicAuth         *BasicAuth             `json:"basic_auth,omitempty"`          // Basic Auth for the HTTP(S) request.
	Headers           []Header               `json:"headers,omitempty"`             // Headers for the HTTP(S) request.
	JSON              string                 `json:"json,omitempty"`                // JSON key to use e.g. version_current.
	Regex             string                 `json:"regex,omitempty"`               // Regex to get the CurrentVersion
	HardDefaults      *DeployedVersionLookup `json:"-"`                             // Hardcoded default values.
	Defaults          *DeployedVersionLookup `json:"-"`                             // Default values.
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

// URLCommand is a command to be ran to filter version from the URL body.
type URLCommand struct {
	Type         string  `json:"type,omitempty"`          // regex/regex_submatch/replace/split
	Regex        *string `json:"regex,omitempty"`         // regex/regex_submatch: regexp.MustCompile(Regex)
	Index        *int    `json:"index,omitempty"`         // regex_submatch/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index]
	Text         *string `json:"text,omitempty"`          // split:                strings.Split(tgtString, "Text")
	New          *string `json:"new,omitempty"`           // replace:              strings.ReplaceAll(tgtString, "Old", "New")
	Old          *string `json:"old,omitempty"`           // replace:              strings.ReplaceAll(tgtString, "Old", "New")
	IgnoreMisses *bool   `json:"ignore_misses,omitempty"` // Ignore this command failing (e.g. split on text that doesn't exist)
}

// SlackSlice is a slice mapping of Slack.
type SlackSlice map[string]*Slack

// Slack is a Slack message w/ destination and from details.
type Slack struct {
	URL         *string `json:"url,omitempty"`        // "https://example.com
	ServiceIcon *string `json:"-"`                    // Service.Icon
	IconEmoji   *string `json:"icon_emoji,omitempty"` // ":github:"
	IconURL     *string `json:"icon_url,omitempty"`   // "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
	Username    *string `json:"username,omitempty"`   // "Hymenaios"
	Message     *string `json:"message,omitempty"`    // "<{{ service_url }}|{{ service_id }}> - {{ version }} released"
	Delay       *string `json:"delay,omitempty"`      // The delay before sending the Slack message.
	MaxTries    *uint   `json:"max_tries,omitempty"`  // Number of times to attempt sending the Slack message if a 200 is not received.
}

// WebHookSlice is a slice mapping of WebHook.
type WebHookSlice map[string]*WebHook

// WebHook is a WebHook to send.
type WebHook struct {
	ServiceID         *string `json:"-"`                             // ID of the service this WebHook is attached to
	Type              *string `json:"type,omitempty"`                // "github"/"url"
	URL               *string `json:"url,omitempty"`                 // "https://example.com"
	Secret            *string `json:"secret,omitempty"`              // "SECRET"
	DesiredStatusCode *int    `json:"desired_status_code,omitempty"` // e.g. 202
	Delay             *string `json:"delay,omitempty"`               // The delay before sending the WebHook.
	MaxTries          *uint   `json:"max_tries,omitempty"`           // Number of times to attempt sending the WebHook if the desired status code is not received.
	SilentFails       *bool   `json:"silent_fails,omitempty"`        // Whether to notify if this WebHook fails MaxTries times.
}

// Notifiers are the notifiers to use when a WebHook fails.
type Notifiers struct {
	Gotify *GotifySlice // Service.Gotify
	Slack  *SlackSlice  // Service.WebHook
}
