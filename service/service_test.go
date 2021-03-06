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

package service

import (
	"regexp"
	"testing"

	db_types "github.com/release-argus/Argus/db/types"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func TestServiceQuery(t *testing.T) {
	var (
		hardDefaults                   Service
		hardDefaultsAllowInvalidCerts  bool   = false
		hardDefaultsAccessToken        string = ""
		hardDefaultsSemanticVersioning bool   = true
		hardDefaultsUsePreRelease      bool   = false

		service                 Slice                 = Slice{}
		serviceID               string                = "GitHub_Query_Test"
		serviceType             string                = "github"
		serviceURL              string                = "go-gitea/gitea"
		serviceURLcommand0Type  string                = "regex"
		serviceURLcommand0Regex string                = "v([0-9.]+)"
		serviceRegexVersion     string                = "^[0-9.]+[0-9]$"
		serviceDatabaseChannel  chan db_types.Message = make(chan db_types.Message, 5)

		want = regexp.MustCompile(`^[0-9.]+[0-9]$`)
	)
	jLog = utils.NewJLog("WARN", false)
	hardDefaults = Service{
		AllowInvalidCerts:  &hardDefaultsAllowInvalidCerts,
		AccessToken:        &hardDefaultsAccessToken,
		SemanticVersioning: &hardDefaultsSemanticVersioning,
		UsePreRelease:      &hardDefaultsUsePreRelease,
	}
	service["GitHub_Query_Test"] = &Service{
		ID:   &serviceID,
		Type: &serviceType,
		URL:  &serviceURL,
		URLCommands: &URLCommandSlice{
			URLCommand{
				Type:  serviceURLcommand0Type,
				Regex: &serviceURLcommand0Regex,
			}},
		RegexVersion:    &serviceRegexVersion,
		Status:          &service_status.Status{},
		DatabaseChannel: &serviceDatabaseChannel,
		Defaults:        &Service{},
		HardDefaults:    &hardDefaults,
	}

	_, _ = service["GitHub_Query_Test"].Query()
	got := service["GitHub_Query_Test"].Status.LatestVersion

	if !want.MatchString(got) {
		t.Errorf(`%s.status.LatestVersion = %v, want match for %s`, *service["GitHub_Query_Test"].ID, got, want)
	}
}
