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

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

var packageName = "auth_provider"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// fakeProvider is a stub Provider with a fixed name.
type fakeProvider struct {
	name string
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Authenticate(_ context.Context, _, _ string) (*auth.Identity, error) {
	return &auth.Identity{Provider: f.name}, nil
}
