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

//go:build unit

package testing

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config"
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func TestServiceTestWithNoService(t *testing.T) {
	// GIVEN a Config with a Service
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	serviceID := "test"
	cfg := config.Config{
		Service: service.Slice{
			serviceID: &service.Service{
				ID: &serviceID,
			},
		},
	}
	flag := ""
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN ServiceTest is called with an empty (undefined) flag
	ServiceTest(&flag, &cfg, jLog)

	// THEN nothing will be run/printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	want := ""
	if want != output {
		t.Errorf("ServiceTest with %q flag shouldn't print anything, got\n%s",
			flag, output)
	}
}
func TestServiceTestWithUnknownService(t *testing.T) {
	// GIVEN a Config with a Service
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		sID                string                = "test"
		sType              string                = "url"
		sURL               string                = "github.com/release-argus/argus/releases"
		sRegexVersion      string                = "[0-9.]+"
		sAllowInvalidCerts bool                  = false
		sIgnoreMisses      bool                  = false
		databaseChannel    chan db_types.Message = make(chan db_types.Message, 5)
	)
	svc := service.Service{
		ID:                &sID,
		Type:              &sType,
		URL:               &sURL,
		RegexVersion:      &sRegexVersion,
		AllowInvalidCerts: &sAllowInvalidCerts,
		IgnoreMisses:      &sIgnoreMisses,
		Status:            &service_status.Status{},
		DatabaseChannel:   &databaseChannel,
		Defaults:          &service.Service{},
		HardDefaults:      &service.Service{},
	}
	svc.Init(jLog, svc.Defaults, svc.HardDefaults)
	cfg := config.Config{
		Service: service.Slice{
			sID: &svc,
		},
	}
	flag := "other_" + sID
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() {
		r := recover()
		if !strings.Contains(r.(string), " could not be found ") {
			t.Error(r)
		}
	}()

	// WHEN ServiceTest is called with a Service not in the config
	ServiceTest(&flag, &cfg, jLog)

	// THEN it will be printed that the command couldn't be found
	t.Error("Should os.Exit(1), err")
}

func TestServiceTestWithGitHubServiceAsURL(t *testing.T) {
	// GIVEN a Config with a Service that should be type github
	jLog = utils.NewJLog("INFO", false)
	InitJLog(jLog)
	var (
		sID                 string                = "test"
		sType               string                = "url"
		sURL                string                = "github.com/release-argus/argus"
		sRegexVersion       string                = "[0-9.]+"
		sAllowInvalidCerts  bool                  = false
		sIgnoreMisses       bool                  = false
		sSemanticVersioning bool                  = false
		databaseChannel     chan db_types.Message = make(chan db_types.Message, 5)
	)
	svc := service.Service{
		ID:                 &sID,
		Type:               &sType,
		URL:                &sURL,
		RegexVersion:       &sRegexVersion,
		AllowInvalidCerts:  &sAllowInvalidCerts,
		IgnoreMisses:       &sIgnoreMisses,
		SemanticVersioning: &sSemanticVersioning,
		Status:             &service_status.Status{},
		DatabaseChannel:    &databaseChannel,
		Defaults:           &service.Service{},
		HardDefaults:       &service.Service{},
	}
	svc.Init(jLog, svc.Defaults, svc.HardDefaults)
	cfg := config.Config{
		Service: service.Slice{
			sID: &svc,
		},
	}
	flag := "test"
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true

	// WHEN ServiceTest is called with a Service not in the config
	ServiceTest(&flag, &cfg, jLog)

	// THEN it will be printed that the command couldn't be found
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "unsupported protocol scheme") {
		t.Errorf("Expected Query to have failed for config reasons\n%s",
			output)
	}
}

func TestServiceTestWithKnownService(t *testing.T) {
	// GIVEN a Config with a Service
	jLog = utils.NewJLog("WARN", false)
	InitJLog(jLog)
	var (
		sID                 string                = "test"
		sType               string                = "url"
		sURL                string                = "release-argus/argus"
		sRegexVersion       string                = "[0-9.]+"
		sAllowInvalidCerts  bool                  = false
		sIgnoreMisses       bool                  = false
		sSemanticVersioning bool                  = false
		databaseChannel     chan db_types.Message = make(chan db_types.Message, 5)
	)
	svc := service.Service{
		ID:                 &sID,
		Type:               &sType,
		URL:                &sURL,
		RegexVersion:       &sRegexVersion,
		AllowInvalidCerts:  &sAllowInvalidCerts,
		IgnoreMisses:       &sIgnoreMisses,
		SemanticVersioning: &sSemanticVersioning,
		Status:             &service_status.Status{},
		DatabaseChannel:    &databaseChannel,
		Defaults:           &service.Service{},
		HardDefaults:       &service.Service{},
	}
	svc.Init(jLog, svc.Defaults, svc.HardDefaults)
	cfg := config.Config{
		Service: service.Slice{
			sID: &svc,
		},
	}
	flag := "test"
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// Switch Fatal to panic to disable os.Exit(0).
	jLog.Testing = true

	// WHEN ServiceTest is called with a Service not in the config
	ServiceTest(&flag, &cfg, jLog)

	// THEN it will have printed the LatestVersion found
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "unsupported protocol scheme") {
		t.Errorf("Expected Query to have failed for config reasons\n%s",
			output)
	}
}

func TestServiceTestWithKnownServiceAndDeployedVersionLookup(t *testing.T) {
	// GIVEN a Config with a Service
	jLog = utils.NewJLog("INFO", false)
	InitJLog(jLog)
	var (
		sID                 string                = "test"
		sType               string                = "github"
		sURL                string                = "release-argus/argus"
		sAllowInvalidCerts  bool                  = false
		sIgnoreMisses       bool                  = false
		sSemanticVersioning bool                  = false
		databaseChannel     chan db_types.Message = make(chan db_types.Message, 5)
	)
	svc := service.Service{
		ID:                 &sID,
		Type:               &sType,
		URL:                &sURL,
		AllowInvalidCerts:  &sAllowInvalidCerts,
		IgnoreMisses:       &sIgnoreMisses,
		SemanticVersioning: &sSemanticVersioning,
		DeployedVersionLookup: &service.DeployedVersionLookup{
			URL:               "https://release-argus.io",
			AllowInvalidCerts: &sAllowInvalidCerts,
			Regex:             "([0-9]+) The Argus Developers",
		},
		Status:          &service_status.Status{},
		DatabaseChannel: &databaseChannel,
		Defaults: &service.Service{
			DeployedVersionLookup: &service.DeployedVersionLookup{},
		},
		HardDefaults: &service.Service{
			DeployedVersionLookup: &service.DeployedVersionLookup{},
		},
	}
	svc.Init(jLog, svc.Defaults, svc.HardDefaults)
	cfg := config.Config{
		Service: service.Slice{
			sID: &svc,
		},
	}
	flag := "test"
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	// Switch Fatal to panic to disable os.Exit(0).
	jLog.Testing = true

	// WHEN ServiceTest is called with a Service not in the config
	ServiceTest(&flag, &cfg, jLog)

	// THEN it should have printed the LatestVersion and DeployedVersion
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	output := string(out)
	if !strings.Contains(output, "Latest Release - ") {
		t.Errorf("Expected LatestVersion to be broadcast, got\n%s",
			output)
	}
	if !strings.Contains(output, "Deployed version - ") {
		t.Errorf("Expected DeployedVersion to be broadcast, got\n%s",
			output)
	}
}
