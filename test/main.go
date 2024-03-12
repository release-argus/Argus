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

//go:build unit || integration

package test

import (
	"fmt"
	"strings"

	"github.com/release-argus/Argus/util"
)

func BoolPtr(val bool) *bool {
	return &val
}
func IntPtr(val int) *int {
	return &val
}
func StringPtr(val string) *string {
	return &val
}
func UIntPtr(val int) *uint {
	converted := uint(val)
	return &converted
}
func StringifyPtr[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}

func CopyMapPtr(tgt map[string]string) *map[string]string {
	ptr := util.CopyMap(tgt)
	return &ptr
}

func TrimJSON(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `": `, `":`)
	str = strings.ReplaceAll(str, `", `, `",`)
	str = strings.ReplaceAll(str, `, "`, `,"`)
	return str
}
