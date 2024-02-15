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

package config

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/release-argus/Argus/util"
)

// Export the flags.
var (
	LogLevel         = flag.String("log.level", "INFO", "ERROR, WARN, INFO, VERBOSE or DEBUG")
	LogTimestamps    = flag.Bool("log.timestamps", false, "Enable timestamps in CLI output.")
	DataDatabaseFile = flag.String("data.database-file", "data/argus.db", "Database file path.")
	WebListenHost    = flag.String("web.listen-host", "0.0.0.0", "IP address to listen on for UI, API, and telemetry.")
	WebListenPort    = flag.String("web.listen-port", "8080", "Port to listen on for UI, API, and telemetry.")
	WebCertFile      = flag.String("web.cert-file", "", "HTTPS certificate file path.")
	WebPKeyFile      = flag.String("web.pkey-file", "", "HTTPS private key file path.")
	WebRoutePrefix   = flag.String("web.route-prefix", "/", "Prefix for web endpoints")
)

// Settings for the binary.
type Settings struct {
	SettingsBase `yaml:",inline"` // SettingsBase for the binary

	FromFlags    SettingsBase `yaml:"-"` // Values from flags
	HardDefaults SettingsBase `yaml:"-"` // Hard defaults
	Indentation  uint8        `yaml:"-"` // Number of spaces used in the config.yml for indentation
}

// String returns a string representation of the Settings.
func (s *Settings) String(prefix string) (str string) {
	if s != nil {
		str = util.ToYAMLString(s, prefix)
	}
	return
}

// SettingsBase for the binary.
//
// (Used in Defaults)
type SettingsBase struct {
	Log  LogSettings  `yaml:"log,omitempty"`  // Log settings
	Data DataSettings `yaml:"data,omitempty"` // Data settings
	Web  WebSettings  `yaml:"web,omitempty"`  // Web settings
}

// MapEnvToStruct maps environment variables to this struct.
func (s *SettingsBase) MapEnvToStruct() {
	err := mapEnvToStruct(s, "", nil)
	if err != nil {
		jLog.Fatal(
			"One or more 'ARGUS_' environment variables are incorrect:\n"+
				strings.ReplaceAll(util.ErrorToString(err), "\\", "\n"),
			util.LogFrom{}, true)
	}
}

// LogSettings for the binary.
type LogSettings struct {
	Timestamps *bool   `yaml:"timestamps,omitempty"` // Timestamps in CLI output
	Level      *string `yaml:"level,omitempty"`      // Log level
}

// DataSettings for the binary.
type DataSettings struct {
	DatabaseFile *string `yaml:"database_file,omitempty"` // Database path
}

// WebSettings for the binary.
type WebSettings struct {
	ListenHost  *string               `yaml:"listen_host,omitempty"`  // Web listen host
	ListenPort  *string               `yaml:"listen_port,omitempty"`  // Web listen port
	RoutePrefix *string               `yaml:"route_prefix,omitempty"` // Web endpoint prefix
	CertFile    *string               `yaml:"cert_file,omitempty"`    // HTTPS certificate path
	KeyFile     *string               `yaml:"pkey_file,omitempty"`    // HTTPS privkey path
	BasicAuth   *WebSettingsBasicAuth `yaml:"basic_auth,omitempty"`   // Basic auth creds
	Favicon     *FaviconSettings      `yaml:"favicon,omitempty"`      // Favicon settings
}

// String returns a string representation of the WebSettings.
func (s *WebSettings) String(prefix string) (str string) {
	if s != nil {
		str = util.ToYAMLString(s, prefix)
	}
	return
}

func (s *WebSettings) CheckValues() {
	// BasicAuth
	if s.BasicAuth != nil {
		// Remove the BasicAuth if both the Username and Password are empty.
		if s.BasicAuth.Username == "" && s.BasicAuth.Password == "" {
			s.BasicAuth = nil
		} else {
			s.BasicAuth.CheckValues()
		}
	}

	// Favicon
	if s.Favicon != nil {
		// Remove the Favicon override if both the SVG and PNG are empty.
		if s.Favicon.SVG == "" && s.Favicon.PNG == "" {
			s.Favicon = nil
		}
	}
}

