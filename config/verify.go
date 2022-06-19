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

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/release-argus/Argus/utils"
)

// CheckValues are valid.
func (c *Config) CheckValues() {
	var errs error
	if err := c.Defaults.CheckValues(); err != nil {
		errs = fmt.Errorf("defaults:\\%w", err)
	}

	if err := c.Notify.CheckValues("  "); err != nil {
		errs = fmt.Errorf("%s%w", utils.ErrorToString(errs), err)
	}

	if err := c.WebHook.CheckValues("  "); err != nil {
		errs = fmt.Errorf("%s%w", utils.ErrorToString(errs), err)
	}

	if err := c.Service.CheckValues("  "); err != nil {
		errs = fmt.Errorf("%sservice:\\%w", utils.ErrorToString(errs), err)
	}

	if errs != nil {
		fmt.Println(strings.Replace(errs.Error(), "\\", "\n", -1))
		fmt.Println("\nERROR: Config could not be parsed successfully.")
		os.Exit(1)
	}
}

// Print the parsed config if *flag.
func (c *Config) Print(flag *bool) {
	if !*flag {
		return
	}

	c.Service.Print("", *c.Order)
	fmt.Println()
	c.Notify.Print("")
	fmt.Println()
	c.WebHook.Print("")
	fmt.Println()
	c.Defaults.Print()
	os.Exit(0)
}
