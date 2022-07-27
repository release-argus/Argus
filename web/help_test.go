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

//go:build testing

package web

import (
	"fmt"
	"net"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

func boolPtr(val bool) *bool {
	return &val
}
func intPtr(val int) *int {
	return &val
}
func stringPtr(val string) *string {
	return &val
}

func testConfig() config.Config {
	port, err := getFreePort()
	if err != nil {
		panic(err)
	}
	var (
		listenHost  string = "0.0.0.0"
		listenPort  string = fmt.Sprint(port)
		routePrefix string = "/"
	)
	webSettings := config.WebSettings{
		ListenHost:  &listenHost,
		ListenPort:  &listenPort,
		RoutePrefix: &routePrefix,
	}
	var (
		logLevel string = "WARN"
	)
	logSettings := config.LogSettings{
		Level: &logLevel,
	}
	var defaults config.Defaults
	defaults.SetDefaults()
	svc := testService("test")
	dvl := testDeployedVersion()
	svc.DeployedVersionLookup = &dvl
	svc.URLCommands = &filters.URLCommandSlice{testURLCommandRegex()}
	emptyNotify := shoutrrr.Shoutrrr{
		Options:   map[string]string{},
		Params:    map[string]string{},
		URLFields: map[string]string{},
	}
	notify := shoutrrr.Slice{
		"test": &shoutrrr.Shoutrrr{
			Options: map[string]string{
				"message": "{{ service_id }} release",
			},
			Params:       map[string]string{},
			URLFields:    map[string]string{},
			Main:         &emptyNotify,
			Defaults:     &emptyNotify,
			HardDefaults: &emptyNotify,
		},
	}
	notify["test"].Params = map[string]string{}
	svc.Notify = notify
	svcComment := "test service's comment"
	svc.Comment = &svcComment
	whPass := testWebHookPass("pass")
	whFail := testWebHookFail("pass")
	return config.Config{
		Settings: config.Settings{
			Web: webSettings,
			Log: logSettings,
		},
		Defaults: defaults,
		WebHook: webhook.Slice{
			whPass.ID: whPass,
			whFail.ID: whFail,
		},
		Notify: defaults.Notify,
		Service: service.Slice{
			svc.ID: &svc,
		},
		Order: &[]string{svc.ID},
		All:   []string{svc.ID},
	}
}

func getFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	ln.Close()
	if err != nil {
		return 0, err
	}
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func testService(id string) service.Service {
	var (
		sAnnounceChannel chan []byte           = make(chan []byte, 2)
		sDatabaseChannel chan db_types.Message = make(chan db_types.Message, 5)
		sSaveChannel     chan bool             = make(chan bool, 5)
	)
	svc := service.Service{
		ID:          id,
		Type:        "url",
		AccessToken: stringPtr("secret"),
		URL:         stringPtr("https://release-argus.io"),
		WebURL:      stringPtr("https://release-argus.io"),
		LatestVersion: latest_version.Lookup{
			AllowInvalidCerts: boolPtr(false),
			UsePreRelease:     boolPtr(false),
			Require: &filters.Require{
				RegexContent: "content",
				RegexVersion: "version",
			},
		},
		Options: options.Options{
			SemanticVersioning: boolPtr(true),
			Interval:           stringPtr("10m"),
			Defaults:           &options.Options{},
			HardDefaults:       &options.Options{},
		},
		Dashboard: service.DashboardOptions{
			AutoApprove:  boolPtr(false),
			Icon:         stringPtr("test"),
			Defaults:     &service.DashboardOptions{},
			HardDefaults: &service.DashboardOptions{},
		},
		Defaults:          &service.Service{},
		HardDefaults:      &service.Service{},
		Command:           command.Slice{command.Command{"ls", "-lah"}},
		CommandController: &command.Controller{},
		WebHook:           webhook.Slice{"test": &webhook.WebHook{URL: stringPtr("example.com")}},
		Status: service_status.Status{
			AnnounceChannel: &sAnnounceChannel,
			DatabaseChannel: &sDatabaseChannel,
			SaveChannel:     &sSaveChannel,
		},
	}
	svc.CommandController.Init(jLog, &id, &svc.Status, &svc.Command, nil, nil)
	return svc
}

func testCommandPass() command.Command {
	return command.Command{"ls", "-lah"}
}

func testCommandFail() command.Command {
	return command.Command{"ls", "-lah", "/root"}
}

func testWebHookPass(id string) *webhook.WebHook {
	var slice *webhook.Slice
	slice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil, nil)

	whType := "github"
	whURL := "https://httpbin.org/anything"
	whSecret := "secret"
	whAllowInvalidCerts := false
	whDesiredStatusCode := 0
	whDelay := "0s"
	whSilentFails := true
	whMaxTries := uint(1)
	parentInterval := "11m"
	return &webhook.WebHook{
		ID:                id,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		Delay:             &whDelay,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		ParentInterval:    &parentInterval,
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
}

func testWebHookFail(id string) *webhook.WebHook {
	var slice *webhook.Slice
	slice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil, nil)

	whType := "github"
	whURL := "https://httpbin.org/hidden-basic-auth/:user/:passwd"
	whSecret := "secret"
	whAllowInvalidCerts := false
	whDesiredStatusCode := 0
	whSilentFails := true
	whMaxTries := uint(1)
	parentInterval := "11m"
	return &webhook.WebHook{
		ID:                id,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		ParentInterval:    &parentInterval,
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
		Notifiers:         &webhook.Notifiers{},
	}
}

func testDeployedVersion() deployed_version.Lookup {
	var (
		allowInvalidCerts bool = false
	)
	dflt := &deployed_version.Lookup{}
	hardDflt := &deployed_version.Lookup{}
	return deployed_version.Lookup{
		URL:               "https://release-argus.io",
		AllowInvalidCerts: &allowInvalidCerts,
		Headers: []deployed_version.Header{
			{Key: "foo", Value: "bar"},
		},
		JSON:  "something",
		Regex: "([0-9]+) The Argus Developers",
		BasicAuth: &deployed_version.BasicAuth{
			Username: "fizz",
			Password: "buzz",
		},
		Defaults:     &dflt,
		HardDefaults: &hardDflt,
	}
}

func testURLCommandRegex() filters.URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	return filters.URLCommand{
		Type:  "regex",
		Regex: &regex,
		Index: index,
	}
}
