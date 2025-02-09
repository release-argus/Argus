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

//go:build unit || integration

package test

import (
	"strings"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
)

// Defaults returns a shoutrrr.Defaults instance for testing.
func Defaults(failing bool, selfSignedCert bool) *shoutrrr.Defaults {
	url := test.ValidCertNoProtocol
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	s := shoutrrr.NewDefaults(
		"gotify",
		map[string]string{"max_tries": "1"},
		map[string]string{
			"host":  url,
			"path":  "/gotify",
			"token": test.ShoutrrrGotifyToken()},
		map[string]string{"title": "default title"})
	if failing {
		s.URLFields["token"] = "invalid"
	}
	return s
}

// Shoutrrr returns a shoutrrr instance for testing.
func Shoutrrr(failing bool, selfSignedCert bool) *shoutrrr.Shoutrrr {
	url := test.ValidCertNoProtocol
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	s := shoutrrr.New(
		nil, "",
		"gotify",
		map[string]string{"max_tries": "1"},
		map[string]string{
			"host":  url,
			"path":  "/gotify",
			"token": test.ShoutrrrGotifyToken()},
		map[string]string{"title": "A Title!"},
		shoutrrr.NewDefaults(
			"",
			make(map[string]string), make(map[string]string), make(map[string]string)),
		shoutrrr.NewDefaults(
			"",
			make(map[string]string), make(map[string]string), make(map[string]string)),
		shoutrrr.NewDefaults(
			"",
			make(map[string]string), make(map[string]string), make(map[string]string)))
	s.Main.InitMaps()
	s.Defaults.InitMaps()
	s.HardDefaults.InitMaps()

	s.ID = "test"
	s.ServiceStatus = &status.Status{
		ServiceID: test.StringPtr("service"),
	}
	s.ServiceStatus.Fails.Shoutrrr.Init(1)
	s.Failed = &s.ServiceStatus.Fails.Shoutrrr

	if failing {
		s.URLFields["token"] = "invalid"
	}
	return s
}
