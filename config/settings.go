// Copyright [2022] [Argus]
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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/release-argus/Argus/utils"
)

// Export the flags.
var (
	LogLevel       = flag.String("log.level", "INFO", "ERROR, WARN, INFO, VERBOSE or DEBUG")
	LogTimestamps  = flag.Bool("log.timestamps", false, "Enable timestamps in CLI output.")
	WebListenHost  = flag.String("web.listen-host", "0.0.0.0", "IP address to listen on for UI, API, and telemetry.")
	WebListenPort  = flag.String("web.listen-port", "8080", "Port to listen on for UI, API, and telemetry.")
	WebCertFile    = flag.String("web.cert-file", "", "HTTPS certificate file path.")
	WebPKeyFile    = flag.String("web.pkey-file", "", "HTTPS private key file path.")
	WebRoutePrefix = flag.String("web.route-prefix", "/", "Prefix for web endpoints")
)

// Settings for the binary.
type Settings struct {
	Log          LogSettings  `yaml:"log,omitempty"` // Log settings
	Web          WebSettings  `yaml:"web,omitempty"` // Web settings
	FromFlags    SettingsBase `yaml:"-"`             // Values from flags
	HardDefaults SettingsBase `yaml:"-"`             // Hard defaults
	Indentation  uint8        `yaml:"-"`             // Number of spaces used in the config.yml for indentation
}

// SettingsBase for the binary.
//
// (Used in Defaults)
type SettingsBase struct {
	Log LogSettings `yaml:"-"`
	Web WebSettings `yaml:"-"`
}

// LogSettings for the binary.
type LogSettings struct {
	Timestamps *bool   `yaml:"timestamps,omitempty"` // Timestamps in CLI output
	Level      *string `yaml:"level,omitempty"`      // Log level
}

// WebSettings for the binary.
type WebSettings struct {
	ListenHost  *string `yaml:"listen_host,omitempty"`  // Web listen host
	ListenPort  *string `yaml:"listen_port,omitempty"`  // Web listen port
	RoutePrefix *string `yaml:"route_prefix,omitempty"` // Web endpoint prefix
	CertFile    *string `yaml:"cert_file,omitempty"`    // HTTPS certificate path
	KeyFile     *string `yaml:"pkey_file,omitempty"`    // HTTPS privkey path
}

func (s *Settings) NilUndefinedFlags(flagset *map[string]bool) {
	if !(*flagset)["log.level"] {
		LogLevel = nil
	}
	if !(*flagset)["log.timestamps"] {
		LogTimestamps = nil
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

// SetDefaults initialises to the defaults.
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
}

// GetLogTimestamps.
func (s *Settings) GetLogTimestamps() *bool {
	return utils.GetFirstNonNilPtr(s.FromFlags.Log.Timestamps, s.Log.Timestamps, s.HardDefaults.Log.Timestamps)
}

// GetLogLevel.
func (s *Settings) GetLogLevel() string {
	return strings.ToUpper(*utils.GetFirstNonNilPtr(s.FromFlags.Log.Level, s.Log.Level, s.HardDefaults.Log.Level))
}

// GetWebListenHost.
func (s *Settings) GetWebListenHost() string {
	return *utils.GetFirstNonNilPtr(s.FromFlags.Web.ListenHost, s.Web.ListenHost, s.HardDefaults.Web.ListenHost)
}

// GetWebListenPort.
func (s *Settings) GetWebListenPort() string {
	return *utils.GetFirstNonNilPtr(s.FromFlags.Web.ListenPort, s.Web.ListenPort, s.HardDefaults.Web.ListenPort)
}

// GetWebRoutePrefix.
func (s *Settings) GetWebRoutePrefix() string {
	return *utils.GetFirstNonNilPtr(s.FromFlags.Web.RoutePrefix, s.Web.RoutePrefix, s.HardDefaults.Web.RoutePrefix)
}

// GetWebCertFile.
func (s *Settings) GetWebCertFile() *string {
	certFile := utils.GetFirstNonNilPtr(s.FromFlags.Web.CertFile, s.Web.CertFile, s.HardDefaults.Web.CertFile)
	if certFile == nil || *certFile == "" {
		return nil
	}
	if _, err := os.Stat(*certFile); err != nil {
		if !filepath.IsAbs(*certFile) {
			path, execErr := os.Executable()
			if execErr != nil {
				jLog.Error(execErr, utils.LogFrom{}, true)
			}
			err = fmt.Errorf(strings.Replace(
				err.Error(),
				" "+*certFile+":",
				" "+path+"/"+*certFile+":",
				1,
			))
		}
		jLog.Fatal("settings.web.cert_file "+err.Error(), utils.LogFrom{}, true)
	}
	return certFile
}

// GetWebKeyFile.
func (s *Settings) GetWebKeyFile() *string {
	keyFile := utils.GetFirstNonNilPtr(s.FromFlags.Web.KeyFile, s.Web.KeyFile, s.HardDefaults.Web.KeyFile)
	if keyFile == nil || *keyFile == "" {
		return nil
	}
	if _, err := os.Stat(*keyFile); err != nil {
		if !filepath.IsAbs(*keyFile) {
			path, execErr := os.Executable()
			if execErr != nil {
				jLog.Error(execErr, utils.LogFrom{}, true)
			}
			err = fmt.Errorf(strings.Replace(
				err.Error(),
				" "+*keyFile+":",
				" "+path+"/"+*keyFile+":",
				1,
			))
		}
		jLog.Fatal("settings.web.key_file "+err.Error(), utils.LogFrom{}, true)
	}
	return keyFile
}
