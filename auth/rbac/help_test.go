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

package rbac

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

var packageName = "auth_rbac"

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

// globalGrant builds a valid global-scope Grant.
func globalGrant(resource Resource, action Action) Grant {
	return Grant{
		Permission: Permission{Resource: resource, Action: action},
		Scope:      Scope{Type: ScopeGlobal},
	}
}

// serviceGrant builds a valid service-scope Grant.
func serviceGrant(resource Resource, action Action, serviceID string) Grant {
	return Grant{
		Permission: Permission{Resource: resource, Action: action},
		Scope:      Scope{Type: ScopeService, Ref: serviceID},
	}
}

// tagGrant builds a valid service_tag-scope Grant.
func tagGrant(resource Resource, action Action, tag string) Grant {
	return Grant{
		Permission: Permission{Resource: resource, Action: action},
		Scope:      Scope{Type: ScopeServiceTag, Ref: tag},
	}
}
