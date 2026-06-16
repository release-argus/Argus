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

// Package config provides the configuration for Argus.
package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
)

// Export the flags.
var (
	LogLevel = flag.String(
		"log.level",
		"INFO",
		"ERROR, WARN, INFO, VERBOSE or DEBUG (env_var=ARGUS_LOG_LEVEL)",
	)
	LogTimestamps = flag.Bool(
		"log.timestamps",
		false,
		"Enable timestamps in CLI output (env_var=ARGUS_LOG_TIMESTAMPS)",
	)
	DataDatabaseFile = flag.String(
		"data.database-file",
		"data/argus.db",
		"Database file path (env_var=ARGUS_DATA_DATABASE_FILE)",
	)
	WebListenHost = flag.String(
		"web.listen-host",
		"0.0.0.0",
		"IP address to listen on for UI, API, and telemetry (env_var=ARGUS_WEB_LISTEN_HOST)",
	)
	WebListenPort = flag.String(
		"web.listen-port",
		"8080",
		"Port to listen on for UI, API, and telemetry (env_var=ARGUS_WEB_LISTEN_PORT)",
	)
	WebCertFile = flag.String(
		"web.cert-file",
		"",
		"HTTPS certificate file path (env_var=ARGUS_WEB_CERT_FILE)",
	)
	WebPKeyFile = flag.String(
		"web.pkey-file",
		"",
		"HTTPS private key file path (env_var=ARGUS_WEB_PKEY_FILE)",
	)
	WebRoutePrefix = flag.String(
		"web.route-prefix",
		"/",
		"Prefix for web endpoints (env_var=ARGUS_WEB_ROUTE_PREFIX)",
	)
	WebBasicAuthUsername = flag.String(
		"web.basic-auth.username",
		"",
		"Username for basic auth (env_var=ARGUS_WEB_BASIC_AUTH_USERNAME)",
	)
	WebBasicAuthPassword = flag.String(
		"web.basic-auth.password",
		"",
		"Password for basic auth (env_var=ARGUS_WEB_BASIC_AUTH_PASSWORD)",
	)
)

// SettingsBase for the binary.
//
// (Used in Defaults).
type SettingsBase struct {
	Log  LogSettings  `json:"log,omitempty" yaml:"log,omitempty"`   // Log settings
	Data DataSettings `json:"data,omitempty" yaml:"data,omitempty"` // Data settings
	Web  WebSettings  `json:"web,omitempty" yaml:"web,omitempty"`   // Web settings
}

// IsZero implements the yaml.IsZeroer interface.
func (s SettingsBase) IsZero() bool {
	return s.Log.IsZero() &&
		s.Data.IsZero() &&
		s.Web.IsZero()
}

// CheckValues validates the fields of the receiver.
func (s *SettingsBase) CheckValues() error {
	var errs []error

	// Web.
	if err := s.Web.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "web",
				Err: err,
			},
		)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// MapEnvToStruct maps environment variables to this struct.
func (s *SettingsBase) MapEnvToStruct() error {
	if err := mapEnvToStruct(s, "", nil); err != nil {
		return fmt.Errorf("one or more 'ARGUS_' environment variables are incorrect: %w", err)
	}

	var errs []error
	// Set hash values, remove empty structs, and validate cert/pkey.
	if err := s.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "hard_defaults",
				Err: &decode.KeyFieldError{
					Key: "settings",
					Err: err,
				},
			},
		)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// DataSettings for the binary.
type DataSettings struct {
	DatabaseFile string `json:"database_file,omitempty" yaml:"database_file,omitempty"` // Database path
}

// IsZero implements the yaml.IsZeroer interface.
func (s DataSettings) IsZero() bool {
	return s.DatabaseFile == ""
}

// LogSettings for the binary.
type LogSettings struct {
	Timestamps *bool  `json:"timestamps,omitempty" yaml:"timestamps,omitempty"` // Timestamps in CLI output
	Level      string `json:"level,omitempty" yaml:"level,omitempty"`           // Log level
}

