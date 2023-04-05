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

//go:build unit || integration

package v1

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func stringifyPointer[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}

func testLogging() {
	jLog := util.NewJLog("DEBUG", false)
	jLog.Testing = true
	service.LogInit(jLog)
}

func testClient() Client {
	hub := NewHub()
	api := API{}
	return Client{
		api:  &api,
		hub:  hub,
		ip:   "1.1.1.1",
		conn: &websocket.Conn{},
		send: make(chan []byte, 5),
	}
}

func testLoad(file string) *config.Config {
	var config config.Config

	flags := make(map[string]bool)
	config.Load(file, &flags, &util.JLog{})
	announceChannel := make(chan []byte, 8)
	config.HardDefaults.Service.Status.AnnounceChannel = &announceChannel

	return &config
}

func testAPI(name string) API {
	testYAML_Argus(name)
	cfg := testLoad(name)
	return API{
		Config: cfg,
		Log:    util.NewJLog("WARN", false),
	}
}

func testService(id string) *service.Service {
	announceChannel := make(chan []byte, 8)
	databaseChannel := make(chan dbtype.Message, 8)
	svc := service.Service{
		ID:      id,
		Comment: "foo",
		LatestVersion: latestver.Lookup{
			Type:              "url",
			URL:               "https://valid.release-argus.io/plain",
			AllowInvalidCerts: boolPtr(false),
			URLCommands: []filter.URLCommand{
				{Type: "regex", Regex: stringPtr(`stable version: "v?([0-9.]+)"`)}}},
		DeployedVersionLookup: &deployedver.Lookup{
			URL:               "https://valid.release-argus.io/json",
			AllowInvalidCerts: boolPtr(false),
			JSON:              "foo.bar.version"},
		Options: opt.Options{
			SemanticVersioning: boolPtr(true)},
		Status: svcstatus.Status{
			AnnounceChannel: &announceChannel,
			DatabaseChannel: &databaseChannel}}
	svc.Init(
		&service.Service{
			Options: opt.Options{}},
		&service.Service{
			Options:               opt.Options{},
			DeployedVersionLookup: &deployedver.Lookup{},
			Status: svcstatus.Status{
				AnnounceChannel: &announceChannel,
				DatabaseChannel: &databaseChannel}},
		&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
		&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{})
	return &svc
}
