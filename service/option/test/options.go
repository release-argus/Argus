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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
)

// PlainOptions returns service options decoded with the given defaults config.
func PlainOptions(t *testing.T, cfg opt.DefaultsConfig) *opt.Options {
	t.Helper()

	options, _ := opt.Decode(
		"yaml", nil,
		cfg,
	)

	return options
}

func Options(t *testing.T) *opt.Options {
	t.Helper()

	optCfg := PlainDefaultsConfig(t)

	options, _ := opt.Decode(
		"yaml", []byte(test.TrimYAML(`
			interval: 1m
		`)),
		optCfg,
	)
	return options
}
