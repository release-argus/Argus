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

package shoutrrr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	shoutrrr_vars "github.com/release-argus/Argus/notifiers/shoutrrr/types"
	"github.com/release-argus/Argus/util"
)

func FromPayload(
	payload *io.ReadCloser,
	original *Shoutrrr,
	logFrom *util.LogFrom,
) (shoutrrr *Shoutrrr, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(*payload); err != nil {
		return
	}

	// Notify
	shoutrrr = &Shoutrrr{}
	dec1 := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	err = dec1.Decode(shoutrrr)
	if err != nil {
		jLog.Error(err, *logFrom, true)
		jLog.Verbose(fmt.Sprintf("Payload: %s", buf.String()), *logFrom, true)
		return
	}
	if shoutrrr.Type == "" {
		err = fmt.Errorf("type is required")
		jLog.Error(err, *logFrom, true)
		return
	}

	shoutrrr.inheritSecrets(original)
	shoutrrr.Main = original.Main
	shoutrrr.Defaults = original.Defaults
	shoutrrr.HardDefaults = original.HardDefaults

	err = shoutrrr.CheckValues("")
	return
}

// inheritSecrets from original will copy all '<secret>'s referenced in `s` from `original`.
func (s *Shoutrrr) inheritSecrets(original *Shoutrrr) {
	util.CopyIfSecret(s.Params, original.Params, shoutrrr_vars.CensorableParams[:])
	util.CopyIfSecret(s.URLFields, original.URLFields, shoutrrr_vars.CensorableURLFields[:])
}
