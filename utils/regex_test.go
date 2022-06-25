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

package utils

import (
	"testing"
)

func TestRegexCheckPass(t *testing.T) {
	// GIVEN a string and matching RegEx
	regex := "^abc"
	str := "abc123"

	// WHEN RegexCheck is called on them
	match := RegexCheck(regex, str)

	// THEN a match is found
	if !match {
		t.Errorf("%q should have matched the %q RegEx", str, regex)
	}
}

func TestRegexCheckFail(t *testing.T) {
	// GIVEN a string and non-matching RegEx
	regex := "^abc"
	str := "123abc"

	// WHEN RegexCheck is called on them
	match := RegexCheck(regex, str)

	// THEN a match is not found
	if match {
		t.Errorf("%q shouldn't have matched the %q RegEx", str, regex)
	}
}

func TestRegexCheckWithParamsRunsOnTemplatedStringPass(t *testing.T) {
	// GIVEN a string and matching RegEx
	regex := "^abc{{ version }}$"
	str := "abc1.2.3"
	version := "1.2.3"

	// WHEN RegexCheck is called on them
	match := RegexCheckWithParams(regex, str, version)

	// THEN a match is found
	if !match {
		t.Errorf("%q (version=%s) should have matched the %q RegEx", str, version, regex)
	}
}

func TestRegexCheckWithParamsRunsOnTemplatedStringFail(t *testing.T) {
	// GIVEN a string and matching RegEx
	regex := "^abc{{ version }}$"
	str := "abc4.5.6"
	version := "1.2.3"

	// WHEN RegexCheck is called on them
	match := RegexCheckWithParams(regex, str, version)

	// THEN a match is found
	if match {
		t.Errorf("%q (version=%s) should not have matched the %q RegEx", str, version, regex)
	}
}
