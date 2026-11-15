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

//go:build unit || integration

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/status"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

var (
	testTempDir string // Process-lifetime home of test config files.
)

func TestMain(m *testing.M) {
	logtest.InitLog()
	logx.SetExitCodeChannel(make(chan string, 64))

	tempDir, err := os.MkdirTemp("", "argus-config-test")
	if err != nil {
		fmt.Printf("os.MkdirTemp() failed: %v\n", err)
		os.Exit(1)
	}
	testTempDir = tempDir

	exitCode := m.Run()
	_ = os.RemoveAll(tempDir)

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

// testConfig creates a Config with an inaccessible directory for the file
// (saves will write to ExitCodeChannel).
func testConfig(t *testing.T) *Config {
	t.Helper()

	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)

	dir := filepath.Join(t.TempDir(), "inaccessible")
	if err := os.Mkdir(dir, 0400); err != nil {
		t.Fatalf("os.Mkdir() failed: %v", err)
	}

	cfg := Config{
		File: filepath.Join(dir, "config.yml"),
		Settings: Settings{
			Indentation: 4,
			Log: LogSettings{
				Level: "DEBUG",
			},
		},
		Notify:  shoutrrr.ShoutrrrsDefaults{},
		Webhook: webhook.WebhooksDefaults{},
		Defaults: Defaults{
			Notify:  shoutrrr.ShoutrrrsDefaults{},
			Webhook: webhook.Defaults{},
		},
		HardDefaults: Defaults{
			Service: service.Defaults{
				Status: status.NewDefaults(
					nil, databaseChannel, saveChannel,
				),
			},
			Notify:  shoutrrr.ShoutrrrsDefaults{},
			Webhook: webhook.Defaults{},
		},
		DatabaseChannel: databaseChannel,
		SaveChannel:     saveChannel,
	}

	cfg.HardDefaults.Default()

	return &cfg
}

func testSettings() Settings {
	return Settings{
		Log: LogSettings{
			Timestamps: new(true),
			Level:      "DEBUG",
		},
		Data: DataSettings{
			DatabaseFile: "somewhere.db",
			Readonly:     new(true),
		},
		Web: WebSettings{
			ListenHost:  "test",
			ListenPort:  "123",
			RoutePrefix: "/something",
			CertFile:    "../README.md",
			KeyFile:     "../LICENSE",
		},
	}
}

var loadMu sync.RWMutex

// snapshotFlags returns a func that restores every flag to its current value. The caller holds
// loadMu.
func snapshotFlags() func() {
	strFlags := []**string{
		&LogLevel, &DataDatabaseFile,
		&WebListenHost, &WebListenPort, &WebCertFile, &WebPKeyFile, &WebRoutePrefix,
		&WebBasicAuthUsername, &WebBasicAuthPassword, &AuthCreateAdmin, &AuthResetPassword,
	}
	boolFlags := []**bool{&LogTimestamps, &DataReadonly}
	strHad := make([]*string, len(strFlags))
	boolHad := make([]*bool, len(boolFlags))
	for i, flag := range strFlags {
		strHad[i] = *flag
	}
	for i, flag := range boolFlags {
		boolHad[i] = *flag
	}

	return func() {
		for i, flag := range strFlags {
			*flag = strHad[i]
		}
		for i, flag := range boolFlags {
			*flag = boolHad[i]
		}
	}
}

// nilUndefinedFlags nils every flag until the test ends. The caller holds loadMu.
func nilUndefinedFlags(t *testing.T, settings *Settings) {
	restore := snapshotFlags()

	settings.NilUndefinedFlags(&map[string]bool{})

	t.Cleanup(func() {
		loadMu.Lock()
		defer loadMu.Unlock()
		restore()
	})
}

func testLoadBasic(t *testing.T, file string) *Config {
	t.Helper()

	cfg := &Config{}

	loadMu.Lock()
	defer loadMu.Unlock()
	nilUndefinedFlags(t, &cfg.Settings)

	cfg.File = file

	//#nosec G304 -- Loading the test config file
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) failed: %v", file, err)
	}

	if err = cfg.Decode(data); err != nil {
		t.Fatalf("initial: Config.Decode(%q) failed: %v", file, err)
	}

	saveChannel := make(chan bool, 32)
	cfg.SaveChannel = saveChannel
	cfg.HardDefaults.Service.Status.SaveChannel = cfg.SaveChannel

	databaseChannel := make(chan dbtype.Message, 32)
	cfg.DatabaseChannel = databaseChannel
	cfg.HardDefaults.Service.Status.DatabaseChannel = cfg.DatabaseChannel

	if err := cfg.Decode(data); err != nil {
		t.Fatalf("with channels: Config.Decode(%q) failed: %v", file, err)
	}

	cfg.GetOrder(data)

	cfg.CheckValues()

	return cfg
}

func testServiceURL(t *testing.T, id string) *service.Service {
	t.Helper()

	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	svc, err := service.DecodeService(
		"yaml", []byte(test.TrimYAML(`
			options:
				interval: 5s
				semantic_versioning: true
			latest_version:
				type: url
				url: `+test.LookupPlain.URLValid+`
				url_commands:
					- type: regex
						regex: 'v([0-9.]+)'
				require:
					regex_content: "{{ version }}-beta"
					regex_version: "[0-9]+"
				access_token: placeholder
			deployed_version_lookup:
				type: url
				method: GET
				url: `+test.LookupJSON.URLValid+`
				json: version
			dashboard:
				auto_approve: false
				icon: test
				icon_link_to: https://release-argus.io
				web_url: https://release-argus.io/docs
		`)),
		id,
		svcCfg, notifyCfg, whCfg,
	)
	if err != nil {
		t.Fatalf("service.DecodeService(%q) failed: %v", id, err)
	}

	svc.Status.SetLastQueried("")
	svc.Status.SetDeployedVersion("0.0.0", "2001-01-01T01:01:01Z", false)
	svc.Status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	svc.Status.SetApprovedVersion("1.1.1", false)
	return svc
}

// testServiceManualDV gives a Service whose 'manual' deployed_version carries a
// pending version.
func testServiceManualDV(t *testing.T, id, version string) *service.Service {
	t.Helper()

	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	svc, err := service.DecodeService(
		"yaml", []byte(test.TrimYAML(`
			options:
				interval: 5s
				semantic_versioning: true
			latest_version:
				type: url
				url: `+test.LookupBare.URLValid+`/v2.2.2
				url_commands:
					- type: regex
						regex: 'v([0-9.]+)'
			deployed_version:
				type: manual
				version: `+version+`
		`)),
		id,
		svcCfg, notifyCfg, whCfg,
	)
	if err != nil {
		t.Fatalf("service.DecodeService(id=%q, version=%q) failed: %v", id, version, err)
	}

	svc.Status.SetLastQueried("")
	svc.Status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	return svc
}

// plainDefaults returns plain defaults and hardDefaults for testing.
func plainDefaults(t *testing.T) (*Defaults, *Defaults) {
	t.Helper()

	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)
	defaults := &Defaults{
		Service: *svcCfg.Soft,
		Notify:  notifyCfg.Defaults,
		Webhook: *whCfg.Defaults,
	}

	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	defaults.SetDefaults(hardDefaults)

	return defaults, hardDefaults
}
