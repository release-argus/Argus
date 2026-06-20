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

package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/release-argus/Argus/config/decode"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	svctest "github.com/release-argus/Argus/service/test"
	whtest "github.com/release-argus/Argus/webhook/test"
	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
)

var (
	packageName           = "api_v1"
	secretValueMarshalled string
)

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	ctx, cancel := context.WithCancel(context.Background())
	g, _ := errgroup.WithContext(ctx)

	config.DebounceDuration = 500 * time.Millisecond
	flags := make(map[string]bool)
	path := "TestWebAPIv1Main.yml"
	testYAML_Argus(path)
	var cfg config.Config
	cfg.Load(ctx, g, path, &flags)

	// Marshal the secret value '<secret>' -> '\u003csecret\u003e'.
	secretValueMarshalledBytes, _ := decode.Marshal("json", util.SecretValue)
	secretValueMarshalled = string(secretValueMarshalledBytes)

	// Run other tests.
	exitCode := m.Run()
	_ = os.Remove(path)
	_ = os.Remove(cfg.Settings.DataDatabaseFile())
	cancel()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

type failWriter struct {
	header http.Header
	code   int
	body   string
}

func (f *failWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failWriter) WriteHeader(code int) {
	f.code = code
}

func (f *failWriter) Write([]byte) (int, error) {
	if f.code == 0 {
		f.code = http.StatusOK
	}
	return 0, errors.New("write failed")
}

// PlainDefaults returns plain defaults and hardDefaults for testing.
func plainDefaults(t *testing.T) (*config.Defaults, *config.Defaults) {
	t.Helper()

	dockerDefaults, _ := docker.DecodeDefaults(
		"yaml", nil,
		nil,
	)
	defaults := config.Defaults{
		Service: service.Defaults{
			LatestVersion: lvbase.Defaults{
				Require: filter.RequireDefaults{
					Docker: *dockerDefaults,
				},
			},
		},
	}
	hardDefaults := config.Defaults{}
	hardDefaults.Default()
	hardDefaults.Service.LatestVersion.AccessToken = test.GitHubToken(t)
	defaults.SetDefaults(&hardDefaults)

	return &defaults, &hardDefaults
}

func testClient() Client {
	hub := NewHub()
	return Client{
		hub:            hub,
		ip:             "1.1.1.1",
		conn:           &websocket.Conn{},
		send: make(chan []byte, 5),
	}
}

func testLoad(t *testing.T, file string) *config.Config {
	var cfg config.Config
	g, _ := errgroup.WithContext(t.Context())

	flags := make(map[string]bool)
	cfg.Load(t.Context(), g, file, &flags)
	announceChannel := make(chan []byte, 8)
	cfg.HardDefaults.Service.Status.AnnounceChannel = announceChannel

	return &cfg
}

func testAPI(t *testing.T, path string) API {
	t.Helper()
	testYAML_Argus(path)

	cfg := testLoad(t, path)
	cfg.HardDefaults.Service.LatestVersion.AccessToken = test.GitHubToken(t)

	t.Cleanup(func() {
		_ = os.RemoveAll(cfg.Settings.Data.DatabaseFile)
		_ = os.RemoveAll(cfg.File)
	})

	return API{Config: cfg}
}

func testService(
	t *testing.T,
	id string,
	lvType, dvType string,
	semVer bool,
) *service.Service {
	if t != nil {
		t.Helper()
	}
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	svc := test.Must(t, func() (*service.Service, error) {
		return service.DecodeService(
			"yaml", []byte(test.TrimYAML(`
			options:
				semantic_versioning: `+fmt.Sprint(semVer)+`
			latest_version:
			`+lvtest.Lookup(t, lvType, false).String("  ")+`
			deployed_version:
			`+dvtest.Lookup(t, dvType, false, "").String("  ")+`
			dashboard:
				icon: https://example.com/icon.png
				icon_link_to: https://example.com/icon-{{ version }}.png
				web_url: https://example.com/{{ version }}
		`)),
			id,
			svcCfg, notifyCfg, whCfg,
		)
	})

	announceChannel := make(chan []byte, 8)
	databaseChannel := make(chan dbtype.Message, 8)

	// Status channels.
	svc.Status.AnnounceChannel = announceChannel
	svc.Status.DatabaseChannel = databaseChannel

	return svc
}

func testCommand(failing bool) command.Command {
	if failing {
		return command.Command{"ls", "-lah", "/root"}
	}
	return command.Command{"ls", "-lah"}
}

func testFaviconSettings(png string, svg string) *config.FaviconSettings {
	if svg == "" && png == "" {
		return nil
	}

	return &config.FaviconSettings{
		SVG: svg,
		PNG: png,
	}
}
