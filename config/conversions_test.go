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
	"testing"

	"github.com/release-argus/Argus/service"
	service_status "github.com/release-argus/Argus/service/status"
)

func TestConvertCurrentVersionToDeployedVersionWithNilStatus(t *testing.T) {
	// GIVEN a Config with Service's that have nil Status
	config := Config{
		Service: service.Slice{
			"test":  &service.Service{},
			"other": &service.Service{},
		},
	}

	// WHEN convertCurrentVersionToDeployedVersion is called on this Config
	config.convertCurrentVersionToDeployedVersion()

	// THEN Status stays nil for the Service's
	for i := range config.Service {
		if config.Service[i].Status != nil {
			t.Errorf("Status is now non-nil - %v", *config.Service[i].Status)
		}
	}
}

func testConvertCurrentVersionToDeployedVersion() Config {
	return Config{
		Service: service.Slice{
			"test": &service.Service{
				Status: &service_status.Status{
					CurrentVersion:          "1.2.3",
					CurrentVersionTimestamp: "2022-01-01T01:01:01Z",
				},
			},
			"other": &service.Service{
				Status: &service_status.Status{
					CurrentVersion:          "4.5.6",
					CurrentVersionTimestamp: "2012-01-01T01:01:01Z",
				},
			},
		},
	}
}

func TestConvertCurrentVersionToDeployedVersionCheckVersion(t *testing.T) {
	// GIVEN a Config with Service's that have nil Status.CurrentVersionTim
	had := testConvertCurrentVersionToDeployedVersion()
	changing := testConvertCurrentVersionToDeployedVersion()

	// WHEN convertCurrentVersionToDeployedVersion is called on this Config
	changing.convertCurrentVersionToDeployedVersion()

	// THEN Status stays nil for the Service's
	for i := range changing.Service {
		if (*changing.Service[i].Status).DeployedVersion != (*had.Service[i].Status).CurrentVersion {
			t.Errorf("CurrentVersion not converted to DeployedVersion. Want %s, got %s", (*had.Service[i].Status).CurrentVersion, (*&changing.Service[i].Status).DeployedVersion)
		}
	}
}

func TestConvertCurrentVersionToDeployedVersionCheckTimestamp(t *testing.T) {
	// GIVEN a Config with Service's that have nil Status
	had := testConvertCurrentVersionToDeployedVersion()
	changing := testConvertCurrentVersionToDeployedVersion()

	// WHEN convertCurrentVersionToDeployedVersion is called on this Config
	changing.convertCurrentVersionToDeployedVersion()

	// THEN Status stays nil for the Service's
	for i := range changing.Service {
		if (*changing.Service[i].Status).DeployedVersionTimestamp != (*had.Service[i].Status).CurrentVersionTimestamp {
			t.Errorf("CurrentVersionTimestamp not converted to DeployedVersionTimestamp. Want %q, got %q", (*had.Service[i].Status).CurrentVersion, (*changing.Service[i].Status).DeployedVersion)
		}
	}
}

func TestConvertCurrentVersionToDeployedVersionCheckAlreadyConvertedVersion(t *testing.T) {
	// GIVEN a Config with Service's that have already had CurrentVersion converted to DeployedVersion
	had := testConvertCurrentVersionToDeployedVersion()
	(*had.Service["test"].Status).DeployedVersion = "a.b.c"
	(*had.Service["other"].Status).DeployedVersion = "d.e.f"
	changing := testConvertCurrentVersionToDeployedVersion()
	(*changing.Service["test"].Status).DeployedVersion = "a.b.c"
	(*changing.Service["other"].Status).DeployedVersion = "d.e.f"

	// WHEN convertCurrentVersionToDeployedVersion is called on this Config
	changing.convertCurrentVersionToDeployedVersion()

	// THEN Status stays nil for the Service's
	for i := range changing.Service {
		want := (*had.Service[i].Status).DeployedVersion
		dontWant := (*had.Service[i].Status).CurrentVersion
		got := (*changing.Service[i].Status).DeployedVersion
		if got == dontWant {
			t.Errorf("CurrentVersion shouldn't have been converted to DeployedVersion again! Want %q, got %q", want, got)
		}
	}
}