// IsZero implements the yaml.IsZeroer interface.
func (s LogSettings) IsZero() bool {
	return s.Timestamps == nil &&
		s.Level == ""
}

// WebSettingsBasicAuth contains the basic auth credentials to use (if any).
type WebSettingsBasicAuth struct {
	Username     string   `json:"username,omitempty" yaml:"username,omitempty"`
	UsernameHash [32]byte `json:"-" yaml:"-"` // SHA256 hash.
	Password     string   `json:"password,omitempty" yaml:"password,omitempty"`
	PasswordHash [32]byte `json:"-" yaml:"-"` // SHA256 hash.
}

// String returns a string representation of the receiver.
func (b *WebSettingsBasicAuth) String(prefix string) string {
	if b == nil {
		return ""
	}
	return decode.ToYAMLString(b, prefix)
}

// CheckValues ensures the fields of the receiver struct are SHA256 hashed.
func (b *WebSettingsBasicAuth) CheckValues() {
	// Username.
	b.UsernameHash = util.GetHash(util.EvalEnvVars(b.Username))

	// Password.
	password := util.EvalEnvVars(b.Password)
	b.PasswordHash = util.GetHash(password)
	if password == b.Password {
		// Password doesn't include an env var, so hash the config val.
		b.Password = util.FmtHash(b.PasswordHash)
	}
}

// FaviconSettings contains the favicon override settings.
type FaviconSettings struct {
	SVG string `json:"svg,omitempty" yaml:"svg,omitempty"`
	PNG string `json:"png,omitempty" yaml:"png,omitempty"`
}

// WebSettings for the binary.
type WebSettings struct {
	ListenHost     string                `json:"listen_host,omitempty" yaml:"listen_host,omitempty"`         // Web listen host.
	ListenPort     string                `json:"listen_port,omitempty" yaml:"listen_port,omitempty"`         // Web listen port.
	RoutePrefix    string                `json:"route_prefix,omitempty" yaml:"route_prefix,omitempty"`       // Web endpoint prefix.
	CertFile       string                `json:"cert_file,omitempty" yaml:"cert_file,omitempty"`             // HTTPS certificate path.
	KeyFile        string                `json:"pkey_file,omitempty" yaml:"pkey_file,omitempty"`             // HTTPS privkey path.
	BasicAuth      *WebSettingsBasicAuth `json:"basic_auth,omitempty" yaml:"basic_auth,omitempty"`           // Basic auth creds.
	DisabledRoutes []string              `json:"disabled_routes,omitempty" yaml:"disabled_routes,omitempty"` // Disabled API routes.
	Favicon        *FaviconSettings      `json:"favicon,omitempty" yaml:"favicon,omitempty"`                 // Favicon settings.
}

// IsZero implements the yaml.IsZeroer interface.
func (s WebSettings) IsZero() bool {
	return s.ListenHost == "" &&
		s.ListenPort == "" &&
		s.RoutePrefix == "" &&
		s.CertFile == "" &&
		s.KeyFile == "" &&
		s.BasicAuth == nil &&
		len(s.DisabledRoutes) == 0 &&
		s.Favicon == nil
}

// String returns a string representation of the receiver.
func (s *WebSettings) String(prefix string) string {
	if s == nil {
		return ""
	}
	return decode.ToYAMLString(s, prefix)
}

// CheckValues validates the fields of the receiver.
func (s *WebSettings) CheckValues() error {
	var errs []error

	// BasicAuth.
	if s.BasicAuth != nil {
		// Remove the BasicAuth if both the Username and Password are empty.
		if s.BasicAuth.Username == "" && s.BasicAuth.Password == "" {
			s.BasicAuth = nil
		} else {
			s.BasicAuth.CheckValues()
		}
	}

	// Route Prefix.
	if s.RoutePrefix != "" {
		// Ensure the RoutePrefix starts with one '/' and doesn't end with a '/'.
		s.RoutePrefix = strings.TrimLeft(s.RoutePrefix, "/")
		s.RoutePrefix = "/" + strings.TrimRight(s.RoutePrefix, "/")
	}

	// Favicon.
	if s.Favicon != nil {
		// Remove the Favicon override if both the SVG and PNG are empty.
		if s.Favicon.SVG == "" && s.Favicon.PNG == "" {
			s.Favicon = nil
		}
	}

	// CertFile.
	if err := util.CheckFileReadable(s.CertFile); err != nil {
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "cert_file",
				Value:       s.CertFile,
				Description: err.Error(),
			},
		)
	}

	// KeyFile.
	if err := util.CheckFileReadable(s.KeyFile); err != nil {
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "pkey_file",
				Value:       s.KeyFile,
				Description: err.Error(),
			},
		)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Settings for the binary.
