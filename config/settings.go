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

// Package config provides the configuration for Argus.
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Export the flags.
var (
	LogLevel             = flag.String("log.level", "INFO", "ERROR, WARN, INFO, VERBOSE or DEBUG")
	LogTimestamps        = flag.Bool("log.timestamps", false, "Enable timestamps in CLI output.")
	DataDatabaseFile     = flag.String("data.database-file", "data/argus.db", "Database file path.")
	WebListenHost        = flag.String("web.listen-host", "0.0.0.0", "IP address to listen on for UI, API, and telemetry.")
	WebListenPort        = flag.String("web.listen-port", "8080", "Port to listen on for UI, API, and telemetry.")
	WebCertFile          = flag.String("web.cert-file", "", "HTTPS certificate file path.")
	WebPKeyFile          = flag.String("web.pkey-file", "", "HTTPS private key file path.")
	WebRoutePrefix       = flag.String("web.route-prefix", "/", "Prefix for web endpoints")
	WebBasicAuthUsername = flag.String("web.basic-auth.username", "", "Username for basic auth")
	WebBasicAuthPassword = flag.String("web.basic-auth.password", "", "Password for basic auth")
)

// Settings for the binary.
type Settings struct {
	SettingsBase `yaml:",inline"` // SettingsBase for the binary.

	FromFlags    SettingsBase `yaml:"-"` // Values from flags.
	HardDefaults SettingsBase `yaml:"-"` // Hard defaults.
	Indentation  uint8        `yaml:"-"` // Number of spaces used in the config.yml for indentation.
}

// String returns a string representation of the Settings.
func (s *Settings) String(prefix string) string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, prefix)
}

// SettingsBase for the binary.
//
// (Used in Defaults).
type SettingsBase struct {
	Log  LogSettings  `yaml:"log,omitempty"`  // Log settings
	Data DataSettings `yaml:"data,omitempty"` // Data settings
	Web  WebSettings  `yaml:"web,omitempty"`  // Web settings
}

// CheckValues validates the fields of the SettingsBase struct.
func (s *SettingsBase) CheckValues() {
	// Web
	s.Web.CheckValues()
}

// MapEnvToStruct maps environment variables to this struct.
func (s *SettingsBase) MapEnvToStruct() {
	err := mapEnvToStruct(s, "", nil)
	if err != nil {
		logutil.Log.Fatal(
			"One or more 'ARGUS_' environment variables are incorrect:\n"+err.Error(),
			logutil.LogFrom{}, true)
	}
	s.CheckValues() // Set hash values and remove empty structs.
}

// LogSettings for the binary.
type LogSettings struct {
	Timestamps *bool  `yaml:"timestamps,omitempty"` // Timestamps in CLI output
	Level      string `yaml:"level,omitempty"`      // Log level
}

// DataSettings for the binary.
type DataSettings struct {
	DatabaseFile string `yaml:"database_file,omitempty"` // Database path
}

// WebSettings for the binary.
type WebSettings struct {
	ListenHost     string                `yaml:"listen_host,omitempty"`     // Web listen host.
	ListenPort     string                `yaml:"listen_port,omitempty"`     // Web listen port.
	RoutePrefix    string                `yaml:"route_prefix,omitempty"`    // Web endpoint prefix.
	CertFile       string                `yaml:"cert_file,omitempty"`       // HTTPS certificate path.
	KeyFile        string                `yaml:"pkey_file,omitempty"`       // HTTPS privkey path.
	BasicAuth      *WebSettingsBasicAuth `yaml:"basic_auth,omitempty"`      // Basic auth creds.
	DisabledRoutes []string              `yaml:"disabled_routes,omitempty"` // Disabled API routes.
	Favicon        *FaviconSettings      `yaml:"favicon,omitempty"`         // Favicon settings.
}

// String returns a string representation of the WebSettings.
func (s *WebSettings) String(prefix string) string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, prefix)
}

// CheckValues validates the fields of the WebSettings struct.
func (s *WebSettings) CheckValues() {
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
}

// WebSettingsBasicAuth contains the basic auth credentials to use (if any).
type WebSettingsBasicAuth struct {
	Username     string   `yaml:"username,omitempty"`
	UsernameHash [32]byte `yaml:"-"` // SHA256 hash.
	Password     string   `yaml:"password,omitempty"`
	PasswordHash [32]byte `yaml:"-"` // SHA256 hash.
}

// String returns a string representation of the WebSettingsBasicAuth.
func (b *WebSettingsBasicAuth) String(prefix string) string {
	if b == nil {
		return ""
	}
	return util.ToYAMLString(b, prefix)
}

// CheckValues ensures the fields of the WebSettingsBasicAuth struct are SHA256 hashed.
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
	SVG string `yaml:"svg,omitempty"`
	PNG string `yaml:"png,omitempty"`
}