func TestConvertCurrentVersionToDeployedVersionCheckAlreadyConvertedTimestamp(t *testing.T) {
	// GIVEN a Config with Service's that have already had CurrentVersionTimestamp converted to DeployedVersionTimestamp
	had := testConvertCurrentVersionToDeployedVersion()
	(*had.Service["test"].Status).DeployedVersion = "2022-02-02T02:02:02Z"
	changing := testConvertCurrentVersionToDeployedVersion()
	(*changing.Service["test"].Status).DeployedVersion = "2012-02-02T02:02:02Z"

	// WHEN convertCurrentVersionToDeployedVersion is called on this Config
	changing.convertCurrentVersionToDeployedVersion()

	// THEN Status stays nil for the Service's
	for i := range changing.Service {
		if (*changing.Service[i].Status).DeployedVersionTimestamp != (*had.Service[i].Status).CurrentVersionTimestamp {
			t.Errorf("CurrentVersionTimestamp shoouldn't have been converted to DeployedVersionTimestamp again! Want %q, got %q", (*had.Service[i].Status).DeployedVersion, (*changing.Service[i].Status).CurrentVersion)
		}
	}
}

func testConvertDeprecatedURLCommands() Config {
	return Config{
		Service: service.Slice{
			"test": &service.Service{
				URLCommands: &service.URLCommandSlice{
					service.URLCommand{Type: "replace"},
					service.URLCommand{Type: "regex_submatch", Index: 1},
					service.URLCommand{Type: "regex_submatch", Index: 1},
					service.URLCommand{Type: "split", Index: 10},
				},
			},
		},
	}
}

func TestConvertDeprecatedURLCommandsWithNil(t *testing.T) {
	// GIVEN a Config with nil URLCommands
	config := testConvertDeprecatedURLCommands()
	for i := range config.Service {
		config.Service[i].URLCommands = nil
	}

	// WHEN convertDeprecatedURLCommands is called on this Config
	config.convertDeprecatedURLCommands()

	// THEN URLCommands stay nil
	for i := range config.Service {
		if config.Service[i].URLCommands != nil {
			t.Errorf("URLCommands should be nil but are %v", *config.Service[i].URLCommands)
		}
	}
}

func TestConvertDeprecatedURLCommandsWithNonNilDidReplace(t *testing.T) {
	// GIVEN a Config with the deprecated type (regex_submatch)
	config := testConvertDeprecatedURLCommands()

	// WHEN convertDeprecatedURLCommands is called on this Config
	config.convertDeprecatedURLCommands()

	// THEN "regex_submatch" Types are replaced by "regex"
	for i := range config.Service {
		for j := range *config.Service[i].URLCommands {
			if (*config.Service[i].URLCommands)[j].Type == "regex_submatch" {
				t.Errorf("'regex_submatch' shouldn't exist anymore, but was found in %v", *config.Service[i].URLCommands)
			}
		}
	}
}

func TestConvertDeprecatedURLCommandsWithNonNilDidntReplaceNonMatches(t *testing.T) {
	// GIVEN a Config with the deprecated type (regex_submatch)
	had := testConvertDeprecatedURLCommands()
	config := testConvertDeprecatedURLCommands()

	// WHEN convertDeprecatedURLCommands is called on this Config
	config.convertDeprecatedURLCommands()

	// THEN types that are neither "regex" nor "regex_submatch" aren't changed
	for i := range config.Service {
		for j := range *config.Service[i].URLCommands {
			if (*config.Service[i].URLCommands)[j].Type == "regex" &&
				((*had.Service[i].URLCommands)[j].Type != "regex" && (*had.Service[i].URLCommands)[j].Type != "regex_submatch") {
				t.Errorf("%s was converted to %s when it isn't the target 'regex_submatch' type",
					(*had.Service[i].URLCommands)[j].Type,
					(*config.Service[i].URLCommands)[j].Type)
			}
		}
	}
}