type Settings struct {
	SettingsBase `json:",inline" yaml:",inline"` // SettingsBase for the binary.

	FromFlags    SettingsBase `json:"-" yaml:"-"` // Values from flags.
	HardDefaults SettingsBase `json:"-" yaml:"-"` // Hard defaults.
	Indentation  uint8        `json:"-" yaml:"-"` // Number of spaces used in the config.yml for indentation.
}

// IsZero implements the yaml.IsZeroer interface.
func (s Settings) IsZero() bool {
	return s.SettingsBase.IsZero()
}

// String returns a string representation of the receiver.
func (s *Settings) String(prefix string) string {
	if s == nil {
		return ""
	}
	return decode.ToYAMLString(s, prefix)
}

// NilUndefinedFlags sets the flags to nil if they are not set.
func (s *Settings) NilUndefinedFlags(flagset *map[string]bool) {
	for _, f := range []struct {
		Flag     string
		Variable any
	}{
		{"log.level", &LogLevel},
		{"log.timestamps", &LogTimestamps},
		{"data.database-file", &DataDatabaseFile},
		{"web.listen-host", &WebListenHost},
		{"web.listen-port", &WebListenPort},
		{"web.cert-file", &WebCertFile},
		{"web.pkey-file", &WebPKeyFile},
		{"web.route-prefix", &WebRoutePrefix},
		{"web.basic-auth.username", &WebBasicAuthUsername},
		{"web.basic-auth.password", &WebBasicAuthPassword},
	} {
		if !(*flagset)[f.Flag] {
			if strPtr, ok := f.Variable.(**string); ok {
				*strPtr = nil
			} else if boolPtr, ok := f.Variable.(**bool); ok {
				*boolPtr = nil
			}
		}
	}
}

// Default sets the values of the receiver to their default values.
func (s *Settings) Default() bool {
	// #######
	// # LOG #
	// #######
	s.FromFlags.Log = LogSettings{}

	// Timestamps.
	s.FromFlags.Log.Timestamps = LogTimestamps
	logTimestamps := false
	s.HardDefaults.Log.Timestamps = &logTimestamps

	// Level.
	s.FromFlags.Log.Level = util.DerefOrZero(LogLevel)
	s.HardDefaults.Log.Level = "INFO"

	// ########
	// # DATA #
	// ########
	s.FromFlags.Data = DataSettings{}

	// DatabaseFile.
	s.FromFlags.Data.DatabaseFile = util.DerefOrZero(DataDatabaseFile)
	s.HardDefaults.Data.DatabaseFile = "data/argus.db"

	// #######
	// # WEB #
	// #######
	s.FromFlags.Web = WebSettings{}

	// ListenHost.
	s.FromFlags.Web.ListenHost = util.DerefOrZero(WebListenHost)
	s.HardDefaults.Web.ListenHost = "0.0.0.0"

	// ListenPort.
	s.FromFlags.Web.ListenPort = util.DerefOrZero(WebListenPort)
	s.HardDefaults.Web.ListenPort = "8080"

	// CertFile.
	s.FromFlags.Web.CertFile = util.DerefOrZero(WebCertFile)
	// KeyFile.
	s.FromFlags.Web.KeyFile = util.DerefOrZero(WebPKeyFile)

	// RoutePrefix.
	s.FromFlags.Web.RoutePrefix = util.DerefOrZero(WebRoutePrefix)
	s.HardDefaults.Web.RoutePrefix = "/"

	// BasicAuth.
	if WebBasicAuthUsername != nil || WebBasicAuthPassword != nil {
		s.FromFlags.Web.BasicAuth = &WebSettingsBasicAuth{}
		s.FromFlags.Web.BasicAuth.Username = util.EvalEnvVars(util.DerefOrZero(WebBasicAuthUsername))
		s.FromFlags.Web.BasicAuth.Password = util.EvalEnvVars(util.DerefOrZero(WebBasicAuthPassword))
		s.FromFlags.Web.BasicAuth.CheckValues()
	}

	// Overwrite defaults with environment variables.
	if err := s.HardDefaults.MapEnvToStruct(); err != nil {
		//nolint:wrapcheck
		logx.Fatal(err, logx.LogFrom{})
		return false
	}

	return true
}

