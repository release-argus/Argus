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

package decode

import (
	"io"
	"strings"

	"github.com/goccy/go-yaml"
)

var (
	// yamlMarshalIndent is the standard indentation to marshal YAML with.
	yamlMarshalIndent = 2
	// YAMLMarshalOpts are the options for yaml.Marshaler.
	YAMLMarshalOpts = []yaml.EncodeOption{
		yaml.Indent(yamlMarshalIndent),
		yaml.IndentSequence(true),
		yaml.UseLiteralStyleIfMultiline(true),
		yaml.UseSingleQuote(true),
	}
)

// NewYAMLEncoder returns a new [yaml.Encoder] that writes to w.
// The Encoder should be closed after use to flush all data to w.
func NewYAMLEncoder(w io.Writer, spaces int) *yaml.Encoder {
	opts := append([]yaml.EncodeOption(nil), YAMLMarshalOpts...)

	// Override indentation.
	if spaces > yamlMarshalIndent {
		opts = append(opts, yaml.Indent(spaces))
	}

	return yaml.NewEncoder(w, opts...)
}

// ToYAMLString converts input to its YAML string representation with the given prefix.
func ToYAMLString(input any, prefix string) string {
	b, err := yaml.MarshalWithOptions(input, YAMLMarshalOpts...)
	if err != nil {
		return ""
	}

	str := string(b)

	// Prefix each line.
	if prefix != "" && str != "" && str != "{}\n" {
		str = strings.Replace(
			str,
			"\n",
			"\n"+prefix,
			strings.Count(str, "\n")-1,
		)
		str = prefix + str
	}

	return str
}
