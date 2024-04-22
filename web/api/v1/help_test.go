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

//go:build unit || integration

package v1

import (
	"os"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

var (
	loadMutex sync.Mutex
	loadCount int
)

func TestMain(m *testing.M) {
	// initialize jLog
	jLog := util.NewJLog("DEBUG", false)
	jLog.Testing = true
	flags := make(map[string]bool)
	path := "TestMain.yml"
	testYAML_Argus(path)
	var config config.Config
	config.Load(path, &flags, jLog)
	os.Remove(path)
	jLog.SetLevel("DEBUG")
	LogInit(jLog)

	// run other tests
	exitCode := m.Run()

	// exit
	os.Exit(exitCode)
}

func testClient() Client {
	hub := NewHub()
	return Client{
		hub:  hub,
		ip:   "1.1.1.1",
		conn: &websocket.Conn{},
		send: make(chan []byte, 5),
	}
}

func testLoad(file string, jLog *util.JLog) *config.Config {
	var config config.Config

	flags := make(map[string]bool)
	config.Load(file, &flags, nil)
	announceChannel := make(chan []byte, 8)
	config.HardDefaults.Service.Status.AnnounceChannel = &announceChannel

	return &config
}

func testAPI(name string) API {
	testYAML_Argus(name)

	// Only give the log once (to avoid potential RACE condition)
	var loadLog *util.JLog
	loadMutex.Lock()
	if loadCount == 0 {
		loadLog = jLog
		loadCount++
	}
	loadMutex.Unlock()

	cfg := testLoad(name, loadLog)
	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken != "" {
		cfg.HardDefaults.Service.LatestVersion.AccessToken = &accessToken
	}
	return API{Config: cfg}
}

func testService(id string) *service.Service {
	announceChannel := make(chan []byte, 8)
	databaseChannel := make(chan dbtype.Message, 8)
	svc := service.Service{
		ID:      id,
		Comment: "foo",
		LatestVersion: *latestver.New(
			nil,
			test.BoolPtr(false),
			nil, nil, nil, nil,
			"url",
			"https://valid.release-argus.io/plain",
			&filter.URLCommandSlice{
				{Type: "regex", Regex: test.StringPtr(`stable version: "v?([0-9.]+)"`)}},
			nil, nil, nil),
		DeployedVersionLookup: deployedver.New(
			test.BoolPtr(false),
			nil, nil,
			"foo.bar.version",
			nil, "", nil,
			&svcstatus.Status{},
			"https://valid.release-argus.io/json",
			nil, nil),
		Options: *opt.New(
			nil, "", test.BoolPtr(true),
			nil, nil)}
	serviceHardDefaults := service.Defaults{}
	serviceHardDefaults.SetDefaults()
	shoutrrrHardDefaults := shoutrrr.SliceDefaults{}
	shoutrrrHardDefaults.SetDefaults()
	webhookHardDefaults := webhook.WebHookDefaults{}
	webhookHardDefaults.SetDefaults()
	svc.Init(
		&service.Defaults{}, &serviceHardDefaults,
		&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrrHardDefaults,
		&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhookHardDefaults)
	svc.Status.AnnounceChannel = &announceChannel
	svc.Status.DatabaseChannel = &databaseChannel
	return &svc
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
		PNG: png}
}
