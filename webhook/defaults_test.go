// Copyright [2024] [Argus]
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

package webhook

import (
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestDefaults_Default(t *testing.T) {
	// GIVEN a Defaults and an expected Defaults
	defaults := Defaults{}
	expected := Defaults{
		Base: Base{
			Type:              "github",
			Delay:             "0s",
			AllowInvalidCerts: test.BoolPtr(false),
			DesiredStatusCode: test.UInt16Ptr(0),
			MaxTries:          test.UInt8Ptr(3),
			SilentFails:       test.BoolPtr(false),
		}}

	// WHEN Default is called
	defaults.Default()

	// THEN it should set the defaults
	if defaults.Type != expected.Type {
		t.Errorf("Type not set correctly, got %q, want %q",
			defaults.Type, expected.Type)
	}
	if defaults.Delay != expected.Delay {
		t.Errorf("Delay not set correctly, got %q, want %q",
			defaults.Delay, expected.Delay)
	}
	if test.StringifyPtr(defaults.AllowInvalidCerts) != test.StringifyPtr(expected.AllowInvalidCerts) {
		t.Errorf("AllowInvalidCerts not set correctly, got %q, want %q",
			test.StringifyPtr(defaults.AllowInvalidCerts), test.StringifyPtr(expected.AllowInvalidCerts))
	}
	if test.StringifyPtr(defaults.DesiredStatusCode) != test.StringifyPtr(expected.DesiredStatusCode) {
		t.Errorf("DesiredStatusCode not set correctly, got %q, want %q",
			test.StringifyPtr(defaults.DesiredStatusCode), test.StringifyPtr(expected.DesiredStatusCode))
	}
	if test.StringifyPtr(defaults.MaxTries) != test.StringifyPtr(expected.MaxTries) {
		t.Errorf("MaxTries not set correctly, got %q, want %q",
			test.StringifyPtr(defaults.MaxTries), test.StringifyPtr(expected.MaxTries))
	}
	if test.StringifyPtr(defaults.SilentFails) != test.StringifyPtr(expected.SilentFails) {
		t.Errorf("SilentFails not set correctly, got %q, want %q",
			test.StringifyPtr(defaults.SilentFails), test.StringifyPtr(expected.SilentFails))
	}
}