// LogTimestamps returns the log timestamps setting.
func (s *Settings) LogTimestamps() *bool {
	return util.FirstNonNilPtr(
		s.FromFlags.Log.Timestamps,
		s.Log.Timestamps,
		s.HardDefaults.Log.Timestamps,
	)
}

// LogLevel resolves the log level.
func (s *Settings) LogLevel() string {
	return strings.ToUpper(
		util.FirstNonDefaultWithEnv(
			s.FromFlags.Log.Level,
			s.Log.Level,
			s.HardDefaults.Log.Level,
		),
	)
}

// DataDatabaseFile resolves the path to the database file.
func (s *Settings) DataDatabaseFile() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Data.DatabaseFile,
		s.Data.DatabaseFile,
		s.HardDefaults.Data.DatabaseFile,
	)
}

// WebListenHost resolves the host to listen on.
func (s *Settings) WebListenHost() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.ListenHost,
		s.Web.ListenHost,
		s.HardDefaults.Web.ListenHost,
	)
}

// WebListenPort resolves the port to listen on.
func (s *Settings) WebListenPort() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.ListenPort,
		s.Web.ListenPort,
		s.HardDefaults.Web.ListenPort,
	)
}

// WebRoutePrefix resolves the prefix for the web endpoints.
func (s *Settings) WebRoutePrefix() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.RoutePrefix,
		s.Web.RoutePrefix,
		s.HardDefaults.Web.RoutePrefix,
	)
}

// WebCertFile resolves the path to the certificate file.
func (s *Settings) WebCertFile() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.CertFile,
		s.Web.CertFile,
		s.HardDefaults.Web.CertFile,
	)
}

// WebKeyFile resolves the path to the private key file.
func (s *Settings) WebKeyFile() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.KeyFile,
		s.Web.KeyFile,
		s.HardDefaults.Web.KeyFile,
	)
}

// WebBasicAuthUsernameHash resolves the SHA256 hash of the basic auth username.
func (s *Settings) WebBasicAuthUsernameHash() [32]byte {
	// Username set through flag.
	if s.FromFlags.Web.BasicAuth != nil && s.FromFlags.Web.BasicAuth.Username != "" {
		return s.FromFlags.Web.BasicAuth.UsernameHash
	}
	// Username set through config.
	if s.Web.BasicAuth != nil && s.Web.BasicAuth.Username != "" {
		return s.Web.BasicAuth.UsernameHash
	}
	return s.HardDefaults.Web.BasicAuth.UsernameHash
}

// WebBasicAuthPasswordHash resolves the SHA256 hash of the password.
func (s *Settings) WebBasicAuthPasswordHash() [32]byte {
	// Password set through flag.
	if s.FromFlags.Web.BasicAuth != nil && s.FromFlags.Web.BasicAuth.Password != "" {
		return s.FromFlags.Web.BasicAuth.PasswordHash
	}
	// Password set through config.
	if s.Web.BasicAuth != nil && s.Web.BasicAuth.Password != "" {
		return s.Web.BasicAuth.PasswordHash
	}
	return s.HardDefaults.Web.BasicAuth.PasswordHash
}
