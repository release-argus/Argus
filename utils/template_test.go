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

func testTemplateStringContext() ServiceInfo {
	return ServiceInfo{
		ID:            "something",
		URL:           "example.com",
		WebURL:        "other.com",
		LatestVersion: "NEW",
	}
}

func TestTemplateStringNoJinja(t *testing.T) {
	// GIVEN a string with no Jinja expressions/vars
	str := "test"

	// WHEN TemplateString is called
	got := TemplateString(str, testTemplateStringContext())
	want := str

	// THEN the string stays the same
	if got != want {
		t.Errorf("TemplateString didn't stay the same! Got %q, want %q",
			got, want)
	}
}

func TestTemplateStringCompilePanic(t *testing.T) {
	// GIVEN a string with an invalid Jinja expression
	str := "test{% if 'a' == 'a' %}hi{%"
	// Turn off the panic.
	defer func() { _ = recover() }()

	// WHEN TemplateString is called
	got := TemplateString(str, testTemplateStringContext())

	// THEN it should have panic'd at the compile and not reach this
	t.Errorf("TemplateString didn't panic on invalid Jinja %q. Got %s",
		str, got)
}

func TestTemplateStringJinjaVars(t *testing.T) {
	// GIVEN a string with Jinja vars
	str := "a={{ version }}, b={{ service_id }}, c={{ service_url }}, d={{ web_url }}"

	// WHEN TemplateString is called
	got := TemplateString(str, testTemplateStringContext())
	want := "a=NEW, b=something, c=example.com, d=other.com"

	// THEN the string stays the same
	if got != want {
		t.Errorf("TemplateString didn't expand %q correctly! Got %q, want %q",
			str, got, want)
	}
}

func TestTemplateStringJinjaExpressions(t *testing.T) {
	// GIVEN a string with Jinja expressions
	str := "{% if 'a' == 'a' %}it_is{% endif %}{% if 'a' == 'b' %}it_isnt{% endif %}"

	// WHEN TemplateString is called
	got := TemplateString(str, testTemplateStringContext())
	want := "it_is"

	// THEN the string stays the same
	if got != want {
		t.Errorf("TemplateString didn't eval %q correctly! Got %q, want %q",
			str, got, want)
	}
}

func TestTemplateStringJinjaExpressionsAndVars(t *testing.T) {
	// GIVEN a string with Jinja expressions and vars
	str := "{% if 'a' == 'a' %}a={{ version }}, b={{ service_id }}, c={{ service_url }}, d={{ web_url }}{% endif %}{% if 'a' == 'b' %}it_isnt{% endif %}"

	// WHEN TemplateString is called
	got := TemplateString(str, testTemplateStringContext())
	want := "a=NEW, b=something, c=example.com, d=other.com"

	// THEN the string stays the same
	if got != want {
		t.Errorf("TemplateString didn't eval %q correctly! Got %q, want %q",
			str, got, want)
	}
}