// WebSettingsBasicAuth contains the basic auth credentials to use (if any)
type WebSettingsBasicAuth struct {
	Username     string   `yaml:"username,omitempty"`
	UsernameHash [32]byte `yaml:"-"` // SHA256 hash
	Password     string   `yaml:"password,omitempty"`
	PasswordHash [32]byte `yaml:"-"` // SHA256 hash
}

// String returns a string representation of the WebSettingsBasicAuth.
func (s *WebSettingsBasicAuth) String(prefix string) (str string) {
	if s != nil {
		str = util.ToYAMLString(s, prefix)
	}
	return
}

// CheckValues will ensure that the values are SHA256 hashed.
func (ba *WebSettingsBasicAuth) CheckValues() {
	// Ensure it's hashed.
	sha256Regex := "^h__[a-f0-9]{64}$"
	// Username - Hash if not already hashed.
	if !util.RegexCheck(sha256Regex, ba.Username) {
		ba.UsernameHash = sha256.Sum256([]byte(ba.Username))
		ba.Username = fmt.Sprintf("h__%x", ba.UsernameHash)
	} else {
		hash, _ := hex.DecodeString(ba.Username[3:])
		copy(ba.UsernameHash[:], hash)
	}
	// Password - Hash if not already hashed.
	if !util.RegexCheck(sha256Regex, ba.Password) {
		ba.PasswordHash = sha256.Sum256([]byte(ba.Password))
		ba.Password = fmt.Sprintf("h__%x", ba.PasswordHash)
	} else {
		hash, _ := hex.DecodeString(ba.Password[3:])
		copy(ba.PasswordHash[:], hash)
	}
}

// FaviconSettings contains the favicon override settings.
type FaviconSettings struct {
	SVG string `yaml:"svg,omitempty"`
	PNG string `yaml:"png,omitempty"`
}

func (s *Settings) NilUndefinedFlags(flagset *map[string]bool) {
	if !(*flagset)["log.level"] {
		LogLevel = nil
	}
	if !(*flagset)["log.timestamps"] {
		LogTimestamps = nil
	}
	if !(*flagset)["data.database-file"] {
		DataDatabaseFile = nil
	}
	if !(*flagset)["web.listen-host"] {
		WebListenHost = nil
	}
	if !(*flagset)["web.listen-port"] {
		WebListenPort = nil
	}
	if !(*flagset)["web.cert-file"] {
		WebCertFile = nil
	}
	if !(*flagset)["web.pkey-file"] {
		WebPKeyFile = nil
	}
	if !(*flagset)["web.route-prefix"] {
		WebRoutePrefix = nil
	}
}

// SetDefaults initialises the Settings to the defaults.
func (s *Settings) SetDefaults() {
	// #######
	// # LOG #
	// #######
	s.FromFlags.Log = LogSettings{}

	// Timestamps
	s.FromFlags.Log.Timestamps = LogTimestamps
	logTimestamps := false
	s.HardDefaults.Log.Timestamps = &logTimestamps

	// Level
	s.FromFlags.Log.Level = LogLevel
	logLevel := "INFO"
	s.HardDefaults.Log.Level = &logLevel

	// ########
	// # DATA #
	// ########
	s.FromFlags.Data = DataSettings{}

	// DatabaseFile
	s.FromFlags.Data.DatabaseFile = DataDatabaseFile
	dataDatabaseFile := "data/argus.db"
	s.HardDefaults.Data.DatabaseFile = &dataDatabaseFile

	// #######
	// # WEB #
	// #######
	s.FromFlags.Web = WebSettings{}

	// ListenHost
	s.FromFlags.Web.ListenHost = WebListenHost
	webListenHost := "0.0.0.0"
	s.HardDefaults.Web.ListenHost = &webListenHost

	// ListenPort
	s.FromFlags.Web.ListenPort = WebListenPort
	webListenPort := "8080"
	s.HardDefaults.Web.ListenPort = &webListenPort

	// RoutePrefix
	s.FromFlags.Web.RoutePrefix = WebRoutePrefix
	webRoutePrefix := "/"
	s.HardDefaults.Web.RoutePrefix = &webRoutePrefix

	// CertFile
	s.FromFlags.Web.CertFile = WebCertFile

	// KeyFile
	s.FromFlags.Web.KeyFile = WebPKeyFile

	// Overwrite defaults with environment variables.
	s.HardDefaults.MapEnvToStruct()
}

