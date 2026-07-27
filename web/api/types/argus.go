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

// Package types provides the types for the Argus API.
package types

import (
	"time"

	"github.com/goccy/go-yaml"

	"github.com/release-argus/Argus/config/decode"
	shoutrrr_types "github.com/release-argus/Argus/notify/shoutrrr/types"
	"github.com/release-argus/Argus/util"
)

// ServiceSummary is the Summary of a Service.
type ServiceSummary struct {
	ID                  string    `json:"id,omitzero" yaml:"id,omitzero"`
	Name                *string   `json:"name,omitzero" yaml:"name,omitzero"`                                   // Name for this Service.
	Active              *bool     `json:"active,omitzero" yaml:"active,omitzero"`                               // Active Service?
	Comment             *string   `json:"comment,omitzero" yaml:"comment,omitzero"`                             // Comment on the Service.
	Type                string    `json:"type,omitzero" yaml:"type,omitzero"`                                   // "github"|"URL".
	WebURL              *string   `json:"url,omitzero" yaml:"url,omitzero"`                                     // URL to provide on the Web UI.
	Icon                *string   `json:"icon,omitzero" yaml:"icon,omitzero"`                                   // Service.Dashboard.Icon / Service.Notify.*.Params.Icon / Service.Notify.*.Defaults.Params.Icon.
	IconLinkTo          *string   `json:"icon_link_to,omitzero" yaml:"icon_link_to,omitzero"`                   // URL to redirect Icon clicks to.
	DeployedVersionType *string   `json:"deployed_version_type,omitzero" yaml:"deployed_version_type,omitzero"` // "manual"|"url", empty string if no DeployedVersionLookup.
	Command             *int      `json:"command,omitzero" yaml:"command,omitzero"`                             // Amount of Commands to send on a new release.
	WebHook             *int      `json:"webhook,omitzero" yaml:"webhook,omitzero"`                             // Amount of WebHooks to send on a new release.
	Status              *Status   `json:"status,omitempty" yaml:"status,omitempty"`                             // Track the Status of this source (version and regex misses).
	Tags                *[]string `json:"tags,omitzero" yaml:"tags,omitzero"`                                   // Tags for the Service.
}

// IsZero implements the yaml.IsZeroer interface.
func (s *ServiceSummary) IsZero() bool {
	return s.ID == "" &&
		s.Name == nil &&
		s.Active == nil &&
		s.Comment == nil &&
		s.Type == "" &&
		s.WebURL == nil &&
		s.Icon == nil &&
		s.IconLinkTo == nil &&
		s.DeployedVersionType == nil &&
		s.Command == nil &&
		s.WebHook == nil &&
		s.Status == nil &&
		s.Tags == nil
}

