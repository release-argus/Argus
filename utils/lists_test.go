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

package utils

import (
	"testing"
)

func TestSort(t *testing.T) {
	lstString := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	Swap(&lstString, 0, 0, 9, 9)
	wantLstString := []string{"9", "1", "2", "3", "4", "5", "6", "7", "8", "0"}
	for i := range lstString {
		if lstString[i] != wantLstString[i] {
			t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %s`, lstString, wantLstString)
		}
	}
	if len(lstString) != len(wantLstString) {
		t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %v`, lstString, wantLstString)
	}

	lstInt := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	Swap(&lstInt, 0, 0, 9, 9)
	wantLstInt := []int{9, 1, 2, 3, 4, 5, 6, 7, 8, 0}
	for i := range lstInt {
		if lstInt[i] != wantLstInt[i] {
			t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %v`, lstInt, wantLstInt)
		}
	}
	if len(lstInt) != len(wantLstInt) {
		t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %v`, lstInt, wantLstInt)
	}

	lstInt = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	Swap(&lstInt, 5, 9, 14, 15)
	wantLstInt = []int{0, 1, 2, 3, 4, 14, 15, 10, 11, 12, 13, 5, 6, 7, 8, 9, 16, 17, 18, 19, 20}
	for i := range lstInt {
		if lstInt[i] != wantLstInt[i] {
			t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %v`, lstInt, wantLstInt)
		}
	}
	if len(lstInt) != len(wantLstInt) {
		t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %v`, lstInt, wantLstInt)
	}

	lstInt = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	Swap(&lstInt, 7, 7, 14, 15)
	wantLstInt = []int{0, 1, 2, 3, 4, 5, 6, 14, 15, 8, 9, 10, 11, 12, 13, 7, 16, 17, 18, 19, 20}
	for i := range lstInt {
		if lstInt[i] != wantLstInt[i] {
			t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %v`, lstInt, wantLstInt)
		}
	}
	if len(lstInt) != len(wantLstInt) {
		t.Fatalf(`config.Defaults.Service.Interval = %v, want match for %v`, lstInt, wantLstInt)
	}
}

func TestRemoveIndex(t *testing.T) {
	lstString := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	RemoveIndex(&lstString, 4)
	wantLstString := []string{"0", "1", "2", "3", "5", "6", "7", "8", "9"}
	if len(lstString) != len(wantLstString) {
		t.Fatalf(`RemoveIndex gave %v, want match for %v`, lstString, wantLstString)
	}

	lstInt := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	RemoveIndex(&lstInt, 9)
	RemoveIndex(&lstInt, 0)
	RemoveIndex(&lstInt, 3)
	wantLstInt := []int{1, 2, 3, 5, 6, 7, 8}
	if len(lstInt) != len(wantLstInt) {
		t.Fatalf(`RemoveIndex gave %v, want match for %v`, lstInt, wantLstInt)
	}
}

func TestGetIndentation(t *testing.T) {
	gotIndentation := GetIndentation("foo: bar", 2)
	if gotIndentation != "" {
		t.Fatalf(`RemoveIndex gave %v, want match for %q`, gotIndentation, "")
	}
	gotIndentation = GetIndentation("  foo: bar", 2)
	if gotIndentation != "  " {
		t.Fatalf(`RemoveIndex gave %v, want match for %q`, gotIndentation, "  ")
	}
	gotIndentation = GetIndentation("    foo: bar", 2)
	if gotIndentation != "    " {
		t.Fatalf(`RemoveIndex gave %v, want match for %q`, gotIndentation, "    ")
	}
	gotIndentation = GetIndentation("    foo: bar", 4)
	if gotIndentation != "    " {
		t.Fatalf(`RemoveIndex gave %v, want match for %q`, gotIndentation, "    ")
	}
}