// NilUndefinedFlags sets the flags to nil if they are not set.
func (s *Settings) NilUndefinedFlags(flagset *map[string]bool) {
	for _, f := range []struct {
		Flag     string
		Variable interface{}
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

// Default sets these Settings to the default values.
func (s *Settings) Default() {
	// #######
	// # LOG #
	// #######
	s.FromFlags.Log = LogSettings{}

	// Timestamps.
	s.FromFlags.Log.Timestamps = LogTimestamps
	logTimestamps := false
	s.HardDefaults.Log.Timestamps = &logTimestamps

	// Level.
	s.FromFlags.Log.Level = util.DereferenceOrDefault(LogLevel)
	s.HardDefaults.Log.Level = "INFO"

	// ########
	// # DATA #
	// ########
	s.FromFlags.Data = DataSettings{}

	// DatabaseFile.
	s.FromFlags.Data.DatabaseFile = util.DereferenceOrDefault(DataDatabaseFile)
	s.HardDefaults.Data.DatabaseFile = "data/argus.db"

	// #######
	// # WEB #
	// #######
	s.FromFlags.Web = WebSettings{}

	// ListenHost.
	s.FromFlags.Web.ListenHost = util.DereferenceOrDefault(WebListenHost)
	s.HardDefaults.Web.ListenHost = "0.0.0.0"

	// ListenPort.
	s.FromFlags.Web.ListenPort = util.DereferenceOrDefault(WebListenPort)
	s.HardDefaults.Web.ListenPort = "8080"

	// CertFile.
	s.FromFlags.Web.CertFile = util.DereferenceOrDefault(WebCertFile)
	// KeyFile.
	s.FromFlags.Web.KeyFile = util.DereferenceOrDefault(WebPKeyFile)

	// RoutePrefix.
	s.FromFlags.Web.RoutePrefix = util.DereferenceOrDefault(WebRoutePrefix)
	s.HardDefaults.Web.RoutePrefix = "/"

	// BasicAuth.
	if WebBasicAuthUsername != nil || WebBasicAuthPassword != nil {
		s.FromFlags.Web.BasicAuth = &WebSettingsBasicAuth{}
		s.FromFlags.Web.BasicAuth.Username = util.EvalEnvVars(util.DereferenceOrDefault(WebBasicAuthUsername))
		s.FromFlags.Web.BasicAuth.Password = util.EvalEnvVars(util.DereferenceOrDefault(WebBasicAuthPassword))
		s.FromFlags.Web.BasicAuth.CheckValues()
	}

	// Overwrite defaults with environment variables.
	s.HardDefaults.MapEnvToStruct()
}

// LogTimestamps returns the log timestamps setting.
func (s *Settings) LogTimestamps() *bool {
	return util.FirstNonNilPtr(
		s.FromFlags.Log.Timestamps,
		s.Log.Timestamps,
		s.HardDefaults.Log.Timestamps)
}

// LogLevel returns the log level.
func (s *Settings) LogLevel() string {
	return strings.ToUpper(util.FirstNonDefaultWithEnv(
		s.FromFlags.Log.Level,
		s.Log.Level,
		s.HardDefaults.Log.Level))
}

// DataDatabaseFile returns the path to the database file.
func (s *Settings) DataDatabaseFile() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Data.DatabaseFile,
		s.Data.DatabaseFile,
		s.HardDefaults.Data.DatabaseFile)
}

// WebListenHost returns the host to listen on.
func (s *Settings) WebListenHost() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.ListenHost,
		s.Web.ListenHost,
		s.HardDefaults.Web.ListenHost)
}

// WebListenPort returns the port to listen on.
func (s *Settings) WebListenPort() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.ListenPort,
		s.Web.ListenPort,
		s.HardDefaults.Web.ListenPort)
}

// WebRoutePrefix returns the prefix for the web endpoints.
func (s *Settings) WebRoutePrefix() string {
	return util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.RoutePrefix,
		s.Web.RoutePrefix,
		s.HardDefaults.Web.RoutePrefix)
}

// WebCertFile returns the path to the certificate file.
func (s *Settings) WebCertFile() string {
	certFile := util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.CertFile,
		s.Web.CertFile,
		s.HardDefaults.Web.CertFile)
	if certFile == "" {
		return ""
	}
	if _, err := os.Stat(certFile); err != nil {
		if !filepath.IsAbs(certFile) {
			path, execErr := os.Executable()
			logutil.Log.Error(execErr, logutil.LogFrom{}, execErr != nil)
			// Add the path to the error message.
			err = errors.New(strings.Replace(
				err.Error(),
				fmt.Sprintf(" %s:", certFile),
				fmt.Sprintf(" %s/%s:", path, certFile),
				1,
			))
		}
		logutil.Log.Fatal("settings.web.cert_file "+err.Error(), logutil.LogFrom{}, true)
	}
	return certFile
}

// WebKeyFile returns the path to the private key file.
func (s *Settings) WebKeyFile() string {
	keyFile := util.FirstNonDefaultWithEnv(
		s.FromFlags.Web.KeyFile,
		s.Web.KeyFile,
		s.HardDefaults.Web.KeyFile)
	if keyFile == "" {
		return ""
	}
	if _, err := os.Stat(keyFile); err != nil {
		if !filepath.IsAbs(keyFile) {
			path, execErr := os.Executable()
			logutil.Log.Error(execErr, logutil.LogFrom{}, execErr != nil)
			// Add the path to the error message.
			err = errors.New(strings.Replace(
				err.Error(),
				fmt.Sprintf(" %s:", keyFile),
				fmt.Sprintf(" %s/%s:", path, keyFile),
				1,
			))
		}
		logutil.Log.Fatal("settings.web.key_file "+err.Error(), logutil.LogFrom{}, true)
	}
	return keyFile
}

// WebBasicAuthUsernameHash returns the SHA256 hash of the basic auth username.
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

// WebBasicAuthPasswordHash returns the SHA256 hash of the password.
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
