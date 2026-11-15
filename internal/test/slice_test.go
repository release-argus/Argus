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

import (
	"reflect"
	"testing"
)

type item struct {
	Name string
}

func TestAttributes(t *testing.T) {
	tests := []struct {
		name                  string
		items                 []item
		wantSorted, wantInput []string
	}{
		{
			name:       "nil slice",
			items:      nil,
			wantSorted: []string{},
			wantInput:  []string{},
		},
		{
			name:       "empty slice",
			items:      []item{},
			wantSorted: []string{},
			wantInput:  []string{},
		},
		{
			name: "single element",
			items: []item{
				{Name: "Charlie"},
			},
			wantSorted: []string{"Charlie"},
			wantInput:  []string{"Charlie"},
		},
		{
			name: "multiple elements",
			items: []item{
				{Name: "Charlie"},
				{Name: "Alice"},
				{Name: "Bob"},
			},
			wantSorted: []string{"Alice", "Bob", "Charlie"},
			wantInput:  []string{"Charlie", "Alice", "Bob"},
		},
		{
			name: "duplicate values",
			items: []item{
				{Name: "Bob"},
				{Name: "Alice"},
				{Name: "Bob"},
			},
			wantSorted: []string{"Alice", "Bob", "Bob"},
			wantInput:  []string{"Bob", "Alice", "Bob"},
		},
		{
			name: "empty attribute values",
			items: []item{
				{Name: "Bob"},
				{Name: ""},
				{Name: "Alice"},
			},
			wantSorted: []string{"", "Alice", "Bob"},
			wantInput:  []string{"Bob", "", "Alice"},
		},
		{
			name: "case-sensitive sorting",
			items: []item{
				{Name: "bob"},
				{Name: "Alice"},
				{Name: "Bob"},
			},
			wantSorted: []string{"Alice", "Bob", "bob"},
			wantInput:  []string{"bob", "Alice", "Bob"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			funcs := []struct {
				name string
				fn   func([]item, func(item) string) []string
				want []string
			}{
				{name: "Attributes", fn: Attributes[item, string], want: tc.wantInput},
				{name: "AttributesSorted", fn: AttributesSorted[item, string], want: tc.wantSorted},
			}
			for _, f := range funcs {
				t.Run(f.name, func(t *testing.T) {
					t.Parallel()

					got := f.fn(tc.items, func(i item) string { return i.Name })

					if !reflect.DeepEqual(got, f.want) {
						t.Errorf("mismatch\ngot:  %q\nwant: %q", got, f.want)
					}
				})
			}
		})
	}
}

func TestAttributesSorted__doesNotModifyInput(t *testing.T) {
	items := []item{
		{Name: "Charlie"},
		{Name: "Alice"},
		{Name: "Bob"},
	}

	original := append([]item(nil), items...)

	AttributesSorted(items, func(i item) string { return i.Name })

	if !reflect.DeepEqual(items, original) {
		t.Errorf("input mismatch\ngot:  %v\nwant: %v", items, original)
	}
}
