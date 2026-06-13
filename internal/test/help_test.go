// Copyright [2026] [Argus]
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

package test

import "github.com/goccy/go-yaml"

var packageName = "test"

type testStruct struct {
	String string `json:"string,omitempty" yaml:"string,omitempty"`
	Int    int    `json:"int,omitempty" yaml:"int,omitempty"`
	Bool   bool   `json:"bool,omitempty" yaml:"bool,omitempty"`
}

func (t *testStruct) Foo() string { return t.String }

func genericStringify[T any](v T) string {
	bytes, _ := yaml.Marshal(v)
	return string(bytes)
}
