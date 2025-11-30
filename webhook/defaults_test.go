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

//go:build unit

package webhook

import (
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestDefaults_Default(t *testing.T) {
	// GIVEN Defaults and an expected Defaults.
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

	// WHEN Default is called.
	defaults.Default()

	// THEN it should set the defaults.
	if defaults.Type != expected.Type {
		t.Errorf("%s\nType mismatch\nwant: %q\ngot:  %q",
			packageName, expected.Type, defaults.Type)
	}
	if defaults.Delay != expected.Delay {
		t.Errorf("%s\nDelay mismatch\nwant: %q\ngot:  %q",
			packageName, expected.Delay, defaults.Delay)
	}
	if test.StringifyPtr(defaults.AllowInvalidCerts) != test.StringifyPtr(expected.AllowInvalidCerts) {
		t.Errorf("%s\nAllowInvalidCerts mismatch\nwant: %q\ngot:  %q",
			packageName,
			test.StringifyPtr(expected.AllowInvalidCerts), test.StringifyPtr(defaults.AllowInvalidCerts))
	}
	if test.StringifyPtr(defaults.DesiredStatusCode) != test.StringifyPtr(expected.DesiredStatusCode) {
		t.Errorf("%s\nDesiredStatusCode mismatch\nwant: %q\ngot:  %q",
			packageName, test.StringifyPtr(expected.DesiredStatusCode), test.StringifyPtr(defaults.DesiredStatusCode))
	}
	if test.StringifyPtr(defaults.MaxTries) != test.StringifyPtr(expected.MaxTries) {
		t.Errorf("%s\nMaxTries mismatch\nwant: %q\ngot:  %q",
			packageName, test.StringifyPtr(expected.MaxTries), test.StringifyPtr(defaults.MaxTries))
	}
	if test.StringifyPtr(defaults.SilentFails) != test.StringifyPtr(expected.SilentFails) {
		t.Errorf("%s\nSilentFails mismatch\nwant: %q\ngot:  %q",
			packageName, test.StringifyPtr(expected.SilentFails), test.StringifyPtr(defaults.SilentFails))
	}
}