// String implements fmt.Stringer and returns a JSON representation.
func (s *ServiceSummary) String() string {
	if s == nil {
		return ""
	}
	return decode.ToJSONString(s)
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
	if util.DerefOr(oldData.Active, true) ==
		util.DerefOr(s.Active, true) {
		s.Active = nil
	} else {
		active := util.DerefOr(s.Active, true)
		s.Active = &active
	}
	// Comment.
	s.Comment = nilIfUnchanged(oldData.Comment, s.Comment)
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
	// DeployedVersionType.
	s.DeployedVersionType = nilIfUnchanged(oldData.DeployedVersionType, s.DeployedVersionType)

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
		emptyTags := make([]string, 0)
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

// ActionSummary is the summary of all Actions for a Service.
type ActionSummary struct {
	Command map[string]CommandSummary `json:"command" yaml:"command"` // Summary of all Commands.
	WebHook map[string]WebHookSummary `json:"webhook" yaml:"webhook"` // Summary of all WebHooks.
}

// WebHookSummary is the summary of a WebHook.
type WebHookSummary struct {
	Failed       *bool     `json:"failed,omitzero" yaml:"failed,omitzero"`                 // Whether this WebHook failed to send successfully for the LatestVersion.
	NextRunnable time.Time `json:"next_runnable,omitempty" yaml:"next_runnable,omitempty"` // Time the WebHook can next run (for staggering).
}

// BuildInfo is information from build time.
type BuildInfo struct {
	Version   string `json:"version,omitzero" yaml:"version,omitzero"`
	BuildDate string `json:"build_date,omitzero" yaml:"build_date,omitzero"`
	GoVersion string `json:"go_version,omitzero" yaml:"go_version,omitzero"`
}

// RuntimeInfo defines current runtime information.
type RuntimeInfo struct {
	StartTime      time.Time `json:"start_time,omitzero" yaml:"start_time,omitzero"`
	CWD            string    `json:"cwd,omitzero" yaml:"cwd,omitzero"`
	GoRoutineCount int       `json:"goroutines,omitzero" yaml:"goroutines,omitzero"`
	GOMAXPROCS     int       `json:"GOMAXPROCS,omitzero" yaml:"GOMAXPROCS,omitzero"`
	GoGC           string    `json:"GOGC,omitzero" yaml:"GOGC,omitzero"`
	GoDebug        string    `json:"GODEBUG,omitzero" yaml:"GODEBUG,omitzero"`
}

// Flags define the runtime flags.
type Flags struct {
	ConfigFile       string `json:"config.file,omitzero" yaml:"config.file,omitzero"`
	LogLevel         string `json:"log.level,omitzero" yaml:"log.level,omitzero"`
	LogTimestamps    *bool  `json:"log.timestamps,omitzero" yaml:"log.timestamps,omitzero"`
	DataDatabaseFile string `json:"data.database-file,omitzero" yaml:"data.database-file,omitzero"`
	DataReadonly     bool   `json:"data.readonly" yaml:"data.readonly"`
	WebListenHost    string `json:"web.listen-host,omitzero" yaml:"web.listen-host,omitzero"`
	WebListenPort    string `json:"web.listen-port,omitzero" yaml:"web.listen-port,omitzero"`
	WebCertFile      string `json:"web.cert-file" yaml:"web.cert-file"`
	WebPKeyFile      string `json:"web.pkey-file" yaml:"web.pkey-file"`
	WebRoutePrefix   string `json:"web.route-prefix,omitzero" yaml:"web.route-prefix,omitzero"`
}

// Defaults are the global defaults for vars.
type Defaults struct {
	Service ServiceDefaults `json:"service,omitzero" yaml:"service,omitzero"`
	Notify  Notifiers       `json:"notify,omitempty" yaml:"notify,omitempty"`
	WebHook WebHook         `json:"webhook,omitzero" yaml:"webhook,omitzero"`
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.Service.IsZero() &&
		len(d.Notify) == 0 &&
		d.WebHook.IsZero()
}

// String implements fmt.Stringer and returns a JSON representation.
func (d *Defaults) String() string {
	if d == nil {
		return ""
	}

	return decode.ToJSONString(d)
}

// Config defines the config for Argus.
type Config struct {
	File         string    `json:"-" yaml:"-"`                                           // Path to the config file (-config.file='').
	Settings     Settings  `json:"settings,omitzero" yaml:"settings,omitzero"`           // Settings for the program.
	HardDefaults Defaults  `json:"hard_defaults,omitzero" yaml:"hard_defaults,omitzero"` // Hard default values.
	Defaults     Defaults  `json:"defaults,omitzero" yaml:"defaults,omitzero"`           // Default values.
	Notify       Notifiers `json:"notify,omitempty" yaml:"notify,omitempty"`             // Notify messages to send on a new release.
	WebHook      WebHooks  `json:"webhook,omitempty" yaml:"webhook,omitempty"`           // WebHooks to send on a new release.
	Service      Services  `json:"service,omitempty" yaml:"service,omitempty"`           // The services to monitor.
	Order        []string  `json:"order,omitempty" yaml:"order,omitempty"`               // Ordering for the Services in the WebUI.
}

// Settings contain settings for the program.
type Settings struct {
	Data DataSettings `json:"data,omitzero" yaml:"data,omitzero"`
	Log  LogSettings  `json:"log,omitzero" yaml:"log,omitzero"`
	Web  WebSettings  `json:"web,omitzero" yaml:"web,omitzero"`
	Auth AuthSettings `json:"auth,omitzero" yaml:"auth,omitzero"`
}

// IsZero implements the yaml.IsZeroer interface.
func (s *Settings) IsZero() bool {
	return s.Log.IsZero() &&
		s.Data.IsZero() &&
		s.Web.IsZero() &&
		s.Auth.IsZero()
}

// DataSettings contains data settings for the program.
type DataSettings struct {
	DatabaseFile string `json:"database_file,omitzero" yaml:"database_file,omitzero"` // Database file path.
	Readonly     *bool  `json:"readonly,omitzero" yaml:"readonly,omitzero"`           // Disable saving config changes to disk.
}

// IsZero implements the yaml.IsZeroer interface.
func (d DataSettings) IsZero() bool {
	return d.DatabaseFile == "" &&
		d.Readonly == nil
}

// LogSettings contains web settings for the program.
type LogSettings struct {
	Timestamps *bool  `json:"timestamps,omitzero" yaml:"timestamps,omitzero"` // Timestamps in command-line tool output.
	Level      string `json:"level,omitzero" yaml:"level,omitzero"`           // Log level.
}

// IsZero implements the yaml.IsZeroer interface.
func (l LogSettings) IsZero() bool {
	return l.Timestamps == nil &&
		l.Level == ""
}

// WebSettings contains web settings for the program.
type WebSettings struct {
	ListenHost     string   `json:"listen_host,omitzero" yaml:"listen_host,omitzero"`           // Web listen host.
	ListenPort     string   `json:"listen_port,omitzero" yaml:"listen_port,omitzero"`           // Web listen port.
	CertFile       string   `json:"cert_file,omitzero" yaml:"cert_file,omitzero"`               // HTTPS certificate path.
	KeyFile        string   `json:"pkey_file,omitzero" yaml:"pkey_file,omitzero"`               // HTTPS privkey path.
	RoutePrefix    string   `json:"route_prefix,omitzero" yaml:"route_prefix,omitzero"`         // Web endpoint prefix.
	DisabledRoutes []string `json:"disabled_routes,omitempty" yaml:"disabled_routes,omitempty"` // Disabled API routes.
	TrustedProxies []string `json:"trusted_proxies,omitempty" yaml:"trusted_proxies,omitempty"` // Peers whose forwarded headers are honoured.
}

// IsZero implements the yaml.IsZeroer interface.
func (w WebSettings) IsZero() bool {
	return w.ListenHost == "" &&
		w.ListenPort == "" &&
		w.CertFile == "" &&
		w.KeyFile == "" &&
		w.RoutePrefix == "" &&
		len(w.DisabledRoutes) == 0 &&
		len(w.TrustedProxies) == 0
}

// AuthSettings contains the authentication settings for the program.
type AuthSettings struct {
	Enabled *bool               `json:"enabled,omitzero" yaml:"enabled,omitzero"` // Enable user/RBAC auth.
	Session AuthSessionSettings `json:"session,omitzero" yaml:"session,omitzero"` // Session bounds.
	Local   AuthLocalSettings   `json:"local,omitzero" yaml:"local,omitzero"`     // Local provider.
}

// IsZero implements the yaml.IsZeroer interface.
func (a AuthSettings) IsZero() bool {
	return a.Enabled == nil &&
		a.Session.IsZero() &&
		a.Local.IsZero()
}

// AuthSessionSettings bounds login session validity.
type AuthSessionSettings struct {
	Lifetime    string `json:"lifetime,omitzero" yaml:"lifetime,omitzero"`         // Absolute cap.
	IdleTimeout string `json:"idle_timeout,omitzero" yaml:"idle_timeout,omitzero"` // Sliding window.
}

// IsZero implements the yaml.IsZeroer interface.
func (a AuthSessionSettings) IsZero() bool {
	return a.Lifetime == "" &&
		a.IdleTimeout == ""
}

// AuthLocalSettings configures the local (username/password) provider.
type AuthLocalSettings struct {
	Enabled *bool `json:"enabled,omitzero" yaml:"enabled,omitzero"`
}

// IsZero implements the yaml.IsZeroer interface.
func (a AuthLocalSettings) IsZero() bool {
	return a.Enabled == nil
}

// Notifiers is a string map of Notify.
type Notifiers map[string]*Notify

// String implements fmt.Stringer and returns a JSON representation.
func (n *Notifiers) String() string {
	if n == nil {
		return ""
	}
	return decode.ToJSONString(n)
}

// Flatten returns the Notifiers as an ordered flat list.
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

// Notify is a message notifier source.
type Notify struct {
	ID        string            `json:"name,omitzero" yaml:"name,omitzero"`               // ID for this Notify sender.
	Type      string            `json:"type,omitzero" yaml:"type,omitzero"`               // Notification type, e.g. slack.
	Options   map[string]string `json:"options,omitempty" yaml:"options,omitempty"`       // Options.
	URLFields map[string]string `json:"url_fields,omitempty" yaml:"url_fields,omitempty"` // URL Fields.
	Params    map[string]string `json:"params,omitempty" yaml:"params,omitempty"`         // Param props.
}

// Censor redacts secret url_fields and params with [util.SecretValue].
func (n *Notify) Censor() {
	if n == nil {
		return
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
}

// NotifyParams is a map of Notify parameters.
type NotifyParams map[string]string

// Services is a string map of Service.
type Services map[string]*Service

// Service defines a software source to track and where/what to notify.
type Service struct {
	Name                  string                 `json:"name,omitzero" yaml:"name,omitzero"`                         // Name for this Service.
	Comment               string                 `json:"comment,omitzero" yaml:"comment,omitzero"`                   // Extra detail on the Service..
	Options               ServiceOptions         `json:"options,omitzero" yaml:"options,omitzero"`                   // Options to give the Service Lookup's.
	LatestVersion         *LatestVersion         `json:"latest_version,omitzero" yaml:"latest_version,omitzero"`     // Latest version lookup for the Service.
	Notify                Notifiers              `json:"notify,omitempty" yaml:"notify,omitempty"`                   // Service-specific Notify configuration.
	Command               Commands               `json:"command,omitempty" yaml:"command,omitempty"`                 // CLI Commands to run on new release.
	WebHook               WebHooks               `json:"webhook,omitempty" yaml:"webhook,omitempty"`                 // Service-specific WebHook configuration.
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitzero" yaml:"deployed_version,omitzero"` // Configuration to scrape the Service's current deployed version.
	Dashboard             DashboardOptions       `json:"dashboard,omitzero" yaml:"dashboard,omitzero"`               // Dashboard options.
	Status                *Status                `json:"status,omitempty" yaml:"status,omitempty"`                   // Track the Status of this source (version and regex misses).
}

// String implements fmt.Stringer and returns a JSON representation.
func (r *Service) String() string {
	if r == nil {
		return ""
	}
	return decode.ToJSONString(r)
}

// ServiceDefaults defines default values for a Service.
type ServiceDefaults struct {
	Comment               string                        `json:"comment,omitzero" yaml:"comment,omitzero"`                   // Comment on the Service.
	Options               ServiceOptions                `json:"options,omitzero" yaml:"options,omitzero"`                   // Options to give the Service.
	LatestVersion         LatestVersionDefaults         `json:"latest_version,omitzero" yaml:"latest_version,omitzero"`     // Latest version lookup for the Service.
	Notify                []string                      `json:"notify,omitempty" yaml:"notify,omitempty"`                   // Service-specific Notify configuration.
	Command               Commands                      `json:"command,omitempty" yaml:"command,omitempty"`                 // CLI Commands to run on new release.
	WebHook               []string                      `json:"webhook,omitempty" yaml:"webhook,omitempty"`                 // Service-specific WebHook configuration.
	DeployedVersionLookup DeployedVersionLookupDefaults `json:"deployed_version,omitzero" yaml:"deployed_version,omitzero"` // Configuration to scrape the Service's current deployed version.
	Dashboard             DashboardOptions              `json:"dashboard,omitzero" yaml:"dashboard,omitzero"`               // Dashboard options.
}

// IsZero implements the yaml.IsZeroer interface.
func (d ServiceDefaults) IsZero() bool {
	return d.Comment == "" &&
		d.Options.IsZero() &&
		d.LatestVersion.IsZero() &&
		len(d.Notify) == 0 &&
		len(d.Command) == 0 &&
		len(d.WebHook) == 0 &&
		d.DeployedVersionLookup.IsZero() &&
		d.Dashboard.IsZero()
}

// ServiceOptions defines configuration options for a service.
type ServiceOptions struct {
	Active             *bool  `json:"active,omitzero" yaml:"active,omitzero"`                           // Active Service?.
	Interval           string `json:"interval,omitzero" yaml:"interval,omitzero"`                       // AhBmCs = Sleep A hours, B minutes and C seconds between queries.
	SemanticVersioning *bool  `json:"semantic_versioning,omitzero" yaml:"semantic_versioning,omitzero"` // Default - true = Version must exceed the previous version to trigger alerts/Commands/WebHooks.
}

// IsZero implements the yaml.IsZeroer interface.
func (o ServiceOptions) IsZero() bool {
	return o.Active == nil &&
		o.Interval == "" &&
		o.SemanticVersioning == nil
}

// DashboardOptions defines configuration options for a service on the Web UI dashboard.
type DashboardOptions struct {
	AutoApprove *bool    `json:"auto_approve,omitzero" yaml:"auto_approve,omitzero"` // Default - true = Require approval before actioning new releases.
	Icon        string   `json:"icon,omitzero" yaml:"icon,omitzero"`                 // Icon URL to use for messages/Web UI.
	IconLinkTo  string   `json:"icon_link_to,omitzero" yaml:"icon_link_to,omitzero"` // URL to redirect Icon clicks to.
	WebURL      string   `json:"web_url,omitzero" yaml:"web_url,omitzero"`           // URL to provide on the Web UI.
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`               // Tags for the Service.
}

// IsZero implements the yaml.IsZeroer interface.
func (d DashboardOptions) IsZero() bool {
	return d.AutoApprove == nil &&
		d.Icon == "" &&
		d.IconLinkTo == "" &&
		d.WebURL == "" &&
		len(d.Tags) == 0
}

// LatestVersion lookup of the service.
type LatestVersion struct {
	Type              string                `json:"type,omitzero" yaml:"type,omitzero"`                               // Service Type, github/url.
	URL               string                `json:"url,omitzero" yaml:"url,omitzero"`                                 // URL to query.
	AccessToken       string                `json:"access_token,omitzero" yaml:"access_token,omitzero"`               // GitHub access token to use.
	AllowInvalidCerts *bool                 `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Default - false = Disallows invalid HTTPS certificates.
	UsePreRelease     *bool                 `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"`           // Whether to use GitHub prereleases.
	URLCommands       URLCommands           `json:"url_commands,omitempty" yaml:"url_commands,omitempty"`             // Commands to filter the release from the URL request.
	Headers           []Header              `json:"headers,omitempty" yaml:"headers,omitempty"`                       // Request Headers.
	Require           *LatestVersionRequire `json:"require,omitzero" yaml:"require,omitzero"`                         // Requirements before treating a release as valid.
}

// String implements fmt.Stringer and returns a JSON representation.
func (r *LatestVersion) String() string {
	if r == nil {
		return ""
	}
	return decode.ToJSONString(r)
}

// LatestVersionDefaults are default values for a LatestVersion - fields common to
// every registered type live under 'common', and fields specific to a single type
// live under that type's own key.
type LatestVersionDefaults struct {
	Type   string                      `json:"type,omitzero" yaml:"type,omitzero"` // "github" | "url".
	Common LatestVersionCommonDefaults `json:"common,omitzero" yaml:"common,omitzero"`
	GitHub LatestVersionGitHubDefaults `json:"github,omitzero" yaml:"github,omitzero"`
	URL    LatestVersionURLDefaults    `json:"url,omitzero" yaml:"url,omitzero"`
}

// IsZero implements the yaml.IsZeroer interface.
func (l LatestVersionDefaults) IsZero() bool {
	return l.Type == "" &&
		l.Common.IsZero() &&
		l.GitHub.IsZero() &&
		l.URL.IsZero()
}

// LatestVersionCommonDefaults are default values common to every latest_version type.
type LatestVersionCommonDefaults struct {
	Require *LatestVersionRequireDefaults `json:"require,omitzero" yaml:"require,omitzero"`
}

// IsZero implements the yaml.IsZeroer interface.
func (l LatestVersionCommonDefaults) IsZero() bool {
	return l.Require == nil || l.Require.IsZero()
}

// LatestVersionGitHubDefaults are GitHub-specific default values for a LatestVersion.
type LatestVersionGitHubDefaults struct {
	AccessToken   string `json:"access_token,omitzero" yaml:"access_token,omitzero"`     // GitHub access token to use.
	UsePreRelease *bool  `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"` // Whether to use GitHub prereleases.
}

// IsZero implements the yaml.IsZeroer interface.
func (l LatestVersionGitHubDefaults) IsZero() bool {
	return l.AccessToken == "" && l.UsePreRelease == nil
}

// LatestVersionURLDefaults are URL-specific default values for a LatestVersion.
type LatestVersionURLDefaults struct {
	AllowInvalidCerts *bool `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Default - false = Disallows invalid HTTPS certificates.
}

// IsZero implements the yaml.IsZeroer interface.
func (l LatestVersionURLDefaults) IsZero() bool {
	return l.AllowInvalidCerts == nil
}

// LatestVersionRequire contains commands, regex, etc. that must pass before considering a release valid.
type LatestVersionRequire struct {
	Command      []string       `json:"command,omitempty" yaml:"command,omitempty"`           // Require Command to pass.
	Docker       *RequireDocker `json:"docker,omitempty" yaml:"docker,omitempty"`             // Docker image tag requirements.
	RegexContent string         `json:"regex_content,omitzero" yaml:"regex_content,omitzero"` // "abc-[a-z]+-{{ version }}_amd64.deb" This regex must exist in the body of the URL to trigger new version actions.
	RegexVersion string         `json:"regex_version,omitzero" yaml:"regex_version,omitzero"` // "v*[0-9.]+" The version found must match this release to trigger new version actions/.
}

// String implements fmt.Stringer and returns a JSON representation.
func (r *LatestVersionRequire) String() string {
	if r == nil {
		return ""
	}
	return decode.ToJSONString(r)
}

// LatestVersionRequireDefaults define the default requirements before considering a release valid.
type LatestVersionRequireDefaults struct {
	Docker RequireDockerDefaults `json:"docker,omitzero" yaml:"docker,omitzero"` // Docker repo defaults.
}

// IsZero implements the yaml.IsZeroer interface.
func (l LatestVersionRequireDefaults) IsZero() bool {
	return l.Docker.IsZero()
}

// String implements fmt.Stringer and returns a JSON representation.
func (r *LatestVersionRequireDefaults) String() string {
	if r == nil {
		return ""
	}

	return decode.ToJSONString(r)
}

type RequireDockerRegistriesDefaults struct {
	GHCR RequireDockerRegistryDefaults `json:"ghcr,omitzero" yaml:"ghcr,omitzero"` // GitHub Container Registry.
	Hub  RequireDockerRegistryDefaults `json:"hub,omitzero" yaml:"hub,omitzero"`   // Docker Hub.
	Quay RequireDockerRegistryDefaults `json:"quay,omitzero" yaml:"quay,omitzero"` // Quay.
}

// IsZero implements the yaml.IsZeroer interface.
func (r RequireDockerRegistriesDefaults) IsZero() bool {
	return (r.GHCR == nil || r.GHCR.IsZero()) &&
		(r.Hub == nil || r.Hub.IsZero()) &&
		(r.Quay == nil || r.Quay.IsZero())
}

type RequireDockerRegistryDefaults interface {
	GetToken() string
	yaml.IsZeroer
}

// RequireDockerRegistryDefaultsAuth are the auth values for a RequireDocker that takes a Token.
type RequireDockerRegistryDefaultsAuth struct {
	Token string `json:"token,omitzero" yaml:"token,omitzero"` // Token to get the token for the queries.
}

// IsZero implements the yaml.IsZeroer interface.
func (r RequireDockerRegistryDefaultsAuth) IsZero() bool {
	return r.Token == ""
}

// RequireDockerRegistryDefaultsAuthWithUsername are the auth values for a RequireDocker that takes a Username with Token.
type RequireDockerRegistryDefaultsAuthWithUsername struct {
	Username                          string `json:"username,omitzero" yaml:"username,omitzero"` // Username to get a new token.
	RequireDockerRegistryDefaultsAuth `json:",inline" yaml:",inline"`
}

// IsZero implements the yaml.IsZeroer interface.
func (r RequireDockerRegistryDefaultsAuthWithUsername) IsZero() bool {
	return r.Username == "" &&
		r.Token == ""
}

// RequireDockerRegistryDefaultsToken are the default values for a RequireDocker with just Token auth.
type RequireDockerRegistryDefaultsToken struct {
	RequireDockerRegistryDefaultsAuth `json:"auth" yaml:"auth"`
}

// IsZero implements the yaml.IsZeroer interface.
func (r *RequireDockerRegistryDefaultsToken) IsZero() bool {
	return r.RequireDockerRegistryDefaultsAuth.IsZero()
}

func (r *RequireDockerRegistryDefaultsToken) GetToken() string { return r.Token }

// RequireDockerCheckRegistryDefaultsTokenWithUsername are the default values for a RequireDocker with Token+Username auth.
type RequireDockerCheckRegistryDefaultsTokenWithUsername struct {
	RequireDockerRegistryDefaultsAuthWithUsername `json:"auth" yaml:"auth"`
}

// IsZero implements the yaml.IsZeroer interface.
func (r *RequireDockerCheckRegistryDefaultsTokenWithUsername) IsZero() bool {
	return r.RequireDockerRegistryDefaultsAuthWithUsername.IsZero()
}

func (r *RequireDockerCheckRegistryDefaultsTokenWithUsername) GetToken() string { return r.Token }

// RequireDockerDefaults are default values for a RequireDocker.
type RequireDockerDefaults struct {
	Type     string                          `json:"type,omitzero" yaml:"type,omitzero"`         // Default DockerCheck Type.
	Tag      string                          `json:"tag,omitzero" yaml:"tag,omitzero"`           // Default Tag template.
	Registry RequireDockerRegistriesDefaults `json:"registry,omitzero" yaml:"registry,omitzero"` // GHCR | Hub | Quay.
}

// IsZero implements the yaml.IsZeroer interface.
func (r RequireDockerDefaults) IsZero() bool {
	return r.Type == "" &&
		r.Tag == "" &&
		r.Registry.IsZero()
}

// RequireDocker points to a Docker repository for a release to qualify as valid.
type RequireDocker struct {
	Type     string `json:"type,omitzero" yaml:"type,omitzero"`         // Where to check, e.g. hub (Docker Hub), GHCR, Quay).
	Image    string `json:"image,omitzero" yaml:"image,omitzero"`       // Image to check.
	Tag      string `json:"tag,omitzero" yaml:"tag,omitzero"`           // Tag to check for.
	Username string `json:"username,omitzero" yaml:"username,omitzero"` // Username to get a new token.
	Token    string `json:"token,omitzero" yaml:"token,omitzero"`       // Token to get the token for the queries.
}

type DeployedVersionLookupDefaults struct {
	Type              string `json:"type,omitzero" yaml:"type,omitzero"`                               // "manual" | "url".
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Disallows invalid HTTPS certificates.
	Method            string `json:"method,omitzero" yaml:"method,omitzero"`                           // HTTP method.
}

// IsZero implements the yaml.IsZeroer interface.
func (d DeployedVersionLookupDefaults) IsZero() bool {
	return d.AllowInvalidCerts == nil &&
		d.Method == ""
}

// DeployedVersionLookup of the service.
type DeployedVersionLookup struct {
	Type string `json:"type,omitzero" yaml:"type,omitzero"` // Service Type, url/manual.

	// manual
	Version string `json:"version,omitzero" yaml:"version,omitzero"` // Deployed version.

	// url
	Method            string                 `json:"method,omitzero" yaml:"method,omitzero"`                           // HTTP method.
	URL               string                 `json:"url,omitzero" yaml:"url,omitzero"`                                 // URL to query.
	AllowInvalidCerts *bool                  `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Default - false = Disallows invalid HTTPS certificates.
	TargetHeader      string                 `json:"target_header,omitzero" yaml:"target_header,omitzero"`             // Header to target for the version.
	BasicAuth         *BasicAuth             `json:"basic_auth,omitzero" yaml:"basic_auth,omitzero"`                   // Basic Auth credentials.
	Headers           []Header               `json:"headers,omitempty" yaml:"headers,omitempty"`                       // Request Headers.
	Body              string                 `json:"body,omitzero" yaml:"body,omitzero"`                               // Request Body.
	JSON              string                 `json:"json,omitzero" yaml:"json,omitzero"`                               // JSON key to use e.g. version_current.
	Regex             string                 `json:"regex,omitzero" yaml:"regex,omitzero"`                             // Regex for the version.
	RegexTemplate     string                 `json:"regex_template,omitzero" yaml:"regex_template,omitzero"`           // Template to apply to the RegEx match.
	HardDefaults      *DeployedVersionLookup `json:"-" yaml:"-"`                                                       // Hardcoded default values.
	Defaults          *DeployedVersionLookup `json:"-" yaml:"-"`                                                       // Default values.
}

// String implements fmt.Stringer and returns a JSON representation.
func (d *DeployedVersionLookup) String() string {
	if d == nil {
		return ""
	}
	return decode.ToJSONString(d)
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

// String implements fmt.Stringer and returns a JSON representation.
func (slice *URLCommands) String() string {
	if slice == nil {
		return ""
	}
	return decode.ToJSONString(slice)
}

// URLCommand is a command to run to filter versions from the URL body.
type URLCommand struct {
	Type     string `json:"type,omitzero" yaml:"type,omitzero"`         // regex/replace/split.
	Regex    string `json:"regex,omitzero" yaml:"regex,omitzero"`       // regex: regexp.MustCompile(Regex).
	Index    *int   `json:"index,omitzero" yaml:"index,omitzero"`       // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index].
	Template string `json:"template,omitzero" yaml:"template,omitzero"` // regex: template.
	Text     string `json:"text,omitzero" yaml:"text,omitzero"`         // split:       strings.Split(tgtString, "Text").
	New      string `json:"new,omitzero" yaml:"new,omitzero"`           // replace:     strings.ReplaceAll(tgtString, "Old", "New").
	Old      string `json:"old,omitzero" yaml:"old,omitzero"`           // replace:     strings.ReplaceAll(tgtString, "Old", "New").
}

// Status is the Status of a Service.
type Status struct {
	ApprovedVersion          string `json:"approved_version,omitzero" yaml:"approved_version,omitzero"`                     // The approved version.
	DeployedVersion          string `json:"deployed_version,omitzero" yaml:"deployed_version,omitzero"`                     // Deployed version of the Service.
	DeployedVersionTimestamp string `json:"deployed_version_timestamp,omitzero" yaml:"deployed_version_timestamp,omitzero"` // UTC timestamp that the deployed version changed.
	LatestVersion            string `json:"latest_version,omitzero" yaml:"latest_version,omitzero"`                         // Latest version of the Service.
	LatestVersionTimestamp   string `json:"latest_version_timestamp,omitzero" yaml:"latest_version_timestamp,omitzero"`     // UTC timestamp that the latest version last changed.
	LastQueried              string `json:"last_queried,omitzero" yaml:"last_queried,omitzero"`                             // UTC timestamp of the last query.
	RegexMissesContent       uint   `json:"regex_misses_content,omitzero" yaml:"regex_misses_content,omitzero"`             // Counter for the number of regular expression misses on URL content.
	RegexMissesVersion       uint   `json:"regex_misses_version,omitzero" yaml:"regex_misses_version,omitzero"`             // Counter for the number of regular expression misses on version.
}

// String implements fmt.Stringer and returns a JSON representation.
func (s *Status) String() string {
	if s == nil {
		return ""
	}
	return decode.ToJSONString(s)
}

// Command is a command to run.
type Command []string

// Commands is a slice of Command.
type Commands []Command

// WebHooks is a slice map of WebHook.
type WebHooks map[string]WebHook

// String implements fmt.Stringer and returns a JSON representation.
func (w *WebHooks) String() string {
	if w == nil {
		return ""
	}
	return decode.ToJSONString(w)
}

// Flatten returns the WebHooks as an ordered flat list.
func (w *WebHooks) Flatten() []WebHook {
	if w == nil {
		return nil
	}

	names := util.SortedKeys(*w)
	list := make([]WebHook, len(names))

	for index, name := range names {
		list[index] = (*w)[name]
		list[index].Censor()
		list[index].ID = name
	}

	return list
}

// WebHook is a WebHook to send.
type WebHook struct {
	ServiceID         string   `json:"-" yaml:"-"`                                                       // ID of the service this WebHook belongs to.
	ID                string   `json:"name,omitzero" yaml:"name,omitzero"`                               // Name of this WebHook.
	Type              string   `json:"type,omitzero" yaml:"type,omitzero"`                               // "github"/"url".
	URL               string   `json:"url,omitzero" yaml:"url,omitzero"`                                 // "https://example.com".
	AllowInvalidCerts *bool    `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Default - false = Disallows invalid HTTPS certificates.
	Secret            string   `json:"secret,omitzero" yaml:"secret,omitzero"`                           // "SECRET".
	Headers           []Header `json:"headers,omitempty" yaml:"headers,omitempty"`                       // Custom Headers for the WebHook.
	DesiredStatusCode *uint16  `json:"desired_status_code,omitzero" yaml:"desired_status_code,omitzero"` // e.g. 202.
	Delay             string   `json:"delay,omitzero" yaml:"delay,omitzero"`                             // The delay before sending the WebHook.
	MaxTries          *uint8   `json:"max_tries,omitzero" yaml:"max_tries,omitzero"`                     // Number of times to send the WebHook until we receive the desired status code.
	SilentFails       *bool    `json:"silent_fails,omitzero" yaml:"silent_fails,omitzero"`               // Whether to notify if this WebHook fails MaxTries times.
}

// IsZero implements the yaml.IsZeroer interface.
func (w WebHook) IsZero() bool {
	return w.ServiceID == "" &&
		w.ID == "" &&
		w.Type == "" &&
		w.URL == "" &&
		w.AllowInvalidCerts == nil &&
		w.Secret == "" &&
		len(w.Headers) == 0 &&
		w.DesiredStatusCode == nil &&
		w.Delay == "" &&
		w.MaxTries == nil &&
		w.SilentFails == nil
}

// String returns a string representation of the receiver.
func (w *WebHook) String(prefix string) string {
	if w == nil {
		return ""
	}
	return decode.ToYAMLString(w, prefix)
}

// Censor replaces the WebHook's secret and header values with [util.SecretValue].
func (w *WebHook) Censor() {
	if w == nil {
		return
	}

	// Secret.
	if w.Secret != "" {
		w.Secret = util.SecretValue
	}

	// Headers.
	if w.Headers != nil {
		for i := range w.Headers {
			w.Headers[i].Value = util.SecretValue
		}
	}
}

// CommandSummary holds the summary of a Command.
type CommandSummary struct {
	Failed       *bool     `json:"failed,omitzero" yaml:"failed,omitzero"`                 // Whether the last run failed.
	NextRunnable time.Time `json:"next_runnable,omitempty" yaml:"next_runnable,omitempty"` // Time at which the Command can next run (for staggering).
}

// CommandStatusUpdate holds the current state of a Command at a given index.
type CommandStatusUpdate struct {
	Command string `json:"command" yaml:"command"` // Index of the Command.
	Failed  bool   `json:"failed" yaml:"failed"`   // Whether the last attempt of this command failed.
}

// ################################################################################
// #                                     EDIT                                     #
// ################################################################################

// CommandEdit for JSON react-hook-form.
type CommandEdit struct {
	Arg string `json:"arg,omitzero" yaml:"arg,omitzero"` // Command argument.
}

// ServiceEdit is a Service in API format.
type ServiceEdit struct {
	Name                  string                 `json:"name,omitzero" yaml:"name,omitzero"`                         // Name of the Service.
	Comment               string                 `json:"comment,omitzero" yaml:"comment,omitzero"`                   // Comment on the Service.
	Options               ServiceOptions         `json:"options,omitzero" yaml:"options,omitzero"`                   // Options to give the Service.
	LatestVersion         *LatestVersion         `json:"latest_version,omitzero" yaml:"latest_version,omitzero"`     // Latest version lookup for the Service.
	Notify                []Notify               `json:"notify,omitempty" yaml:"notify,omitempty"`                   // Service-specific Notify vars.
	Command               Commands               `json:"command,omitempty" yaml:"command,omitempty"`                 // OS Commands to run on new release.
	WebHook               []WebHook              `json:"webhook,omitempty" yaml:"webhook,omitempty"`                 // Service-specific WebHook vars.
	DeployedVersionLookup *DeployedVersionLookup `json:"deployed_version,omitzero" yaml:"deployed_version,omitzero"` // Deployed version lookup for the Service.
	Dashboard             DashboardOptions       `json:"dashboard,omitzero" yaml:"dashboard,omitzero"`               // Dashboard options.
	Status                *Status                `json:"status,omitempty" yaml:"status,omitempty"`                   // Track the Status of this source (version and regex misses).
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
