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
	packageName = "config"
	testTempDir string // Process-lifetime home of test config files.
)

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Sweep strays from earlier killed runs.
	strays, _ := filepath.Glob(filepath.Join(os.TempDir(), "argus-config-test*"))
	for _, stray := range strays {
		_ = os.RemoveAll(stray)
	}
	tempDir, err := os.MkdirTemp("", "argus-config-test")
	if err != nil {
		fmt.Printf("%s\ncreate temp dir: %v", packageName, err)
		os.Exit(1)
	}
	testTempDir = tempDir

	// Run other tests.
	exitCode := m.Run()
	_ = os.RemoveAll(tempDir)

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// testConfig creates a Config with an inaccessible directory for the file
// (saves will write to ExitCodeChannel).
func testConfig(t *testing.T) *Config {
	t.Helper()

	logLevel := "DEBUG"
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)

	dir := filepath.Join(t.TempDir(), "inaccessible")
	_ = os.Mkdir(dir, 0400)

	cfg := Config{
		File: filepath.Join(dir, "config.yml"),
		Settings: Settings{
			Indentation: 4,
			SettingsBase: SettingsBase{
				Log: LogSettings{
					Level: logLevel,
				},
			},
		},
		Notify:  shoutrrr.ShoutrrrsDefaults{},
		WebHook: webhook.WebHooksDefaults{},
		Defaults: Defaults{
			Notify:  shoutrrr.ShoutrrrsDefaults{},
			WebHook: webhook.Defaults{},
		},
		HardDefaults: Defaults{
			Service: service.Defaults{
				Status: status.NewDefaults(
					nil, databaseChannel, saveChannel,
				),
			},
			Notify:  shoutrrr.ShoutrrrsDefaults{},
			WebHook: webhook.Defaults{},
		},
		DatabaseChannel: databaseChannel,
		SaveChannel:     saveChannel,
	}

	cfg.HardDefaults.Default()

	return &cfg
}

func testSettings(t *testing.T) Settings {
	t.Helper()

	return Settings{
		SettingsBase: SettingsBase{
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
		},
	}
}

var loadMu sync.RWMutex

func testLoadBasic(t *testing.T, file string) *Config {
	cfg := &Config{}

	strFlags := []**string{
		&LogLevel, &DataDatabaseFile,
		&WebListenHost, &WebListenPort, &WebCertFile, &WebPKeyFile, &WebRoutePrefix,
		&WebBasicAuthUsername, &WebBasicAuthPassword,
	}
	boolFlags := []**bool{&LogTimestamps, &DataReadonly}
	strHad := make([]*string, len(strFlags))
	boolHad := make([]*bool, len(boolFlags))
	loadMu.Lock()
	defer loadMu.Unlock()
	for i, flag := range strFlags {
		strHad[i] = *flag
	}
	for i, flag := range boolFlags {
		boolHad[i] = *flag
	}

	cfg.Settings.NilUndefinedFlags(&map[string]bool{})

	t.Cleanup(func() {
		loadMu.Lock()
		defer loadMu.Unlock()
		for i, flag := range strFlags {
			*flag = strHad[i]
		}
		for i, flag := range boolFlags {
			*flag = boolHad[i]
		}
	})

	cfg.File = file

	//#nosec G304 -- Loading the test config file
	data, err := os.ReadFile(file)
	if err != nil {
		logx.Fatal(
			fmt.Sprintf(
				"%s\nError reading %q\n%s",
				packageName, file, err,
			),
			logx.LogFrom{},
		)
	}

	if err = cfg.Decode(data); err != nil {
		logx.Fatal(
			fmt.Sprintf(
				"%q\nUnmarshal of %q failed\n%s",
				packageName, file, err,
			),
			logx.LogFrom{},
		)
	}

	saveChannel := make(chan bool, 32)
	cfg.SaveChannel = saveChannel
	cfg.HardDefaults.Service.Status.SaveChannel = cfg.SaveChannel

	databaseChannel := make(chan dbtype.Message, 32)
	cfg.DatabaseChannel = databaseChannel
	cfg.HardDefaults.Service.Status.DatabaseChannel = cfg.DatabaseChannel

	if err := cfg.Decode(data); err != nil {
		t.Fatalf(
			"Unmarshal of %q failed\n%s",
			file, err,
		)
	}

	cfg.GetOrder(data)

	cfg.CheckValues()
	t.Logf(
		"%s - Loaded %q",
		packageName, file,
	)

	return cfg
}

func testServiceURL(t *testing.T, id string) *service.Service {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	svc, _ := service.DecodeService(
		"yaml", []byte(test.TrimYAML(`
			options:
				interval: 5s
				semantic_versioning: true
			latest_version:
				type: url
				url: `+test.LookupPlain["url_valid"]+`
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
				url: `+test.LookupJSON["url_valid"]+`
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

	svc.Status.SetLastQueried("")
	svc.Status.SetDeployedVersion("0.0.0", "2001-01-01T01:01:01Z", false)
	svc.Status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	svc.Status.SetApprovedVersion("1.1.1", false)
	return svc
}

// testServiceManualDV gives a Service whose `manual` deployed_version carries a
// pending version.
func testServiceManualDV(t *testing.T, id, version string) *service.Service {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	svc, err := service.DecodeService(
		"yaml", []byte(test.TrimYAML(`
			options:
				interval: 5s
				semantic_versioning: true
			latest_version:
				type: url
				url: `+test.LookupBare["url_valid"]+`/v2.2.2
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
		t.Fatalf("%s testServiceManualDV(id=%q, version=%q) unexpected error: %v",
			packageName, id, version, err,
		)
	}

	svc.Status.SetLastQueried("")
	svc.Status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	return svc
}

func plainDefaults(t *testing.T) (*Defaults, *Defaults) {
	t.Helper()

	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)
	defaults := &Defaults{
		Service: *svcCfg.Soft,
		Notify:  notifyCfg.Defaults,
		WebHook: *whCfg.Defaults,
	}

	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	defaults.SetDefaults(hardDefaults)

	return defaults, hardDefaults
}
