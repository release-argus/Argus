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

//go:build unit || integration

// Package test provides test helpers for the option package.
package test

import (
	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
)

// PlainOptions returns service options decoded with the given defaults config.
func PlainOptions(cfg opt.DefaultsConfig) *opt.Options {
	options, _ := opt.Decode("yaml", nil, cfg)

	return options
}

// Options returns service options with a 1m interval, decoded with plain defaults.
func Options(t test.TB) *opt.Options {
	t.Helper()

	optCfg := PlainDefaultsConfig()

	options, err := opt.Decode(
		"yaml", []byte(test.TrimYAML(`
			interval: 1m
		`)),
		optCfg,
	)
	if err != nil {
		t.Fatalf("opt.Decode() failed: %v", err)
	}
	return options
}