// LogTimestamps.
func (s *Settings) LogTimestamps() *bool {
	return util.FirstNonNilPtr(
		s.FromFlags.Log.Timestamps,
		s.Log.Timestamps,
		s.HardDefaults.Log.Timestamps)
}

// LogLevel.
func (s *Settings) LogLevel() string {
	return strings.ToUpper(*util.FirstNonNilPtr(
		s.FromFlags.Log.Level,
		s.FromFlags.Log.Level,
		s.Log.Level,
		s.HardDefaults.Log.Level))
}

// DataDatabaseFile.
func (s *Settings) DataDatabaseFile() *string {
	return util.FirstNonNilPtr(
		s.FromFlags.Data.DatabaseFile,
		s.Data.DatabaseFile,
		s.HardDefaults.Data.DatabaseFile)
}

// WebListenHost.
func (s *Settings) WebListenHost() string {
	return *util.FirstNonNilPtr(
		s.FromFlags.Web.ListenHost,
		s.Web.ListenHost,
		s.HardDefaults.Web.ListenHost)
}

// WebListenPort.
func (s *Settings) WebListenPort() string {
	return *util.FirstNonNilPtr(
		s.FromFlags.Web.ListenPort,
		s.Web.ListenPort,
		s.HardDefaults.Web.ListenPort)
}

// WebRoutePrefix.
func (s *Settings) WebRoutePrefix() string {
	return *util.FirstNonNilPtr(
		s.FromFlags.Web.RoutePrefix,
		s.Web.RoutePrefix,
		s.HardDefaults.Web.RoutePrefix)
}

// WebCertFile.
func (s *Settings) WebCertFile() *string {
	certFile := util.FirstNonNilPtr(
		s.FromFlags.Web.CertFile,
		s.Web.CertFile,
		s.HardDefaults.Web.CertFile)
	if certFile == nil || *certFile == "" {
		return nil
	}
	if _, err := os.Stat(*certFile); err != nil {
		if !filepath.IsAbs(*certFile) {
			path, execErr := os.Executable()
			jLog.Error(execErr, util.LogFrom{}, execErr != nil)
			err = fmt.Errorf(strings.Replace(
				err.Error(),
				" "+*certFile+":",
				" "+path+"/"+*certFile+":",
				1,
			))
		}
		jLog.Fatal("settings.web.cert_file "+err.Error(), util.LogFrom{}, true)
	}
	return certFile
}

// WebKeyFile.
func (s *Settings) WebKeyFile() *string {
	keyFile := util.FirstNonNilPtr(
		s.FromFlags.Web.KeyFile,
		s.Web.KeyFile,
		s.HardDefaults.Web.KeyFile)
	if keyFile == nil || *keyFile == "" {
		return nil
	}
	if _, err := os.Stat(*keyFile); err != nil {
		if !filepath.IsAbs(*keyFile) {
			path, execErr := os.Executable()
			jLog.Error(execErr, util.LogFrom{}, execErr != nil)
			err = fmt.Errorf(strings.Replace(
				err.Error(),
				" "+*keyFile+":",
				" "+path+"/"+*keyFile+":",
				1,
			))
		}
		jLog.Fatal("settings.web.key_file "+err.Error(), util.LogFrom{}, true)
	}
	return keyFile
}

// CheckValues of the Settings.
func (s *Settings) CheckValues() {
	// Web
	s.Web.CheckValues()
}
